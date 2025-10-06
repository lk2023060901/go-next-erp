package adapter

import "github.com/google/wire"

// ProviderSet 适配器层的 Wire Provider Set
var ProviderSet = wire.NewSet(
	NewAuthAdapter,
	NewUserAdapter,
	NewRoleAdapter,
)
