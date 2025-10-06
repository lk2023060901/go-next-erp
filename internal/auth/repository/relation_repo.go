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

// RelationRepository 关系仓储接口（ReBAC）
type RelationRepository interface {
	// 基础操作
	Create(ctx context.Context, tuple *model.RelationTuple) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByTuple(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) error

	// 查询
	FindBySubject(ctx context.Context, tenantID uuid.UUID, subject string) ([]*model.RelationTuple, error)
	FindByObject(ctx context.Context, tenantID uuid.UUID, object string) ([]*model.RelationTuple, error)
	FindByRelation(ctx context.Context, tenantID uuid.UUID, subject, relation string) ([]*model.RelationTuple, error)

	// 检查
	Check(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) (bool, error)
	Expand(ctx context.Context, tenantID uuid.UUID, object, relation string) ([]string, error) // 展开所有拥有该关系的主体
}

type relationRepo struct {
	db    *database.DB
	cache *cache.Cache
}

func NewRelationRepository(db *database.DB, cache *cache.Cache) RelationRepository {
	return &relationRepo{
		db:    db,
		cache: cache,
	}
}

// Create 创建关系元组
func (r *relationRepo) Create(ctx context.Context, tuple *model.RelationTuple) error {
	tuple.ID = uuid.Must(uuid.NewV7())
	tuple.CreatedAt = time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 检查是否已存在
		var exists bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM relation_tuples
				WHERE tenant_id = $1 AND subject = $2 AND relation = $3 AND object = $4 AND deleted_at IS NULL
			)
		`, tuple.TenantID, tuple.Subject, tuple.Relation, tuple.Object).Scan(&exists)

		if err != nil {
			return err
		}

		if exists {
			return nil // 已存在，跳过
		}

		// 插入关系元组
		_, err = tx.Exec(ctx, `
			INSERT INTO relation_tuples (id, tenant_id, subject, relation, object, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, tuple.ID, tuple.TenantID, tuple.Subject, tuple.Relation, tuple.Object, tuple.CreatedAt)

		if err == nil {
			r.invalidateCache(tuple.TenantID, tuple.Subject, tuple.Relation, tuple.Object)
		}

		return err
	})
}

// Delete 删除关系元组（软删除）
func (r *relationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 获取元组信息用于清除缓存
		var tenantID uuid.UUID
		var subject, relation, object string
		err := tx.QueryRow(ctx, `
			SELECT tenant_id, subject, relation, object FROM relation_tuples WHERE id = $1
		`, id).Scan(&tenantID, &subject, &relation, &object)

		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, "UPDATE relation_tuples SET deleted_at = $1 WHERE id = $2", now, id)

		if err == nil {
			r.invalidateCache(tenantID, subject, relation, object)
		}

		return err
	})
}

// DeleteByTuple 根据元组删除关系
func (r *relationRepo) DeleteByTuple(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) error {
	now := time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE relation_tuples SET deleted_at = $1
			WHERE tenant_id = $2 AND subject = $3 AND relation = $4 AND object = $5 AND deleted_at IS NULL
		`, now, tenantID, subject, relation, object)

		if err == nil {
			r.invalidateCache(tenantID, subject, relation, object)
		}

		return err
	})
}

// FindBySubject 查询主体的所有关系
func (r *relationRepo) FindBySubject(ctx context.Context, tenantID uuid.UUID, subject string) ([]*model.RelationTuple, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, subject, relation, object, created_at, deleted_at
		FROM relation_tuples
		WHERE tenant_id = $1 AND subject = $2 AND deleted_at IS NULL
	`, tenantID, subject)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTuples(rows)
}

// FindByObject 查询客体的所有关系
func (r *relationRepo) FindByObject(ctx context.Context, tenantID uuid.UUID, object string) ([]*model.RelationTuple, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, subject, relation, object, created_at, deleted_at
		FROM relation_tuples
		WHERE tenant_id = $1 AND object = $2 AND deleted_at IS NULL
	`, tenantID, object)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTuples(rows)
}

// FindByRelation 查询主体在指定关系下的所有客体
func (r *relationRepo) FindByRelation(ctx context.Context, tenantID uuid.UUID, subject, relation string) ([]*model.RelationTuple, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, subject, relation, object, created_at, deleted_at
		FROM relation_tuples
		WHERE tenant_id = $1 AND subject = $2 AND relation = $3 AND deleted_at IS NULL
	`, tenantID, subject, relation)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTuples(rows)
}

// Check 检查关系是否存在
func (r *relationRepo) Check(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) (bool, error) {
	cacheKey := fmt.Sprintf("relation:check:%s:%s:%s:%s", tenantID.String(), subject, relation, object)

	// 尝试从缓存获取
	var exists bool
	if err := r.cache.Get(ctx, cacheKey, &exists); err == nil {
		return exists, nil
	}

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM relation_tuples
			WHERE tenant_id = $1 AND subject = $2 AND relation = $3 AND object = $4 AND deleted_at IS NULL
		)
	`, tenantID, subject, relation, object).Scan(&exists)

	if err != nil {
		return false, err
	}

	// 缓存结果
	_ = r.cache.Set(ctx, cacheKey, exists, 300)

	return exists, nil
}

// Expand 展开所有拥有该关系的主体
func (r *relationRepo) Expand(ctx context.Context, tenantID uuid.UUID, object, relation string) ([]string, error) {
	cacheKey := fmt.Sprintf("relation:expand:%s:%s:%s", tenantID.String(), object, relation)

	// 尝试从缓存获取
	var subjects []string
	if err := r.cache.Get(ctx, cacheKey, &subjects); err == nil {
		return subjects, nil
	}

	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT subject FROM relation_tuples
		WHERE tenant_id = $1 AND object = $2 AND relation = $3 AND deleted_at IS NULL
	`, tenantID, object, relation)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var subject string
		if err := rows.Scan(&subject); err != nil {
			return nil, err
		}
		subjects = append(subjects, subject)
	}

	// 缓存结果
	_ = r.cache.Set(ctx, cacheKey, subjects, 300)

	return subjects, nil
}

// scanTuples 扫描关系元组列表
func (r *relationRepo) scanTuples(rows pgx.Rows) ([]*model.RelationTuple, error) {
	var tuples []*model.RelationTuple

	for rows.Next() {
		var tuple model.RelationTuple
		if err := rows.Scan(
			&tuple.ID, &tuple.TenantID, &tuple.Subject,
			&tuple.Relation, &tuple.Object, &tuple.CreatedAt, &tuple.DeletedAt,
		); err != nil {
			return nil, err
		}
		tuples = append(tuples, &tuple)
	}

	return tuples, nil
}

// invalidateCache 清除缓存
func (r *relationRepo) invalidateCache(tenantID uuid.UUID, subject, relation, object string) {
	ctx := context.Background()
	r.cache.Delete(ctx, fmt.Sprintf("relation:check:%s:%s:%s:%s", tenantID.String(), subject, relation, object))
	r.cache.Delete(ctx, fmt.Sprintf("relation:expand:%s:%s:%s", tenantID.String(), object, relation))
}
