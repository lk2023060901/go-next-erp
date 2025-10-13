package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestOrganizationType_CanBeParentOf 测试父类型判断
func TestOrganizationType_CanBeParentOf(t *testing.T) {
	tests := []struct {
		name          string
		orgType       *OrganizationType
		childTypeCode string
		expected      bool
	}{
		{
			name: "允许的子类型",
			orgType: &OrganizationType{
				AllowedChildTypes: []string{"department", "team"},
			},
			childTypeCode: "department",
			expected:      true,
		},
		{
			name: "不允许的子类型",
			orgType: &OrganizationType{
				AllowedChildTypes: []string{"department", "team"},
			},
			childTypeCode: "company",
			expected:      false,
		},
		{
			name: "空允许列表时允许所有",
			orgType: &OrganizationType{
				AllowedChildTypes: []string{},
			},
			childTypeCode: "any_type",
			expected:      true,
		},
		{
			name: "nil允许列表时允许所有",
			orgType: &OrganizationType{
				AllowedChildTypes: nil,
			},
			childTypeCode: "any_type",
			expected:      true,
		},
		{
			name: "空字符串子类型",
			orgType: &OrganizationType{
				AllowedChildTypes: []string{"department"},
			},
			childTypeCode: "",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.orgType.CanBeParentOf(tt.childTypeCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOrganizationType_CanBeChildOf 测试子类型判断
func TestOrganizationType_CanBeChildOf(t *testing.T) {
	tests := []struct {
		name           string
		orgType        *OrganizationType
		parentTypeCode string
		expected       bool
	}{
		{
			name: "允许的父类型",
			orgType: &OrganizationType{
				AllowedParentTypes: []string{"company", "group"},
			},
			parentTypeCode: "company",
			expected:       true,
		},
		{
			name: "不允许的父类型",
			orgType: &OrganizationType{
				AllowedParentTypes: []string{"company", "group"},
			},
			parentTypeCode: "department",
			expected:       false,
		},
		{
			name: "空允许列表时允许所有",
			orgType: &OrganizationType{
				AllowedParentTypes: []string{},
			},
			parentTypeCode: "any_type",
			expected:       true,
		},
		{
			name: "nil允许列表时允许所有",
			orgType: &OrganizationType{
				AllowedParentTypes: nil,
			},
			parentTypeCode: "any_type",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.orgType.CanBeChildOf(tt.parentTypeCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOrganizationType_IsActive 测试激活状态判断
func TestOrganizationType_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		orgType  *OrganizationType
		expected bool
	}{
		{
			name: "状态为active时是激活状态",
			orgType: &OrganizationType{
				Status: "active",
			},
			expected: true,
		},
		{
			name: "状态为inactive时不是激活状态",
			orgType: &OrganizationType{
				Status: "inactive",
			},
			expected: false,
		},
		{
			name: "状态为空时不是激活状态",
			orgType: &OrganizationType{
				Status: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.orgType.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOrganizationType_TableName 测试表名
func TestOrganizationType_TableName(t *testing.T) {
	orgType := &OrganizationType{}
	tableName := orgType.TableName()
	assert.Equal(t, "organization_types", tableName)
}

// TestOrganizationType_CompleteWorkflow 测试完整工作流
func TestOrganizationType_CompleteWorkflow(t *testing.T) {
	tenantID := uuid.New()
	createdBy := uuid.New()

	orgType := &OrganizationType{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		Code:               "department",
		Name:               "部门",
		Icon:               "department-icon",
		Level:              2,
		MaxLevel:           5,
		AllowRoot:          false,
		AllowMulti:         true,
		AllowedParentTypes: []string{"company", "division"},
		AllowedChildTypes:  []string{"team", "group"},
		EnableLeader:       true,
		EnableLegalInfo:    false,
		EnableAddress:      true,
		Sort:               10,
		Status:             "active",
		IsSystem:           false,
		CreatedBy:          createdBy,
		UpdatedBy:          createdBy,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// 验证基本属性
	assert.NotEqual(t, uuid.Nil, orgType.ID)
	assert.Equal(t, "department", orgType.Code)
	assert.Equal(t, "部门", orgType.Name)
	assert.Equal(t, 2, orgType.Level)
	assert.Equal(t, 5, orgType.MaxLevel)
	assert.True(t, orgType.IsActive())
	assert.False(t, orgType.AllowRoot)
	assert.True(t, orgType.AllowMulti)

	// 验证类型关系
	assert.True(t, orgType.CanBeChildOf("company"))
	assert.True(t, orgType.CanBeChildOf("division"))
	assert.False(t, orgType.CanBeChildOf("team"))

	assert.True(t, orgType.CanBeParentOf("team"))
	assert.True(t, orgType.CanBeParentOf("group"))
	assert.False(t, orgType.CanBeParentOf("company"))

	// 验证功能开关
	assert.True(t, orgType.EnableLeader)
	assert.False(t, orgType.EnableLegalInfo)
	assert.True(t, orgType.EnableAddress)
}

// TestOrganizationType_EdgeCases 测试边界情况
func TestOrganizationType_EdgeCases(t *testing.T) {
	t.Run("极大MaxLevel", func(t *testing.T) {
		orgType := &OrganizationType{
			MaxLevel: 10000,
		}
		assert.Equal(t, 10000, orgType.MaxLevel)
	})

	t.Run("负数Level", func(t *testing.T) {
		orgType := &OrganizationType{
			Level: -1,
		}
		assert.Equal(t, -1, orgType.Level)
	})

	t.Run("大量AllowedTypes", func(t *testing.T) {
		allowedTypes := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			allowedTypes[i] = "type-" + string(rune(i))
		}

		orgType := &OrganizationType{
			AllowedParentTypes: allowedTypes,
			AllowedChildTypes:  allowedTypes,
		}

		assert.Len(t, orgType.AllowedParentTypes, 1000)
		assert.Len(t, orgType.AllowedChildTypes, 1000)
	})

	t.Run("空数组vs nil数组", func(t *testing.T) {
		emptyArray := &OrganizationType{
			AllowedParentTypes: []string{},
			AllowedChildTypes:  []string{},
		}
		nilArray := &OrganizationType{
			AllowedParentTypes: nil,
			AllowedChildTypes:  nil,
		}

		// 两者行为应该相同（都允许所有）
		assert.True(t, emptyArray.CanBeChildOf("any"))
		assert.True(t, nilArray.CanBeChildOf("any"))
		assert.True(t, emptyArray.CanBeParentOf("any"))
		assert.True(t, nilArray.CanBeParentOf("any"))
	})
}

// TestOrganizationType_SpecialCharacters 测试特殊字符
func TestOrganizationType_SpecialCharacters(t *testing.T) {
	specialCode := "dept-类型-🎉"
	specialName := "部门@#$%^&*()"

	orgType := &OrganizationType{
		Code: specialCode,
		Name: specialName,
		Icon: "icon-✓",
	}

	assert.Equal(t, specialCode, orgType.Code)
	assert.Equal(t, specialName, orgType.Name)
	assert.Equal(t, "icon-✓", orgType.Icon)
}

// TestOrganizationType_ZeroValues 测试零值
func TestOrganizationType_ZeroValues(t *testing.T) {
	orgType := &OrganizationType{}

	assert.Equal(t, uuid.Nil, orgType.ID)
	assert.Equal(t, uuid.Nil, orgType.TenantID)
	assert.Equal(t, "", orgType.Code)
	assert.Equal(t, "", orgType.Name)
	assert.Equal(t, 0, orgType.Level)
	assert.Equal(t, 0, orgType.MaxLevel)
	assert.False(t, orgType.AllowRoot)
	assert.False(t, orgType.AllowMulti)
	assert.Nil(t, orgType.AllowedParentTypes)
	assert.Nil(t, orgType.AllowedChildTypes)
	assert.False(t, orgType.IsActive())
}

// TestOrganizationType_CircularReference 测试循环引用检测
func TestOrganizationType_CircularReference(t *testing.T) {
	// 类型A允许类型B作为子节点
	typeA := &OrganizationType{
		Code:              "type_a",
		AllowedChildTypes: []string{"type_b"},
	}

	// 类型B允许类型A作为子节点（循环）
	typeB := &OrganizationType{
		Code:              "type_b",
		AllowedChildTypes: []string{"type_a"},
	}

	// 模型层面不阻止循环引用，这应该在业务逻辑层处理
	assert.True(t, typeA.CanBeParentOf("type_b"))
	assert.True(t, typeB.CanBeParentOf("type_a"))
}

// TestOrganizationType_MultipleMatches 测试多重匹配
func TestOrganizationType_MultipleMatches(t *testing.T) {
	orgType := &OrganizationType{
		AllowedParentTypes: []string{"type_a", "type_b", "type_a", "type_c", "type_a"},
		AllowedChildTypes:  []string{"type_x", "type_y", "type_x"},
	}

	// 重复的类型应该都能匹配
	assert.True(t, orgType.CanBeChildOf("type_a"))
	assert.True(t, orgType.CanBeParentOf("type_x"))
}

// TestOrganizationType_CaseSensitivity 测试大小写敏感性
func TestOrganizationType_CaseSensitivity(t *testing.T) {
	orgType := &OrganizationType{
		AllowedParentTypes: []string{"Company"},
		AllowedChildTypes:  []string{"Department"},
	}

	// 大小写不匹配应该返回false
	assert.False(t, orgType.CanBeChildOf("company"))
	assert.False(t, orgType.CanBeChildOf("COMPANY"))
	assert.True(t, orgType.CanBeChildOf("Company"))

	assert.False(t, orgType.CanBeParentOf("department"))
	assert.False(t, orgType.CanBeParentOf("DEPARTMENT"))
	assert.True(t, orgType.CanBeParentOf("Department"))
}
