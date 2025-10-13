package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
	"github.com/lk2023060901/go-next-erp/internal/organization/repository"
)

// PositionService 职位服务接口
type PositionService interface {
	// Create 创建职位
	Create(ctx context.Context, req *CreatePositionRequest) (*model.Position, error)

	// Update 更新职位
	Update(ctx context.Context, id uuid.UUID, req *UpdatePositionRequest) (*model.Position, error)

	// Delete 删除职位
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取职位
	GetByID(ctx context.Context, id uuid.UUID) (*model.Position, error)

	// GetByCode 根据编码获取职位
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Position, error)

	// List 列出所有职位
	List(ctx context.Context, tenantID uuid.UUID) ([]*model.Position, error)

	// ListByOrg 列出组织职位
	ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*model.Position, error)

	// ListGlobal 列出全局职位
	ListGlobal(ctx context.Context, tenantID uuid.UUID) ([]*model.Position, error)

	// ListByCategory 列出指定类别职位
	ListByCategory(ctx context.Context, tenantID uuid.UUID, category string) ([]*model.Position, error)

	// ListActive 列出激活的职位
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.Position, error)
}

type positionService struct {
	posRepo repository.PositionRepository
	empRepo repository.EmployeeRepository
}

// NewPositionService 创建职位服务
func NewPositionService(
	posRepo repository.PositionRepository,
	empRepo repository.EmployeeRepository,
) PositionService {
	return &positionService{
		posRepo: posRepo,
		empRepo: empRepo,
	}
}

// CreatePositionRequest 创建职位请求
type CreatePositionRequest struct {
	TenantID    uuid.UUID
	Code        string
	Name        string
	Description string
	OrgID       *uuid.UUID
	Level       int
	Category    string
	Sort        int
	Status      string
	CreatedBy   uuid.UUID
}

// UpdatePositionRequest 更新职位请求
type UpdatePositionRequest struct {
	Name        string
	Description string
	Level       int
	Category    string
	Sort        int
	Status      string
	UpdatedBy   uuid.UUID
}

func (s *positionService) Create(ctx context.Context, req *CreatePositionRequest) (*model.Position, error) {
	// 验证编码唯一性
	exists, err := s.posRepo.Exists(ctx, req.TenantID, req.Code)
	if err != nil {
		return nil, fmt.Errorf("check code exists failed: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("position code '%s' already exists", req.Code)
	}

	// 创建职位
	pos := &model.Position{
		ID:          uuid.New(),
		TenantID:    req.TenantID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		OrgID:       req.OrgID,
		Level:       req.Level,
		Category:    req.Category,
		Sort:        req.Sort,
		Status:      req.Status,
		CreatedBy:   req.CreatedBy,
		UpdatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.posRepo.Create(ctx, pos); err != nil {
		return nil, fmt.Errorf("create position failed: %w", err)
	}

	return pos, nil
}

func (s *positionService) Update(ctx context.Context, id uuid.UUID, req *UpdatePositionRequest) (*model.Position, error) {
	// 获取现有职位
	pos, err := s.posRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get position failed: %w", err)
	}

	// 更新字段
	pos.Name = req.Name
	pos.Description = req.Description
	pos.Level = req.Level
	pos.Category = req.Category
	pos.Sort = req.Sort
	pos.Status = req.Status
	pos.UpdatedBy = req.UpdatedBy
	pos.UpdatedAt = time.Now()

	if err := s.posRepo.Update(ctx, pos); err != nil {
		return nil, fmt.Errorf("update position failed: %w", err)
	}

	return pos, nil
}

func (s *positionService) Delete(ctx context.Context, id uuid.UUID) error {
	// 检查是否有员工使用该职位
	employees, err := s.empRepo.ListByPosition(ctx, id)
	if err != nil {
		return fmt.Errorf("list employees failed: %w", err)
	}
	if len(employees) > 0 {
		return fmt.Errorf("position has %d employees, cannot delete", len(employees))
	}

	return s.posRepo.Delete(ctx, id)
}

func (s *positionService) GetByID(ctx context.Context, id uuid.UUID) (*model.Position, error) {
	return s.posRepo.GetByID(ctx, id)
}

func (s *positionService) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Position, error) {
	return s.posRepo.GetByCode(ctx, tenantID, code)
}

func (s *positionService) List(ctx context.Context, tenantID uuid.UUID) ([]*model.Position, error) {
	return s.posRepo.List(ctx, tenantID)
}

func (s *positionService) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*model.Position, error) {
	return s.posRepo.ListByOrg(ctx, orgID)
}

func (s *positionService) ListGlobal(ctx context.Context, tenantID uuid.UUID) ([]*model.Position, error) {
	return s.posRepo.ListGlobal(ctx, tenantID)
}

func (s *positionService) ListByCategory(ctx context.Context, tenantID uuid.UUID, category string) ([]*model.Position, error) {
	return s.posRepo.ListByCategory(ctx, tenantID, category)
}

func (s *positionService) ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.Position, error) {
	return s.posRepo.ListActive(ctx, tenantID)
}
