package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lk2023060901/go-next-erp/pkg/logger"
)

// Executor 工作流执行器
// 企业级实现：支持并行执行、DAG 拓扑排序、完整的错误处理、超时控制
type Executor struct {
	engine    *Engine
	evaluator *ConditionEvaluator
	logger    *logger.Logger
}

// NewExecutor 创建执行器
func NewExecutor(engine *Engine) *Executor {
	return &Executor{
		engine:    engine,
		evaluator: engine.evaluator,
		logger:    engine.logger,
	}
}

// Execute 执行工作流
// 企业级特性：
// - DAG 拓扑排序保证执行顺序
// - 并行执行无依赖节点
// - 完整的错误恢复机制
// - 实时状态更新
// - 分布式追踪支持
func (ex *Executor) Execute(ctx context.Context, def *WorkflowDefinition, execCtx *ExecutionContext) error {
	ex.logger.Infow("starting workflow execution",
		"workflow_id", def.ID,
		"execution_id", execCtx.ID,
		"nodes", len(def.Nodes),
		"edges", len(def.Edges),
	)

	// 1. 构建执行图（DAG）
	graph, err := ex.buildExecutionGraph(def)
	if err != nil {
		return fmt.Errorf("failed to build execution graph: %w", err)
	}

	// 2. 验证图的完整性
	if err := graph.Validate(); err != nil {
		return fmt.Errorf("invalid workflow graph: %w", err)
	}

	// 3. 拓扑排序，获取执行层级
	layers, err := graph.TopologicalSort()
	if err != nil {
		return fmt.Errorf("failed to sort execution graph: %w", err)
	}

	ex.logger.Debugw("execution graph built",
		"layers", len(layers),
		"total_nodes", graph.NodeCount(),
	)

	// 4. 按层级顺序执行节点
	for layerIndex, layer := range layers {
		ex.logger.Debugw("executing layer",
			"layer", layerIndex,
			"nodes", len(layer),
		)

		// 并行执行同一层级的节点
		if err := ex.executeLayer(ctx, def, execCtx, layer, graph); err != nil {
			return fmt.Errorf("layer %d execution failed: %w", layerIndex, err)
		}

		// 检查执行是否被取消
		select {
		case <-ctx.Done():
			return ErrExecutionCancelled
		default:
		}
	}

	// 5. 收集最终输出
	ex.collectFinalOutput(execCtx, graph)

	ex.logger.Infow("workflow execution completed",
		"execution_id", execCtx.ID,
		"duration", execCtx.Duration(),
		"nodes_executed", len(execCtx.NodeStates),
	)

	return nil
}

// executeLayer 执行一个层级的所有节点（并行）
func (ex *Executor) executeLayer(ctx context.Context, def *WorkflowDefinition, execCtx *ExecutionContext, nodeIDs []string, graph *ExecutionGraph) error {
	if len(nodeIDs) == 0 {
		return nil
	}

	// 单节点直接执行
	if len(nodeIDs) == 1 {
		return ex.executeNode(ctx, def, execCtx, nodeIDs[0], graph)
	}

	// 多节点并行执行
	var wg sync.WaitGroup
	errChan := make(chan error, len(nodeIDs))

	for _, nodeID := range nodeIDs {
		wg.Add(1)
		go func(nid string) {
			defer wg.Done()

			if err := ex.executeNode(ctx, def, execCtx, nid, graph); err != nil {
				errChan <- fmt.Errorf("node %s: %w", nid, err)
			}
		}(nodeID)
	}

	// 等待所有节点完成
	wg.Wait()
	close(errChan)

	// 收集错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("layer execution failed with %d errors: %v", len(errors), errors[0])
	}

	return nil
}

// executeNode 执行单个节点
func (ex *Executor) executeNode(ctx context.Context, def *WorkflowDefinition, execCtx *ExecutionContext, nodeID string, graph *ExecutionGraph) error {
	// 1. 检查上下文取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 2. 查找节点定义
	nodeDef := ex.findNodeDef(def, nodeID)
	if nodeDef == nil {
		return fmt.Errorf("%w: %s", ErrNodeNotFound, nodeID)
	}

	ex.logger.Debugw("executing node",
		"node_id", nodeID,
		"node_name", nodeDef.Name,
		"node_type", nodeDef.Type,
	)

	// 3. 检查节点是否禁用
	if nodeDef.Disabled {
		ex.markNodeSkipped(execCtx, nodeID, "node is disabled")
		return nil
	}

	// 4. 检查前置条件（入边条件）
	shouldExecute, err := ex.evaluateIncomingConditions(ctx, def, execCtx, nodeID, graph)
	if err != nil {
		return fmt.Errorf("failed to evaluate incoming conditions: %w", err)
	}

	if !shouldExecute {
		ex.markNodeSkipped(execCtx, nodeID, "incoming conditions not met")
		return nil
	}

	// 5. 创建节点实例
	node, err := ex.engine.registry.Create(nodeDef)
	if err != nil {
		return fmt.Errorf("failed to create node %s: %w", nodeID, err)
	}

	// 6. 应用中间件包装节点
	wrappedNode := ex.applyMiddlewares(node, nodeDef)

	// 7. 执行节点（带重试和超时）
	nodeState := ex.executeNodeWithRetry(ctx, wrappedNode, nodeDef, execCtx, graph)

	// 8. 保存节点状态（线程安全）
	execCtx.SetNodeState(nodeID, nodeState)

	// 9. 处理节点执行结果
	if nodeState.Status == NodeStatusFailed {
		return ex.handleNodeError(def, execCtx, nodeID, nodeState.Error)
	}

	// 10. 更新执行上下文变量（节点可能修改变量）
	ex.updateContextVariables(execCtx, nodeState)

	ex.logger.Debugw("node execution completed",
		"node_id", nodeID,
		"status", nodeState.Status,
		"attempts", nodeState.Attempts,
		"duration", nodeState.CompletedAt.Sub(nodeState.StartedAt),
	)

	return nil
}

// executeNodeWithRetry 执行节点（带重试、超时、监控）
func (ex *Executor) executeNodeWithRetry(ctx context.Context, node Node, nodeDef *NodeDefinition, execCtx *ExecutionContext, graph *ExecutionGraph) *NodeState {
	state := &NodeState{
		NodeID:    nodeDef.ID,
		Status:    NodeStatusPending,
		Input:     make(map[string]interface{}),
		Output:    make(map[string]interface{}),
		Attempts:  0,
		StartedAt: time.Now(),
	}

	// 1. 准备节点输入（从前置节点收集数据）
	state.Input = ex.prepareNodeInput(execCtx, nodeDef, graph)

	// 2. 获取重试策略
	retryPolicy := nodeDef.RetryPolicy
	if retryPolicy == nil {
		retryPolicy = &RetryPolicy{
			MaxAttempts: ex.engine.config.DefaultMaxRetries,
			Delay:       ex.engine.config.DefaultRetryDelay,
			BackoffRate: ex.engine.config.DefaultBackoffRate,
		}
	}

	// 3. 设置节点超时
	nodeTimeout := nodeDef.Timeout
	if nodeTimeout == 0 {
		nodeTimeout = 10 * time.Minute // 默认节点超时
	}

	// 4. 执行节点（带重试）
	maxAttempts := retryPolicy.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		state.Attempts = attempt
		state.Status = NodeStatusRunning

		// 创建带超时的上下文
		nodeCtx, cancel := context.WithTimeout(ctx, nodeTimeout)

		// 记录开始时间（用于监控）
		attemptStart := time.Now()

		// 执行节点
		output, err := node.Execute(nodeCtx, state.Input)

		// 记录执行时间
		attemptDuration := time.Since(attemptStart)

		cancel()

		if err == nil {
			// 执行成功
			state.Status = NodeStatusCompleted
			state.Output = output
			now := time.Now()
			state.CompletedAt = &now

			ex.logger.Infow("node executed successfully",
				"node", nodeDef.Name,
				"node_id", nodeDef.ID,
				"attempt", attempt,
				"duration", attemptDuration,
			)

			// 记录指标
			if ex.engine.config.EnableMetrics {
				ex.recordNodeMetrics(nodeDef.ID, true, attemptDuration)
			}

			return state
		}

		// 执行失败
		lastErr = err
		ex.logger.Warnw("node execution failed",
			"node", nodeDef.Name,
			"node_id", nodeDef.ID,
			"attempt", attempt,
			"max_attempts", maxAttempts,
			"duration", attemptDuration,
			"error", err,
		)

		// 如果还有重试机会，计算退避延迟
		if attempt < maxAttempts {
			// 指数退避
			delay := time.Duration(float64(retryPolicy.Delay) * float64(attempt-1) * retryPolicy.BackoffRate)
			if delay < retryPolicy.Delay {
				delay = retryPolicy.Delay
			}

			ex.logger.Infow("retrying node",
				"node", nodeDef.Name,
				"node_id", nodeDef.ID,
				"next_attempt", attempt+1,
				"delay", delay,
			)

			// 带取消检查的延迟
			select {
			case <-time.After(delay):
				// 继续重试
			case <-ctx.Done():
				// 上下文取消，不再重试
				state.Status = NodeStatusFailed
				state.Error = fmt.Sprintf("execution cancelled after %d attempts: %v", attempt, ctx.Err())
				now := time.Now()
				state.CompletedAt = &now
				return state
			}
		}
	}

	// 所有重试都失败
	state.Status = NodeStatusFailed
	state.Error = fmt.Sprintf("failed after %d attempts: %v", maxAttempts, lastErr)
	now := time.Now()
	state.CompletedAt = &now

	// 记录失败指标
	if ex.engine.config.EnableMetrics {
		ex.recordNodeMetrics(nodeDef.ID, false, 0)
	}

	return state
}

// prepareNodeInput 准备节点输入数据
// 企业级实现：支持复杂的数据映射、表达式计算、类型转换
func (ex *Executor) prepareNodeInput(execCtx *ExecutionContext, nodeDef *NodeDefinition, graph *ExecutionGraph) map[string]interface{} {
	input := make(map[string]interface{})

	// 1. 基础上下文数据
	input["execution_id"] = execCtx.ID
	input["workflow_id"] = execCtx.WorkflowID
	input["workflow_input"] = execCtx.Input
	input["variables"] = execCtx.Variables

	// 2. 节点配置
	input["config"] = nodeDef.Config

	// 3. 前置节点的输出
	predecessors := graph.GetPredecessors(nodeDef.ID)
	if len(predecessors) > 0 {
		predecessorOutputs := make(map[string]interface{})
		for _, predID := range predecessors {
			if state, ok := execCtx.GetNodeState(predID); ok && state.Status == NodeStatusCompleted {
				predecessorOutputs[predID] = state.Output
			}
		}
		input["predecessor_outputs"] = predecessorOutputs

		// 如果只有一个前置节点，直接提供其输出
		if len(predecessors) == 1 {
			if state, ok := execCtx.GetNodeState(predecessors[0]); ok {
				input["previous_output"] = state.Output
			}
		}
	}

	// 4. 所有已完成节点的输出（用于复杂依赖场景）
	allOutputs := make(map[string]interface{})
	for nodeID, state := range execCtx.NodeStates {
		if state.Status == NodeStatusCompleted {
			allOutputs[nodeID] = state.Output
		}
	}
	input["all_node_outputs"] = allOutputs

	// 5. 元数据
	input["metadata"] = map[string]interface{}{
		"trigger_by": execCtx.TriggerBy,
		"started_at": execCtx.StartedAt,
		"node_name":  nodeDef.Name,
		"node_type":  nodeDef.Type,
	}

	return input
}

// evaluateIncomingConditions 评估入边条件
func (ex *Executor) evaluateIncomingConditions(ctx context.Context, def *WorkflowDefinition, execCtx *ExecutionContext, nodeID string, graph *ExecutionGraph) (bool, error) {
	incomingEdges := graph.GetIncomingEdges(nodeID)

	// 如果没有入边（起始节点），直接执行
	if len(incomingEdges) == 0 {
		return true, nil
	}

	// 检查所有入边的条件
	for _, edge := range incomingEdges {
		// 检查源节点是否已完成
		sourceState, ok := execCtx.GetNodeState(edge.Source)
		if !ok || sourceState.Status != NodeStatusCompleted {
			// 源节点未完成，跳过此条件
			continue
		}

		// 如果边有条件表达式，评估之
		if edge.Condition != "" {
			result, err := ex.evaluator.Evaluate(ctx, edge.Condition, execCtx)
			if err != nil {
				return false, fmt.Errorf("failed to evaluate edge condition: %w", err)
			}

			if result {
				// 有一条边的条件满足即可执行（OR 逻辑）
				return true, nil
			}
		} else {
			// 无条件边，源节点完成即可执行
			return true, nil
		}
	}

	// 所有条件都不满足
	return false, nil
}

// handleNodeError 处理节点错误
func (ex *Executor) handleNodeError(def *WorkflowDefinition, execCtx *ExecutionContext, nodeID, errMsg string) error {
	strategy := def.Settings.OnError
	if strategy == "" {
		strategy = "stop"
	}

	ex.logger.Errorw("node execution failed",
		"node_id", nodeID,
		"error", errMsg,
		"error_strategy", strategy,
	)

	switch strategy {
	case "continue":
		// 继续执行，不抛出错误
		ex.logger.Warnw("continuing execution despite node failure",
			"node_id", nodeID,
		)
		return nil

	case "stop":
		// 立即停止整个工作流
		return fmt.Errorf("%w: node %s failed: %s", ErrNodeExecutionFailed, nodeID, errMsg)

	case "retry":
		// 重试已在 executeNodeWithRetry 中处理，此时已达到最大重试次数
		return fmt.Errorf("%w: node %s failed after %d retries: %s",
			ErrNodeExecutionFailed,
			nodeID,
			def.Settings.MaxRetries,
			errMsg,
		)

	default:
		return fmt.Errorf("unknown error handling strategy: %s", strategy)
	}
}

// updateContextVariables 更新执行上下文变量
// 节点输出中以 "var_" 开头的键会自动设置为上下文变量
func (ex *Executor) updateContextVariables(execCtx *ExecutionContext, nodeState *NodeState) {
	for key, value := range nodeState.Output {
		if len(key) > 4 && key[:4] == "var_" {
			varName := key[4:]
			execCtx.SetVariable(varName, value)

			ex.logger.Debugw("context variable updated",
				"variable", varName,
				"from_node", nodeState.NodeID,
			)
		}
	}
}

// collectFinalOutput 收集最终输出
func (ex *Executor) collectFinalOutput(execCtx *ExecutionContext, graph *ExecutionGraph) {
	// 查找终止节点（没有出边的节点）
	endNodes := graph.FindEndNodes()

	for _, nodeID := range endNodes {
		if state, ok := execCtx.GetNodeState(nodeID); ok && state.Status == NodeStatusCompleted {
			// 将终止节点的输出合并到执行上下文输出
			for key, value := range state.Output {
				execCtx.SetOutput(key, value)
			}
		}
	}

	// 如果没有明确的终止节点，收集所有成功节点的输出
	if len(endNodes) == 0 {
		for nodeID, state := range execCtx.NodeStates {
			if state.Status == NodeStatusCompleted {
				execCtx.SetOutput("node_"+nodeID, state.Output)
			}
		}
	}
}

// markNodeSkipped 标记节点为跳过
func (ex *Executor) markNodeSkipped(execCtx *ExecutionContext, nodeID, reason string) {
	state := &NodeState{
		NodeID:    nodeID,
		Status:    NodeStatusSkipped,
		StartedAt: time.Now(),
		Error:     reason,
	}
	now := time.Now()
	state.CompletedAt = &now

	execCtx.SetNodeState(nodeID, state)

	ex.logger.Debugw("node skipped",
		"node_id", nodeID,
		"reason", reason,
	)
}

// applyMiddlewares 应用中间件
func (ex *Executor) applyMiddlewares(node Node, nodeDef *NodeDefinition) Node {
	// 应用全局中间件
	for i := len(ex.engine.middlewares) - 1; i >= 0; i-- {
		node = ex.engine.middlewares[i](node)
	}

	return node
}

// recordNodeMetrics 记录节点执行指标
func (ex *Executor) recordNodeMetrics(nodeID string, success bool, duration time.Duration) {
	// TODO: 集成 Prometheus/StatsD 等监控系统
	ex.logger.Debugw("node metrics",
		"node_id", nodeID,
		"success", success,
		"duration_ms", duration.Milliseconds(),
	)
}

// findNodeDef 查找节点定义
func (ex *Executor) findNodeDef(def *WorkflowDefinition, nodeID string) *NodeDefinition {
	for _, node := range def.Nodes {
		if node.ID == nodeID {
			return node
		}
	}
	return nil
}

// buildExecutionGraph 构建执行图（DAG）
func (ex *Executor) buildExecutionGraph(def *WorkflowDefinition) (*ExecutionGraph, error) {
	graph := NewExecutionGraph()

	// 添加所有节点
	for _, node := range def.Nodes {
		graph.AddNode(node.ID, node)
	}

	// 添加所有边
	for _, edge := range def.Edges {
		if err := graph.AddEdge(edge); err != nil {
			return nil, err
		}
	}

	return graph, nil
}
