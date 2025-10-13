package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// AuditLogRepository 审计日志仓储接口
type AuditLogRepository interface {
	// 基础操作
	Create(ctx context.Context, log *model.AuditLog) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error)
	FindByEventID(ctx context.Context, eventID string) (*model.AuditLog, error)

	// 查询
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.AuditLog, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.AuditLog, error)
	ListByAction(ctx context.Context, tenantID uuid.UUID, action string, limit, offset int) ([]*model.AuditLog, error)
	ListByTimeRange(ctx context.Context, tenantID uuid.UUID, start, end time.Time, limit, offset int) ([]*model.AuditLog, error)

	// 统计
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CountByAction(ctx context.Context, tenantID uuid.UUID, action string) (int64, error)

	// 清理
	CleanupOldLogs(ctx context.Context, before time.Time) error
}

type auditLogRepo struct {
	db *database.DB
}

func NewAuditLogRepository(db *database.DB) AuditLogRepository {
	return &auditLogRepo{
		db: db,
	}
}

// Create 创建审计日志
func (r *auditLogRepo) Create(ctx context.Context, log *model.AuditLog) error {
	log.ID = uuid.Must(uuid.NewV7())
	log.CreatedAt = time.Now()

	beforeJSON, _ := json.Marshal(log.BeforeData)
	afterJSON, _ := json.Marshal(log.AfterData)

	// 合并before/after data到changes字段
	changes := make(map[string]interface{})
	if len(beforeJSON) > 0 {
		changes["before"] = json.RawMessage(beforeJSON)
	}
	if len(afterJSON) > 0 {
		changes["after"] = json.RawMessage(afterJSON)
	}
	changesJSON, _ := json.Marshal(changes)

	// 审计日志直接写入主库，不使用事务（避免影响业务操作）
	_, err := r.db.Master().Exec(ctx, `
		INSERT INTO audit_logs (
			id, request_id, tenant_id, user_id, action, resource_type, resource_id,
			changes, ip_address, user_agent, result, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`,
		log.ID, log.EventID, log.TenantID, log.UserID, log.Action,
		log.Resource, log.ResourceID, changesJSON,
		log.IPAddress, log.UserAgent, log.Result, log.ErrorMsg, log.CreatedAt,
	)

	return err
}

// FindByID 根据 ID 查找审计日志
func (r *auditLogRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error) {
	var log model.AuditLog

	row := r.db.QueryRow(ctx, `
		SELECT id, request_id, tenant_id, user_id, action, resource_type, resource_id,
			   changes, ip_address, user_agent, result, error_message, created_at
		FROM audit_logs
		WHERE id = $1
	`, id)

	if err := r.scanAuditLog(row, &log); err != nil {
		return nil, err
	}

	return &log, nil
}

// FindByEventID 根据事件 ID 查找审计日志
func (r *auditLogRepo) FindByEventID(ctx context.Context, eventID string) (*model.AuditLog, error) {
	var log model.AuditLog

	row := r.db.QueryRow(ctx, `
		SELECT id, request_id, tenant_id, user_id, action, resource_type, resource_id,
			   changes, ip_address, user_agent, result, error_message, created_at
		FROM audit_logs
		WHERE request_id = $1
	`, eventID)

	if err := r.scanAuditLog(row, &log); err != nil {
		return nil, err
	}

	return &log, nil
}

// ListByUser 查询用户的审计日志
func (r *auditLogRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.AuditLog, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, request_id, tenant_id, user_id, action, resource_type, resource_id,
			   changes, ip_address, user_agent, result, error_message, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// ListByTenant 查询租户的审计日志
func (r *auditLogRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.AuditLog, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, request_id, tenant_id, user_id, action, resource_type, resource_id,
			   changes, ip_address, user_agent, result, error_message, created_at
		FROM audit_logs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tenantID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// ListByAction 查询指定动作的审计日志
func (r *auditLogRepo) ListByAction(ctx context.Context, tenantID uuid.UUID, action string, limit, offset int) ([]*model.AuditLog, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, request_id, tenant_id, user_id, action, resource_type, resource_id,
			   changes, ip_address, user_agent, result, error_message, created_at
		FROM audit_logs
		WHERE tenant_id = $1 AND action = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, tenantID, action, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// ListByTimeRange 查询时间范围内的审计日志
func (r *auditLogRepo) ListByTimeRange(ctx context.Context, tenantID uuid.UUID, start, end time.Time, limit, offset int) ([]*model.AuditLog, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, request_id, tenant_id, user_id, action, resource_type, resource_id,
			   changes, ip_address, user_agent, result, error_message, created_at
		FROM audit_logs
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`, tenantID, start, end, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanAuditLogs(rows)
}

// CountByUser 统计用户的审计日志数量
func (r *auditLogRepo) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM audit_logs WHERE user_id = $1
	`, userID).Scan(&count)

	return count, err
}

// CountByAction 统计指定动作的审计日志数量
func (r *auditLogRepo) CountByAction(ctx context.Context, tenantID uuid.UUID, action string) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1 AND action = $2
	`, tenantID, action).Scan(&count)

	return count, err
}

// CleanupOldLogs 清理旧的审计日志
func (r *auditLogRepo) CleanupOldLogs(ctx context.Context, before time.Time) error {
	_, err := r.db.Master().Exec(ctx, `
		DELETE FROM audit_logs WHERE created_at < $1
	`, before)

	return err
}

// scanAuditLog 扫描单条审计日志
func (r *auditLogRepo) scanAuditLog(row pgx.Row, log *model.AuditLog) error {
	var changesJSON []byte

	err := row.Scan(
		&log.ID, &log.EventID, &log.TenantID, &log.UserID, &log.Action,
		&log.Resource, &log.ResourceID, &changesJSON,
		&log.IPAddress, &log.UserAgent, &log.Result, &log.ErrorMsg,
		&log.CreatedAt,
	)

	if err != nil {
		return err
	}

	// 解析 changes JSON 字段
	if len(changesJSON) > 0 {
		var changes map[string]json.RawMessage
		if err := json.Unmarshal(changesJSON, &changes); err == nil {
			if before, ok := changes["before"]; ok {
				log.BeforeData = before
			}
			if after, ok := changes["after"]; ok {
				log.AfterData = after
			}
		}
	}

	return nil
}

// scanAuditLogs 扫描多条审计日志
func (r *auditLogRepo) scanAuditLogs(rows pgx.Rows) ([]*model.AuditLog, error) {
	var logs []*model.AuditLog

	for rows.Next() {
		var log model.AuditLog
		if err := r.scanAuditLog(rows, &log); err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}

	return logs, nil
}
