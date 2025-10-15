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

type shiftRepo struct {
	db *database.DB
}

// NewShiftRepository 创建班次仓储
func NewShiftRepository(db *database.DB) repository.ShiftRepository {
	return &shiftRepo{db: db}
}

func (r *shiftRepo) Create(ctx context.Context, shift *model.Shift) error {
	restPeriodsJSON, _ := json.Marshal(shift.RestPeriods)

	sql := `
		INSERT INTO hrm_shifts (
			id, tenant_id, code, name, type, description,
			work_start, work_end, flexible_start, flexible_end, work_duration,
			check_in_required, check_out_required,
			late_grace_period, early_grace_period,
			rest_periods, is_cross_days,
			allow_overtime, overtime_start_buffer, overtime_min_duration, overtime_pay_rate,
			workday_types, color, is_active, sort,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13,
			$14, $15,
			$16, $17,
			$18, $19, $20, $21,
			$22, $23, $24, $25,
			$26, $27, $28, $29
		)
	`

	_, err := r.db.Exec(ctx, sql,
		shift.ID, shift.TenantID, shift.Code, shift.Name, shift.Type, shift.Description,
		shift.WorkStart, shift.WorkEnd, shift.FlexibleStart, shift.FlexibleEnd, shift.WorkDuration,
		shift.CheckInRequired, shift.CheckOutRequired,
		shift.LateGracePeriod, shift.EarlyGracePeriod,
		restPeriodsJSON, shift.IsCrossDays,
		shift.AllowOvertime, shift.OvertimeStartBuffer, shift.OvertimeMinDuration, shift.OvertimePayRate,
		shift.WorkdayTypes, shift.Color, shift.IsActive, shift.Sort,
		shift.CreatedBy, shift.UpdatedBy, shift.CreatedAt, shift.UpdatedAt,
	)

	return err
}

func (r *shiftRepo) Update(ctx context.Context, shift *model.Shift) error {
	restPeriodsJSON, _ := json.Marshal(shift.RestPeriods)

	sql := `
		UPDATE hrm_shifts SET
			name = $1, description = $2,
			work_start = $3, work_end = $4, flexible_start = $5, flexible_end = $6, work_duration = $7,
			check_in_required = $8, check_out_required = $9,
			late_grace_period = $10, early_grace_period = $11,
			rest_periods = $12, is_cross_days = $13,
			allow_overtime = $14, overtime_start_buffer = $15, overtime_min_duration = $16, overtime_pay_rate = $17,
			workday_types = $18, color = $19, is_active = $20, sort = $21,
			updated_by = $22, updated_at = $23
		WHERE id = $24 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql,
		shift.Name, shift.Description,
		shift.WorkStart, shift.WorkEnd, shift.FlexibleStart, shift.FlexibleEnd, shift.WorkDuration,
		shift.CheckInRequired, shift.CheckOutRequired,
		shift.LateGracePeriod, shift.EarlyGracePeriod,
		restPeriodsJSON, shift.IsCrossDays,
		shift.AllowOvertime, shift.OvertimeStartBuffer, shift.OvertimeMinDuration, shift.OvertimePayRate,
		shift.WorkdayTypes, shift.Color, shift.IsActive, shift.Sort,
		shift.UpdatedBy, shift.UpdatedAt,
		shift.ID,
	)

	return err
}

func (r *shiftRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `UPDATE hrm_shifts SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *shiftRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Shift, error) {
	sql := `
		SELECT id, tenant_id, code, name, type, description,
		       work_start, work_end, flexible_start, flexible_end, work_duration,
		       check_in_required, check_out_required,
		       late_grace_period, early_grace_period,
		       rest_periods, is_cross_days,
		       allow_overtime, overtime_start_buffer, overtime_min_duration, overtime_pay_rate,
		       workday_types, color, is_active, sort,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM hrm_shifts
		WHERE id = $1 AND deleted_at IS NULL
	`

	shift := &model.Shift{}
	var restPeriodsJSON []byte
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&shift.ID, &shift.TenantID, &shift.Code, &shift.Name, &shift.Type, &shift.Description,
		&shift.WorkStart, &shift.WorkEnd, &shift.FlexibleStart, &shift.FlexibleEnd, &shift.WorkDuration,
		&shift.CheckInRequired, &shift.CheckOutRequired,
		&shift.LateGracePeriod, &shift.EarlyGracePeriod,
		&restPeriodsJSON, &shift.IsCrossDays,
		&shift.AllowOvertime, &shift.OvertimeStartBuffer, &shift.OvertimeMinDuration, &shift.OvertimePayRate,
		&shift.WorkdayTypes, &shift.Color, &shift.IsActive, &shift.Sort,
		&shift.CreatedBy, &shift.UpdatedBy, &shift.CreatedAt, &shift.UpdatedAt, &shift.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("shift not found")
		}
		return nil, err
	}

	if len(restPeriodsJSON) > 0 {
		json.Unmarshal(restPeriodsJSON, &shift.RestPeriods)
	}

	return shift, nil
}

func (r *shiftRepo) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Shift, error) {
	sql := `
		SELECT id, tenant_id, code, name, type, description,
		       work_start, work_end, flexible_start, flexible_end, work_duration,
		       check_in_required, check_out_required,
		       late_grace_period, early_grace_period,
		       rest_periods, is_cross_days,
		       allow_overtime, overtime_start_buffer, overtime_min_duration, overtime_pay_rate,
		       workday_types, color, is_active, sort,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM hrm_shifts
		WHERE tenant_id = $1 AND code = $2 AND deleted_at IS NULL
	`

	shift := &model.Shift{}
	var restPeriodsJSON []byte
	err := r.db.QueryRow(ctx, sql, tenantID, code).Scan(
		&shift.ID, &shift.TenantID, &shift.Code, &shift.Name, &shift.Type, &shift.Description,
		&shift.WorkStart, &shift.WorkEnd, &shift.FlexibleStart, &shift.FlexibleEnd, &shift.WorkDuration,
		&shift.CheckInRequired, &shift.CheckOutRequired,
		&shift.LateGracePeriod, &shift.EarlyGracePeriod,
		&restPeriodsJSON, &shift.IsCrossDays,
		&shift.AllowOvertime, &shift.OvertimeStartBuffer, &shift.OvertimeMinDuration, &shift.OvertimePayRate,
		&shift.WorkdayTypes, &shift.Color, &shift.IsActive, &shift.Sort,
		&shift.CreatedBy, &shift.UpdatedBy, &shift.CreatedAt, &shift.UpdatedAt, &shift.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("shift not found")
		}
		return nil, err
	}

	if len(restPeriodsJSON) > 0 {
		json.Unmarshal(restPeriodsJSON, &shift.RestPeriods)
	}

	return shift, nil
}

func (r *shiftRepo) List(ctx context.Context, tenantID uuid.UUID, filter *repository.ShiftFilter, offset, limit int) ([]*model.Shift, int, error) {
	// 构建查询条件
	where := "tenant_id = $1 AND deleted_at IS NULL"
	args := []interface{}{tenantID}
	argIdx := 2

	if filter != nil {
		if filter.Type != nil {
			where += fmt.Sprintf(" AND type = $%d", argIdx)
			args = append(args, *filter.Type)
			argIdx++
		}
		if filter.IsActive != nil {
			where += fmt.Sprintf(" AND is_active = $%d", argIdx)
			args = append(args, *filter.IsActive)
			argIdx++
		}
		if filter.Keyword != "" {
			where += fmt.Sprintf(" AND (name LIKE $%d OR code LIKE $%d)", argIdx, argIdx)
			args = append(args, "%"+filter.Keyword+"%")
			argIdx++
		}
	}

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM hrm_shifts WHERE %s", where)
	var total int
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	dataSQL := fmt.Sprintf(`
		SELECT id, tenant_id, code, name, type, description,
		       work_start, work_end, is_active, sort, created_at, updated_at
		FROM hrm_shifts
		WHERE %s
		ORDER BY sort ASC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var shifts []*model.Shift
	for rows.Next() {
		shift := &model.Shift{}
		err := rows.Scan(
			&shift.ID, &shift.TenantID, &shift.Code, &shift.Name, &shift.Type, &shift.Description,
			&shift.WorkStart, &shift.WorkEnd, &shift.IsActive, &shift.Sort,
			&shift.CreatedAt, &shift.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		shifts = append(shifts, shift)
	}

	return shifts, total, rows.Err()
}

func (r *shiftRepo) ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.Shift, error) {
	sql := `
		SELECT id, tenant_id, code, name, type, description,
		       work_start, work_end, is_active, sort, created_at
		FROM hrm_shifts
		WHERE tenant_id = $1 AND is_active = TRUE AND deleted_at IS NULL
		ORDER BY sort ASC, created_at DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []*model.Shift
	for rows.Next() {
		shift := &model.Shift{}
		err := rows.Scan(
			&shift.ID, &shift.TenantID, &shift.Code, &shift.Name, &shift.Type, &shift.Description,
			&shift.WorkStart, &shift.WorkEnd, &shift.IsActive, &shift.Sort, &shift.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		shifts = append(shifts, shift)
	}

	return shifts, rows.Err()
}

func (r *shiftRepo) ListByType(ctx context.Context, tenantID uuid.UUID, shiftType model.ShiftType) ([]*model.Shift, error) {
	sql := `
		SELECT id, tenant_id, code, name, type, description,
		       work_start, work_end, is_active, sort, created_at
		FROM hrm_shifts
		WHERE tenant_id = $1 AND type = $2 AND deleted_at IS NULL
		ORDER BY sort ASC, created_at DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, shiftType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []*model.Shift
	for rows.Next() {
		shift := &model.Shift{}
		err := rows.Scan(
			&shift.ID, &shift.TenantID, &shift.Code, &shift.Name, &shift.Type, &shift.Description,
			&shift.WorkStart, &shift.WorkEnd, &shift.IsActive, &shift.Sort, &shift.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		shifts = append(shifts, shift)
	}

	return shifts, rows.Err()
}

// ListWithCursor 游标分页查询班次（高性能）
func (r *shiftRepo) ListWithCursor(
	ctx context.Context,
	tenantID uuid.UUID,
	filter *repository.ShiftFilter,
	cursor *time.Time,
	limit int,
) ([]*model.Shift, *time.Time, bool, error) {
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
		if filter.Type != nil {
			argIdx++
			where += fmt.Sprintf(" AND type = $%d", argIdx)
			args = append(args, *filter.Type)
		}
		if filter.IsActive != nil {
			argIdx++
			where += fmt.Sprintf(" AND is_active = $%d", argIdx)
			args = append(args, *filter.IsActive)
		}
		if filter.Keyword != "" {
			argIdx++
			where += fmt.Sprintf(" AND (name LIKE $%d OR code LIKE $%d)", argIdx, argIdx)
			args = append(args, "%"+filter.Keyword+"%")
		}
	}

	// 构建查询（多查1条用于判断是否有下一页）
	argIdx++
	sql := fmt.Sprintf(`
		SELECT id, tenant_id, code, name, type, description,
		       work_start, work_end, is_active, sort, created_at
		FROM hrm_shifts
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
	var shifts []*model.Shift
	for rows.Next() {
		shift := &model.Shift{}
		err := rows.Scan(
			&shift.ID, &shift.TenantID, &shift.Code, &shift.Name, &shift.Type, &shift.Description,
			&shift.WorkStart, &shift.WorkEnd, &shift.IsActive, &shift.Sort, &shift.CreatedAt,
		)
		if err != nil {
			return nil, nil, false, err
		}
		shifts = append(shifts, shift)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, false, err
	}

	// 判断是否有下一页
	hasNext := len(shifts) > limit
	if hasNext {
		shifts = shifts[:limit]
	}

	// 生成下一页游标
	var nextCursor *time.Time
	if hasNext && len(shifts) > 0 {
		lastShift := shifts[len(shifts)-1]
		nextCursor = &lastShift.CreatedAt
	}

	return shifts, nextCursor, hasNext, nil
}
