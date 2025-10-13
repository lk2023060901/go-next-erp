package approval

import (
	"github.com/google/wire"
	"github.com/lk2023060901/go-next-erp/internal/approval/repository"
	"github.com/lk2023060901/go-next-erp/internal/approval/service"
	"github.com/lk2023060901/go-next-erp/pkg/workflow"
)

// ProviderSet approval 模块的 Wire Provider Set
var ProviderSet = wire.NewSet(
	// Repositories
	repository.NewProcessDefinitionRepository,
	repository.NewProcessInstanceRepository,
	repository.NewApprovalTaskRepository,
	repository.NewProcessHistoryRepository,

	// Services
	ProvideWorkflowEngine,
	service.NewAssigneeResolver,
	service.NewApprovalService,
)

// ProvideWorkflowEngine 提供工作流引擎
func ProvideWorkflowEngine() *workflow.Engine {
	engine, err := workflow.New()
	if err != nil {
		panic(err)
	}
	return engine
}
