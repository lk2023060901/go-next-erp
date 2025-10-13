package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
	"github.com/lk2023060901/go-next-erp/internal/organization/repository"
)

// OrganizationTypeService 组织类型服务接口
type OrganizationTypeService interface {
	// Create 创建组织类型
	Create(ctx context.Context, req *CreateOrganizationTypeRequest) (*model.OrganizationType, error)

	// Update 更新组织类型
	Update(ctx context.Context, id uuid.UUID, req *UpdateOrganizationTypeRequest) (*model.OrganizationType, error)

	// Delete 删除组织类型
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取组织类型
	GetByID(ctx context.Context, id uuid.UUID) (*model.OrganizationType, error)

	// GetByCode 根据编码获取组织类型
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.OrganizationType, error)

	// List 列出所有组织类型
	List(ctx context.Context, tenantID uuid.UUID) ([]*model.OrganizationType, error)

	// ListActive 列出激活的组织类型
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.OrganizationType, error)

	// ValidateParentType 验证父类型是否允许
	ValidateParentType(ctx context.Context, typeID, parentTypeID uuid.UUID) error

	// ValidateChildType 验证子类型是否允许
	ValidateChildType(ctx context.Context, typeID, childTypeID uuid.UUID) error
}

type organizationTypeService struct {
	repo    repository.OrganizationTypeRepository
	orgRepo repository.OrganizationRepository
}

// NewOrganizationTypeService 创建组织类型服务
func NewOrganizationTypeService(repo repository.OrganizationTypeRepository, orgRepo repository.OrganizationRepository) OrganizationTypeService {
	return &organizationTypeService{
		repo:    repo,
		orgRepo: orgRepo,
	}
}

// CreateOrganizationTypeRequest 创建组织类型请求
type CreateOrganizationTypeRequest struct {
	TenantID           uuid.UUID
	Code               string
	Name               string
	Icon               string
	Level              int
	MaxLevel           int
	AllowRoot          bool
	AllowMulti         bool
	AllowedParentTypes []string
	AllowedChildTypes  []string
	EnableLeader       bool
	EnableLegalInfo    bool
	EnableAddress      bool
	Sort               int
	Status             string
	CreatedBy          uuid.UUID
}

// UpdateOrganizationTypeRequest 更新组织类型请求
type UpdateOrganizationTypeRequest struct {
	Name               string
	Icon               string
	MaxLevel           int
	AllowRoot          bool
	AllowMulti         bool
	AllowedParentTypes []string
	AllowedChildTypes  []string
	EnableLeader       bool
	EnableLegalInfo    bool
	EnableAddress      bool
	Sort               int
	Status             string
	UpdatedBy          uuid.UUID
}

func (s *organizationTypeService) Create(ctx context.Context, req *CreateOrganizationTypeRequest) (*model.OrganizationType, error) {
	// 验证编码唯一性
	exists, err := s.repo.Exists(ctx, req.TenantID, req.Code)
	if err != nil {
		return nil, fmt.Errorf("check code exists failed: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("organization type code '%s' already exists", req.Code)
	}

	// 创建组织类型
	orgType := &model.OrganizationType{
		ID:                 uuid.New(),
		TenantID:           req.TenantID,
		Code:               req.Code,
		Name:               req.Name,
		Icon:               req.Icon,
		Level:              req.Level,
		MaxLevel:           req.MaxLevel,
		AllowRoot:          req.AllowRoot,
		AllowMulti:         req.AllowMulti,
		AllowedParentTypes: req.AllowedParentTypes,
		AllowedChildTypes:  req.AllowedChildTypes,
		EnableLeader:       req.EnableLeader,
		EnableLegalInfo:    req.EnableLegalInfo,
		EnableAddress:      req.EnableAddress,
		Sort:               req.Sort,
		Status:             req.Status,
		IsSystem:           false,
		CreatedBy:          req.CreatedBy,
		UpdatedBy:          req.CreatedBy,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.repo.Create(ctx, orgType); err != nil {
		return nil, fmt.Errorf("create organization type failed: %w", err)
	}

	return orgType, nil
}

func (s *organizationTypeService) Update(ctx context.Context, id uuid.UUID, req *UpdateOrganizationTypeRequest) (*model.OrganizationType, error) {
	// 获取现有组织类型
	orgType, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get organization type failed: %w", err)
	}

	// 系统类型不允许修改
	if orgType.IsSystem {
		return nil, fmt.Errorf("system organization type cannot be modified")
	}

	// 更新字段
	orgType.Name = req.Name
	orgType.Icon = req.Icon
	orgType.MaxLevel = req.MaxLevel
	orgType.AllowRoot = req.AllowRoot
	orgType.AllowMulti = req.AllowMulti
	orgType.AllowedParentTypes = req.AllowedParentTypes
	orgType.AllowedChildTypes = req.AllowedChildTypes
	orgType.EnableLeader = req.EnableLeader
	orgType.EnableLegalInfo = req.EnableLegalInfo
	orgType.EnableAddress = req.EnableAddress
	orgType.Sort = req.Sort
	orgType.Status = req.Status
	orgType.UpdatedBy = req.UpdatedBy
	orgType.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, orgType); err != nil {
		return nil, fmt.Errorf("update organization type failed: %w", err)
	}

	return orgType, nil
}

func (s *organizationTypeService) Delete(ctx context.Context, id uuid.UUID) error {
	// 获取组织类型
	orgType, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get organization type failed: %w", err)
	}

	// 系统类型不允许删除
	if orgType.IsSystem {
		return fmt.Errorf("system organization type cannot be deleted")
	}

	// 检查是否有组织在使用该类型
	count, err := s.orgRepo.CountByTypeID(ctx, id)
	if err != nil {
		return fmt.Errorf("count organizations by type failed: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("organization type is in use by %d organizations, cannot delete", count)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete organization type failed: %w", err)
	}

	return nil
}

func (s *organizationTypeService) GetByID(ctx context.Context, id uuid.UUID) (*model.OrganizationType, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *organizationTypeService) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.OrganizationType, error) {
	return s.repo.GetByCode(ctx, tenantID, code)
}

func (s *organizationTypeService) List(ctx context.Context, tenantID uuid.UUID) ([]*model.OrganizationType, error) {
	return s.repo.List(ctx, tenantID)
}

func (s *organizationTypeService) ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.OrganizationType, error) {
	return s.repo.ListActive(ctx, tenantID)
}

func (s *organizationTypeService) ValidateParentType(ctx context.Context, typeID, parentTypeID uuid.UUID) error {
	// 获取子类型
	childType, err := s.repo.GetByID(ctx, typeID)
	if err != nil {
		return fmt.Errorf("get child type failed: %w", err)
	}

	// 获取父类型
	parentType, err := s.repo.GetByID(ctx, parentTypeID)
	if err != nil {
		return fmt.Errorf("get parent type failed: %w", err)
	}

	// 检查是否允许该父类型
	if len(childType.AllowedParentTypes) > 0 {
		allowed := false
		for _, code := range childType.AllowedParentTypes {
			if code == parentType.Code {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("parent type '%s' is not allowed for type '%s'", parentType.Code, childType.Code)
		}
	}

	return nil
}

func (s *organizationTypeService) ValidateChildType(ctx context.Context, typeID, childTypeID uuid.UUID) error {
	// 获取父类型
	parentType, err := s.repo.GetByID(ctx, typeID)
	if err != nil {
		return fmt.Errorf("get parent type failed: %w", err)
	}

	// 获取子类型
	childType, err := s.repo.GetByID(ctx, childTypeID)
	if err != nil {
		return fmt.Errorf("get child type failed: %w", err)
	}

	// 检查是否允许该子类型
	if len(parentType.AllowedChildTypes) > 0 {
		allowed := false
		for _, code := range parentType.AllowedChildTypes {
			if code == childType.Code {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("child type '%s' is not allowed for type '%s'", childType.Code, parentType.Code)
		}
	}

	return nil
}
