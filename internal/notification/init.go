package notification

import (
	"github.com/lk2023060901/go-next-erp/internal/notification/service"
	"github.com/lk2023060901/go-next-erp/internal/notification/websocket"
)

// InitNotificationWebSocket 初始化通知 WebSocket 支持
func InitNotificationWebSocket(notifService service.NotificationService, wsHandler *websocket.Handler) {
	notifService.SetPushHandler(wsHandler)
}
