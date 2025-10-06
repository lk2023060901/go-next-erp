package adapter

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	authv1 "github.com/lk2023060901/go-next-erp/api/auth/v1"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/internal/auth/repository"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RoleAdapter 角色服务适配器 (实现 RoleServiceServer)
type RoleAdapter struct {
	authv1.UnimplementedRoleServiceServer
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
}

// NewRoleAdapter 创建角色服务适配器
func NewRoleAdapter(
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
) *RoleAdapter {
	return &RoleAdapter{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// ListRoles 查询角色列表
func (a *RoleAdapter) ListRoles(ctx context.Context, req *authv1.ListRolesRequest) (*authv1.ListRolesResponse, error) {
	// 从上下文中获取租户 ID
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		tenantID = uuid.Nil
	}

	roles, err := a.roleRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf 响应
	roleInfos := make([]*authv1.RoleInfo, 0, len(roles))
	for _, role := range roles {
		roleInfos = append(roleInfos, convertRoleToProto(role))
	}

	return &authv1.ListRolesResponse{
		Roles: roleInfos,
	}, nil
}

// GetRole 获取角色详情
func (a *RoleAdapter) GetRole(ctx context.Context, req *authv1.GetRoleRequest) (*authv1.RoleInfo, error) {
	roleID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid role id: %w", err)
	}

	role, err := a.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return convertRoleToProto(role), nil
}

// CreateRole 创建角色
func (a *RoleAdapter) CreateRole(ctx context.Context, req *authv1.CreateRoleRequest) (*authv1.RoleInfo, error) {
	// 从上下文中获取租户 ID
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		tenantID = uuid.Nil
	}

	// 构建角色模型
	role := &model.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		TenantID:    tenantID,
	}

	// 处理父角色 ID
	if req.ParentId != "" {
		parentID, err := uuid.Parse(req.ParentId)
		if err != nil {
			return nil, fmt.Errorf("invalid parent role id: %w", err)
		}
		role.ParentID = &parentID
	}

	// 创建角色
	if err := a.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	return convertRoleToProto(role), nil
}

// UpdateRole 更新角色
func (a *RoleAdapter) UpdateRole(ctx context.Context, req *authv1.UpdateRoleRequest) (*authv1.RoleInfo, error) {
	roleID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid role id: %w", err)
	}

	// 查询角色
	role, err := a.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.DisplayName != "" {
		role.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	// 保存更新
	if err := a.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}

	return convertRoleToProto(role), nil
}

// DeleteRole 删除角色
func (a *RoleAdapter) DeleteRole(ctx context.Context, req *authv1.DeleteRoleRequest) (*emptypb.Empty, error) {
	roleID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid role id: %w", err)
	}

	if err := a.roleRepo.Delete(ctx, roleID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// GetRolePermissions 获取角色的权限列表
func (a *RoleAdapter) GetRolePermissions(ctx context.Context, req *authv1.GetRolePermissionsRequest) (*authv1.GetRolePermissionsResponse, error) {
	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, fmt.Errorf("invalid role id: %w", err)
	}

	permissions, err := a.permissionRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// 转换为 Protobuf 响应
	permissionInfos := make([]*authv1.PermissionInfo, 0, len(permissions))
	for _, perm := range permissions {
		permissionInfos = append(permissionInfos, convertPermissionToProto(perm))
	}

	return &authv1.GetRolePermissionsResponse{
		Permissions: permissionInfos,
	}, nil
}

// AssignPermissions 分配权限给角色
func (a *RoleAdapter) AssignPermissions(ctx context.Context, req *authv1.AssignPermissionsRequest) (*emptypb.Empty, error) {
	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, fmt.Errorf("invalid role id: %w", err)
	}

	// 从上下文中获取租户 ID
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		tenantID = uuid.Nil
	}

	// 解析权限 ID 列表
	for _, permID := range req.PermissionIds {
		permissionID, err := uuid.Parse(permID)
		if err != nil {
			return nil, fmt.Errorf("invalid permission id: %w", err)
		}

		// 分配权限
		if err := a.permissionRepo.AssignPermissionToRole(ctx, roleID, permissionID, tenantID); err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}

// RevokePermissions 撤销角色的权限
func (a *RoleAdapter) RevokePermissions(ctx context.Context, req *authv1.RevokePermissionsRequest) (*emptypb.Empty, error) {
	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, fmt.Errorf("invalid role id: %w", err)
	}

	// 解析权限 ID 列表
	for _, permID := range req.PermissionIds {
		permissionID, err := uuid.Parse(permID)
		if err != nil {
			return nil, fmt.Errorf("invalid permission id: %w", err)
		}

		// 撤销权限
		if err := a.permissionRepo.RemovePermissionFromRole(ctx, roleID, permissionID); err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}

// AssignRoleToUser 分配角色给用户
func (a *RoleAdapter) AssignRoleToUser(ctx context.Context, req *authv1.AssignRoleToUserRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, fmt.Errorf("invalid role id: %w", err)
	}

	// 从上下文中获取租户 ID
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		tenantID = uuid.Nil
	}

	if err := a.roleRepo.AssignRoleToUser(ctx, userID, roleID, tenantID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RemoveRoleFromUser 移除用户的角色
func (a *RoleAdapter) RemoveRoleFromUser(ctx context.Context, req *authv1.RemoveRoleFromUserRequest) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, fmt.Errorf("invalid role id: %w", err)
	}

	if err := a.roleRepo.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
