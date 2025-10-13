package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/approval/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// ProcessHistoryRepository 流程历史仓储接口
type ProcessHistoryRepository interface {
	Create(ctx context.Context, history *model.ProcessHistory) error
	ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]*model.ProcessHistory, error)
	ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]*model.ProcessHistory, error)
}

type processHistoryRepo struct {
	db *database.DB
}

// NewProcessHistoryRepository 创建流程历史仓储
func NewProcessHistoryRepository(db *database.DB) ProcessHistoryRepository {
	return &processHistoryRepo{db: db}
}

func (r *processHistoryRepo) Create(ctx context.Context, history *model.ProcessHistory) error {
	sql := `
		INSERT INTO approval_process_histories (
			id, tenant_id, process_instance_id, task_id, node_id, node_name,
			operator_id, operator_name, action, comment, from_status, to_status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.Exec(ctx, sql,
		history.ID,
		history.TenantID,
		history.ProcessInstanceID,
		history.TaskID,
		history.NodeID,
		history.NodeName,
		history.OperatorID,
		history.OperatorName,
		history.Action,
		history.Comment,
		history.FromStatus,
		history.ToStatus,
		history.CreatedAt,
	)

	return err
}

func (r *processHistoryRepo) ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]*model.ProcessHistory, error) {
	sql := `
		SELECT id, tenant_id, process_instance_id, task_id, node_id,
		       action, operator_id, comment, created_at
		FROM approval_process_histories
		WHERE process_instance_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, sql, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []*model.ProcessHistory
	for rows.Next() {
		var history model.ProcessHistory
		err := rows.Scan(
			&history.ID,
			&history.TenantID,
			&history.ProcessInstanceID,
			&history.TaskID,
			&history.NodeID,
			&history.Action,
			&history.OperatorID,
			&history.Comment,
			&history.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		histories = append(histories, &history)
	}

	return histories, rows.Err()
}

func (r *processHistoryRepo) ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]*model.ProcessHistory, error) {
	sql := `
		SELECT id, tenant_id, process_instance_id, task_id, node_id,
		       action, operator_id, comment, created_at
		FROM approval_process_histories
		WHERE task_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, sql, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []*model.ProcessHistory
	for rows.Next() {
		var history model.ProcessHistory
		err := rows.Scan(
			&history.ID,
			&history.TenantID,
			&history.ProcessInstanceID,
			&history.TaskID,
			&history.NodeID,
			&history.Action,
			&history.OperatorID,
			&history.Comment,
			&history.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		histories = append(histories, &history)
	}

	return histories, rows.Err()
}
