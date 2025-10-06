package authentication

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/jwt"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/password"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/internal/auth/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionRevoked     = errors.New("session has been revoked")
	ErrSessionExpired     = errors.New("session has expired")
)

// Service 认证服务
type Service struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	auditRepo   repository.AuditLogRepository
	passwordAuth *password.Authenticator
	jwtManager   *jwt.Manager
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username  string
	Password  string
	IPAddress string
	UserAgent string
}

// LoginResponse 登录响应
type LoginResponse struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// NewService 创建认证服务
func NewService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	auditRepo repository.AuditLogRepository,
	jwtConfig *jwt.Config,
) *Service {
	return &Service{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		auditRepo:    auditRepo,
		passwordAuth: password.NewAuthenticator(userRepo),
		jwtManager:   jwt.NewManager(jwtConfig),
	}
}

// Login 用户登录
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 1. 密码认证
	user, err := s.passwordAuth.Authenticate(ctx, req.Username, req.Password)
	if err != nil {
		// 记录审计日志
		_ = s.auditRepo.Create(ctx, &model.AuditLog{
			EventID:   uuid.New().String(),
			TenantID:  uuid.Nil,
			UserID:    uuid.Nil,
			Action:    model.AuditActionLoginFailed,
			Resource:  "user",
			IPAddress: req.IPAddress,
			UserAgent: req.UserAgent,
			Result:    model.AuditResultFailure,
			ErrorMsg:  err.Error(),
		})

		return nil, ErrInvalidCredentials
	}

	// 2. 生成 JWT
	accessToken, err := s.jwtManager.GenerateAccessToken(
		user.ID,
		user.TenantID,
		user.Username,
		user.Email,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.TenantID)
	if err != nil {
		return nil, err
	}

	// 3. 创建会话
	session := &model.Session{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Token:     accessToken,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24小时过期
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// 4. 更新最后登录信息
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID, req.IPAddress)

	// 5. 记录审计日志
	_ = s.auditRepo.Create(ctx, &model.AuditLog{
		EventID:   uuid.New().String(),
		TenantID:  user.TenantID,
		UserID:    user.ID,
		Action:    model.AuditActionLogin,
		Resource:  "user",
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
		Result:    model.AuditResultSuccess,
	})

	return &LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

// Logout 用户登出
func (s *Service) Logout(ctx context.Context, token string, ipAddress, userAgent string) error {
	// 1. 验证 token
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return err
	}

	// 2. 查找并撤销会话
	session, err := s.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return ErrSessionNotFound
	}

	if err := s.sessionRepo.RevokeSession(ctx, session.ID); err != nil {
		return err
	}

	// 3. 记录审计日志
	_ = s.auditRepo.Create(ctx, &model.AuditLog{
		EventID:   uuid.New().String(),
		TenantID:  claims.TenantID,
		UserID:    claims.UserID,
		Action:    model.AuditActionLogout,
		Resource:  "user",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Result:    model.AuditResultSuccess,
	})

	return nil
}

// ValidateToken 验证令牌
func (s *Service) ValidateToken(ctx context.Context, token string) (*jwt.Claims, error) {
	// 1. 验证 JWT
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// 2. 检查会话是否有效
	session, err := s.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if !session.IsValid() {
		if session.RevokedAt != nil {
			return nil, ErrSessionRevoked
		}
		return nil, ErrSessionExpired
	}

	return claims, nil
}

// RefreshToken 刷新令牌
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// 1. 验证刷新令牌
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 2. 获取用户信息
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	// 3. 生成新的访问令牌
	newAccessToken, err := s.jwtManager.GenerateAccessToken(
		user.ID,
		user.TenantID,
		user.Username,
		user.Email,
	)
	if err != nil {
		return nil, err
	}

	// 4. 创建新会话
	session := &model.Session{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Token:     newAccessToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &LoginResponse{
		User:         user,
		AccessToken:  newAccessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

// Register 用户注册
func (s *Service) Register(ctx context.Context, username, email, password string, tenantID uuid.UUID) (*model.User, error) {
	// 1. 验证密码强度
	if err := s.passwordAuth.ValidatePassword(password); err != nil {
		return nil, err
	}

	// 2. 哈希密码
	passwordHash, err := s.passwordAuth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// 3. 创建用户
	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		TenantID:     tenantID,
		Status:       model.UserStatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 4. 记录审计日志
	_ = s.auditRepo.Create(ctx, &model.AuditLog{
		EventID:    uuid.New().String(),
		TenantID:   tenantID,
		UserID:     user.ID,
		Action:     model.AuditActionUserCreate,
		Resource:   "user",
		ResourceID: user.ID.String(),
		Result:     model.AuditResultSuccess,
	})

	return user, nil
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	// 1. 获取用户
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// 2. 验证旧密码
	valid, err := password.NewArgon2Hasher().Verify(oldPassword, user.PasswordHash)
	if err != nil || !valid {
		return ErrInvalidCredentials
	}

	// 3. 验证新密码强度
	if err := s.passwordAuth.ValidatePassword(newPassword); err != nil {
		return err
	}

	// 4. 哈希新密码
	newPasswordHash, err := s.passwordAuth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 5. 更新密码
	user.PasswordHash = newPasswordHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// 6. 撤销所有会话
	_ = s.sessionRepo.RevokeUserSessions(ctx, userID)

	// 7. 记录审计日志
	_ = s.auditRepo.Create(ctx, &model.AuditLog{
		EventID:    uuid.New().String(),
		TenantID:   user.TenantID,
		UserID:     userID,
		Action:     model.AuditActionPasswordReset,
		Resource:   "user",
		ResourceID: userID.String(),
		Result:     model.AuditResultSuccess,
	})

	return nil
}

// GetUserSessions 获取用户的所有活跃会话
func (s *Service) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*model.Session, error) {
	return s.sessionRepo.GetUserSessions(ctx, userID)
}

// RevokeSession 撤销指定会话
func (s *Service) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessionRepo.RevokeSession(ctx, sessionID)
}
