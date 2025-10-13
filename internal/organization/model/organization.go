package model

import (
	"time"

	"github.com/google/uuid"
)

// Organization 组织实体
// 支持多级树形结构：集团 → 公司 → 事业部 → 部门 → 小组
type Organization struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 组织基本信息
	Code        string `json:"code"`         // 组织编码
	Name        string `json:"name"`         // 组织名称
	ShortName   string `json:"short_name"`   // 简称
	Description string `json:"description"`  // 描述

	// 组织类型（关联组织类型表）
	TypeID   uuid.UUID `json:"type_id"`
	TypeCode string    `json:"type_code"` // 冗余字段，便于查询

	// 树形结构字段
	ParentID    *uuid.UUID `json:"parent_id"`     // 父组织 ID
	Level       int        `json:"level"`         // 层级（1=根节点）
	Path        string     `json:"path"`          // 路径："/uuid1/uuid2/uuid3/"
	PathNames   string     `json:"path_names"`    // 路径名称："/集团/公司A/部门B/"
	AncestorIDs []string   `json:"ancestor_ids"`  // 所有祖先节点 ID
	IsLeaf      bool       `json:"is_leaf"`       // 是否叶子节点

	// 组织负责人
	LeaderID   *uuid.UUID `json:"leader_id"`   // 负责人用户 ID
	LeaderName string     `json:"leader_name"` // 负责人姓名（冗余）

	// 公司法人信息（仅公司类型需要）
	LegalPerson  string     `json:"legal_person"`  // 法定代表人
	UnifiedCode  string     `json:"unified_code"`  // 统一社会信用代码
	RegisterDate *time.Time `json:"register_date"` // 注册日期
	RegisterAddr string     `json:"register_addr"` // 注册地址

	// 联系方式
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`

	// 统计信息（冗余字段，定期更新）
	EmployeeCount  int `json:"employee_count"`   // 员工数量（含子组织）
	DirectEmpCount int `json:"direct_emp_count"` // 直属员工数量

	// 排序和状态
	Sort   int      `json:"sort"`   // 同级排序
	Status string   `json:"status"` // active, inactive, disbanded
	Tags   []string `json:"tags"`   // 标签

	// 审计字段
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy uuid.UUID  `json:"updated_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// IsRoot 是否根节点
func (o *Organization) IsRoot() bool {
	return o.ParentID == nil || o.Level == 1
}

// HasChildren 是否有子节点
func (o *Organization) HasChildren() bool {
	return !o.IsLeaf
}

// GetFullPath 获取完整路径名称
func (o *Organization) GetFullPath() string {
	if o.PathNames != "" {
		return o.PathNames
	}
	return o.Name
}

// IsActive 是否激活
func (o *Organization) IsActive() bool {
	return o.Status == "active"
}
