package model

import (
	"time"

	"github.com/google/uuid"
)

// Role 角色模型
type Role struct {
	ID          uuid.UUID  `json:"id"`          // UUID v7
	Name        string     `json:"name"`        // 角色名称
	DisplayName string     `json:"display_name"` // 显示名称
	Description string     `json:"description,omitempty"`
	TenantID    uuid.UUID  `json:"tenant_id"`   // 租户 ID
	ParentID    *uuid.UUID `json:"parent_id,omitempty"` // 父角色 ID（支持角色继承）

	// 时间戳
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"` // 软删除
}

// UserRole 用户-角色关联
type UserRole struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	RoleID    uuid.UUID `json:"role_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
}
