package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

type employeeSyncMappingRepo struct {
	db *database.DB
}

// NewEmployeeSyncMappingRepository 创建员工同步映射仓储
func NewEmployeeSyncMappingRepository(db *database.DB) repository.EmployeeSyncMappingRepository {
	return &employeeSyncMappingRepo{db: db}
}

func (r *employeeSyncMappingRepo) Create(ctx context.Context, mapping *model.EmployeeSyncMapping) error {
	rawDataJSON, _ := json.Marshal(mapping.RawData)

	sql := `
		INSERT INTO hrm_employee_sync_mappings (
			id, tenant_id, employee_id, platform, platform_id,
			sync_enabled, last_sync_at, sync_status, sync_error,
			raw_data, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12
		)
	`

	_, err := r.db.Exec(ctx, sql,
		mapping.ID, mapping.TenantID, mapping.EmployeeID, mapping.Platform, mapping.PlatformID,
		mapping.SyncEnabled, mapping.LastSyncAt, mapping.SyncStatus, mapping.SyncError,
		rawDataJSON, mapping.CreatedAt, mapping.UpdatedAt,
	)

	return err
}

func (r *employeeSyncMappingRepo) BatchCreate(ctx context.Context, mappings []*model.EmployeeSyncMapping) error {
	if len(mappings) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	sql := `
		INSERT INTO hrm_employee_sync_mappings (
			id, tenant_id, employee_id, platform, platform_id,
			sync_enabled, sync_status, raw_data, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	for _, mapping := range mappings {
		rawDataJSON, _ := json.Marshal(mapping.RawData)
		_, err := tx.Exec(ctx, sql,
			mapping.ID, mapping.TenantID, mapping.EmployeeID, mapping.Platform, mapping.PlatformID,
			mapping.SyncEnabled, mapping.SyncStatus, rawDataJSON,
			mapping.CreatedAt, mapping.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *employeeSyncMappingRepo) Update(ctx context.Context, mapping *model.EmployeeSyncMapping) error {
	rawDataJSON, _ := json.Marshal(mapping.RawData)

	sql := `
		UPDATE hrm_employee_sync_mappings SET
			sync_enabled = $1, last_sync_at = $2, sync_status = $3, sync_error = $4,
			raw_data = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(ctx, sql,
		mapping.SyncEnabled, mapping.LastSyncAt, mapping.SyncStatus, mapping.SyncError,
		rawDataJSON, mapping.UpdatedAt,
		mapping.ID,
	)

	return err
}

func (r *employeeSyncMappingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `DELETE FROM hrm_employee_sync_mappings WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *employeeSyncMappingRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.EmployeeSyncMapping, error) {
	sql := `
		SELECT id, tenant_id, employee_id, platform, platform_id,
		       sync_enabled, last_sync_at, sync_status, sync_error,
		       raw_data, created_at, updated_at
		FROM hrm_employee_sync_mappings
		WHERE id = $1
	`

	mapping := &model.EmployeeSyncMapping{}
	var rawDataJSON []byte

	err := r.db.QueryRow(ctx, sql, id).Scan(
		&mapping.ID, &mapping.TenantID, &mapping.EmployeeID, &mapping.Platform, &mapping.PlatformID,
		&mapping.SyncEnabled, &mapping.LastSyncAt, &mapping.SyncStatus, &mapping.SyncError,
		&rawDataJSON, &mapping.CreatedAt, &mapping.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("employee sync mapping not found")
		}
		return nil, err
	}

	if len(rawDataJSON) > 0 {
		json.Unmarshal(rawDataJSON, &mapping.RawData)
	}

	return mapping, nil
}

func (r *employeeSyncMappingRepo) FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) ([]*model.EmployeeSyncMapping, error) {
	sql := `
		SELECT id, tenant_id, employee_id, platform, platform_id,
		       sync_enabled, last_sync_at, sync_status, sync_error, created_at
		FROM hrm_employee_sync_mappings
		WHERE tenant_id = $1 AND employee_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mappings []*model.EmployeeSyncMapping
	for rows.Next() {
		mapping := &model.EmployeeSyncMapping{}
		err := rows.Scan(
			&mapping.ID, &mapping.TenantID, &mapping.EmployeeID, &mapping.Platform, &mapping.PlatformID,
			&mapping.SyncEnabled, &mapping.LastSyncAt, &mapping.SyncStatus, &mapping.SyncError, &mapping.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, mapping)
	}

	return mappings, rows.Err()
}

func (r *employeeSyncMappingRepo) FindByPlatform(ctx context.Context, tenantID uuid.UUID, platform model.PlatformType, platformID string) (*model.EmployeeSyncMapping, error) {
	sql := `
		SELECT id, tenant_id, employee_id, platform, platform_id,
		       sync_enabled, last_sync_at, sync_status, created_at
		FROM hrm_employee_sync_mappings
		WHERE tenant_id = $1 AND platform = $2 AND platform_id = $3
	`

	mapping := &model.EmployeeSyncMapping{}
	err := r.db.QueryRow(ctx, sql, tenantID, platform, platformID).Scan(
		&mapping.ID, &mapping.TenantID, &mapping.EmployeeID, &mapping.Platform, &mapping.PlatformID,
		&mapping.SyncEnabled, &mapping.LastSyncAt, &mapping.SyncStatus, &mapping.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("employee sync mapping not found")
		}
		return nil, err
	}

	return mapping, nil
}

func (r *employeeSyncMappingRepo) ListByPlatform(ctx context.Context, tenantID uuid.UUID, platform model.PlatformType) ([]*model.EmployeeSyncMapping, error) {
	sql := `
		SELECT id, tenant_id, employee_id, platform, platform_id,
		       sync_enabled, sync_status, created_at
		FROM hrm_employee_sync_mappings
		WHERE tenant_id = $1 AND platform = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, sql, tenantID, platform)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mappings []*model.EmployeeSyncMapping
	for rows.Next() {
		mapping := &model.EmployeeSyncMapping{}
		err := rows.Scan(
			&mapping.ID, &mapping.TenantID, &mapping.EmployeeID, &mapping.Platform, &mapping.PlatformID,
			&mapping.SyncEnabled, &mapping.SyncStatus, &mapping.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, mapping)
	}

	return mappings, rows.Err()
}

func (r *employeeSyncMappingRepo) ListSyncEnabled(ctx context.Context, tenantID uuid.UUID, platform model.PlatformType) ([]*model.EmployeeSyncMapping, error) {
	sql := `
		SELECT id, tenant_id, employee_id, platform, platform_id,
		       sync_enabled, last_sync_at, sync_status, created_at
		FROM hrm_employee_sync_mappings
		WHERE tenant_id = $1 AND platform = $2 AND sync_enabled = TRUE
		ORDER BY last_sync_at ASC NULLS FIRST
	`

	rows, err := r.db.Query(ctx, sql, tenantID, platform)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mappings []*model.EmployeeSyncMapping
	for rows.Next() {
		mapping := &model.EmployeeSyncMapping{}
		err := rows.Scan(
			&mapping.ID, &mapping.TenantID, &mapping.EmployeeID, &mapping.Platform, &mapping.PlatformID,
			&mapping.SyncEnabled, &mapping.LastSyncAt, &mapping.SyncStatus, &mapping.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, mapping)
	}

	return mappings, rows.Err()
}

func (r *employeeSyncMappingRepo) UpdateSyncStatus(ctx context.Context, id uuid.UUID, status, errorMsg string) error {
	sql := `
		UPDATE hrm_employee_sync_mappings 
		SET sync_status = $1, sync_error = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, sql, status, errorMsg, id)
	return err
}

func (r *employeeSyncMappingRepo) UpdateLastSyncTime(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	sql := `
		UPDATE hrm_employee_sync_mappings 
		SET last_sync_at = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, sql, now, now, id)
	return err
}

func (r *employeeSyncMappingRepo) ExistsByPlatform(ctx context.Context, tenantID uuid.UUID, platform model.PlatformType, platformID string) (bool, error) {
	sql := `
		SELECT COUNT(*) FROM hrm_employee_sync_mappings
		WHERE tenant_id = $1 AND platform = $2 AND platform_id = $3
	`
	var count int
	err := r.db.QueryRow(ctx, sql, tenantID, platform, platformID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
