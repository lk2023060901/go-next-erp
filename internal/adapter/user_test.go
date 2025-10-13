package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	authv1 "github.com/lk2023060901/go-next-erp/api/auth/v1"
	"github.com/lk2023060901/go-next-erp/internal/auth/authorization"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRoleRepository mocks the RoleRepository interface
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

func (m *MockRoleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.Role, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Role), args.Error(1)
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

// MockAuthorizationService mocks the AuthorizationService interface
type MockAuthorizationService struct {
	mock.Mock
}

func (m *MockAuthorizationService) CheckPermission(ctx context.Context, userID, tenantID uuid.UUID, resource, action string, resourceAttrs map[string]interface{}) (bool, error) {
	args := m.Called(ctx, userID, tenantID, resource, action, resourceAttrs)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthorizationService) CheckPermissionWithAudit(ctx context.Context, userID, tenantID uuid.UUID, resource, action string, resourceAttrs map[string]interface{}, ipAddress, userAgent string) (bool, error) {
	args := m.Called(ctx, userID, tenantID, resource, action, resourceAttrs, ipAddress, userAgent)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthorizationService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Permission), args.Error(1)
}

func (m *MockAuthorizationService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Role), args.Error(1)
}

func (m *MockAuthorizationService) GrantRelation(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) error {
	args := m.Called(ctx, tenantID, subject, relation, object)
	return args.Error(0)
}

func (m *MockAuthorizationService) RevokeRelation(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) error {
	args := m.Called(ctx, tenantID, subject, relation, object)
	return args.Error(0)
}

func (m *MockAuthorizationService) ValidatePolicyExpression(expression string) error {
	args := m.Called(expression)
	return args.Error(0)
}

// TestUserAdapter_ListUsers tests listing users via adapter
func TestUserAdapter_ListUsers(t *testing.T) {
	t.Run("ListUsers successfully", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		var mockAuthzService *authorization.Service = nil // Use real service for now

		adapter := NewUserAdapter(mockUserRepo, mockRoleRepo, mockAuthzService)

		tenantID := uuid.New()
		user1 := &model.User{
			ID:        uuid.New(),
			Username:  "user1",
			Email:     "user1@example.com",
			TenantID:  tenantID,
			Status:    model.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		user2 := &model.User{
			ID:        uuid.New(),
			Username:  "user2",
			Email:     "user2@example.com",
			TenantID:  tenantID,
			Status:    model.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Mock: list users
		mockUserRepo.On("ListByTenant", mock.Anything, tenantID, 20, 0).
			Return([]*model.User{user1, user2}, nil).Once()
		mockUserRepo.On("CountByTenant", mock.Anything, tenantID).
			Return(int64(2), nil).Once()

		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := &authv1.ListUsersRequest{
			Page: &authv1.PageRequest{
				Page:     1,
				PageSize: 20,
			},
		}

		resp, err := adapter.ListUsers(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Users, 2)
		assert.Equal(t, user1.Username, resp.Users[0].Username)
		assert.Equal(t, user2.Username, resp.Users[1].Username)
		assert.Equal(t, int64(2), resp.Page.Total)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("ListUsers with empty result", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		var mockAuthzService *authorization.Service = nil

		adapter := NewUserAdapter(mockUserRepo, mockRoleRepo, mockAuthzService)

		tenantID := uuid.New()

		// Mock: empty list
		mockUserRepo.On("ListByTenant", mock.Anything, tenantID, 20, 0).
			Return([]*model.User{}, nil).Once()
		mockUserRepo.On("CountByTenant", mock.Anything, tenantID).
			Return(int64(0), nil).Once()

		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := &authv1.ListUsersRequest{
			Page: &authv1.PageRequest{
				Page:     1,
				PageSize: 20,
			},
		}

		resp, err := adapter.ListUsers(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Users, 0)
		assert.Equal(t, int64(0), resp.Page.Total)
		mockUserRepo.AssertExpectations(t)
	})
}

// TestUserAdapter_GetUser tests getting a user via adapter
func TestUserAdapter_GetUser(t *testing.T) {
	t.Run("GetUser successfully", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		var mockAuthzService *authorization.Service = nil

		adapter := NewUserAdapter(mockUserRepo, mockRoleRepo, mockAuthzService)

		userID := uuid.New()
		tenantID := uuid.New()
		expectedUser := &model.User{
			ID:        userID,
			Username:  "testuser",
			Email:     "test@example.com",
			TenantID:  tenantID,
			Status:    model.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Mock: find user by ID
		mockUserRepo.On("FindByID", mock.Anything, userID).
			Return(expectedUser, nil).Once()

		req := &authv1.GetUserRequest{
			Id: userID.String(),
		}

		resp, err := adapter.GetUser(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, userID.String(), resp.Id)
		assert.Equal(t, expectedUser.Username, resp.Username)
		assert.Equal(t, expectedUser.Email, resp.Email)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("GetUser with invalid ID", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		var mockAuthzService *authorization.Service = nil

		adapter := NewUserAdapter(mockUserRepo, mockRoleRepo, mockAuthzService)

		req := &authv1.GetUserRequest{
			Id: "invalid-uuid",
		}

		resp, err := adapter.GetUser(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid user id")
	})
}

// TestUserAdapter_UpdateUser tests updating a user via adapter
func TestUserAdapter_UpdateUser(t *testing.T) {
	t.Run("UpdateUser successfully", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		var mockAuthzService *authorization.Service = nil

		adapter := NewUserAdapter(mockUserRepo, mockRoleRepo, mockAuthzService)

		userID := uuid.New()
		tenantID := uuid.New()
		existingUser := &model.User{
			ID:        userID,
			Username:  "testuser",
			Email:     "old@example.com",
			TenantID:  tenantID,
			Status:    model.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Mock: find and update user
		mockUserRepo.On("FindByID", mock.Anything, userID).
			Return(existingUser, nil).Once()
		mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.User")).
			Return(nil).Once()

		req := &authv1.UpdateUserRequest{
			Id:     userID.String(),
			Email:  "new@example.com",
			Status: "inactive",
		}

		resp, err := adapter.UpdateUser(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "new@example.com", resp.Email)
		assert.Equal(t, "inactive", resp.Status)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("UpdateUser with invalid ID", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		var mockAuthzService *authorization.Service = nil

		adapter := NewUserAdapter(mockUserRepo, mockRoleRepo, mockAuthzService)

		req := &authv1.UpdateUserRequest{
			Id:    "invalid-uuid",
			Email: "new@example.com",
		}

		resp, err := adapter.UpdateUser(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid user id")
	})
}

// TestUserAdapter_DeleteUser tests deleting a user via adapter
func TestUserAdapter_DeleteUser(t *testing.T) {
	t.Run("DeleteUser successfully", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		var mockAuthzService *authorization.Service = nil

		adapter := NewUserAdapter(mockUserRepo, mockRoleRepo, mockAuthzService)

		userID := uuid.New()

		// Mock: delete user
		mockUserRepo.On("Delete", mock.Anything, userID).
			Return(nil).Once()

		req := &authv1.DeleteUserRequest{
			Id: userID.String(),
		}

		resp, err := adapter.DeleteUser(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("DeleteUser with invalid ID", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		var mockAuthzService *authorization.Service = nil

		adapter := NewUserAdapter(mockUserRepo, mockRoleRepo, mockAuthzService)

		req := &authv1.DeleteUserRequest{
			Id: "invalid-uuid",
		}

		resp, err := adapter.DeleteUser(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid user id")
	})
}

// TestUserAdapter_GetUserRoles tests getting user roles via adapter
func TestUserAdapter_GetUserRoles(t *testing.T) {
	t.Run("GetUserRoles with invalid ID", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		var mockAuthzService *authorization.Service = nil

		adapter := NewUserAdapter(mockUserRepo, mockRoleRepo, mockAuthzService)

		req := &authv1.GetUserRolesRequest{
			UserId: "invalid-uuid",
		}

		resp, err := adapter.GetUserRoles(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid user id")
	})

	// TODO: Add test with real authorization.Service once we create an interface for it
}

// TestUserAdapter_GetUserPermissions tests getting user permissions via adapter
func TestUserAdapter_GetUserPermissions(t *testing.T) {
	t.Run("GetUserPermissions with invalid ID", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockRoleRepo := new(MockRoleRepository)
		var mockAuthzService *authorization.Service = nil

		adapter := NewUserAdapter(mockUserRepo, mockRoleRepo, mockAuthzService)

		req := &authv1.GetUserPermissionsRequest{
			UserId: "invalid-uuid",
		}

		resp, err := adapter.GetUserPermissions(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid user id")
	})

	// TODO: Add test with real authorization.Service once we create an interface for it
}
