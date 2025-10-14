package integration

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
)

// AttendanceIntegration 考勤集成接口
// 统一的考勤数据同步接口,支持钉钉、企业微信、飞书等平台
type AttendanceIntegration interface {
	// SyncAttendanceRecords 同步考勤记录
	SyncAttendanceRecords(ctx context.Context, req *SyncAttendanceRequest) ([]*model.AttendanceRecord, error)

	// SyncEmployees 同步员工信息
	SyncEmployees(ctx context.Context, req *SyncEmployeeRequest) ([]*model.EmployeeSyncMapping, error)

	// PushAttendanceRule 推送考勤规则
	PushAttendanceRule(ctx context.Context, req *PushAttendanceRuleRequest) error

	// GetEmployeeAttendance 获取员工考勤数据
	GetEmployeeAttendance(ctx context.Context, req *GetEmployeeAttendanceRequest) ([]*model.AttendanceRecord, error)
}

// SyncAttendanceRequest 同步考勤记录请求
type SyncAttendanceRequest struct {
	TenantID   uuid.UUID
	EmployeeID uuid.UUID
	StartDate  time.Time
	EndDate    time.Time
}

// SyncEmployeeRequest 同步员工请求
type SyncEmployeeRequest struct {
	TenantID     uuid.UUID
	DepartmentID string
}

// PushAttendanceRuleRequest 推送考勤规则请求
type PushAttendanceRuleRequest struct {
	TenantID uuid.UUID
	RuleID   uuid.UUID
}

// GetEmployeeAttendanceRequest 获取员工考勤请求
type GetEmployeeAttendanceRequest struct {
	TenantID   uuid.UUID
	EmployeeID uuid.UUID
	StartDate  time.Time
	EndDate    time.Time
}

// DeviceAdapter 考勤设备适配器接口
// 用于对接不同品牌的考勤机：中控智慧、钉钉M2、得力、海康威视等
type DeviceAdapter interface {
	// GetType 获取设备类型
	GetType() model.DeviceType

	// Connect 连接设备
	Connect(ctx context.Context, config *DeviceConfig) error

	// Disconnect 断开连接
	Disconnect() error

	// Ping 测试连接（心跳检测）
	Ping(ctx context.Context) error

	// PullRecords 拉取考勤记录
	// startTime: 开始时间
	// endTime: 结束时间
	// 返回：考勤记录列表、错误
	PullRecords(ctx context.Context, startTime, endTime time.Time) ([]*AttendanceRecordDTO, error)

	// PushEmployee 推送员工信息到设备
	PushEmployee(ctx context.Context, employee *EmployeeDTO) error

	// BatchPushEmployees 批量推送员工信息
	BatchPushEmployees(ctx context.Context, employees []*EmployeeDTO) error

	// DeleteEmployee 从设备删除员工
	DeleteEmployee(ctx context.Context, employeeID string) error

	// GetDeviceInfo 获取设备信息
	GetDeviceInfo(ctx context.Context) (*DeviceInfo, error)

	// GetDeviceStatus 获取设备状态
	GetDeviceStatus(ctx context.Context) (*DeviceStatus, error)

	// ClearRecords 清空设备记录
	ClearRecords(ctx context.Context) error
}

// PlatformAdapter 第三方平台适配器接口
// 用于对接钉钉、企业微信、飞书等平台的考勤API
type PlatformAdapter interface {
	// GetPlatform 获取平台类型
	GetPlatform() model.PlatformType

	// Init 初始化配置
	Init(ctx context.Context, config *PlatformConfig) error

	// GetAccessToken 获取访问令牌
	GetAccessToken(ctx context.Context) (string, error)

	// RefreshToken 刷新令牌
	RefreshToken(ctx context.Context) (string, error)

	// PullAttendanceRecords 拉取考勤记录
	PullAttendanceRecords(ctx context.Context, startTime, endTime time.Time, userIDs []string) ([]*AttendanceRecordDTO, error)

	// PullEmployees 拉取员工信息
	PullEmployees(ctx context.Context) ([]*EmployeeDTO, error)

	// PullDepartments 拉取部门信息
	PullDepartments(ctx context.Context) ([]*DepartmentDTO, error)

	// PushSchedule 推送排班信息
	PushSchedule(ctx context.Context, schedules []*ScheduleDTO) error

	// RegisterWebhook 注册Webhook回调
	RegisterWebhook(ctx context.Context, events []string, callbackURL string) error

	// HandleWebhook 处理Webhook回调
	HandleWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error)

	// ValidateSignature 验证签名
	ValidateSignature(payload []byte, signature string) bool
}

// DeviceConfig 设备配置
type DeviceConfig struct {
	DeviceType model.DeviceType `json:"device_type"`
	IPAddress  string           `json:"ip_address"`
	Port       int              `json:"port"`
	Username   string           `json:"username"`
	Password   string           `json:"password"`
	APIKey     string           `json:"api_key,omitempty"`
	SecretKey  string           `json:"secret_key,omitempty"`
	Timeout    int              `json:"timeout"` // 超时时间（秒）
}

// PlatformConfig 平台配置
type PlatformConfig struct {
	Platform    model.PlatformType `json:"platform"`
	AppID       string             `json:"app_id"`
	AppKey      string             `json:"app_key"`
	AppSecret   string             `json:"app_secret"`
	CorpID      string             `json:"corp_id,omitempty"`
	AgentID     string             `json:"agent_id,omitempty"`
	SuiteKey    string             `json:"suite_key,omitempty"`
	SuiteSecret string             `json:"suite_secret,omitempty"`
}

// AttendanceRecordDTO 考勤记录数据传输对象
type AttendanceRecordDTO struct {
	UserID        string                    `json:"user_id"`               // 用户ID（第三方平台ID或工号）
	UserName      string                    `json:"user_name"`             // 用户姓名
	ClockTime     time.Time                 `json:"clock_time"`            // 打卡时间
	ClockType     model.AttendanceClockType `json:"clock_type"`            // 打卡类型
	CheckInMethod model.AttendanceMethod    `json:"check_in_method"`       // 打卡方式
	Location      *model.LocationInfo       `json:"location,omitempty"`    // 定位
	Address       string                    `json:"address,omitempty"`     // 地址
	PhotoURL      string                    `json:"photo_url,omitempty"`   // 照片
	Temperature   float64                   `json:"temperature,omitempty"` // 体温
	DeviceID      string                    `json:"device_id,omitempty"`   // 设备ID
	RawData       map[string]interface{}    `json:"raw_data,omitempty"`    // 原始数据
}

// EmployeeDTO 员工数据传输对象
type EmployeeDTO struct {
	EmployeeID   string `json:"employee_id"`           // 员工ID
	EmployeeNo   string `json:"employee_no"`           // 工号
	Name         string `json:"name"`                  // 姓名
	DepartmentID string `json:"department_id"`         // 部门ID
	Mobile       string `json:"mobile,omitempty"`      // 手机号
	Email        string `json:"email,omitempty"`       // 邮箱
	CardNo       string `json:"card_no,omitempty"`     // 卡号
	FaceData     string `json:"face_data,omitempty"`   // 人脸数据
	Fingerprint  string `json:"fingerprint,omitempty"` // 指纹数据
	Status       string `json:"status"`                // 状态
}

// DepartmentDTO 部门数据传输对象
type DepartmentDTO struct {
	DepartmentID string `json:"department_id"`       // 部门ID
	Name         string `json:"name"`                // 部门名称
	ParentID     string `json:"parent_id,omitempty"` // 父部门ID
	Sort         int    `json:"sort"`                // 排序
}

// ScheduleDTO 排班数据传输对象
type ScheduleDTO struct {
	UserID       string    `json:"user_id"`       // 用户ID
	ScheduleDate time.Time `json:"schedule_date"` // 排班日期
	ShiftName    string    `json:"shift_name"`    // 班次名称
	WorkStart    string    `json:"work_start"`    // 上班时间
	WorkEnd      string    `json:"work_end"`      // 下班时间
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	DeviceSN       string `json:"device_sn"`       // 设备序列号
	DeviceModel    string `json:"device_model"`    // 设备型号
	FirmwareVer    string `json:"firmware_ver"`    // 固件版本
	UserCapacity   int    `json:"user_capacity"`   // 用户容量
	RecordCapacity int    `json:"record_capacity"` // 记录容量
	UserCount      int    `json:"user_count"`      // 当前用户数
	RecordCount    int    `json:"record_count"`    // 当前记录数
}

// DeviceStatus 设备状态
type DeviceStatus struct {
	IsOnline      bool      `json:"is_online"`               // 是否在线
	CPUUsage      float64   `json:"cpu_usage"`               // CPU使用率
	MemoryUsage   float64   `json:"memory_usage"`            // 内存使用率
	DiskUsage     float64   `json:"disk_usage"`              // 磁盘使用率
	LastHeartbeat time.Time `json:"last_heartbeat"`          // 最后心跳时间
	ErrorMessage  string    `json:"error_message,omitempty"` // 错误信息
}

// WebhookEvent Webhook事件
type WebhookEvent struct {
	EventType string                 `json:"event_type"` // 事件类型
	Timestamp time.Time              `json:"timestamp"`  // 时间戳
	Data      map[string]interface{} `json:"data"`       // 事件数据
}
