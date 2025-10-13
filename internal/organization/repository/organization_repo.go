package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// OrganizationRepository 组织仓储接口
type OrganizationRepository interface {
	// Create 创建组织
	Create(ctx context.Context, org *model.Organization) error

	// Update 更新组织
	Update(ctx context.Context, org *model.Organization) error

	// Delete 删除组织（软删除）
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取组织
	GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error)

	// GetByCode 根据编码获取组织
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Organization, error)

	// GetByPath 根据路径获取组织
	GetByPath(ctx context.Context, path string) (*model.Organization, error)

	// List 列出租户的所有组织
	List(ctx context.Context, tenantID uuid.UUID) ([]*model.Organization, error)

	// ListByParent 列出指定父组织的子组织
	ListByParent(ctx context.Context, parentID uuid.UUID) ([]*model.Organization, error)

	// ListByLevel 列出指定层级的组织
	ListByLevel(ctx context.Context, tenantID uuid.UUID, level int) ([]*model.Organization, error)

	// ListByTypeCode 列出指定类型的组织
	ListByTypeCode(ctx context.Context, tenantID uuid.UUID, typeCode string) ([]*model.Organization, error)

	// GetRoots 获取根组织
	GetRoots(ctx context.Context, tenantID uuid.UUID) ([]*model.Organization, error)

	// GetChildren 获取直接子组织
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*model.Organization, error)

	// GetDescendants 获取所有后代组织（通过路径）
	GetDescendants(ctx context.Context, orgID uuid.UUID, path string) ([]*model.Organization, error)

	// UpdateChildrenLeafStatus 更新子节点的叶子状态
	UpdateChildrenLeafStatus(ctx context.Context, parentID uuid.UUID, isLeaf bool) error

	// UpdatePath 更新组织路径
	UpdatePath(ctx context.Context, orgID uuid.UUID, path, pathNames string, ancestorIDs []string, level int) error

	// Move 移动组织到新父节点
	Move(ctx context.Context, orgID, newParentID uuid.UUID) error

	// Exists 检查组织是否存在
	Exists(ctx context.Context, tenantID uuid.UUID, code string) (bool, error)

	// CountChildren 统计子组织数量
	CountChildren(ctx context.Context, parentID uuid.UUID) (int, error)

	// CountByTypeID 统计使用指定类型的组织数量
	CountByTypeID(ctx context.Context, typeID uuid.UUID) (int, error)
}

type organizationRepo struct {
	db *database.DB
}

// NewOrganizationRepository 创建组织仓储
func NewOrganizationRepository(db *database.DB) OrganizationRepository {
	return &organizationRepo{db: db}
}

func (r *organizationRepo) Create(ctx context.Context, org *model.Organization) error {
	sql := `
		INSERT INTO organizations (
			id, tenant_id, code, name, short_name, description,
			type_id, type_code, parent_id, level, path, path_names,
			ancestor_ids, is_leaf, leader_id, leader_name,
			legal_person, unified_code, register_date, register_addr,
			phone, email, address, employee_count, direct_emp_count,
			sort, status, tags, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16,
			$17, $18, $19, $20,
			$21, $22, $23, $24, $25,
			$26, $27, $28, $29, $30, $31, $32
		)
	`

	_, err := r.db.Exec(ctx, sql,
		org.ID, org.TenantID, org.Code, org.Name, org.ShortName, org.Description,
		org.TypeID, org.TypeCode, org.ParentID, org.Level, org.Path, org.PathNames,
		org.AncestorIDs, org.IsLeaf, org.LeaderID, org.LeaderName,
		org.LegalPerson, org.UnifiedCode, org.RegisterDate, org.RegisterAddr,
		org.Phone, org.Email, org.Address, org.EmployeeCount, org.DirectEmpCount,
		org.Sort, org.Status, org.Tags, org.CreatedBy, org.UpdatedBy, org.CreatedAt, org.UpdatedAt,
	)

	return err
}

func (r *organizationRepo) Update(ctx context.Context, org *model.Organization) error {
	sql := `
		UPDATE organizations SET
			name = $1, short_name = $2, description = $3,
			leader_id = $4, leader_name = $5,
			legal_person = $6, unified_code = $7, register_date = $8, register_addr = $9,
			phone = $10, email = $11, address = $12,
			sort = $13, status = $14, tags = $15,
			updated_by = $16, updated_at = $17
		WHERE id = $18 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql,
		org.Name, org.ShortName, org.Description,
		org.LeaderID, org.LeaderName,
		org.LegalPerson, org.UnifiedCode, org.RegisterDate, org.RegisterAddr,
		org.Phone, org.Email, org.Address,
		org.Sort, org.Status, org.Tags,
		org.UpdatedBy, org.UpdatedAt,
		org.ID,
	)

	return err
}

func (r *organizationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `UPDATE organizations SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *organizationRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	sql := `
		SELECT id, tenant_id, code, name, short_name, description,
		       type_id, type_code, parent_id, level, path, path_names,
		       ancestor_ids, is_leaf, leader_id, leader_name,
		       legal_person, unified_code, register_date, register_addr,
		       phone, email, address, employee_count, direct_emp_count,
		       sort, status, tags, created_by, updated_by, created_at, updated_at, deleted_at
		FROM organizations
		WHERE id = $1 AND deleted_at IS NULL
	`

	org := &model.Organization{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&org.ID, &org.TenantID, &org.Code, &org.Name, &org.ShortName, &org.Description,
		&org.TypeID, &org.TypeCode, &org.ParentID, &org.Level, &org.Path, &org.PathNames,
		&org.AncestorIDs, &org.IsLeaf, &org.LeaderID, &org.LeaderName,
		&org.LegalPerson, &org.UnifiedCode, &org.RegisterDate, &org.RegisterAddr,
		&org.Phone, &org.Email, &org.Address, &org.EmployeeCount, &org.DirectEmpCount,
		&org.Sort, &org.Status, &org.Tags, &org.CreatedBy, &org.UpdatedBy, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, err
	}

	return org, nil
}

func (r *organizationRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Organization, error) {
	sql := `
		SELECT id, tenant_id, code, name, short_name, description,
		       type_id, type_code, parent_id, level, path, path_names,
		       ancestor_ids, is_leaf, leader_id, leader_name,
		       legal_person, unified_code, register_date, register_addr,
		       phone, email, address, employee_count, direct_emp_count,
		       sort, status, tags, created_by, updated_by, created_at, updated_at, deleted_at
		FROM organizations
		WHERE tenant_id = $1 AND code = $2 AND deleted_at IS NULL
	`

	org := &model.Organization{}
	err := r.db.QueryRow(ctx, sql, tenantID, code).Scan(
		&org.ID, &org.TenantID, &org.Code, &org.Name, &org.ShortName, &org.Description,
		&org.TypeID, &org.TypeCode, &org.ParentID, &org.Level, &org.Path, &org.PathNames,
		&org.AncestorIDs, &org.IsLeaf, &org.LeaderID, &org.LeaderName,
		&org.LegalPerson, &org.UnifiedCode, &org.RegisterDate, &org.RegisterAddr,
		&org.Phone, &org.Email, &org.Address, &org.EmployeeCount, &org.DirectEmpCount,
		&org.Sort, &org.Status, &org.Tags, &org.CreatedBy, &org.UpdatedBy, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, err
	}

	return org, nil
}

func (r *organizationRepo) GetByPath(ctx context.Context, path string) (*model.Organization, error) {
	sql := `
		SELECT id, tenant_id, code, name, short_name, description,
		       type_id, type_code, parent_id, level, path, path_names,
		       ancestor_ids, is_leaf, leader_id, leader_name,
		       legal_person, unified_code, register_date, register_addr,
		       phone, email, address, employee_count, direct_emp_count,
		       sort, status, tags, created_by, updated_by, created_at, updated_at, deleted_at
		FROM organizations
		WHERE path = $1 AND deleted_at IS NULL
	`

	org := &model.Organization{}
	err := r.db.QueryRow(ctx, sql, path).Scan(
		&org.ID, &org.TenantID, &org.Code, &org.Name, &org.ShortName, &org.Description,
		&org.TypeID, &org.TypeCode, &org.ParentID, &org.Level, &org.Path, &org.PathNames,
		&org.AncestorIDs, &org.IsLeaf, &org.LeaderID, &org.LeaderName,
		&org.LegalPerson, &org.UnifiedCode, &org.RegisterDate, &org.RegisterAddr,
		&org.Phone, &org.Email, &org.Address, &org.EmployeeCount, &org.DirectEmpCount,
		&org.Sort, &org.Status, &org.Tags, &org.CreatedBy, &org.UpdatedBy, &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("organization not found for path: %s", path)
		}
		return nil, err
	}

	return org, nil
}

func (r *organizationRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*model.Organization, error) {
	sql := `
		SELECT id, tenant_id, code, name, short_name, description,
		       type_id, type_code, parent_id, level, path, path_names,
		       ancestor_ids, is_leaf, leader_id, leader_name,
		       employee_count, direct_emp_count, sort, status, tags,
		       created_at, updated_at
		FROM organizations
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY level ASC, sort ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*model.Organization
	for rows.Next() {
		org := &model.Organization{}
		err := rows.Scan(
			&org.ID, &org.TenantID, &org.Code, &org.Name, &org.ShortName, &org.Description,
			&org.TypeID, &org.TypeCode, &org.ParentID, &org.Level, &org.Path, &org.PathNames,
			&org.AncestorIDs, &org.IsLeaf, &org.LeaderID, &org.LeaderName,
			&org.EmployeeCount, &org.DirectEmpCount, &org.Sort, &org.Status, &org.Tags,
			&org.CreatedAt, &org.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}

func (r *organizationRepo) ListByParent(ctx context.Context, parentID uuid.UUID) ([]*model.Organization, error) {
	sql := `
		SELECT id, tenant_id, code, name, short_name, type_code, level, path,
		       is_leaf, leader_name, employee_count, status, sort, created_at
		FROM organizations
		WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY sort ASC
	`

	rows, err := r.db.Query(ctx, sql, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*model.Organization
	for rows.Next() {
		org := &model.Organization{}
		err := rows.Scan(
			&org.ID, &org.TenantID, &org.Code, &org.Name, &org.ShortName, &org.TypeCode, &org.Level, &org.Path,
			&org.IsLeaf, &org.LeaderName, &org.EmployeeCount, &org.Status, &org.Sort, &org.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}

func (r *organizationRepo) ListByLevel(ctx context.Context, tenantID uuid.UUID, level int) ([]*model.Organization, error) {
	sql := `
		SELECT id, tenant_id, code, name, type_code, level, path, is_leaf, status, sort
		FROM organizations
		WHERE tenant_id = $1 AND level = $2 AND deleted_at IS NULL
		ORDER BY sort ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, level)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*model.Organization
	for rows.Next() {
		org := &model.Organization{}
		err := rows.Scan(
			&org.ID, &org.TenantID, &org.Code, &org.Name, &org.TypeCode, &org.Level, &org.Path, &org.IsLeaf, &org.Status, &org.Sort,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}

func (r *organizationRepo) ListByTypeCode(ctx context.Context, tenantID uuid.UUID, typeCode string) ([]*model.Organization, error) {
	sql := `
		SELECT id, tenant_id, code, name, type_code, level, path, is_leaf, status, sort
		FROM organizations
		WHERE tenant_id = $1 AND type_code = $2 AND deleted_at IS NULL
		ORDER BY level ASC, sort ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, typeCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*model.Organization
	for rows.Next() {
		org := &model.Organization{}
		err := rows.Scan(
			&org.ID, &org.TenantID, &org.Code, &org.Name, &org.TypeCode, &org.Level, &org.Path, &org.IsLeaf, &org.Status, &org.Sort,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}

func (r *organizationRepo) GetRoots(ctx context.Context, tenantID uuid.UUID) ([]*model.Organization, error) {
	sql := `
		SELECT id, tenant_id, code, name, type_code, level, path, is_leaf, status, sort
		FROM organizations
		WHERE tenant_id = $1 AND parent_id IS NULL AND deleted_at IS NULL
		ORDER BY sort ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*model.Organization
	for rows.Next() {
		org := &model.Organization{}
		err := rows.Scan(
			&org.ID, &org.TenantID, &org.Code, &org.Name, &org.TypeCode, &org.Level, &org.Path, &org.IsLeaf, &org.Status, &org.Sort,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}

func (r *organizationRepo) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*model.Organization, error) {
	return r.ListByParent(ctx, parentID)
}

func (r *organizationRepo) GetDescendants(ctx context.Context, orgID uuid.UUID, path string) ([]*model.Organization, error) {
	sql := `
		SELECT id, tenant_id, code, name, type_code, level, path, is_leaf, status, sort
		FROM organizations
		WHERE path LIKE $1 AND id != $2 AND deleted_at IS NULL
		ORDER BY level ASC, sort ASC
	`

	rows, err := r.db.Query(ctx, sql, path+"%", orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*model.Organization
	for rows.Next() {
		org := &model.Organization{}
		err := rows.Scan(
			&org.ID, &org.TenantID, &org.Code, &org.Name, &org.TypeCode, &org.Level, &org.Path, &org.IsLeaf, &org.Status, &org.Sort,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}

func (r *organizationRepo) UpdateChildrenLeafStatus(ctx context.Context, parentID uuid.UUID, isLeaf bool) error {
	sql := `UPDATE organizations SET is_leaf = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, sql, isLeaf, parentID)
	return err
}

func (r *organizationRepo) UpdatePath(ctx context.Context, orgID uuid.UUID, path, pathNames string, ancestorIDs []string, level int) error {
	sql := `
		UPDATE organizations SET
			path = $1, path_names = $2, ancestor_ids = $3, level = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, sql, path, pathNames, ancestorIDs, level, orgID)
	return err
}

func (r *organizationRepo) Move(ctx context.Context, orgID, newParentID uuid.UUID) error {
	sql := `UPDATE organizations SET parent_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, sql, newParentID, orgID)
	return err
}

func (r *organizationRepo) Exists(ctx context.Context, tenantID uuid.UUID, code string) (bool, error) {
	sql := `
		SELECT COUNT(*) FROM organizations
		WHERE tenant_id = $1 AND code = $2 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, sql, tenantID, code).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *organizationRepo) CountChildren(ctx context.Context, parentID uuid.UUID) (int, error) {
	sql := `
		SELECT COUNT(*) FROM organizations
		WHERE parent_id = $1 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, sql, parentID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *organizationRepo) CountByTypeID(ctx context.Context, typeID uuid.UUID) (int, error) {
	sql := `
		SELECT COUNT(*) FROM organizations
		WHERE type_id = $1 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, sql, typeID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
