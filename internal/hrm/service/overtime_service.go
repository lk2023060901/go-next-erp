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

// OvertimeService 加班服务接口
type OvertimeService interface {
	// Create 创建加班申请
	Create(ctx context.Context, overtime *model.Overtime) error

	// GetByID 根据ID获取加班记录
	GetByID(ctx context.Context, id uuid.UUID) (*model.Overtime, error)

	// Update 更新加班申请
	Update(ctx context.Context, overtime *model.Overtime) error

	// Delete 删除加班记录
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询
	List(ctx context.Context, tenantID uuid.UUID, filter *repository.OvertimeFilter, offset, limit int) ([]*model.Overtime, int, error)

	// ListByEmployee 查询员工加班记录
	ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.Overtime, error)

	// ListPending 查询待审批的加班
	ListPending(ctx context.Context, tenantID uuid.UUID) ([]*model.Overtime, error)

	// Submit 提交加班审批
	Submit(ctx context.Context, overtimeID, submitterID uuid.UUID) error

	// Approve 批准加班
	Approve(ctx context.Context, overtimeID, approverID uuid.UUID) error

	// Reject 拒绝加班
	Reject(ctx context.Context, overtimeID, approverID uuid.UUID, reason string) error

	// SumHoursByEmployee 统计员工加班时长
	SumHoursByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) (float64, error)

	// SumCompOffDays 统计可调休天数
	SumCompOffDays(ctx context.Context, tenantID, employeeID uuid.UUID) (float64, error)

	// UseCompOffDays 使用调休
	UseCompOffDays(ctx context.Context, tenantID, employeeID uuid.UUID, days float64) error
}

type overtimeService struct {
	db             *database.DB
	overtimeRepo   repository.OvertimeRepository
	workflowEngine *integration.OvertimeWorkflowEngine
}

// NewOvertimeService 创建加班服务
func NewOvertimeService(
	db *database.DB,
	overtimeRepo repository.OvertimeRepository,
	workflowEngine *workflow.Engine,
) OvertimeService {
	return &overtimeService{
		db:             db,
		overtimeRepo:   overtimeRepo,
		workflowEngine: integration.NewOvertimeWorkflowEngine(workflowEngine),
	}
}

// Create 创建加班申请
func (s *overtimeService) Create(ctx context.Context, overtime *model.Overtime) error {
	overtime.ID = uuid.Must(uuid.NewV7())
	now := time.Now()
	overtime.CreatedAt = now
	overtime.UpdatedAt = now
	overtime.ApprovalStatus = "pending" // 默认待审批
	overtime.CompOffDays = 0
	overtime.CompOffUsed = 0

	// 根据加班类型计算倍率
	if overtime.PayRate == 0 {
		overtime.PayRate = s.calculatePayRate(overtime.OvertimeType)
	}

	// 如果是调休类型，计算可调休天数
	if overtime.PayType == "leave" {
		overtime.CompOffDays = s.calculateCompOffDays(overtime.Duration, overtime.OvertimeType)
	}

	return s.overtimeRepo.Create(ctx, overtime)
}

// calculatePayRate 计算加班倍率
func (s *overtimeService) calculatePayRate(overtimeType model.OvertimeType) float64 {
	switch overtimeType {
	case model.OvertimeTypeWorkday:
		return 1.5 // 工作日加班1.5倍
	case model.OvertimeTypeWeekend:
		return 2.0 // 周末加班2倍
	case model.OvertimeTypeHoliday:
		return 3.0 // 节假日加班3倍
	default:
		return 1.5
	}
}

// calculateCompOffDays 计算可调休天数
func (s *overtimeService) calculateCompOffDays(hours float64, overtimeType model.OvertimeType) float64 {
	// 8小时=1天
	// 工作日加班：按实际时长
	// 周末/节假日加班：按倍率计算
	days := hours / 8.0

	switch overtimeType {
	case model.OvertimeTypeWeekend:
		return days * 2.0 // 周末加班可调休2倍
	case model.OvertimeTypeHoliday:
		return days * 3.0 // 节假日加班可调休3倍
	default:
		return days // 工作日按实际时长
	}
}

// GetByID 根据ID获取加班记录
func (s *overtimeService) GetByID(ctx context.Context, id uuid.UUID) (*model.Overtime, error) {
	return s.overtimeRepo.FindByID(ctx, id)
}

// Update 更新加班申请
func (s *overtimeService) Update(ctx context.Context, overtime *model.Overtime) error {
	// 只有待审批状态才能更新
	existing, err := s.overtimeRepo.FindByID(ctx, overtime.ID)
	if err != nil {
		return fmt.Errorf("failed to get overtime: %w", err)
	}

	if existing.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending overtime can be updated")
	}

	overtime.UpdatedAt = time.Now()

	// 重新计算倍率和调休天数
	if overtime.PayRate == 0 {
		overtime.PayRate = s.calculatePayRate(overtime.OvertimeType)
	}

	if overtime.PayType == "leave" {
		overtime.CompOffDays = s.calculateCompOffDays(overtime.Duration, overtime.OvertimeType)
	}

	return s.overtimeRepo.Update(ctx, overtime)
}

// Delete 删除加班记录
func (s *overtimeService) Delete(ctx context.Context, id uuid.UUID) error {
	// 只有待审批状态才能删除
	existing, err := s.overtimeRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get overtime: %w", err)
	}

	if existing.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending overtime can be deleted")
	}

	return s.overtimeRepo.Delete(ctx, id)
}

// List 列表查询
func (s *overtimeService) List(ctx context.Context, tenantID uuid.UUID, filter *repository.OvertimeFilter, offset, limit int) ([]*model.Overtime, int, error) {
	return s.overtimeRepo.List(ctx, tenantID, filter, offset, limit)
}

// ListByEmployee 查询员工加班记录
func (s *overtimeService) ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.Overtime, error) {
	return s.overtimeRepo.FindByEmployee(ctx, tenantID, employeeID, year)
}

// ListPending 查询待审批的加班
func (s *overtimeService) ListPending(ctx context.Context, tenantID uuid.UUID) ([]*model.Overtime, error) {
	return s.overtimeRepo.FindPending(ctx, tenantID)
}

// Submit 提交加班审批
func (s *overtimeService) Submit(ctx context.Context, overtimeID, submitterID uuid.UUID) error {
	overtime, err := s.overtimeRepo.FindByID(ctx, overtimeID)
	if err != nil {
		return fmt.Errorf("failed to get overtime: %w", err)
	}

	if overtime.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending overtime can be submitted")
	}

	// 使用工作流引擎执行审批流程
	workflowID := fmt.Sprintf("overtime-approval-%s", overtime.TenantID.String())

	// 启动工作流执行
	_, err = s.workflowEngine.ExecuteOvertimeApproval(
		ctx,
		workflowID,
		overtime,
		submitterID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to start workflow execution: %w", err)
	}

	// 更新状态为待审批
	overtime.ApprovalStatus = "pending"
	overtime.UpdatedAt = time.Now()

	return s.overtimeRepo.Update(ctx, overtime)
}

// Approve 批准加班（使用事务）
func (s *overtimeService) Approve(ctx context.Context, overtimeID, approverID uuid.UUID) error {
	// 使用工作流引擎处理审批（待实现）
	// TODO: 获取execution ID，调用workflowEngine.HandleApproval
	// 目前保持原有逻辑

	overtime, err := s.overtimeRepo.FindByID(ctx, overtimeID)
	if err != nil {
		return fmt.Errorf("failed to get overtime: %w", err)
	}

	if overtime.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending overtime can be approved")
	}

	// 使用事务处理审批
	return s.db.Transaction(ctx, func(tx pgx.Tx) error {
		now := time.Now()
		overtime.ApprovalStatus = "approved"
		overtime.ApprovedBy = &approverID
		overtime.ApprovedAt = &now
		overtime.UpdatedAt = now

		// 如果是调休类型，设置过期时间（1年后）
		if overtime.PayType == "leave" {
			expireAt := now.AddDate(1, 0, 0) // 1年后过期
			overtime.CompOffExpireAt = &expireAt
		}

		if err := s.overtimeRepo.Update(ctx, overtime); err != nil {
			return fmt.Errorf("failed to update overtime status: %w", err)
		}

		return nil
	})
}

// Reject 拒绝加班（使用事务）
func (s *overtimeService) Reject(ctx context.Context, overtimeID, approverID uuid.UUID, reason string) error {
	// 使用工作流引擎处理拒绝（待实现）
	// TODO: 获取execution ID，调用workflowEngine.HandleApproval(approved=false)
	// 目前保持原有逻辑

	overtime, err := s.overtimeRepo.FindByID(ctx, overtimeID)
	if err != nil {
		return fmt.Errorf("failed to get overtime: %w", err)
	}

	if overtime.ApprovalStatus != "pending" {
		return fmt.Errorf("only pending overtime can be rejected")
	}

	// 使用事务处理拒绝
	return s.db.Transaction(ctx, func(tx pgx.Tx) error {
		now := time.Now()
		overtime.ApprovalStatus = "rejected"
		overtime.ApprovedBy = &approverID
		overtime.ApprovedAt = &now
		overtime.RejectReason = reason
		overtime.UpdatedAt = now

		if err := s.overtimeRepo.Update(ctx, overtime); err != nil {
			return fmt.Errorf("failed to update overtime status: %w", err)
		}

		return nil
	})
}

// SumHoursByEmployee 统计员工加班时长
func (s *overtimeService) SumHoursByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	return s.overtimeRepo.SumHoursByEmployee(ctx, tenantID, employeeID, startDate, endDate)
}

// SumCompOffDays 统计可调休天数
func (s *overtimeService) SumCompOffDays(ctx context.Context, tenantID, employeeID uuid.UUID) (float64, error) {
	return s.overtimeRepo.SumCompOffDays(ctx, tenantID, employeeID)
}

// UseCompOffDays 使用调休（使用事务）
func (s *overtimeService) UseCompOffDays(ctx context.Context, tenantID, employeeID uuid.UUID, days float64) error {
	if days <= 0 {
		return fmt.Errorf("invalid days: %f", days)
	}

	// 使用事务确保数据一致性
	return s.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 查询可用的调休记录（按过期时间排序，优先使用即将过期的）
		query := `
			SELECT id, comp_off_days, comp_off_used
			FROM hrm_overtimes
			WHERE tenant_id = $1 AND employee_id = $2
				AND approval_status = 'approved'
				AND pay_type = 'leave'
				AND (comp_off_days - comp_off_used) > 0
				AND (comp_off_expire_at IS NULL OR comp_off_expire_at > NOW())
				AND deleted_at IS NULL
			ORDER BY comp_off_expire_at NULLS LAST
		`

		rows, err := s.db.Query(ctx, query, tenantID, employeeID)
		if err != nil {
			return fmt.Errorf("failed to query comp off records: %w", err)
		}
		defer rows.Close()

		remainingDays := days
		for rows.Next() && remainingDays > 0 {
			var id uuid.UUID
			var compOffDays, compOffUsed float64

			if err := rows.Scan(&id, &compOffDays, &compOffUsed); err != nil {
				return fmt.Errorf("failed to scan comp off record: %w", err)
			}

			available := compOffDays - compOffUsed
			if available <= 0 {
				continue
			}

			// 计算本次使用的天数
			useDays := remainingDays
			if useDays > available {
				useDays = available
			}

			// 更新已使用天数
			updateQuery := `
				UPDATE hrm_overtimes
				SET comp_off_used = comp_off_used + $1, updated_at = $2
				WHERE id = $3
			`
			_, err := s.db.Exec(ctx, updateQuery, useDays, time.Now(), id)
			if err != nil {
				return fmt.Errorf("failed to update comp off used: %w", err)
			}

			remainingDays -= useDays
		}

		if remainingDays > 0 {
			return fmt.Errorf("insufficient comp off days: need %f, available %f", days, days-remainingDays)
		}

		return nil
	})
}
