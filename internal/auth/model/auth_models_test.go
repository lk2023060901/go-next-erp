package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Role 模型测试
// ============================================================================

// TestRole_Creation 测试角色创建
func TestRole_Creation(t *testing.T) {
	tenantID := uuid.New()
	roleID := uuid.New()
	parentID := uuid.New()
	now := time.Now()

	role := &Role{
		ID:          roleID,
		Name:        "admin",
		DisplayName: "系统管理员",
		Description: "拥有系统所有权限",
		TenantID:    tenantID,
		ParentID:    &parentID,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   nil,
	}

	assert.Equal(t, roleID, role.ID)
	assert.Equal(t, "admin", role.Name)
	assert.Equal(t, "系统管理员", role.DisplayName)
	assert.Equal(t, tenantID, role.TenantID)
	assert.NotNil(t, role.ParentID)
	assert.Equal(t, parentID, *role.ParentID)
	assert.Nil(t, role.DeletedAt)
}

// TestRole_EdgeCases 测试角色边界情况
func TestRole_EdgeCases(t *testing.T) {
	t.Run("零值角色", func(t *testing.T) {
		role := &Role{}
		assert.Equal(t, uuid.Nil, role.ID)
		assert.Empty(t, role.Name)
		assert.Nil(t, role.ParentID)
	})

	t.Run("nil ParentID", func(t *testing.T) {
		role := &Role{
			Name:     "admin",
			ParentID: nil,
		}
		assert.Nil(t, role.ParentID)
	})

	t.Run("极长名称", func(t *testing.T) {
		longName := string(make([]byte, 10000))
		role := &Role{
			Name:        longName,
			DisplayName: longName,
		}
		assert.Len(t, role.Name, 10000)
	})

	t.Run("特殊字符", func(t *testing.T) {
		specialChars := "!@#$%^&*()_+-=中文🎉"
		role := &Role{
			Name:        specialChars,
			DisplayName: specialChars,
			Description: specialChars,
		}
		assert.Equal(t, specialChars, role.Name)
	})
}

// TestUserRole_Creation 测试用户角色关联创建
func TestUserRole_Creation(t *testing.T) {
	userRole := &UserRole{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		RoleID:    uuid.New(),
		TenantID:  uuid.New(),
		CreatedAt: time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, userRole.ID)
	assert.NotEqual(t, uuid.Nil, userRole.UserID)
	assert.NotEqual(t, uuid.Nil, userRole.RoleID)
}

// ============================================================================
// Permission 模型测试
// ============================================================================

// TestPermission_String 测试权限字符串表示
func TestPermission_String(t *testing.T) {
	tests := []struct {
		name       string
		permission *Permission
		expected   string
	}{
		{
			name: "标准权限",
			permission: &Permission{
				Resource: "document",
				Action:   "read",
			},
			expected: "document:read",
		},
		{
			name: "通配符资源",
			permission: &Permission{
				Resource: "*",
				Action:   "read",
			},
			expected: "*:read",
		},
		{
			name: "通配符操作",
			permission: &Permission{
				Resource: "document",
				Action:   "*",
			},
			expected: "document:*",
		},
		{
			name: "全通配符",
			permission: &Permission{
				Resource: "*",
				Action:   "*",
			},
			expected: "*:*",
		},
		{
			name: "空字符串",
			permission: &Permission{
				Resource: "",
				Action:   "",
			},
			expected: ":",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.permission.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPermission_Match 测试权限匹配
func TestPermission_Match(t *testing.T) {
	tests := []struct {
		name       string
		permission *Permission
		resource   string
		action     string
		expected   bool
	}{
		{
			name: "完全匹配",
			permission: &Permission{
				Resource: "document",
				Action:   "read",
			},
			resource: "document",
			action:   "read",
			expected: true,
		},
		{
			name: "不匹配",
			permission: &Permission{
				Resource: "document",
				Action:   "read",
			},
			resource: "document",
			action:   "write",
			expected: false,
		},
		{
			name: "资源通配符匹配",
			permission: &Permission{
				Resource: "document",
				Action:   "*",
			},
			resource: "document",
			action:   "read",
			expected: true,
		},
		{
			name: "操作通配符匹配",
			permission: &Permission{
				Resource: "*",
				Action:   "read",
			},
			resource: "document",
			action:   "read",
			expected: true,
		},
		{
			name: "全通配符匹配",
			permission: &Permission{
				Resource: "*",
				Action:   "*",
			},
			resource: "document",
			action:   "read",
			expected: true,
		},
		{
			name: "资源不匹配",
			permission: &Permission{
				Resource: "document",
				Action:   "read",
			},
			resource: "user",
			action:   "read",
			expected: false,
		},
		{
			name: "资源通配符但资源不匹配",
			permission: &Permission{
				Resource: "document",
				Action:   "read",
			},
			resource: "user",
			action:   "write",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.permission.Match(tt.resource, tt.action)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPermission_StandardActions 测试标准操作常量
func TestPermission_StandardActions(t *testing.T) {
	assert.Equal(t, "create", ActionCreate)
	assert.Equal(t, "read", ActionRead)
	assert.Equal(t, "update", ActionUpdate)
	assert.Equal(t, "delete", ActionDelete)
	assert.Equal(t, "list", ActionList)
	assert.Equal(t, "export", ActionExport)
	assert.Equal(t, "*", ActionAll)
}

// TestPermission_Creation 测试权限创建
func TestPermission_Creation(t *testing.T) {
	tenantID := uuid.New()
	now := time.Now()

	permission := &Permission{
		ID:          uuid.New(),
		Resource:    "document",
		Action:      "read",
		DisplayName: "读取文档",
		Description: "允许读取文档内容",
		TenantID:    tenantID,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   nil,
	}

	assert.NotEqual(t, uuid.Nil, permission.ID)
	assert.Equal(t, "document", permission.Resource)
	assert.Equal(t, "read", permission.Action)
	assert.Equal(t, "document:read", permission.String())
}

// TestPermission_EdgeCases 测试权限边界情况
func TestPermission_EdgeCases(t *testing.T) {
	t.Run("空资源和操作", func(t *testing.T) {
		perm := &Permission{
			Resource: "",
			Action:   "",
		}
		assert.False(t, perm.Match("document", "read"))
		assert.True(t, perm.Match("", ""))
	})

	t.Run("特殊字符资源", func(t *testing.T) {
		special := "document:v2:🎉"
		perm := &Permission{
			Resource: special,
			Action:   "read",
		}
		assert.True(t, perm.Match(special, "read"))
		assert.Contains(t, perm.String(), special)
	})

	t.Run("极长资源名", func(t *testing.T) {
		longResource := string(make([]byte, 1000))
		perm := &Permission{
			Resource: longResource,
			Action:   "read",
		}
		assert.Len(t, perm.Resource, 1000)
	})
}

// TestRolePermission_Creation 测试角色权限关联创建
func TestRolePermission_Creation(t *testing.T) {
	rolePermission := &RolePermission{
		ID:           uuid.New(),
		RoleID:       uuid.New(),
		PermissionID: uuid.New(),
		TenantID:     uuid.New(),
		CreatedAt:    time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, rolePermission.ID)
	assert.NotEqual(t, uuid.Nil, rolePermission.RoleID)
	assert.NotEqual(t, uuid.Nil, rolePermission.PermissionID)
}

// ============================================================================
// Policy 模型测试
// ============================================================================

// TestPolicy_Creation 测试策略创建
func TestPolicy_Creation(t *testing.T) {
	tenantID := uuid.New()
	now := time.Now()

	policy := &Policy{
		ID:          uuid.New(),
		Name:        "department_access",
		Description: "部门访问策略",
		TenantID:    tenantID,
		Resource:    "document",
		Action:      "read",
		Expression:  "user.department_id == resource.department_id",
		Effect:      PolicyEffectAllow,
		Priority:    10,
		Enabled:     true,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   nil,
	}

	assert.NotEqual(t, uuid.Nil, policy.ID)
	assert.Equal(t, "department_access", policy.Name)
	assert.Equal(t, PolicyEffectAllow, policy.Effect)
	assert.Equal(t, 10, policy.Priority)
	assert.True(t, policy.Enabled)
}

// TestPolicy_Effects 测试策略效果
func TestPolicy_Effects(t *testing.T) {
	assert.Equal(t, PolicyEffect("allow"), PolicyEffectAllow)
	assert.Equal(t, PolicyEffect("deny"), PolicyEffectDeny)

	t.Run("Allow策略", func(t *testing.T) {
		policy := &Policy{
			Effect: PolicyEffectAllow,
		}
		assert.Equal(t, PolicyEffectAllow, policy.Effect)
	})

	t.Run("Deny策略", func(t *testing.T) {
		policy := &Policy{
			Effect: PolicyEffectDeny,
		}
		assert.Equal(t, PolicyEffectDeny, policy.Effect)
	})
}

// TestPolicy_Priority 测试策略优先级
func TestPolicy_Priority(t *testing.T) {
	tests := []struct {
		name     string
		priority int
	}{
		{"最低优先级", 0},
		{"低优先级", 1},
		{"中优先级", 10},
		{"高优先级", 100},
		{"最高优先级", 1000},
		{"负数优先级", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &Policy{
				Priority: tt.priority,
			}
			assert.Equal(t, tt.priority, policy.Priority)
		})
	}
}

// TestPolicy_Expression 测试策略表达式
func TestPolicy_Expression(t *testing.T) {
	expressions := []string{
		"user.department_id == resource.department_id",
		"user.level >= 3",
		"user.roles contains 'manager'",
		"time.hour >= 9 && time.hour <= 18",
		"resource.status == 'published' || user.id == resource.owner_id",
		"user.age >= 18 && user.verified == true",
	}

	for _, expr := range expressions {
		t.Run(expr, func(t *testing.T) {
			policy := &Policy{
				Expression: expr,
			}
			assert.Equal(t, expr, policy.Expression)
			assert.NotEmpty(t, policy.Expression)
		})
	}
}

// TestPolicy_EdgeCases 测试策略边界情况
func TestPolicy_EdgeCases(t *testing.T) {
	t.Run("零值策略", func(t *testing.T) {
		policy := &Policy{}
		assert.Equal(t, uuid.Nil, policy.ID)
		assert.Empty(t, policy.Name)
		assert.Equal(t, 0, policy.Priority)
		assert.False(t, policy.Enabled)
	})

	t.Run("禁用策略", func(t *testing.T) {
		policy := &Policy{
			Enabled: false,
			Effect:  PolicyEffectAllow,
		}
		assert.False(t, policy.Enabled)
	})

	t.Run("极长表达式", func(t *testing.T) {
		longExpr := string(make([]byte, 10000))
		policy := &Policy{
			Expression: longExpr,
		}
		assert.Len(t, policy.Expression, 10000)
	})

	t.Run("特殊字符表达式", func(t *testing.T) {
		policy := &Policy{
			Expression: "user.name == '张三' && user.部门 == 'IT'",
		}
		assert.Contains(t, policy.Expression, "张三")
	})

	t.Run("空表达式", func(t *testing.T) {
		policy := &Policy{
			Expression: "",
		}
		assert.Empty(t, policy.Expression)
	})
}

// TestPolicy_Scenarios 测试策略场景
func TestPolicy_Scenarios(t *testing.T) {
	t.Run("时间限制策略", func(t *testing.T) {
		policy := &Policy{
			Name:       "work_hours_access",
			Expression: "time.hour >= 9 && time.hour <= 18",
			Effect:     PolicyEffectAllow,
			Enabled:    true,
		}
		assert.Contains(t, policy.Expression, "time.hour")
	})

	t.Run("部门隔离策略", func(t *testing.T) {
		policy := &Policy{
			Name:       "department_isolation",
			Expression: "user.department_id == resource.department_id",
			Effect:     PolicyEffectAllow,
			Enabled:    true,
		}
		assert.Contains(t, policy.Expression, "department_id")
	})

	t.Run("角色检查策略", func(t *testing.T) {
		policy := &Policy{
			Name:       "manager_only",
			Expression: "user.roles contains 'manager'",
			Effect:     PolicyEffectAllow,
			Enabled:    true,
		}
		assert.Contains(t, policy.Expression, "manager")
	})

	t.Run("拒绝策略", func(t *testing.T) {
		policy := &Policy{
			Name:       "deny_external_access",
			Expression: "user.is_external == true",
			Effect:     PolicyEffectDeny,
			Priority:   100,
			Enabled:    true,
		}
		assert.Equal(t, PolicyEffectDeny, policy.Effect)
		assert.Equal(t, 100, policy.Priority)
	})
}

// TestPolicy_MultipleConditions 测试复杂条件策略
func TestPolicy_MultipleConditions(t *testing.T) {
	complexExpressions := []string{
		"(user.level >= 3 && user.department == 'IT') || user.role == 'admin'",
		"user.age >= 18 && user.verified == true && user.status == 'active'",
		"time.hour >= 9 && time.hour <= 18 && time.day != 'Sunday'",
		"resource.visibility == 'public' || (resource.visibility == 'private' && user.id == resource.owner_id)",
	}

	for i, expr := range complexExpressions {
		t.Run("复杂条件"+string(rune('A'+i)), func(t *testing.T) {
			policy := &Policy{
				Name:       "complex_policy_" + string(rune('A'+i)),
				Expression: expr,
				Effect:     PolicyEffectAllow,
				Priority:   i + 1,
				Enabled:    true,
			}
			assert.Equal(t, expr, policy.Expression)
			assert.True(t, policy.Enabled)
		})
	}
}
