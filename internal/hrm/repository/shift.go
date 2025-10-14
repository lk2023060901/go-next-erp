package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
)

// ShiftRepository 班次仓储接口
type ShiftRepository interface {
	// Create 创建班次
	Create(ctx context.Context, shift *model.Shift) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.Shift, error)

	// FindByCode 根据编码查找
	FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Shift, error)

	// Update 更新班次
	Update(ctx context.Context, shift *model.Shift) error

	// Delete 删除班次
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *ShiftFilter, offset, limit int) ([]*model.Shift, int, error)

	// ListActive 查询启用的班次
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.Shift, error)

	// ListByType 按类型查询
	ListByType(ctx context.Context, tenantID uuid.UUID, shiftType model.ShiftType) ([]*model.Shift, error)
}

// ShiftFilter 班次查询过滤器
type ShiftFilter struct {
	Type     *model.ShiftType
	IsActive *bool
	Keyword  string // 搜索关键词（名称、编码）
}

// ScheduleRepository 排班仓储接口
type ScheduleRepository interface {
	// Create 创建排班
	Create(ctx context.Context, schedule *model.Schedule) error

	// BatchCreate 批量创建排班
	BatchCreate(ctx context.Context, schedules []*model.Schedule) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error)

	// Update 更新排班
	Update(ctx context.Context, schedule *model.Schedule) error

	// Delete 删除排班
	Delete(ctx context.Context, id uuid.UUID) error

	// BatchDelete 批量删除排班
	BatchDelete(ctx context.Context, ids []uuid.UUID) error

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *ScheduleFilter, offset, limit int) ([]*model.Schedule, int, error)

	// FindByEmployee 查询员工排班
	FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, month string) ([]*model.Schedule, error)

	// FindByDepartment 查询部门排班
	FindByDepartment(ctx context.Context, tenantID, departmentID uuid.UUID, month string) ([]*model.Schedule, error)

	// FindByDate 查询某日排班
	FindByDate(ctx context.Context, tenantID uuid.UUID, date string) ([]*model.Schedule, error)
}

// ScheduleFilter 排班查询过滤器
type ScheduleFilter struct {
	EmployeeID   *uuid.UUID
	DepartmentID *uuid.UUID
	ShiftID      *uuid.UUID
	StartDate    *string
	EndDate      *string
	Status       *string
}

// AttendanceRuleRepository 考勤规则仓储接口
type AttendanceRuleRepository interface {
	// Create 创建考勤规则
	Create(ctx context.Context, rule *model.AttendanceRule) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.AttendanceRule, error)

	// FindByCode 根据编码查找
	FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.AttendanceRule, error)

	// Update 更新考勤规则
	Update(ctx context.Context, rule *model.AttendanceRule) error

	// Delete 删除考勤规则
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *AttendanceRuleFilter, offset, limit int) ([]*model.AttendanceRule, int, error)

	// ListActive 查询启用的规则
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.AttendanceRule, error)

	// FindByEmployee 查询员工适用的规则（按优先级）
	FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) (*model.AttendanceRule, error)
}

// AttendanceRuleFilter 考勤规则查询过滤器
type AttendanceRuleFilter struct {
	ApplyType *model.ApplyType
	IsActive  *bool
	Keyword   string
}
