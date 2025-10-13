package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestOrganization_IsRoot 测试是否根节点判断
func TestOrganization_IsRoot(t *testing.T) {
	tests := []struct {
		name     string
		org      *Organization
		expected bool
	}{
		{
			name: "ParentID为nil时是根节点",
			org: &Organization{
				ParentID: nil,
				Level:    1,
			},
			expected: true,
		},
		{
			name: "Level为1时是根节点",
			org: &Organization{
				ParentID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				Level:    1,
			},
			expected: true,
		},
		{
			name: "ParentID不为nil且Level大于1不是根节点",
			org: &Organization{
				ParentID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				Level:    2,
			},
			expected: false,
		},
		{
			name: "Level为0且ParentID不为nil",
			org: &Organization{
				ParentID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				Level:    0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.org.IsRoot()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOrganization_HasChildren 测试是否有子节点
func TestOrganization_HasChildren(t *testing.T) {
	tests := []struct {
		name     string
		org      *Organization
		expected bool
	}{
		{
			name: "IsLeaf为false有子节点",
			org: &Organization{
				IsLeaf: false,
			},
			expected: true,
		},
		{
			name: "IsLeaf为true没有子节点",
			org: &Organization{
				IsLeaf: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.org.HasChildren()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOrganization_GetFullPath 测试获取完整路径名称
func TestOrganization_GetFullPath(t *testing.T) {
	tests := []struct {
		name      string
		org       *Organization
		expected  string
	}{
		{
			name: "有PathNames时返回PathNames",
			org: &Organization{
				Name:      "部门A",
				PathNames: "/公司/事业部/部门A/",
			},
			expected: "/公司/事业部/部门A/",
		},
		{
			name: "PathNames为空时返回Name",
			org: &Organization{
				Name:      "部门A",
				PathNames: "",
			},
			expected: "部门A",
		},
		{
			name: "Name和PathNames都为空",
			org: &Organization{
				Name:      "",
				PathNames: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.org.GetFullPath()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOrganization_IsActive 测试是否激活状态
func TestOrganization_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		org      *Organization
		expected bool
	}{
		{
			name: "状态为active时是激活状态",
			org: &Organization{
				Status: "active",
			},
			expected: true,
		},
		{
			name: "状态为inactive时不是激活状态",
			org: &Organization{
				Status: "inactive",
			},
			expected: false,
		},
		{
			name: "状态为disbanded时不是激活状态",
			org: &Organization{
				Status: "disbanded",
			},
			expected: false,
		},
		{
			name: "状态为空时不是激活状态",
			org: &Organization{
				Status: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.org.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOrganization_CompleteWorkflow 测试完整工作流
func TestOrganization_CompleteWorkflow(t *testing.T) {
	tenantID := uuid.New()
	parentID := uuid.New()
	leaderID := uuid.New()
	createdBy := uuid.New()
	registerDate := time.Now().AddDate(-1, 0, 0)

	org := &Organization{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Code:          "DEPT001",
		Name:          "技术部",
		ShortName:     "Tech",
		Description:   "技术研发部门",
		TypeID:        uuid.New(),
		TypeCode:      "department",
		ParentID:      &parentID,
		Level:         2,
		Path:          fmt.Sprintf("/%s/%s/", parentID, uuid.New()),
		PathNames:     "/公司/技术部/",
		AncestorIDs:   []string{parentID.String()},
		IsLeaf:        true,
		LeaderID:      &leaderID,
		LeaderName:    "张三",
		LegalPerson:   "李四",
		UnifiedCode:   "91110000MA01234567",
		RegisterDate:  &registerDate,
		RegisterAddr:  "北京市海淀区",
		Phone:         "010-12345678",
		Email:         "tech@example.com",
		Address:       "北京市海淀区中关村大街1号",
		EmployeeCount: 50,
		DirectEmpCount: 20,
		Sort:          1,
		Status:        "active",
		Tags:          []string{"研发", "核心"},
		CreatedBy:     createdBy,
		UpdatedBy:     createdBy,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 验证所有字段
	assert.NotEqual(t, uuid.Nil, org.ID)
	assert.Equal(t, tenantID, org.TenantID)
	assert.Equal(t, "DEPT001", org.Code)
	assert.Equal(t, "技术部", org.Name)
	assert.Equal(t, "Tech", org.ShortName)
	assert.Equal(t, 2, org.Level)
	assert.False(t, org.IsRoot())
	assert.False(t, org.HasChildren())
	assert.True(t, org.IsActive())
	assert.Equal(t, "/公司/技术部/", org.GetFullPath())
	assert.Len(t, org.Tags, 2)
	assert.Equal(t, 50, org.EmployeeCount)
}

// TestOrganization_EdgeCases 测试边界情况
func TestOrganization_EdgeCases(t *testing.T) {
	t.Run("极长字符串字段", func(t *testing.T) {
		longString := string(make([]byte, 10000))
		org := &Organization{
			Name:        longString,
			Description: longString,
			Path:        longString,
			PathNames:   longString,
		}

		assert.Len(t, org.Name, 10000)
		assert.Equal(t, longString, org.GetFullPath())
	})

	t.Run("空AncestorIDs", func(t *testing.T) {
		org := &Organization{
			AncestorIDs: []string{},
		}
		assert.NotNil(t, org.AncestorIDs)
		assert.Len(t, org.AncestorIDs, 0)
	})

	t.Run("空Tags", func(t *testing.T) {
		org := &Organization{
			Tags: []string{},
		}
		assert.NotNil(t, org.Tags)
		assert.Len(t, org.Tags, 0)
	})

	t.Run("nil Tags", func(t *testing.T) {
		org := &Organization{
			Tags: nil,
		}
		assert.Nil(t, org.Tags)
	})

	t.Run("大量Tags", func(t *testing.T) {
		tags := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			tags[i] = fmt.Sprintf("tag-%d", i)
		}
		org := &Organization{
			Tags: tags,
		}
		assert.Len(t, org.Tags, 1000)
	})

	t.Run("极深层级", func(t *testing.T) {
		parentID := uuid.New()
		org := &Organization{
			Level:    100,
			ParentID: &parentID,
		}
		assert.Equal(t, 100, org.Level)
		assert.False(t, org.IsRoot())
	})

	t.Run("负数员工数", func(t *testing.T) {
		org := &Organization{
			EmployeeCount:  -1,
			DirectEmpCount: -10,
		}
		assert.Equal(t, -1, org.EmployeeCount)
		assert.Equal(t, -10, org.DirectEmpCount)
	})
}

// TestOrganization_SpecialCharacters 测试特殊字符
func TestOrganization_SpecialCharacters(t *testing.T) {
	specialChars := "!@#$%^&*()_+-=[]{}|;':\",./<>?`~中文字符🎉"

	org := &Organization{
		Code:        specialChars,
		Name:        specialChars,
		ShortName:   specialChars,
		Description: specialChars,
		Phone:       specialChars,
		Email:       specialChars,
		Address:     specialChars,
	}

	assert.Equal(t, specialChars, org.Code)
	assert.Equal(t, specialChars, org.Name)
	assert.Equal(t, specialChars, org.ShortName)
	assert.Equal(t, specialChars, org.Description)
}

// TestOrganization_NilFields 测试nil字段
func TestOrganization_NilFields(t *testing.T) {
	org := &Organization{
		ParentID:     nil,
		LeaderID:     nil,
		RegisterDate: nil,
		DeletedAt:    nil,
	}

	assert.Nil(t, org.ParentID)
	assert.Nil(t, org.LeaderID)
	assert.Nil(t, org.RegisterDate)
	assert.Nil(t, org.DeletedAt)
	assert.True(t, org.IsRoot())
}

// TestOrganization_ZeroValues 测试零值
func TestOrganization_ZeroValues(t *testing.T) {
	org := &Organization{}

	assert.Equal(t, uuid.Nil, org.ID)
	assert.Equal(t, uuid.Nil, org.TenantID)
	assert.Equal(t, "", org.Code)
	assert.Equal(t, "", org.Name)
	assert.Equal(t, 0, org.Level)
	assert.Equal(t, 0, org.EmployeeCount)
	assert.Equal(t, "", org.Status)
	assert.False(t, org.IsActive())
	assert.False(t, org.IsLeaf) // bool零值为false
	// Level=0 且 ParentID=nil，因此是根节点（OR条件：ParentID == nil）
	assert.True(t, org.IsRoot())
}
