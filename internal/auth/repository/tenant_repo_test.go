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

func setupTenantRepoTest(t *testing.T) (*database.DB, TenantRepository) {
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

	repo := NewTenantRepository(db, nil)
	return db, repo
}

func TestTenantRepo_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo := setupTenantRepoTest(t)
	defer db.Close()

	ctx := context.Background()
	tenant := &model.Tenant{
		Name:   "TestTenant_" + uuid.New().String()[:8],
		Domain: "test-" + uuid.New().String()[:8] + ".example.com",
		Status: model.TenantStatusActive,
		Settings: map[string]interface{}{
			"theme": "dark",
		},
	}

	err := repo.Create(ctx, tenant)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, tenant.ID)

	db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
}

func TestTenantRepo_FindByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo := setupTenantRepoTest(t)
	defer db.Close()

	ctx := context.Background()
	tenant := &model.Tenant{
		Name:   "FindByIDTenant_" + uuid.New().String()[:8],
		Status: model.TenantStatusActive,
	}
	require.NoError(t, repo.Create(ctx, tenant))
	defer db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)

	found, err := repo.FindByID(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, found.ID)
	assert.Equal(t, tenant.Name, found.Name)
	assert.Equal(t, model.TenantStatusActive, found.Status)
}

func TestTenantRepo_FindByDomain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo := setupTenantRepoTest(t)
	defer db.Close()

	ctx := context.Background()
	domain := "findbydomain-" + uuid.New().String()[:8] + ".example.com"
	tenant := &model.Tenant{
		Name:   "DomainTenant_" + uuid.New().String()[:8],
		Domain: domain,
		Status: model.TenantStatusActive,
	}
	require.NoError(t, repo.Create(ctx, tenant))
	defer db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)

	found, err := repo.FindByDomain(ctx, domain)
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, found.ID)
	assert.Equal(t, domain, found.Domain)
}

func TestTenantRepo_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo := setupTenantRepoTest(t)
	defer db.Close()

	ctx := context.Background()
	tenant := &model.Tenant{
		Name:   "UpdateTenant_" + uuid.New().String()[:8],
		Status: model.TenantStatusActive,
		Settings: map[string]interface{}{
			"feature": "old",
		},
	}
	require.NoError(t, repo.Create(ctx, tenant))
	defer db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)

	tenant.Settings = map[string]interface{}{
		"feature": "new",
	}
	err := repo.Update(ctx, tenant)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Equal(t, "new", found.Settings["feature"])
}

func TestTenantRepo_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo := setupTenantRepoTest(t)
	defer db.Close()

	ctx := context.Background()
	tenant := &model.Tenant{
		Name:   "DeleteTenant_" + uuid.New().String()[:8],
		Status: model.TenantStatusActive,
	}
	require.NoError(t, repo.Create(ctx, tenant))

	err := repo.Delete(ctx, tenant.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, tenant.ID)
	assert.Error(t, err)

	db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
}

func TestTenantRepo_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo := setupTenantRepoTest(t)
	defer db.Close()

	ctx := context.Background()
	tenants := []*model.Tenant{
		{
			Name:   "ListTenant1_" + uuid.New().String()[:8],
			Status: model.TenantStatusActive,
		},
		{
			Name:   "ListTenant2_" + uuid.New().String()[:8],
			Status: model.TenantStatusActive,
		},
	}

	for _, tenant := range tenants {
		require.NoError(t, repo.Create(ctx, tenant))
		defer db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)
	}

	found, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 2)
}

func TestTenantRepo_UpdateStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo := setupTenantRepoTest(t)
	defer db.Close()

	ctx := context.Background()
	tenant := &model.Tenant{
		Name:   "StatusTenant_" + uuid.New().String()[:8],
		Status: model.TenantStatusActive,
	}
	require.NoError(t, repo.Create(ctx, tenant))
	defer db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)

	err := repo.UpdateStatus(ctx, tenant.ID, model.TenantStatusSuspended)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Equal(t, model.TenantStatusSuspended, found.Status)
}

func TestTenantRepo_UpdateSettings(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db, repo := setupTenantRepoTest(t)
	defer db.Close()

	ctx := context.Background()
	tenant := &model.Tenant{
		Name:   "SettingsTenant_" + uuid.New().String()[:8],
		Status: model.TenantStatusActive,
		Settings: map[string]interface{}{
			"key1": "value1",
		},
	}
	require.NoError(t, repo.Create(ctx, tenant))
	defer db.Exec(ctx, "DELETE FROM tenants WHERE id = $1", tenant.ID)

	newSettings := map[string]interface{}{
		"key1": "updated",
		"key2": "value2",
	}
	err := repo.UpdateSettings(ctx, tenant.ID, newSettings)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated", found.Settings["key1"])
	assert.Equal(t, "value2", found.Settings["key2"])
}
