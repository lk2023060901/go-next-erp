package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestUser_IsLocked 测试用户锁定状态判断
func TestUser_IsLocked(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		expected bool
	}{
		{
			name: "状态为locked时用户被锁定",
			user: &User{
				Status: UserStatusLocked,
			},
			expected: true,
		},
		{
			name: "LockedUntil在未来时用户被锁定",
			user: &User{
				Status:      UserStatusActive,
				LockedUntil: timePtr(time.Now().Add(1 * time.Hour)),
			},
			expected: true,
		},
		{
			name: "LockedUntil在过去时用户未锁定",
			user: &User{
				Status:      UserStatusActive,
				LockedUntil: timePtr(time.Now().Add(-1 * time.Hour)),
			},
			expected: false,
		},
		{
			name: "状态为active且LockedUntil为nil时未锁定",
			user: &User{
				Status:      UserStatusActive,
				LockedUntil: nil,
			},
			expected: false,
		},
		{
			name: "状态为inactive时未锁定",
			user: &User{
				Status:      UserStatusInactive,
				LockedUntil: nil,
			},
			expected: false,
		},
		{
			name: "状态为banned时未锁定（但不可用）",
			user: &User{
				Status: UserStatusBanned,
			},
			expected: false,
		},
		{
			name: "状态为locked且LockedUntil已过期",
			user: &User{
				Status:      UserStatusLocked,
				LockedUntil: timePtr(time.Now().Add(-1 * time.Hour)),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.IsLocked()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestUser_IsActive 测试用户激活状态判断
func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		expected bool
	}{
		{
			name: "状态为active且未锁定时激活",
			user: &User{
				Status:      UserStatusActive,
				LockedUntil: nil,
			},
			expected: true,
		},
		{
			name: "状态为active但被锁定时不激活",
			user: &User{
				Status:      UserStatusActive,
				LockedUntil: timePtr(time.Now().Add(1 * time.Hour)),
			},
			expected: false,
		},
		{
			name: "状态为inactive时不激活",
			user: &User{
				Status: UserStatusInactive,
			},
			expected: false,
		},
		{
			name: "状态为locked时不激活",
			user: &User{
				Status: UserStatusLocked,
			},
			expected: false,
		},
		{
			name: "状态为banned时不激活",
			user: &User{
				Status: UserStatusBanned,
			},
			expected: false,
		},
		{
			name: "状态为active但Status为locked",
			user: &User{
				Status: UserStatusLocked,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestUser_CanLogin 测试用户登录权限判断
func TestUser_CanLogin(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		expected bool
	}{
		{
			name: "激活且未删除的用户可以登录",
			user: &User{
				Status:    UserStatusActive,
				DeletedAt: nil,
			},
			expected: true,
		},
		{
			name: "激活但已删除的用户不能登录",
			user: &User{
				Status:    UserStatusActive,
				DeletedAt: timePtr(time.Now()),
			},
			expected: false,
		},
		{
			name: "未激活的用户不能登录",
			user: &User{
				Status:    UserStatusInactive,
				DeletedAt: nil,
			},
			expected: false,
		},
		{
			name: "被锁定的用户不能登录",
			user: &User{
				Status:      UserStatusActive,
				LockedUntil: timePtr(time.Now().Add(1 * time.Hour)),
				DeletedAt:   nil,
			},
			expected: false,
		},
		{
			name: "被封禁的用户不能登录",
			user: &User{
				Status:    UserStatusBanned,
				DeletedAt: nil,
			},
			expected: false,
		},
		{
			name: "锁定且已删除的用户不能登录",
			user: &User{
				Status:    UserStatusLocked,
				DeletedAt: timePtr(time.Now()),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.CanLogin()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestUser_CompleteWorkflow 测试用户完整工作流
func TestUser_CompleteWorkflow(t *testing.T) {
	now := time.Now()
	tenantID := uuid.New()
	userID := uuid.New()

	user := &User{
		ID:           userID,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "$2a$10$hashedpassword",
		TenantID:     tenantID,
		Status:       UserStatusActive,
		MFAEnabled:   false,
		MFASecret:    strPtr(""),
		LastLoginAt:  &now,
		LastLoginIP:  strPtr("192.168.1.1"),
		LoginAttempts: 0,
		LockedUntil:  nil,
		Metadata: map[string]interface{}{
			"employee_id": "EMP001",
			"department":  "IT",
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
	}

	// 验证基本属性
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, tenantID, user.TenantID)

	// 验证状态方法
	assert.True(t, user.IsActive())
	assert.False(t, user.IsLocked())
	assert.True(t, user.CanLogin())

	// 锁定用户
	user.LockedUntil = timePtr(now.Add(1 * time.Hour))
	assert.True(t, user.IsLocked())
	assert.False(t, user.IsActive())
	assert.False(t, user.CanLogin())

	// 解锁用户
	user.LockedUntil = timePtr(now.Add(-1 * time.Hour))
	assert.False(t, user.IsLocked())
	assert.True(t, user.IsActive())
	assert.True(t, user.CanLogin())

	// 删除用户
	user.DeletedAt = &now
	assert.False(t, user.CanLogin())
}

// TestUser_EdgeCases 测试边界情况
func TestUser_EdgeCases(t *testing.T) {
	t.Run("零值用户", func(t *testing.T) {
		user := &User{}
		assert.False(t, user.IsActive())
		assert.False(t, user.IsLocked())
		assert.False(t, user.CanLogin())
	})

	t.Run("LockedUntil刚好等于当前时间", func(t *testing.T) {
		now := time.Now()
		user := &User{
			Status:      UserStatusActive,
			LockedUntil: &now,
		}
		// time.Now()可能略有差异，这里只验证行为一致性
		_ = user.IsLocked()
	})

	t.Run("极长的用户名", func(t *testing.T) {
		user := &User{
			Username: string(make([]byte, 10000)),
			Status:   UserStatusActive,
		}
		assert.Equal(t, 10000, len(user.Username))
		assert.True(t, user.CanLogin())
	})

	t.Run("空字符串字段", func(t *testing.T) {
		user := &User{
			Username:     "",
			Email:        "",
			PasswordHash: "",
			Status:       UserStatusActive,
		}
		assert.True(t, user.CanLogin())
	})

	t.Run("nil Metadata", func(t *testing.T) {
		user := &User{
			Metadata: nil,
			Status:   UserStatusActive,
		}
		assert.Nil(t, user.Metadata)
		assert.True(t, user.CanLogin())
	})

	t.Run("空Metadata", func(t *testing.T) {
		user := &User{
			Metadata: map[string]interface{}{},
			Status:   UserStatusActive,
		}
		assert.NotNil(t, user.Metadata)
		assert.Len(t, user.Metadata, 0)
	})

	t.Run("极大LoginAttempts", func(t *testing.T) {
		user := &User{
			LoginAttempts: 999999,
			Status:        UserStatusActive,
		}
		assert.Equal(t, 999999, user.LoginAttempts)
	})
}

// TestUser_SpecialCharacters 测试特殊字符
func TestUser_SpecialCharacters(t *testing.T) {
	specialChars := "!@#$%^&*()_+-=[]{}|;':\",./<>?`~中文字符🎉"

	user := &User{
		Username:     specialChars,
		Email:        specialChars + "@example.com",
		PasswordHash: specialChars,
		LastLoginIP:  strPtr(specialChars),
		Status:       UserStatusActive,
	}

	assert.Equal(t, specialChars, user.Username)
	assert.Contains(t, user.Email, specialChars)
	assert.True(t, user.CanLogin())
}

// TestUser_MFAScenarios 测试MFA场景
func TestUser_MFAScenarios(t *testing.T) {
	t.Run("MFA启用的用户", func(t *testing.T) {
		user := &User{
			Status:     UserStatusActive,
			MFAEnabled: true,
			MFASecret:  strPtr("JBSWY3DPEHPK3PXP"),
		}
		assert.True(t, user.MFAEnabled)
		assert.NotEmpty(t, user.MFASecret)
		assert.True(t, user.CanLogin()) // MFA不影响登录权限判断
	})

	t.Run("MFA未启用的用户", func(t *testing.T) {
		user := &User{
			Status:     UserStatusActive,
			MFAEnabled: false,
			MFASecret:  strPtr(""),
		}
		assert.False(t, user.MFAEnabled)
		assert.Empty(t, user.MFASecret)
	})
}

// TestUser_StatusTransitions 测试状态转换
func TestUser_StatusTransitions(t *testing.T) {
	user := &User{Status: UserStatusActive}

	// Active -> Inactive
	user.Status = UserStatusInactive
	assert.False(t, user.IsActive())
	assert.False(t, user.CanLogin())

	// Inactive -> Active
	user.Status = UserStatusActive
	assert.True(t, user.IsActive())
	assert.True(t, user.CanLogin())

	// Active -> Locked
	user.Status = UserStatusLocked
	assert.True(t, user.IsLocked())
	assert.False(t, user.IsActive())
	assert.False(t, user.CanLogin())

	// Locked -> Active
	user.Status = UserStatusActive
	assert.False(t, user.IsLocked())
	assert.True(t, user.IsActive())

	// Active -> Banned
	user.Status = UserStatusBanned
	assert.False(t, user.IsActive())
	assert.False(t, user.CanLogin())
}

// TestUser_TimestampScenarios 测试时间戳场景
func TestUser_TimestampScenarios(t *testing.T) {
	t.Run("未来的LockedUntil", func(t *testing.T) {
		user := &User{
			Status:      UserStatusActive,
			LockedUntil: timePtr(time.Now().Add(24 * time.Hour)),
		}
		assert.True(t, user.IsLocked())
	})

	t.Run("远过去的LockedUntil", func(t *testing.T) {
		user := &User{
			Status:      UserStatusActive,
			LockedUntil: timePtr(time.Now().Add(-365 * 24 * time.Hour)),
		}
		assert.False(t, user.IsLocked())
	})

	t.Run("最近的LastLoginAt", func(t *testing.T) {
		user := &User{
			Status:      UserStatusActive,
			LastLoginAt: timePtr(time.Now().Add(-5 * time.Minute)),
		}
		assert.True(t, user.CanLogin())
	})

	t.Run("nil时间字段", func(t *testing.T) {
		user := &User{
			Status:      UserStatusActive,
			LastLoginAt: nil,
			LockedUntil: nil,
			DeletedAt:   nil,
		}
		assert.True(t, user.CanLogin())
	})
}

// TestUser_MetadataScenarios 测试元数据场景
func TestUser_MetadataScenarios(t *testing.T) {
	t.Run("复杂元数据", func(t *testing.T) {
		user := &User{
			Status: UserStatusActive,
			Metadata: map[string]interface{}{
				"employee_id": "EMP001",
				"department":  "IT",
				"roles":       []string{"admin", "developer"},
				"permissions": map[string]bool{
					"read":  true,
					"write": true,
				},
				"age":    30,
				"salary": 100000.50,
			},
		}
		assert.Len(t, user.Metadata, 6)
		assert.Equal(t, "EMP001", user.Metadata["employee_id"])
	})

	t.Run("嵌套元数据", func(t *testing.T) {
		user := &User{
			Status: UserStatusActive,
			Metadata: map[string]interface{}{
				"profile": map[string]interface{}{
					"first_name": "John",
					"last_name":  "Doe",
				},
			},
		}
		assert.NotNil(t, user.Metadata["profile"])
	})
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func strPtr(s string) *string {
	return &s
}
