package form

import (
	"github.com/google/wire"
	"github.com/lk2023060901/go-next-erp/internal/form/repository"
)

// ProviderSet form 模块的 Wire Provider Set
var ProviderSet = wire.NewSet(
	repository.NewFormDefinitionRepository,
	repository.NewFormDataRepository,
)
