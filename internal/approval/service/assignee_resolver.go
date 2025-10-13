package service

import (
	"context"
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/google/uuid"
	authRepo "github.com/lk2023060901/go-next-erp/internal/auth/repository"
	orgService "github.com/lk2023060901/go-next-erp/internal/organization/service"
	"github.com/lk2023060901/go-next-erp/pkg/workflow"
)

// AssigneeType 审批人分配类型
type AssigneeType string

const (
	AssigneeTypeUser         AssigneeType = "user"          // 指定用户
	AssigneeTypeRole         AssigneeType = "role"          // 按角色
	AssigneeTypeDepartment   AssigneeType = "department"    // 按部门
	AssigneeTypeRelation     AssigneeType = "relation"      // 按关系（如：申请人上级）
	AssigneeTypeExpression   AssigneeType = "expression"    // 按表达式
)

// AssigneeConfig 审批人配置
type AssigneeConfig struct {
	Type       AssigneeType `json:"type"`
	Value      string       `json:"value"`       // 用户ID/角色ID/部门ID/关系类型
	Expression string       `json:"expression"`  // 条件表达式（可选）
}

// AssigneeResolver 审批人解析器
type AssigneeResolver struct {
	userRepo    authRepo.UserRepository
	empService  orgService.EmployeeService
	orgService  orgService.OrganizationService
}

// NewAssigneeResolver 创建审批人解析器
func NewAssigneeResolver(
	userRepo authRepo.UserRepository,
	empService orgService.EmployeeService,
	orgService orgService.OrganizationService,
) *AssigneeResolver {
	return &AssigneeResolver{
		userRepo:   userRepo,
		empService: empService,
		orgService: orgService,
	}
}

// ResolveAssignee 解析审批人
func (r *AssigneeResolver) ResolveAssignee(
	ctx context.Context,
	node *workflow.NodeDefinition,
	processVariables map[string]interface{},
) ([]uuid.UUID, error) {

	// 从节点配置中获取审批人配置
	assigneeConfig, err := r.parseAssigneeConfig(node.Config)
	if err != nil {
		return nil, err
	}

	switch assigneeConfig.Type {
	case AssigneeTypeUser:
		// 指定用户
		return r.resolveUserAssignee(assigneeConfig.Value)

	case AssigneeTypeRole:
		// 按角色
		return r.resolveRoleAssignee(ctx, assigneeConfig.Value)

	case AssigneeTypeDepartment:
		// 按部门
		return r.resolveDepartmentAssignee(ctx, assigneeConfig.Value)

	case AssigneeTypeRelation:
		// 按关系
		return r.resolveRelationAssignee(ctx, assigneeConfig.Value, processVariables)

	case AssigneeTypeExpression:
		// 按表达式
		return r.resolveExpressionAssignee(ctx, assigneeConfig.Expression, processVariables)

	default:
		return nil, fmt.Errorf("unknown assignee type: %s", assigneeConfig.Type)
	}
}

// parseAssigneeConfig 解析审批人配置
func (r *AssigneeResolver) parseAssigneeConfig(nodeConfig map[string]interface{}) (*AssigneeConfig, error) {
	// 兼容旧的 assignee_id 字段
	if assigneeIDStr, ok := nodeConfig["assignee_id"].(string); ok {
		return &AssigneeConfig{
			Type:  AssigneeTypeUser,
			Value: assigneeIDStr,
		}, nil
	}

	// 新的配置格式
	assigneeType, ok := nodeConfig["assignee_type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing assignee_type in node config")
	}

	assigneeValue, _ := nodeConfig["assignee_value"].(string)
	expression, _ := nodeConfig["assignee_expression"].(string)

	return &AssigneeConfig{
		Type:       AssigneeType(assigneeType),
		Value:      assigneeValue,
		Expression: expression,
	}, nil
}

// resolveUserAssignee 解析指定用户
func (r *AssigneeResolver) resolveUserAssignee(userIDStr string) ([]uuid.UUID, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	return []uuid.UUID{userID}, nil
}

// resolveRoleAssignee 解析角色审批人
func (r *AssigneeResolver) resolveRoleAssignee(ctx context.Context, roleIDStr string) ([]uuid.UUID, error) {
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid role id: %w", err)
	}

	// 查询该角色下的所有用户
	users, err := r.userRepo.ListUsersByRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to list users by role: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no active users found for role %s", roleID)
	}

	userIDs := make([]uuid.UUID, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	return userIDs, nil
}

// resolveDepartmentAssignee 解析部门审批人（查找部门负责人）
func (r *AssigneeResolver) resolveDepartmentAssignee(ctx context.Context, deptIDStr string) ([]uuid.UUID, error) {
	deptID, err := uuid.Parse(deptIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid department id: %w", err)
	}

	// 查询部门信息
	dept, err := r.orgService.GetByID(ctx, deptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get department: %w", err)
	}

	// 检查部门是否有负责人
	if dept.LeaderID == nil {
		return nil, fmt.Errorf("department %s has no leader assigned", dept.Name)
	}

	return []uuid.UUID{*dept.LeaderID}, nil
}

// resolveRelationAssignee 解析关系审批人
func (r *AssigneeResolver) resolveRelationAssignee(
	ctx context.Context,
	relationType string,
	processVariables map[string]interface{},
) ([]uuid.UUID, error) {

	switch relationType {
	case "applicant_manager":
		// 申请人的直属上级
		applicantIDStr, ok := processVariables["applicant_id"].(string)
		if !ok {
			return nil, fmt.Errorf("applicant_id not found in process variables")
		}

		applicantID, err := uuid.Parse(applicantIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid applicant_id: %w", err)
		}

		// 通过员工信息查询直属上级
		tenantIDStr, ok := processVariables["tenant_id"].(string)
		if !ok {
			return nil, fmt.Errorf("tenant_id not found in process variables")
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid tenant_id: %w", err)
		}

		employee, err := r.empService.GetByUserID(ctx, tenantID, applicantID)
		if err != nil {
			return nil, fmt.Errorf("failed to get employee info: %w", err)
		}

		if employee.DirectLeaderID == nil {
			return nil, fmt.Errorf("applicant has no direct manager assigned")
		}

		return []uuid.UUID{*employee.DirectLeaderID}, nil

	case "applicant_dept_manager":
		// 申请人部门负责人
		applicantIDStr, ok := processVariables["applicant_id"].(string)
		if !ok {
			return nil, fmt.Errorf("applicant_id not found in process variables")
		}

		applicantID, err := uuid.Parse(applicantIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid applicant_id: %w", err)
		}

		tenantIDStr, ok := processVariables["tenant_id"].(string)
		if !ok {
			return nil, fmt.Errorf("tenant_id not found in process variables")
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid tenant_id: %w", err)
		}

		// 获取申请人的部门信息
		employee, err := r.empService.GetByUserID(ctx, tenantID, applicantID)
		if err != nil {
			return nil, fmt.Errorf("failed to get employee info: %w", err)
		}

		// 获取部门负责人
		dept, err := r.orgService.GetByID(ctx, employee.OrgID)
		if err != nil {
			return nil, fmt.Errorf("failed to get department: %w", err)
		}

		if dept.LeaderID == nil {
			return nil, fmt.Errorf("applicant's department has no leader assigned")
		}

		return []uuid.UUID{*dept.LeaderID}, nil

	default:
		return nil, fmt.Errorf("unknown relation type: %s", relationType)
	}
}

// resolveExpressionAssignee 解析表达式审批人
func (r *AssigneeResolver) resolveExpressionAssignee(
	ctx context.Context,
	expression string,
	processVariables map[string]interface{},
) ([]uuid.UUID, error) {

	// 编译并执行表达式
	program, err := expr.Compile(expression, expr.Env(processVariables))
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", err)
	}

	result, err := expr.Run(program, processVariables)
	if err != nil {
		return nil, fmt.Errorf("failed to execute expression: %w", err)
	}

	// 解析结果为审批人ID
	switch v := result.(type) {
	case string:
		// 单个用户ID字符串
		userID, err := uuid.Parse(v)
		if err != nil {
			return nil, fmt.Errorf("expression result is not a valid UUID: %w", err)
		}
		return []uuid.UUID{userID}, nil

	case []interface{}:
		// 多个用户ID数组
		userIDs := make([]uuid.UUID, 0, len(v))
		for _, item := range v {
			if idStr, ok := item.(string); ok {
				userID, err := uuid.Parse(idStr)
				if err != nil {
					return nil, fmt.Errorf("invalid UUID in array: %w", err)
				}
				userIDs = append(userIDs, userID)
			}
		}
		if len(userIDs) == 0 {
			return nil, fmt.Errorf("expression result array contains no valid UUIDs")
		}
		return userIDs, nil

	case []string:
		// 字符串数组
		userIDs := make([]uuid.UUID, 0, len(v))
		for _, idStr := range v {
			userID, err := uuid.Parse(idStr)
			if err != nil {
				return nil, fmt.Errorf("invalid UUID in array: %w", err)
			}
			userIDs = append(userIDs, userID)
		}
		return userIDs, nil

	default:
		return nil, fmt.Errorf("expression result type %T is not supported, expected string or []string", result)
	}
}
