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

// QuotaRepository 存储配额仓库接口
type QuotaRepository interface {
	Create(ctx context.Context, quota *model.StorageQuota) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.StorageQuota, error)
	FindBySubject(ctx context.Context, tenantID uuid.UUID, subjectType model.SubjectType, subjectID *uuid.UUID) (*model.StorageQuota, error)
	Update(ctx context.Context, quota *model.StorageQuota) error
	UpdateUsage(ctx context.Context, id uuid.UUID, usedDelta int64, fileCountDelta int) error
	ReserveQuota(ctx context.Context, id uuid.UUID, size int64) error
	CommitReservation(ctx context.Context, id uuid.UUID, size int64) error
	ReleaseReservation(ctx context.Context, id uuid.UUID, size int64) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.StorageQuota, error)
	GetOrCreateTenantQuota(ctx context.Context, tenantID uuid.UUID, defaultLimit int64) (*model.StorageQuota, error)
	GetOrCreateUserQuota(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, defaultLimit int64) (*model.StorageQuota, error)
}

type quotaRepo struct {
	db *database.DB
}

// NewQuotaRepository 创建配额仓库
func NewQuotaRepository(db *database.DB) QuotaRepository {
	return &quotaRepo{db: db}
}

// Create 创建配额记录
func (r *quotaRepo) Create(ctx context.Context, quota *model.StorageQuota) error {
	quota.ID = uuid.Must(uuid.NewV7())
	now := time.Now()
	quota.CreatedAt = now
	quota.UpdatedAt = now

	var settingsJSON []byte
	if quota.Settings != nil {
		settingsJSON, _ = json.Marshal(quota.Settings)
	}

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO storage_quotas (
				id, tenant_id, subject_type, subject_id,
				quota_limit, quota_used, quota_reserved,
				file_count_limit, file_count_used,
				settings, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`,
			quota.ID, quota.TenantID, quota.SubjectType, quota.SubjectID,
			quota.QuotaLimit, quota.QuotaUsed, quota.QuotaReserved,
			quota.FileCountLimit, quota.FileCountUsed,
			settingsJSON, quota.CreatedAt, quota.UpdatedAt,
		)

		return err
	})
}

// FindByID 根据ID查找配额
func (r *quotaRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.StorageQuota, error) {
	quota := &model.StorageQuota{}
	var settingsJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT
			id, tenant_id, subject_type, subject_id,
			quota_limit, quota_used, quota_reserved,
			file_count_limit, file_count_used,
			settings, created_at, updated_at
		FROM storage_quotas
		WHERE id = $1
	`, id).Scan(
		&quota.ID, &quota.TenantID, &quota.SubjectType, &quota.SubjectID,
		&quota.QuotaLimit, &quota.QuotaUsed, &quota.QuotaReserved,
		&quota.FileCountLimit, &quota.FileCountUsed,
		&settingsJSON, &quota.CreatedAt, &quota.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("quota not found")
		}
		return nil, fmt.Errorf("failed to find quota: %w", err)
	}

	if settingsJSON != nil {
		json.Unmarshal(settingsJSON, &quota.Settings)
	}

	return quota, nil
}

// FindBySubject 根据主体查找配额
func (r *quotaRepo) FindBySubject(ctx context.Context, tenantID uuid.UUID, subjectType model.SubjectType, subjectID *uuid.UUID) (*model.StorageQuota, error) {
	quota := &model.StorageQuota{}
	var settingsJSON []byte

	query := `
		SELECT
			id, tenant_id, subject_type, subject_id,
			quota_limit, quota_used, quota_reserved,
			file_count_limit, file_count_used,
			settings, created_at, updated_at
		FROM storage_quotas
		WHERE tenant_id = $1 AND subject_type = $2
	`

	var err error
	if subjectID == nil {
		query += " AND subject_id IS NULL"
		err = r.db.QueryRow(ctx, query, tenantID, subjectType).Scan(
			&quota.ID, &quota.TenantID, &quota.SubjectType, &quota.SubjectID,
			&quota.QuotaLimit, &quota.QuotaUsed, &quota.QuotaReserved,
			&quota.FileCountLimit, &quota.FileCountUsed,
			&settingsJSON, &quota.CreatedAt, &quota.UpdatedAt,
		)
	} else {
		query += " AND subject_id = $3"
		err = r.db.QueryRow(ctx, query, tenantID, subjectType, *subjectID).Scan(
			&quota.ID, &quota.TenantID, &quota.SubjectType, &quota.SubjectID,
			&quota.QuotaLimit, &quota.QuotaUsed, &quota.QuotaReserved,
			&quota.FileCountLimit, &quota.FileCountUsed,
			&settingsJSON, &quota.CreatedAt, &quota.UpdatedAt,
		)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found is OK
		}
		return nil, fmt.Errorf("failed to find quota by subject: %w", err)
	}

	if settingsJSON != nil {
		json.Unmarshal(settingsJSON, &quota.Settings)
	}

	return quota, nil
}

// Update 更新配额
func (r *quotaRepo) Update(ctx context.Context, quota *model.StorageQuota) error {
	quota.UpdatedAt = time.Now()

	var settingsJSON []byte
	if quota.Settings != nil {
		settingsJSON, _ = json.Marshal(quota.Settings)
	}

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		result, err := tx.Exec(ctx, `
			UPDATE storage_quotas SET
				quota_limit = $2,
				quota_used = $3,
				quota_reserved = $4,
				file_count_limit = $5,
				file_count_used = $6,
				settings = $7,
				updated_at = $8
			WHERE id = $1
		`,
			quota.ID,
			quota.QuotaLimit,
			quota.QuotaUsed,
			quota.QuotaReserved,
			quota.FileCountLimit,
			quota.FileCountUsed,
			settingsJSON,
			quota.UpdatedAt,
		)

		if err != nil {
			return fmt.Errorf("failed to update quota: %w", err)
		}

		if result.RowsAffected() == 0 {
			return fmt.Errorf("quota not found")
		}

		return nil
	})
}

// UpdateUsage 更新使用量（增量）
func (r *quotaRepo) UpdateUsage(ctx context.Context, id uuid.UUID, usedDelta int64, fileCountDelta int) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE storage_quotas SET
				quota_used = quota_used + $2,
				file_count_used = file_count_used + $3,
				updated_at = $4
			WHERE id = $1
		`, id, usedDelta, fileCountDelta, time.Now())

		if err != nil {
			return fmt.Errorf("failed to update usage: %w", err)
		}

		return nil
	})
}

// ReserveQuota 预留配额
func (r *quotaRepo) ReserveQuota(ctx context.Context, id uuid.UUID, size int64) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		// Check if there's enough space
		var available int64
		err := tx.QueryRow(ctx, `
			SELECT quota_limit - quota_used - quota_reserved
			FROM storage_quotas
			WHERE id = $1
			FOR UPDATE
		`, id).Scan(&available)

		if err != nil {
			return fmt.Errorf("failed to check available quota: %w", err)
		}

		if available < size {
			return fmt.Errorf("insufficient quota: available %d bytes, requested %d bytes", available, size)
		}

		// Reserve the space
		_, err = tx.Exec(ctx, `
			UPDATE storage_quotas SET
				quota_reserved = quota_reserved + $2,
				updated_at = $3
			WHERE id = $1
		`, id, size, time.Now())

		if err != nil {
			return fmt.Errorf("failed to reserve quota: %w", err)
		}

		return nil
	})
}

// CommitReservation 提交预留（转为已使用）
func (r *quotaRepo) CommitReservation(ctx context.Context, id uuid.UUID, size int64) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE storage_quotas SET
				quota_reserved = GREATEST(quota_reserved - $2, 0),
				quota_used = quota_used + $2,
				file_count_used = file_count_used + 1,
				updated_at = $3
			WHERE id = $1
		`, id, size, time.Now())

		if err != nil {
			return fmt.Errorf("failed to commit reservation: %w", err)
		}

		return nil
	})
}

// ReleaseReservation 释放预留
func (r *quotaRepo) ReleaseReservation(ctx context.Context, id uuid.UUID, size int64) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE storage_quotas SET
				quota_reserved = GREATEST(quota_reserved - $2, 0),
				updated_at = $3
			WHERE id = $1
		`, id, size, time.Now())

		if err != nil {
			return fmt.Errorf("failed to release reservation: %w", err)
		}

		return nil
	})
}

// ListByTenant 列出租户的所有配额
func (r *quotaRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.StorageQuota, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id, tenant_id, subject_type, subject_id,
			quota_limit, quota_used, quota_reserved,
			file_count_limit, file_count_used,
			settings, created_at, updated_at
		FROM storage_quotas
		WHERE tenant_id = $1
		ORDER BY subject_type, created_at
	`, tenantID)

	if err != nil {
		return nil, fmt.Errorf("failed to list quotas: %w", err)
	}
	defer rows.Close()

	quotas := []*model.StorageQuota{}
	for rows.Next() {
		quota := &model.StorageQuota{}
		var settingsJSON []byte

		err := rows.Scan(
			&quota.ID, &quota.TenantID, &quota.SubjectType, &quota.SubjectID,
			&quota.QuotaLimit, &quota.QuotaUsed, &quota.QuotaReserved,
			&quota.FileCountLimit, &quota.FileCountUsed,
			&settingsJSON, &quota.CreatedAt, &quota.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quota: %w", err)
		}

		if settingsJSON != nil {
			json.Unmarshal(settingsJSON, &quota.Settings)
		}

		quotas = append(quotas, quota)
	}

	return quotas, nil
}

// GetOrCreateTenantQuota 获取或创建租户配额
func (r *quotaRepo) GetOrCreateTenantQuota(ctx context.Context, tenantID uuid.UUID, defaultLimit int64) (*model.StorageQuota, error) {
	// Try to find existing
	quota, err := r.FindBySubject(ctx, tenantID, model.SubjectTypeTenant, nil)
	if err != nil {
		return nil, err
	}

	if quota != nil {
		return quota, nil
	}

	// Create new tenant quota
	quota = &model.StorageQuota{
		TenantID:       tenantID,
		SubjectType:    model.SubjectTypeTenant,
		SubjectID:      nil,
		QuotaLimit:     defaultLimit,
		QuotaUsed:      0,
		QuotaReserved:  0,
		FileCountLimit: nil,
		FileCountUsed:  0,
		Settings:       nil,
	}

	if err := r.Create(ctx, quota); err != nil {
		return nil, err
	}

	return quota, nil
}

// GetOrCreateUserQuota 获取或创建用户配额
func (r *quotaRepo) GetOrCreateUserQuota(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, defaultLimit int64) (*model.StorageQuota, error) {
	// Try to find existing
	quota, err := r.FindBySubject(ctx, tenantID, model.SubjectTypeUser, &userID)
	if err != nil {
		return nil, err
	}

	if quota != nil {
		return quota, nil
	}

	// Create new user quota
	quota = &model.StorageQuota{
		TenantID:       tenantID,
		SubjectType:    model.SubjectTypeUser,
		SubjectID:      &userID,
		QuotaLimit:     defaultLimit,
		QuotaUsed:      0,
		QuotaReserved:  0,
		FileCountLimit: nil,
		FileCountUsed:  0,
		Settings:       nil,
	}

	if err := r.Create(ctx, quota); err != nil {
		return nil, err
	}

	return quota, nil
}
