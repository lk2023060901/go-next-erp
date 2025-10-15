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

type leaveRequestRepo struct {
	db *database.DB
}

// NewLeaveRequestRepository 创建请假申请仓储
func NewLeaveRequestRepository(db *database.DB) repository.LeaveRequestRepository {
	return &leaveRequestRepo{db: db}
}

func (r *leaveRequestRepo) Create(ctx context.Context, request *model.LeaveRequest) error {
	proofURLsJSON, _ := json.Marshal(request.ProofURLs)

	sql := `
		INSERT INTO hrm_leave_requests (
			id, tenant_id, employee_id, employee_name, department_id,
			leave_type_id, leave_type_name, start_time, end_time, duration, unit,
			reason, proof_urls, status, current_approver_id,
			submitted_at, remark,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17,
			$18, $19, $20, $21
		)
	`

	_, err := r.db.Exec(ctx, sql,
		request.ID, request.TenantID, request.EmployeeID, request.EmployeeName, request.DepartmentID,
		request.LeaveTypeID, request.LeaveTypeName, request.StartTime, request.EndTime, request.Duration, request.Unit,
		request.Reason, proofURLsJSON, request.Status, request.CurrentApproverID,
		request.SubmittedAt, request.Remark,
		request.CreatedBy, request.UpdatedBy, request.CreatedAt, request.UpdatedAt,
	)

	return err
}

func (r *leaveRequestRepo) Update(ctx context.Context, request *model.LeaveRequest) error {
	proofURLsJSON, _ := json.Marshal(request.ProofURLs)

	sql := `
		UPDATE hrm_leave_requests SET
			start_time = $1, end_time = $2, duration = $3, unit = $4,
			reason = $5, proof_urls = $6, status = $7, current_approver_id = $8,
			submitted_at = $9, approved_at = $10, rejected_at = $11, cancelled_at = $12,
			remark = $13, updated_by = $14, updated_at = $15
		WHERE id = $16 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql,
		request.StartTime, request.EndTime, request.Duration, request.Unit,
		request.Reason, proofURLsJSON, request.Status, request.CurrentApproverID,
		request.SubmittedAt, request.ApprovedAt, request.RejectedAt, request.CancelledAt,
		request.Remark, request.UpdatedBy, request.UpdatedAt,
		request.ID,
	)

	return err
}

func (r *leaveRequestRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `UPDATE hrm_leave_requests SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *leaveRequestRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveRequest, error) {
	sql := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
		       leave_type_id, leave_type_name, start_time, end_time, duration, unit,
		       reason, proof_urls, status, current_approver_id,
		       submitted_at, approved_at, rejected_at, cancelled_at, remark,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM hrm_leave_requests
		WHERE id = $1 AND deleted_at IS NULL
	`

	request := &model.LeaveRequest{}
	var proofURLsJSON []byte

	err := r.db.QueryRow(ctx, sql, id).Scan(
		&request.ID, &request.TenantID, &request.EmployeeID, &request.EmployeeName, &request.DepartmentID,
		&request.LeaveTypeID, &request.LeaveTypeName, &request.StartTime, &request.EndTime, &request.Duration, &request.Unit,
		&request.Reason, &proofURLsJSON, &request.Status, &request.CurrentApproverID,
		&request.SubmittedAt, &request.ApprovedAt, &request.RejectedAt, &request.CancelledAt, &request.Remark,
		&request.CreatedBy, &request.UpdatedBy, &request.CreatedAt, &request.UpdatedAt, &request.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("leave request not found")
		}
		return nil, err
	}

	// 解析proof_urls
	if len(proofURLsJSON) > 0 {
		json.Unmarshal(proofURLsJSON, &request.ProofURLs)
	}

	return request, nil
}

func (r *leaveRequestRepo) FindByIDWithApprovals(ctx context.Context, id uuid.UUID) (*model.LeaveRequestWithApprovals, error) {
	// 先查询请假申请
	request, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 查询审批记录
	approvalSQL := `
		SELECT id, tenant_id, leave_request_id, approver_id, approver_name,
		       level, status, action, comment, approved_at,
		       created_at, updated_at
		FROM hrm_leave_approvals
		WHERE leave_request_id = $1
		ORDER BY level ASC, created_at ASC
	`

	rows, err := r.db.Query(ctx, approvalSQL, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []*model.LeaveApproval
	for rows.Next() {
		approval := &model.LeaveApproval{}
		err := rows.Scan(
			&approval.ID, &approval.TenantID, &approval.LeaveRequestID, &approval.ApproverID, &approval.ApproverName,
			&approval.Level, &approval.Status, &approval.Action, &approval.Comment, &approval.ApprovedAt,
			&approval.CreatedAt, &approval.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		approvals = append(approvals, approval)
	}

	return &model.LeaveRequestWithApprovals{
		LeaveRequest: *request,
		Approvals:    approvals,
	}, rows.Err()
}

func (r *leaveRequestRepo) List(ctx context.Context, tenantID uuid.UUID, filter *repository.LeaveRequestFilter, offset, limit int) ([]*model.LeaveRequest, int, error) {
	// 构建查询条件
	where := "tenant_id = $1 AND deleted_at IS NULL"
	args := []interface{}{tenantID}
	argIdx := 2

	if filter != nil {
		if filter.LeaveTypeID != nil {
			where += fmt.Sprintf(" AND leave_type_id = $%d", argIdx)
			args = append(args, *filter.LeaveTypeID)
			argIdx++
		}
		if filter.DepartmentID != nil {
			where += fmt.Sprintf(" AND department_id = $%d", argIdx)
			args = append(args, *filter.DepartmentID)
			argIdx++
		}
		if filter.Status != nil {
			where += fmt.Sprintf(" AND status = $%d", argIdx)
			args = append(args, *filter.Status)
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
			where += fmt.Sprintf(" AND (employee_name LIKE $%d OR reason LIKE $%d)", argIdx, argIdx)
			args = append(args, "%"+filter.Keyword+"%")
			argIdx++
		}
	}

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM hrm_leave_requests WHERE %s", where)
	var total int
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	dataSQL := fmt.Sprintf(`
		SELECT id, tenant_id, employee_id, employee_name, leave_type_id, leave_type_name,
		       start_time, end_time, duration, unit, status, created_at
		FROM hrm_leave_requests
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

	var requests []*model.LeaveRequest
	for rows.Next() {
		req := &model.LeaveRequest{}
		err := rows.Scan(
			&req.ID, &req.TenantID, &req.EmployeeID, &req.EmployeeName, &req.LeaveTypeID, &req.LeaveTypeName,
			&req.StartTime, &req.EndTime, &req.Duration, &req.Unit, &req.Status, &req.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		requests = append(requests, req)
	}

	return requests, total, rows.Err()
}

func (r *leaveRequestRepo) ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, filter *repository.LeaveRequestFilter, offset, limit int) ([]*model.LeaveRequest, int, error) {
	// 构建查询条件
	where := "tenant_id = $1 AND employee_id = $2 AND deleted_at IS NULL"
	args := []interface{}{tenantID, employeeID}
	argIdx := 3

	if filter != nil {
		if filter.LeaveTypeID != nil {
			where += fmt.Sprintf(" AND leave_type_id = $%d", argIdx)
			args = append(args, *filter.LeaveTypeID)
			argIdx++
		}
		if filter.Status != nil {
			where += fmt.Sprintf(" AND status = $%d", argIdx)
			args = append(args, *filter.Status)
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
	}

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM hrm_leave_requests WHERE %s", where)
	var total int
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	dataSQL := fmt.Sprintf(`
		SELECT id, tenant_id, employee_id, employee_name, leave_type_id, leave_type_name,
		       start_time, end_time, duration, unit, status, submitted_at, created_at
		FROM hrm_leave_requests
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

	var requests []*model.LeaveRequest
	for rows.Next() {
		req := &model.LeaveRequest{}
		err := rows.Scan(
			&req.ID, &req.TenantID, &req.EmployeeID, &req.EmployeeName, &req.LeaveTypeID, &req.LeaveTypeName,
			&req.StartTime, &req.EndTime, &req.Duration, &req.Unit, &req.Status, &req.SubmittedAt, &req.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		requests = append(requests, req)
	}

	return requests, total, rows.Err()
}

func (r *leaveRequestRepo) ListPendingApprovals(ctx context.Context, tenantID, approverID uuid.UUID, offset, limit int) ([]*model.LeaveRequest, int, error) {
	where := "tenant_id = $1 AND current_approver_id = $2 AND status = $3 AND deleted_at IS NULL"
	args := []interface{}{tenantID, approverID, model.LeaveRequestStatusPending}

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM hrm_leave_requests WHERE %s", where)
	var total int
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	dataSQL := fmt.Sprintf(`
		SELECT id, tenant_id, employee_id, employee_name, leave_type_id, leave_type_name,
		       start_time, end_time, duration, unit, reason, status, submitted_at, created_at
		FROM hrm_leave_requests
		WHERE %s
		ORDER BY submitted_at ASC, created_at ASC
		LIMIT $4 OFFSET $5
	`, where)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var requests []*model.LeaveRequest
	for rows.Next() {
		req := &model.LeaveRequest{}
		err := rows.Scan(
			&req.ID, &req.TenantID, &req.EmployeeID, &req.EmployeeName, &req.LeaveTypeID, &req.LeaveTypeName,
			&req.StartTime, &req.EndTime, &req.Duration, &req.Unit, &req.Reason, &req.Status, &req.SubmittedAt, &req.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		requests = append(requests, req)
	}

	return requests, total, rows.Err()
}

func (r *leaveRequestRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.LeaveRequestStatus, operatedAt *time.Time) error {
	var sql string
	var args []interface{}

	switch status {
	case model.LeaveRequestStatusPending:
		sql = `UPDATE hrm_leave_requests SET status = $1, submitted_at = $2, updated_at = NOW() WHERE id = $3`
		args = []interface{}{status, operatedAt, id}
	case model.LeaveRequestStatusApproved:
		sql = `UPDATE hrm_leave_requests SET status = $1, approved_at = $2, updated_at = NOW() WHERE id = $3`
		args = []interface{}{status, operatedAt, id}
	case model.LeaveRequestStatusRejected:
		sql = `UPDATE hrm_leave_requests SET status = $1, rejected_at = $2, updated_at = NOW() WHERE id = $3`
		args = []interface{}{status, operatedAt, id}
	case model.LeaveRequestStatusCancelled:
		sql = `UPDATE hrm_leave_requests SET status = $1, cancelled_at = $2, updated_at = NOW() WHERE id = $3`
		args = []interface{}{status, operatedAt, id}
	default:
		sql = `UPDATE hrm_leave_requests SET status = $1, updated_at = NOW() WHERE id = $2`
		args = []interface{}{status, id}
	}

	_, err := r.db.Exec(ctx, sql, args...)
	return err
}

func (r *leaveRequestRepo) SetCurrentApprover(ctx context.Context, id uuid.UUID, approverID *uuid.UUID) error {
	sql := `UPDATE hrm_leave_requests SET current_approver_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, sql, approverID, id)
	return err
}

func (r *leaveRequestRepo) CheckTimeConflict(ctx context.Context, tenantID, employeeID uuid.UUID, startTime, endTime time.Time, excludeID *uuid.UUID) (bool, error) {
	where := `
		tenant_id = $1 AND employee_id = $2 
		AND status = $3 
		AND deleted_at IS NULL
		AND (
			(start_time <= $4 AND end_time >= $5) OR
			(start_time <= $6 AND end_time >= $7) OR
			(start_time >= $8 AND end_time <= $9)
		)
	`
	args := []interface{}{
		tenantID, employeeID, model.LeaveRequestStatusApproved,
		startTime, startTime, endTime, endTime,
		startTime, endTime,
	}

	if excludeID != nil {
		where += " AND id != $10"
		args = append(args, *excludeID)
	}

	sql := fmt.Sprintf("SELECT COUNT(*) FROM hrm_leave_requests WHERE %s", where)
	var count int
	err := r.db.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// 游标分页实现
func (r *leaveRequestRepo) ListWithCursor(
	ctx context.Context,
	tenantID uuid.UUID,
	filter *repository.LeaveRequestFilter,
	cursor *time.Time,
	limit int,
) ([]*model.LeaveRequest, *time.Time, bool, error) {
	where := "tenant_id = $1 AND deleted_at IS NULL"
	args := []interface{}{tenantID}
	argIdx := 1

	if cursor != nil {
		argIdx++
		where += fmt.Sprintf(" AND created_at < $%d", argIdx)
		args = append(args, *cursor)
	}

	if filter != nil {
		if filter.LeaveTypeID != nil {
			argIdx++
			where += fmt.Sprintf(" AND leave_type_id = $%d", argIdx)
			args = append(args, *filter.LeaveTypeID)
		}
		if filter.DepartmentID != nil {
			argIdx++
			where += fmt.Sprintf(" AND department_id = $%d", argIdx)
			args = append(args, *filter.DepartmentID)
		}
		if filter.Status != nil {
			argIdx++
			where += fmt.Sprintf(" AND status = $%d", argIdx)
			args = append(args, *filter.Status)
		}
		if filter.StartDate != nil {
			argIdx++
			where += fmt.Sprintf(" AND start_time >= $%d", argIdx)
			args = append(args, *filter.StartDate)
		}
		if filter.EndDate != nil {
			argIdx++
			where += fmt.Sprintf(" AND end_time <= $%d", argIdx)
			args = append(args, *filter.EndDate)
		}
		if filter.Keyword != "" {
			argIdx++
			where += fmt.Sprintf(" AND (employee_name LIKE $%d OR reason LIKE $%d)", argIdx, argIdx)
			args = append(args, "%"+filter.Keyword+"%")
		}
	}

	argIdx++
	sql := fmt.Sprintf(`
		SELECT id, tenant_id, employee_id, employee_name, leave_type_id, leave_type_name,
		       start_time, end_time, duration, unit, status, created_at
		FROM hrm_leave_requests
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
	`, where, argIdx)
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, nil, false, err
	}
	defer rows.Close()

	var requests []*model.LeaveRequest
	for rows.Next() {
		req := &model.LeaveRequest{}
		err := rows.Scan(
			&req.ID, &req.TenantID, &req.EmployeeID, &req.EmployeeName, &req.LeaveTypeID, &req.LeaveTypeName,
			&req.StartTime, &req.EndTime, &req.Duration, &req.Unit, &req.Status, &req.CreatedAt,
		)
		if err != nil {
			return nil, nil, false, err
		}
		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, false, err
	}

	hasNext := len(requests) > limit
	if hasNext {
		requests = requests[:limit]
	}

	var nextCursor *time.Time
	if hasNext && len(requests) > 0 {
		lastReq := requests[len(requests)-1]
		nextCursor = &lastReq.CreatedAt
	}

	return requests, nextCursor, hasNext, nil
}

func (r *leaveRequestRepo) ListByEmployeeWithCursor(
	ctx context.Context,
	tenantID, employeeID uuid.UUID,
	filter *repository.LeaveRequestFilter,
	cursor *time.Time,
	limit int,
) ([]*model.LeaveRequest, *time.Time, bool, error) {
	where := "tenant_id = $1 AND employee_id = $2 AND deleted_at IS NULL"
	args := []interface{}{tenantID, employeeID}
	argIdx := 2

	if cursor != nil {
		argIdx++
		where += fmt.Sprintf(" AND created_at < $%d", argIdx)
		args = append(args, *cursor)
	}

	if filter != nil {
		if filter.LeaveTypeID != nil {
			argIdx++
			where += fmt.Sprintf(" AND leave_type_id = $%d", argIdx)
			args = append(args, *filter.LeaveTypeID)
		}
		if filter.Status != nil {
			argIdx++
			where += fmt.Sprintf(" AND status = $%d", argIdx)
			args = append(args, *filter.Status)
		}
		if filter.StartDate != nil {
			argIdx++
			where += fmt.Sprintf(" AND start_time >= $%d", argIdx)
			args = append(args, *filter.StartDate)
		}
		if filter.EndDate != nil {
			argIdx++
			where += fmt.Sprintf(" AND end_time <= $%d", argIdx)
			args = append(args, *filter.EndDate)
		}
	}

	argIdx++
	sql := fmt.Sprintf(`
		SELECT id, tenant_id, employee_id, employee_name, leave_type_id, leave_type_name,
		       start_time, end_time, duration, unit, status, submitted_at, created_at
		FROM hrm_leave_requests
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
	`, where, argIdx)
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, nil, false, err
	}
	defer rows.Close()

	var requests []*model.LeaveRequest
	for rows.Next() {
		req := &model.LeaveRequest{}
		err := rows.Scan(
			&req.ID, &req.TenantID, &req.EmployeeID, &req.EmployeeName, &req.LeaveTypeID, &req.LeaveTypeName,
			&req.StartTime, &req.EndTime, &req.Duration, &req.Unit, &req.Status, &req.SubmittedAt, &req.CreatedAt,
		)
		if err != nil {
			return nil, nil, false, err
		}
		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, false, err
	}

	hasNext := len(requests) > limit
	if hasNext {
		requests = requests[:limit]
	}

	var nextCursor *time.Time
	if hasNext && len(requests) > 0 {
		lastReq := requests[len(requests)-1]
		nextCursor = &lastReq.CreatedAt
	}

	return requests, nextCursor, hasNext, nil
}

func (r *leaveRequestRepo) ListPendingApprovalsWithCursor(
	ctx context.Context,
	tenantID, approverID uuid.UUID,
	cursor *time.Time,
	limit int,
) ([]*model.LeaveRequest, *time.Time, bool, error) {
	where := "tenant_id = $1 AND current_approver_id = $2 AND status = $3 AND deleted_at IS NULL"
	args := []interface{}{tenantID, approverID, model.LeaveRequestStatusPending}
	argIdx := 3

	if cursor != nil {
		argIdx++
		where += fmt.Sprintf(" AND created_at < $%d", argIdx)
		args = append(args, *cursor)
	}

	argIdx++
	sql := fmt.Sprintf(`
		SELECT id, tenant_id, employee_id, employee_name, leave_type_id, leave_type_name,
		       start_time, end_time, duration, unit, reason, status, submitted_at, created_at
		FROM hrm_leave_requests
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
	`, where, argIdx)
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, nil, false, err
	}
	defer rows.Close()

	var requests []*model.LeaveRequest
	for rows.Next() {
		req := &model.LeaveRequest{}
		err := rows.Scan(
			&req.ID, &req.TenantID, &req.EmployeeID, &req.EmployeeName, &req.LeaveTypeID, &req.LeaveTypeName,
			&req.StartTime, &req.EndTime, &req.Duration, &req.Unit, &req.Reason, &req.Status, &req.SubmittedAt, &req.CreatedAt,
		)
		if err != nil {
			return nil, nil, false, err
		}
		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, false, err
	}

	hasNext := len(requests) > limit
	if hasNext {
		requests = requests[:limit]
	}

	var nextCursor *time.Time
	if hasNext && len(requests) > 0 {
		lastReq := requests[len(requests)-1]
		nextCursor = &lastReq.CreatedAt
	}

	return requests, nextCursor, hasNext, nil
}
