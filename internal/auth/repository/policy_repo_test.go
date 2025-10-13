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

func setupPolicyRepoTest(t *testing.T) (*database.DB, PolicyRepository, uuid.UUID) {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping integration test")
	}

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

	repo := NewPolicyRepository(db, nil)
	return db, repo, tenantID
}

func TestPolicyRepo_Create(t *testing.T) {
	db, repo, tenantID := setupPolicyRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	policy := &model.Policy{
		Name:        "Test Policy",
		Description: "Test policy description",
		TenantID:    tenantID,
		Resource:    "document",
		Action:      "read",
		Effect:      "allow",
		Expression:  "true",
		Priority:    100,
		Enabled:     true,
	}

	err := repo.Create(ctx, policy)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, policy.ID)
	assert.False(t, policy.CreatedAt.IsZero())
	assert.False(t, policy.UpdatedAt.IsZero())
}

func TestPolicyRepo_FindByID(t *testing.T) {
	db, repo, tenantID := setupPolicyRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	policy := &model.Policy{
		Name:        "Find Test Policy",
		Description: "Test",
		TenantID:    tenantID,
		Resource:    "file",
		Action:      "write",
		Effect:      "allow",
		Expression:  "true",
		Priority:    50,
		Enabled:     true,
	}

	err := repo.Create(ctx, policy)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, policy.ID)
	require.NoError(t, err)
	assert.Equal(t, policy.Name, found.Name)
	assert.Equal(t, policy.Resource, found.Resource)
	assert.Equal(t, policy.Action, found.Action)
	assert.Equal(t, policy.Effect, found.Effect)
}

func TestPolicyRepo_FindByID_NotFound(t *testing.T) {
	db, repo, _ := setupPolicyRepoTest(t)
	defer db.Close()

	ctx := context.Background()
	_, err := repo.FindByID(ctx, uuid.New())
	assert.Error(t, err)
}

func TestPolicyRepo_Update(t *testing.T) {
	db, repo, tenantID := setupPolicyRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	policy := &model.Policy{
		Name:        "Update Test",
		Description: "Original",
		TenantID:    tenantID,
		Resource:    "api",
		Action:      "execute",
		Effect:      "allow",
		Expression:  "true",
		Priority:    75,
		Enabled:     true,
	}

	err := repo.Create(ctx, policy)
	require.NoError(t, err)

	policy.Description = "Updated"
	policy.Priority = 90
	policy.Enabled = false

	err = repo.Update(ctx, policy)
	require.NoError(t, err)

	updated, err := repo.FindByID(ctx, policy.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Description)
	assert.Equal(t, 90, updated.Priority)
	assert.False(t, updated.Enabled)
}

func TestPolicyRepo_Delete(t *testing.T) {
	db, repo, tenantID := setupPolicyRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	policy := &model.Policy{
		Name:        "Delete Test",
		Description: "Will be deleted",
		TenantID:    tenantID,
		Resource:    "resource",
		Action:      "delete",
		Effect:      "deny",
		Expression:  "false",
		Priority:    10,
		Enabled:     true,
	}

	err := repo.Create(ctx, policy)
	require.NoError(t, err)

	err = repo.Delete(ctx, policy.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, policy.ID)
	assert.Error(t, err)
}

func TestPolicyRepo_List(t *testing.T) {
	db, repo, tenantID := setupPolicyRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	// 创建多个策略
	policies := []*model.Policy{
		{
			Name:        "Policy 1",
			Description: "First",
			TenantID:    tenantID,
			Resource:    "resource1",
			Action:      "read",
			Effect:      "allow",
			Expression:  "true",
			Priority:    100,
			Enabled:     true,
		},
		{
			Name:        "Policy 2",
			Description: "Second",
			TenantID:    tenantID,
			Resource:    "resource2",
			Action:      "write",
			Effect:      "allow",
			Expression:  "true",
			Priority:    200,
			Enabled:     true,
		},
	}

	for _, p := range policies {
		err := repo.Create(ctx, p)
		require.NoError(t, err)
	}

	result, err := repo.ListByTenant(ctx, tenantID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 2)
}

func TestPolicyRepo_GetApplicablePolicies(t *testing.T) {
	db, repo, tenantID := setupPolicyRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	// 创建不同资源和动作的策略
	policies := []*model.Policy{
		{
			Name:        "Read Document",
			Description: "Allow read document",
			TenantID:    tenantID,
			Resource:    "document",
			Action:      "read",
			Effect:      "allow",
			Expression:  "true",
			Priority:    100,
			Enabled:     true,
		},
		{
			Name:        "Write Document",
			Description: "Allow write document",
			TenantID:    tenantID,
			Resource:    "document",
			Action:      "write",
			Effect:      "allow",
			Expression:  "true",
			Priority:    100,
			Enabled:     true,
		},
		{
			Name:        "Read File",
			Description: "Allow read file",
			TenantID:    tenantID,
			Resource:    "file",
			Action:      "read",
			Effect:      "allow",
			Expression:  "true",
			Priority:    100,
			Enabled:     true,
		},
	}

	for _, p := range policies {
		err := repo.Create(ctx, p)
		require.NoError(t, err)
	}

	// 查询 document:read
	result, err := repo.GetApplicablePolicies(ctx, tenantID, "document", "read")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Read Document", result[0].Name)

	// 查询 document:write
	result, err = repo.GetApplicablePolicies(ctx, tenantID, "document", "write")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Write Document", result[0].Name)
}

func TestPolicyRepo_Wildcard(t *testing.T) {
	db, repo, tenantID := setupPolicyRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	// 创建通配符策略
	policy := &model.Policy{
		Name:        "Wildcard Policy",
		Description: "Applies to all resources",
		TenantID:    tenantID,
		Resource:    "*",
		Action:      "*",
		Effect:      "allow",
		Expression:  "true",
		Priority:    1,
		Enabled:     true,
	}

	err := repo.Create(ctx, policy)
	require.NoError(t, err)

	// 查询任意资源和动作都应该找到通配符策略
	result, err := repo.GetApplicablePolicies(ctx, tenantID, "any_resource", "any_action")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 1)

	found := false
	for _, p := range result {
		if p.Resource == "*" && p.Action == "*" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find wildcard policy")
}

func TestPolicyRepo_Priority(t *testing.T) {
	db, repo, tenantID := setupPolicyRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	// 创建不同优先级的策略
	policies := []*model.Policy{
		{
			Name:        "High Priority",
			Description: "High",
			TenantID:    tenantID,
			Resource:    "test",
			Action:      "action",
			Effect:      "allow",
			Expression:  "true",
			Priority:    100,
			Enabled:     true,
		},
		{
			Name:        "Low Priority",
			Description: "Low",
			TenantID:    tenantID,
			Resource:    "test",
			Action:      "action",
			Effect:      "deny",
			Expression:  "true",
			Priority:    50,
			Enabled:     true,
		},
	}

	for _, p := range policies {
		err := repo.Create(ctx, p)
		require.NoError(t, err)
	}

	// 查询应该按优先级降序返回
	result, err := repo.GetApplicablePolicies(ctx, tenantID, "test", "action")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 2)

	// 验证第一个是高优先级
	assert.Equal(t, 100, result[0].Priority)
	assert.Equal(t, "High Priority", result[0].Name)
}
