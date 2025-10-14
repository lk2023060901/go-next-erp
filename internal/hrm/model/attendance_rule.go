package model

import (
	"time"

	"github.com/google/uuid"
)

// AttendanceRule 考勤规则（考勤组）
type AttendanceRule struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 规则基本信息
	Code        string `json:"code"`        // 规则编码
	Name        string `json:"name"`        // 规则名称
	Description string `json:"description"` // 描述

	// 适用范围
	ApplyType     ApplyType   `json:"apply_type"`               // all, department, employee
	DepartmentIDs []uuid.UUID `json:"department_ids,omitempty"` // 适用部门
	EmployeeIDs   []uuid.UUID `json:"employee_ids,omitempty"`   // 适用员工

	// 工作制
	WorkdayType WorkdayType `json:"workday_type"` // five_day, six_day, custom
	WeekendDays []int       `json:"weekend_days"` // 0=周日, 1=周一, ..., 6=周六

	// 默认班次
	DefaultShiftID *uuid.UUID `json:"default_shift_id,omitempty"`

	// 打卡位置限制
	LocationRequired bool              `json:"location_required"`           // 是否必须定位
	AllowedLocations []AllowedLocation `json:"allowed_locations,omitempty"` // 允许的打卡位置

	// WiFi限制
	WiFiRequired bool     `json:"wifi_required"`          // 是否必须连接指定WiFi
	AllowedWiFi  []string `json:"allowed_wifi,omitempty"` // 允许的WiFi列表（SSID或MAC）

	// 人脸识别
	FaceRequired     bool    `json:"face_required"`      // 是否必须人脸识别
	FaceThreshold    float64 `json:"face_threshold"`     // 人脸识别阈值
	FaceAntiSpoofing bool    `json:"face_anti_spoofing"` // 是否开启活体检测

	// 外勤打卡
	AllowFieldWork bool `json:"allow_field_work"` // 是否允许外勤打卡

	// 节假日设置
	HolidayCalendarID *uuid.UUID `json:"holiday_calendar_id,omitempty"` // 关联假期日历

	// 审批设置
	RequireApprovalForLate  bool `json:"require_approval_for_late"`  // 迟到需要审批
	RequireApprovalForEarly bool `json:"require_approval_for_early"` // 早退需要审批

	// 状态
	IsActive bool `json:"is_active"` // 是否启用
	Priority int  `json:"priority"`  // 优先级，数字越大优先级越高

	// 审计字段
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy uuid.UUID  `json:"updated_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ApplyType 适用类型
type ApplyType string

const (
	ApplyTypeAll        ApplyType = "all"        // 全员
	ApplyTypeDepartment ApplyType = "department" // 按部门
	ApplyTypeEmployee   ApplyType = "employee"   // 按员工
)

// WorkdayType 工作制类型
type WorkdayType string

const (
	WorkdayTypeFiveDay WorkdayType = "five_day" // 标准工作制（周一到周五）
	WorkdayTypeSixDay  WorkdayType = "six_day"  // 大小周（六天工作制）
	WorkdayTypeCustom  WorkdayType = "custom"   // 自定义
)

// AllowedLocation 允许的打卡位置
type AllowedLocation struct {
	Name      string  `json:"name"`      // 位置名称
	Latitude  float64 `json:"latitude"`  // 纬度
	Longitude float64 `json:"longitude"` // 经度
	Radius    int     `json:"radius"`    // 半径（米）
	Address   string  `json:"address"`   // 地址
}
