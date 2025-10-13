package rbac

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
)

// ============================================================================
// RBAC Engine Benchmark Tests
// ============================================================================

// BenchmarkEngine_CheckPermission_Hit 基准测试：权限检查（匹配）
func BenchmarkEngine_CheckPermission_Hit(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()

	roleRepo := new(MockRoleRepository)
	permRepo := new(MockPermissionRepository)
	roles := []*model.Role{{ID: roleID, Name: "admin"}}
	perms := []*model.Permission{{Resource: "document", Action: "read"}}

	roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
	roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{}, nil)
	permRepo.On("GetRolePermissions", ctx, roleID).Return(perms, nil)

	engine := NewEngine(roleRepo, permRepo, nil)

	// 预热缓存
	_, _ = engine.CheckPermission(ctx, userID, "document", "read")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.CheckPermission(ctx, userID, "document", "read")
	}
}

// BenchmarkEngine_CheckPermission_Miss 基准测试：权限检查（不匹配）
func BenchmarkEngine_CheckPermission_Miss(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()

	roleRepo := new(MockRoleRepository)
	permRepo := new(MockPermissionRepository)
	roles := []*model.Role{{ID: roleID, Name: "viewer"}}
	perms := []*model.Permission{{Resource: "document", Action: "read"}}

	roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
	roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{}, nil)
	permRepo.On("GetRolePermissions", ctx, roleID).Return(perms, nil)

	engine := NewEngine(roleRepo, permRepo, nil)

	// 预热缓存
	_, _ = engine.CheckPermission(ctx, userID, "document", "write")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.CheckPermission(ctx, userID, "document", "write")
	}
}

// BenchmarkEngine_CheckPermission_Wildcard 基准测试：通配符权限匹配
func BenchmarkEngine_CheckPermission_Wildcard(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()

	roleRepo := new(MockRoleRepository)
	permRepo := new(MockPermissionRepository)

	roles := []*model.Role{{ID: roleID, Name: "admin"}}
	perms := []*model.Permission{{Resource: "*", Action: "*"}}

	roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
	roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{}, nil)
	permRepo.On("GetRolePermissions", ctx, roleID).Return(perms, nil)

	engine := NewEngine(roleRepo, permRepo, nil)

	// 预热缓存
	_, _ = engine.CheckPermission(ctx, userID, "any", "any")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.CheckPermission(ctx, userID, "any", "any")
	}
}

// BenchmarkEngine_CheckPermission_NoCache 基准测试：无缓存权限检查
func BenchmarkEngine_CheckPermission_NoCache(b *testing.B) {
	ctx := context.Background()
	roleID := uuid.New()

	roleRepo := new(MockRoleRepository)
	permRepo := new(MockPermissionRepository)

	roles := []*model.Role{{ID: roleID, Name: "admin"}}
	perms := []*model.Permission{{Resource: "document", Action: "read"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userID := uuid.New() // 每次不同的用户ID，避免缓存

		roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil).Once()
		roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{}, nil).Once()
		permRepo.On("GetRolePermissions", ctx, roleID).Return(perms, nil).Once()

		engine := NewEngine(roleRepo, permRepo, nil)
		_, _ = engine.CheckPermission(ctx, userID, "document", "read")
	}
}

// BenchmarkEngine_GetUserRoles 基准测试：获取用户角色
func BenchmarkEngine_GetUserRoles(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	role1 := uuid.New()
	role2 := uuid.New()

	roleRepo := new(MockRoleRepository)
	permRepo := new(MockPermissionRepository)

	roles := []*model.Role{
		{ID: role1, Name: "developer"},
		{ID: role2, Name: "viewer"},
	}

	roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
	roleRepo.On("GetRoleHierarchy", ctx, role1).Return([]*model.Role{}, nil)
	roleRepo.On("GetRoleHierarchy", ctx, role2).Return([]*model.Role{}, nil)

	engine := NewEngine(roleRepo, permRepo, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.GetUserRoles(ctx, userID)
	}
}

// BenchmarkEngine_GetUserRoles_WithHierarchy 基准测试：获取角色（含继承）
func BenchmarkEngine_GetUserRoles_WithHierarchy(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	devRole := uuid.New()
	employeeRole := uuid.New()

	roleRepo := new(MockRoleRepository)
	permRepo := new(MockPermissionRepository)

	directRoles := []*model.Role{
		{ID: devRole, Name: "developer"},
	}
	parentRoles := []*model.Role{
		{ID: employeeRole, Name: "employee"},
	}

	roleRepo.On("GetUserRoles", ctx, userID).Return(directRoles, nil)
	roleRepo.On("GetRoleHierarchy", ctx, devRole).Return(parentRoles, nil)

	engine := NewEngine(roleRepo, permRepo, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.GetUserRoles(ctx, userID)
	}
}

// BenchmarkEngine_HasRole 基准测试：检查用户角色
func BenchmarkEngine_HasRole(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()

	roleRepo := new(MockRoleRepository)
	permRepo := new(MockPermissionRepository)

	roles := []*model.Role{{ID: roleID, Name: "admin"}}

	roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
	roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{}, nil)

	engine := NewEngine(roleRepo, permRepo, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.HasRole(ctx, userID, roleID)
	}
}

// BenchmarkEngine_GetUserPermissions 基准测试：获取用户权限
func BenchmarkEngine_GetUserPermissions(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()

	roleRepo := new(MockRoleRepository)
	permRepo := new(MockPermissionRepository)

	roles := []*model.Role{{ID: roleID, Name: "admin"}}
	perms := []*model.Permission{
		{ID: uuid.New(), Resource: "document", Action: "read"},
		{ID: uuid.New(), Resource: "document", Action: "write"},
		{ID: uuid.New(), Resource: "user", Action: "read"},
	}

	roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
	roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{}, nil)
	permRepo.On("GetRolePermissions", ctx, roleID).Return(perms, nil)

	engine := NewEngine(roleRepo, permRepo, nil)

	// 预热缓存
	_, _ = engine.GetUserPermissions(ctx, userID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.GetUserPermissions(ctx, userID)
	}
}

// BenchmarkEngine_GetUserPermissions_MultiRole 基准测试：多角色权限获取
func BenchmarkEngine_GetUserPermissions_MultiRole(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	role1 := uuid.New()
	role2 := uuid.New()
	role3 := uuid.New()

	roleRepo := new(MockRoleRepository)
	permRepo := new(MockPermissionRepository)

	roles := []*model.Role{
		{ID: role1, Name: "admin"},
		{ID: role2, Name: "developer"},
		{ID: role3, Name: "viewer"},
	}

	perms1 := []*model.Permission{
		{ID: uuid.New(), Resource: "system", Action: "*"},
	}
	perms2 := []*model.Permission{
		{ID: uuid.New(), Resource: "code", Action: "write"},
	}
	perms3 := []*model.Permission{
		{ID: uuid.New(), Resource: "document", Action: "read"},
	}

	roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
	roleRepo.On("GetRoleHierarchy", ctx, role1).Return([]*model.Role{}, nil)
	roleRepo.On("GetRoleHierarchy", ctx, role2).Return([]*model.Role{}, nil)
	roleRepo.On("GetRoleHierarchy", ctx, role3).Return([]*model.Role{}, nil)
	permRepo.On("GetRolePermissions", ctx, role1).Return(perms1, nil)
	permRepo.On("GetRolePermissions", ctx, role2).Return(perms2, nil)
	permRepo.On("GetRolePermissions", ctx, role3).Return(perms3, nil)

	engine := NewEngine(roleRepo, permRepo, nil)

	// 预热缓存
	_, _ = engine.GetUserPermissions(ctx, userID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.GetUserPermissions(ctx, userID)
	}
}

// ============================================================================
// 并发基准测试
// ============================================================================

// BenchmarkConcurrentEngine_CheckPermission 基准测试：并发权限检查
func BenchmarkConcurrentEngine_CheckPermission(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()

	roleRepo := new(MockRoleRepository)
	permRepo := new(MockPermissionRepository)
	roles := []*model.Role{{ID: roleID, Name: "admin"}}
	perms := []*model.Permission{{Resource: "document", Action: "read"}}

	roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
	roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{}, nil)
	permRepo.On("GetRolePermissions", ctx, roleID).Return(perms, nil)

	engine := NewEngine(roleRepo, permRepo, nil)

	// 预热缓存
	_, _ = engine.CheckPermission(ctx, userID, "document", "read")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = engine.CheckPermission(ctx, userID, "document", "read")
		}
	})
}

// BenchmarkConcurrentEngine_GetUserRoles 基准测试：并发获取角色
func BenchmarkConcurrentEngine_GetUserRoles(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()

	roleRepo := new(MockRoleRepository)
	permRepo := new(MockPermissionRepository)

	roles := []*model.Role{{ID: roleID, Name: "admin"}}

	roleRepo.On("GetUserRoles", ctx, userID).Return(roles, nil)
	roleRepo.On("GetRoleHierarchy", ctx, roleID).Return([]*model.Role{}, nil)

	engine := NewEngine(roleRepo, permRepo, nil)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = engine.GetUserRoles(ctx, userID)
		}
	})
}
