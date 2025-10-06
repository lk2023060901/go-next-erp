package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditLog 审计日志模型
type AuditLog struct {
	ID        uuid.UUID `json:"id"`         // UUID v7
	EventID   string    `json:"event_id"`   // 事件唯一标识
	TenantID  uuid.UUID `json:"tenant_id"`  // 租户 ID
	UserID    uuid.UUID `json:"user_id"`    // 操作用户 ID
	Action    string    `json:"action"`     // 操作动作（如：user.login, document.delete）
	Resource  string    `json:"resource"`   // 资源类型
	ResourceID string   `json:"resource_id,omitempty"` // 资源 ID

	// 操作详情
	BeforeData json.RawMessage `json:"before_data,omitempty"` // 操作前数据
	AfterData  json.RawMessage `json:"after_data,omitempty"`  // 操作后数据

	// 请求信息
	IPAddress string `json:"ip_address"` // 客户端 IP
	UserAgent string `json:"user_agent"` // 用户代理

	// 结果
	Result   AuditResult `json:"result"`   // 操作结果
	ErrorMsg string      `json:"error_msg,omitempty"` // 错误信息

	// 元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"` // 额外信息

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
}

// AuditResult 审计结果
type AuditResult string

const (
	AuditResultSuccess AuditResult = "success" // 成功
	AuditResultFailure AuditResult = "failure" // 失败
	AuditResultDenied  AuditResult = "denied"  // 拒绝（权限不足）
)

// Action 定义标准审计动作
const (
	// 认证相关
	AuditActionLogin        = "user.login"
	AuditActionLogout       = "user.logout"
	AuditActionLoginFailed  = "user.login_failed"
	AuditActionPasswordReset = "user.password_reset"

	// 用户管理
	AuditActionUserCreate = "user.create"
	AuditActionUserUpdate = "user.update"
	AuditActionUserDelete = "user.delete"

	// 角色权限
	AuditActionRoleAssign   = "role.assign"
	AuditActionRoleRevoke   = "role.revoke"
	AuditActionPermissionGrant = "permission.grant"

	// 数据操作
	AuditActionDataRead   = "data.read"
	AuditActionDataCreate = "data.create"
	AuditActionDataUpdate = "data.update"
	AuditActionDataDelete = "data.delete"
)
