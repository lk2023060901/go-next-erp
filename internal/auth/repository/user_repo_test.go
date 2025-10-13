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

func setupUserRepoTest(t *testing.T) (*database.DB, UserRepository, uuid.UUID) {
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

	// Create test tenant
	ctx := context.Background()
	tenantID := uuid.New()
	_, err = db.Exec(ctx, `
		INSERT INTO tenants (id, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, tenantID, "Test Tenant", "active")
	require.NoError(t, err)

	repo := NewUserRepository(db, nil)
	return db, repo, tenantID
}

func TestUserRepo_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupUserRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()

	user := &model.User{
		TenantID:     tenantID,
		Username:     "testuser_" + uuid.New().String()[:8],
		Email:        "test_" + uuid.New().String()[:8] + "@example.com",
		PasswordHash: "hash123",
		Status:       model.UserStatusActive,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.NotZero(t, user.CreatedAt)

	// Cleanup
	_, _ = db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
}

func TestUserRepo_FindByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupUserRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()
	
	ctx := context.Background()

	user := &model.User{
		TenantID:     tenantID,
		Username:     "findtest_" + uuid.New().String()[:8],
		Email:        "find_" + uuid.New().String()[:8] + "@example.com",
		PasswordHash: "hash123",
		Status:       model.UserStatusActive,
	}
	require.NoError(t, repo.Create(ctx, user))
	defer db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)

	found, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Username, found.Username)
}

func TestUserRepo_FindByUsername(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupUserRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()
	
	ctx := context.Background()
	username := "usertest_" + uuid.New().String()[:8]

	user := &model.User{
		TenantID:     tenantID,
		Username:     username,
		Email:        "usertest_" + uuid.New().String()[:8] + "@example.com",
		PasswordHash: "hash123",
		Status:       model.UserStatusActive,
	}
	require.NoError(t, repo.Create(ctx, user))
	defer db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)

	found, err := repo.FindByUsername(ctx, username)
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, username, found.Username)
}

func TestUserRepo_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupUserRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()
	
	ctx := context.Background()

	user := &model.User{
		TenantID:     tenantID,
		Username:     "updatetest_" + uuid.New().String()[:8],
		Email:        "update_" + uuid.New().String()[:8] + "@example.com",
		PasswordHash: "hash123",
		Status:       model.UserStatusActive,
	}
	require.NoError(t, repo.Create(ctx, user))
	defer db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)

	user.Email = "newemail_" + uuid.New().String()[:8] + "@example.com"
	user.Status = model.UserStatusInactive
	err := repo.Update(ctx, user)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, model.UserStatusInactive, found.Status)
}

func TestUserRepo_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupUserRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()
	
	ctx := context.Background()

	user := &model.User{
		TenantID:     tenantID,
		Username:     "deletetest_" + uuid.New().String()[:8],
		Email:        "delete_" + uuid.New().String()[:8] + "@example.com",
		PasswordHash: "hash123",
		Status:       model.UserStatusActive,
	}
	require.NoError(t, repo.Create(ctx, user))

	err := repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, user.ID)
	assert.Error(t, err)

	db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
}

func TestUserRepo_ListByTenant(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupUserRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()
	
	ctx := context.Background()

	users := []*model.User{
		{
			TenantID:     tenantID,
			Username:     "list1_" + uuid.New().String()[:8],
			Email:        "list1_" + uuid.New().String()[:8] + "@example.com",
			PasswordHash: "hash123",
			Status:       model.UserStatusActive,
		},
		{
			TenantID:     tenantID,
			Username:     "list2_" + uuid.New().String()[:8],
			Email:        "list2_" + uuid.New().String()[:8] + "@example.com",
			PasswordHash: "hash123",
			Status:       model.UserStatusActive,
		},
	}

	for _, u := range users {
		require.NoError(t, repo.Create(ctx, u))
		defer db.Exec(ctx, "DELETE FROM users WHERE id = $1", u.ID)
	}

	found, err := repo.ListByTenant(ctx, tenantID, 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 2)

	total, err := repo.CountByTenant(ctx, tenantID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(2))
}

func TestUserRepo_UpdateLastLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupUserRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	user := &model.User{
		TenantID:     tenantID,
		Username:     "logintest_" + uuid.New().String()[:8],
		Email:        "login_" + uuid.New().String()[:8] + "@example.com",
		PasswordHash: "hash123",
		Status:       model.UserStatusActive,
	}
	require.NoError(t, repo.Create(ctx, user))
	defer db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)

	err := repo.UpdateLastLogin(ctx, user.ID, "192.168.1.1")
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.NotNil(t, found.LastLoginAt)
	require.NotNil(t, found.LastLoginIP)
	assert.Equal(t, "192.168.1.1", *found.LastLoginIP)
}

func TestUserRepo_LockUnlock(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupUserRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	user := &model.User{
		TenantID:     tenantID,
		Username:     "locktest_" + uuid.New().String()[:8],
		Email:        "lock_" + uuid.New().String()[:8] + "@example.com",
		PasswordHash: "hash123",
		Status:       model.UserStatusActive,
	}
	require.NoError(t, repo.Create(ctx, user))
	defer db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)

	// Lock user
	lockUntil := time.Now().Add(1 * time.Hour)
	err := repo.LockUser(ctx, user.ID, lockUntil)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.NotNil(t, found.LockedUntil)

	// Unlock user
	err = repo.UnlockUser(ctx, user.ID)
	require.NoError(t, err)

	found, err = repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Nil(t, found.LockedUntil)
}
