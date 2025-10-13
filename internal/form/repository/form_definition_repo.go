package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/form/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// FormDefinitionRepository 表单定义仓储接口
type FormDefinitionRepository interface {
	Create(ctx context.Context, form *model.FormDefinition) error
	Update(ctx context.Context, form *model.FormDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.FormDefinition, error)
	FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.FormDefinition, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]*model.FormDefinition, error)
	ListEnabled(ctx context.Context, tenantID uuid.UUID) ([]*model.FormDefinition, error)
}

type formDefinitionRepo struct {
	db *database.DB
}

// NewFormDefinitionRepository 创建表单定义仓储
func NewFormDefinitionRepository(db *database.DB) FormDefinitionRepository {
	return &formDefinitionRepo{db: db}
}

func (r *formDefinitionRepo) Create(ctx context.Context, form *model.FormDefinition) error {
	fieldsJSON, err := json.Marshal(form.Fields)
	if err != nil {
		return err
	}

	sql := `
		INSERT INTO form_definitions (
			id, tenant_id, code, name, fields, enabled,
			created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.Exec(ctx, sql,
		form.ID,
		form.TenantID,
		form.Code,
		form.Name,
		fieldsJSON,
		form.Enabled,
		form.CreatedBy,
		form.CreatedAt,
		form.UpdatedAt,
	)

	return err
}

func (r *formDefinitionRepo) Update(ctx context.Context, form *model.FormDefinition) error {
	fieldsJSON, err := json.Marshal(form.Fields)
	if err != nil {
		return err
	}

	sql := `
		UPDATE form_definitions
		SET name = $1, fields = $2, enabled = $3,
		    updated_by = $4, updated_at = $5
		WHERE id = $6 AND deleted_at IS NULL
	`

	_, err = r.db.Exec(ctx, sql,
		form.Name,
		fieldsJSON,
		form.Enabled,
		form.UpdatedBy,
		form.UpdatedAt,
		form.ID,
	)

	return err
}

func (r *formDefinitionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `
		UPDATE form_definitions
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql, time.Now(), id)
	return err
}

func (r *formDefinitionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.FormDefinition, error) {
	sql := `
		SELECT id, tenant_id, code, name, fields, enabled,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM form_definitions
		WHERE id = $1 AND deleted_at IS NULL
	`

	var form model.FormDefinition
	var fieldsJSON []byte

	err := r.db.QueryRow(ctx, sql, id).Scan(
		&form.ID,
		&form.TenantID,
		&form.Code,
		&form.Name,
		&fieldsJSON,
		&form.Enabled,
		&form.CreatedBy,
		&form.UpdatedBy,
		&form.CreatedAt,
		&form.UpdatedAt,
		&form.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(fieldsJSON, &form.Fields); err != nil {
		return nil, err
	}

	return &form, nil
}

func (r *formDefinitionRepo) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.FormDefinition, error) {
	sql := `
		SELECT id, tenant_id, code, name, fields, enabled,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM form_definitions
		WHERE tenant_id = $1 AND code = $2 AND deleted_at IS NULL
	`

	var form model.FormDefinition
	var fieldsJSON []byte

	err := r.db.QueryRow(ctx, sql, tenantID, code).Scan(
		&form.ID,
		&form.TenantID,
		&form.Code,
		&form.Name,
		&fieldsJSON,
		&form.Enabled,
		&form.CreatedBy,
		&form.UpdatedBy,
		&form.CreatedAt,
		&form.UpdatedAt,
		&form.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(fieldsJSON, &form.Fields); err != nil {
		return nil, err
	}

	return &form, nil
}

func (r *formDefinitionRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*model.FormDefinition, error) {
	sql := `
		SELECT id, tenant_id, code, name, fields, enabled,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM form_definitions
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var forms []*model.FormDefinition
	for rows.Next() {
		var form model.FormDefinition
		var fieldsJSON []byte

		err := rows.Scan(
			&form.ID,
			&form.TenantID,
			&form.Code,
			&form.Name,
			&fieldsJSON,
			&form.Enabled,
			&form.CreatedBy,
			&form.UpdatedBy,
			&form.CreatedAt,
			&form.UpdatedAt,
			&form.DeletedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(fieldsJSON, &form.Fields); err != nil {
			return nil, err
		}

		forms = append(forms, &form)
	}

	return forms, rows.Err()
}

func (r *formDefinitionRepo) ListEnabled(ctx context.Context, tenantID uuid.UUID) ([]*model.FormDefinition, error) {
	sql := `
		SELECT id, tenant_id, code, name, fields, enabled,
		       created_by, updated_by, created_at, updated_at, deleted_at
		FROM form_definitions
		WHERE tenant_id = $1 AND enabled = true AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var forms []*model.FormDefinition
	for rows.Next() {
		var form model.FormDefinition
		var fieldsJSON []byte

		err := rows.Scan(
			&form.ID,
			&form.TenantID,
			&form.Code,
			&form.Name,
			&fieldsJSON,
			&form.Enabled,
			&form.CreatedBy,
			&form.UpdatedBy,
			&form.CreatedAt,
			&form.UpdatedAt,
			&form.DeletedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(fieldsJSON, &form.Fields); err != nil {
			return nil, err
		}

		forms = append(forms, &form)
	}

	return forms, rows.Err()
}
