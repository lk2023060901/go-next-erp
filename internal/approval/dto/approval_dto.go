package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/approval/model"
)

// CreateProcessDefRequest 创建流程定义请求
type CreateProcessDefRequest struct {
	TenantID   uuid.UUID `json:"-"`
	Code       string    `json:"code" binding:"required,max=50"`
	Name       string    `json:"name" binding:"required,max=100"`
	Category   string    `json:"category" binding:"required,max=50"`
	FormID     uuid.UUID `json:"form_id" binding:"required"`
	WorkflowID uuid.UUID `json:"workflow_id" binding:"required"`
	CreatedBy  uuid.UUID `json:"-"`
}

// CreateProcessDefinitionRequest 创建流程定义请求（向后兼容）
type CreateProcessDefinitionRequest struct {
	Code        string  `json:"code" binding:"required,max=50"`
	Name        string  `json:"name" binding:"required,max=100"`
	Category    string  `json:"category" binding:"required,max=50"`
	FormID      string  `json:"form_id" binding:"required"`
	WorkflowID  string  `json:"workflow_id" binding:"required"`
	Icon        *string `json:"icon"`
	Description *string `json:"description"`
	Sort        *int    `json:"sort"`
}

// UpdateProcessDefRequest 更新流程定义请求
type UpdateProcessDefRequest struct {
	Name       string    `json:"name" binding:"required,max=100"`
	FormID     uuid.UUID `json:"form_id" binding:"required"`
	WorkflowID uuid.UUID `json:"workflow_id" binding:"required"`
	Enabled    bool      `json:"enabled"`
	UpdatedBy  uuid.UUID `json:"-"`
}

// UpdateProcessDefinitionRequest 更新流程定义请求（向后兼容）
type UpdateProcessDefinitionRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Icon        *string `json:"icon"`
	Description *string `json:"description"`
	Enabled     *bool   `json:"enabled"`
	Sort        *int    `json:"sort"`
}

// StartProcessRequest 发起流程请求
type StartProcessRequest struct {
	TenantID     uuid.UUID              `json:"-"`
	ProcessDefID uuid.UUID              `json:"process_def_id" binding:"required"`
	ApplicantID  uuid.UUID              `json:"-"`
	FormData     map[string]interface{} `json:"form_data" binding:"required"`
}

// StartProcessRequestOld 发起流程请求（向后兼容）
type StartProcessRequestOld struct {
	ProcessDefCode string                 `json:"process_def_code" binding:"required"`
	Title          string                 `json:"title" binding:"required,max=200"`
	FormData       map[string]interface{} `json:"form_data" binding:"required"`
	Variables      map[string]interface{} `json:"variables"`
}

// ProcessTaskRequest 处理任务请求
type ProcessTaskRequest struct {
	TaskID     uuid.UUID            `json:"-"`
	OperatorID uuid.UUID            `json:"-"`
	Action     model.ApprovalAction `json:"action" binding:"required"`
	Comment    *string              `json:"comment"`
}

// ApproveTaskRequest 审批任务请求（向后兼容）
type ApproveTaskRequest struct {
	Action       string   `json:"action" binding:"required,oneof=approve reject transfer"`
	Comment      *string  `json:"comment"`
	Attachments  []string `json:"attachments"`
	TransferToID *string  `json:"transfer_to_id"`
}

// WithdrawProcessRequest 撤回流程请求
type WithdrawProcessRequest struct {
	Reason *string `json:"reason"`
}

// ProcessDefResponse 流程定义响应
type ProcessDefResponse struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	FormID       uuid.UUID `json:"form_id"`
	FormName     string    `json:"form_name"`
	WorkflowID   uuid.UUID `json:"workflow_id"`
	WorkflowName string    `json:"workflow_name"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ProcessDefinitionResponse 流程定义响应（向后兼容）
type ProcessDefinitionResponse struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Category    string     `json:"category"`
	FormID      uuid.UUID  `json:"form_id"`
	WorkflowID  uuid.UUID  `json:"workflow_id"`
	Icon        *string    `json:"icon,omitempty"`
	Description *string    `json:"description,omitempty"`
	Enabled     bool       `json:"enabled"`
	Sort        int        `json:"sort"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ProcessInstanceResponse 流程实例响应
type ProcessInstanceResponse struct {
	ID                 uuid.UUID              `json:"id"`
	TenantID           uuid.UUID              `json:"tenant_id"`
	ProcessDefID       uuid.UUID              `json:"process_def_id"`
	ProcessDefCode     string                 `json:"process_def_code"`
	ProcessDefName     string                 `json:"process_def_name"`
	WorkflowInstanceID uuid.UUID              `json:"workflow_instance_id"`
	FormDataID         uuid.UUID              `json:"form_data_id"`
	ApplicantID        uuid.UUID              `json:"applicant_id"`
	ApplicantName      string                 `json:"applicant_name"`
	Title              string                 `json:"title"`
	Status             model.ProcessStatus    `json:"status"`
	CurrentNodeID      *string                `json:"current_node_id,omitempty"`
	CurrentNodeName    *string                `json:"current_node_name,omitempty"`
	Variables          map[string]interface{} `json:"variables,omitempty"`
	StartedAt          time.Time              `json:"started_at"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// ApprovalTaskResponse 审批任务响应
type ApprovalTaskResponse struct {
	ID                uuid.UUID             `json:"id"`
	TenantID          uuid.UUID             `json:"tenant_id"`
	ProcessInstanceID uuid.UUID             `json:"process_instance_id"`
	NodeID            string                `json:"node_id"`
	NodeName          string                `json:"node_name"`
	AssigneeID        uuid.UUID             `json:"assignee_id"`
	AssigneeName      string                `json:"assignee_name"`
	Status            model.TaskStatus      `json:"status"`
	Action            *model.ApprovalAction `json:"action,omitempty"`
	Comment           *string               `json:"comment,omitempty"`
	Attachments       []string              `json:"attachments,omitempty"`
	TransferToID      *uuid.UUID            `json:"transfer_to_id,omitempty"`
	TransferToName    *string               `json:"transfer_to_name,omitempty"`
	ApprovedAt        *time.Time            `json:"approved_at,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
}

// ProcessHistoryResponse 流程历史响应
type ProcessHistoryResponse struct {
	ID                uuid.UUID            `json:"id"`
	ProcessInstanceID uuid.UUID            `json:"process_instance_id"`
	TaskID            *uuid.UUID           `json:"task_id,omitempty"`
	NodeID            string               `json:"node_id"`
	NodeName          string               `json:"node_name"`
	OperatorID        uuid.UUID            `json:"operator_id"`
	OperatorName      string               `json:"operator_name"`
	Action            model.ApprovalAction `json:"action"`
	Comment           *string              `json:"comment,omitempty"`
	FromStatus        *model.ProcessStatus `json:"from_status,omitempty"`
	ToStatus          model.ProcessStatus  `json:"to_status"`
	CreatedAt         time.Time            `json:"created_at"`
}

// Response 通用响应
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Success 成功响应
func Success(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

// Error 错误响应
func Error(code int, message string) Response {
	return Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
}

// ToProcessDefinitionResponse 转换为流程定义响应
func ToProcessDefinitionResponse(pd *model.ProcessDefinition) *ProcessDefinitionResponse {
	return &ProcessDefinitionResponse{
		ID:          pd.ID,
		TenantID:    pd.TenantID,
		Code:        pd.Code,
		Name:        pd.Name,
		Category:    pd.Category,
		FormID:      pd.FormID,
		WorkflowID:  pd.WorkflowID,
		Icon:        pd.Icon,
		Description: pd.Description,
		Enabled:     pd.Enabled,
		Sort:        pd.Sort,
		CreatedBy:   pd.CreatedBy,
		UpdatedBy:   pd.UpdatedBy,
		CreatedAt:   pd.CreatedAt,
		UpdatedAt:   pd.UpdatedAt,
	}
}

// ToProcessDefinitionResponseList 转换为流程定义响应列表
func ToProcessDefinitionResponseList(defs []*model.ProcessDefinition) []*ProcessDefinitionResponse {
	result := make([]*ProcessDefinitionResponse, len(defs))
	for i, pd := range defs {
		result[i] = ToProcessDefinitionResponse(pd)
	}
	return result
}

// ToProcessInstanceResponse 转换为流程实例响应
func ToProcessInstanceResponse(pi *model.ProcessInstance) *ProcessInstanceResponse {
	return &ProcessInstanceResponse{
		ID:                 pi.ID,
		TenantID:           pi.TenantID,
		ProcessDefID:       pi.ProcessDefID,
		ProcessDefCode:     pi.ProcessDefCode,
		ProcessDefName:     pi.ProcessDefName,
		WorkflowInstanceID: pi.WorkflowInstanceID,
		FormDataID:         pi.FormDataID,
		ApplicantID:        pi.ApplicantID,
		ApplicantName:      pi.ApplicantName,
		Title:              pi.Title,
		Status:             pi.Status,
		CurrentNodeID:      pi.CurrentNodeID,
		CurrentNodeName:    pi.CurrentNodeName,
		Variables:          pi.Variables,
		StartedAt:          pi.StartedAt,
		CompletedAt:        pi.CompletedAt,
		CreatedAt:          pi.CreatedAt,
		UpdatedAt:          pi.UpdatedAt,
	}
}

// ToProcessInstanceResponseList 转换为流程实例响应列表
func ToProcessInstanceResponseList(instances []*model.ProcessInstance) []*ProcessInstanceResponse {
	result := make([]*ProcessInstanceResponse, len(instances))
	for i, pi := range instances {
		result[i] = ToProcessInstanceResponse(pi)
	}
	return result
}

// ToApprovalTaskResponse 转换为审批任务响应
func ToApprovalTaskResponse(task *model.ApprovalTask) *ApprovalTaskResponse {
	return &ApprovalTaskResponse{
		ID:                task.ID,
		TenantID:          task.TenantID,
		ProcessInstanceID: task.ProcessInstanceID,
		NodeID:            task.NodeID,
		NodeName:          task.NodeName,
		AssigneeID:        task.AssigneeID,
		AssigneeName:      task.AssigneeName,
		Status:            task.Status,
		Action:            task.Action,
		Comment:           task.Comment,
		Attachments:       task.Attachments,
		TransferToID:      task.TransferToID,
		TransferToName:    task.TransferToName,
		ApprovedAt:        task.ApprovedAt,
		CreatedAt:         task.CreatedAt,
		UpdatedAt:         task.UpdatedAt,
	}
}

// ToApprovalTaskResponseList 转换为审批任务响应列表
func ToApprovalTaskResponseList(tasks []*model.ApprovalTask) []*ApprovalTaskResponse {
	result := make([]*ApprovalTaskResponse, len(tasks))
	for i, task := range tasks {
		result[i] = ToApprovalTaskResponse(task)
	}
	return result
}

// ToProcessHistoryResponse 转换为流程历史响应
func ToProcessHistoryResponse(ph *model.ProcessHistory) *ProcessHistoryResponse {
	return &ProcessHistoryResponse{
		ID:                ph.ID,
		ProcessInstanceID: ph.ProcessInstanceID,
		TaskID:            ph.TaskID,
		NodeID:            ph.NodeID,
		NodeName:          ph.NodeName,
		OperatorID:        ph.OperatorID,
		OperatorName:      ph.OperatorName,
		Action:            ph.Action,
		Comment:           ph.Comment,
		FromStatus:        ph.FromStatus,
		ToStatus:          ph.ToStatus,
		CreatedAt:         ph.CreatedAt,
	}
}

// ToProcessHistoryResponseList 转换为流程历史响应列表
func ToProcessHistoryResponseList(histories []*model.ProcessHistory) []*ProcessHistoryResponse {
	result := make([]*ProcessHistoryResponse, len(histories))
	for i, ph := range histories {
		result[i] = ToProcessHistoryResponse(ph)
	}
	return result
}

// BatchProcessResult 批量处理结果
type BatchProcessResult struct {
	TaskID  uuid.UUID `json:"task_id"`
	Success bool      `json:"success"`
	Error   *string   `json:"error,omitempty"`
}

// InstanceStatsSummary 实例统计汇总
type InstanceStatsSummary struct {
	Total        int            `json:"total"`
	Pending      int            `json:"pending"`
	Approved     int            `json:"approved"`
	Rejected     int            `json:"rejected"`
	Withdrawn    int            `json:"withdrawn"`
	Cancelled    int            `json:"cancelled"`
	ByStatus     map[string]int `json:"by_status"`
	ByProcessDef map[string]int `json:"by_process_def,omitempty"`
}

// DashboardResponse 工作台响应
type DashboardResponse struct {
	MyPendingTasks      int                        `json:"my_pending_tasks"`
	MyCompletedTasks    int                        `json:"my_completed_tasks"`
	MyApplications      int                        `json:"my_applications"`
	PendingApplications int                        `json:"pending_applications"`
	RecentTasks         []*ApprovalTaskResponse    `json:"recent_tasks"`
	RecentApplications  []*ProcessInstanceResponse `json:"recent_applications"`
}

// ProcessMetrics 流程指标
type ProcessMetrics struct {
	ProcessDefID       uuid.UUID `json:"process_def_id"`
	ProcessCode        string    `json:"process_code"`
	ProcessName        string    `json:"process_name"`
	TotalInstances     int       `json:"total_instances"`
	CompletedInstances int       `json:"completed_instances"`
	PendingInstances   int       `json:"pending_instances"`
	ApprovedInstances  int       `json:"approved_instances"`
	RejectedInstances  int       `json:"rejected_instances"`
	AvgDurationSeconds int64     `json:"avg_duration_seconds"`
	AvgApprovalTime    int64     `json:"avg_approval_time"` // 平均审批耗时（秒）
	MaxDuration        int64     `json:"max_duration"`
	MinDuration        int64     `json:"min_duration"`
	ApprovalRate       float64   `json:"approval_rate"`  // 通过率
	RejectionRate      float64   `json:"rejection_rate"` // 拒绝率
}

// UserWorkload 用户工作负载
type UserWorkload struct {
	UserID           uuid.UUID `json:"user_id"`
	UserName         string    `json:"user_name"`
	PendingTasks     int       `json:"pending_tasks"`
	CompletedTasks   int       `json:"completed_tasks"`
	ApprovedTasks    int       `json:"approved_tasks"`
	RejectedTasks    int       `json:"rejected_tasks"`
	TransferredTasks int       `json:"transferred_tasks"`
	AvgProcessTime   int64     `json:"avg_process_time"` // 平均处理时间（秒）
	TodayTasks       int       `json:"today_tasks"`
	ThisWeekTasks    int       `json:"this_week_tasks"`
	ThisMonthTasks   int       `json:"this_month_tasks"`
}
