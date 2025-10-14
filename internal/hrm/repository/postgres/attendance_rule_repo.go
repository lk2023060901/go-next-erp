package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

type attendanceRuleRepo struct {
	db *database.DB
}

// NewAttendanceRuleRepository 创建考勤规则仓储
func NewAttendanceRuleRepository(db *database.DB) repository.AttendanceRuleRepository {
	return &attendanceRuleRepo{db: db}
}

func (r *attendanceRuleRepo) Create(ctx context.Context, rule *model.AttendanceRule) error {
	allowedLocationsJSON, _ := json.Marshal(rule.AllowedLocations)

	sql := `
		INSERT INTO hrm_attendance_rules (
			id, tenant_id, code, name, description, apply_type,
			department_ids, employee_ids,
			workday_type, weekend_days, default_shift_id,
			location_required, allowed_locations,
			wifi_required, allowed_wifi,
			face_required, face_threshold, face_anti_spoofing,
			allow_field_work, holiday_calendar_id,
			require_approval_for_late, require_approval_for_early,
			is_active, priority,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8,
			$9, $10, $11,
			$12, $13,
			$14, $15,
			$16, $17, $18,
			$19, $20,
			$21, $22,
			$23, $24,
			$25, $26, $27, $28
		)
	`

	_, err := r.db.Exec(ctx, sql,
		rule.ID, rule.TenantID, rule.Code, rule.Name, rule.Description, rule.ApplyType,
		rule.DepartmentIDs, rule.EmployeeIDs,
		rule.WorkdayType, rule.WeekendDays, rule.DefaultShiftID,
		rule.LocationRequired, allowedLocationsJSON,
		rule.WiFiRequired, rule.AllowedWiFi,
		rule.FaceRequired, rule.FaceThreshold, rule.FaceAntiSpoofing,
		rule.AllowFieldWork, rule.HolidayCalendarID,
		rule.RequireApprovalForLate, rule.RequireApprovalForEarly,
		rule.IsActive, rule.Priority,
		rule.CreatedBy, rule.UpdatedBy, rule.CreatedAt, rule.UpdatedAt,
	)

	return err
}

func (r *attendanceRuleRepo) Update(ctx context.Context, rule *model.AttendanceRule) error {
	allowedLocationsJSON, _ := json.Marshal(rule.AllowedLocations)

	sql := `
		UPDATE hrm_attendance_rules SET
			name = $1, description = $2, apply_type = $3,
			department_ids = $4, employee_ids = $5,
			workday_type = $6, weekend_days = $7, default_shift_id = $8,
			location_required = $9, allowed_locations = $10,
			wifi_required = $11, allowed_wifi = $12,
			face_required = $13, face_threshold = $14, face_anti_spoofing = $15,
			allow_field_work = $16, holiday_calendar_id = $17,
			require_approval_for_late = $18, require_approval_for_early = $19,
			is_active = $20, priority = $21,
			updated_by = $22, updated_at = $23
		WHERE id = $24 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql,
		rule.Name, rule.Description, rule.ApplyType,
		rule.DepartmentIDs, rule.EmployeeIDs,
		rule.WorkdayType, rule.WeekendDays, rule.DefaultShiftID,
		rule.LocationRequired, allowedLocationsJSON,
		rule.WiFiRequired, rule.AllowedWiFi,
		rule.FaceRequired, rule.FaceThreshold, rule.FaceAntiSpoofing,
		rule.AllowFieldWork, rule.HolidayCalendarID,
		rule.RequireApprovalForLate, rule.RequireApprovalForEarly,
		rule.IsActive, rule.Priority,
		rule.UpdatedBy, rule.UpdatedAt,
		rule.ID,
	)

	return err
}

func (r *attendanceRuleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `UPDATE hrm_attendance_rules SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *attendanceRuleRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.AttendanceRule, error) {
	sql := `
		SELECT id, tenant_id, code, name, description, apply_type,
		       department_ids, employee_ids,
		       workday_type, weekend_days, default_shift_id,
		       location_required, allowed_locations,
		       wifi_required, allowed_wifi,
		       face_required, face_threshold, face_anti_spoofing,
		       allow_field_work, holiday_calendar_id,
		       require_approval_for_late, require_approval_for_early,
		       is_active, priority,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM hrm_attendance_rules
		WHERE id = $1 AND deleted_at IS NULL
	`

	rule := &model.AttendanceRule{}
	var allowedLocationsJSON []byte

	err := r.db.QueryRow(ctx, sql, id).Scan(
		&rule.ID, &rule.TenantID, &rule.Code, &rule.Name, &rule.Description, &rule.ApplyType,
		&rule.DepartmentIDs, &rule.EmployeeIDs,
		&rule.WorkdayType, &rule.WeekendDays, &rule.DefaultShiftID,
		&rule.LocationRequired, &allowedLocationsJSON,
		&rule.WiFiRequired, &rule.AllowedWiFi,
		&rule.FaceRequired, &rule.FaceThreshold, &rule.FaceAntiSpoofing,
		&rule.AllowFieldWork, &rule.HolidayCalendarID,
		&rule.RequireApprovalForLate, &rule.RequireApprovalForEarly,
		&rule.IsActive, &rule.Priority,
		&rule.CreatedBy, &rule.UpdatedBy, &rule.CreatedAt, &rule.UpdatedAt, &rule.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("attendance rule not found")
		}
		return nil, err
	}

	// 反序列化 JSONB 字段
	if len(allowedLocationsJSON) > 0 {
		json.Unmarshal(allowedLocationsJSON, &rule.AllowedLocations)
	}

	return rule, nil
}

func (r *attendanceRuleRepo) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.AttendanceRule, error) {
	sql := `
		SELECT id, tenant_id, code, name, description, apply_type, is_active, priority
		FROM hrm_attendance_rules
		WHERE tenant_id = $1 AND code = $2 AND deleted_at IS NULL
	`

	rule := &model.AttendanceRule{}
	err := r.db.QueryRow(ctx, sql, tenantID, code).Scan(
		&rule.ID, &rule.TenantID, &rule.Code, &rule.Name, &rule.Description, &rule.ApplyType, &rule.IsActive, &rule.Priority,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("attendance rule not found")
		}
		return nil, err
	}

	return rule, nil
}

func (r *attendanceRuleRepo) List(ctx context.Context, tenantID uuid.UUID, filter *repository.AttendanceRuleFilter, offset, limit int) ([]*model.AttendanceRule, int, error) {
	where := "tenant_id = $1 AND deleted_at IS NULL"
	args := []interface{}{tenantID}
	argIdx := 2

	if filter != nil {
		if filter.ApplyType != nil {
			where += fmt.Sprintf(" AND apply_type = $%d", argIdx)
			args = append(args, *filter.ApplyType)
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
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM hrm_attendance_rules WHERE %s", where)
	var total int
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	dataSQL := fmt.Sprintf(`
		SELECT id, tenant_id, code, name, description, apply_type, is_active, priority, created_at
		FROM hrm_attendance_rules
		WHERE %s
		ORDER BY priority DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rules []*model.AttendanceRule
	for rows.Next() {
		rule := &model.AttendanceRule{}
		err := rows.Scan(
			&rule.ID, &rule.TenantID, &rule.Code, &rule.Name, &rule.Description,
			&rule.ApplyType, &rule.IsActive, &rule.Priority, &rule.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		rules = append(rules, rule)
	}

	return rules, total, rows.Err()
}

func (r *attendanceRuleRepo) ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.AttendanceRule, error) {
	sql := `
		SELECT id, tenant_id, code, name, apply_type, is_active, priority
		FROM hrm_attendance_rules
		WHERE tenant_id = $1 AND is_active = TRUE AND deleted_at IS NULL
		ORDER BY priority DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*model.AttendanceRule
	for rows.Next() {
		rule := &model.AttendanceRule{}
		err := rows.Scan(
			&rule.ID, &rule.TenantID, &rule.Code, &rule.Name,
			&rule.ApplyType, &rule.IsActive, &rule.Priority,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

func (r *attendanceRuleRepo) FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) (*model.AttendanceRule, error) {
	// 按优先级查询适用于该员工的规则
	// 1. 优先查询直接指定员工的规则
	// 2. 其次查询员工职位的规则
	// 3. 最后查询员工组织的规则
	// 4. 全员规则

	sql := `
		SELECT ar.id, ar.tenant_id, ar.code, ar.name, ar.apply_type, ar.priority
		FROM hrm_attendance_rules ar
		WHERE ar.tenant_id = $1 AND ar.is_active = TRUE AND ar.deleted_at IS NULL
		  AND (
		    ar.apply_type = 'all'
		    OR (ar.apply_type = 'employee' AND ar.employee_ids::jsonb ? $2::text)
		  )
		ORDER BY 
		  CASE ar.apply_type
		    WHEN 'employee' THEN 1
		    WHEN 'position' THEN 2
		    WHEN 'organization' THEN 3
		    WHEN 'all' THEN 4
		  END,
		  ar.priority DESC
		LIMIT 1
	`

	rule := &model.AttendanceRule{}
	err := r.db.QueryRow(ctx, sql, tenantID, employeeID.String()).Scan(
		&rule.ID, &rule.TenantID, &rule.Code, &rule.Name, &rule.ApplyType, &rule.Priority,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no applicable attendance rule found for employee")
		}
		return nil, err
	}

	return rule, nil
}
