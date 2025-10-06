package workflow

import (
	"sync/atomic"
	"time"
)

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
	WorkflowStatusDraft    WorkflowStatus = "draft"    // 草稿
	WorkflowStatusActive   WorkflowStatus = "active"   // 启用
	WorkflowStatusInactive WorkflowStatus = "inactive" // 禁用
	WorkflowStatusArchived WorkflowStatus = "archived" // 归档
)

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"   // 待执行
	ExecutionStatusRunning   ExecutionStatus = "running"   // 执行中
	ExecutionStatusCompleted ExecutionStatus = "completed" // 已完成
	ExecutionStatusFailed    ExecutionStatus = "failed"    // 失败
	ExecutionStatusCancelled ExecutionStatus = "cancelled" // 已取消
	ExecutionStatusTimeout   ExecutionStatus = "timeout"   // 超时
)

// NodeStatus 节点执行状态
type NodeStatus string

const (
	NodeStatusPending   NodeStatus = "pending"   // 待执行
	NodeStatusRunning   NodeStatus = "running"   // 执行中
	NodeStatusCompleted NodeStatus = "completed" // 已完成
	NodeStatusFailed    NodeStatus = "failed"    // 失败
	NodeStatusSkipped   NodeStatus = "skipped"   // 跳过
)

// WorkflowDefinition 工作流定义
type WorkflowDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     int                    `json:"version"`
	Status      WorkflowStatus         `json:"status"`
	Nodes       []*NodeDefinition      `json:"nodes"`
	Edges       []*Edge                `json:"edges"`
	Variables   map[string]interface{} `json:"variables,omitempty"` // 全局变量
	Settings    *WorkflowSettings      `json:"settings"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
}

// WorkflowSettings 工作流配置
type WorkflowSettings struct {
	ExecutionTimeout time.Duration          `json:"execution_timeout"` // 执行超时时间
	MaxRetries       int                    `json:"max_retries"`       // 最大重试次数
	RetryDelay       time.Duration          `json:"retry_delay"`       // 重试延迟
	OnError          string                 `json:"on_error"`          // 错误处理策略: continue, stop, retry
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// NodeDefinition 节点定义
type NodeDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`        // 节点类型
	Config      map[string]interface{} `json:"config"`      // 节点配置
	Position    *Position              `json:"position"`    // UI 位置（可选）
	Disabled    bool                   `json:"disabled"`    // 是否禁用
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
}

// Edge 节点连接
type Edge struct {
	ID        string `json:"id"`
	Source    string `json:"source"`     // 源节点 ID
	Target    string `json:"target"`     // 目标节点 ID
	Condition string `json:"condition"`  // 条件表达式（可选）
	Label     string `json:"label"`      // 标签（如 "true"/"false" 分支）
}

// Position UI 位置
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxAttempts int           `json:"max_attempts"`
	Delay       time.Duration `json:"delay"`
	BackoffRate float64       `json:"backoff_rate"` // 退避倍率
}

// ExecutionContext 执行上下文
type ExecutionContext struct {
	ID            string                 `json:"id"`
	WorkflowID    string                 `json:"workflow_id"`
	Status        ExecutionStatus        `json:"status"`
	Input         map[string]interface{} `json:"input"`          // 输入数据
	Output        map[string]interface{} `json:"output"`         // 输出数据
	Variables     map[string]interface{} `json:"variables"`      // 上下文变量
	NodeStates    map[string]*NodeState  `json:"node_states"`    // 节点状态
	CurrentNodeID string                 `json:"current_node_id"`
	Error         string                 `json:"error,omitempty"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	TriggerBy     string                 `json:"trigger_by"`     // 触发者
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// NodeState 节点执行状态
type NodeState struct {
	NodeID      string                 `json:"node_id"`
	Status      NodeStatus             `json:"status"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output"`
	Error       string                 `json:"error,omitempty"`
	Attempts    int                    `json:"attempts"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// ExecutionMetrics 执行指标
type ExecutionMetrics struct {
	TotalExecutions   atomic.Int64
	SuccessExecutions atomic.Int64
	FailedExecutions  atomic.Int64
	AverageDuration   atomic.Int64 // 平均耗时（毫秒）
}

// WorkflowStats 工作流统计
type WorkflowStats struct {
	WorkflowID        string
	TotalExecutions   int64
	SuccessExecutions int64
	FailedExecutions  int64
	LastExecutedAt    *time.Time
	AverageDuration   time.Duration
}
