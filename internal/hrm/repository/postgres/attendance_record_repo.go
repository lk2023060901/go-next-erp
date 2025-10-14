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

type attendanceRecordRepo struct {
	db *database.DB
}

// NewAttendanceRecordRepository 创建考勤记录仓储
func NewAttendanceRecordRepository(db *database.DB) repository.AttendanceRecordRepository {
	return &attendanceRecordRepo{db: db}
}

func (r *attendanceRecordRepo) Create(ctx context.Context, record *model.AttendanceRecord) error {
	sql := `
		INSERT INTO hrm_attendance_records (
			id, tenant_id, employee_id, employee_name, department_id,
			shift_id, shift_name,
			clock_time, clock_type, status, check_in_method, source_type, source_id,
			location, address, wifi_ssid, wifi_mac,
			photo_url, face_score, temperature,
			is_exception, exception_reason, exception_type,
			approval_id, raw_data, remark,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7,
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17,
			$18, $19, $20,
			$21, $22, $23,
			$24, $25, $26,
			$27, $28
		)
	`

	locationJSON, _ := json.Marshal(record.Location)
	rawDataJSON, _ := json.Marshal(record.RawData)

	_, err := r.db.Exec(ctx, sql,
		record.ID, record.TenantID, record.EmployeeID, record.EmployeeName, record.DepartmentID,
		record.ShiftID, record.ShiftName,
		record.ClockTime, record.ClockType, record.Status, record.CheckInMethod, record.SourceType, record.SourceID,
		locationJSON, record.Address, record.WiFiSSID, record.WiFiMAC,
		record.PhotoURL, record.FaceScore, record.Temperature,
		record.IsException, record.ExceptionReason, record.ExceptionType,
		record.ApprovalID, rawDataJSON, record.Remark,
		record.CreatedAt, record.UpdatedAt,
	)

	return err
}

func (r *attendanceRecordRepo) BatchCreate(ctx context.Context, records []*model.AttendanceRecord) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, record := range records {
		if err := r.Create(ctx, record); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *attendanceRecordRepo) Update(ctx context.Context, record *model.AttendanceRecord) error {
	sql := `
		UPDATE hrm_attendance_records SET
			status = $1, is_exception = $2, exception_reason = $3,
			exception_type = $4, approval_id = $5, remark = $6, updated_at = $7
		WHERE id = $8 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql,
		record.Status, record.IsException, record.ExceptionReason,
		record.ExceptionType, record.ApprovalID, record.Remark, record.UpdatedAt,
		record.ID,
	)

	return err
}

func (r *attendanceRecordRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `UPDATE hrm_attendance_records SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *attendanceRecordRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.AttendanceRecord, error) {
	sql := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
		       shift_id, shift_name,
		       clock_time, clock_type, status, check_in_method, source_type, source_id,
		       location, address, wifi_ssid, wifi_mac,
		       photo_url, face_score, temperature,
		       is_exception, exception_reason, exception_type,
		       approval_id, raw_data, remark,
		       created_at, updated_at, deleted_at
		FROM hrm_attendance_records
		WHERE id = $1 AND deleted_at IS NULL
	`

	record := &model.AttendanceRecord{}
	var locationJSON, rawDataJSON []byte
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&record.ID, &record.TenantID, &record.EmployeeID, &record.EmployeeName, &record.DepartmentID,
		&record.ShiftID, &record.ShiftName,
		&record.ClockTime, &record.ClockType, &record.Status, &record.CheckInMethod, &record.SourceType, &record.SourceID,
		&locationJSON, &record.Address, &record.WiFiSSID, &record.WiFiMAC,
		&record.PhotoURL, &record.FaceScore, &record.Temperature,
		&record.IsException, &record.ExceptionReason, &record.ExceptionType,
		&record.ApprovalID, &rawDataJSON, &record.Remark,
		&record.CreatedAt, &record.UpdatedAt, &record.DeletedAt,
	)

	if err == nil {
		if len(locationJSON) > 0 {
			json.Unmarshal(locationJSON, &record.Location)
		}
		if len(rawDataJSON) > 0 {
			json.Unmarshal(rawDataJSON, &record.RawData)
		}
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("attendance record not found")
		}
		return nil, err
	}

	return record, nil
}

func (r *attendanceRecordRepo) FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) ([]*model.AttendanceRecord, error) {
	// 添加调试日志
	fmt.Printf("[DEBUG] FindByEmployee - tenantID: %s, employeeID: %s, startDate: %v, endDate: %v\n",
		tenantID, employeeID, startDate, endDate)

	sql := `
		SELECT id, tenant_id, employee_id, employee_name,
		       clock_time, clock_type, status, check_in_method, source_type,
		       is_exception, exception_reason, exception_type, created_at
		FROM hrm_attendance_records
		WHERE tenant_id = $1 AND employee_id = $2 AND clock_time >= $3 AND clock_time < $4 AND deleted_at IS NULL
		ORDER BY clock_time ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, employeeID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*model.AttendanceRecord
	for rows.Next() {
		record := &model.AttendanceRecord{}
		err := rows.Scan(
			&record.ID, &record.TenantID, &record.EmployeeID, &record.EmployeeName,
			&record.ClockTime, &record.ClockType, &record.Status, &record.CheckInMethod, &record.SourceType,
			&record.IsException, &record.ExceptionReason, &record.ExceptionType, &record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *attendanceRecordRepo) FindByDepartment(ctx context.Context, tenantID, departmentID uuid.UUID, startDate, endDate time.Time) ([]*model.AttendanceRecord, error) {
	sql := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
		       clock_time, clock_type, status, source_type,
		       is_exception, exception_reason, created_at
		FROM hrm_attendance_records
		WHERE tenant_id = $1 AND department_id = $2 AND clock_time >= $3 AND clock_time < $4 AND deleted_at IS NULL
		ORDER BY clock_time DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, departmentID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*model.AttendanceRecord
	for rows.Next() {
		record := &model.AttendanceRecord{}
		err := rows.Scan(
			&record.ID, &record.TenantID, &record.EmployeeID, &record.EmployeeName, &record.DepartmentID,
			&record.ClockTime, &record.ClockType, &record.Status, &record.SourceType,
			&record.IsException, &record.ExceptionReason, &record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *attendanceRecordRepo) FindByDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*model.AttendanceRecord, error) {
	sql := `
		SELECT id, tenant_id, employee_id, employee_name,
		       clock_time, clock_type, status, source_type, created_at
		FROM hrm_attendance_records
		WHERE tenant_id = $1 AND clock_time >= $2 AND clock_time < $3 AND deleted_at IS NULL
		ORDER BY clock_time DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*model.AttendanceRecord
	for rows.Next() {
		record := &model.AttendanceRecord{}
		err := rows.Scan(
			&record.ID, &record.TenantID, &record.EmployeeID, &record.EmployeeName,
			&record.ClockTime, &record.ClockType, &record.Status, &record.SourceType, &record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *attendanceRecordRepo) List(ctx context.Context, tenantID uuid.UUID, filter *repository.AttendanceRecordFilter, offset, limit int) ([]*model.AttendanceRecord, int, error) {
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
		if filter.StartDate != nil {
			where += fmt.Sprintf(" AND clock_time >= $%d", argIdx)
			args = append(args, *filter.StartDate)
			argIdx++
		}
		if filter.EndDate != nil {
			where += fmt.Sprintf(" AND clock_time < $%d", argIdx)
			args = append(args, *filter.EndDate)
			argIdx++
		}
		if filter.ClockType != nil {
			where += fmt.Sprintf(" AND clock_type = $%d", argIdx)
			args = append(args, *filter.ClockType)
			argIdx++
		}
		if filter.Status != nil {
			where += fmt.Sprintf(" AND status = $%d", argIdx)
			args = append(args, *filter.Status)
			argIdx++
		}
		if filter.SourceType != nil {
			where += fmt.Sprintf(" AND source_type = $%d", argIdx)
			args = append(args, *filter.SourceType)
			argIdx++
		}
		if filter.IsException != nil {
			where += fmt.Sprintf(" AND is_exception = $%d", argIdx)
			args = append(args, *filter.IsException)
			argIdx++
		}
		if filter.Keyword != "" {
			where += fmt.Sprintf(" AND employee_name LIKE $%d", argIdx)
			args = append(args, "%"+filter.Keyword+"%")
			argIdx++
		}
	}

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM hrm_attendance_records WHERE %s", where)
	var total int
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	dataSQL := fmt.Sprintf(`
		SELECT id, tenant_id, employee_id, employee_name, clock_time, clock_type, status, 
		       source_type, is_exception, exception_reason, created_at
		FROM hrm_attendance_records 
		WHERE %s 
		ORDER BY clock_time DESC 
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []*model.AttendanceRecord
	for rows.Next() {
		record := &model.AttendanceRecord{}
		err := rows.Scan(
			&record.ID, &record.TenantID, &record.EmployeeID, &record.EmployeeName,
			&record.ClockTime, &record.ClockType, &record.Status,
			&record.SourceType, &record.IsException, &record.ExceptionReason, &record.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}

	return records, total, rows.Err()
}

func (r *attendanceRecordRepo) CountByStatus(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[model.AttendanceStatus]int, error) {
	sql := `
		SELECT status, COUNT(*) as count
		FROM hrm_attendance_records
		WHERE tenant_id = $1 AND clock_time >= $2 AND clock_time < $3 AND deleted_at IS NULL
		GROUP BY status
	`

	rows, err := r.db.Query(ctx, sql, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[model.AttendanceStatus]int)
	for rows.Next() {
		var status model.AttendanceStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		result[status] = count
	}

	return result, rows.Err()
}

func (r *attendanceRecordRepo) FindExceptions(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*model.AttendanceRecord, error) {
	sql := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
		       clock_time, clock_type, status, source_type,
		       is_exception, exception_reason, exception_type, created_at
		FROM hrm_attendance_records
		WHERE tenant_id = $1 AND is_exception = TRUE AND clock_time >= $2 AND clock_time < $3 AND deleted_at IS NULL
		ORDER BY clock_time DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*model.AttendanceRecord
	for rows.Next() {
		record := &model.AttendanceRecord{}
		err := rows.Scan(
			&record.ID, &record.TenantID, &record.EmployeeID, &record.EmployeeName, &record.DepartmentID,
			&record.ClockTime, &record.ClockType, &record.Status, &record.SourceType,
			&record.IsException, &record.ExceptionReason, &record.ExceptionType, &record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}
