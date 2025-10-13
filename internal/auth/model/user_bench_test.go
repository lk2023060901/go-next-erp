package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// BenchmarkUser_IsLocked 基准测试：判断用户是否被锁定
func BenchmarkUser_IsLocked(b *testing.B) {
	user := &User{
		Status:      UserStatusActive,
		LockedUntil: timePtr(time.Now().Add(1 * time.Hour)),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.IsLocked()
	}
}

// BenchmarkUser_IsLocked_WithStatusLocked 基准测试：状态为locked的判断
func BenchmarkUser_IsLocked_WithStatusLocked(b *testing.B) {
	user := &User{
		Status: UserStatusLocked,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.IsLocked()
	}
}

// BenchmarkUser_IsActive 基准测试：判断用户是否激活
func BenchmarkUser_IsActive(b *testing.B) {
	user := &User{
		Status:      UserStatusActive,
		LockedUntil: nil,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.IsActive()
	}
}

// BenchmarkUser_CanLogin 基准测试：判断用户是否可以登录
func BenchmarkUser_CanLogin(b *testing.B) {
	user := &User{
		Status:    UserStatusActive,
		DeletedAt: nil,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.CanLogin()
	}
}

// BenchmarkUser_Creation 基准测试：创建用户对象
func BenchmarkUser_Creation(b *testing.B) {
	tenantID := uuid.New()
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &User{
			ID:           uuid.New(),
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "$2a$10$hashedpassword",
			TenantID:     tenantID,
			Status:       UserStatusActive,
			MFAEnabled:   false,
			LastLoginAt:  &now,
			LastLoginIP:  strPtr("192.168.1.1"),
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	}
}

// BenchmarkUser_CreationWithMetadata 基准测试：创建带元数据的用户
func BenchmarkUser_CreationWithMetadata(b *testing.B) {
	tenantID := uuid.New()
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &User{
			ID:           uuid.New(),
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "$2a$10$hashedpassword",
			TenantID:     tenantID,
			Status:       UserStatusActive,
			Metadata: map[string]interface{}{
				"employee_id": "EMP001",
				"department":  "IT",
			},
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
}

// BenchmarkUser_StatusCheck_Active 基准测试：检查Active状态
func BenchmarkUser_StatusCheck_Active(b *testing.B) {
	user := &User{Status: UserStatusActive}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.Status == UserStatusActive
	}
}

// BenchmarkUser_AllChecks 基准测试：完整的状态检查流程
func BenchmarkUser_AllChecks(b *testing.B) {
	user := &User{
		Status:    UserStatusActive,
		DeletedAt: nil,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.IsLocked()
		_ = user.IsActive()
		_ = user.CanLogin()
	}
}

// BenchmarkUser_FieldAccess 基准测试：字段访问性能
func BenchmarkUser_FieldAccess(b *testing.B) {
	user := &User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		TenantID:     uuid.New(),
		Status:       UserStatusActive,
		LoginAttempts: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.ID
		_ = user.Username
		_ = user.Email
		_ = user.TenantID
		_ = user.Status
		_ = user.LoginAttempts
	}
}

// BenchmarkUser_MetadataAccess 基准测试：元数据访问性能
func BenchmarkUser_MetadataAccess(b *testing.B) {
	user := &User{
		Status: UserStatusActive,
		Metadata: map[string]interface{}{
			"employee_id": "EMP001",
			"department":  "IT",
			"role":        "developer",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.Metadata["employee_id"]
		_ = user.Metadata["department"]
		_ = user.Metadata["role"]
	}
}

// BenchmarkUser_LargeMetadata 基准测试：大元数据性能
func BenchmarkUser_LargeMetadata(b *testing.B) {
	metadata := make(map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		metadata[string(rune(i))] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &User{
			Status:   UserStatusActive,
			Metadata: metadata,
		}
		_ = user.Metadata
	}
}

// BenchmarkConcurrentUser_IsActive 基准测试：并发判断激活状态
func BenchmarkConcurrentUser_IsActive(b *testing.B) {
	user := &User{
		Status:      UserStatusActive,
		LockedUntil: nil,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = user.IsActive()
		}
	})
}

// BenchmarkConcurrentUser_CanLogin 基准测试：并发判断登录权限
func BenchmarkConcurrentUser_CanLogin(b *testing.B) {
	user := &User{
		Status:    UserStatusActive,
		DeletedAt: nil,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = user.CanLogin()
		}
	})
}

// BenchmarkUser_StatusTransition 基准测试：状态转换性能
func BenchmarkUser_StatusTransition(b *testing.B) {
	user := &User{Status: UserStatusActive}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user.Status = UserStatusInactive
		user.Status = UserStatusActive
		user.Status = UserStatusLocked
		user.Status = UserStatusActive
	}
}
