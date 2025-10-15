package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/approval/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// ApprovalTaskRepository 审批任务仓储接口
type ApprovalTaskRepository interface {
	Create(ctx context.Context, task *model.ApprovalTask) error
	Update(ctx context.Context, task *model.ApprovalTask) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.ApprovalTask, error)
	ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]*model.ApprovalTask, error)
	ListByAssignee(ctx context.Context, assigneeID uuid.UUID, status *model.TaskStatus, limit, offset int) ([]*model.ApprovalTask, error)
	ListPendingByAssignee(ctx context.Context, assigneeID uuid.UUID) ([]*model.ApprovalTask, error)
	CountPendingByAssignee(ctx context.Context, assigneeID uuid.UUID) (int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.TaskStatus, action *model.ApprovalAction, comment *string, approvedAt *time.Time) error

	// 游标分页查询（高性能，适用于大数据量）
	ListByAssigneeWithCursor(ctx context.Context, assigneeID uuid.UUID, status *model.TaskStatus, cursor *time.Time, limit int) ([]*model.ApprovalTask, *time.Time, bool, error)
}

type approvalTaskRepo struct {
	db *database.DB
}

// NewApprovalTaskRepository 创建审批任务仓储
func NewApprovalTaskRepository(db *database.DB) ApprovalTaskRepository {
	return &approvalTaskRepo{db: db}
}

func (r *approvalTaskRepo) Create(ctx context.Context, task *model.ApprovalTask) error {
	sql := `
		INSERT INTO approval_tasks (
			id, tenant_id, process_instance_id, node_id, node_name,
			assignee_id, assignee_name, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(ctx, sql,
		task.ID,
		task.TenantID,
		task.ProcessInstanceID,
		task.NodeID,
		task.NodeName,
		task.AssigneeID,
		task.AssigneeName,
		task.Status,
		task.CreatedAt,
		task.UpdatedAt,
	)

	return err
}

func (r *approvalTaskRepo) Update(ctx context.Context, task *model.ApprovalTask) error {
	sql := `
		UPDATE approval_tasks
		SET status = $1, action = $2, comment = $3, approved_at = $4, updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.Exec(ctx, sql,
		task.Status,
		task.Action,
		task.Comment,
		task.ApprovedAt,
		task.UpdatedAt,
		task.ID,
	)

	return err
}

func (r *approvalTaskRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ApprovalTask, error) {
	sql := `
		SELECT id, tenant_id, process_instance_id, node_id, assignee_id,
		       status, action, comment, approved_at, created_at, updated_at
		FROM approval_tasks
		WHERE id = $1
	`

	var task model.ApprovalTask
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&task.ID,
		&task.TenantID,
		&task.ProcessInstanceID,
		&task.NodeID,
		&task.AssigneeID,
		&task.Status,
		&task.Action,
		&task.Comment,
		&task.ApprovedAt,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *approvalTaskRepo) ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]*model.ApprovalTask, error) {
	sql := `
		SELECT id, tenant_id, process_instance_id, node_id, assignee_id,
		       status, action, comment, approved_at, created_at, updated_at
		FROM approval_tasks
		WHERE process_instance_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, sql, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*model.ApprovalTask
	for rows.Next() {
		var task model.ApprovalTask
		err := rows.Scan(
			&task.ID,
			&task.TenantID,
			&task.ProcessInstanceID,
			&task.NodeID,
			&task.AssigneeID,
			&task.Status,
			&task.Action,
			&task.Comment,
			&task.ApprovedAt,
			&task.CreatedAt,
			&task.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		tasks = append(tasks, &task)
	}

	return tasks, rows.Err()
}

func (r *approvalTaskRepo) ListByAssignee(ctx context.Context, assigneeID uuid.UUID, status *model.TaskStatus, limit, offset int) ([]*model.ApprovalTask, error) {
	var sql string
	var args []interface{}

	if status != nil {
		sql = `
			SELECT id, tenant_id, process_instance_id, node_id, assignee_id,
			       status, action, comment, approved_at, created_at, updated_at
			FROM approval_tasks
			WHERE assignee_id = $1 AND status = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{assigneeID, *status, limit, offset}
	} else {
		sql = `
			SELECT id, tenant_id, process_instance_id, node_id, assignee_id,
			       status, action, comment, approved_at, created_at, updated_at
			FROM approval_tasks
			WHERE assignee_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{assigneeID, limit, offset}
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*model.ApprovalTask
	for rows.Next() {
		var task model.ApprovalTask
		err := rows.Scan(
			&task.ID,
			&task.TenantID,
			&task.ProcessInstanceID,
			&task.NodeID,
			&task.AssigneeID,
			&task.Status,
			&task.Action,
			&task.Comment,
			&task.ApprovedAt,
			&task.CreatedAt,
			&task.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		tasks = append(tasks, &task)
	}

	return tasks, rows.Err()
}

func (r *approvalTaskRepo) ListPendingByAssignee(ctx context.Context, assigneeID uuid.UUID) ([]*model.ApprovalTask, error) {
	sql := `
		SELECT id, tenant_id, process_instance_id, node_id, assignee_id,
		       status, action, comment, approved_at, created_at, updated_at
		FROM approval_tasks
		WHERE assignee_id = $1 AND status = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, sql, assigneeID, model.TaskStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*model.ApprovalTask
	for rows.Next() {
		var task model.ApprovalTask
		err := rows.Scan(
			&task.ID,
			&task.TenantID,
			&task.ProcessInstanceID,
			&task.NodeID,
			&task.AssigneeID,
			&task.Status,
			&task.Action,
			&task.Comment,
			&task.ApprovedAt,
			&task.CreatedAt,
			&task.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		tasks = append(tasks, &task)
	}

	return tasks, rows.Err()
}

func (r *approvalTaskRepo) CountPendingByAssignee(ctx context.Context, assigneeID uuid.UUID) (int, error) {
	sql := `
		SELECT COUNT(*)
		FROM approval_tasks
		WHERE assignee_id = $1 AND status = $2
	`

	var count int
	err := r.db.QueryRow(ctx, sql, assigneeID, model.TaskStatusPending).Scan(&count)
	return count, err
}

func (r *approvalTaskRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.TaskStatus, action *model.ApprovalAction, comment *string, approvedAt *time.Time) error {
	sql := `
		UPDATE approval_tasks
		SET status = $1, action = $2, comment = $3, approved_at = $4, updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.Exec(ctx, sql, status, action, comment, approvedAt, time.Now(), id)
	return err
}

// ListByAssigneeWithCursor 游标分页查询审批人的任务（高性能，适用于大数据量）
// 使用 created_at 作为游标字段，配合 id 保证排序稳定性
func (r *approvalTaskRepo) ListByAssigneeWithCursor(
	ctx context.Context,
	assigneeID uuid.UUID,
	status *model.TaskStatus,
	cursor *time.Time,
	limit int,
) ([]*model.ApprovalTask, *time.Time, bool, error) {
	// 构建 WHERE 条件
	where := "assignee_id = $1"
	args := []interface{}{assigneeID}
	argIdx := 1

	// 添加游标条件
	if cursor != nil {
		argIdx++
		where += fmt.Sprintf(" AND created_at < $%d", argIdx)
		args = append(args, *cursor)
	}

	// 添加状态过滤
	if status != nil {
		argIdx++
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *status)
	}

	// 构建查询（多查1条用于判断是否有下一页）
	argIdx++
	sql := fmt.Sprintf(`
		SELECT id, tenant_id, process_instance_id, node_id, assignee_id,
		       status, action, comment, approved_at, created_at, updated_at
		FROM approval_tasks
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
	var tasks []*model.ApprovalTask
	for rows.Next() {
		var task model.ApprovalTask
		err := rows.Scan(
			&task.ID,
			&task.TenantID,
			&task.ProcessInstanceID,
			&task.NodeID,
			&task.AssigneeID,
			&task.Status,
			&task.Action,
			&task.Comment,
			&task.ApprovedAt,
			&task.CreatedAt,
			&task.UpdatedAt,
		)

		if err != nil {
			return nil, nil, false, err
		}

		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, false, err
	}

	// 判断是否有下一页
	hasNext := len(tasks) > limit
	if hasNext {
		tasks = tasks[:limit]
	}

	// 生成下一页游标
	var nextCursor *time.Time
	if hasNext && len(tasks) > 0 {
		lastTask := tasks[len(tasks)-1]
		nextCursor = &lastTask.CreatedAt
	}

	return tasks, nextCursor, hasNext, nil
}
