package websocket

import (
	"github.com/lk2023060901/go-next-erp/internal/notification/service"
)

// InitHub 初始化 Hub 并关联到 NotificationService
func InitHub(hub *Hub, handler *Handler, notifService service.NotificationService) {
	// 启动 Hub
	go hub.Run()

	// 将 Handler 设置为 NotificationService 的推送处理器
	notifService.SetPushHandler(handler)
}
