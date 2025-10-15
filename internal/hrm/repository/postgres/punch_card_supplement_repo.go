package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// PunchCardSupplementRepo 补卡申请PostgreSQL仓储实现
type PunchCardSupplementRepo struct {
	db *database.DB
}

// NewPunchCardSupplementRepo 创建补卡申请仓储
func NewPunchCardSupplementRepo(db *database.DB) repository.PunchCardSupplementRepository {
	return &PunchCardSupplementRepo{db: db}
}

// Create 创建补卡申请
func (r *PunchCardSupplementRepo) Create(ctx context.Context, supplement *model.PunchCardSupplement) error {
	// 序列化证明材料
	evidenceJSON, err := json.Marshal(supplement.Evidence)
	if err != nil {
		return fmt.Errorf("failed to marshal evidence: %w", err)
	}

	query := `
		INSERT INTO hrm.punch_card_supplements (
			id, tenant_id, employee_id, employee_name, department_id,
			supplement_date, supplement_type, supplement_time,
			missing_type, reason, evidence,
			attendance_record_id,
			approval_id, approval_status, approved_by, approved_at, reject_reason,
			process_status, processed_at, processed_by,
			remark, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12,
			$13, $14, $15, $16, $17,
			$18, $19, $20,
			$21, $22, $23
		)
	`

	now := time.Now()
	supplement.ID = uuid.New()
	supplement.CreatedAt = now
	supplement.UpdatedAt = now

	_, err = r.db.Exec(ctx, query,
		supplement.ID, supplement.TenantID, supplement.EmployeeID, supplement.EmployeeName, supplement.DepartmentID,
		supplement.SupplementDate, supplement.SupplementType, supplement.SupplementTime,
		supplement.MissingType, supplement.Reason, evidenceJSON,
		supplement.AttendanceRecordID,
		supplement.ApprovalID, supplement.ApprovalStatus, supplement.ApprovedBy, supplement.ApprovedAt, supplement.RejectReason,
		supplement.ProcessStatus, supplement.ProcessedAt, supplement.ProcessedBy,
		supplement.Remark, supplement.CreatedAt, supplement.UpdatedAt,
	)

	return err
}

// FindByID 根据ID查找
func (r *PunchCardSupplementRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.PunchCardSupplement, error) {
	query := `
		SELECT 
			id, tenant_id, employee_id, employee_name, department_id,
			supplement_date, supplement_type, supplement_time,
			missing_type, reason, evidence,
			attendance_record_id,
			approval_id, approval_status, approved_by, approved_at, reject_reason,
			process_status, processed_at, processed_by,
			remark, created_at, updated_at, deleted_at
		FROM hrm.punch_card_supplements
		WHERE id = $1 AND deleted_at IS NULL
	`

	supplement := &model.PunchCardSupplement{}
	var evidenceJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&supplement.ID, &supplement.TenantID, &supplement.EmployeeID, &supplement.EmployeeName, &supplement.DepartmentID,
		&supplement.SupplementDate, &supplement.SupplementType, &supplement.SupplementTime,
		&supplement.MissingType, &supplement.Reason, &evidenceJSON,
		&supplement.AttendanceRecordID,
		&supplement.ApprovalID, &supplement.ApprovalStatus, &supplement.ApprovedBy, &supplement.ApprovedAt, &supplement.RejectReason,
		&supplement.ProcessStatus, &supplement.ProcessedAt, &supplement.ProcessedBy,
		&supplement.Remark, &supplement.CreatedAt, &supplement.UpdatedAt, &supplement.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("punch card supplement not found")
		}
		return nil, err
	}

	// 反序列化证明材料
	if len(evidenceJSON) > 0 {
		if err := json.Unmarshal(evidenceJSON, &supplement.Evidence); err != nil {
			return nil, fmt.Errorf("failed to unmarshal evidence: %w", err)
		}
	}

	return supplement, nil
}

// Update 更新补卡申请
func (r *PunchCardSupplementRepo) Update(ctx context.Context, supplement *model.PunchCardSupplement) error {
	// 序列化证明材料
	evidenceJSON, err := json.Marshal(supplement.Evidence)
	if err != nil {
		return fmt.Errorf("failed to marshal evidence: %w", err)
	}

	query := `
		UPDATE hrm.punch_card_supplements SET
			supplement_date = $2,
			supplement_type = $3,
			supplement_time = $4,
			missing_type = $5,
			reason = $6,
			evidence = $7,
			attendance_record_id = $8,
			approval_id = $9,
			approval_status = $10,
			approved_by = $11,
			approved_at = $12,
			reject_reason = $13,
			process_status = $14,
			processed_at = $15,
			processed_by = $16,
			remark = $17,
			updated_at = $18
		WHERE id = $1 AND deleted_at IS NULL
	`

	supplement.UpdatedAt = time.Now()

	_, err = r.db.Exec(ctx, query,
		supplement.ID,
		supplement.SupplementDate, supplement.SupplementType, supplement.SupplementTime,
		supplement.MissingType, supplement.Reason, evidenceJSON,
		supplement.AttendanceRecordID,
		supplement.ApprovalID, supplement.ApprovalStatus, supplement.ApprovedBy, supplement.ApprovedAt, supplement.RejectReason,
		supplement.ProcessStatus, supplement.ProcessedAt, supplement.ProcessedBy,
		supplement.Remark, supplement.UpdatedAt,
	)

	return err
}

// Delete 删除补卡申请（软删除）
func (r *PunchCardSupplementRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE hrm.punch_card_supplements 
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, id, now)
	return err
}

// List 列表查询（分页）
func (r *PunchCardSupplementRepo) List(ctx context.Context, tenantID uuid.UUID, filter *repository.PunchCardSupplementFilter, offset, limit int) ([]*model.PunchCardSupplement, int, error) {
	// 构建查询条件
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

		if filter.SupplementType != nil {
			conditions = append(conditions, fmt.Sprintf("supplement_type = $%d", argIndex))
			args = append(args, *filter.SupplementType)
			argIndex++
		}

		if filter.MissingType != nil {
			conditions = append(conditions, fmt.Sprintf("missing_type = $%d", argIndex))
			args = append(args, *filter.MissingType)
			argIndex++
		}

		if filter.ApprovalStatus != nil {
			conditions = append(conditions, fmt.Sprintf("approval_status = $%d", argIndex))
			args = append(args, *filter.ApprovalStatus)
			argIndex++
		}

		if filter.ProcessStatus != nil {
			conditions = append(conditions, fmt.Sprintf("process_status = $%d", argIndex))
			args = append(args, *filter.ProcessStatus)
			argIndex++
		}

		if filter.StartDate != nil {
			conditions = append(conditions, fmt.Sprintf("supplement_date >= $%d", argIndex))
			args = append(args, *filter.StartDate)
			argIndex++
		}

		if filter.EndDate != nil {
			conditions = append(conditions, fmt.Sprintf("supplement_date < $%d", argIndex))
			args = append(args, *filter.EndDate)
			argIndex++
		}

		if filter.Keyword != "" {
			conditions = append(conditions, fmt.Sprintf("(employee_name ILIKE $%d OR reason ILIKE $%d)", argIndex, argIndex))
			args = append(args, "%"+filter.Keyword+"%")
			argIndex++
		}
	}

	whereClause := "WHERE " + conditions[0]
	for i := 1; i < len(conditions); i++ {
		whereClause += " AND " + conditions[i]
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM hrm.punch_card_supplements " + whereClause
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := `
		SELECT 
			id, tenant_id, employee_id, employee_name, department_id,
			supplement_date, supplement_type, supplement_time,
			missing_type, reason, evidence,
			attendance_record_id,
			approval_id, approval_status, approved_by, approved_at, reject_reason,
			process_status, processed_at, processed_by,
			remark, created_at, updated_at
		FROM hrm.punch_card_supplements
		` + whereClause + `
		ORDER BY supplement_date DESC, created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	supplements := make([]*model.PunchCardSupplement, 0)
	for rows.Next() {
		supplement := &model.PunchCardSupplement{}
		var evidenceJSON []byte

		err := rows.Scan(
			&supplement.ID, &supplement.TenantID, &supplement.EmployeeID, &supplement.EmployeeName, &supplement.DepartmentID,
			&supplement.SupplementDate, &supplement.SupplementType, &supplement.SupplementTime,
			&supplement.MissingType, &supplement.Reason, &evidenceJSON,
			&supplement.AttendanceRecordID,
			&supplement.ApprovalID, &supplement.ApprovalStatus, &supplement.ApprovedBy, &supplement.ApprovedAt, &supplement.RejectReason,
			&supplement.ProcessStatus, &supplement.ProcessedAt, &supplement.ProcessedBy,
			&supplement.Remark, &supplement.CreatedAt, &supplement.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// 反序列化证明材料
		if len(evidenceJSON) > 0 {
			if err := json.Unmarshal(evidenceJSON, &supplement.Evidence); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal evidence: %w", err)
			}
		}

		supplements = append(supplements, supplement)
	}

	return supplements, total, rows.Err()
}

// FindByEmployee 查询员工补卡申请记录
func (r *PunchCardSupplementRepo) FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.PunchCardSupplement, error) {
	query := `
		SELECT 
			id, tenant_id, employee_id, employee_name, department_id,
			supplement_date, supplement_type, supplement_time,
			missing_type, reason, evidence,
			attendance_record_id,
			approval_id, approval_status, approved_by, approved_at, reject_reason,
			process_status, processed_at, processed_by,
			remark, created_at, updated_at
		FROM hrm.punch_card_supplements
		WHERE tenant_id = $1 
			AND employee_id = $2 
			AND EXTRACT(YEAR FROM supplement_date) = $3
			AND deleted_at IS NULL
		ORDER BY supplement_date DESC, created_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID, employeeID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	supplements := make([]*model.PunchCardSupplement, 0)
	for rows.Next() {
		supplement := &model.PunchCardSupplement{}
		var evidenceJSON []byte

		err := rows.Scan(
			&supplement.ID, &supplement.TenantID, &supplement.EmployeeID, &supplement.EmployeeName, &supplement.DepartmentID,
			&supplement.SupplementDate, &supplement.SupplementType, &supplement.SupplementTime,
			&supplement.MissingType, &supplement.Reason, &evidenceJSON,
			&supplement.AttendanceRecordID,
			&supplement.ApprovalID, &supplement.ApprovalStatus, &supplement.ApprovedBy, &supplement.ApprovedAt, &supplement.RejectReason,
			&supplement.ProcessStatus, &supplement.ProcessedAt, &supplement.ProcessedBy,
			&supplement.Remark, &supplement.CreatedAt, &supplement.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 反序列化证明材料
		if len(evidenceJSON) > 0 {
			if err := json.Unmarshal(evidenceJSON, &supplement.Evidence); err != nil {
				return nil, fmt.Errorf("failed to unmarshal evidence: %w", err)
			}
		}

		supplements = append(supplements, supplement)
	}

	return supplements, rows.Err()
}

// FindPending 查询待审批的补卡申请
func (r *PunchCardSupplementRepo) FindPending(ctx context.Context, tenantID uuid.UUID) ([]*model.PunchCardSupplement, error) {
	query := `
		SELECT 
			id, tenant_id, employee_id, employee_name, department_id,
			supplement_date, supplement_type, supplement_time,
			missing_type, reason, evidence,
			attendance_record_id,
			approval_id, approval_status, approved_by, approved_at, reject_reason,
			process_status, processed_at, processed_by,
			remark, created_at, updated_at
		FROM hrm.punch_card_supplements
		WHERE tenant_id = $1 
			AND approval_status = 'pending'
			AND deleted_at IS NULL
		ORDER BY supplement_date DESC, created_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	supplements := make([]*model.PunchCardSupplement, 0)
	for rows.Next() {
		supplement := &model.PunchCardSupplement{}
		var evidenceJSON []byte

		err := rows.Scan(
			&supplement.ID, &supplement.TenantID, &supplement.EmployeeID, &supplement.EmployeeName, &supplement.DepartmentID,
			&supplement.SupplementDate, &supplement.SupplementType, &supplement.SupplementTime,
			&supplement.MissingType, &supplement.Reason, &evidenceJSON,
			&supplement.AttendanceRecordID,
			&supplement.ApprovalID, &supplement.ApprovalStatus, &supplement.ApprovedBy, &supplement.ApprovedAt, &supplement.RejectReason,
			&supplement.ProcessStatus, &supplement.ProcessedAt, &supplement.ProcessedBy,
			&supplement.Remark, &supplement.CreatedAt, &supplement.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 反序列化证明材料
		if len(evidenceJSON) > 0 {
			if err := json.Unmarshal(evidenceJSON, &supplement.Evidence); err != nil {
				return nil, fmt.Errorf("failed to unmarshal evidence: %w", err)
			}
		}

		supplements = append(supplements, supplement)
	}

	return supplements, rows.Err()
}

// FindByDate 查询指定日期的补卡申请
func (r *PunchCardSupplementRepo) FindByDate(ctx context.Context, tenantID, employeeID uuid.UUID, date time.Time, supplementType model.SupplementType) (*model.PunchCardSupplement, error) {
	query := `
		SELECT 
			id, tenant_id, employee_id, employee_name, department_id,
			supplement_date, supplement_type, supplement_time,
			missing_type, reason, evidence,
			attendance_record_id,
			approval_id, approval_status, approved_by, approved_at, reject_reason,
			process_status, processed_at, processed_by,
			remark, created_at, updated_at
		FROM hrm.punch_card_supplements
		WHERE tenant_id = $1 
			AND employee_id = $2 
			AND supplement_date = $3
			AND supplement_type = $4
			AND approval_status != 'rejected'
			AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	supplement := &model.PunchCardSupplement{}
	var evidenceJSON []byte

	err := r.db.QueryRow(ctx, query, tenantID, employeeID, date, supplementType).Scan(
		&supplement.ID, &supplement.TenantID, &supplement.EmployeeID, &supplement.EmployeeName, &supplement.DepartmentID,
		&supplement.SupplementDate, &supplement.SupplementType, &supplement.SupplementTime,
		&supplement.MissingType, &supplement.Reason, &evidenceJSON,
		&supplement.AttendanceRecordID,
		&supplement.ApprovalID, &supplement.ApprovalStatus, &supplement.ApprovedBy, &supplement.ApprovedAt, &supplement.RejectReason,
		&supplement.ProcessStatus, &supplement.ProcessedAt, &supplement.ProcessedBy,
		&supplement.Remark, &supplement.CreatedAt, &supplement.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 未找到返回nil，不是错误
		}
		return nil, err
	}

	// 反序列化证明材料
	if len(evidenceJSON) > 0 {
		if err := json.Unmarshal(evidenceJSON, &supplement.Evidence); err != nil {
			return nil, fmt.Errorf("failed to unmarshal evidence: %w", err)
		}
	}

	return supplement, nil
}

// CountByEmployee 统计员工补卡次数
func (r *PunchCardSupplementRepo) CountByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM hrm.punch_card_supplements
		WHERE tenant_id = $1 
			AND employee_id = $2 
			AND supplement_date >= $3 
			AND supplement_date < $4
			AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, query, tenantID, employeeID, startDate, endDate).Scan(&count)
	return count, err
}
