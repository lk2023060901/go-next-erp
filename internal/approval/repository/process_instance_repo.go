package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/approval/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// ProcessInstanceRepository 流程实例仓储接口
type ProcessInstanceRepository interface {
	Create(ctx context.Context, instance *model.ProcessInstance) error
	Update(ctx context.Context, instance *model.ProcessInstance) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.ProcessInstance, error)
	FindByWorkflowInstanceID(ctx context.Context, workflowInstanceID uuid.UUID) (*model.ProcessInstance, error)
	ListByApplicant(ctx context.Context, applicantID uuid.UUID, limit, offset int) ([]*model.ProcessInstance, error)
	ListByStatus(ctx context.Context, tenantID uuid.UUID, status model.ProcessStatus, limit, offset int) ([]*model.ProcessInstance, error)
	ListByProcessDef(ctx context.Context, processDefID uuid.UUID, limit, offset int) ([]*model.ProcessInstance, error)
	CountByStatus(ctx context.Context, tenantID uuid.UUID, status model.ProcessStatus) (int, error)

	// 游标分页查询（高性能，适用于大数据量）
	ListByApplicantWithCursor(ctx context.Context, applicantID uuid.UUID, cursor *time.Time, limit int) ([]*model.ProcessInstance, *time.Time, bool, error)
	ListByStatusWithCursor(ctx context.Context, tenantID uuid.UUID, status model.ProcessStatus, cursor *time.Time, limit int) ([]*model.ProcessInstance, *time.Time, bool, error)
}

type processInstanceRepo struct {
	db *database.DB
}

// NewProcessInstanceRepository 创建流程实例仓储
func NewProcessInstanceRepository(db *database.DB) ProcessInstanceRepository {
	return &processInstanceRepo{db: db}
}

func (r *processInstanceRepo) Create(ctx context.Context, instance *model.ProcessInstance) error {
	varsJSON, err := json.Marshal(instance.Variables)
	if err != nil {
		return err
	}

	sql := `
		INSERT INTO approval_process_instances (
			id, tenant_id, process_def_id, process_def_code, process_def_name,
			workflow_instance_id, form_data_id, applicant_id, applicant_name,
			title, status, current_node_id, current_node_name, variables,
			started_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err = r.db.Exec(ctx, sql,
		instance.ID,
		instance.TenantID,
		instance.ProcessDefID,
		instance.ProcessDefCode,
		instance.ProcessDefName,
		instance.WorkflowInstanceID,
		instance.FormDataID,
		instance.ApplicantID,
		instance.ApplicantName,
		instance.Title,
		instance.Status,
		instance.CurrentNodeID,
		instance.CurrentNodeName,
		varsJSON,
		instance.StartedAt,
		instance.CreatedAt,
		instance.UpdatedAt,
	)

	return err
}

func (r *processInstanceRepo) Update(ctx context.Context, instance *model.ProcessInstance) error {
	varsJSON, err := json.Marshal(instance.Variables)
	if err != nil {
		return err
	}

	sql := `
		UPDATE approval_process_instances
		SET status = $1, current_node_id = $2, variables = $3,
		    completed_at = $4, updated_at = $5
		WHERE id = $6
	`

	_, err = r.db.Exec(ctx, sql,
		instance.Status,
		instance.CurrentNodeID,
		varsJSON,
		instance.CompletedAt,
		instance.UpdatedAt,
		instance.ID,
	)

	return err
}

func (r *processInstanceRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ProcessInstance, error) {
	sql := `
		SELECT id, tenant_id, process_def_id, process_def_code, process_def_name,
		       workflow_instance_id, form_data_id, applicant_id, applicant_name,
		       title, status, current_node_id, current_node_name,
		       variables, started_at, completed_at, created_at, updated_at
		FROM approval_process_instances
		WHERE id = $1
	`

	var instance model.ProcessInstance
	var varsJSON []byte

	err := r.db.QueryRow(ctx, sql, id).Scan(
		&instance.ID,
		&instance.TenantID,
		&instance.ProcessDefID,
		&instance.ProcessDefCode,
		&instance.ProcessDefName,
		&instance.WorkflowInstanceID,
		&instance.FormDataID,
		&instance.ApplicantID,
		&instance.ApplicantName,
		&instance.Title,
		&instance.Status,
		&instance.CurrentNodeID,
		&instance.CurrentNodeName,
		&varsJSON,
		&instance.StartedAt,
		&instance.CompletedAt,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(varsJSON, &instance.Variables); err != nil {
		return nil, err
	}

	return &instance, nil
}

func (r *processInstanceRepo) FindByWorkflowInstanceID(ctx context.Context, workflowInstanceID uuid.UUID) (*model.ProcessInstance, error) {
	sql := `
		SELECT id, tenant_id, process_def_id, process_def_code, process_def_name,
		       workflow_instance_id, form_data_id, applicant_id, applicant_name,
		       title, status, current_node_id, current_node_name,
		       variables, started_at, completed_at, created_at, updated_at
		FROM approval_process_instances
		WHERE workflow_instance_id = $1
	`

	var instance model.ProcessInstance
	var varsJSON []byte

	err := r.db.QueryRow(ctx, sql, workflowInstanceID).Scan(
		&instance.ID,
		&instance.TenantID,
		&instance.ProcessDefID,
		&instance.ProcessDefCode,
		&instance.ProcessDefName,
		&instance.WorkflowInstanceID,
		&instance.FormDataID,
		&instance.ApplicantID,
		&instance.ApplicantName,
		&instance.Title,
		&instance.Status,
		&instance.CurrentNodeID,
		&instance.CurrentNodeName,
		&varsJSON,
		&instance.StartedAt,
		&instance.CompletedAt,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(varsJSON, &instance.Variables); err != nil {
		return nil, err
	}

	return &instance, nil
}

func (r *processInstanceRepo) ListByApplicant(ctx context.Context, applicantID uuid.UUID, limit, offset int) ([]*model.ProcessInstance, error) {
	sql := `
		SELECT id, tenant_id, process_def_id, process_def_code, process_def_name,
		       workflow_instance_id, form_data_id, applicant_id, applicant_name,
		       title, status, current_node_id, current_node_name,
		       variables, started_at, completed_at, created_at, updated_at
		FROM approval_process_instances
		WHERE applicant_id = $1
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, sql, applicantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*model.ProcessInstance
	for rows.Next() {
		var instance model.ProcessInstance
		var varsJSON []byte

		err := rows.Scan(
			&instance.ID,
			&instance.TenantID,
			&instance.ProcessDefID,
			&instance.ProcessDefCode,
			&instance.ProcessDefName,
			&instance.WorkflowInstanceID,
			&instance.FormDataID,
			&instance.ApplicantID,
			&instance.ApplicantName,
			&instance.Title,
			&instance.Status,
			&instance.CurrentNodeID,
			&instance.CurrentNodeName,
			&varsJSON,
			&instance.StartedAt,
			&instance.CompletedAt,
			&instance.CreatedAt,
			&instance.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(varsJSON, &instance.Variables); err != nil {
			return nil, err
		}

		instances = append(instances, &instance)
	}

	return instances, rows.Err()
}

func (r *processInstanceRepo) ListByStatus(ctx context.Context, tenantID uuid.UUID, status model.ProcessStatus, limit, offset int) ([]*model.ProcessInstance, error) {
	sql := `
		SELECT id, tenant_id, process_def_id, process_def_code, process_def_name,
		       workflow_instance_id, form_data_id, applicant_id, applicant_name,
		       title, status, current_node_id, current_node_name,
		       variables, started_at, completed_at, created_at, updated_at
		FROM approval_process_instances
		WHERE tenant_id = $1 AND status = $2
		ORDER BY started_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, sql, tenantID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*model.ProcessInstance
	for rows.Next() {
		var instance model.ProcessInstance
		var varsJSON []byte

		err := rows.Scan(
			&instance.ID,
			&instance.TenantID,
			&instance.ProcessDefID,
			&instance.ProcessDefCode,
			&instance.ProcessDefName,
			&instance.WorkflowInstanceID,
			&instance.FormDataID,
			&instance.ApplicantID,
			&instance.ApplicantName,
			&instance.Title,
			&instance.Status,
			&instance.CurrentNodeID,
			&instance.CurrentNodeName,
			&varsJSON,
			&instance.StartedAt,
			&instance.CompletedAt,
			&instance.CreatedAt,
			&instance.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(varsJSON, &instance.Variables); err != nil {
			return nil, err
		}

		instances = append(instances, &instance)
	}

	return instances, rows.Err()
}

func (r *processInstanceRepo) ListByProcessDef(ctx context.Context, processDefID uuid.UUID, limit, offset int) ([]*model.ProcessInstance, error) {
	sql := `
		SELECT id, tenant_id, process_def_id, process_def_code, process_def_name,
		       workflow_instance_id, form_data_id, applicant_id, applicant_name,
		       title, status, current_node_id, current_node_name,
		       variables, started_at, completed_at, created_at, updated_at
		FROM approval_process_instances
		WHERE process_def_id = $1
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, sql, processDefID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*model.ProcessInstance
	for rows.Next() {
		var instance model.ProcessInstance
		var varsJSON []byte

		err := rows.Scan(
			&instance.ID,
			&instance.TenantID,
			&instance.ProcessDefID,
			&instance.ProcessDefCode,
			&instance.ProcessDefName,
			&instance.WorkflowInstanceID,
			&instance.FormDataID,
			&instance.ApplicantID,
			&instance.ApplicantName,
			&instance.Title,
			&instance.Status,
			&instance.CurrentNodeID,
			&instance.CurrentNodeName,
			&varsJSON,
			&instance.StartedAt,
			&instance.CompletedAt,
			&instance.CreatedAt,
			&instance.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(varsJSON, &instance.Variables); err != nil {
			return nil, err
		}

		instances = append(instances, &instance)
	}

	return instances, rows.Err()
}

func (r *processInstanceRepo) CountByStatus(ctx context.Context, tenantID uuid.UUID, status model.ProcessStatus) (int, error) {
	sql := `
		SELECT COUNT(*)
		FROM approval_process_instances
		WHERE tenant_id = $1 AND status = $2
	`

	var count int
	err := r.db.QueryRow(ctx, sql, tenantID, status).Scan(&count)
	return count, err
}

// ListByApplicantWithCursor 游标分页查询申请人的流程实例（高性能）
func (r *processInstanceRepo) ListByApplicantWithCursor(
	ctx context.Context,
	applicantID uuid.UUID,
	cursor *time.Time,
	limit int,
) ([]*model.ProcessInstance, *time.Time, bool, error) {
	// 构建 WHERE 条件
	where := "applicant_id = $1"
	args := []interface{}{applicantID}
	argIdx := 1

	// 添加游标条件
	if cursor != nil {
		argIdx++
		where += fmt.Sprintf(" AND started_at < $%d", argIdx)
		args = append(args, *cursor)
	}

	// 构建查询（多查1条用于判断是否有下一页）
	argIdx++
	sql := fmt.Sprintf(`
		SELECT id, tenant_id, process_def_id, process_def_code, process_def_name,
		       workflow_instance_id, form_data_id, applicant_id, applicant_name,
		       title, status, current_node_id, current_node_name,
		       variables, started_at, completed_at, created_at, updated_at
		FROM approval_process_instances
		WHERE %s
		ORDER BY started_at DESC, id DESC
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
	var instances []*model.ProcessInstance
	for rows.Next() {
		var instance model.ProcessInstance
		var varsJSON []byte

		err := rows.Scan(
			&instance.ID,
			&instance.TenantID,
			&instance.ProcessDefID,
			&instance.ProcessDefCode,
			&instance.ProcessDefName,
			&instance.WorkflowInstanceID,
			&instance.FormDataID,
			&instance.ApplicantID,
			&instance.ApplicantName,
			&instance.Title,
			&instance.Status,
			&instance.CurrentNodeID,
			&instance.CurrentNodeName,
			&varsJSON,
			&instance.StartedAt,
			&instance.CompletedAt,
			&instance.CreatedAt,
			&instance.UpdatedAt,
		)

		if err != nil {
			return nil, nil, false, err
		}

		if err := json.Unmarshal(varsJSON, &instance.Variables); err != nil {
			return nil, nil, false, err
		}

		instances = append(instances, &instance)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, false, err
	}

	// 判断是否有下一页
	hasNext := len(instances) > limit
	if hasNext {
		instances = instances[:limit]
	}

	// 生成下一页游标
	var nextCursor *time.Time
	if hasNext && len(instances) > 0 {
		lastInstance := instances[len(instances)-1]
		nextCursor = &lastInstance.StartedAt
	}

	return instances, nextCursor, hasNext, nil
}

// ListByStatusWithCursor 游标分页查询指定状态的流程实例（高性能）
func (r *processInstanceRepo) ListByStatusWithCursor(
	ctx context.Context,
	tenantID uuid.UUID,
	status model.ProcessStatus,
	cursor *time.Time,
	limit int,
) ([]*model.ProcessInstance, *time.Time, bool, error) {
	// 构建 WHERE 条件
	where := "tenant_id = $1 AND status = $2"
	args := []interface{}{tenantID, status}
	argIdx := 2

	// 添加游标条件
	if cursor != nil {
		argIdx++
		where += fmt.Sprintf(" AND started_at < $%d", argIdx)
		args = append(args, *cursor)
	}

	// 构建查询（多查1条用于判断是否有下一页）
	argIdx++
	sql := fmt.Sprintf(`
		SELECT id, tenant_id, process_def_id, process_def_code, process_def_name,
		       workflow_instance_id, form_data_id, applicant_id, applicant_name,
		       title, status, current_node_id, current_node_name,
		       variables, started_at, completed_at, created_at, updated_at
		FROM approval_process_instances
		WHERE %s
		ORDER BY started_at DESC, id DESC
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
	var instances []*model.ProcessInstance
	for rows.Next() {
		var instance model.ProcessInstance
		var varsJSON []byte

		err := rows.Scan(
			&instance.ID,
			&instance.TenantID,
			&instance.ProcessDefID,
			&instance.ProcessDefCode,
			&instance.ProcessDefName,
			&instance.WorkflowInstanceID,
			&instance.FormDataID,
			&instance.ApplicantID,
			&instance.ApplicantName,
			&instance.Title,
			&instance.Status,
			&instance.CurrentNodeID,
			&instance.CurrentNodeName,
			&varsJSON,
			&instance.StartedAt,
			&instance.CompletedAt,
			&instance.CreatedAt,
			&instance.UpdatedAt,
		)

		if err != nil {
			return nil, nil, false, err
		}

		if err := json.Unmarshal(varsJSON, &instance.Variables); err != nil {
			return nil, nil, false, err
		}

		instances = append(instances, &instance)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, false, err
	}

	// 判断是否有下一页
	hasNext := len(instances) > limit
	if hasNext {
		instances = instances[:limit]
	}

	// 生成下一页游标
	var nextCursor *time.Time
	if hasNext && len(instances) > 0 {
		lastInstance := instances[len(instances)-1]
		nextCursor = &lastInstance.StartedAt
	}

	return instances, nextCursor, hasNext, nil
}
