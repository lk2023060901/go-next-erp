package model

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeSystem   NotificationType = "system"    // 系统通知
	NotificationTypeApproval NotificationType = "approval"  // 审批通知
	NotificationTypeTask     NotificationType = "task"      // 任务通知
	NotificationTypeMessage  NotificationType = "message"   // 消息通知
	NotificationTypeAlert    NotificationType = "alert"     // 警告通知
)

// NotificationChannel 通知渠道
type NotificationChannel string

const (
	NotificationChannelInApp NotificationChannel = "in_app" // 站内消息
	NotificationChannelEmail NotificationChannel = "email"  // 邮件
	NotificationChannelSMS   NotificationChannel = "sms"    // 短信
	NotificationChannelWebhook NotificationChannel = "webhook" // Webhook
)

// NotificationStatus 通知状态
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"   // 待发送
	NotificationStatusSent      NotificationStatus = "sent"      // 已发送
	NotificationStatusDelivered NotificationStatus = "delivered" // 已送达
	NotificationStatusRead      NotificationStatus = "read"      // 已读
	NotificationStatusFailed    NotificationStatus = "failed"    // 发送失败
)

// NotificationPriority 通知优先级
type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"    // 低优先级
	NotificationPriorityNormal NotificationPriority = "normal" // 普通优先级
	NotificationPriorityHigh   NotificationPriority = "high"   // 高优先级
	NotificationPriorityUrgent NotificationPriority = "urgent" // 紧急
)

// Notification 通知模型
type Notification struct {
	ID             uuid.UUID               `json:"id"`
	TenantID       uuid.UUID               `json:"tenant_id"`
	Type           NotificationType        `json:"type"`
	Channel        NotificationChannel     `json:"channel"`
	RecipientID    uuid.UUID               `json:"recipient_id"`     // 接收人ID
	RecipientEmail *string                 `json:"recipient_email"`  // 接收人邮箱（邮件通知）
	RecipientPhone *string                 `json:"recipient_phone"`  // 接收人手机（短信通知）
	Title          string                  `json:"title"`            // 通知标题
	Content        string                  `json:"content"`          // 通知内容
	Data           map[string]interface{}  `json:"data"`             // 附加数据（JSON）
	Priority       NotificationPriority    `json:"priority"`
	Status         NotificationStatus      `json:"status"`
	SentAt         *time.Time              `json:"sent_at"`
	DeliveredAt    *time.Time              `json:"delivered_at"`
	ReadAt         *time.Time              `json:"read_at"`
	RelatedType    *string                 `json:"related_type"`     // 关联类型（如 approval_task）
	RelatedID      *uuid.UUID              `json:"related_id"`       // 关联ID
	ErrorMessage   *string                 `json:"error_message"`    // 错误信息
	RetryCount     int                     `json:"retry_count"`      // 重试次数
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

// NotificationTemplate 通知模板
type NotificationTemplate struct {
	ID          uuid.UUID            `json:"id"`
	TenantID    uuid.UUID            `json:"tenant_id"`
	Code        string               `json:"code"`         // 模板编码
	Name        string               `json:"name"`         // 模板名称
	Type        NotificationType     `json:"type"`
	Channel     NotificationChannel  `json:"channel"`
	Subject     string               `json:"subject"`      // 主题（邮件）
	Template    string               `json:"template"`     // 模板内容（支持变量）
	Variables   []string             `json:"variables"`    // 变量列表
	Enabled     bool                 `json:"enabled"`
	CreatedBy   uuid.UUID            `json:"created_by"`
	UpdatedBy   *uuid.UUID           `json:"updated_by"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	DeletedAt   *time.Time           `json:"deleted_at"`
}
