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

func setupRelationRepoTest(t *testing.T) (*database.DB, RelationRepository, uuid.UUID) {
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

	repo := NewRelationRepository(db, nil)
	return db, repo, tenantID
}

func TestRelationRepo_Create(t *testing.T) {
	db, repo, tenantID := setupRelationRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	tuple := &model.RelationTuple{
		TenantID: tenantID,
		Subject:  "user:123",
		Relation: "member",
		Object:   "group:456",
	}

	err := repo.Create(ctx, tuple)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, tuple.ID)
	assert.False(t, tuple.CreatedAt.IsZero())

	// 验证已创建
	exists, err := repo.Check(ctx, tenantID, "user:123", "member", "group:456")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRelationRepo_Create_Duplicate(t *testing.T) {
	db, repo, tenantID := setupRelationRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	tuple := &model.RelationTuple{
		TenantID: tenantID,
		Subject:  "user:789",
		Relation: "owner",
		Object:   "document:101",
	}

	err := repo.Create(ctx, tuple)
	require.NoError(t, err)

	// 重复创建应该成功但不报错
	err = repo.Create(ctx, tuple)
	require.NoError(t, err)
}

func TestRelationRepo_Delete(t *testing.T) {
	db, repo, tenantID := setupRelationRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	tuple := &model.RelationTuple{
		TenantID: tenantID,
		Subject:  "user:111",
		Relation: "viewer",
		Object:   "resource:222",
	}

	err := repo.Create(ctx, tuple)
	require.NoError(t, err)

	err = repo.Delete(ctx, tuple.ID)
	require.NoError(t, err)

	// 验证已删除
	exists, err := repo.Check(ctx, tenantID, "user:111", "viewer", "resource:222")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestRelationRepo_DeleteByTuple(t *testing.T) {
	db, repo, tenantID := setupRelationRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	tuple := &model.RelationTuple{
		TenantID: tenantID,
		Subject:  "user:333",
		Relation: "editor",
		Object:   "file:444",
	}

	err := repo.Create(ctx, tuple)
	require.NoError(t, err)

	err = repo.DeleteByTuple(ctx, tenantID, "user:333", "editor", "file:444")
	require.NoError(t, err)

	// 验证已删除
	exists, err := repo.Check(ctx, tenantID, "user:333", "editor", "file:444")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestRelationRepo_FindBySubject(t *testing.T) {
	db, repo, tenantID := setupRelationRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	// 创建多个关系
	tuples := []*model.RelationTuple{
		{TenantID: tenantID, Subject: "user:555", Relation: "owner", Object: "resource:666"},
		{TenantID: tenantID, Subject: "user:555", Relation: "editor", Object: "resource:777"},
		{TenantID: tenantID, Subject: "user:888", Relation: "viewer", Object: "resource:666"},
	}

	for _, tuple := range tuples {
		err := repo.Create(ctx, tuple)
		require.NoError(t, err)
	}

	// 查询 user:555 的所有关系
	result, err := repo.FindBySubject(ctx, tenantID, "user:555")
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestRelationRepo_FindByObject(t *testing.T) {
	db, repo, tenantID := setupRelationRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	// 创建多个关系
	tuples := []*model.RelationTuple{
		{TenantID: tenantID, Subject: "user:aaa", Relation: "owner", Object: "document:bbb"},
		{TenantID: tenantID, Subject: "user:ccc", Relation: "viewer", Object: "document:bbb"},
		{TenantID: tenantID, Subject: "user:aaa", Relation: "editor", Object: "document:ddd"},
	}

	for _, tuple := range tuples {
		err := repo.Create(ctx, tuple)
		require.NoError(t, err)
	}

	// 查询 document:bbb 的所有关系
	result, err := repo.FindByObject(ctx, tenantID, "document:bbb")
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestRelationRepo_FindByRelation(t *testing.T) {
	db, repo, tenantID := setupRelationRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	// 创建多个关系
	tuples := []*model.RelationTuple{
		{TenantID: tenantID, Subject: "user:eee", Relation: "member", Object: "group:fff"},
		{TenantID: tenantID, Subject: "user:eee", Relation: "member", Object: "group:ggg"},
		{TenantID: tenantID, Subject: "user:eee", Relation: "owner", Object: "group:fff"},
	}

	for _, tuple := range tuples {
		err := repo.Create(ctx, tuple)
		require.NoError(t, err)
	}

	// 查询 user:eee 的 member 关系
	result, err := repo.FindByRelation(ctx, tenantID, "user:eee", "member")
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestRelationRepo_Check(t *testing.T) {
	db, repo, tenantID := setupRelationRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	tuple := &model.RelationTuple{
		TenantID: tenantID,
		Subject:  "user:hhh",
		Relation: "admin",
		Object:   "system:iii",
	}

	err := repo.Create(ctx, tuple)
	require.NoError(t, err)

	// 检查存在的关系
	exists, err := repo.Check(ctx, tenantID, "user:hhh", "admin", "system:iii")
	require.NoError(t, err)
	assert.True(t, exists)

	// 检查不存在的关系
	exists, err = repo.Check(ctx, tenantID, "user:hhh", "viewer", "system:iii")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestRelationRepo_Expand(t *testing.T) {
	db, repo, tenantID := setupRelationRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	// 创建多个主体与同一客体的关系
	tuples := []*model.RelationTuple{
		{TenantID: tenantID, Subject: "user:jjj", Relation: "viewer", Object: "file:kkk"},
		{TenantID: tenantID, Subject: "user:lll", Relation: "viewer", Object: "file:kkk"},
		{TenantID: tenantID, Subject: "user:mmm", Relation: "editor", Object: "file:kkk"},
	}

	for _, tuple := range tuples {
		err := repo.Create(ctx, tuple)
		require.NoError(t, err)
	}

	// 展开 file:kkk 的 viewer 关系
	subjects, err := repo.Expand(ctx, tenantID, "file:kkk", "viewer")
	require.NoError(t, err)
	assert.Len(t, subjects, 2)
	assert.Contains(t, subjects, "user:jjj")
	assert.Contains(t, subjects, "user:lll")
}

func TestRelationRepo_NamespaceHandling(t *testing.T) {
	db, repo, tenantID := setupRelationRepoTest(t)
	defer db.Exec(context.Background(), "DELETE FROM tenants WHERE id = $1", tenantID)
	defer db.Close()

	ctx := context.Background()

	// 测试不同 namespace 格式
	testCases := []struct {
		name     string
		subject  string
		object   string
		relation string
	}{
		{
			name:     "Standard format",
			subject:  "user:123",
			object:   "document:456",
			relation: "owner",
		},
		{
			name:     "No namespace in subject",
			subject:  "789",
			object:   "file:abc",
			relation: "viewer",
		},
		{
			name:     "Complex namespace",
			subject:  "service:auth:token",
			object:   "resource:api:endpoint",
			relation: "access",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tuple := &model.RelationTuple{
				TenantID: tenantID,
				Subject:  tc.subject,
				Relation: tc.relation,
				Object:   tc.object,
			}

			err := repo.Create(ctx, tuple)
			require.NoError(t, err)

			// 验证可以正确查询
			exists, err := repo.Check(ctx, tenantID, tc.subject, tc.relation, tc.object)
			require.NoError(t, err)
			assert.True(t, exists)

			// 验证可以正确查找
			tuples, err := repo.FindBySubject(ctx, tenantID, tc.subject)
			require.NoError(t, err)
			assert.Greater(t, len(tuples), 0)

			found := false
			for _, t := range tuples {
				if t.Subject == tc.subject && t.Object == tc.object && t.Relation == tc.relation {
					found = true
					break
				}
			}
			assert.True(t, found, "Should find the created tuple")
		})
	}
}
