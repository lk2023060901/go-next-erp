package dto

import "github.com/google/uuid"

// ===========================
// 角色管理 DTO
// ===========================

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Code        string                 `json:"code" binding:"required,min=2,max=50"`
	Name        string                 `json:"name" binding:"required,min=2,max=100"`
	TenantID    uuid.UUID              `json:"tenant_id" binding:"required"`
	ParentID    *uuid.UUID             `json:"parent_id,omitempty"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status" binding:"omitempty,oneof=active inactive"`
	IsSystem    bool                   `json:"is_system,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        *string                `json:"name,omitempty"`
	ParentID    *uuid.UUID             `json:"parent_id,omitempty"`
	Description *string                `json:"description,omitempty"`
	Status      *string                `json:"status,omitempty" binding:"omitempty,oneof=active inactive"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          uuid.UUID              `json:"id"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	ParentID    *uuid.UUID             `json:"parent_id,omitempty"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status"`
	IsSystem    bool                   `json:"is_system"`
	Level       int                    `json:"level"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   int64                  `json:"created_at"`
	UpdatedAt   int64                  `json:"updated_at"`
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
	UserID   uuid.UUID   `json:"user_id" binding:"required"`
	RoleIDs  []uuid.UUID `json:"role_ids" binding:"required,min=1"`
}

// RemoveRoleRequest 移除角色请求
type RemoveRoleRequest struct {
	UserID   uuid.UUID   `json:"user_id" binding:"required"`
	RoleIDs  []uuid.UUID `json:"role_ids" binding:"required,min=1"`
}

// ListRolesRequest 角色列表请求
type ListRolesRequest struct {
	TenantID uuid.UUID `form:"tenant_id" binding:"required"`
	Status   string    `form:"status" binding:"omitempty,oneof=active inactive"`
	ParentID *uuid.UUID `form:"parent_id,omitempty"`
	Page     int       `form:"page" binding:"omitempty,min=1"`
	PageSize int       `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	Items      []*RoleResponse `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// UserRolesResponse 用户角色响应
type UserRolesResponse struct {
	UserID uuid.UUID      `json:"user_id"`
	Roles  []*RoleResponse `json:"roles"`
}

// RoleUsersResponse 角色用户响应
type RoleUsersResponse struct {
	RoleID uuid.UUID      `json:"role_id"`
	Users  []*UserResponse `json:"users"`
}

// ===========================
// 权限管理 DTO
// ===========================

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Code        string                 `json:"code" binding:"required,min=2,max=50"`
	Name        string                 `json:"name" binding:"required,min=2,max=100"`
	TenantID    uuid.UUID              `json:"tenant_id" binding:"required"`
	Resource    string                 `json:"resource" binding:"required"`
	Action      string                 `json:"action" binding:"required"`
	Effect      string                 `json:"effect" binding:"required,oneof=allow deny"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status" binding:"omitempty,oneof=active inactive"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	Name        *string                `json:"name,omitempty"`
	Resource    *string                `json:"resource,omitempty"`
	Action      *string                `json:"action,omitempty"`
	Effect      *string                `json:"effect,omitempty" binding:"omitempty,oneof=allow deny"`
	Description *string                `json:"description,omitempty"`
	Status      *string                `json:"status,omitempty" binding:"omitempty,oneof=active inactive"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PermissionResponse 权限响应
type PermissionResponse struct {
	ID          uuid.UUID              `json:"id"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Effect      string                 `json:"effect"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   int64                  `json:"created_at"`
	UpdatedAt   int64                  `json:"updated_at"`
}

// AssignPermissionRequest 分配权限请求
type AssignPermissionRequest struct {
	RoleID        uuid.UUID   `json:"role_id" binding:"required"`
	PermissionIDs []uuid.UUID `json:"permission_ids" binding:"required,min=1"`
}

// RemovePermissionRequest 移除权限请求
type RemovePermissionRequest struct {
	RoleID        uuid.UUID   `json:"role_id" binding:"required"`
	PermissionIDs []uuid.UUID `json:"permission_ids" binding:"required,min=1"`
}

// ListPermissionsRequest 权限列表请求
type ListPermissionsRequest struct {
	TenantID uuid.UUID `form:"tenant_id" binding:"required"`
	Resource string    `form:"resource,omitempty"`
	Action   string    `form:"action,omitempty"`
	Status   string    `form:"status" binding:"omitempty,oneof=active inactive"`
	Page     int       `form:"page" binding:"omitempty,min=1"`
	PageSize int       `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// PermissionListResponse 权限列表响应
type PermissionListResponse struct {
	Items      []*PermissionResponse `json:"items"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// CheckPermissionRequest 检查权限请求
type CheckPermissionRequest struct {
	UserID     uuid.UUID              `json:"user_id" binding:"required"`
	TenantID   uuid.UUID              `json:"tenant_id" binding:"required"`
	Resource   string                 `json:"resource" binding:"required"`
	Action     string                 `json:"action" binding:"required"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// CheckPermissionResponse 检查权限响应
type CheckPermissionResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// UserPermissionsResponse 用户权限响应
type UserPermissionsResponse struct {
	UserID      uuid.UUID            `json:"user_id"`
	Permissions []*PermissionResponse `json:"permissions"`
}

// RolePermissionsResponse 角色权限响应
type RolePermissionsResponse struct {
	RoleID      uuid.UUID            `json:"role_id"`
	Permissions []*PermissionResponse `json:"permissions"`
}
