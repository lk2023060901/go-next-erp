package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/notification/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// NotificationRepository 通知仓储接口
type NotificationRepository interface {
	Create(ctx context.Context, notification *model.Notification) error
	Update(ctx context.Context, notification *model.Notification) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Notification, error)
	ListByRecipient(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]*model.Notification, error)
	ListUnread(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]*model.Notification, error)
	MarkAsRead(ctx context.Context, ids []uuid.UUID, readAt time.Time) error
	MarkAllAsRead(ctx context.Context, recipientID uuid.UUID, readAt time.Time) error
	CountUnread(ctx context.Context, recipientID uuid.UUID) (int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.NotificationStatus, sentAt *time.Time, errorMsg *string) error
}

type notificationRepo struct {
	db *database.DB
}

// NewNotificationRepository 创建通知仓储
func NewNotificationRepository(db *database.DB) NotificationRepository {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) Create(ctx context.Context, notification *model.Notification) error {
	dataJSON, err := json.Marshal(notification.Data)
	if err != nil {
		return err
	}

	sql := `
		INSERT INTO notifications (
			id, tenant_id, type, channel, recipient_id, recipient_email, recipient_phone,
			title, content, data, priority, status, sent_at, delivered_at, read_at,
			related_type, related_id, error_message, retry_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`

	_, err = r.db.Exec(ctx, sql,
		notification.ID,
		notification.TenantID,
		notification.Type,
		notification.Channel,
		notification.RecipientID,
		notification.RecipientEmail,
		notification.RecipientPhone,
		notification.Title,
		notification.Content,
		dataJSON,
		notification.Priority,
		notification.Status,
		notification.SentAt,
		notification.DeliveredAt,
		notification.ReadAt,
		notification.RelatedType,
		notification.RelatedID,
		notification.ErrorMessage,
		notification.RetryCount,
		notification.CreatedAt,
		notification.UpdatedAt,
	)

	return err
}

func (r *notificationRepo) Update(ctx context.Context, notification *model.Notification) error {
	dataJSON, err := json.Marshal(notification.Data)
	if err != nil {
		return err
	}

	sql := `
		UPDATE notifications
		SET status = $1, data = $2, sent_at = $3, delivered_at = $4, read_at = $5,
		    error_message = $6, retry_count = $7, updated_at = $8
		WHERE id = $9
	`

	_, err = r.db.Exec(ctx, sql,
		notification.Status,
		dataJSON,
		notification.SentAt,
		notification.DeliveredAt,
		notification.ReadAt,
		notification.ErrorMessage,
		notification.RetryCount,
		notification.UpdatedAt,
		notification.ID,
	)

	return err
}

func (r *notificationRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Notification, error) {
	sql := `
		SELECT id, tenant_id, type, channel, recipient_id, recipient_email, recipient_phone,
		       title, content, data, priority, status, sent_at, delivered_at, read_at,
		       related_type, related_id, error_message, retry_count, created_at, updated_at
		FROM notifications
		WHERE id = $1
	`

	var n model.Notification
	var dataJSON []byte

	err := r.db.QueryRow(ctx, sql, id).Scan(
		&n.ID, &n.TenantID, &n.Type, &n.Channel, &n.RecipientID, &n.RecipientEmail, &n.RecipientPhone,
		&n.Title, &n.Content, &dataJSON, &n.Priority, &n.Status, &n.SentAt, &n.DeliveredAt, &n.ReadAt,
		&n.RelatedType, &n.RelatedID, &n.ErrorMessage, &n.RetryCount, &n.CreatedAt, &n.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(dataJSON, &n.Data); err != nil {
		return nil, err
	}

	return &n, nil
}

func (r *notificationRepo) ListByRecipient(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]*model.Notification, error) {
	sql := `
		SELECT id, tenant_id, type, channel, recipient_id, recipient_email, recipient_phone,
		       title, content, data, priority, status, sent_at, delivered_at, read_at,
		       related_type, related_id, error_message, retry_count, created_at, updated_at
		FROM notifications
		WHERE recipient_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, sql, recipientID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*model.Notification
	for rows.Next() {
		var n model.Notification
		var dataJSON []byte

		err := rows.Scan(
			&n.ID, &n.TenantID, &n.Type, &n.Channel, &n.RecipientID, &n.RecipientEmail, &n.RecipientPhone,
			&n.Title, &n.Content, &dataJSON, &n.Priority, &n.Status, &n.SentAt, &n.DeliveredAt, &n.ReadAt,
			&n.RelatedType, &n.RelatedID, &n.ErrorMessage, &n.RetryCount, &n.CreatedAt, &n.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(dataJSON, &n.Data); err != nil {
			return nil, err
		}

		notifications = append(notifications, &n)
	}

	return notifications, rows.Err()
}

func (r *notificationRepo) ListUnread(ctx context.Context, recipientID uuid.UUID, limit, offset int) ([]*model.Notification, error) {
	sql := `
		SELECT id, tenant_id, type, channel, recipient_id, recipient_email, recipient_phone,
		       title, content, data, priority, status, sent_at, delivered_at, read_at,
		       related_type, related_id, error_message, retry_count, created_at, updated_at
		FROM notifications
		WHERE recipient_id = $1 AND read_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, sql, recipientID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*model.Notification
	for rows.Next() {
		var n model.Notification
		var dataJSON []byte

		err := rows.Scan(
			&n.ID, &n.TenantID, &n.Type, &n.Channel, &n.RecipientID, &n.RecipientEmail, &n.RecipientPhone,
			&n.Title, &n.Content, &dataJSON, &n.Priority, &n.Status, &n.SentAt, &n.DeliveredAt, &n.ReadAt,
			&n.RelatedType, &n.RelatedID, &n.ErrorMessage, &n.RetryCount, &n.CreatedAt, &n.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(dataJSON, &n.Data); err != nil {
			return nil, err
		}

		notifications = append(notifications, &n)
	}

	return notifications, rows.Err()
}

func (r *notificationRepo) MarkAsRead(ctx context.Context, ids []uuid.UUID, readAt time.Time) error {
	sql := `
		UPDATE notifications
		SET read_at = $1, status = $2, updated_at = $3
		WHERE id = ANY($4) AND read_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql, readAt, model.NotificationStatusRead, time.Now(), ids)
	return err
}

func (r *notificationRepo) MarkAllAsRead(ctx context.Context, recipientID uuid.UUID, readAt time.Time) error {
	sql := `
		UPDATE notifications
		SET read_at = $1, status = $2, updated_at = $3
		WHERE recipient_id = $4 AND read_at IS NULL
	`

	_, err := r.db.Exec(ctx, sql, readAt, model.NotificationStatusRead, time.Now(), recipientID)
	return err
}

func (r *notificationRepo) CountUnread(ctx context.Context, recipientID uuid.UUID) (int, error) {
	sql := `
		SELECT COUNT(*)
		FROM notifications
		WHERE recipient_id = $1 AND read_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, sql, recipientID).Scan(&count)
	return count, err
}

func (r *notificationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.NotificationStatus, sentAt *time.Time, errorMsg *string) error {
	sql := `
		UPDATE notifications
		SET status = $1, sent_at = $2, error_message = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.Exec(ctx, sql, status, sentAt, errorMsg, time.Now(), id)
	return err
}
