package websocket

import (
	"github.com/google/wire"
)

// ProviderSet WebSocket 模块的 Wire Provider Set
var ProviderSet = wire.NewSet(
	NewHub,
	NewHandler,
)
