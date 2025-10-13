package abac

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Mock Repositories
// ============================================================================

// MockPolicyRepository mock 策略仓储
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

// MockUserRepository mock 用户仓储
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
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

// ============================================================================
// Engine Tests
// ============================================================================

// TestEngine_CheckPermission 测试权限检查
func TestEngine_CheckPermission(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	tests := []struct {
		name          string
		resource      string
		action        string
		resourceAttrs map[string]interface{}
		envAttrs      map[string]interface{}
		setupMocks    func(*MockPolicyRepository, *MockUserRepository)
		expectResult  bool
		expectError   bool
	}{
		{
			name:     "简单条件匹配-允许",
			resource: "document",
			action:   "read",
			resourceAttrs: map[string]interface{}{
				"DepartmentID": "IT",
			},
			setupMocks: func(policyRepo *MockPolicyRepository, userRepo *MockUserRepository) {
				user := &model.User{
					ID:       userID,
					Username: "alice",
					Email:    "alice@example.com",
					TenantID: tenantID,
					Status:   model.UserStatusActive,
					Metadata: map[string]interface{}{
						"DepartmentID": "IT",
					},
				}

				policies := []*model.Policy{
					{
						ID:         uuid.New(),
						Expression: "User.DepartmentID == Resource.DepartmentID",
						Effect:     model.PolicyEffectAllow,
						Priority:   10,
						Enabled:    true,
					},
				}

				userRepo.On("FindByID", ctx, userID).Return(user, nil)
				policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "read").Return(policies, nil)
			},
			expectResult: true,
			expectError:  false,
		},
		{
			name:     "简单条件匹配-拒绝",
			resource: "document",
			action:   "read",
			resourceAttrs: map[string]interface{}{
				"DepartmentID": "HR",
			},
			setupMocks: func(policyRepo *MockPolicyRepository, userRepo *MockUserRepository) {
				user := &model.User{
					ID:       userID,
					Username: "alice",
					TenantID: tenantID,
					Metadata: map[string]interface{}{
						"DepartmentID": "IT",
					},
				}

				policies := []*model.Policy{
					{
						Expression: "User.DepartmentID == Resource.DepartmentID",
						Effect:     model.PolicyEffectAllow,
						Priority:   10,
					},
				}

				userRepo.On("FindByID", ctx, userID).Return(user, nil)
				policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "read").Return(policies, nil)
			},
			expectResult: false, // 不匹配，默认拒绝
			expectError:  false,
		},
		{
			name:     "显式拒绝策略",
			resource: "document",
			action:   "delete",
			setupMocks: func(policyRepo *MockPolicyRepository, userRepo *MockUserRepository) {
				user := &model.User{
					ID:       userID,
					TenantID: tenantID,
					Metadata: map[string]interface{}{
						"Role": "viewer",
					},
				}

				policies := []*model.Policy{
					{
						Expression: "User.Role == \"viewer\"",
						Effect:     model.PolicyEffectDeny,
						Priority:   100,
					},
				}

				userRepo.On("FindByID", ctx, userID).Return(user, nil)
				policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "delete").Return(policies, nil)
			},
			expectResult: false,
			expectError:  false,
		},
		{
			name:     "时间限制策略",
			resource: "system",
			action:   "access",
			setupMocks: func(policyRepo *MockPolicyRepository, userRepo *MockUserRepository) {
				user := &model.User{
					ID:       userID,
					TenantID: tenantID,
				}

				// 只允许9-18点访问
				policies := []*model.Policy{
					{
						Expression: "Time.Hour >= 9 && Time.Hour <= 18",
						Effect:     model.PolicyEffectAllow,
						Priority:   10,
					},
				}

				userRepo.On("FindByID", ctx, userID).Return(user, nil)
				policyRepo.On("GetApplicablePolicies", ctx, tenantID, "system", "access").Return(policies, nil)
			},
			// 结果取决于当前时间
			expectError: false,
		},
		{
			name:     "复杂条件-逻辑运算",
			resource: "document",
			action:   "write",
			resourceAttrs: map[string]interface{}{
				"Status":  "draft",
				"OwnerID": userID.String(),
			},
			setupMocks: func(policyRepo *MockPolicyRepository, userRepo *MockUserRepository) {
				user := &model.User{
					ID:       userID,
					TenantID: tenantID,
					Metadata: map[string]interface{}{
						"Level": 3,
					},
				}

				policies := []*model.Policy{
					{
						// 允许：文档草稿且是所有者，或者用户级别>=3
						Expression: "(Resource.Status == \"draft\" && Resource.OwnerID == User.ID) || User.Level >= 3",
						Effect:     model.PolicyEffectAllow,
						Priority:   10,
					},
				}

				userRepo.On("FindByID", ctx, userID).Return(user, nil)
				policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "write").Return(policies, nil)
			},
			expectResult: true, // Level >= 3 满足条件
			expectError:  false,
		},
		{
			name:     "无适用策略-默认拒绝",
			resource: "document",
			action:   "read",
			setupMocks: func(policyRepo *MockPolicyRepository, userRepo *MockUserRepository) {
				user := &model.User{
					ID:       userID,
					TenantID: tenantID,
				}

				userRepo.On("FindByID", ctx, userID).Return(user, nil)
				policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "read").Return([]*model.Policy{}, nil)
			},
			expectResult: false,
			expectError:  false,
		},
		{
			name:     "获取用户失败",
			resource: "document",
			action:   "read",
			setupMocks: func(policyRepo *MockPolicyRepository, userRepo *MockUserRepository) {
				userRepo.On("FindByID", ctx, userID).Return(nil, errors.New("user not found"))
			},
			expectResult: false,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policyRepo := new(MockPolicyRepository)
			userRepo := new(MockUserRepository)

			tt.setupMocks(policyRepo, userRepo)

			engine := NewEngine(policyRepo, userRepo)
			result, err := engine.CheckPermission(ctx, userID, tenantID, tt.resource, tt.action, tt.resourceAttrs, tt.envAttrs)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.name != "时间限制策略" {
					assert.Equal(t, tt.expectResult, result)
				}
			}

			policyRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

// TestEngine_EvaluatePolicy 测试单策略评估
func TestEngine_EvaluatePolicy(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name          string
		policy        *model.Policy
		resourceAttrs map[string]interface{}
		setupMocks    func(*MockUserRepository)
		expectResult  bool
		expectError   bool
	}{
		{
			name: "简单相等判断",
			policy: &model.Policy{
				Expression: "User.Status == \"active\"",
			},
			setupMocks: func(userRepo *MockUserRepository) {
				user := &model.User{
					ID:     userID,
					Status: model.UserStatusActive,
				}
				userRepo.On("FindByID", ctx, userID).Return(user, nil)
			},
			expectResult: true,
			expectError:  false,
		},
		{
			name: "数值比较",
			policy: &model.Policy{
				Expression: "User.Age >= 18",
			},
			setupMocks: func(userRepo *MockUserRepository) {
				user := &model.User{
					ID: userID,
					Metadata: map[string]interface{}{
						"Age": 25,
					},
				}
				userRepo.On("FindByID", ctx, userID).Return(user, nil)
			},
			expectResult: true,
			expectError:  false,
		},
		{
			name: "字符串包含",
			policy: &model.Policy{
				Expression: "User.Email contains \"@example.com\"",
			},
			setupMocks: func(userRepo *MockUserRepository) {
				user := &model.User{
					ID:    userID,
					Email: "alice@example.com",
				}
				userRepo.On("FindByID", ctx, userID).Return(user, nil)
			},
			expectResult: true,
			expectError:  false,
		},
		{
			name: "资源属性访问",
			policy: &model.Policy{
				Expression: "Resource.Visibility == \"public\"",
			},
			resourceAttrs: map[string]interface{}{
				"Visibility": "public",
			},
			setupMocks: func(userRepo *MockUserRepository) {
				user := &model.User{ID: userID}
				userRepo.On("FindByID", ctx, userID).Return(user, nil)
			},
			expectResult: true,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policyRepo := new(MockPolicyRepository)
			userRepo := new(MockUserRepository)

			tt.setupMocks(userRepo)

			engine := NewEngine(policyRepo, userRepo)
			result, err := engine.EvaluatePolicy(ctx, tt.policy, userID, tt.resourceAttrs, nil)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectResult, result)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

// TestEngine_ValidatePolicyExpression 测试表达式验证
func TestEngine_ValidatePolicyExpression(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectError bool
	}{
		{
			name:        "有效表达式-简单条件",
			expression:  "User.Level >= 3",
			expectError: false,
		},
		{
			name:        "有效表达式-复杂逻辑",
			expression:  "(User.DepartmentID == Resource.DepartmentID) && (Time.Hour >= 9 && Time.Hour <= 18)",
			expectError: false,
		},
		{
			name:        "有效表达式-字符串操作",
			expression:  "User.Email contains \"@example.com\"",
			expectError: false,
		},
		{
			name:        "无效表达式-语法错误",
			expression:  "User.Level >= ",
			expectError: true,
		},
		{
			name:        "无效表达式-语法错误2",
			expression:  "User.Level >= && 3",
			expectError: true,
		},
		{
			name:        "空表达式",
			expression:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policyRepo := new(MockPolicyRepository)
			userRepo := new(MockUserRepository)

			engine := NewEngine(policyRepo, userRepo)
			err := engine.ValidatePolicyExpression(tt.expression)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEngine_GetApplicablePolicies 测试获取适用策略
func TestEngine_GetApplicablePolicies(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	t.Run("按优先级排序", func(t *testing.T) {
		policyRepo := new(MockPolicyRepository)
		userRepo := new(MockUserRepository)

		policies := []*model.Policy{
			{ID: uuid.New(), Priority: 10},
			{ID: uuid.New(), Priority: 100}, // 最高优先级
			{ID: uuid.New(), Priority: 5},
		}

		policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "read").Return(policies, nil)

		engine := NewEngine(policyRepo, userRepo)
		result, err := engine.GetApplicablePolicies(ctx, tenantID, "document", "read")

		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, 100, result[0].Priority) // 第一个应该是最高优先级
		assert.Equal(t, 10, result[1].Priority)
		assert.Equal(t, 5, result[2].Priority)
	})
}

// TestEngine_BuildContext 测试上下文构建
func TestEngine_BuildContext(t *testing.T) {
	t.Run("完整用户属性", func(t *testing.T) {
		user := &model.User{
			ID:       uuid.New(),
			Username: "alice",
			Email:    "alice@example.com",
			TenantID: uuid.New(),
			Status:   model.UserStatusActive,
			Metadata: map[string]interface{}{
				"Department": "IT",
				"Level":      3,
			},
		}

		resourceAttrs := map[string]interface{}{
			"Type": "document",
		}

		envAttrs := map[string]interface{}{
			"IP": "192.168.1.1",
		}

		policyRepo := new(MockPolicyRepository)
		userRepo := new(MockUserRepository)
		engine := NewEngine(policyRepo, userRepo)

		ctx := engine.buildContext(user, resourceAttrs, envAttrs)

		assert.Equal(t, user.ID.String(), ctx.User["ID"])
		assert.Equal(t, "alice", ctx.User["Username"])
		assert.Equal(t, "active", ctx.User["Status"])
		assert.Equal(t, "IT", ctx.User["Department"])
		assert.Equal(t, 3, ctx.User["Level"])
		assert.Equal(t, "document", ctx.Resource["Type"])
		assert.Equal(t, "192.168.1.1", ctx.Environment["IP"])
		assert.NotNil(t, ctx.Time["Hour"])
	})

	t.Run("nil属性处理", func(t *testing.T) {
		user := &model.User{
			ID: uuid.New(),
		}

		policyRepo := new(MockPolicyRepository)
		userRepo := new(MockUserRepository)
		engine := NewEngine(policyRepo, userRepo)

		ctx := engine.buildContext(user, nil, nil)

		assert.NotNil(t, ctx.User)
		assert.NotNil(t, ctx.Resource)
		assert.NotNil(t, ctx.Environment)
		assert.NotNil(t, ctx.Time)
	})
}

// TestEngine_EdgeCases 测试边界情况
func TestEngine_EdgeCases(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("表达式错误被跳过", func(t *testing.T) {
		policyRepo := new(MockPolicyRepository)
		userRepo := new(MockUserRepository)

		user := &model.User{ID: userID}

		policies := []*model.Policy{
			{
				Expression: "invalid expression >>>",
				Effect:     model.PolicyEffectAllow,
				Priority:   100,
			},
			{
				Expression: "1 == 1", // 总是true
				Effect:     model.PolicyEffectAllow,
				Priority:   10,
			},
		}

		userRepo.On("FindByID", ctx, userID).Return(user, nil)
		policyRepo.On("GetApplicablePolicies", ctx, tenantID, "doc", "read").Return(policies, nil)

		engine := NewEngine(policyRepo, userRepo)
		result, err := engine.CheckPermission(ctx, userID, tenantID, "doc", "read", nil, nil)

		assert.NoError(t, err)
		assert.True(t, result) // 第二个有效策略生效
	})

	t.Run("高优先级Deny覆盖低优先级Allow", func(t *testing.T) {
		policyRepo := new(MockPolicyRepository)
		userRepo := new(MockUserRepository)

		user := &model.User{ID: userID}

		policies := []*model.Policy{
			{
				Expression: "1 == 1",
				Effect:     model.PolicyEffectDeny,
				Priority:   100, // 高优先级
			},
			{
				Expression: "1 == 1",
				Effect:     model.PolicyEffectAllow,
				Priority:   10, // 低优先级
			},
		}

		userRepo.On("FindByID", ctx, userID).Return(user, nil)
		policyRepo.On("GetApplicablePolicies", ctx, tenantID, "doc", "read").Return(policies, nil)

		engine := NewEngine(policyRepo, userRepo)
		result, err := engine.CheckPermission(ctx, userID, tenantID, "doc", "read", nil, nil)

		assert.NoError(t, err)
		assert.False(t, result) // Deny优先
	})
}

// TestEngine_ComplexScenarios 测试复杂场景
func TestEngine_ComplexScenarios(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("多条件组合策略", func(t *testing.T) {
		policyRepo := new(MockPolicyRepository)
		userRepo := new(MockUserRepository)

		user := &model.User{
			ID:    userID,
			Email: "alice@example.com",
			Metadata: map[string]interface{}{
				"Department": "IT",
				"Level":      5,
				"Verified":   true,
			},
		}

		resourceAttrs := map[string]interface{}{
			"Department":  "IT",
			"Sensitivity": "high",
		}

		policies := []*model.Policy{
			{
				// 复杂条件：同部门 && 级别>=3 && 已验证 && 邮箱包含@example.com
				Expression: "User.Department == Resource.Department && User.Level >= 3 && User.Verified == true && User.Email contains \"@example.com\"",
				Effect:     model.PolicyEffectAllow,
				Priority:   10,
			},
		}

		userRepo.On("FindByID", ctx, userID).Return(user, nil)
		policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "read").Return(policies, nil)

		engine := NewEngine(policyRepo, userRepo)
		result, err := engine.CheckPermission(ctx, userID, tenantID, "document", "read", resourceAttrs, nil)

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("时间窗口限制", func(t *testing.T) {
		policyRepo := new(MockPolicyRepository)
		userRepo := new(MockUserRepository)

		user := &model.User{ID: userID}

		now := time.Now()
		hour := now.Hour()

		policies := []*model.Policy{
			{
				Expression: "Time.Hour >= 0 && Time.Hour <= 23", // 总是满足
				Effect:     model.PolicyEffectAllow,
				Priority:   10,
			},
		}

		userRepo.On("FindByID", ctx, userID).Return(user, nil)
		policyRepo.On("GetApplicablePolicies", ctx, tenantID, "system", "access").Return(policies, nil)

		engine := NewEngine(policyRepo, userRepo)
		result, err := engine.CheckPermission(ctx, userID, tenantID, "system", "access", nil, nil)

		assert.NoError(t, err)
		assert.True(t, result)
		_ = hour // 使用变量避免未使用警告
	})
}
