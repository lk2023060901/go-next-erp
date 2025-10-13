package model

import "github.com/google/uuid"

// OrganizationClosure 组织闭包表
// 用于高效查询组织树的祖先和后代关系
// 采用闭包表（Closure Table）设计模式
type OrganizationClosure struct {
	TenantID     uuid.UUID `json:"tenant_id"`
	AncestorID   uuid.UUID `json:"ancestor_id"`   // 祖先节点
	DescendantID uuid.UUID `json:"descendant_id"` // 后代节点
	Depth        int       `json:"depth"`         // 距离（0=自己，1=直接子节点）
}

// IsSelf 是否自身关系
func (c *OrganizationClosure) IsSelf() bool {
	return c.Depth == 0
}

// IsDirectChild 是否直接子节点
func (c *OrganizationClosure) IsDirectChild() bool {
	return c.Depth == 1
}
