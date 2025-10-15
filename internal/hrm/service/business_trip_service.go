package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/hrm/integration"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/pkg/database"
	"github.com/lk2023060901/go-next-erp/pkg/workflow"
)

// BusinessTripService 出差服务接口
type BusinessTripService interface {
	// Create 创建出差申请
	Create(ctx context.Context, trip *model.BusinessTrip) error

	// GetByID 根据ID获取出差记录
	GetByID(ctx context.Context, id uuid.UUID) (*model.BusinessTrip, error)

	// Update 更新出差申请
	Update(ctx context.Context, trip *model.BusinessTrip) error

	// Delete 删除出差记录
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询
	List(ctx context.Context, tenantID uuid.UUID, filter *repository.BusinessTripFilter, offset, limit int) ([]*model.BusinessTrip, int, error)

	// ListByEmployee 查询员工出差记录
	ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.BusinessTrip, error)

	// ListPending 查询待审批的出差
	ListPending(ctx context.Context, tenantID uuid.UUID) ([]*model.BusinessTrip, error)

	// Submit 提交出差审批
	Submit(ctx context.Context, tripID, submitterID uuid.UUID) error

	// Approve 批准出差
	Approve(ctx context.Context, tripID, approverID uuid.UUID, comment string) error

	// Reject 拒绝出差
	Reject(ctx context.Context, tripID, approverID uuid.UUID, reason string) error

	// SubmitReport 提交出差报告
	SubmitReport(ctx context.Context, tripID uuid.UUID, report string, actualCost float64) error

	// SumDaysByEmployee 统计员工出差天数
	SumDaysByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) (float64, error)

	// CheckTimeConflict 检查时间冲突
	CheckTimeConflict(ctx context.Context, tenantID, employeeID uuid.UUID, startTime, endTime time.Time, excludeID *uuid.UUID) (bool, error)
}

type businessTripService struct {
	db             *database.DB
	tripRepo       repository.BusinessTripRepository
	workflowEngine *integration.BusinessTripWorkflowEngine
}

// NewBusinessTripService 创建出差服务
func NewBusinessTripService(
	db *database.DB,
	tripRepo repository.BusinessTripRepository,
	workflowEngine *workflow.Engine,
) BusinessTripService {
	return &businessTripService{
		db:             db,
		tripRepo:       tripRepo,
		workflowEngine: integration.NewBusinessTripWorkflowEngine(workflowEngine),
	}
}

// Create 创建出差申请
func (s *businessTripService) Create(ctx context.Context, trip *model.BusinessTrip) error {
	// 1. 验证时间
	if err := s.validateTime(trip.StartTime, trip.EndTime); err != nil {
		return err
	}

	// 2. 检查时间冲突
	hasConflict, err := s.CheckTimeConflict(ctx, trip.TenantID, trip.EmployeeID, trip.StartTime, trip.EndTime, nil)
	if err != nil {
		return fmt.Errorf("failed to check time conflict: %w", err)
	}
	if hasConflict {
		return fmt.Errorf("time conflict: employee already has a business trip during this period")
	}

	// 3. 计算出差天数
	trip.Duration = trip.EndTime.Sub(trip.StartTime).Hours() / 24

	// 4. 设置默认值
	trip.ID = uuid.Must(uuid.NewV7())
	now := time.Now()
	trip.CreatedAt = now
	trip.UpdatedAt = now
	trip.ApprovalStatus = "pending" // 默认待审批
	if trip.EstimatedCost == 0 {
		trip.EstimatedCost = 0 // 默认预算为0
	}
	trip.ActualCost = 0 // 初始实际费用为0

	// 5. 创建出差申请
	return s.tripRepo.Create(ctx, trip)
}

// GetByID 根据ID获取出差记录
func (s *businessTripService) GetByID(ctx context.Context, id uuid.UUID) (*model.BusinessTrip, error) {
	return s.tripRepo.FindByID(ctx, id)
}

// Update 更新出差申请
func (s *businessTripService) Update(ctx context.Context, trip *model.BusinessTrip) error {
	// 只有待审批状态才能更新
	existing, err := s.tripRepo.FindByID(ctx, trip.ID)
	if err != nil {
		return fmt.Errorf("failed to get business trip: %w", err)
	}

	if existing.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending business trips can be updated")
	}

	// 验证时间
	if err := s.validateTime(trip.StartTime, trip.EndTime); err != nil {
		return err
	}

	// 检查时间冲突（排除当前记录）
	hasConflict, err := s.CheckTimeConflict(ctx, trip.TenantID, trip.EmployeeID, trip.StartTime, trip.EndTime, &trip.ID)
	if err != nil {
		return fmt.Errorf("failed to check time conflict: %w", err)
	}
	if hasConflict {
		return fmt.Errorf("time conflict: employee already has a business trip during this period")
	}

	// 重新计算天数
	trip.Duration = trip.EndTime.Sub(trip.StartTime).Hours() / 24
	trip.UpdatedAt = time.Now()

	return s.tripRepo.Update(ctx, trip)
}

// Delete 删除出差记录
func (s *businessTripService) Delete(ctx context.Context, id uuid.UUID) error {
	// 只有待审批状态才能删除
	existing, err := s.tripRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get business trip: %w", err)
	}

	if existing.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending business trips can be deleted")
	}

	return s.tripRepo.Delete(ctx, id)
}

// List 列表查询
func (s *businessTripService) List(ctx context.Context, tenantID uuid.UUID, filter *repository.BusinessTripFilter, offset, limit int) ([]*model.BusinessTrip, int, error) {
	return s.tripRepo.List(ctx, tenantID, filter, offset, limit)
}

// ListByEmployee 查询员工出差记录
func (s *businessTripService) ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.BusinessTrip, error) {
	return s.tripRepo.FindByEmployee(ctx, tenantID, employeeID, year)
}

// ListPending 查询待审批的出差
func (s *businessTripService) ListPending(ctx context.Context, tenantID uuid.UUID) ([]*model.BusinessTrip, error) {
	return s.tripRepo.FindPending(ctx, tenantID)
}

// Submit 提交出差审批
func (s *businessTripService) Submit(ctx context.Context, tripID, submitterID uuid.UUID) error {
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return fmt.Errorf("failed to get business trip: %w", err)
	}

	if trip.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending business trips can be submitted")
	}

	// 验证提前申请天数（至少提前1天）
	if time.Now().Add(24 * time.Hour).After(trip.StartTime) {
		return fmt.Errorf("business trip must be submitted at least 1 day in advance")
	}

	// 使用工作流引擎执行审批流程
	workflowID := fmt.Sprintf("business-trip-approval-%s", trip.TenantID.String())

	// 启动工作流执行
	_, err = s.workflowEngine.ExecuteBusinessTripApproval(
		ctx,
		workflowID,
		trip,
		submitterID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to start workflow execution: %w", err)
	}

	// 更新状态为待审批（已提交）
	trip.ApprovalStatus = "pending"
	trip.UpdatedAt = time.Now()

	return s.tripRepo.Update(ctx, trip)
}

// Approve 批准出差
func (s *businessTripService) Approve(ctx context.Context, tripID, approverID uuid.UUID, comment string) error {
	// 使用工作流引擎处理审批（待实现）
	// TODO: 获取execution ID，调用workflowEngine.HandleApproval
	// 目前保持原有逻辑

	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return fmt.Errorf("failed to get business trip: %w", err)
	}

	if trip.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending business trips can be approved")
	}

	// 使用事务处理审批
	return s.db.Transaction(ctx, func(tx pgx.Tx) error {
		now := time.Now()
		trip.ApprovalStatus = "approved"
		trip.ApprovedBy = &approverID
		trip.ApprovedAt = &now
		trip.UpdatedAt = now

		if err := s.tripRepo.Update(ctx, trip); err != nil {
			return fmt.Errorf("failed to update business trip status: %w", err)
		}

		return nil
	})
}

// Reject 拒绝出差
func (s *businessTripService) Reject(ctx context.Context, tripID, approverID uuid.UUID, reason string) error {
	// 使用工作流引擎处理拒绝（待实现）
	// TODO: 获取execution ID，调用workflowEngine.HandleApproval(approved=false)
	// 目前保持原有逻辑

	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return fmt.Errorf("failed to get business trip: %w", err)
	}

	if trip.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending business trips can be rejected")
	}

	// 使用事务处理拒绝
	return s.db.Transaction(ctx, func(tx pgx.Tx) error {
		now := time.Now()
		trip.ApprovalStatus = "rejected"
		trip.ApprovedBy = &approverID
		trip.ApprovedAt = &now
		trip.RejectReason = reason
		trip.UpdatedAt = now

		if err := s.tripRepo.Update(ctx, trip); err != nil {
			return fmt.Errorf("failed to update business trip status: %w", err)
		}

		return nil
	})
}

// SubmitReport 提交出差报告
func (s *businessTripService) SubmitReport(ctx context.Context, tripID uuid.UUID, report string, actualCost float64) error {
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return fmt.Errorf("failed to get business trip: %w", err)
	}

	if trip.ApprovalStatus != "approved" {
		return fmt.Errorf("only approved business trips can submit report")
	}

	// 验证出差是否已结束
	if time.Now().Before(trip.EndTime) {
		return fmt.Errorf("cannot submit report before trip ends")
	}

	now := time.Now()
	trip.Report = report
	trip.ReportAt = &now
	trip.ActualCost = actualCost
	trip.UpdatedAt = now

	return s.tripRepo.Update(ctx, trip)
}

// SumDaysByEmployee 统计员工出差天数
func (s *businessTripService) SumDaysByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	// 查询时间范围内的出差记录
	trips, err := s.tripRepo.FindByEmployee(ctx, tenantID, employeeID, startDate.Year())
	if err != nil {
		return 0, err
	}

	var totalDays float64
	for _, trip := range trips {
		// 只统计已批准的出差
		if trip.ApprovalStatus == "approved" {
			// 检查是否在统计范围内
			if (trip.StartTime.Equal(startDate) || trip.StartTime.After(startDate)) &&
				(trip.EndTime.Equal(endDate) || trip.EndTime.Before(endDate)) {
				totalDays += trip.Duration
			}
		}
	}

	return totalDays, nil
}

// CheckTimeConflict 检查时间冲突
func (s *businessTripService) CheckTimeConflict(ctx context.Context, tenantID, employeeID uuid.UUID, startTime, endTime time.Time, excludeID *uuid.UUID) (bool, error) {
	overlapping, err := s.tripRepo.FindOverlapping(ctx, tenantID, employeeID, startTime, endTime)
	if err != nil {
		return false, err
	}

	// 如果需要排除特定ID
	if excludeID != nil {
		for i, trip := range overlapping {
			if trip.ID == *excludeID {
				overlapping = append(overlapping[:i], overlapping[i+1:]...)
				break
			}
		}
	}

	return len(overlapping) > 0, nil
}

// validateTime 验证时间
func (s *businessTripService) validateTime(startTime, endTime time.Time) error {
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return fmt.Errorf("end time must be after start time")
	}

	if startTime.Before(time.Now()) {
		return fmt.Errorf("start time cannot be in the past")
	}

	return nil
}
