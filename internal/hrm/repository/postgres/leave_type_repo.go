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

type leaveTypeRepo struct {
	db *database.DB
}

// NewLeaveTypeRepository 创建请假类型仓储
func NewLeaveTypeRepository(db *database.DB) repository.LeaveTypeRepository {
	return &leaveTypeRepo{db: db}
}

func (r *leaveTypeRepo) Create(ctx context.Context, leaveType *model.LeaveType) error {
	// 序列化审批规则
	var approvalRulesJSON []byte
	if leaveType.ApprovalRules != nil {
		var err error
		approvalRulesJSON, err = json.Marshal(leaveType.ApprovalRules)
		if err != nil {
			return fmt.Errorf("failed to marshal approval rules: %w", err)
		}
	}

	sql := `
		INSERT INTO hrm_leave_types (
			id, tenant_id, code, name, description,
			is_paid, requires_approval, requires_proof, deduct_quota,
			unit, min_duration, max_duration, advance_days,
			approval_rules, color, is_active, sort,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15, $16, $17,
			$18, $19, $20, $21
		)
	`

	_, err := r.db.Exec(ctx, sql,
		leaveType.ID, leaveType.TenantID, leaveType.Code, leaveType.Name, leaveType.Description,
		leaveType.IsPaid, leaveType.RequiresApproval, leaveType.RequiresProof, leaveType.DeductQuota,
		leaveType.Unit, leaveType.MinDuration, leaveType.MaxDuration, leaveType.AdvanceDays,
		approvalRulesJSON, leaveType.Color, leaveType.IsActive, leaveType.Sort,
		leaveType.CreatedBy, leaveType.UpdatedBy, leaveType.CreatedAt, leaveType.UpdatedAt,
	)

	return err
}

func (r *leaveTypeRepo) Update(ctx context.Context, leaveType *model.LeaveType) error {
	// 序列化审批规则
	var approvalRulesJSON []byte
	if leaveType.ApprovalRules != nil {
		var err error
		approvalRulesJSON, err = json.Marshal(leaveType.ApprovalRules)
		if err != nil {
			return fmt.Errorf("failed to marshal approval rules: %w", err)
		}
	}

	sql := `
		UPDATE hrm_leave_types SET
			name = $1, description = $2,
			is_paid = $3, requires_approval = $4, requires_proof = $5, deduct_quota = $6,
			unit = $7, min_duration = $8, max_duration = $9, advance_days = $10,
			approval_rules = $11, color = $12, is_active = $13, sort = $14,
			updated_by = $15, updated_at = $16
		WHERE id = $17 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql,
		leaveType.Name, leaveType.Description,
		leaveType.IsPaid, leaveType.RequiresApproval, leaveType.RequiresProof, leaveType.DeductQuota,
		leaveType.Unit, leaveType.MinDuration, leaveType.MaxDuration, leaveType.AdvanceDays,
		approvalRulesJSON, leaveType.Color, leaveType.IsActive, leaveType.Sort,
		leaveType.UpdatedBy, leaveType.UpdatedAt,
		leaveType.ID,
	)

	return err
}

func (r *leaveTypeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `UPDATE hrm_leave_types SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *leaveTypeRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.LeaveType, error) {
	sql := `
		SELECT id, tenant_id, code, name, description,
		       is_paid, requires_approval, requires_proof, deduct_quota,
		       unit, min_duration, max_duration, advance_days,
		       approval_rules, color, is_active, sort,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM hrm_leave_types
		WHERE id = $1 AND deleted_at IS NULL
	`

	leaveType := &model.LeaveType{}
	var approvalRulesJSON []byte

	err := r.db.QueryRow(ctx, sql, id).Scan(
		&leaveType.ID, &leaveType.TenantID, &leaveType.Code, &leaveType.Name, &leaveType.Description,
		&leaveType.IsPaid, &leaveType.RequiresApproval, &leaveType.RequiresProof, &leaveType.DeductQuota,
		&leaveType.Unit, &leaveType.MinDuration, &leaveType.MaxDuration, &leaveType.AdvanceDays,
		&approvalRulesJSON, &leaveType.Color, &leaveType.IsActive, &leaveType.Sort,
		&leaveType.CreatedBy, &leaveType.UpdatedBy, &leaveType.CreatedAt, &leaveType.UpdatedAt, &leaveType.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("leave type not found")
		}
		return nil, err
	}

	// 解析审批规则
	if len(approvalRulesJSON) > 0 {
		var rules model.ApprovalRules
		if err := json.Unmarshal(approvalRulesJSON, &rules); err == nil {
			leaveType.ApprovalRules = &rules
		}
	}

	return leaveType, nil
}

func (r *leaveTypeRepo) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.LeaveType, error) {
	sql := `
		SELECT id, tenant_id, code, name, description,
		       is_paid, requires_approval, requires_proof, deduct_quota,
		       unit, min_duration, max_duration, advance_days,
		       approval_rules, color, is_active, sort,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM hrm_leave_types
		WHERE tenant_id = $1 AND code = $2 AND deleted_at IS NULL
	`

	leaveType := &model.LeaveType{}
	var approvalRulesJSON []byte

	err := r.db.QueryRow(ctx, sql, tenantID, code).Scan(
		&leaveType.ID, &leaveType.TenantID, &leaveType.Code, &leaveType.Name, &leaveType.Description,
		&leaveType.IsPaid, &leaveType.RequiresApproval, &leaveType.RequiresProof, &leaveType.DeductQuota,
		&leaveType.Unit, &leaveType.MinDuration, &leaveType.MaxDuration, &leaveType.AdvanceDays,
		&approvalRulesJSON, &leaveType.Color, &leaveType.IsActive, &leaveType.Sort,
		&leaveType.CreatedBy, &leaveType.UpdatedBy, &leaveType.CreatedAt, &leaveType.UpdatedAt, &leaveType.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("leave type not found")
		}
		return nil, err
	}

	// 解析审批规则
	if len(approvalRulesJSON) > 0 {
		var rules model.ApprovalRules
		if err := json.Unmarshal(approvalRulesJSON, &rules); err == nil {
			leaveType.ApprovalRules = &rules
		}
	}

	return leaveType, nil
}

func (r *leaveTypeRepo) List(ctx context.Context, tenantID uuid.UUID, filter *repository.LeaveTypeFilter, offset, limit int) ([]*model.LeaveType, int, error) {
	// 构建查询条件
	where := "tenant_id = $1 AND deleted_at IS NULL"
	args := []interface{}{tenantID}
	argIdx := 2

	if filter != nil {
		if filter.IsActive != nil {
			where += fmt.Sprintf(" AND is_active = $%d", argIdx)
			args = append(args, *filter.IsActive)
			argIdx++
		}
		if filter.RequiresProof != nil {
			where += fmt.Sprintf(" AND requires_proof = $%d", argIdx)
			args = append(args, *filter.RequiresProof)
			argIdx++
		}
		if filter.DeductQuota != nil {
			where += fmt.Sprintf(" AND deduct_quota = $%d", argIdx)
			args = append(args, *filter.DeductQuota)
			argIdx++
		}
		if filter.Keyword != "" {
			where += fmt.Sprintf(" AND (name LIKE $%d OR code LIKE $%d)", argIdx, argIdx)
			args = append(args, "%"+filter.Keyword+"%")
			argIdx++
		}
	}

	// 查询总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM hrm_leave_types WHERE %s", where)
	var total int
	err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	dataSQL := fmt.Sprintf(`
		SELECT id, tenant_id, code, name, is_paid, requires_approval,
		       deduct_quota, unit, color, is_active, sort, created_at
		FROM hrm_leave_types
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

	var leaveTypes []*model.LeaveType
	for rows.Next() {
		lt := &model.LeaveType{}
		err := rows.Scan(
			&lt.ID, &lt.TenantID, &lt.Code, &lt.Name, &lt.IsPaid, &lt.RequiresApproval,
			&lt.DeductQuota, &lt.Unit, &lt.Color, &lt.IsActive, &lt.Sort, &lt.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		leaveTypes = append(leaveTypes, lt)
	}

	return leaveTypes, total, rows.Err()
}

func (r *leaveTypeRepo) ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.LeaveType, error) {
	sql := `
		SELECT id, tenant_id, code, name, description,
		       is_paid, requires_approval, requires_proof, deduct_quota,
		       unit, min_duration, max_duration, advance_days,
		       color, sort
		FROM hrm_leave_types
		WHERE tenant_id = $1 AND is_active = true AND deleted_at IS NULL
		ORDER BY sort ASC, created_at DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaveTypes []*model.LeaveType
	for rows.Next() {
		lt := &model.LeaveType{}
		err := rows.Scan(
			&lt.ID, &lt.TenantID, &lt.Code, &lt.Name, &lt.Description,
			&lt.IsPaid, &lt.RequiresApproval, &lt.RequiresProof, &lt.DeductQuota,
			&lt.Unit, &lt.MinDuration, &lt.MaxDuration, &lt.AdvanceDays,
			&lt.Color, &lt.Sort,
		)
		if err != nil {
			return nil, err
		}
		leaveTypes = append(leaveTypes, lt)
	}

	return leaveTypes, rows.Err()
}

// ListWithCursor 游标分页查询请假类型（高性能）
func (r *leaveTypeRepo) ListWithCursor(
	ctx context.Context,
	tenantID uuid.UUID,
	filter *repository.LeaveTypeFilter,
	cursor *time.Time,
	limit int,
) ([]*model.LeaveType, *time.Time, bool, error) {
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
		if filter.IsActive != nil {
			argIdx++
			where += fmt.Sprintf(" AND is_active = $%d", argIdx)
			args = append(args, *filter.IsActive)
		}
		if filter.RequiresProof != nil {
			argIdx++
			where += fmt.Sprintf(" AND requires_proof = $%d", argIdx)
			args = append(args, *filter.RequiresProof)
		}
		if filter.DeductQuota != nil {
			argIdx++
			where += fmt.Sprintf(" AND deduct_quota = $%d", argIdx)
			args = append(args, *filter.DeductQuota)
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
		SELECT id, tenant_id, code, name, is_paid, requires_approval,
		       deduct_quota, unit, color, is_active, sort, created_at
		FROM hrm_leave_types
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
	var leaveTypes []*model.LeaveType
	for rows.Next() {
		lt := &model.LeaveType{}
		err := rows.Scan(
			&lt.ID, &lt.TenantID, &lt.Code, &lt.Name, &lt.IsPaid, &lt.RequiresApproval,
			&lt.DeductQuota, &lt.Unit, &lt.Color, &lt.IsActive, &lt.Sort, &lt.CreatedAt,
		)
		if err != nil {
			return nil, nil, false, err
		}
		leaveTypes = append(leaveTypes, lt)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, false, err
	}

	// 判断是否有下一页
	hasNext := len(leaveTypes) > limit
	if hasNext {
		leaveTypes = leaveTypes[:limit]
	}

	// 生成下一页游标
	var nextCursor *time.Time
	if hasNext && len(leaveTypes) > 0 {
		lastType := leaveTypes[len(leaveTypes)-1]
		nextCursor = &lastType.CreatedAt
	}

	return leaveTypes, nextCursor, hasNext, nil
}
