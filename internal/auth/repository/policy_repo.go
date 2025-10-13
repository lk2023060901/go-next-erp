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

// PolicyRepository 策略仓储接口
type PolicyRepository interface {
	// 基础 CRUD
	Create(ctx context.Context, policy *model.Policy) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Policy, error)
	Update(ctx context.Context, policy *model.Policy) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 策略查询
	GetApplicablePolicies(ctx context.Context, tenantID uuid.UUID, resource, action string) ([]*model.Policy, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.Policy, error)
	EnablePolicy(ctx context.Context, id uuid.UUID) error
	DisablePolicy(ctx context.Context, id uuid.UUID) error
}

type policyRepo struct {
	db    *database.DB
	cache *cache.Cache
}

func NewPolicyRepository(db *database.DB, cache *cache.Cache) PolicyRepository {
	return &policyRepo{
		db:    db,
		cache: cache,
	}
}

// Create 创建策略
func (r *policyRepo) Create(ctx context.Context, policy *model.Policy) error {
	policy.ID = uuid.Must(uuid.NewV7())
	now := time.Now()
	policy.CreatedAt = now
	policy.UpdatedAt = now

	// 将string字段转换为JSONB
	subjectJSON := []byte(`{"type": "*"}`)
	resourceJSON, _ := json.Marshal(map[string]string{"type": policy.Resource})
	conditionsJSON, _ := json.Marshal(map[string]string{"expression": policy.Expression})

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO policies (
				id, name, description, tenant_id, subject, resource, action,
				effect, conditions, priority, enabled, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`,
			policy.ID, policy.Name, policy.Description, policy.TenantID,
			subjectJSON, resourceJSON, policy.Action, policy.Effect,
			conditionsJSON, policy.Priority, policy.Enabled, policy.CreatedAt, policy.UpdatedAt,
		)

		if err == nil {
			// 清除缓存
			r.invalidatePolicyCache(policy.TenantID, policy.Resource, policy.Action)
		}

		return err
	})
}

// FindByID 根据 ID 查找策略
func (r *policyRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Policy, error) {
	cacheKey := fmt.Sprintf("policy:id:%s", id.String())

	var policy model.Policy
	if r.cache != nil {
		if err := r.cache.Get(ctx, cacheKey, &policy); err == nil {
			return &policy, nil
		}
	}

	row := r.db.QueryRow(ctx, `
		SELECT id, name, description, tenant_id, subject, resource, action,
			   effect, conditions, priority, enabled, created_at, updated_at
		FROM policies
		WHERE id = $1
	`, id)

	if err := r.scanPolicy(row, &policy); err != nil {
		return nil, err
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, cacheKey, &policy, 600)
	}
	return &policy, nil
}

// Update 更新策略
func (r *policyRepo) Update(ctx context.Context, policy *model.Policy) error {
	policy.UpdatedAt = time.Now()

	resourceJSON, _ := json.Marshal(map[string]string{"type": policy.Resource})
	conditionsJSON, _ := json.Marshal(map[string]string{"expression": policy.Expression})

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE policies SET
				name = $2, description = $3, resource = $4, action = $5,
				conditions = $6, effect = $7, priority = $8, enabled = $9, updated_at = $10
			WHERE id = $1
		`,
			policy.ID, policy.Name, policy.Description, resourceJSON, policy.Action,
			conditionsJSON, policy.Effect, policy.Priority, policy.Enabled, policy.UpdatedAt,
		)

		if err == nil {
			if r.cache != nil {
				r.cache.Delete(ctx, fmt.Sprintf("policy:id:%s", policy.ID.String()))
			}
			r.invalidatePolicyCache(policy.TenantID, policy.Resource, policy.Action)
		}

		return err
	})
}

// Delete 删除策略
func (r *policyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 获取策略信息用于清除缓存
		var tenantID uuid.UUID
		var resourceJSON []byte
		var action string
		err := tx.QueryRow(ctx,
			"SELECT tenant_id, resource, action FROM policies WHERE id = $1",
			id,
		).Scan(&tenantID, &resourceJSON, &action)

		if err != nil {
			return err
		}

		// 从JSONB中提取resource type
		var resourceMap map[string]string
		json.Unmarshal(resourceJSON, &resourceMap)
		resource := resourceMap["type"]

		_, err = tx.Exec(ctx, "DELETE FROM policies WHERE id = $1", id)

		if err == nil {
			if r.cache != nil {
				r.cache.Delete(ctx, fmt.Sprintf("policy:id:%s", id.String()))
			}
			r.invalidatePolicyCache(tenantID, resource, action)
		}

		return err
	})
}

// GetApplicablePolicies 获取适用的策略（按优先级排序）
func (r *policyRepo) GetApplicablePolicies(ctx context.Context, tenantID uuid.UUID, resource, action string) ([]*model.Policy, error) {
	cacheKey := fmt.Sprintf("policies:%s:%s:%s", tenantID.String(), resource, action)

	var policies []*model.Policy
	if r.cache != nil {
		if err := r.cache.Get(ctx, cacheKey, &policies); err == nil {
			return policies, nil
		}
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, name, description, tenant_id, subject, resource, action,
			   effect, conditions, priority, enabled, created_at, updated_at
		FROM policies
		WHERE tenant_id = $1
		  AND (resource->>'type' = $2 OR resource->>'type' = '*')
		  AND (action = $3 OR action = '*')
		  AND enabled = true
		ORDER BY priority DESC
	`, tenantID, resource, action)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var policy model.Policy
		if err := r.scanPolicy(rows, &policy); err != nil {
			return nil, err
		}
		policies = append(policies, &policy)
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, cacheKey, policies, 300) // 5分钟缓存
	}
	return policies, nil
}

// ListByTenant 获取租户的所有策略
func (r *policyRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.Policy, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, description, tenant_id, subject, resource, action,
			   effect, conditions, priority, enabled, created_at, updated_at
		FROM policies
		WHERE tenant_id = $1
		ORDER BY priority DESC, created_at DESC
	`, tenantID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*model.Policy
	for rows.Next() {
		var policy model.Policy
		if err := r.scanPolicy(rows, &policy); err != nil {
			return nil, err
		}
		policies = append(policies, &policy)
	}

	return policies, nil
}

// EnablePolicy 启用策略
func (r *policyRepo) EnablePolicy(ctx context.Context, id uuid.UUID) error {
	return r.updatePolicyStatus(ctx, id, true)
}

// DisablePolicy 禁用策略
func (r *policyRepo) DisablePolicy(ctx context.Context, id uuid.UUID) error {
	return r.updatePolicyStatus(ctx, id, false)
}

// updatePolicyStatus 更新策略状态
func (r *policyRepo) updatePolicyStatus(ctx context.Context, id uuid.UUID, enabled bool) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 获取策略信息
		var tenantID uuid.UUID
		var resourceJSON []byte
		var action string
		err := tx.QueryRow(ctx,
			"SELECT tenant_id, resource, action FROM policies WHERE id = $1",
			id,
		).Scan(&tenantID, &resourceJSON, &action)

		if err != nil {
			return err
		}

		// 从JSONB中提取resource type
		var resourceMap map[string]string
		json.Unmarshal(resourceJSON, &resourceMap)
		resource := resourceMap["type"]

		_, err = tx.Exec(ctx,
			"UPDATE policies SET enabled = $1, updated_at = $2 WHERE id = $3",
			enabled, time.Now(), id,
		)

		if err == nil {
			if r.cache != nil {
				r.cache.Delete(ctx, fmt.Sprintf("policy:id:%s", id.String()))
			}
			r.invalidatePolicyCache(tenantID, resource, action)
		}

		return err
	})
}

// scanPolicy 扫描策略数据
func (r *policyRepo) scanPolicy(row pgx.Row, policy *model.Policy) error {
	var subjectJSON, resourceJSON, conditionsJSON []byte

	err := row.Scan(
		&policy.ID, &policy.Name, &policy.Description, &policy.TenantID,
		&subjectJSON, &resourceJSON, &policy.Action, &policy.Effect,
		&conditionsJSON, &policy.Priority, &policy.Enabled, &policy.CreatedAt, &policy.UpdatedAt,
	)

	if err != nil {
		return err
	}

	// 从JSONB中提取resource type
	if len(resourceJSON) > 0 {
		var resourceMap map[string]string
		if err := json.Unmarshal(resourceJSON, &resourceMap); err == nil {
			policy.Resource = resourceMap["type"]
		}
	}

	// 从JSONB中提取expression
	if len(conditionsJSON) > 0 {
		var conditionsMap map[string]string
		if err := json.Unmarshal(conditionsJSON, &conditionsMap); err == nil {
			policy.Expression = conditionsMap["expression"]
		}
	}

	return nil
}

// invalidatePolicyCache 清除策略缓存
func (r *policyRepo) invalidatePolicyCache(tenantID uuid.UUID, resource, action string) {
	if r.cache != nil {
		ctx := context.Background()
		// 清除精确匹配的缓存
		r.cache.Delete(ctx, fmt.Sprintf("policies:%s:%s:%s", tenantID.String(), resource, action))
		// 清除通配符缓存
		r.cache.Delete(ctx, fmt.Sprintf("policies:%s:*:%s", tenantID.String(), action))
		r.cache.Delete(ctx, fmt.Sprintf("policies:%s:%s:*", tenantID.String(), resource))
		r.cache.Delete(ctx, fmt.Sprintf("policies:%s:*:*", tenantID.String()))
	}
}
