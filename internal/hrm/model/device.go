package model

import (
	"time"

	"github.com/google/uuid"
)

// AttendanceDevice 考勤设备（考勤机）
type AttendanceDevice struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 设备信息
	DeviceType  DeviceType `json:"device_type"`  // zkteco, dingtalk_m2, deli等
	DeviceSN    string     `json:"device_sn"`    // 设备序列号（唯一标识）
	DeviceName  string     `json:"device_name"`  // 设备名称
	DeviceModel string     `json:"device_model"` // 设备型号

	// 网络信息
	IPAddress  string `json:"ip_address,omitempty"`  // IP地址
	Port       int    `json:"port,omitempty"`        // 端口号
	MACAddress string `json:"mac_address,omitempty"` // MAC地址

	// 位置信息
	Location       *LocationInfo `json:"location,omitempty"`        // 设备位置
	InstallAddress string        `json:"install_address,omitempty"` // 安装地址
	DepartmentID   *uuid.UUID    `json:"department_id,omitempty"`   // 关联部门

	// 认证信息
	AuthType  string `json:"auth_type,omitempty"` // 认证方式：password, apikey, certificate
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"` // 加密存储
	APIKey    string `json:"api_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`

	// 同步配置
	SyncEnabled  bool       `json:"sync_enabled"`           // 是否启用同步
	SyncInterval int        `json:"sync_interval"`          // 同步间隔（分钟）
	SyncMode     string     `json:"sync_mode"`              // push, pull
	LastSyncAt   *time.Time `json:"last_sync_at,omitempty"` // 最后同步时间

	// 功能支持
	SupportFace        bool `json:"support_face"`        // 支持人脸识别
	SupportFingerprint bool `json:"support_fingerprint"` // 支持指纹
	SupportCard        bool `json:"support_card"`        // 支持刷卡
	SupportTemperature bool `json:"support_temperature"` // 支持体温检测

	// 设备状态
	Status        DeviceStatus `json:"status"`                   // online, offline, error
	IsActive      bool         `json:"is_active"`                // 是否启用
	LastHeartbeat *time.Time   `json:"last_heartbeat,omitempty"` // 最后心跳时间
	ErrorMessage  string       `json:"error_message,omitempty"`  // 错误信息

	// 统计信息
	TotalRecords int `json:"total_records"` // 总记录数
	TodayRecords int `json:"today_records"` // 今日记录数

	// 备注
	Remark string `json:"remark,omitempty"`

	// 审计字段
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy uuid.UUID  `json:"updated_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// DeviceType 设备类型
type DeviceType string

const (
	DeviceTypeZKTeco     DeviceType = "zkteco"      // 中控智慧
	DeviceTypeDingTalkM2 DeviceType = "dingtalk_m2" // 钉钉M2
	DeviceTypeDeli       DeviceType = "deli"        // 得力
	DeviceTypeHikvision  DeviceType = "hikvision"   // 海康威视
	DeviceTypeDahua      DeviceType = "dahua"       // 大华
	DeviceTypeOther      DeviceType = "other"       // 其他
)

// DeviceStatus 设备状态
type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"  // 在线
	DeviceStatusOffline DeviceStatus = "offline" // 离线
	DeviceStatusError   DeviceStatus = "error"   // 异常
)

// ThirdPartyIntegration 第三方平台集成配置
type ThirdPartyIntegration struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 平台信息
	Platform  PlatformType `json:"platform"`   // dingtalk, wecom, feishu
	AppName   string       `json:"app_name"`   // 应用名称
	AppID     string       `json:"app_id"`     // 应用ID
	AppKey    string       `json:"app_key"`    // AppKey
	AppSecret string       `json:"app_secret"` // AppSecret（加密存储）

	// 企业信息
	CorpID      string `json:"corp_id,omitempty"`      // 企业ID
	AgentID     string `json:"agent_id,omitempty"`     // 应用AgentID（企微）
	SuiteKey    string `json:"suite_key,omitempty"`    // 套件Key（飞书）
	SuiteSecret string `json:"suite_secret,omitempty"` // 套件Secret（飞书）

	// Webhook配置
	WebhookURL    string `json:"webhook_url,omitempty"`    // Webhook接收地址
	WebhookToken  string `json:"webhook_token,omitempty"`  // Webhook验证Token
	WebhookSecret string `json:"webhook_secret,omitempty"` // Webhook加密Secret

	// 同步配置
	SyncEnabled    bool       `json:"sync_enabled"`    // 是否启用同步
	SyncAttendance bool       `json:"sync_attendance"` // 同步考勤记录
	SyncEmployee   bool       `json:"sync_employee"`   // 同步员工信息
	SyncDepartment bool       `json:"sync_department"` // 同步部门信息
	SyncSchedule   bool       `json:"sync_schedule"`   // 同步排班
	SyncInterval   int        `json:"sync_interval"`   // 同步间隔（分钟）
	SyncDirection  string     `json:"sync_direction"`  // both, pull, push
	LastSyncAt     *time.Time `json:"last_sync_at,omitempty"`

	// 字段映射配置
	FieldMapping map[string]string `json:"field_mapping,omitempty"` // 字段映射关系

	// 状态
	Status       IntegrationStatus `json:"status"`                  // active, inactive, error
	IsActive     bool              `json:"is_active"`               // 是否启用
	ErrorMessage string            `json:"error_message,omitempty"` // 错误信息

	// 统计信息
	TotalSyncCount int `json:"total_sync_count"` // 总同步次数
	LastSyncCount  int `json:"last_sync_count"`  // 最后一次同步数量

	// 备注
	Remark string `json:"remark,omitempty"`

	// 审计字段
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy uuid.UUID  `json:"updated_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// PlatformType 平台类型
type PlatformType string

const (
	PlatformDingTalk PlatformType = "dingtalk" // 钉钉
	PlatformWeCom    PlatformType = "wecom"    // 企业微信
	PlatformFeishu   PlatformType = "feishu"   // 飞书
)

// IntegrationStatus 集成状态
type IntegrationStatus string

const (
	IntegrationStatusActive   IntegrationStatus = "active"   // 正常
	IntegrationStatusInactive IntegrationStatus = "inactive" // 未激活
	IntegrationStatusError    IntegrationStatus = "error"    // 异常
)

// SyncLog 同步日志
type SyncLog struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 同步来源
	SourceType string    `json:"source_type"` // device, dingtalk, wecom, feishu
	SourceID   uuid.UUID `json:"source_id"`   // 设备或集成配置ID

	// 同步信息
	SyncType      string    `json:"sync_type"`      // attendance, employee, department
	SyncDirection string    `json:"sync_direction"` // pull, push
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Duration      int       `json:"duration"` // 毫秒

	// 同步结果
	Status       string `json:"status"`        // success, failed, partial
	TotalCount   int    `json:"total_count"`   // 总数
	SuccessCount int    `json:"success_count"` // 成功数
	FailedCount  int    `json:"failed_count"`  // 失败数
	ErrorMessage string `json:"error_message,omitempty"`

	// 详细信息
	Details map[string]interface{} `json:"details,omitempty"`

	// 审计字段
	CreatedAt time.Time `json:"created_at"`
}
