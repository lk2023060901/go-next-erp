package abac

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
)

// ============================================================================
// ABAC Engine Benchmark Tests
// ============================================================================

// BenchmarkEngine_CheckPermission_SimpleExpression 基准测试：简单表达式
func BenchmarkEngine_CheckPermission_SimpleExpression(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)

	user := &model.User{
		ID:       userID,
		TenantID: tenantID,
		Metadata: map[string]interface{}{
			"Level": 3,
		},
	}

	policies := []*model.Policy{
		{
			Expression: "User.Level >= 3",
			Effect:     model.PolicyEffectAllow,
			Priority:   10,
		},
	}

	userRepo.On("FindByID", ctx, userID).Return(user, nil)
	policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "read").Return(policies, nil)

	engine := NewEngine(policyRepo, userRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.CheckPermission(ctx, userID, tenantID, "document", "read", nil, nil)
	}
}

// BenchmarkEngine_CheckPermission_ComplexExpression 基准测试：复杂表达式
func BenchmarkEngine_CheckPermission_ComplexExpression(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)

	user := &model.User{
		ID:    userID,
		Email: "alice@example.com",
		Metadata: map[string]interface{}{
			"Department": "IT",
			"Level":      5,
			"Verified":   true,
		},
	}

	resourceAttrs := map[string]interface{}{
		"Department": "IT",
	}

	policies := []*model.Policy{
		{
			Expression: "(User.Department == Resource.Department) && (User.Level >= 3) && (User.Verified == true) && (User.Email contains \"@example.com\")",
			Effect:     model.PolicyEffectAllow,
			Priority:   10,
		},
	}

	userRepo.On("FindByID", ctx, userID).Return(user, nil)
	policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "read").Return(policies, nil)

	engine := NewEngine(policyRepo, userRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.CheckPermission(ctx, userID, tenantID, "document", "read", resourceAttrs, nil)
	}
}

// BenchmarkEngine_CheckPermission_MultiplePolicies 基准测试：多策略评估
func BenchmarkEngine_CheckPermission_MultiplePolicies(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)

	user := &model.User{
		ID: userID,
		Metadata: map[string]interface{}{
			"Level": 3,
			"Role":  "developer",
		},
	}

	policies := []*model.Policy{
		{
			Expression: "User.Role == \"admin\"",
			Effect:     model.PolicyEffectAllow,
			Priority:   100,
		},
		{
			Expression: "User.Role == \"manager\"",
			Effect:     model.PolicyEffectAllow,
			Priority:   50,
		},
		{
			Expression: "User.Level >= 3",
			Effect:     model.PolicyEffectAllow,
			Priority:   10,
		},
	}

	userRepo.On("FindByID", ctx, userID).Return(user, nil)
	policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "read").Return(policies, nil)

	engine := NewEngine(policyRepo, userRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.CheckPermission(ctx, userID, tenantID, "document", "read", nil, nil)
	}
}

// BenchmarkEngine_EvaluatePolicy 基准测试：单策略评估
func BenchmarkEngine_EvaluatePolicy(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()

	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)

	user := &model.User{
		ID: userID,
		Metadata: map[string]interface{}{
			"Level": 5,
		},
	}

	policy := &model.Policy{
		Expression: "User.Level >= 3",
	}

	userRepo.On("FindByID", ctx, userID).Return(user, nil)

	engine := NewEngine(policyRepo, userRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.EvaluatePolicy(ctx, policy, userID, nil, nil)
	}
}

// BenchmarkEngine_ValidatePolicyExpression 基准测试：表达式验证
func BenchmarkEngine_ValidatePolicyExpression(b *testing.B) {
	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)

	engine := NewEngine(policyRepo, userRepo)
	expression := "User.Level >= 3 && User.Department == \"IT\""

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.ValidatePolicyExpression(expression)
	}
}

// BenchmarkEngine_GetApplicablePolicies 基准测试：获取适用策略
func BenchmarkEngine_GetApplicablePolicies(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)

	policies := []*model.Policy{
		{ID: uuid.New(), Priority: 10},
		{ID: uuid.New(), Priority: 100},
		{ID: uuid.New(), Priority: 5},
		{ID: uuid.New(), Priority: 50},
		{ID: uuid.New(), Priority: 1},
	}

	policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "read").Return(policies, nil)

	engine := NewEngine(policyRepo, userRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.GetApplicablePolicies(ctx, tenantID, "document", "read")
	}
}

// BenchmarkEngine_BuildContext 基准测试：构建评估上下文
func BenchmarkEngine_BuildContext(b *testing.B) {
	user := &model.User{
		ID:       uuid.New(),
		Username: "alice",
		Email:    "alice@example.com",
		TenantID: uuid.New(),
		Status:   model.UserStatusActive,
		Metadata: map[string]interface{}{
			"Department": "IT",
			"Level":      5,
			"Verified":   true,
		},
	}

	resourceAttrs := map[string]interface{}{
		"Type":   "document",
		"Status": "published",
	}

	envAttrs := map[string]interface{}{
		"IPAddress": "192.168.1.1",
	}

	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)
	engine := NewEngine(policyRepo, userRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.buildContext(user, resourceAttrs, envAttrs)
	}
}

// ============================================================================
// Evaluator Benchmark Tests
// ============================================================================

// BenchmarkEvaluator_Evaluate_Simple 基准测试：简单表达式评估
func BenchmarkEvaluator_Evaluate_Simple(b *testing.B) {
	evaluator := NewEvaluator()
	expression := "User.Level >= 3"
	ctx := &EvaluationContext{
		User: map[string]interface{}{
			"Level": 5,
		},
	}

	// 预热缓存
	_, _ = evaluator.Evaluate(expression, ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.Evaluate(expression, ctx)
	}
}

// BenchmarkEvaluator_Evaluate_Complex 基准测试：复杂表达式评估
func BenchmarkEvaluator_Evaluate_Complex(b *testing.B) {
	evaluator := NewEvaluator()
	expression := "(User.Level >= 3 && User.Department == \"IT\") || (User.Role == \"admin\" && User.Active == true)"
	ctx := &EvaluationContext{
		User: map[string]interface{}{
			"Level":      5,
			"Department": "IT",
			"Role":       "developer",
			"Active":     true,
		},
	}

	// 预热缓存
	_, _ = evaluator.Evaluate(expression, ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.Evaluate(expression, ctx)
	}
}

// BenchmarkEvaluator_Evaluate_NoCache 基准测试：无缓存评估
func BenchmarkEvaluator_Evaluate_NoCache(b *testing.B) {
	ctx := &EvaluationContext{
		User: map[string]interface{}{
			"Level": 5,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator := NewEvaluator() // 每次新建，无缓存
		_, _ = evaluator.Evaluate("User.Level >= 3", ctx)
	}
}

// BenchmarkEvaluator_Evaluate_String 基准测试：字符串操作
func BenchmarkEvaluator_Evaluate_String(b *testing.B) {
	evaluator := NewEvaluator()
	expression := "User.Email contains \"@example.com\" && User.Username == \"alice\""
	ctx := &EvaluationContext{
		User: map[string]interface{}{
			"Email":    "alice@example.com",
			"Username": "alice",
		},
	}

	// 预热缓存
	_, _ = evaluator.Evaluate(expression, ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.Evaluate(expression, ctx)
	}
}

// BenchmarkEvaluator_Evaluate_Numeric 基准测试：数值运算
func BenchmarkEvaluator_Evaluate_Numeric(b *testing.B) {
	evaluator := NewEvaluator()
	expression := "User.Age >= 18 && User.Score > 60 && User.Level <= 10"
	ctx := &EvaluationContext{
		User: map[string]interface{}{
			"Age":   25,
			"Score": 85,
			"Level": 5,
		},
	}

	// 预热缓存
	_, _ = evaluator.Evaluate(expression, ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.Evaluate(expression, ctx)
	}
}

// BenchmarkEvaluator_Evaluate_Boolean 基准测试：布尔运算
func BenchmarkEvaluator_Evaluate_Boolean(b *testing.B) {
	evaluator := NewEvaluator()
	expression := "User.Active == true && User.Verified == true && !User.Blocked"
	ctx := &EvaluationContext{
		User: map[string]interface{}{
			"Active":   true,
			"Verified": true,
			"Blocked":  false,
		},
	}

	// 预热缓存
	_, _ = evaluator.Evaluate(expression, ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.Evaluate(expression, ctx)
	}
}

// BenchmarkEvaluator_Evaluate_CrossContext 基准测试：跨上下文访问
func BenchmarkEvaluator_Evaluate_CrossContext(b *testing.B) {
	evaluator := NewEvaluator()
	expression := "User.DepartmentID == Resource.DepartmentID && Time.Hour >= 9 && Environment.Secure == true"
	ctx := &EvaluationContext{
		User: map[string]interface{}{
			"DepartmentID": "IT-001",
		},
		Resource: map[string]interface{}{
			"DepartmentID": "IT-001",
		},
		Time: map[string]interface{}{
			"Hour": 14,
		},
		Environment: map[string]interface{}{
			"Secure": true,
		},
	}

	// 预热缓存
	_, _ = evaluator.Evaluate(expression, ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.Evaluate(expression, ctx)
	}
}

// BenchmarkEvaluator_ValidateExpression 基准测试：表达式验证
func BenchmarkEvaluator_ValidateExpression(b *testing.B) {
	evaluator := NewEvaluator()
	expression := "User.Level >= 3 && User.Department == \"IT\""

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = evaluator.ValidateExpression(expression)
	}
}

// BenchmarkEvaluator_ClearCache 基准测试：清除缓存
func BenchmarkEvaluator_ClearCache(b *testing.B) {
	evaluator := NewEvaluator()
	ctx := &EvaluationContext{
		User: map[string]interface{}{"Level": 5},
	}

	// 填充缓存
	for i := 0; i < 100; i++ {
		expr := "User.Level >= " + string(rune('0'+i%10))
		_, _ = evaluator.Evaluate(expr, ctx)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.ClearCache()
	}
}

// ============================================================================
// 并发基准测试
// ============================================================================

// BenchmarkConcurrentEngine_CheckPermission 基准测试：并发权限检查
func BenchmarkConcurrentEngine_CheckPermission(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	policyRepo := new(MockPolicyRepository)
	userRepo := new(MockUserRepository)

	user := &model.User{
		ID: userID,
		Metadata: map[string]interface{}{
			"Level": 5,
		},
	}

	policies := []*model.Policy{
		{
			Expression: "User.Level >= 3",
			Effect:     model.PolicyEffectAllow,
			Priority:   10,
		},
	}

	userRepo.On("FindByID", ctx, userID).Return(user, nil)
	policyRepo.On("GetApplicablePolicies", ctx, tenantID, "document", "read").Return(policies, nil)

	engine := NewEngine(policyRepo, userRepo)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = engine.CheckPermission(ctx, userID, tenantID, "document", "read", nil, nil)
		}
	})
}

// BenchmarkConcurrentEvaluator_Evaluate 基准测试：并发表达式评估
func BenchmarkConcurrentEvaluator_Evaluate(b *testing.B) {
	evaluator := NewEvaluator()
	expression := "User.Level >= 3 && User.Department == \"IT\""
	ctx := &EvaluationContext{
		User: map[string]interface{}{
			"Level":      5,
			"Department": "IT",
		},
	}

	// 预热缓存
	_, _ = evaluator.Evaluate(expression, ctx)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = evaluator.Evaluate(expression, ctx)
		}
	})
}
