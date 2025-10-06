package server

import "github.com/google/wire"

// ProviderSet 服务器层的 Wire Provider Set
var ProviderSet = wire.NewSet(
	NewHTTPServer,
	NewGRPCServer,
)
