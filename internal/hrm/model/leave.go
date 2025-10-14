package model

import (
	"time"

	"github.com/google/uuid"
)

// Leave 请假记录
type Leave struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 申请人信息（关联 organization.employees）
	EmployeeID   uuid.UUID `json:"employee_id"`   // 对应 organization.employees.id
	EmployeeName string    `json:"employee_name"` // 冗余
	DepartmentID uuid.UUID `json:"department_id"` // 冗余

	// 请假类型
	LeaveTypeID   uuid.UUID `json:"leave_type_id"`
	LeaveTypeName string    `json:"leave_type_name"` // 冗余

	// 请假时间
	StartTime time.Time `json:"start_time"` // 开始时间
	EndTime   time.Time `json:"end_time"`   // 结束时间
	Duration  float64   `json:"duration"`   // 请假天数（支持小数）
	Unit      string    `json:"unit"`       // 单位：day, hour

	// 请假理由
	Reason     string   `json:"reason"`
	Attachment []string `json:"attachment,omitempty"` // 附件（证明材料）

	// 审批信息
	ApprovalID     *uuid.UUID `json:"approval_id,omitempty"` // 关联审批流程
	ApprovalStatus string     `json:"approval_status"`       // pending, approved, rejected
	ApprovedBy     *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	RejectReason   string     `json:"reject_reason,omitempty"`

	// 销假信息
	ActualEndTime *time.Time `json:"actual_end_time,omitempty"` // 实际结束时间
	IsCanceled    bool       `json:"is_canceled"`               // 是否撤销

	// 备注
	Remark string `json:"remark,omitempty"`

	// 审计字段
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// LeaveType 请假类型
type LeaveType struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 类型信息
	Code        string `json:"code"`        // 编码（唯一）
	Name        string `json:"name"`        // 名称
	Description string `json:"description"` // 描述

	// 假期属性
	IsPaid         bool    `json:"is_paid"`         // 是否带薪
	NeedApproval   bool    `json:"need_approval"`   // 是否需要审批
	NeedAttachment bool    `json:"need_attachment"` // 是否需要附件
	DeductSalary   bool    `json:"deduct_salary"`   // 是否扣薪
	PayRate        float64 `json:"pay_rate"`        // 薪资比例（0-1）

	// 额度设置
	HasQuota     bool    `json:"has_quota"`      // 是否有额度限制
	QuotaType    string  `json:"quota_type"`     // annual, monthly, total
	QuotaDays    float64 `json:"quota_days"`     // 额度天数
	CarryForward bool    `json:"carry_forward"`  // 是否可结转
	MaxCarryDays float64 `json:"max_carry_days"` // 最大结转天数

	// 申请限制
	MinUnit        string  `json:"min_unit"`         // 最小单位：day, half_day, hour
	MinDuration    float64 `json:"min_duration"`     // 最小时长
	MaxDuration    float64 `json:"max_duration"`     // 最大时长
	MinAdvanceDays int     `json:"min_advance_days"` // 最少提前天数
	MaxAdvanceDays int     `json:"max_advance_days"` // 最多提前天数

	// 适用范围
	ApplyType     ApplyType   `json:"apply_type"` // all, department, employee
	DepartmentIDs []uuid.UUID `json:"department_ids,omitempty"`
	EmployeeIDs   []uuid.UUID `json:"employee_ids,omitempty"`

	// 颜色标识
	Color string `json:"color,omitempty"`

	// 状态
	IsActive bool `json:"is_active"`
	Sort     int  `json:"sort"`

	// 审计字段
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy uuid.UUID  `json:"updated_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// LeaveQuota 请假额度
type LeaveQuota struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	EmployeeID  uuid.UUID `json:"employee_id"`
	LeaveTypeID uuid.UUID `json:"leave_type_id"`

	// 额度信息
	Year          int     `json:"year"`           // 年份
	TotalDays     float64 `json:"total_days"`     // 总额度
	UsedDays      float64 `json:"used_days"`      // 已使用
	RemainingDays float64 `json:"remaining_days"` // 剩余
	CarriedDays   float64 `json:"carried_days"`   // 结转

	// 审计字段
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
