package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
)

// HRMEmployeeRepository HRM员工扩展信息仓储接口
type HRMEmployeeRepository interface {
	// Create 创建HRM员工扩展信息
	Create(ctx context.Context, emp *model.HRMEmployee) error

	// Update 更新HRM员工扩展信息
	Update(ctx context.Context, emp *model.HRMEmployee) error

	// Delete 删除HRM员工扩展信息
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.HRMEmployee, error)

	// FindByEmployeeID 根据组织员工ID查找
	FindByEmployeeID(ctx context.Context, tenantID, employeeID uuid.UUID) (*model.HRMEmployee, error)

	// FindByCardNo 根据考勤卡号查找
	FindByCardNo(ctx context.Context, tenantID uuid.UUID, cardNo string) (*model.HRMEmployee, error)

	// FindByThirdPartyID 根据第三方平台ID查找
	FindByThirdPartyID(ctx context.Context, tenantID uuid.UUID, platform model.PlatformType, platformID string) (*model.HRMEmployee, error)

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *HRMEmployeeFilter, offset, limit int) ([]*model.HRMEmployee, int, error)

	// ListByAttendanceRule 根据考勤规则查询员工
	ListByAttendanceRule(ctx context.Context, tenantID, ruleID uuid.UUID) ([]*model.HRMEmployee, error)

	// ListByShift 根据班次查询员工
	ListByShift(ctx context.Context, tenantID, shiftID uuid.UUID) ([]*model.HRMEmployee, error)

	// ListActive 查询启用考勤的员工
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.HRMEmployee, error)

	// UpdateFaceData 更新人脸数据
	UpdateFaceData(ctx context.Context, id uuid.UUID, faceData string) error

	// UpdateFingerprint 更新指纹数据
	UpdateFingerprint(ctx context.Context, id uuid.UUID, fingerprint string) error

	// UpdateCardNo 更新考勤卡号
	UpdateCardNo(ctx context.Context, id uuid.UUID, cardNo string) error

	// UpdateThirdPartyID 更新第三方平台ID
	UpdateThirdPartyID(ctx context.Context, id uuid.UUID, platform model.PlatformType, platformID string) error

	// BatchCreate 批量创建
	BatchCreate(ctx context.Context, employees []*model.HRMEmployee) error

	// ExistsByEmployeeID 检查员工是否已有HRM扩展信息
	ExistsByEmployeeID(ctx context.Context, tenantID, employeeID uuid.UUID) (bool, error)

	// 游标分页查询（高性能，适用于大数据量）
	ListWithCursor(ctx context.Context, tenantID uuid.UUID, filter *HRMEmployeeFilter, cursor *time.Time, limit int) ([]*model.HRMEmployee, *time.Time, bool, error)
}

// HRMEmployeeFilter HRM员工查询过滤器
type HRMEmployeeFilter struct {
	AttendanceRuleID *uuid.UUID
	DefaultShiftID   *uuid.UUID
	WorkLocation     *string
	IsActive         *bool
	HasFaceData      *bool
	HasFingerprint   *bool
	Keyword          string // 搜索关键词
}

// EmployeeSyncMappingRepository 员工同步映射仓储接口
type EmployeeSyncMappingRepository interface {
	// Create 创建同步映射
	Create(ctx context.Context, mapping *model.EmployeeSyncMapping) error

	// Update 更新同步映射
	Update(ctx context.Context, mapping *model.EmployeeSyncMapping) error

	// Delete 删除同步映射
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.EmployeeSyncMapping, error)

	// FindByEmployee 查询员工的所有平台映射
	FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) ([]*model.EmployeeSyncMapping, error)

	// FindByPlatform 根据平台和平台ID查找
	FindByPlatform(ctx context.Context, tenantID uuid.UUID, platform model.PlatformType, platformID string) (*model.EmployeeSyncMapping, error)

	// ListByPlatform 查询某平台的所有映射
	ListByPlatform(ctx context.Context, tenantID uuid.UUID, platform model.PlatformType) ([]*model.EmployeeSyncMapping, error)

	// ListSyncEnabled 查询启用同步的映射
	ListSyncEnabled(ctx context.Context, tenantID uuid.UUID, platform model.PlatformType) ([]*model.EmployeeSyncMapping, error)

	// UpdateSyncStatus 更新同步状态
	UpdateSyncStatus(ctx context.Context, id uuid.UUID, status, errorMsg string) error

	// UpdateLastSyncTime 更新最后同步时间
	UpdateLastSyncTime(ctx context.Context, id uuid.UUID) error

	// BatchCreate 批量创建
	BatchCreate(ctx context.Context, mappings []*model.EmployeeSyncMapping) error

	// ExistsByPlatform 检查平台映射是否存在
	ExistsByPlatform(ctx context.Context, tenantID uuid.UUID, platform model.PlatformType, platformID string) (bool, error)
}

// EmployeeWorkScheduleRepository 员工工作时间表仓储接口
type EmployeeWorkScheduleRepository interface {
	// Create 创建工作时间表
	Create(ctx context.Context, schedule *model.EmployeeWorkSchedule) error

	// Update 更新工作时间表
	Update(ctx context.Context, schedule *model.EmployeeWorkSchedule) error

	// Delete 删除工作时间表
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.EmployeeWorkSchedule, error)

	// FindActiveByEmployee 查询员工当前生效的工作时间表
	FindActiveByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) (*model.EmployeeWorkSchedule, error)

	// ListByEmployee 查询员工的所有工作时间表
	ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) ([]*model.EmployeeWorkSchedule, error)

	// ListByType 根据时间表类型查询
	ListByType(ctx context.Context, tenantID uuid.UUID, scheduleType string) ([]*model.EmployeeWorkSchedule, error)

	// ListActive 查询所有生效的工作时间表
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.EmployeeWorkSchedule, error)

	// DeactivateOld 停用员工的旧时间表
	DeactivateOld(ctx context.Context, tenantID, employeeID uuid.UUID) error
}
