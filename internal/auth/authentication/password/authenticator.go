package password

import (
	"context"
	"errors"

	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/internal/auth/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserLocked         = errors.New("user account is locked")
	ErrUserInactive       = errors.New("user account is inactive")
)

// Authenticator 密码认证器
type Authenticator struct {
	userRepo  repository.UserRepository
	hasher    Hasher
	validator *Validator
}

// NewAuthenticator 创建密码认证器
func NewAuthenticator(userRepo repository.UserRepository) *Authenticator {
	return &Authenticator{
		userRepo:  userRepo,
		hasher:    NewArgon2Hasher(),
		validator: NewValidator(DefaultPolicy()),
	}
}

// Authenticate 认证用户
func (a *Authenticator) Authenticate(ctx context.Context, username, password string) (*model.User, error) {
	// 查找用户
	user, err := a.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// 检查用户状态
	if user.IsLocked() {
		return nil, ErrUserLocked
	}

	if !user.IsActive() {
		return nil, ErrUserInactive
	}

	// 验证密码
	valid, err := a.hasher.Verify(password, user.PasswordHash)
	if err != nil {
		return nil, err
	}

	if !valid {
		// 增加登录失败次数
		_ = a.userRepo.IncrementLoginAttempts(ctx, user.ID)
		return nil, ErrInvalidCredentials
	}

	// 重置登录失败次数
	_ = a.userRepo.ResetLoginAttempts(ctx, user.ID)

	return user, nil
}

// HashPassword 哈希密码
func (a *Authenticator) HashPassword(password string) (string, error) {
	// 验证密码强度
	if err := a.validator.Validate(password); err != nil {
		return "", err
	}

	return a.hasher.Hash(password)
}

// ValidatePassword 验证密码强度
func (a *Authenticator) ValidatePassword(password string) error {
	return a.validator.Validate(password)
}

// PasswordStrength 计算密码强度
func (a *Authenticator) PasswordStrength(password string) int {
	return a.validator.Strength(password)
}
