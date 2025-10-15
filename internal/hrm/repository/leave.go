package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
)

// LeaveTypeRepository 请假类型仓储接口
type LeaveTypeRepository interface {
	// Create 创建请假类型
	Create(ctx context.Context, leaveType *model.LeaveType) error

	// Update 更新请假类型
	Update(ctx context.Context, leaveType *model.LeaveType) error

	// Delete 删除请假类型
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveType, error)

	// FindByCode 根据编码查找
	FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.LeaveType, error)

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *LeaveTypeFilter, offset, limit int) ([]*model.LeaveType, int, error)

	// ListActive 查询启用的请假类型
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.LeaveType, error)

	// 游标分页查询（高性能，适用于大数据量）
	ListWithCursor(ctx context.Context, tenantID uuid.UUID, filter *LeaveTypeFilter, cursor *time.Time, limit int) ([]*model.LeaveType, *time.Time, bool, error)
}

// LeaveTypeFilter 请假类型查询过滤器
type LeaveTypeFilter struct {
	IsActive      *bool
	RequiresProof *bool
	DeductQuota   *bool
	Keyword       string
}

// LeaveQuotaRepository 请假额度仓储接口
type LeaveQuotaRepository interface {
	// Create 创建请假额度
	Create(ctx context.Context, quota *model.LeaveQuota) error

	// Update 更新请假额度
	Update(ctx context.Context, quota *model.LeaveQuota) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveQuota, error)

	// FindByEmployeeAndType 根据员工和类型查找
	FindByEmployeeAndType(ctx context.Context, tenantID, employeeID, leaveTypeID uuid.UUID, year int) (*model.LeaveQuota, error)

	// ListByEmployee 查询员工的所有假期额度
	ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.LeaveQuota, error)

	// ListByEmployeeWithType 查询员工的所有假期额度（含类型信息）
	ListByEmployeeWithType(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.LeaveQuotaWithType, error)

	// IncrementUsedQuota 增加已使用额度
	IncrementUsedQuota(ctx context.Context, id uuid.UUID, amount float64) error

	// DecrementUsedQuota 减少已使用额度
	DecrementUsedQuota(ctx context.Context, id uuid.UUID, amount float64) error

	// IncrementPendingQuota 增加待审批额度
	IncrementPendingQuota(ctx context.Context, id uuid.UUID, amount float64) error

	// DecrementPendingQuota 减少待审批额度
	DecrementPendingQuota(ctx context.Context, id uuid.UUID, amount float64) error

	// BatchCreate 批量创建额度
	BatchCreate(ctx context.Context, quotas []*model.LeaveQuota) error
}

// LeaveRequestRepository 请假申请仓储接口
type LeaveRequestRepository interface {
	// Create 创建请假申请
	Create(ctx context.Context, request *model.LeaveRequest) error

	// Update 更新请假申请
	Update(ctx context.Context, request *model.LeaveRequest) error

	// Delete 删除请假申请
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveRequest, error)

	// FindByIDWithApprovals 根据ID查找（含审批记录）
	FindByIDWithApprovals(ctx context.Context, id uuid.UUID) (*model.LeaveRequestWithApprovals, error)

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *LeaveRequestFilter, offset, limit int) ([]*model.LeaveRequest, int, error)

	// ListByEmployee 查询员工的请假记录
	ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, filter *LeaveRequestFilter, offset, limit int) ([]*model.LeaveRequest, int, error)

	// ListPendingApprovals 查询待审批的请假申请
	ListPendingApprovals(ctx context.Context, tenantID, approverID uuid.UUID, offset, limit int) ([]*model.LeaveRequest, int, error)

	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.LeaveRequestStatus, operatedAt *time.Time) error

	// SetCurrentApprover 设置当前审批人
	SetCurrentApprover(ctx context.Context, id uuid.UUID, approverID *uuid.UUID) error

	// CheckTimeConflict 检查时间冲突
	CheckTimeConflict(ctx context.Context, tenantID, employeeID uuid.UUID, startTime, endTime time.Time, excludeID *uuid.UUID) (bool, error)

	// 游标分页查询（高性能，适用于大数据量）
	ListWithCursor(ctx context.Context, tenantID uuid.UUID, filter *LeaveRequestFilter, cursor *time.Time, limit int) ([]*model.LeaveRequest, *time.Time, bool, error)
	ListByEmployeeWithCursor(ctx context.Context, tenantID, employeeID uuid.UUID, filter *LeaveRequestFilter, cursor *time.Time, limit int) ([]*model.LeaveRequest, *time.Time, bool, error)
	ListPendingApprovalsWithCursor(ctx context.Context, tenantID, approverID uuid.UUID, cursor *time.Time, limit int) ([]*model.LeaveRequest, *time.Time, bool, error)
}

// LeaveRequestFilter 请假申请查询过滤器
type LeaveRequestFilter struct {
	LeaveTypeID  *uuid.UUID
	DepartmentID *uuid.UUID
	Status       *model.LeaveRequestStatus
	StartDate    *time.Time
	EndDate      *time.Time
	Keyword      string
}

// LeaveApprovalRepository 请假审批记录仓储接口
type LeaveApprovalRepository interface {
	// Create 创建审批记录
	Create(ctx context.Context, approval *model.LeaveApproval) error

	// Update 更新审批记录
	Update(ctx context.Context, approval *model.LeaveApproval) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveApproval, error)

	// ListByRequest 查询请假申请的所有审批记录
	ListByRequest(ctx context.Context, leaveRequestID uuid.UUID) ([]*model.LeaveApproval, error)

	// FindPendingApproval 查找待审批的记录
	FindPendingApproval(ctx context.Context, leaveRequestID uuid.UUID, approverID uuid.UUID) (*model.LeaveApproval, error)

	// UpdateStatus 更新审批状态
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.LeaveApprovalStatus, action *model.LeaveApprovalAction, comment string, approvedAt *time.Time) error

	// BatchCreate 批量创建审批记录
	BatchCreate(ctx context.Context, approvals []*model.LeaveApproval) error
}
