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

type leaveQuotaRepo struct {
	db *database.DB
}

// NewLeaveQuotaRepository 创建请假额度仓储
func NewLeaveQuotaRepository(db *database.DB) repository.LeaveQuotaRepository {
	return &leaveQuotaRepo{db: db}
}

func (r *leaveQuotaRepo) Create(ctx context.Context, quota *model.LeaveQuota) error {
	sql := `
		INSERT INTO hrm_leave_quotas (
			id, tenant_id, employee_id, leave_type_id, year,
			total_quota, used_quota, pending_quota, expired_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11
		)
	`

	_, err := r.db.Exec(ctx, sql,
		quota.ID, quota.TenantID, quota.EmployeeID, quota.LeaveTypeID, quota.Year,
		quota.TotalQuota, quota.UsedQuota, quota.PendingQuota, quota.ExpiredAt,
		quota.CreatedAt, quota.UpdatedAt,
	)

	return err
}

func (r *leaveQuotaRepo) Update(ctx context.Context, quota *model.LeaveQuota) error {
	sql := `
		UPDATE hrm_leave_quotas SET
			total_quota = $1, used_quota = $2, pending_quota = $3,
			expired_at = $4, updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.Exec(ctx, sql,
		quota.TotalQuota, quota.UsedQuota, quota.PendingQuota,
		quota.ExpiredAt, quota.UpdatedAt,
		quota.ID,
	)

	return err
}

func (r *leaveQuotaRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveQuota, error) {
	sql := `
		SELECT id, tenant_id, employee_id, leave_type_id, year,
		       total_quota, used_quota, pending_quota, expired_at,
		       created_at, updated_at
		FROM hrm_leave_quotas
		WHERE id = $1
	`

	quota := &model.LeaveQuota{}
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&quota.ID, &quota.TenantID, &quota.EmployeeID, &quota.LeaveTypeID, &quota.Year,
		&quota.TotalQuota, &quota.UsedQuota, &quota.PendingQuota, &quota.ExpiredAt,
		&quota.CreatedAt, &quota.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("leave quota not found")
		}
		return nil, err
	}

	return quota, nil
}

func (r *leaveQuotaRepo) FindByEmployeeAndType(ctx context.Context, tenantID, employeeID, leaveTypeID uuid.UUID, year int) (*model.LeaveQuota, error) {
	sql := `
		SELECT id, tenant_id, employee_id, leave_type_id, year,
		       total_quota, used_quota, pending_quota, expired_at,
		       created_at, updated_at
		FROM hrm_leave_quotas
		WHERE tenant_id = $1 AND employee_id = $2 AND leave_type_id = $3 AND year = $4
	`

	quota := &model.LeaveQuota{}
	err := r.db.QueryRow(ctx, sql, tenantID, employeeID, leaveTypeID, year).Scan(
		&quota.ID, &quota.TenantID, &quota.EmployeeID, &quota.LeaveTypeID, &quota.Year,
		&quota.TotalQuota, &quota.UsedQuota, &quota.PendingQuota, &quota.ExpiredAt,
		&quota.CreatedAt, &quota.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("leave quota not found")
		}
		return nil, err
	}

	return quota, nil
}

func (r *leaveQuotaRepo) ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.LeaveQuota, error) {
	sql := `
		SELECT id, tenant_id, employee_id, leave_type_id, year,
		       total_quota, used_quota, pending_quota, expired_at,
		       created_at, updated_at
		FROM hrm_leave_quotas
		WHERE tenant_id = $1 AND employee_id = $2 AND year = $3
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, employeeID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotas []*model.LeaveQuota
	for rows.Next() {
		quota := &model.LeaveQuota{}
		err := rows.Scan(
			&quota.ID, &quota.TenantID, &quota.EmployeeID, &quota.LeaveTypeID, &quota.Year,
			&quota.TotalQuota, &quota.UsedQuota, &quota.PendingQuota, &quota.ExpiredAt,
			&quota.CreatedAt, &quota.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		quotas = append(quotas, quota)
	}

	return quotas, rows.Err()
}

func (r *leaveQuotaRepo) ListByEmployeeWithType(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.LeaveQuotaWithType, error) {
	sql := `
		SELECT q.id, q.tenant_id, q.employee_id, q.leave_type_id, q.year,
		       q.total_quota, q.used_quota, q.pending_quota, q.expired_at,
		       q.created_at, q.updated_at,
		       lt.code, lt.name, lt.unit, lt.color, lt.is_paid, lt.deduct_quota
		FROM hrm_leave_quotas q
		INNER JOIN hrm_leave_types lt ON q.leave_type_id = lt.id
		WHERE q.tenant_id = $1 AND q.employee_id = $2 AND q.year = $3 AND lt.deleted_at IS NULL
		ORDER BY lt.sort ASC, q.created_at ASC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, employeeID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotasWithType []*model.LeaveQuotaWithType
	for rows.Next() {
		qt := &model.LeaveQuotaWithType{
			LeaveType: &model.LeaveType{},
		}
		err := rows.Scan(
			&qt.ID, &qt.TenantID, &qt.EmployeeID, &qt.LeaveTypeID, &qt.Year,
			&qt.TotalQuota, &qt.UsedQuota, &qt.PendingQuota, &qt.ExpiredAt,
			&qt.CreatedAt, &qt.UpdatedAt,
			&qt.LeaveType.Code, &qt.LeaveType.Name, &qt.LeaveType.Unit,
			&qt.LeaveType.Color, &qt.LeaveType.IsPaid, &qt.LeaveType.DeductQuota,
		)
		if err != nil {
			return nil, err
		}
		qt.LeaveType.ID = qt.LeaveTypeID
		quotasWithType = append(quotasWithType, qt)
	}

	return quotasWithType, rows.Err()
}

func (r *leaveQuotaRepo) IncrementUsedQuota(ctx context.Context, id uuid.UUID, amount float64) error {
	sql := `
		UPDATE hrm_leave_quotas 
		SET used_quota = used_quota + $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, sql, amount, id)
	return err
}

func (r *leaveQuotaRepo) DecrementUsedQuota(ctx context.Context, id uuid.UUID, amount float64) error {
	sql := `
		UPDATE hrm_leave_quotas 
		SET used_quota = GREATEST(0, used_quota - $1), updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, sql, amount, id)
	return err
}

func (r *leaveQuotaRepo) IncrementPendingQuota(ctx context.Context, id uuid.UUID, amount float64) error {
	sql := `
		UPDATE hrm_leave_quotas 
		SET pending_quota = pending_quota + $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, sql, amount, id)
	return err
}

func (r *leaveQuotaRepo) DecrementPendingQuota(ctx context.Context, id uuid.UUID, amount float64) error {
	sql := `
		UPDATE hrm_leave_quotas 
		SET pending_quota = GREATEST(0, pending_quota - $1), updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, sql, amount, id)
	return err
}

func (r *leaveQuotaRepo) BatchCreate(ctx context.Context, quotas []*model.LeaveQuota) error {
	if len(quotas) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql := `
		INSERT INTO hrm_leave_quotas (
			id, tenant_id, employee_id, leave_type_id, year,
			total_quota, used_quota, pending_quota, expired_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	for _, quota := range quotas {
		_, err := tx.Exec(ctx, sql,
			quota.ID, quota.TenantID, quota.EmployeeID, quota.LeaveTypeID, quota.Year,
			quota.TotalQuota, quota.UsedQuota, quota.PendingQuota, quota.ExpiredAt,
			quota.CreatedAt, quota.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
