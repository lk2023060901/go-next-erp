package abac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Evaluator Tests
// ============================================================================

// TestEvaluator_Evaluate 测试表达式评估
func TestEvaluator_Evaluate(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		context     *EvaluationContext
		expectTrue  bool
		expectError bool
	}{
		{
			name:       "简单相等判断-true",
			expression: "User.Level == 3",
			context: &EvaluationContext{
				User: map[string]interface{}{
					"Level": 3,
				},
			},
			expectTrue:  true,
			expectError: false,
		},
		{
			name:       "简单相等判断-false",
			expression: "User.Level == 5",
			context: &EvaluationContext{
				User: map[string]interface{}{
					"Level": 3,
				},
			},
			expectTrue:  false,
			expectError: false,
		},
		{
			name:       "数值比较-大于等于",
			expression: "User.Age >= 18",
			context: &EvaluationContext{
				User: map[string]interface{}{
					"Age": 25,
				},
			},
			expectTrue:  true,
			expectError: false,
		},
		{
			name:       "数值比较-小于",
			expression: "User.Score < 60",
			context: &EvaluationContext{
				User: map[string]interface{}{
					"Score": 45,
				},
			},
			expectTrue:  true,
			expectError: false,
		},
		{
			name:       "字符串相等",
			expression: "User.Role == \"admin\"",
			context: &EvaluationContext{
				User: map[string]interface{}{
					"Role": "admin",
				},
			},
			expectTrue:  true,
			expectError: false,
		},
		{
			name:       "字符串包含",
			expression: "User.Email contains \"@example.com\"",
			context: &EvaluationContext{
				User: map[string]interface{}{
					"Email": "alice@example.com",
				},
			},
			expectTrue:  true,
			expectError: false,
		},
		{
			name:       "逻辑与-true",
			expression: "User.Active == true && User.Verified == true",
			context: &EvaluationContext{
				User: map[string]interface{}{
					"Active":   true,
					"Verified": true,
				},
			},
			expectTrue:  true,
			expectError: false,
		},
		{
			name:       "逻辑与-false",
			expression: "User.Active == true && User.Verified == true",
			context: &EvaluationContext{
				User: map[string]interface{}{
					"Active":   true,
					"Verified": false,
				},
			},
			expectTrue:  false,
			expectError: false,
		},
		{
			name:       "逻辑或-true",
			expression: "User.IsAdmin == true || User.IsSuperUser == true",
			context: &EvaluationContext{
				User: map[string]interface{}{
					"IsAdmin":     false,
					"IsSuperUser": true,
				},
			},
			expectTrue:  true,
			expectError: false,
		},
		{
			name:       "逻辑非",
			expression: "!User.IsGuest",
			context: &EvaluationContext{
				User: map[string]interface{}{
					"IsGuest": false,
				},
			},
			expectTrue:  true,
			expectError: false,
		},
		{
			name:       "复杂逻辑表达式",
			expression: "(User.Level >= 3 && User.Department == \"IT\") || User.IsAdmin == true",
			context: &EvaluationContext{
				User: map[string]interface{}{
					"Level":      5,
					"Department": "IT",
					"IsAdmin":    false,
				},
			},
			expectTrue:  true,
			expectError: false,
		},
		{
			name:       "资源属性访问",
			expression: "Resource.Status == \"published\"",
			context: &EvaluationContext{
				Resource: map[string]interface{}{
					"Status": "published",
				},
			},
			expectTrue:  true,
			expectError: false,
		},
		{
			name:       "环境属性访问",
			expression: "Environment.IPAddress == \"192.168.1.1\"",
			context: &EvaluationContext{
				Environment: map[string]interface{}{
					"IPAddress": "192.168.1.1",
				},
			},
			expectTrue:  true,
			expectError: false,
		},
		{
			name:       "时间属性访问",
			expression: "Time.Hour >= 9 && Time.Hour <= 18",
			context: &EvaluationContext{
				Time: map[string]interface{}{
					"Hour": 14,
				},
			},
			expectTrue:  true,
			expectError: false,
		},
		{
			name:       "跨上下文访问",
			expression: "User.DepartmentID == Resource.DepartmentID",
			context: &EvaluationContext{
				User: map[string]interface{}{
					"DepartmentID": "IT-001",
				},
				Resource: map[string]interface{}{
					"DepartmentID": "IT-001",
				},
			},
			expectTrue:  true,
			expectError: false,
		},
		{
			name:       "无效表达式-语法错误",
			expression: "User.Level >=",
			context: &EvaluationContext{
				User: map[string]interface{}{
					"Level": 3,
				},
			},
			expectTrue:  false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewEvaluator()
			result, err := evaluator.Evaluate(tt.expression, tt.context)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectTrue, result)
			}
		})
	}
}

// TestEvaluator_ValidateExpression 测试表达式验证
func TestEvaluator_ValidateExpression(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectError bool
	}{
		{
			name:        "有效-简单比较",
			expression:  "User.Level >= 3",
			expectError: false,
		},
		{
			name:        "有效-逻辑运算",
			expression:  "User.Active == true && User.Level >= 3",
			expectError: false,
		},
		{
			name:        "有效-字符串操作",
			expression:  "User.Email contains \"@example.com\"",
			expectError: false,
		},
		{
			name:        "有效-复杂嵌套",
			expression:  "(User.Level >= 3 && User.Department == \"IT\") || (User.Role == \"admin\" && Time.Hour >= 9)",
			expectError: false,
		},
		{
			name:        "无效-语法错误",
			expression:  "User.Level >= ",
			expectError: true,
		},
		{
			name:        "无效-未闭合括号",
			expression:  "(User.Level >= 3",
			expectError: true,
		},
		{
			name:        "无效-语法错误2",
			expression:  "User.Level >= && 3",
			expectError: true,
		},
		{
			name:        "空表达式",
			expression:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewEvaluator()
			err := evaluator.ValidateExpression(tt.expression)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEvaluator_ProgramCache 测试程序缓存
func TestEvaluator_ProgramCache(t *testing.T) {
	evaluator := NewEvaluator()
	expression := "User.Level >= 3"
	ctx := &EvaluationContext{
		User: map[string]interface{}{
			"Level": 5,
		},
	}

	// 第一次评估（未缓存）
	result1, err1 := evaluator.Evaluate(expression, ctx)
	assert.NoError(t, err1)
	assert.True(t, result1)

	// 第二次评估（应从缓存读取）
	result2, err2 := evaluator.Evaluate(expression, ctx)
	assert.NoError(t, err2)
	assert.True(t, result2)

	// 验证缓存中有该表达式
	_, ok := evaluator.programCache.Load(expression)
	assert.True(t, ok, "表达式应该被缓存")
}

// TestEvaluator_ClearCache 测试清除缓存
func TestEvaluator_ClearCache(t *testing.T) {
	evaluator := NewEvaluator()
	expression := "User.Level >= 3"
	ctx := &EvaluationContext{
		User: map[string]interface{}{
			"Level": 5,
		},
	}

	// 评估以填充缓存
	_, _ = evaluator.Evaluate(expression, ctx)

	// 验证缓存存在
	_, ok := evaluator.programCache.Load(expression)
	assert.True(t, ok)

	// 清除缓存
	evaluator.ClearCache()

	// 验证缓存已清除
	_, ok = evaluator.programCache.Load(expression)
	assert.False(t, ok)
}

// TestEvaluator_EdgeCases 测试边界情况
func TestEvaluator_EdgeCases(t *testing.T) {
	t.Run("nil上下文值", func(t *testing.T) {
		evaluator := NewEvaluator()
		ctx := &EvaluationContext{
			User: map[string]interface{}{
				"Name": nil,
			},
		}

		result, err := evaluator.Evaluate("User.Name == nil", ctx)
		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("空map", func(t *testing.T) {
		evaluator := NewEvaluator()
		ctx := &EvaluationContext{
			User:        map[string]interface{}{},
			Resource:    map[string]interface{}{},
			Environment: map[string]interface{}{},
			Time:        map[string]interface{}{},
		}

		// 访问不存在的字段会返回nil
		_, err := evaluator.Evaluate("User.NonExistent == nil", ctx)
		assert.NoError(t, err)
	})

	t.Run("类型不匹配", func(t *testing.T) {
		evaluator := NewEvaluator()
		ctx := &EvaluationContext{
			User: map[string]interface{}{
				"Level": "high", // 字符串而非数字
			},
		}

		_, err := evaluator.Evaluate("User.Level >= 3", ctx)
		assert.Error(t, err) // 类型不匹配应该报错
	})

	t.Run("布尔值直接返回", func(t *testing.T) {
		evaluator := NewEvaluator()
		ctx := &EvaluationContext{
			User: map[string]interface{}{
				"IsActive": true,
			},
		}

		result, err := evaluator.Evaluate("User.IsActive", ctx)
		assert.NoError(t, err)
		assert.True(t, result)
	})
}

// TestEvaluator_ComplexExpressions 测试复杂表达式
func TestEvaluator_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		context    *EvaluationContext
		expectTrue bool
	}{
		{
			name: "嵌套逻辑-场景1",
			expression: `
				(User.Department == "IT" && User.Level >= 3) ||
				(User.Department == "HR" && User.Level >= 5) ||
				User.Role == "admin"
			`,
			context: &EvaluationContext{
				User: map[string]interface{}{
					"Department": "IT",
					"Level":      4,
					"Role":       "developer",
				},
			},
			expectTrue: true,
		},
		{
			name: "嵌套逻辑-场景2",
			expression: `
				Resource.Visibility == "public" ||
				(Resource.Visibility == "private" && Resource.OwnerID == User.ID) ||
				(Resource.Visibility == "shared" && Resource.SharedWith contains User.ID)
			`,
			context: &EvaluationContext{
				User: map[string]interface{}{
					"ID": "user-123",
				},
				Resource: map[string]interface{}{
					"Visibility": "private",
					"OwnerID":    "user-123",
				},
			},
			expectTrue: true,
		},
		{
			name: "时间窗口和用户条件组合",
			expression: `
				(Time.Hour >= 9 && Time.Hour <= 18 && Time.Weekday >= 1 && Time.Weekday <= 5) &&
				(User.IsEmployee == true) &&
				(Resource.Type == "document")
			`,
			context: &EvaluationContext{
				User: map[string]interface{}{
					"IsEmployee": true,
				},
				Resource: map[string]interface{}{
					"Type": "document",
				},
				Time: map[string]interface{}{
					"Hour":    14,
					"Weekday": 3, // 周三
				},
			},
			expectTrue: true,
		},
		{
			name: "字符串匹配和数值比较组合",
			expression: `
				User.Email contains "@example.com" &&
				User.Age >= 18 &&
				User.Status == "active" &&
				!User.IsBlocked
			`,
			context: &EvaluationContext{
				User: map[string]interface{}{
					"Email":     "alice@example.com",
					"Age":       25,
					"Status":    "active",
					"IsBlocked": false,
				},
			},
			expectTrue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewEvaluator()
			result, err := evaluator.Evaluate(tt.expression, tt.context)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectTrue, result)
		})
	}
}

// TestEvaluator_PerformanceCache 测试缓存性能
func TestEvaluator_PerformanceCache(t *testing.T) {
	evaluator := NewEvaluator()

	// 预编译多个不同的表达式
	expressions := []string{
		"User.Level >= 3",
		"User.Department == \"IT\"",
		"User.Active == true",
		"User.Email contains \"@example.com\"",
		"Resource.Status == \"published\"",
	}

	ctx := &EvaluationContext{
		User: map[string]interface{}{
			"Level":      5,
			"Department": "IT",
			"Active":     true,
			"Email":      "test@example.com",
		},
		Resource: map[string]interface{}{
			"Status": "published",
		},
	}

	// 预热缓存
	for _, expr := range expressions {
		_, _ = evaluator.Evaluate(expr, ctx)
	}

	// 验证所有表达式都被缓存
	for _, expr := range expressions {
		_, ok := evaluator.programCache.Load(expr)
		assert.True(t, ok, "表达式应该被缓存: "+expr)
	}
}
