package model

import (
	"time"

	"github.com/google/uuid"
)

// User 用户模型
type User struct {
	ID           uuid.UUID  `json:"id"`            // UUID v7
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`             // 不暴露到 JSON
	TenantID     uuid.UUID  `json:"tenant_id"`     // 租户 ID
	Status       UserStatus `json:"status"`

	// MFA 多因素认证
	MFAEnabled bool   `json:"mfa_enabled"`
	MFASecret  string `json:"-"` // TOTP 密钥

	// 安全字段
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP   string     `json:"last_login_ip,omitempty"`   // 上次登录 IP
	LoginAttempts int        `json:"-"`
	LockedUntil   *time.Time `json:"-"`

	// 元数据（可存储 employee_id 等扩展字段）
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// 时间戳
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"` // 软删除
}

// UserStatus 用户状态
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"   // 活跃
	UserStatusInactive UserStatus = "inactive" // 未激活
	UserStatusLocked   UserStatus = "locked"   // 锁定
	UserStatusBanned   UserStatus = "banned"   // 封禁
)

// IsLocked 检查用户是否被锁定
func (u *User) IsLocked() bool {
	if u.Status == UserStatusLocked {
		return true
	}

	if u.LockedUntil != nil && u.LockedUntil.After(time.Now()) {
		return true
	}

	return false
}

// IsActive 检查用户是否活跃
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && !u.IsLocked()
}

// CanLogin 检查用户是否可以登录
func (u *User) CanLogin() bool {
	return u.IsActive() && u.DeletedAt == nil
}
