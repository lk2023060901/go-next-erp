package repository

import (
"context"
"encoding/json"
"fmt"
"time"

"github.com/google/uuid"
"github.com/lk2023060901/go-next-erp/internal/file/model"
"github.com/lk2023060901/go-next-erp/pkg/database"
)

// FileAccessLogRepository 文件访问日志仓库接口
type FileAccessLogRepository interface {
	Create(ctx context.Context, log *model.FileAccessLog) error
	FindByFileID(ctx context.Context, fileID uuid.UUID, limit int) ([]*model.FileAccessLog, error)
	FindByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*model.FileAccessLog, error)
	DeleteOldLogs(ctx context.Context, before time.Time) (int64, error)
}

type fileAccessLogRepo struct {
	db *database.DB
}

// NewFileAccessLogRepository 创建文件访问日志仓库
func NewFileAccessLogRepository(db *database.DB) FileAccessLogRepository {
	return &fileAccessLogRepo{db: db}
}

// Create 创建访问日志
func (r *fileAccessLogRepo) Create(ctx context.Context, log *model.FileAccessLog) error {
	log.ID = uuid.Must(uuid.NewV7())
	log.CreatedAt = time.Now()

	var metadataJSON []byte
	if log.Metadata != nil {
		metadataJSON, _ = json.Marshal(log.Metadata)
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO file_access_logs (
id, file_id, tenant_id, action, user_id, ip_address, user_agent,
success, error_message, metadata, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
log.ID, log.FileID, log.TenantID, log.Action, log.UserID, log.IPAddress, log.UserAgent,
log.Success, log.ErrorMessage, metadataJSON, log.CreatedAt,
)

	if err != nil {
		return fmt.Errorf("failed to create access log: %w", err)
	}

	return nil
}

// FindByFileID 查找文件的访问日志
func (r *fileAccessLogRepo) FindByFileID(ctx context.Context, fileID uuid.UUID, limit int) ([]*model.FileAccessLog, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, file_id, tenant_id, action, user_id, ip_address, user_agent,
			success, error_message, metadata, created_at
		FROM file_access_logs
		WHERE file_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, fileID, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to find access logs: %w", err)
	}
	defer rows.Close()

	logs := []*model.FileAccessLog{}
	for rows.Next() {
		log := &model.FileAccessLog{}
		var metadataJSON []byte

		err := rows.Scan(
&log.ID, &log.FileID, &log.TenantID, &log.Action, &log.UserID, &log.IPAddress, &log.UserAgent,
			&log.Success, &log.ErrorMessage, &metadataJSON, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan access log: %w", err)
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &log.Metadata)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// FindByUser 查找用户的访问日志
func (r *fileAccessLogRepo) FindByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*model.FileAccessLog, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, file_id, tenant_id, action, user_id, ip_address, user_agent,
			success, error_message, metadata, created_at
		FROM file_access_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to find access logs: %w", err)
	}
	defer rows.Close()

	logs := []*model.FileAccessLog{}
	for rows.Next() {
		log := &model.FileAccessLog{}
		var metadataJSON []byte

		err := rows.Scan(
&log.ID, &log.FileID, &log.TenantID, &log.Action, &log.UserID, &log.IPAddress, &log.UserAgent,
			&log.Success, &log.ErrorMessage, &metadataJSON, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan access log: %w", err)
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &log.Metadata)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// DeleteOldLogs 删除旧日志
func (r *fileAccessLogRepo) DeleteOldLogs(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.Exec(ctx, `
		DELETE FROM file_access_logs
		WHERE created_at < $1
	`, before)

	if err != nil {
		return 0, fmt.Errorf("failed to delete old logs: %w", err)
	}

	return result.RowsAffected(), nil
}
