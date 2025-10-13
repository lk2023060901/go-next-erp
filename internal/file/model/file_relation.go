package model

import (
"time"

"github.com/google/uuid"
)

// FileRelation 文件关联模型
type FileRelation struct {
	ID       uuid.UUID `json:"id"`
	FileID   uuid.UUID `json:"file_id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 关联实体
	EntityType string    `json:"entity_type"` // approval_task/form_data/workflow_instance/employee
	EntityID   uuid.UUID `json:"entity_id"`   // 关联实体 ID

	// 关联元数据
	FieldName    *string `json:"field_name"`     // 字段名（表单附件）
	RelationType string  `json:"relation_type"`  // attachment/avatar/evidence/report
	Description  *string `json:"description"`    // 关联描述
	SortOrder    int     `json:"sort_order"`     // 显示顺序

	// 访问控制
	CreatedBy uuid.UUID `json:"created_by"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
}

// RelationType 关联类型
type RelationType string

const (
RelationTypeAttachment RelationType = "attachment" // 附件
RelationTypeAvatar     RelationType = "avatar"     // 头像
RelationTypeEvidence   RelationType = "evidence"   // 证据
RelationTypeReport     RelationType = "report"     // 报告
RelationTypeDocument   RelationType = "document"   // 文档
RelationTypeCertificate RelationType = "certificate" // 证书
)

// EntityType 实体类型
type EntityType string

const (
EntityTypeApprovalTask      EntityType = "approval_task"       // 审批任务
EntityTypeFormData          EntityType = "form_data"           // 表单数据
EntityTypeWorkflowInstance  EntityType = "workflow_instance"   // 工作流实例
EntityTypeEmployee          EntityType = "employee"            // 员工
EntityTypeOrganization      EntityType = "organization"        // 组织
EntityTypeContract          EntityType = "contract"            // 合同
EntityTypeProject           EntityType = "project"             // 项目
)
