package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	authv1 "github.com/lk2023060901/go-next-erp/api/auth/v1"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/jwt"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/password"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockUserRepository mocks the user repository
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

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
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

func (m *MockUserRepository) LockUser(ctx context.Context, userID uuid.UUID, lockUntil time.Time) error {
	args := m.Called(ctx, userID, lockUntil)
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

func (m *MockUserRepository) IncrementLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) ResetLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, ip string) error {
	args := m.Called(ctx, userID, ip)
	return args.Error(0)
}

// MockSessionRepository mocks the session repository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *model.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Session), args.Error(1)
}

func (m *MockSessionRepository) FindByToken(ctx context.Context, token string) (*model.Session, error) {
	args := m.Called(ctx, token)
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

func (m *MockSessionRepository) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*model.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Session), args.Error(1)
}

func (m *MockSessionRepository) RevokeSession(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockAuditLogRepository mocks the audit log repository
type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditLogRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) FindByEventID(ctx context.Context, eventID string) (*model.AuditLog, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.AuditLog, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.AuditLog, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) ListByAction(ctx context.Context, tenantID uuid.UUID, action string, limit, offset int) ([]*model.AuditLog, error) {
	args := m.Called(ctx, tenantID, action, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) ListByTimeRange(ctx context.Context, tenantID uuid.UUID, start, end time.Time, limit, offset int) ([]*model.AuditLog, error) {
	args := m.Called(ctx, tenantID, start, end, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuditLogRepository) CountByAction(ctx context.Context, tenantID uuid.UUID, action string) (int64, error) {
	args := m.Called(ctx, tenantID, action)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuditLogRepository) CleanupOldLogs(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *MockAuditLogRepository) ListByActionWithCursor(ctx context.Context, tenantID uuid.UUID, action string, cursor *time.Time, limit int) ([]*model.AuditLog, *time.Time, bool, error) {
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

func (m *MockAuditLogRepository) ListByUserWithCursor(ctx context.Context, userID uuid.UUID, cursor *time.Time, limit int) ([]*model.AuditLog, *time.Time, bool, error) {
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

func (m *MockAuditLogRepository) ListByTenantWithCursor(ctx context.Context, tenantID uuid.UUID, cursor *time.Time, limit int) ([]*model.AuditLog, *time.Time, bool, error) {
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

// Helper function to create test authentication service
func createTestAuthService(userRepo *MockUserRepository, sessionRepo *MockSessionRepository, auditRepo *MockAuditLogRepository) *authentication.Service {
	// Use test JWT config
	jwtConfig := &jwt.Config{
		SecretKey:       "test-secret-key-for-testing-only",
		AccessTokenTTL:  24 * time.Hour,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test-issuer",
	}

	return authentication.NewService(userRepo, sessionRepo, auditRepo, jwtConfig)
}

// TestAuthAdapter_Register tests user registration via adapter
func TestAuthAdapter_Register(t *testing.T) {
	t.Run("Register successfully", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockSessionRepo := new(MockSessionRepository)
		mockAuditRepo := new(MockAuditLogRepository)

		authService := createTestAuthService(mockUserRepo, mockSessionRepo, mockAuditRepo)
		adapter := NewAuthAdapter(authService, mockUserRepo)

		tenantID := uuid.New()

		// Mock: create user successfully
		mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).
			Return(nil).Once()

		// Mock: audit log creation
		mockAuditRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.AuditLog")).
			Return(nil).Once()

		req := &authv1.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "Password123!",
			TenantId: tenantID.String(),
		}

		resp, err := adapter.Register(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.User)
		assert.Equal(t, "testuser", resp.User.Username)
		assert.Equal(t, "test@example.com", resp.User.Email)
		mockUserRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t)
	})

	t.Run("Register with existing username", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockSessionRepo := new(MockSessionRepository)
		mockAuditRepo := new(MockAuditLogRepository)

		authService := createTestAuthService(mockUserRepo, mockSessionRepo, mockAuditRepo)
		adapter := NewAuthAdapter(authService, mockUserRepo)

		tenantID := uuid.New()

		// Mock: Create returns error (username already exists)
		mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).
			Return(authentication.ErrUserAlreadyExists).Once()

		req := &authv1.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "Password123!",
			TenantId: tenantID.String(),
		}

		resp, err := adapter.Register(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockUserRepo.AssertExpectations(t)
	})
}

// TestAuthAdapter_Login tests user login via adapter
func TestAuthAdapter_Login(t *testing.T) {
	t.Run("Login successfully", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockSessionRepo := new(MockSessionRepository)
		mockAuditRepo := new(MockAuditLogRepository)

		authService := createTestAuthService(mockUserRepo, mockSessionRepo, mockAuditRepo)
		adapter := NewAuthAdapter(authService, mockUserRepo)

		userID := uuid.New()
		tenantID := uuid.New()

		// Create password hash
		hasher := password.NewArgon2Hasher()
		passwordHash, _ := hasher.Hash("password123")

		existingUser := &model.User{
			ID:           userID,
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: passwordHash,
			TenantID:     tenantID,
			Status:       model.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Mock: find user by username
		mockUserRepo.On("FindByUsername", mock.Anything, "testuser").
			Return(existingUser, nil).Once()

		// Mock: reset login attempts and update last login
		mockUserRepo.On("ResetLoginAttempts", mock.Anything, userID).
			Return(nil).Once()
		mockUserRepo.On("UpdateLastLogin", mock.Anything, userID, "192.168.1.1").
			Return(nil).Once()

		// Mock: create session
		mockSessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Session")).
			Return(nil).Once()

		// Mock: audit log creation
		mockAuditRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.AuditLog")).
			Return(nil).Once()

		req := &authv1.LoginRequest{
			Username:  "testuser",
			Password:  "password123",
			IpAddress: "192.168.1.1",
			UserAgent: "Mozilla/5.0",
		}

		resp, err := adapter.Login(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.NotNil(t, resp.User)
		assert.Equal(t, userID.String(), resp.User.Id)
		mockUserRepo.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t)
	})

	t.Run("Login with invalid password", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockSessionRepo := new(MockSessionRepository)
		mockAuditRepo := new(MockAuditLogRepository)

		authService := createTestAuthService(mockUserRepo, mockSessionRepo, mockAuditRepo)
		adapter := NewAuthAdapter(authService, mockUserRepo)

		userID := uuid.New()
		tenantID := uuid.New()

		// Create password hash
		hasher := password.NewArgon2Hasher()
		passwordHash, _ := hasher.Hash("correctpassword")

		existingUser := &model.User{
			ID:           userID,
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: passwordHash,
			TenantID:     tenantID,
			Status:       model.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Mock: find user by username
		mockUserRepo.On("FindByUsername", mock.Anything, "testuser").
			Return(existingUser, nil).Once()

		// Mock: increment login attempts
		mockUserRepo.On("IncrementLoginAttempts", mock.Anything, userID).
			Return(nil).Once()

		// Mock: audit log for failed login
		mockAuditRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.AuditLog")).
			Return(nil).Once()

		req := &authv1.LoginRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}

		resp, err := adapter.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockUserRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t)
	})
}

// TestAuthAdapter_Logout tests user logout via adapter
func TestAuthAdapter_Logout(t *testing.T) {
	t.Run("Logout successfully", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockSessionRepo := new(MockSessionRepository)
		mockAuditRepo := new(MockAuditLogRepository)

		authService := createTestAuthService(mockUserRepo, mockSessionRepo, mockAuditRepo)
		adapter := NewAuthAdapter(authService, mockUserRepo)

		sessionID := uuid.New()
		userID := uuid.New()
		tenantID := uuid.New()

		// Create a valid token
		jwtConfig := &jwt.Config{
			SecretKey:       "test-secret-key-for-testing-only",
			AccessTokenTTL:  24 * time.Hour,
			RefreshTokenTTL: 7 * 24 * time.Hour,
			Issuer:          "test-issuer",
		}
		jwtManager := jwt.NewManager(jwtConfig)
		validToken, _ := jwtManager.GenerateAccessToken(userID, tenantID, "testuser", "test@example.com")

		// Mock: find and revoke session
		mockSessionRepo.On("FindByToken", mock.Anything, validToken).
			Return(&model.Session{
				ID:     sessionID,
				UserID: userID,
				Token:  validToken,
			}, nil).Once()
		mockSessionRepo.On("RevokeSession", mock.Anything, sessionID).
			Return(nil).Once()
		mockAuditRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.AuditLog")).
			Return(nil).Once()

		req := &authv1.LogoutRequest{
			Token:     validToken,
			IpAddress: "192.168.1.1",
			UserAgent: "Mozilla/5.0",
		}

		resp, err := adapter.Logout(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockSessionRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t)
	})
}

// TestAuthAdapter_GetCurrentUser tests getting current user via adapter
func TestAuthAdapter_GetCurrentUser(t *testing.T) {
	t.Run("GetCurrentUser successfully", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockSessionRepo := new(MockSessionRepository)
		mockAuditRepo := new(MockAuditLogRepository)

		authService := createTestAuthService(mockUserRepo, mockSessionRepo, mockAuditRepo)
		adapter := NewAuthAdapter(authService, mockUserRepo)

		userID := uuid.New()
		tenantID := uuid.New()
		expectedUser := &model.User{
			ID:        userID,
			Username:  "testuser",
			Email:     "test@example.com",
			TenantID:  tenantID,
			Status:    model.UserStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockUserRepo.On("FindByID", mock.Anything, userID).
			Return(expectedUser, nil).Once()

		// Create context with user ID
		ctx := context.WithValue(context.Background(), "user_id", userID)

		resp, err := adapter.GetCurrentUser(ctx, &emptypb.Empty{})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, userID.String(), resp.Id)
		assert.Equal(t, expectedUser.Username, resp.Username)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("GetCurrentUser without user_id in context", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockSessionRepo := new(MockSessionRepository)
		mockAuditRepo := new(MockAuditLogRepository)

		authService := createTestAuthService(mockUserRepo, mockSessionRepo, mockAuditRepo)
		adapter := NewAuthAdapter(authService, mockUserRepo)

		resp, err := adapter.GetCurrentUser(context.Background(), &emptypb.Empty{})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, authentication.ErrInvalidCredentials, err)
	})
}

// TestAuthAdapter_ChangePassword tests password change via adapter
func TestAuthAdapter_ChangePassword(t *testing.T) {
	t.Run("ChangePassword successfully", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockSessionRepo := new(MockSessionRepository)
		mockAuditRepo := new(MockAuditLogRepository)

		authService := createTestAuthService(mockUserRepo, mockSessionRepo, mockAuditRepo)
		adapter := NewAuthAdapter(authService, mockUserRepo)

		userID := uuid.New()

		// Create password hash for old password
		hasher := password.NewArgon2Hasher()
		oldPasswordHash, _ := hasher.Hash("old-password")

		existingUser := &model.User{
			ID:           userID,
			Username:     "testuser",
			PasswordHash: oldPasswordHash,
			Status:       model.UserStatusActive,
		}

		// Mock: find user by ID
		mockUserRepo.On("FindByID", mock.Anything, userID).
			Return(existingUser, nil).Once()

		// Mock: update user with new password
		mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.User")).
			Return(nil).Once()

		// Mock: revoke all user sessions after password change
		mockSessionRepo.On("RevokeUserSessions", mock.Anything, userID).
			Return(nil).Once()

		// Mock: audit log for password change
		mockAuditRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.AuditLog")).
			Return(nil).Once()

		// Create context with user ID
		ctx := context.WithValue(context.Background(), "user_id", userID)

		req := &authv1.ChangePasswordRequest{
			OldPassword: "old-password",
			NewPassword: "NewPassword123!",
		}

		resp, err := adapter.ChangePassword(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockUserRepo.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
		mockAuditRepo.AssertExpectations(t)
	})

	t.Run("ChangePassword with wrong old password", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockSessionRepo := new(MockSessionRepository)
		mockAuditRepo := new(MockAuditLogRepository)

		authService := createTestAuthService(mockUserRepo, mockSessionRepo, mockAuditRepo)
		adapter := NewAuthAdapter(authService, mockUserRepo)

		userID := uuid.New()

		// Create password hash
		hasher := password.NewArgon2Hasher()
		oldPasswordHash, _ := hasher.Hash("correct-old-password")

		existingUser := &model.User{
			ID:           userID,
			Username:     "testuser",
			PasswordHash: oldPasswordHash,
			Status:       model.UserStatusActive,
		}

		// Mock: find user by ID
		mockUserRepo.On("FindByID", mock.Anything, userID).
			Return(existingUser, nil).Once()

		ctx := context.WithValue(context.Background(), "user_id", userID)

		req := &authv1.ChangePasswordRequest{
			OldPassword: "wrong-old-password",
			NewPassword: "NewPassword123!",
		}

		resp, err := adapter.ChangePassword(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockUserRepo.AssertExpectations(t)
	})
}
