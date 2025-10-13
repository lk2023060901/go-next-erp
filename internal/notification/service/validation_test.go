package service

import (
	"testing"

	"github.com/lk2023060901/go-next-erp/internal/notification/model"
	"github.com/stretchr/testify/assert"
)

// TestIsValidNotificationType tests notification type validation
func TestIsValidNotificationType(t *testing.T) {
	tests := []struct {
		name     string
		notifType model.NotificationType
		want     bool
	}{
		{
			name:     "valid - system",
			notifType: model.NotificationTypeSystem,
			want:     true,
		},
		{
			name:     "valid - approval",
			notifType: model.NotificationTypeApproval,
			want:     true,
		},
		{
			name:     "valid - task",
			notifType: model.NotificationTypeTask,
			want:     true,
		},
		{
			name:     "valid - message",
			notifType: model.NotificationTypeMessage,
			want:     true,
		},
		{
			name:     "valid - alert",
			notifType: model.NotificationTypeAlert,
			want:     true,
		},
		{
			name:     "invalid - unknown type",
			notifType: model.NotificationType("unknown"),
			want:     false,
		},
		{
			name:     "invalid - empty string",
			notifType: model.NotificationType(""),
			want:     false,
		},
		{
			name:     "invalid - random string",
			notifType: model.NotificationType("random_type"),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidNotificationType(tt.notifType)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIsValidNotificationChannel tests notification channel validation
func TestIsValidNotificationChannel(t *testing.T) {
	tests := []struct {
		name    string
		channel model.NotificationChannel
		want    bool
	}{
		{
			name:    "valid - in_app",
			channel: model.NotificationChannelInApp,
			want:    true,
		},
		{
			name:    "valid - email",
			channel: model.NotificationChannelEmail,
			want:    true,
		},
		{
			name:    "valid - sms",
			channel: model.NotificationChannelSMS,
			want:    true,
		},
		{
			name:    "valid - webhook",
			channel: model.NotificationChannelWebhook,
			want:    true,
		},
		{
			name:    "invalid - unknown channel",
			channel: model.NotificationChannel("unknown"),
			want:    false,
		},
		{
			name:    "invalid - empty string",
			channel: model.NotificationChannel(""),
			want:    false,
		},
		{
			name:    "invalid - push notification",
			channel: model.NotificationChannel("push"),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidNotificationChannel(tt.channel)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIsValidPriority tests notification priority validation
func TestIsValidPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority model.NotificationPriority
		want     bool
	}{
		{
			name:     "valid - low",
			priority: model.NotificationPriorityLow,
			want:     true,
		},
		{
			name:     "valid - normal",
			priority: model.NotificationPriorityNormal,
			want:     true,
		},
		{
			name:     "valid - high",
			priority: model.NotificationPriorityHigh,
			want:     true,
		},
		{
			name:     "valid - urgent",
			priority: model.NotificationPriorityUrgent,
			want:     true,
		},
		{
			name:     "invalid - unknown priority",
			priority: model.NotificationPriority("unknown"),
			want:     false,
		},
		{
			name:     "invalid - empty string",
			priority: model.NotificationPriority(""),
			want:     false,
		},
		{
			name:     "invalid - critical",
			priority: model.NotificationPriority("critical"),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidPriority(tt.priority)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestNotificationServiceErrors tests error definitions
func TestNotificationServiceErrors(t *testing.T) {
	t.Run("error definitions exist", func(t *testing.T) {
		assert.NotNil(t, ErrNotificationNotFound)
		assert.NotNil(t, ErrInvalidPriority)
		assert.NotNil(t, ErrInvalidType)
		assert.NotNil(t, ErrInvalidChannel)
	})

	t.Run("error messages are descriptive", func(t *testing.T) {
		assert.Contains(t, ErrNotificationNotFound.Error(), "notification")
		assert.Contains(t, ErrInvalidPriority.Error(), "priority")
		assert.Contains(t, ErrInvalidType.Error(), "type")
		assert.Contains(t, ErrInvalidChannel.Error(), "channel")
	})
}

// TestNotificationConstants tests model constants
func TestNotificationConstants(t *testing.T) {
	t.Run("notification types are defined", func(t *testing.T) {
		assert.Equal(t, model.NotificationType("system"), model.NotificationTypeSystem)
		assert.Equal(t, model.NotificationType("approval"), model.NotificationTypeApproval)
		assert.Equal(t, model.NotificationType("task"), model.NotificationTypeTask)
		assert.Equal(t, model.NotificationType("message"), model.NotificationTypeMessage)
		assert.Equal(t, model.NotificationType("alert"), model.NotificationTypeAlert)
	})

	t.Run("notification channels are defined", func(t *testing.T) {
		assert.Equal(t, model.NotificationChannel("in_app"), model.NotificationChannelInApp)
		assert.Equal(t, model.NotificationChannel("email"), model.NotificationChannelEmail)
		assert.Equal(t, model.NotificationChannel("sms"), model.NotificationChannelSMS)
		assert.Equal(t, model.NotificationChannel("webhook"), model.NotificationChannelWebhook)
	})

	t.Run("notification priorities are defined", func(t *testing.T) {
		assert.Equal(t, model.NotificationPriority("low"), model.NotificationPriorityLow)
		assert.Equal(t, model.NotificationPriority("normal"), model.NotificationPriorityNormal)
		assert.Equal(t, model.NotificationPriority("high"), model.NotificationPriorityHigh)
		assert.Equal(t, model.NotificationPriority("urgent"), model.NotificationPriorityUrgent)
	})

	t.Run("notification statuses are defined", func(t *testing.T) {
		assert.Equal(t, model.NotificationStatus("pending"), model.NotificationStatusPending)
		assert.Equal(t, model.NotificationStatus("sent"), model.NotificationStatusSent)
		assert.Equal(t, model.NotificationStatus("delivered"), model.NotificationStatusDelivered)
		assert.Equal(t, model.NotificationStatus("read"), model.NotificationStatusRead)
		assert.Equal(t, model.NotificationStatus("failed"), model.NotificationStatusFailed)
	})
}
