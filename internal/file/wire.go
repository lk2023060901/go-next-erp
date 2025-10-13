package file

import (
	"time"

	"github.com/google/wire"
	"github.com/lk2023060901/go-next-erp/internal/file/repository"
	"github.com/lk2023060901/go-next-erp/internal/file/service"
)

// ProviderSet 是文件模块的 Wire Provider Set
var ProviderSet = wire.NewSet(
	// Repository 层
	repository.NewFileRepository,
	repository.NewQuotaRepository,
	repository.NewMultipartUploadRepository,
	repository.NewDownloadStatsRepository,
	repository.NewFileRelationRepository,
	repository.NewFileAccessLogRepository,

	// Service 层
	service.NewUploadService,
	service.NewDownloadService,
	service.NewQuotaService,
	service.NewQuotaAlertService,
	service.NewMultipartUploadService,
	service.NewFileManagementService,
	service.NewFileRelationService,
	service.NewCleanupService,

	// Service 配置
	ProvideUploadServiceConfig,
	ProvideQuotaServiceConfig,
	ProvideQuotaAlertConfig,
	ProvideMultipartUploadServiceConfig,
)

// ProvideUploadServiceConfig 提供上传服务配置
func ProvideUploadServiceConfig() *service.UploadServiceConfig {
	return &service.UploadServiceConfig{
		MaxFileSize:         100 * 1024 * 1024, // 100MB
		EnableVirusScan:     false,
		EnableDeduplication: true,
	}
}

// ProvideQuotaServiceConfig 提供配额服务配置
func ProvideQuotaServiceConfig() *service.QuotaServiceConfig {
	return &service.QuotaServiceConfig{
		DefaultTenantQuota: 10 * 1024 * 1024 * 1024, // 10GB
		DefaultUserQuota:   1 * 1024 * 1024 * 1024,  // 1GB
	}
}

// ProvideQuotaAlertConfig 提供配额预警配置
func ProvideQuotaAlertConfig() *service.QuotaAlertConfig {
	return &service.QuotaAlertConfig{
		WarningThreshold:  80.0,
		CriticalThreshold: 90.0,
		UrgentThreshold:   95.0,
	}
}

// ProvideMultipartUploadServiceConfig 提供分片上传服务配置
func ProvideMultipartUploadServiceConfig() *service.MultipartUploadServiceConfig {
	return &service.MultipartUploadServiceConfig{
		PartSize:          5 * 1024 * 1024,        // 5MB
		MaxPartSize:       5 * 1024 * 1024 * 1024, // 5GB
		MinPartSize:       5 * 1024 * 1024,        // 5MB
		UploadExpiration:  7 * 24 * time.Hour,     // 7 days
		MaxPartsPerUpload: 10000,
	}
}
