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

// LeaveService 请假服务接口
type LeaveService interface {
	// 请假类型管理
	CreateLeaveType(ctx context.Context, leaveType *model.LeaveType) error
	UpdateLeaveType(ctx context.Context, leaveType *model.LeaveType) error
	DeleteLeaveType(ctx context.Context, id uuid.UUID) error
	GetLeaveType(ctx context.Context, id uuid.UUID) (*model.LeaveType, error)
	ListLeaveTypes(ctx context.Context, tenantID uuid.UUID, filter *repository.LeaveTypeFilter, offset, limit int) ([]*model.LeaveType, int, error)
	ListActiveLeaveTypes(ctx context.Context, tenantID uuid.UUID) ([]*model.LeaveType, error)

	// 请假额度管理
	InitEmployeeQuota(ctx context.Context, tenantID, employeeID uuid.UUID, year int) error
	UpdateQuota(ctx context.Context, quota *model.LeaveQuota) error
	GetEmployeeQuotas(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.LeaveQuotaWithType, error)
	CalculateAnnualLeaveQuota(ctx context.Context, workYears int) float64

	// 请假申请
	CreateLeaveRequest(ctx context.Context, request *model.LeaveRequest) error
	UpdateLeaveRequest(ctx context.Context, request *model.LeaveRequest) error
	SubmitLeaveRequest(ctx context.Context, requestID, submitterID uuid.UUID) error
	WithdrawLeaveRequest(ctx context.Context, requestID, operatorID uuid.UUID) error
	CancelLeaveRequest(ctx context.Context, requestID, operatorID uuid.UUID, reason string) error
	GetLeaveRequest(ctx context.Context, requestID uuid.UUID) (*model.LeaveRequestWithApprovals, error)
	ListMyLeaveRequests(ctx context.Context, tenantID, employeeID uuid.UUID, filter *repository.LeaveRequestFilter, offset, limit int) ([]*model.LeaveRequest, int, error)
	ListLeaveRequests(ctx context.Context, tenantID uuid.UUID, filter *repository.LeaveRequestFilter, offset, limit int) ([]*model.LeaveRequest, int, error)
	ListPendingApprovals(ctx context.Context, tenantID, approverID uuid.UUID, offset, limit int) ([]*model.LeaveRequest, int, error)

	// 审批
	ApproveLeaveRequest(ctx context.Context, requestID, approverID uuid.UUID, comment string) error
	RejectLeaveRequest(ctx context.Context, requestID, approverID uuid.UUID, comment string) error
	GetApprovalHistory(ctx context.Context, requestID uuid.UUID) ([]*model.LeaveApproval, error)
}

type leaveService struct {
	db                *database.DB
	leaveTypeRepo     repository.LeaveTypeRepository
	leaveQuotaRepo    repository.LeaveQuotaRepository
	leaveRequestRepo  repository.LeaveRequestRepository
	leaveApprovalRepo repository.LeaveApprovalRepository
	workflowEngine    *integration.LeaveWorkflowEngine
}

// NewLeaveService 创建请假服务
func NewLeaveService(
	db *database.DB,
	leaveTypeRepo repository.LeaveTypeRepository,
	leaveQuotaRepo repository.LeaveQuotaRepository,
	leaveRequestRepo repository.LeaveRequestRepository,
	leaveApprovalRepo repository.LeaveApprovalRepository,
	workflowEngine *workflow.Engine,
) LeaveService {
	return &leaveService{
		db:                db,
		leaveTypeRepo:     leaveTypeRepo,
		leaveQuotaRepo:    leaveQuotaRepo,
		leaveRequestRepo:  leaveRequestRepo,
		leaveApprovalRepo: leaveApprovalRepo,
		workflowEngine:    integration.NewLeaveWorkflowEngine(workflowEngine),
	}
}

// CreateLeaveType 创建请假类型
func (s *leaveService) CreateLeaveType(ctx context.Context, leaveType *model.LeaveType) error {
	leaveType.ID = uuid.Must(uuid.NewV7())
	now := time.Now()
	leaveType.CreatedAt = now
	leaveType.UpdatedAt = now

	// 创建请假类型
	if err := s.leaveTypeRepo.Create(ctx, leaveType); err != nil {
		return err
	}

	// 为请假类型创建工作流
	workflowDef, err := s.workflowEngine.CreateLeaveApprovalWorkflow(leaveType)
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	// 注册工作流到引擎（忽略错误，工作流可后续创建）
	_ = workflowDef

	return nil
}

// UpdateLeaveType 更新请假类型
func (s *leaveService) UpdateLeaveType(ctx context.Context, leaveType *model.LeaveType) error {
	leaveType.UpdatedAt = time.Now()
	return s.leaveTypeRepo.Update(ctx, leaveType)
}

// DeleteLeaveType 删除请假类型
func (s *leaveService) DeleteLeaveType(ctx context.Context, id uuid.UUID) error {
	return s.leaveTypeRepo.Delete(ctx, id)
}

// GetLeaveType 获取请假类型详情
func (s *leaveService) GetLeaveType(ctx context.Context, id uuid.UUID) (*model.LeaveType, error) {
	return s.leaveTypeRepo.FindByID(ctx, id)
}

// ListLeaveTypes 列表查询请假类型
func (s *leaveService) ListLeaveTypes(ctx context.Context, tenantID uuid.UUID, filter *repository.LeaveTypeFilter, offset, limit int) ([]*model.LeaveType, int, error) {
	return s.leaveTypeRepo.List(ctx, tenantID, filter, offset, limit)
}

// ListActiveLeaveTypes 查询启用的请假类型
func (s *leaveService) ListActiveLeaveTypes(ctx context.Context, tenantID uuid.UUID) ([]*model.LeaveType, error) {
	return s.leaveTypeRepo.ListActive(ctx, tenantID)
}

// InitEmployeeQuota 初始化员工请假额度
func (s *leaveService) InitEmployeeQuota(ctx context.Context, tenantID, employeeID uuid.UUID, year int) error {
	// 获取所有启用的请假类型
	leaveTypes, err := s.leaveTypeRepo.ListActive(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to list leave types: %w", err)
	}

	// 为每种请假类型创建额度
	quotas := make([]*model.LeaveQuota, 0, len(leaveTypes))
	now := time.Now()

	for _, lt := range leaveTypes {
		// 只为需要扣除额度的类型创建额度记录
		if !lt.DeductQuota {
			continue
		}

		quota := &model.LeaveQuota{
			ID:           uuid.Must(uuid.NewV7()),
			TenantID:     tenantID,
			EmployeeID:   employeeID,
			LeaveTypeID:  lt.ID,
			Year:         year,
			TotalQuota:   s.getDefaultQuota(lt.Code),
			UsedQuota:    0,
			PendingQuota: 0,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		// 设置过期时间（次年12月31日）
		expiredAt := time.Date(year+1, 12, 31, 23, 59, 59, 0, time.Local)
		quota.ExpiredAt = &expiredAt

		quotas = append(quotas, quota)
	}

	if len(quotas) > 0 {
		return s.leaveQuotaRepo.BatchCreate(ctx, quotas)
	}

	return nil
}

// getDefaultQuota 获取默认额度
func (s *leaveService) getDefaultQuota(code string) float64 {
	switch code {
	case "annual_leave":
		return 5.0 // 默认5天年假
	case "sick_leave":
		return 10.0 // 默认10天病假
	case "compensatory_leave":
		return 0.0 // 调休初始为0
	case "marriage_leave":
		return 3.0 // 婚假3天
	case "maternity_leave":
		return 98.0 // 产假98天
	case "paternity_leave":
		return 15.0 // 陪产假15天
	default:
		return 0.0
	}
}

// UpdateQuota 更新请假额度
func (s *leaveService) UpdateQuota(ctx context.Context, quota *model.LeaveQuota) error {
	quota.UpdatedAt = time.Now()
	return s.leaveQuotaRepo.Update(ctx, quota)
}

// GetEmployeeQuotas 获取员工的所有假期额度
func (s *leaveService) GetEmployeeQuotas(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.LeaveQuotaWithType, error) {
	return s.leaveQuotaRepo.ListByEmployeeWithType(ctx, tenantID, employeeID, year)
}

// CalculateAnnualLeaveQuota 计算年假额度（根据工龄）
func (s *leaveService) CalculateAnnualLeaveQuota(ctx context.Context, workYears int) float64 {
	// 根据工龄计算年假
	// 1年以下：0天
	// 1-10年：5天
	// 10-20年：10天
	// 20年以上：15天
	if workYears < 1 {
		return 0
	} else if workYears < 10 {
		return 5
	} else if workYears < 20 {
		return 10
	}
	return 15
}

// CreateLeaveRequest 创建请假申请
func (s *leaveService) CreateLeaveRequest(ctx context.Context, request *model.LeaveRequest) error {
	request.ID = uuid.Must(uuid.NewV7())
	now := time.Now()
	request.CreatedAt = now
	request.UpdatedAt = now
	request.Status = model.LeaveRequestStatusDraft

	// 验证请假时间
	if request.StartTime.After(request.EndTime) {
		return fmt.Errorf("start time must be before end time")
	}

	// 检查时间冲突
	hasConflict, err := s.leaveRequestRepo.CheckTimeConflict(ctx, request.TenantID, request.EmployeeID, request.StartTime, request.EndTime, nil)
	if err != nil {
		return fmt.Errorf("failed to check time conflict: %w", err)
	}
	if hasConflict {
		return fmt.Errorf("leave time conflicts with existing approved leave")
	}

	// 获取请假类型
	leaveType, err := s.leaveTypeRepo.FindByID(ctx, request.LeaveTypeID)
	if err != nil {
		return fmt.Errorf("failed to get leave type: %w", err)
	}

	request.LeaveTypeName = leaveType.Name
	request.Unit = leaveType.Unit

	// 如果需要扣除额度，检查额度是否足够
	if leaveType.DeductQuota {
		year := request.StartTime.Year()
		quota, err := s.leaveQuotaRepo.FindByEmployeeAndType(ctx, request.TenantID, request.EmployeeID, request.LeaveTypeID, year)
		if err != nil {
			return fmt.Errorf("failed to get leave quota: %w", err)
		}

		remaining := quota.RemainingQuota()
		if remaining < request.Duration {
			return fmt.Errorf("insufficient leave quota: remaining %.1f %s, requested %.1f %s", remaining, request.Unit, request.Duration, request.Unit)
		}
	}

	return s.leaveRequestRepo.Create(ctx, request)
}

// UpdateLeaveRequest 更新请假申请
func (s *leaveService) UpdateLeaveRequest(ctx context.Context, request *model.LeaveRequest) error {
	// 只有草稿状态才能更新
	existing, err := s.leaveRequestRepo.FindByID(ctx, request.ID)
	if err != nil {
		return fmt.Errorf("failed to get leave request: %w", err)
	}

	if existing.Status != model.LeaveRequestStatusDraft {
		return fmt.Errorf("only draft leave requests can be updated")
	}

	request.UpdatedAt = time.Now()
	return s.leaveRequestRepo.Update(ctx, request)
}

// SubmitLeaveRequest 提交请假申请
func (s *leaveService) SubmitLeaveRequest(ctx context.Context, requestID, submitterID uuid.UUID) error {
	// 获取请假申请
	request, err := s.leaveRequestRepo.FindByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get leave request: %w", err)
	}

	// 验证状态
	if request.Status != model.LeaveRequestStatusDraft {
		return fmt.Errorf("only draft leave requests can be submitted")
	}

	// 获取请假类型
	leaveType, err := s.leaveTypeRepo.FindByID(ctx, request.LeaveTypeID)
	if err != nil {
		return fmt.Errorf("failed to get leave type: %w", err)
	}

	// 使用工作流引擎执行审批流程
	workflowID := fmt.Sprintf("leave-approval-%s", leaveType.ID.String())

	// 启动工作流执行
	_, err = s.workflowEngine.ExecuteLeaveApproval(
		ctx,
		workflowID,
		request,
		submitterID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to start workflow execution: %w", err)
	}

	// 保存executionID并更新状态为Pending
	now := time.Now()
	if err := s.leaveRequestRepo.UpdateStatus(ctx, requestID, model.LeaveRequestStatusPending, &now); err != nil {
		return fmt.Errorf("failed to update leave request status: %w", err)
	}

	// 如果需要扣减额度，增加待审批额度
	if leaveType.DeductQuota {
		year := request.StartTime.Year()
		quota, err := s.leaveQuotaRepo.FindByEmployeeAndType(ctx, request.TenantID, request.EmployeeID, request.LeaveTypeID, year)
		if err != nil {
			return fmt.Errorf("failed to get quota: %w", err)
		}
		if err := s.leaveQuotaRepo.IncrementPendingQuota(ctx, quota.ID, request.Duration); err != nil {
			return fmt.Errorf("failed to increment pending quota: %w", err)
		}
	}

	return nil
}

// WithdrawLeaveRequest 撤回请假申请
func (s *leaveService) WithdrawLeaveRequest(ctx context.Context, requestID, operatorID uuid.UUID) error {
	request, err := s.leaveRequestRepo.FindByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get leave request: %w", err)
	}

	// 只有草稿和待审批状态才能撤回
	if request.Status != model.LeaveRequestStatusDraft && request.Status != model.LeaveRequestStatusPending {
		return fmt.Errorf("only draft or pending leave requests can be withdrawn")
	}

	// 如果是草稿状态，直接更新状态即可
	if request.Status == model.LeaveRequestStatusDraft {
		now := time.Now()
		return s.leaveRequestRepo.UpdateStatus(ctx, requestID, model.LeaveRequestStatusWithdrawn, &now)
	}

	// 待审批状态：使用事务（需要同时更新状态和额度）
	return s.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 更新状态
		now := time.Now()
		if err := s.leaveRequestRepo.UpdateStatus(ctx, requestID, model.LeaveRequestStatusWithdrawn, &now); err != nil {
			return fmt.Errorf("failed to withdraw leave request: %w", err)
		}

		// 减少待审批额度
		leaveType, err := s.leaveTypeRepo.FindByID(ctx, request.LeaveTypeID)
		if err != nil {
			return fmt.Errorf("failed to get leave type: %w", err)
		}

		if leaveType.DeductQuota {
			year := request.StartTime.Year()
			quota, err := s.leaveQuotaRepo.FindByEmployeeAndType(ctx, request.TenantID, request.EmployeeID, request.LeaveTypeID, year)
			if err != nil {
				return fmt.Errorf("failed to get quota: %w", err)
			}
			if err := s.leaveQuotaRepo.DecrementPendingQuota(ctx, quota.ID, request.Duration); err != nil {
				return fmt.Errorf("failed to decrement pending quota: %w", err)
			}
		}

		return nil
	})
}

// CancelLeaveRequest 取消已批准的请假
func (s *leaveService) CancelLeaveRequest(ctx context.Context, requestID, operatorID uuid.UUID, reason string) error {
	request, err := s.leaveRequestRepo.FindByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get leave request: %w", err)
	}

	// 只有已批准的才能取消
	if request.Status != model.LeaveRequestStatusApproved {
		return fmt.Errorf("only approved leave requests can be cancelled")
	}

	// 不能取消已经开始的请假
	if time.Now().After(request.StartTime) {
		return fmt.Errorf("cannot cancel leave that has already started")
	}

	// 使用事务（需要同时更新状态和退还额度）
	return s.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 更新状态
		now := time.Now()
		if err := s.leaveRequestRepo.UpdateStatus(ctx, requestID, model.LeaveRequestStatusCancelled, &now); err != nil {
			return fmt.Errorf("failed to cancel leave request: %w", err)
		}

		// 退还已使用额度
		leaveType, err := s.leaveTypeRepo.FindByID(ctx, request.LeaveTypeID)
		if err != nil {
			return fmt.Errorf("failed to get leave type: %w", err)
		}

		if leaveType.DeductQuota {
			year := request.StartTime.Year()
			quota, err := s.leaveQuotaRepo.FindByEmployeeAndType(ctx, request.TenantID, request.EmployeeID, request.LeaveTypeID, year)
			if err != nil {
				return fmt.Errorf("failed to get quota: %w", err)
			}
			if err := s.leaveQuotaRepo.DecrementUsedQuota(ctx, quota.ID, request.Duration); err != nil {
				return fmt.Errorf("failed to decrement used quota: %w", err)
			}
		}

		return nil
	})
}

// GetLeaveRequest 获取请假申请详情
func (s *leaveService) GetLeaveRequest(ctx context.Context, requestID uuid.UUID) (*model.LeaveRequestWithApprovals, error) {
	return s.leaveRequestRepo.FindByIDWithApprovals(ctx, requestID)
}

// ListMyLeaveRequests 查询我的请假记录
func (s *leaveService) ListMyLeaveRequests(ctx context.Context, tenantID, employeeID uuid.UUID, filter *repository.LeaveRequestFilter, offset, limit int) ([]*model.LeaveRequest, int, error) {
	return s.leaveRequestRepo.ListByEmployee(ctx, tenantID, employeeID, filter, offset, limit)
}

// ListLeaveRequests 查询请假记录（管理员/HR）
func (s *leaveService) ListLeaveRequests(ctx context.Context, tenantID uuid.UUID, filter *repository.LeaveRequestFilter, offset, limit int) ([]*model.LeaveRequest, int, error) {
	return s.leaveRequestRepo.List(ctx, tenantID, filter, offset, limit)
}

// ListPendingApprovals 查询待我审批的请假
func (s *leaveService) ListPendingApprovals(ctx context.Context, tenantID, approverID uuid.UUID, offset, limit int) ([]*model.LeaveRequest, int, error) {
	return s.leaveRequestRepo.ListPendingApprovals(ctx, tenantID, approverID, offset, limit)
}

// ApproveLeaveRequest 批准请假
func (s *leaveService) ApproveLeaveRequest(ctx context.Context, requestID, approverID uuid.UUID, comment string) error {
	// 使用工作流引擎处理审批（待实现）
	// TODO: 获取execution ID，调用workflowEngine.HandleApproval
	// 目前保持原有逻辑

	// 获取请假申请
	request, err := s.leaveRequestRepo.FindByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get leave request: %w", err)
	}

	// 验证状态
	if request.Status != model.LeaveRequestStatusPending {
		return fmt.Errorf("only pending leave requests can be approved")
	}

	// 查找待审批记录
	approval, err := s.leaveApprovalRepo.FindPendingApproval(ctx, requestID, approverID)
	if err != nil {
		return fmt.Errorf("no pending approval found for this approver: %w", err)
	}

	// 使用事务处理审批流程
	return s.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 更新审批记录
		now := time.Now()
		action := model.LeaveApprovalActionApprove
		if err := s.leaveApprovalRepo.UpdateStatus(ctx, approval.ID, model.LeaveApprovalStatusApproved, &action, comment, &now); err != nil {
			return fmt.Errorf("failed to update approval status: %w", err)
		}

		// 获取所有审批记录，判断是否还有下一级
		allApprovals, err := s.leaveApprovalRepo.ListByRequest(ctx, requestID)
		if err != nil {
			return fmt.Errorf("failed to get approval history: %w", err)
		}

		// 查找下一级待审批的记录
		var nextApproval *model.LeaveApproval
		for _, ap := range allApprovals {
			if ap.Level > approval.Level && ap.Status == model.LeaveApprovalStatusPending {
				if nextApproval == nil || ap.Level < nextApproval.Level {
					nextApproval = ap
				}
			}
		}

		if nextApproval != nil {
			// 还有下一级审批，更新当前审批人
			if err := s.leaveRequestRepo.SetCurrentApprover(ctx, requestID, &nextApproval.ApproverID); err != nil {
				return fmt.Errorf("failed to set next approver: %w", err)
			}
			return nil
		}

		// 所有审批完成，更新请假申请状态为已批准
		if err := s.leaveRequestRepo.UpdateStatus(ctx, requestID, model.LeaveRequestStatusApproved, &now); err != nil {
			return fmt.Errorf("failed to approve leave request: %w", err)
		}

		// 清除当前审批人
		if err := s.leaveRequestRepo.SetCurrentApprover(ctx, requestID, nil); err != nil {
			return fmt.Errorf("failed to clear current approver: %w", err)
		}

		// 额度处理：从待审批转为已使用
		leaveType, err := s.leaveTypeRepo.FindByID(ctx, request.LeaveTypeID)
		if err != nil {
			return fmt.Errorf("failed to get leave type: %w", err)
		}

		if leaveType.DeductQuota {
			year := request.StartTime.Year()
			quota, err := s.leaveQuotaRepo.FindByEmployeeAndType(ctx, request.TenantID, request.EmployeeID, request.LeaveTypeID, year)
			if err != nil {
				return fmt.Errorf("failed to get quota: %w", err)
			}
			if err := s.leaveQuotaRepo.DecrementPendingQuota(ctx, quota.ID, request.Duration); err != nil {
				return fmt.Errorf("failed to decrement pending quota: %w", err)
			}
			if err := s.leaveQuotaRepo.IncrementUsedQuota(ctx, quota.ID, request.Duration); err != nil {
				return fmt.Errorf("failed to increment used quota: %w", err)
			}
		}

		return nil
	})
}

// RejectLeaveRequest 拒绝请假
func (s *leaveService) RejectLeaveRequest(ctx context.Context, requestID, approverID uuid.UUID, comment string) error {
	// 使用工作流引擎处理拒绝（待实现）
	// TODO: 获取execution ID，调用workflowEngine.HandleApproval(approved=false)
	// 目前保持原有逻辑

	// 获取请假申请
	request, err := s.leaveRequestRepo.FindByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get leave request: %w", err)
	}

	// 验证状态
	if request.Status != model.LeaveRequestStatusPending {
		return fmt.Errorf("only pending leave requests can be rejected")
	}

	// 查找待审批记录
	approval, err := s.leaveApprovalRepo.FindPendingApproval(ctx, requestID, approverID)
	if err != nil {
		return fmt.Errorf("no pending approval found for this approver: %w", err)
	}

	// 使用事务处理拒绝流程
	return s.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 更新审批记录
		now := time.Now()
		action := model.LeaveApprovalActionReject
		if err := s.leaveApprovalRepo.UpdateStatus(ctx, approval.ID, model.LeaveApprovalStatusRejected, &action, comment, &now); err != nil {
			return fmt.Errorf("failed to update approval status: %w", err)
		}

		// 更新请假申请状态为已拒绝
		if err := s.leaveRequestRepo.UpdateStatus(ctx, requestID, model.LeaveRequestStatusRejected, &now); err != nil {
			return fmt.Errorf("failed to reject leave request: %w", err)
		}

		// 清除当前审批人
		if err := s.leaveRequestRepo.SetCurrentApprover(ctx, requestID, nil); err != nil {
			return fmt.Errorf("failed to clear current approver: %w", err)
		}

		// 减少待审批额度
		leaveType, err := s.leaveTypeRepo.FindByID(ctx, request.LeaveTypeID)
		if err != nil {
			return fmt.Errorf("failed to get leave type: %w", err)
		}

		if leaveType.DeductQuota {
			year := request.StartTime.Year()
			quota, err := s.leaveQuotaRepo.FindByEmployeeAndType(ctx, request.TenantID, request.EmployeeID, request.LeaveTypeID, year)
			if err != nil {
				return fmt.Errorf("failed to get quota: %w", err)
			}
			if err := s.leaveQuotaRepo.DecrementPendingQuota(ctx, quota.ID, request.Duration); err != nil {
				return fmt.Errorf("failed to decrement pending quota: %w", err)
			}
		}

		return nil
	})
}

// GetApprovalHistory 获取审批历史
func (s *leaveService) GetApprovalHistory(ctx context.Context, requestID uuid.UUID) ([]*model.LeaveApproval, error) {
	return s.leaveApprovalRepo.ListByRequest(ctx, requestID)
}
