package service

import (
"context"
"fmt"
"time"

"github.com/lk2023060901/go-next-erp/internal/file/repository"
"github.com/lk2023060901/go-next-erp/pkg/logger"
"github.com/lk2023060901/go-next-erp/pkg/storage"
"go.uber.org/zap"
)

// CleanupService 清理服务接口
type CleanupService interface {
	// 清理过期的临时文件
	CleanExpiredFiles(ctx context.Context) (int64, error)
	
	// 清理过期的分片上传
	CleanExpiredUploads(ctx context.Context) (int64, error)
	
	// 清理旧的访问日志
	CleanOldAccessLogs(ctx context.Context, retentionDays int) (int64, error)
	
	// 清理已删除文件的物理存储
	CleanDeletedFiles(ctx context.Context) (int64, error)
}

// CleanupServiceConfig 清理服务配置
type CleanupServiceConfig struct {
	TempFileRetentionDays   int // 临时文件保留天数，默认 7 天
	UploadRetentionDays     int // 未完成上传保留天数，默认 1 天
	AccessLogRetentionDays  int // 访问日志保留天数，默认 90 天
	DeletedFileGracePeriod  int // 删除文件宽限期（天），默认 30 天
}

type cleanupService struct {
	fileRepo          repository.FileRepository
	multipartRepo     repository.MultipartUploadRepository
	accessLogRepo     repository.FileAccessLogRepository
	storageClient     storage.Storage
	logger            *logger.Logger
	config            *CleanupServiceConfig
}

// NewCleanupService 创建清理服务
func NewCleanupService(
fileRepo repository.FileRepository,
multipartRepo repository.MultipartUploadRepository,
accessLogRepo repository.FileAccessLogRepository,
storageClient storage.Storage,
logger *logger.Logger,
) CleanupService {
	config := &CleanupServiceConfig{
		TempFileRetentionDays:  7,
		UploadRetentionDays:    1,
		AccessLogRetentionDays: 90,
		DeletedFileGracePeriod: 30,
	}

	return &cleanupService{
		fileRepo:      fileRepo,
		multipartRepo: multipartRepo,
		accessLogRepo: accessLogRepo,
		storageClient: storageClient,
		logger:        logger.With(zap.String("service", "cleanup")),
		config:        config,
	}
}

// CleanExpiredFiles 清理过期的临时文件
func (s *cleanupService) CleanExpiredFiles(ctx context.Context) (int64, error) {
	now := time.Now()
	
	// 标记过期文件
	count, err := s.fileRepo.MarkAsExpired(ctx, now)
	if err != nil {
		return 0, fmt.Errorf("failed to mark expired files: %w", err)
	}

	s.logger.Info("Marked expired files",
zap.Int64("count", count),
zap.Time("before", now))

	// 清理超过保留期的临时文件
	retentionPeriod := time.Duration(s.config.TempFileRetentionDays) * 24 * time.Hour
	deleteBefore := now.Add(-retentionPeriod)
	
	deletedCount, err := s.fileRepo.CleanTemporaryFiles(ctx, deleteBefore)
	if err != nil {
		return count, fmt.Errorf("failed to clean temporary files: %w", err)
	}

	s.logger.Info("Cleaned temporary files",
zap.Int64("count", deletedCount),
zap.Time("before", deleteBefore))

	return count + deletedCount, nil
}

// CleanExpiredUploads 清理过期的分片上传
func (s *cleanupService) CleanExpiredUploads(ctx context.Context) (int64, error) {
	now := time.Now()
	
	count, err := s.multipartRepo.CleanExpiredUploads(ctx, now)
	if err != nil {
		return 0, fmt.Errorf("failed to clean expired uploads: %w", err)
	}

	s.logger.Info("Cleaned expired uploads",
zap.Int64("count", count),
zap.Time("before", now))

	return count, nil
}

// CleanOldAccessLogs 清理旧的访问日志
func (s *cleanupService) CleanOldAccessLogs(ctx context.Context, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = s.config.AccessLogRetentionDays
	}

	retentionPeriod := time.Duration(retentionDays) * 24 * time.Hour
	deleteBefore := time.Now().Add(-retentionPeriod)
	
	count, err := s.accessLogRepo.DeleteOldLogs(ctx, deleteBefore)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old logs: %w", err)
	}

	s.logger.Info("Cleaned old access logs",
zap.Int64("count", count),
zap.Int("retention_days", retentionDays),
zap.Time("before", deleteBefore))

	return count, nil
}

// CleanDeletedFiles 清理已删除文件的物理存储
func (s *cleanupService) CleanDeletedFiles(ctx context.Context) (int64, error) {
	// 此功能需要额外实现：
	// 1. 查找超过宽限期的软删除文件
	// 2. 从对象存储中删除物理文件
	// 3. 从数据库中硬删除记录
	
	s.logger.Info("Deleted files cleanup not fully implemented yet")
	
	// TODO: 实现物理删除逻辑
	return 0, nil
}
