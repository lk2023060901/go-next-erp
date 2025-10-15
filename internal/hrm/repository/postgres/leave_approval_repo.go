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

type leaveApprovalRepo struct {
	db *database.DB
}

// NewLeaveApprovalRepository 创建请假审批记录仓储
func NewLeaveApprovalRepository(db *database.DB) repository.LeaveApprovalRepository {
	return &leaveApprovalRepo{db: db}
}

func (r *leaveApprovalRepo) Create(ctx context.Context, approval *model.LeaveApproval) error {
	sql := `
		INSERT INTO hrm_leave_approvals (
			id, tenant_id, leave_request_id, approver_id, approver_name,
			level, status, action, comment, approved_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12
		)
	`

	_, err := r.db.Exec(ctx, sql,
		approval.ID, approval.TenantID, approval.LeaveRequestID, approval.ApproverID, approval.ApproverName,
		approval.Level, approval.Status, approval.Action, approval.Comment, approval.ApprovedAt,
		approval.CreatedAt, approval.UpdatedAt,
	)

	return err
}

func (r *leaveApprovalRepo) Update(ctx context.Context, approval *model.LeaveApproval) error {
	sql := `
		UPDATE hrm_leave_approvals SET
			status = $1, action = $2, comment = $3, approved_at = $4, updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.Exec(ctx, sql,
		approval.Status, approval.Action, approval.Comment, approval.ApprovedAt, approval.UpdatedAt,
		approval.ID,
	)

	return err
}

func (r *leaveApprovalRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveApproval, error) {
	sql := `
		SELECT id, tenant_id, leave_request_id, approver_id, approver_name,
		       level, status, action, comment, approved_at,
		       created_at, updated_at
		FROM hrm_leave_approvals
		WHERE id = $1
	`

	approval := &model.LeaveApproval{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&approval.ID, &approval.TenantID, &approval.LeaveRequestID, &approval.ApproverID, &approval.ApproverName,
		&approval.Level, &approval.Status, &approval.Action, &approval.Comment, &approval.ApprovedAt,
		&approval.CreatedAt, &approval.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("leave approval not found")
		}
		return nil, err
	}

	return approval, nil
}

func (r *leaveApprovalRepo) ListByRequest(ctx context.Context, leaveRequestID uuid.UUID) ([]*model.LeaveApproval, error) {
	sql := `
		SELECT id, tenant_id, leave_request_id, approver_id, approver_name,
		       level, status, action, comment, approved_at,
		       created_at, updated_at
		FROM hrm_leave_approvals
		WHERE leave_request_id = $1
		ORDER BY level ASC, created_at ASC
	`

	rows, err := r.db.Query(ctx, sql, leaveRequestID)
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

	return approvals, rows.Err()
}

func (r *leaveApprovalRepo) FindPendingApproval(ctx context.Context, leaveRequestID uuid.UUID, approverID uuid.UUID) (*model.LeaveApproval, error) {
	sql := `
		SELECT id, tenant_id, leave_request_id, approver_id, approver_name,
		       level, status, action, comment, approved_at,
		       created_at, updated_at
		FROM hrm_leave_approvals
		WHERE leave_request_id = $1 AND approver_id = $2 AND status = $3
		ORDER BY level ASC
		LIMIT 1
	`

	approval := &model.LeaveApproval{}
	err := r.db.QueryRow(ctx, sql, leaveRequestID, approverID, model.LeaveApprovalStatusPending).Scan(
		&approval.ID, &approval.TenantID, &approval.LeaveRequestID, &approval.ApproverID, &approval.ApproverName,
		&approval.Level, &approval.Status, &approval.Action, &approval.Comment, &approval.ApprovedAt,
		&approval.CreatedAt, &approval.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("pending approval not found")
		}
		return nil, err
	}

	return approval, nil
}

func (r *leaveApprovalRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.LeaveApprovalStatus, action *model.LeaveApprovalAction, comment string, approvedAt *time.Time) error {
	sql := `
		UPDATE hrm_leave_approvals 
		SET status = $1, action = $2, comment = $3, approved_at = $4, updated_at = NOW()
		WHERE id = $5
	`

	_, err := r.db.Exec(ctx, sql, status, action, comment, approvedAt, id)
	return err
}

func (r *leaveApprovalRepo) BatchCreate(ctx context.Context, approvals []*model.LeaveApproval) error {
	if len(approvals) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql := `
		INSERT INTO hrm_leave_approvals (
			id, tenant_id, leave_request_id, approver_id, approver_name,
			level, status, action, comment, approved_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	for _, approval := range approvals {
		_, err := tx.Exec(ctx, sql,
			approval.ID, approval.TenantID, approval.LeaveRequestID, approval.ApproverID, approval.ApproverName,
			approval.Level, approval.Status, approval.Action, approval.Comment, approval.ApprovedAt,
			approval.CreatedAt, approval.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
