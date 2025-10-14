package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// ScheduleService 排班服务接口
type ScheduleService interface {
	// Create 创建排班
	Create(ctx context.Context, req *CreateScheduleRequest) (*model.Schedule, error)

	// BatchCreate 批量创建排班
	BatchCreate(ctx context.Context, reqs []*CreateScheduleRequest) ([]*model.Schedule, error)

	// GetByID 根据ID获取排班
	GetByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error)

	// Update 更新排班
	Update(ctx context.Context, id uuid.UUID, req *UpdateScheduleRequest) (*model.Schedule, error)

	// Delete 删除排班
	Delete(ctx context.Context, id uuid.UUID) error

	// ListEmployeeSchedules 查询员工排班
	ListEmployeeSchedules(ctx context.Context, tenantID, employeeID uuid.UUID, month string) ([]*model.Schedule, error)

	// ListDepartmentSchedules 查询部门排班
	ListDepartmentSchedules(ctx context.Context, tenantID, departmentID uuid.UUID, month string) ([]*model.Schedule, error)
}

// CreateScheduleRequest 创建排班请求
type CreateScheduleRequest struct {
	TenantID     uuid.UUID
	EmployeeID   uuid.UUID
	EmployeeName string
	DepartmentID uuid.UUID
	ShiftID      uuid.UUID
	ShiftName    string
	ScheduleDate time.Time
	WorkdayType  string
	Status       string
	Remark       string
	CreatedBy    uuid.UUID
}

// UpdateScheduleRequest 更新排班请求
type UpdateScheduleRequest struct {
	ShiftID      *uuid.UUID
	ShiftName    *string
	ScheduleDate *time.Time
	WorkdayType  *string
	Status       *string
	Remark       *string
	UpdatedBy    uuid.UUID
}

type scheduleService struct {
	scheduleRepo repository.ScheduleRepository
	shiftRepo    repository.ShiftRepository
	hrmEmpRepo   repository.HRMEmployeeRepository
	db           *database.DB
}

// NewScheduleService 创建排班服务
func NewScheduleService(
	scheduleRepo repository.ScheduleRepository,
	shiftRepo repository.ShiftRepository,
	hrmEmpRepo repository.HRMEmployeeRepository,
	db *database.DB,
) ScheduleService {
	return &scheduleService{
		scheduleRepo: scheduleRepo,
		shiftRepo:    shiftRepo,
		hrmEmpRepo:   hrmEmpRepo,
		db:           db,
	}
}

func (s *scheduleService) Create(ctx context.Context, req *CreateScheduleRequest) (*model.Schedule, error) {
	// 验证班次存在
	shift, err := s.shiftRepo.FindByID(ctx, req.ShiftID)
	if err != nil {
		return nil, fmt.Errorf("shift not found: %w", err)
	}

	// 验证员工存在
	_, err = s.hrmEmpRepo.FindByEmployeeID(ctx, req.TenantID, req.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	// 如果没有提供 DepartmentID，从 employees 表获取
	departmentID := req.DepartmentID
	if departmentID == uuid.Nil {
		var orgID uuid.UUID
		sql := `SELECT org_id FROM employees WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`
		err = s.db.QueryRow(ctx, sql, req.EmployeeID, req.TenantID).Scan(&orgID)
		if err == nil {
			departmentID = orgID
		}
	}

	schedule := &model.Schedule{
		ID:           uuid.New(),
		TenantID:     req.TenantID,
		EmployeeID:   req.EmployeeID,
		EmployeeName: req.EmployeeName,
		DepartmentID: departmentID,
		ShiftID:      req.ShiftID,
		ShiftName:    shift.Name,
		ScheduleDate: req.ScheduleDate,
		WorkdayType:  req.WorkdayType,
		Status:       req.Status,
		Remark:       req.Remark,
		CreatedBy:    req.CreatedBy,
		UpdatedBy:    req.CreatedBy,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.scheduleRepo.Create(ctx, schedule); err != nil {
		return nil, fmt.Errorf("create schedule failed: %w", err)
	}

	return schedule, nil
}

func (s *scheduleService) BatchCreate(ctx context.Context, reqs []*CreateScheduleRequest) ([]*model.Schedule, error) {
	schedules := make([]*model.Schedule, 0, len(reqs))

	for _, req := range reqs {
		schedule, err := s.Create(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("batch create failed at index: %w", err)
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

func (s *scheduleService) GetByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error) {
	return s.scheduleRepo.FindByID(ctx, id)
}

func (s *scheduleService) Update(ctx context.Context, id uuid.UUID, req *UpdateScheduleRequest) (*model.Schedule, error) {
	schedule, err := s.scheduleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.ShiftID != nil {
		schedule.ShiftID = *req.ShiftID
		// 更新班次名称
		shift, err := s.shiftRepo.FindByID(ctx, *req.ShiftID)
		if err == nil {
			schedule.ShiftName = shift.Name
		}
	}
	if req.ShiftName != nil {
		schedule.ShiftName = *req.ShiftName
	}
	if req.ScheduleDate != nil {
		schedule.ScheduleDate = *req.ScheduleDate
	}
	if req.WorkdayType != nil {
		schedule.WorkdayType = *req.WorkdayType
	}
	if req.Status != nil {
		schedule.Status = *req.Status
	}
	if req.Remark != nil {
		schedule.Remark = *req.Remark
	}

	schedule.UpdatedBy = req.UpdatedBy
	schedule.UpdatedAt = time.Now()

	if err := s.scheduleRepo.Update(ctx, schedule); err != nil {
		return nil, fmt.Errorf("update schedule failed: %w", err)
	}

	return schedule, nil
}

func (s *scheduleService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.scheduleRepo.Delete(ctx, id)
}

func (s *scheduleService) ListEmployeeSchedules(ctx context.Context, tenantID, employeeID uuid.UUID, month string) ([]*model.Schedule, error) {
	return s.scheduleRepo.FindByEmployee(ctx, tenantID, employeeID, month)
}

func (s *scheduleService) ListDepartmentSchedules(ctx context.Context, tenantID, departmentID uuid.UUID, month string) ([]*model.Schedule, error) {
	return s.scheduleRepo.FindByDepartment(ctx, tenantID, departmentID, month)
}
