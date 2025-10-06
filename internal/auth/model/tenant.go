package model

import (
	"time"

	"github.com/google/uuid"
)

// Tenant 租户模型
type Tenant struct {
	ID          uuid.UUID    `json:"id"`          // UUID v7
	Name        string       `json:"name"`        // 租户名称
	DisplayName string       `json:"display_name"` // 显示名称
	Domain      string       `json:"domain,omitempty"` // 域名（如：company.example.com）
	Status      TenantStatus `json:"status"`

	// 配置
	MaxUsers    int                    `json:"max_users"`    // 最大用户数
	MaxStorage  int64                  `json:"max_storage"`  // 最大存储（字节）
	Settings    map[string]interface{} `json:"settings,omitempty"` // 租户配置

	// 时间戳
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"` // 软删除
}

// TenantStatus 租户状态
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"    // 活跃
	TenantStatusSuspended TenantStatus = "suspended" // 暂停
	TenantStatusExpired   TenantStatus = "expired"   // 过期
)

// IsActive 检查租户是否活跃
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive && t.DeletedAt == nil
}
