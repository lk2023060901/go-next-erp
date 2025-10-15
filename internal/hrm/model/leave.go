package model

import (
	"time"

	"github.com/google/uuid"
)

// LeaveUnit 请假单位
type LeaveUnit string

const (
	LeaveUnitDay     LeaveUnit = "day"      // 天
	LeaveUnitHalfDay LeaveUnit = "half_day" // 半天
	LeaveUnitHour    LeaveUnit = "hour"     // 小时
)

// LeaveRequestStatus 请假申请状态
type LeaveRequestStatus string

const (
	LeaveRequestStatusDraft     LeaveRequestStatus = "draft"     // 草稿
	LeaveRequestStatusPending   LeaveRequestStatus = "pending"   // 待审批
	LeaveRequestStatusApproved  LeaveRequestStatus = "approved"  // 已批准
	LeaveRequestStatusRejected  LeaveRequestStatus = "rejected"  // 已拒绝
	LeaveRequestStatusWithdrawn LeaveRequestStatus = "withdrawn" // 已撤回
	LeaveRequestStatusCancelled LeaveRequestStatus = "cancelled" // 已取消
)

// LeaveApprovalStatus 审批状态
type LeaveApprovalStatus string

const (
	LeaveApprovalStatusPending  LeaveApprovalStatus = "pending"  // 待审批
	LeaveApprovalStatusApproved LeaveApprovalStatus = "approved" // 已批准
	LeaveApprovalStatusRejected LeaveApprovalStatus = "rejected" // 已拒绝
	LeaveApprovalStatusSkipped  LeaveApprovalStatus = "skipped"  // 已跳过
)

// LeaveApprovalAction 审批动作
type LeaveApprovalAction string

const (
	LeaveApprovalActionApprove LeaveApprovalAction = "approve" // 批准
	LeaveApprovalActionReject  LeaveApprovalAction = "reject"  // 拒绝
)

// ApproverType 审批人类型
type ApproverType string

const (
	ApproverTypeDirectManager  ApproverType = "direct_manager"  // 直属上级
	ApproverTypeDeptManager    ApproverType = "dept_manager"    // 部门负责人
	ApproverTypeHR             ApproverType = "hr"              // HR
	ApproverTypeGeneralManager ApproverType = "general_manager" // 总经理
	ApproverTypeCustom         ApproverType = "custom"          // 自定义审批人
)

// ApprovalNode 审批节点
type ApprovalNode struct {
	Level        int          `json:"level"`         // 审批层级
	ApproverType ApproverType `json:"approver_type"` // 审批人类型
	ApproverID   *string      `json:"approver_id"`   // 自定义审批人ID（当type为custom时使用）
	Required     bool         `json:"required"`      // 是否必须审批
}

// DurationRule 基于请假天数的审批规则
type DurationRule struct {
	MinDuration   float64         `json:"min_duration"`   // 最小天数（包含）
	MaxDuration   *float64        `json:"max_duration"`   // 最大天数（不包含），null表示无上限
	ApprovalChain []*ApprovalNode `json:"approval_chain"` // 该天数范围的审批链
}

// ApprovalRules 审批规则配置
type ApprovalRules struct {
	DefaultChain  []*ApprovalNode `json:"default_chain"`  // 默认审批链（天数未匹配时使用）
	DurationRules []*DurationRule `json:"duration_rules"` // 基于天数的规则列表
}

// GetApprovalChain 根据请假天数获取适用的审批链
func (r *ApprovalRules) GetApprovalChain(duration float64) []*ApprovalNode {
	if r == nil {
		return nil
	}

	// 遍历天数规则，找到匹配的规则
	for _, rule := range r.DurationRules {
		if duration >= rule.MinDuration {
			// 如果没有上限，或者在上限范围内
			if rule.MaxDuration == nil || duration < *rule.MaxDuration {
				return rule.ApprovalChain
			}
		}
	}

	// 没有匹配的规则，使用默认审批链
	return r.DefaultChain
}

// LeaveType 请假类型
type LeaveType struct {
	ID               uuid.UUID      `json:"id"`
	TenantID         uuid.UUID      `json:"tenant_id"`
	Code             string         `json:"code"`              // 类型编码
	Name             string         `json:"name"`              // 类型名称
	Description      string         `json:"description"`       // 描述
	IsPaid           bool           `json:"is_paid"`           // 是否带薪
	RequiresApproval bool           `json:"requires_approval"` // 是否需要审批
	RequiresProof    bool           `json:"requires_proof"`    // 是否需要证明材料
	DeductQuota      bool           `json:"deduct_quota"`      // 是否扣除额度
	Unit             LeaveUnit      `json:"unit"`              // 最小单位
	MinDuration      float64        `json:"min_duration"`      // 最小请假时长
	MaxDuration      *float64       `json:"max_duration"`      // 最大请假时长
	AdvanceDays      int            `json:"advance_days"`      // 需要提前申请的天数
	ApprovalRules    *ApprovalRules `json:"approval_rules"`    // 审批规则配置
	Color            string         `json:"color"`             // UI显示颜色
	IsActive         bool           `json:"is_active"`         // 是否启用
	Sort             int            `json:"sort"`              // 排序
	CreatedBy        *uuid.UUID     `json:"created_by"`
	UpdatedBy        *uuid.UUID     `json:"updated_by"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        *time.Time     `json:"deleted_at"`
}

// LeaveQuota 请假额度
type LeaveQuota struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	EmployeeID   uuid.UUID  `json:"employee_id"`   // 员工ID
	LeaveTypeID  uuid.UUID  `json:"leave_type_id"` // 请假类型ID
	Year         int        `json:"year"`          // 年份
	TotalQuota   float64    `json:"total_quota"`   // 总额度
	UsedQuota    float64    `json:"used_quota"`    // 已使用额度
	PendingQuota float64    `json:"pending_quota"` // 待审批额度
	ExpiredAt    *time.Time `json:"expired_at"`    // 过期时间
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// RemainingQuota 剩余额度（计算属性）
func (q *LeaveQuota) RemainingQuota() float64 {
	return q.TotalQuota - q.UsedQuota - q.PendingQuota
}

// LeaveRequest 请假申请
type LeaveRequest struct {
	ID                uuid.UUID          `json:"id"`
	TenantID          uuid.UUID          `json:"tenant_id"`
	EmployeeID        uuid.UUID          `json:"employee_id"`         // 申请人ID
	EmployeeName      string             `json:"employee_name"`       // 申请人姓名
	DepartmentID      *uuid.UUID         `json:"department_id"`       // 部门ID
	LeaveTypeID       uuid.UUID          `json:"leave_type_id"`       // 请假类型ID
	LeaveTypeName     string             `json:"leave_type_name"`     // 请假类型名称
	StartTime         time.Time          `json:"start_time"`          // 开始时间
	EndTime           time.Time          `json:"end_time"`            // 结束时间
	Duration          float64            `json:"duration"`            // 请假时长
	Unit              LeaveUnit          `json:"unit"`                // 单位
	Reason            string             `json:"reason"`              // 请假原因
	ProofURLs         []string           `json:"proof_urls"`          // 证明材料附件URL数组
	Status            LeaveRequestStatus `json:"status"`              // 状态
	CurrentApproverID *uuid.UUID         `json:"current_approver_id"` // 当前审批人ID
	SubmittedAt       *time.Time         `json:"submitted_at"`        // 提交时间
	ApprovedAt        *time.Time         `json:"approved_at"`         // 审批通过时间
	RejectedAt        *time.Time         `json:"rejected_at"`         // 拒绝时间
	CancelledAt       *time.Time         `json:"cancelled_at"`        // 取消时间
	Remark            string             `json:"remark"`              // 备注
	CreatedBy         *uuid.UUID         `json:"created_by"`
	UpdatedBy         *uuid.UUID         `json:"updated_by"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
	DeletedAt         *time.Time         `json:"deleted_at"`
}

// LeaveApproval 请假审批记录
type LeaveApproval struct {
	ID             uuid.UUID            `json:"id"`
	TenantID       uuid.UUID            `json:"tenant_id"`
	LeaveRequestID uuid.UUID            `json:"leave_request_id"` // 请假申请ID
	ApproverID     uuid.UUID            `json:"approver_id"`      // 审批人ID
	ApproverName   string               `json:"approver_name"`    // 审批人姓名
	Level          int                  `json:"level"`            // 审批层级
	Status         LeaveApprovalStatus  `json:"status"`           // 审批状态
	Action         *LeaveApprovalAction `json:"action"`           // 审批动作
	Comment        string               `json:"comment"`          // 审批意见
	ApprovedAt     *time.Time           `json:"approved_at"`      // 审批时间
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

// LeaveRequestWithApprovals 请假申请（含审批记录）
type LeaveRequestWithApprovals struct {
	LeaveRequest
	Approvals []*LeaveApproval `json:"approvals"` // 审批记录列表
}

// LeaveQuotaWithType 请假额度（含类型信息）
type LeaveQuotaWithType struct {
	LeaveQuota
	LeaveType *LeaveType `json:"leave_type"` // 请假类型
}
