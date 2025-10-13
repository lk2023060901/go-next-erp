package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// RelationRepository 文件关联仓库接口
type RelationRepository interface {
	Create(ctx context.Context, relation *model.FileRelation) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.FileRelation, error)
	ListByFile(ctx context.Context, fileID uuid.UUID) ([]*model.FileRelation, error)
	ListByEntity(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) ([]*model.FileRelation, error)
	FindByFileAndEntity(ctx context.Context, fileID uuid.UUID, entityType model.EntityType, entityID uuid.UUID) (*model.FileRelation, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByFile(ctx context.Context, fileID uuid.UUID) error
	DeleteByEntity(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) error
	CountByFile(ctx context.Context, fileID uuid.UUID) (int, error)
	CountByEntity(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) (int, error)
}

type relationRepo struct {
	db *database.DB
}

// NewRelationRepository 创建关联仓库
func NewRelationRepository(db *database.DB) RelationRepository {
	return &relationRepo{db: db}
}

// Create 创建文件关联
func (r *relationRepo) Create(ctx context.Context, relation *model.FileRelation) error {
	relation.ID = uuid.Must(uuid.NewV7())
	relation.CreatedAt = time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO file_relations (
				id, file_id, tenant_id,
				entity_type, entity_id,
				field_name, relation_type, description, sort_order,
				created_by, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`,
			relation.ID, relation.FileID, relation.TenantID,
			relation.EntityType, relation.EntityID,
			relation.FieldName, relation.RelationType, relation.Description, relation.SortOrder,
			relation.CreatedBy, relation.CreatedAt,
		)

		return err
	})
}

// FindByID 根据ID查找关联
func (r *relationRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.FileRelation, error) {
	relation := &model.FileRelation{}

	err := r.db.QueryRow(ctx, `
		SELECT
			id, file_id, tenant_id,
			entity_type, entity_id,
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("relation not found")
		}
		return nil, fmt.Errorf("failed to find relation: %w", err)
	}

	return relation, nil
}

// ListByFile 列出文件的所有关联
func (r *relationRepo) ListByFile(ctx context.Context, fileID uuid.UUID) ([]*model.FileRelation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, file_id, tenant_id,
			entity_type, entity_id,
			field_name, relation_type, description, sort_order,
			created_by, created_at
		FROM file_relations
		WHERE file_id = $1
		ORDER BY sort_order, created_at
	`, fileID)

	if err != nil {
		return nil, fmt.Errorf("failed to list relations by file: %w", err)
	}
	defer rows.Close()

	return r.scanRelations(rows)
}

// ListByEntity 列出实体的所有关联文件
func (r *relationRepo) ListByEntity(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) ([]*model.FileRelation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, file_id, tenant_id,
			entity_type, entity_id,
			field_name, relation_type, description, sort_order,
			created_by, created_at
		FROM file_relations
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY sort_order, created_at
	`, entityType, entityID)

	if err != nil {
		return nil, fmt.Errorf("failed to list relations by entity: %w", err)
	}
	defer rows.Close()

	return r.scanRelations(rows)
}

// FindByFileAndEntity 根据文件和实体查找关联
func (r *relationRepo) FindByFileAndEntity(ctx context.Context, fileID uuid.UUID, entityType model.EntityType, entityID uuid.UUID) (*model.FileRelation, error) {
	relation := &model.FileRelation{}

	err := r.db.QueryRow(ctx, `
		SELECT
			id, file_id, tenant_id,
			entity_type, entity_id,
			field_name, relation_type, description, sort_order,
			created_by, created_at
		FROM file_relations
		WHERE file_id = $1 AND entity_type = $2 AND entity_id = $3
		LIMIT 1
	`, fileID, entityType, entityID).Scan(
		&relation.ID, &relation.FileID, &relation.TenantID,
		&relation.EntityType, &relation.EntityID,
		&relation.FieldName, &relation.RelationType, &relation.Description, &relation.SortOrder,
		&relation.CreatedBy, &relation.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found is OK
		}
		return nil, fmt.Errorf("failed to find relation: %w", err)
	}

	return relation, nil
}

// Delete 删除关联
func (r *relationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, `DELETE FROM file_relations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete relation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("relation not found")
	}

	return nil
}

// DeleteByFile 删除文件的所有关联
func (r *relationRepo) DeleteByFile(ctx context.Context, fileID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM file_relations WHERE file_id = $1`, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete relations by file: %w", err)
	}
	return nil
}

// DeleteByEntity 删除实体的所有关联
func (r *relationRepo) DeleteByEntity(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM file_relations
		WHERE entity_type = $1 AND entity_id = $2
	`, entityType, entityID)

	if err != nil {
		return fmt.Errorf("failed to delete relations by entity: %w", err)
	}
	return nil
}

// CountByFile 统计文件的关联数量
func (r *relationRepo) CountByFile(ctx context.Context, fileID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM file_relations WHERE file_id = $1
	`, fileID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count relations by file: %w", err)
	}

	return count, nil
}

// CountByEntity 统计实体的关联文件数量
func (r *relationRepo) CountByEntity(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM file_relations
		WHERE entity_type = $1 AND entity_id = $2
	`, entityType, entityID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count relations by entity: %w", err)
	}

	return count, nil
}

// scanRelations 扫描关联记录
func (r *relationRepo) scanRelations(rows pgx.Rows) ([]*model.FileRelation, error) {
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
			return nil, fmt.Errorf("failed to scan relation: %w", err)
		}
		relations = append(relations, relation)
	}

	return relations, nil
}
