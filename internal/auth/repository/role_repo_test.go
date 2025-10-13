package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRoleRepoTest(t *testing.T) (*database.DB, RoleRepository, uuid.UUID) {
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

	repo := NewRoleRepository(db, nil)
	return db, repo, tenantID
}

func TestRoleRepo_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupRoleRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	role := &model.Role{
		TenantID:    tenantID,
		Name:        "TestRole_" + uuid.New().String()[:8],
		Description: "Test role",
	}

	err := repo.Create(ctx, role)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, role.ID)

	db.Exec(ctx, "DELETE FROM roles WHERE id = $1", role.ID)
}

func TestRoleRepo_FindByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupRoleRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	role := &model.Role{
		TenantID:    tenantID,
		Name:        "FindRole_" + uuid.New().String()[:8],
		Description: "Find test",
	}
	require.NoError(t, repo.Create(ctx, role))
	defer db.Exec(ctx, "DELETE FROM roles WHERE id = $1", role.ID)

	found, err := repo.FindByID(ctx, role.ID)
	require.NoError(t, err)
	assert.Equal(t, role.ID, found.ID)
	assert.Equal(t, role.Name, found.Name)
}

func TestRoleRepo_FindByName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupRoleRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	roleName := "UniqueName_" + uuid.New().String()[:8]
	role := &model.Role{
		TenantID:    tenantID,
		Name:        roleName,
		Description: "Name test",
	}
	require.NoError(t, repo.Create(ctx, role))
	defer db.Exec(ctx, "DELETE FROM roles WHERE id = $1", role.ID)

	found, err := repo.FindByName(ctx, tenantID, roleName)
	require.NoError(t, err)
	assert.Equal(t, role.ID, found.ID)
	assert.Equal(t, roleName, found.Name)
}

func TestRoleRepo_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupRoleRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	role := &model.Role{
		TenantID:    tenantID,
		Name:        "UpdateRole_" + uuid.New().String()[:8],
		Description: "Original",
	}
	require.NoError(t, repo.Create(ctx, role))
	defer db.Exec(ctx, "DELETE FROM roles WHERE id = $1", role.ID)

	role.Description = "Updated description"
	err := repo.Update(ctx, role)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, role.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", found.Description)
}

func TestRoleRepo_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupRoleRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	role := &model.Role{
		TenantID:    tenantID,
		Name:        "DeleteRole_" + uuid.New().String()[:8],
		Description: "To be deleted",
	}
	require.NoError(t, repo.Create(ctx, role))

	err := repo.Delete(ctx, role.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, role.ID)
	assert.Error(t, err)

	db.Exec(ctx, "DELETE FROM roles WHERE id = $1", role.ID)
}

func TestRoleRepo_ListByTenant(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupRoleRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	roles := []*model.Role{
		{
			TenantID:    tenantID,
			Name:        "ListRole1_" + uuid.New().String()[:8],
			Description: "Role 1",
		},
		{
			TenantID:    tenantID,
			Name:        "ListRole2_" + uuid.New().String()[:8],
			Description: "Role 2",
		},
	}

	for _, r := range roles {
		require.NoError(t, repo.Create(ctx, r))
		defer db.Exec(ctx, "DELETE FROM roles WHERE id = $1", r.ID)
	}

	found, err := repo.ListByTenant(ctx, tenantID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 2)
}

func TestRoleRepo_AssignRoleToUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupRoleRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()

	// Create user
	userID := uuid.New()
	_, err := db.Exec(ctx, `
		INSERT INTO users (id, tenant_id, username, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`, userID, tenantID, "testuser", "test@example.com", "hash", "active")
	require.NoError(t, err)
	defer db.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)

	// Create role
	role := &model.Role{
		TenantID:    tenantID,
		Name:        "AssignRole_" + uuid.New().String()[:8],
		Description: "Assign test",
	}
	require.NoError(t, repo.Create(ctx, role))
	defer db.Exec(ctx, "DELETE FROM roles WHERE id = $1", role.ID)

	// Assign role to user
	err = repo.AssignRoleToUser(ctx, userID, role.ID, tenantID)
	require.NoError(t, err)

	// Verify assignment
	hasRole, err := repo.HasRole(ctx, userID, role.ID)
	require.NoError(t, err)
	assert.True(t, hasRole)

	// Cleanup
	db.Exec(ctx, "DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2", userID, role.ID)
}

func TestRoleRepo_GetUserRoles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupRoleRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()

	// Create user
	userID := uuid.New()
	_, err := db.Exec(ctx, `
		INSERT INTO users (id, tenant_id, username, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`, userID, tenantID, "roleuser", "roleuser@example.com", "hash", "active")
	require.NoError(t, err)
	defer db.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)

	// Create and assign roles
	roles := []*model.Role{
		{
			TenantID:    tenantID,
			Name:        "UserRole1_" + uuid.New().String()[:8],
			Description: "User Role 1",
		},
		{
			TenantID:    tenantID,
			Name:        "UserRole2_" + uuid.New().String()[:8],
			Description: "User Role 2",
		},
	}

	for _, r := range roles {
		require.NoError(t, repo.Create(ctx, r))
		defer db.Exec(ctx, "DELETE FROM roles WHERE id = $1", r.ID)

		err = repo.AssignRoleToUser(ctx, userID, r.ID, tenantID)
		require.NoError(t, err)
		defer db.Exec(ctx, "DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2", userID, r.ID)
	}

	// Get user roles
	userRoles, err := repo.GetUserRoles(ctx, userID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(userRoles), 2)
}
