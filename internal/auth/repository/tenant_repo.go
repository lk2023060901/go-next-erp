package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/pkg/cache"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// TenantRepository 租户仓储接口
type TenantRepository interface {
	// 基础 CRUD
	Create(ctx context.Context, tenant *model.Tenant) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error)
	FindByDomain(ctx context.Context, domain string) (*model.Tenant, error)
	Update(ctx context.Context, tenant *model.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 租户管理
	List(ctx context.Context, limit, offset int) ([]*model.Tenant, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.TenantStatus) error
	UpdateSettings(ctx context.Context, id uuid.UUID, settings map[string]interface{}) error
}

type tenantRepo struct {
	db    *database.DB
	cache *cache.Cache
}

func NewTenantRepository(db *database.DB, cache *cache.Cache) TenantRepository {
	return &tenantRepo{
		db:    db,
		cache: cache,
	}
}

// Create 创建租户
func (r *tenantRepo) Create(ctx context.Context, tenant *model.Tenant) error {
	tenant.ID = uuid.Must(uuid.NewV7())
	now := time.Now()
	tenant.CreatedAt = now
	tenant.UpdatedAt = now

	settingsJSON, _ := json.Marshal(tenant.Settings)

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO tenants (
				id, name, domain, status, settings, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
			tenant.ID, tenant.Name, tenant.Domain, tenant.Status,
			settingsJSON, tenant.CreatedAt, tenant.UpdatedAt,
		)

		return err
	})
}

// FindByID 根据 ID 查找租户
func (r *tenantRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error) {
	cacheKey := fmt.Sprintf("tenant:id:%s", id.String())

	var tenant model.Tenant
	if r.cache != nil {
		if err := r.cache.Get(ctx, cacheKey, &tenant); err == nil {
			return &tenant, nil
		}
	}

	row := r.db.QueryRow(ctx, `
		SELECT id, name, domain, status, settings, created_at, updated_at, deleted_at
		FROM tenants
		WHERE id = $1 AND deleted_at IS NULL
	`, id)

	if err := r.scanTenant(row, &tenant); err != nil {
		return nil, err
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, cacheKey, &tenant, 600)
	}
	return &tenant, nil
}

// FindByDomain 根据域名查找租户
func (r *tenantRepo) FindByDomain(ctx context.Context, domain string) (*model.Tenant, error) {
	cacheKey := fmt.Sprintf("tenant:domain:%s", domain)

	var tenant model.Tenant
	if r.cache != nil {
		if err := r.cache.Get(ctx, cacheKey, &tenant); err == nil {
			return &tenant, nil
		}
	}

	row := r.db.QueryRow(ctx, `
		SELECT id, name, domain, status, settings, created_at, updated_at, deleted_at
		FROM tenants
		WHERE domain = $1 AND deleted_at IS NULL
	`, domain)

	if err := r.scanTenant(row, &tenant); err != nil {
		return nil, err
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, cacheKey, &tenant, 600)
	}
	return &tenant, nil
}

// Update 更新租户
func (r *tenantRepo) Update(ctx context.Context, tenant *model.Tenant) error {
	tenant.UpdatedAt = time.Now()
	settingsJSON, _ := json.Marshal(tenant.Settings)

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE tenants SET
				name = $2, domain = $3, status = $4, settings = $5, updated_at = $6
			WHERE id = $1 AND deleted_at IS NULL
		`,
			tenant.ID, tenant.Name, tenant.Domain, tenant.Status,
			settingsJSON, tenant.UpdatedAt,
		)

		if err == nil {
			r.invalidateCache(tenant.ID, tenant.Domain)
		}

		return err
	})
}

// Delete 软删除租户
func (r *tenantRepo) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		var domain string
		err := tx.QueryRow(ctx, "SELECT domain FROM tenants WHERE id = $1", id).Scan(&domain)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, "UPDATE tenants SET deleted_at = $1 WHERE id = $2", now, id)

		if err == nil {
			r.invalidateCache(id, domain)
		}

		return err
	})
}

// List 查询租户列表
func (r *tenantRepo) List(ctx context.Context, limit, offset int) ([]*model.Tenant, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, domain, status, settings, created_at, updated_at, deleted_at
		FROM tenants
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*model.Tenant
	for rows.Next() {
		var tenant model.Tenant
		if err := r.scanTenant(rows, &tenant); err != nil {
			return nil, err
		}
		tenants = append(tenants, &tenant)
	}

	return tenants, nil
}

// UpdateStatus 更新租户状态
func (r *tenantRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.TenantStatus) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		var domain string
		err := tx.QueryRow(ctx, "SELECT domain FROM tenants WHERE id = $1", id).Scan(&domain)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx,
			"UPDATE tenants SET status = $1, updated_at = $2 WHERE id = $3",
			status, time.Now(), id,
		)

		if err == nil {
			r.invalidateCache(id, domain)
		}

		return err
	})
}

// UpdateSettings 更新租户配置
func (r *tenantRepo) UpdateSettings(ctx context.Context, id uuid.UUID, settings map[string]interface{}) error {
	settingsJSON, _ := json.Marshal(settings)

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		var domain string
		err := tx.QueryRow(ctx, "SELECT domain FROM tenants WHERE id = $1", id).Scan(&domain)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx,
			"UPDATE tenants SET settings = $1, updated_at = $2 WHERE id = $3",
			settingsJSON, time.Now(), id,
		)

		if err == nil {
			r.invalidateCache(id, domain)
		}

		return err
	})
}

// scanTenant 扫描租户数据
func (r *tenantRepo) scanTenant(row pgx.Row, tenant *model.Tenant) error {
	var settingsJSON []byte
	var domain *string

	err := row.Scan(
		&tenant.ID, &tenant.Name, &domain, &tenant.Status,
		&settingsJSON, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt,
	)

	if err != nil {
		return err
	}

	if domain != nil {
		tenant.Domain = *domain
	}

	// 解析 settings
	if len(settingsJSON) > 0 {
		_ = json.Unmarshal(settingsJSON, &tenant.Settings)
	}

	return nil
}

// invalidateCache 清除缓存
func (r *tenantRepo) invalidateCache(id uuid.UUID, domain string) {
	if r.cache != nil {
		ctx := context.Background()
		r.cache.Delete(ctx, fmt.Sprintf("tenant:id:%s", id.String()))
		if domain != "" {
			r.cache.Delete(ctx, fmt.Sprintf("tenant:domain:%s", domain))
		}
	}
}
