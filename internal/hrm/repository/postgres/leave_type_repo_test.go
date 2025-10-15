package postgres

import (
"context"
"encoding/json"
"testing"

"github.com/google/uuid"
"github.com/lk2023060901/go-next-erp/internal/hrm/model"
"github.com/stretchr/testify/assert"
)

// TestApprovalRules_JSONSerialization 测试审批规则的 JSON 序列化
func TestApprovalRules_JSONSerialization(t *testing.T) {
	tests := []struct {
		name  string
		rules *model.ApprovalRules
	}{
		{
			name: "简单的单级审批",
			rules: &model.ApprovalRules{
				DefaultChain: []*model.ApprovalNode{
					{Level: 1, ApproverType: model.ApproverTypeDirectManager, Required: true},
				},
				DurationRules: []*model.DurationRule{},
			},
		},
		{
			name: "完整的三级分段审批",
			rules: &model.ApprovalRules{
				DefaultChain: []*model.ApprovalNode{
					{Level: 1, ApproverType: model.ApproverTypeDirectManager, Required: true},
				},
				DurationRules: []*model.DurationRule{
					{
						MinDuration: 0,
						MaxDuration: floatPtr(3),
						ApprovalChain: []*model.ApprovalNode{
							{Level: 1, ApproverType: model.ApproverTypeDirectManager, Required: true},
						},
					},
					{
						MinDuration: 3,
						MaxDuration: floatPtr(7),
						ApprovalChain: []*model.ApprovalNode{
							{Level: 1, ApproverType: model.ApproverTypeDirectManager, Required: true},
							{Level: 2, ApproverType: model.ApproverTypeDeptManager, Required: true},
						},
					},
					{
						MinDuration: 7,
						MaxDuration: nil,
						ApprovalChain: []*model.ApprovalNode{
							{Level: 1, ApproverType: model.ApproverTypeDirectManager, Required: true},
							{Level: 2, ApproverType: model.ApproverTypeDeptManager, Required: true},
							{Level: 3, ApproverType: model.ApproverTypeHR, Required: true},
						},
					},
				},
			},
		},
		{
			name: "包含自定义审批人",
			rules: &model.ApprovalRules{
				DefaultChain: []*model.ApprovalNode{
					{Level: 1, ApproverType: model.ApproverTypeCustom, ApproverID: strPtr("custom-approver-123"), Required: true},
				},
				DurationRules: []*model.DurationRule{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
// 序列化
data, err := json.Marshal(tt.rules)
assert.NoError(t, err)
assert.NotEmpty(t, data)

// 反序列化
var decoded model.ApprovalRules
err = json.Unmarshal(data, &decoded)
			assert.NoError(t, err)

			// 验证数据一致性
			assert.Equal(t, len(tt.rules.DefaultChain), len(decoded.DefaultChain))
			assert.Equal(t, len(tt.rules.DurationRules), len(decoded.DurationRules))

			// 验证默认链
			if len(tt.rules.DefaultChain) > 0 {
				assert.Equal(t, tt.rules.DefaultChain[0].Level, decoded.DefaultChain[0].Level)
				assert.Equal(t, tt.rules.DefaultChain[0].ApproverType, decoded.DefaultChain[0].ApproverType)
				assert.Equal(t, tt.rules.DefaultChain[0].Required, decoded.DefaultChain[0].Required)
			}

			// 验证天数规则
			if len(tt.rules.DurationRules) > 0 {
				for i, rule := range tt.rules.DurationRules {
					assert.Equal(t, rule.MinDuration, decoded.DurationRules[i].MinDuration)
					if rule.MaxDuration != nil {
						assert.NotNil(t, decoded.DurationRules[i].MaxDuration)
						assert.Equal(t, *rule.MaxDuration, *decoded.DurationRules[i].MaxDuration)
					}
				}
			}
		})
	}
}

// TestApprovalRules_JSONFormat 测试 JSON 格式是否符合预期
func TestApprovalRules_JSONFormat(t *testing.T) {
	rules := &model.ApprovalRules{
		DefaultChain: []*model.ApprovalNode{
			{Level: 1, ApproverType: model.ApproverTypeDirectManager, Required: true},
		},
		DurationRules: []*model.DurationRule{
			{
				MinDuration: 0,
				MaxDuration: floatPtr(3),
				ApprovalChain: []*model.ApprovalNode{
					{Level: 1, ApproverType: model.ApproverTypeDirectManager, Required: true},
				},
			},
		},
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	assert.NoError(t, err)

	// 验证 JSON 包含关键字段
	jsonStr := string(data)
	assert.Contains(t, jsonStr, "default_chain")
	assert.Contains(t, jsonStr, "duration_rules")
	assert.Contains(t, jsonStr, "min_duration")
	assert.Contains(t, jsonStr, "max_duration")
	assert.Contains(t, jsonStr, "approval_chain")
	assert.Contains(t, jsonStr, "level")
	assert.Contains(t, jsonStr, "approver_type")
	assert.Contains(t, jsonStr, "required")
}

// TestApprovalRules_EmptyRules 测试空规则的序列化
func TestApprovalRules_EmptyRules(t *testing.T) {
	tests := []struct {
		name  string
		rules *model.ApprovalRules
	}{
		{
			name:  "nil规则",
			rules: nil,
		},
		{
			name: "空的默认链和天数规则",
			rules: &model.ApprovalRules{
				DefaultChain:  []*model.ApprovalNode{},
				DurationRules: []*model.DurationRule{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
data, err := json.Marshal(tt.rules)
assert.NoError(t, err)

if tt.rules != nil {
assert.NotEmpty(t, data)
}
})
	}
}

// TestLeaveType_ApprovalRulesSerialization 测试 LeaveType 中的审批规则序列化
func TestLeaveType_ApprovalRulesSerialization(t *testing.T) {
	ctx := context.Background()
	
	// 创建测试数据
	leaveType := &model.LeaveType{
		ID:       uuid.Must(uuid.NewV7()),
		TenantID: uuid.Must(uuid.NewV7()),
		Code:     "annual_leave",
		Name:     "年假",
		ApprovalRules: &model.ApprovalRules{
			DefaultChain: []*model.ApprovalNode{
				{Level: 1, ApproverType: model.ApproverTypeDirectManager, Required: true},
			},
			DurationRules: []*model.DurationRule{
				{
					MinDuration: 0,
					MaxDuration: floatPtr(3),
					ApprovalChain: []*model.ApprovalNode{
						{Level: 1, ApproverType: model.ApproverTypeDirectManager, Required: true},
					},
				},
			},
		},
	}

	// 序列化审批规则
	rulesJSON, err := json.Marshal(leaveType.ApprovalRules)
	assert.NoError(t, err)
	assert.NotEmpty(t, rulesJSON)

	// 反序列化
	var decodedRules model.ApprovalRules
	err = json.Unmarshal(rulesJSON, &decodedRules)
	assert.NoError(t, err)

	// 验证数据
	assert.NotNil(t, decodedRules.DefaultChain)
	assert.NotNil(t, decodedRules.DurationRules)
	assert.Equal(t, 1, len(decodedRules.DefaultChain))
	assert.Equal(t, 1, len(decodedRules.DurationRules))

	// 测试反序列化后的规则匹配功能
	chain := decodedRules.GetApprovalChain(2.5)
	assert.NotNil(t, chain)
	assert.Equal(t, 1, len(chain))

	_ = ctx // 避免未使用变量警告
}

// 辅助函数
func floatPtr(f float64) *float64 {
	return &f
}

func strPtr(s string) *string {
	return &s
}
