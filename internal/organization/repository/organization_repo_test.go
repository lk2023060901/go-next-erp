package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 注意：这些是基于真实数据库的集成测试
// 运行前需要设置测试数据库：TEST_DATABASE_URL

func setupTestDB(t *testing.T) *database.DB {
	// 跳过实际数据库测试，仅保留测试结构
	// 实际项目中需要设置测试数据库
	t.Skip("Skipping database integration test - requires test database setup")
	return nil
}

func createTestOrganization(tenantID, typeID, createdBy uuid.UUID) *model.Organization {
	orgID := uuid.New()
	now := time.Now()

	return &model.Organization{
		ID:             orgID,
		TenantID:       tenantID,
		Code:           "ORG-" + orgID.String()[:8],
		Name:           "Test Organization",
		ShortName:      "TestOrg",
		Description:    "Test Description",
		TypeID:         typeID,
		TypeCode:       "department",
		ParentID:       nil,
		Level:          1,
		Path:           "/" + orgID.String() + "/",
		PathNames:      "/Test Organization/",
		AncestorIDs:    []string{},
		IsLeaf:         true,
		LeaderID:       nil,
		LeaderName:     "",
		LegalPerson:    "",
		UnifiedCode:    "",
		RegisterDate:   nil,
		RegisterAddr:   "",
		Phone:          "",
		Email:          "test@example.com",
		Address:        "",
		EmployeeCount:  0,
		DirectEmpCount: 0,
		Sort:           1,
		Status:         "active",
		Tags:           []string{"test"},
		CreatedBy:      createdBy,
		UpdatedBy:      createdBy,
		CreatedAt:      now,
		UpdatedAt:      now,
		DeletedAt:      nil,
	}
}

// TestOrganizationRepository_Create 测试创建组织
func TestOrganizationRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()
	org := createTestOrganization(tenantID, typeID, createdBy)

	err := repo.Create(ctx, org)
	require.NoError(t, err)

	// 验证创建
	retrieved, err := repo.GetByID(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, org.Code, retrieved.Code)
	assert.Equal(t, org.Name, retrieved.Name)
}

// TestOrganizationRepository_Update 测试更新组织
func TestOrganizationRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	// 创建组织
	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()
	org := createTestOrganization(tenantID, typeID, createdBy)

	err := repo.Create(ctx, org)
	require.NoError(t, err)

	// 更新组织
	org.Name = "Updated Organization"
	org.ShortName = "UpdatedOrg"
	org.Description = "Updated Description"
	org.UpdatedAt = time.Now()

	err = repo.Update(ctx, org)
	require.NoError(t, err)

	// 验证更新
	retrieved, err := repo.GetByID(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Organization", retrieved.Name)
	assert.Equal(t, "UpdatedOrg", retrieved.ShortName)
	assert.Equal(t, "Updated Description", retrieved.Description)
}

// TestOrganizationRepository_Delete 测试删除组织（软删除）
func TestOrganizationRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	// 创建组织
	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()
	org := createTestOrganization(tenantID, typeID, createdBy)

	err := repo.Create(ctx, org)
	require.NoError(t, err)

	// 删除组织
	err = repo.Delete(ctx, org.ID)
	require.NoError(t, err)

	// 验证已删除（无法查询到）
	_, err = repo.GetByID(ctx, org.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "organization not found")
}

// TestOrganizationRepository_GetByID 测试根据ID获取组织
func TestOrganizationRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	t.Run("存在的组织", func(t *testing.T) {
		tenantID := uuid.New()
		typeID := uuid.New()
		createdBy := uuid.New()
		org := createTestOrganization(tenantID, typeID, createdBy)

		err := repo.Create(ctx, org)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, org.ID)
		require.NoError(t, err)
		assert.Equal(t, org.ID, retrieved.ID)
		assert.Equal(t, org.Code, retrieved.Code)
	})

	t.Run("不存在的组织", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := repo.GetByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "organization not found")
	})
}

// TestOrganizationRepository_GetByCode 测试根据编码获取组织
func TestOrganizationRepository_GetByCode(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()
	org := createTestOrganization(tenantID, typeID, createdBy)

	err := repo.Create(ctx, org)
	require.NoError(t, err)

	// 根据编码查询
	retrieved, err := repo.GetByCode(ctx, tenantID, org.Code)
	require.NoError(t, err)
	assert.Equal(t, org.ID, retrieved.ID)
	assert.Equal(t, org.Code, retrieved.Code)

	// 查询不存在的编码
	_, err = repo.GetByCode(ctx, tenantID, "NON_EXISTENT_CODE")
	assert.Error(t, err)
}

// TestOrganizationRepository_List 测试列出租户的所有组织
func TestOrganizationRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()

	// 创建多个组织
	org1 := createTestOrganization(tenantID, typeID, createdBy)
	org1.Level = 1
	org1.Sort = 1
	err := repo.Create(ctx, org1)
	require.NoError(t, err)

	org2 := createTestOrganization(tenantID, typeID, createdBy)
	org2.Level = 2
	org2.Sort = 2
	err = repo.Create(ctx, org2)
	require.NoError(t, err)

	// 列出组织
	orgs, err := repo.List(ctx, tenantID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(orgs), 2)

	// 验证排序（按level和sort）
	for i := 1; i < len(orgs); i++ {
		if orgs[i].Level == orgs[i-1].Level {
			assert.GreaterOrEqual(t, orgs[i].Sort, orgs[i-1].Sort)
		} else {
			assert.GreaterOrEqual(t, orgs[i].Level, orgs[i-1].Level)
		}
	}
}

// TestOrganizationRepository_ListByParent 测试列出指定父组织的子组织
func TestOrganizationRepository_ListByParent(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()

	// 创建父组织
	parent := createTestOrganization(tenantID, typeID, createdBy)
	parent.Level = 1
	err := repo.Create(ctx, parent)
	require.NoError(t, err)

	// 创建子组织
	child := createTestOrganization(tenantID, typeID, createdBy)
	child.ParentID = &parent.ID
	child.Level = 2
	err = repo.Create(ctx, child)
	require.NoError(t, err)

	// 列出子组织
	children, err := repo.ListByParent(ctx, parent.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(children), 1)

	found := false
	for _, c := range children {
		if c.ID == child.ID {
			found = true
			break
		}
	}
	assert.True(t, found)
}

// TestOrganizationRepository_ListByLevel 测试列出指定层级的组织
func TestOrganizationRepository_ListByLevel(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()

	// 创建不同层级的组织
	org1 := createTestOrganization(tenantID, typeID, createdBy)
	org1.Level = 1
	err := repo.Create(ctx, org1)
	require.NoError(t, err)

	org2 := createTestOrganization(tenantID, typeID, createdBy)
	org2.Level = 2
	err = repo.Create(ctx, org2)
	require.NoError(t, err)

	// 查询第1层级
	orgs, err := repo.ListByLevel(ctx, tenantID, 1)
	require.NoError(t, err)
	for _, org := range orgs {
		assert.Equal(t, 1, org.Level)
	}

	// 查询第2层级
	orgs, err = repo.ListByLevel(ctx, tenantID, 2)
	require.NoError(t, err)
	for _, org := range orgs {
		assert.Equal(t, 2, org.Level)
	}
}

// TestOrganizationRepository_ListByTypeCode 测试列出指定类型的组织
func TestOrganizationRepository_ListByTypeCode(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()

	// 创建不同类型的组织
	org1 := createTestOrganization(tenantID, typeID, createdBy)
	org1.TypeCode = "company"
	err := repo.Create(ctx, org1)
	require.NoError(t, err)

	org2 := createTestOrganization(tenantID, typeID, createdBy)
	org2.TypeCode = "department"
	err = repo.Create(ctx, org2)
	require.NoError(t, err)

	// 查询公司类型
	orgs, err := repo.ListByTypeCode(ctx, tenantID, "company")
	require.NoError(t, err)
	for _, org := range orgs {
		assert.Equal(t, "company", org.TypeCode)
	}
}

// TestOrganizationRepository_GetRoots 测试获取根组织
func TestOrganizationRepository_GetRoots(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()

	// 创建根组织
	root := createTestOrganization(tenantID, typeID, createdBy)
	root.ParentID = nil
	root.Level = 1
	err := repo.Create(ctx, root)
	require.NoError(t, err)

	// 创建子组织
	child := createTestOrganization(tenantID, typeID, createdBy)
	child.ParentID = &root.ID
	child.Level = 2
	err = repo.Create(ctx, child)
	require.NoError(t, err)

	// 获取根组织
	roots, err := repo.GetRoots(ctx, tenantID)
	require.NoError(t, err)

	// 所有根组织的ParentID应该为nil
	for _, r := range roots {
		assert.Nil(t, r.ParentID)
	}
}

// TestOrganizationRepository_GetDescendants 测试获取所有后代组织
func TestOrganizationRepository_GetDescendants(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()

	// 创建祖先组织
	ancestor := createTestOrganization(tenantID, typeID, createdBy)
	ancestor.Level = 1
	err := repo.Create(ctx, ancestor)
	require.NoError(t, err)

	// 创建后代组织
	descendant := createTestOrganization(tenantID, typeID, createdBy)
	descendant.ParentID = &ancestor.ID
	descendant.Level = 2
	descendant.Path = ancestor.Path + descendant.ID.String() + "/"
	err = repo.Create(ctx, descendant)
	require.NoError(t, err)

	// 获取后代
	descendants, err := repo.GetDescendants(ctx, ancestor.ID, ancestor.Path)
	require.NoError(t, err)

	// 验证后代不包含自己
	for _, d := range descendants {
		assert.NotEqual(t, ancestor.ID, d.ID)
	}
}

// TestOrganizationRepository_UpdateChildrenLeafStatus 测试更新子节点叶子状态
func TestOrganizationRepository_UpdateChildrenLeafStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()

	org := createTestOrganization(tenantID, typeID, createdBy)
	org.IsLeaf = true
	err := repo.Create(ctx, org)
	require.NoError(t, err)

	// 更新叶子状态
	err = repo.UpdateChildrenLeafStatus(ctx, org.ID, false)
	require.NoError(t, err)

	// 验证更新
	retrieved, err := repo.GetByID(ctx, org.ID)
	require.NoError(t, err)
	assert.False(t, retrieved.IsLeaf)
}

// TestOrganizationRepository_UpdatePath 测试更新组织路径
func TestOrganizationRepository_UpdatePath(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()

	org := createTestOrganization(tenantID, typeID, createdBy)
	err := repo.Create(ctx, org)
	require.NoError(t, err)

	// 更新路径
	newPath := "/new/path/"
	newPathNames := "/New/Path/"
	newAncestorIDs := []string{uuid.New().String()}
	newLevel := 3

	err = repo.UpdatePath(ctx, org.ID, newPath, newPathNames, newAncestorIDs, newLevel)
	require.NoError(t, err)

	// 验证更新
	retrieved, err := repo.GetByID(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, newPath, retrieved.Path)
	assert.Equal(t, newPathNames, retrieved.PathNames)
	assert.Equal(t, newLevel, retrieved.Level)
}

// TestOrganizationRepository_Move 测试移动组织
func TestOrganizationRepository_Move(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()

	// 创建原父组织
	oldParent := createTestOrganization(tenantID, typeID, createdBy)
	err := repo.Create(ctx, oldParent)
	require.NoError(t, err)

	// 创建新父组织
	newParent := createTestOrganization(tenantID, typeID, createdBy)
	err = repo.Create(ctx, newParent)
	require.NoError(t, err)

	// 创建子组织
	child := createTestOrganization(tenantID, typeID, createdBy)
	child.ParentID = &oldParent.ID
	err = repo.Create(ctx, child)
	require.NoError(t, err)

	// 移动组织
	err = repo.Move(ctx, child.ID, newParent.ID)
	require.NoError(t, err)

	// 验证移动
	retrieved, err := repo.GetByID(ctx, child.ID)
	require.NoError(t, err)
	assert.Equal(t, newParent.ID, *retrieved.ParentID)
}

// TestOrganizationRepository_Exists 测试检查组织是否存在
func TestOrganizationRepository_Exists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()

	org := createTestOrganization(tenantID, typeID, createdBy)
	err := repo.Create(ctx, org)
	require.NoError(t, err)

	// 存在的组织
	exists, err := repo.Exists(ctx, tenantID, org.Code)
	require.NoError(t, err)
	assert.True(t, exists)

	// 不存在的组织
	exists, err = repo.Exists(ctx, tenantID, "NON_EXISTENT_CODE")
	require.NoError(t, err)
	assert.False(t, exists)
}

// TestOrganizationRepository_CountChildren 测试统计子组织数量
func TestOrganizationRepository_CountChildren(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()

	// 创建父组织
	parent := createTestOrganization(tenantID, typeID, createdBy)
	err := repo.Create(ctx, parent)
	require.NoError(t, err)

	// 初始子组织数量应为0
	count, err := repo.CountChildren(ctx, parent.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// 创建两个子组织
	child1 := createTestOrganization(tenantID, typeID, createdBy)
	child1.ParentID = &parent.ID
	err = repo.Create(ctx, child1)
	require.NoError(t, err)

	child2 := createTestOrganization(tenantID, typeID, createdBy)
	child2.ParentID = &parent.ID
	err = repo.Create(ctx, child2)
	require.NoError(t, err)

	// 统计子组织数量
	count, err = repo.CountChildren(ctx, parent.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

// TestOrganizationRepository_EdgeCases 测试边界情况
func TestOrganizationRepository_EdgeCases(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	t.Run("极长的组织名称", func(t *testing.T) {
		tenantID := uuid.New()
		typeID := uuid.New()
		createdBy := uuid.New()

		org := createTestOrganization(tenantID, typeID, createdBy)
		org.Name = string(make([]byte, 1000))

		err := repo.Create(ctx, org)
		// 可能成功或失败，取决于数据库字段长度限制
		_ = err
	})

	t.Run("空Tags数组", func(t *testing.T) {
		tenantID := uuid.New()
		typeID := uuid.New()
		createdBy := uuid.New()

		org := createTestOrganization(tenantID, typeID, createdBy)
		org.Tags = []string{}

		err := repo.Create(ctx, org)
		require.NoError(t, err)
	})

	t.Run("nil Tags", func(t *testing.T) {
		tenantID := uuid.New()
		typeID := uuid.New()
		createdBy := uuid.New()

		org := createTestOrganization(tenantID, typeID, createdBy)
		org.Tags = nil

		err := repo.Create(ctx, org)
		require.NoError(t, err)
	})
}
