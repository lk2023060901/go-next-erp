package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/pkg/cache"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// 基础 CRUD
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 多租户查询
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.User, error)
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)

	// 认证相关
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, ip string) error
	IncrementLoginAttempts(ctx context.Context, userID uuid.UUID) error
	ResetLoginAttempts(ctx context.Context, userID uuid.UUID) error
	LockUser(ctx context.Context, userID uuid.UUID, until time.Time) error
	UnlockUser(ctx context.Context, userID uuid.UUID) error
}

// userRepo 用户仓储实现
type userRepo struct {
	db    *database.DB
	cache *cache.Cache
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *database.DB, cache *cache.Cache) UserRepository {
	return &userRepo{
		db:    db,
		cache: cache,
	}
}

// Create 创建用户
func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	// 生成 UUID v7
	user.ID = uuid.Must(uuid.NewV7())
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// 序列化 metadata
	metadataJSON, _ := json.Marshal(user.Metadata)

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO users (
				id, username, email, password_hash, tenant_id, status,
				mfa_enabled, mfa_secret, last_login_at, last_login_ip,
				login_attempts, locked_until, metadata, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		`,
			user.ID, user.Username, user.Email, user.PasswordHash, user.TenantID, user.Status,
			user.MFAEnabled, user.MFASecret, user.LastLoginAt, user.LastLoginIP,
			user.LoginAttempts, user.LockedUntil, metadataJSON, user.CreatedAt, user.UpdatedAt,
		)

		if err != nil {
			return err
		}

		// 清除缓存
		r.invalidateCache(user.Username, user.Email, user.ID)

		return nil
	})
}

// FindByID 根据 ID 查找用户
func (r *userRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	cacheKey := fmt.Sprintf("user:id:%s", id.String())

	// 尝试从缓存获取
	var user model.User
	if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	// 从数据库查询（自动路由到从库）
	row := r.db.QueryRow(ctx, `
		SELECT id, username, email, password_hash, tenant_id, status,
			   mfa_enabled, mfa_secret, last_login_at, last_login_ip,
			   login_attempts, locked_until, metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, id)

	if err := r.scanUser(row, &user); err != nil {
		return nil, err
	}

	// 写入缓存
	_ = r.cache.Set(ctx, cacheKey, &user, 300) // 5分钟

	return &user, nil
}

// FindByUsername 根据用户名查找
func (r *userRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	cacheKey := fmt.Sprintf("user:username:%s", username)

	var user model.User
	if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	row := r.db.QueryRow(ctx, `
		SELECT id, username, email, password_hash, tenant_id, status,
			   mfa_enabled, mfa_secret, last_login_at, last_login_ip,
			   login_attempts, locked_until, metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL
	`, username)

	if err := r.scanUser(row, &user); err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, cacheKey, &user, 300)
	return &user, nil
}

// FindByEmail 根据邮箱查找
func (r *userRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	cacheKey := fmt.Sprintf("user:email:%s", email)

	var user model.User
	if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	row := r.db.QueryRow(ctx, `
		SELECT id, username, email, password_hash, tenant_id, status,
			   mfa_enabled, mfa_secret, last_login_at, last_login_ip,
			   login_attempts, locked_until, metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`, email)

	if err := r.scanUser(row, &user); err != nil {
		return nil, err
	}

	_ = r.cache.Set(ctx, cacheKey, &user, 300)
	return &user, nil
}

// Update 更新用户
func (r *userRepo) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()
	metadataJSON, _ := json.Marshal(user.Metadata)

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE users SET
				username = $2, email = $3, password_hash = $4, status = $5,
				mfa_enabled = $6, mfa_secret = $7, metadata = $8, updated_at = $9
			WHERE id = $1 AND deleted_at IS NULL
		`,
			user.ID, user.Username, user.Email, user.PasswordHash, user.Status,
			user.MFAEnabled, user.MFASecret, metadataJSON, user.UpdatedAt,
		)

		if err != nil {
			return err
		}

		// 清除缓存
		r.invalidateCache(user.Username, user.Email, user.ID)

		return nil
	})
}

// Delete 软删除用户
func (r *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		// 先查询用户信息（用于清除缓存）
		var username, email string
		err := tx.QueryRow(ctx, "SELECT username, email FROM users WHERE id = $1", id).Scan(&username, &email)
		if err != nil {
			return err
		}

		// 软删除
		_, err = tx.Exec(ctx, "UPDATE users SET deleted_at = $1 WHERE id = $2", now, id)
		if err != nil {
			return err
		}

		// 清除缓存
		r.invalidateCache(username, email, id)

		return nil
	})
}

// ListByTenant 根据租户查询用户列表
func (r *userRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, username, email, password_hash, tenant_id, status,
			   mfa_enabled, mfa_secret, last_login_at, last_login_ip,
			   login_attempts, locked_until, metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tenantID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var user model.User
		if err := r.scanUser(rows, &user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

// CountByTenant 统计租户用户数
func (r *userRepo) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND deleted_at IS NULL
	`, tenantID).Scan(&count)

	return count, err
}

// UpdateLastLogin 更新最后登录信息
func (r *userRepo) UpdateLastLogin(ctx context.Context, userID uuid.UUID, ip string) error {
	now := time.Now()

	_, err := r.db.Master().Exec(ctx, `
		UPDATE users SET last_login_at = $1, last_login_ip = $2, updated_at = $1
		WHERE id = $3
	`, now, ip, userID)

	if err == nil {
		// 清除缓存
		r.cache.Delete(ctx, fmt.Sprintf("user:id:%s", userID.String()))
	}

	return err
}

// IncrementLoginAttempts 增加登录失败次数
func (r *userRepo) IncrementLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Master().Exec(ctx, `
		UPDATE users SET login_attempts = login_attempts + 1, updated_at = NOW()
		WHERE id = $1
	`, userID)

	return err
}

// ResetLoginAttempts 重置登录失败次数
func (r *userRepo) ResetLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Master().Exec(ctx, `
		UPDATE users SET login_attempts = 0, updated_at = NOW()
		WHERE id = $1
	`, userID)

	return err
}

// LockUser 锁定用户
func (r *userRepo) LockUser(ctx context.Context, userID uuid.UUID, until time.Time) error {
	_, err := r.db.Master().Exec(ctx, `
		UPDATE users SET locked_until = $1, status = $2, updated_at = NOW()
		WHERE id = $3
	`, until, model.UserStatusLocked, userID)

	if err == nil {
		r.cache.Delete(ctx, fmt.Sprintf("user:id:%s", userID.String()))
	}

	return err
}

// UnlockUser 解锁用户
func (r *userRepo) UnlockUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Master().Exec(ctx, `
		UPDATE users SET locked_until = NULL, status = $1, updated_at = NOW()
		WHERE id = $2
	`, model.UserStatusActive, userID)

	if err == nil {
		r.cache.Delete(ctx, fmt.Sprintf("user:id:%s", userID.String()))
	}

	return err
}

// scanUser 扫描用户数据
func (r *userRepo) scanUser(row pgx.Row, user *model.User) error {
	var metadataJSON []byte

	err := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.TenantID, &user.Status,
		&user.MFAEnabled, &user.MFASecret, &user.LastLoginAt, &user.LastLoginIP,
		&user.LoginAttempts, &user.LockedUntil, &metadataJSON, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err != nil {
		return err
	}

	// 解析 metadata
	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &user.Metadata)
	}

	return nil
}

// invalidateCache 清除相关缓存
func (r *userRepo) invalidateCache(username, email string, id uuid.UUID) {
	ctx := context.Background()
	r.cache.Delete(ctx, fmt.Sprintf("user:username:%s", username))
	r.cache.Delete(ctx, fmt.Sprintf("user:email:%s", email))
	r.cache.Delete(ctx, fmt.Sprintf("user:id:%s", id.String()))
}
