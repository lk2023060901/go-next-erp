package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/integration"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/pkg/database"
	"github.com/lk2023060901/go-next-erp/pkg/workflow"
)

// LeaveOfficeService 外出服务接口
type LeaveOfficeService interface {
	// Create 创建外出申请
	Create(ctx context.Context, leaveOffice *model.LeaveOffice) error

	// Update 更新外出申请
	Update(ctx context.Context, leaveOffice *model.LeaveOffice) error

	// Delete 删除外出申请
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据ID获取外出记录
	GetByID(ctx context.Context, id uuid.UUID) (*model.LeaveOffice, error)

	// List 列表查询
	List(ctx context.Context, tenantID uuid.UUID, filter *repository.LeaveOfficeFilter, offset, limit int) ([]*model.LeaveOffice, int, error)

	// ListByEmployee 查询员工外出记录
	ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.LeaveOffice, error)

	// Submit 提交外出审批
	Submit(ctx context.Context, leaveOfficeID, submitterID uuid.UUID) error

	// Approve 批准外出
	Approve(ctx context.Context, leaveOfficeID, approverID uuid.UUID, comment string) error

	// Reject 拒绝外出
	Reject(ctx context.Context, leaveOfficeID, approverID uuid.UUID, reason string) error

	// CheckTimeConflict 检查时间冲突
	CheckTimeConflict(ctx context.Context, tenantID, employeeID uuid.UUID, startTime, endTime time.Time, excludeID *uuid.UUID) (bool, error)
}

type leaveOfficeService struct {
	db              *database.DB
	leaveOfficeRepo repository.LeaveOfficeRepository
	workflowEngine  *integration.LeaveOfficeWorkflowEngine
}

// NewLeaveOfficeService 创建外出服务
func NewLeaveOfficeService(
	db *database.DB,
	leaveOfficeRepo repository.LeaveOfficeRepository,
	workflowEngine *workflow.Engine,
) LeaveOfficeService {
	return &leaveOfficeService{
		db:              db,
		leaveOfficeRepo: leaveOfficeRepo,
		workflowEngine:  integration.NewLeaveOfficeWorkflowEngine(workflowEngine),
	}
}

// Create 创建外出申请
func (s *leaveOfficeService) Create(ctx context.Context, leaveOffice *model.LeaveOffice) error {
	// 1. 验证时间
	if err := s.validateTime(leaveOffice.StartTime, leaveOffice.EndTime); err != nil {
		return err
	}

	// 2. 检查时间冲突
	hasConflict, err := s.CheckTimeConflict(ctx, leaveOffice.TenantID, leaveOffice.EmployeeID, leaveOffice.StartTime, leaveOffice.EndTime, nil)
	if err != nil {
		return fmt.Errorf("failed to check time conflict: %w", err)
	}
	if hasConflict {
		return fmt.Errorf("time conflict: employee already has a leave office record during this period")
	}

	// 3. 计算外出时长（小时）
	leaveOffice.Duration = leaveOffice.EndTime.Sub(leaveOffice.StartTime).Hours()

	// 4. 设置默认值
	leaveOffice.ID = uuid.Must(uuid.NewV7())
	leaveOffice.ApprovalStatus = "pending" // 默认待审批
	now := time.Now()
	leaveOffice.CreatedAt = now
	leaveOffice.UpdatedAt = now

	// 5. 创建记录
	return s.leaveOfficeRepo.Create(ctx, leaveOffice)
}

// Update 更新外出申请
func (s *leaveOfficeService) Update(ctx context.Context, leaveOffice *model.LeaveOffice) error {
	// 1. 获取原记录
	existing, err := s.leaveOfficeRepo.FindByID(ctx, leaveOffice.ID)
	if err != nil {
		return fmt.Errorf("failed to get leave office: %w", err)
	}

	// 2. 只有pending状态可以编辑
	if existing.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending leave office records can be updated")
	}

	// 3. 验证时间
	if err := s.validateTime(leaveOffice.StartTime, leaveOffice.EndTime); err != nil {
		return err
	}

	// 4. 检查时间冲突（排除自己）
	hasConflict, err := s.CheckTimeConflict(ctx, leaveOffice.TenantID, leaveOffice.EmployeeID, leaveOffice.StartTime, leaveOffice.EndTime, &leaveOffice.ID)
	if err != nil {
		return fmt.Errorf("failed to check time conflict: %w", err)
	}
	if hasConflict {
		return fmt.Errorf("time conflict: employee already has a leave office record during this period")
	}

	// 5. 计算外出时长
	leaveOffice.Duration = leaveOffice.EndTime.Sub(leaveOffice.StartTime).Hours()

	// 6. 更新时间戳
	leaveOffice.UpdatedAt = time.Now()

	return s.leaveOfficeRepo.Update(ctx, leaveOffice)
}

// Delete 删除外出申请
func (s *leaveOfficeService) Delete(ctx context.Context, id uuid.UUID) error {
	// 1. 获取记录
	leaveOffice, err := s.leaveOfficeRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get leave office: %w", err)
	}

	// 2. 只有pending状态可以删除
	if leaveOffice.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending leave office records can be deleted")
	}

	return s.leaveOfficeRepo.Delete(ctx, id)
}

// GetByID 根据ID获取外出记录
func (s *leaveOfficeService) GetByID(ctx context.Context, id uuid.UUID) (*model.LeaveOffice, error) {
	return s.leaveOfficeRepo.FindByID(ctx, id)
}

// List 列表查询
func (s *leaveOfficeService) List(ctx context.Context, tenantID uuid.UUID, filter *repository.LeaveOfficeFilter, offset, limit int) ([]*model.LeaveOffice, int, error) {
	return s.leaveOfficeRepo.List(ctx, tenantID, filter, offset, limit)
}

// ListByEmployee 查询员工外出记录
func (s *leaveOfficeService) ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.LeaveOffice, error) {
	return s.leaveOfficeRepo.FindByEmployee(ctx, tenantID, employeeID, year)
}

// Submit 提交外出审批
func (s *leaveOfficeService) Submit(ctx context.Context, leaveOfficeID, submitterID uuid.UUID) error {
	// 1. 获取外出记录
	leaveOffice, err := s.leaveOfficeRepo.FindByID(ctx, leaveOfficeID)
	if err != nil {
		return fmt.Errorf("failed to get leave office: %w", err)
	}

	// 2. 验证状态
	if leaveOffice.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending leave office records can be submitted")
	}

	// 3. 验证外出时间（不能是过去时间）
	if time.Now().After(leaveOffice.StartTime) {
		return fmt.Errorf("leave office start time cannot be in the past")
	}

	// 4. 使用工作流引擎执行审批流程
	workflowID := fmt.Sprintf("leave-office-approval-%s", leaveOffice.TenantID.String())
	executionID, err := s.workflowEngine.ExecuteLeaveOfficeApproval(ctx, workflowID, leaveOffice, submitterID.String())
	if err != nil {
		return fmt.Errorf("failed to execute workflow: %w", err)
	}

	// 5. 保存工作流执行ID到ApprovalID
	approvalID, _ := uuid.Parse(executionID)
	leaveOffice.ApprovalID = &approvalID
	leaveOffice.UpdatedAt = time.Now()

	return s.leaveOfficeRepo.Update(ctx, leaveOffice)
}

// Approve 批准外出
func (s *leaveOfficeService) Approve(ctx context.Context, leaveOfficeID, approverID uuid.UUID, comment string) error {
	// 1. 获取外出记录
	leaveOffice, err := s.leaveOfficeRepo.FindByID(ctx, leaveOfficeID)
	if err != nil {
		return fmt.Errorf("failed to get leave office: %w", err)
	}

	// 2. 验证状态
	if leaveOffice.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending leave office records can be approved")
	}

	// 3. 更新状态
	leaveOffice.ApprovalStatus = "approved"
	leaveOffice.ApprovedBy = &approverID
	now := time.Now()
	leaveOffice.ApprovedAt = &now
	leaveOffice.UpdatedAt = now

	return s.leaveOfficeRepo.Update(ctx, leaveOffice)
}

// Reject 拒绝外出
func (s *leaveOfficeService) Reject(ctx context.Context, leaveOfficeID, approverID uuid.UUID, reason string) error {
	// 1. 获取外出记录
	leaveOffice, err := s.leaveOfficeRepo.FindByID(ctx, leaveOfficeID)
	if err != nil {
		return fmt.Errorf("failed to get leave office: %w", err)
	}

	// 2. 验证状态
	if leaveOffice.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending leave office records can be rejected")
	}

	// 3. 更新状态
	leaveOffice.ApprovalStatus = "rejected"
	leaveOffice.ApprovedBy = &approverID
	leaveOffice.RejectReason = reason
	leaveOffice.UpdatedAt = time.Now()

	return s.leaveOfficeRepo.Update(ctx, leaveOffice)
}

// CheckTimeConflict 检查时间冲突
func (s *leaveOfficeService) CheckTimeConflict(ctx context.Context, tenantID, employeeID uuid.UUID, startTime, endTime time.Time, excludeID *uuid.UUID) (bool, error) {
	overlapping, err := s.leaveOfficeRepo.FindOverlapping(ctx, tenantID, employeeID, startTime, endTime)
	if err != nil {
		return false, err
	}

	// 排除指定ID
	if excludeID != nil {
		filtered := make([]*model.LeaveOffice, 0, len(overlapping))
		for _, lo := range overlapping {
			if lo.ID != *excludeID {
				filtered = append(filtered, lo)
			}
		}
		overlapping = filtered
	}

	return len(overlapping) > 0, nil
}

// validateTime 验证时间有效性
func (s *leaveOfficeService) validateTime(startTime, endTime time.Time) error {
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return fmt.Errorf("end time must be after start time")
	}

	// 外出通常是当天往返，建议不超过24小时
	duration := endTime.Sub(startTime).Hours()
	if duration > 24 {
		return fmt.Errorf("leave office duration should not exceed 24 hours, consider using business trip instead")
	}

	return nil
}
