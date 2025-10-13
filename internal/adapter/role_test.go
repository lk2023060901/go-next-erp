package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	authv1 "github.com/lk2023060901/go-next-erp/api/auth/v1"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPermissionRepository mocks the PermissionRepository interface
type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) Create(ctx context.Context, perm *model.Permission) error {
	args := m.Called(ctx, perm)
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

func (m *MockPermissionRepository) ListAll(ctx context.Context) ([]*model.Permission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Permission), args.Error(1)
}

func (m *MockPermissionRepository) Update(ctx context.Context, perm *model.Permission) error {
	args := m.Called(ctx, perm)
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

func (m *MockPermissionRepository) HasPermission(ctx context.Context, roleID, permissionID uuid.UUID) (bool, error) {
	args := m.Called(ctx, roleID, permissionID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*model.Permission, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Permission), args.Error(1)
}

// TestRoleAdapter_ListRoles tests listing roles via adapter
func TestRoleAdapter_ListRoles(t *testing.T) {
	t.Run("ListRoles successfully", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		tenantID := uuid.New()
		role1 := &model.Role{
			ID:          uuid.New(),
			Name:        "admin",
			DisplayName: "Administrator",
			TenantID:    tenantID,
			CreatedAt:   time.Now(),
		}
		role2 := &model.Role{
			ID:          uuid.New(),
			Name:        "user",
			DisplayName: "User",
			TenantID:    tenantID,
			CreatedAt:   time.Now(),
		}

		// Mock: list roles
		mockRoleRepo.On("ListByTenant", mock.Anything, tenantID).
			Return([]*model.Role{role1, role2}, nil).Once()

		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := &authv1.ListRolesRequest{}

		resp, err := adapter.ListRoles(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Roles, 2)
		assert.Equal(t, role1.Name, resp.Roles[0].Name)
		assert.Equal(t, role2.Name, resp.Roles[1].Name)
		mockRoleRepo.AssertExpectations(t)
	})

	t.Run("ListRoles with empty result", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		tenantID := uuid.New()

		// Mock: empty list
		mockRoleRepo.On("ListByTenant", mock.Anything, tenantID).
			Return([]*model.Role{}, nil).Once()

		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := &authv1.ListRolesRequest{}

		resp, err := adapter.ListRoles(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Roles, 0)
		mockRoleRepo.AssertExpectations(t)
	})
}

// TestRoleAdapter_GetRole tests getting a role via adapter
func TestRoleAdapter_GetRole(t *testing.T) {
	t.Run("GetRole successfully", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		roleID := uuid.New()
		tenantID := uuid.New()
		expectedRole := &model.Role{
			ID:          roleID,
			Name:        "admin",
			DisplayName: "Administrator",
			Description: "Admin role",
			TenantID:    tenantID,
			CreatedAt:   time.Now(),
		}

		// Mock: find role by ID
		mockRoleRepo.On("FindByID", mock.Anything, roleID).
			Return(expectedRole, nil).Once()

		req := &authv1.GetRoleRequest{
			Id: roleID.String(),
		}

		resp, err := adapter.GetRole(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, roleID.String(), resp.Id)
		assert.Equal(t, expectedRole.Name, resp.Name)
		assert.Equal(t, expectedRole.DisplayName, resp.DisplayName)
		mockRoleRepo.AssertExpectations(t)
	})

	t.Run("GetRole with invalid ID", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		req := &authv1.GetRoleRequest{
			Id: "invalid-uuid",
		}

		resp, err := adapter.GetRole(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid role id")
	})
}

// TestRoleAdapter_CreateRole tests creating a role via adapter
func TestRoleAdapter_CreateRole(t *testing.T) {
	t.Run("CreateRole successfully", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		tenantID := uuid.New()

		// Mock: create role
		mockRoleRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Role")).
			Return(nil).Once()

		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := &authv1.CreateRoleRequest{
			Name:        "editor",
			DisplayName: "Editor",
			Description: "Editor role",
		}

		resp, err := adapter.CreateRole(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "editor", resp.Name)
		assert.Equal(t, "Editor", resp.DisplayName)
		mockRoleRepo.AssertExpectations(t)
	})

	t.Run("CreateRole with invalid parent ID", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		tenantID := uuid.New()
		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := &authv1.CreateRoleRequest{
			Name:        "editor",
			DisplayName: "Editor",
			ParentId:    "invalid-uuid",
		}

		resp, err := adapter.CreateRole(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid parent role id")
	})
}

// TestRoleAdapter_UpdateRole tests updating a role via adapter
func TestRoleAdapter_UpdateRole(t *testing.T) {
	t.Run("UpdateRole successfully", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		roleID := uuid.New()
		tenantID := uuid.New()
		existingRole := &model.Role{
			ID:          roleID,
			Name:        "editor",
			DisplayName: "Editor",
			Description: "Old description",
			TenantID:    tenantID,
			CreatedAt:   time.Now(),
		}

		// Mock: find and update role
		mockRoleRepo.On("FindByID", mock.Anything, roleID).
			Return(existingRole, nil).Once()
		mockRoleRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Role")).
			Return(nil).Once()

		req := &authv1.UpdateRoleRequest{
			Id:          roleID.String(),
			DisplayName: "Senior Editor",
			Description: "New description",
		}

		resp, err := adapter.UpdateRole(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Senior Editor", resp.DisplayName)
		assert.Equal(t, "New description", resp.Description)
		mockRoleRepo.AssertExpectations(t)
	})

	t.Run("UpdateRole with invalid ID", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		req := &authv1.UpdateRoleRequest{
			Id:          "invalid-uuid",
			DisplayName: "New Name",
		}

		resp, err := adapter.UpdateRole(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid role id")
	})
}

// TestRoleAdapter_DeleteRole tests deleting a role via adapter
func TestRoleAdapter_DeleteRole(t *testing.T) {
	t.Run("DeleteRole successfully", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		roleID := uuid.New()

		// Mock: delete role
		mockRoleRepo.On("Delete", mock.Anything, roleID).
			Return(nil).Once()

		req := &authv1.DeleteRoleRequest{
			Id: roleID.String(),
		}

		resp, err := adapter.DeleteRole(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockRoleRepo.AssertExpectations(t)
	})

	t.Run("DeleteRole with invalid ID", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		req := &authv1.DeleteRoleRequest{
			Id: "invalid-uuid",
		}

		resp, err := adapter.DeleteRole(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid role id")
	})
}

// TestRoleAdapter_GetRolePermissions tests getting role permissions via adapter
func TestRoleAdapter_GetRolePermissions(t *testing.T) {
	t.Run("GetRolePermissions successfully", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		roleID := uuid.New()
		perm1 := &model.Permission{
			ID:       uuid.New(),
			Resource: "user",
			Action:   "read",
		}
		perm2 := &model.Permission{
			ID:       uuid.New(),
			Resource: "user",
			Action:   "write",
		}

		// Mock: get role permissions
		mockPermRepo.On("GetRolePermissions", mock.Anything, roleID).
			Return([]*model.Permission{perm1, perm2}, nil).Once()

		req := &authv1.GetRolePermissionsRequest{
			RoleId: roleID.String(),
		}

		resp, err := adapter.GetRolePermissions(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Permissions, 2)
		assert.Equal(t, perm1.Resource, resp.Permissions[0].Resource)
		assert.Equal(t, perm1.Action, resp.Permissions[0].Action)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("GetRolePermissions with invalid ID", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		req := &authv1.GetRolePermissionsRequest{
			RoleId: "invalid-uuid",
		}

		resp, err := adapter.GetRolePermissions(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid role id")
	})
}

// TestRoleAdapter_AssignPermissions tests assigning permissions to role via adapter
func TestRoleAdapter_AssignPermissions(t *testing.T) {
	t.Run("AssignPermissions successfully", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		roleID := uuid.New()
		permID1 := uuid.New()
		permID2 := uuid.New()
		tenantID := uuid.New()

		// Mock: assign permissions
		mockPermRepo.On("AssignPermissionToRole", mock.Anything, roleID, permID1, tenantID).
			Return(nil).Once()
		mockPermRepo.On("AssignPermissionToRole", mock.Anything, roleID, permID2, tenantID).
			Return(nil).Once()

		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := &authv1.AssignPermissionsRequest{
			RoleId:        roleID.String(),
			PermissionIds: []string{permID1.String(), permID2.String()},
		}

		resp, err := adapter.AssignPermissions(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("AssignPermissions with invalid role ID", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		tenantID := uuid.New()
		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := &authv1.AssignPermissionsRequest{
			RoleId:        "invalid-uuid",
			PermissionIds: []string{uuid.New().String()},
		}

		resp, err := adapter.AssignPermissions(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid role id")
	})

	t.Run("AssignPermissions with invalid permission ID", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		roleID := uuid.New()
		tenantID := uuid.New()
		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := &authv1.AssignPermissionsRequest{
			RoleId:        roleID.String(),
			PermissionIds: []string{"invalid-uuid"},
		}

		resp, err := adapter.AssignPermissions(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid permission id")
	})
}

// TestRoleAdapter_RevokePermissions tests revoking permissions from role via adapter
func TestRoleAdapter_RevokePermissions(t *testing.T) {
	t.Run("RevokePermissions successfully", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		roleID := uuid.New()
		permID1 := uuid.New()
		permID2 := uuid.New()

		// Mock: revoke permissions
		mockPermRepo.On("RemovePermissionFromRole", mock.Anything, roleID, permID1).
			Return(nil).Once()
		mockPermRepo.On("RemovePermissionFromRole", mock.Anything, roleID, permID2).
			Return(nil).Once()

		req := &authv1.RevokePermissionsRequest{
			RoleId:        roleID.String(),
			PermissionIds: []string{permID1.String(), permID2.String()},
		}

		resp, err := adapter.RevokePermissions(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockPermRepo.AssertExpectations(t)
	})

	t.Run("RevokePermissions with invalid role ID", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		req := &authv1.RevokePermissionsRequest{
			RoleId:        "invalid-uuid",
			PermissionIds: []string{uuid.New().String()},
		}

		resp, err := adapter.RevokePermissions(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid role id")
	})
}

// TestRoleAdapter_AssignRoleToUser tests assigning role to user via adapter
func TestRoleAdapter_AssignRoleToUser(t *testing.T) {
	t.Run("AssignRoleToUser successfully", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		userID := uuid.New()
		roleID := uuid.New()
		tenantID := uuid.New()

		// Mock: assign role to user
		mockRoleRepo.On("AssignRoleToUser", mock.Anything, userID, roleID, tenantID).
			Return(nil).Once()

		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := &authv1.AssignRoleToUserRequest{
			UserId: userID.String(),
			RoleId: roleID.String(),
		}

		resp, err := adapter.AssignRoleToUser(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockRoleRepo.AssertExpectations(t)
	})

	t.Run("AssignRoleToUser with invalid user ID", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		tenantID := uuid.New()
		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := &authv1.AssignRoleToUserRequest{
			UserId: "invalid-uuid",
			RoleId: uuid.New().String(),
		}

		resp, err := adapter.AssignRoleToUser(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid user id")
	})

	t.Run("AssignRoleToUser with invalid role ID", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		tenantID := uuid.New()
		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := &authv1.AssignRoleToUserRequest{
			UserId: uuid.New().String(),
			RoleId: "invalid-uuid",
		}

		resp, err := adapter.AssignRoleToUser(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid role id")
	})
}

// TestRoleAdapter_RemoveRoleFromUser tests removing role from user via adapter
func TestRoleAdapter_RemoveRoleFromUser(t *testing.T) {
	t.Run("RemoveRoleFromUser successfully", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		userID := uuid.New()
		roleID := uuid.New()

		// Mock: remove role from user
		mockRoleRepo.On("RemoveRoleFromUser", mock.Anything, userID, roleID).
			Return(nil).Once()

		req := &authv1.RemoveRoleFromUserRequest{
			UserId: userID.String(),
			RoleId: roleID.String(),
		}

		resp, err := adapter.RemoveRoleFromUser(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockRoleRepo.AssertExpectations(t)
	})

	t.Run("RemoveRoleFromUser with invalid user ID", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		req := &authv1.RemoveRoleFromUserRequest{
			UserId: "invalid-uuid",
			RoleId: uuid.New().String(),
		}

		resp, err := adapter.RemoveRoleFromUser(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid user id")
	})

	t.Run("RemoveRoleFromUser with invalid role ID", func(t *testing.T) {
		mockRoleRepo := new(MockRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		adapter := NewRoleAdapter(mockRoleRepo, mockPermRepo)

		req := &authv1.RemoveRoleFromUserRequest{
			UserId: uuid.New().String(),
			RoleId: "invalid-uuid",
		}

		resp, err := adapter.RemoveRoleFromUser(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid role id")
	})
}
