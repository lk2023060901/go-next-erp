package model

import (
	"time"

	"github.com/google/uuid"
)

// HRMEmployee HRM员工扩展信息
// 扩展 organization.Employee，添加考勤系统所需的专属字段
type HRMEmployee struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 关联组织模块的员工
	EmployeeID uuid.UUID `json:"employee_id"` // 对应 organization.employees.id

	// 身份信息
	IDCardNo string `json:"id_card_no,omitempty"` // 身份证号（加密存储）

	// 考勤设备信息
	CardNo      string `json:"card_no,omitempty"`     // 考勤卡号
	FaceData    string `json:"face_data,omitempty"`   // 人脸特征数据（加密存储）
	Fingerprint string `json:"fingerprint,omitempty"` // 指纹数据（加密存储）

	// 第三方平台映射
	DingTalkUserID string `json:"dingtalk_user_id,omitempty"` // 钉钉 UserID
	WeComUserID    string `json:"wecom_user_id,omitempty"`    // 企业微信 UserID
	FeishuUserID   string `json:"feishu_user_id,omitempty"`   // 飞书 UserID
	FeishuOpenID   string `json:"feishu_open_id,omitempty"`   // 飞书 OpenID

	// 工作信息
	WorkLocation     string     `json:"work_location,omitempty"`      // 工作地点
	WorkScheduleType string     `json:"work_schedule_type,omitempty"` // 工作时间表类型：weekday, shift, flexible
	AttendanceRuleID *uuid.UUID `json:"attendance_rule_id,omitempty"` // 关联考勤规则

	// 默认班次（用于固定班次员工）
	DefaultShiftID *uuid.UUID `json:"default_shift_id,omitempty"`

	// 考勤设置
	AllowFieldWork  bool `json:"allow_field_work"` // 是否允许外勤打卡
	RequireFace     bool `json:"require_face"`     // 是否必须人脸识别
	RequireLocation bool `json:"require_location"` // 是否必须定位
	RequireWiFi     bool `json:"require_wifi"`     // 是否必须WiFi

	// 紧急联系人
	EmergencyContact  string `json:"emergency_contact,omitempty"`  // 紧急联系人姓名
	EmergencyPhone    string `json:"emergency_phone,omitempty"`    // 紧急联系人电话
	EmergencyRelation string `json:"emergency_relation,omitempty"` // 与紧急联系人关系

	// 状态
	IsActive bool `json:"is_active"` // 是否启用考勤（默认true）

	// 备注
	Remark string `json:"remark,omitempty"`

	// 审计字段
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// EmployeeSyncMapping 员工第三方平台同步映射
// 用于记录员工在各个第三方平台的账号映射关系和同步状态
type EmployeeSyncMapping struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 关联员工
	EmployeeID uuid.UUID `json:"employee_id"`

	// 平台信息
	Platform   PlatformType `json:"platform"`    // dingtalk, wecom, feishu
	PlatformID string       `json:"platform_id"` // 第三方平台的用户ID

	// 同步信息
	SyncEnabled bool       `json:"sync_enabled"`           // 是否启用同步
	LastSyncAt  *time.Time `json:"last_sync_at,omitempty"` // 最后同步时间
	SyncStatus  string     `json:"sync_status"`            // success, failed
	SyncError   string     `json:"sync_error,omitempty"`   // 同步错误信息

	// 映射数据（原始数据，JSON格式）
	RawData map[string]interface{} `json:"raw_data,omitempty"`

	// 审计字段
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EmployeeWorkSchedule 员工工作时间表
// 记录员工的工作安排（如固定班次、轮班等）
type EmployeeWorkSchedule struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 关联员工
	EmployeeID uuid.UUID `json:"employee_id"`

	// 时间表类型
	ScheduleType string `json:"schedule_type"` // weekday(标准工作日), shift(轮班), flexible(弹性), custom(自定义)

	// 工作日配置（适用于 weekday 类型）
	WorkDays  []int  `json:"work_days,omitempty"`  // 工作日：1-7 (1=周一)
	WorkHours int    `json:"work_hours,omitempty"` // 每日工作小时数
	WorkStart string `json:"work_start,omitempty"` // 标准上班时间 HH:MM
	WorkEnd   string `json:"work_end,omitempty"`   // 标准下班时间 HH:MM

	// 轮班配置（适用于 shift 类型）
	ShiftCycle   int    `json:"shift_cycle,omitempty"`   // 轮班周期（天）
	ShiftPattern string `json:"shift_pattern,omitempty"` // 轮班模式（如：早中晚休）

	// 生效时间
	EffectiveFrom time.Time  `json:"effective_from"`         // 生效开始时间
	EffectiveTo   *time.Time `json:"effective_to,omitempty"` // 生效结束时间

	// 状态
	IsActive bool `json:"is_active"` // 是否启用

	// 备注
	Remark string `json:"remark,omitempty"`

	// 审计字段
	CreatedBy uuid.UUID `json:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HasFaceData 是否有人脸数据
func (e *HRMEmployee) HasFaceData() bool {
	return e.FaceData != ""
}

// HasFingerprint 是否有指纹数据
func (e *HRMEmployee) HasFingerprint() bool {
	return e.Fingerprint != ""
}

// HasCardNo 是否有考勤卡号
func (e *HRMEmployee) HasCardNo() bool {
	return e.CardNo != ""
}

// GetThirdPartyID 获取第三方平台ID
func (e *HRMEmployee) GetThirdPartyID(platform PlatformType) string {
	switch platform {
	case PlatformDingTalk:
		return e.DingTalkUserID
	case PlatformWeCom:
		return e.WeComUserID
	case PlatformFeishu:
		return e.FeishuUserID
	default:
		return ""
	}
}

// SetThirdPartyID 设置第三方平台ID
func (e *HRMEmployee) SetThirdPartyID(platform PlatformType, id string) {
	switch platform {
	case PlatformDingTalk:
		e.DingTalkUserID = id
	case PlatformWeCom:
		e.WeComUserID = id
	case PlatformFeishu:
		e.FeishuUserID = id
	}
}
