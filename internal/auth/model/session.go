package model

import (
	"time"

	"github.com/google/uuid"
)

// Session 会话模型
type Session struct {
	ID        uuid.UUID `json:"id"`         // UUID v7
	UserID    uuid.UUID `json:"user_id"`    // 用户 ID
	TenantID  uuid.UUID `json:"tenant_id"`  // 租户 ID
	Token     string    `json:"token"`      // 会话令牌（JWT 或 UUID）
	IPAddress string    `json:"ip_address"` // 客户端 IP
	UserAgent string    `json:"user_agent"` // 用户代理（浏览器/设备信息）

	// 时间控制
	ExpiresAt time.Time  `json:"expires_at"` // 过期时间
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"` // 撤销时间
}

// IsValid 检查会话是否有效
func (s *Session) IsValid() bool {
	now := time.Now()

	// 检查是否已撤销
	if s.RevokedAt != nil {
		return false
	}

	// 检查是否过期
	if s.ExpiresAt.Before(now) {
		return false
	}

	return true
}

// Revoke 撤销会话
func (s *Session) Revoke() {
	now := time.Now()
	s.RevokedAt = &now
}
