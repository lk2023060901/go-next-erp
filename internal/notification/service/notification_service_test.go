package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lk2023060901/go-next-erp/internal/notification/dto"
	"github.com/lk2023060901/go-next-erp/internal/notification/model"
)

// ===== Mock Repository =====

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) Update(ctx context.Context, notification *model.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Notification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Notification), args.Error(1)
}

func (m *MockNotificationRepository) ListByRecipient(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]*model.Notification, error) {
	args := m.Called(ctx, recipientID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Notification), args.Error(1)
}

func (m *MockNotificationRepository) ListUnread(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]*model.Notification, error) {
	args := m.Called(ctx, recipientID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Notification), args.Error(1)
}

func (m *MockNotificationRepository) CountUnread(ctx context.Context, recipientID uuid.UUID) (int, error) {
	args := m.Called(ctx, recipientID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockNotificationRepository) MarkAsRead(ctx context.Context, ids []uuid.UUID, readAt time.Time) error {
	args := m.Called(ctx, ids, readAt)
	return args.Error(0)
}

func (m *MockNotificationRepository) MarkAllAsRead(ctx context.Context, recipientID uuid.UUID, readAt time.Time) error {
	args := m.Called(ctx, recipientID, readAt)
	return args.Error(0)
}

func (m *MockNotificationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.NotificationStatus, sentAt *time.Time, errorMsg *string) error {
	args := m.Called(ctx, id, status, sentAt, errorMsg)
	return args.Error(0)
}

// ===== Unit Tests =====

// TestSendNotification 测试发送通知
func TestSendNotification(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	t.Run("Send in-app notification successfully", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		service := &notificationService{
			repo: mockRepo,
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Notification")).Return(nil)
		// 异步发送时会调用 Update
		mockRepo.On("Update", ctx, mock.AnythingOfType("*model.Notification")).Return(nil).Maybe()

		req := &dto.SendNotificationRequest{
			Type:        "approval",
			Channel:     "in_app",
			RecipientID: uuid.New().String(),
			Title:       "Test Notification",
			Content:     "This is a test",
		}

		notification, err := service.SendNotification(ctx, tenantID, req)

		assert.NoError(t, err)
		assert.NotNil(t, notification)
		assert.Equal(t, string(model.NotificationTypeApproval), string(notification.Type))
		assert.Equal(t, string(model.NotificationChannelInApp), string(notification.Channel))
		assert.Equal(t, "Test Notification", notification.Title)

		// 等待异步goroutine完成
		time.Sleep(10 * time.Millisecond)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid recipient ID", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		service := &notificationService{
			repo: mockRepo,
		}

		req := &dto.SendNotificationRequest{
			Type:        "approval",
			Channel:     "in_app",
			RecipientID: "invalid-uuid",
			Title:       "Test Notification",
			Content:     "This is a test",
		}

		_, err := service.SendNotification(ctx, tenantID, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid recipient_id")
	})
}

// TestListMyNotifications 测试查询我的通知
func TestListMyNotifications(t *testing.T) {
	ctx := context.Background()
	recipientID := uuid.New()

	t.Run("List notifications successfully", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		service := &notificationService{
			repo: mockRepo,
		}

		now := time.Now()
		readAt := now.Add(-1 * time.Hour)

		expectedNotifications := []*model.Notification{
			{
				ID:          uuid.New(),
				Type:        "approval",
				Channel:     model.NotificationChannelInApp,
				RecipientID: recipientID,
				Title:       "Notification 1",
				Content:     "Content 1",
				Status:      model.NotificationStatusSent,
				ReadAt:      nil,
				CreatedAt:   now,
			},
			{
				ID:          uuid.New(),
				Type:        "system",
				Channel:     model.NotificationChannelInApp,
				RecipientID: recipientID,
				Title:       "Notification 2",
				Content:     "Content 2",
				Status:      model.NotificationStatusRead,
				ReadAt:      &readAt,
				CreatedAt:   now,
			},
		}

		mockRepo.On("ListByRecipient", ctx, recipientID, 10, 0).Return(expectedNotifications, nil)

		notifications, err := service.ListMyNotifications(ctx, recipientID, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, notifications, 2)
		assert.Equal(t, "Notification 1", notifications[0].Title)
		assert.Nil(t, notifications[0].ReadAt)
		assert.Equal(t, "Notification 2", notifications[1].Title)
		assert.NotNil(t, notifications[1].ReadAt)

		mockRepo.AssertExpectations(t)
	})
}

// TestCountUnread 测试统计未读数量
func TestCountUnread(t *testing.T) {
	ctx := context.Background()
	recipientID := uuid.New()

	t.Run("Count unread successfully", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		service := &notificationService{
			repo: mockRepo,
		}

		mockRepo.On("CountUnread", ctx, recipientID).Return(5, nil)

		count, err := service.CountUnread(ctx, recipientID)

		assert.NoError(t, err)
		assert.Equal(t, 5, count)

		mockRepo.AssertExpectations(t)
	})
}

// TestMarkAsRead 测试标记为已读
func TestMarkAsRead(t *testing.T) {
	ctx := context.Background()
	recipientID := uuid.New()

	t.Run("Mark as read successfully", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		service := &notificationService{
			repo: mockRepo,
		}

		notificationID := uuid.New()
		notification := &model.Notification{
			ID:          notificationID,
			RecipientID: recipientID,
			Status:      model.NotificationStatusSent,
		}

		mockRepo.On("FindByID", ctx, notificationID).Return(notification, nil)
		mockRepo.On("MarkAsRead", ctx, []uuid.UUID{notificationID}, mock.AnythingOfType("time.Time")).Return(nil)

		err := service.MarkAsRead(ctx, recipientID, []uuid.UUID{notificationID})

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Unauthorized - notification belongs to different user", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		service := &notificationService{
			repo: mockRepo,
		}

		notificationID := uuid.New()
		differentRecipient := uuid.New()
		notification := &model.Notification{
			ID:          notificationID,
			RecipientID: differentRecipient, // 不同的接收人
			Status:      model.NotificationStatusSent,
		}

		mockRepo.On("FindByID", ctx, notificationID).Return(notification, nil)

		err := service.MarkAsRead(ctx, recipientID, []uuid.UUID{notificationID})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not belong")

		mockRepo.AssertExpectations(t)
	})
}

// TestMarkAllAsRead 测试全部标记为已读
func TestMarkAllAsRead(t *testing.T) {
	ctx := context.Background()
	recipientID := uuid.New()

	t.Run("Mark all as read successfully", func(t *testing.T) {
		mockRepo := new(MockNotificationRepository)
		service := &notificationService{
			repo: mockRepo,
		}

		mockRepo.On("MarkAllAsRead", ctx, recipientID, mock.AnythingOfType("time.Time")).Return(nil)

		err := service.MarkAllAsRead(ctx, recipientID)

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})
}

// ===== Benchmark Tests =====

// BenchmarkSendInAppNotification 基准测试：发送站内消息
func BenchmarkSendInAppNotification(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()
	recipientID := uuid.New()

	mockRepo := new(MockNotificationRepository)
	service := &notificationService{
		repo: mockRepo,
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Notification")).Return(nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*model.Notification")).Return(nil).Maybe()

	req := &dto.SendNotificationRequest{
		Type:        "approval",
		Channel:     "in_app",
		RecipientID: recipientID.String(),
		Title:       "Test Notification",
		Content:     "This is a test notification",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.SendNotification(ctx, tenantID, req)
	}
}

// BenchmarkListNotifications 基准测试：查询通知列表
func BenchmarkListNotifications(b *testing.B) {
	ctx := context.Background()
	recipientID := uuid.New()

	mockRepo := new(MockNotificationRepository)
	service := &notificationService{
		repo: mockRepo,
	}

	notifications := make([]*model.Notification, 10)
	for i := 0; i < 10; i++ {
		notifications[i] = &model.Notification{
			ID:          uuid.New(),
			Type:        "approval",
			Channel:     model.NotificationChannelInApp,
			RecipientID: recipientID,
			Title:       "Test Notification",
			Content:     "Test Content",
			Status:      model.NotificationStatusSent,
			ReadAt:      nil,
			CreatedAt:   time.Now(),
		}
	}

	mockRepo.On("ListByRecipient", ctx, recipientID, 10, 0).Return(notifications, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ListMyNotifications(ctx, recipientID, 10, 0)
	}
}

// BenchmarkCountUnread 基准测试：统计未读数量
func BenchmarkCountUnread(b *testing.B) {
	ctx := context.Background()
	recipientID := uuid.New()

	mockRepo := new(MockNotificationRepository)
	service := &notificationService{
		repo: mockRepo,
	}

	mockRepo.On("CountUnread", ctx, recipientID).Return(5, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CountUnread(ctx, recipientID)
	}
}
