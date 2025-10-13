package pkg

import (
	"context"

	"github.com/google/wire"
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
func ProvideDatabase(ctx context.Context) (*database.DB, func(), error) {
	db, err := database.New(ctx,
		database.WithHost("localhost"),
		database.WithPort(5432),
		database.WithUsername("postgres"),
		database.WithPassword("password"),
		database.WithDatabase("go_next_erp"),
		database.WithMaxConns(25),
		database.WithMinConns(5),
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
func ProvideCache(ctx context.Context) (*cache.Cache, func(), error) {
	cache, err := cache.New(ctx,
		cache.WithHost("localhost"),
		cache.WithPort(6379),
		cache.WithPassword(""),
		cache.WithDB(0),
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
func ProvideStorage(ctx context.Context) (storage.Storage, func(), error) {
	storage, err := storage.New(ctx,
		storage.WithEndpoint("localhost:9000"),
		storage.WithCredentials("minioadmin", "minioadmin"),
		storage.WithSSL(false),
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
