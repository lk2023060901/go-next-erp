package adapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	authv1 "github.com/lk2023060901/go-next-erp/api/auth/v1"
	"github.com/lk2023060901/go-next-erp/internal/auth/authorization"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/internal/auth/repository"
	"google.golang.org/protobuf/types/known/emptypb"
)

// UserAdapter 用户服务适配器 (实现 UserServiceServer)
type UserAdapter struct {
	authv1.UnimplementedUserServiceServer
	userRepo    repository.UserRepository
	roleRepo    repository.RoleRepository
	authzService *authorization.Service
}

// NewUserAdapter 创建用户服务适配器
func NewUserAdapter(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	authzService *authorization.Service,
) *UserAdapter {
	return &UserAdapter{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		authzService: authzService,
	}
}

// 字段白名单 (安全性)
var allowedUserFields = map[string]bool{
	"username":    true,
	"email":       true,
	"status":      true,
	"created_at":  true,
	"updated_at":  true,
	"tenant_id":   true,
}

// ListUsers 查询用户列表 (支持分页、排序、筛选)
func (a *UserAdapter) ListUsers(ctx context.Context, req *authv1.ListUsersRequest) (*authv1.ListUsersResponse, error) {
	// 从上下文中获取租户 ID
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		tenantID = uuid.Nil
	}

	// 解析分页参数
	page := int(req.Page.Page)
	pageSize := int(req.Page.PageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// TODO: 实现高级筛选 (FilterGroup)
	// 当前只实现基础分页查询
	users, err := a.userRepo.ListByTenant(ctx, tenantID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	total, err := a.userRepo.CountByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf 响应
	userInfos := make([]*authv1.UserInfo, 0, len(users))
	for _, user := range users {
		userInfos = append(userInfos, convertUserToProto(user))
	}

	return &authv1.ListUsersResponse{
		Users: userInfos,
		Page: &authv1.PageResponse{
			Total:       total,
			Page:        int32(page),
			PageSize:    int32(pageSize),
			TotalPages:  int32((total + int64(pageSize) - 1) / int64(pageSize)),
		},
	}, nil
}

// GetUser 获取用户详情
func (a *UserAdapter) GetUser(ctx context.Context, req *authv1.GetUserRequest) (*authv1.UserInfo, error) {
	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	user, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return convertUserToProto(user), nil
}

// UpdateUser 更新用户信息
func (a *UserAdapter) UpdateUser(ctx context.Context, req *authv1.UpdateUserRequest) (*authv1.UserInfo, error) {
	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	// 查询用户
	user, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Status != "" {
		user.Status = model.UserStatus(req.Status)
	}

	// 保存更新
	if err := a.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return convertUserToProto(user), nil
}

// DeleteUser 删除用户
func (a *UserAdapter) DeleteUser(ctx context.Context, req *authv1.DeleteUserRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	if err := a.userRepo.Delete(ctx, userID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// GetUserRoles 获取用户的角色列表
func (a *UserAdapter) GetUserRoles(ctx context.Context, req *authv1.GetUserRolesRequest) (*authv1.GetUserRolesResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	roles, err := a.authzService.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf 响应
	roleInfos := make([]*authv1.RoleInfo, 0, len(roles))
	for _, role := range roles {
		roleInfos = append(roleInfos, convertRoleToProto(role))
	}

	return &authv1.GetUserRolesResponse{
		Roles: roleInfos,
	}, nil
}

// GetUserPermissions 获取用户的权限列表
func (a *UserAdapter) GetUserPermissions(ctx context.Context, req *authv1.GetUserPermissionsRequest) (*authv1.GetUserPermissionsResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	permissions, err := a.authzService.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf 响应
	permissionInfos := make([]*authv1.PermissionInfo, 0, len(permissions))
	for _, perm := range permissions {
		permissionInfos = append(permissionInfos, convertPermissionToProto(perm))
	}

	return &authv1.GetUserPermissionsResponse{
		Permissions: permissionInfos,
	}, nil
}

// convertUserToProto 转换用户模型为 Protobuf
func convertUserToProto(user *model.User) *authv1.UserInfo {
	return &authv1.UserInfo{
		Id:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		Status:    string(user.Status),
		TenantId:  user.TenantID.String(),
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}
}

// convertRoleToProto 转换角色模型为 Protobuf
func convertRoleToProto(role *model.Role) *authv1.RoleInfo {
	return &authv1.RoleInfo{
		Id:          role.ID.String(),
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		TenantId:    role.TenantID.String(),
		CreatedAt:   role.CreatedAt.Unix(),
	}
}

// convertPermissionToProto 转换权限模型为 Protobuf
func convertPermissionToProto(perm *model.Permission) *authv1.PermissionInfo {
	return &authv1.PermissionInfo{
		Id:          perm.ID.String(),
		Resource:    perm.Resource,
		Action:      perm.Action,
		Description: perm.Description,
	}
}

// validateSortField 验证排序字段 (白名单)
func validateSortField(field string) error {
	// 规范化字段名 (小写 + 去空格)
	field = strings.TrimSpace(strings.ToLower(field))

	if !allowedUserFields[field] {
		return fmt.Errorf("invalid sort field: %s", field)
	}

	return nil
}
