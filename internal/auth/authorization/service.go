package authorization

import (
	"context"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/authorization/abac"
	"github.com/lk2023060901/go-next-erp/internal/auth/authorization/rbac"
	"github.com/lk2023060901/go-next-erp/internal/auth/authorization/rebac"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/internal/auth/repository"
	"github.com/lk2023060901/go-next-erp/pkg/cache"
)

// Service 统一授权服务（整合 RBAC、ABAC、ReBAC）
type Service struct {
	rbacEngine  *rbac.Engine
	abacEngine  *abac.Engine
	rebacEngine *rebac.Engine
	auditRepo   repository.AuditLogRepository
}

// NewService 创建授权服务
func NewService(
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	policyRepo repository.PolicyRepository,
	userRepo repository.UserRepository,
	relationRepo repository.RelationRepository,
	auditRepo repository.AuditLogRepository,
	cache *cache.Cache,
) *Service {
	return &Service{
		rbacEngine:  rbac.NewEngine(roleRepo, permissionRepo, cache),
		abacEngine:  abac.NewEngine(policyRepo, userRepo),
		rebacEngine: rebac.NewEngine(relationRepo),
		auditRepo:   auditRepo,
	}
}

// CheckPermission 检查权限（综合 RBAC + ABAC + ReBAC）
func (s *Service) CheckPermission(
	ctx context.Context,
	userID, tenantID uuid.UUID,
	resource, action string,
	resourceAttrs map[string]interface{},
) (bool, error) {
	// 1. RBAC 检查
	rbacAllowed, err := s.rbacEngine.CheckPermission(ctx, userID, resource, action)
	if err == nil && rbacAllowed {
		return true, nil
	}

	// 2. ABAC 检查
	abacAllowed, err := s.abacEngine.CheckPermission(
		ctx, userID, tenantID, resource, action, resourceAttrs, nil,
	)
	if err == nil && abacAllowed {
		return true, nil
	}

	// 3. ReBAC 检查（如果提供了 resource_id）
	if resourceID, ok := resourceAttrs["ID"].(string); ok {
		subject := "user:" + userID.String()
		object := resource + ":" + resourceID
		rebacAllowed, err := s.rebacEngine.Check(ctx, tenantID, subject, action, object)
		if err == nil && rebacAllowed {
			return true, nil
		}
	}

	// 4. 默认拒绝
	return false, nil
}

// CheckPermissionWithAudit 检查权限并记录审计
func (s *Service) CheckPermissionWithAudit(
	ctx context.Context,
	userID, tenantID uuid.UUID,
	resource, action string,
	resourceAttrs map[string]interface{},
	ipAddress, userAgent string,
) (bool, error) {
	// 检查权限
	allowed, err := s.CheckPermission(ctx, userID, tenantID, resource, action, resourceAttrs)

	// 记录审计日志
	result := model.AuditResultSuccess
	if !allowed {
		result = model.AuditResultDenied
	}

	resourceID := ""
	if id, ok := resourceAttrs["ID"].(string); ok {
		resourceID = id
	}

	_ = s.auditRepo.Create(ctx, &model.AuditLog{
		EventID:    uuid.New().String(),
		TenantID:   tenantID,
		UserID:     userID,
		Action:     resource + ":" + action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Result:     result,
	})

	return allowed, err
}

// GetUserPermissions 获取用户的所有权限（RBAC）
func (s *Service) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error) {
	return s.rbacEngine.GetUserPermissions(ctx, userID)
}

// GetUserRoles 获取用户的所有角色（含继承）
func (s *Service) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*model.Role, error) {
	return s.rbacEngine.GetUserRoles(ctx, userID)
}

// GrantRelation 授予关系（ReBAC）
func (s *Service) GrantRelation(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) error {
	return s.rebacEngine.Grant(ctx, tenantID, subject, relation, object)
}

// RevokeRelation 撤销关系（ReBAC）
func (s *Service) RevokeRelation(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) error {
	return s.rebacEngine.Revoke(ctx, tenantID, subject, relation, object)
}

// ValidatePolicyExpression 验证 ABAC 策略表达式
func (s *Service) ValidatePolicyExpression(expression string) error {
	return s.abacEngine.ValidatePolicyExpression(expression)
}
