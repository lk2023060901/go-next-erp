package password

import (
	"errors"
	"regexp"
	"unicode"
)

var (
	ErrPasswordTooShort       = errors.New("password is too short")
	ErrPasswordTooLong        = errors.New("password is too long")
	ErrPasswordNoUppercase    = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLowercase    = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit        = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecialChar  = errors.New("password must contain at least one special character")
	ErrPasswordCommonPassword = errors.New("password is too common")
)

// Policy 密码策略
type Policy struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
	ForbidCommon   bool
}

// DefaultPolicy 默认密码策略
func DefaultPolicy() *Policy {
	return &Policy{
		MinLength:      8,
		MaxLength:      128,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
		ForbidCommon:   true,
	}
}

// Validator 密码验证器
type Validator struct {
	policy *Policy
}

// NewValidator 创建密码验证器
func NewValidator(policy *Policy) *Validator {
	if policy == nil {
		policy = DefaultPolicy()
	}

	return &Validator{
		policy: policy,
	}
}

// Validate 验证密码
func (v *Validator) Validate(password string) error {
	// 长度检查
	if len(password) < v.policy.MinLength {
		return ErrPasswordTooShort
	}

	if len(password) > v.policy.MaxLength {
		return ErrPasswordTooLong
	}

	// 字符类型检查
	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if v.policy.RequireUpper && !hasUpper {
		return ErrPasswordNoUppercase
	}

	if v.policy.RequireLower && !hasLower {
		return ErrPasswordNoLowercase
	}

	if v.policy.RequireDigit && !hasDigit {
		return ErrPasswordNoDigit
	}

	if v.policy.RequireSpecial && !hasSpecial {
		return ErrPasswordNoSpecialChar
	}

	// 常见密码检查
	if v.policy.ForbidCommon && isCommonPassword(password) {
		return ErrPasswordCommonPassword
	}

	return nil
}

// isCommonPassword 检查是否为常见密码
func isCommonPassword(password string) bool {
	// 常见密码列表（实际应该更全面）
	commonPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123",
		"monkey", "1234567", "letmein", "trustno1", "dragon",
		"baseball", "111111", "iloveyou", "master", "sunshine",
		"ashley", "bailey", "passw0rd", "shadow", "123123",
		"654321", "superman", "qazwsx", "michael", "football",
	}

	// 简单的字符串匹配（实际应使用更复杂的算法）
	passwordLower := regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(password, "")

	for _, common := range commonPasswords {
		if passwordLower == common {
			return true
		}
	}

	return false
}

// Strength 计算密码强度（0-100）
func (v *Validator) Strength(password string) int {
	score := 0

	// 长度得分
	length := len(password)
	if length >= 8 {
		score += 20
	}
	if length >= 12 {
		score += 10
	}
	if length >= 16 {
		score += 10
	}

	// 字符类型得分
	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if hasUpper {
		score += 15
	}
	if hasLower {
		score += 15
	}
	if hasDigit {
		score += 15
	}
	if hasSpecial {
		score += 15
	}

	// 常见密码扣分
	if isCommonPassword(password) {
		score -= 50
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}
