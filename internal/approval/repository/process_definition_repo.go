package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/approval/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// ProcessDefinitionRepository 流程定义仓储接口
type ProcessDefinitionRepository interface {
	Create(ctx context.Context, def *model.ProcessDefinition) error
	Update(ctx context.Context, def *model.ProcessDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.ProcessDefinition, error)
	FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.ProcessDefinition, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*model.ProcessDefinition, error)
	ListEnabled(ctx context.Context, tenantID uuid.UUID) ([]*model.ProcessDefinition, error)
}

type processDefinitionRepo struct {
	db *database.DB
}

// NewProcessDefinitionRepository 创建流程定义仓储
func NewProcessDefinitionRepository(db *database.DB) ProcessDefinitionRepository {
	return &processDefinitionRepo{db: db}
}

func (r *processDefinitionRepo) Create(ctx context.Context, def *model.ProcessDefinition) error {
	sql := `
		INSERT INTO approval_process_definitions (
			id, tenant_id, code, name, category, form_id, workflow_id, enabled,
			created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(ctx, sql,
		def.ID,
		def.TenantID,
		def.Code,
		def.Name,
		def.Category,
		def.FormID,
		def.WorkflowID,
		def.Enabled,
		def.CreatedBy,
		def.CreatedAt,
		def.UpdatedAt,
	)

	return err
}

func (r *processDefinitionRepo) Update(ctx context.Context, def *model.ProcessDefinition) error {
	sql := `
		UPDATE approval_process_definitions
		SET name = $1, form_id = $2, workflow_id = $3, enabled = $4,
		    updated_by = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql,
		def.Name,
		def.FormID,
		def.WorkflowID,
		def.Enabled,
		def.UpdatedBy,
		def.UpdatedAt,
		def.ID,
	)

	return err
}

func (r *processDefinitionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `
		UPDATE approval_process_definitions
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql, time.Now(), id)
	return err
}

func (r *processDefinitionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ProcessDefinition, error) {
	sql := `
		SELECT id, tenant_id, code, name, category, form_id, workflow_id, enabled,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM approval_process_definitions
		WHERE id = $1 AND deleted_at IS NULL
	`

	var def model.ProcessDefinition
	err := r.db.QueryRow(ctx, sql, id).Scan(
		&def.ID,
		&def.TenantID,
		&def.Code,
		&def.Name,
		&def.Category,
		&def.FormID,
		&def.WorkflowID,
		&def.Enabled,
		&def.CreatedBy,
		&def.UpdatedBy,
		&def.CreatedAt,
		&def.UpdatedAt,
		&def.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &def, nil
}

func (r *processDefinitionRepo) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.ProcessDefinition, error) {
	sql := `
		SELECT id, tenant_id, code, name, category, form_id, workflow_id, enabled,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM approval_process_definitions
		WHERE tenant_id = $1 AND code = $2 AND deleted_at IS NULL
	`

	var def model.ProcessDefinition
	err := r.db.QueryRow(ctx, sql, tenantID, code).Scan(
		&def.ID,
		&def.TenantID,
		&def.Code,
		&def.Name,
		&def.Category,
		&def.FormID,
		&def.WorkflowID,
		&def.Enabled,
		&def.CreatedBy,
		&def.UpdatedBy,
		&def.CreatedAt,
		&def.UpdatedAt,
		&def.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &def, nil
}

func (r *processDefinitionRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*model.ProcessDefinition, error) {
	sql := `
		SELECT id, tenant_id, code, name, category, form_id, workflow_id, enabled,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM approval_process_definitions
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var defs []*model.ProcessDefinition
	for rows.Next() {
		var def model.ProcessDefinition
		err := rows.Scan(
			&def.ID,
			&def.TenantID,
			&def.Code,
			&def.Name,
			&def.Category,
			&def.FormID,
			&def.WorkflowID,
			&def.Enabled,
			&def.CreatedBy,
			&def.UpdatedBy,
			&def.CreatedAt,
			&def.UpdatedAt,
			&def.DeletedAt,
		)

		if err != nil {
			return nil, err
		}

		defs = append(defs, &def)
	}

	return defs, rows.Err()
}

func (r *processDefinitionRepo) ListEnabled(ctx context.Context, tenantID uuid.UUID) ([]*model.ProcessDefinition, error) {
	sql := `
		SELECT id, tenant_id, code, name, category, form_id, workflow_id, enabled,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM approval_process_definitions
		WHERE tenant_id = $1 AND enabled = true AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var defs []*model.ProcessDefinition
	for rows.Next() {
		var def model.ProcessDefinition
		err := rows.Scan(
			&def.ID,
			&def.TenantID,
			&def.Code,
			&def.Name,
			&def.Category,
			&def.FormID,
			&def.WorkflowID,
			&def.Enabled,
			&def.CreatedBy,
			&def.UpdatedBy,
			&def.CreatedAt,
			&def.UpdatedAt,
			&def.DeletedAt,
		)

		if err != nil {
			return nil, err
		}

		defs = append(defs, &def)
	}

	return defs, rows.Err()
}
