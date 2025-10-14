package sse

import "github.com/google/wire"

// ProviderSet SSE 模块的 Wire Provider Set
var ProviderSet = wire.NewSet(
	ProvideBroker,
	ProvideHandler,
)

// ProvideBroker 提供 SSE Broker
func ProvideBroker() *Broker {
	return NewBroker(DefaultBrokerConfig())
}

// ProvideHandler 提供 SSE Handler（使用默认认证和主题解析）
func ProvideHandler(broker *Broker) *Handler {
	return NewHandler(broker, DefaultAuthenticator, DefaultTopicResolver)
}
