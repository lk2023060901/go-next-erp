package model

import (
	"time"

	"github.com/google/uuid"
)

// Policy ABAC 策略模型
type Policy struct {
	ID          uuid.UUID    `json:"id"`          // UUID v7
	Name        string       `json:"name"`        // 策略名称
	Description string       `json:"description,omitempty"`
	TenantID    uuid.UUID    `json:"tenant_id"`   // 租户 ID
	Resource    string       `json:"resource"`    // 资源类型
	Action      string       `json:"action"`      // 操作
	Expression  string       `json:"expression"`  // 策略表达式
	Effect      PolicyEffect `json:"effect"`      // 效果（Allow/Deny）
	Priority    int          `json:"priority"`    // 优先级（数字越大优先级越高）
	Enabled     bool         `json:"enabled"`     // 是否启用

	// 时间戳
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"` // 软删除
}

// PolicyEffect 策略效果
type PolicyEffect string

const (
	PolicyEffectAllow PolicyEffect = "allow" // 允许
	PolicyEffectDeny  PolicyEffect = "deny"  // 拒绝
)

// 策略表达式示例：
// user.department_id == resource.department_id && user.level >= 3
// user.roles contains "manager" && time.hour >= 9 && time.hour <= 18
// resource.status == "published" || user.id == resource.owner_id
