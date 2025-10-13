package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// ClosureRepository 组织闭包表仓储接口
type ClosureRepository interface {
	// Insert 插入闭包关系
	Insert(ctx context.Context, closure *model.OrganizationClosure) error

	// BatchInsert 批量插入闭包关系
	BatchInsert(ctx context.Context, closures []*model.OrganizationClosure) error

	// Delete 删除闭包关系
	Delete(ctx context.Context, tenantID, descendantID uuid.UUID) error

	// GetAncestors 获取所有祖先
	GetAncestors(ctx context.Context, tenantID, descendantID uuid.UUID) ([]*model.OrganizationClosure, error)

	// GetDescendants 获取所有后代
	GetDescendants(ctx context.Context, tenantID, ancestorID uuid.UUID) ([]*model.OrganizationClosure, error)

	// GetDirectChildren 获取直接子节点
	GetDirectChildren(ctx context.Context, tenantID, ancestorID uuid.UUID) ([]*model.OrganizationClosure, error)

	// GetParent 获取父节点
	GetParent(ctx context.Context, tenantID, descendantID uuid.UUID) (*model.OrganizationClosure, error)

	// GetByDepth 获取指定深度的后代
	GetByDepth(ctx context.Context, tenantID, ancestorID uuid.UUID, depth int) ([]*model.OrganizationClosure, error)

	// DeleteSubtree 删除子树的所有闭包关系
	DeleteSubtree(ctx context.Context, tenantID, ancestorID uuid.UUID) error

	// Move 移动节点到新父节点
	Move(ctx context.Context, tenantID, nodeID, oldParentID, newParentID uuid.UUID) error
}

type closureRepo struct {
	db *database.DB
}

// NewClosureRepository 创建闭包表仓储
func NewClosureRepository(db *database.DB) ClosureRepository {
	return &closureRepo{db: db}
}

func (r *closureRepo) Insert(ctx context.Context, closure *model.OrganizationClosure) error {
	sql := `
		INSERT INTO organization_closures (tenant_id, ancestor_id, descendant_id, depth)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, sql,
		closure.TenantID, closure.AncestorID, closure.DescendantID, closure.Depth,
	)

	return err
}

func (r *closureRepo) BatchInsert(ctx context.Context, closures []*model.OrganizationClosure) error {
	if len(closures) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	sql := `
		INSERT INTO organization_closures (tenant_id, ancestor_id, descendant_id, depth)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING
	`

	for _, closure := range closures {
		batch.Queue(sql, closure.TenantID, closure.AncestorID, closure.DescendantID, closure.Depth)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for range closures {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r *closureRepo) Delete(ctx context.Context, tenantID, descendantID uuid.UUID) error {
	sql := `DELETE FROM organization_closures WHERE tenant_id = $1 AND descendant_id = $2`
	_, err := r.db.Exec(ctx, sql, tenantID, descendantID)
	return err
}

func (r *closureRepo) GetAncestors(ctx context.Context, tenantID, descendantID uuid.UUID) ([]*model.OrganizationClosure, error) {
	sql := `
		SELECT tenant_id, ancestor_id, descendant_id, depth
		FROM organization_closures
		WHERE tenant_id = $1 AND descendant_id = $2 AND depth > 0
		ORDER BY depth ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, descendantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var closures []*model.OrganizationClosure
	for rows.Next() {
		closure := &model.OrganizationClosure{}
		err := rows.Scan(
			&closure.TenantID, &closure.AncestorID, &closure.DescendantID, &closure.Depth,
		)
		if err != nil {
			return nil, err
		}
		closures = append(closures, closure)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return closures, nil
}

func (r *closureRepo) GetDescendants(ctx context.Context, tenantID, ancestorID uuid.UUID) ([]*model.OrganizationClosure, error) {
	sql := `
		SELECT tenant_id, ancestor_id, descendant_id, depth
		FROM organization_closures
		WHERE tenant_id = $1 AND ancestor_id = $2 AND depth > 0
		ORDER BY depth ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, ancestorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var closures []*model.OrganizationClosure
	for rows.Next() {
		closure := &model.OrganizationClosure{}
		err := rows.Scan(
			&closure.TenantID, &closure.AncestorID, &closure.DescendantID, &closure.Depth,
		)
		if err != nil {
			return nil, err
		}
		closures = append(closures, closure)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return closures, nil
}

func (r *closureRepo) GetDirectChildren(ctx context.Context, tenantID, ancestorID uuid.UUID) ([]*model.OrganizationClosure, error) {
	sql := `
		SELECT tenant_id, ancestor_id, descendant_id, depth
		FROM organization_closures
		WHERE tenant_id = $1 AND ancestor_id = $2 AND depth = 1
		ORDER BY descendant_id
	`

	rows, err := r.db.Query(ctx, sql, tenantID, ancestorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var closures []*model.OrganizationClosure
	for rows.Next() {
		closure := &model.OrganizationClosure{}
		err := rows.Scan(
			&closure.TenantID, &closure.AncestorID, &closure.DescendantID, &closure.Depth,
		)
		if err != nil {
			return nil, err
		}
		closures = append(closures, closure)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return closures, nil
}

func (r *closureRepo) GetParent(ctx context.Context, tenantID, descendantID uuid.UUID) (*model.OrganizationClosure, error) {
	sql := `
		SELECT tenant_id, ancestor_id, descendant_id, depth
		FROM organization_closures
		WHERE tenant_id = $1 AND descendant_id = $2 AND depth = 1
		LIMIT 1
	`

	closure := &model.OrganizationClosure{}
	err := r.db.QueryRow(ctx, sql, tenantID, descendantID).Scan(
		&closure.TenantID, &closure.AncestorID, &closure.DescendantID, &closure.Depth,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // 没有父节点（根节点）
		}
		return nil, err
	}

	return closure, nil
}

func (r *closureRepo) GetByDepth(ctx context.Context, tenantID, ancestorID uuid.UUID, depth int) ([]*model.OrganizationClosure, error) {
	sql := `
		SELECT tenant_id, ancestor_id, descendant_id, depth
		FROM organization_closures
		WHERE tenant_id = $1 AND ancestor_id = $2 AND depth = $3
		ORDER BY descendant_id
	`

	rows, err := r.db.Query(ctx, sql, tenantID, ancestorID, depth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var closures []*model.OrganizationClosure
	for rows.Next() {
		closure := &model.OrganizationClosure{}
		err := rows.Scan(
			&closure.TenantID, &closure.AncestorID, &closure.DescendantID, &closure.Depth,
		)
		if err != nil {
			return nil, err
		}
		closures = append(closures, closure)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return closures, nil
}

func (r *closureRepo) DeleteSubtree(ctx context.Context, tenantID, ancestorID uuid.UUID) error {
	// 删除所有以 ancestorID 为祖先的闭包关系（包括自身）
	sql := `
		DELETE FROM organization_closures
		WHERE tenant_id = $1 AND descendant_id IN (
			SELECT descendant_id FROM organization_closures
			WHERE tenant_id = $1 AND ancestor_id = $2
		)
	`
	_, err := r.db.Exec(ctx, sql, tenantID, ancestorID)
	return err
}

func (r *closureRepo) Move(ctx context.Context, tenantID, nodeID, oldParentID, newParentID uuid.UUID) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 1. 删除旧的祖先关系（不包括自身）
		deleteSql := `
			DELETE FROM organization_closures
			WHERE tenant_id = $1 AND descendant_id IN (
				SELECT descendant_id FROM organization_closures
				WHERE tenant_id = $1 AND ancestor_id = $2
			) AND ancestor_id IN (
				SELECT ancestor_id FROM organization_closures
				WHERE tenant_id = $1 AND descendant_id = $3 AND depth > 0
			)
		`
		_, err := tx.Exec(ctx, deleteSql, tenantID, nodeID, nodeID)
		if err != nil {
			return err
		}

		// 2. 插入新的祖先关系
		insertSql := `
			INSERT INTO organization_closures (tenant_id, ancestor_id, descendant_id, depth)
			SELECT $1, p.ancestor_id, c.descendant_id, p.depth + c.depth + 1
			FROM organization_closures p, organization_closures c
			WHERE p.tenant_id = $1 AND c.tenant_id = $1
			  AND p.descendant_id = $2  -- 新父节点的所有祖先
			  AND c.ancestor_id = $3    -- 要移动节点的所有后代
			ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING
		`
		_, err = tx.Exec(ctx, insertSql, tenantID, newParentID, nodeID)
		if err != nil {
			return err
		}

		// 3. 插入与新父节点的直接关系
		directSql := `
			INSERT INTO organization_closures (tenant_id, ancestor_id, descendant_id, depth)
			VALUES ($1, $2, $3, 1)
			ON CONFLICT (tenant_id, ancestor_id, descendant_id) DO NOTHING
		`
		_, err = tx.Exec(ctx, directSql, tenantID, newParentID, nodeID)
		return err
	})
}
