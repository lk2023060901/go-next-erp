package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// BenchmarkGenerateAccessToken 基准测试：生成访问令牌
func BenchmarkGenerateAccessToken(b *testing.B) {
	manager := NewManager(&Config{
		SecretKey:       "benchmark-secret-key-32-bytes-long",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "benchmark-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()
	username := "benchuser"
	email := "bench@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.GenerateAccessToken(userID, tenantID, username, email)
	}
}

// BenchmarkGenerateAccessTokenWithMetadata 基准测试：生成带元数据的访问令牌
func BenchmarkGenerateAccessTokenWithMetadata(b *testing.B) {
	manager := NewManager(&Config{
		SecretKey:       "benchmark-secret-key-32-bytes-long",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "benchmark-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()
	metadata := map[string]interface{}{
		"employee_id": "emp-123",
		"org_id":      "org-456",
		"roles":       []string{"admin", "user"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.GenerateAccessTokenWithMetadata(userID, tenantID, "user", "user@example.com", metadata)
	}
}

// BenchmarkGenerateRefreshToken 基准测试：生成刷新令牌
func BenchmarkGenerateRefreshToken(b *testing.B) {
	manager := NewManager(&Config{
		SecretKey:       "benchmark-secret-key-32-bytes-long",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "benchmark-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.GenerateRefreshToken(userID, tenantID)
	}
}

// BenchmarkValidateToken 基准测试：验证令牌
func BenchmarkValidateToken(b *testing.B) {
	manager := NewManager(&Config{
		SecretKey:       "benchmark-secret-key-32-bytes-long",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "benchmark-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()
	token, _ := manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.ValidateToken(token)
	}
}

// BenchmarkExtractUserID 基准测试：提取用户ID
func BenchmarkExtractUserID(b *testing.B) {
	manager := NewManager(&Config{
		SecretKey:       "benchmark-secret-key-32-bytes-long",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "benchmark-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()
	token, _ := manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.ExtractUserID(token)
	}
}

// BenchmarkConcurrentTokenGeneration 基准测试：并发令牌生成
func BenchmarkConcurrentTokenGeneration(b *testing.B) {
	manager := NewManager(&Config{
		SecretKey:       "benchmark-secret-key-32-bytes-long",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "benchmark-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
		}
	})
}

// BenchmarkConcurrentTokenValidation 基准测试：并发令牌验证
func BenchmarkConcurrentTokenValidation(b *testing.B) {
	manager := NewManager(&Config{
		SecretKey:       "benchmark-secret-key-32-bytes-long",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "benchmark-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()
	token, _ := manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = manager.ValidateToken(token)
		}
	})
}
