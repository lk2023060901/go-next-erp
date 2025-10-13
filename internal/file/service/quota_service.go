package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
	"github.com/lk2023060901/go-next-erp/internal/file/repository"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"go.uber.org/zap"
)

// QuotaService 配额服务接口
type QuotaService interface {
	// 获取配额信息
	GetTenantQuota(ctx context.Context, tenantID uuid.UUID) (*model.StorageQuota, error)
	GetUserQuota(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (*model.StorageQuota, error)

	// 更新配额限制
	UpdateQuotaLimit(ctx context.Context, quotaID uuid.UUID, newLimit int64) error

	// 检查配额
	CheckQuota(ctx context.Context, tenantID uuid.UUID, size int64) (bool, error)
	CheckUserQuota(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, size int64) (bool, error)

	// 配额统计
	GetQuotaUsage(ctx context.Context, tenantID uuid.UUID) (*QuotaUsageInfo, error)
	GetQuotaList(ctx context.Context, tenantID uuid.UUID) ([]*QuotaInfo, error)

	// 配额预警
	CheckQuotaWarning(ctx context.Context, tenantID uuid.UUID, threshold float64) (bool, string, error)
}

// QuotaUsageInfo 配额使用信息
type QuotaUsageInfo struct {
	TenantID       uuid.UUID
	QuotaLimit     int64
	QuotaUsed      int64
	QuotaReserved  int64
	QuotaAvailable int64
	UsagePercent   float64
	FileCount      int
	IsNearLimit    bool // 是否接近限制（>80%）
	IsExceeded     bool // 是否超限
}

// QuotaInfo 配额信息
type QuotaInfo struct {
	ID            uuid.UUID
	SubjectType   model.SubjectType
	SubjectID     *uuid.UUID
	QuotaLimit    int64
	QuotaUsed     int64
	UsagePercent  float64
	FileCount     int
	FormattedUsed  string
	FormattedLimit string
}

type quotaService struct {
	fileRepo  repository.FileRepository
	quotaRepo repository.QuotaRepository
	logger    *logger.Logger

	// Default limits
	defaultTenantQuota int64
	defaultUserQuota   int64
}

// QuotaServiceConfig 配额服务配置
type QuotaServiceConfig struct {
	DefaultTenantQuota int64 // Default 10GB
	DefaultUserQuota   int64 // Default 1GB
}

// NewQuotaService 创建配额服务
func NewQuotaService(
	fileRepo repository.FileRepository,
	quotaRepo repository.QuotaRepository,
	logger *logger.Logger,
	config *QuotaServiceConfig,
) QuotaService {
	if config == nil {
		config = &QuotaServiceConfig{
			DefaultTenantQuota: 10 * 1024 * 1024 * 1024, // 10GB
			DefaultUserQuota:   1 * 1024 * 1024 * 1024,  // 1GB
		}
	}

	return &quotaService{
		fileRepo:           fileRepo,
		quotaRepo:          quotaRepo,
		logger:             logger.With(zap.String("service", "quota")),
		defaultTenantQuota: config.DefaultTenantQuota,
		defaultUserQuota:   config.DefaultUserQuota,
	}
}

// GetTenantQuota 获取租户配额
func (s *quotaService) GetTenantQuota(ctx context.Context, tenantID uuid.UUID) (*model.StorageQuota, error) {
	quota, err := s.quotaRepo.GetOrCreateTenantQuota(ctx, tenantID, s.defaultTenantQuota)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant quota: %w", err)
	}

	// Sync with actual usage
	actualUsed, err := s.fileRepo.GetTotalSize(ctx, tenantID)
	if err != nil {
		s.logger.Warn("Failed to get actual file size", zap.Error(err))
	} else if actualUsed != quota.QuotaUsed {
		s.logger.Info("Syncing quota usage",
			zap.Int64("recorded", quota.QuotaUsed),
			zap.Int64("actual", actualUsed))
		quota.QuotaUsed = actualUsed
		s.quotaRepo.Update(ctx, quota)
	}

	return quota, nil
}

// GetUserQuota 获取用户配额
func (s *quotaService) GetUserQuota(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (*model.StorageQuota, error) {
	return s.quotaRepo.GetOrCreateUserQuota(ctx, tenantID, userID, s.defaultUserQuota)
}

// UpdateQuotaLimit 更新配额限制
func (s *quotaService) UpdateQuotaLimit(ctx context.Context, quotaID uuid.UUID, newLimit int64) error {
	quota, err := s.quotaRepo.FindByID(ctx, quotaID)
	if err != nil {
		return fmt.Errorf("quota not found: %w", err)
	}

	quota.QuotaLimit = newLimit
	if err := s.quotaRepo.Update(ctx, quota); err != nil {
		return fmt.Errorf("failed to update quota limit: %w", err)
	}

	s.logger.Info("Quota limit updated",
		zap.String("quota_id", quotaID.String()),
		zap.Int64("new_limit", newLimit))

	return nil
}

// CheckQuota 检查租户配额
func (s *quotaService) CheckQuota(ctx context.Context, tenantID uuid.UUID, size int64) (bool, error) {
	quota, err := s.GetTenantQuota(ctx, tenantID)
	if err != nil {
		return false, err
	}

	return quota.CanAllocate(size), nil
}

// CheckUserQuota 检查用户配额
func (s *quotaService) CheckUserQuota(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, size int64) (bool, error) {
	// Check tenant quota first
	canAllocate, err := s.CheckQuota(ctx, tenantID, size)
	if err != nil || !canAllocate {
		return false, err
	}

	// Then check user quota
	userQuota, err := s.GetUserQuota(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	return userQuota.CanAllocate(size), nil
}

// GetQuotaUsage 获取配额使用信息
func (s *quotaService) GetQuotaUsage(ctx context.Context, tenantID uuid.UUID) (*QuotaUsageInfo, error) {
	quota, err := s.GetTenantQuota(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	fileCount, err := s.fileRepo.GetFileCount(ctx, tenantID)
	if err != nil {
		s.logger.Warn("Failed to get file count", zap.Error(err))
		fileCount = int64(quota.FileCountUsed)
	}

	usage := &QuotaUsageInfo{
		TenantID:       tenantID,
		QuotaLimit:     quota.QuotaLimit,
		QuotaUsed:      quota.QuotaUsed,
		QuotaReserved:  quota.QuotaReserved,
		QuotaAvailable: quota.GetAvailableQuota(),
		UsagePercent:   quota.GetUsagePercentage(),
		FileCount:      int(fileCount),
		IsNearLimit:    quota.IsNearLimit(80.0),
		IsExceeded:     quota.IsQuotaExceeded(),
	}

	return usage, nil
}

// GetQuotaList 获取配额列表
func (s *quotaService) GetQuotaList(ctx context.Context, tenantID uuid.UUID) ([]*QuotaInfo, error) {
	quotas, err := s.quotaRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list quotas: %w", err)
	}

	quotaInfos := make([]*QuotaInfo, len(quotas))
	for i, q := range quotas {
		quotaInfos[i] = &QuotaInfo{
			ID:             q.ID,
			SubjectType:    q.SubjectType,
			SubjectID:      q.SubjectID,
			QuotaLimit:     q.QuotaLimit,
			QuotaUsed:      q.QuotaUsed,
			UsagePercent:   q.GetUsagePercentage(),
			FileCount:      q.FileCountUsed,
			FormattedUsed:  formatBytes(q.QuotaUsed),
			FormattedLimit: formatBytes(q.QuotaLimit),
		}
	}

	return quotaInfos, nil
}

// CheckQuotaWarning 检查配额预警
func (s *quotaService) CheckQuotaWarning(ctx context.Context, tenantID uuid.UUID, threshold float64) (bool, string, error) {
	quota, err := s.GetTenantQuota(ctx, tenantID)
	if err != nil {
		return false, "", err
	}

	usagePercent := quota.GetUsagePercentage()
	if usagePercent >= threshold {
		message := fmt.Sprintf("Storage quota warning: %.1f%% used (%s / %s)",
			usagePercent,
			formatBytes(quota.QuotaUsed),
			formatBytes(quota.QuotaLimit))

		s.logger.Warn("Quota warning triggered",
			zap.String("tenant_id", tenantID.String()),
			zap.Float64("usage_percent", usagePercent),
			zap.Float64("threshold", threshold))

		return true, message, nil
	}

	return false, "", nil
}
