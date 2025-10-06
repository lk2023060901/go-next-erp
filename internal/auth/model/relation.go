package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RelationTuple ReBAC 关系元组（类似 Google Zanzibar）
type RelationTuple struct {
	ID       uuid.UUID `json:"id"`        // UUID v7
	TenantID uuid.UUID `json:"tenant_id"` // 租户 ID

	// 关系元组：(subject, relation, object)
	Subject  string `json:"subject"`  // 主体（如：user:123, group:456）
	Relation string `json:"relation"` // 关系（如：owner, editor, viewer）
	Object   string `json:"object"`   // 客体（如：document:789, folder:abc）

	// 时间戳
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"-"` // 软删除
}

// String 返回元组字符串表示
func (t *RelationTuple) String() string {
	return fmt.Sprintf("(%s, %s, %s)", t.Subject, t.Relation, t.Object)
}

// 关系类型定义
const (
	RelationOwner  = "owner"  // 所有者
	RelationEditor = "editor" // 编辑者
	RelationViewer = "viewer" // 查看者
	RelationMember = "member" // 成员
	RelationAdmin  = "admin"  // 管理员
	RelationParent = "parent" // 父级（层级关系）
)

// 元组示例：
// (user:123, owner, document:789) - 用户123是文档789的所有者
// (user:456, editor, document:789) - 用户456是文档789的编辑者
// (group:sales, member, user:123) - 用户123是销售组的成员
// (folder:abc, parent, document:789) - 文件夹abc是文档789的父级
