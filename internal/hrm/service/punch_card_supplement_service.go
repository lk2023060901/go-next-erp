package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
)

// PunchCardSupplementService 补卡申请服务接口
type PunchCardSupplementService interface {
	// Create 创建补卡申请
	Create(ctx context.Context, supplement *model.PunchCardSupplement) error

	// Update 更新补卡申请
	Update(ctx context.Context, supplement *model.PunchCardSupplement) error

	// Delete 删除补卡申请
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据ID获取补卡申请
	GetByID(ctx context.Context, id uuid.UUID) (*model.PunchCardSupplement, error)

	// List 列表查询
	List(ctx context.Context, tenantID uuid.UUID, filter *repository.PunchCardSupplementFilter, offset, limit int) ([]*model.PunchCardSupplement, int, error)

	// ListByEmployee 查询员工补卡申请记录
	ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.PunchCardSupplement, error)

	// ListPending 查询待审批的补卡申请
	ListPending(ctx context.Context, tenantID uuid.UUID) ([]*model.PunchCardSupplement, error)

	// Submit 提交补卡申请审批
	Submit(ctx context.Context, supplementID, submitterID uuid.UUID) error

	// Approve 批准补卡申请
	Approve(ctx context.Context, supplementID, approverID uuid.UUID, comment string) error

	// Reject 拒绝补卡申请
	Reject(ctx context.Context, supplementID, approverID uuid.UUID, reason string) error

	// Process 处理补卡（生成补卡记录）
	Process(ctx context.Context, supplementID, processorID uuid.UUID) error

	// Cancel 取消补卡申请
	Cancel(ctx context.Context, supplementID uuid.UUID) error

	// ValidateSupplement 验证补卡申请的合理性
	ValidateSupplement(ctx context.Context, supplement *model.PunchCardSupplement) error

	// CheckDuplicate 检查是否存在重复的补卡申请
	CheckDuplicate(ctx context.Context, tenantID, employeeID uuid.UUID, date time.Time, supplementType model.SupplementType) (bool, error)

	// CountByEmployee 统计员工补卡次数
	CountByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) (int, error)
}

// punchCardSupplementService 补卡申请服务实现
type punchCardSupplementService struct {
	repo repository.PunchCardSupplementRepository
}

// NewPunchCardSupplementService 创建补卡申请服务
func NewPunchCardSupplementService(repo repository.PunchCardSupplementRepository) PunchCardSupplementService {
	return &punchCardSupplementService{
		repo: repo,
	}
}

// Create 创建补卡申请
func (s *punchCardSupplementService) Create(ctx context.Context, supplement *model.PunchCardSupplement) error {
	// 1. 验证补卡申请的合理性
	if err := s.ValidateSupplement(ctx, supplement); err != nil {
		return err
	}

	// 2. 检查是否存在重复的补卡申请
	duplicate, err := s.CheckDuplicate(ctx, supplement.TenantID, supplement.EmployeeID, supplement.SupplementDate, supplement.SupplementType)
	if err != nil {
		return err
	}
	if duplicate {
		return fmt.Errorf("duplicate supplement application exists for this date and type")
	}

	// 3. 设置初始状态
	supplement.ApprovalStatus = "pending"
	supplement.ProcessStatus = "pending"

	// 4. 创建补卡申请
	return s.repo.Create(ctx, supplement)
}

// Update 更新补卡申请
func (s *punchCardSupplementService) Update(ctx context.Context, supplement *model.PunchCardSupplement) error {
	// 1. 获取原记录
	existing, err := s.repo.FindByID(ctx, supplement.ID)
	if err != nil {
		return err
	}

	// 2. 检查状态（只有pending状态的申请可以修改）
	if existing.ApprovalStatus != "pending" {
		return fmt.Errorf("cannot update supplement with status: %s", existing.ApprovalStatus)
	}

	// 3. 验证补卡申请的合理性
	if err := s.ValidateSupplement(ctx, supplement); err != nil {
		return err
	}

	// 4. 更新补卡申请
	return s.repo.Update(ctx, supplement)
}

// Delete 删除补卡申请
func (s *punchCardSupplementService) Delete(ctx context.Context, id uuid.UUID) error {
	// 1. 获取原记录
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. 检查状态（只有pending状态的申请可以删除）
	if existing.ApprovalStatus != "pending" {
		return fmt.Errorf("cannot delete supplement with status: %s", existing.ApprovalStatus)
	}

	// 3. 删除补卡申请
	return s.repo.Delete(ctx, id)
}

// GetByID 根据ID获取补卡申请
func (s *punchCardSupplementService) GetByID(ctx context.Context, id uuid.UUID) (*model.PunchCardSupplement, error) {
	return s.repo.FindByID(ctx, id)
}

// List 列表查询
func (s *punchCardSupplementService) List(ctx context.Context, tenantID uuid.UUID, filter *repository.PunchCardSupplementFilter, offset, limit int) ([]*model.PunchCardSupplement, int, error) {
	return s.repo.List(ctx, tenantID, filter, offset, limit)
}

// ListByEmployee 查询员工补卡申请记录
func (s *punchCardSupplementService) ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.PunchCardSupplement, error) {
	return s.repo.FindByEmployee(ctx, tenantID, employeeID, year)
}

// ListPending 查询待审批的补卡申请
func (s *punchCardSupplementService) ListPending(ctx context.Context, tenantID uuid.UUID) ([]*model.PunchCardSupplement, error) {
	return s.repo.FindPending(ctx, tenantID)
}

// Submit 提交补卡申请审批
func (s *punchCardSupplementService) Submit(ctx context.Context, supplementID, submitterID uuid.UUID) error {
	// 1. 获取补卡申请
	supplement, err := s.repo.FindByID(ctx, supplementID)
	if err != nil {
		return err
	}

	// 2. 检查状态
	if supplement.ApprovalStatus != "pending" {
		return fmt.Errorf("supplement is already submitted")
	}

	// 3. 这里可以集成工作流引擎创建审批流程
	// TODO: 调用工作流引擎创建审批实例

	// 4. 更新状态（暂时不修改，等待审批）
	return nil
}

// Approve 批准补卡申请
func (s *punchCardSupplementService) Approve(ctx context.Context, supplementID, approverID uuid.UUID, comment string) error {
	// 1. 获取补卡申请
	supplement, err := s.repo.FindByID(ctx, supplementID)
	if err != nil {
		return err
	}

	// 2. 检查状态
	if supplement.ApprovalStatus != "pending" {
		return fmt.Errorf("supplement status is not pending")
	}

	// 3. 更新审批信息
	now := time.Now()
	supplement.ApprovalStatus = "approved"
	supplement.ApprovedBy = &approverID
	supplement.ApprovedAt = &now

	// 4. 更新补卡申请
	return s.repo.Update(ctx, supplement)
}

// Reject 拒绝补卡申请
func (s *punchCardSupplementService) Reject(ctx context.Context, supplementID, approverID uuid.UUID, reason string) error {
	// 1. 获取补卡申请
	supplement, err := s.repo.FindByID(ctx, supplementID)
	if err != nil {
		return err
	}

	// 2. 检查状态
	if supplement.ApprovalStatus != "pending" {
		return fmt.Errorf("supplement status is not pending")
	}

	// 3. 更新审批信息
	now := time.Now()
	supplement.ApprovalStatus = "rejected"
	supplement.ApprovedBy = &approverID
	supplement.ApprovedAt = &now
	supplement.RejectReason = reason
	supplement.ProcessStatus = "cancelled"

	// 4. 更新补卡申请
	return s.repo.Update(ctx, supplement)
}

// Process 处理补卡（生成补卡记录）
func (s *punchCardSupplementService) Process(ctx context.Context, supplementID, processorID uuid.UUID) error {
	// 1. 获取补卡申请
	supplement, err := s.repo.FindByID(ctx, supplementID)
	if err != nil {
		return err
	}

	// 2. 检查审批状态
	if supplement.ApprovalStatus != "approved" {
		return fmt.Errorf("supplement is not approved")
	}

	// 3. 检查处理状态
	if supplement.ProcessStatus == "processed" {
		return fmt.Errorf("supplement is already processed")
	}

	// 4. 这里应该创建考勤记录或更新现有记录
	// TODO: 调用考勤服务创建补卡记录

	// 5. 更新处理状态
	now := time.Now()
	supplement.ProcessStatus = "processed"
	supplement.ProcessedAt = &now
	supplement.ProcessedBy = &processorID

	// 6. 更新补卡申请
	return s.repo.Update(ctx, supplement)
}

// Cancel 取消补卡申请
func (s *punchCardSupplementService) Cancel(ctx context.Context, supplementID uuid.UUID) error {
	// 1. 获取补卡申请
	supplement, err := s.repo.FindByID(ctx, supplementID)
	if err != nil {
		return err
	}

	// 2. 检查状态
	if supplement.ApprovalStatus != "pending" {
		return fmt.Errorf("cannot cancel supplement with status: %s", supplement.ApprovalStatus)
	}

	// 3. 更新状态
	supplement.ProcessStatus = "cancelled"

	// 4. 更新补卡申请
	return s.repo.Update(ctx, supplement)
}

// ValidateSupplement 验证补卡申请的合理性
func (s *punchCardSupplementService) ValidateSupplement(ctx context.Context, supplement *model.PunchCardSupplement) error {
	// 1. 验证补卡日期不能是未来日期
	now := time.Now()
	if supplement.SupplementDate.After(now) {
		return fmt.Errorf("supplement date cannot be in the future")
	}

	// 2. 验证补卡日期不能太久远（例如不超过30天）
	maxDaysAgo := 30
	minDate := now.AddDate(0, 0, -maxDaysAgo)
	if supplement.SupplementDate.Before(minDate) {
		return fmt.Errorf("supplement date cannot be more than %d days ago", maxDaysAgo)
	}

	// 3. 验证补卡时间格式
	if supplement.SupplementTime.IsZero() {
		return fmt.Errorf("supplement time is required")
	}

	// 4. 验证补卡时间在补卡日期当天
	supplementDate := supplement.SupplementDate.Format("2006-01-02")
	supplementTimeDate := supplement.SupplementTime.Format("2006-01-02")
	if supplementDate != supplementTimeDate {
		return fmt.Errorf("supplement time must be on the supplement date")
	}

	// 5. 验证补卡类型
	if supplement.SupplementType != model.SupplementTypeCheckIn && supplement.SupplementType != model.SupplementTypeCheckOut {
		return fmt.Errorf("invalid supplement type")
	}

	// 6. 验证缺卡类型
	validMissingTypes := map[model.PunchCardMissingType]bool{
		model.MissingTypeForgot:      true,
		model.MissingTypeMalfunction: true,
		model.MissingTypeOutside:     true,
		model.MissingTypeOther:       true,
	}
	if !validMissingTypes[supplement.MissingType] {
		return fmt.Errorf("invalid missing type")
	}

	// 7. 验证补卡原因不能为空
	if supplement.Reason == "" {
		return fmt.Errorf("reason is required")
	}

	// 8. 验证补卡原因长度
	if len(supplement.Reason) > 500 {
		return fmt.Errorf("reason is too long (max 500 characters)")
	}

	return nil
}

// CheckDuplicate 检查是否存在重复的补卡申请
func (s *punchCardSupplementService) CheckDuplicate(ctx context.Context, tenantID, employeeID uuid.UUID, date time.Time, supplementType model.SupplementType) (bool, error) {
	existing, err := s.repo.FindByDate(ctx, tenantID, employeeID, date, supplementType)
	if err != nil {
		return false, err
	}
	return existing != nil, nil
}

// CountByEmployee 统计员工补卡次数
func (s *punchCardSupplementService) CountByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) (int, error) {
	return s.repo.CountByEmployee(ctx, tenantID, employeeID, startDate, endDate)
}
