package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/notification/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	ctx := context.Background()

	db, err := database.New(ctx,
		database.WithHost("localhost"),
		database.WithPort(15000),
		database.WithDatabase("erp_test"),
		database.WithUsername("postgres"),
		database.WithPassword("postgres123"),
		database.WithSSLMode("disable"),
	)

	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return nil
	}

	return db
}

// Helper function to create test notification
func createTestNotification(t *testing.T, tenantID, recipientID uuid.UUID) *model.Notification {
	t.Helper()

	return &model.Notification{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Type:        model.NotificationTypeSystem,
		Channel:     model.NotificationChannelInApp,
		RecipientID: recipientID,
		Title:       "测试通知",
		Content:     "这是一条测试通知内容",
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
		Priority:   model.NotificationPriorityNormal,
		Status:     model.NotificationStatusPending,
		RetryCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// Cleanup helper
func cleanupNotifications(t *testing.T, db *database.DB, tenantID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	_, _ = db.Exec(ctx, "DELETE FROM notifications WHERE tenant_id = $1", tenantID)
}

func cleanupNotificationsByRecipient(t *testing.T, db *database.DB, recipientID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	_, _ = db.Exec(ctx, "DELETE FROM notifications WHERE recipient_id = $1", recipientID)
}

// TestNotificationRepository_Create tests notification creation
func TestNotificationRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	recipientID := uuid.New()
	defer cleanupNotifications(t, db, tenantID)

	t.Run("Create successfully", func(t *testing.T) {
		notification := createTestNotification(t, tenantID, recipientID)

		err := repo.Create(ctx, notification)
		assert.NoError(t, err)

		// Verify
		found, err := repo.FindByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, notification.ID, found.ID)
		assert.Equal(t, notification.Title, found.Title)
		assert.Equal(t, notification.Content, found.Content)
		assert.Equal(t, notification.Status, found.Status)
	})

	t.Run("Create with email channel", func(t *testing.T) {
		notification := createTestNotification(t, tenantID, recipientID)
		notification.Channel = model.NotificationChannelEmail
		email := "test@example.com"
		notification.RecipientEmail = &email

		err := repo.Create(ctx, notification)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, model.NotificationChannelEmail, found.Channel)
		assert.NotNil(t, found.RecipientEmail)
		assert.Equal(t, "test@example.com", *found.RecipientEmail)
	})

	t.Run("Create with SMS channel", func(t *testing.T) {
		notification := createTestNotification(t, tenantID, recipientID)
		notification.Channel = model.NotificationChannelSMS
		phone := "13800138000"
		notification.RecipientPhone = &phone

		err := repo.Create(ctx, notification)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, model.NotificationChannelSMS, found.Channel)
		assert.NotNil(t, found.RecipientPhone)
		assert.Equal(t, "13800138000", *found.RecipientPhone)
	})

	t.Run("Create with different priorities", func(t *testing.T) {
		priorities := []model.NotificationPriority{
			model.NotificationPriorityLow,
			model.NotificationPriorityNormal,
			model.NotificationPriorityHigh,
			model.NotificationPriorityUrgent,
		}

		for _, priority := range priorities {
			notification := createTestNotification(t, tenantID, recipientID)
			notification.Priority = priority

			err := repo.Create(ctx, notification)
			assert.NoError(t, err)

			found, err := repo.FindByID(ctx, notification.ID)
			require.NoError(t, err)
			assert.Equal(t, priority, found.Priority)
		}
	})

	t.Run("Create with related entity", func(t *testing.T) {
		notification := createTestNotification(t, tenantID, recipientID)
		relatedType := "approval_task"
		relatedID := uuid.New()
		notification.RelatedType = &relatedType
		notification.RelatedID = &relatedID

		err := repo.Create(ctx, notification)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.NotNil(t, found.RelatedType)
		assert.Equal(t, "approval_task", *found.RelatedType)
		assert.NotNil(t, found.RelatedID)
		assert.Equal(t, relatedID, *found.RelatedID)
	})

	t.Run("Create with complex data", func(t *testing.T) {
		notification := createTestNotification(t, tenantID, recipientID)
		notification.Data = map[string]interface{}{
			"nested": map[string]interface{}{
				"field1": "value1",
				"field2": 456,
			},
			"array": []string{"item1", "item2", "item3"},
		}

		err := repo.Create(ctx, notification)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.NotNil(t, found.Data["nested"])
		assert.NotNil(t, found.Data["array"])
	})
}

// TestNotificationRepository_Update tests notification updates
func TestNotificationRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	recipientID := uuid.New()
	defer cleanupNotifications(t, db, tenantID)

	t.Run("Update status to sent", func(t *testing.T) {
		notification := createTestNotification(t, tenantID, recipientID)
		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		// Update
		notification.Status = model.NotificationStatusSent
		sentAt := time.Now()
		notification.SentAt = &sentAt
		notification.UpdatedAt = time.Now()

		err = repo.Update(ctx, notification)
		assert.NoError(t, err)

		// Verify
		found, err := repo.FindByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, model.NotificationStatusSent, found.Status)
		assert.NotNil(t, found.SentAt)
	})

	t.Run("Update status to failed with error message", func(t *testing.T) {
		notification := createTestNotification(t, tenantID, recipientID)
		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		// Update
		notification.Status = model.NotificationStatusFailed
		errorMsg := "发送失败：网络错误"
		notification.ErrorMessage = &errorMsg
		notification.RetryCount = 3
		notification.UpdatedAt = time.Now()

		err = repo.Update(ctx, notification)
		assert.NoError(t, err)

		// Verify
		found, err := repo.FindByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, model.NotificationStatusFailed, found.Status)
		assert.NotNil(t, found.ErrorMessage)
		assert.Equal(t, "发送失败：网络错误", *found.ErrorMessage)
		assert.Equal(t, 3, found.RetryCount)
	})

	t.Run("Update status to delivered", func(t *testing.T) {
		notification := createTestNotification(t, tenantID, recipientID)
		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		// Update
		notification.Status = model.NotificationStatusDelivered
		deliveredAt := time.Now()
		notification.DeliveredAt = &deliveredAt

		err = repo.Update(ctx, notification)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, model.NotificationStatusDelivered, found.Status)
		assert.NotNil(t, found.DeliveredAt)
	})
}

// TestNotificationRepository_FindByID tests finding notification by ID
func TestNotificationRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	recipientID := uuid.New()
	defer cleanupNotifications(t, db, tenantID)

	t.Run("Find existing notification", func(t *testing.T) {
		notification := createTestNotification(t, tenantID, recipientID)
		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, notification.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, notification.ID, found.ID)
	})

	t.Run("Find non-existent notification", func(t *testing.T) {
		found, err := repo.FindByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}

// TestNotificationRepository_ListByRecipient tests listing notifications by recipient
func TestNotificationRepository_ListByRecipient(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	recipientID := uuid.New()
	defer cleanupNotifications(t, db, tenantID)

	t.Run("List notifications for recipient", func(t *testing.T) {
		// Create multiple notifications
		for i := 0; i < 5; i++ {
			notification := createTestNotification(t, tenantID, recipientID)
			notification.CreatedAt = time.Now().Add(time.Duration(-i) * time.Hour)
			err := repo.Create(ctx, notification)
			require.NoError(t, err)
		}

		// List
		notifications, err := repo.ListByRecipient(ctx, recipientID, 10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(notifications), 5)

		// Verify all belong to recipient
		for _, n := range notifications {
			assert.Equal(t, recipientID, n.RecipientID)
		}

		// Verify ordering (newest first)
		for i := 0; i < len(notifications)-1; i++ {
			assert.True(t, notifications[i].CreatedAt.After(notifications[i+1].CreatedAt) ||
				notifications[i].CreatedAt.Equal(notifications[i+1].CreatedAt))
		}
	})

	t.Run("List with pagination", func(t *testing.T) {
		recipientID := uuid.New()
		tenantID := uuid.New()
		defer cleanupNotifications(t, db, tenantID)

		// Create 10 notifications
		for i := 0; i < 10; i++ {
			notification := createTestNotification(t, tenantID, recipientID)
			err := repo.Create(ctx, notification)
			require.NoError(t, err)
		}

		// First page
		page1, err := repo.ListByRecipient(ctx, recipientID, 3, 0)
		assert.NoError(t, err)
		assert.Len(t, page1, 3)

		// Second page
		page2, err := repo.ListByRecipient(ctx, recipientID, 3, 3)
		assert.NoError(t, err)
		assert.Len(t, page2, 3)

		// No overlap
		for _, n1 := range page1 {
			for _, n2 := range page2 {
				assert.NotEqual(t, n1.ID, n2.ID)
			}
		}
	})

	t.Run("List empty for recipient with no notifications", func(t *testing.T) {
		notifications, err := repo.ListByRecipient(ctx, uuid.New(), 10, 0)
		assert.NoError(t, err)
		assert.Empty(t, notifications)
	})
}

// TestNotificationRepository_ListUnread tests listing unread notifications
func TestNotificationRepository_ListUnread(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	recipientID := uuid.New()
	defer cleanupNotifications(t, db, tenantID)

	t.Run("List only unread notifications", func(t *testing.T) {
		// Create unread notifications
		for i := 0; i < 3; i++ {
			notification := createTestNotification(t, tenantID, recipientID)
			notification.Status = model.NotificationStatusDelivered
			err := repo.Create(ctx, notification)
			require.NoError(t, err)
		}

		// Create read notifications
		for i := 0; i < 2; i++ {
			notification := createTestNotification(t, tenantID, recipientID)
			notification.Status = model.NotificationStatusRead
			readAt := time.Now()
			notification.ReadAt = &readAt
			err := repo.Create(ctx, notification)
			require.NoError(t, err)
		}

		// List unread
		unread, err := repo.ListUnread(ctx, recipientID, 10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(unread), 3)

		// Verify all are unread
		for _, n := range unread {
			assert.Nil(t, n.ReadAt)
		}
	})

	t.Run("List unread with pagination", func(t *testing.T) {
		recipientID := uuid.New()
		tenantID := uuid.New()
		defer cleanupNotifications(t, db, tenantID)

		// Create 5 unread notifications
		for i := 0; i < 5; i++ {
			notification := createTestNotification(t, tenantID, recipientID)
			err := repo.Create(ctx, notification)
			require.NoError(t, err)
		}

		// First page
		page1, err := repo.ListUnread(ctx, recipientID, 2, 0)
		assert.NoError(t, err)
		assert.Len(t, page1, 2)

		// Second page
		page2, err := repo.ListUnread(ctx, recipientID, 2, 2)
		assert.NoError(t, err)
		assert.Len(t, page2, 2)
	})
}

// TestNotificationRepository_MarkAsRead tests marking notifications as read
func TestNotificationRepository_MarkAsRead(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	recipientID := uuid.New()
	defer cleanupNotifications(t, db, tenantID)

	t.Run("Mark single notification as read", func(t *testing.T) {
		notification := createTestNotification(t, tenantID, recipientID)
		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		// Mark as read
		readAt := time.Now()
		err = repo.MarkAsRead(ctx, []uuid.UUID{notification.ID}, readAt)
		assert.NoError(t, err)

		// Verify
		found, err := repo.FindByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.NotNil(t, found.ReadAt)
		assert.Equal(t, model.NotificationStatusRead, found.Status)
	})

	t.Run("Mark multiple notifications as read", func(t *testing.T) {
		recipientID := uuid.New()
		tenantID := uuid.New()
		defer cleanupNotifications(t, db, tenantID)

		// Create notifications
		ids := make([]uuid.UUID, 3)
		for i := 0; i < 3; i++ {
			notification := createTestNotification(t, tenantID, recipientID)
			err := repo.Create(ctx, notification)
			require.NoError(t, err)
			ids[i] = notification.ID
		}

		// Mark all as read
		readAt := time.Now()
		err := repo.MarkAsRead(ctx, ids, readAt)
		assert.NoError(t, err)

		// Verify all are read
		for _, id := range ids {
			found, err := repo.FindByID(ctx, id)
			require.NoError(t, err)
			assert.NotNil(t, found.ReadAt)
			assert.Equal(t, model.NotificationStatusRead, found.Status)
		}
	})

	t.Run("Mark already read notification should not duplicate", func(t *testing.T) {
		notification := createTestNotification(t, tenantID, recipientID)
		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		// First mark
		readAt1 := time.Now()
		err = repo.MarkAsRead(ctx, []uuid.UUID{notification.ID}, readAt1)
		assert.NoError(t, err)

		// Second mark (should not error)
		readAt2 := time.Now()
		err = repo.MarkAsRead(ctx, []uuid.UUID{notification.ID}, readAt2)
		assert.NoError(t, err)
	})
}

// TestNotificationRepository_MarkAllAsRead tests marking all notifications as read
func TestNotificationRepository_MarkAllAsRead(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	recipientID := uuid.New()
	defer cleanupNotifications(t, db, tenantID)

	t.Run("Mark all as read for recipient", func(t *testing.T) {
		// Create unread notifications
		for i := 0; i < 5; i++ {
			notification := createTestNotification(t, tenantID, recipientID)
			err := repo.Create(ctx, notification)
			require.NoError(t, err)
		}

		// Mark all as read
		readAt := time.Now()
		err := repo.MarkAllAsRead(ctx, recipientID, readAt)
		assert.NoError(t, err)

		// Verify all are read
		notifications, err := repo.ListByRecipient(ctx, recipientID, 10, 0)
		assert.NoError(t, err)

		for _, n := range notifications {
			assert.NotNil(t, n.ReadAt)
			assert.Equal(t, model.NotificationStatusRead, n.Status)
		}
	})

	t.Run("Mark all does not affect other recipients", func(t *testing.T) {
		recipient1 := uuid.New()
		recipient2 := uuid.New()
		tenantID := uuid.New()
		defer cleanupNotifications(t, db, tenantID)

		// Create for recipient1
		for i := 0; i < 3; i++ {
			notification := createTestNotification(t, tenantID, recipient1)
			err := repo.Create(ctx, notification)
			require.NoError(t, err)
		}

		// Create for recipient2
		for i := 0; i < 2; i++ {
			notification := createTestNotification(t, tenantID, recipient2)
			err := repo.Create(ctx, notification)
			require.NoError(t, err)
		}

		// Mark all for recipient1
		err := repo.MarkAllAsRead(ctx, recipient1, time.Now())
		assert.NoError(t, err)

		// Verify recipient1 all read
		count1, err := repo.CountUnread(ctx, recipient1)
		assert.NoError(t, err)
		assert.Equal(t, 0, count1)

		// Verify recipient2 still has unread
		count2, err := repo.CountUnread(ctx, recipient2)
		assert.NoError(t, err)
		assert.Equal(t, 2, count2)
	})
}

// TestNotificationRepository_CountUnread tests counting unread notifications
func TestNotificationRepository_CountUnread(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	recipientID := uuid.New()
	cleanupNotificationsByRecipient(t, db, recipientID) // 清理该recipient的残留数据
	defer cleanupNotifications(t, db, tenantID)

	t.Run("Count unread notifications", func(t *testing.T) {
		// Create 5 unread
		for i := 0; i < 5; i++ {
			notification := createTestNotification(t, tenantID, recipientID)
			err := repo.Create(ctx, notification)
			require.NoError(t, err)
		}

		// Create 3 read
		for i := 0; i < 3; i++ {
			notification := createTestNotification(t, tenantID, recipientID)
			readAt := time.Now()
			notification.ReadAt = &readAt
			notification.Status = model.NotificationStatusRead
			err := repo.Create(ctx, notification)
			require.NoError(t, err)
		}

		count, err := repo.CountUnread(ctx, recipientID)
		assert.NoError(t, err)
		assert.Equal(t, 5, count)
	})

	t.Run("Count zero for no unread", func(t *testing.T) {
		count, err := repo.CountUnread(ctx, uuid.New())
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

// TestNotificationRepository_UpdateStatus tests updating notification status
func TestNotificationRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	recipientID := uuid.New()
	defer cleanupNotifications(t, db, tenantID)

	t.Run("Update status to sent", func(t *testing.T) {
		notification := createTestNotification(t, tenantID, recipientID)
		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		sentAt := time.Now()
		err = repo.UpdateStatus(ctx, notification.ID, model.NotificationStatusSent, &sentAt, nil)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, model.NotificationStatusSent, found.Status)
		assert.NotNil(t, found.SentAt)
	})

	t.Run("Update status to failed with error", func(t *testing.T) {
		notification := createTestNotification(t, tenantID, recipientID)
		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		errorMsg := "发送失败"
		err = repo.UpdateStatus(ctx, notification.ID, model.NotificationStatusFailed, nil, &errorMsg)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, model.NotificationStatusFailed, found.Status)
		assert.NotNil(t, found.ErrorMessage)
		assert.Equal(t, "发送失败", *found.ErrorMessage)
	})
}

// ======================== Benchmark Tests ========================

// BenchmarkNotificationRepository_Create benchmarks notification creation
func BenchmarkNotificationRepository_Create(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	recipientID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		notification := createTestNotification(&testing.T{}, tenantID, recipientID)
		_ = repo.Create(ctx, notification)
	}
}

// BenchmarkNotificationRepository_FindByID benchmarks finding by ID
func BenchmarkNotificationRepository_FindByID(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	recipientID := uuid.New()

	notification := createTestNotification(&testing.T{}, tenantID, recipientID)
	_ = repo.Create(ctx, notification)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByID(ctx, notification.ID)
	}
}

// BenchmarkNotificationRepository_ListByRecipient benchmarks listing by recipient
func BenchmarkNotificationRepository_ListByRecipient(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	recipientID := uuid.New()

	// Create 10 notifications
	for i := 0; i < 10; i++ {
		notification := createTestNotification(&testing.T{}, tenantID, recipientID)
		_ = repo.Create(ctx, notification)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.ListByRecipient(ctx, recipientID, 10, 0)
	}
}

// BenchmarkNotificationRepository_CountUnread benchmarks counting unread
func BenchmarkNotificationRepository_CountUnread(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewNotificationRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	recipientID := uuid.New()

	// Create 10 unread notifications
	for i := 0; i < 10; i++ {
		notification := createTestNotification(&testing.T{}, tenantID, recipientID)
		_ = repo.Create(ctx, notification)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.CountUnread(ctx, recipientID)
	}
}
