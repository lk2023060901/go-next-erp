package dto

import "github.com/google/uuid"

// ===========================
// 用户管理 DTO
// ===========================

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username     string    `json:"username" binding:"required,min=3,max=50"`
	Email        string    `json:"email" binding:"required,email"`
	Password     string    `json:"password" binding:"required,min=8,max=100"`
	TenantID     uuid.UUID `json:"tenant_id" binding:"required"`
	Nickname     string    `json:"nickname,omitempty"`
	Phone        string    `json:"phone,omitempty"`
	Avatar       string    `json:"avatar,omitempty"`
	Status       string    `json:"status" binding:"omitempty,oneof=active inactive locked banned"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname     *string `json:"nickname,omitempty"`
	Phone        *string `json:"phone,omitempty"`
	Avatar       *string `json:"avatar,omitempty"`
	Email        *string `json:"email,omitempty" binding:"omitempty,email"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive locked banned"`
	Reason string `json:"reason,omitempty"`
}

// LockUserRequest 锁定用户请求
type LockUserRequest struct {
	Duration int64  `json:"duration" binding:"required,min=60"` // 秒数，最少1分钟
	Reason   string `json:"reason" binding:"required"`
}

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	TenantID uuid.UUID `form:"tenant_id" binding:"required"`
	Status   string    `form:"status" binding:"omitempty,oneof=active inactive locked banned"`
	Keyword  string    `form:"keyword"` // 搜索关键词（用户名/邮箱/昵称）
	Page     int       `form:"page" binding:"omitempty,min=1"`
	PageSize int       `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID           uuid.UUID              `json:"id"`
	Username     string                 `json:"username"`
	Email        string                 `json:"email"`
	Nickname     string                 `json:"nickname,omitempty"`
	Phone        string                 `json:"phone,omitempty"`
	Avatar       string                 `json:"avatar,omitempty"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	Status       string                 `json:"status"`
	EmailVerified bool                  `json:"email_verified"`
	PhoneVerified bool                  `json:"phone_verified"`
	MFAEnabled   bool                   `json:"mfa_enabled"`
	LastLoginAt  *int64                 `json:"last_login_at,omitempty"`
	LastLoginIP  string                 `json:"last_login_ip,omitempty"`
	LoginAttempts int                   `json:"login_attempts"`
	LockedUntil  *int64                 `json:"locked_until,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    int64                  `json:"created_at"`
	UpdatedAt    int64                  `json:"updated_at"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Items      []*UserResponse `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// ===========================
// 租户管理 DTO
// ===========================

// CreateTenantRequest 创建租户请求
type CreateTenantRequest struct {
	Code       string                 `json:"code" binding:"required,min=2,max=50"`
	Name       string                 `json:"name" binding:"required,min=2,max=100"`
	Domain     string                 `json:"domain,omitempty"`
	Logo       string                 `json:"logo,omitempty"`
	Contact    string                 `json:"contact,omitempty"`
	Email      string                 `json:"email,omitempty" binding:"omitempty,email"`
	Phone      string                 `json:"phone,omitempty"`
	Address    string                 `json:"address,omitempty"`
	Status     string                 `json:"status" binding:"omitempty,oneof=active inactive suspended"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
}

// UpdateTenantRequest 更新租户请求
type UpdateTenantRequest struct {
	Name       *string                `json:"name,omitempty"`
	Domain     *string                `json:"domain,omitempty"`
	Logo       *string                `json:"logo,omitempty"`
	Contact    *string                `json:"contact,omitempty"`
	Email      *string                `json:"email,omitempty" binding:"omitempty,email"`
	Phone      *string                `json:"phone,omitempty"`
	Address    *string                `json:"address,omitempty"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
}

// TenantResponse 租户响应
type TenantResponse struct {
	ID            uuid.UUID              `json:"id"`
	Code          string                 `json:"code"`
	Name          string                 `json:"name"`
	Domain        string                 `json:"domain,omitempty"`
	Logo          string                 `json:"logo,omitempty"`
	Contact       string                 `json:"contact,omitempty"`
	Email         string                 `json:"email,omitempty"`
	Phone         string                 `json:"phone,omitempty"`
	Address       string                 `json:"address,omitempty"`
	Status        string                 `json:"status"`
	UserQuota     int                    `json:"user_quota"`
	UserCount     int                    `json:"user_count"`
	StorageQuota  int64                  `json:"storage_quota"`
	StorageUsed   int64                  `json:"storage_used"`
	Settings      map[string]interface{} `json:"settings,omitempty"`
	CreatedAt     int64                  `json:"created_at"`
	UpdatedAt     int64                  `json:"updated_at"`
}
