package model

import (
	"time"

	"github.com/google/uuid"
)

// AttendanceRecord 考勤记录
type AttendanceRecord struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 员工信息（关联 organization.employees）
	EmployeeID   uuid.UUID `json:"employee_id"`   // 对应 organization.employees.id
	EmployeeName string    `json:"employee_name"` // 冗余字段，便于查询
	DepartmentID uuid.UUID `json:"department_id"` // 冗余字段，对应 organization.organizations.id

	// 打卡信息
	ClockTime time.Time           `json:"clock_time"` // 打卡时间
	ClockType AttendanceClockType `json:"clock_type"` // 上班/下班
	Status    AttendanceStatus    `json:"status"`     // 正常/迟到/早退/旷工

	// 班次信息
	ShiftID   *uuid.UUID `json:"shift_id,omitempty"`   // 关联班次
	ShiftName string     `json:"shift_name,omitempty"` // 冗余字段

	// 打卡方式和来源
	CheckInMethod AttendanceMethod `json:"check_in_method"` // 打卡方式
	SourceType    SourceType       `json:"source_type"`     // 数据来源
	SourceID      string           `json:"source_id"`       // 来源标识（设备ID或平台ID）

	// 定位信息（支持地理围栏打卡）
	Location *LocationInfo `json:"location,omitempty"`  // GPS定位
	Address  string        `json:"address,omitempty"`   // 地址
	WiFiSSID string        `json:"wifi_ssid,omitempty"` // WiFi名称
	WiFiMAC  string        `json:"wifi_mac,omitempty"`  // WiFi MAC地址

	// 生物识别信息
	PhotoURL    string  `json:"photo_url,omitempty"`   // 打卡照片
	FaceScore   float64 `json:"face_score,omitempty"`  // 人脸识别分数
	Temperature float64 `json:"temperature,omitempty"` // 体温（疫情期间）

	// 异常信息
	IsException     bool   `json:"is_exception"`               // 是否异常
	ExceptionReason string `json:"exception_reason,omitempty"` // 异常原因
	ExceptionType   string `json:"exception_type,omitempty"`   // 异常类型（迟到、早退、缺卡等）

	// 审批关联
	ApprovalID *uuid.UUID `json:"approval_id,omitempty"` // 补卡审批ID

	// 原始数据（用于问题排查和审计）
	RawData map[string]interface{} `json:"raw_data,omitempty"`

	// 备注
	Remark string `json:"remark,omitempty"`

	// 审计字段
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// LocationInfo GPS定位信息
type LocationInfo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Accuracy  float64 `json:"accuracy,omitempty"` // 精度（米）
}

// AttendanceClockType 打卡类型
type AttendanceClockType string

const (
	ClockTypeCheckIn  AttendanceClockType = "check_in"  // 上班打卡
	ClockTypeCheckOut AttendanceClockType = "check_out" // 下班打卡
)

// AttendanceStatus 考勤状态
type AttendanceStatus string

const (
	AttendanceStatusNormal   AttendanceStatus = "normal"   // 正常
	AttendanceStatusLate     AttendanceStatus = "late"     // 迟到
	AttendanceStatusEarly    AttendanceStatus = "early"    // 早退
	AttendanceStatusAbsent   AttendanceStatus = "absent"   // 旷工
	AttendanceStatusLeave    AttendanceStatus = "leave"    // 请假
	AttendanceStatusOvertime AttendanceStatus = "overtime" // 加班
	AttendanceStatusTrip     AttendanceStatus = "trip"     // 出差
)

// AttendanceMethod 打卡方式
type AttendanceMethod string

const (
	MethodDevice      AttendanceMethod = "device"      // 考勤机
	MethodMobile      AttendanceMethod = "mobile"      // 手机APP
	MethodWeb         AttendanceMethod = "web"         // 网页端
	MethodFace        AttendanceMethod = "face"        // 人脸识别
	MethodFingerprint AttendanceMethod = "fingerprint" // 指纹
	MethodCard        AttendanceMethod = "card"        // 刷卡
	MethodManual      AttendanceMethod = "manual"      // 手动补卡
)

// SourceType 数据来源类型
type SourceType string

const (
	SourceTypeSystem   SourceType = "system"   // 系统内部
	SourceTypeDevice   SourceType = "device"   // 考勤机设备
	SourceTypeDingTalk SourceType = "dingtalk" // 钉钉
	SourceTypeWeCom    SourceType = "wecom"    // 企业微信
	SourceTypeFeishu   SourceType = "feishu"   // 飞书
	SourceTypeManual   SourceType = "manual"   // 手动录入
)

// IsLate 判断是否迟到
func (r *AttendanceRecord) IsLate() bool {
	return r.Status == AttendanceStatusLate
}

// IsEarly 判断是否早退
func (r *AttendanceRecord) IsEarly() bool {
	return r.Status == AttendanceStatusEarly
}

// IsAbsent 判断是否旷工
func (r *AttendanceRecord) IsAbsent() bool {
	return r.Status == AttendanceStatusAbsent
}

// IsNormal 判断是否正常
func (r *AttendanceRecord) IsNormal() bool {
	return r.Status == AttendanceStatusNormal
}

// FromThirdParty 判断是否来自第三方平台
func (r *AttendanceRecord) FromThirdParty() bool {
	return r.SourceType == SourceTypeDingTalk ||
		r.SourceType == SourceTypeWeCom ||
		r.SourceType == SourceTypeFeishu
}

// FromDevice 判断是否来自考勤机
func (r *AttendanceRecord) FromDevice() bool {
	return r.SourceType == SourceTypeDevice
}
