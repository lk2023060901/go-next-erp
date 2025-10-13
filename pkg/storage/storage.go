package storage

import (
	"context"
	"fmt"

	"github.com/lk2023060901/go-next-erp/pkg/logger"
)

// Storage MinIO 对象存储客户端封装
// 为了保持向后兼容，保留这个接口
type Storage interface {
	// 获取底层存储客户端
	GetObjectStorage() ObjectStorage

	// 关闭连接
	Close() error
}

// storage 存储实现
type storage struct {
	objectStorage ObjectStorage
	config        *Config
	logger        *logger.Logger
}

// New 创建存储客户端
// 默认使用 MinIO 后端
func New(ctx context.Context, opts ...Option) (Storage, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建 MinIO 存储后端
	log := logger.GetLogger()
	objectStorage, err := NewMinIOStorage(ctx, cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create minio storage: %w", err)
	}

	s := &storage{
		objectStorage: objectStorage,
		config:        cfg,
		logger:        log,
	}

	return s, nil
}

// GetObjectStorage 获取底层对象存储客户端
func (s *storage) GetObjectStorage() ObjectStorage {
	return s.objectStorage
}

// Close 关闭连接
func (s *storage) Close() error {
	return s.objectStorage.Close()
}
