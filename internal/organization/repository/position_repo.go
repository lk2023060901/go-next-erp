package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// PositionRepository 职位仓储接口
type PositionRepository interface {
	// Create 创建职位
	Create(ctx context.Context, pos *model.Position) error

	// Update 更新职位
	Update(ctx context.Context, pos *model.Position) error

	// Delete 删除职位（软删除）
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取职位
	GetByID(ctx context.Context, id uuid.UUID) (*model.Position, error)

	// GetByCode 根据编码获取职位
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Position, error)

	// List 列出租户的所有职位
	List(ctx context.Context, tenantID uuid.UUID) ([]*model.Position, error)

	// ListByOrg 列出指定组织的职位
	ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*model.Position, error)

	// ListGlobal 列出全局职位
	ListGlobal(ctx context.Context, tenantID uuid.UUID) ([]*model.Position, error)

	// ListByCategory 列出指定类别的职位
	ListByCategory(ctx context.Context, tenantID uuid.UUID, category string) ([]*model.Position, error)

	// ListByLevel 列出指定职级范围的职位
	ListByLevel(ctx context.Context, tenantID uuid.UUID, minLevel, maxLevel int) ([]*model.Position, error)

	// ListActive 列出激活的职位
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.Position, error)

	// Exists 检查职位是否存在
	Exists(ctx context.Context, tenantID uuid.UUID, code string) (bool, error)
}

type positionRepo struct {
	db *database.DB
}

// NewPositionRepository 创建职位仓储
func NewPositionRepository(db *database.DB) PositionRepository {
	return &positionRepo{db: db}
}

func (r *positionRepo) Create(ctx context.Context, pos *model.Position) error {
	sql := `
		INSERT INTO positions (
			id, tenant_id, code, name, description, org_id, level, category,
			sort, status, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14
		)
	`

	_, err := r.db.Exec(ctx, sql,
		pos.ID, pos.TenantID, pos.Code, pos.Name, pos.Description, pos.OrgID, pos.Level, pos.Category,
		pos.Sort, pos.Status, pos.CreatedBy, pos.UpdatedBy, pos.CreatedAt, pos.UpdatedAt,
	)

	return err
}

func (r *positionRepo) Update(ctx context.Context, pos *model.Position) error {
	sql := `
		UPDATE positions SET
			name = $1, description = $2, level = $3, category = $4,
			sort = $5, status = $6, updated_by = $7, updated_at = $8
		WHERE id = $9 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql,
		pos.Name, pos.Description, pos.Level, pos.Category,
		pos.Sort, pos.Status, pos.UpdatedBy, pos.UpdatedAt,
		pos.ID,
	)

	return err
}

func (r *positionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `UPDATE positions SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *positionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Position, error) {
	sql := `
		SELECT id, tenant_id, code, name, description, org_id, level, category,
		       sort, status, created_by, updated_by, created_at, updated_at, deleted_at
		FROM positions
		WHERE id = $1 AND deleted_at IS NULL
	`

	pos := &model.Position{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&pos.ID, &pos.TenantID, &pos.Code, &pos.Name, &pos.Description, &pos.OrgID, &pos.Level, &pos.Category,
		&pos.Sort, &pos.Status, &pos.CreatedBy, &pos.UpdatedBy, &pos.CreatedAt, &pos.UpdatedAt, &pos.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("position not found")
		}
		return nil, err
	}

	return pos, nil
}

func (r *positionRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Position, error) {
	sql := `
		SELECT id, tenant_id, code, name, description, org_id, level, category,
		       sort, status, created_by, updated_by, created_at, updated_at, deleted_at
		FROM positions
		WHERE (tenant_id = $1 OR tenant_id IS NULL) AND code = $2 AND deleted_at IS NULL
		ORDER BY tenant_id DESC NULLS LAST
		LIMIT 1
	`

	pos := &model.Position{}
	err := r.db.QueryRow(ctx, sql, tenantID, code).Scan(
		&pos.ID, &pos.TenantID, &pos.Code, &pos.Name, &pos.Description, &pos.OrgID, &pos.Level, &pos.Category,
		&pos.Sort, &pos.Status, &pos.CreatedBy, &pos.UpdatedBy, &pos.CreatedAt, &pos.UpdatedAt, &pos.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("position not found")
		}
		return nil, err
	}

	return pos, nil
}

func (r *positionRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*model.Position, error) {
	sql := `
		SELECT id, tenant_id, code, name, description, org_id, level, category,
		       sort, status, created_at, updated_at
		FROM positions
		WHERE (tenant_id = $1 OR tenant_id IS NULL) AND deleted_at IS NULL
		ORDER BY level DESC, sort ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []*model.Position
	for rows.Next() {
		pos := &model.Position{}
		err := rows.Scan(
			&pos.ID, &pos.TenantID, &pos.Code, &pos.Name, &pos.Description, &pos.OrgID, &pos.Level, &pos.Category,
			&pos.Sort, &pos.Status, &pos.CreatedAt, &pos.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		positions = append(positions, pos)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}

func (r *positionRepo) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*model.Position, error) {
	sql := `
		SELECT id, tenant_id, code, name, description, org_id, level, category,
		       sort, status, created_at, updated_at
		FROM positions
		WHERE org_id = $1 AND deleted_at IS NULL
		ORDER BY level DESC, sort ASC
	`

	rows, err := r.db.Query(ctx, sql, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []*model.Position
	for rows.Next() {
		pos := &model.Position{}
		err := rows.Scan(
			&pos.ID, &pos.TenantID, &pos.Code, &pos.Name, &pos.Description, &pos.OrgID, &pos.Level, &pos.Category,
			&pos.Sort, &pos.Status, &pos.CreatedAt, &pos.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		positions = append(positions, pos)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}

func (r *positionRepo) ListGlobal(ctx context.Context, tenantID uuid.UUID) ([]*model.Position, error) {
	sql := `
		SELECT id, tenant_id, code, name, description, org_id, level, category,
		       sort, status, created_at, updated_at
		FROM positions
		WHERE (tenant_id = $1 OR tenant_id IS NULL) AND org_id IS NULL AND deleted_at IS NULL
		ORDER BY level DESC, sort ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []*model.Position
	for rows.Next() {
		pos := &model.Position{}
		err := rows.Scan(
			&pos.ID, &pos.TenantID, &pos.Code, &pos.Name, &pos.Description, &pos.OrgID, &pos.Level, &pos.Category,
			&pos.Sort, &pos.Status, &pos.CreatedAt, &pos.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		positions = append(positions, pos)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}

func (r *positionRepo) ListByCategory(ctx context.Context, tenantID uuid.UUID, category string) ([]*model.Position, error) {
	sql := `
		SELECT id, tenant_id, code, name, description, org_id, level, category,
		       sort, status, created_at, updated_at
		FROM positions
		WHERE (tenant_id = $1 OR tenant_id IS NULL) AND category = $2 AND deleted_at IS NULL
		ORDER BY level DESC, sort ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []*model.Position
	for rows.Next() {
		pos := &model.Position{}
		err := rows.Scan(
			&pos.ID, &pos.TenantID, &pos.Code, &pos.Name, &pos.Description, &pos.OrgID, &pos.Level, &pos.Category,
			&pos.Sort, &pos.Status, &pos.CreatedAt, &pos.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		positions = append(positions, pos)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}

func (r *positionRepo) ListByLevel(ctx context.Context, tenantID uuid.UUID, minLevel, maxLevel int) ([]*model.Position, error) {
	sql := `
		SELECT id, tenant_id, code, name, description, org_id, level, category,
		       sort, status, created_at, updated_at
		FROM positions
		WHERE (tenant_id = $1 OR tenant_id IS NULL) AND level BETWEEN $2 AND $3 AND deleted_at IS NULL
		ORDER BY level DESC, sort ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, minLevel, maxLevel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []*model.Position
	for rows.Next() {
		pos := &model.Position{}
		err := rows.Scan(
			&pos.ID, &pos.TenantID, &pos.Code, &pos.Name, &pos.Description, &pos.OrgID, &pos.Level, &pos.Category,
			&pos.Sort, &pos.Status, &pos.CreatedAt, &pos.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		positions = append(positions, pos)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}

func (r *positionRepo) ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.Position, error) {
	sql := `
		SELECT id, tenant_id, code, name, description, org_id, level, category,
		       sort, status, created_at, updated_at
		FROM positions
		WHERE (tenant_id = $1 OR tenant_id IS NULL) AND status = 'active' AND deleted_at IS NULL
		ORDER BY level DESC, sort ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []*model.Position
	for rows.Next() {
		pos := &model.Position{}
		err := rows.Scan(
			&pos.ID, &pos.TenantID, &pos.Code, &pos.Name, &pos.Description, &pos.OrgID, &pos.Level, &pos.Category,
			&pos.Sort, &pos.Status, &pos.CreatedAt, &pos.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		positions = append(positions, pos)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}

func (r *positionRepo) Exists(ctx context.Context, tenantID uuid.UUID, code string) (bool, error) {
	sql := `
		SELECT COUNT(*) FROM positions
		WHERE (tenant_id = $1 OR tenant_id IS NULL) AND code = $2 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, sql, tenantID, code).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
