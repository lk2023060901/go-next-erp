package abac

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/internal/auth/repository"
)

// Engine ABAC 授权引擎
type Engine struct {
	policyRepo repository.PolicyRepository
	userRepo   repository.UserRepository
	evaluator  *Evaluator
}

// NewEngine 创建 ABAC 引擎
func NewEngine(
	policyRepo repository.PolicyRepository,
	userRepo repository.UserRepository,
) *Engine {
	return &Engine{
		policyRepo: policyRepo,
		userRepo:   userRepo,
		evaluator:  NewEvaluator(),
	}
}

// CheckPermission 检查用户是否有权限访问资源
func (e *Engine) CheckPermission(
	ctx context.Context,
	userID, tenantID uuid.UUID,
	resource, action string,
	resourceAttrs map[string]interface{},
	envAttrs map[string]interface{},
) (bool, error) {
	// 1. 获取用户信息
	user, err := e.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}

	// 2. 获取适用的策略（按优先级排序）
	policies, err := e.policyRepo.GetApplicablePolicies(ctx, tenantID, resource, action)
	if err != nil {
		return false, err
	}

	// 3. 构建评估上下文
	evalCtx := e.buildContext(user, resourceAttrs, envAttrs)

	// 4. 按优先级评估策略
	for _, policy := range policies {
		matched, err := e.evaluator.Evaluate(policy.Expression, evalCtx)
		if err != nil {
			// 表达式错误，跳过该策略
			continue
		}

		if matched {
			// 策略匹配，返回效果
			return policy.Effect == model.PolicyEffectAllow, nil
		}
	}

	// 5. 默认拒绝
	return false, nil
}

// EvaluatePolicy 评估单个策略
func (e *Engine) EvaluatePolicy(
	ctx context.Context,
	policy *model.Policy,
	userID uuid.UUID,
	resourceAttrs map[string]interface{},
	envAttrs map[string]interface{},
) (bool, error) {
	// 获取用户信息
	user, err := e.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}

	// 构建上下文
	evalCtx := e.buildContext(user, resourceAttrs, envAttrs)

	// 评估表达式
	return e.evaluator.Evaluate(policy.Expression, evalCtx)
}

// ValidatePolicyExpression 验证策略表达式
func (e *Engine) ValidatePolicyExpression(expression string) error {
	return e.evaluator.ValidateExpression(expression)
}

// buildContext 构建评估上下文
func (e *Engine) buildContext(
	user *model.User,
	resourceAttrs map[string]interface{},
	envAttrs map[string]interface{},
) *EvaluationContext {
	// 用户属性
	userAttrs := map[string]interface{}{
		"ID":       user.ID.String(),
		"Username": user.Username,
		"Email":    user.Email,
		"TenantID": user.TenantID.String(),
		"Status":   string(user.Status),
	}

	// 合并用户元数据
	if user.Metadata != nil {
		for k, v := range user.Metadata {
			userAttrs[k] = v
		}
	}

	// 资源属性（如果为空，初始化）
	if resourceAttrs == nil {
		resourceAttrs = make(map[string]interface{})
	}

	// 环境属性（如果为空，初始化）
	if envAttrs == nil {
		envAttrs = make(map[string]interface{})
	}

	// 时间属性
	now := time.Now()
	timeAttrs := map[string]interface{}{
		"Hour":    now.Hour(),
		"Day":     now.Day(),
		"Weekday": int(now.Weekday()),
		"Month":   int(now.Month()),
		"Year":    now.Year(),
	}

	return &EvaluationContext{
		User:        userAttrs,
		Resource:    resourceAttrs,
		Environment: envAttrs,
		Time:        timeAttrs,
	}
}

// GetApplicablePolicies 获取适用的策略（已排序）
func (e *Engine) GetApplicablePolicies(
	ctx context.Context,
	tenantID uuid.UUID,
	resource, action string,
) ([]*model.Policy, error) {
	policies, err := e.policyRepo.GetApplicablePolicies(ctx, tenantID, resource, action)
	if err != nil {
		return nil, err
	}

	// 按优先级排序（从高到低）
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Priority > policies[j].Priority
	})

	return policies, nil
}
