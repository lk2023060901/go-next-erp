package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
	"github.com/lk2023060901/go-next-erp/internal/organization/repository"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// OrganizationService 组织服务接口
type OrganizationService interface {
	// Create 创建组织
	Create(ctx context.Context, req *CreateOrganizationRequest) (*model.Organization, error)

	// Update 更新组织
	Update(ctx context.Context, id uuid.UUID, req *UpdateOrganizationRequest) (*model.Organization, error)

	// Delete 删除组织
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取组织
	GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error)

	// GetByCode 根据编码获取组织
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Organization, error)

	// List 列出所有组织
	List(ctx context.Context, tenantID uuid.UUID) ([]*model.Organization, error)

	// GetTree 获取组织树
	GetTree(ctx context.Context, tenantID uuid.UUID, rootID *uuid.UUID) ([]*OrganizationTreeNode, error)

	// GetChildren 获取子组织
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*model.Organization, error)

	// GetDescendants 获取所有后代组织
	GetDescendants(ctx context.Context, orgID uuid.UUID) ([]*model.Organization, error)

	// Move 移动组织到新父节点
	Move(ctx context.Context, orgID, newParentID uuid.UUID, operatorID uuid.UUID) error

	// UpdateSort 更新排序
	UpdateSort(ctx context.Context, orgID uuid.UUID, sort int) error

	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, orgID uuid.UUID, status string) error
}

type organizationService struct {
	db          *database.DB
	orgRepo     repository.OrganizationRepository
	closureRepo repository.ClosureRepository
	typeRepo    repository.OrganizationTypeRepository
	empRepo     repository.EmployeeRepository
}

// NewOrganizationService 创建组织服务
func NewOrganizationService(
	db *database.DB,
	orgRepo repository.OrganizationRepository,
	closureRepo repository.ClosureRepository,
	typeRepo repository.OrganizationTypeRepository,
	empRepo repository.EmployeeRepository,
) OrganizationService {
	return &organizationService{
		db:          db,
		orgRepo:     orgRepo,
		closureRepo: closureRepo,
		typeRepo:    typeRepo,
		empRepo:     empRepo,
	}
}

// CreateOrganizationRequest 创建组织请求
type CreateOrganizationRequest struct {
	TenantID     uuid.UUID
	Code         string
	Name         string
	ShortName    string
	Description  string
	TypeID       uuid.UUID
	ParentID     *uuid.UUID
	LeaderID     *uuid.UUID
	LeaderName   string
	LegalPerson  string
	UnifiedCode  string
	RegisterDate *time.Time
	RegisterAddr string
	Phone        string
	Email        string
	Address      string
	Sort         int
	Status       string
	Tags         []string
	CreatedBy    uuid.UUID
}

// UpdateOrganizationRequest 更新组织请求
type UpdateOrganizationRequest struct {
	Name         string
	ShortName    string
	Description  string
	LeaderID     *uuid.UUID
	LeaderName   string
	LegalPerson  string
	UnifiedCode  string
	RegisterDate *time.Time
	RegisterAddr string
	Phone        string
	Email        string
	Address      string
	Sort         int
	Status       string
	Tags         []string
	UpdatedBy    uuid.UUID
}

// OrganizationTreeNode 组织树节点
type OrganizationTreeNode struct {
	*model.Organization
	Children []*OrganizationTreeNode `json:"children,omitempty"`
}

func (s *organizationService) Create(ctx context.Context, req *CreateOrganizationRequest) (*model.Organization, error) {
	// 验证编码唯一性
	exists, err := s.orgRepo.Exists(ctx, req.TenantID, req.Code)
	if err != nil {
		return nil, fmt.Errorf("check code exists failed: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("organization code '%s' already exists", req.Code)
	}

	// 获取组织类型
	orgType, err := s.typeRepo.GetByID(ctx, req.TypeID)
	if err != nil {
		return nil, fmt.Errorf("get organization type failed: %w", err)
	}

	// 计算层级和路径
	var level int
	var path, pathNames string
	var ancestorIDs []string
	var parentOrg *model.Organization

	if req.ParentID != nil {
		// 获取父组织
		parentOrg, err = s.orgRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("get parent organization failed: %w", err)
		}

		// 验证父类型是否允许
		parentType, err := s.typeRepo.GetByID(ctx, parentOrg.TypeID)
		if err != nil {
			return nil, fmt.Errorf("get parent type failed: %w", err)
		}

		if len(parentType.AllowedChildTypes) > 0 {
			allowed := false
			for _, code := range parentType.AllowedChildTypes {
				if code == orgType.Code {
					allowed = true
					break
				}
			}
			if !allowed {
				return nil, fmt.Errorf("child type '%s' is not allowed for parent type '%s'", orgType.Code, parentType.Code)
			}
		}

		// 计算层级和路径
		level = parentOrg.Level + 1
		path = parentOrg.Path + req.Code + "/"
		pathNames = parentOrg.PathNames + req.Name + "/"
		ancestorIDs = append(parentOrg.AncestorIDs, parentOrg.ID.String())
	} else {
		// 根节点
		if !orgType.AllowRoot {
			return nil, fmt.Errorf("organization type '%s' cannot be root", orgType.Code)
		}
		level = 1
		path = "/" + req.Code + "/"
		pathNames = "/" + req.Name + "/"
		ancestorIDs = []string{}
	}

	// 创建组织
	org := &model.Organization{
		ID:             uuid.New(),
		TenantID:       req.TenantID,
		Code:           req.Code,
		Name:           req.Name,
		ShortName:      req.ShortName,
		Description:    req.Description,
		TypeID:         req.TypeID,
		TypeCode:       orgType.Code,
		ParentID:       req.ParentID,
		Level:          level,
		Path:           path,
		PathNames:      pathNames,
		AncestorIDs:    ancestorIDs,
		IsLeaf:         true,
		LeaderID:       req.LeaderID,
		LeaderName:     req.LeaderName,
		LegalPerson:    req.LegalPerson,
		UnifiedCode:    req.UnifiedCode,
		RegisterDate:   req.RegisterDate,
		RegisterAddr:   req.RegisterAddr,
		Phone:          req.Phone,
		Email:          req.Email,
		Address:        req.Address,
		EmployeeCount:  0,
		DirectEmpCount: 0,
		Sort:           req.Sort,
		Status:         req.Status,
		Tags:           req.Tags,
		CreatedBy:      req.CreatedBy,
		UpdatedBy:      req.CreatedBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 使用事务创建组织和闭包关系
	err = s.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 创建组织
		if err := s.orgRepo.Create(ctx, org); err != nil {
			return fmt.Errorf("create organization failed: %w", err)
		}

		// 创建闭包关系
		closures := []*model.OrganizationClosure{
			// 自己到自己的关系
			{
				TenantID:     req.TenantID,
				AncestorID:   org.ID,
				DescendantID: org.ID,
				Depth:        0,
			},
		}

		// 添加祖先关系
		if req.ParentID != nil {
			// 获取父节点的所有祖先
			ancestors, err := s.closureRepo.GetAncestors(ctx, req.TenantID, *req.ParentID)
			if err != nil {
				return fmt.Errorf("get parent ancestors failed: %w", err)
			}

			// 添加父节点
			closures = append(closures, &model.OrganizationClosure{
				TenantID:     req.TenantID,
				AncestorID:   *req.ParentID,
				DescendantID: org.ID,
				Depth:        1,
			})

			// 添加所有祖先节点
			for _, ancestor := range ancestors {
				closures = append(closures, &model.OrganizationClosure{
					TenantID:     req.TenantID,
					AncestorID:   ancestor.AncestorID,
					DescendantID: org.ID,
					Depth:        ancestor.Depth + 1,
				})
			}

			// 更新父节点的 IsLeaf 状态
			if err := s.orgRepo.UpdateChildrenLeafStatus(ctx, *req.ParentID, false); err != nil {
				return fmt.Errorf("update parent leaf status failed: %w", err)
			}
		}

		// 批量插入闭包关系
		if err := s.closureRepo.BatchInsert(ctx, closures); err != nil {
			return fmt.Errorf("insert closures failed: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return org, nil
}

func (s *organizationService) Update(ctx context.Context, id uuid.UUID, req *UpdateOrganizationRequest) (*model.Organization, error) {
	// 获取现有组织
	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get organization failed: %w", err)
	}

	// 更新字段
	org.Name = req.Name
	org.ShortName = req.ShortName
	org.Description = req.Description
	org.LeaderID = req.LeaderID
	org.LeaderName = req.LeaderName
	org.LegalPerson = req.LegalPerson
	org.UnifiedCode = req.UnifiedCode
	org.RegisterDate = req.RegisterDate
	org.RegisterAddr = req.RegisterAddr
	org.Phone = req.Phone
	org.Email = req.Email
	org.Address = req.Address
	org.Sort = req.Sort
	org.Status = req.Status
	org.Tags = req.Tags
	org.UpdatedBy = req.UpdatedBy
	org.UpdatedAt = time.Now()

	// 更新路径名称（如果名称变更）
	if org.Name != req.Name {
		oldPath := org.PathNames
		newPath := strings.Replace(oldPath, "/"+org.Name+"/", "/"+req.Name+"/", 1)
		org.PathNames = newPath

		// 更新所有子节点的路径名称
		if err := s.updateChildrenPathNames(ctx, org.Path, oldPath, newPath); err != nil {
			return nil, fmt.Errorf("update children path names failed: %w", err)
		}
	}

	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, fmt.Errorf("update organization failed: %w", err)
	}

	return org, nil
}

func (s *organizationService) Delete(ctx context.Context, id uuid.UUID) error {
	// 检查是否有子组织
	count, err := s.orgRepo.CountChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("count children failed: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("organization has %d children, cannot delete", count)
	}

	// 检查是否有员工
	empCount, err := s.empRepo.CountByOrgDirect(ctx, id)
	if err != nil {
		return fmt.Errorf("count employees failed: %w", err)
	}
	if empCount > 0 {
		return fmt.Errorf("organization has %d employees, cannot delete", empCount)
	}

	// 使用事务删除组织和闭包关系
	err = s.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 获取组织信息
		org, err := s.orgRepo.GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("get organization failed: %w", err)
		}

		// 删除闭包关系
		if err := s.closureRepo.Delete(ctx, org.TenantID, id); err != nil {
			return fmt.Errorf("delete closures failed: %w", err)
		}

		// 删除组织
		if err := s.orgRepo.Delete(ctx, id); err != nil {
			return fmt.Errorf("delete organization failed: %w", err)
		}

		// 如果有父节点，检查父节点是否还有其他子节点
		if org.ParentID != nil {
			childCount, err := s.orgRepo.CountChildren(ctx, *org.ParentID)
			if err != nil {
				return fmt.Errorf("count parent children failed: %w", err)
			}
			if childCount == 0 {
				// 更新父节点为叶子节点
				if err := s.orgRepo.UpdateChildrenLeafStatus(ctx, *org.ParentID, true); err != nil {
					return fmt.Errorf("update parent leaf status failed: %w", err)
				}
			}
		}

		return nil
	})

	return err
}

func (s *organizationService) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	return s.orgRepo.GetByID(ctx, id)
}

func (s *organizationService) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Organization, error) {
	return s.orgRepo.GetByCode(ctx, tenantID, code)
}

func (s *organizationService) List(ctx context.Context, tenantID uuid.UUID) ([]*model.Organization, error) {
	return s.orgRepo.List(ctx, tenantID)
}

func (s *organizationService) GetTree(ctx context.Context, tenantID uuid.UUID, rootID *uuid.UUID) ([]*OrganizationTreeNode, error) {
	var orgs []*model.Organization
	var err error

	if rootID != nil {
		// 获取指定节点的子树
		root, err := s.orgRepo.GetByID(ctx, *rootID)
		if err != nil {
			return nil, fmt.Errorf("get root organization failed: %w", err)
		}
		descendants, err := s.orgRepo.GetDescendants(ctx, *rootID, root.Path)
		if err != nil {
			return nil, fmt.Errorf("get descendants failed: %w", err)
		}
		orgs = append([]*model.Organization{root}, descendants...)
	} else {
		// 获取所有组织
		orgs, err = s.orgRepo.List(ctx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("list organizations failed: %w", err)
		}
	}

	// 构建树形结构
	return s.buildTree(orgs, rootID), nil
}

func (s *organizationService) buildTree(orgs []*model.Organization, rootID *uuid.UUID) []*OrganizationTreeNode {
	// 创建 ID 到节点的映射
	nodeMap := make(map[uuid.UUID]*OrganizationTreeNode)
	for _, org := range orgs {
		nodeMap[org.ID] = &OrganizationTreeNode{
			Organization: org,
			Children:     []*OrganizationTreeNode{},
		}
	}

	// 构建父子关系
	var roots []*OrganizationTreeNode
	for _, org := range orgs {
		node := nodeMap[org.ID]
		if org.ParentID == nil || (rootID != nil && org.ID == *rootID) {
			// 根节点
			roots = append(roots, node)
		} else if parentNode, ok := nodeMap[*org.ParentID]; ok {
			// 添加到父节点的子节点列表
			parentNode.Children = append(parentNode.Children, node)
		}
	}

	return roots
}

func (s *organizationService) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*model.Organization, error) {
	return s.orgRepo.GetChildren(ctx, parentID)
}

func (s *organizationService) GetDescendants(ctx context.Context, orgID uuid.UUID) ([]*model.Organization, error) {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get organization failed: %w", err)
	}
	return s.orgRepo.GetDescendants(ctx, orgID, org.Path)
}

func (s *organizationService) Move(ctx context.Context, orgID, newParentID uuid.UUID, operatorID uuid.UUID) error {
	return s.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 获取要移动的组织
		org, err := s.orgRepo.GetByID(ctx, orgID)
		if err != nil {
			return fmt.Errorf("get organization failed: %w", err)
		}

		// 不能移动到自己的子节点下
		descendants, err := s.orgRepo.GetDescendants(ctx, orgID, org.Path)
		if err != nil {
			return fmt.Errorf("get descendants failed: %w", err)
		}
		for _, desc := range descendants {
			if desc.ID == newParentID {
				return fmt.Errorf("cannot move organization to its descendant")
			}
		}

		// 获取新父节点
		newParent, err := s.orgRepo.GetByID(ctx, newParentID)
		if err != nil {
			return fmt.Errorf("get new parent failed: %w", err)
		}

		// 验证类型兼容性
		newParentType, err := s.typeRepo.GetByID(ctx, newParent.TypeID)
		if err != nil {
			return fmt.Errorf("get parent type failed: %w", err)
		}

		orgType, err := s.typeRepo.GetByID(ctx, org.TypeID)
		if err != nil {
			return fmt.Errorf("get org type failed: %w", err)
		}

		if len(newParentType.AllowedChildTypes) > 0 {
			allowed := false
			for _, code := range newParentType.AllowedChildTypes {
				if code == orgType.Code {
					allowed = true
					break
				}
			}
			if !allowed {
				return fmt.Errorf("child type '%s' is not allowed for parent type '%s'", orgType.Code, newParentType.Code)
			}
		}

		// 更新闭包表
		if org.ParentID != nil {
			if err := s.closureRepo.Move(ctx, org.TenantID, orgID, *org.ParentID, newParentID); err != nil {
				return fmt.Errorf("move closure failed: %w", err)
			}
		}

		// 更新组织的父节点
		if err := s.orgRepo.Move(ctx, orgID, newParentID); err != nil {
			return fmt.Errorf("move organization failed: %w", err)
		}

		// 计算新路径
		newLevel := newParent.Level + 1
		newPath := newParent.Path + org.Code + "/"
		newPathNames := newParent.PathNames + org.Name + "/"
		newAncestorIDs := append(newParent.AncestorIDs, newParent.ID.String())

		// 更新路径
		if err := s.orgRepo.UpdatePath(ctx, orgID, newPath, newPathNames, newAncestorIDs, newLevel); err != nil {
			return fmt.Errorf("update path failed: %w", err)
		}

		// 递归更新所有子节点的路径
		if err := s.updateChildrenPaths(ctx, org.Path, newPath, newPathNames); err != nil {
			return fmt.Errorf("update children paths failed: %w", err)
		}

		// 更新旧父节点的叶子状态
		if org.ParentID != nil {
			childCount, err := s.orgRepo.CountChildren(ctx, *org.ParentID)
			if err != nil {
				return fmt.Errorf("count old parent children failed: %w", err)
			}
			if childCount == 0 {
				if err := s.orgRepo.UpdateChildrenLeafStatus(ctx, *org.ParentID, true); err != nil {
					return fmt.Errorf("update old parent leaf status failed: %w", err)
				}
			}
		}

		// 更新新父节点的叶子状态
		if err := s.orgRepo.UpdateChildrenLeafStatus(ctx, newParentID, false); err != nil {
			return fmt.Errorf("update new parent leaf status failed: %w", err)
		}

		return nil
	})
}

func (s *organizationService) UpdateSort(ctx context.Context, orgID uuid.UUID, sort int) error {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return fmt.Errorf("get organization failed: %w", err)
	}

	org.Sort = sort
	org.UpdatedAt = time.Now()

	return s.orgRepo.Update(ctx, org)
}

func (s *organizationService) UpdateStatus(ctx context.Context, orgID uuid.UUID, status string) error {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return fmt.Errorf("get organization failed: %w", err)
	}

	org.Status = status
	org.UpdatedAt = time.Now()

	return s.orgRepo.Update(ctx, org)
}

// updateChildrenPathNames 更新所有子节点的路径名称
func (s *organizationService) updateChildrenPathNames(ctx context.Context, parentPath, oldPathNames, newPathNames string) error {
	// 直接通过路径获取父组织
	parentOrg, err := s.orgRepo.GetByPath(ctx, parentPath)
	if err != nil {
		return fmt.Errorf("get parent organization by path failed: %w", err)
	}

	// 获取所有子节点
	children, err := s.orgRepo.GetDescendants(ctx, parentOrg.ID, parentPath)
	if err != nil {
		return fmt.Errorf("get descendants failed: %w", err)
	}

	// 批量更新子节点的 PathNames
	for _, child := range children {
		// 替换路径名称
		updatedPathNames := strings.Replace(child.PathNames, oldPathNames, newPathNames, 1)
		child.PathNames = updatedPathNames
		if err := s.orgRepo.Update(ctx, child); err != nil {
			return fmt.Errorf("update child %s path names failed: %w", child.ID, err)
		}
	}

	return nil
}

// updateChildrenPaths 递归更新所有子节点的路径
func (s *organizationService) updateChildrenPaths(ctx context.Context, oldPath, newPath, newPathNames string) error {
	// 直接通过路径获取父组织
	parentOrg, err := s.orgRepo.GetByPath(ctx, oldPath)
	if err != nil {
		return fmt.Errorf("get parent organization by path failed: %w", err)
	}

	// 获取所有子节点
	children, err := s.orgRepo.GetDescendants(ctx, parentOrg.ID, oldPath)
	if err != nil {
		return fmt.Errorf("get descendants failed: %w", err)
	}

	// 批量更新子节点的 Path 和 PathNames
	for _, child := range children {
		// 替换路径
		updatedPath := strings.Replace(child.Path, oldPath, newPath, 1)
		updatedPathNames := strings.Replace(child.PathNames, oldPath, newPathNames, 1)

		// 重新计算层级
		newLevel := strings.Count(updatedPath, "/") - 1

		// 更新路径、路径名称和层级
		ancestorIDs := strings.Split(strings.Trim(updatedPath, "/"), "/")
		if err := s.orgRepo.UpdatePath(ctx, child.ID, updatedPath, updatedPathNames, ancestorIDs[:len(ancestorIDs)-1], newLevel); err != nil {
			return fmt.Errorf("update child %s paths failed: %w", child.ID, err)
		}
	}

	return nil
}
