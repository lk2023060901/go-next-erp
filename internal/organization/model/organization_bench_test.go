package model

import (
	"testing"

	"github.com/google/uuid"
)

// BenchmarkOrganization_IsRoot 基准测试：判断是否根节点
func BenchmarkOrganization_IsRoot(b *testing.B) {
	org := &Organization{
		ParentID: nil,
		Level:    1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = org.IsRoot()
	}
}

// BenchmarkOrganization_HasChildren 基准测试：判断是否有子节点
func BenchmarkOrganization_HasChildren(b *testing.B) {
	org := &Organization{
		IsLeaf: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = org.HasChildren()
	}
}

// BenchmarkOrganization_GetFullPath 基准测试：获取完整路径
func BenchmarkOrganization_GetFullPath(b *testing.B) {
	org := &Organization{
		Name:      "部门A",
		PathNames: "/公司/事业部/部门A/",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = org.GetFullPath()
	}
}

// BenchmarkOrganization_IsActive 基准测试：判断激活状态
func BenchmarkOrganization_IsActive(b *testing.B) {
	org := &Organization{
		Status: "active",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = org.IsActive()
	}
}

// BenchmarkOrganizationType_CanBeParentOf 基准测试：判断父类型关系
func BenchmarkOrganizationType_CanBeParentOf(b *testing.B) {
	orgType := &OrganizationType{
		AllowedChildTypes: []string{"department", "team", "group"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = orgType.CanBeParentOf("department")
	}
}

// BenchmarkOrganizationType_CanBeChildOf 基准测试：判断子类型关系
func BenchmarkOrganizationType_CanBeChildOf(b *testing.B) {
	orgType := &OrganizationType{
		AllowedParentTypes: []string{"company", "division"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = orgType.CanBeChildOf("company")
	}
}

// BenchmarkOrganizationType_IsActive 基准测试：判断类型激活状态
func BenchmarkOrganizationType_IsActive(b *testing.B) {
	orgType := &OrganizationType{
		Status: "active",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = orgType.IsActive()
	}
}

// BenchmarkOrganizationClosure_IsSelf 基准测试：判断自身关系
func BenchmarkOrganizationClosure_IsSelf(b *testing.B) {
	closure := &OrganizationClosure{
		Depth: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = closure.IsSelf()
	}
}

// BenchmarkOrganizationClosure_IsDirectChild 基准测试：判断直接子节点
func BenchmarkOrganizationClosure_IsDirectChild(b *testing.B) {
	closure := &OrganizationClosure{
		Depth: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = closure.IsDirectChild()
	}
}

// BenchmarkOrganization_Creation 基准测试：创建组织对象
func BenchmarkOrganization_Creation(b *testing.B) {
	tenantID := uuid.New()
	typeID := uuid.New()
	createdBy := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &Organization{
			ID:       uuid.New(),
			TenantID: tenantID,
			Code:     "DEPT001",
			Name:     "技术部",
			TypeID:   typeID,
			Level:    2,
			Status:   "active",
			IsLeaf:   true,
			Tags:     []string{"研发", "核心"},
			CreatedBy: createdBy,
		}
	}
}

// BenchmarkOrganizationType_Creation 基准测试：创建组织类型对象
func BenchmarkOrganizationType_Creation(b *testing.B) {
	tenantID := uuid.New()
	createdBy := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &OrganizationType{
			ID:                 uuid.New(),
			TenantID:           tenantID,
			Code:               "department",
			Name:               "部门",
			Level:              2,
			MaxLevel:           5,
			AllowedParentTypes: []string{"company", "division"},
			AllowedChildTypes:  []string{"team", "group"},
			Status:             "active",
			CreatedBy:          createdBy,
		}
	}
}

// BenchmarkOrganizationClosure_Creation 基准测试：创建闭包表记录
func BenchmarkOrganizationClosure_Creation(b *testing.B) {
	tenantID := uuid.New()
	ancestorID := uuid.New()
	descendantID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &OrganizationClosure{
			TenantID:     tenantID,
			AncestorID:   ancestorID,
			DescendantID: descendantID,
			Depth:        1,
		}
	}
}

// BenchmarkOrganization_FieldAccess 基准测试：字段访问性能
func BenchmarkOrganization_FieldAccess(b *testing.B) {
	org := &Organization{
		ID:       uuid.New(),
		Code:     "DEPT001",
		Name:     "技术部",
		Level:    2,
		Status:   "active",
		IsLeaf:   true,
		Tags:     []string{"研发", "核心"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = org.ID
		_ = org.Code
		_ = org.Name
		_ = org.Level
		_ = org.Status
		_ = org.IsLeaf
		_ = org.Tags
	}
}

// BenchmarkOrganization_LargeTagsSlice 基准测试：大标签数组性能
func BenchmarkOrganization_LargeTagsSlice(b *testing.B) {
	tags := make([]string, 100)
	for i := 0; i < 100; i++ {
		tags[i] = "tag-" + string(rune(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		org := &Organization{
			Tags: tags,
		}
		_ = org.Tags
	}
}

// BenchmarkOrganizationType_LargeAllowedTypes 基准测试：大允许类型数组性能
func BenchmarkOrganizationType_LargeAllowedTypes(b *testing.B) {
	allowedTypes := make([]string, 50)
	for i := 0; i < 50; i++ {
		allowedTypes[i] = "type-" + string(rune(i))
	}

	orgType := &OrganizationType{
		AllowedParentTypes: allowedTypes,
		AllowedChildTypes:  allowedTypes,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = orgType.CanBeParentOf("type-10")
		_ = orgType.CanBeChildOf("type-20")
	}
}

// BenchmarkConcurrentOrganization_IsRoot 基准测试：并发判断根节点
func BenchmarkConcurrentOrganization_IsRoot(b *testing.B) {
	org := &Organization{
		ParentID: nil,
		Level:    1,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = org.IsRoot()
		}
	})
}

// BenchmarkConcurrentOrganizationType_CanBeParentOf 基准测试：并发判断父类型
func BenchmarkConcurrentOrganizationType_CanBeParentOf(b *testing.B) {
	orgType := &OrganizationType{
		AllowedChildTypes: []string{"department", "team", "group"},
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = orgType.CanBeParentOf("department")
		}
	})
}
