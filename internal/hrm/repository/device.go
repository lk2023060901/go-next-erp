package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
)

// AttendanceDeviceRepository 考勤设备仓储接口
type AttendanceDeviceRepository interface {
	// Create 创建设备
	Create(ctx context.Context, device *model.AttendanceDevice) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.AttendanceDevice, error)

	// FindBySN 根据序列号查找
	FindBySN(ctx context.Context, tenantID uuid.UUID, deviceSN string) (*model.AttendanceDevice, error)

	// Update 更新设备
	Update(ctx context.Context, device *model.AttendanceDevice) error

	// Delete 删除设备
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *DeviceFilter, offset, limit int) ([]*model.AttendanceDevice, int, error)

	// ListActive 查询启用的设备
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.AttendanceDevice, error)

	// ListOnline 查询在线设备
	ListOnline(ctx context.Context, tenantID uuid.UUID) ([]*model.AttendanceDevice, error)

	// UpdateHeartbeat 更新心跳时间
	UpdateHeartbeat(ctx context.Context, id uuid.UUID) error

	// UpdateSyncTime 更新同步时间
	UpdateSyncTime(ctx context.Context, id uuid.UUID) error

	// UpdateStatus 更新设备状态
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.DeviceStatus, errorMsg string) error
}

// DeviceFilter 设备查询过滤器
type DeviceFilter struct {
	DeviceType   *model.DeviceType
	Status       *model.DeviceStatus
	DepartmentID *uuid.UUID
	IsActive     *bool
	Keyword      string // 搜索关键词（设备名称、序列号）
}

// ThirdPartyIntegrationRepository 第三方集成仓储接口
type ThirdPartyIntegrationRepository interface {
	// Create 创建集成配置
	Create(ctx context.Context, integration *model.ThirdPartyIntegration) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.ThirdPartyIntegration, error)

	// FindByPlatform 根据平台查找
	FindByPlatform(ctx context.Context, tenantID uuid.UUID, platform model.PlatformType) (*model.ThirdPartyIntegration, error)

	// Update 更新集成配置
	Update(ctx context.Context, integration *model.ThirdPartyIntegration) error

	// Delete 删除集成配置
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *IntegrationFilter, offset, limit int) ([]*model.ThirdPartyIntegration, int, error)

	// ListActive 查询启用的集成
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.ThirdPartyIntegration, error)

	// UpdateSyncTime 更新同步时间
	UpdateSyncTime(ctx context.Context, id uuid.UUID, syncCount int) error

	// UpdateStatus 更新集成状态
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.IntegrationStatus, errorMsg string) error
}

// IntegrationFilter 集成查询过滤器
type IntegrationFilter struct {
	Platform *model.PlatformType
	Status   *model.IntegrationStatus
	IsActive *bool
	Keyword  string
}

// SyncLogRepository 同步日志仓储接口
type SyncLogRepository interface {
	// Create 创建同步日志
	Create(ctx context.Context, log *model.SyncLog) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.SyncLog, error)

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *SyncLogFilter, offset, limit int) ([]*model.SyncLog, int, error)

	// ListBySource 查询某来源的同步日志
	ListBySource(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID, limit int) ([]*model.SyncLog, error)

	// Delete 删除日志（清理历史日志）
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteBefore 删除指定时间之前的日志
	DeleteBefore(ctx context.Context, tenantID uuid.UUID, beforeDate string) error
}

// SyncLogFilter 同步日志查询过滤器
type SyncLogFilter struct {
	SourceType    *string
	SourceID      *uuid.UUID
	SyncType      *string
	SyncDirection *string
	Status        *string
	StartDate     *string
	EndDate       *string
}
