package pkg

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/wire"
	"github.com/lk2023060901/go-next-erp/internal/conf"
	"github.com/lk2023060901/go-next-erp/pkg/cache"
	"github.com/lk2023060901/go-next-erp/pkg/database"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"github.com/lk2023060901/go-next-erp/pkg/storage"
)

// ProviderSet pkg 包的 Wire Provider Set
var ProviderSet = wire.NewSet(
	ProvideDatabase,
	ProvideCache,
	ProvideStorage,
	ProvideLogger,
)

// ProvideDatabase 提供数据库连接
func ProvideDatabase(ctx context.Context, cfg *conf.Config) (*database.DB, func(), error) {
	// 解析 master DSN
	dsn, err := url.Parse(cfg.Database.Master)
	if err != nil {
		return nil, nil, err
	}

	password, _ := dsn.User.Password()
	db, err := database.New(ctx,
		database.WithHost(dsn.Hostname()),
		database.WithPort(func() int {
			if dsn.Port() != "" {
				var port int
				_, _ = fmt.Sscanf(dsn.Port(), "%d", &port)
				return port
			}
			return 5432
		}()),
		database.WithUsername(dsn.User.Username()),
		database.WithPassword(password),
		database.WithDatabase(dsn.Path[1:]), // 去掉开头的 /
		database.WithMaxConns(int32(cfg.Database.MaxOpenConns)),
		database.WithMinConns(int32(cfg.Database.MaxIdleConns)),
	)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup, nil
}

// ProvideCache 提供缓存连接
func ProvideCache(ctx context.Context, cfg *conf.Config) (*cache.Cache, func(), error) {
	// 解析 Redis 地址 (格式: host:port)
	host, port := "localhost", 6379
	if cfg.Redis.Addr != "" {
		// 直接分割 host:port
		parts := strings.Split(cfg.Redis.Addr, ":")
		if len(parts) == 2 {
			host = parts[0]
			fmt.Sscanf(parts[1], "%d", &port)
		} else {
			host = cfg.Redis.Addr
		}
	}

	cache, err := cache.New(ctx,
		cache.WithHost(host),
		cache.WithPort(port),
		cache.WithPassword(cfg.Redis.Password),
		cache.WithDB(cfg.Redis.DB),
		cache.WithPoolSize(10),
	)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		cache.Close()
	}

	return cache, cleanup, nil
}

// ProvideStorage 提供存储连接
func ProvideStorage(ctx context.Context, cfg *conf.Config) (storage.Storage, func(), error) {
	storage, err := storage.New(ctx,
		storage.WithEndpoint(cfg.MinIO.Endpoint),
		storage.WithCredentials(cfg.MinIO.AccessKeyID, cfg.MinIO.SecretAccessKey),
		storage.WithSSL(cfg.MinIO.UseSSL),
		storage.WithBucket(cfg.MinIO.BucketName),
	)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		storage.Close()
	}

	return storage, cleanup, nil
}

// ProvideLogger 提供日志器
func ProvideLogger() (*logger.Logger, func(), error) {
	err := logger.InitGlobal(
		logger.WithLevel("info"),
		logger.WithDevelopmentConfig(),
		logger.WithConsole(true),
	)
	if err != nil {
		return nil, nil, err
	}

	log := logger.GetLogger()

	cleanup := func() {
		// Flush logger
		log.Sync()
	}

	return log, cleanup, nil
}
