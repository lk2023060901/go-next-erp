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

func setupPermissionRepoTest(t *testing.T) (*database.DB, PermissionRepository, uuid.UUID) {
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

	repo := NewPermissionRepository(db, nil)
	return db, repo, tenantID
}

func TestPermissionRepo_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupPermissionRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	permission := &model.Permission{
		TenantID:    tenantID,
		Resource:    "users_" + uuid.New().String()[:8],
		Action:      "read",
		DisplayName: "Read Users",
		Description: "Permission to read users",
	}

	err := repo.Create(ctx, permission)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, permission.ID)

	db.Exec(ctx, "DELETE FROM permissions WHERE id = $1", permission.ID)
}

func TestPermissionRepo_FindByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupPermissionRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	permission := &model.Permission{
		TenantID:    tenantID,
		Resource:    "products_" + uuid.New().String()[:8],
		Action:      "write",
		DisplayName: "Write Products",
		Description: "Permission to write products",
	}
	require.NoError(t, repo.Create(ctx, permission))
	defer db.Exec(ctx, "DELETE FROM permissions WHERE id = $1", permission.ID)

	found, err := repo.FindByID(ctx, permission.ID)
	require.NoError(t, err)
	assert.Equal(t, permission.ID, found.ID)
	assert.Equal(t, permission.Resource, found.Resource)
	assert.Equal(t, permission.Action, found.Action)
}

func TestPermissionRepo_FindByResourceAction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupPermissionRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	resource := "orders_" + uuid.New().String()[:8]
	permission := &model.Permission{
		TenantID:    tenantID,
		Resource:    resource,
		Action:      "delete",
		DisplayName: "Delete Orders",
		Description: "Permission to delete orders",
	}
	require.NoError(t, repo.Create(ctx, permission))
	defer db.Exec(ctx, "DELETE FROM permissions WHERE id = $1", permission.ID)

	found, err := repo.FindByResourceAction(ctx, tenantID, resource, "delete")
	require.NoError(t, err)
	assert.Equal(t, permission.ID, found.ID)
	assert.Equal(t, resource, found.Resource)
	assert.Equal(t, "delete", found.Action)
}

func TestPermissionRepo_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupPermissionRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	permission := &model.Permission{
		TenantID:    tenantID,
		Resource:    "reports_" + uuid.New().String()[:8],
		Action:      "read",
		DisplayName: "Read Reports",
		Description: "Original description",
	}
	require.NoError(t, repo.Create(ctx, permission))
	defer db.Exec(ctx, "DELETE FROM permissions WHERE id = $1", permission.ID)

	permission.Description = "Updated description"
	err := repo.Update(ctx, permission)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, permission.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", found.Description)
}

func TestPermissionRepo_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupPermissionRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	permission := &model.Permission{
		TenantID:    tenantID,
		Resource:    "invoices_" + uuid.New().String()[:8],
		Action:      "read",
		DisplayName: "Read Invoices",
		Description: "To be deleted",
	}
	require.NoError(t, repo.Create(ctx, permission))

	err := repo.Delete(ctx, permission.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, permission.ID)
	assert.Error(t, err)

	db.Exec(ctx, "DELETE FROM permissions WHERE id = $1", permission.ID)
}

func TestPermissionRepo_ListByTenant(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupPermissionRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()
	permissions := []*model.Permission{
		{
			TenantID:    tenantID,
			Resource:    "customers_" + uuid.New().String()[:8],
			Action:      "read",
			DisplayName: "Read Customers",
			Description: "Permission 1",
		},
		{
			TenantID:    tenantID,
			Resource:    "customers_" + uuid.New().String()[:8],
			Action:      "write",
			DisplayName: "Write Customers",
			Description: "Permission 2",
		},
	}

	for _, p := range permissions {
		require.NoError(t, repo.Create(ctx, p))
		defer db.Exec(ctx, "DELETE FROM permissions WHERE id = $1", p.ID)
	}

	found, err := repo.ListByTenant(ctx, tenantID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 2)
}

func TestPermissionRepo_AssignPermissionToRole(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupPermissionRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()

	// Create role
	roleID := uuid.New()
	_, err := db.Exec(ctx, `
		INSERT INTO roles (id, tenant_id, name, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, roleID, tenantID, "TestRole_"+uuid.New().String()[:8], "Test Role")
	require.NoError(t, err)
	defer db.Exec(ctx, "DELETE FROM roles WHERE id = $1", roleID)

	// Create permission
	permission := &model.Permission{
		TenantID:    tenantID,
		Resource:    "settings_" + uuid.New().String()[:8],
		Action:      "manage",
		DisplayName: "Manage Settings",
		Description: "Assign test",
	}
	require.NoError(t, repo.Create(ctx, permission))
	defer db.Exec(ctx, "DELETE FROM permissions WHERE id = $1", permission.ID)

	// Assign permission to role
	err = repo.AssignPermissionToRole(ctx, roleID, permission.ID, tenantID)
	require.NoError(t, err)

	// Verify assignment
	hasPermission, err := repo.HasPermission(ctx, roleID, permission.ID)
	require.NoError(t, err)
	assert.True(t, hasPermission)

	// Cleanup
	db.Exec(ctx, "DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2", roleID, permission.ID)
}

func TestPermissionRepo_GetRolePermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupPermissionRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()

	// Create role
	roleID := uuid.New()
	_, err := db.Exec(ctx, `
		INSERT INTO roles (id, tenant_id, name, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, roleID, tenantID, "PermRole_"+uuid.New().String()[:8], "Permission Role")
	require.NoError(t, err)
	defer db.Exec(ctx, "DELETE FROM roles WHERE id = $1", roleID)

	// Create and assign permissions
	permissions := []*model.Permission{
		{
			TenantID:    tenantID,
			Resource:    "dashboard_" + uuid.New().String()[:8],
			Action:      "view",
			DisplayName: "View Dashboard",
			Description: "Permission 1",
		},
		{
			TenantID:    tenantID,
			Resource:    "analytics_" + uuid.New().String()[:8],
			Action:      "read",
			DisplayName: "Read Analytics",
			Description: "Permission 2",
		},
	}

	for _, p := range permissions {
		require.NoError(t, repo.Create(ctx, p))
		defer db.Exec(ctx, "DELETE FROM permissions WHERE id = $1", p.ID)

		err = repo.AssignPermissionToRole(ctx, roleID, p.ID, tenantID)
		require.NoError(t, err)
		defer db.Exec(ctx, "DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2", roleID, p.ID)
	}

	// Get role permissions
	rolePerms, err := repo.GetRolePermissions(ctx, roleID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(rolePerms), 2)
}

func TestPermissionRepo_GetUserPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo, tenantID := setupPermissionRepoTest(t)
	defer db.Close()
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)

	ctx := context.Background()

	// Create user
	userID := uuid.New()
	_, err := db.Exec(ctx, `
		INSERT INTO users (id, tenant_id, username, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`, userID, tenantID, "permuser", "permuser@example.com", "hash", "active")
	require.NoError(t, err)
	defer db.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)

	// Create role
	roleID := uuid.New()
	_, err = db.Exec(ctx, `
		INSERT INTO roles (id, tenant_id, name, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, roleID, tenantID, "UserPermRole_"+uuid.New().String()[:8], "User Permission Role")
	require.NoError(t, err)
	defer db.Exec(ctx, "DELETE FROM roles WHERE id = $1", roleID)

	// Assign role to user
	urID := uuid.New()
	_, err = db.Exec(ctx, `
		INSERT INTO user_roles (id, user_id, role_id, tenant_id, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, urID, userID, roleID, tenantID)
	require.NoError(t, err)
	defer db.Exec(ctx, "DELETE FROM user_roles WHERE id = $1", urID)

	// Create and assign permissions to role
	permission := &model.Permission{
		TenantID:    tenantID,
		Resource:    "exports_" + uuid.New().String()[:8],
		Action:      "download",
		DisplayName: "Download Exports",
		Description: "User permission test",
	}
	require.NoError(t, repo.Create(ctx, permission))
	defer db.Exec(ctx, "DELETE FROM permissions WHERE id = $1", permission.ID)

	err = repo.AssignPermissionToRole(ctx, roleID, permission.ID, tenantID)
	require.NoError(t, err)
	defer db.Exec(ctx, "DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2", roleID, permission.ID)

	// Get user permissions
	userPerms, err := repo.GetUserPermissions(ctx, userID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(userPerms), 1)
}
