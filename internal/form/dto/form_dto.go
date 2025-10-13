package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/form/model"
)

// CreateFormDefinitionRequest 创建表单定义请求
type CreateFormDefinitionRequest struct {
	Code   string            `json:"code" binding:"required,max=50"`
	Name   string            `json:"name" binding:"required,max=100"`
	Fields []model.FormField `json:"fields" binding:"required,min=1"`
}

// UpdateFormDefinitionRequest 更新表单定义请求
type UpdateFormDefinitionRequest struct {
	Name   *string            `json:"name" binding:"omitempty,max=100"`
	Fields *[]model.FormField `json:"fields" binding:"omitempty,min=1"`
	Enabled *bool             `json:"enabled"`
}

// FormDefinitionResponse 表单定义响应
type FormDefinitionResponse struct {
	ID        uuid.UUID          `json:"id"`
	TenantID  uuid.UUID          `json:"tenant_id"`
	Code      string             `json:"code"`
	Name      string             `json:"name"`
	Fields    []model.FormField  `json:"fields"`
	Enabled   bool               `json:"enabled"`
	CreatedBy uuid.UUID          `json:"created_by"`
	UpdatedBy *uuid.UUID         `json:"updated_by,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// SubmitFormDataRequest 提交表单数据请求
type SubmitFormDataRequest struct {
	FormID      string                 `json:"form_id" binding:"required"`
	Data        map[string]interface{} `json:"data" binding:"required"`
	RelatedType *string                `json:"related_type"`
	RelatedID   *string                `json:"related_id"`
}

// FormDataResponse 表单数据响应
type FormDataResponse struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	FormID      uuid.UUID              `json:"form_id"`
	Data        map[string]interface{} `json:"data"`
	SubmittedBy uuid.UUID              `json:"submitted_by"`
	SubmittedAt time.Time              `json:"submitted_at"`
	RelatedType *string                `json:"related_type,omitempty"`
	RelatedID   *uuid.UUID             `json:"related_id,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Response 通用响应
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Success 成功响应
func Success(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

// Error 错误响应
func Error(code int, message string) Response {
	return Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
}

// ToFormDefinitionResponse 转换为表单定义响应
func ToFormDefinitionResponse(form *model.FormDefinition) *FormDefinitionResponse {
	return &FormDefinitionResponse{
		ID:        form.ID,
		TenantID:  form.TenantID,
		Code:      form.Code,
		Name:      form.Name,
		Fields:    form.Fields,
		Enabled:   form.Enabled,
		CreatedBy: form.CreatedBy,
		UpdatedBy: form.UpdatedBy,
		CreatedAt: form.CreatedAt,
		UpdatedAt: form.UpdatedAt,
	}
}

// ToFormDefinitionResponseList 转换为表单定义响应列表
func ToFormDefinitionResponseList(forms []*model.FormDefinition) []*FormDefinitionResponse {
	result := make([]*FormDefinitionResponse, len(forms))
	for i, form := range forms {
		result[i] = ToFormDefinitionResponse(form)
	}
	return result
}

// ToFormDataResponse 转换为表单数据响应
func ToFormDataResponse(data *model.FormData) *FormDataResponse {
	return &FormDataResponse{
		ID:          data.ID,
		TenantID:    data.TenantID,
		FormID:      data.FormID,
		Data:        data.Data,
		SubmittedBy: data.SubmittedBy,
		SubmittedAt: data.SubmittedAt,
		RelatedType: data.RelatedType,
		RelatedID:   data.RelatedID,
		CreatedAt:   data.CreatedAt,
		UpdatedAt:   data.UpdatedAt,
	}
}
