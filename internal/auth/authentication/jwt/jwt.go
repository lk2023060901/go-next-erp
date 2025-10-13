package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrTokenNotYetValid = errors.New("token not yet valid")
)

// Claims JWT 声明
type Claims struct {
	UserID   uuid.UUID              `json:"user_id"`
	TenantID uuid.UUID              `json:"tenant_id"`
	Username string                 `json:"username"`
	Email    string                 `json:"email"`
	Metadata map[string]interface{} `json:"metadata,omitempty"` // 扩展元数据(employee_id, org_id, org_path, roles等)
	jwt.RegisteredClaims
}

// Manager JWT 管理器
type Manager struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
}

// Config JWT 配置
type Config struct {
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
}

// NewManager 创建 JWT 管理器
func NewManager(config *Config) *Manager {
	return &Manager{
		secretKey:       []byte(config.SecretKey),
		accessTokenTTL:  config.AccessTokenTTL,
		refreshTokenTTL: config.RefreshTokenTTL,
		issuer:          config.Issuer,
	}
}

// GenerateAccessToken 生成访问令牌
func (m *Manager) GenerateAccessToken(userID, tenantID uuid.UUID, username, email string) (string, error) {
	return m.GenerateAccessTokenWithMetadata(userID, tenantID, username, email, nil)
}

// GenerateAccessTokenWithMetadata 生成带元数据的访问令牌
func (m *Manager) GenerateAccessTokenWithMetadata(userID, tenantID uuid.UUID, username, email string, metadata map[string]interface{}) (string, error) {
	now := time.Now()
	expiresAt := now.Add(m.accessTokenTTL)

	claims := &Claims{
		UserID:   userID,
		TenantID: tenantID,
		Username: username,
		Email:    email,
		Metadata: metadata,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// GenerateRefreshToken 生成刷新令牌
func (m *Manager) GenerateRefreshToken(userID, tenantID uuid.UUID) (string, error) {
	now := time.Now()
	expiresAt := now.Add(m.refreshTokenTTL)

	claims := &Claims{
		UserID:   userID,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   userID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// ValidateToken 验证令牌
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotYetValid
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshAccessToken 刷新访问令牌
func (m *Manager) RefreshAccessToken(refreshTokenString string, username, email string) (string, error) {
	// 验证刷新令牌
	token, err := jwt.ParseWithClaims(refreshTokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secretKey, nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}

	// 解析用户 ID
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return "", ErrInvalidToken
	}

	// 生成新的访问令牌
	// 注意：这里 tenantID 需要从其他地方获取，暂时使用空值
	return m.GenerateAccessToken(userID, uuid.Nil, username, email)
}

// ExtractUserID 从令牌中提取用户 ID
func (m *Manager) ExtractUserID(tokenString string) (uuid.UUID, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	return claims.UserID, nil
}

// ExtractTenantID 从令牌中提取租户 ID
func (m *Manager) ExtractTenantID(tokenString string) (uuid.UUID, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	return claims.TenantID, nil
}
