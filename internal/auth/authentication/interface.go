package authentication

import (
	"context"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/jwt"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
)

// AuthenticationService 认证服务接口
type AuthenticationService interface {
	// Login 用户登录
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)

	// Logout 用户登出
	Logout(ctx context.Context, token string, ipAddress, userAgent string) error

	// ValidateToken 验证令牌
	ValidateToken(ctx context.Context, token string) (*jwt.Claims, error)

	// RefreshToken 刷新令牌
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)

	// Register 用户注册
	Register(ctx context.Context, username, email, password string, tenantID uuid.UUID) (*model.User, error)

	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error

	// GetUserSessions 获取用户会话列表
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*model.Session, error)

	// RevokeSession 撤销会话
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
}

// 确保 Service 实现了 AuthenticationService 接口
var _ AuthenticationService = (*Service)(nil)
