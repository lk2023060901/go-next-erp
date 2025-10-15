package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

type leaveOfficeRepository struct {
	db *database.DB
}

// NewLeaveOfficeRepository 创建外出仓储实例
func NewLeaveOfficeRepository(db *database.DB) repository.LeaveOfficeRepository {
	return &leaveOfficeRepository{db: db}
}

// Create 创建外出记录
func (r *leaveOfficeRepository) Create(ctx context.Context, leaveOffice *model.LeaveOffice) error {
	query := `
		INSERT INTO hrm_leave_offices (
			id, tenant_id, employee_id, employee_name, department_id,
			start_time, end_time, duration,
			destination, purpose, contact,
			approval_id, approval_status, approved_by, approved_at, reject_reason,
			remark, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15, $16,
			$17, $18, $19
		)
	`

	_, err := r.db.Exec(ctx, query,
		leaveOffice.ID,
		leaveOffice.TenantID,
		leaveOffice.EmployeeID,
		leaveOffice.EmployeeName,
		leaveOffice.DepartmentID,
		leaveOffice.StartTime,
		leaveOffice.EndTime,
		leaveOffice.Duration,
		leaveOffice.Destination,
		leaveOffice.Purpose,
		leaveOffice.Contact,
		leaveOffice.ApprovalID,
		leaveOffice.ApprovalStatus,
		leaveOffice.ApprovedBy,
		leaveOffice.ApprovedAt,
		leaveOffice.RejectReason,
		leaveOffice.Remark,
		leaveOffice.CreatedAt,
		leaveOffice.UpdatedAt,
	)

	return err
}

// FindByID 根据ID查找外出记录
func (r *leaveOfficeRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveOffice, error) {
	query := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
		       start_time, end_time, duration,
		       destination, purpose, contact,
		       approval_id, approval_status, approved_by, approved_at, reject_reason,
		       remark, created_at, updated_at, deleted_at
		FROM hrm_leave_offices
		WHERE id = $1 AND deleted_at IS NULL
	`

	leaveOffice := &model.LeaveOffice{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&leaveOffice.ID,
		&leaveOffice.TenantID,
		&leaveOffice.EmployeeID,
		&leaveOffice.EmployeeName,
		&leaveOffice.DepartmentID,
		&leaveOffice.StartTime,
		&leaveOffice.EndTime,
		&leaveOffice.Duration,
		&leaveOffice.Destination,
		&leaveOffice.Purpose,
		&leaveOffice.Contact,
		&leaveOffice.ApprovalID,
		&leaveOffice.ApprovalStatus,
		&leaveOffice.ApprovedBy,
		&leaveOffice.ApprovedAt,
		&leaveOffice.RejectReason,
		&leaveOffice.Remark,
		&leaveOffice.CreatedAt,
		&leaveOffice.UpdatedAt,
		&leaveOffice.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return leaveOffice, nil
}

// Update 更新外出记录
func (r *leaveOfficeRepository) Update(ctx context.Context, leaveOffice *model.LeaveOffice) error {
	query := `
		UPDATE hrm_leave_offices
		SET start_time = $2,
		    end_time = $3,
		    duration = $4,
		    destination = $5,
		    purpose = $6,
		    contact = $7,
		    approval_id = $8,
		    approval_status = $9,
		    approved_by = $10,
		    approved_at = $11,
		    reject_reason = $12,
		    remark = $13,
		    updated_at = $14
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query,
		leaveOffice.ID,
		leaveOffice.StartTime,
		leaveOffice.EndTime,
		leaveOffice.Duration,
		leaveOffice.Destination,
		leaveOffice.Purpose,
		leaveOffice.Contact,
		leaveOffice.ApprovalID,
		leaveOffice.ApprovalStatus,
		leaveOffice.ApprovedBy,
		leaveOffice.ApprovedAt,
		leaveOffice.RejectReason,
		leaveOffice.Remark,
		leaveOffice.UpdatedAt,
	)

	return err
}

// Delete 删除外出记录（软删除）
func (r *leaveOfficeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE hrm_leave_offices
		SET deleted_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, id, time.Now())
	return err
}

// List 列表查询（分页）
func (r *leaveOfficeRepository) List(ctx context.Context, tenantID uuid.UUID, filter *repository.LeaveOfficeFilter, offset, limit int) ([]*model.LeaveOffice, int, error) {
	where := "tenant_id = $1 AND deleted_at IS NULL"
	args := []interface{}{tenantID}
	argIdx := 2

	if filter != nil {
		if filter.EmployeeID != nil {
			where += fmt.Sprintf(" AND employee_id = $%d", argIdx)
			args = append(args, *filter.EmployeeID)
			argIdx++
		}
		if filter.DepartmentID != nil {
			where += fmt.Sprintf(" AND department_id = $%d", argIdx)
			args = append(args, *filter.DepartmentID)
			argIdx++
		}
		if filter.ApprovalStatus != nil {
			where += fmt.Sprintf(" AND approval_status = $%d", argIdx)
			args = append(args, *filter.ApprovalStatus)
			argIdx++
		}
		if filter.StartDate != nil {
			where += fmt.Sprintf(" AND start_time >= $%d", argIdx)
			args = append(args, *filter.StartDate)
			argIdx++
		}
		if filter.EndDate != nil {
			where += fmt.Sprintf(" AND end_time <= $%d", argIdx)
			args = append(args, *filter.EndDate)
			argIdx++
		}
		if filter.Keyword != "" {
			where += fmt.Sprintf(" AND (employee_name LIKE $%d OR destination LIKE $%d OR purpose LIKE $%d)", argIdx, argIdx, argIdx)
			args = append(args, "%"+filter.Keyword+"%")
			argIdx++
		}
	}

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM hrm_leave_offices WHERE %s", where)
	var total int
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	dataSQL := fmt.Sprintf(`
		SELECT id, tenant_id, employee_id, employee_name, department_id,
		       start_time, end_time, duration,
		       destination, purpose, contact,
		       approval_status, created_at, updated_at
		FROM hrm_leave_offices
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	leaveOffices := make([]*model.LeaveOffice, 0)
	for rows.Next() {
		lo := &model.LeaveOffice{}
		err := rows.Scan(
			&lo.ID,
			&lo.TenantID,
			&lo.EmployeeID,
			&lo.EmployeeName,
			&lo.DepartmentID,
			&lo.StartTime,
			&lo.EndTime,
			&lo.Duration,
			&lo.Destination,
			&lo.Purpose,
			&lo.Contact,
			&lo.ApprovalStatus,
			&lo.CreatedAt,
			&lo.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		leaveOffices = append(leaveOffices, lo)
	}

	return leaveOffices, total, nil
}

// FindByEmployee 查询员工外出记录
func (r *leaveOfficeRepository) FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.LeaveOffice, error) {
	query := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
		       start_time, end_time, duration,
		       destination, purpose, contact,
		       approval_status, created_at, updated_at
		FROM hrm_leave_offices
		WHERE tenant_id = $1 AND employee_id = $2
		      AND EXTRACT(YEAR FROM start_time) = $3
		      AND deleted_at IS NULL
		ORDER BY start_time DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID, employeeID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leaveOffices := make([]*model.LeaveOffice, 0)
	for rows.Next() {
		lo := &model.LeaveOffice{}
		err := rows.Scan(
			&lo.ID,
			&lo.TenantID,
			&lo.EmployeeID,
			&lo.EmployeeName,
			&lo.DepartmentID,
			&lo.StartTime,
			&lo.EndTime,
			&lo.Duration,
			&lo.Destination,
			&lo.Purpose,
			&lo.Contact,
			&lo.ApprovalStatus,
			&lo.CreatedAt,
			&lo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		leaveOffices = append(leaveOffices, lo)
	}

	return leaveOffices, nil
}

// FindPending 查询待审批的外出
func (r *leaveOfficeRepository) FindPending(ctx context.Context, tenantID uuid.UUID) ([]*model.LeaveOffice, error) {
	query := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
		       start_time, end_time, duration,
		       destination, purpose, contact,
		       approval_status, created_at, updated_at
		FROM hrm_leave_offices
		WHERE tenant_id = $1 AND approval_status = 'pending' AND deleted_at IS NULL
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leaveOffices := make([]*model.LeaveOffice, 0)
	for rows.Next() {
		lo := &model.LeaveOffice{}
		err := rows.Scan(
			&lo.ID,
			&lo.TenantID,
			&lo.EmployeeID,
			&lo.EmployeeName,
			&lo.DepartmentID,
			&lo.StartTime,
			&lo.EndTime,
			&lo.Duration,
			&lo.Destination,
			&lo.Purpose,
			&lo.Contact,
			&lo.ApprovalStatus,
			&lo.CreatedAt,
			&lo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		leaveOffices = append(leaveOffices, lo)
	}

	return leaveOffices, nil
}

// FindOverlapping 查询时间重叠的外出记录（用于时间冲突检测）
func (r *leaveOfficeRepository) FindOverlapping(ctx context.Context, tenantID, employeeID uuid.UUID, startTime, endTime time.Time) ([]*model.LeaveOffice, error) {
	query := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
		       start_time, end_time, duration,
		       destination, purpose, contact,
		       approval_status, created_at, updated_at
		FROM hrm_leave_offices
		WHERE tenant_id = $1 AND employee_id = $2
			AND approval_status IN ('pending', 'approved')
			AND deleted_at IS NULL
			AND (
				(start_time <= $3 AND end_time >= $3) OR
				(start_time <= $4 AND end_time >= $4) OR
				(start_time >= $3 AND end_time <= $4)
			)
		ORDER BY start_time
	`

	rows, err := r.db.Query(ctx, query, tenantID, employeeID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leaveOffices := make([]*model.LeaveOffice, 0)
	for rows.Next() {
		lo := &model.LeaveOffice{}
		err := rows.Scan(
			&lo.ID,
			&lo.TenantID,
			&lo.EmployeeID,
			&lo.EmployeeName,
			&lo.DepartmentID,
			&lo.StartTime,
			&lo.EndTime,
			&lo.Duration,
			&lo.Destination,
			&lo.Purpose,
			&lo.Contact,
			&lo.ApprovalStatus,
			&lo.CreatedAt,
			&lo.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		leaveOffices = append(leaveOffices, lo)
	}

	return leaveOffices, nil
}
