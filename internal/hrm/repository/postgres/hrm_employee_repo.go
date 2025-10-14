package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

type hrmEmployeeRepo struct {
	db *database.DB
}

// NewHRMEmployeeRepository 创建HRM员工仓储
func NewHRMEmployeeRepository(db *database.DB) repository.HRMEmployeeRepository {
	return &hrmEmployeeRepo{db: db}
}

func (r *hrmEmployeeRepo) Create(ctx context.Context, emp *model.HRMEmployee) error {
	sql := `
		INSERT INTO hrm_employees (
			id, tenant_id, employee_id, id_card_no, card_no, face_data, fingerprint,
			dingtalk_user_id, wecom_user_id, feishu_user_id, feishu_open_id,
			work_location, work_schedule_type, attendance_rule_id, default_shift_id,
			allow_field_work, require_face, require_location, require_wifi,
			emergency_contact, emergency_phone, emergency_relation,
			is_active, remark, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17, $18, $19,
			$20, $21, $22,
			$23, $24, $25, $26
		)
	`

	_, err := r.db.Exec(ctx, sql,
		emp.ID, emp.TenantID, emp.EmployeeID, emp.IDCardNo, emp.CardNo, emp.FaceData, emp.Fingerprint,
		emp.DingTalkUserID, emp.WeComUserID, emp.FeishuUserID, emp.FeishuOpenID,
		emp.WorkLocation, emp.WorkScheduleType, emp.AttendanceRuleID, emp.DefaultShiftID,
		emp.AllowFieldWork, emp.RequireFace, emp.RequireLocation, emp.RequireWiFi,
		emp.EmergencyContact, emp.EmergencyPhone, emp.EmergencyRelation,
		emp.IsActive, emp.Remark, emp.CreatedAt, emp.UpdatedAt,
	)

	return err
}

func (r *hrmEmployeeRepo) BatchCreate(ctx context.Context, employees []*model.HRMEmployee) error {
	if len(employees) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, emp := range employees {
		if err := r.Create(ctx, emp); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *hrmEmployeeRepo) Update(ctx context.Context, emp *model.HRMEmployee) error {
	sql := `
		UPDATE hrm_employees SET
			card_no = $1, face_data = $2, fingerprint = $3,
			dingtalk_user_id = $4, wecom_user_id = $5, feishu_user_id = $6,
			work_location = $7, work_schedule_type = $8, attendance_rule_id = $9, default_shift_id = $10,
			allow_field_work = $11, require_face = $12, require_location = $13, require_wifi = $14,
			emergency_contact = $15, emergency_phone = $16, emergency_relation = $17,
			is_active = $18, remark = $19, updated_at = $20
		WHERE id = $21 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql,
		emp.CardNo, emp.FaceData, emp.Fingerprint,
		emp.DingTalkUserID, emp.WeComUserID, emp.FeishuUserID,
		emp.WorkLocation, emp.WorkScheduleType, emp.AttendanceRuleID, emp.DefaultShiftID,
		emp.AllowFieldWork, emp.RequireFace, emp.RequireLocation, emp.RequireWiFi,
		emp.EmergencyContact, emp.EmergencyPhone, emp.EmergencyRelation,
		emp.IsActive, emp.Remark, emp.UpdatedAt,
		emp.ID,
	)

	return err
}

func (r *hrmEmployeeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `UPDATE hrm_employees SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *hrmEmployeeRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.HRMEmployee, error) {
	sql := `
		SELECT id, tenant_id, employee_id, id_card_no, card_no, face_data, fingerprint,
		       dingtalk_user_id, wecom_user_id, feishu_user_id, feishu_open_id,
		       work_location, work_schedule_type, attendance_rule_id, default_shift_id,
		       allow_field_work, require_face, require_location, require_wifi,
		       emergency_contact, emergency_phone, emergency_relation,
		       is_active, remark, created_at, updated_at, deleted_at
		FROM hrm_employees
		WHERE id = $1 AND deleted_at IS NULL
	`

	emp := &model.HRMEmployee{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&emp.ID, &emp.TenantID, &emp.EmployeeID, &emp.IDCardNo, &emp.CardNo, &emp.FaceData, &emp.Fingerprint,
		&emp.DingTalkUserID, &emp.WeComUserID, &emp.FeishuUserID, &emp.FeishuOpenID,
		&emp.WorkLocation, &emp.WorkScheduleType, &emp.AttendanceRuleID, &emp.DefaultShiftID,
		&emp.AllowFieldWork, &emp.RequireFace, &emp.RequireLocation, &emp.RequireWiFi,
		&emp.EmergencyContact, &emp.EmergencyPhone, &emp.EmergencyRelation,
		&emp.IsActive, &emp.Remark, &emp.CreatedAt, &emp.UpdatedAt, &emp.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("hrm employee not found")
		}
		return nil, err
	}

	return emp, nil
}

func (r *hrmEmployeeRepo) FindByEmployeeID(ctx context.Context, tenantID, employeeID uuid.UUID) (*model.HRMEmployee, error) {
	sql := `
		SELECT id, tenant_id, employee_id, 
		       COALESCE(card_no, '') as card_no, 
		       COALESCE(face_data, '') as face_data, 
		       COALESCE(fingerprint, '') as fingerprint,
		       COALESCE(dingtalk_user_id, '') as dingtalk_user_id, 
		       COALESCE(wecom_user_id, '') as wecom_user_id, 
		       COALESCE(feishu_user_id, '') as feishu_user_id,
		       attendance_rule_id, default_shift_id, is_active, created_at
		FROM hrm_employees
		WHERE tenant_id = $1 AND employee_id = $2 AND deleted_at IS NULL
	`

	emp := &model.HRMEmployee{}
	err := r.db.QueryRow(ctx, sql, tenantID, employeeID).Scan(
		&emp.ID, &emp.TenantID, &emp.EmployeeID, &emp.CardNo, &emp.FaceData, &emp.Fingerprint,
		&emp.DingTalkUserID, &emp.WeComUserID, &emp.FeishuUserID,
		&emp.AttendanceRuleID, &emp.DefaultShiftID, &emp.IsActive, &emp.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("hrm employee not found")
		}
		return nil, err
	}

	return emp, nil
}

func (r *hrmEmployeeRepo) FindByCardNo(ctx context.Context, tenantID uuid.UUID, cardNo string) (*model.HRMEmployee, error) {
	sql := `SELECT id, tenant_id, employee_id, card_no FROM hrm_employees WHERE tenant_id = $1 AND card_no = $2 AND deleted_at IS NULL`
	emp := &model.HRMEmployee{}
	err := r.db.QueryRow(ctx, sql, tenantID, cardNo).Scan(&emp.ID, &emp.TenantID, &emp.EmployeeID, &emp.CardNo)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("hrm employee not found")
		}
		return nil, err
	}
	return emp, nil
}

func (r *hrmEmployeeRepo) FindByThirdPartyID(ctx context.Context, tenantID uuid.UUID, platform model.PlatformType, platformID string) (*model.HRMEmployee, error) {
	var sql string
	switch platform {
	case model.PlatformDingTalk:
		sql = `SELECT id, tenant_id, employee_id, dingtalk_user_id FROM hrm_employees WHERE tenant_id = $1 AND dingtalk_user_id = $2 AND deleted_at IS NULL`
	case model.PlatformWeCom:
		sql = `SELECT id, tenant_id, employee_id, wecom_user_id FROM hrm_employees WHERE tenant_id = $1 AND wecom_user_id = $2 AND deleted_at IS NULL`
	case model.PlatformFeishu:
		sql = `SELECT id, tenant_id, employee_id, feishu_user_id FROM hrm_employees WHERE tenant_id = $1 AND feishu_user_id = $2 AND deleted_at IS NULL`
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}

	emp := &model.HRMEmployee{}
	var thirdPartyID string
	err := r.db.QueryRow(ctx, sql, tenantID, platformID).Scan(&emp.ID, &emp.TenantID, &emp.EmployeeID, &thirdPartyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("hrm employee not found")
		}
		return nil, err
	}

	// 设置对应的第三方ID
	switch platform {
	case model.PlatformDingTalk:
		emp.DingTalkUserID = thirdPartyID
	case model.PlatformWeCom:
		emp.WeComUserID = thirdPartyID
	case model.PlatformFeishu:
		emp.FeishuUserID = thirdPartyID
	}

	return emp, nil
}

func (r *hrmEmployeeRepo) List(ctx context.Context, tenantID uuid.UUID, filter *repository.HRMEmployeeFilter, offset, limit int) ([]*model.HRMEmployee, int, error) {
	where := "tenant_id = $1 AND deleted_at IS NULL"
	args := []interface{}{tenantID}
	argIdx := 2

	if filter != nil {
		if filter.AttendanceRuleID != nil {
			where += fmt.Sprintf(" AND attendance_rule_id = $%d", argIdx)
			args = append(args, *filter.AttendanceRuleID)
			argIdx++
		}
		if filter.IsActive != nil {
			where += fmt.Sprintf(" AND is_active = $%d", argIdx)
			args = append(args, *filter.IsActive)
			argIdx++
		}
	}

	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM hrm_employees WHERE %s", where)
	var total int
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataSQL := fmt.Sprintf(`
		SELECT id, tenant_id, employee_id, card_no, is_active, created_at
		FROM hrm_employees WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var emps []*model.HRMEmployee
	for rows.Next() {
		emp := &model.HRMEmployee{}
		rows.Scan(&emp.ID, &emp.TenantID, &emp.EmployeeID, &emp.CardNo, &emp.IsActive, &emp.CreatedAt)
		emps = append(emps, emp)
	}

	return emps, total, rows.Err()
}

func (r *hrmEmployeeRepo) ListByAttendanceRule(ctx context.Context, tenantID, ruleID uuid.UUID) ([]*model.HRMEmployee, error) {
	sql := `SELECT id, employee_id FROM hrm_employees WHERE tenant_id = $1 AND attendance_rule_id = $2 AND deleted_at IS NULL`
	rows, err := r.db.Query(ctx, sql, tenantID, ruleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emps []*model.HRMEmployee
	for rows.Next() {
		emp := &model.HRMEmployee{TenantID: tenantID}
		rows.Scan(&emp.ID, &emp.EmployeeID)
		emps = append(emps, emp)
	}
	return emps, rows.Err()
}

func (r *hrmEmployeeRepo) ListByShift(ctx context.Context, tenantID, shiftID uuid.UUID) ([]*model.HRMEmployee, error) {
	sql := `SELECT id, employee_id FROM hrm_employees WHERE tenant_id = $1 AND default_shift_id = $2 AND deleted_at IS NULL`
	rows, err := r.db.Query(ctx, sql, tenantID, shiftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emps []*model.HRMEmployee
	for rows.Next() {
		emp := &model.HRMEmployee{TenantID: tenantID}
		rows.Scan(&emp.ID, &emp.EmployeeID)
		emps = append(emps, emp)
	}
	return emps, rows.Err()
}

func (r *hrmEmployeeRepo) ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.HRMEmployee, error) {
	sql := `SELECT id, employee_id FROM hrm_employees WHERE tenant_id = $1 AND is_active = TRUE AND deleted_at IS NULL`
	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emps []*model.HRMEmployee
	for rows.Next() {
		emp := &model.HRMEmployee{TenantID: tenantID, IsActive: true}
		rows.Scan(&emp.ID, &emp.EmployeeID)
		emps = append(emps, emp)
	}
	return emps, rows.Err()
}

func (r *hrmEmployeeRepo) UpdateFaceData(ctx context.Context, id uuid.UUID, faceData string) error {
	sql := `UPDATE hrm_employees SET face_data = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, sql, faceData, id)
	return err
}

func (r *hrmEmployeeRepo) UpdateFingerprint(ctx context.Context, id uuid.UUID, fingerprint string) error {
	sql := `UPDATE hrm_employees SET fingerprint = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, sql, fingerprint, id)
	return err
}

func (r *hrmEmployeeRepo) UpdateCardNo(ctx context.Context, id uuid.UUID, cardNo string) error {
	sql := `UPDATE hrm_employees SET card_no = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, sql, cardNo, id)
	return err
}

func (r *hrmEmployeeRepo) UpdateThirdPartyID(ctx context.Context, id uuid.UUID, platform model.PlatformType, platformID string) error {
	var sql string
	switch platform {
	case model.PlatformDingTalk:
		sql = `UPDATE hrm_employees SET dingtalk_user_id = $1, updated_at = NOW() WHERE id = $2`
	case model.PlatformWeCom:
		sql = `UPDATE hrm_employees SET wecom_user_id = $1, updated_at = NOW() WHERE id = $2`
	case model.PlatformFeishu:
		sql = `UPDATE hrm_employees SET feishu_user_id = $1, updated_at = NOW() WHERE id = $2`
	default:
		return fmt.Errorf("unsupported platform: %s", platform)
	}
	_, err := r.db.Exec(ctx, sql, platformID, id)
	return err
}

func (r *hrmEmployeeRepo) ExistsByEmployeeID(ctx context.Context, tenantID, employeeID uuid.UUID) (bool, error) {
	sql := `SELECT COUNT(*) FROM hrm_employees WHERE tenant_id = $1 AND employee_id = $2 AND deleted_at IS NULL`
	var count int
	err := r.db.QueryRow(ctx, sql, tenantID, employeeID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
