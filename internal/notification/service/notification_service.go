package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/notification/dto"
	"github.com/lk2023060901/go-next-erp/internal/notification/model"
	"github.com/lk2023060901/go-next-erp/internal/notification/repository"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrInvalidPriority      = errors.New("invalid priority")
	ErrInvalidType          = errors.New("invalid notification type")
	ErrInvalidChannel       = errors.New("invalid notification channel")
)

// NotificationService 通知服务接口
type NotificationService interface {
	// 发送通知
	SendNotification(ctx context.Context, tenantID uuid.UUID, req *dto.SendNotificationRequest) (*dto.NotificationResponse, error)

	// 获取通知
	GetNotification(ctx context.Context, id uuid.UUID) (*dto.NotificationResponse, error)

	// 列出用户的通知
	ListMyNotifications(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]*dto.NotificationResponse, error)

	// 列出未读通知
	ListUnreadNotifications(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]*dto.NotificationResponse, error)

	// 标记为已读
	MarkAsRead(ctx context.Context, recipientID uuid.UUID, notificationIDs []uuid.UUID) error

	// 标记全部为已读
	MarkAllAsRead(ctx context.Context, recipientID uuid.UUID) error

	// 统计未读数量
	CountUnread(ctx context.Context, recipientID uuid.UUID) (int, error)

	// 设置 WebSocket 推送处理器
	SetPushHandler(handler PushHandler)
}

type notificationService struct {
	repo        repository.NotificationRepository
	emailSender *EmailSender
	pushHandler PushHandler // WebSocket 推送处理器
}

// PushHandler WebSocket 推送处理器接口
type PushHandler interface {
	SendNotificationToUser(userID uuid.UUID, notification map[string]interface{}) error
}

// NewNotificationService 创建通知服务
func NewNotificationService(repo repository.NotificationRepository, emailConfig *EmailConfig) NotificationService {
	var emailSender *EmailSender
	if emailConfig != nil {
		emailSender = NewEmailSender(emailConfig)
	}

	return &notificationService{
		repo:        repo,
		emailSender: emailSender,
		pushHandler: nil, // 稍后通过 SetPushHandler 设置
	}
}

// SetPushHandler 设置 WebSocket 推送处理器
func (s *notificationService) SetPushHandler(handler PushHandler) {
	s.pushHandler = handler
}

func (s *notificationService) SendNotification(ctx context.Context, tenantID uuid.UUID, req *dto.SendNotificationRequest) (*dto.NotificationResponse, error) {
	// 验证类型
	notifType := model.NotificationType(req.Type)
	if !isValidNotificationType(notifType) {
		return nil, ErrInvalidType
	}

	// 验证渠道
	channel := model.NotificationChannel(req.Channel)
	if !isValidNotificationChannel(channel) {
		return nil, ErrInvalidChannel
	}

	// 解析优先级
	priority := model.NotificationPriorityNormal
	if req.Priority != nil {
		priority = model.NotificationPriority(*req.Priority)
		if !isValidPriority(priority) {
			return nil, ErrInvalidPriority
		}
	}

	// 解析接收人ID
	recipientID, err := uuid.Parse(req.RecipientID)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient_id: %w", err)
	}

	// 解析关联ID
	var relatedID *uuid.UUID
	if req.RelatedID != nil {
		id, err := uuid.Parse(*req.RelatedID)
		if err != nil {
			return nil, fmt.Errorf("invalid related_id: %w", err)
		}
		relatedID = &id
	}

	// 创建通知
	now := time.Now()
	notification := &model.Notification{
		ID:             uuid.New(),
		TenantID:       tenantID,
		Type:           notifType,
		Channel:        channel,
		RecipientID:    recipientID,
		RecipientEmail: req.RecipientEmail,
		RecipientPhone: req.RecipientPhone,
		Title:          req.Title,
		Content:        req.Content,
		Data:           req.Data,
		Priority:       priority,
		Status:         model.NotificationStatusPending,
		RelatedType:    req.RelatedType,
		RelatedID:      relatedID,
		RetryCount:     0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// 异步发送通知（根据 channel 类型）
	go s.sendNotification(context.Background(), notification)

	return dto.ToNotificationResponse(notification), nil
}

func (s *notificationService) GetNotification(ctx context.Context, id uuid.UUID) (*dto.NotificationResponse, error) {
	notification, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrNotificationNotFound
	}

	return dto.ToNotificationResponse(notification), nil
}

func (s *notificationService) ListMyNotifications(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]*dto.NotificationResponse, error) {
	notifications, err := s.repo.ListByRecipient(ctx, recipientID, limit, offset)
	if err != nil {
		return nil, err
	}

	return dto.ToNotificationResponseList(notifications), nil
}

func (s *notificationService) ListUnreadNotifications(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]*dto.NotificationResponse, error) {
	notifications, err := s.repo.ListUnread(ctx, recipientID, limit, offset)
	if err != nil {
		return nil, err
	}

	return dto.ToNotificationResponseList(notifications), nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, recipientID uuid.UUID, notificationIDs []uuid.UUID) error {
	// 验证通知是否属于该用户
	for _, id := range notificationIDs {
		notification, err := s.repo.FindByID(ctx, id)
		if err != nil {
			return ErrNotificationNotFound
		}
		if notification.RecipientID != recipientID {
			return fmt.Errorf("notification does not belong to user")
		}
	}

	return s.repo.MarkAsRead(ctx, notificationIDs, time.Now())
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, recipientID uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, recipientID, time.Now())
}

func (s *notificationService) CountUnread(ctx context.Context, recipientID uuid.UUID) (int, error) {
	return s.repo.CountUnread(ctx, recipientID)
}

// 辅助函数

func isValidNotificationType(t model.NotificationType) bool {
	switch t {
	case model.NotificationTypeSystem, model.NotificationTypeApproval,
		model.NotificationTypeTask, model.NotificationTypeMessage,
		model.NotificationTypeAlert:
		return true
	}
	return false
}

func isValidNotificationChannel(c model.NotificationChannel) bool {
	switch c {
	case model.NotificationChannelInApp, model.NotificationChannelEmail,
		model.NotificationChannelSMS, model.NotificationChannelWebhook:
		return true
	}
	return false
}

func isValidPriority(p model.NotificationPriority) bool {
	switch p {
	case model.NotificationPriorityLow, model.NotificationPriorityNormal,
		model.NotificationPriorityHigh, model.NotificationPriorityUrgent:
		return true
	}
	return false
}

// sendNotification 异步发送通知
func (s *notificationService) sendNotification(ctx context.Context, notification *model.Notification) {
	now := time.Now()

	switch notification.Channel {
	case model.NotificationChannelInApp:
		// 站内消息已经创建，直接标记为已发送
		notification.Status = model.NotificationStatusSent
		notification.SentAt = &now
		s.repo.Update(ctx, notification)

		// 通过 WebSocket 推送
		if s.pushHandler != nil {
			notifData := map[string]interface{}{
				"id":         notification.ID.String(),
				"type":       string(notification.Type),
				"title":      notification.Title,
				"content":    notification.Content,
				"priority":   string(notification.Priority),
				"created_at": notification.CreatedAt.Format(time.RFC3339),
			}
			if notification.Data != nil {
				notifData["data"] = notification.Data
			}
			s.pushHandler.SendNotificationToUser(notification.RecipientID, notifData)
		}

	case model.NotificationChannelEmail:
		// 发送邮件
		if s.emailSender == nil {
			errMsg := "email sender not configured"
			notification.Status = model.NotificationStatusFailed
			notification.ErrorMessage = &errMsg
			s.repo.Update(ctx, notification)
			return
		}

		// 获取收件人邮箱（从 notification.RecipientEmail 或其他来源）
		if notification.RecipientEmail == nil || *notification.RecipientEmail == "" {
			errMsg := "recipient email not provided"
			notification.Status = model.NotificationStatusFailed
			notification.ErrorMessage = &errMsg
			s.repo.Update(ctx, notification)
			return
		}

		// 发送邮件
		err := s.emailSender.SendEmail(*notification.RecipientEmail, notification.Title, notification.Content)
		if err != nil {
			errMsg := fmt.Sprintf("failed to send email: %v", err)
			notification.Status = model.NotificationStatusFailed
			notification.ErrorMessage = &errMsg
			s.repo.Update(ctx, notification)
			return
		}

		notification.Status = model.NotificationStatusSent
		notification.SentAt = &now
		s.repo.Update(ctx, notification)

	case model.NotificationChannelSMS:
		// TODO: 集成短信服务
		// 暂时标记为已发送
		notification.Status = model.NotificationStatusSent
		notification.SentAt = &now
		s.repo.Update(ctx, notification)

	case model.NotificationChannelWebhook:
		// TODO: 实现 HTTP 回调
		// 暂时标记为已发送
		notification.Status = model.NotificationStatusSent
		notification.SentAt = &now
		s.repo.Update(ctx, notification)

	default:
		// 未知渠道，标记为失败
		errMsg := "unknown notification channel"
		notification.Status = model.NotificationStatusFailed
		notification.ErrorMessage = &errMsg
		s.repo.Update(ctx, notification)
	}
}
