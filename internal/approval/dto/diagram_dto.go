package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/approval/model"
)

// Position 节点位置
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// DiagramNode 流程图节点
type DiagramNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`     // start, approval, condition, end
	Name     string                 `json:"name"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Position *Position              `json:"position,omitempty"`
	Disabled bool                   `json:"disabled,omitempty"`
}

// DiagramEdge 流程图连接线
type DiagramEdge struct {
	ID        string  `json:"id"`
	Source    string  `json:"source"`
	Target    string  `json:"target"`
	Condition string  `json:"condition,omitempty"` // 条件表达式
	Label     string  `json:"label,omitempty"`     // 边标签，如 "同意"/"拒绝"
}

// ProcessDiagramResponse 流程定义图响应
type ProcessDiagramResponse struct {
	ProcessDefID uuid.UUID      `json:"process_def_id"`
	ProcessCode  string         `json:"process_code"`
	ProcessName  string         `json:"process_name"`
	Version      int            `json:"version"`
	Nodes        []*DiagramNode `json:"nodes"`
	Edges        []*DiagramEdge `json:"edges"`
}

// NodeExecutionState 节点执行状态
type NodeExecutionState struct {
	NodeID      string                 `json:"node_id"`
	NodeName    string                 `json:"node_name"`
	Status      string                 `json:"status"` // pending, running, completed, failed, skipped
	Operator    *string                `json:"operator,omitempty"`
	OperatorID  *uuid.UUID             `json:"operator_id,omitempty"`
	Action      *model.ApprovalAction  `json:"action,omitempty"`
	Comment     *string                `json:"comment,omitempty"`
	EnteredAt   *time.Time             `json:"entered_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    *int64                 `json:"duration,omitempty"` // 耗时（秒）
	Data        map[string]interface{} `json:"data,omitempty"`
}

// ProcessInstanceDiagramResponse 流程实例图响应（带执行状态）
type ProcessInstanceDiagramResponse struct {
	InstanceID       uuid.UUID             `json:"instance_id"`
	ProcessDefID     uuid.UUID             `json:"process_def_id"`
	ProcessCode      string                `json:"process_code"`
	ProcessName      string                `json:"process_name"`
	Status           model.ProcessStatus   `json:"status"`
	CurrentNodeID    *string               `json:"current_node_id,omitempty"`
	StartedAt        time.Time             `json:"started_at"`
	CompletedAt      *time.Time            `json:"completed_at,omitempty"`
	Nodes            []*DiagramNode        `json:"nodes"`
	Edges            []*DiagramEdge        `json:"edges"`
	NodeStates       []*NodeExecutionState `json:"node_states"`        // 节点执行状态
	CompletedNodeIDs []string              `json:"completed_node_ids"` // 已完成节点ID列表
	ActiveNodeIDs    []string              `json:"active_node_ids"`    // 当前活动节点ID列表
}

// ProcessTraceNode 流程轨迹节点
type ProcessTraceNode struct {
	NodeID      string                `json:"node_id"`
	NodeName    string                `json:"node_name"`
	NodeType    string                `json:"node_type"`
	Operator    *string               `json:"operator,omitempty"`
	OperatorID  *uuid.UUID            `json:"operator_id,omitempty"`
	Action      *model.ApprovalAction `json:"action,omitempty"`
	Comment     *string               `json:"comment,omitempty"`
	EnteredAt   time.Time             `json:"entered_at"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`
	Duration    *int64                `json:"duration,omitempty"` // 耗时（秒）
}

// ProcessTraceResponse 流程轨迹响应（历史路径）
type ProcessTraceResponse struct {
	InstanceID  uuid.UUID           `json:"instance_id"`
	ProcessCode string              `json:"process_code"`
	ProcessName string              `json:"process_name"`
	Status      model.ProcessStatus `json:"status"`
	StartedAt   time.Time           `json:"started_at"`
	CompletedAt *time.Time          `json:"completed_at,omitempty"`
	Path        []*ProcessTraceNode `json:"path"`          // 执行路径
	TotalNodes  int                 `json:"total_nodes"`   // 总节点数
	Duration    int64               `json:"duration"`      // 总耗时（秒）
}

// ProcessStatsResponse 流程统计响应
type ProcessStatsResponse struct {
	ProcessDefID      uuid.UUID `json:"process_def_id"`
	ProcessCode       string    `json:"process_code"`
	ProcessName       string    `json:"process_name"`
	TotalInstances    int       `json:"total_instances"`
	PendingInstances  int       `json:"pending_instances"`
	ApprovedInstances int       `json:"approved_instances"`
	RejectedInstances int       `json:"rejected_instances"`
	AvgDuration       int64     `json:"avg_duration"` // 平均耗时（秒）
}
