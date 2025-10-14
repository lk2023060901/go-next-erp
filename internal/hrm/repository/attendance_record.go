package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
)

// AttendanceRecordRepository 考勤记录仓储接口
type AttendanceRecordRepository interface {
	// Create 创建考勤记录
	Create(ctx context.Context, record *model.AttendanceRecord) error

	// BatchCreate 批量创建考勤记录
	BatchCreate(ctx context.Context, records []*model.AttendanceRecord) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.AttendanceRecord, error)

	// Update 更新考勤记录
	Update(ctx context.Context, record *model.AttendanceRecord) error

	// Delete 删除考勤记录
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *AttendanceRecordFilter, offset, limit int) ([]*model.AttendanceRecord, int, error)

	// FindByEmployee 查询员工考勤记录
	FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) ([]*model.AttendanceRecord, error)

	// FindByDepartment 查询部门考勤记录
	FindByDepartment(ctx context.Context, tenantID, departmentID uuid.UUID, startDate, endDate time.Time) ([]*model.AttendanceRecord, error)

	// FindByDateRange 按日期范围查询
	FindByDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*model.AttendanceRecord, error)

	// CountByStatus 统计各状态考勤记录数
	CountByStatus(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[model.AttendanceStatus]int, error)

	// FindExceptions 查询异常考勤记录
	FindExceptions(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*model.AttendanceRecord, error)
}

// AttendanceRecordFilter 考勤记录查询过滤器
type AttendanceRecordFilter struct {
	EmployeeID   *uuid.UUID
	DepartmentID *uuid.UUID
	StartDate    *time.Time
	EndDate      *time.Time
	ClockType    *model.AttendanceClockType
	Status       *model.AttendanceStatus
	SourceType   *model.SourceType
	IsException  *bool
	Keyword      string // 搜索关键词（员工姓名）
}
