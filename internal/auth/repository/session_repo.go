package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/pkg/cache"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// SessionRepository 会话仓储接口
type SessionRepository interface {
	// 基础 CRUD
	Create(ctx context.Context, session *model.Session) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Session, error)
	FindByToken(ctx context.Context, token string) (*model.Session, error)
	Update(ctx context.Context, session *model.Session) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 会话管理
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*model.Session, error)
	RevokeSession(ctx context.Context, id uuid.UUID) error
	RevokeUserSessions(ctx context.Context, userID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) error
}

type sessionRepo struct {
	db    *database.DB
	cache *cache.Cache
}

func NewSessionRepository(db *database.DB, cache *cache.Cache) SessionRepository {
	return &sessionRepo{
		db:    db,
		cache: cache,
	}
}

// Create 创建会话
func (r *sessionRepo) Create(ctx context.Context, session *model.Session) error {
	session.ID = uuid.Must(uuid.NewV7())
	now := time.Now()
	session.CreatedAt = now
	session.UpdatedAt = now

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO sessions (
				id, user_id, tenant_id, token, ip_address, user_agent,
				expires_at, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
			session.ID, session.UserID, session.TenantID, session.Token,
			session.IPAddress, session.UserAgent, session.ExpiresAt,
			session.CreatedAt, session.UpdatedAt,
		)

		if err == nil {
			// 缓存会话
			cacheKey := fmt.Sprintf("session:token:%s", session.Token)
			_ = r.cache.Set(ctx, cacheKey, session, int(time.Until(session.ExpiresAt).Seconds()))
		}

		return err
	})
}

// FindByID 根据 ID 查找会话
func (r *sessionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	cacheKey := fmt.Sprintf("session:id:%s", id.String())

	var session model.Session
	if err := r.cache.Get(ctx, cacheKey, &session); err == nil {
		return &session, nil
	}

	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, tenant_id, token, ip_address, user_agent,
			   expires_at, created_at, updated_at, revoked_at
		FROM sessions
		WHERE id = $1
	`, id)

	if err := r.scanSession(row, &session); err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, cacheKey, &session, 300)
	return &session, nil
}

// FindByToken 根据 Token 查找会话
func (r *sessionRepo) FindByToken(ctx context.Context, token string) (*model.Session, error) {
	cacheKey := fmt.Sprintf("session:token:%s", token)

	var session model.Session
	if err := r.cache.Get(ctx, cacheKey, &session); err == nil {
		return &session, nil
	}

	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, tenant_id, token, ip_address, user_agent,
			   expires_at, created_at, updated_at, revoked_at
		FROM sessions
		WHERE token = $1
	`, token)

	if err := r.scanSession(row, &session); err != nil {
		return nil, err
	}

	// 缓存到过期时间
	ttl := int(time.Until(session.ExpiresAt).Seconds())
	if ttl > 0 {
		_ = r.cache.Set(ctx, cacheKey, &session, ttl)
	}

	return &session, nil
}

// Update 更新会话
func (r *sessionRepo) Update(ctx context.Context, session *model.Session) error {
	session.UpdatedAt = time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE sessions SET
				ip_address = $2, user_agent = $3, expires_at = $4,
				updated_at = $5, revoked_at = $6
			WHERE id = $1
		`,
			session.ID, session.IPAddress, session.UserAgent,
			session.ExpiresAt, session.UpdatedAt, session.RevokedAt,
		)

		if err == nil {
			r.invalidateCache(session.ID, session.Token)
		}

		return err
	})
}

// Delete 删除会话
func (r *sessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		var token string
		err := tx.QueryRow(ctx, "SELECT token FROM sessions WHERE id = $1", id).Scan(&token)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, "DELETE FROM sessions WHERE id = $1", id)

		if err == nil {
			r.invalidateCache(id, token)
		}

		return err
	})
}

// GetUserSessions 获取用户的所有会话
func (r *sessionRepo) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*model.Session, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, tenant_id, token, ip_address, user_agent,
			   expires_at, created_at, updated_at, revoked_at
		FROM sessions
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*model.Session
	for rows.Next() {
		var session model.Session
		if err := r.scanSession(rows, &session); err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// RevokeSession 撤销会话
func (r *sessionRepo) RevokeSession(ctx context.Context, id uuid.UUID) error {
	now := time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		var token string
		err := tx.QueryRow(ctx, "SELECT token FROM sessions WHERE id = $1", id).Scan(&token)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx,
			"UPDATE sessions SET revoked_at = $1, updated_at = $1 WHERE id = $2",
			now, id,
		)

		if err == nil {
			r.invalidateCache(id, token)
		}

		return err
	})
}

// RevokeUserSessions 撤销用户的所有会话
func (r *sessionRepo) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 获取所有会话 token 用于清除缓存
		rows, err := tx.Query(ctx,
			"SELECT id, token FROM sessions WHERE user_id = $1 AND revoked_at IS NULL",
			userID,
		)
		if err != nil {
			return err
		}

		var sessionTokens []struct {
			id    uuid.UUID
			token string
		}

		for rows.Next() {
			var st struct {
				id    uuid.UUID
				token string
			}
			if err := rows.Scan(&st.id, &st.token); err != nil {
				rows.Close()
				return err
			}
			sessionTokens = append(sessionTokens, st)
		}
		rows.Close()

		// 撤销所有会话
		_, err = tx.Exec(ctx,
			"UPDATE sessions SET revoked_at = $1, updated_at = $1 WHERE user_id = $2 AND revoked_at IS NULL",
			now, userID,
		)

		if err == nil {
			// 清除缓存
			for _, st := range sessionTokens {
				r.invalidateCache(st.id, st.token)
			}
		}

		return err
	})
}

// CleanupExpiredSessions 清理过期会话
func (r *sessionRepo) CleanupExpiredSessions(ctx context.Context) error {
	now := time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			DELETE FROM sessions
			WHERE expires_at < $1 OR revoked_at < $1 - INTERVAL '7 days'
		`, now)

		return err
	})
}

// scanSession 扫描会话数据
func (r *sessionRepo) scanSession(row pgx.Row, session *model.Session) error {
	return row.Scan(
		&session.ID, &session.UserID, &session.TenantID, &session.Token,
		&session.IPAddress, &session.UserAgent, &session.ExpiresAt,
		&session.CreatedAt, &session.UpdatedAt, &session.RevokedAt,
	)
}

// invalidateCache 清除缓存
func (r *sessionRepo) invalidateCache(id uuid.UUID, token string) {
	ctx := context.Background()
	r.cache.Delete(ctx, fmt.Sprintf("session:id:%s", id.String()))
	r.cache.Delete(ctx, fmt.Sprintf("session:token:%s", token))
}
