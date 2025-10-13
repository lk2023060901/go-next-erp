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
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// MultipartUploadRepository 分片上传仓库接口
type MultipartUploadRepository interface {
	// 基础 CRUD
	Create(ctx context.Context, upload *model.MultipartUpload) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.MultipartUpload, error)
	FindByUploadID(ctx context.Context, uploadID string) (*model.MultipartUpload, error)
	Update(ctx context.Context, upload *model.MultipartUpload) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 分片跟踪
	AddCompletedPart(ctx context.Context, id uuid.UUID, partNumber int) error
	GetCompletedParts(ctx context.Context, id uuid.UUID) ([]int, error)

	// 状态管理
	MarkAsCompleted(ctx context.Context, id uuid.UUID) error
	MarkAsAborted(ctx context.Context, id uuid.UUID) error

	// 清理
	CleanExpiredUploads(ctx context.Context, before time.Time) (int64, error)
	ListExpiredUploads(ctx context.Context, before time.Time) ([]*model.MultipartUpload, error)

	// 查询
	ListByTenant(ctx context.Context, tenantID uuid.UUID, status *model.UploadStatus, limit, offset int) ([]*model.MultipartUpload, error)
	ListByUser(ctx context.Context, userID uuid.UUID, status *model.UploadStatus, limit, offset int) ([]*model.MultipartUpload, error)
}

type multipartUploadRepo struct {
	db *database.DB
}

// NewMultipartUploadRepository 创建分片上传仓库
func NewMultipartUploadRepository(db *database.DB) MultipartUploadRepository {
	return &multipartUploadRepo{
		db: db,
	}
}

// Create 创建分片上传记录
func (r *multipartUploadRepo) Create(ctx context.Context, upload *model.MultipartUpload) error {
	upload.ID = uuid.Must(uuid.NewV7())
	now := time.Now()
	upload.CreatedAt = now
	upload.UpdatedAt = now

	// 序列化 metadata
	var metadataJSON []byte
	if upload.Metadata != nil {
		metadataJSON, _ = json.Marshal(upload.Metadata)
	}

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO multipart_uploads (
				id, tenant_id, upload_id, filename, storage_key, total_size, part_size,
				uploaded_parts, total_parts, mime_type, metadata,
				status, created_by, expires_at, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7,
				$8, $9, $10, $11,
				$12, $13, $14, $15, $16
			)
		`,
			upload.ID, upload.TenantID, upload.UploadID, upload.Filename, upload.StorageKey,
			upload.TotalSize, upload.PartSize,
			upload.UploadedParts, upload.TotalParts, upload.MimeType, metadataJSON,
			upload.Status, upload.CreatedBy, upload.ExpiresAt, upload.CreatedAt, upload.UpdatedAt,
		)

		if err != nil {
			return fmt.Errorf("failed to create multipart upload: %w", err)
		}

		return nil
	})
}

// FindByID 根据 ID 查找分片上传
func (r *multipartUploadRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.MultipartUpload, error) {
	var upload model.MultipartUpload
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT
			id, tenant_id, upload_id, filename, storage_key, total_size, part_size,
			uploaded_parts, total_parts, mime_type, metadata,
			status, created_by, expires_at, created_at, updated_at, completed_at
		FROM multipart_uploads
		WHERE id = $1
	`, id).Scan(
		&upload.ID, &upload.TenantID, &upload.UploadID, &upload.Filename, &upload.StorageKey,
		&upload.TotalSize, &upload.PartSize,
		&upload.UploadedParts, &upload.TotalParts, &upload.MimeType, &metadataJSON,
		&upload.Status, &upload.CreatedBy, &upload.ExpiresAt, &upload.CreatedAt, &upload.UpdatedAt,
		&upload.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("multipart upload not found")
		}
		return nil, fmt.Errorf("failed to find multipart upload: %w", err)
	}

	// 反序列化 metadata
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &upload.Metadata)
	}

	return &upload, nil
}

// FindByUploadID 根据上传 ID 查找
func (r *multipartUploadRepo) FindByUploadID(ctx context.Context, uploadID string) (*model.MultipartUpload, error) {
	var upload model.MultipartUpload
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT
			id, tenant_id, upload_id, filename, storage_key, total_size, part_size,
			uploaded_parts, total_parts, mime_type, metadata,
			status, created_by, expires_at, created_at, updated_at, completed_at
		FROM multipart_uploads
		WHERE upload_id = $1
	`, uploadID).Scan(
		&upload.ID, &upload.TenantID, &upload.UploadID, &upload.Filename, &upload.StorageKey,
		&upload.TotalSize, &upload.PartSize,
		&upload.UploadedParts, &upload.TotalParts, &upload.MimeType, &metadataJSON,
		&upload.Status, &upload.CreatedBy, &upload.ExpiresAt, &upload.CreatedAt, &upload.UpdatedAt,
		&upload.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("multipart upload not found")
		}
		return nil, fmt.Errorf("failed to find multipart upload: %w", err)
	}

	// 反序列化 metadata
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &upload.Metadata)
	}

	return &upload, nil
}

// Update 更新分片上传
func (r *multipartUploadRepo) Update(ctx context.Context, upload *model.MultipartUpload) error {
	upload.UpdatedAt = time.Now()

	var metadataJSON []byte
	if upload.Metadata != nil {
		metadataJSON, _ = json.Marshal(upload.Metadata)
	}

	result, err := r.db.Exec(ctx, `
		UPDATE multipart_uploads
		SET
			uploaded_parts = $1,
			total_parts = $2,
			status = $3,
			metadata = $4,
			updated_at = $5,
			completed_at = $6
		WHERE id = $7
	`,
		upload.UploadedParts, upload.TotalParts, upload.Status, metadataJSON,
		upload.UpdatedAt, upload.CompletedAt, upload.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update multipart upload: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("multipart upload not found")
	}

	return nil
}

// Delete 删除分片上传
func (r *multipartUploadRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, `
		DELETE FROM multipart_uploads
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete multipart upload: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("multipart upload not found")
	}

	return nil
}

// AddCompletedPart 添加已完成的分片
func (r *multipartUploadRepo) AddCompletedPart(ctx context.Context, id uuid.UUID, partNumber int) error {
	// 使用 PostgreSQL array_append 函数
	result, err := r.db.Exec(ctx, `
		UPDATE multipart_uploads
		SET
			uploaded_parts = array_append(uploaded_parts, $1),
			updated_at = $2
		WHERE id = $3
		AND NOT ($1 = ANY(uploaded_parts))
	`, partNumber, time.Now(), id)

	if err != nil {
		return fmt.Errorf("failed to add completed part: %w", err)
	}

	if result.RowsAffected() == 0 {
		// 可能是分片已存在或上传不存在
		return nil
	}

	return nil
}

// GetCompletedParts 获取已完成的分片
func (r *multipartUploadRepo) GetCompletedParts(ctx context.Context, id uuid.UUID) ([]int, error) {
	var parts []int

	err := r.db.QueryRow(ctx, `
		SELECT uploaded_parts
		FROM multipart_uploads
		WHERE id = $1
	`, id).Scan(&parts)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("multipart upload not found")
		}
		return nil, fmt.Errorf("failed to get completed parts: %w", err)
	}

	if parts == nil {
		parts = []int{}
	}

	return parts, nil
}

// MarkAsCompleted 标记为已完成
func (r *multipartUploadRepo) MarkAsCompleted(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result, err := r.db.Exec(ctx, `
		UPDATE multipart_uploads
		SET
			status = $1,
			completed_at = $2,
			updated_at = $3
		WHERE id = $4
	`, model.UploadStatusCompleted, now, now, id)

	if err != nil {
		return fmt.Errorf("failed to mark as completed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("multipart upload not found")
	}

	return nil
}

// MarkAsAborted 标记为已中止
func (r *multipartUploadRepo) MarkAsAborted(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result, err := r.db.Exec(ctx, `
		UPDATE multipart_uploads
		SET
			status = $1,
			updated_at = $2
		WHERE id = $3
	`, model.UploadStatusAborted, now, id)

	if err != nil {
		return fmt.Errorf("failed to mark as aborted: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("multipart upload not found")
	}

	return nil
}

// CleanExpiredUploads 清理过期的上传
func (r *multipartUploadRepo) CleanExpiredUploads(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.Exec(ctx, `
		DELETE FROM multipart_uploads
		WHERE expires_at < $1
		AND status = $2
	`, before, model.UploadStatusInProgress)

	if err != nil {
		return 0, fmt.Errorf("failed to clean expired uploads: %w", err)
	}

	return result.RowsAffected(), nil
}

// ListExpiredUploads 列出过期的上传
func (r *multipartUploadRepo) ListExpiredUploads(ctx context.Context, before time.Time) ([]*model.MultipartUpload, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, tenant_id, upload_id, filename, storage_key, total_size, part_size,
			uploaded_parts, total_parts, mime_type, metadata,
			status, created_by, expires_at, created_at, updated_at, completed_at
		FROM multipart_uploads
		WHERE expires_at < $1
		AND status = $2
		ORDER BY expires_at ASC
	`, before, model.UploadStatusInProgress)

	if err != nil {
		return nil, fmt.Errorf("failed to list expired uploads: %w", err)
	}
	defer rows.Close()

	var uploads []*model.MultipartUpload
	for rows.Next() {
		var upload model.MultipartUpload
		var metadataJSON []byte

		err := rows.Scan(
			&upload.ID, &upload.TenantID, &upload.UploadID, &upload.Filename, &upload.StorageKey,
			&upload.TotalSize, &upload.PartSize,
			&upload.UploadedParts, &upload.TotalParts, &upload.MimeType, &metadataJSON,
			&upload.Status, &upload.CreatedBy, &upload.ExpiresAt, &upload.CreatedAt, &upload.UpdatedAt,
			&upload.CompletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan upload: %w", err)
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &upload.Metadata)
		}

		uploads = append(uploads, &upload)
	}

	return uploads, nil
}

// ListByTenant 按租户列出
func (r *multipartUploadRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, status *model.UploadStatus, limit, offset int) ([]*model.MultipartUpload, error) {
	query := `
		SELECT
			id, tenant_id, upload_id, filename, storage_key, total_size, part_size,
			uploaded_parts, total_parts, mime_type, metadata,
			status, created_by, expires_at, created_at, updated_at, completed_at
		FROM multipart_uploads
		WHERE tenant_id = $1
	`

	args := []interface{}{tenantID}
	argPos := 2

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *status)
		argPos++
	}

	query += ` ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, limit)
		argPos++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list uploads by tenant: %w", err)
	}
	defer rows.Close()

	return r.scanUploads(rows)
}

// ListByUser 按用户列出
func (r *multipartUploadRepo) ListByUser(ctx context.Context, userID uuid.UUID, status *model.UploadStatus, limit, offset int) ([]*model.MultipartUpload, error) {
	query := `
		SELECT
			id, tenant_id, upload_id, filename, storage_key, total_size, part_size,
			uploaded_parts, total_parts, mime_type, metadata,
			status, created_by, expires_at, created_at, updated_at, completed_at
		FROM multipart_uploads
		WHERE created_by = $1
	`

	args := []interface{}{userID}
	argPos := 2

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *status)
		argPos++
	}

	query += ` ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, limit)
		argPos++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list uploads by user: %w", err)
	}
	defer rows.Close()

	return r.scanUploads(rows)
}

// scanUploads 扫描上传列表
func (r *multipartUploadRepo) scanUploads(rows pgx.Rows) ([]*model.MultipartUpload, error) {
	var uploads []*model.MultipartUpload

	for rows.Next() {
		var upload model.MultipartUpload
		var metadataJSON []byte

		err := rows.Scan(
			&upload.ID, &upload.TenantID, &upload.UploadID, &upload.Filename, &upload.StorageKey,
			&upload.TotalSize, &upload.PartSize,
			&upload.UploadedParts, &upload.TotalParts, &upload.MimeType, &metadataJSON,
			&upload.Status, &upload.CreatedBy, &upload.ExpiresAt, &upload.CreatedAt, &upload.UpdatedAt,
			&upload.CompletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan upload: %w", err)
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &upload.Metadata)
		}

		uploads = append(uploads, &upload)
	}

	return uploads, nil
}
