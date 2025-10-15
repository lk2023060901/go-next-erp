package authentication

import (
	"context"
	"errors"
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

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.User, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
}

func (m *MockUserRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) LockUser(ctx context.Context, userID uuid.UUID, until time.Time) error {
	args := m.Called(ctx, userID, until)
	return args.Error(0)
}

func (m *MockUserRepository) UnlockUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) ListUsersByRole(ctx context.Context, roleID uuid.UUID) ([]*model.User, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.User), args.Error(1)
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

func (m *MockSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *model.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) FindByEventID(ctx context.Context, eventID string) (*model.AuditLog, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.AuditLog, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.AuditLog, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) ListByAction(ctx context.Context, tenantID uuid.UUID, action string, limit, offset int) ([]*model.AuditLog, error) {
	args := m.Called(ctx, tenantID, action, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) ListByTimeRange(ctx context.Context, tenantID uuid.UUID, start, end time.Time, limit, offset int) ([]*model.AuditLog, error) {
	args := m.Called(ctx, tenantID, start, end, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuditRepository) CountByAction(ctx context.Context, tenantID uuid.UUID, action string) (int64, error) {
	args := m.Called(ctx, tenantID, action)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuditRepository) CleanupOldLogs(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *MockAuditRepository) ListByActionWithCursor(ctx context.Context, tenantID uuid.UUID, action string, cursor *time.Time, limit int) ([]*model.AuditLog, *time.Time, bool, error) {
	args := m.Called(ctx, tenantID, action, cursor, limit)
	if args.Get(0) == nil {
		return nil, nil, false, args.Error(3)
	}
	nextCursor := args.Get(1)
	hasNext := args.Bool(2)
	if nextCursor == nil {
		return args.Get(0).([]*model.AuditLog), nil, hasNext, args.Error(3)
	}
	return args.Get(0).([]*model.AuditLog), nextCursor.(*time.Time), hasNext, args.Error(3)
}

func (m *MockAuditRepository) ListByUserWithCursor(ctx context.Context, userID uuid.UUID, cursor *time.Time, limit int) ([]*model.AuditLog, *time.Time, bool, error) {
	args := m.Called(ctx, userID, cursor, limit)
	if args.Get(0) == nil {
		return nil, nil, false, args.Error(3)
	}
	nextCursor := args.Get(1)
	hasNext := args.Bool(2)
	if nextCursor == nil {
		return args.Get(0).([]*model.AuditLog), nil, hasNext, args.Error(3)
	}
	return args.Get(0).([]*model.AuditLog), nextCursor.(*time.Time), hasNext, args.Error(3)
}

func (m *MockAuditRepository) ListByTenantWithCursor(ctx context.Context, tenantID uuid.UUID, cursor *time.Time, limit int) ([]*model.AuditLog, *time.Time, bool, error) {
	args := m.Called(ctx, tenantID, cursor, limit)
	if args.Get(0) == nil {
		return nil, nil, false, args.Error(3)
	}
	nextCursor := args.Get(1)
	hasNext := args.Bool(2)
	if nextCursor == nil {
		return args.Get(0).([]*model.AuditLog), nil, hasNext, args.Error(3)
	}
	return args.Get(0).([]*model.AuditLog), nextCursor.(*time.Time), hasNext, args.Error(3)
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

// TestService_Logout 测试登出
func TestService_Logout(t *testing.T) {
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	auditRepo := new(MockAuditRepository)

	jwtConfig := &jwt.Config{
		SecretKey:       "test-secret-key",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
	}

	service := NewService(userRepo, sessionRepo, auditRepo, jwtConfig)

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	// 生成token
	token, _ := service.jwtManager.GenerateAccessToken(userID, tenantID, "testuser", "test@example.com")

	t.Run("成功登出", func(t *testing.T) {
		session := &model.Session{
			ID:       uuid.New(),
			UserID:   userID,
			TenantID: tenantID,
			Token:    token,
		}

		sessionRepo.On("FindByToken", ctx, token).Return(session, nil).Once()
		sessionRepo.On("RevokeSession", ctx, session.ID).Return(nil).Once()
		auditRepo.On("Create", ctx, mock.AnythingOfType("*model.AuditLog")).Return(nil).Once()

		err := service.Logout(ctx, token, "127.0.0.1", "test-agent")
		assert.NoError(t, err)

		sessionRepo.AssertExpectations(t)
		auditRepo.AssertExpectations(t)
	})

	t.Run("会话不存在", func(t *testing.T) {
		sessionRepo.On("FindByToken", ctx, token).Return(nil, ErrSessionNotFound).Once()

		err := service.Logout(ctx, token, "127.0.0.1", "test-agent")
		assert.Error(t, err)
		assert.Equal(t, ErrSessionNotFound, err)

		sessionRepo.AssertExpectations(t)
	})
}

// TestService_ValidateToken 测试验证令牌
func TestService_ValidateToken(t *testing.T) {
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	auditRepo := new(MockAuditRepository)

	jwtConfig := &jwt.Config{
		SecretKey:       "test-secret-key",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
	}

	service := NewService(userRepo, sessionRepo, auditRepo, jwtConfig)

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	// 生成token
	token, _ := service.jwtManager.GenerateAccessToken(userID, tenantID, "testuser", "test@example.com")

	t.Run("有效令牌", func(t *testing.T) {
		session := &model.Session{
			ID:        uuid.New(),
			UserID:    userID,
			TenantID:  tenantID,
			Token:     token,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		sessionRepo.On("FindByToken", ctx, token).Return(session, nil).Once()

		claims, err := service.ValidateToken(ctx, token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, tenantID, claims.TenantID)

		sessionRepo.AssertExpectations(t)
	})

	t.Run("会话不存在", func(t *testing.T) {
		sessionRepo.On("FindByToken", ctx, token).Return(nil, ErrSessionNotFound).Once()

		_, err := service.ValidateToken(ctx, token)
		assert.Error(t, err)
		assert.Equal(t, ErrSessionNotFound, err)

		sessionRepo.AssertExpectations(t)
	})

	t.Run("会话已过期", func(t *testing.T) {
		expiredSession := &model.Session{
			ID:        uuid.New(),
			UserID:    userID,
			TenantID:  tenantID,
			Token:     token,
			ExpiresAt: time.Now().Add(-1 * time.Hour), // 过期
		}

		sessionRepo.On("FindByToken", ctx, token).Return(expiredSession, nil).Once()

		_, err := service.ValidateToken(ctx, token)
		assert.Error(t, err)
		assert.Equal(t, ErrSessionExpired, err)

		sessionRepo.AssertExpectations(t)
	})

	t.Run("会话已撤销", func(t *testing.T) {
		now := time.Now()
		revokedSession := &model.Session{
			ID:        uuid.New(),
			UserID:    userID,
			TenantID:  tenantID,
			Token:     token,
			ExpiresAt: time.Now().Add(1 * time.Hour),
			RevokedAt: &now, // 已撤销
		}

		sessionRepo.On("FindByToken", ctx, token).Return(revokedSession, nil).Once()

		_, err := service.ValidateToken(ctx, token)
		assert.Error(t, err)
		assert.Equal(t, ErrSessionRevoked, err)

		sessionRepo.AssertExpectations(t)
	})
}

// TestService_RefreshToken 测试刷新令牌
func TestService_RefreshToken(t *testing.T) {
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	auditRepo := new(MockAuditRepository)

	jwtConfig := &jwt.Config{
		SecretKey:       "test-secret-key",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
	}

	service := NewService(userRepo, sessionRepo, auditRepo, jwtConfig)

	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	user := &model.User{
		ID:       userID,
		TenantID: tenantID,
		Username: "testuser",
		Email:    "test@example.com",
		Status:   model.UserStatusActive,
	}

	// 生成refresh token
	refreshToken, _ := service.jwtManager.GenerateRefreshToken(userID, tenantID)

	t.Run("成功刷新", func(t *testing.T) {
		userRepo.On("FindByID", ctx, userID).Return(user, nil).Once()
		sessionRepo.On("Create", ctx, mock.AnythingOfType("*model.Session")).Return(nil).Once()

		response, err := service.RefreshToken(ctx, refreshToken)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.Equal(t, refreshToken, response.RefreshToken)
		assert.Equal(t, user, response.User)

		userRepo.AssertExpectations(t)
		sessionRepo.AssertExpectations(t)
	})

	t.Run("无效的refresh token", func(t *testing.T) {
		_, err := service.RefreshToken(ctx, "invalid-token")
		assert.Error(t, err)
	})

	t.Run("用户不存在", func(t *testing.T) {
		userRepo.On("FindByID", ctx, userID).Return(nil, errors.New("user not found")).Once()

		_, err := service.RefreshToken(ctx, refreshToken)
		assert.Error(t, err)

		userRepo.AssertExpectations(t)
	})
}

// TestService_RevokeSession 测试撤销会话
func TestService_RevokeSession(t *testing.T) {
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	auditRepo := new(MockAuditRepository)

	jwtConfig := &jwt.Config{
		SecretKey:       "test-secret-key",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
	}

	service := NewService(userRepo, sessionRepo, auditRepo, jwtConfig)

	ctx := context.Background()
	sessionID := uuid.New()

	t.Run("成功撤销", func(t *testing.T) {
		sessionRepo.On("RevokeSession", ctx, sessionID).Return(nil).Once()

		err := service.RevokeSession(ctx, sessionID)
		assert.NoError(t, err)

		sessionRepo.AssertExpectations(t)
	})

	t.Run("撤销失败", func(t *testing.T) {
		sessionRepo.On("RevokeSession", ctx, sessionID).Return(ErrSessionNotFound).Once()

		err := service.RevokeSession(ctx, sessionID)
		assert.Error(t, err)
		assert.Equal(t, ErrSessionNotFound, err)

		sessionRepo.AssertExpectations(t)
	})
}

// 运行所有测试
func TestAll_Success(t *testing.T) {
	t.Run("用户注册", TestService_Register_Success)
	t.Run("用户登录", TestService_Login_Success)
	t.Run("登出", TestService_Logout)
	t.Run("验证令牌", TestService_ValidateToken)
	t.Run("刷新令牌", TestService_RefreshToken)
	t.Run("撤销会话", TestService_RevokeSession)
	t.Run("JWT管理", TestJWT_Success)
	t.Run("密码哈希", TestPasswordHasher_Success)
	t.Run("密码验证", TestPasswordValidator_Success)
	t.Run("会话管理", TestService_SessionManagement_Success)
	t.Run("修改密码", TestService_ChangePassword_Success)

	t.Log("\n========================================")
	t.Log("✅ 所有认证模块测试通过！")
	t.Log("========================================")
}
