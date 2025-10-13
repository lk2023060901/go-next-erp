package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	workflowModel "github.com/lk2023060901/go-next-erp/pkg/workflow"
	"github.com/stretchr/testify/assert"
)

// ===== Expression Assignee Tests =====

// TestExpressionAssignee 测试表达式审批人解析
func TestExpressionAssignee(t *testing.T) {
	resolver := &AssigneeResolver{}
	ctx := context.Background()

	t.Run("Simple condition - single assignee", func(t *testing.T) {
		cfoID := uuid.New()
		managerID := uuid.New()

		// 表达式：金额 > 10000 则 CFO 审批，否则经理审批
		expression := `amount > 10000 ? cfo_id : manager_id`

		processVariables := map[string]interface{}{
			"amount":     15000,
			"cfo_id":     cfoID.String(),
			"manager_id": managerID.String(),
		}

		node := &workflowModel.NodeDefinition{
			ID:   "approval_node",
			Name: "条件审批",
			Config: map[string]interface{}{
				"assignee_type":       "expression",
				"assignee_expression": expression,
			},
		}

		assignees, err := resolver.ResolveAssignee(ctx, node, processVariables)
		assert.NoError(t, err)
		assert.Len(t, assignees, 1)
		assert.Equal(t, cfoID, assignees[0])
	})

	t.Run("Simple condition - low amount", func(t *testing.T) {
		cfoID := uuid.New()
		managerID := uuid.New()

		expression := `amount > 10000 ? cfo_id : manager_id`

		processVariables := map[string]interface{}{
			"amount":     5000, // 小于 10000
			"cfo_id":     cfoID.String(),
			"manager_id": managerID.String(),
		}

		node := &workflowModel.NodeDefinition{
			ID:   "approval_node",
			Name: "条件审批",
			Config: map[string]interface{}{
				"assignee_type":       "expression",
				"assignee_expression": expression,
			},
		}

		assignees, err := resolver.ResolveAssignee(ctx, node, processVariables)
		assert.NoError(t, err)
		assert.Len(t, assignees, 1)
		assert.Equal(t, managerID, assignees[0])
	})

	t.Run("Complex condition - multiple criteria", func(t *testing.T) {
		ceoID := uuid.New()
		cfoID := uuid.New()
		managerID := uuid.New()

		// 复杂表达式：部门=财务 且 金额>50000 则 CEO，金额>10000 则 CFO，否则经理
		expression := `department == "finance" && amount > 50000 ? ceo_id : (amount > 10000 ? cfo_id : manager_id)`

		processVariables := map[string]interface{}{
			"department": "finance",
			"amount":     60000,
			"ceo_id":     ceoID.String(),
			"cfo_id":     cfoID.String(),
			"manager_id": managerID.String(),
		}

		node := &workflowModel.NodeDefinition{
			ID:   "approval_node",
			Name: "分级审批",
			Config: map[string]interface{}{
				"assignee_type":       "expression",
				"assignee_expression": expression,
			},
		}

		assignees, err := resolver.ResolveAssignee(ctx, node, processVariables)
		assert.NoError(t, err)
		assert.Len(t, assignees, 1)
		assert.Equal(t, ceoID, assignees[0])
	})

	t.Run("Array result - multiple assignees", func(t *testing.T) {
		approver1 := uuid.New()
		approver2 := uuid.New()

		// 表达式返回数组（需要多人会签）
		expression := `is_critical ? [approver1_id, approver2_id] : [approver1_id]`

		processVariables := map[string]interface{}{
			"is_critical":  true,
			"approver1_id": approver1.String(),
			"approver2_id": approver2.String(),
		}

		node := &workflowModel.NodeDefinition{
			ID:   "approval_node",
			Name: "会签审批",
			Config: map[string]interface{}{
				"assignee_type":       "expression",
				"assignee_expression": expression,
			},
		}

		assignees, err := resolver.ResolveAssignee(ctx, node, processVariables)
		assert.NoError(t, err)
		assert.Len(t, assignees, 2)
		assert.Contains(t, assignees, approver1)
		assert.Contains(t, assignees, approver2)
	})

	t.Run("Invalid expression", func(t *testing.T) {
		expression := `amount > 10000 ? invalid syntax`

		processVariables := map[string]interface{}{
			"amount": 15000,
		}

		node := &workflowModel.NodeDefinition{
			ID:   "approval_node",
			Name: "错误表达式",
			Config: map[string]interface{}{
				"assignee_type":       "expression",
				"assignee_expression": expression,
			},
		}

		_, err := resolver.ResolveAssignee(ctx, node, processVariables)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to compile expression")
	})

	t.Run("Invalid UUID result", func(t *testing.T) {
		expression := `"not-a-uuid"`

		processVariables := map[string]interface{}{}

		node := &workflowModel.NodeDefinition{
			ID:   "approval_node",
			Name: "无效UUID",
			Config: map[string]interface{}{
				"assignee_type":       "expression",
				"assignee_expression": expression,
			},
		}

		_, err := resolver.ResolveAssignee(ctx, node, processVariables)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a valid UUID")
	})
}

// ===== User Assignee Tests =====

// TestUserAssignee 测试指定用户审批人
func TestUserAssignee(t *testing.T) {
	resolver := &AssigneeResolver{}
	ctx := context.Background()

	t.Run("Single user", func(t *testing.T) {
		userID := uuid.New()

		node := &workflowModel.NodeDefinition{
			ID:   "approval_node",
			Name: "指定用户审批",
			Config: map[string]interface{}{
				"assignee_type": "user",
				"assignee_id":   userID.String(),
			},
		}

		assignees, err := resolver.ResolveAssignee(ctx, node, nil)
		assert.NoError(t, err)
		assert.Len(t, assignees, 1)
		assert.Equal(t, userID, assignees[0])
	})

	t.Run("Invalid user ID", func(t *testing.T) {
		node := &workflowModel.NodeDefinition{
			ID:   "approval_node",
			Name: "无效用户ID",
			Config: map[string]interface{}{
				"assignee_type": "user",
				"assignee_id":   "invalid-uuid",
			},
		}

		_, err := resolver.ResolveAssignee(ctx, node, nil)
		assert.Error(t, err)
	})

	t.Run("Missing assignee_id", func(t *testing.T) {
		node := &workflowModel.NodeDefinition{
			ID:   "approval_node",
			Name: "缺少用户ID",
			Config: map[string]interface{}{
				"assignee_type": "user",
			},
		}

		_, err := resolver.ResolveAssignee(ctx, node, nil)
		assert.Error(t, err)
		// 实际错误信息是 "invalid user id: invalid UUID length: 0"
		assert.Contains(t, err.Error(), "invalid user id")
	})
}

// ===== Benchmark Tests =====

// BenchmarkExpressionResolution 基准测试：简单表达式解析性能
func BenchmarkExpressionResolution(b *testing.B) {
	resolver := &AssigneeResolver{}
	ctx := context.Background()

	cfoID := uuid.New()
	managerID := uuid.New()

	node := &workflowModel.NodeDefinition{
		ID:   "approval_node",
		Name: "条件审批",
		Config: map[string]interface{}{
			"assignee_type":       "expression",
			"assignee_expression": `amount > 10000 ? cfo_id : manager_id`,
		},
	}

	processVariables := map[string]interface{}{
		"amount":     15000,
		"cfo_id":     cfoID.String(),
		"manager_id": managerID.String(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.ResolveAssignee(ctx, node, processVariables)
	}
}

// BenchmarkComplexExpression 基准测试：复杂表达式性能
func BenchmarkComplexExpression(b *testing.B) {
	resolver := &AssigneeResolver{}
	ctx := context.Background()

	ceoID := uuid.New()
	cfoID := uuid.New()
	managerID := uuid.New()

	node := &workflowModel.NodeDefinition{
		ID:   "approval_node",
		Name: "分级审批",
		Config: map[string]interface{}{
			"assignee_type":       "expression",
			"assignee_expression": `department == "finance" && amount > 50000 ? ceo_id : (amount > 10000 ? cfo_id : manager_id)`,
		},
	}

	processVariables := map[string]interface{}{
		"department": "finance",
		"amount":     60000,
		"ceo_id":     ceoID.String(),
		"cfo_id":     cfoID.String(),
		"manager_id": managerID.String(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.ResolveAssignee(ctx, node, processVariables)
	}
}

// BenchmarkMultipleAssignees 基准测试：多审批人数组返回
func BenchmarkMultipleAssignees(b *testing.B) {
	resolver := &AssigneeResolver{}
	ctx := context.Background()

	approver1 := uuid.New()
	approver2 := uuid.New()
	approver3 := uuid.New()

	node := &workflowModel.NodeDefinition{
		ID:   "approval_node",
		Name: "会签审批",
		Config: map[string]interface{}{
			"assignee_type":       "expression",
			"assignee_expression": `is_critical ? [approver1_id, approver2_id, approver3_id] : [approver1_id]`,
		},
	}

	processVariables := map[string]interface{}{
		"is_critical":  true,
		"approver1_id": approver1.String(),
		"approver2_id": approver2.String(),
		"approver3_id": approver3.String(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.ResolveAssignee(ctx, node, processVariables)
	}
}

// BenchmarkUserAssignee 基准测试：指定用户解析
func BenchmarkUserAssignee(b *testing.B) {
	resolver := &AssigneeResolver{}
	ctx := context.Background()

	userID := uuid.New()

	node := &workflowModel.NodeDefinition{
		ID:   "approval_node",
		Name: "指定用户审批",
		Config: map[string]interface{}{
			"assignee_type": "user",
			"assignee_id":   userID.String(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.ResolveAssignee(ctx, node, nil)
	}
}

// ===== Expression Examples Documentation =====

// TestExpressionExamples 表达式示例文档
func TestExpressionExamples(t *testing.T) {
	t.Skip("Documentation examples only")

	// 示例1: 金额分级审批
	_ = `amount > 100000 ? ceo_id : (amount > 10000 ? cfo_id : manager_id)`

	// 示例2: 部门路由
	_ = `department == "finance" ? finance_manager_id : dept_manager_id`

	// 示例3: 请假天数分级
	_ = `days > 10 ? ceo_id : (days > 3 ? hr_manager_id : direct_manager_id)`

	// 示例4: 紧急情况多人会签
	_ = `is_urgent ? [ceo_id, coo_id, cfo_id] : manager_id`

	// 示例5: 组合条件
	_ = `(department == "finance" && amount > 50000) || (department == "it" && amount > 30000) ? ceo_id : manager_id`

	// 示例6: 使用逻辑运算
	_ = `amount > 10000 && risk_level == "high" ? [cfo_id, risk_manager_id] : manager_id`

	// 示例7: 复杂嵌套条件
	_ = `
		status == "urgent" ?
			(amount > 100000 ? [ceo_id, cfo_id] : cfo_id) :
			(amount > 50000 ? cfo_id : manager_id)
	`

	// 示例8: 字符串匹配
	_ = `priority == "critical" && department in ["finance", "legal"] ? [ceo_id, general_counsel_id] : manager_id`
}
