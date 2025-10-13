package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
	"github.com/lk2023060901/go-next-erp/internal/file/repository"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"github.com/lk2023060901/go-next-erp/pkg/storage"
	"go.uber.org/zap"
)

// DownloadService 文件下载服务接口
type DownloadService interface {
	// 生成下载URL
	GetDownloadURL(ctx context.Context, fileID uuid.UUID, userID uuid.UUID, tenantID uuid.UUID, expiry time.Duration) (string, error)

	// 生成预览URL
	GetPreviewURL(ctx context.Context, fileID uuid.UUID, userID uuid.UUID, tenantID uuid.UUID, expiry time.Duration) (string, error)

	// 检查访问权限
	CheckAccess(ctx context.Context, fileID uuid.UUID, userID uuid.UUID, tenantID uuid.UUID) (bool, error)

	// 批量获取下载URL
	GetBatchDownloadURLs(ctx context.Context, fileIDs []uuid.UUID, userID uuid.UUID, tenantID uuid.UUID, expiry time.Duration) (map[uuid.UUID]string, error)

	// 记录下载（用于统计）
	RecordDownload(ctx context.Context, req *RecordDownloadRequest) error

	// 获取下载统计
	GetFileDownloadStats(ctx context.Context, fileID uuid.UUID) (*model.FileDownloadSummary, error)
	GetTenantDownloadStats(ctx context.Context, tenantID uuid.UUID, period string) (*model.TenantDownloadSummary, error)
	GetUserDownloadStats(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, period string) (*model.UserDownloadSummary, error)
}

// RecordDownloadRequest 记录下载请求
type RecordDownloadRequest struct {
	FileID          uuid.UUID
	TenantID        uuid.UUID
	DownloadedBy    uuid.UUID
	IPAddress       string
	UserAgent       string
	BytesDownloaded int64
	DownloadTime    time.Duration
	IsComplete      bool
	IsResumed       bool
}

type downloadService struct {
	fileRepo  repository.FileRepository
	statsRepo repository.DownloadStatsRepository
	storage   storage.Storage
	logger    *logger.Logger
}

// NewDownloadService 创建下载服务
func NewDownloadService(
	fileRepo repository.FileRepository,
	statsRepo repository.DownloadStatsRepository,
	storage storage.Storage,
	logger *logger.Logger,
) DownloadService {
	return &downloadService{
		fileRepo:  fileRepo,
		statsRepo: statsRepo,
		storage:   storage,
		logger:    logger.With(zap.String("service", "download")),
	}
}

// GetDownloadURL 获取下载URL
func (s *downloadService) GetDownloadURL(ctx context.Context, fileID uuid.UUID, userID uuid.UUID, tenantID uuid.UUID, expiry time.Duration) (string, error) {
	// 1. 查找文件
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	// 2. 检查访问权限
	if !file.CanAccess(userID, tenantID) {
		return "", fmt.Errorf("access denied: user %s cannot access file %s", userID, fileID)
	}

	// 3. 检查文件状态
	if !file.IsActive() {
		return "", fmt.Errorf("file is not active")
	}

	// 4. 检查病毒扫描结果（如果启用）
	if file.VirusScanned && file.VirusScanResult != nil {
		if *file.VirusScanResult == model.VirusScanInfected {
			return "", fmt.Errorf("file is infected with virus")
		}
	}

	// 5. 生成预签名URL
	objectStorage := s.storage.GetObjectStorage()
	url, err := objectStorage.PresignedGetObject(ctx, file.Bucket, file.StorageKey, expiry, storage.PresignedGetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	s.logger.Info("Download URL generated",
		zap.String("file_id", fileID.String()),
		zap.String("user_id", userID.String()),
		zap.Duration("expiry", expiry))

	return url, nil
}

// GetPreviewURL 获取预览URL
func (s *downloadService) GetPreviewURL(ctx context.Context, fileID uuid.UUID, userID uuid.UUID, tenantID uuid.UUID, expiry time.Duration) (string, error) {
	// 1. 查找文件
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	// 2. 检查访问权限
	if !file.CanAccess(userID, tenantID) {
		return "", fmt.Errorf("access denied")
	}

	// 3. 检查文件状态
	if !file.IsActive() {
		return "", fmt.Errorf("file is not active")
	}

	// 4. 如果有缓存的预览URL且未过期，直接返回
	if file.PreviewURL != nil && file.PreviewExpiresAt != nil {
		if file.PreviewExpiresAt.After(time.Now()) {
			return *file.PreviewURL, nil
		}
	}

	// 5. 生成新的预览URL
	objectStorage := s.storage.GetObjectStorage()
	url, err := objectStorage.PresignedGetObject(ctx, file.Bucket, file.StorageKey, expiry, storage.PresignedGetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to generate preview URL: %w", err)
	}

	// 6. 更新文件记录的预览URL
	expiresAt := time.Now().Add(expiry)
	file.PreviewURL = &url
	file.PreviewExpiresAt = &expiresAt
	if err := s.fileRepo.Update(ctx, file); err != nil {
		s.logger.Warn("Failed to update preview URL in database", zap.Error(err))
		// Non-fatal, continue
	}

	s.logger.Info("Preview URL generated",
		zap.String("file_id", fileID.String()),
		zap.String("user_id", userID.String()))

	return url, nil
}

// CheckAccess 检查访问权限
func (s *downloadService) CheckAccess(ctx context.Context, fileID uuid.UUID, userID uuid.UUID, tenantID uuid.UUID) (bool, error) {
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return false, err
	}

	return file.CanAccess(userID, tenantID), nil
}

// GetBatchDownloadURLs 批量获取下载URL
func (s *downloadService) GetBatchDownloadURLs(ctx context.Context, fileIDs []uuid.UUID, userID uuid.UUID, tenantID uuid.UUID, expiry time.Duration) (map[uuid.UUID]string, error) {
	urls := make(map[uuid.UUID]string)

	for _, fileID := range fileIDs {
		url, err := s.GetDownloadURL(ctx, fileID, userID, tenantID, expiry)
		if err != nil {
			s.logger.Warn("Failed to generate download URL for file",
				zap.String("file_id", fileID.String()),
				zap.Error(err))
			continue
		}
		urls[fileID] = url
	}

	return urls, nil
}

// RecordDownload 记录下载
func (s *downloadService) RecordDownload(ctx context.Context, req *RecordDownloadRequest) error {
	stats := &model.DownloadStats{
		TenantID:        req.TenantID,
		FileID:          req.FileID,
		DownloadedBy:    req.DownloadedBy,
		IPAddress:       req.IPAddress,
		UserAgent:       req.UserAgent,
		BytesDownloaded: req.BytesDownloaded,
		DownloadTime:    req.DownloadTime,
		IsComplete:      req.IsComplete,
		IsResumed:       req.IsResumed,
		DownloadedAt:    time.Now(),
	}

	if err := s.statsRepo.RecordDownload(ctx, stats); err != nil {
		s.logger.Error("Failed to record download",
			zap.String("file_id", req.FileID.String()),
			zap.Error(err))
		// 非致命错误，不影响下载功能
	}

	return nil
}

// GetFileDownloadStats 获取文件下载统计
func (s *downloadService) GetFileDownloadStats(ctx context.Context, fileID uuid.UUID) (*model.FileDownloadSummary, error) {
	return s.statsRepo.GetFileDownloadSummary(ctx, fileID)
}

// GetTenantDownloadStats 获取租户下载统计
func (s *downloadService) GetTenantDownloadStats(ctx context.Context, tenantID uuid.UUID, period string) (*model.TenantDownloadSummary, error) {
	start, end := s.getPeriodRange(period)
	return s.statsRepo.GetTenantDownloadSummary(ctx, tenantID, period, start, end)
}

// GetUserDownloadStats 获取用户下载统计
func (s *downloadService) GetUserDownloadStats(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, period string) (*model.UserDownloadSummary, error) {
	start, end := s.getPeriodRange(period)
	return s.statsRepo.GetUserDownloadSummary(ctx, tenantID, userID, period, start, end)
}

// getPeriodRange 获取统计周期范围
func (s *downloadService) getPeriodRange(period string) (time.Time, time.Time) {
	now := time.Now()
	var start time.Time

	switch period {
	case "day":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = now.AddDate(0, 0, -weekday+1)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	default:
		start = now.AddDate(0, 0, -7) // 默认 7 天
	}

	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	return start, end
}
