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

// VersionRepository 文件版本仓库接口
type VersionRepository interface {
	Create(ctx context.Context, version *model.FileVersion) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.FileVersion, error)
	ListByFile(ctx context.Context, fileID uuid.UUID) ([]*model.FileVersion, error)
	FindByFileAndVersion(ctx context.Context, fileID uuid.UUID, versionNumber int) (*model.FileVersion, error)
	GetLatestVersion(ctx context.Context, fileID uuid.UUID) (*model.FileVersion, error)
	CountByFile(ctx context.Context, fileID uuid.UUID) (int, error)
	DeleteByFile(ctx context.Context, fileID uuid.UUID) error
}

type versionRepo struct {
	db *database.DB
}

// NewVersionRepository 创建版本仓库
func NewVersionRepository(db *database.DB) VersionRepository {
	return &versionRepo{db: db}
}

// Create 创建版本记录
func (r *versionRepo) Create(ctx context.Context, version *model.FileVersion) error {
	version.ID = uuid.Must(uuid.NewV7())
	version.CreatedAt = time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO file_versions (
				id, file_id, tenant_id,
				version_number, storage_key, size, checksum,
				filename, mime_type, comment,
				changed_by, change_type,
				created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`,
			version.ID, version.FileID, version.TenantID,
			version.VersionNumber, version.StorageKey, version.Size, version.Checksum,
			version.Filename, version.MimeType, version.Comment,
			version.ChangedBy, version.ChangeType,
			version.CreatedAt,
		)

		return err
	})
}

// FindByID 根据ID查找版本
func (r *versionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.FileVersion, error) {
	version := &model.FileVersion{}

	err := r.db.QueryRow(ctx, `
		SELECT
			id, file_id, tenant_id,
			version_number, storage_key, size, checksum,
			filename, mime_type, comment,
			changed_by, change_type,
			created_at
		FROM file_versions
		WHERE id = $1
	`, id).Scan(
		&version.ID, &version.FileID, &version.TenantID,
		&version.VersionNumber, &version.StorageKey, &version.Size, &version.Checksum,
		&version.Filename, &version.MimeType, &version.Comment,
		&version.ChangedBy, &version.ChangeType,
		&version.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("version not found")
		}
		return nil, fmt.Errorf("failed to find version: %w", err)
	}

	return version, nil
}

// ListByFile 列出文件的所有版本
func (r *versionRepo) ListByFile(ctx context.Context, fileID uuid.UUID) ([]*model.FileVersion, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, file_id, tenant_id,
			version_number, storage_key, size, checksum,
			filename, mime_type, comment,
			changed_by, change_type,
			created_at
		FROM file_versions
		WHERE file_id = $1
		ORDER BY version_number DESC
	`, fileID)

	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}
	defer rows.Close()

	versions := []*model.FileVersion{}
	for rows.Next() {
		version := &model.FileVersion{}
		err := rows.Scan(
			&version.ID, &version.FileID, &version.TenantID,
			&version.VersionNumber, &version.StorageKey, &version.Size, &version.Checksum,
			&version.Filename, &version.MimeType, &version.Comment,
			&version.ChangedBy, &version.ChangeType,
			&version.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, version)
	}

	return versions, nil
}

// FindByFileAndVersion 根据文件ID和版本号查找
func (r *versionRepo) FindByFileAndVersion(ctx context.Context, fileID uuid.UUID, versionNumber int) (*model.FileVersion, error) {
	version := &model.FileVersion{}

	err := r.db.QueryRow(ctx, `
		SELECT
			id, file_id, tenant_id,
			version_number, storage_key, size, checksum,
			filename, mime_type, comment,
			changed_by, change_type,
			created_at
		FROM file_versions
		WHERE file_id = $1 AND version_number = $2
	`, fileID, versionNumber).Scan(
		&version.ID, &version.FileID, &version.TenantID,
		&version.VersionNumber, &version.StorageKey, &version.Size, &version.Checksum,
		&version.Filename, &version.MimeType, &version.Comment,
		&version.ChangedBy, &version.ChangeType,
		&version.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("version not found")
		}
		return nil, fmt.Errorf("failed to find version: %w", err)
	}

	return version, nil
}

// GetLatestVersion 获取最新版本
func (r *versionRepo) GetLatestVersion(ctx context.Context, fileID uuid.UUID) (*model.FileVersion, error) {
	version := &model.FileVersion{}

	err := r.db.QueryRow(ctx, `
		SELECT
			id, file_id, tenant_id,
			version_number, storage_key, size, checksum,
			filename, mime_type, comment,
			changed_by, change_type,
			created_at
		FROM file_versions
		WHERE file_id = $1
		ORDER BY version_number DESC
		LIMIT 1
	`, fileID).Scan(
		&version.ID, &version.FileID, &version.TenantID,
		&version.VersionNumber, &version.StorageKey, &version.Size, &version.Checksum,
		&version.Filename, &version.MimeType, &version.Comment,
		&version.ChangedBy, &version.ChangeType,
		&version.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("no versions found")
		}
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	return version, nil
}

// CountByFile 统计文件的版本数量
func (r *versionRepo) CountByFile(ctx context.Context, fileID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM file_versions WHERE file_id = $1
	`, fileID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count versions: %w", err)
	}

	return count, nil
}

// DeleteByFile 删除文件的所有版本
func (r *versionRepo) DeleteByFile(ctx context.Context, fileID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM file_versions WHERE file_id = $1`, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete versions: %w", err)
	}
	return nil
}
