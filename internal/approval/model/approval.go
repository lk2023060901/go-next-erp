package model

import (
	"time"

	"github.com/google/uuid"
)

// ApprovalAction 审批操作
type ApprovalAction string

const (
	ApprovalActionApprove  ApprovalAction = "approve"  // 同意
	ApprovalActionReject   ApprovalAction = "reject"   // 拒绝
	ApprovalActionTransfer ApprovalAction = "transfer" // 转审
	ApprovalActionWithdraw ApprovalAction = "withdraw" // 撤回
)

// ProcessStatus 流程状态
type ProcessStatus string

const (
	ProcessStatusPending   ProcessStatus = "pending"   // 待审批
	ProcessStatusApproved  ProcessStatus = "approved"  // 已通过
	ProcessStatusRejected  ProcessStatus = "rejected"  // 已拒绝
	ProcessStatusWithdrawn ProcessStatus = "withdrawn" // 已撤回
	ProcessStatusCancelled ProcessStatus = "cancelled" // 已取消
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 待处理
	TaskStatusApproved  TaskStatus = "approved"  // 已同意
	TaskStatusRejected  TaskStatus = "rejected"  // 已拒绝
	TaskStatusTransferred TaskStatus = "transferred" // 已转审
	TaskStatusSkipped   TaskStatus = "skipped"   // 已跳过
)

// ProcessDefinition 流程定义
type ProcessDefinition struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Code        string     `json:"code"`         // 流程编码（如 LEAVE_REQUEST）
	Name        string     `json:"name"`         // 流程名称
	Category    string     `json:"category"`     // 流程分类
	FormID      uuid.UUID  `json:"form_id"`      // 关联表单ID
	WorkflowID  uuid.UUID  `json:"workflow_id"`  // 关联工作流ID
	Icon        *string    `json:"icon"`         // 图标
	Description *string    `json:"description"`  // 描述
	Enabled     bool       `json:"enabled"`      // 是否启用
	Sort        int        `json:"sort"`         // 排序
	CreatedBy   uuid.UUID  `json:"created_by"`
	UpdatedBy   *uuid.UUID `json:"updated_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

// ProcessInstance 流程实例
type ProcessInstance struct {
	ID                uuid.UUID              `json:"id"`
	TenantID          uuid.UUID              `json:"tenant_id"`
	ProcessDefID      uuid.UUID              `json:"process_def_id"`      // 流程定义ID
	ProcessDefCode    string                 `json:"process_def_code"`    // 流程定义编码
	ProcessDefName    string                 `json:"process_def_name"`    // 流程定义名称
	WorkflowInstanceID uuid.UUID             `json:"workflow_instance_id"` // 工作流实例ID
	FormDataID        uuid.UUID              `json:"form_data_id"`        // 表单数据ID
	ApplicantID       uuid.UUID              `json:"applicant_id"`        // 申请人ID
	ApplicantName     string                 `json:"applicant_name"`      // 申请人姓名
	Title             string                 `json:"title"`               // 流程标题
	Status            ProcessStatus          `json:"status"`
	CurrentNodeID     *string                `json:"current_node_id"`     // 当前节点ID
	CurrentNodeName   *string                `json:"current_node_name"`   // 当前节点名称
	Variables         map[string]interface{} `json:"variables"`           // 流程变量
	StartedAt         time.Time              `json:"started_at"`
	CompletedAt       *time.Time             `json:"completed_at"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// ApprovalTask 审批任务
type ApprovalTask struct {
	ID                 uuid.UUID      `json:"id"`
	TenantID           uuid.UUID      `json:"tenant_id"`
	ProcessInstanceID  uuid.UUID      `json:"process_instance_id"`  // 流程实例ID
	NodeID             string         `json:"node_id"`              // 工作流节点ID
	NodeName           string         `json:"node_name"`            // 节点名称
	AssigneeID         uuid.UUID      `json:"assignee_id"`          // 审批人ID
	AssigneeName       string         `json:"assignee_name"`        // 审批人姓名
	Status             TaskStatus     `json:"status"`
	Action             *ApprovalAction `json:"action"`              // 审批操作
	Comment            *string        `json:"comment"`              // 审批意见
	Attachments        []string       `json:"attachments"`          // 附件URL列表
	TransferToID       *uuid.UUID     `json:"transfer_to_id"`       // 转审目标人ID
	TransferToName     *string        `json:"transfer_to_name"`     // 转审目标人姓名
	ApprovedAt         *time.Time     `json:"approved_at"`          // 审批时间
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

// ProcessHistory 流程历史
type ProcessHistory struct {
	ID                uuid.UUID      `json:"id"`
	TenantID          uuid.UUID      `json:"tenant_id"`
	ProcessInstanceID uuid.UUID      `json:"process_instance_id"`
	TaskID            *uuid.UUID     `json:"task_id"`              // 关联任务ID
	NodeID            string         `json:"node_id"`
	NodeName          string         `json:"node_name"`
	OperatorID        uuid.UUID      `json:"operator_id"`          // 操作人ID
	OperatorName      string         `json:"operator_name"`        // 操作人姓名
	Action            ApprovalAction `json:"action"`
	Comment           *string        `json:"comment"`
	FromStatus        *ProcessStatus `json:"from_status"`
	ToStatus          ProcessStatus  `json:"to_status"`
	CreatedAt         time.Time      `json:"created_at"`
}
