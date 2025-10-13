package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSessionRepoTest(t *testing.T) (*database.DB, SessionRepository, uuid.UUID, uuid.UUID) {
	t.Helper()

	cfg := database.DefaultConfig()
	cfg.Host = "localhost"
	cfg.Port = 15000
	cfg.Database = "erp_test"
	cfg.Username = "postgres"
	cfg.Password = "postgres123"
	cfg.SSLMode = "disable"

	db, err := database.New(context.Background(), database.WithConfig(cfg))
	require.NoError(t, err)

	ctx := context.Background()
	tenantID := uuid.New()
	_, err = db.Exec(ctx, `
		INSERT INTO tenants (id, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, tenantID, "Test Tenant", "active")
	require.NoError(t, err)

	userID := uuid.New()
	_, err = db.Exec(ctx, `
		INSERT INTO users (id, tenant_id, username, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, userID, tenantID, "testuser", "test@example.com", "hash", "active")
	require.NoError(t, err)

	repo := NewSessionRepository(db, nil)
	return db, repo, tenantID, userID
}

func TestSessionRepo_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID, userID := setupSessionRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	session := &model.Session{
		UserID:       userID,
		TenantID:     tenantID,
		Token: "refresh_token_123",
		IPAddress:    "192.168.1.1",
		UserAgent:    "Test Agent",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, session.ID)

	db.Exec(ctx, "DELETE FROM sessions WHERE id = $1", session.ID)
}

func TestSessionRepo_FindByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID, userID := setupSessionRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	session := &model.Session{
		UserID:       userID,
		TenantID:     tenantID,
		Token: "refresh_token_456",
		IPAddress:    "192.168.1.2",
		UserAgent:    "Test Agent",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, repo.Create(ctx, session))
	defer db.Exec(ctx, "DELETE FROM sessions WHERE id = $1", session.ID)

	found, err := repo.FindByID(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, found.ID)
	assert.Equal(t, userID, found.UserID)
}

func TestSessionRepo_FindByToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID, userID := setupSessionRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	token := "unique_refresh_" + uuid.New().String()
	session := &model.Session{
		UserID:       userID,
		TenantID:     tenantID,
		Token: token,
		IPAddress:    "192.168.1.3",
		UserAgent:    "Test Agent",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, repo.Create(ctx, session))
	defer db.Exec(ctx, "DELETE FROM sessions WHERE id = $1", session.ID)

	found, err := repo.FindByToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, session.ID, found.ID)
	assert.Equal(t, token, found.Token)
}

func TestSessionRepo_Revoke(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID, userID := setupSessionRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	session := &model.Session{
		UserID:       userID,
		TenantID:     tenantID,
		Token: "revoke_token",
		IPAddress:    "192.168.1.4",
		UserAgent:    "Test Agent",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	require.NoError(t, repo.Create(ctx, session))
	defer db.Exec(ctx, "DELETE FROM sessions WHERE id = $1", session.ID)

	err := repo.RevokeSession(ctx, session.ID)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, session.ID)
	require.NoError(t, err)
	assert.NotNil(t, found.RevokedAt)
}

func TestSessionRepo_ListByUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID, userID := setupSessionRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	sessions := []*model.Session{
		{
			UserID:       userID,
			TenantID:     tenantID,
			Token: "list_token_1_" + uuid.New().String(),
			IPAddress:    "192.168.1.5",
			UserAgent:    "Agent 1",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		},
		{
			UserID:       userID,
			TenantID:     tenantID,
			Token: "list_token_2_" + uuid.New().String(),
			IPAddress:    "192.168.1.6",
			UserAgent:    "Agent 2",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		},
	}

	for _, s := range sessions {
		require.NoError(t, repo.Create(ctx, s))
		defer db.Exec(ctx, "DELETE FROM sessions WHERE id = $1", s.ID)
	}

	found, err := repo.GetUserSessions(ctx, userID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 2)
}
