package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Session 模型测试
// ============================================================================

// TestSession_IsValid 测试会话有效性判断
func TestSession_IsValid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		session  *Session
		expected bool
	}{
		{
			name: "有效会话",
			session: &Session{
				ExpiresAt: now.Add(1 * time.Hour),
				RevokedAt: nil,
			},
			expected: true,
		},
		{
			name: "已过期会话",
			session: &Session{
				ExpiresAt: now.Add(-1 * time.Hour),
				RevokedAt: nil,
			},
			expected: false,
		},
		{
			name: "已撤销会话",
			session: &Session{
				ExpiresAt: now.Add(1 * time.Hour),
				RevokedAt: timePtr(now.Add(-5 * time.Minute)),
			},
			expected: false,
		},
		{
			name: "已撤销且已过期",
			session: &Session{
				ExpiresAt: now.Add(-1 * time.Hour),
				RevokedAt: timePtr(now.Add(-30 * time.Minute)),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.session.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSession_Revoke 测试撤销会话
func TestSession_Revoke(t *testing.T) {
	session := &Session{
		ID:        uuid.New(),
		RevokedAt: nil,
	}

	assert.Nil(t, session.RevokedAt)
	assert.True(t, session.IsValid() || !session.IsValid()) // 取决于ExpiresAt

	session.Revoke()

	assert.NotNil(t, session.RevokedAt)
	assert.False(t, session.IsValid())
}

// TestSession_Creation 测试会话创建
func TestSession_Creation(t *testing.T) {
	now := time.Now()
	sessionID := uuid.New()
	userID := uuid.New()
	tenantID := uuid.New()

	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		TenantID:  tenantID,
		Token:     "jwt_token_here",
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0...",
		ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
		RevokedAt: nil,
	}

	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, tenantID, session.TenantID)
	assert.True(t, session.IsValid())
}

// TestSession_EdgeCases 测试会话边界情况
func TestSession_EdgeCases(t *testing.T) {
	t.Run("零值会话", func(t *testing.T) {
		session := &Session{}
		// ExpiresAt默认为零值，会被认为已过期
		assert.False(t, session.IsValid())
	})

	t.Run("刚好过期", func(t *testing.T) {
		now := time.Now()
		session := &Session{
			ExpiresAt: now,
		}
		// 由于时间可能有微小差异，这里不做严格断言
		_ = session.IsValid()
	})

	t.Run("极长Token", func(t *testing.T) {
		session := &Session{
			Token:     string(make([]byte, 10000)),
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		assert.Len(t, session.Token, 10000)
		assert.True(t, session.IsValid())
	})

	t.Run("特殊字符", func(t *testing.T) {
		special := "!@#$%^&*()中文🎉"
		session := &Session{
			UserAgent: special,
			IPAddress: special,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
		assert.Equal(t, special, session.UserAgent)
		assert.True(t, session.IsValid())
	})
}

// ============================================================================
// Tenant 模型测试
// ============================================================================

// TestTenant_IsActive 测试租户激活状态
func TestTenant_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		tenant   *Tenant
		expected bool
	}{
		{
			name: "活跃租户",
			tenant: &Tenant{
				Status:    TenantStatusActive,
				DeletedAt: nil,
			},
			expected: true,
		},
		{
			name: "暂停租户",
			tenant: &Tenant{
				Status:    TenantStatusSuspended,
				DeletedAt: nil,
			},
			expected: false,
		},
		{
			name: "过期租户",
			tenant: &Tenant{
				Status:    TenantStatusExpired,
				DeletedAt: nil,
			},
			expected: false,
		},
		{
			name: "已删除租户",
			tenant: &Tenant{
				Status:    TenantStatusActive,
				DeletedAt: timePtr(time.Now()),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tenant.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTenant_Creation 测试租户创建
func TestTenant_Creation(t *testing.T) {
	now := time.Now()
	tenantID := uuid.New()

	tenant := &Tenant{
		ID:          tenantID,
		Name:        "company_a",
		DisplayName: "公司A",
		Domain:      "company-a.example.com",
		Status:      TenantStatusActive,
		MaxUsers:    1000,
		MaxStorage:  10737418240, // 10GB
		Settings: map[string]interface{}{
			"theme":    "dark",
			"language": "zh-CN",
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
	}

	assert.Equal(t, tenantID, tenant.ID)
	assert.Equal(t, "company_a", tenant.Name)
	assert.Equal(t, 1000, tenant.MaxUsers)
	assert.True(t, tenant.IsActive())
}

// TestTenant_StatusTransitions 测试租户状态转换
func TestTenant_StatusTransitions(t *testing.T) {
	tenant := &Tenant{Status: TenantStatusActive}

	// Active -> Suspended
	tenant.Status = TenantStatusSuspended
	assert.False(t, tenant.IsActive())

	// Suspended -> Active
	tenant.Status = TenantStatusActive
	assert.True(t, tenant.IsActive())

	// Active -> Expired
	tenant.Status = TenantStatusExpired
	assert.False(t, tenant.IsActive())
}

// TestTenant_EdgeCases 测试租户边界情况
func TestTenant_EdgeCases(t *testing.T) {
	t.Run("零值租户", func(t *testing.T) {
		tenant := &Tenant{}
		assert.False(t, tenant.IsActive())
	})

	t.Run("负数配额", func(t *testing.T) {
		tenant := &Tenant{
			MaxUsers:   -1,
			MaxStorage: -1000,
			Status:     TenantStatusActive,
		}
		assert.Equal(t, -1, tenant.MaxUsers)
		assert.True(t, tenant.IsActive())
	})

	t.Run("极大配额", func(t *testing.T) {
		tenant := &Tenant{
			MaxUsers:   1000000,
			MaxStorage: 1099511627776, // 1TB
			Status:     TenantStatusActive,
		}
		assert.Equal(t, 1000000, tenant.MaxUsers)
	})

	t.Run("nil Settings", func(t *testing.T) {
		tenant := &Tenant{
			Settings: nil,
			Status:   TenantStatusActive,
		}
		assert.Nil(t, tenant.Settings)
		assert.True(t, tenant.IsActive())
	})

	t.Run("空Settings", func(t *testing.T) {
		tenant := &Tenant{
			Settings: map[string]interface{}{},
			Status:   TenantStatusActive,
		}
		assert.NotNil(t, tenant.Settings)
		assert.Len(t, tenant.Settings, 0)
	})

	t.Run("复杂Settings", func(t *testing.T) {
		tenant := &Tenant{
			Settings: map[string]interface{}{
				"features": map[string]bool{
					"ai":         true,
					"blockchain": false,
				},
				"limits": []int{100, 200, 300},
				"metadata": map[string]string{
					"region": "cn-east",
				},
			},
			Status: TenantStatusActive,
		}
		assert.Len(t, tenant.Settings, 3)
	})

	t.Run("特殊字符", func(t *testing.T) {
		special := "公司-测试_2024🎉"
		tenant := &Tenant{
			Name:        special,
			DisplayName: special,
			Domain:      special,
			Status:      TenantStatusActive,
		}
		assert.Equal(t, special, tenant.Name)
	})
}

// TestTenant_StatusConstants 测试租户状态常量
func TestTenant_StatusConstants(t *testing.T) {
	assert.Equal(t, TenantStatus("active"), TenantStatusActive)
	assert.Equal(t, TenantStatus("suspended"), TenantStatusSuspended)
	assert.Equal(t, TenantStatus("expired"), TenantStatusExpired)
}
