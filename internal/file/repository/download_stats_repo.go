package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
	"github.com/lk2023060901/go-next-erp/pkg/cache"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// DownloadStatsRepository 下载统计仓库接口
type DownloadStatsRepository interface {
	// 记录下载
	RecordDownload(ctx context.Context, stats *model.DownloadStats) error

	// 批量记录下载（用于从 Redis 同步）
	BatchRecordDownloads(ctx context.Context, stats []*model.DownloadStats) error

	// 获取文件下载汇总
	GetFileDownloadSummary(ctx context.Context, fileID uuid.UUID) (*model.FileDownloadSummary, error)

	// 获取租户下载汇总
	GetTenantDownloadSummary(ctx context.Context, tenantID uuid.UUID, period string, start, end time.Time) (*model.TenantDownloadSummary, error)

	// 获取用户下载汇总
	GetUserDownloadSummary(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, period string, start, end time.Time) (*model.UserDownloadSummary, error)

	// 获取热门文件列表
	GetTopDownloadedFiles(ctx context.Context, tenantID uuid.UUID, limit int, start, end time.Time) ([]uuid.UUID, error)

	// 清理旧数据
	CleanOldStats(ctx context.Context, before time.Time) (int64, error)
}

type downloadStatsRepo struct {
	db    *database.DB
	cache *cache.Cache
}

// NewDownloadStatsRepository 创建下载统计仓库
func NewDownloadStatsRepository(db *database.DB, cache *cache.Cache) DownloadStatsRepository {
	return &downloadStatsRepo{
		db:    db,
		cache: cache,
	}
}

// RecordDownload 记录下载
func (r *downloadStatsRepo) RecordDownload(ctx context.Context, stats *model.DownloadStats) error {
	stats.ID = uuid.Must(uuid.NewV7())

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO download_stats (
				id, tenant_id, file_id, downloaded_by, ip_address, user_agent,
				bytes_downloaded, download_time, is_complete, is_resumed, downloaded_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`,
			stats.ID, stats.TenantID, stats.FileID, stats.DownloadedBy,
			stats.IPAddress, stats.UserAgent,
			stats.BytesDownloaded, stats.DownloadTime, stats.IsComplete, stats.IsResumed,
			stats.DownloadedAt,
		)

		if err != nil {
			return fmt.Errorf("failed to record download: %w", err)
		}

		// 更新缓存中的实时计数器
		if r.cache != nil {
			r.incrementCacheCounter(ctx, stats)
		}

		return nil
	})
}

// BatchRecordDownloads 批量记录下载
func (r *downloadStatsRepo) BatchRecordDownloads(ctx context.Context, stats []*model.DownloadStats) error {
	if len(stats) == 0 {
		return nil
	}

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}

		for _, stat := range stats {
			if stat.ID == uuid.Nil {
				stat.ID = uuid.Must(uuid.NewV7())
			}

			batch.Queue(`
				INSERT INTO download_stats (
					id, tenant_id, file_id, downloaded_by, ip_address, user_agent,
					bytes_downloaded, download_time, is_complete, is_resumed, downloaded_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			`,
				stat.ID, stat.TenantID, stat.FileID, stat.DownloadedBy,
				stat.IPAddress, stat.UserAgent,
				stat.BytesDownloaded, stat.DownloadTime, stat.IsComplete, stat.IsResumed,
				stat.DownloadedAt,
			)
		}

		br := tx.SendBatch(ctx, batch)
		defer br.Close()

		for range stats {
			if _, err := br.Exec(); err != nil {
				return fmt.Errorf("failed to batch record downloads: %w", err)
			}
		}

		return nil
	})
}

// GetFileDownloadSummary 获取文件下载汇总
func (r *downloadStatsRepo) GetFileDownloadSummary(ctx context.Context, fileID uuid.UUID) (*model.FileDownloadSummary, error) {
	summary := &model.FileDownloadSummary{
		FileID: fileID,
	}

	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) as total_downloads,
			COUNT(DISTINCT downloaded_by) as unique_downloads,
			COALESCE(SUM(bytes_downloaded), 0) as total_bytes,
			COUNT(*) FILTER (WHERE is_complete = true) as complete_downloads,
			COUNT(*) FILTER (WHERE is_resumed = true) as resumed_downloads,
			MAX(downloaded_at) as last_downloaded_at
		FROM download_stats
		WHERE file_id = $1
	`, fileID).Scan(
		&summary.TotalDownloads,
		&summary.UniqueDownloads,
		&summary.TotalBytes,
		&summary.CompleteDownloads,
		&summary.ResumedDownloads,
		&summary.LastDownloadedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get file download summary: %w", err)
	}

	return summary, nil
}

// GetTenantDownloadSummary 获取租户下载汇总
func (r *downloadStatsRepo) GetTenantDownloadSummary(ctx context.Context, tenantID uuid.UUID, period string, start, end time.Time) (*model.TenantDownloadSummary, error) {
	summary := &model.TenantDownloadSummary{
		TenantID:    tenantID,
		Period:      period,
		PeriodStart: start,
		PeriodEnd:   end,
	}

	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) as total_downloads,
			COALESCE(SUM(bytes_downloaded), 0) as total_bytes,
			COUNT(DISTINCT file_id) as total_files,
			COUNT(DISTINCT downloaded_by) as active_users
		FROM download_stats
		WHERE tenant_id = $1
		AND downloaded_at >= $2
		AND downloaded_at < $3
	`, tenantID, start, end).Scan(
		&summary.TotalDownloads,
		&summary.TotalBytes,
		&summary.TotalFiles,
		&summary.ActiveUsers,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get tenant download summary: %w", err)
	}

	return summary, nil
}

// GetUserDownloadSummary 获取用户下载汇总
func (r *downloadStatsRepo) GetUserDownloadSummary(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, period string, start, end time.Time) (*model.UserDownloadSummary, error) {
	summary := &model.UserDownloadSummary{
		TenantID:    tenantID,
		UserID:      userID,
		Period:      period,
		PeriodStart: start,
		PeriodEnd:   end,
	}

	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) as total_downloads,
			COALESCE(SUM(bytes_downloaded), 0) as total_bytes,
			COUNT(DISTINCT file_id) as total_files
		FROM download_stats
		WHERE tenant_id = $1
		AND downloaded_by = $2
		AND downloaded_at >= $3
		AND downloaded_at < $4
	`, tenantID, userID, start, end).Scan(
		&summary.TotalDownloads,
		&summary.TotalBytes,
		&summary.TotalFiles,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user download summary: %w", err)
	}

	return summary, nil
}

// GetTopDownloadedFiles 获取热门文件列表
func (r *downloadStatsRepo) GetTopDownloadedFiles(ctx context.Context, tenantID uuid.UUID, limit int, start, end time.Time) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `
		SELECT file_id, COUNT(*) as download_count
		FROM download_stats
		WHERE tenant_id = $1
		AND downloaded_at >= $2
		AND downloaded_at < $3
		GROUP BY file_id
		ORDER BY download_count DESC
		LIMIT $4
	`, tenantID, start, end, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to get top downloaded files: %w", err)
	}
	defer rows.Close()

	var fileIDs []uuid.UUID
	for rows.Next() {
		var fileID uuid.UUID
		var count int
		if err := rows.Scan(&fileID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan file ID: %w", err)
		}
		fileIDs = append(fileIDs, fileID)
	}

	return fileIDs, nil
}

// CleanOldStats 清理旧数据
func (r *downloadStatsRepo) CleanOldStats(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.db.Exec(ctx, `
		DELETE FROM download_stats
		WHERE downloaded_at < $1
	`, before)

	if err != nil {
		return 0, fmt.Errorf("failed to clean old stats: %w", err)
	}

	return result.RowsAffected(), nil
}

// incrementCacheCounter 增加缓存计数器
func (r *downloadStatsRepo) incrementCacheCounter(ctx context.Context, stats *model.DownloadStats) {
	// 文件下载次数
	fileKey := fmt.Sprintf("download:file:%s:count", stats.FileID.String())
	r.cache.Incr(ctx, fileKey)

	// 文件下载流量（简化处理，实际可用 Redis Hash 或者定时从 DB 读取）
	bytesKey := fmt.Sprintf("download:file:%s:bytes", stats.FileID.String())
	var currentBytes int64
	if err := r.cache.Get(ctx, bytesKey, &currentBytes); err == nil {
		r.cache.Set(ctx, bytesKey, currentBytes+stats.BytesDownloaded, 0)
	} else {
		r.cache.Set(ctx, bytesKey, stats.BytesDownloaded, 0)
	}

	// 租户下载次数（按天）
	today := stats.DownloadedAt.Format("2006-01-02")
	tenantDayKey := fmt.Sprintf("download:tenant:%s:day:%s:count", stats.TenantID.String(), today)
	r.cache.Incr(ctx, tenantDayKey)
	r.cache.Expire(ctx, tenantDayKey, 7*24*3600) // 保留 7 天

	// 用户下载次数（按天）
	userDayKey := fmt.Sprintf("download:user:%s:day:%s:count", stats.DownloadedBy.String(), today)
	r.cache.Incr(ctx, userDayKey)
	r.cache.Expire(ctx, userDayKey, 7*24*3600) // 保留 7 天
}
