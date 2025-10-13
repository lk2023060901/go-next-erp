package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/pkg/cache"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// 分割 namespace:id 格式的字符串
func splitNamespaceID(s string) (string, string) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", s
}

// 合并 namespace 和 id
func joinNamespaceID(namespace, id string) string {
	if namespace == "" {
		return id
	}
	return namespace + ":" + id
}

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

	// 分割 subject 和 object
	subjectNS, subjectID := splitNamespaceID(tuple.Subject)
	objectNS, objectID := splitNamespaceID(tuple.Object)

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 检查是否已存在
		var exists bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM relation_tuples
				WHERE tenant_id = $1 AND subject_namespace = $2 AND subject_id = $3
				  AND relation = $4 AND namespace = $5 AND object_id = $6
			)
		`, tuple.TenantID, subjectNS, subjectID, tuple.Relation, objectNS, objectID).Scan(&exists)

		if err != nil {
			return err
		}

		if exists {
			return nil // 已存在，跳过
		}

		// 插入关系元组
		_, err = tx.Exec(ctx, `
			INSERT INTO relation_tuples (id, tenant_id, namespace, object_id, relation, subject_namespace, subject_id, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, tuple.ID, tuple.TenantID, objectNS, objectID, tuple.Relation, subjectNS, subjectID, tuple.CreatedAt)

		if err == nil {
			r.invalidateCache(tuple.TenantID, tuple.Subject, tuple.Relation, tuple.Object)
		}

		return err
	})
}

// Delete 删除关系元组
func (r *relationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 获取元组信息用于清除缓存
		var tenantID uuid.UUID
		var subjectNS, subjectID, relation, objectNS, objectID string
		err := tx.QueryRow(ctx, `
			SELECT tenant_id, subject_namespace, subject_id, relation, namespace, object_id FROM relation_tuples WHERE id = $1
		`, id).Scan(&tenantID, &subjectNS, &subjectID, &relation, &objectNS, &objectID)

		if err != nil {
			return err
		}

		subject := joinNamespaceID(subjectNS, subjectID)
		object := joinNamespaceID(objectNS, objectID)

		_, err = tx.Exec(ctx, "DELETE FROM relation_tuples WHERE id = $1", id)

		if err == nil {
			r.invalidateCache(tenantID, subject, relation, object)
		}

		return err
	})
}

// DeleteByTuple 根据元组删除关系
func (r *relationRepo) DeleteByTuple(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) error {
	subjectNS, subjectID := splitNamespaceID(subject)
	objectNS, objectID := splitNamespaceID(object)

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			DELETE FROM relation_tuples
			WHERE tenant_id = $1 AND subject_namespace = $2 AND subject_id = $3
			  AND relation = $4 AND namespace = $5 AND object_id = $6
		`, tenantID, subjectNS, subjectID, relation, objectNS, objectID)

		if err == nil {
			r.invalidateCache(tenantID, subject, relation, object)
		}

		return err
	})
}

// FindBySubject 查询主体的所有关系
func (r *relationRepo) FindBySubject(ctx context.Context, tenantID uuid.UUID, subject string) ([]*model.RelationTuple, error) {
	subjectNS, subjectID := splitNamespaceID(subject)

	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, subject_namespace, subject_id, relation, namespace, object_id, created_at
		FROM relation_tuples
		WHERE tenant_id = $1 AND subject_namespace = $2 AND subject_id = $3
	`, tenantID, subjectNS, subjectID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTuples(rows)
}

// FindByObject 查询客体的所有关系
func (r *relationRepo) FindByObject(ctx context.Context, tenantID uuid.UUID, object string) ([]*model.RelationTuple, error) {
	objectNS, objectID := splitNamespaceID(object)

	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, subject_namespace, subject_id, relation, namespace, object_id, created_at
		FROM relation_tuples
		WHERE tenant_id = $1 AND namespace = $2 AND object_id = $3
	`, tenantID, objectNS, objectID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTuples(rows)
}

// FindByRelation 查询主体在指定关系下的所有客体
func (r *relationRepo) FindByRelation(ctx context.Context, tenantID uuid.UUID, subject, relation string) ([]*model.RelationTuple, error) {
	subjectNS, subjectID := splitNamespaceID(subject)

	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, subject_namespace, subject_id, relation, namespace, object_id, created_at
		FROM relation_tuples
		WHERE tenant_id = $1 AND subject_namespace = $2 AND subject_id = $3 AND relation = $4
	`, tenantID, subjectNS, subjectID, relation)

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
	if r.cache != nil {
		if err := r.cache.Get(ctx, cacheKey, &exists); err == nil {
			return exists, nil
		}
	}

	subjectNS, subjectID := splitNamespaceID(subject)
	objectNS, objectID := splitNamespaceID(object)

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM relation_tuples
			WHERE tenant_id = $1 AND subject_namespace = $2 AND subject_id = $3
			  AND relation = $4 AND namespace = $5 AND object_id = $6
		)
	`, tenantID, subjectNS, subjectID, relation, objectNS, objectID).Scan(&exists)

	if err != nil {
		return false, err
	}

	// 缓存结果
	if r.cache != nil {
		_ = r.cache.Set(ctx, cacheKey, exists, 300)
	}

	return exists, nil
}

// Expand 展开所有拥有该关系的主体
func (r *relationRepo) Expand(ctx context.Context, tenantID uuid.UUID, object, relation string) ([]string, error) {
	cacheKey := fmt.Sprintf("relation:expand:%s:%s:%s", tenantID.String(), object, relation)

	// 尝试从缓存获取
	var subjects []string
	if r.cache != nil {
		if err := r.cache.Get(ctx, cacheKey, &subjects); err == nil {
			return subjects, nil
		}
	}

	objectNS, objectID := splitNamespaceID(object)

	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT subject_namespace, subject_id FROM relation_tuples
		WHERE tenant_id = $1 AND namespace = $2 AND object_id = $3 AND relation = $4
	`, tenantID, objectNS, objectID, relation)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var subjectNS, subjectID string
		if err := rows.Scan(&subjectNS, &subjectID); err != nil {
			return nil, err
		}
		subject := joinNamespaceID(subjectNS, subjectID)
		subjects = append(subjects, subject)
	}

	// 缓存结果
	if r.cache != nil {
		_ = r.cache.Set(ctx, cacheKey, subjects, 300)
	}

	return subjects, nil
}

// scanTuples 扫描关系元组列表
func (r *relationRepo) scanTuples(rows pgx.Rows) ([]*model.RelationTuple, error) {
	var tuples []*model.RelationTuple

	for rows.Next() {
		var tuple model.RelationTuple
		var subjectNS, subjectID, objectNS, objectID string

		if err := rows.Scan(
			&tuple.ID, &tuple.TenantID, &subjectNS, &subjectID,
			&tuple.Relation, &objectNS, &objectID, &tuple.CreatedAt,
		); err != nil {
			return nil, err
		}

		tuple.Subject = joinNamespaceID(subjectNS, subjectID)
		tuple.Object = joinNamespaceID(objectNS, objectID)

		tuples = append(tuples, &tuple)
	}

	return tuples, nil
}

// invalidateCache 清除缓存
func (r *relationRepo) invalidateCache(tenantID uuid.UUID, subject, relation, object string) {
	if r.cache != nil {
		ctx := context.Background()
		r.cache.Delete(ctx, fmt.Sprintf("relation:check:%s:%s:%s:%s", tenantID.String(), subject, relation, object))
		r.cache.Delete(ctx, fmt.Sprintf("relation:expand:%s:%s:%s", tenantID.String(), object, relation))
	}
}
