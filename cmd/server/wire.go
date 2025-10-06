//go:build wireinject
// +build wireinject

package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/lk2023060901/go-next-erp/internal/adapter"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/jwt"
	"github.com/lk2023060901/go-next-erp/internal/auth/authorization"
	"github.com/lk2023060901/go-next-erp/internal/auth/repository"
	"github.com/lk2023060901/go-next-erp/internal/conf"
	"github.com/lk2023060901/go-next-erp/internal/server"
	"github.com/lk2023060901/go-next-erp/pkg/cache"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// wireApp 通过 Wire 自动生成依赖注入代码
func wireApp(*conf.Config, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		// 基础设施层
		provideDatabase,
		provideCache,
		provideJWTConfig,

		// 仓储层
		repository.NewUserRepository,
		repository.NewRoleRepository,
		repository.NewPermissionRepository,
		repository.NewPolicyRepository,
		repository.NewSessionRepository,
		repository.NewAuditLogRepository,
		repository.NewRelationRepository,

		// 领域服务层
		authentication.NewService,
		authorization.NewService,

		// 适配器层
		adapter.ProviderSet,

		// 服务器层
		server.ProviderSet,

		// 应用
		newApp,
	))
}

// provideDatabase 提供数据库连接
func provideDatabase(c *conf.Config) (*database.DB, func(), error) {
	ctx := context.Background()

	// 解析连接字符串 (postgres://user:pass@host:port/dbname?sslmode=disable)
	parts := strings.TrimPrefix(c.Database.Master, "postgres://")

	var dbCfg *database.Config
	if userPass, hostInfo, found := strings.Cut(parts, "@"); found {
		user, pass, _ := strings.Cut(userPass, ":")
		hostPort, dbParams, _ := strings.Cut(hostInfo, "/")
		host, portStr, _ := strings.Cut(hostPort, ":")
		dbName, _, _ := strings.Cut(dbParams, "?")

		port := 5432
		fmt.Sscanf(portStr, "%d", &port)

		dbCfg = &database.Config{
			Host:        host,
			Port:        port,
			Database:    dbName,
			Username:    user,
			Password:    pass,
			SSLMode:     "disable",
			MaxConns:    int32(c.Database.MaxOpenConns),
			MinConns:    int32(c.Database.MaxIdleConns),
		}
	} else {
		return nil, nil, fmt.Errorf("invalid database connection string")
	}

	// 使用配置函数创建
	opts := []database.Option{
		func(cfg *database.Config) {
			*cfg = *dbCfg
		},
	}

	db, err := database.New(ctx, opts...)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup, nil
}

// provideCache 提供缓存客户端
func provideCache(c *conf.Config) (*cache.Redis, func(), error) {
	ctx := context.Background()

	// 解析地址
	host, portStr, _ := strings.Cut(c.Redis.Addr, ":")
	port := 6379
	fmt.Sscanf(portStr, "%d", &port)

	cacheCfg := &cache.Config{
		Host:     host,
		Port:     port,
		Password: c.Redis.Password,
		DB:       c.Redis.DB,
	}

	// 使用配置函数创建
	opts := []cache.Option{
		func(cfg *cache.Config) {
			*cfg = *cacheCfg
		},
	}

	redis, err := cache.New(ctx, opts...)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		redis.Close()
	}

	return redis, cleanup, nil
}

// provideJWTConfig 提供 JWT 配置
func provideJWTConfig(c *conf.Config) *jwt.Config {
	return &jwt.Config{
		SecretKey:       c.JWT.Secret,
		AccessTokenTTL:  time.Duration(c.JWT.AccessExpire) * time.Second,
		RefreshTokenTTL: time.Duration(c.JWT.RefreshExpire) * time.Second,
		Issuer:          "go-next-erp",
	}
}
