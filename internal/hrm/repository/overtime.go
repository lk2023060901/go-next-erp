package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
)

// OvertimeRepository 加班仓储接口
type OvertimeRepository interface {
	// Create 创建加班记录
	Create(ctx context.Context, overtime *model.Overtime) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.Overtime, error)

	// Update 更新加班记录
	Update(ctx context.Context, overtime *model.Overtime) error

	// Delete 删除加班记录
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *OvertimeFilter, offset, limit int) ([]*model.Overtime, int, error)

	// FindByEmployee 查询员工加班记录
	FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.Overtime, error)

	// FindPending 查询待审批的加班
	FindPending(ctx context.Context, tenantID uuid.UUID) ([]*model.Overtime, error)

	// SumHoursByEmployee 统计员工加班时长
	SumHoursByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) (float64, error)

	// SumCompOffDays 统计可调休天数
	SumCompOffDays(ctx context.Context, tenantID, employeeID uuid.UUID) (float64, error)
}

// OvertimeFilter 加班查询过滤器
type OvertimeFilter struct {
	EmployeeID     *uuid.UUID
	DepartmentID   *uuid.UUID
	OvertimeType   *model.OvertimeType
	ApprovalStatus *string
	StartDate      *time.Time
	EndDate        *time.Time
	Keyword        string
}

// BusinessTripRepository 出差仓储接口
type BusinessTripRepository interface {
	// Create 创建出差记录
	Create(ctx context.Context, trip *model.BusinessTrip) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.BusinessTrip, error)

	// Update 更新出差记录
	Update(ctx context.Context, trip *model.BusinessTrip) error

	// Delete 删除出差记录
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *BusinessTripFilter, offset, limit int) ([]*model.BusinessTrip, int, error)

	// FindByEmployee 查询员工出差记录
	FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.BusinessTrip, error)

	// FindPending 查询待审批的出差
	FindPending(ctx context.Context, tenantID uuid.UUID) ([]*model.BusinessTrip, error)

	// FindOverlapping 查询时间重叠的出差记录
	FindOverlapping(ctx context.Context, tenantID, employeeID uuid.UUID, startTime, endTime time.Time) ([]*model.BusinessTrip, error)
}

// BusinessTripFilter 出差查询过滤器
type BusinessTripFilter struct {
	EmployeeID     *uuid.UUID
	DepartmentID   *uuid.UUID
	ApprovalStatus *string
	StartDate      *time.Time
	EndDate        *time.Time
	Keyword        string
}

// LeaveOfficeRepository 外出仓储接口
type LeaveOfficeRepository interface {
	// Create 创建外出记录
	Create(ctx context.Context, leaveOffice *model.LeaveOffice) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveOffice, error)

	// Update 更新外出记录
	Update(ctx context.Context, leaveOffice *model.LeaveOffice) error

	// Delete 删除外出记录
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *LeaveOfficeFilter, offset, limit int) ([]*model.LeaveOffice, int, error)

	// FindByEmployee 查询员工外出记录
	FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.LeaveOffice, error)

	// FindPending 查询待审批的外出
	FindPending(ctx context.Context, tenantID uuid.UUID) ([]*model.LeaveOffice, error)

	// FindOverlapping 查询时间重叠的外出记录
	FindOverlapping(ctx context.Context, tenantID, employeeID uuid.UUID, startTime, endTime time.Time) ([]*model.LeaveOffice, error)
}

// LeaveOfficeFilter 外出查询过滤器
type LeaveOfficeFilter struct {
	EmployeeID     *uuid.UUID
	DepartmentID   *uuid.UUID
	ApprovalStatus *string
	StartDate      *time.Time
	EndDate        *time.Time
	Keyword        string
}

// AttendanceSummaryRepository 考勤汇总仓储接口
type AttendanceSummaryRepository interface {
	// Create 创建考勤汇总
	Create(ctx context.Context, summary *model.AttendanceSummary) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.AttendanceSummary, error)

	// FindByEmployee 查询员工考勤汇总
	FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year, month int) (*model.AttendanceSummary, error)

	// Update 更新考勤汇总
	Update(ctx context.Context, summary *model.AttendanceSummary) error

	// Delete 删除考勤汇总
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *AttendanceSummaryFilter, offset, limit int) ([]*model.AttendanceSummary, int, error)

	// ListByDepartment 按部门查询考勤汇总
	ListByDepartment(ctx context.Context, tenantID, departmentID uuid.UUID, year, month int) ([]*model.AttendanceSummary, error)

	// ConfirmSummary 确认考勤汇总
	ConfirmSummary(ctx context.Context, id, confirmedBy uuid.UUID) error

	// LockSummary 锁定考勤汇总
	LockSummary(ctx context.Context, tenantID uuid.UUID, year, month int) error
}

// AttendanceSummaryFilter 考勤汇总查询过滤器
type AttendanceSummaryFilter struct {
	EmployeeID   *uuid.UUID
	DepartmentID *uuid.UUID
	Year         *int
	Month        *int
	Status       *string
	Keyword      string
}

// PunchCardSupplementRepository 补卡申请仓储接口
type PunchCardSupplementRepository interface {
	// Create 创建补卡申请
	Create(ctx context.Context, supplement *model.PunchCardSupplement) error

	// FindByID 根据ID查找
	FindByID(ctx context.Context, id uuid.UUID) (*model.PunchCardSupplement, error)

	// Update 更新补卡申请
	Update(ctx context.Context, supplement *model.PunchCardSupplement) error

	// Delete 删除补卡申请
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询（分页）
	List(ctx context.Context, tenantID uuid.UUID, filter *PunchCardSupplementFilter, offset, limit int) ([]*model.PunchCardSupplement, int, error)

	// FindByEmployee 查询员工补卡申请记录
	FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.PunchCardSupplement, error)

	// FindPending 查询待审批的补卡申请
	FindPending(ctx context.Context, tenantID uuid.UUID) ([]*model.PunchCardSupplement, error)

	// FindByDate 查询指定日期的补卡申请
	FindByDate(ctx context.Context, tenantID, employeeID uuid.UUID, date time.Time, supplementType model.SupplementType) (*model.PunchCardSupplement, error)

	// CountByEmployee 统计员工补卡次数
	CountByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) (int, error)
}

// PunchCardSupplementFilter 补卡申请查询过滤器
type PunchCardSupplementFilter struct {
	EmployeeID     *uuid.UUID
	DepartmentID   *uuid.UUID
	SupplementType *model.SupplementType
	MissingType    *model.PunchCardMissingType
	ApprovalStatus *string
	ProcessStatus  *string
	StartDate      *time.Time
	EndDate        *time.Time
	Keyword        string
}
