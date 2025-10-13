package authorization

import (
	"context"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
)

// AuthorizationService 授权服务接口
type AuthorizationService interface {
	// CheckPermission 检查权限（综合 RBAC + ABAC + ReBAC）
	CheckPermission(
		ctx context.Context,
		userID, tenantID uuid.UUID,
		resource, action string,
		resourceAttrs map[string]interface{},
	) (bool, error)

	// CheckPermissionWithAudit 检查权限并记录审计
	CheckPermissionWithAudit(
		ctx context.Context,
		userID, tenantID uuid.UUID,
		resource, action string,
		resourceAttrs map[string]interface{},
		ipAddress, userAgent string,
	) (bool, error)

	// GetUserPermissions 获取用户的所有权限（RBAC）
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error)

	// GetUserRoles 获取用户的所有角色（含继承）
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.Role, error)

	// GrantRelation 授予关系（ReBAC）
	GrantRelation(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) error

	// RevokeRelation 撤销关系（ReBAC）
	RevokeRelation(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) error

	// ValidatePolicyExpression 验证策略表达式（ABAC）
	ValidatePolicyExpression(expression string) error
}

// 确保 Service 实现了 AuthorizationService 接口
var _ AuthorizationService = (*Service)(nil)
