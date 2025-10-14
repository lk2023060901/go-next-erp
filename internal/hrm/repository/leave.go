package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
)

// LeaveRepository 请假仓储接口
type LeaveRepository interface {
	// Create 创建请假记录
	Create(ctx context.Context, leave *model.Leave) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.Leave, error)

	// Update 更新请假记录
	Update(ctx context.Context, leave *model.Leave) error

	// Delete 删除请假记录
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *LeaveFilter, offset, limit int) ([]*model.Leave, int, error)

	// FindByEmployee 查询员工请假记录
	FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.Leave, error)

	// FindPending 查询待审批的请假
	FindPending(ctx context.Context, tenantID uuid.UUID) ([]*model.Leave, error)

	// FindOverlapping 查询时间重叠的请假记录
	FindOverlapping(ctx context.Context, tenantID, employeeID uuid.UUID, startTime, endTime time.Time) ([]*model.Leave, error)
}

// LeaveFilter 请假查询过滤器
type LeaveFilter struct {
	EmployeeID     *uuid.UUID
	DepartmentID   *uuid.UUID
	LeaveTypeID    *uuid.UUID
	ApprovalStatus *string
	StartDate      *time.Time
	EndDate        *time.Time
	Keyword        string
}

// LeaveTypeRepository 请假类型仓储接口
type LeaveTypeRepository interface {
	// Create 创建请假类型
	Create(ctx context.Context, leaveType *model.LeaveType) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveType, error)

	// FindByCode 根据编码查找
	FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.LeaveType, error)

	// Update 更新请假类型
	Update(ctx context.Context, leaveType *model.LeaveType) error

	// Delete 删除请假类型
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *LeaveTypeFilter, offset, limit int) ([]*model.LeaveType, int, error)

	// ListActive 查询启用的请假类型
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.LeaveType, error)

	// ListByEmployee 查询员工可用的请假类型
	ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) ([]*model.LeaveType, error)
}

// LeaveTypeFilter 请假类型查询过滤器
type LeaveTypeFilter struct {
	IsPaid   *bool
	IsActive *bool
	Keyword  string
}

// LeaveQuotaRepository 请假额度仓储接口
type LeaveQuotaRepository interface {
	// Create 创建请假额度
	Create(ctx context.Context, quota *model.LeaveQuota) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveQuota, error)

	// FindByEmployee 查询员工请假额度
	FindByEmployee(ctx context.Context, tenantID, employeeID, leaveTypeID uuid.UUID, year int) (*model.LeaveQuota, error)

	// Update 更新请假额度
	Update(ctx context.Context, quota *model.LeaveQuota) error

	// ListByEmployee 查询员工所有请假额度
	ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.LeaveQuota, error)

	// BatchCreate 批量创建请假额度
	BatchCreate(ctx context.Context, quotas []*model.LeaveQuota) error
}
