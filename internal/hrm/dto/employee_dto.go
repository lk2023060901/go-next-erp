package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	orgmodel "github.com/lk2023060901/go-next-erp/internal/organization/model"
)

// EmployeeWithHRM 员工完整信息（组织信息 + HRM扩展信息）
type EmployeeWithHRM struct {
	// 基础员工信息（来自 organization 模块）
	*orgmodel.Employee

	// HRM 扩展信息
	HRMInfo *model.HRMEmployee `json:"hrm_info,omitempty"`

	// 考勤规则信息（冗余字段，便于展示）
	AttendanceRuleName string `json:"attendance_rule_name,omitempty"`

	// 默认班次信息（冗余字段，便于展示）
	DefaultShiftName string `json:"default_shift_name,omitempty"`
}

// EmployeeAttendanceInfo 员工考勤相关信息
type EmployeeAttendanceInfo struct {
	// 员工基本信息
	EmployeeID uuid.UUID `json:"employee_id"`
	EmployeeNo string    `json:"employee_no"`
	Name       string    `json:"name"`
	Gender     string    `json:"gender"`
	Mobile     string    `json:"mobile"`
	Avatar     string    `json:"avatar"`

	// 组织信息
	OrgID   uuid.UUID `json:"org_id"`
	OrgName string    `json:"org_name"`
	OrgPath string    `json:"org_path"`

	// 职位信息
	PositionID   *uuid.UUID `json:"position_id,omitempty"`
	PositionName string     `json:"position_name,omitempty"`

	// 入职状态
	JoinDate *time.Time `json:"join_date,omitempty"`
	Status   string     `json:"status"` // probation, active, resigned

	// HRM 考勤信息
	CardNo             string     `json:"card_no,omitempty"`
	WorkLocation       string     `json:"work_location,omitempty"`
	AttendanceRuleID   *uuid.UUID `json:"attendance_rule_id,omitempty"`
	AttendanceRuleName string     `json:"attendance_rule_name,omitempty"`
	DefaultShiftID     *uuid.UUID `json:"default_shift_id,omitempty"`
	DefaultShiftName   string     `json:"default_shift_name,omitempty"`

	// 第三方平台映射
	DingTalkUserID string `json:"dingtalk_user_id,omitempty"`
	WeComUserID    string `json:"wecom_user_id,omitempty"`
	FeishuUserID   string `json:"feishu_user_id,omitempty"`

	// 考勤设置
	AllowFieldWork  bool `json:"allow_field_work"`
	RequireFace     bool `json:"require_face"`
	RequireLocation bool `json:"require_location"`
	RequireWiFi     bool `json:"require_wifi"`

	// 生物识别状态
	HasFaceData    bool `json:"has_face_data"`
	HasFingerprint bool `json:"has_fingerprint"`

	// 考勤启用状态
	IsAttendanceActive bool `json:"is_attendance_active"`
}

// CreateHRMEmployeeRequest 创建HRM员工扩展信息请求
type CreateHRMEmployeeRequest struct {
	EmployeeID uuid.UUID `json:"employee_id" binding:"required"` // 组织员工ID

	// 身份信息
	IDCardNo string `json:"id_card_no,omitempty"`

	// 考勤设备信息
	CardNo      string `json:"card_no,omitempty"`
	FaceData    string `json:"face_data,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`

	// 第三方平台映射
	DingTalkUserID string `json:"dingtalk_user_id,omitempty"`
	WeComUserID    string `json:"wecom_user_id,omitempty"`
	FeishuUserID   string `json:"feishu_user_id,omitempty"`
	FeishuOpenID   string `json:"feishu_open_id,omitempty"`

	// 工作信息
	WorkLocation     string     `json:"work_location,omitempty"`
	WorkScheduleType string     `json:"work_schedule_type,omitempty"`
	AttendanceRuleID *uuid.UUID `json:"attendance_rule_id,omitempty"`
	DefaultShiftID   *uuid.UUID `json:"default_shift_id,omitempty"`

	// 考勤设置
	AllowFieldWork  bool `json:"allow_field_work"`
	RequireFace     bool `json:"require_face"`
	RequireLocation bool `json:"require_location"`
	RequireWiFi     bool `json:"require_wifi"`

	// 紧急联系人
	EmergencyContact  string `json:"emergency_contact,omitempty"`
	EmergencyPhone    string `json:"emergency_phone,omitempty"`
	EmergencyRelation string `json:"emergency_relation,omitempty"`

	// 备注
	Remark string `json:"remark,omitempty"`
}

// UpdateHRMEmployeeRequest 更新HRM员工扩展信息请求
type UpdateHRMEmployeeRequest struct {
	// 考勤设备信息
	CardNo      *string `json:"card_no,omitempty"`
	FaceData    *string `json:"face_data,omitempty"`
	Fingerprint *string `json:"fingerprint,omitempty"`

	// 第三方平台映射
	DingTalkUserID *string `json:"dingtalk_user_id,omitempty"`
	WeComUserID    *string `json:"wecom_user_id,omitempty"`
	FeishuUserID   *string `json:"feishu_user_id,omitempty"`
	FeishuOpenID   *string `json:"feishu_open_id,omitempty"`

	// 工作信息
	WorkLocation     *string    `json:"work_location,omitempty"`
	WorkScheduleType *string    `json:"work_schedule_type,omitempty"`
	AttendanceRuleID *uuid.UUID `json:"attendance_rule_id,omitempty"`
	DefaultShiftID   *uuid.UUID `json:"default_shift_id,omitempty"`

	// 考勤设置
	AllowFieldWork  *bool `json:"allow_field_work,omitempty"`
	RequireFace     *bool `json:"require_face,omitempty"`
	RequireLocation *bool `json:"require_location,omitempty"`
	RequireWiFi     *bool `json:"require_wifi,omitempty"`

	// 紧急联系人
	EmergencyContact  *string `json:"emergency_contact,omitempty"`
	EmergencyPhone    *string `json:"emergency_phone,omitempty"`
	EmergencyRelation *string `json:"emergency_relation,omitempty"`

	// 状态
	IsActive *bool `json:"is_active,omitempty"`

	// 备注
	Remark *string `json:"remark,omitempty"`
}

// EmployeeSyncRequest 员工同步请求
type EmployeeSyncRequest struct {
	Platform    model.PlatformType     `json:"platform" binding:"required"`    // dingtalk, wecom, feishu
	PlatformID  string                 `json:"platform_id" binding:"required"` // 第三方平台用户ID
	EmployeeID  uuid.UUID              `json:"employee_id" binding:"required"` // 内部员工ID
	SyncEnabled bool                   `json:"sync_enabled"`                   // 是否启用同步
	RawData     map[string]interface{} `json:"raw_data,omitempty"`             // 原始数据
}

// EmployeeBatchSyncRequest 批量同步请求
type EmployeeBatchSyncRequest struct {
	Platform model.PlatformType `json:"platform" binding:"required"` // dingtalk, wecom, feishu
	Mappings []struct {
		PlatformID string    `json:"platform_id" binding:"required"`
		EmployeeID uuid.UUID `json:"employee_id" binding:"required"`
	} `json:"mappings" binding:"required"`
}

// EmployeeSyncStatusResponse 同步状态响应
type EmployeeSyncStatusResponse struct {
	EmployeeID  uuid.UUID          `json:"employee_id"`
	Platform    model.PlatformType `json:"platform"`
	PlatformID  string             `json:"platform_id"`
	SyncEnabled bool               `json:"sync_enabled"`
	LastSyncAt  *time.Time         `json:"last_sync_at,omitempty"`
	SyncStatus  string             `json:"sync_status"`
	SyncError   string             `json:"sync_error,omitempty"`
}
