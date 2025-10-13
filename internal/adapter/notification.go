package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	notifyv1 "github.com/lk2023060901/go-next-erp/api/notification/v1"
	"github.com/lk2023060901/go-next-erp/internal/notification/service"
)

// NotificationAdapter 通知适配器
type NotificationAdapter struct {
	notifyv1.UnimplementedNotificationServiceServer
	notifyService service.NotificationService
}

// NewNotificationAdapter 创建通知适配器
func NewNotificationAdapter(notifyService service.NotificationService) *NotificationAdapter {
	return &NotificationAdapter{
		notifyService: notifyService,
	}
}

// SendNotification 发送通知（简化实现）
func (a *NotificationAdapter) SendNotification(ctx context.Context, req *notifyv1.SendNotificationRequest) (*notifyv1.NotificationResponse, error) {
	// TODO: 实现完整的通知发送逻辑
	return &notifyv1.NotificationResponse{
		Id:        uuid.New().String(),
		TenantId:  req.TenantId,
		UserId:    req.UserId,
		Type:      req.Type,
		Title:     req.Title,
		Content:   req.Content,
		Link:      req.Link,
		Priority:  req.Priority,
		IsRead:    false,
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// GetNotification 获取通知（简化实现）
func (a *NotificationAdapter) GetNotification(ctx context.Context, req *notifyv1.GetNotificationRequest) (*notifyv1.NotificationResponse, error) {
	// TODO: 实现通知查询逻辑
	return &notifyv1.NotificationResponse{
		Id:        req.Id,
		IsRead:    false,
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// ListNotifications 列出通知（简化实现）
func (a *NotificationAdapter) ListNotifications(ctx context.Context, req *notifyv1.ListNotificationsRequest) (*notifyv1.ListNotificationsResponse, error) {
	// TODO: 实现通知列表查询逻辑
	return &notifyv1.ListNotificationsResponse{
		Items: []*notifyv1.NotificationResponse{},
		Total: 0,
	}, nil
}

// MarkAsRead 标记为已读（简化实现）
func (a *NotificationAdapter) MarkAsRead(ctx context.Context, req *notifyv1.MarkAsReadRequest) (*notifyv1.NotificationResponse, error) {
	// TODO: 实现标记已读逻辑
	return &notifyv1.NotificationResponse{
		Id:     req.Id,
		IsRead: true,
		ReadAt: time.Now().Format(time.RFC3339),
	}, nil
}

// BatchMarkAsRead 批量标记为已读（简化实现）
func (a *NotificationAdapter) BatchMarkAsRead(ctx context.Context, req *notifyv1.BatchMarkAsReadRequest) (*notifyv1.BatchMarkAsReadResponse, error) {
	// TODO: 实现批量标记已读逻辑
	return &notifyv1.BatchMarkAsReadResponse{
		Count:   int32(len(req.Ids)),
		Success: true,
	}, nil
}

// DeleteNotification 删除通知（简化实现）
func (a *NotificationAdapter) DeleteNotification(ctx context.Context, req *notifyv1.DeleteNotificationRequest) (*notifyv1.DeleteNotificationResponse, error) {
	// TODO: 实现删除逻辑
	return &notifyv1.DeleteNotificationResponse{Success: true}, nil
}

// GetUnreadCount 获取未读数量（简化实现）
func (a *NotificationAdapter) GetUnreadCount(ctx context.Context, req *notifyv1.GetUnreadCountRequest) (*notifyv1.UnreadCountResponse, error) {
	// TODO: 实现未读数量查询逻辑
	return &notifyv1.UnreadCountResponse{Count: 0}, nil
}
