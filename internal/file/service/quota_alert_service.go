package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
	"github.com/lk2023060901/go-next-erp/internal/file/repository"
	"github.com/lk2023060901/go-next-erp/pkg/cache"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"go.uber.org/zap"
)

// QuotaAlertService 配额预警服务接口
type QuotaAlertService interface {
	// 检查所有租户的配额预警
	CheckAllTenantsQuota(ctx context.Context) (*QuotaAlertSummary, error)

	// 检查单个租户的配额预警
	CheckTenantQuota(ctx context.Context, tenantID uuid.UUID) (*QuotaAlertResult, error)

	// 获取预警配置
	GetAlertConfig() *QuotaAlertConfig

	// 更新预警配置
	UpdateAlertConfig(config *QuotaAlertConfig) error
}

// QuotaAlertLevel 预警级别
type QuotaAlertLevel string

const (
	AlertLevelNone     QuotaAlertLevel = "none"     // 无预警
	AlertLevelWarning  QuotaAlertLevel = "warning"  // 警告（80%）
	AlertLevelCritical QuotaAlertLevel = "critical" // 严重（90%）
	AlertLevelUrgent   QuotaAlertLevel = "urgent"   // 紧急（95%）
)

// QuotaAlertConfig 配额预警配置
type QuotaAlertConfig struct {
	// 预警阈值
	WarningThreshold  float64 // 默认 80%
	CriticalThreshold float64 // 默认 90%
	UrgentThreshold   float64 // 默认 95%

	// 通知频率（防止重复通知）
	WarningInterval  time.Duration // 默认 24 小时
	CriticalInterval time.Duration // 默认 6 小时
	UrgentInterval   time.Duration // 默认 1 小时

	// 是否启用各级别预警
	EnableWarning  bool // 默认 true
	EnableCritical bool // 默认 true
	EnableUrgent   bool // 默认 true

	// 通知方式
	NotifyBySystem bool // 系统通知，默认 true
	NotifyByEmail  bool // 邮件通知，默认 false（需要邮件服务）
	NotifyBySMS    bool // 短信通知，默认 false（需要短信服务）
}

// QuotaAlertResult 单个租户的预警结果
type QuotaAlertResult struct {
	TenantID       uuid.UUID       `json:"tenant_id"`
	AlertLevel     QuotaAlertLevel `json:"alert_level"`
	UsagePercent   float64         `json:"usage_percent"`
	QuotaUsed      int64           `json:"quota_used"`
	QuotaLimit     int64           `json:"quota_limit"`
	QuotaAvailable int64           `json:"quota_available"`
	Message        string          `json:"message"`
	ShouldNotify   bool            `json:"should_notify"` // 是否需要发送通知
	LastNotified   *time.Time      `json:"last_notified"` // 上次通知时间
}

// QuotaAlertSummary 预警汇总
type QuotaAlertSummary struct {
	TotalTenants  int                 `json:"total_tenants"`
	WarningCount  int                 `json:"warning_count"`
	CriticalCount int                 `json:"critical_count"`
	UrgentCount   int                 `json:"urgent_count"`
	NotifiedCount int                 `json:"notified_count"`
	AlertResults  []*QuotaAlertResult `json:"alert_results"`
	CheckedAt     time.Time           `json:"checked_at"`
}

type quotaAlertService struct {
	quotaRepo repository.QuotaRepository
	cache     *cache.Cache
	logger    *logger.Logger
	config    *QuotaAlertConfig

	// 通知服务（可选，暂时使用日志记录）
	// notificationService notification.NotificationService
}

// DefaultQuotaAlertConfig 返回默认配额预警配置
func DefaultQuotaAlertConfig() *QuotaAlertConfig {
	return &QuotaAlertConfig{
		WarningThreshold:  80.0,
		CriticalThreshold: 90.0,
		UrgentThreshold:   95.0,
		WarningInterval:   24 * time.Hour,
		CriticalInterval:  6 * time.Hour,
		UrgentInterval:    1 * time.Hour,
		EnableWarning:     true,
		EnableCritical:    true,
		EnableUrgent:      true,
		NotifyBySystem:    true,
		NotifyByEmail:     false,
		NotifyBySMS:       false,
	}
}

// NewQuotaAlertService 创建配额预警服务
func NewQuotaAlertService(
	quotaRepo repository.QuotaRepository,
	cache *cache.Cache,
	logger *logger.Logger,
	config *QuotaAlertConfig,
) QuotaAlertService {
	if config == nil {
		config = DefaultQuotaAlertConfig()
	}

	return &quotaAlertService{
		quotaRepo: quotaRepo,
		cache:     cache,
		logger:    logger.With(zap.String("service", "quota_alert")),
		config:    config,
	}
}

// CheckAllTenantsQuota 检查所有租户的配额预警
func (s *quotaAlertService) CheckAllTenantsQuota(ctx context.Context) (*QuotaAlertSummary, error) {
	s.logger.Info("Starting quota alert check for all tenants")

	// 获取所有租户的配额列表
	// 注意：这里简化处理，实际中应该从租户服务获取租户列表
	// 目前只检查 tenant 类型的配额
	quotas, err := s.quotaRepo.ListByTenant(ctx, uuid.Nil) // 使用 Nil UUID 表示获取所有
	if err != nil {
		return nil, fmt.Errorf("failed to list quotas: %w", err)
	}

	summary := &QuotaAlertSummary{
		TotalTenants:  0,
		WarningCount:  0,
		CriticalCount: 0,
		UrgentCount:   0,
		NotifiedCount: 0,
		AlertResults:  make([]*QuotaAlertResult, 0),
		CheckedAt:     time.Now(),
	}

	// 只检查租户级别的配额
	for _, quota := range quotas {
		if quota.SubjectType != model.SubjectTypeTenant {
			continue
		}

		summary.TotalTenants++

		result := s.checkQuotaAlert(ctx, quota)

		// 统计预警级别
		switch result.AlertLevel {
		case AlertLevelWarning:
			summary.WarningCount++
		case AlertLevelCritical:
			summary.CriticalCount++
		case AlertLevelUrgent:
			summary.UrgentCount++
		}

		// 如果需要通知，发送通知
		if result.ShouldNotify {
			if err := s.sendNotification(ctx, result); err != nil {
				s.logger.Error("Failed to send notification",
					zap.String("tenant_id", result.TenantID.String()),
					zap.Error(err))
			} else {
				summary.NotifiedCount++
			}
		}

		// 只记录有预警的结果
		if result.AlertLevel != AlertLevelNone {
			summary.AlertResults = append(summary.AlertResults, result)
		}
	}

	s.logger.Info("Quota alert check completed",
		zap.Int("total_tenants", summary.TotalTenants),
		zap.Int("warning_count", summary.WarningCount),
		zap.Int("critical_count", summary.CriticalCount),
		zap.Int("urgent_count", summary.UrgentCount),
		zap.Int("notified_count", summary.NotifiedCount))

	return summary, nil
}

// CheckTenantQuota 检查单个租户的配额预警
func (s *quotaAlertService) CheckTenantQuota(ctx context.Context, tenantID uuid.UUID) (*QuotaAlertResult, error) {
	quota, err := s.quotaRepo.GetOrCreateTenantQuota(ctx, tenantID, 10*1024*1024*1024) // 10GB default
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant quota: %w", err)
	}

	result := s.checkQuotaAlert(ctx, quota)

	// 如果需要通知，发送通知
	if result.ShouldNotify {
		if err := s.sendNotification(ctx, result); err != nil {
			s.logger.Error("Failed to send notification",
				zap.String("tenant_id", result.TenantID.String()),
				zap.Error(err))
		}
	}

	return result, nil
}

// checkQuotaAlert 检查单个配额的预警状态
func (s *quotaAlertService) checkQuotaAlert(ctx context.Context, quota *model.StorageQuota) *QuotaAlertResult {
	usagePercent := quota.GetUsagePercentage()

	result := &QuotaAlertResult{
		TenantID:       quota.TenantID,
		AlertLevel:     AlertLevelNone,
		UsagePercent:   usagePercent,
		QuotaUsed:      quota.QuotaUsed,
		QuotaLimit:     quota.QuotaLimit,
		QuotaAvailable: quota.GetAvailableQuota(),
		ShouldNotify:   false,
	}

	// 确定预警级别
	if s.config.EnableUrgent && usagePercent >= s.config.UrgentThreshold {
		result.AlertLevel = AlertLevelUrgent
		result.Message = fmt.Sprintf("紧急预警：存储配额已使用 %.1f%%（%s / %s），剩余 %s",
			usagePercent,
			formatBytes(quota.QuotaUsed),
			formatBytes(quota.QuotaLimit),
			formatBytes(quota.GetAvailableQuota()))
	} else if s.config.EnableCritical && usagePercent >= s.config.CriticalThreshold {
		result.AlertLevel = AlertLevelCritical
		result.Message = fmt.Sprintf("严重预警：存储配额已使用 %.1f%%（%s / %s），剩余 %s",
			usagePercent,
			formatBytes(quota.QuotaUsed),
			formatBytes(quota.QuotaLimit),
			formatBytes(quota.GetAvailableQuota()))
	} else if s.config.EnableWarning && usagePercent >= s.config.WarningThreshold {
		result.AlertLevel = AlertLevelWarning
		result.Message = fmt.Sprintf("警告：存储配额已使用 %.1f%%（%s / %s），剩余 %s",
			usagePercent,
			formatBytes(quota.QuotaUsed),
			formatBytes(quota.QuotaLimit),
			formatBytes(quota.GetAvailableQuota()))
	}

	// 检查是否需要通知（防止重复通知）
	if result.AlertLevel != AlertLevelNone {
		shouldNotify, lastNotified := s.shouldSendNotification(ctx, quota.TenantID, result.AlertLevel)
		result.ShouldNotify = shouldNotify
		result.LastNotified = lastNotified
	}

	return result
}

// shouldSendNotification 判断是否应该发送通知
func (s *quotaAlertService) shouldSendNotification(ctx context.Context, tenantID uuid.UUID, level QuotaAlertLevel) (bool, *time.Time) {
	cacheKey := fmt.Sprintf("quota_alert:%s:%s", tenantID.String(), level)

	// 从缓存中获取上次通知时间
	var lastNotified time.Time
	if s.cache != nil {
		if err := s.cache.Get(ctx, cacheKey, &lastNotified); err == nil {
			// 计算间隔
			var interval time.Duration
			switch level {
			case AlertLevelWarning:
				interval = s.config.WarningInterval
			case AlertLevelCritical:
				interval = s.config.CriticalInterval
			case AlertLevelUrgent:
				interval = s.config.UrgentInterval
			default:
				return false, &lastNotified
			}

			// 如果未超过通知间隔，不发送通知
			if time.Since(lastNotified) < interval {
				return false, &lastNotified
			}
		}
	}

	return true, nil
}

// sendNotification 发送预警通知
func (s *quotaAlertService) sendNotification(ctx context.Context, result *QuotaAlertResult) error {
	s.logger.Warn("Quota alert notification",
		zap.String("tenant_id", result.TenantID.String()),
		zap.String("level", string(result.AlertLevel)),
		zap.Float64("usage_percent", result.UsagePercent),
		zap.String("message", result.Message))

	// TODO: 集成通知服务
	// 1. 系统通知
	if s.config.NotifyBySystem {
		// notificationService.SendSystemNotification(...)
		s.logger.Info("System notification sent (simulated)",
			zap.String("tenant_id", result.TenantID.String()),
			zap.String("level", string(result.AlertLevel)))
	}

	// 2. 邮件通知
	if s.config.NotifyByEmail && result.AlertLevel >= AlertLevelCritical {
		// notificationService.SendEmail(...)
		s.logger.Info("Email notification sent (simulated)",
			zap.String("tenant_id", result.TenantID.String()),
			zap.String("level", string(result.AlertLevel)))
	}

	// 3. 短信通知
	if s.config.NotifyBySMS && result.AlertLevel == AlertLevelUrgent {
		// notificationService.SendSMS(...)
		s.logger.Info("SMS notification sent (simulated)",
			zap.String("tenant_id", result.TenantID.String()),
			zap.String("level", string(result.AlertLevel)))
	}

	// 更新缓存，记录通知时间
	if s.cache != nil {
		cacheKey := fmt.Sprintf("quota_alert:%s:%s", result.TenantID.String(), result.AlertLevel)
		now := time.Now()

		// 缓存时间设置为通知间隔的 2 倍，确保不会过早清除
		var ttl time.Duration
		switch result.AlertLevel {
		case AlertLevelWarning:
			ttl = s.config.WarningInterval * 2
		case AlertLevelCritical:
			ttl = s.config.CriticalInterval * 2
		case AlertLevelUrgent:
			ttl = s.config.UrgentInterval * 2
		default:
			ttl = 24 * time.Hour
		}

		if err := s.cache.Set(ctx, cacheKey, now, int(ttl.Seconds())); err != nil {
			s.logger.Error("Failed to cache notification time",
				zap.String("cache_key", cacheKey),
				zap.Error(err))
		}
	}

	return nil
}

// GetAlertConfig 获取预警配置
func (s *quotaAlertService) GetAlertConfig() *QuotaAlertConfig {
	return s.config
}

// UpdateAlertConfig 更新预警配置
func (s *quotaAlertService) UpdateAlertConfig(config *QuotaAlertConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 验证配置
	if config.WarningThreshold <= 0 || config.WarningThreshold > 100 {
		return fmt.Errorf("warning threshold must be between 0 and 100")
	}
	if config.CriticalThreshold <= 0 || config.CriticalThreshold > 100 {
		return fmt.Errorf("critical threshold must be between 0 and 100")
	}
	if config.UrgentThreshold <= 0 || config.UrgentThreshold > 100 {
		return fmt.Errorf("urgent threshold must be between 0 and 100")
	}

	// 确保阈值递增
	if config.CriticalThreshold <= config.WarningThreshold {
		return fmt.Errorf("critical threshold must be greater than warning threshold")
	}
	if config.UrgentThreshold <= config.CriticalThreshold {
		return fmt.Errorf("urgent threshold must be greater than critical threshold")
	}

	s.config = config
	s.logger.Info("Alert config updated",
		zap.Float64("warning_threshold", config.WarningThreshold),
		zap.Float64("critical_threshold", config.CriticalThreshold),
		zap.Float64("urgent_threshold", config.UrgentThreshold))

	return nil
}
