package repository

import (
"context"
"fmt"
"time"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/lk2023060901/go-next-erp/internal/file/model"
"github.com/lk2023060901/go-next-erp/pkg/database"
)

// FileRelationRepository 文件关联仓库接口
type FileRelationRepository interface {
	Create(ctx context.Context, relation *model.FileRelation) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.FileRelation, error)
	FindByFileID(ctx context.Context, fileID uuid.UUID) ([]*model.FileRelation, error)
	FindByEntity(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) ([]*model.FileRelation, error)
	FindByEntityAndField(ctx context.Context, entityType model.EntityType, entityID uuid.UUID, fieldName string) ([]*model.FileRelation, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByFileID(ctx context.Context, fileID uuid.UUID) error
	DeleteByEntity(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) error
}

type fileRelationRepo struct {
	db *database.DB
}

// NewFileRelationRepository 创建文件关联仓库
func NewFileRelationRepository(db *database.DB) FileRelationRepository {
	return &fileRelationRepo{db: db}
}

// Create 创建文件关联
func (r *fileRelationRepo) Create(ctx context.Context, relation *model.FileRelation) error {
	relation.ID = uuid.Must(uuid.NewV7())
	relation.CreatedAt = time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
_, err := tx.Exec(ctx, `
			INSERT INTO file_relations (
id, file_id, tenant_id, entity_type, entity_id,
field_name, relation_type, description, sort_order,
created_by, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`,
relation.ID, relation.FileID, relation.TenantID,
relation.EntityType, relation.EntityID,
relation.FieldName, relation.RelationType, relation.Description, relation.SortOrder,
relation.CreatedBy, relation.CreatedAt,
)

if err != nil {
return fmt.Errorf("failed to create file relation: %w", err)
}

return nil
})
}

// FindByID 根据 ID 查找文件关联
func (r *fileRelationRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.FileRelation, error) {
	var relation model.FileRelation

	err := r.db.QueryRow(ctx, `
		SELECT
			id, file_id, tenant_id, entity_type, entity_id,
			field_name, relation_type, description, sort_order,
			created_by, created_at
		FROM file_relations
		WHERE id = $1
	`, id).Scan(
&relation.ID, &relation.FileID, &relation.TenantID,
		&relation.EntityType, &relation.EntityID,
		&relation.FieldName, &relation.RelationType, &relation.Description, &relation.SortOrder,
		&relation.CreatedBy, &relation.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("file relation not found")
		}
		return nil, fmt.Errorf("failed to find file relation: %w", err)
	}

	return &relation, nil
}

// FindByFileID 根据文件 ID 查找所有关联
func (r *fileRelationRepo) FindByFileID(ctx context.Context, fileID uuid.UUID) ([]*model.FileRelation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, file_id, tenant_id, entity_type, entity_id,
			field_name, relation_type, description, sort_order,
			created_by, created_at
		FROM file_relations
		WHERE file_id = $1
		ORDER BY sort_order, created_at
	`, fileID)

	if err != nil {
		return nil, fmt.Errorf("failed to find file relations: %w", err)
	}
	defer rows.Close()

	relations := []*model.FileRelation{}
	for rows.Next() {
		relation := &model.FileRelation{}
		err := rows.Scan(
&relation.ID, &relation.FileID, &relation.TenantID,
			&relation.EntityType, &relation.EntityID,
			&relation.FieldName, &relation.RelationType, &relation.Description, &relation.SortOrder,
			&relation.CreatedBy, &relation.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file relation: %w", err)
		}
		relations = append(relations, relation)
	}

	return relations, nil
}

// FindByEntity 根据实体查找所有关联文件
func (r *fileRelationRepo) FindByEntity(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) ([]*model.FileRelation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, file_id, tenant_id, entity_type, entity_id,
			field_name, relation_type, description, sort_order,
			created_by, created_at
		FROM file_relations
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY sort_order, created_at
	`, entityType, entityID)

	if err != nil {
		return nil, fmt.Errorf("failed to find relations by entity: %w", err)
	}
	defer rows.Close()

	relations := []*model.FileRelation{}
	for rows.Next() {
		relation := &model.FileRelation{}
		err := rows.Scan(
&relation.ID, &relation.FileID, &relation.TenantID,
			&relation.EntityType, &relation.EntityID,
			&relation.FieldName, &relation.RelationType, &relation.Description, &relation.SortOrder,
			&relation.CreatedBy, &relation.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file relation: %w", err)
		}
		relations = append(relations, relation)
	}

	return relations, nil
}

// FindByEntityAndField 根据实体和字段查找关联文件
func (r *fileRelationRepo) FindByEntityAndField(ctx context.Context, entityType model.EntityType, entityID uuid.UUID, fieldName string) ([]*model.FileRelation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, file_id, tenant_id, entity_type, entity_id,
			field_name, relation_type, description, sort_order,
			created_by, created_at
		FROM file_relations
		WHERE entity_type = $1 AND entity_id = $2 AND field_name = $3
		ORDER BY sort_order, created_at
	`, entityType, entityID, fieldName)

	if err != nil {
		return nil, fmt.Errorf("failed to find relations by entity and field: %w", err)
	}
	defer rows.Close()

	relations := []*model.FileRelation{}
	for rows.Next() {
		relation := &model.FileRelation{}
		err := rows.Scan(
&relation.ID, &relation.FileID, &relation.TenantID,
			&relation.EntityType, &relation.EntityID,
			&relation.FieldName, &relation.RelationType, &relation.Description, &relation.SortOrder,
			&relation.CreatedBy, &relation.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file relation: %w", err)
		}
		relations = append(relations, relation)
	}

	return relations, nil
}

// Delete 删除文件关联
func (r *fileRelationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
result, err := tx.Exec(ctx, `DELETE FROM file_relations WHERE id = $1`, id)
if err != nil {
return fmt.Errorf("failed to delete file relation: %w", err)
}

if result.RowsAffected() == 0 {
			return fmt.Errorf("file relation not found")
		}

		return nil
	})
}

// DeleteByFileID 删除文件的所有关联
func (r *fileRelationRepo) DeleteByFileID(ctx context.Context, fileID uuid.UUID) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
_, err := tx.Exec(ctx, `DELETE FROM file_relations WHERE file_id = $1`, fileID)
if err != nil {
return fmt.Errorf("failed to delete file relations: %w", err)
}
return nil
})
}

// DeleteByEntity 删除实体的所有文件关联
func (r *fileRelationRepo) DeleteByEntity(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
_, err := tx.Exec(ctx, `
			DELETE FROM file_relations 
			WHERE entity_type = $1 AND entity_id = $2
		`, entityType, entityID)
if err != nil {
return fmt.Errorf("failed to delete entity file relations: %w", err)
}
return nil
})
}
