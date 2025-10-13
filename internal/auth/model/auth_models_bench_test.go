package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Role 模型基准测试
// ============================================================================

// BenchmarkRole_Creation 基准测试：创建角色对象
func BenchmarkRole_Creation(b *testing.B) {
	tenantID := uuid.New()
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &Role{
			ID:          uuid.New(),
			Name:        "admin",
			DisplayName: "管理员",
			TenantID:    tenantID,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}
}

// BenchmarkRole_FieldAccess 基准测试：字段访问
func BenchmarkRole_FieldAccess(b *testing.B) {
	role := &Role{
		ID:          uuid.New(),
		Name:        "admin",
		DisplayName: "管理员",
		TenantID:    uuid.New(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = role.ID
		_ = role.Name
		_ = role.DisplayName
		_ = role.TenantID
	}
}

// BenchmarkUserRole_Creation 基准测试：创建用户角色关联
func BenchmarkUserRole_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &UserRole{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			RoleID:    uuid.New(),
			TenantID:  uuid.New(),
			CreatedAt: time.Now(),
		}
	}
}

// ============================================================================
// Permission 模型基准测试
// ============================================================================

// BenchmarkPermission_String 基准测试：权限字符串表示
func BenchmarkPermission_String(b *testing.B) {
	perm := &Permission{
		Resource: "document",
		Action:   "read",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = perm.String()
	}
}

// BenchmarkPermission_Match_Exact 基准测试：精确匹配
func BenchmarkPermission_Match_Exact(b *testing.B) {
	perm := &Permission{
		Resource: "document",
		Action:   "read",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = perm.Match("document", "read")
	}
}

// BenchmarkPermission_Match_Wildcard 基准测试：通配符匹配
func BenchmarkPermission_Match_Wildcard(b *testing.B) {
	perm := &Permission{
		Resource: "*",
		Action:   "*",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = perm.Match("document", "read")
	}
}

// BenchmarkPermission_Match_ResourceWildcard 基准测试：资源通配符匹配
func BenchmarkPermission_Match_ResourceWildcard(b *testing.B) {
	perm := &Permission{
		Resource: "document",
		Action:   "*",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = perm.Match("document", "read")
	}
}

// BenchmarkPermission_Match_NoMatch 基准测试：不匹配情况
func BenchmarkPermission_Match_NoMatch(b *testing.B) {
	perm := &Permission{
		Resource: "document",
		Action:   "read",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = perm.Match("user", "write")
	}
}

// BenchmarkPermission_Creation 基准测试：创建权限对象
func BenchmarkPermission_Creation(b *testing.B) {
	tenantID := uuid.New()
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &Permission{
			ID:          uuid.New(),
			Resource:    "document",
			Action:      "read",
			DisplayName: "读取文档",
			TenantID:    tenantID,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}
}

// BenchmarkPermission_AllOperations 基准测试：所有操作组合
func BenchmarkPermission_AllOperations(b *testing.B) {
	perm := &Permission{
		Resource: "document",
		Action:   "read",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = perm.String()
		_ = perm.Match("document", "read")
		_ = perm.Match("document", "*")
		_ = perm.Match("*", "read")
	}
}

// BenchmarkRolePermission_Creation 基准测试：创建角色权限关联
func BenchmarkRolePermission_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &RolePermission{
			ID:           uuid.New(),
			RoleID:       uuid.New(),
			PermissionID: uuid.New(),
			TenantID:     uuid.New(),
			CreatedAt:    time.Now(),
		}
	}
}

// ============================================================================
// Policy 模型基准测试
// ============================================================================

// BenchmarkPolicy_Creation 基准测试：创建策略对象
func BenchmarkPolicy_Creation(b *testing.B) {
	tenantID := uuid.New()
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &Policy{
			ID:          uuid.New(),
			Name:        "department_access",
			TenantID:    tenantID,
			Resource:    "document",
			Action:      "read",
			Expression:  "user.department_id == resource.department_id",
			Effect:      PolicyEffectAllow,
			Priority:    10,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}
}

// BenchmarkPolicy_FieldAccess 基准测试：字段访问
func BenchmarkPolicy_FieldAccess(b *testing.B) {
	policy := &Policy{
		ID:         uuid.New(),
		Name:       "test_policy",
		Expression: "user.level >= 3",
		Effect:     PolicyEffectAllow,
		Priority:   10,
		Enabled:    true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = policy.ID
		_ = policy.Name
		_ = policy.Expression
		_ = policy.Effect
		_ = policy.Priority
		_ = policy.Enabled
	}
}

// BenchmarkPolicy_EffectCheck 基准测试：效果检查
func BenchmarkPolicy_EffectCheck(b *testing.B) {
	policy := &Policy{
		Effect: PolicyEffectAllow,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = policy.Effect == PolicyEffectAllow
	}
}

// BenchmarkPolicy_EnabledCheck 基准测试：启用状态检查
func BenchmarkPolicy_EnabledCheck(b *testing.B) {
	policy := &Policy{
		Enabled: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = policy.Enabled
	}
}

// BenchmarkPolicy_PriorityCompare 基准测试：优先级比较
func BenchmarkPolicy_PriorityCompare(b *testing.B) {
	policy1 := &Policy{Priority: 10}
	policy2 := &Policy{Priority: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = policy1.Priority > policy2.Priority
	}
}

// BenchmarkPolicy_ComplexExpression 基准测试：复杂表达式
func BenchmarkPolicy_ComplexExpression(b *testing.B) {
	expr := "(user.level >= 3 && user.department == 'IT') || user.role == 'admin'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		policy := &Policy{
			Expression: expr,
		}
		_ = policy.Expression
	}
}

// ============================================================================
// 并发基准测试
// ============================================================================

// BenchmarkConcurrentPermission_Match 基准测试：并发权限匹配
func BenchmarkConcurrentPermission_Match(b *testing.B) {
	perm := &Permission{
		Resource: "document",
		Action:   "read",
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = perm.Match("document", "read")
		}
	})
}

// BenchmarkConcurrentPermission_String 基准测试：并发字符串转换
func BenchmarkConcurrentPermission_String(b *testing.B) {
	perm := &Permission{
		Resource: "document",
		Action:   "read",
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = perm.String()
		}
	})
}

// BenchmarkConcurrentPolicy_EffectCheck 基准测试：并发效果检查
func BenchmarkConcurrentPolicy_EffectCheck(b *testing.B) {
	policy := &Policy{
		Effect: PolicyEffectAllow,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = policy.Effect == PolicyEffectAllow
		}
	})
}
