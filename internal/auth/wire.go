package auth

import (
	"time"

	"github.com/google/wire"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/jwt"
	"github.com/lk2023060901/go-next-erp/internal/auth/authorization"
	"github.com/lk2023060901/go-next-erp/internal/auth/repository"
)

// ProviderSet auth 模块的 Wire Provider Set
var ProviderSet = wire.NewSet(
	// Repositories
	repository.NewUserRepository,
	repository.NewRoleRepository,
	repository.NewPermissionRepository,
	repository.NewSessionRepository,
	repository.NewAuditLogRepository,
	repository.NewPolicyRepository,
	repository.NewRelationRepository,
	repository.NewTenantRepository,

	// Services
	ProvideJWTConfig,
	authentication.NewService,
	authorization.NewService,
)

// ProvideJWTConfig 提供 JWT 配置
func ProvideJWTConfig() *jwt.Config {
	return &jwt.Config{
		SecretKey:       "your-secret-key-change-in-production",
		AccessTokenTTL:  3600 * time.Second,  // 1 hour
		RefreshTokenTTL: 86400 * time.Second, // 24 hours
		Issuer:          "go-next-erp",
	}
}
