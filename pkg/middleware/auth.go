package middleware

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/jwt"
)

var (
	ErrUnauthorized     = errors.Unauthorized("UNAUTHORIZED", "未授权访问")
	ErrInvalidToken     = errors.Unauthorized("INVALID_TOKEN", "无效的令牌")
	ErrMissingToken     = errors.Unauthorized("MISSING_TOKEN", "缺少令牌")
)

// Auth JWT 认证中间件
func Auth(jwtManager *jwt.Manager) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 从传输层获取 Header
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, ErrUnauthorized
			}

			// 获取 Authorization Header
			authHeader := tr.RequestHeader().Get("Authorization")
			if authHeader == "" {
				return nil, ErrMissingToken
			}

			// 提取 Bearer Token
			token := extractToken(authHeader)
			if token == "" {
				return nil, ErrInvalidToken
			}

			// 验证 Token
			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				return nil, ErrInvalidToken
			}

			// 将用户信息注入上下文
			ctx = context.WithValue(ctx, "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "tenant_id", claims.TenantID)
			ctx = context.WithValue(ctx, "username", claims.Username)
			ctx = context.WithValue(ctx, "email", claims.Email)

			return handler(ctx, req)
		}
	}
}

// extractToken 从 Authorization Header 提取 Token
func extractToken(authHeader string) string {
	// 格式：Bearer <token>
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

// GetUserID 从上下文获取用户 ID
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	return userID, ok
}

// GetTenantID 从上下文获取租户 ID
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	return tenantID, ok
}

// GetUsername 从上下文获取用户名
func GetUsername(ctx context.Context) (string, bool) {
	username, ok := ctx.Value("username").(string)
	return username, ok
}
