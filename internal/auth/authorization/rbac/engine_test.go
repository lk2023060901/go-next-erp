package rbac

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCache is a simple mock cache for testing
type MockCache struct {
	data map[string]interface{}
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]interface{}),
	}
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	if val, ok := m.data[key]; ok {
		// Simple type assertion (in real code, use reflection or json marshal/unmarshal)
		switch v := dest.(type) {
		case *[]*model.Permission:
			if perms, ok := val.([]*model.Permission); ok {
				*v = perms
				return nil
			}
		}
	}
	return errors.New("cache miss")
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	m.data[key] = value
	return nil
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

// ============================================================================
// Mock Repositories
// ============================================================================

// MockRoleRepository mock 角色仓储
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

// MockPermissionRepository mock 权限仓储
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

func (m *MockPermissionRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID, tenantID uuid.UUID) error {
	args := m.Called(ctx, roleID, permissionID, tenantID)
	return args.Error(0)
}

func (m *MockPermissionRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	args := m.Called(ctx, roleID, permissionID)
	return args.Error(0)
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

func (m *MockPermissionRepository) HasPermission(ctx context.Context, roleID, permissionID uuid.UUID) (bool, error) {
	args := m.Called(ctx, roleID, permissionID)
	return args.Bool(0), args.Error(1)
}

// ============================================================================
// Engine Tests
// ============================================================================

// TestEngine_CheckPermission 测试权限检查
func TestEngine_CheckPermission(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()

	tests := []struct {
		name         string
		resource     string
		action       string
		setupMocks   func(*MockRoleRepository, *MockPermissionRepository)
		expectResult bool
		expectError  bool
	}{
		{
			name:     "精确权限匹配成功",
			resource: "document",
			action:   "read",
			setupMocks: func(roleRepo *MockRoleRepository, permRepo *MockPermissionRepository) {
				roles := []*model.Role{
					{ID: roleID, Name: "admin"},
				}
				perms := []*model.Permission{
					{Resource: "document", Action: "read"},
				}

				roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
				roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{}, nil)
				permRepo.On("GetRolePermissions", ctx, roleID).Return(perms, nil)
			},
			expectResult: true,
			expectError:  false,
		},
		{
			name:     "通配符权限匹配成功",
			resource: "document",
			action:   "read",
			setupMocks: func(roleRepo *MockRoleRepository, permRepo *MockPermissionRepository) {
				roles := []*model.Role{
					{ID: roleID, Name: "admin"},
				}
				perms := []*model.Permission{
					{Resource: "*", Action: "*"},
				}

				roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
				roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{}, nil)
				permRepo.On("GetRolePermissions", ctx, roleID).Return(perms, nil)
			},
			expectResult: true,
			expectError:  false,
		},
		{
			name:     "权限不匹配",
			resource: "document",
			action:   "delete",
			setupMocks: func(roleRepo *MockRoleRepository, permRepo *MockPermissionRepository) {
				roles := []*model.Role{
					{ID: roleID, Name: "viewer"},
				}
				perms := []*model.Permission{
					{Resource: "document", Action: "read"},
				}

				roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
				roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{}, nil)
				permRepo.On("GetRolePermissions", ctx, roleID).Return(perms, nil)
			},
			expectResult: false,
			expectError:  false,
		},
		{
			name:     "用户无角色",
			resource: "document",
			action:   "read",
			setupMocks: func(roleRepo *MockRoleRepository, permRepo *MockPermissionRepository) {
				roleRepo.On("GetUserRoles", ctx, userID).Return([]*model.Role{}, nil)
			},
			expectResult: false,
			expectError:  false,
		},
		{
			name:     "获取角色失败",
			resource: "document",
			action:   "read",
			setupMocks: func(roleRepo *MockRoleRepository, permRepo *MockPermissionRepository) {
				roleRepo.On("GetUserRoles", ctx, userID).Return(nil, errors.New("database error"))
			},
			expectResult: false,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleRepo := new(MockRoleRepository)
			permRepo := new(MockPermissionRepository)
			tt.setupMocks(roleRepo, permRepo)

			engine := NewEngine(roleRepo, permRepo, nil)
			result, err := engine.CheckPermission(ctx, userID, tt.resource, tt.action)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectResult, result)
			}

			roleRepo.AssertExpectations(t)
			permRepo.AssertExpectations(t)
		})
	}
}

// TestEngine_GetUserRoles 测试获取用户角色（含继承）
func TestEngine_GetUserRoles(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	role1 := uuid.New()
	role2 := uuid.New()
	parentRole := uuid.New()

	tests := []struct {
		name        string
		setupMocks  func(*MockRoleRepository)
		expectRoles int
		expectError bool
	}{
		{
			name: "获取用户直接角色",
			setupMocks: func(roleRepo *MockRoleRepository) {
				roles := []*model.Role{
					{ID: role1, Name: "developer"},
					{ID: role2, Name: "viewer"},
				}
				roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
				roleRepo.On("GetRoleHierarchy", ctx, role1).Return([]*model.Role{}, nil)
				roleRepo.On("GetRoleHierarchy", ctx, role2).Return([]*model.Role{}, nil)
			},
			expectRoles: 2,
			expectError: false,
		},
		{
			name: "获取角色含父级继承",
			setupMocks: func(roleRepo *MockRoleRepository) {
				directRoles := []*model.Role{
					{ID: role1, Name: "developer"},
				}
				parentRoles := []*model.Role{
					{ID: parentRole, Name: "employee"},
				}

				roleRepo.On("GetUserRoles", ctx, userID).Return(directRoles, nil)
				roleRepo.On("GetRoleHierarchy", ctx, role1).Return(parentRoles, nil)
			},
			expectRoles: 2, // developer + employee
			expectError: false,
		},
		{
			name: "用户无角色",
			setupMocks: func(roleRepo *MockRoleRepository) {
				roleRepo.On("GetUserRoles", ctx, userID).Return([]*model.Role{}, nil)
			},
			expectRoles: 0,
			expectError: false,
		},
		{
			name: "获取角色失败",
			setupMocks: func(roleRepo *MockRoleRepository) {
				roleRepo.On("GetUserRoles", ctx, userID).Return(nil, errors.New("database error"))
			},
			expectRoles: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleRepo := new(MockRoleRepository)
			permRepo := new(MockPermissionRepository)
	
			tt.setupMocks(roleRepo)

			engine := NewEngine(roleRepo, permRepo, nil)
			roles, err := engine.GetUserRoles(ctx, userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, roles, tt.expectRoles)
			}

			roleRepo.AssertExpectations(t)
		})
	}
}

// TestEngine_HasRole 测试角色检查
func TestEngine_HasRole(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()
	otherRoleID := uuid.New()

	tests := []struct {
		name        string
		checkRoleID uuid.UUID
		setupMocks  func(*MockRoleRepository)
		expectHas   bool
		expectError bool
	}{
		{
			name:        "用户拥有直接角色",
			checkRoleID: roleID,
			setupMocks: func(roleRepo *MockRoleRepository) {
				roles := []*model.Role{
					{ID: roleID, Name: "admin"},
				}
				roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
				roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{}, nil)
			},
			expectHas:   true,
			expectError: false,
		},
		{
			name:        "用户拥有继承角色",
			checkRoleID: roleID,
			setupMocks: func(roleRepo *MockRoleRepository) {
				directRoles := []*model.Role{
					{ID: otherRoleID, Name: "developer"},
				}
				parentRoles := []*model.Role{
					{ID: roleID, Name: "employee"},
				}

				roleRepo.On("GetUserRoles", ctx, userID).Return(directRoles, nil)
				roleRepo.On("GetRoleHierarchy", ctx, otherRoleID).Return(parentRoles, nil)
			},
			expectHas:   true,
			expectError: false,
		},
		{
			name:        "用户没有该角色",
			checkRoleID: roleID,
			setupMocks: func(roleRepo *MockRoleRepository) {
				roles := []*model.Role{
					{ID: otherRoleID, Name: "viewer"},
				}
				roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
				roleRepo.On("GetRoleHierarchy", ctx, otherRoleID).Return([]*model.Role{}, nil)
			},
			expectHas:   false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleRepo := new(MockRoleRepository)
			permRepo := new(MockPermissionRepository)
	
			tt.setupMocks(roleRepo)

			engine := NewEngine(roleRepo, permRepo, nil)
			hasRole, err := engine.HasRole(ctx, userID, tt.checkRoleID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectHas, hasRole)
			}

			roleRepo.AssertExpectations(t)
		})
	}
}

// TestEngine_GetUserPermissions 测试获取用户权限
func TestEngine_GetUserPermissions(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	role1 := uuid.New()
	role2 := uuid.New()

	tests := []struct {
		name         string
		setupMocks   func(*MockRoleRepository, *MockPermissionRepository)
		expectPerms  int
		expectError  bool
	}{
		{
			name: "获取单角色权限",
			setupMocks: func(roleRepo *MockRoleRepository, permRepo *MockPermissionRepository) {
				roles := []*model.Role{
					{ID: role1, Name: "admin"},
				}
				perms := []*model.Permission{
					{ID: uuid.New(), Resource: "document", Action: "read"},
					{ID: uuid.New(), Resource: "document", Action: "write"},
				}

				roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
				roleRepo.On("GetRoleHierarchy", ctx, role1).Return([]*model.Role{}, nil)
				permRepo.On("GetRolePermissions", ctx, role1).Return(perms, nil)
			},
			expectPerms: 2,
			expectError: false,
		},
		{
			name: "获取多角色权限（去重）",
			setupMocks: func(roleRepo *MockRoleRepository, permRepo *MockPermissionRepository) {
				roles := []*model.Role{
					{ID: role1, Name: "admin"},
					{ID: role2, Name: "developer"},
				}
				sharedPermID := uuid.New()
				perms1 := []*model.Permission{
					{ID: sharedPermID, Resource: "document", Action: "read"},
				}
				perms2 := []*model.Permission{
					{ID: sharedPermID, Resource: "document", Action: "read"}, // 重复权限
					{ID: uuid.New(), Resource: "code", Action: "write"},
				}

				roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
				roleRepo.On("GetRoleHierarchy", ctx, role1).Return([]*model.Role{}, nil)
				roleRepo.On("GetRoleHierarchy", ctx, role2).Return([]*model.Role{}, nil)
				permRepo.On("GetRolePermissions", ctx, role1).Return(perms1, nil)
				permRepo.On("GetRolePermissions", ctx, role2).Return(perms2, nil)
			},
			expectPerms: 2, // 去重后只有2个唯一权限
			expectError: false,
		},
		{
			name: "用户无角色则无权限",
			setupMocks: func(roleRepo *MockRoleRepository, permRepo *MockPermissionRepository) {
				roleRepo.On("GetUserRoles", ctx, userID).Return([]*model.Role{}, nil)
			},
			expectPerms: 0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleRepo := new(MockRoleRepository)
			permRepo := new(MockPermissionRepository)
			tt.setupMocks(roleRepo, permRepo)

			engine := NewEngine(roleRepo, permRepo, nil)
			perms, err := engine.GetUserPermissions(ctx, userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, perms, tt.expectPerms)
			}

			roleRepo.AssertExpectations(t)
			permRepo.AssertExpectations(t)
		})
	}
}

// TestEngine_CacheScenarios 测试缓存场景
func TestEngine_CacheScenarios(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()

	t.Run("无缓存情况-每次都调用仓储", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		permRepo := new(MockPermissionRepository)

		roles := []*model.Role{
			{ID: roleID, Name: "admin"},
		}
		perms := []*model.Permission{
			{Resource: "document", Action: "read"},
		}

		// 因为没有缓存，每次调用都会访问仓储
		roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil).Times(2)
		roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{}, nil).Times(2)
		permRepo.On("GetRolePermissions", ctx, roleID).Return(perms, nil).Times(2)

		engine := NewEngine(roleRepo, permRepo, nil)

		// 第一次检查
		result1, err1 := engine.CheckPermission(ctx, userID, "document", "read")
		assert.NoError(t, err1)
		assert.True(t, result1)

		// 第二次检查：无缓存，会再次调用仓储
		result2, err2 := engine.CheckPermission(ctx, userID, "document", "read")
		assert.NoError(t, err2)
		assert.True(t, result2)

		roleRepo.AssertExpectations(t)
		permRepo.AssertExpectations(t)
	})
}

// TestEngine_EdgeCases 测试边界情况
func TestEngine_EdgeCases(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()

	t.Run("空资源空操作", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		permRepo := new(MockPermissionRepository)

		roles := []*model.Role{{ID: roleID, Name: "admin"}}
		perms := []*model.Permission{{Resource: "", Action: ""}}

		roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
		roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{}, nil)
		permRepo.On("GetRolePermissions", ctx, roleID).Return(perms, nil)

		engine := NewEngine(roleRepo, permRepo, nil)
		result, err := engine.CheckPermission(ctx, userID, "", "")

		assert.NoError(t, err)
		assert.True(t, result) // 精确匹配
	})

	t.Run("nil角色列表", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		permRepo := new(MockPermissionRepository)

		roleRepo.On("GetUserRoles", ctx, userID).Return(nil, nil)

		engine := NewEngine(roleRepo, permRepo, nil)
		result, err := engine.CheckPermission(ctx, userID, "document", "read")

		assert.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("角色继承失败不影响直接角色", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		permRepo := new(MockPermissionRepository)

		roles := []*model.Role{{ID: roleID, Name: "admin"}}
		perms := []*model.Permission{{Resource: "document", Action: "read"}}

		roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
		roleRepo.On("GetRoleHierarchy", ctx, roleID).Return(nil, errors.New("hierarchy error"))
		permRepo.On("GetRolePermissions", ctx, roleID).Return(perms, nil)

		engine := NewEngine(roleRepo, permRepo, nil)
		result, err := engine.CheckPermission(ctx, userID, "document", "read")

		assert.NoError(t, err)
		assert.True(t, result) // 仍然能使用直接角色权限
	})
}

// TestEngine_ComplexScenarios 测试复杂场景
func TestEngine_ComplexScenarios(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	devRole := uuid.New()
	employeeRole := uuid.New()
	managerRole := uuid.New()

	t.Run("多角色多层级继承", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		permRepo := new(MockPermissionRepository)

		// 用户直接拥有developer和manager角色
		directRoles := []*model.Role{
			{ID: devRole, Name: "developer"},
			{ID: managerRole, Name: "manager"},
		}

		// developer继承employee
		devParents := []*model.Role{
			{ID: employeeRole, Name: "employee"},
		}

		// developer权限
		devPerms := []*model.Permission{
			{ID: uuid.New(), Resource: "code", Action: "write"},
		}

		// manager权限
		managerPerms := []*model.Permission{
			{ID: uuid.New(), Resource: "team", Action: "manage"},
		}

		// employee权限
		employeePerms := []*model.Permission{
			{ID: uuid.New(), Resource: "document", Action: "read"},
		}

		roleRepo.On("GetUserRoles", ctx, userID).Return(directRoles, nil)
		roleRepo.On("GetRoleHierarchy", ctx, devRole).Return(devParents, nil)
		roleRepo.On("GetRoleHierarchy", ctx, managerRole).Return([]*model.Role{}, nil)
		permRepo.On("GetRolePermissions", ctx, devRole).Return(devPerms, nil)
		permRepo.On("GetRolePermissions", ctx, managerRole).Return(managerPerms, nil)
		permRepo.On("GetRolePermissions", ctx, employeeRole).Return(employeePerms, nil)

		engine := NewEngine(roleRepo, permRepo, nil)

		// 应该拥有所有三个角色的权限
		result1, _ := engine.CheckPermission(ctx, userID, "code", "write")      // dev权限
		result2, _ := engine.CheckPermission(ctx, userID, "team", "manage")     // manager权限
		result3, _ := engine.CheckPermission(ctx, userID, "document", "read")   // employee权限（继承）

		assert.True(t, result1)
		assert.True(t, result2)
		assert.True(t, result3)

		// 验证角色数量（3个：developer, manager, employee）
		roles, _ := engine.GetUserRoles(ctx, userID)
		assert.Len(t, roles, 3)
	})
}
