package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

type overtimeRepository struct {
	db *database.DB
}

// NewOvertimeRepository 创建加班仓储实例
func NewOvertimeRepository(db *database.DB) repository.OvertimeRepository {
	return &overtimeRepository{db: db}
}

// Create 创建加班记录
func (r *overtimeRepository) Create(ctx context.Context, overtime *model.Overtime) error {
	query := `
		INSERT INTO hrm_overtimes (
			id, tenant_id, employee_id, employee_name, department_id,
			start_time, end_time, duration, overtime_type, pay_type, pay_rate,
			reason, tasks, approval_status, comp_off_days, comp_off_used,
			remark, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
	`

	_, err := r.db.Exec(ctx, query,
		overtime.ID, overtime.TenantID, overtime.EmployeeID, overtime.EmployeeName, overtime.DepartmentID,
		overtime.StartTime, overtime.EndTime, overtime.Duration, overtime.OvertimeType, overtime.PayType, overtime.PayRate,
		overtime.Reason, overtime.Tasks, overtime.ApprovalStatus, overtime.CompOffDays, overtime.CompOffUsed,
		overtime.Remark, overtime.CreatedAt, overtime.UpdatedAt,
	)

	return err
}

// FindByID 根据ID查找
func (r *overtimeRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Overtime, error) {
	query := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
			start_time, end_time, duration, overtime_type, pay_type, pay_rate,
			reason, tasks, approval_id, approval_status, approved_by, approved_at, reject_reason,
			comp_off_days, comp_off_used, comp_off_expire_at,
			remark, created_at, updated_at, deleted_at
		FROM hrm_overtimes
		WHERE id = $1 AND deleted_at IS NULL
	`

	overtime := &model.Overtime{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&overtime.ID, &overtime.TenantID, &overtime.EmployeeID, &overtime.EmployeeName, &overtime.DepartmentID,
		&overtime.StartTime, &overtime.EndTime, &overtime.Duration, &overtime.OvertimeType, &overtime.PayType, &overtime.PayRate,
		&overtime.Reason, &overtime.Tasks, &overtime.ApprovalID, &overtime.ApprovalStatus, &overtime.ApprovedBy, &overtime.ApprovedAt, &overtime.RejectReason,
		&overtime.CompOffDays, &overtime.CompOffUsed, &overtime.CompOffExpireAt,
		&overtime.Remark, &overtime.CreatedAt, &overtime.UpdatedAt, &overtime.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return overtime, nil
}

// Update 更新加班记录
func (r *overtimeRepository) Update(ctx context.Context, overtime *model.Overtime) error {
	query := `
		UPDATE hrm_overtimes SET
			employee_name = $2,
			department_id = $3,
			start_time = $4,
			end_time = $5,
			duration = $6,
			overtime_type = $7,
			pay_type = $8,
			pay_rate = $9,
			reason = $10,
			tasks = $11,
			approval_id = $12,
			approval_status = $13,
			approved_by = $14,
			approved_at = $15,
			reject_reason = $16,
			comp_off_days = $17,
			comp_off_used = $18,
			comp_off_expire_at = $19,
			remark = $20,
			updated_at = $21
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query,
		overtime.ID, overtime.EmployeeName, overtime.DepartmentID,
		overtime.StartTime, overtime.EndTime, overtime.Duration, overtime.OvertimeType, overtime.PayType, overtime.PayRate,
		overtime.Reason, overtime.Tasks, overtime.ApprovalID, overtime.ApprovalStatus, overtime.ApprovedBy, overtime.ApprovedAt, overtime.RejectReason,
		overtime.CompOffDays, overtime.CompOffUsed, overtime.CompOffExpireAt,
		overtime.Remark, overtime.UpdatedAt,
	)

	return err
}

// Delete 删除加班记录（软删除）
func (r *overtimeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE hrm_overtimes SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, time.Now(), id)
	return err
}

// List 列表查询（分页）
func (r *overtimeRepository) List(ctx context.Context, tenantID uuid.UUID, filter *repository.OvertimeFilter, offset, limit int) ([]*model.Overtime, int, error) {
	conditions := []string{"tenant_id = $1", "deleted_at IS NULL"}
	args := []interface{}{tenantID}
	argIndex := 2

	if filter != nil {
		if filter.EmployeeID != nil {
			conditions = append(conditions, fmt.Sprintf("employee_id = $%d", argIndex))
			args = append(args, *filter.EmployeeID)
			argIndex++
		}
		if filter.DepartmentID != nil {
			conditions = append(conditions, fmt.Sprintf("department_id = $%d", argIndex))
			args = append(args, *filter.DepartmentID)
			argIndex++
		}
		if filter.OvertimeType != nil {
			conditions = append(conditions, fmt.Sprintf("overtime_type = $%d", argIndex))
			args = append(args, *filter.OvertimeType)
			argIndex++
		}
		if filter.ApprovalStatus != nil {
			conditions = append(conditions, fmt.Sprintf("approval_status = $%d", argIndex))
			args = append(args, *filter.ApprovalStatus)
			argIndex++
		}
		if filter.StartDate != nil {
			conditions = append(conditions, fmt.Sprintf("start_time >= $%d", argIndex))
			args = append(args, *filter.StartDate)
			argIndex++
		}
		if filter.EndDate != nil {
			conditions = append(conditions, fmt.Sprintf("end_time <= $%d", argIndex))
			args = append(args, *filter.EndDate)
			argIndex++
		}
		if filter.Keyword != "" {
			conditions = append(conditions, fmt.Sprintf("(employee_name ILIKE $%d OR reason ILIKE $%d)", argIndex, argIndex))
			args = append(args, "%"+filter.Keyword+"%")
			argIndex++
		}
	}

	whereClause := strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM hrm_overtimes WHERE %s", whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf(`
		SELECT id, tenant_id, employee_id, employee_name, department_id,
			start_time, end_time, duration, overtime_type, pay_type, pay_rate,
			reason, tasks, approval_id, approval_status, approved_by, approved_at, reject_reason,
			comp_off_days, comp_off_used, comp_off_expire_at,
			remark, created_at, updated_at, deleted_at
		FROM hrm_overtimes
		WHERE %s
		ORDER BY start_time DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	overtimes := make([]*model.Overtime, 0)
	for rows.Next() {
		overtime := &model.Overtime{}
		err := rows.Scan(
			&overtime.ID, &overtime.TenantID, &overtime.EmployeeID, &overtime.EmployeeName, &overtime.DepartmentID,
			&overtime.StartTime, &overtime.EndTime, &overtime.Duration, &overtime.OvertimeType, &overtime.PayType, &overtime.PayRate,
			&overtime.Reason, &overtime.Tasks, &overtime.ApprovalID, &overtime.ApprovalStatus, &overtime.ApprovedBy, &overtime.ApprovedAt, &overtime.RejectReason,
			&overtime.CompOffDays, &overtime.CompOffUsed, &overtime.CompOffExpireAt,
			&overtime.Remark, &overtime.CreatedAt, &overtime.UpdatedAt, &overtime.DeletedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		overtimes = append(overtimes, overtime)
	}

	return overtimes, total, rows.Err()
}

// FindByEmployee 查询员工加班记录
func (r *overtimeRepository) FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.Overtime, error) {
	query := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
			start_time, end_time, duration, overtime_type, pay_type, pay_rate,
			reason, tasks, approval_id, approval_status, approved_by, approved_at, reject_reason,
			comp_off_days, comp_off_used, comp_off_expire_at,
			remark, created_at, updated_at, deleted_at
		FROM hrm_overtimes
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

	overtimes := make([]*model.Overtime, 0)
	for rows.Next() {
		overtime := &model.Overtime{}
		err := rows.Scan(
			&overtime.ID, &overtime.TenantID, &overtime.EmployeeID, &overtime.EmployeeName, &overtime.DepartmentID,
			&overtime.StartTime, &overtime.EndTime, &overtime.Duration, &overtime.OvertimeType, &overtime.PayType, &overtime.PayRate,
			&overtime.Reason, &overtime.Tasks, &overtime.ApprovalID, &overtime.ApprovalStatus, &overtime.ApprovedBy, &overtime.ApprovedAt, &overtime.RejectReason,
			&overtime.CompOffDays, &overtime.CompOffUsed, &overtime.CompOffExpireAt,
			&overtime.Remark, &overtime.CreatedAt, &overtime.UpdatedAt, &overtime.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		overtimes = append(overtimes, overtime)
	}

	return overtimes, rows.Err()
}

// FindPending 查询待审批的加班
func (r *overtimeRepository) FindPending(ctx context.Context, tenantID uuid.UUID) ([]*model.Overtime, error) {
	query := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
			start_time, end_time, duration, overtime_type, pay_type, pay_rate,
			reason, tasks, approval_id, approval_status, approved_by, approved_at, reject_reason,
			comp_off_days, comp_off_used, comp_off_expire_at,
			remark, created_at, updated_at, deleted_at
		FROM hrm_overtimes
		WHERE tenant_id = $1 AND approval_status = 'pending' AND deleted_at IS NULL
		ORDER BY start_time DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	overtimes := make([]*model.Overtime, 0)
	for rows.Next() {
		overtime := &model.Overtime{}
		err := rows.Scan(
			&overtime.ID, &overtime.TenantID, &overtime.EmployeeID, &overtime.EmployeeName, &overtime.DepartmentID,
			&overtime.StartTime, &overtime.EndTime, &overtime.Duration, &overtime.OvertimeType, &overtime.PayType, &overtime.PayRate,
			&overtime.Reason, &overtime.Tasks, &overtime.ApprovalID, &overtime.ApprovalStatus, &overtime.ApprovedBy, &overtime.ApprovedAt, &overtime.RejectReason,
			&overtime.CompOffDays, &overtime.CompOffUsed, &overtime.CompOffExpireAt,
			&overtime.Remark, &overtime.CreatedAt, &overtime.UpdatedAt, &overtime.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		overtimes = append(overtimes, overtime)
	}

	return overtimes, rows.Err()
}

// SumHoursByEmployee 统计员工加班时长
func (r *overtimeRepository) SumHoursByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	query := `
		SELECT COALESCE(SUM(duration), 0)
		FROM hrm_overtimes
		WHERE tenant_id = $1 AND employee_id = $2
			AND start_time >= $3 AND end_time <= $4
			AND approval_status = 'approved'
			AND deleted_at IS NULL
	`

	var total float64
	err := r.db.QueryRow(ctx, query, tenantID, employeeID, startDate, endDate).Scan(&total)
	return total, err
}

// SumCompOffDays 统计可调休天数
func (r *overtimeRepository) SumCompOffDays(ctx context.Context, tenantID, employeeID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(comp_off_days - comp_off_used), 0)
		FROM hrm_overtimes
		WHERE tenant_id = $1 AND employee_id = $2
			AND approval_status = 'approved'
			AND pay_type = 'leave'
			AND (comp_off_expire_at IS NULL OR comp_off_expire_at > NOW())
			AND deleted_at IS NULL
	`

	var total float64
	err := r.db.QueryRow(ctx, query, tenantID, employeeID).Scan(&total)
	return total, err
}
