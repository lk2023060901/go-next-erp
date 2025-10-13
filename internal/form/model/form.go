package model

import (
	"time"

	"github.com/google/uuid"
)

// FieldType 表单字段类型
type FieldType string

const (
	FieldTypeText     FieldType = "text"      // 单行文本
	FieldTypeTextarea FieldType = "textarea"  // 多行文本
	FieldTypeNumber   FieldType = "number"    // 数字
	FieldTypeDate     FieldType = "date"      // 日期
	FieldTypeDateTime FieldType = "datetime"  // 日期时间
	FieldTypeSelect   FieldType = "select"    // 下拉选择
	FieldTypeRadio    FieldType = "radio"     // 单选
	FieldTypeCheckbox FieldType = "checkbox"  // 复选
	FieldTypeFile     FieldType = "file"      // 文件上传
	FieldTypeUser     FieldType = "user"      // 用户选择
	FieldTypeDept     FieldType = "department" // 部门选择
)

// FormField 表单字段定义
type FormField struct {
	Key          string                 `json:"key"`           // 字段标识
	Label        string                 `json:"label"`         // 字段标签
	Type         FieldType              `json:"type"`          // 字段类型
	Required     bool                   `json:"required"`      // 是否必填
	DefaultValue interface{}            `json:"default_value"` // 默认值
	Placeholder  *string                `json:"placeholder"`   // 占位符
	Options      []FieldOption          `json:"options"`       // 选项（用于 select/radio/checkbox）
	Rules        []ValidationRule       `json:"rules"`         // 验证规则
	Properties   map[string]interface{} `json:"properties"`    // 其他属性
	Sort         int                    `json:"sort"`          // 排序
}

// FieldOption 字段选项
type FieldOption struct {
	Label string      `json:"label"` // 选项标签
	Value interface{} `json:"value"` // 选项值
}

// ValidationRule 验证规则
type ValidationRule struct {
	Type    string      `json:"type"`    // 规则类型：required/min/max/pattern/custom
	Value   interface{} `json:"value"`   // 规则值
	Message string      `json:"message"` // 错误消息
}

// FormDefinition 表单定义
type FormDefinition struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	Code      string     `json:"code"`        // 表单编码（唯一标识）
	Name      string     `json:"name"`        // 表单名称
	Fields    []FormField `json:"fields"`     // 字段列表
	Enabled   bool       `json:"enabled"`     // 是否启用
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// FormData 表单数据
type FormData struct {
	ID             uuid.UUID              `json:"id"`
	TenantID       uuid.UUID              `json:"tenant_id"`
	FormID         uuid.UUID              `json:"form_id"`          // 关联表单定义
	Data           map[string]interface{} `json:"data"`             // 表单数据（JSON）
	SubmittedBy    uuid.UUID              `json:"submitted_by"`     // 提交人
	SubmittedAt    time.Time              `json:"submitted_at"`
	RelatedType    *string                `json:"related_type"`     // 关联类型（如 approval_process）
	RelatedID      *uuid.UUID             `json:"related_id"`       // 关联ID
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}
