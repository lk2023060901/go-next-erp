package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
	"github.com/lk2023060901/go-next-erp/internal/organization/repository"
)

// EmployeeService 员工服务接口
type EmployeeService interface {
	// Create 创建员工（入职）
	Create(ctx context.Context, req *CreateEmployeeRequest) (*model.Employee, error)

	// Update 更新员工信息
	Update(ctx context.Context, id uuid.UUID, req *UpdateEmployeeRequest) (*model.Employee, error)

	// Delete 删除员工（软删除）
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取员工
	GetByID(ctx context.Context, id uuid.UUID) (*model.Employee, error)

	// GetByUserID 根据用户 ID 获取员工
	GetByUserID(ctx context.Context, tenantID, userID uuid.UUID) (*model.Employee, error)

	// GetByEmployeeNo 根据工号获取员工
	GetByEmployeeNo(ctx context.Context, tenantID uuid.UUID, employeeNo string) (*model.Employee, error)

	// List 列出所有员工
	List(ctx context.Context, tenantID uuid.UUID) ([]*model.Employee, error)

	// ListByOrg 列出组织员工（包含子组织）
	ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*model.Employee, error)

	// ListByPosition 列出职位员工
	ListByPosition(ctx context.Context, positionID uuid.UUID) ([]*model.Employee, error)

	// ListByStatus 列出指定状态员工
	ListByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*model.Employee, error)

	// Transfer 调岗（更换组织或职位）
	Transfer(ctx context.Context, empID, newOrgID uuid.UUID, newPositionID *uuid.UUID, operatorID uuid.UUID) error

	// ChangePosition 更换职位
	ChangePosition(ctx context.Context, empID uuid.UUID, newPositionID uuid.UUID, operatorID uuid.UUID) error

	// ChangeLeader 更换上级
	ChangeLeader(ctx context.Context, empID, newLeaderID uuid.UUID, operatorID uuid.UUID) error

	// Regularize 转正
	Regularize(ctx context.Context, empID uuid.UUID, formalDate time.Time, operatorID uuid.UUID) error

	// Resign 离职
	Resign(ctx context.Context, empID uuid.UUID, leaveDate time.Time, operatorID uuid.UUID) error

	// Reinstate 复职
	Reinstate(ctx context.Context, empID uuid.UUID, operatorID uuid.UUID) error
}

type employeeService struct {
	empRepo repository.EmployeeRepository
	orgRepo repository.OrganizationRepository
	posRepo repository.PositionRepository
}

// NewEmployeeService 创建员工服务
func NewEmployeeService(
	empRepo repository.EmployeeRepository,
	orgRepo repository.OrganizationRepository,
	posRepo repository.PositionRepository,
) EmployeeService {
	return &employeeService{
		empRepo: empRepo,
		orgRepo: orgRepo,
		posRepo: posRepo,
	}
}

// CreateEmployeeRequest 创建员工请求
type CreateEmployeeRequest struct {
	TenantID       uuid.UUID
	UserID         uuid.UUID
	EmployeeNo     string
	Name           string
	Gender         string
	Mobile         string
	Email          string
	Avatar         string
	OrgID          uuid.UUID
	PositionID     *uuid.UUID
	DirectLeaderID *uuid.UUID
	JoinDate       *time.Time
	ProbationEnd   *time.Time
	Status         string // probation, active
	CreatedBy      uuid.UUID
}

// UpdateEmployeeRequest 更新员工请求
type UpdateEmployeeRequest struct {
	Name           string
	Gender         string
	Mobile         string
	Email          string
	Avatar         string
	DirectLeaderID *uuid.UUID
	UpdatedBy      uuid.UUID
}

func (s *employeeService) Create(ctx context.Context, req *CreateEmployeeRequest) (*model.Employee, error) {
	// 验证工号唯一性
	exists, err := s.empRepo.Exists(ctx, req.TenantID, req.EmployeeNo)
	if err != nil {
		return nil, fmt.Errorf("check employee no exists failed: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("employee no '%s' already exists", req.EmployeeNo)
	}

	// 获取组织信息（用于设置组织路径）
	org, err := s.orgRepo.GetByID(ctx, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("get organization failed: %w", err)
	}

	// 验证职位
	if req.PositionID != nil {
		_, err := s.posRepo.GetByID(ctx, *req.PositionID)
		if err != nil {
			return nil, fmt.Errorf("get position failed: %w", err)
		}
	}

	// 验证上级
	if req.DirectLeaderID != nil {
		_, err := s.empRepo.GetByID(ctx, *req.DirectLeaderID)
		if err != nil {
			return nil, fmt.Errorf("get direct leader failed: %w", err)
		}
	}

	// 创建员工
	emp := &model.Employee{
		ID:             uuid.New(),
		TenantID:       req.TenantID,
		UserID:         req.UserID,
		EmployeeNo:     req.EmployeeNo,
		Name:           req.Name,
		Gender:         req.Gender,
		Mobile:         req.Mobile,
		Email:          req.Email,
		Avatar:         req.Avatar,
		OrgID:          req.OrgID,
		OrgPath:        org.Path,
		PositionID:     req.PositionID,
		DirectLeaderID: req.DirectLeaderID,
		JoinDate:       req.JoinDate,
		ProbationEnd:   req.ProbationEnd,
		Status:         req.Status,
		CreatedBy:      req.CreatedBy,
		UpdatedBy:      req.CreatedBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.empRepo.Create(ctx, emp); err != nil {
		return nil, fmt.Errorf("create employee failed: %w", err)
	}

	return emp, nil
}

func (s *employeeService) Update(ctx context.Context, id uuid.UUID, req *UpdateEmployeeRequest) (*model.Employee, error) {
	// 获取现有员工
	emp, err := s.empRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get employee failed: %w", err)
	}

	// 验证上级
	if req.DirectLeaderID != nil {
		// 不能设置自己为上级
		if *req.DirectLeaderID == id {
			return nil, fmt.Errorf("cannot set self as direct leader")
		}

		_, err := s.empRepo.GetByID(ctx, *req.DirectLeaderID)
		if err != nil {
			return nil, fmt.Errorf("get direct leader failed: %w", err)
		}
	}

	// 更新字段
	emp.Name = req.Name
	emp.Gender = req.Gender
	emp.Mobile = req.Mobile
	emp.Email = req.Email
	emp.Avatar = req.Avatar
	emp.DirectLeaderID = req.DirectLeaderID
	emp.UpdatedBy = req.UpdatedBy
	emp.UpdatedAt = time.Now()

	if err := s.empRepo.Update(ctx, emp); err != nil {
		return nil, fmt.Errorf("update employee failed: %w", err)
	}

	return emp, nil
}

func (s *employeeService) Delete(ctx context.Context, id uuid.UUID) error {
	// 检查是否有下属
	subordinates, err := s.empRepo.ListByLeader(ctx, id)
	if err != nil {
		return fmt.Errorf("list subordinates failed: %w", err)
	}
	if len(subordinates) > 0 {
		return fmt.Errorf("employee has %d subordinates, cannot delete", len(subordinates))
	}

	return s.empRepo.Delete(ctx, id)
}

func (s *employeeService) GetByID(ctx context.Context, id uuid.UUID) (*model.Employee, error) {
	return s.empRepo.GetByID(ctx, id)
}

func (s *employeeService) GetByUserID(ctx context.Context, tenantID, userID uuid.UUID) (*model.Employee, error) {
	return s.empRepo.GetByUserID(ctx, tenantID, userID)
}

func (s *employeeService) GetByEmployeeNo(ctx context.Context, tenantID uuid.UUID, employeeNo string) (*model.Employee, error) {
	return s.empRepo.GetByEmployeeNo(ctx, tenantID, employeeNo)
}

func (s *employeeService) List(ctx context.Context, tenantID uuid.UUID) ([]*model.Employee, error) {
	return s.empRepo.List(ctx, tenantID)
}

func (s *employeeService) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*model.Employee, error) {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get organization failed: %w", err)
	}
	return s.empRepo.ListByOrg(ctx, org.Path)
}

func (s *employeeService) ListByPosition(ctx context.Context, positionID uuid.UUID) ([]*model.Employee, error) {
	return s.empRepo.ListByPosition(ctx, positionID)
}

func (s *employeeService) ListByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*model.Employee, error) {
	return s.empRepo.ListByStatus(ctx, tenantID, status)
}

func (s *employeeService) Transfer(ctx context.Context, empID, newOrgID uuid.UUID, newPositionID *uuid.UUID, operatorID uuid.UUID) error {
	// 获取员工
	emp, err := s.empRepo.GetByID(ctx, empID)
	if err != nil {
		return fmt.Errorf("get employee failed: %w", err)
	}

	// 检查状态
	if !emp.IsActive() {
		return fmt.Errorf("employee is not active, cannot transfer")
	}

	// 获取新组织
	newOrg, err := s.orgRepo.GetByID(ctx, newOrgID)
	if err != nil {
		return fmt.Errorf("get new organization failed: %w", err)
	}

	// 验证新职位
	if newPositionID != nil {
		_, err := s.posRepo.GetByID(ctx, *newPositionID)
		if err != nil {
			return fmt.Errorf("get new position failed: %w", err)
		}
	}

	// 更新组织
	if err := s.empRepo.UpdateOrg(ctx, empID, newOrgID, newOrg.Path); err != nil {
		return fmt.Errorf("update org failed: %w", err)
	}

	// 更新职位
	if newPositionID != nil {
		if err := s.empRepo.UpdatePosition(ctx, empID, newPositionID); err != nil {
			return fmt.Errorf("update position failed: %w", err)
		}
	}

	return nil
}

func (s *employeeService) ChangePosition(ctx context.Context, empID uuid.UUID, newPositionID uuid.UUID, operatorID uuid.UUID) error {
	// 获取员工
	emp, err := s.empRepo.GetByID(ctx, empID)
	if err != nil {
		return fmt.Errorf("get employee failed: %w", err)
	}

	// 检查状态
	if !emp.IsActive() {
		return fmt.Errorf("employee is not active, cannot change position")
	}

	// 验证新职位
	_, err = s.posRepo.GetByID(ctx, newPositionID)
	if err != nil {
		return fmt.Errorf("get new position failed: %w", err)
	}

	// 更新职位
	return s.empRepo.UpdatePosition(ctx, empID, &newPositionID)
}

func (s *employeeService) ChangeLeader(ctx context.Context, empID, newLeaderID uuid.UUID, operatorID uuid.UUID) error {
	// 获取员工
	emp, err := s.empRepo.GetByID(ctx, empID)
	if err != nil {
		return fmt.Errorf("get employee failed: %w", err)
	}

	// 不能设置自己为上级
	if newLeaderID == empID {
		return fmt.Errorf("cannot set self as direct leader")
	}

	// 验证新上级
	_, err = s.empRepo.GetByID(ctx, newLeaderID)
	if err != nil {
		return fmt.Errorf("get new leader failed: %w", err)
	}

	// 更新上级
	emp.DirectLeaderID = &newLeaderID
	emp.UpdatedAt = time.Now()

	return s.empRepo.Update(ctx, emp)
}

func (s *employeeService) Regularize(ctx context.Context, empID uuid.UUID, formalDate time.Time, operatorID uuid.UUID) error {
	// 获取员工
	emp, err := s.empRepo.GetByID(ctx, empID)
	if err != nil {
		return fmt.Errorf("get employee failed: %w", err)
	}

	// 检查状态
	if emp.Status != "probation" {
		return fmt.Errorf("employee is not in probation, cannot regularize")
	}

	// 更新状态为正式员工
	emp.Status = "active"
	emp.FormalDate = &formalDate
	emp.UpdatedAt = time.Now()

	return s.empRepo.Update(ctx, emp)
}

func (s *employeeService) Resign(ctx context.Context, empID uuid.UUID, leaveDate time.Time, operatorID uuid.UUID) error {
	// 获取员工
	emp, err := s.empRepo.GetByID(ctx, empID)
	if err != nil {
		return fmt.Errorf("get employee failed: %w", err)
	}

	// 检查状态
	if emp.Status == "resigned" {
		return fmt.Errorf("employee has already resigned")
	}

	// 检查是否有下属
	subordinates, err := s.empRepo.ListByLeader(ctx, empID)
	if err != nil {
		return fmt.Errorf("list subordinates failed: %w", err)
	}
	if len(subordinates) > 0 {
		return fmt.Errorf("employee has %d subordinates, cannot resign", len(subordinates))
	}

	// 更新状态为离职
	emp.Status = "resigned"
	emp.LeaveDate = &leaveDate
	emp.UpdatedAt = time.Now()

	return s.empRepo.Update(ctx, emp)
}

func (s *employeeService) Reinstate(ctx context.Context, empID uuid.UUID, operatorID uuid.UUID) error {
	// 获取员工
	emp, err := s.empRepo.GetByID(ctx, empID)
	if err != nil {
		return fmt.Errorf("get employee failed: %w", err)
	}

	// 检查状态
	if emp.Status != "resigned" {
		return fmt.Errorf("employee is not resigned, cannot reinstate")
	}

	// 更新状态为在职
	emp.Status = "active"
	emp.LeaveDate = nil
	emp.UpdatedAt = time.Now()

	return s.empRepo.Update(ctx, emp)
}
