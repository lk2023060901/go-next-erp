package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/form/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// FormDataRepository 表单数据仓储接口
type FormDataRepository interface {
	Create(ctx context.Context, data *model.FormData) error
	Update(ctx context.Context, data *model.FormData) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.FormData, error)
	FindByRelated(ctx context.Context, relatedType string, relatedID uuid.UUID) (*model.FormData, error)
	ListByForm(ctx context.Context, formID uuid.UUID) ([]*model.FormData, error)
	ListBySubmitter(ctx context.Context, submitterID uuid.UUID) ([]*model.FormData, error)
}

type formDataRepo struct {
	db *database.DB
}

// NewFormDataRepository 创建表单数据仓储
func NewFormDataRepository(db *database.DB) FormDataRepository {
	return &formDataRepo{db: db}
}

func (r *formDataRepo) Create(ctx context.Context, data *model.FormData) error {
	dataJSON, err := json.Marshal(data.Data)
	if err != nil {
		return err
	}

	sql := `
		INSERT INTO form_data (
			id, tenant_id, form_id, data, submitted_by, submitted_at,
			related_type, related_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.Exec(ctx, sql,
		data.ID,
		data.TenantID,
		data.FormID,
		dataJSON,
		data.SubmittedBy,
		data.SubmittedAt,
		data.RelatedType,
		data.RelatedID,
		data.CreatedAt,
		data.UpdatedAt,
	)

	return err
}

func (r *formDataRepo) Update(ctx context.Context, data *model.FormData) error {
	dataJSON, err := json.Marshal(data.Data)
	if err != nil {
		return err
	}

	sql := `
		UPDATE form_data
		SET data = $1, updated_at = $2
		WHERE id = $3
	`

	_, err = r.db.Exec(ctx, sql,
		dataJSON,
		data.UpdatedAt,
		data.ID,
	)

	return err
}

func (r *formDataRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.FormData, error) {
	sql := `
		SELECT id, tenant_id, form_id, data, submitted_by, submitted_at,
		       related_type, related_id, created_at, updated_at
		FROM form_data
		WHERE id = $1
	`

	var formData model.FormData
	var dataJSON []byte

	err := r.db.QueryRow(ctx, sql, id).Scan(
		&formData.ID,
		&formData.TenantID,
		&formData.FormID,
		&dataJSON,
		&formData.SubmittedBy,
		&formData.SubmittedAt,
		&formData.RelatedType,
		&formData.RelatedID,
		&formData.CreatedAt,
		&formData.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(dataJSON, &formData.Data); err != nil {
		return nil, err
	}

	return &formData, nil
}

func (r *formDataRepo) FindByRelated(ctx context.Context, relatedType string, relatedID uuid.UUID) (*model.FormData, error) {
	sql := `
		SELECT id, tenant_id, form_id, data, submitted_by, submitted_at,
		       related_type, related_id, created_at, updated_at
		FROM form_data
		WHERE related_type = $1 AND related_id = $2
	`

	var formData model.FormData
	var dataJSON []byte

	err := r.db.QueryRow(ctx, sql, relatedType, relatedID).Scan(
		&formData.ID,
		&formData.TenantID,
		&formData.FormID,
		&dataJSON,
		&formData.SubmittedBy,
		&formData.SubmittedAt,
		&formData.RelatedType,
		&formData.RelatedID,
		&formData.CreatedAt,
		&formData.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(dataJSON, &formData.Data); err != nil {
		return nil, err
	}

	return &formData, nil
}

func (r *formDataRepo) ListByForm(ctx context.Context, formID uuid.UUID) ([]*model.FormData, error) {
	sql := `
		SELECT id, tenant_id, form_id, data, submitted_by, submitted_at,
		       related_type, related_id, created_at, updated_at
		FROM form_data
		WHERE form_id = $1
		ORDER BY submitted_at DESC
	`

	rows, err := r.db.Query(ctx, sql, formID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataList []*model.FormData
	for rows.Next() {
		var formData model.FormData
		var dataJSON []byte

		err := rows.Scan(
			&formData.ID,
			&formData.TenantID,
			&formData.FormID,
			&dataJSON,
			&formData.SubmittedBy,
			&formData.SubmittedAt,
			&formData.RelatedType,
			&formData.RelatedID,
			&formData.CreatedAt,
			&formData.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(dataJSON, &formData.Data); err != nil {
			return nil, err
		}

		dataList = append(dataList, &formData)
	}

	return dataList, rows.Err()
}

func (r *formDataRepo) ListBySubmitter(ctx context.Context, submitterID uuid.UUID) ([]*model.FormData, error) {
	sql := `
		SELECT id, tenant_id, form_id, data, submitted_by, submitted_at,
		       related_type, related_id, created_at, updated_at
		FROM form_data
		WHERE submitted_by = $1
		ORDER BY submitted_at DESC
	`

	rows, err := r.db.Query(ctx, sql, submitterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataList []*model.FormData
	for rows.Next() {
		var formData model.FormData
		var dataJSON []byte

		err := rows.Scan(
			&formData.ID,
			&formData.TenantID,
			&formData.FormID,
			&dataJSON,
			&formData.SubmittedBy,
			&formData.SubmittedAt,
			&formData.RelatedType,
			&formData.RelatedID,
			&formData.CreatedAt,
			&formData.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(dataJSON, &formData.Data); err != nil {
			return nil, err
		}

		dataList = append(dataList, &formData)
	}

	return dataList, rows.Err()
}
