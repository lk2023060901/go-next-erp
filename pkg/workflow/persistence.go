package workflow

import (
	"context"
	"time"
)

// PersistenceProvider 持久化提供者接口
// 支持多种后端：PostgreSQL, MongoDB, Redis 等
type PersistenceProvider interface {
	// 工作流定义持久化
	SaveWorkflow(ctx context.Context, def *WorkflowDefinition) error
	GetWorkflow(ctx context.Context, workflowID string) (*WorkflowDefinition, error)
	ListWorkflows(ctx context.Context, filter *WorkflowFilter) ([]*WorkflowDefinition, error)
	DeleteWorkflow(ctx context.Context, workflowID string) error
	UpdateWorkflowStatus(ctx context.Context, workflowID string, status WorkflowStatus) error

	// 执行上下文持久化
	SaveExecution(ctx context.Context, execCtx *ExecutionContext) error
	GetExecution(ctx context.Context, executionID string) (*ExecutionContext, error)
	ListExecutions(ctx context.Context, filter *ExecutionFilter) ([]*ExecutionContext, error)
	DeleteExecution(ctx context.Context, executionID string) error
	UpdateExecutionStatus(ctx context.Context, executionID string, status ExecutionStatus) error

	// 节点状态持久化
	SaveNodeState(ctx context.Context, executionID string, state *NodeState) error
	GetNodeStates(ctx context.Context, executionID string) (map[string]*NodeState, error)

	// 统计和查询
	GetWorkflowStats(ctx context.Context, workflowID string, timeRange *TimeRange) (*WorkflowStats, error)
	GetExecutionHistory(ctx context.Context, workflowID string, limit int) ([]*ExecutionSummary, error)

	// 清理
	CleanupOldExecutions(ctx context.Context, olderThan time.Time) (int, error)

	// 健康检查
	Ping(ctx context.Context) error
	Close() error
}

// WorkflowFilter 工作流过滤器
type WorkflowFilter struct {
	Status    []WorkflowStatus
	CreatedBy string
	Search    string // 搜索名称或描述
	Limit     int
	Offset    int
	SortBy    string // created_at, updated_at, name
	SortOrder string // asc, desc
}

// ExecutionFilter 执行过滤器
type ExecutionFilter struct {
	WorkflowID string
	Status     []ExecutionStatus
	TriggerBy  string
	StartTime  *time.Time
	EndTime    *time.Time
	Limit      int
	Offset     int
	SortBy     string // started_at, completed_at, duration
	SortOrder  string // asc, desc
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// ExecutionSummary 执行摘要（用于历史记录）
type ExecutionSummary struct {
	ID          string
	WorkflowID  string
	Status      ExecutionStatus
	TriggerBy   string
	StartedAt   time.Time
	CompletedAt *time.Time
	Duration    time.Duration
	Error       string
}

// NopPersistence 空操作持久化（默认实现，不保存数据）
type NopPersistence struct{}

func NewNopPersistence() *NopPersistence {
	return &NopPersistence{}
}

func (n *NopPersistence) SaveWorkflow(ctx context.Context, def *WorkflowDefinition) error {
	return nil
}

func (n *NopPersistence) GetWorkflow(ctx context.Context, workflowID string) (*WorkflowDefinition, error) {
	return nil, ErrWorkflowNotFound
}

func (n *NopPersistence) ListWorkflows(ctx context.Context, filter *WorkflowFilter) ([]*WorkflowDefinition, error) {
	return []*WorkflowDefinition{}, nil
}

func (n *NopPersistence) DeleteWorkflow(ctx context.Context, workflowID string) error {
	return nil
}

func (n *NopPersistence) UpdateWorkflowStatus(ctx context.Context, workflowID string, status WorkflowStatus) error {
	return nil
}

func (n *NopPersistence) SaveExecution(ctx context.Context, execCtx *ExecutionContext) error {
	return nil
}

func (n *NopPersistence) GetExecution(ctx context.Context, executionID string) (*ExecutionContext, error) {
	return nil, ErrExecutionNotFound
}

func (n *NopPersistence) ListExecutions(ctx context.Context, filter *ExecutionFilter) ([]*ExecutionContext, error) {
	return []*ExecutionContext{}, nil
}

func (n *NopPersistence) DeleteExecution(ctx context.Context, executionID string) error {
	return nil
}

func (n *NopPersistence) UpdateExecutionStatus(ctx context.Context, executionID string, status ExecutionStatus) error {
	return nil
}

func (n *NopPersistence) SaveNodeState(ctx context.Context, executionID string, state *NodeState) error {
	return nil
}

func (n *NopPersistence) GetNodeStates(ctx context.Context, executionID string) (map[string]*NodeState, error) {
	return make(map[string]*NodeState), nil
}

func (n *NopPersistence) GetWorkflowStats(ctx context.Context, workflowID string, timeRange *TimeRange) (*WorkflowStats, error) {
	return &WorkflowStats{WorkflowID: workflowID}, nil
}

func (n *NopPersistence) GetExecutionHistory(ctx context.Context, workflowID string, limit int) ([]*ExecutionSummary, error) {
	return []*ExecutionSummary{}, nil
}

func (n *NopPersistence) CleanupOldExecutions(ctx context.Context, olderThan time.Time) (int, error) {
	return 0, nil
}

func (n *NopPersistence) Ping(ctx context.Context) error {
	return nil
}

func (n *NopPersistence) Close() error {
	return nil
}
