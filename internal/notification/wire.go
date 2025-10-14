package notification

import (
	"github.com/google/wire"
	"github.com/lk2023060901/go-next-erp/internal/notification/repository"
	"github.com/lk2023060901/go-next-erp/internal/notification/service"
	"github.com/lk2023060901/go-next-erp/internal/notification/websocket"
)

// ProvideEmailConfig 提供邮件配置（临时返回 nil，后续从配置文件读取）
func ProvideEmailConfig() *service.EmailConfig {
	// TODO: 从配置文件读取邮件配置
	// 暂时返回 nil，表示不启用邮件发送功能
	return nil
}

// ProviderSet notification 模块的 Wire Provider Set
var ProviderSet = wire.NewSet(
	repository.NewNotificationRepository,
	ProvideEmailConfig,
	service.NewNotificationService,
	websocket.ProviderSet, // WebSocket 支持
)
