package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

type scheduleRepo struct {
	db *database.DB
}

// NewScheduleRepository 创建排班仓储
func NewScheduleRepository(db *database.DB) repository.ScheduleRepository {
	return &scheduleRepo{db: db}
}

func (r *scheduleRepo) Create(ctx context.Context, schedule *model.Schedule) error {
	sql := `
		INSERT INTO hrm_schedules (
			id, tenant_id, employee_id, employee_name, department_id,
			shift_id, shift_name, schedule_date, workday_type,
			status, remark,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11,
			$12, $13, $14, $15
		)
	`

	_, err := r.db.Exec(ctx, sql,
		schedule.ID, schedule.TenantID, schedule.EmployeeID, schedule.EmployeeName, schedule.DepartmentID,
		schedule.ShiftID, schedule.ShiftName, schedule.ScheduleDate, schedule.WorkdayType,
		schedule.Status, schedule.Remark,
		schedule.CreatedBy, schedule.UpdatedBy, schedule.CreatedAt, schedule.UpdatedAt,
	)

	return err
}

func (r *scheduleRepo) BatchCreate(ctx context.Context, schedules []*model.Schedule) error {
	if len(schedules) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql := `
		INSERT INTO hrm_schedules (
			id, tenant_id, employee_id, employee_name, department_id,
			shift_id, shift_name, schedule_date, workday_type,
			status, remark, created_by, updated_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	for _, schedule := range schedules {
		_, err := tx.Exec(ctx, sql,
			schedule.ID, schedule.TenantID, schedule.EmployeeID, schedule.EmployeeName, schedule.DepartmentID,
			schedule.ShiftID, schedule.ShiftName, schedule.ScheduleDate, schedule.WorkdayType,
			schedule.Status, schedule.Remark,
			schedule.CreatedBy, schedule.UpdatedBy, schedule.CreatedAt, schedule.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *scheduleRepo) Update(ctx context.Context, schedule *model.Schedule) error {
	sql := `
		UPDATE hrm_schedules SET
			shift_id = $1, shift_name = $2, workday_type = $3,
			status = $4, remark = $5, updated_by = $6, updated_at = $7
		WHERE id = $8 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql,
		schedule.ShiftID, schedule.ShiftName, schedule.WorkdayType,
		schedule.Status, schedule.Remark, schedule.UpdatedBy, schedule.UpdatedAt,
		schedule.ID,
	)

	return err
}

func (r *scheduleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `UPDATE hrm_schedules SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *scheduleRepo) BatchDelete(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	sql := `UPDATE hrm_schedules SET deleted_at = NOW() WHERE id = ANY($1)`
	_, err := r.db.Exec(ctx, sql, ids)
	return err
}

func (r *scheduleRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error) {
	sql := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
		       shift_id, shift_name, schedule_date, workday_type,
		       status, remark,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM hrm_schedules
		WHERE id = $1 AND deleted_at IS NULL
	`

	schedule := &model.Schedule{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&schedule.ID, &schedule.TenantID, &schedule.EmployeeID, &schedule.EmployeeName, &schedule.DepartmentID,
		&schedule.ShiftID, &schedule.ShiftName, &schedule.ScheduleDate, &schedule.WorkdayType,
		&schedule.Status, &schedule.Remark,
		&schedule.CreatedBy, &schedule.UpdatedBy, &schedule.CreatedAt, &schedule.UpdatedAt, &schedule.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("schedule not found")
		}
		return nil, err
	}

	return schedule, nil
}

func (r *scheduleRepo) List(ctx context.Context, tenantID uuid.UUID, filter *repository.ScheduleFilter, offset, limit int) ([]*model.Schedule, int, error) {
	// 构建查询条件
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
		if filter.ShiftID != nil {
			where += fmt.Sprintf(" AND shift_id = $%d", argIdx)
			args = append(args, *filter.ShiftID)
			argIdx++
		}
		if filter.StartDate != nil {
			where += fmt.Sprintf(" AND schedule_date >= $%d", argIdx)
			args = append(args, *filter.StartDate)
			argIdx++
		}
		if filter.EndDate != nil {
			where += fmt.Sprintf(" AND schedule_date <= $%d", argIdx)
			args = append(args, *filter.EndDate)
			argIdx++
		}
		if filter.Status != nil {
			where += fmt.Sprintf(" AND status = $%d", argIdx)
			args = append(args, *filter.Status)
			argIdx++
		}
	}

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM hrm_schedules WHERE %s", where)
	var total int
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	dataSQL := fmt.Sprintf(`
		SELECT id, tenant_id, employee_id, employee_name, shift_id, shift_name,
		       schedule_date, workday_type, status, created_at
		FROM hrm_schedules
		WHERE %s
		ORDER BY schedule_date DESC, employee_name ASC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var schedules []*model.Schedule
	for rows.Next() {
		schedule := &model.Schedule{}
		err := rows.Scan(
			&schedule.ID, &schedule.TenantID, &schedule.EmployeeID, &schedule.EmployeeName,
			&schedule.ShiftID, &schedule.ShiftName,
			&schedule.ScheduleDate, &schedule.WorkdayType, &schedule.Status, &schedule.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, total, rows.Err()
}

// ListWithCursor 游标分页查询排班（高性能）
func (r *scheduleRepo) ListWithCursor(
	ctx context.Context,
	tenantID uuid.UUID,
	filter *repository.ScheduleFilter,
	cursor *time.Time,
	limit int,
) ([]*model.Schedule, *time.Time, bool, error) {
	// 构建 WHERE 条件
	where := "tenant_id = $1 AND deleted_at IS NULL"
	args := []interface{}{tenantID}
	argIdx := 1

	// 添加游标条件
	if cursor != nil {
		argIdx++
		where += fmt.Sprintf(" AND created_at < $%d", argIdx)
		args = append(args, *cursor)
	}

	// 添加过滤条件
	if filter != nil {
		if filter.EmployeeID != nil {
			argIdx++
			where += fmt.Sprintf(" AND employee_id = $%d", argIdx)
			args = append(args, *filter.EmployeeID)
		}
		if filter.DepartmentID != nil {
			argIdx++
			where += fmt.Sprintf(" AND department_id = $%d", argIdx)
			args = append(args, *filter.DepartmentID)
		}
		if filter.ShiftID != nil {
			argIdx++
			where += fmt.Sprintf(" AND shift_id = $%d", argIdx)
			args = append(args, *filter.ShiftID)
		}
		if filter.StartDate != nil {
			argIdx++
			where += fmt.Sprintf(" AND schedule_date >= $%d", argIdx)
			args = append(args, *filter.StartDate)
		}
		if filter.EndDate != nil {
			argIdx++
			where += fmt.Sprintf(" AND schedule_date <= $%d", argIdx)
			args = append(args, *filter.EndDate)
		}
		if filter.Status != nil {
			argIdx++
			where += fmt.Sprintf(" AND status = $%d", argIdx)
			args = append(args, *filter.Status)
		}
	}

	// 构建查询（多查1条用于判断是否有下一页）
	argIdx++
	sql := fmt.Sprintf(`
		SELECT id, tenant_id, employee_id, employee_name, department_id,
		       shift_id, shift_name, schedule_date, workday_type, status, created_at
		FROM hrm_schedules
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
	`, where, argIdx)
	args = append(args, limit+1)

	// 执行查询
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, nil, false, err
	}
	defer rows.Close()

	// 扫描结果
	var schedules []*model.Schedule
	for rows.Next() {
		schedule := &model.Schedule{}
		err := rows.Scan(
			&schedule.ID, &schedule.TenantID, &schedule.EmployeeID, &schedule.EmployeeName, &schedule.DepartmentID,
			&schedule.ShiftID, &schedule.ShiftName,
			&schedule.ScheduleDate, &schedule.WorkdayType, &schedule.Status, &schedule.CreatedAt,
		)
		if err != nil {
			return nil, nil, false, err
		}
		schedules = append(schedules, schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, false, err
	}

	// 判断是否有下一页
	hasNext := len(schedules) > limit
	if hasNext {
		schedules = schedules[:limit]
	}

	// 生成下一页游标
	var nextCursor *time.Time
	if hasNext && len(schedules) > 0 {
		lastSchedule := schedules[len(schedules)-1]
		nextCursor = &lastSchedule.CreatedAt
	}

	return schedules, nextCursor, hasNext, nil
}

func (r *scheduleRepo) FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, month string) ([]*model.Schedule, error) {
	// month 格式: "2025-01"
	startDate, _ := time.Parse("2006-01", month)
	endDate := startDate.AddDate(0, 1, 0)

	sql := `
		SELECT id, tenant_id, employee_id, shift_id, shift_name,
		       schedule_date, workday_type, status, created_at
		FROM hrm_schedules
		WHERE tenant_id = $1 AND employee_id = $2 
		  AND schedule_date >= $3 AND schedule_date < $4 
		  AND deleted_at IS NULL
		ORDER BY schedule_date ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, employeeID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*model.Schedule
	for rows.Next() {
		schedule := &model.Schedule{}
		err := rows.Scan(
			&schedule.ID, &schedule.TenantID, &schedule.EmployeeID,
			&schedule.ShiftID, &schedule.ShiftName,
			&schedule.ScheduleDate, &schedule.WorkdayType, &schedule.Status, &schedule.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, rows.Err()
}

func (r *scheduleRepo) FindByDepartment(ctx context.Context, tenantID, departmentID uuid.UUID, month string) ([]*model.Schedule, error) {
	startDate, _ := time.Parse("2006-01", month)
	endDate := startDate.AddDate(0, 1, 0)

	sql := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
		       shift_id, shift_name, schedule_date, workday_type, status
		FROM hrm_schedules
		WHERE tenant_id = $1 AND department_id = $2 
		  AND schedule_date >= $3 AND schedule_date < $4 
		  AND deleted_at IS NULL
		ORDER BY schedule_date ASC, employee_name ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, departmentID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*model.Schedule
	for rows.Next() {
		schedule := &model.Schedule{}
		err := rows.Scan(
			&schedule.ID, &schedule.TenantID, &schedule.EmployeeID, &schedule.EmployeeName, &schedule.DepartmentID,
			&schedule.ShiftID, &schedule.ShiftName,
			&schedule.ScheduleDate, &schedule.WorkdayType, &schedule.Status,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, rows.Err()
}

func (r *scheduleRepo) FindByDate(ctx context.Context, tenantID uuid.UUID, date string) ([]*model.Schedule, error) {
	scheduleDate, _ := time.Parse("2006-01-02", date)

	sql := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
		       shift_id, shift_name, schedule_date, workday_type, status
		FROM hrm_schedules
		WHERE tenant_id = $1 AND schedule_date = $2 AND deleted_at IS NULL
		ORDER BY employee_name ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, scheduleDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*model.Schedule
	for rows.Next() {
		schedule := &model.Schedule{}
		err := rows.Scan(
			&schedule.ID, &schedule.TenantID, &schedule.EmployeeID, &schedule.EmployeeName, &schedule.DepartmentID,
			&schedule.ShiftID, &schedule.ShiftName,
			&schedule.ScheduleDate, &schedule.WorkdayType, &schedule.Status,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, rows.Err()
}
