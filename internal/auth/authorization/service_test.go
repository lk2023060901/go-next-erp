package authorization

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ==================== Mock Repositories ====================

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleRepository) FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*model.Role, error) {
	args := m.Called(ctx, tenantID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Role), args.Error(1)
}

func (m *MockRoleRepository) GetRoleHierarchy(ctx context.Context, roleID uuid.UUID) ([]*model.Role, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Role), args.Error(1)
}

func (m *MockRoleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.Role, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Role), args.Error(1)
}

func (m *MockRoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID, tenantID uuid.UUID) error {
	args := m.Called(ctx, userID, roleID, tenantID)
	return args.Error(0)
}

func (m *MockRoleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockRoleRepository) HasRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, roleID)
	return args.Bool(0), args.Error(1)
}

type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) Create(ctx context.Context, permission *model.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *MockPermissionRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

func (m *MockPermissionRepository) FindByResourceAction(ctx context.Context, tenantID uuid.UUID, resource, action string) (*model.Permission, error) {
	args := m.Called(ctx, tenantID, resource, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Permission), args.Error(1)
}

func (m *MockPermissionRepository) Update(ctx context.Context, permission *model.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *MockPermissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPermissionRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*model.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Permission), args.Error(1)
}

func (m *MockPermissionRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.Permission, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Permission), args.Error(1)
}

func (m *MockPermissionRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID, tenantID uuid.UUID) error {
	args := m.Called(ctx, roleID, permissionID, tenantID)
	return args.Error(0)
}

func (m *MockPermissionRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
}

func (m *MockPermissionRepository) HasPermission(ctx context.Context, roleID, permissionID uuid.UUID) (bool, error) {
	args := m.Called(ctx, roleID, permissionID)
	return args.Bool(0), args.Error(1)
}

type MockPolicyRepository struct {
	mock.Mock
}

func (m *MockPolicyRepository) Create(ctx context.Context, policy *model.Policy) error {
	args := m.Called(ctx, policy)
	return args.Error(0)
}

func (m *MockPolicyRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Policy, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Policy), args.Error(1)
}

func (m *MockPolicyRepository) Update(ctx context.Context, policy *model.Policy) error {
	args := m.Called(ctx, policy)
	return args.Error(0)
}

func (m *MockPolicyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPolicyRepository) GetApplicablePolicies(ctx context.Context, tenantID uuid.UUID, resource, action string) ([]*model.Policy, error) {
	args := m.Called(ctx, tenantID, resource, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Policy), args.Error(1)
}

func (m *MockPolicyRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.Policy, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Policy), args.Error(1)
}

func (m *MockPolicyRepository) EnablePolicy(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPolicyRepository) DisablePolicy(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.User, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *MockUserRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, ip string) error {
	args := m.Called(ctx, userID, ip)
	return args.Error(0)
}

func (m *MockUserRepository) IncrementLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) ResetLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) LockUser(ctx context.Context, userID uuid.UUID, until time.Time) error {
	args := m.Called(ctx, userID, until)
	return args.Error(0)
}

func (m *MockUserRepository) UnlockUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) ListUsersByRole(ctx context.Context, roleID uuid.UUID) ([]*model.User, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

type MockRelationRepository struct {
	mock.Mock
}

func (m *MockRelationRepository) Create(ctx context.Context, tuple *model.RelationTuple) error {
	args := m.Called(ctx, tuple)
	return args.Error(0)
}

func (m *MockRelationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRelationRepository) DeleteByTuple(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) error {
	args := m.Called(ctx, tenantID, subject, relation, object)
	return args.Error(0)
}

func (m *MockRelationRepository) FindBySubject(ctx context.Context, tenantID uuid.UUID, subject string) ([]*model.RelationTuple, error) {
	args := m.Called(ctx, tenantID, subject)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.RelationTuple), args.Error(1)
}

func (m *MockRelationRepository) FindByObject(ctx context.Context, tenantID uuid.UUID, object string) ([]*model.RelationTuple, error) {
	args := m.Called(ctx, tenantID, object)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.RelationTuple), args.Error(1)
}

func (m *MockRelationRepository) FindByRelation(ctx context.Context, tenantID uuid.UUID, subject, relation string) ([]*model.RelationTuple, error) {
	args := m.Called(ctx, tenantID, subject, relation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.RelationTuple), args.Error(1)
}

func (m *MockRelationRepository) Check(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) (bool, error) {
	args := m.Called(ctx, tenantID, subject, relation, object)
	return args.Bool(0), args.Error(1)
}

func (m *MockRelationRepository) Expand(ctx context.Context, tenantID uuid.UUID, object, relation string) ([]string, error) {
	args := m.Called(ctx, tenantID, object, relation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) FindByEventID(ctx context.Context, eventID string) (*model.AuditLog, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.AuditLog, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.AuditLog, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) ListByAction(ctx context.Context, tenantID uuid.UUID, action string, limit, offset int) ([]*model.AuditLog, error) {
	args := m.Called(ctx, tenantID, action, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) ListByTimeRange(ctx context.Context, tenantID uuid.UUID, start, end time.Time, limit, offset int) ([]*model.AuditLog, error) {
	args := m.Called(ctx, tenantID, start, end, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuditRepository) CountByAction(ctx context.Context, tenantID uuid.UUID, action string) (int64, error) {
	args := m.Called(ctx, tenantID, action)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuditRepository) CleanupOldLogs(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *MockAuditRepository) ListByActionWithCursor(ctx context.Context, tenantID uuid.UUID, action string, cursor *time.Time, limit int) ([]*model.AuditLog, *time.Time, bool, error) {
	args := m.Called(ctx, tenantID, action, cursor, limit)
	if args.Get(0) == nil {
		return nil, nil, false, args.Error(3)
	}
	nextCursor := args.Get(1)
	hasNext := args.Bool(2)
	if nextCursor == nil {
		return args.Get(0).([]*model.AuditLog), nil, hasNext, args.Error(3)
	}
	return args.Get(0).([]*model.AuditLog), nextCursor.(*time.Time), hasNext, args.Error(3)
}

func (m *MockAuditRepository) ListByUserWithCursor(ctx context.Context, userID uuid.UUID, cursor *time.Time, limit int) ([]*model.AuditLog, *time.Time, bool, error) {
	args := m.Called(ctx, userID, cursor, limit)
	if args.Get(0) == nil {
		return nil, nil, false, args.Error(3)
	}
	nextCursor := args.Get(1)
	hasNext := args.Bool(2)
	if nextCursor == nil {
		return args.Get(0).([]*model.AuditLog), nil, hasNext, args.Error(3)
	}
	return args.Get(0).([]*model.AuditLog), nextCursor.(*time.Time), hasNext, args.Error(3)
}

func (m *MockAuditRepository) ListByTenantWithCursor(ctx context.Context, tenantID uuid.UUID, cursor *time.Time, limit int) ([]*model.AuditLog, *time.Time, bool, error) {
	args := m.Called(ctx, tenantID, cursor, limit)
	if args.Get(0) == nil {
		return nil, nil, false, args.Error(3)
	}
	nextCursor := args.Get(1)
	hasNext := args.Bool(2)
	if nextCursor == nil {
		return args.Get(0).([]*model.AuditLog), nil, hasNext, args.Error(3)
	}
	return args.Get(0).([]*model.AuditLog), nextCursor.(*time.Time), hasNext, args.Error(3)
}

// Note: We cannot easily mock *cache.Cache since it's a concrete type.
// Instead, we'll use dependency injection pattern or skip cache in tests.

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

	service := NewService(roleRepo, permissionRepo, policyRepo, userRepo, relationRepo, auditRepo, nil)

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
	roleRepo.On("GetUserRoles", ctx, userID).Return([]*model.Role{testRole}, nil)
	roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{testRole}, nil)
	permissionRepo.On("GetRolePermissions", ctx, roleID).Return([]*model.Permission{testPerm}, nil)

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

	service := NewService(roleRepo, permissionRepo, policyRepo, userRepo, relationRepo, auditRepo, nil)

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
	roleRepo.On("GetUserRoles", ctx, userID).Return([]*model.Role{}, nil)
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

	service := NewService(roleRepo, permissionRepo, policyRepo, userRepo, relationRepo, auditRepo, nil)

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
	roleRepo.On("GetUserRoles", ctx, userID).Return([]*model.Role{}, nil)
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

	service := NewService(roleRepo, permissionRepo, policyRepo, userRepo, relationRepo, auditRepo, nil)

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
			name:       "时间检查",
			expression: "Time.Hour >= 0 && Time.Hour <= 23", // 任何时间都通过
			userAttrs:  map[string]interface{}{},
			resAttrs:   map[string]interface{}{},
			expected:   true,
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

			roleRepo.On("GetUserRoles", ctx, userID).Return([]*model.Role{}, nil)
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

// TestService_CheckPermissionWithAudit 测试带审计的权限检查
func TestService_CheckPermissionWithAudit(t *testing.T) {
	roleRepo := new(MockRoleRepository)
	permissionRepo := new(MockPermissionRepository)
	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)
	relationRepo := new(MockRelationRepository)
	auditRepo := new(MockAuditRepository)

	service := NewService(roleRepo, permissionRepo, policyRepo, userRepo, relationRepo, auditRepo, nil)

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("权限允许带审计", func(t *testing.T) {
		role := &model.Role{
			ID:   uuid.New(),
			Name: "admin",
		}

		permission := &model.Permission{
			ID:       uuid.New(),
			Resource: "document",
			Action:   "read",
		}

		roleRepo.On("GetUserRoles", ctx, userID).Return([]*model.Role{role}, nil).Once()
		roleRepo.On("GetRoleHierarchy", ctx, role.ID).Return([]*model.Role{}, nil).Once()
		permissionRepo.On("GetRolePermissions", ctx, role.ID).Return([]*model.Permission{permission}, nil).Once()
		auditRepo.On("Create", ctx, mock.AnythingOfType("*model.AuditLog")).Return(nil).Once()

		allowed, err := service.CheckPermissionWithAudit(ctx, userID, tenantID, "document", "read", nil, "127.0.0.1", "test-agent")
		assert.NoError(t, err)
		assert.True(t, allowed)

		roleRepo.AssertExpectations(t)
		permissionRepo.AssertExpectations(t)
		auditRepo.AssertExpectations(t)
	})

	t.Run("权限拒绝带审计", func(t *testing.T) {
		user := &model.User{
			ID:       userID,
			TenantID: tenantID,
		}

		roleRepo.On("GetUserRoles", ctx, userID).Return([]*model.Role{}, nil).Once()
		userRepo.On("FindByID", ctx, userID).Return(user, nil).Once()
		policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "delete").Return([]*model.Policy{}, nil).Once()
		auditRepo.On("Create", ctx, mock.AnythingOfType("*model.AuditLog")).Return(nil).Once()

		allowed, err := service.CheckPermissionWithAudit(ctx, userID, tenantID, "document", "delete", nil, "127.0.0.1", "test-agent")
		assert.NoError(t, err)
		assert.False(t, allowed)

		roleRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
		policyRepo.AssertExpectations(t)
		auditRepo.AssertExpectations(t)
	})
}

// TestService_GetUserPermissions 测试获取用户权限
func TestService_GetUserPermissions(t *testing.T) {
	roleRepo := new(MockRoleRepository)
	permissionRepo := new(MockPermissionRepository)
	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)
	relationRepo := new(MockRelationRepository)
	auditRepo := new(MockAuditRepository)

	service := NewService(roleRepo, permissionRepo, policyRepo, userRepo, relationRepo, auditRepo, nil)

	ctx := context.Background()
	userID := uuid.New()

	role := &model.Role{
		ID:   uuid.New(),
		Name: "editor",
	}

	permissions := []*model.Permission{
		{ID: uuid.New(), Resource: "document", Action: "read"},
		{ID: uuid.New(), Resource: "document", Action: "write"},
	}

	roleRepo.On("GetUserRoles", ctx, userID).Return([]*model.Role{role}, nil).Once()
	roleRepo.On("GetRoleHierarchy", ctx, role.ID).Return([]*model.Role{}, nil).Once()
	permissionRepo.On("GetRolePermissions", ctx, role.ID).Return(permissions, nil).Once()

	result, err := service.GetUserPermissions(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	roleRepo.AssertExpectations(t)
	permissionRepo.AssertExpectations(t)
}

// TestService_GetUserRoles 测试获取用户角色
func TestService_GetUserRoles(t *testing.T) {
	roleRepo := new(MockRoleRepository)
	permissionRepo := new(MockPermissionRepository)
	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)
	relationRepo := new(MockRelationRepository)
	auditRepo := new(MockAuditRepository)

	service := NewService(roleRepo, permissionRepo, policyRepo, userRepo, relationRepo, auditRepo, nil)

	ctx := context.Background()
	userID := uuid.New()

	roles := []*model.Role{
		{ID: uuid.New(), Name: "admin"},
		{ID: uuid.New(), Name: "editor"},
	}

	roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil).Once()
	for _, role := range roles {
		roleRepo.On("GetRoleHierarchy", ctx, role.ID).Return([]*model.Role{}, nil).Once()
	}

	result, err := service.GetUserRoles(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// 验证返回了正确的角色（不依赖顺序）
	roleNames := []string{result[0].Name, result[1].Name}
	assert.Contains(t, roleNames, "admin")
	assert.Contains(t, roleNames, "editor")

	roleRepo.AssertExpectations(t)
}

// TestService_GrantRelation 测试授予关系
func TestService_GrantRelation(t *testing.T) {
	roleRepo := new(MockRoleRepository)
	permissionRepo := new(MockPermissionRepository)
	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)
	relationRepo := new(MockRelationRepository)
	auditRepo := new(MockAuditRepository)

	service := NewService(roleRepo, permissionRepo, policyRepo, userRepo, relationRepo, auditRepo, nil)

	ctx := context.Background()
	tenantID := uuid.New()

	tuple := &model.RelationTuple{
		Subject:  "user:alice",
		Relation: "viewer",
		Object:   "document:123",
	}

	relationRepo.On("Create", ctx, mock.MatchedBy(func(t *model.RelationTuple) bool {
		return t.Subject == "user:alice" && t.Relation == "viewer" && t.Object == "document:123"
	})).Return(nil).Once()

	err := service.GrantRelation(ctx, tenantID, tuple.Subject, tuple.Relation, tuple.Object)
	assert.NoError(t, err)

	relationRepo.AssertExpectations(t)
}

// TestService_RevokeRelation 测试撤销关系
func TestService_RevokeRelation(t *testing.T) {
	roleRepo := new(MockRoleRepository)
	permissionRepo := new(MockPermissionRepository)
	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)
	relationRepo := new(MockRelationRepository)
	auditRepo := new(MockAuditRepository)

	service := NewService(roleRepo, permissionRepo, policyRepo, userRepo, relationRepo, auditRepo, nil)

	ctx := context.Background()
	tenantID := uuid.New()

	relationRepo.On("DeleteByTuple", ctx, tenantID, "user:alice", "viewer", "document:123").Return(nil).Once()

	err := service.RevokeRelation(ctx, tenantID, "user:alice", "viewer", "document:123")
	assert.NoError(t, err)

	relationRepo.AssertExpectations(t)
}

// TestService_ValidatePolicyExpression 测试验证策略表达式
func TestService_ValidatePolicyExpression(t *testing.T) {
	roleRepo := new(MockRoleRepository)
	permissionRepo := new(MockPermissionRepository)
	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)
	relationRepo := new(MockRelationRepository)
	auditRepo := new(MockAuditRepository)

	service := NewService(roleRepo, permissionRepo, policyRepo, userRepo, relationRepo, auditRepo, nil)

	t.Run("有效表达式", func(t *testing.T) {
		err := service.ValidatePolicyExpression("User.Level >= 5")
		assert.NoError(t, err)
	})

	t.Run("无效表达式", func(t *testing.T) {
		err := service.ValidatePolicyExpression("User.Level >>>= 5")
		assert.Error(t, err)
	})
}

// 运行所有测试
func TestAll_Success(t *testing.T) {
	t.Run("RBAC授权", TestService_RBAC_Success)
	t.Run("ABAC授权", TestService_ABAC_Success)
	t.Run("ReBAC授权", TestService_ReBAC_Success)
	t.Run("权限通配符", TestPermission_WildcardMatch_Success)
	t.Run("ABAC表达式", TestABAC_ExpressionEvaluation_Success)
	t.Run("带审计权限检查", TestService_CheckPermissionWithAudit)
	t.Run("获取用户权限", TestService_GetUserPermissions)
	t.Run("获取用户角色", TestService_GetUserRoles)
	t.Run("授予关系", TestService_GrantRelation)
	t.Run("撤销关系", TestService_RevokeRelation)
	t.Run("验证策略表达式", TestService_ValidatePolicyExpression)

	t.Log("\n========================================")
	t.Log("✅ 所有授权模块测试通过！")
	t.Log("========================================")
}
