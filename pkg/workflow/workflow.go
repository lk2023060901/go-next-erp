package workflow

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
)

// Engine 工作流引擎
type Engine struct {
	config      *Config
	logger      *logger.Logger
	registry    *Registry
	evaluator   *ConditionEvaluator
	ctxMgr      *ContextManager
	executor    *Executor
	persistence PersistenceProvider

	// 工作流定义存储
	workflows sync.Map // workflowID -> *WorkflowDefinition

	// 中间件
	middlewares []Middleware

	// 执行指标
	metrics map[string]*ExecutionMetrics
	metricsMu sync.RWMutex

	// 运行状态
	running atomic.Bool
	mu      sync.RWMutex
}

// New 创建工作流引擎
func New(opts ...Option) (*Engine, error) {
	// 创建默认日志器
	defaultLogger, _ := logger.New()

	e := &Engine{
		config:      DefaultConfig(),
		logger:      defaultLogger,
		registry:    NewRegistry(),
		evaluator:   NewConditionEvaluator(),
		ctxMgr:      NewContextManager(),
		persistence: NewNopPersistence(), // 默认使用空持久化
		metrics:     make(map[string]*ExecutionMetrics),
	}

	// 应用选项
	for _, opt := range opts {
		opt(e)
	}

	// 验证配置
	if err := e.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建执行器
	e.executor = NewExecutor(e)

	// 注册内置节点类型
	if err := e.registerBuiltinNodes(); err != nil {
		return nil, fmt.Errorf("failed to register builtin nodes: %w", err)
	}

	return e, nil
}

// registerBuiltinNodes 注册内置节点类型
func (e *Engine) registerBuiltinNodes() error {
	// 这里将在实现 nodes/ 包后注册
	// 示例:
	// e.registry.Register("trigger", nodes.NewTriggerNode)
	// e.registry.Register("http", nodes.NewHTTPNode)
	// e.registry.Register("condition", nodes.NewConditionNode)
	return nil
}

// RegisterNodeType 注册自定义节点类型
func (e *Engine) RegisterNodeType(nodeType string, factory NodeFactory) error {
	return e.registry.Register(nodeType, factory)
}

// CreateWorkflow 创建工作流
func (e *Engine) CreateWorkflow(def *WorkflowDefinition) error {
	// 验证工作流定义
	if err := e.validateWorkflow(def); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidWorkflowDef, err)
	}

	// 检查是否已存在
	if _, exists := e.workflows.Load(def.ID); exists {
		return ErrWorkflowAlreadyExists
	}

	// 设置时间戳
	now := time.Now()
	if def.CreatedAt.IsZero() {
		def.CreatedAt = now
	}
	def.UpdatedAt = now

	// 设置默认版本
	if def.Version == 0 {
		def.Version = 1
	}

	// 设置默认配置
	if def.Settings == nil {
		def.Settings = &WorkflowSettings{
			ExecutionTimeout: e.config.DefaultExecutionTimeout,
			MaxRetries:       e.config.DefaultMaxRetries,
			RetryDelay:       e.config.DefaultRetryDelay,
			OnError:          "stop",
		}
	}

	// 保存工作流到内存
	e.workflows.Store(def.ID, def)

	// 持久化
	if e.config.EnablePersistence {
		ctx := context.Background()
		if err := e.persistence.SaveWorkflow(ctx, def); err != nil {
			e.logger.Errorw("failed to persist workflow", "error", err)
			// 不返回错误，允许继续（已保存到内存）
		}
	}

	// 初始化指标
	if e.config.EnableMetrics {
		e.metricsMu.Lock()
		e.metrics[def.ID] = &ExecutionMetrics{}
		e.metricsMu.Unlock()
	}

	e.logger.Infow("workflow created",
		"workflow_id", def.ID,
		"name", def.Name,
		"version", def.Version,
		"nodes", len(def.Nodes),
	)

	return nil
}

// UpdateWorkflow 更新工作流
func (e *Engine) UpdateWorkflow(def *WorkflowDefinition) error {
	// 验证工作流定义
	if err := e.validateWorkflow(def); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidWorkflowDef, err)
	}

	// 检查是否存在
	existing, exists := e.workflows.Load(def.ID)
	if !exists {
		return ErrWorkflowNotFound
	}

	// 增加版本号
	oldDef := existing.(*WorkflowDefinition)
	def.Version = oldDef.Version + 1
	def.CreatedAt = oldDef.CreatedAt
	def.UpdatedAt = time.Now()

	// 更新工作流
	e.workflows.Store(def.ID, def)

	// 清空表达式缓存
	e.evaluator.ClearCache()

	e.logger.Infow("workflow updated",
		"workflow_id", def.ID,
		"version", def.Version,
	)

	return nil
}

// DeleteWorkflow 删除工作流
func (e *Engine) DeleteWorkflow(workflowID string) error {
	if _, exists := e.workflows.Load(workflowID); !exists {
		return ErrWorkflowNotFound
	}

	e.workflows.Delete(workflowID)

	// 清理指标
	e.metricsMu.Lock()
	delete(e.metrics, workflowID)
	e.metricsMu.Unlock()

	e.logger.Infow("workflow deleted", "workflow_id", workflowID)

	return nil
}

// GetWorkflow 获取工作流定义
func (e *Engine) GetWorkflow(workflowID string) (*WorkflowDefinition, error) {
	value, exists := e.workflows.Load(workflowID)
	if !exists {
		return nil, ErrWorkflowNotFound
	}

	return value.(*WorkflowDefinition), nil
}

// ListWorkflows 列出所有工作流
func (e *Engine) ListWorkflows() []*WorkflowDefinition {
	var workflows []*WorkflowDefinition

	e.workflows.Range(func(key, value interface{}) bool {
		workflows = append(workflows, value.(*WorkflowDefinition))
		return true
	})

	return workflows
}

// Execute 执行工作流
//
// workflowID: 工作流 ID
// input: 输入数据
// triggerBy: 触发者标识
//
// 返回: 执行 ID 和错误
func (e *Engine) Execute(ctx context.Context, workflowID string, input map[string]interface{}, triggerBy string) (string, error) {
	// 获取工作流定义
	def, err := e.GetWorkflow(workflowID)
	if err != nil {
		return "", err
	}

	// 检查工作流状态
	if def.Status != WorkflowStatusActive {
		return "", ErrWorkflowInvalidState
	}

	// 生成执行 ID
	executionID := uuid.New().String()

	// 创建执行上下文
	execCtx := NewExecutionContext(workflowID, executionID, triggerBy, input)

	// 复制全局变量到执行上下文
	for k, v := range def.Variables {
		execCtx.SetVariable(k, v)
	}

	// 保存到上下文管理器
	e.ctxMgr.Store(execCtx)

	// 异步执行工作流
	go func() {
		e.executeWorkflow(ctx, def, execCtx)
	}()

	e.logger.Infow("workflow execution started",
		"workflow_id", workflowID,
		"execution_id", executionID,
		"trigger_by", triggerBy,
	)

	return executionID, nil
}

// ExecuteSync 同步执行工作流（阻塞直到完成）
func (e *Engine) ExecuteSync(ctx context.Context, workflowID string, input map[string]interface{}, triggerBy string) (*ExecutionContext, error) {
	// 获取工作流定义
	def, err := e.GetWorkflow(workflowID)
	if err != nil {
		return nil, err
	}

	// 检查工作流状态
	if def.Status != WorkflowStatusActive {
		return nil, ErrWorkflowInvalidState
	}

	// 生成执行 ID
	executionID := uuid.New().String()

	// 创建执行上下文
	execCtx := NewExecutionContext(workflowID, executionID, triggerBy, input)

	// 复制全局变量
	for k, v := range def.Variables {
		execCtx.SetVariable(k, v)
	}

	// 保存到上下文管理器
	e.ctxMgr.Store(execCtx)
	defer e.ctxMgr.Delete(executionID)

	// 同步执行
	e.executeWorkflow(ctx, def, execCtx)

	return execCtx, nil
}

// GetExecution 获取执行上下文
func (e *Engine) GetExecution(executionID string) (*ExecutionContext, error) {
	execCtx, ok := e.ctxMgr.Load(executionID)
	if !ok {
		return nil, ErrExecutionNotFound
	}

	return execCtx, nil
}

// CancelExecution 取消执行
func (e *Engine) CancelExecution(executionID string) error {
	execCtx, err := e.GetExecution(executionID)
	if err != nil {
		return err
	}

	if execCtx.Status == ExecutionStatusCompleted || execCtx.Status == ExecutionStatusFailed {
		return ErrExecutionAlreadyDone
	}

	execCtx.MarkCancelled()

	e.logger.Infow("workflow execution cancelled",
		"execution_id", executionID,
	)

	return nil
}

// Use 添加全局中间件
func (e *Engine) Use(middlewares ...Middleware) {
	e.middlewares = append(e.middlewares, middlewares...)
}

// GetStats 获取工作流统计
func (e *Engine) GetStats(workflowID string) (*WorkflowStats, error) {
	if !e.config.EnableMetrics {
		return nil, fmt.Errorf("metrics not enabled")
	}

	e.metricsMu.RLock()
	metrics, exists := e.metrics[workflowID]
	e.metricsMu.RUnlock()

	if !exists {
		return nil, ErrWorkflowNotFound
	}

	total := metrics.TotalExecutions.Load()
	success := metrics.SuccessExecutions.Load()
	failed := metrics.FailedExecutions.Load()
	avgDuration := metrics.AverageDuration.Load()

	return &WorkflowStats{
		WorkflowID:        workflowID,
		TotalExecutions:   total,
		SuccessExecutions: success,
		FailedExecutions:  failed,
		AverageDuration:   time.Duration(avgDuration) * time.Millisecond,
	}, nil
}

// validateWorkflow 验证工作流定义
func (e *Engine) validateWorkflow(def *WorkflowDefinition) error {
	if def.ID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	if def.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(def.Nodes) == 0 {
		return fmt.Errorf("workflow must have at least one node")
	}

	// 验证节点
	nodeIDs := make(map[string]bool)
	for _, node := range def.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node ID is required")
		}

		if nodeIDs[node.ID] {
			return fmt.Errorf("duplicate node ID: %s", node.ID)
		}
		nodeIDs[node.ID] = true

		// 检查节点类型是否已注册
		if !e.registry.HasType(node.Type) {
			return fmt.Errorf("node type not registered: %s", node.Type)
		}

		// 验证节点配置
		nodeInstance, err := e.registry.Create(node)
		if err != nil {
			return fmt.Errorf("invalid node %s: %w", node.ID, err)
		}

		if err := nodeInstance.Validate(); err != nil {
			return fmt.Errorf("node %s validation failed: %w", node.ID, err)
		}
	}

	// 验证边
	for _, edge := range def.Edges {
		if !nodeIDs[edge.Source] {
			return fmt.Errorf("edge source node not found: %s", edge.Source)
		}

		if !nodeIDs[edge.Target] {
			return fmt.Errorf("edge target node not found: %s", edge.Target)
		}

		// 验证条件表达式
		if edge.Condition != "" {
			if err := e.evaluator.ValidateExpression(edge.Condition); err != nil {
				return fmt.Errorf("invalid edge condition: %w", err)
			}
		}
	}

	// TODO: 检测循环依赖

	return nil
}

// executeWorkflow 执行工作流（内部实现）
func (e *Engine) executeWorkflow(ctx context.Context, def *WorkflowDefinition, execCtx *ExecutionContext) {
	// 更新状态
	execCtx.Status = ExecutionStatusRunning

	// 记录开始时间
	startTime := time.Now()

	// 设置超时
	timeout := def.Settings.ExecutionTimeout
	if timeout == 0 {
		timeout = e.config.DefaultExecutionTimeout
	}

	execCtxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 执行工作流（这里简化实现，完整实现在 executor.go）
	err := e.runWorkflow(execCtxWithTimeout, def, execCtx)

	// 更新执行结果
	duration := time.Since(startTime)

	if err != nil {
		execCtx.MarkFailed(err)
		e.logger.Errorw("workflow execution failed",
			"execution_id", execCtx.ID,
			"error", err,
			"duration", duration,
		)
	} else {
		execCtx.MarkCompleted()
		e.logger.Infow("workflow execution completed",
			"execution_id", execCtx.ID,
			"duration", duration,
		)
	}

	// 更新指标
	if e.config.EnableMetrics {
		e.updateMetrics(def.ID, execCtx.Status, duration)
	}
}

// runWorkflow 运行工作流逻辑
func (e *Engine) runWorkflow(ctx context.Context, def *WorkflowDefinition, execCtx *ExecutionContext) error {
	// 使用执行器执行工作流
	return e.executor.Execute(ctx, def, execCtx)
}

// updateMetrics 更新执行指标
func (e *Engine) updateMetrics(workflowID string, status ExecutionStatus, duration time.Duration) {
	e.metricsMu.RLock()
	metrics, exists := e.metrics[workflowID]
	e.metricsMu.RUnlock()

	if !exists {
		return
	}

	metrics.TotalExecutions.Add(1)

	if status == ExecutionStatusCompleted {
		metrics.SuccessExecutions.Add(1)
	} else if status == ExecutionStatusFailed {
		metrics.FailedExecutions.Add(1)
	}

	// 更新平均耗时（简化计算）
	currentAvg := metrics.AverageDuration.Load()
	total := metrics.TotalExecutions.Load()
	newAvg := (currentAvg*int64(total-1) + duration.Milliseconds()) / int64(total)
	metrics.AverageDuration.Store(newAvg)
}
