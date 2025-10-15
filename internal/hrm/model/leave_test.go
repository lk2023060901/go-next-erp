package model

import (
"testing"

"github.com/stretchr/testify/assert"
)

// floatPtr 辅助函数，返回 float64 指针
func floatPtr(f float64) *float64 {
	return &f
}

// TestApprovalRules_GetApprovalChain 测试审批链匹配算法
func TestApprovalRules_GetApprovalChain(t *testing.T) {
	tests := []struct {
		name     string
		rules    *ApprovalRules
		duration float64
		expected int // 期望的审批链长度
	}{
		{
			name: "匹配第一个区间(0-3天)",
			rules: &ApprovalRules{
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
			},
			duration: 2.5,
			expected: 1,
		},
		{
			name: "匹配第二个区间(3-7天)",
			rules: &ApprovalRules{
				DefaultChain: []*ApprovalNode{
					{Level: 1, ApproverType: ApproverTypeDirectManager, Required: true},
				},
				DurationRules: []*DurationRule{
					{
						MinDuration: 3,
						MaxDuration: floatPtr(7),
						ApprovalChain: []*ApprovalNode{
							{Level: 1, ApproverType: ApproverTypeDirectManager, Required: true},
							{Level: 2, ApproverType: ApproverTypeDeptManager, Required: true},
						},
					},
				},
			},
			duration: 5,
			expected: 2,
		},
		{
			name: "无上限区间",
			rules: &ApprovalRules{
				DefaultChain: []*ApprovalNode{
					{Level: 1, ApproverType: ApproverTypeDirectManager, Required: true},
				},
				DurationRules: []*DurationRule{
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
			},
			duration: 100,
			expected: 3,
		},
		{
			name: "没有匹配规则-使用默认链",
			rules: &ApprovalRules{
				DefaultChain: []*ApprovalNode{
					{Level: 1, ApproverType: ApproverTypeDirectManager, Required: true},
					{Level: 2, ApproverType: ApproverTypeHR, Required: true},
				},
				DurationRules: []*DurationRule{},
			},
			duration: 5,
			expected: 2,
		},
		{
			name:     "nil规则-返回nil",
			rules:    nil,
			duration: 5,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
chain := tt.rules.GetApprovalChain(tt.duration)
if tt.expected == 0 {
assert.Nil(t, chain)
} else {
assert.NotNil(t, chain)
assert.Equal(t, tt.expected, len(chain))
}
})
	}
}

// TestApprovalRules_ComplexScenario 测试年假的完整审批规则
func TestApprovalRules_ComplexScenario(t *testing.T) {
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

	testCases := []struct {
		duration      float64
		expectedLen   int
		expectedType  ApproverType
	}{
		{0.5, 1, ApproverTypeDirectManager},
		{2.5, 1, ApproverTypeDirectManager},
		{3, 2, ApproverTypeDeptManager},
		{5, 2, ApproverTypeDeptManager},
		{7, 3, ApproverTypeHR},
		{10, 3, ApproverTypeHR},
	}

	for _, tc := range testCases {
		chain := rules.GetApprovalChain(tc.duration)
		assert.NotNil(t, chain)
		assert.Equal(t, tc.expectedLen, len(chain))
		assert.Equal(t, tc.expectedType, chain[len(chain)-1].ApproverType)
	}
}
