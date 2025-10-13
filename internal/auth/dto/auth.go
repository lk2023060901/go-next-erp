package dto

import "github.com/google/uuid"

// ===========================
// 认证相关 DTO
// ===========================

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username  string    `json:"username" binding:"required,min=3,max=50"`
	Email     string    `json:"email" binding:"required,email"`
	Password  string    `json:"password" binding:"required,min=8,max=100"`
	TenantID  uuid.UUID `json:"tenant_id" binding:"required"`
	Nickname  string    `json:"nickname,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Avatar    string    `json:"avatar,omitempty"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	TenantID  uuid.UUID `json:"tenant_id" binding:"required"`
	IPAddress string `json:"-"` // 从请求中提取
	UserAgent string `json:"-"` // 从请求中提取
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=100"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	TenantID    uuid.UUID `json:"tenant_id" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=100"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	ExpiresAt    int64      `json:"expires_at"`
	TokenType    string     `json:"token_type"`
	User         *UserInfo  `json:"user"`
}

// RefreshTokenResponse 刷新令牌响应
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	TokenType    string `json:"token_type"`
}

// UserInfo 用户基本信息
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Nickname  string    `json:"nickname,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Avatar    string    `json:"avatar,omitempty"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Status    string    `json:"status"`
	Roles     []string  `json:"roles,omitempty"`
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	Token     string `json:"-"` // 从Header提取
	IPAddress string `json:"-"` // 从请求中提取
	UserAgent string `json:"-"` // 从请求中提取
}

// ===========================
// 会话相关 DTO
// ===========================

// SessionInfo 会话信息
type SessionInfo struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	LastActive int64     `json:"last_active"`
	CreatedAt  int64     `json:"created_at"`
	ExpiresAt  int64     `json:"expires_at"`
}

// RevokeSessionRequest 撤销会话请求
type RevokeSessionRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}
