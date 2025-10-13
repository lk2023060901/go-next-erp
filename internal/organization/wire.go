package organization

import (
	"github.com/google/wire"
	"github.com/lk2023060901/go-next-erp/internal/organization/repository"
	"github.com/lk2023060901/go-next-erp/internal/organization/service"
)

// ProviderSet organization 模块的 Wire Provider Set
var ProviderSet = wire.NewSet(
	repository.NewOrganizationRepository,
	repository.NewEmployeeRepository,
	repository.NewPositionRepository,
	repository.NewClosureRepository,
	repository.NewOrganizationTypeRepository,
	service.NewOrganizationService,
	service.NewEmployeeService,
)
