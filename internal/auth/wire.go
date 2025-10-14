package auth

import (
	"time"

	"github.com/google/wire"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/jwt"
	"github.com/lk2023060901/go-next-erp/internal/auth/authorization"
	"github.com/lk2023060901/go-next-erp/internal/auth/repository"
	"github.com/lk2023060901/go-next-erp/internal/conf"
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

	// JWT
	ProvideJWTConfig,
	ProvideJWTManager,

	// Services
	authentication.NewService,
	authorization.NewService,
)

// ProvideJWTConfig 提供 JWT 配置
func ProvideJWTConfig(cfg *conf.Config) *jwt.Config {
	return &jwt.Config{
		SecretKey:       cfg.JWT.Secret,
		AccessTokenTTL:  time.Duration(cfg.JWT.AccessExpire) * time.Second,
		RefreshTokenTTL: time.Duration(cfg.JWT.RefreshExpire) * time.Second,
		Issuer:          "go-next-erp",
	}
}

// ProvideJWTManager 提供 JWT 管理器
func ProvideJWTManager(config *jwt.Config) *jwt.Manager {
	return jwt.NewManager(config)
}
