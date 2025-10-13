package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/notification/model"
)

// SendNotificationRequest 发送通知请求
type SendNotificationRequest struct {
	Type           string                 `json:"type" binding:"required"`
	Channel        string                 `json:"channel" binding:"required"`
	RecipientID    string                 `json:"recipient_id" binding:"required"`
	RecipientEmail *string                `json:"recipient_email"`
	RecipientPhone *string                `json:"recipient_phone"`
	Title          string                 `json:"title" binding:"required,max=200"`
	Content        string                 `json:"content" binding:"required"`
	Data           map[string]interface{} `json:"data"`
	Priority       *string                `json:"priority"`
	RelatedType    *string                `json:"related_type"`
	RelatedID      *string                `json:"related_id"`
}

// SendByTemplateRequest 使用模板发送通知请求
type SendByTemplateRequest struct {
	TemplateCode   string                 `json:"template_code" binding:"required"`
	RecipientID    string                 `json:"recipient_id" binding:"required"`
	RecipientEmail *string                `json:"recipient_email"`
	RecipientPhone *string                `json:"recipient_phone"`
	Variables      map[string]interface{} `json:"variables" binding:"required"`
	Priority       *string                `json:"priority"`
	RelatedType    *string                `json:"related_type"`
	RelatedID      *string                `json:"related_id"`
}

// MarkAsReadRequest 标记为已读请求
type MarkAsReadRequest struct {
	NotificationIDs []string `json:"notification_ids" binding:"required,min=1"`
}

// NotificationResponse 通知响应
type NotificationResponse struct {
	ID             uuid.UUID                   `json:"id"`
	TenantID       uuid.UUID                   `json:"tenant_id"`
	Type           model.NotificationType      `json:"type"`
	Channel        model.NotificationChannel   `json:"channel"`
	RecipientID    uuid.UUID                   `json:"recipient_id"`
	RecipientEmail *string                     `json:"recipient_email,omitempty"`
	RecipientPhone *string                     `json:"recipient_phone,omitempty"`
	Title          string                      `json:"title"`
	Content        string                      `json:"content"`
	Data           map[string]interface{}      `json:"data,omitempty"`
	Priority       model.NotificationPriority  `json:"priority"`
	Status         model.NotificationStatus    `json:"status"`
	SentAt         *time.Time                  `json:"sent_at,omitempty"`
	DeliveredAt    *time.Time                  `json:"delivered_at,omitempty"`
	ReadAt         *time.Time                  `json:"read_at,omitempty"`
	RelatedType    *string                     `json:"related_type,omitempty"`
	RelatedID      *uuid.UUID                  `json:"related_id,omitempty"`
	ErrorMessage   *string                     `json:"error_message,omitempty"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
}

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	Code      string   `json:"code" binding:"required,max=50"`
	Name      string   `json:"name" binding:"required,max=100"`
	Type      string   `json:"type" binding:"required"`
	Channel   string   `json:"channel" binding:"required"`
	Subject   string   `json:"subject" binding:"max=200"`
	Template  string   `json:"template" binding:"required"`
	Variables []string `json:"variables"`
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	Name      *string  `json:"name" binding:"omitempty,max=100"`
	Subject   *string  `json:"subject" binding:"omitempty,max=200"`
	Template  *string  `json:"template"`
	Variables *[]string `json:"variables"`
	Enabled   *bool    `json:"enabled"`
}

// TemplateResponse 模板响应
type TemplateResponse struct {
	ID        uuid.UUID                  `json:"id"`
	TenantID  uuid.UUID                  `json:"tenant_id"`
	Code      string                     `json:"code"`
	Name      string                     `json:"name"`
	Type      model.NotificationType     `json:"type"`
	Channel   model.NotificationChannel  `json:"channel"`
	Subject   string                     `json:"subject"`
	Template  string                     `json:"template"`
	Variables []string                   `json:"variables"`
	Enabled   bool                       `json:"enabled"`
	CreatedBy uuid.UUID                  `json:"created_by"`
	UpdatedBy *uuid.UUID                 `json:"updated_by,omitempty"`
	CreatedAt time.Time                  `json:"created_at"`
	UpdatedAt time.Time                  `json:"updated_at"`
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

// ToNotificationResponse 转换为通知响应
func ToNotificationResponse(n *model.Notification) *NotificationResponse {
	return &NotificationResponse{
		ID:             n.ID,
		TenantID:       n.TenantID,
		Type:           n.Type,
		Channel:        n.Channel,
		RecipientID:    n.RecipientID,
		RecipientEmail: n.RecipientEmail,
		RecipientPhone: n.RecipientPhone,
		Title:          n.Title,
		Content:        n.Content,
		Data:           n.Data,
		Priority:       n.Priority,
		Status:         n.Status,
		SentAt:         n.SentAt,
		DeliveredAt:    n.DeliveredAt,
		ReadAt:         n.ReadAt,
		RelatedType:    n.RelatedType,
		RelatedID:      n.RelatedID,
		ErrorMessage:   n.ErrorMessage,
		CreatedAt:      n.CreatedAt,
		UpdatedAt:      n.UpdatedAt,
	}
}

// ToNotificationResponseList 转换为通知响应列表
func ToNotificationResponseList(notifications []*model.Notification) []*NotificationResponse {
	result := make([]*NotificationResponse, len(notifications))
	for i, n := range notifications {
		result[i] = ToNotificationResponse(n)
	}
	return result
}

// ToTemplateResponse 转换为模板响应
func ToTemplateResponse(t *model.NotificationTemplate) *TemplateResponse {
	return &TemplateResponse{
		ID:        t.ID,
		TenantID:  t.TenantID,
		Code:      t.Code,
		Name:      t.Name,
		Type:      t.Type,
		Channel:   t.Channel,
		Subject:   t.Subject,
		Template:  t.Template,
		Variables: t.Variables,
		Enabled:   t.Enabled,
		CreatedBy: t.CreatedBy,
		UpdatedBy: t.UpdatedBy,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// ToTemplateResponseList 转换为模板响应列表
func ToTemplateResponseList(templates []*model.NotificationTemplate) []*TemplateResponse {
	result := make([]*TemplateResponse, len(templates))
	for i, t := range templates {
		result[i] = ToTemplateResponse(t)
	}
	return result
}
