package rbac

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/internal/auth/repository"
	"github.com/lk2023060901/go-next-erp/pkg/cache"
)

// Engine RBAC 授权引擎
type Engine struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
	cache          *cache.Cache
}

// NewEngine 创建 RBAC 引擎
func NewEngine(
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	cache *cache.Cache,
) *Engine {
	return &Engine{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		cache:          cache,
	}
}

// CheckPermission 检查用户是否拥有指定权限
func (e *Engine) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	// 1. 获取用户的所有权限（带缓存）
	permissions, err := e.getUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// 2. 检查权限匹配（支持通配符）
	for _, perm := range permissions {
		if perm.Match(resource, action) {
			return true, nil
		}
	}

	return false, nil
}

// GetUserPermissions 获取用户的所有权限
func (e *Engine) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error) {
	return e.getUserPermissions(ctx, userID)
}

// GetUserRoles 获取用户的所有角色（包含继承）
func (e *Engine) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.Role, error) {
	// 1. 获取用户直接角色
	directRoles, err := e.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. 获取角色继承链
	allRoles := make(map[uuid.UUID]*model.Role)
	for _, role := range directRoles {
		// 添加当前角色
		allRoles[role.ID] = role

		// 获取父级角色
		parentRoles, err := e.roleRepo.GetRoleHierarchy(ctx, role.ID)
		if err != nil {
			continue
		}

		for _, parent := range parentRoles {
			allRoles[parent.ID] = parent
		}
	}

	// 3. 转换为切片
	result := make([]*model.Role, 0, len(allRoles))
	for _, role := range allRoles {
		result = append(result, role)
	}

	return result, nil
}

// HasRole 检查用户是否拥有指定角色
func (e *Engine) HasRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error) {
	roles, err := e.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.ID == roleID {
			return true, nil
		}
	}

	return false, nil
}

// getUserPermissions 获取用户权限（带缓存）
func (e *Engine) getUserPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error) {
	cacheKey := fmt.Sprintf("rbac:user:permissions:%s", userID.String())

	// 尝试从缓存获取
	var permissions []*model.Permission
	if e.cache != nil {
		if err := e.cache.Get(ctx, cacheKey, &permissions); err == nil {
			return permissions, nil
		}
	}

	// 1. 获取用户的所有角色（包含继承）
	roles, err := e.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. 收集所有角色的权限（去重）
	permMap := make(map[uuid.UUID]*model.Permission)
	for _, role := range roles {
		rolePerms, err := e.permissionRepo.GetRolePermissions(ctx, role.ID)
		if err != nil {
			continue
		}

		for _, perm := range rolePerms {
			permMap[perm.ID] = perm
		}
	}

	// 3. 转换为切片
	permissions = make([]*model.Permission, 0, len(permMap))
	for _, perm := range permMap {
		permissions = append(permissions, perm)
	}

	// 4. 缓存结果
	if e.cache != nil {
		_ = e.cache.Set(ctx, cacheKey, permissions, 600) // 10分钟
	}

	return permissions, nil
}
