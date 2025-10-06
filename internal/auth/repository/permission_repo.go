package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/pkg/cache"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// PermissionRepository 权限仓储接口
type PermissionRepository interface {
	// 基础 CRUD
	Create(ctx context.Context, permission *model.Permission) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Permission, error)
	FindByResourceAction(ctx context.Context, tenantID uuid.UUID, resource, action string) (*model.Permission, error)
	Update(ctx context.Context, permission *model.Permission) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 权限查询
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*model.Permission, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.Permission, error)

	// 权限分配
	AssignPermissionToRole(ctx context.Context, roleID, permissionID, tenantID uuid.UUID) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	HasPermission(ctx context.Context, roleID, permissionID uuid.UUID) (bool, error)
}

type permissionRepo struct {
	db    *database.DB
	cache *cache.Cache
}

func NewPermissionRepository(db *database.DB, cache *cache.Cache) PermissionRepository {
	return &permissionRepo{
		db:    db,
		cache: cache,
	}
}

// Create 创建权限
func (r *permissionRepo) Create(ctx context.Context, permission *model.Permission) error {
	permission.ID = uuid.Must(uuid.NewV7())
	now := time.Now()
	permission.CreatedAt = now
	permission.UpdatedAt = now

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO permissions (id, resource, action, display_name, description, tenant_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, permission.ID, permission.Resource, permission.Action, permission.DisplayName,
			permission.Description, permission.TenantID, permission.CreatedAt, permission.UpdatedAt)

		return err
	})
}

// FindByID 根据 ID 查找权限
func (r *permissionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Permission, error) {
	cacheKey := fmt.Sprintf("permission:id:%s", id.String())

	var perm model.Permission
	if err := r.cache.Get(ctx, cacheKey, &perm); err == nil {
		return &perm, nil
	}

	row := r.db.QueryRow(ctx, `
		SELECT id, resource, action, display_name, description, tenant_id, created_at, updated_at, deleted_at
		FROM permissions
		WHERE id = $1 AND deleted_at IS NULL
	`, id)

	if err := r.scanPermission(row, &perm); err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, cacheKey, &perm, 600)
	return &perm, nil
}

// FindByResourceAction 根据资源和操作查找权限
func (r *permissionRepo) FindByResourceAction(ctx context.Context, tenantID uuid.UUID, resource, action string) (*model.Permission, error) {
	var perm model.Permission

	row := r.db.QueryRow(ctx, `
		SELECT id, resource, action, display_name, description, tenant_id, created_at, updated_at, deleted_at
		FROM permissions
		WHERE tenant_id = $1 AND resource = $2 AND action = $3 AND deleted_at IS NULL
	`, tenantID, resource, action)

	if err := r.scanPermission(row, &perm); err != nil {
		return nil, err
	}

	return &perm, nil
}

// Update 更新权限
func (r *permissionRepo) Update(ctx context.Context, permission *model.Permission) error {
	permission.UpdatedAt = time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE permissions SET
				resource = $2, action = $3, display_name = $4, description = $5, updated_at = $6
			WHERE id = $1 AND deleted_at IS NULL
		`, permission.ID, permission.Resource, permission.Action,
			permission.DisplayName, permission.Description, permission.UpdatedAt)

		if err == nil {
			r.cache.Delete(ctx, fmt.Sprintf("permission:id:%s", permission.ID.String()))
		}

		return err
	})
}

// Delete 软删除权限
func (r *permissionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "UPDATE permissions SET deleted_at = $1 WHERE id = $2", now, id)

		if err == nil {
			r.cache.Delete(ctx, fmt.Sprintf("permission:id:%s", id.String()))
		}

		return err
	})
}

// GetRolePermissions 获取角色的所有权限
func (r *permissionRepo) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*model.Permission, error) {
	cacheKey := fmt.Sprintf("role:permissions:%s", roleID.String())

	var permissions []*model.Permission
	if err := r.cache.Get(ctx, cacheKey, &permissions); err == nil {
		return permissions, nil
	}

	rows, err := r.db.Query(ctx, `
		SELECT p.id, p.resource, p.action, p.display_name, p.description, p.tenant_id, p.created_at, p.updated_at, p.deleted_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1 AND p.deleted_at IS NULL
	`, roleID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var perm model.Permission
		if err := r.scanPermission(rows, &perm); err != nil {
			return nil, err
		}
		permissions = append(permissions, &perm)
	}

	_ = r.cache.Set(ctx, cacheKey, permissions, 600)
	return permissions, nil
}

// GetUserPermissions 获取用户的所有权限（通过角色）
func (r *permissionRepo) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error) {
	cacheKey := fmt.Sprintf("user:permissions:%s", userID.String())

	var permissions []*model.Permission
	if err := r.cache.Get(ctx, cacheKey, &permissions); err == nil {
		return permissions, nil
	}

	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT p.id, p.resource, p.action, p.display_name, p.description, p.tenant_id, p.created_at, p.updated_at, p.deleted_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1 AND p.deleted_at IS NULL
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var perm model.Permission
		if err := r.scanPermission(rows, &perm); err != nil {
			return nil, err
		}
		permissions = append(permissions, &perm)
	}

	_ = r.cache.Set(ctx, cacheKey, permissions, 600)
	return permissions, nil
}

// ListByTenant 获取租户的所有权限
func (r *permissionRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.Permission, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, resource, action, display_name, description, tenant_id, created_at, updated_at, deleted_at
		FROM permissions
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY resource, action
	`, tenantID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*model.Permission
	for rows.Next() {
		var perm model.Permission
		if err := r.scanPermission(rows, &perm); err != nil {
			return nil, err
		}
		permissions = append(permissions, &perm)
	}

	return permissions, nil
}

// AssignPermissionToRole 分配权限给角色
func (r *permissionRepo) AssignPermissionToRole(ctx context.Context, roleID, permissionID, tenantID uuid.UUID) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 检查是否已分配
		var exists bool
		err := tx.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM role_permissions WHERE role_id = $1 AND permission_id = $2)",
			roleID, permissionID,
		).Scan(&exists)

		if err != nil {
			return err
		}

		if exists {
			return nil
		}

		// 插入权限关联
		id := uuid.Must(uuid.NewV7())
		_, err = tx.Exec(ctx, `
			INSERT INTO role_permissions (id, role_id, permission_id, tenant_id, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, id, roleID, permissionID, tenantID, time.Now())

		if err == nil {
			r.cache.Delete(ctx, fmt.Sprintf("role:permissions:%s", roleID.String()))
		}

		return err
	})
}

// RemovePermissionFromRole 移除角色权限
func (r *permissionRepo) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			"DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2",
			roleID, permissionID,
		)

		if err == nil {
			r.cache.Delete(ctx, fmt.Sprintf("role:permissions:%s", roleID.String()))
		}

		return err
	})
}

// HasPermission 检查角色是否拥有指定权限
func (r *permissionRepo) HasPermission(ctx context.Context, roleID, permissionID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM role_permissions WHERE role_id = $1 AND permission_id = $2)",
		roleID, permissionID,
	).Scan(&exists)

	return exists, err
}

// scanPermission 扫描权限数据
func (r *permissionRepo) scanPermission(row pgx.Row, perm *model.Permission) error {
	return row.Scan(
		&perm.ID, &perm.Resource, &perm.Action, &perm.DisplayName,
		&perm.Description, &perm.TenantID, &perm.CreatedAt, &perm.UpdatedAt, &perm.DeletedAt,
	)
}
