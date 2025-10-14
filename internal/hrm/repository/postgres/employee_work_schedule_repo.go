package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

type employeeWorkScheduleRepo struct {
	db *database.DB
}

// NewEmployeeWorkScheduleRepository 创建员工工作时间表仓储
func NewEmployeeWorkScheduleRepository(db *database.DB) repository.EmployeeWorkScheduleRepository {
	return &employeeWorkScheduleRepo{db: db}
}

func (r *employeeWorkScheduleRepo) Create(ctx context.Context, schedule *model.EmployeeWorkSchedule) error {
	workDaysJSON, _ := json.Marshal(schedule.WorkDays)

	sql := `
		INSERT INTO hrm_employee_work_schedules (
			id, tenant_id, employee_id, schedule_type,
			work_days, work_hours, work_start, work_end,
			shift_cycle, shift_pattern,
			effective_from, effective_to, is_active, remark,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10,
			$11, $12, $13, $14,
			$15, $16, $17, $18
		)
	`

	_, err := r.db.Exec(ctx, sql,
		schedule.ID, schedule.TenantID, schedule.EmployeeID, schedule.ScheduleType,
		workDaysJSON, schedule.WorkHours, schedule.WorkStart, schedule.WorkEnd,
		schedule.ShiftCycle, schedule.ShiftPattern,
		schedule.EffectiveFrom, schedule.EffectiveTo, schedule.IsActive, schedule.Remark,
		schedule.CreatedBy, schedule.UpdatedBy, schedule.CreatedAt, schedule.UpdatedAt,
	)

	return err
}

func (r *employeeWorkScheduleRepo) Update(ctx context.Context, schedule *model.EmployeeWorkSchedule) error {
	workDaysJSON, _ := json.Marshal(schedule.WorkDays)

	sql := `
		UPDATE hrm_employee_work_schedules SET
			schedule_type = $1,
			work_days = $2, work_hours = $3, work_start = $4, work_end = $5,
			shift_cycle = $6, shift_pattern = $7,
			effective_from = $8, effective_to = $9, is_active = $10, remark = $11,
			updated_by = $12, updated_at = $13
		WHERE id = $14
	`

	_, err := r.db.Exec(ctx, sql,
		schedule.ScheduleType,
		workDaysJSON, schedule.WorkHours, schedule.WorkStart, schedule.WorkEnd,
		schedule.ShiftCycle, schedule.ShiftPattern,
		schedule.EffectiveFrom, schedule.EffectiveTo, schedule.IsActive, schedule.Remark,
		schedule.UpdatedBy, schedule.UpdatedAt,
		schedule.ID,
	)

	return err
}

func (r *employeeWorkScheduleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `DELETE FROM hrm_employee_work_schedules WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *employeeWorkScheduleRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.EmployeeWorkSchedule, error) {
	sql := `
		SELECT id, tenant_id, employee_id, schedule_type,
		       work_days, work_hours, work_start, work_end,
		       shift_cycle, shift_pattern,
		       effective_from, effective_to, is_active, remark,
		       created_by, updated_by, created_at, updated_at
		FROM hrm_employee_work_schedules
		WHERE id = $1
	`

	schedule := &model.EmployeeWorkSchedule{}
	var workDaysJSON []byte

	err := r.db.QueryRow(ctx, sql, id).Scan(
		&schedule.ID, &schedule.TenantID, &schedule.EmployeeID, &schedule.ScheduleType,
		&workDaysJSON, &schedule.WorkHours, &schedule.WorkStart, &schedule.WorkEnd,
		&schedule.ShiftCycle, &schedule.ShiftPattern,
		&schedule.EffectiveFrom, &schedule.EffectiveTo, &schedule.IsActive, &schedule.Remark,
		&schedule.CreatedBy, &schedule.UpdatedBy, &schedule.CreatedAt, &schedule.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("employee work schedule not found")
		}
		return nil, err
	}

	if len(workDaysJSON) > 0 {
		json.Unmarshal(workDaysJSON, &schedule.WorkDays)
	}

	return schedule, nil
}

func (r *employeeWorkScheduleRepo) FindActiveByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) (*model.EmployeeWorkSchedule, error) {
	now := time.Now()

	sql := `
		SELECT id, tenant_id, employee_id, schedule_type,
		       work_days, work_hours, work_start, work_end,
		       shift_cycle, shift_pattern,
		       effective_from, effective_to, is_active, created_at
		FROM hrm_employee_work_schedules
		WHERE tenant_id = $1 AND employee_id = $2 
		  AND is_active = TRUE
		  AND effective_from <= $3
		  AND (effective_to IS NULL OR effective_to >= $3)
		ORDER BY effective_from DESC
		LIMIT 1
	`

	schedule := &model.EmployeeWorkSchedule{}
	var workDaysJSON []byte

	err := r.db.QueryRow(ctx, sql, tenantID, employeeID, now).Scan(
		&schedule.ID, &schedule.TenantID, &schedule.EmployeeID, &schedule.ScheduleType,
		&workDaysJSON, &schedule.WorkHours, &schedule.WorkStart, &schedule.WorkEnd,
		&schedule.ShiftCycle, &schedule.ShiftPattern,
		&schedule.EffectiveFrom, &schedule.EffectiveTo, &schedule.IsActive, &schedule.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no active work schedule found for employee")
		}
		return nil, err
	}

	if len(workDaysJSON) > 0 {
		json.Unmarshal(workDaysJSON, &schedule.WorkDays)
	}

	return schedule, nil
}

func (r *employeeWorkScheduleRepo) ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) ([]*model.EmployeeWorkSchedule, error) {
	sql := `
		SELECT id, tenant_id, employee_id, schedule_type,
		       work_hours, work_start, work_end,
		       effective_from, effective_to, is_active, created_at
		FROM hrm_employee_work_schedules
		WHERE tenant_id = $1 AND employee_id = $2
		ORDER BY effective_from DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*model.EmployeeWorkSchedule
	for rows.Next() {
		schedule := &model.EmployeeWorkSchedule{}
		err := rows.Scan(
			&schedule.ID, &schedule.TenantID, &schedule.EmployeeID, &schedule.ScheduleType,
			&schedule.WorkHours, &schedule.WorkStart, &schedule.WorkEnd,
			&schedule.EffectiveFrom, &schedule.EffectiveTo, &schedule.IsActive, &schedule.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, rows.Err()
}

func (r *employeeWorkScheduleRepo) ListByType(ctx context.Context, tenantID uuid.UUID, scheduleType string) ([]*model.EmployeeWorkSchedule, error) {
	sql := `
		SELECT id, tenant_id, employee_id, schedule_type,
		       work_hours, effective_from, is_active, created_at
		FROM hrm_employee_work_schedules
		WHERE tenant_id = $1 AND schedule_type = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, scheduleType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*model.EmployeeWorkSchedule
	for rows.Next() {
		schedule := &model.EmployeeWorkSchedule{}
		err := rows.Scan(
			&schedule.ID, &schedule.TenantID, &schedule.EmployeeID, &schedule.ScheduleType,
			&schedule.WorkHours, &schedule.EffectiveFrom, &schedule.IsActive, &schedule.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, rows.Err()
}

func (r *employeeWorkScheduleRepo) ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.EmployeeWorkSchedule, error) {
	now := time.Now()

	sql := `
		SELECT id, tenant_id, employee_id, schedule_type,
		       work_hours, work_start, work_end, effective_from, created_at
		FROM hrm_employee_work_schedules
		WHERE tenant_id = $1 
		  AND is_active = TRUE
		  AND effective_from <= $2
		  AND (effective_to IS NULL OR effective_to >= $2)
		ORDER BY employee_id ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []*model.EmployeeWorkSchedule
	for rows.Next() {
		schedule := &model.EmployeeWorkSchedule{}
		err := rows.Scan(
			&schedule.ID, &schedule.TenantID, &schedule.EmployeeID, &schedule.ScheduleType,
			&schedule.WorkHours, &schedule.WorkStart, &schedule.WorkEnd, &schedule.EffectiveFrom, &schedule.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, rows.Err()
}

func (r *employeeWorkScheduleRepo) DeactivateOld(ctx context.Context, tenantID, employeeID uuid.UUID) error {
	sql := `
		UPDATE hrm_employee_work_schedules 
		SET is_active = FALSE, updated_at = NOW()
		WHERE tenant_id = $1 AND employee_id = $2 AND is_active = TRUE
	`
	_, err := r.db.Exec(ctx, sql, tenantID, employeeID)
	return err
}
