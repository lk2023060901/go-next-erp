package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
	"github.com/lk2023060901/go-next-erp/pkg/cache"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// FileRepository 文件仓库接口
type FileRepository interface {
	// 基础 CRUD
	Create(ctx context.Context, file *model.File) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.File, error)
	FindByStorageKey(ctx context.Context, storageKey string) (*model.File, error)
	FindByChecksum(ctx context.Context, checksum string, tenantID uuid.UUID) (*model.File, error)
	Update(ctx context.Context, file *model.File) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// 列表查询
	List(ctx context.Context, filter *FileFilter) ([]*model.File, int64, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.File, error)
	ListByUploader(ctx context.Context, uploaderID uuid.UUID, limit, offset int) ([]*model.File, error)

	// 病毒扫描
	UpdateVirusScanResult(ctx context.Context, id uuid.UUID, result model.VirusScanResult) error

	// 生命周期管理
	MarkAsExpired(ctx context.Context, before time.Time) (int64, error)
	CleanTemporaryFiles(ctx context.Context, before time.Time) (int64, error)

	// 统计
	GetTotalSize(ctx context.Context, tenantID uuid.UUID) (int64, error)
	GetFileCount(ctx context.Context, tenantID uuid.UUID) (int64, error)
}

// FileFilter 文件筛选条件
type FileFilter struct {
	TenantID      uuid.UUID
	UploadedBy    *uuid.UUID
	Category      *string
	Status        *model.FileStatus
	IsTemporary   *bool
	MimeType      *string
	MinSize       *int64
	MaxSize       *int64
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Tags          []string
	SearchQuery   *string
	Page          int
	PageSize      int
	OrderBy       string
	OrderDesc     bool
}

type fileRepo struct {
	db    *database.DB
	cache *cache.Cache
}

// NewFileRepository 创建文件仓库
func NewFileRepository(db *database.DB, cache *cache.Cache) FileRepository {
	return &fileRepo{
		db:    db,
		cache: cache,
	}
}

// Create 创建文件记录
func (r *fileRepo) Create(ctx context.Context, file *model.File) error {
	// Generate UUID v7
	file.ID = uuid.Must(uuid.NewV7())
	now := time.Now()
	file.CreatedAt = now
	file.UpdatedAt = now

	// Serialize metadata
	var metadataJSON []byte
	if file.Metadata != nil {
		metadataJSON, _ = json.Marshal(file.Metadata)
	}

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO files (
				id, tenant_id, filename, storage_key, size, mime_type, content_type,
				checksum, virus_scanned, virus_scan_result, virus_scanned_at,
				extension, bucket, category, tags, metadata,
				status, is_temporary, is_public,
				version_number, parent_file_id,
				is_compressed, has_watermark, watermark_text,
				uploaded_by, access_level,
				thumbnail_key, preview_url, preview_expires_at,
				expires_at, archived_at,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7,
				$8, $9, $10, $11,
				$12, $13, $14, $15, $16,
				$17, $18, $19,
				$20, $21,
				$22, $23, $24,
				$25, $26,
				$27, $28, $29,
				$30, $31,
				$32, $33
			)
		`,
			file.ID, file.TenantID, file.Filename, file.StorageKey, file.Size, file.MimeType, file.ContentType,
			file.Checksum, file.VirusScanned, file.VirusScanResult, file.VirusScannedAt,
			file.Extension, file.Bucket, file.Category, file.Tags, metadataJSON,
			file.Status, file.IsTemporary, file.IsPublic,
			file.VersionNumber, file.ParentFileID,
			file.IsCompressed, file.HasWatermark, file.WatermarkText,
			file.UploadedBy, file.AccessLevel,
			file.ThumbnailKey, file.PreviewURL, file.PreviewExpiresAt,
			file.ExpiresAt, file.ArchivedAt,
			file.CreatedAt, file.UpdatedAt,
		)

		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		// Invalidate cache
		r.invalidateCache(file.ID, file.StorageKey)
		return nil
	})
}

// FindByID 根据 ID 查找文件
func (r *fileRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.File, error) {
	cacheKey := fmt.Sprintf("file:id:%s", id.String())

	// Try cache
	var file model.File
	if r.cache != nil {
		if err := r.cache.Get(ctx, cacheKey, &file); err == nil {
			return &file, nil
		}
	}

	// Query database
	var metadataJSON []byte
	err := r.db.QueryRow(ctx, `
		SELECT
			id, tenant_id, filename, storage_key, size, mime_type, content_type,
			checksum, virus_scanned, virus_scan_result, virus_scanned_at,
			extension, bucket, category, tags, metadata,
			status, is_temporary, is_public,
			version_number, parent_file_id,
			is_compressed, has_watermark, watermark_text,
			uploaded_by, access_level,
			thumbnail_key, preview_url, preview_expires_at,
			expires_at, archived_at,
			created_at, updated_at, deleted_at
		FROM files
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&file.ID, &file.TenantID, &file.Filename, &file.StorageKey, &file.Size, &file.MimeType, &file.ContentType,
		&file.Checksum, &file.VirusScanned, &file.VirusScanResult, &file.VirusScannedAt,
		&file.Extension, &file.Bucket, &file.Category, &file.Tags, &metadataJSON,
		&file.Status, &file.IsTemporary, &file.IsPublic,
		&file.VersionNumber, &file.ParentFileID,
		&file.IsCompressed, &file.HasWatermark, &file.WatermarkText,
		&file.UploadedBy, &file.AccessLevel,
		&file.ThumbnailKey, &file.PreviewURL, &file.PreviewExpiresAt,
		&file.ExpiresAt, &file.ArchivedAt,
		&file.CreatedAt, &file.UpdatedAt, &file.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to find file: %w", err)
	}

	// Deserialize metadata
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &file.Metadata)
	}

	// Set cache
	if r.cache != nil {
		r.cache.Set(ctx, cacheKey, &file, 300) // 5 minutes = 300 seconds
	}

	return &file, nil
}

// FindByStorageKey 根据存储键查找文件
func (r *fileRepo) FindByStorageKey(ctx context.Context, storageKey string) (*model.File, error) {
	var file model.File
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT
			id, tenant_id, filename, storage_key, size, mime_type, content_type,
			checksum, virus_scanned, virus_scan_result, virus_scanned_at,
			extension, bucket, category, tags, metadata,
			status, is_temporary, is_public,
			version_number, parent_file_id,
			is_compressed, has_watermark, watermark_text,
			uploaded_by, access_level,
			thumbnail_key, preview_url, preview_expires_at,
			expires_at, archived_at,
			created_at, updated_at, deleted_at
		FROM files
		WHERE storage_key = $1 AND deleted_at IS NULL
	`, storageKey).Scan(
		&file.ID, &file.TenantID, &file.Filename, &file.StorageKey, &file.Size, &file.MimeType, &file.ContentType,
		&file.Checksum, &file.VirusScanned, &file.VirusScanResult, &file.VirusScannedAt,
		&file.Extension, &file.Bucket, &file.Category, &file.Tags, &metadataJSON,
		&file.Status, &file.IsTemporary, &file.IsPublic,
		&file.VersionNumber, &file.ParentFileID,
		&file.IsCompressed, &file.HasWatermark, &file.WatermarkText,
		&file.UploadedBy, &file.AccessLevel,
		&file.ThumbnailKey, &file.PreviewURL, &file.PreviewExpiresAt,
		&file.ExpiresAt, &file.ArchivedAt,
		&file.CreatedAt, &file.UpdatedAt, &file.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to find file by storage key: %w", err)
	}

	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &file.Metadata)
	}

	return &file, nil
}

// FindByChecksum 根据校验和查找文件（去重用）
func (r *fileRepo) FindByChecksum(ctx context.Context, checksum string, tenantID uuid.UUID) (*model.File, error) {
	var file model.File
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT
			id, tenant_id, filename, storage_key, size, mime_type, content_type,
			checksum, virus_scanned, virus_scan_result, virus_scanned_at,
			extension, bucket, category, tags, metadata,
			status, is_temporary, is_public,
			version_number, parent_file_id,
			is_compressed, has_watermark, watermark_text,
			uploaded_by, access_level,
			thumbnail_key, preview_url, preview_expires_at,
			expires_at, archived_at,
			created_at, updated_at, deleted_at
		FROM files
		WHERE checksum = $1 AND tenant_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, checksum, tenantID).Scan(
		&file.ID, &file.TenantID, &file.Filename, &file.StorageKey, &file.Size, &file.MimeType, &file.ContentType,
		&file.Checksum, &file.VirusScanned, &file.VirusScanResult, &file.VirusScannedAt,
		&file.Extension, &file.Bucket, &file.Category, &file.Tags, &metadataJSON,
		&file.Status, &file.IsTemporary, &file.IsPublic,
		&file.VersionNumber, &file.ParentFileID,
		&file.IsCompressed, &file.HasWatermark, &file.WatermarkText,
		&file.UploadedBy, &file.AccessLevel,
		&file.ThumbnailKey, &file.PreviewURL, &file.PreviewExpiresAt,
		&file.ExpiresAt, &file.ArchivedAt,
		&file.CreatedAt, &file.UpdatedAt, &file.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found is OK for dedup
		}
		return nil, fmt.Errorf("failed to find file by checksum: %w", err)
	}

	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &file.Metadata)
	}

	return &file, nil
}

// Update 更新文件
func (r *fileRepo) Update(ctx context.Context, file *model.File) error {
	file.UpdatedAt = time.Now()

	var metadataJSON []byte
	if file.Metadata != nil {
		metadataJSON, _ = json.Marshal(file.Metadata)
	}

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		result, err := tx.Exec(ctx, `
			UPDATE files SET
				filename = $2,
				size = $3,
				mime_type = $4,
				content_type = $5,
				virus_scanned = $6,
				virus_scan_result = $7,
				virus_scanned_at = $8,
				category = $9,
				tags = $10,
				metadata = $11,
				status = $12,
				is_compressed = $13,
				has_watermark = $14,
				watermark_text = $15,
				access_level = $16,
				thumbnail_key = $17,
				preview_url = $18,
				preview_expires_at = $19,
				expires_at = $20,
				archived_at = $21,
				updated_at = $22
			WHERE id = $1 AND deleted_at IS NULL
		`,
			file.ID, file.Filename, file.Size, file.MimeType, file.ContentType,
			file.VirusScanned, file.VirusScanResult, file.VirusScannedAt,
			file.Category, file.Tags, metadataJSON,
			file.Status, file.IsCompressed, file.HasWatermark, file.WatermarkText,
			file.AccessLevel, file.ThumbnailKey, file.PreviewURL, file.PreviewExpiresAt,
			file.ExpiresAt, file.ArchivedAt, file.UpdatedAt,
		)

		if err != nil {
			return fmt.Errorf("failed to update file: %w", err)
		}

		if result.RowsAffected() == 0 {
			return fmt.Errorf("file not found or already deleted")
		}

		r.invalidateCache(file.ID, file.StorageKey)
		return nil
	})
}

// Delete 硬删除文件
func (r *fileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		// Get storage key first for cache invalidation
		var storageKey string
		err := tx.QueryRow(ctx, `SELECT storage_key FROM files WHERE id = $1`, id).Scan(&storageKey)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("file not found")
			}
			return err
		}

		result, err := tx.Exec(ctx, `DELETE FROM files WHERE id = $1`, id)
		if err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}

		if result.RowsAffected() == 0 {
			return fmt.Errorf("file not found")
		}

		r.invalidateCache(id, storageKey)
		return nil
	})
}

// SoftDelete 软删除文件
func (r *fileRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		var storageKey string
		err := tx.QueryRow(ctx, `SELECT storage_key FROM files WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&storageKey)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("file not found or already deleted")
			}
			return err
		}

		result, err := tx.Exec(ctx, `
			UPDATE files
			SET deleted_at = $2, updated_at = $2
			WHERE id = $1 AND deleted_at IS NULL
		`, id, now)

		if err != nil {
			return fmt.Errorf("failed to soft delete file: %w", err)
		}

		if result.RowsAffected() == 0 {
			return fmt.Errorf("file not found or already deleted")
		}

		r.invalidateCache(id, storageKey)
		return nil
	})
}

// List 列出文件（带分页和筛选）
func (r *fileRepo) List(ctx context.Context, filter *FileFilter) ([]*model.File, int64, error) {
	// Build WHERE clause
	where := "deleted_at IS NULL AND tenant_id = $1"
	args := []interface{}{filter.TenantID}
	argIdx := 1

	if filter.UploadedBy != nil {
		argIdx++
		where += fmt.Sprintf(" AND uploaded_by = $%d", argIdx)
		args = append(args, *filter.UploadedBy)
	}

	if filter.Category != nil {
		argIdx++
		where += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, *filter.Category)
	}

	if filter.Status != nil {
		argIdx++
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *filter.Status)
	}

	if filter.IsTemporary != nil {
		argIdx++
		where += fmt.Sprintf(" AND is_temporary = $%d", argIdx)
		args = append(args, *filter.IsTemporary)
	}

	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		argIdx++
		where += fmt.Sprintf(" AND filename ILIKE $%d", argIdx)
		args = append(args, "%"+*filter.SearchQuery+"%")
	}

	// Count total
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM files WHERE %s", where)
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count files: %w", err)
	}

	// Build ORDER BY
	orderBy := "created_at"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}
	orderDir := "DESC"
	if !filter.OrderDesc {
		orderDir = "ASC"
	}

	// Pagination
	offset := (filter.Page - 1) * filter.PageSize
	argIdx++
	limitIdx := argIdx
	argIdx++
	offsetIdx := argIdx

	query := fmt.Sprintf(`
		SELECT
			id, tenant_id, filename, storage_key, size, mime_type, content_type,
			checksum, virus_scanned, virus_scan_result, virus_scanned_at,
			extension, bucket, category, tags, metadata,
			status, is_temporary, is_public,
			version_number, parent_file_id,
			is_compressed, has_watermark, watermark_text,
			uploaded_by, access_level,
			thumbnail_key, preview_url, preview_expires_at,
			expires_at, archived_at,
			created_at, updated_at, deleted_at
		FROM files
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, where, orderBy, orderDir, limitIdx, offsetIdx)

	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list files: %w", err)
	}
	defer rows.Close()

	files := []*model.File{}
	for rows.Next() {
		file := &model.File{}
		var metadataJSON []byte

		err := rows.Scan(
			&file.ID, &file.TenantID, &file.Filename, &file.StorageKey, &file.Size, &file.MimeType, &file.ContentType,
			&file.Checksum, &file.VirusScanned, &file.VirusScanResult, &file.VirusScannedAt,
			&file.Extension, &file.Bucket, &file.Category, &file.Tags, &metadataJSON,
			&file.Status, &file.IsTemporary, &file.IsPublic,
			&file.VersionNumber, &file.ParentFileID,
			&file.IsCompressed, &file.HasWatermark, &file.WatermarkText,
			&file.UploadedBy, &file.AccessLevel,
			&file.ThumbnailKey, &file.PreviewURL, &file.PreviewExpiresAt,
			&file.ExpiresAt, &file.ArchivedAt,
			&file.CreatedAt, &file.UpdatedAt, &file.DeletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan file: %w", err)
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &file.Metadata)
		}

		files = append(files, file)
	}

	return files, total, nil
}

// ListByTenant 按租户列出文件
func (r *fileRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.File, error) {
	filter := &FileFilter{
		TenantID: tenantID,
		Page:     (offset / limit) + 1,
		PageSize: limit,
		OrderBy:  "created_at",
		OrderDesc: true,
	}

	files, _, err := r.List(ctx, filter)
	return files, err
}

// ListByUploader 按上传者列出文件
func (r *fileRepo) ListByUploader(ctx context.Context, uploaderID uuid.UUID, limit, offset int) ([]*model.File, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, tenant_id, filename, storage_key, size, mime_type, content_type,
			checksum, virus_scanned, virus_scan_result, virus_scanned_at,
			extension, bucket, category, tags, metadata,
			status, is_temporary, is_public,
			version_number, parent_file_id,
			is_compressed, has_watermark, watermark_text,
			uploaded_by, access_level,
			thumbnail_key, preview_url, preview_expires_at,
			expires_at, archived_at,
			created_at, updated_at, deleted_at
		FROM files
		WHERE uploaded_by = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, uploaderID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to list files by uploader: %w", err)
	}
	defer rows.Close()

	files := []*model.File{}
	for rows.Next() {
		file := &model.File{}
		var metadataJSON []byte

		err := rows.Scan(
			&file.ID, &file.TenantID, &file.Filename, &file.StorageKey, &file.Size, &file.MimeType, &file.ContentType,
			&file.Checksum, &file.VirusScanned, &file.VirusScanResult, &file.VirusScannedAt,
			&file.Extension, &file.Bucket, &file.Category, &file.Tags, &metadataJSON,
			&file.Status, &file.IsTemporary, &file.IsPublic,
			&file.VersionNumber, &file.ParentFileID,
			&file.IsCompressed, &file.HasWatermark, &file.WatermarkText,
			&file.UploadedBy, &file.AccessLevel,
			&file.ThumbnailKey, &file.PreviewURL, &file.PreviewExpiresAt,
			&file.ExpiresAt, &file.ArchivedAt,
			&file.CreatedAt, &file.UpdatedAt, &file.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &file.Metadata)
		}

		files = append(files, file)
	}

	return files, nil
}

// UpdateVirusScanResult 更新病毒扫描结果
func (r *fileRepo) UpdateVirusScanResult(ctx context.Context, id uuid.UUID, result model.VirusScanResult) error {
	now := time.Now()

	_, err := r.db.Exec(ctx, `
		UPDATE files
		SET virus_scanned = true,
		    virus_scan_result = $2,
		    virus_scanned_at = $3,
		    updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`, id, result, now)

	if err != nil {
		return fmt.Errorf("failed to update virus scan result: %w", err)
	}

	r.invalidateCache(id, "")
	return nil
}

// MarkAsExpired 标记过期的临时文件
func (r *fileRepo) MarkAsExpired(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.Exec(ctx, `
		UPDATE files
		SET status = 'archived', archived_at = $2, updated_at = $2
		WHERE is_temporary = true
		  AND expires_at IS NOT NULL
		  AND expires_at < $1
		  AND status = 'active'
		  AND deleted_at IS NULL
	`, before, time.Now())

	if err != nil {
		return 0, fmt.Errorf("failed to mark expired files: %w", err)
	}

	return result.RowsAffected(), nil
}

// CleanTemporaryFiles 清理过期的临时文件
func (r *fileRepo) CleanTemporaryFiles(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.Exec(ctx, `
		UPDATE files
		SET deleted_at = $2, updated_at = $2
		WHERE is_temporary = true
		  AND expires_at < $1
		  AND deleted_at IS NULL
	`, before, time.Now())

	if err != nil {
		return 0, fmt.Errorf("failed to clean temporary files: %w", err)
	}

	return result.RowsAffected(), nil
}

// GetTotalSize 获取租户总存储大小
func (r *fileRepo) GetTotalSize(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(size), 0)
		FROM files
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`, tenantID).Scan(&total)

	if err != nil {
		return 0, fmt.Errorf("failed to get total size: %w", err)
	}

	return total, nil
}

// GetFileCount 获取租户文件数量
func (r *fileRepo) GetFileCount(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM files
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`, tenantID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to get file count: %w", err)
	}

	return count, nil
}

// invalidateCache 清除缓存
func (r *fileRepo) invalidateCache(id uuid.UUID, storageKey string) {
	if r.cache == nil {
		return
	}

	ctx := context.Background()
	r.cache.Delete(ctx, fmt.Sprintf("file:id:%s", id.String()))
	if storageKey != "" {
		r.cache.Delete(ctx, fmt.Sprintf("file:key:%s", storageKey))
	}
}
