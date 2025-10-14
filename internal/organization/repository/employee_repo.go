package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// EmployeeRepository 员工仓储接口
type EmployeeRepository interface {
	// Create 创建员工
	Create(ctx context.Context, emp *model.Employee) error

	// Update 更新员工
	Update(ctx context.Context, emp *model.Employee) error

	// Delete 删除员工（软删除）
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据 ID 获取员工
	GetByID(ctx context.Context, id uuid.UUID) (*model.Employee, error)

	// GetByUserID 根据用户 ID 获取员工
	GetByUserID(ctx context.Context, tenantID, userID uuid.UUID) (*model.Employee, error)

	// GetByEmployeeNo 根据工号获取员工
	GetByEmployeeNo(ctx context.Context, tenantID uuid.UUID, employeeNo string) (*model.Employee, error)

	// List 列出租户的所有员工
	List(ctx context.Context, tenantID uuid.UUID) ([]*model.Employee, error)

	// ListByOrg 列出指定组织的员工（包含子组织）
	ListByOrg(ctx context.Context, orgPath string) ([]*model.Employee, error)

	// ListByOrgDirect 列出指定组织的直属员工
	ListByOrgDirect(ctx context.Context, orgID uuid.UUID) ([]*model.Employee, error)

	// ListByPosition 列出指定职位的员工
	ListByPosition(ctx context.Context, positionID uuid.UUID) ([]*model.Employee, error)

	// ListByStatus 列出指定状态的员工
	ListByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*model.Employee, error)

	// ListByLeader 列出指定上级的员工
	ListByLeader(ctx context.Context, leaderID uuid.UUID) ([]*model.Employee, error)

	// CountByOrg 统计组织员工数量（包含子组织）
	CountByOrg(ctx context.Context, orgPath string) (int, error)

	// CountByOrgDirect 统计组织直属员工数量
	CountByOrgDirect(ctx context.Context, orgID uuid.UUID) (int, error)

	// UpdateOrg 更新员工组织
	UpdateOrg(ctx context.Context, empID, orgID uuid.UUID, orgPath string) error

	// UpdatePosition 更新员工职位
	UpdatePosition(ctx context.Context, empID uuid.UUID, positionID *uuid.UUID) error

	// UpdateStatus 更新员工状态
	UpdateStatus(ctx context.Context, empID uuid.UUID, status string) error

	// Exists 检查员工是否存在
	Exists(ctx context.Context, tenantID uuid.UUID, employeeNo string) (bool, error)
}

type employeeRepo struct {
	db *database.DB
}

// NewEmployeeRepository 创建员工仓储
func NewEmployeeRepository(db *database.DB) EmployeeRepository {
	return &employeeRepo{db: db}
}

func (r *employeeRepo) Create(ctx context.Context, emp *model.Employee) error {
	sql := `
		INSERT INTO employees (
			id, tenant_id, user_id, employee_no, name, gender, mobile, email, avatar,
			org_id, org_path, position_id, superior_id,
			status, created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15, $16, $17, $18
		)
	`

	_, err := r.db.Exec(ctx, sql,
		emp.ID, emp.TenantID, emp.UserID, emp.EmployeeNo, emp.Name, emp.Gender, emp.Mobile, emp.Email, emp.Avatar,
		emp.OrgID, emp.OrgPath, emp.PositionID, emp.DirectLeaderID,
		emp.Status, emp.CreatedBy, emp.UpdatedBy, emp.CreatedAt, emp.UpdatedAt,
	)

	return err
}

func (r *employeeRepo) Update(ctx context.Context, emp *model.Employee) error {
	sql := `
		UPDATE employees SET
			name = $1, gender = $2, mobile = $3, email = $4, avatar = $5,
			direct_leader_id = $6,
			join_date = $7, probation_end = $8, formal_date = $9, leave_date = $10,
			status = $11, updated_by = $12, updated_at = $13
		WHERE id = $14 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql,
		emp.Name, emp.Gender, emp.Mobile, emp.Email, emp.Avatar,
		emp.DirectLeaderID,
		emp.JoinDate, emp.ProbationEnd, emp.FormalDate, emp.LeaveDate,
		emp.Status, emp.UpdatedBy, emp.UpdatedAt,
		emp.ID,
	)

	return err
}

func (r *employeeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `UPDATE employees SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *employeeRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Employee, error) {
	sql := `
		SELECT id, tenant_id, user_id, employee_no, name, gender, mobile, email, avatar,
		       org_id, org_path, position_id, direct_leader_id,
		       join_date, probation_end, formal_date, leave_date,
		       status, created_by, updated_by, created_at, updated_at, deleted_at
		FROM employees
		WHERE id = $1 AND deleted_at IS NULL
	`

	emp := &model.Employee{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&emp.ID, &emp.TenantID, &emp.UserID, &emp.EmployeeNo, &emp.Name, &emp.Gender, &emp.Mobile, &emp.Email, &emp.Avatar,
		&emp.OrgID, &emp.OrgPath, &emp.PositionID, &emp.DirectLeaderID,
		&emp.JoinDate, &emp.ProbationEnd, &emp.FormalDate, &emp.LeaveDate,
		&emp.Status, &emp.CreatedBy, &emp.UpdatedBy, &emp.CreatedAt, &emp.UpdatedAt, &emp.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, err
	}

	return emp, nil
}

func (r *employeeRepo) GetByUserID(ctx context.Context, tenantID, userID uuid.UUID) (*model.Employee, error) {
	sql := `
		SELECT id, tenant_id, user_id, employee_no, name, gender, mobile, email, avatar,
		       org_id, org_path, position_id, direct_leader_id,
		       join_date, probation_end, formal_date, leave_date,
		       status, created_by, updated_by, created_at, updated_at, deleted_at
		FROM employees
		WHERE tenant_id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	emp := &model.Employee{}
	err := r.db.QueryRow(ctx, sql, tenantID, userID).Scan(
		&emp.ID, &emp.TenantID, &emp.UserID, &emp.EmployeeNo, &emp.Name, &emp.Gender, &emp.Mobile, &emp.Email, &emp.Avatar,
		&emp.OrgID, &emp.OrgPath, &emp.PositionID, &emp.DirectLeaderID,
		&emp.JoinDate, &emp.ProbationEnd, &emp.FormalDate, &emp.LeaveDate,
		&emp.Status, &emp.CreatedBy, &emp.UpdatedBy, &emp.CreatedAt, &emp.UpdatedAt, &emp.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, err
	}

	return emp, nil
}

func (r *employeeRepo) GetByEmployeeNo(ctx context.Context, tenantID uuid.UUID, employeeNo string) (*model.Employee, error) {
	sql := `
		SELECT id, tenant_id, user_id, employee_no, name, gender, mobile, email, avatar,
		       org_id, org_path, position_id, direct_leader_id,
		       join_date, probation_end, formal_date, leave_date,
		       status, created_by, updated_by, created_at, updated_at, deleted_at
		FROM employees
		WHERE tenant_id = $1 AND employee_no = $2 AND deleted_at IS NULL
	`

	emp := &model.Employee{}
	err := r.db.QueryRow(ctx, sql, tenantID, employeeNo).Scan(
		&emp.ID, &emp.TenantID, &emp.UserID, &emp.EmployeeNo, &emp.Name, &emp.Gender, &emp.Mobile, &emp.Email, &emp.Avatar,
		&emp.OrgID, &emp.OrgPath, &emp.PositionID, &emp.DirectLeaderID,
		&emp.JoinDate, &emp.ProbationEnd, &emp.FormalDate, &emp.LeaveDate,
		&emp.Status, &emp.CreatedBy, &emp.UpdatedBy, &emp.CreatedAt, &emp.UpdatedAt, &emp.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, err
	}

	return emp, nil
}

func (r *employeeRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*model.Employee, error) {
	sql := `
		SELECT id, tenant_id, user_id, employee_no, name, gender, mobile, email,
		       org_id, org_path, position_id, status, join_date, created_at
		FROM employees
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY employee_no ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emps []*model.Employee
	for rows.Next() {
		emp := &model.Employee{}
		err := rows.Scan(
			&emp.ID, &emp.TenantID, &emp.UserID, &emp.EmployeeNo, &emp.Name, &emp.Gender, &emp.Mobile, &emp.Email,
			&emp.OrgID, &emp.OrgPath, &emp.PositionID, &emp.Status, &emp.JoinDate, &emp.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		emps = append(emps, emp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return emps, nil
}

func (r *employeeRepo) ListByOrg(ctx context.Context, orgPath string) ([]*model.Employee, error) {
	sql := `
		SELECT id, tenant_id, user_id, employee_no, name, gender, mobile, email,
		       org_id, org_path, position_id, status, join_date, created_at
		FROM employees
		WHERE org_path LIKE $1 AND status IN ('active', 'probation') AND deleted_at IS NULL
		ORDER BY employee_no ASC
	`

	rows, err := r.db.Query(ctx, sql, orgPath+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emps []*model.Employee
	for rows.Next() {
		emp := &model.Employee{}
		err := rows.Scan(
			&emp.ID, &emp.TenantID, &emp.UserID, &emp.EmployeeNo, &emp.Name, &emp.Gender, &emp.Mobile, &emp.Email,
			&emp.OrgID, &emp.OrgPath, &emp.PositionID, &emp.Status, &emp.JoinDate, &emp.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		emps = append(emps, emp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return emps, nil
}

func (r *employeeRepo) ListByOrgDirect(ctx context.Context, orgID uuid.UUID) ([]*model.Employee, error) {
	sql := `
		SELECT id, tenant_id, user_id, employee_no, name, gender, mobile, email,
		       org_id, org_path, position_id, status, join_date, created_at
		FROM employees
		WHERE org_id = $1 AND status IN ('active', 'probation') AND deleted_at IS NULL
		ORDER BY employee_no ASC
	`

	rows, err := r.db.Query(ctx, sql, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emps []*model.Employee
	for rows.Next() {
		emp := &model.Employee{}
		err := rows.Scan(
			&emp.ID, &emp.TenantID, &emp.UserID, &emp.EmployeeNo, &emp.Name, &emp.Gender, &emp.Mobile, &emp.Email,
			&emp.OrgID, &emp.OrgPath, &emp.PositionID, &emp.Status, &emp.JoinDate, &emp.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		emps = append(emps, emp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return emps, nil
}

func (r *employeeRepo) ListByPosition(ctx context.Context, positionID uuid.UUID) ([]*model.Employee, error) {
	sql := `
		SELECT id, tenant_id, user_id, employee_no, name, gender, mobile, email,
		       org_id, position_id, status, join_date
		FROM employees
		WHERE position_id = $1 AND status IN ('active', 'probation') AND deleted_at IS NULL
		ORDER BY employee_no ASC
	`

	rows, err := r.db.Query(ctx, sql, positionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emps []*model.Employee
	for rows.Next() {
		emp := &model.Employee{}
		err := rows.Scan(
			&emp.ID, &emp.TenantID, &emp.UserID, &emp.EmployeeNo, &emp.Name, &emp.Gender, &emp.Mobile, &emp.Email,
			&emp.OrgID, &emp.PositionID, &emp.Status, &emp.JoinDate,
		)
		if err != nil {
			return nil, err
		}
		emps = append(emps, emp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return emps, nil
}

func (r *employeeRepo) ListByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*model.Employee, error) {
	sql := `
		SELECT id, tenant_id, user_id, employee_no, name, gender, mobile, email,
		       org_id, position_id, status, join_date
		FROM employees
		WHERE tenant_id = $1 AND status = $2 AND deleted_at IS NULL
		ORDER BY employee_no ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emps []*model.Employee
	for rows.Next() {
		emp := &model.Employee{}
		err := rows.Scan(
			&emp.ID, &emp.TenantID, &emp.UserID, &emp.EmployeeNo, &emp.Name, &emp.Gender, &emp.Mobile, &emp.Email,
			&emp.OrgID, &emp.PositionID, &emp.Status, &emp.JoinDate,
		)
		if err != nil {
			return nil, err
		}
		emps = append(emps, emp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return emps, nil
}

func (r *employeeRepo) ListByLeader(ctx context.Context, leaderID uuid.UUID) ([]*model.Employee, error) {
	sql := `
		SELECT id, tenant_id, user_id, employee_no, name, gender, mobile, email,
		       org_id, position_id, status, join_date
		FROM employees
		WHERE direct_leader_id = $1 AND status IN ('active', 'probation') AND deleted_at IS NULL
		ORDER BY employee_no ASC
	`

	rows, err := r.db.Query(ctx, sql, leaderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emps []*model.Employee
	for rows.Next() {
		emp := &model.Employee{}
		err := rows.Scan(
			&emp.ID, &emp.TenantID, &emp.UserID, &emp.EmployeeNo, &emp.Name, &emp.Gender, &emp.Mobile, &emp.Email,
			&emp.OrgID, &emp.PositionID, &emp.Status, &emp.JoinDate,
		)
		if err != nil {
			return nil, err
		}
		emps = append(emps, emp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return emps, nil
}

func (r *employeeRepo) CountByOrg(ctx context.Context, orgPath string) (int, error) {
	sql := `
		SELECT COUNT(*) FROM employees
		WHERE org_path LIKE $1 AND status IN ('active', 'probation') AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, sql, orgPath+"%").Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *employeeRepo) CountByOrgDirect(ctx context.Context, orgID uuid.UUID) (int, error) {
	sql := `
		SELECT COUNT(*) FROM employees
		WHERE org_id = $1 AND status IN ('active', 'probation') AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, sql, orgID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *employeeRepo) UpdateOrg(ctx context.Context, empID, orgID uuid.UUID, orgPath string) error {
	sql := `UPDATE employees SET org_id = $1, org_path = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(ctx, sql, orgID, orgPath, empID)
	return err
}

func (r *employeeRepo) UpdatePosition(ctx context.Context, empID uuid.UUID, positionID *uuid.UUID) error {
	sql := `UPDATE employees SET position_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, sql, positionID, empID)
	return err
}

func (r *employeeRepo) UpdateStatus(ctx context.Context, empID uuid.UUID, status string) error {
	sql := `UPDATE employees SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, sql, status, empID)
	return err
}

func (r *employeeRepo) Exists(ctx context.Context, tenantID uuid.UUID, employeeNo string) (bool, error) {
	sql := `
		SELECT COUNT(*) FROM employees
		WHERE tenant_id = $1 AND employee_no = $2 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, sql, tenantID, employeeNo).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
