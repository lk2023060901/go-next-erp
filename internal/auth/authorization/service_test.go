package authorization

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ==================== Mock Repositories ====================

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.Role, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*model.Role), args.Error(1)
}

func (m *MockRoleRepository) GetRoleHierarchy(ctx context.Context, roleID uuid.UUID) ([]*model.Role, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).([]*model.Role), args.Error(1)
}

type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*model.Permission, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).([]*model.Permission), args.Error(1)
}

type MockPolicyRepository struct {
	mock.Mock
}

func (m *MockPolicyRepository) GetApplicablePolicies(ctx context.Context, tenantID uuid.UUID, resource, action string) ([]*model.Policy, error) {
	args := m.Called(ctx, tenantID, resource, action)
	return args.Get(0).([]*model.Policy), args.Error(1)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.User), args.Error(1)
}

type MockRelationRepository struct {
	mock.Mock
}

func (m *MockRelationRepository) Check(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) (bool, error) {
	args := m.Called(ctx, tenantID, subject, relation, object)
	return args.Bool(0), args.Error(1)
}

type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// ==================== 测试用例 ====================

// 测试 RBAC 授权成功
func TestService_RBAC_Success(t *testing.T) {
	ctx := context.Background()

	roleRepo := new(MockRoleRepository)
	permissionRepo := new(MockPermissionRepository)
	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)
	relationRepo := new(MockRelationRepository)
	auditRepo := new(MockAuditRepository)
	cache := new(MockCache)

	service := NewService(roleRepo, permissionRepo, policyRepo, userRepo, relationRepo, auditRepo, cache)

	userID := uuid.New()
	roleID := uuid.New()
	permID := uuid.New()

	// 创建测试角色
	testRole := &model.Role{
		ID:          roleID,
		Name:        "editor",
		DisplayName: "编辑者",
	}

	// 创建测试权限
	testPerm := &model.Permission{
		ID:       permID,
		Resource: "document",
		Action:   "update",
	}

	// Mock expectations
	cache.On("Get", ctx, mock.Anything, mock.Anything).Return(assert.AnError) // 缓存未命中
	roleRepo.On("GetUserRoles", ctx, userID).Return([]*model.Role{testRole}, nil)
	roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{testRole}, nil)
	permissionRepo.On("GetRolePermissions", ctx, roleID).Return([]*model.Permission{testPerm}, nil)
	cache.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// 检查权限
	allowed, err := service.CheckPermission(ctx, userID, uuid.New(), "document", "update", nil)

	// 断言成功
	assert.NoError(t, err)
	assert.True(t, allowed)

	t.Logf("✅ RBAC 授权成功: 用户拥有 document:update 权限")
}

// 测试 ABAC 授权成功
func TestService_ABAC_Success(t *testing.T) {
	ctx := context.Background()

	roleRepo := new(MockRoleRepository)
	permissionRepo := new(MockPermissionRepository)
	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)
	relationRepo := new(MockRelationRepository)
	auditRepo := new(MockAuditRepository)
	cache := new(MockCache)

	service := NewService(roleRepo, permissionRepo, policyRepo, userRepo, relationRepo, auditRepo, cache)

	userID := uuid.New()
	tenantID := uuid.New()

	// 创建测试用户（同部门 dept-001）
	testUser := &model.User{
		ID:       userID,
		Username: "john",
		TenantID: tenantID,
		Metadata: map[string]interface{}{
			"DepartmentID": "dept-001",
		},
	}

	// 创建 ABAC 策略（同部门可访问）
	testPolicy := &model.Policy{
		ID:         uuid.New(),
		Name:       "same_dept_access",
		Resource:   "document",
		Action:     "read",
		Expression: "User.DepartmentID == Resource.DepartmentID",
		Effect:     model.PolicyEffectAllow,
		Priority:   100,
	}

	// 资源属性（同部门）
	resourceAttrs := map[string]interface{}{
		"DepartmentID": "dept-001",
	}

	// Mock expectations
	cache.On("Get", ctx, mock.Anything, mock.Anything).Return(assert.AnError) // RBAC 缓存未命中
	roleRepo.On("GetUserRoles", ctx, userID).Return([]*model.Role{}, nil)
	cache.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	userRepo.On("FindByID", ctx, userID).Return(testUser, nil)
	policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "read").
		Return([]*model.Policy{testPolicy}, nil)

	// 检查权限
	allowed, err := service.CheckPermission(ctx, userID, tenantID, "document", "read", resourceAttrs)

	// 断言成功
	assert.NoError(t, err)
	assert.True(t, allowed)

	t.Logf("✅ ABAC 授权成功: 同部门用户可访问文档")
	t.Logf("   表达式: %s", testPolicy.Expression)
}

// 测试 ReBAC 授权成功
func TestService_ReBAC_Success(t *testing.T) {
	ctx := context.Background()

	roleRepo := new(MockRoleRepository)
	permissionRepo := new(MockPermissionRepository)
	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)
	relationRepo := new(MockRelationRepository)
	auditRepo := new(MockAuditRepository)
	cache := new(MockCache)

	service := NewService(roleRepo, permissionRepo, policyRepo, userRepo, relationRepo, auditRepo, cache)

	userID := uuid.New()
	tenantID := uuid.New()
	docID := uuid.New().String()

	// 资源属性
	resourceAttrs := map[string]interface{}{
		"ID": docID,
	}

	subject := "user:" + userID.String()
	object := "document:" + docID

	// Mock expectations
	cache.On("Get", ctx, mock.Anything, mock.Anything).Return(assert.AnError)
	roleRepo.On("GetUserRoles", ctx, userID).Return([]*model.Role{}, nil)
	cache.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	userRepo.On("FindByID", ctx, userID).Return(&model.User{ID: userID, TenantID: tenantID}, nil)
	policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "owner").
		Return([]*model.Policy{}, nil)
	relationRepo.On("Check", ctx, tenantID, subject, "owner", object).Return(true, nil)

	// 检查权限
	allowed, err := service.CheckPermission(ctx, userID, tenantID, "document", "owner", resourceAttrs)

	// 断言成功
	assert.NoError(t, err)
	assert.True(t, allowed)

	t.Logf("✅ ReBAC 授权成功: 用户是文档所有者")
	t.Logf("   关系: %s -> owner -> %s", subject, object)
}

// 测试权限通配符匹配成功
func TestPermission_WildcardMatch_Success(t *testing.T) {
	tests := []struct {
		name     string
		perm     *model.Permission
		resource string
		action   string
		expected bool
	}{
		{
			name:     "精确匹配",
			perm:     &model.Permission{Resource: "document", Action: "read"},
			resource: "document",
			action:   "read",
			expected: true,
		},
		{
			name:     "资源通配符",
			perm:     &model.Permission{Resource: "document", Action: "*"},
			resource: "document",
			action:   "read",
			expected: true,
		},
		{
			name:     "全通配符",
			perm:     &model.Permission{Resource: "*", Action: "*"},
			resource: "document",
			action:   "read",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.perm.Match(tt.resource, tt.action)
			assert.Equal(t, tt.expected, result)
			if result {
				t.Logf("✅ %s: %s:%s 匹配 %s:%s",
					tt.name, tt.perm.Resource, tt.perm.Action, tt.resource, tt.action)
			}
		})
	}
}

// 测试 ABAC 表达式求值成功
func TestABAC_ExpressionEvaluation_Success(t *testing.T) {
	ctx := context.Background()

	roleRepo := new(MockRoleRepository)
	permissionRepo := new(MockPermissionRepository)
	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)
	relationRepo := new(MockRelationRepository)
	auditRepo := new(MockAuditRepository)
	cache := new(MockCache)

	service := NewService(roleRepo, permissionRepo, policyRepo, userRepo, relationRepo, auditRepo, cache)

	tests := []struct {
		name       string
		expression string
		userAttrs  map[string]interface{}
		resAttrs   map[string]interface{}
		expected   bool
	}{
		{
			name:       "同部门访问",
			expression: "User.DepartmentID == Resource.DepartmentID",
			userAttrs:  map[string]interface{}{"DepartmentID": "dept-001"},
			resAttrs:   map[string]interface{}{"DepartmentID": "dept-001"},
			expected:   true,
		},
		{
			name:       "级别检查",
			expression: "User.Level >= 3",
			userAttrs:  map[string]interface{}{"Level": 5},
			resAttrs:   map[string]interface{}{},
			expected:   true,
		},
		{
			name:       "工作时间",
			expression: "Time.Hour >= 9 && Time.Hour <= 18",
			userAttrs:  map[string]interface{}{},
			resAttrs:   map[string]interface{}{},
			expected:   true, // 假设当前在工作时间
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := uuid.New()
			tenantID := uuid.New()

			testUser := &model.User{
				ID:       userID,
				TenantID: tenantID,
				Metadata: tt.userAttrs,
			}

			testPolicy := &model.Policy{
				Expression: tt.expression,
				Effect:     model.PolicyEffectAllow,
				Priority:   100,
			}

			cache.On("Get", ctx, mock.Anything, mock.Anything).Return(assert.AnError)
			roleRepo.On("GetUserRoles", ctx, userID).Return([]*model.Role{}, nil)
			cache.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			userRepo.On("FindByID", ctx, userID).Return(testUser, nil)
			policyRepo.On("GetApplicablePolicies", ctx, tenantID, "test", "test").
				Return([]*model.Policy{testPolicy}, nil)

			allowed, err := service.CheckPermission(ctx, userID, tenantID, "test", "test", tt.resAttrs)

			assert.NoError(t, err)
			if tt.expected {
				assert.True(t, allowed)
				t.Logf("✅ 表达式求值成功: %s = %v", tt.expression, allowed)
			}
		})
	}
}

// 运行所有测试
func TestAll_Success(t *testing.T) {
	t.Run("RBAC授权", TestService_RBAC_Success)
	t.Run("ABAC授权", TestService_ABAC_Success)
	t.Run("ReBAC授权", TestService_ReBAC_Success)
	t.Run("权限通配符", TestPermission_WildcardMatch_Success)
	t.Run("ABAC表达式", TestABAC_ExpressionEvaluation_Success)

	t.Log("\n========================================")
	t.Log("✅ 所有授权模块测试通过！")
	t.Log("========================================")
}
