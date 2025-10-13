package model

import (
	"time"

	"github.com/google/uuid"
)

// OrganizationType 组织类型（租户级别配置）
// 允许租户自定义组织层级和类型
// 示例：
//   - 制造业：集团 → 公司 → 工厂 → 车间 → 班组
//   - 互联网：集团 → 公司 → 事业部 → 部门 → 小组
//   - 零售业：集团 → 公司 → 区域 → 门店
type OrganizationType struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;index:idx_org_type_tenant"`

	// 类型标识
	Code string `gorm:"type:varchar(50);not null;uniqueIndex:idx_org_type_code_tenant"` // 类型编码
	Name string `gorm:"type:varchar(100);not null"`                                      // 类型名称
	Icon string `gorm:"type:varchar(100)"`                                               // 图标

	// 层级定义
	Level      int  `gorm:"not null"`      // 建议层级（1=根级别）
	MaxLevel   int  `gorm:"default:10"`    // 允许的最大深度
	AllowRoot  bool `gorm:"default:false"` // 是否允许作为根节点
	AllowMulti bool `gorm:"default:true"`  // 是否允许同一父节点下有多个此类型

	// 父子类型约束（JSON 数组）
	AllowedParentTypes []string `gorm:"type:text[]"` // 允许的父类型编码列表
	AllowedChildTypes  []string `gorm:"type:text[]"` // 允许的子类型编码列表

	// 功能开关
	EnableLeader    bool `gorm:"default:true"`  // 是否启用负责人
	EnableLegalInfo bool `gorm:"default:false"` // 是否启用法人信息
	EnableAddress   bool `gorm:"default:true"`  // 是否启用地址信息

	// 排序和状态
	Sort     int    `gorm:"default:0"`
	Status   string `gorm:"type:varchar(20);default:'active'"` // active, inactive
	IsSystem bool   `gorm:"default:false"`                     // 是否系统预设

	// 审计字段
	CreatedBy uuid.UUID  `gorm:"type:uuid"`
	UpdatedBy uuid.UUID  `gorm:"type:uuid"`
	CreatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time `gorm:"index"`
}

// TableName 指定表名
func (OrganizationType) TableName() string {
	return "organization_types"
}

// CanBeParentOf 判断是否可以作为某类型的父节点
func (t *OrganizationType) CanBeParentOf(childTypeCode string) bool {
	if len(t.AllowedChildTypes) == 0 {
		return true
	}
	for _, allowed := range t.AllowedChildTypes {
		if allowed == childTypeCode {
			return true
		}
	}
	return false
}

// CanBeChildOf 判断是否可以作为某类型的子节点
func (t *OrganizationType) CanBeChildOf(parentTypeCode string) bool {
	if len(t.AllowedParentTypes) == 0 {
		return true
	}
	for _, allowed := range t.AllowedParentTypes {
		if allowed == parentTypeCode {
			return true
		}
	}
	return false
}

// IsActive 是否激活状态
func (t *OrganizationType) IsActive() bool {
	return t.Status == "active"
}
