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

// RoleRepository 角色仓储接口
type RoleRepository interface {
	// 基础 CRUD
	Create(ctx context.Context, role *model.Role) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Role, error)
	FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*model.Role, error)
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 角色关系
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.Role, error)
	GetRoleHierarchy(ctx context.Context, roleID uuid.UUID) ([]*model.Role, error) // 获取父级角色链
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.Role, error)

	// 角色分配
	AssignRoleToUser(ctx context.Context, userID, roleID, tenantID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	HasRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error)
}

type roleRepo struct {
	db    *database.DB
	cache *cache.Cache
}

func NewRoleRepository(db *database.DB, cache *cache.Cache) RoleRepository {
	return &roleRepo{
		db:    db,
		cache: cache,
	}
}

// Create 创建角色
func (r *roleRepo) Create(ctx context.Context, role *model.Role) error {
	role.ID = uuid.Must(uuid.NewV7())
	now := time.Now()
	role.CreatedAt = now
	role.UpdatedAt = now

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO roles (id, name, display_name, description, tenant_id, parent_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, role.ID, role.Name, role.DisplayName, role.Description, role.TenantID, role.ParentID, role.CreatedAt, role.UpdatedAt)

		return err
	})
}

// FindByID 根据 ID 查找角色
func (r *roleRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	cacheKey := fmt.Sprintf("role:id:%s", id.String())

	var role model.Role
	if r.cache != nil {
		if err := r.cache.Get(ctx, cacheKey, &role); err == nil {
			return &role, nil
		}
	}

	row := r.db.QueryRow(ctx, `
		SELECT id, name, display_name, description, tenant_id, parent_id, created_at, updated_at, deleted_at
		FROM roles
		WHERE id = $1 AND deleted_at IS NULL
	`, id)

	if err := r.scanRole(row, &role); err != nil {
		return nil, err
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, cacheKey, &role, 600) // 10分钟
	}
	return &role, nil
}

// FindByName 根据名称查找角色
func (r *roleRepo) FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*model.Role, error) {
	var role model.Role

	row := r.db.QueryRow(ctx, `
		SELECT id, name, display_name, description, tenant_id, parent_id, created_at, updated_at, deleted_at
		FROM roles
		WHERE tenant_id = $1 AND name = $2 AND deleted_at IS NULL
	`, tenantID, name)

	if err := r.scanRole(row, &role); err != nil {
		return nil, err
	}

	return &role, nil
}

// Update 更新角色
func (r *roleRepo) Update(ctx context.Context, role *model.Role) error {
	role.UpdatedAt = time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE roles SET
				name = $2, display_name = $3, description = $4, parent_id = $5, updated_at = $6
			WHERE id = $1 AND deleted_at IS NULL
		`, role.ID, role.Name, role.DisplayName, role.Description, role.ParentID, role.UpdatedAt)

		if err == nil {
			if r.cache != nil {
			r.cache.Delete(ctx, fmt.Sprintf("role:id:%s", role.ID.String()))
			}
		}

		return err
	})
}

// Delete 软删除角色
func (r *roleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "UPDATE roles SET deleted_at = $1 WHERE id = $2", now, id)

		if err == nil {
			if r.cache != nil {
			r.cache.Delete(ctx, fmt.Sprintf("role:id:%s", id.String()))
			}
		}

		return err
	})
}

// GetUserRoles 获取用户的所有角色
func (r *roleRepo) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.Role, error) {
	cacheKey := fmt.Sprintf("user:roles:%s", userID.String())

	var roles []*model.Role
	if r.cache != nil {
		if err := r.cache.Get(ctx, cacheKey, &roles); err == nil {
			return roles, nil
		}
	}

	rows, err := r.db.Query(ctx, `
		SELECT r.id, r.name, r.display_name, r.description, r.tenant_id, r.parent_id, r.created_at, r.updated_at, r.deleted_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.deleted_at IS NULL
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var role model.Role
		if err := r.scanRole(rows, &role); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, cacheKey, roles, 600) // 10分钟
	}
	return roles, nil
}

// GetRoleHierarchy 获取角色的父级链（包含自己）
func (r *roleRepo) GetRoleHierarchy(ctx context.Context, roleID uuid.UUID) ([]*model.Role, error) {
	var roles []*model.Role

	// 递归查询父级角色（使用 CTE）
	rows, err := r.db.Query(ctx, `
		WITH RECURSIVE role_hierarchy AS (
			SELECT id, name, display_name, description, tenant_id, parent_id, created_at, updated_at, deleted_at
			FROM roles
			WHERE id = $1 AND deleted_at IS NULL

			UNION ALL

			SELECT r.id, r.name, r.display_name, r.description, r.tenant_id, r.parent_id, r.created_at, r.updated_at, r.deleted_at
			FROM roles r
			INNER JOIN role_hierarchy rh ON r.id = rh.parent_id
			WHERE r.deleted_at IS NULL
		)
		SELECT * FROM role_hierarchy
	`, roleID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var role model.Role
		if err := r.scanRole(rows, &role); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}

	return roles, nil
}

// ListByTenant 获取租户的所有角色
func (r *roleRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.Role, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, display_name, description, tenant_id, parent_id, created_at, updated_at, deleted_at
		FROM roles
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, tenantID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*model.Role
	for rows.Next() {
		var role model.Role
		if err := r.scanRole(rows, &role); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}

	return roles, nil
}

// AssignRoleToUser 分配角色给用户
func (r *roleRepo) AssignRoleToUser(ctx context.Context, userID, roleID, tenantID uuid.UUID) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 检查是否已分配
		var exists bool
		err := tx.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2)",
			userID, roleID,
		).Scan(&exists)

		if err != nil {
			return err
		}

		if exists {
			return nil // 已存在，跳过
		}

		// 插入角色关联
		id := uuid.Must(uuid.NewV7())
		_, err = tx.Exec(ctx, `
			INSERT INTO user_roles (id, user_id, role_id, tenant_id, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, id, userID, roleID, tenantID, time.Now())

		if err == nil {
			// 清除缓存
			if r.cache != nil {
			r.cache.Delete(ctx, fmt.Sprintf("user:roles:%s", userID.String()))
			}
		}

		return err
	})
}

// RemoveRoleFromUser 移除用户角色
func (r *roleRepo) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			"DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2",
			userID, roleID,
		)

		if err == nil {
			if r.cache != nil {
			r.cache.Delete(ctx, fmt.Sprintf("user:roles:%s", userID.String()))
			}
		}

		return err
	})
}

// HasRole 检查用户是否拥有指定角色
func (r *roleRepo) HasRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2)",
		userID, roleID,
	).Scan(&exists)

	return exists, err
}

// scanRole 扫描角色数据
func (r *roleRepo) scanRole(row pgx.Row, role *model.Role) error {
	return row.Scan(
		&role.ID, &role.Name, &role.DisplayName, &role.Description,
		&role.TenantID, &role.ParentID, &role.CreatedAt, &role.UpdatedAt, &role.DeletedAt,
	)
}
