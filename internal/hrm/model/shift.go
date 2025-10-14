package model

import (
	"time"

	"github.com/google/uuid"
)

// Shift 班次（定义工作时间段）
type Shift struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 班次基本信息
	Code        string `json:"code"`        // 班次编码（唯一）
	Name        string `json:"name"`        // 班次名称
	Description string `json:"description"` // 描述

	// 班次类型
	Type ShiftType `json:"type"` // fixed, flexible, free

	// 固定班次配置
	WorkStart string `json:"work_start,omitempty"` // 上班时间 HH:MM
	WorkEnd   string `json:"work_end,omitempty"`   // 下班时间 HH:MM

	// 弹性班次配置
	FlexibleStart string `json:"flexible_start,omitempty"` // 弹性上班开始时间
	FlexibleEnd   string `json:"flexible_end,omitempty"`   // 弹性上班结束时间
	WorkDuration  int    `json:"work_duration,omitempty"`  // 工作时长（分钟）

	// 打卡规则
	CheckInRequired  bool `json:"check_in_required"`  // 是否必须上班打卡
	CheckOutRequired bool `json:"check_out_required"` // 是否必须下班打卡

	// 迟到早退规则
	LateGracePeriod  int `json:"late_grace_period"`  // 迟到宽限时间（分钟）
	EarlyGracePeriod int `json:"early_grace_period"` // 早退宽限时间（分钟）

	// 休息时间
	RestPeriods []RestPeriod `json:"rest_periods,omitempty"` // 休息时间段

	// 跨天标识
	IsCrossDays bool `json:"is_cross_days"` // 是否跨天班次（如夜班）

	// 加班规则
	AllowOvertime       bool    `json:"allow_overtime"`        // 是否允许加班
	OvertimeStartBuffer int     `json:"overtime_start_buffer"` // 加班开始缓冲（分钟）
	OvertimeMinDuration int     `json:"overtime_min_duration"` // 最小加班时长（分钟）
	OvertimePayRate     float64 `json:"overtime_pay_rate"`     // 加班倍率

	// 工作日类型
	WorkdayTypes []string `json:"workday_types"` // workday, weekend, holiday

	// 颜色标识（用于排班表显示）
	Color string `json:"color,omitempty"`

	// 状态
	IsActive bool `json:"is_active"` // 是否启用
	Sort     int  `json:"sort"`      // 排序

	// 审计字段
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy uuid.UUID  `json:"updated_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ShiftType 班次类型
type ShiftType string

const (
	ShiftTypeFixed    ShiftType = "fixed"    // 固定班次
	ShiftTypeFlexible ShiftType = "flexible" // 弹性班次
	ShiftTypeFree     ShiftType = "free"     // 自由班次（不打卡）
)

// RestPeriod 休息时间段
type RestPeriod struct {
	StartTime string `json:"start_time"` // HH:MM
	EndTime   string `json:"end_time"`   // HH:MM
	Duration  int    `json:"duration"`   // 分钟数
}

// Schedule 排班记录
type Schedule struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 员工信息（关联 organization.employees）
	EmployeeID   uuid.UUID `json:"employee_id"`   // 对应 organization.employees.id
	EmployeeName string    `json:"employee_name"` // 冗余
	DepartmentID uuid.UUID `json:"department_id"` // 冗余

	// 班次信息
	ShiftID   uuid.UUID `json:"shift_id"`
	ShiftName string    `json:"shift_name"` // 冗余

	// 排班日期
	ScheduleDate time.Time `json:"schedule_date"` // 排班日期
	WorkdayType  string    `json:"workday_type"`  // workday, weekend, holiday

	// 状态
	Status string `json:"status"` // draft, published, executed

	// 备注
	Remark string `json:"remark,omitempty"`

	// 审计字段
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy uuid.UUID  `json:"updated_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
