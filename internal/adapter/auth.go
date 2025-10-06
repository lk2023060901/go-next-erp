package adapter

import (
	"context"

	"github.com/google/uuid"
	authv1 "github.com/lk2023060901/go-next-erp/api/auth/v1"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication"
	"github.com/lk2023060901/go-next-erp/internal/auth/repository"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AuthAdapter 认证服务适配器 (实现 AuthServiceServer)
type AuthAdapter struct {
	authv1.UnimplementedAuthServiceServer
	authService *authentication.Service
	userRepo    repository.UserRepository
}

// NewAuthAdapter 创建认证服务适配器
func NewAuthAdapter(
	authService *authentication.Service,
	userRepo repository.UserRepository,
) *AuthAdapter {
	return &AuthAdapter{
		authService: authService,
		userRepo:    userRepo,
	}
}

// Register 注册新用户
func (a *AuthAdapter) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	// 解析租户 ID (如果没有提供，使用默认租户)
	tenantID := uuid.Nil
	if req.TenantId != "" {
		tid, err := uuid.Parse(req.TenantId)
		if err == nil {
			tenantID = tid
		}
	}

	// 调用领域服务
	user, err := a.authService.Register(ctx, req.Username, req.Email, req.Password, tenantID)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf 响应
	return &authv1.RegisterResponse{
		User: &authv1.UserInfo{
			Id:        user.ID.String(),
			Username:  user.Username,
			Email:     user.Email,
			Status:    string(user.Status),
			TenantId:  user.TenantID.String(),
			CreatedAt: user.CreatedAt.Unix(),
			UpdatedAt: user.UpdatedAt.Unix(),
		},
	}, nil
}

// Login 用户登录
func (a *AuthAdapter) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	// 构建认证请求
	loginReq := &authentication.LoginRequest{
		Username:  req.Username,
		Password:  req.Password,
		IPAddress: req.IpAddress,
		UserAgent: req.UserAgent,
	}

	// 调用领域服务
	loginResp, err := a.authService.Login(ctx, loginReq)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf 响应
	return &authv1.LoginResponse{
		AccessToken:  loginResp.AccessToken,
		RefreshToken: loginResp.RefreshToken,
		ExpiresIn:    int64(loginResp.ExpiresAt.Sub(loginResp.ExpiresAt.Add(-24 * 3600000000000)).Seconds()),
		User: &authv1.UserInfo{
			Id:        loginResp.User.ID.String(),
			Username:  loginResp.User.Username,
			Email:     loginResp.User.Email,
			Status:    string(loginResp.User.Status),
			TenantId:  loginResp.User.TenantID.String(),
			CreatedAt: loginResp.User.CreatedAt.Unix(),
			UpdatedAt: loginResp.User.UpdatedAt.Unix(),
		},
	}, nil
}

// Logout 用户登出
func (a *AuthAdapter) Logout(ctx context.Context, req *authv1.LogoutRequest) (*emptypb.Empty, error) {
	err := a.authService.Logout(ctx, req.Token, req.IpAddress, req.UserAgent)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RefreshToken 刷新访问令牌
func (a *AuthAdapter) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.LoginResponse, error) {
	// 调用领域服务
	loginResp, err := a.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf 响应
	return &authv1.LoginResponse{
		AccessToken:  loginResp.AccessToken,
		RefreshToken: loginResp.RefreshToken,
		ExpiresIn:    int64(loginResp.ExpiresAt.Sub(loginResp.ExpiresAt.Add(-24 * 3600000000000)).Seconds()),
		User: &authv1.UserInfo{
			Id:        loginResp.User.ID.String(),
			Username:  loginResp.User.Username,
			Email:     loginResp.User.Email,
			Status:    string(loginResp.User.Status),
			TenantId:  loginResp.User.TenantID.String(),
			CreatedAt: loginResp.User.CreatedAt.Unix(),
			UpdatedAt: loginResp.User.UpdatedAt.Unix(),
		},
	}, nil
}

// GetCurrentUser 获取当前用户信息
func (a *AuthAdapter) GetCurrentUser(ctx context.Context, _ *emptypb.Empty) (*authv1.UserInfo, error) {
	// 从上下文中获取用户 ID (由中间件注入)
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, authentication.ErrInvalidCredentials
	}

	// 查询用户信息
	user, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf 响应
	return &authv1.UserInfo{
		Id:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Status:    string(user.Status),
		TenantId:  user.TenantID.String(),
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}, nil
}

// ChangePassword 修改密码
func (a *AuthAdapter) ChangePassword(ctx context.Context, req *authv1.ChangePasswordRequest) (*emptypb.Empty, error) {
	// 从上下文中获取用户 ID
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, authentication.ErrInvalidCredentials
	}

	// 调用领域服务
	err := a.authService.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
