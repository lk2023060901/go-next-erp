package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewManager 测试创建JWT管理器
func TestNewManager(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   *Manager
	}{
		{
			name: "正常配置",
			config: &Config{
				SecretKey:       "test-secret-key",
				AccessTokenTTL:  15 * time.Minute,
				RefreshTokenTTL: 7 * 24 * time.Hour,
				Issuer:          "test-issuer",
			},
			want: &Manager{
				secretKey:       []byte("test-secret-key"),
				accessTokenTTL:  15 * time.Minute,
				refreshTokenTTL: 7 * 24 * time.Hour,
				issuer:          "test-issuer",
			},
		},
		{
			name: "空密钥",
			config: &Config{
				SecretKey:       "",
				AccessTokenTTL:  15 * time.Minute,
				RefreshTokenTTL: 7 * 24 * time.Hour,
				Issuer:          "test-issuer",
			},
			want: &Manager{
				secretKey:       []byte(""),
				accessTokenTTL:  15 * time.Minute,
				refreshTokenTTL: 7 * 24 * time.Hour,
				issuer:          "test-issuer",
			},
		},
		{
			name: "极短TTL",
			config: &Config{
				SecretKey:       "test-secret-key",
				AccessTokenTTL:  1 * time.Nanosecond,
				RefreshTokenTTL: 1 * time.Nanosecond,
				Issuer:          "test-issuer",
			},
			want: &Manager{
				secretKey:       []byte("test-secret-key"),
				accessTokenTTL:  1 * time.Nanosecond,
				refreshTokenTTL: 1 * time.Nanosecond,
				issuer:          "test-issuer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewManager(tt.config)
			assert.Equal(t, tt.want.secretKey, got.secretKey)
			assert.Equal(t, tt.want.accessTokenTTL, got.accessTokenTTL)
			assert.Equal(t, tt.want.refreshTokenTTL, got.refreshTokenTTL)
			assert.Equal(t, tt.want.issuer, got.issuer)
		})
	}
}

// TestGenerateAccessToken 测试生成访问令牌
func TestGenerateAccessToken(t *testing.T) {
	manager := NewManager(&Config{
		SecretKey:       "test-secret-key-for-access-token",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()
	username := "testuser"
	email := "test@example.com"

	t.Run("正常生成访问令牌", func(t *testing.T) {
		token, err := manager.GenerateAccessToken(userID, tenantID, username, email)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// 验证令牌
		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, tenantID, claims.TenantID)
		assert.Equal(t, username, claims.Username)
		assert.Equal(t, email, claims.Email)
		assert.Nil(t, claims.Metadata)
		assert.Equal(t, "test-issuer", claims.Issuer)
	})

	t.Run("空用户名和邮箱", func(t *testing.T) {
		token, err := manager.GenerateAccessToken(userID, tenantID, "", "")
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, "", claims.Username)
		assert.Equal(t, "", claims.Email)
	})

	t.Run("Nil UUID", func(t *testing.T) {
		token, err := manager.GenerateAccessToken(uuid.Nil, uuid.Nil, username, email)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, uuid.Nil, claims.UserID)
		assert.Equal(t, uuid.Nil, claims.TenantID)
	})

	t.Run("特殊字符用户名和邮箱", func(t *testing.T) {
		specialUsername := "用户@#$%^&*()"
		specialEmail := "测试+tag@example.com"
		token, err := manager.GenerateAccessToken(userID, tenantID, specialUsername, specialEmail)
		require.NoError(t, err)

		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, specialUsername, claims.Username)
		assert.Equal(t, specialEmail, claims.Email)
	})
}

// TestGenerateAccessTokenWithMetadata 测试生成带元数据的访问令牌
func TestGenerateAccessTokenWithMetadata(t *testing.T) {
	manager := NewManager(&Config{
		SecretKey:       "test-secret-key-for-metadata",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("带元数据", func(t *testing.T) {
		metadata := map[string]interface{}{
			"employee_id": "emp-123",
			"org_id":      "org-456",
			"org_path":    "/company/dept/team",
			"roles":       []string{"admin", "user"},
			"level":       5,
		}

		token, err := manager.GenerateAccessTokenWithMetadata(userID, tenantID, "user", "user@example.com", metadata)
		require.NoError(t, err)

		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, "emp-123", claims.Metadata["employee_id"])
		assert.Equal(t, "org-456", claims.Metadata["org_id"])
		assert.Equal(t, "/company/dept/team", claims.Metadata["org_path"])
		assert.Equal(t, float64(5), claims.Metadata["level"])
	})

	t.Run("空元数据", func(t *testing.T) {
		token, err := manager.GenerateAccessTokenWithMetadata(userID, tenantID, "user", "user@example.com", nil)
		require.NoError(t, err)

		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.Nil(t, claims.Metadata)
	})

	t.Run("空map元数据", func(t *testing.T) {
		metadata := map[string]interface{}{}
		token, err := manager.GenerateAccessTokenWithMetadata(userID, tenantID, "user", "user@example.com", metadata)
		require.NoError(t, err)

		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		// 空map在JSON序列化后可能变为nil
		if claims.Metadata != nil {
			assert.Empty(t, claims.Metadata)
		}
	})

	t.Run("复杂嵌套元数据", func(t *testing.T) {
		metadata := map[string]interface{}{
			"nested": map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": "value",
				},
			},
			"array": []interface{}{1, "two", true},
		}

		token, err := manager.GenerateAccessTokenWithMetadata(userID, tenantID, "user", "user@example.com", metadata)
		require.NoError(t, err)

		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.NotNil(t, claims.Metadata["nested"])
		assert.NotNil(t, claims.Metadata["array"])
	})
}

// TestGenerateRefreshToken 测试生成刷新令牌
func TestGenerateRefreshToken(t *testing.T) {
	manager := NewManager(&Config{
		SecretKey:       "test-secret-key-for-refresh",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("正常生成刷新令牌", func(t *testing.T) {
		token, err := manager.GenerateRefreshToken(userID, tenantID)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// 验证令牌
		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, tenantID, claims.TenantID)
		assert.Empty(t, claims.Username)
		assert.Empty(t, claims.Email)
	})

	t.Run("Nil UUID", func(t *testing.T) {
		token, err := manager.GenerateRefreshToken(uuid.Nil, uuid.Nil)
		require.NoError(t, err)

		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, uuid.Nil, claims.UserID)
		assert.Equal(t, uuid.Nil, claims.TenantID)
	})
}

// TestValidateToken 测试验证令牌
func TestValidateToken(t *testing.T) {
	manager := NewManager(&Config{
		SecretKey:       "test-secret-key-for-validation",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("验证有效令牌", func(t *testing.T) {
		token, err := manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
		require.NoError(t, err)

		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, tenantID, claims.TenantID)
	})

	t.Run("验证空令牌", func(t *testing.T) {
		_, err := manager.ValidateToken("")
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("验证格式错误的令牌", func(t *testing.T) {
		_, err := manager.ValidateToken("invalid.token.format")
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("验证错误签名的令牌", func(t *testing.T) {
		wrongManager := NewManager(&Config{
			SecretKey:       "wrong-secret-key",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
			Issuer:          "test-issuer",
		})

		token, err := wrongManager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
		require.NoError(t, err)

		_, err = manager.ValidateToken(token)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("验证过期令牌", func(t *testing.T) {
		expiredManager := NewManager(&Config{
			SecretKey:       "test-secret-key",
			AccessTokenTTL:  -1 * time.Hour, // 负数会立即过期
			RefreshTokenTTL: 7 * 24 * time.Hour,
			Issuer:          "test-issuer",
		})

		token, err := expiredManager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
		require.NoError(t, err)

		_, err = expiredManager.ValidateToken(token)
		assert.ErrorIs(t, err, ErrExpiredToken)
	})

	t.Run("验证错误算法的令牌", func(t *testing.T) {
		// 手动创建使用RS256的令牌（HMAC管理器期望HS256）
		claims := &Claims{
			UserID:   userID,
			TenantID: tenantID,
			Username: "user",
			Email:    "user@example.com",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "test-issuer",
			},
		}

		// 使用不同的算法签名
		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		_, err = manager.ValidateToken(tokenString)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("验证NotBefore未到的令牌", func(t *testing.T) {
		futureManager := NewManager(&Config{
			SecretKey:       "test-secret-key",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
			Issuer:          "test-issuer",
		})

		// 手动创建NotBefore在未来的令牌
		claims := &Claims{
			UserID:   userID,
			TenantID: tenantID,
			Username: "user",
			Email:    "user@example.com",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // 未来1小时
				Issuer:    "test-issuer",
				Subject:   userID.String(),
				ID:        uuid.New().String(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret-key"))
		require.NoError(t, err)

		_, err = futureManager.ValidateToken(tokenString)
		assert.ErrorIs(t, err, ErrTokenNotYetValid)
	})

	t.Run("验证损坏的令牌", func(t *testing.T) {
		token, err := manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
		require.NoError(t, err)

		// 破坏令牌
		corruptedToken := token[:len(token)-5] + "xxxxx"
		_, err = manager.ValidateToken(corruptedToken)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
}

// TestExtractUserID 测试提取用户ID
func TestExtractUserID(t *testing.T) {
	manager := NewManager(&Config{
		SecretKey:       "test-secret-key",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("提取有效令牌的用户ID", func(t *testing.T) {
		token, err := manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
		require.NoError(t, err)

		extractedID, err := manager.ExtractUserID(token)
		require.NoError(t, err)
		assert.Equal(t, userID, extractedID)
	})

	t.Run("提取无效令牌的用户ID", func(t *testing.T) {
		extractedID, err := manager.ExtractUserID("invalid-token")
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, extractedID)
	})

	t.Run("提取过期令牌的用户ID", func(t *testing.T) {
		expiredManager := NewManager(&Config{
			SecretKey:       "test-secret-key",
			AccessTokenTTL:  -1 * time.Hour,
			RefreshTokenTTL: 7 * 24 * time.Hour,
			Issuer:          "test-issuer",
		})

		token, err := expiredManager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
		require.NoError(t, err)

		extractedID, err := expiredManager.ExtractUserID(token)
		assert.ErrorIs(t, err, ErrExpiredToken)
		assert.Equal(t, uuid.Nil, extractedID)
	})
}

// TestExtractTenantID 测试提取租户ID
func TestExtractTenantID(t *testing.T) {
	manager := NewManager(&Config{
		SecretKey:       "test-secret-key",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("提取有效令牌的租户ID", func(t *testing.T) {
		token, err := manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
		require.NoError(t, err)

		extractedID, err := manager.ExtractTenantID(token)
		require.NoError(t, err)
		assert.Equal(t, tenantID, extractedID)
	})

	t.Run("提取无效令牌的租户ID", func(t *testing.T) {
		extractedID, err := manager.ExtractTenantID("invalid-token")
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, extractedID)
	})
}

// TestTokenLifecycle 测试令牌生命周期
func TestTokenLifecycle(t *testing.T) {
	t.Run("令牌在过期前有效", func(t *testing.T) {
		manager := NewManager(&Config{
			SecretKey:       "test-secret-key",
			AccessTokenTTL:  5 * time.Second, // 5秒过期
			RefreshTokenTTL: 7 * 24 * time.Hour,
			Issuer:          "test-issuer",
		})

		userID := uuid.New()
		tenantID := uuid.New()

		token, err := manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
		require.NoError(t, err)

		// 立即验证应该成功
		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
	})

	t.Run("令牌过期后无效", func(t *testing.T) {
		manager := NewManager(&Config{
			SecretKey:       "test-secret-key",
			AccessTokenTTL:  100 * time.Millisecond, // 100ms过期
			RefreshTokenTTL: 7 * 24 * time.Hour,
			Issuer:          "test-issuer",
		})

		userID := uuid.New()
		tenantID := uuid.New()

		token, err := manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
		require.NoError(t, err)

		// 等待令牌过期
		time.Sleep(150 * time.Millisecond)

		_, err = manager.ValidateToken(token)
		assert.ErrorIs(t, err, ErrExpiredToken)
	})
}

// TestConcurrentTokenGeneration 测试并发令牌生成
func TestConcurrentTokenGeneration(t *testing.T) {
	manager := NewManager(&Config{
		SecretKey:       "test-secret-key",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()

	concurrency := 100
	tokens := make(chan string, concurrency)
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			token, err := manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
			if err != nil {
				errors <- err
				return
			}
			tokens <- token
		}()
	}

	// 收集结果
	successCount := 0
	for i := 0; i < concurrency; i++ {
		select {
		case token := <-tokens:
			assert.NotEmpty(t, token)
			successCount++
		case err := <-errors:
			t.Errorf("并发生成令牌失败: %v", err)
		}
	}

	assert.Equal(t, concurrency, successCount, "所有并发请求应该成功")
}

// TestTokenUniqueness 测试令牌唯一性
func TestTokenUniqueness(t *testing.T) {
	manager := NewManager(&Config{
		SecretKey:       "test-secret-key",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
	})

	userID := uuid.New()
	tenantID := uuid.New()

	// 生成多个令牌
	tokens := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		token, err := manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
		require.NoError(t, err)

		// 检查是否重复
		if tokens[token] {
			t.Errorf("发现重复的令牌: %s", token)
		}
		tokens[token] = true
	}

	assert.Equal(t, 1000, len(tokens), "应该生成1000个不同的令牌")
}

// TestEdgeCases 测试边界情况
func TestEdgeCases(t *testing.T) {
	t.Run("极长的用户名和邮箱", func(t *testing.T) {
		manager := NewManager(&Config{
			SecretKey:       "test-secret-key",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
			Issuer:          "test-issuer",
		})

		userID := uuid.New()
		tenantID := uuid.New()
		longString := string(make([]byte, 10000))

		token, err := manager.GenerateAccessToken(userID, tenantID, longString, longString)
		require.NoError(t, err)

		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.Len(t, claims.Username, 10000)
		assert.Len(t, claims.Email, 10000)
	})

	t.Run("极长的密钥", func(t *testing.T) {
		longKey := string(make([]byte, 100000))
		manager := NewManager(&Config{
			SecretKey:       longKey,
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
			Issuer:          "test-issuer",
		})

		userID := uuid.New()
		tenantID := uuid.New()

		token, err := manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
		require.NoError(t, err)

		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
	})

	t.Run("零时长TTL", func(t *testing.T) {
		manager := NewManager(&Config{
			SecretKey:       "test-secret-key",
			AccessTokenTTL:  0,
			RefreshTokenTTL: 0,
			Issuer:          "test-issuer",
		})

		userID := uuid.New()
		tenantID := uuid.New()

		token, err := manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
		require.NoError(t, err)

		// 零TTL的令牌应该立即过期
		_, err = manager.ValidateToken(token)
		assert.Error(t, err)
	})
}

// TestRefreshAccessToken 测试刷新访问令牌
func TestRefreshAccessToken(t *testing.T) {
	manager := NewManager(&Config{
		SecretKey:       "test-secret-key",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
	})

	t.Run("成功刷新", func(t *testing.T) {
		userID := uuid.New()
		tenantID := uuid.New()

		// 生成refresh token
		refreshToken, err := manager.GenerateRefreshToken(userID, tenantID)
		require.NoError(t, err)
		require.NotEmpty(t, refreshToken)

		// 使用refresh token刷新access token
		newAccessToken, err := manager.RefreshAccessToken(refreshToken, "testuser", "test@example.com")
		require.NoError(t, err)
		require.NotEmpty(t, newAccessToken)

		// 验证新access token有效
		claims, err := manager.ValidateToken(newAccessToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
	})

	t.Run("无效的refresh token", func(t *testing.T) {
		_, err := manager.RefreshAccessToken("invalid-token", "user", "user@test.com")
		assert.Error(t, err)
	})

	t.Run("过期的refresh token", func(t *testing.T) {
		expiredManager := NewManager(&Config{
			SecretKey:       "test-secret-key",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: -1 * time.Hour, // 负数TTL，立即过期
			Issuer:          "test-issuer",
		})

		userID := uuid.New()
		tenantID := uuid.New()

		refreshToken, err := expiredManager.GenerateRefreshToken(userID, tenantID)
		require.NoError(t, err)

		_, err = expiredManager.RefreshAccessToken(refreshToken, "user", "user@test.com")
		assert.Error(t, err)
	})

	t.Run("使用access token刷新", func(t *testing.T) {
		userID := uuid.New()
		tenantID := uuid.New()

		// 生成access token (不是refresh token)
		accessToken, err := manager.GenerateAccessToken(userID, tenantID, "user", "user@example.com")
		require.NoError(t, err)

		// 使用access token刷新（技术上也能工作）
		newToken, err := manager.RefreshAccessToken(accessToken, "user", "user@example.com")
		// 应该返回有效token
		assert.NoError(t, err)
		assert.NotEmpty(t, newToken)
	})
}
