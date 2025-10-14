package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	notifyv1 "github.com/lk2023060901/go-next-erp/api/notification/v1"
	"github.com/lk2023060901/go-next-erp/internal/notification/dto"
	"github.com/lk2023060901/go-next-erp/internal/notification/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNotificationService mocks the notification service
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendNotification(ctx context.Context, tenantID uuid.UUID, req *dto.SendNotificationRequest) (*dto.NotificationResponse, error) {
	args := m.Called(ctx, tenantID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.NotificationResponse), args.Error(1)
}

func (m *MockNotificationService) GetNotification(ctx context.Context, id uuid.UUID) (*dto.NotificationResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.NotificationResponse), args.Error(1)
}

func (m *MockNotificationService) ListMyNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*dto.NotificationResponse, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.NotificationResponse), args.Error(1)
}

func (m *MockNotificationService) ListUnreadNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*dto.NotificationResponse, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.NotificationResponse), args.Error(1)
}

func (m *MockNotificationService) MarkAsRead(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error {
	args := m.Called(ctx, userID, ids)
	return args.Error(0)
}

func (m *MockNotificationService) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockNotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockNotificationService) SetPushHandler(handler service.PushHandler) {
	m.Called(handler)
}

// TestNotificationAdapter_SendNotification tests sending notifications
func TestNotificationAdapter_SendNotification(t *testing.T) {
	t.Run("SendNotification successfully", func(t *testing.T) {
		mockService := new(MockNotificationService)
		adapter := NewNotificationAdapter(mockService)

		tenantID := uuid.New()
		userID := uuid.New()
		notifID := uuid.New()

		// Create context with tenant ID
		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)

		expectedNotif := &dto.NotificationResponse{
			ID:          notifID,
			TenantID:    tenantID,
			RecipientID: userID,
			Type:        "system",
			Title:       "Test Notification",
			Content:     "This is a test notification",
			Priority:    "normal",
			CreatedAt:   time.Now(),
		}

		mockService.On("SendNotification", mock.Anything, tenantID, mock.AnythingOfType("*dto.SendNotificationRequest")).
			Return(expectedNotif, nil).Once()

		req := &notifyv1.SendNotificationRequest{
			UserId:   userID.String(),
			Type:     "system",
			Title:    "Test Notification",
			Content:  "This is a test notification",
			Priority: "normal",
		}

		resp, err := adapter.SendNotification(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, notifID.String(), resp.Id)
		assert.Equal(t, "Test Notification", resp.Title)
		assert.Equal(t, "This is a test notification", resp.Content)
		mockService.AssertExpectations(t)
	})

	t.Run("SendNotification without tenant ID", func(t *testing.T) {
		mockService := new(MockNotificationService)
		adapter := NewNotificationAdapter(mockService)

		req := &notifyv1.SendNotificationRequest{
			UserId:  uuid.New().String(),
			Type:    "system",
			Title:   "Test",
			Content: "Test",
		}

		resp, err := adapter.SendNotification(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "tenant_id not found")
	})
}

// TestNotificationAdapter_GetNotification tests getting a notification
func TestNotificationAdapter_GetNotification(t *testing.T) {
	t.Run("GetNotification successfully", func(t *testing.T) {
		mockService := new(MockNotificationService)
		adapter := NewNotificationAdapter(mockService)

		notifID := uuid.New()
		tenantID := uuid.New()
		userID := uuid.New()

		expectedNotif := &dto.NotificationResponse{
			ID:          notifID,
			TenantID:    tenantID,
			RecipientID: userID,
			Type:        "system",
			Title:       "Test Notification",
			Content:     "Test Content",
			Priority:    "normal",
			CreatedAt:   time.Now(),
		}

		mockService.On("GetNotification", mock.Anything, notifID).
			Return(expectedNotif, nil).Once()

		req := &notifyv1.GetNotificationRequest{
			Id: notifID.String(),
		}

		resp, err := adapter.GetNotification(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, notifID.String(), resp.Id)
		assert.Equal(t, "Test Notification", resp.Title)
		mockService.AssertExpectations(t)
	})

	t.Run("GetNotification with invalid ID", func(t *testing.T) {
		mockService := new(MockNotificationService)
		adapter := NewNotificationAdapter(mockService)

		req := &notifyv1.GetNotificationRequest{
			Id: "invalid-uuid",
		}

		resp, err := adapter.GetNotification(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid id")
	})
}

// TestNotificationAdapter_ListNotifications tests listing notifications
func TestNotificationAdapter_ListNotifications(t *testing.T) {
	t.Run("ListNotifications successfully", func(t *testing.T) {
		mockService := new(MockNotificationService)
		adapter := NewNotificationAdapter(mockService)

		userID := uuid.New()
		tenantID := uuid.New()

		// Create context with user ID
		ctx := context.WithValue(context.Background(), "user_id", userID)

		expectedNotifs := []*dto.NotificationResponse{
			{
				ID:          uuid.New(),
				TenantID:    tenantID,
				RecipientID: userID,
				Type:        "system",
				Title:       "Notification 1",
				Content:     "Content 1",
				Priority:    "normal",
				CreatedAt:   time.Now(),
			},
			{
				ID:          uuid.New(),
				TenantID:    tenantID,
				RecipientID: userID,
				Type:        "approval",
				Title:       "Notification 2",
				Content:     "Content 2",
				Priority:    "high",
				CreatedAt:   time.Now(),
			},
		}

		mockService.On("ListMyNotifications", mock.Anything, userID, 20, 0).
			Return(expectedNotifs, nil).Once()

		req := &notifyv1.ListNotificationsRequest{
			Page:     1,
			PageSize: 20,
		}

		resp, err := adapter.ListNotifications(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Items, 2)
		assert.Equal(t, int32(2), resp.Total)
		mockService.AssertExpectations(t)
	})

	t.Run("ListNotifications only unread", func(t *testing.T) {
		mockService := new(MockNotificationService)
		adapter := NewNotificationAdapter(mockService)

		userID := uuid.New()

		ctx := context.WithValue(context.Background(), "user_id", userID)

		expectedNotifs := []*dto.NotificationResponse{
			{
				ID:          uuid.New(),
				TenantID:    uuid.New(),
				RecipientID: userID,
				Type:        "system",
				Title:       "Unread Notification",
				Content:     "Content",
				Priority:    "normal",
				CreatedAt:   time.Now(),
			},
		}

		mockService.On("ListUnreadNotifications", mock.Anything, userID, 20, 0).
			Return(expectedNotifs, nil).Once()

		req := &notifyv1.ListNotificationsRequest{
			Page:       1,
			PageSize:   20,
			OnlyUnread: true,
		}

		resp, err := adapter.ListNotifications(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Items, 1)
		mockService.AssertExpectations(t)
	})

	t.Run("ListNotifications without user ID", func(t *testing.T) {
		mockService := new(MockNotificationService)
		adapter := NewNotificationAdapter(mockService)

		req := &notifyv1.ListNotificationsRequest{
			Page:     1,
			PageSize: 20,
		}

		resp, err := adapter.ListNotifications(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "user_id not found")
	})
}

// TestNotificationAdapter_MarkAsRead tests marking notifications as read
func TestNotificationAdapter_MarkAsRead(t *testing.T) {
	t.Run("MarkAsRead successfully", func(t *testing.T) {
		mockService := new(MockNotificationService)
		adapter := NewNotificationAdapter(mockService)

		userID := uuid.New()
		notifID := uuid.New()
		tenantID := uuid.New()

		ctx := context.WithValue(context.Background(), "user_id", userID)

		mockService.On("MarkAsRead", mock.Anything, userID, mock.MatchedBy(func(ids []uuid.UUID) bool {
			return len(ids) == 1 && ids[0] == notifID
		})).Return(nil).Once()

		readAt := time.Now()
		updatedNotif := &dto.NotificationResponse{
			ID:          notifID,
			TenantID:    tenantID,
			RecipientID: userID,
			Type:        "system",
			Title:       "Test",
			Content:     "Test",
			Priority:    "normal",
			ReadAt:      &readAt,
			CreatedAt:   time.Now(),
		}

		mockService.On("GetNotification", mock.Anything, notifID).
			Return(updatedNotif, nil).Once()

		req := &notifyv1.MarkAsReadRequest{
			Id: notifID.String(),
		}

		resp, err := adapter.MarkAsRead(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, notifID.String(), resp.Id)
		assert.True(t, resp.IsRead)
		mockService.AssertExpectations(t)
	})

	t.Run("MarkAsRead without user ID", func(t *testing.T) {
		mockService := new(MockNotificationService)
		adapter := NewNotificationAdapter(mockService)

		req := &notifyv1.MarkAsReadRequest{
			Id: uuid.New().String(),
		}

		resp, err := adapter.MarkAsRead(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "user_id not found")
	})
}

// TestNotificationAdapter_BatchMarkAsRead tests batch marking as read
func TestNotificationAdapter_BatchMarkAsRead(t *testing.T) {
	t.Run("BatchMarkAsRead successfully", func(t *testing.T) {
		mockService := new(MockNotificationService)
		adapter := NewNotificationAdapter(mockService)

		userID := uuid.New()
		notifID1 := uuid.New()
		notifID2 := uuid.New()

		ctx := context.WithValue(context.Background(), "user_id", userID)

		mockService.On("MarkAsRead", mock.Anything, userID, mock.MatchedBy(func(ids []uuid.UUID) bool {
			return len(ids) == 2
		})).Return(nil).Once()

		req := &notifyv1.BatchMarkAsReadRequest{
			Ids: []string{notifID1.String(), notifID2.String()},
		}

		resp, err := adapter.BatchMarkAsRead(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(2), resp.Count)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})

	t.Run("BatchMarkAsRead with invalid ID", func(t *testing.T) {
		mockService := new(MockNotificationService)
		adapter := NewNotificationAdapter(mockService)

		userID := uuid.New()
		ctx := context.WithValue(context.Background(), "user_id", userID)

		req := &notifyv1.BatchMarkAsReadRequest{
			Ids: []string{"invalid-uuid"},
		}

		resp, err := adapter.BatchMarkAsRead(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid id")
	})
}

// TestNotificationAdapter_GetUnreadCount tests getting unread count
func TestNotificationAdapter_GetUnreadCount(t *testing.T) {
	t.Run("GetUnreadCount successfully", func(t *testing.T) {
		mockService := new(MockNotificationService)
		adapter := NewNotificationAdapter(mockService)

		userID := uuid.New()
		ctx := context.WithValue(context.Background(), "user_id", userID)

		mockService.On("CountUnread", mock.Anything, userID).
			Return(5, nil).Once()

		req := &notifyv1.GetUnreadCountRequest{}

		resp, err := adapter.GetUnreadCount(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(5), resp.Count)
		mockService.AssertExpectations(t)
	})

	t.Run("GetUnreadCount without user ID", func(t *testing.T) {
		mockService := new(MockNotificationService)
		adapter := NewNotificationAdapter(mockService)

		req := &notifyv1.GetUnreadCountRequest{}

		resp, err := adapter.GetUnreadCount(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "user_id not found")
	})
}

// TestNotificationAdapter_DeleteNotification tests deleting notifications
func TestNotificationAdapter_DeleteNotification(t *testing.T) {
	t.Run("DeleteNotification successfully", func(t *testing.T) {
		mockService := new(MockNotificationService)
		adapter := NewNotificationAdapter(mockService)

		req := &notifyv1.DeleteNotificationRequest{
			Id: uuid.New().String(),
		}

		resp, err := adapter.DeleteNotification(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)
	})
}
