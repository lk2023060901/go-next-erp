package rebac

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Mock Repository
// ============================================================================

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

func (m *MockRelationRepository) Check(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) (bool, error) {
	args := m.Called(ctx, tenantID, subject, relation, object)
	return args.Bool(0), args.Error(1)
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

func (m *MockRelationRepository) FindByRelation(ctx context.Context, tenantID uuid.UUID, subject, object string) ([]*model.RelationTuple, error) {
	args := m.Called(ctx, tenantID, subject, object)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.RelationTuple), args.Error(1)
}

func (m *MockRelationRepository) Expand(ctx context.Context, tenantID uuid.UUID, object, relation string) ([]string, error) {
	args := m.Called(ctx, tenantID, object, relation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRelationRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.RelationTuple, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.RelationTuple), args.Error(1)
}

// ============================================================================
// Engine Tests
// ============================================================================

func TestEngine_Check(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	tests := []struct {
		name         string
		subject      string
		relation     string
		object       string
		setupMock    func(*MockRelationRepository)
		expectResult bool
		expectError  bool
	}{
		{
			name:     "直接关系存在",
			subject:  "user:alice",
			relation: "viewer",
			object:   "document:123",
			setupMock: func(repo *MockRelationRepository) {
				repo.On("Check", ctx, tenantID, "user:alice", "viewer", "document:123").Return(true, nil)
			},
			expectResult: true,
			expectError:  false,
		},
		{
			name:     "关系不存在",
			subject:  "user:bob",
			relation: "editor",
			object:   "document:456",
			setupMock: func(repo *MockRelationRepository) {
				repo.On("Check", ctx, tenantID, "user:bob", "editor", "document:456").Return(false, nil)
			},
			expectResult: false,
			expectError:  false,
		},
		{
			name:     "检查失败",
			subject:  "user:charlie",
			relation: "owner",
			object:   "document:789",
			setupMock: func(repo *MockRelationRepository) {
				repo.On("Check", ctx, tenantID, "user:charlie", "owner", "document:789").Return(false, errors.New("database error"))
			},
			expectResult: false,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRelationRepository)
			tt.setupMock(repo)

			engine := NewEngine(repo)
			result, err := engine.Check(ctx, tenantID, tt.subject, tt.relation, tt.object)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectResult, result)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestEngine_CheckTransitive(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	tests := []struct {
		name         string
		subject      string
		relation     string
		object       string
		setupMock    func(*MockRelationRepository)
		expectResult bool
		expectError  bool
	}{
		{
			name:     "直接关系匹配",
			subject:  "user:alice",
			relation: "viewer",
			object:   "document:123",
			setupMock: func(repo *MockRelationRepository) {
				repo.On("Check", ctx, tenantID, "user:alice", "viewer", "document:123").Return(true, nil)
			},
			expectResult: true,
			expectError:  false,
		},
		{
			name:     "继承关系-owner继承viewer",
			subject:  "user:bob",
			relation: "viewer",
			object:   "document:456",
			setupMock: func(repo *MockRelationRepository) {
				repo.On("Check", ctx, tenantID, "user:bob", "viewer", "document:456").Return(false, nil)
				tuples := []*model.RelationTuple{
					{Subject: "user:bob", Relation: "owner", Object: "document:456"},
				}
				repo.On("FindByRelation", ctx, tenantID, "user:bob", "document:456").Return(tuples, nil)
			},
			expectResult: true,
			expectError:  false,
		},
		{
			name:     "继承关系-editor继承viewer",
			subject:  "user:charlie",
			relation: "viewer",
			object:   "document:789",
			setupMock: func(repo *MockRelationRepository) {
				repo.On("Check", ctx, tenantID, "user:charlie", "viewer", "document:789").Return(false, nil)
				tuples := []*model.RelationTuple{
					{Subject: "user:charlie", Relation: "editor", Object: "document:789"},
				}
				repo.On("FindByRelation", ctx, tenantID, "user:charlie", "document:789").Return(tuples, nil)
			},
			expectResult: true,
			expectError:  false,
		},
		{
			name:     "无继承关系",
			subject:  "user:dave",
			relation: "owner",
			object:   "document:999",
			setupMock: func(repo *MockRelationRepository) {
				repo.On("Check", ctx, tenantID, "user:dave", "owner", "document:999").Return(false, nil)
				tuples := []*model.RelationTuple{
					{Subject: "user:dave", Relation: "viewer", Object: "document:999"},
				}
				repo.On("FindByRelation", ctx, tenantID, "user:dave", "document:999").Return(tuples, nil)
			},
			expectResult: false,
			expectError:  false,
		},
		{
			name:     "无任何关系",
			subject:  "user:eve",
			relation: "viewer",
			object:   "document:111",
			setupMock: func(repo *MockRelationRepository) {
				repo.On("Check", ctx, tenantID, "user:eve", "viewer", "document:111").Return(false, nil)
				repo.On("FindByRelation", ctx, tenantID, "user:eve", "document:111").Return([]*model.RelationTuple{}, nil)
			},
			expectResult: false,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRelationRepository)
			tt.setupMock(repo)

			engine := NewEngine(repo)
			result, err := engine.CheckTransitive(ctx, tenantID, tt.subject, tt.relation, tt.object)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectResult, result)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestEngine_Expand(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	tests := []struct {
		name          string
		object        string
		relation      string
		setupMock     func(*MockRelationRepository)
		expectSubjects []string
		expectError   bool
	}{
		{
			name:     "展开多个主体",
			object:   "document:123",
			relation: "viewer",
			setupMock: func(repo *MockRelationRepository) {
				subjects := []string{"user:alice", "user:bob", "user:charlie"}
				repo.On("Expand", ctx, tenantID, "document:123", "viewer").Return(subjects, nil)
			},
			expectSubjects: []string{"user:alice", "user:bob", "user:charlie"},
			expectError:    false,
		},
		{
			name:     "无主体",
			object:   "document:456",
			relation: "owner",
			setupMock: func(repo *MockRelationRepository) {
				repo.On("Expand", ctx, tenantID, "document:456", "owner").Return([]string{}, nil)
			},
			expectSubjects: []string{},
			expectError:    false,
		},
		{
			name:     "查询失败",
			object:   "document:789",
			relation: "editor",
			setupMock: func(repo *MockRelationRepository) {
				repo.On("Expand", ctx, tenantID, "document:789", "editor").Return(nil, errors.New("database error"))
			},
			expectSubjects: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRelationRepository)
			tt.setupMock(repo)

			engine := NewEngine(repo)
			subjects, err := engine.Expand(ctx, tenantID, tt.object, tt.relation)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectSubjects, subjects)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestEngine_ListUserObjects(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	userID := "alice"

	tests := []struct {
		name          string
		relation      string
		objectType    string
		setupMock     func(*MockRelationRepository)
		expectObjects []string
		expectError   bool
	}{
		{
			name:       "列出用户的所有文档（viewer权限）",
			relation:   "viewer",
			objectType: "document",
			setupMock: func(repo *MockRelationRepository) {
				tuples := []*model.RelationTuple{
					{Subject: "user:alice", Relation: "viewer", Object: "document:123"},
					{Subject: "user:alice", Relation: "viewer", Object: "document:456"},
					{Subject: "user:alice", Relation: "editor", Object: "document:789"}, // 不同关系
				}
				repo.On("FindBySubject", ctx, tenantID, "user:alice").Return(tuples, nil)
			},
			expectObjects: []string{"document:123", "document:456"},
			expectError:   false,
		},
		{
			name:       "不限定对象类型",
			relation:   "owner",
			objectType: "",
			setupMock: func(repo *MockRelationRepository) {
				tuples := []*model.RelationTuple{
					{Subject: "user:alice", Relation: "owner", Object: "document:123"},
					{Subject: "user:alice", Relation: "owner", Object: "folder:456"},
				}
				repo.On("FindBySubject", ctx, tenantID, "user:alice").Return(tuples, nil)
			},
			expectObjects: []string{"document:123", "folder:456"},
			expectError:   false,
		},
		{
			name:       "用户无对象",
			relation:   "viewer",
			objectType: "document",
			setupMock: func(repo *MockRelationRepository) {
				repo.On("FindBySubject", ctx, tenantID, "user:alice").Return([]*model.RelationTuple{}, nil)
			},
			expectObjects: nil,
			expectError:   false,
		},
		{
			name:       "查询失败",
			relation:   "viewer",
			objectType: "document",
			setupMock: func(repo *MockRelationRepository) {
				repo.On("FindBySubject", ctx, tenantID, "user:alice").Return(nil, errors.New("database error"))
			},
			expectObjects: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRelationRepository)
			tt.setupMock(repo)

			engine := NewEngine(repo)
			objects, err := engine.ListUserObjects(ctx, tenantID, userID, tt.relation, tt.objectType)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectObjects, objects)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestEngine_Grant(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	t.Run("授予关系成功", func(t *testing.T) {
		repo := new(MockRelationRepository)
		repo.On("Create", ctx, mock.MatchedBy(func(tuple *model.RelationTuple) bool {
			return tuple.TenantID == tenantID &&
				tuple.Subject == "user:alice" &&
				tuple.Relation == "viewer" &&
				tuple.Object == "document:123"
		})).Return(nil)

		engine := NewEngine(repo)
		err := engine.Grant(ctx, tenantID, "user:alice", "viewer", "document:123")

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("授予关系失败", func(t *testing.T) {
		repo := new(MockRelationRepository)
		repo.On("Create", ctx, mock.Anything).Return(errors.New("duplicate relation"))

		engine := NewEngine(repo)
		err := engine.Grant(ctx, tenantID, "user:bob", "editor", "document:456")

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}

func TestEngine_Revoke(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	t.Run("撤销关系成功", func(t *testing.T) {
		repo := new(MockRelationRepository)
		repo.On("DeleteByTuple", ctx, tenantID, "user:alice", "viewer", "document:123").Return(nil)

		engine := NewEngine(repo)
		err := engine.Revoke(ctx, tenantID, "user:alice", "viewer", "document:123")

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("撤销关系失败", func(t *testing.T) {
		repo := new(MockRelationRepository)
		repo.On("DeleteByTuple", ctx, tenantID, "user:bob", "editor", "document:456").Return(errors.New("not found"))

		engine := NewEngine(repo)
		err := engine.Revoke(ctx, tenantID, "user:bob", "editor", "document:456")

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}

func TestEngine_ComplexScenarios(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	t.Run("多层继承关系", func(t *testing.T) {
		repo := new(MockRelationRepository)

		// user:alice 是 owner，应该继承 editor 和 viewer
		repo.On("Check", ctx, tenantID, "user:alice", "viewer", "document:123").Return(false, nil)
		tuples := []*model.RelationTuple{
			{Subject: "user:alice", Relation: "owner", Object: "document:123"},
		}
		repo.On("FindByRelation", ctx, tenantID, "user:alice", "document:123").Return(tuples, nil)

		engine := NewEngine(repo)
		result, err := engine.CheckTransitive(ctx, tenantID, "user:alice", "viewer", "document:123")

		assert.NoError(t, err)
		assert.True(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("组合关系场景", func(t *testing.T) {
		repo := new(MockRelationRepository)

		// 1. 授予关系
		repo.On("Create", ctx, mock.Anything).Return(nil).Once()

		// 2. 检查关系
		repo.On("Check", ctx, tenantID, "user:alice", "viewer", "document:123").Return(true, nil).Once()

		// 3. 撤销关系
		repo.On("DeleteByTuple", ctx, tenantID, "user:alice", "viewer", "document:123").Return(nil).Once()

		engine := NewEngine(repo)

		// 授予
		err := engine.Grant(ctx, tenantID, "user:alice", "viewer", "document:123")
		assert.NoError(t, err)

		// 检查
		exists, err := engine.Check(ctx, tenantID, "user:alice", "viewer", "document:123")
		assert.NoError(t, err)
		assert.True(t, exists)

		// 撤销
		err = engine.Revoke(ctx, tenantID, "user:alice", "viewer", "document:123")
		assert.NoError(t, err)

		repo.AssertExpectations(t)
	})
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		prefix   string
		expected bool
	}{
		{"完全匹配", "document:123", "document:", true},
		{"前缀匹配", "document:123", "doc", true},
		{"不匹配", "folder:456", "document:", false},
		{"空前缀", "anything", "", true},
		{"空字符串", "", "prefix", false},
		{"相等", "same", "same", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasPrefix(tt.s, tt.prefix)
			assert.Equal(t, tt.expected, result)
		})
	}
}
