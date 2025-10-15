package model

import (
"encoding/json"
"testing"

"github.com/stretchr/testify/assert"
)

// TestApprovalRules_JSONSerialization 测试JSON序列化
func TestApprovalRules_JSONSerialization(t *testing.T) {
	rules := &ApprovalRules{
		DefaultChain: []*ApprovalNode{
			{
				Level:        1,
				ApproverType: ApproverTypeDirectManager,
				Required:     true,
			},
		},
		DurationRules: []*DurationRule{
			{
				MinDuration: 0,
				MaxDuration: floatPtr(3),
				ApprovalChain: []*ApprovalNode{
					{Level: 1, ApproverType: ApproverTypeDirectManager, Required: true},
				},
			},
			{
				MinDuration: 3,
				MaxDuration: floatPtr(7),
				ApprovalChain: []*ApprovalNode{
					{Level: 1, ApproverType: ApproverTypeDirectManager, Required: true},
					{Level: 2, ApproverType: ApproverTypeDeptManager, Required: true},
				},
			},
			{
				MinDuration: 7,
				MaxDuration: nil,
				ApprovalChain: []*ApprovalNode{
					{Level: 1, ApproverType: ApproverTypeDirectManager, Required: true},
					{Level: 2, ApproverType: ApproverTypeDeptManager, Required: true},
					{Level: 3, ApproverType: ApproverTypeHR, Required: true},
				},
			},
		},
	}

	// 序列化
	data, err := json.Marshal(rules)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// 反序列化
	var decoded ApprovalRules
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	// 验证默认链
	assert.Equal(t, len(rules.DefaultChain), len(decoded.DefaultChain))
	assert.Equal(t, rules.DefaultChain[0].Level, decoded.DefaultChain[0].Level)
	assert.Equal(t, rules.DefaultChain[0].ApproverType, decoded.DefaultChain[0].ApproverType)

	// 验证天数规则
	assert.Equal(t, len(rules.DurationRules), len(decoded.DurationRules))
	
	// 第一个规则
	assert.Equal(t, rules.DurationRules[0].MinDuration, decoded.DurationRules[0].MinDuration)
	assert.Equal(t, *rules.DurationRules[0].MaxDuration, *decoded.DurationRules[0].MaxDuration)
	
	// 第三个规则（无上限）
	assert.Nil(t, decoded.DurationRules[2].MaxDuration)
	assert.Equal(t, 3, len(decoded.DurationRules[2].ApprovalChain))
}

// TestApprovalNode_JSONSerialization 测试审批节点JSON序列化
func TestApprovalNode_JSONSerialization(t *testing.T) {
	customID := "custom-approver-123"
	
	tests := []struct {
		name string
		node *ApprovalNode
	}{
		{
			name: "普通节点",
			node: &ApprovalNode{
				Level:        1,
				ApproverType: ApproverTypeDirectManager,
				Required:     true,
			},
		},
		{
			name: "自定义审批人节点",
			node: &ApprovalNode{
				Level:        2,
				ApproverType: ApproverTypeCustom,
				ApproverID:   &customID,
				Required:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
// 序列化
data, err := json.Marshal(tt.node)
assert.NoError(t, err)

// 反序列化
var decoded ApprovalNode
err = json.Unmarshal(data, &decoded)
			assert.NoError(t, err)
			
			// 验证
			assert.Equal(t, tt.node.Level, decoded.Level)
			assert.Equal(t, tt.node.ApproverType, decoded.ApproverType)
			assert.Equal(t, tt.node.Required, decoded.Required)
			
			if tt.node.ApproverID != nil {
				assert.NotNil(t, decoded.ApproverID)
				assert.Equal(t, *tt.node.ApproverID, *decoded.ApproverID)
			}
		})
	}
}

// TestApprovalRules_JSONFormat 测试JSON格式
func TestApprovalRules_JSONFormat(t *testing.T) {
	rules := &ApprovalRules{
		DefaultChain: []*ApprovalNode{
			{Level: 1, ApproverType: ApproverTypeDirectManager, Required: true},
		},
		DurationRules: []*DurationRule{
			{
				MinDuration: 0,
				MaxDuration: floatPtr(3),
				ApprovalChain: []*ApprovalNode{
					{Level: 1, ApproverType: ApproverTypeDirectManager, Required: true},
				},
			},
		},
	}

	// 美化输出
	data, err := json.MarshalIndent(rules, "", "  ")
	assert.NoError(t, err)
	
	// 验证JSON格式
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)
	
	// 验证字段存在
	assert.Contains(t, result, "default_chain")
	assert.Contains(t, result, "duration_rules")
}

// TestApprovalRules_EmptyRules 测试空规则
func TestApprovalRules_EmptyRules(t *testing.T) {
	tests := []struct {
		name  string
		rules *ApprovalRules
	}{
		{
			name: "空的DurationRules",
			rules: &ApprovalRules{
				DefaultChain: []*ApprovalNode{
					{Level: 1, ApproverType: ApproverTypeDirectManager, Required: true},
				},
				DurationRules: []*DurationRule{},
			},
		},
		{
			name: "nil的DefaultChain",
			rules: &ApprovalRules{
				DefaultChain: nil,
				DurationRules: []*DurationRule{
					{
						MinDuration: 0,
						MaxDuration: floatPtr(3),
						ApprovalChain: []*ApprovalNode{
							{Level: 1, ApproverType: ApproverTypeDirectManager, Required: true},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
data, err := json.Marshal(tt.rules)
assert.NoError(t, err)

var decoded ApprovalRules
err = json.Unmarshal(data, &decoded)
			assert.NoError(t, err)
		})
	}
}
