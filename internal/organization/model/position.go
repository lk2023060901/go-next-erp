package model

import (
	"time"

	"github.com/google/uuid"
)

// Position 职位
// 定义组织内的职位（如：总经理、部门经理、工程师等）
type Position struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 职位信息
	Code        string `json:"code"`        // 职位编码
	Name        string `json:"name"`        // 职位名称
	Description string `json:"description"` // 职位描述

	// 关联组织（可选，NULL 表示全局职位）
	OrgID *uuid.UUID `json:"org_id"` // 所属组织

	// 职级
	Level int `json:"level"` // 职级（数字越大级别越高）

	// 职位类别
	Category string `json:"category"` // 职位类别：management, technical, sales, support

	// 排序和状态
	Sort   int    `json:"sort"`
	Status string `json:"status"` // active, inactive

	// 审计字段
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy uuid.UUID  `json:"updated_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// IsActive 是否激活
func (p *Position) IsActive() bool {
	return p.Status == "active"
}

// IsGlobal 是否全局职位
func (p *Position) IsGlobal() bool {
	return p.OrgID == nil
}
