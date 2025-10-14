package model

import (
	"time"

	"github.com/google/uuid"
)

// Overtime 加班记录
type Overtime struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 申请人信息（关联 organization.employees）
	EmployeeID   uuid.UUID `json:"employee_id"`   // 对应 organization.employees.id
	EmployeeName string    `json:"employee_name"` // 冗余
	DepartmentID uuid.UUID `json:"department_id"` // 冗余

	// 加班时间
	StartTime time.Time `json:"start_time"` // 开始时间
	EndTime   time.Time `json:"end_time"`   // 结束时间
	Duration  float64   `json:"duration"`   // 加班时长（小时）

	// 加班类型
	OvertimeType OvertimeType `json:"overtime_type"` // workday, weekend, holiday
	PayType      string       `json:"pay_type"`      // money, leave（调休）

	// 加班倍率（根据劳动法）
	PayRate float64 `json:"pay_rate"` // 工作日1.5倍，周末2倍，节假日3倍

	// 加班原因
	Reason string   `json:"reason"`
	Tasks  []string `json:"tasks,omitempty"` // 加班任务

	// 审批信息
	ApprovalID     *uuid.UUID `json:"approval_id,omitempty"`
	ApprovalStatus string     `json:"approval_status"` // pending, approved, rejected
	ApprovedBy     *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	RejectReason   string     `json:"reject_reason,omitempty"`

	// 调休信息
	CompOffDays     float64    `json:"comp_off_days"`                // 可调休天数
	CompOffUsed     float64    `json:"comp_off_used"`                // 已调休天数
	CompOffExpireAt *time.Time `json:"comp_off_expire_at,omitempty"` // 调休过期时间

	// 备注
	Remark string `json:"remark,omitempty"`

	// 审计字段
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// OvertimeType 加班类型
type OvertimeType string

const (
	OvertimeTypeWorkday OvertimeType = "workday" // 工作日加班
	OvertimeTypeWeekend OvertimeType = "weekend" // 周末加班
	OvertimeTypeHoliday OvertimeType = "holiday" // 节假日加班
)

// BusinessTrip 出差记录
type BusinessTrip struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 申请人信息（关联 organization.employees）
	EmployeeID   uuid.UUID `json:"employee_id"`   // 对应 organization.employees.id
	EmployeeName string    `json:"employee_name"` // 冗余
	DepartmentID uuid.UUID `json:"department_id"` // 冗余

	// 出差时间
	StartTime time.Time `json:"start_time"` // 开始时间
	EndTime   time.Time `json:"end_time"`   // 结束时间
	Duration  float64   `json:"duration"`   // 出差天数

	// 出差地点
	Destination    string      `json:"destination"`          // 目的地
	Transportation string      `json:"transportation"`       // 交通方式
	Accommodation  string      `json:"accommodation"`        // 住宿安排
	Companions     []uuid.UUID `json:"companions,omitempty"` // 同行人员

	// 出差原因
	Purpose string `json:"purpose"`
	Tasks   string `json:"tasks,omitempty"`

	// 预算信息
	EstimatedCost float64 `json:"estimated_cost"` // 预计费用
	ActualCost    float64 `json:"actual_cost"`    // 实际费用

	// 审批信息
	ApprovalID     *uuid.UUID `json:"approval_id,omitempty"`
	ApprovalStatus string     `json:"approval_status"` // pending, approved, rejected
	ApprovedBy     *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	RejectReason   string     `json:"reject_reason,omitempty"`

	// 出差报告
	Report   string     `json:"report,omitempty"` // 出差报告
	ReportAt *time.Time `json:"report_at,omitempty"`

	// 备注
	Remark string `json:"remark,omitempty"`

	// 审计字段
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// AttendanceSummary 考勤汇总（按员工按月统计）
type AttendanceSummary struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 员工信息（关联 organization.employees）
	EmployeeID   uuid.UUID `json:"employee_id"`   // 对应 organization.employees.id
	EmployeeName string    `json:"employee_name"` // 冗余
	DepartmentID uuid.UUID `json:"department_id"` // 冗余

	// 统计周期
	Year  int `json:"year"`
	Month int `json:"month"`

	// 出勤统计
	WorkDays      int     `json:"work_days"`      // 应出勤天数
	ActualDays    int     `json:"actual_days"`    // 实际出勤天数
	LateCount     int     `json:"late_count"`     // 迟到次数
	LateDuration  int     `json:"late_duration"`  // 迟到总时长（分钟）
	EarlyCount    int     `json:"early_count"`    // 早退次数
	EarlyDuration int     `json:"early_duration"` // 早退总时长（分钟）
	AbsentCount   int     `json:"absent_count"`   // 旷工次数
	AbsentDays    float64 `json:"absent_days"`    // 旷工天数
	MissingCount  int     `json:"missing_count"`  // 缺卡次数

	// 请假统计
	LeaveCount int     `json:"leave_count"` // 请假次数
	LeaveDays  float64 `json:"leave_days"`  // 请假天数

	// 加班统计
	OvertimeCount  int     `json:"overtime_count"`   // 加班次数
	OvertimeHours  float64 `json:"overtime_hours"`   // 加班小时数
	WeekendOTHours float64 `json:"weekend_ot_hours"` // 周末加班小时
	HolidayOTHours float64 `json:"holiday_ot_hours"` // 节假日加班小时
	CompOffDays    float64 `json:"comp_off_days"`    // 可调休天数

	// 出差统计
	TripCount int     `json:"trip_count"` // 出差次数
	TripDays  float64 `json:"trip_days"`  // 出差天数

	// 工时统计
	WorkHours         float64 `json:"work_hours"`          // 工作总时长
	StandardWorkHours float64 `json:"standard_work_hours"` // 标准工时

	// 状态
	Status string `json:"status"` // draft, confirmed, locked

	// 审计字段
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
	ConfirmedBy *uuid.UUID `json:"confirmed_by,omitempty"`
}
