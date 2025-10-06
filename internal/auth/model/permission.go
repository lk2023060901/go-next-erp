package model

import (
	"time"

	"github.com/google/uuid"
)

// Permission 权限模型
type Permission struct {
	ID          uuid.UUID `json:"id"`          // UUID v7
	Resource    string    `json:"resource"`    // 资源名称（如：document, user, department）
	Action      string    `json:"action"`      // 操作（create, read, update, delete）
	DisplayName string    `json:"display_name"` // 显示名称
	Description string    `json:"description,omitempty"`
	TenantID    uuid.UUID `json:"tenant_id"`   // 租户 ID

	// 时间戳
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"` // 软删除
}

// RolePermission 角色-权限关联
type RolePermission struct {
	ID           uuid.UUID `json:"id"`
	RoleID       uuid.UUID `json:"role_id"`
	PermissionID uuid.UUID `json:"permission_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// Action 定义标准操作
const (
	ActionCreate = "create" // 创建
	ActionRead   = "read"   // 读取
	ActionUpdate = "update" // 更新
	ActionDelete = "delete" // 删除
	ActionList   = "list"   // 列表（可选）
	ActionExport = "export" // 导出（可选）
	ActionAll    = "*"      // 所有操作
)

// String 返回权限字符串表示（格式：resource:action）
func (p *Permission) String() string {
	return p.Resource + ":" + p.Action
}

// Match 检查权限是否匹配（支持通配符）
func (p *Permission) Match(resource, action string) bool {
	// 完全匹配
	if p.Resource == resource && p.Action == action {
		return true
	}

	// 通配符匹配
	if p.Resource == "*" || p.Action == "*" {
		return true
	}

	// 资源通配符（如：document:*）
	if p.Resource == resource && p.Action == "*" {
		return true
	}

	// 操作通配符（如：*:read）
	if p.Resource == "*" && p.Action == action {
		return true
	}

	return false
}
