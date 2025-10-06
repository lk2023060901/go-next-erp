package authentication

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/jwt"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/password"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ==================== Mock Repositories ====================

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, ip string) error {
	args := m.Called(ctx, userID, ip)
	return args.Error(0)
}

func (m *MockUserRepository) IncrementLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) ResetLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *model.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) FindByToken(ctx context.Context, token string) (*model.Session, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Session), args.Error(1)
}

func (m *MockSessionRepository) RevokeSession(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*model.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Session), args.Error(1)
}

type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

// ==================== 测试用例 ====================

// 测试用户注册成功
func TestService_Register_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	auditRepo := new(MockAuditRepository)

	service := NewService(userRepo, sessionRepo, auditRepo, &jwt.Config{
		SecretKey:       "test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "test",
	})

	tenantID := uuid.New()

	// Mock expectations - 确保成功
	userRepo.On("Create", ctx, mock.AnythingOfType("*model.User")).Return(nil)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*model.AuditLog")).Return(nil)

	// 执行注册
	user, err := service.Register(ctx, "testuser", "test@example.com", "SecurePass@123", tenantID)

	// 断言成功
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NotEmpty(t, user.PasswordHash)
	assert.Equal(t, tenantID, user.TenantID)

	userRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)

	t.Logf("✅ 用户注册成功: %s (ID: %s)", user.Username, user.ID)
}

// 测试用户登录成功
func TestService_Login_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	auditRepo := new(MockAuditRepository)

	service := NewService(userRepo, sessionRepo, auditRepo, &jwt.Config{
		SecretKey:       "test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "test",
	})

	// 生成真实的密码哈希
	hasher := password.NewArgon2Hasher()
	realPassword := "SecurePass@123"
	passwordHash, err := hasher.Hash(realPassword)
	assert.NoError(t, err)

	// 创建测试用户
	userID := uuid.New()
	tenantID := uuid.New()
	testUser := &model.User{
		ID:           userID,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		TenantID:     tenantID,
		Status:       model.UserStatusActive,
	}

	// Mock expectations
	userRepo.On("FindByUsername", ctx, "testuser").Return(testUser, nil)
	userRepo.On("ResetLoginAttempts", ctx, userID).Return(nil)
	userRepo.On("UpdateLastLogin", ctx, userID, "192.168.1.1").Return(nil)
	sessionRepo.On("Create", ctx, mock.AnythingOfType("*model.Session")).Return(nil)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*model.AuditLog")).Return(nil)

	// 执行登录
	loginReq := &LoginRequest{
		Username:  "testuser",
		Password:  realPassword, // 使用正确的密码
		IPAddress: "192.168.1.1",
		UserAgent: "Test Agent",
	}

	loginResp, err := service.Login(ctx, loginReq)

	// 断言成功
	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, loginResp.RefreshToken)
	assert.Equal(t, testUser, loginResp.User)

	userRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)

	t.Logf("✅ 用户登录成功: %s", testUser.Username)
	t.Logf("   Access Token: %s...", loginResp.AccessToken[:20])
	t.Logf("   过期时间: %s", loginResp.ExpiresAt.Format(time.RFC3339))
}

// 测试 JWT Token 生成和验证成功
func TestJWT_Success(t *testing.T) {
	manager := jwt.NewManager(&jwt.Config{
		SecretKey:       "test-secret-key",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "test",
	})

	userID := uuid.New()
	tenantID := uuid.New()

	// 生成 Access Token
	accessToken, err := manager.GenerateAccessToken(userID, tenantID, "testuser", "test@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	t.Logf("✅ Access Token 生成成功: %s...", accessToken[:30])

	// 生成 Refresh Token
	refreshToken, err := manager.GenerateRefreshToken(userID, tenantID)
	assert.NoError(t, err)
	assert.NotEmpty(t, refreshToken)
	t.Logf("✅ Refresh Token 生成成功: %s...", refreshToken[:30])

	// 验证 Access Token
	claims, err := manager.ValidateToken(accessToken)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, tenantID, claims.TenantID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "test@example.com", claims.Email)
	t.Logf("✅ Token 验证成功: UserID=%s, Username=%s", claims.UserID, claims.Username)

	// 提取用户 ID
	extractedUserID, err := manager.ExtractUserID(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
	t.Logf("✅ 提取 UserID 成功: %s", extractedUserID)
}

// 测试密码哈希和验证成功
func TestPasswordHasher_Success(t *testing.T) {
	hasher := password.NewArgon2Hasher()

	testPassword := "SecurePass@123"

	// 生成哈希
	hash, err := hasher.Hash(testPassword)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	t.Logf("✅ 密码哈希生成成功: %s", hash[:50]+"...")

	// 验证正确密码
	valid, err := hasher.Verify(testPassword, hash)
	assert.NoError(t, err)
	assert.True(t, valid)
	t.Logf("✅ 密码验证成功")

	// 哈希应该每次不同（因为盐不同）
	hash2, err := hasher.Hash(testPassword)
	assert.NoError(t, err)
	assert.NotEqual(t, hash, hash2)
	t.Logf("✅ 每次哈希不同（盐随机）")

	// 两个哈希都能验证成功
	valid2, err := hasher.Verify(testPassword, hash2)
	assert.NoError(t, err)
	assert.True(t, valid2)
	t.Logf("✅ 第二个哈希也验证成功")
}

// 测试密码强度验证成功
func TestPasswordValidator_Success(t *testing.T) {
	validator := password.NewValidator(password.DefaultPolicy())

	validPasswords := []string{
		"SecurePass@123",
		"MyP@ssw0rd!",
		"C0mpl3x!ty",
		"Str0ng&Valid",
	}

	for _, pwd := range validPasswords {
		err := validator.Validate(pwd)
		assert.NoError(t, err)
		t.Logf("✅ 密码通过验证: %s", pwd)

		// 计算强度
		strength := validator.Strength(pwd)
		assert.Greater(t, strength, 50) // 强度应该大于 50
		t.Logf("   密码强度: %d/100", strength)
	}
}

// 测试会话管理成功
func TestService_SessionManagement_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	auditRepo := new(MockAuditRepository)

	service := NewService(userRepo, sessionRepo, auditRepo, &jwt.Config{
		SecretKey:       "test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "test",
	})

	userID := uuid.New()

	// 创建测试会话
	testSessions := []*model.Session{
		{
			ID:        uuid.New(),
			UserID:    userID,
			TenantID:  uuid.New(),
			IPAddress: "192.168.1.1",
			UserAgent: "Chrome",
			ExpiresAt: time.Now().Add(time.Hour),
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			TenantID:  uuid.New(),
			IPAddress: "192.168.1.2",
			UserAgent: "Safari",
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}

	// Mock expectations
	sessionRepo.On("GetUserSessions", ctx, userID).Return(testSessions, nil)

	// 获取用户会话
	sessions, err := service.GetUserSessions(ctx, userID)

	// 断言成功
	assert.NoError(t, err)
	assert.Len(t, sessions, 2)
	assert.Equal(t, "192.168.1.1", sessions[0].IPAddress)
	assert.Equal(t, "Chrome", sessions[0].UserAgent)

	sessionRepo.AssertExpectations(t)

	t.Logf("✅ 获取用户会话成功 (%d 个)", len(sessions))
	for i, s := range sessions {
		t.Logf("   %d. IP: %s, 设备: %s", i+1, s.IPAddress, s.UserAgent)
	}
}

// 测试修改密码成功
func TestService_ChangePassword_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	auditRepo := new(MockAuditRepository)

	service := NewService(userRepo, sessionRepo, auditRepo, &jwt.Config{
		SecretKey: "test-secret",
	})

	// 生成旧密码哈希
	hasher := password.NewArgon2Hasher()
	oldPassword := "OldPass@123"
	oldPasswordHash, _ := hasher.Hash(oldPassword)

	userID := uuid.New()
	testUser := &model.User{
		ID:           userID,
		PasswordHash: oldPasswordHash,
		TenantID:     uuid.New(),
	}

	// Mock expectations
	userRepo.On("FindByID", ctx, userID).Return(testUser, nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*model.User")).Return(nil)
	sessionRepo.On("RevokeUserSessions", ctx, userID).Return(nil)
	auditRepo.On("Create", ctx, mock.AnythingOfType("*model.AuditLog")).Return(nil)

	// 执行修改密码
	newPassword := "NewPass@456"
	err := service.ChangePassword(ctx, userID, oldPassword, newPassword)

	// 断言成功
	assert.NoError(t, err)

	userRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)

	t.Logf("✅ 密码修改成功")
	t.Logf("   旧密码: %s", oldPassword)
	t.Logf("   新密码: %s", newPassword)
}

// 运行所有测试
func TestAll_Success(t *testing.T) {
	t.Run("用户注册", TestService_Register_Success)
	t.Run("用户登录", TestService_Login_Success)
	t.Run("JWT管理", TestJWT_Success)
	t.Run("密码哈希", TestPasswordHasher_Success)
	t.Run("密码验证", TestPasswordValidator_Success)
	t.Run("会话管理", TestService_SessionManagement_Success)
	t.Run("修改密码", TestService_ChangePassword_Success)

	t.Log("\n========================================")
	t.Log("✅ 所有认证模块测试通过！")
	t.Log("========================================")
}
