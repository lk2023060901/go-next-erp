package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestOrganizationClosure_IsSelf 测试自身关系判断
func TestOrganizationClosure_IsSelf(t *testing.T) {
	tests := []struct {
		name     string
		closure  *OrganizationClosure
		expected bool
	}{
		{
			name: "Depth为0时是自身关系",
			closure: &OrganizationClosure{
				Depth: 0,
			},
			expected: true,
		},
		{
			name: "Depth为1时不是自身关系",
			closure: &OrganizationClosure{
				Depth: 1,
			},
			expected: false,
		},
		{
			name: "Depth为负数时不是自身关系",
			closure: &OrganizationClosure{
				Depth: -1,
			},
			expected: false,
		},
		{
			name: "Depth为大于1时不是自身关系",
			closure: &OrganizationClosure{
				Depth: 10,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.closure.IsSelf()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOrganizationClosure_IsDirectChild 测试直接子节点判断
func TestOrganizationClosure_IsDirectChild(t *testing.T) {
	tests := []struct {
		name     string
		closure  *OrganizationClosure
		expected bool
	}{
		{
			name: "Depth为1时是直接子节点",
			closure: &OrganizationClosure{
				Depth: 1,
			},
			expected: true,
		},
		{
			name: "Depth为0时不是直接子节点",
			closure: &OrganizationClosure{
				Depth: 0,
			},
			expected: false,
		},
		{
			name: "Depth为2时不是直接子节点",
			closure: &OrganizationClosure{
				Depth: 2,
			},
			expected: false,
		},
		{
			name: "Depth为负数时不是直接子节点",
			closure: &OrganizationClosure{
				Depth: -1,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.closure.IsDirectChild()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOrganizationClosure_CompleteWorkflow 测试完整工作流
func TestOrganizationClosure_CompleteWorkflow(t *testing.T) {
	tenantID := uuid.New()
	ancestorID := uuid.New()
	descendantID := uuid.New()

	// 自身关系
	selfClosure := &OrganizationClosure{
		TenantID:     tenantID,
		AncestorID:   ancestorID,
		DescendantID: ancestorID,
		Depth:        0,
	}

	assert.Equal(t, ancestorID, selfClosure.AncestorID)
	assert.Equal(t, ancestorID, selfClosure.DescendantID)
	assert.True(t, selfClosure.IsSelf())
	assert.False(t, selfClosure.IsDirectChild())

	// 直接子节点关系
	directChildClosure := &OrganizationClosure{
		TenantID:     tenantID,
		AncestorID:   ancestorID,
		DescendantID: descendantID,
		Depth:        1,
	}

	assert.NotEqual(t, directChildClosure.AncestorID, directChildClosure.DescendantID)
	assert.False(t, directChildClosure.IsSelf())
	assert.True(t, directChildClosure.IsDirectChild())

	// 间接子节点关系
	indirectChildClosure := &OrganizationClosure{
		TenantID:     tenantID,
		AncestorID:   ancestorID,
		DescendantID: descendantID,
		Depth:        3,
	}

	assert.False(t, indirectChildClosure.IsSelf())
	assert.False(t, indirectChildClosure.IsDirectChild())
	assert.Equal(t, 3, indirectChildClosure.Depth)
}

// TestOrganizationClosure_EdgeCases 测试边界情况
func TestOrganizationClosure_EdgeCases(t *testing.T) {
	t.Run("极大Depth值", func(t *testing.T) {
		closure := &OrganizationClosure{
			Depth: 1000,
		}
		assert.Equal(t, 1000, closure.Depth)
		assert.False(t, closure.IsSelf())
		assert.False(t, closure.IsDirectChild())
	})

	t.Run("Depth为最小整数", func(t *testing.T) {
		closure := &OrganizationClosure{
			Depth: -2147483648,
		}
		assert.False(t, closure.IsSelf())
		assert.False(t, closure.IsDirectChild())
	})

	t.Run("Depth为最大整数", func(t *testing.T) {
		closure := &OrganizationClosure{
			Depth: 2147483647,
		}
		assert.False(t, closure.IsSelf())
		assert.False(t, closure.IsDirectChild())
	})

	t.Run("相同的AncestorID和DescendantID", func(t *testing.T) {
		id := uuid.New()
		closure := &OrganizationClosure{
			AncestorID:   id,
			DescendantID: id,
			Depth:        0,
		}
		assert.Equal(t, closure.AncestorID, closure.DescendantID)
		assert.True(t, closure.IsSelf())
	})
}

// TestOrganizationClosure_ZeroValues 测试零值
func TestOrganizationClosure_ZeroValues(t *testing.T) {
	closure := &OrganizationClosure{}

	assert.Equal(t, uuid.Nil, closure.TenantID)
	assert.Equal(t, uuid.Nil, closure.AncestorID)
	assert.Equal(t, uuid.Nil, closure.DescendantID)
	assert.Equal(t, 0, closure.Depth)
	assert.True(t, closure.IsSelf())
	assert.False(t, closure.IsDirectChild())
}

// TestOrganizationClosure_TreeTraversal 测试树遍历场景
func TestOrganizationClosure_TreeTraversal(t *testing.T) {
	tenantID := uuid.New()
	rootID := uuid.New()
	level1ID := uuid.New()
	level2ID := uuid.New()
	level3ID := uuid.New()

	// 构建树形结构的闭包关系
	closures := []*OrganizationClosure{
		// Root的自身关系
		{TenantID: tenantID, AncestorID: rootID, DescendantID: rootID, Depth: 0},

		// Level1
		{TenantID: tenantID, AncestorID: level1ID, DescendantID: level1ID, Depth: 0},
		{TenantID: tenantID, AncestorID: rootID, DescendantID: level1ID, Depth: 1},

		// Level2
		{TenantID: tenantID, AncestorID: level2ID, DescendantID: level2ID, Depth: 0},
		{TenantID: tenantID, AncestorID: level1ID, DescendantID: level2ID, Depth: 1},
		{TenantID: tenantID, AncestorID: rootID, DescendantID: level2ID, Depth: 2},

		// Level3
		{TenantID: tenantID, AncestorID: level3ID, DescendantID: level3ID, Depth: 0},
		{TenantID: tenantID, AncestorID: level2ID, DescendantID: level3ID, Depth: 1},
		{TenantID: tenantID, AncestorID: level1ID, DescendantID: level3ID, Depth: 2},
		{TenantID: tenantID, AncestorID: rootID, DescendantID: level3ID, Depth: 3},
	}

	// 验证每个级别的关系
	selfClosures := 0
	directChildren := 0
	indirectChildren := 0

	for _, c := range closures {
		if c.IsSelf() {
			selfClosures++
		} else if c.IsDirectChild() {
			directChildren++
		} else {
			indirectChildren++
		}
	}

	assert.Equal(t, 4, selfClosures)       // 4个节点的自身关系
	assert.Equal(t, 3, directChildren)     // 3个直接子节点关系
	assert.Equal(t, 3, indirectChildren)   // 3个间接子节点关系
}
