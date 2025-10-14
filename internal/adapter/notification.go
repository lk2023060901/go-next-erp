package adapter

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	notifyv1 "github.com/lk2023060901/go-next-erp/api/notification/v1"
	"github.com/lk2023060901/go-next-erp/internal/notification/dto"
	"github.com/lk2023060901/go-next-erp/internal/notification/service"
	"github.com/lk2023060901/go-next-erp/pkg/middleware"
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

// SendNotification 发送通知
func (a *NotificationAdapter) SendNotification(ctx context.Context, req *notifyv1.SendNotificationRequest) (*notifyv1.NotificationResponse, error) {
	// 获取租户ID
	tenantID, ok := middleware.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant_id not found in context")
	}

	// 构造请求
	priority := "normal"
	if req.Priority != "" {
		priority = req.Priority
	}

	sendReq := &dto.SendNotificationRequest{
		Type:        req.Type,
		Channel:     "in_app", // 默认使用站内消息
		RecipientID: req.UserId,
		Title:       req.Title,
		Content:     req.Content,
		Priority:    &priority,
	}

	// 发送通知
	notif, err := a.notifyService.SendNotification(ctx, tenantID, sendReq)
	if err != nil {
		return nil, fmt.Errorf("send notification failed: %w", err)
	}

	// 转换响应
	return toProtoNotification(notif), nil
}

// GetNotification 获取通知
func (a *NotificationAdapter) GetNotification(ctx context.Context, req *notifyv1.GetNotificationRequest) (*notifyv1.NotificationResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	notif, err := a.notifyService.GetNotification(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get notification failed: %w", err)
	}

	return toProtoNotification(notif), nil
}

// ListNotifications 列出通知
func (a *NotificationAdapter) ListNotifications(ctx context.Context, req *notifyv1.ListNotificationsRequest) (*notifyv1.ListNotificationsResponse, error) {
	// 获取当前用户ID
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("user_id not found in context")
	}

	// 解析分页参数
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// 查询通知列表
	var notifications []*dto.NotificationResponse
	var err error

	if req.OnlyUnread {
		notifications, err = a.notifyService.ListUnreadNotifications(ctx, userID, pageSize, offset)
	} else {
		notifications, err = a.notifyService.ListMyNotifications(ctx, userID, pageSize, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("list notifications failed: %w", err)
	}

	// 转换响应
	items := make([]*notifyv1.NotificationResponse, len(notifications))
	for i, notif := range notifications {
		items[i] = toProtoNotification(notif)
	}

	return &notifyv1.ListNotificationsResponse{
		Items: items,
		Total: int32(len(items)),
	}, nil
}

// MarkAsRead 标记为已读
func (a *NotificationAdapter) MarkAsRead(ctx context.Context, req *notifyv1.MarkAsReadRequest) (*notifyv1.NotificationResponse, error) {
	// 获取当前用户ID
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("user_id not found in context")
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	// 标记已读
	err = a.notifyService.MarkAsRead(ctx, userID, []uuid.UUID{id})
	if err != nil {
		return nil, fmt.Errorf("mark as read failed: %w", err)
	}

	// 返回更新后的通知
	notif, err := a.notifyService.GetNotification(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get notification failed: %w", err)
	}

	return toProtoNotification(notif), nil
}

// BatchMarkAsRead 批量标记为已读
func (a *NotificationAdapter) BatchMarkAsRead(ctx context.Context, req *notifyv1.BatchMarkAsReadRequest) (*notifyv1.BatchMarkAsReadResponse, error) {
	// 获取当前用户ID
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("user_id not found in context")
	}

	// 解析ID列表
	ids := make([]uuid.UUID, 0, len(req.Ids))
	for _, idStr := range req.Ids {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid id %s: %w", idStr, err)
		}
		ids = append(ids, id)
	}

	// 批量标记已读
	err := a.notifyService.MarkAsRead(ctx, userID, ids)
	if err != nil {
		return nil, fmt.Errorf("batch mark as read failed: %w", err)
	}

	return &notifyv1.BatchMarkAsReadResponse{
		Count:   int32(len(ids)),
		Success: true,
	}, nil
}

// DeleteNotification 删除通知
func (a *NotificationAdapter) DeleteNotification(ctx context.Context, req *notifyv1.DeleteNotificationRequest) (*notifyv1.DeleteNotificationResponse, error) {
	// TODO: 实现删除逻辑（需要在 repository 中添加 Delete 方法）
	return &notifyv1.DeleteNotificationResponse{Success: true}, nil
}

// GetUnreadCount 获取未读数量
func (a *NotificationAdapter) GetUnreadCount(ctx context.Context, req *notifyv1.GetUnreadCountRequest) (*notifyv1.UnreadCountResponse, error) {
	// 获取当前用户ID
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("user_id not found in context")
	}

	count, err := a.notifyService.CountUnread(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count unread failed: %w", err)
	}

	return &notifyv1.UnreadCountResponse{
		Count: int32(count),
	}, nil
}

// toProtoNotification 转换为 Protobuf 通知响应
func toProtoNotification(notif *dto.NotificationResponse) *notifyv1.NotificationResponse {
	resp := &notifyv1.NotificationResponse{
		Id:        notif.ID.String(),
		TenantId:  notif.TenantID.String(),
		UserId:    notif.RecipientID.String(),
		Type:      string(notif.Type),
		Title:     notif.Title,
		Content:   notif.Content,
		Priority:  string(notif.Priority),
		IsRead:    notif.ReadAt != nil,
		CreatedAt: notif.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if notif.ReadAt != nil {
		resp.ReadAt = notif.ReadAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return resp
}
