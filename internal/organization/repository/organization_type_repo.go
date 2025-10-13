package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// OrganizationTypeRepository 组织类型仓储接口
type OrganizationTypeRepository interface {
	// Create 创建组织类型
	Create(ctx context.Context, orgType *model.OrganizationType) error

	// Update 更新组织类型
	Update(ctx context.Context, orgType *model.OrganizationType) error

	// Delete 删除组织类型（软删除）
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取组织类型
	GetByID(ctx context.Context, id uuid.UUID) (*model.OrganizationType, error)

	// GetByCode 根据编码获取组织类型
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.OrganizationType, error)

	// List 列出租户的所有组织类型
	List(ctx context.Context, tenantID uuid.UUID) ([]*model.OrganizationType, error)

	// ListActive 列出激活的组织类型
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.OrganizationType, error)

	// Exists 检查组织类型是否存在
	Exists(ctx context.Context, tenantID uuid.UUID, code string) (bool, error)
}

type organizationTypeRepo struct {
	db *database.DB
}

// NewOrganizationTypeRepository 创建组织类型仓储
func NewOrganizationTypeRepository(db *database.DB) OrganizationTypeRepository {
	return &organizationTypeRepo{db: db}
}

func (r *organizationTypeRepo) Create(ctx context.Context, orgType *model.OrganizationType) error {
	sql := `
		INSERT INTO organization_types (
			id, tenant_id, code, name, icon,
			level, max_level, allow_root, allow_multi,
			allowed_parent_types, allowed_child_types,
			enable_leader, enable_legal_info, enable_address,
			sort, status, is_system,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11,
			$12, $13, $14,
			$15, $16, $17,
			$18, $19, $20, $21
		)
	`

	_, err := r.db.Exec(ctx, sql,
		orgType.ID, orgType.TenantID, orgType.Code, orgType.Name, orgType.Icon,
		orgType.Level, orgType.MaxLevel, orgType.AllowRoot, orgType.AllowMulti,
		orgType.AllowedParentTypes, orgType.AllowedChildTypes,
		orgType.EnableLeader, orgType.EnableLegalInfo, orgType.EnableAddress,
		orgType.Sort, orgType.Status, orgType.IsSystem,
		orgType.CreatedBy, orgType.UpdatedBy, orgType.CreatedAt, orgType.UpdatedAt,
	)

	return err
}

func (r *organizationTypeRepo) Update(ctx context.Context, orgType *model.OrganizationType) error {
	sql := `
		UPDATE organization_types SET
			name = $1, icon = $2,
			level = $3, max_level = $4, allow_root = $5, allow_multi = $6,
			allowed_parent_types = $7, allowed_child_types = $8,
			enable_leader = $9, enable_legal_info = $10, enable_address = $11,
			sort = $12, status = $13,
			updated_by = $14, updated_at = $15
		WHERE id = $16 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql,
		orgType.Name, orgType.Icon,
		orgType.Level, orgType.MaxLevel, orgType.AllowRoot, orgType.AllowMulti,
		orgType.AllowedParentTypes, orgType.AllowedChildTypes,
		orgType.EnableLeader, orgType.EnableLegalInfo, orgType.EnableAddress,
		orgType.Sort, orgType.Status,
		orgType.UpdatedBy, orgType.UpdatedAt,
		orgType.ID,
	)

	return err
}

func (r *organizationTypeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `UPDATE organization_types SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *organizationTypeRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.OrganizationType, error) {
	sql := `
		SELECT id, tenant_id, code, name, icon,
		       level, max_level, allow_root, allow_multi,
		       allowed_parent_types, allowed_child_types,
		       enable_leader, enable_legal_info, enable_address,
		       sort, status, is_system,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM organization_types
		WHERE id = $1 AND deleted_at IS NULL
	`

	orgType := &model.OrganizationType{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&orgType.ID, &orgType.TenantID, &orgType.Code, &orgType.Name, &orgType.Icon,
		&orgType.Level, &orgType.MaxLevel, &orgType.AllowRoot, &orgType.AllowMulti,
		&orgType.AllowedParentTypes, &orgType.AllowedChildTypes,
		&orgType.EnableLeader, &orgType.EnableLegalInfo, &orgType.EnableAddress,
		&orgType.Sort, &orgType.Status, &orgType.IsSystem,
		&orgType.CreatedBy, &orgType.UpdatedBy, &orgType.CreatedAt, &orgType.UpdatedAt, &orgType.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("organization type not found")
		}
		return nil, err
	}

	return orgType, nil
}

func (r *organizationTypeRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.OrganizationType, error) {
	sql := `
		SELECT id, tenant_id, code, name, icon,
		       level, max_level, allow_root, allow_multi,
		       allowed_parent_types, allowed_child_types,
		       enable_leader, enable_legal_info, enable_address,
		       sort, status, is_system,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM organization_types
		WHERE (tenant_id = $1 OR tenant_id IS NULL) AND code = $2 AND deleted_at IS NULL
		ORDER BY tenant_id DESC NULLS LAST
		LIMIT 1
	`

	orgType := &model.OrganizationType{}
	err := r.db.QueryRow(ctx, sql, tenantID, code).Scan(
		&orgType.ID, &orgType.TenantID, &orgType.Code, &orgType.Name, &orgType.Icon,
		&orgType.Level, &orgType.MaxLevel, &orgType.AllowRoot, &orgType.AllowMulti,
		&orgType.AllowedParentTypes, &orgType.AllowedChildTypes,
		&orgType.EnableLeader, &orgType.EnableLegalInfo, &orgType.EnableAddress,
		&orgType.Sort, &orgType.Status, &orgType.IsSystem,
		&orgType.CreatedBy, &orgType.UpdatedBy, &orgType.CreatedAt, &orgType.UpdatedAt, &orgType.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("organization type not found")
		}
		return nil, err
	}

	return orgType, nil
}

func (r *organizationTypeRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*model.OrganizationType, error) {
	sql := `
		SELECT id, tenant_id, code, name, icon,
		       level, max_level, allow_root, allow_multi,
		       allowed_parent_types, allowed_child_types,
		       enable_leader, enable_legal_info, enable_address,
		       sort, status, is_system,
		       created_by, updated_by, created_at, updated_at
		FROM organization_types
		WHERE (tenant_id = $1 OR tenant_id IS NULL) AND deleted_at IS NULL
		ORDER BY level ASC, sort ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgTypes []*model.OrganizationType
	for rows.Next() {
		orgType := &model.OrganizationType{}
		err := rows.Scan(
			&orgType.ID, &orgType.TenantID, &orgType.Code, &orgType.Name, &orgType.Icon,
			&orgType.Level, &orgType.MaxLevel, &orgType.AllowRoot, &orgType.AllowMulti,
			&orgType.AllowedParentTypes, &orgType.AllowedChildTypes,
			&orgType.EnableLeader, &orgType.EnableLegalInfo, &orgType.EnableAddress,
			&orgType.Sort, &orgType.Status, &orgType.IsSystem,
			&orgType.CreatedBy, &orgType.UpdatedBy, &orgType.CreatedAt, &orgType.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orgTypes = append(orgTypes, orgType)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orgTypes, nil
}

func (r *organizationTypeRepo) ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.OrganizationType, error) {
	sql := `
		SELECT id, tenant_id, code, name, icon,
		       level, max_level, allow_root, allow_multi,
		       allowed_parent_types, allowed_child_types,
		       enable_leader, enable_legal_info, enable_address,
		       sort, status, is_system,
		       created_by, updated_by, created_at, updated_at
		FROM organization_types
		WHERE (tenant_id = $1 OR tenant_id IS NULL) AND status = 'active' AND deleted_at IS NULL
		ORDER BY level ASC, sort ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgTypes []*model.OrganizationType
	for rows.Next() {
		orgType := &model.OrganizationType{}
		err := rows.Scan(
			&orgType.ID, &orgType.TenantID, &orgType.Code, &orgType.Name, &orgType.Icon,
			&orgType.Level, &orgType.MaxLevel, &orgType.AllowRoot, &orgType.AllowMulti,
			&orgType.AllowedParentTypes, &orgType.AllowedChildTypes,
			&orgType.EnableLeader, &orgType.EnableLegalInfo, &orgType.EnableAddress,
			&orgType.Sort, &orgType.Status, &orgType.IsSystem,
			&orgType.CreatedBy, &orgType.UpdatedBy, &orgType.CreatedAt, &orgType.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orgTypes = append(orgTypes, orgType)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orgTypes, nil
}

func (r *organizationTypeRepo) Exists(ctx context.Context, tenantID uuid.UUID, code string) (bool, error) {
	sql := `
		SELECT COUNT(*) FROM organization_types
		WHERE (tenant_id = $1 OR tenant_id IS NULL) AND code = $2 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, sql, tenantID, code).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
