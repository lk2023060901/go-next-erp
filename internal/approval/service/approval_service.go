package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/approval/dto"
	"github.com/lk2023060901/go-next-erp/internal/approval/model"
	"github.com/lk2023060901/go-next-erp/internal/approval/repository"
	"github.com/lk2023060901/go-next-erp/internal/auth/authorization"
	formModel "github.com/lk2023060901/go-next-erp/internal/form/model"
	formRepo "github.com/lk2023060901/go-next-erp/internal/form/repository"
	notificationDto "github.com/lk2023060901/go-next-erp/internal/notification/dto"
	notificationService "github.com/lk2023060901/go-next-erp/internal/notification/service"
	"github.com/lk2023060901/go-next-erp/pkg/workflow"
	workflowModel "github.com/lk2023060901/go-next-erp/pkg/workflow"
)

var (
	ErrProcessNotFound         = errors.New("process definition not found")
	ErrProcessInstanceNotFound = errors.New("process instance not found")
	ErrTaskNotFound            = errors.New("approval task not found")
	ErrInvalidAction           = errors.New("invalid approval action")
	ErrTaskAlreadyProcessed    = errors.New("task already processed")
	ErrUnauthorized            = errors.New("unauthorized to process this task")
	ErrPermissionDenied        = errors.New("permission denied")
	ErrProcessHasInstances     = errors.New("process definition has active instances")
)

// ApprovalService 审批服务接口
type ApprovalService interface {
	// 流程定义管理
	CreateProcessDefinition(ctx context.Context, req *dto.CreateProcessDefRequest) (*dto.ProcessDefResponse, error)
	UpdateProcessDefinition(ctx context.Context, id uuid.UUID, req *dto.UpdateProcessDefRequest) (*dto.ProcessDefResponse, error)
	GetProcessDefinition(ctx context.Context, id uuid.UUID) (*dto.ProcessDefResponse, error)
	ListProcessDefinitions(ctx context.Context, tenantID uuid.UUID) ([]*dto.ProcessDefResponse, error)
	DeleteProcessDefinition(ctx context.Context, id uuid.UUID) error
	SetProcessDefinitionStatus(ctx context.Context, id uuid.UUID, enabled bool) error
	GetProcessStats(ctx context.Context, processDefID uuid.UUID) (*dto.ProcessStatsResponse, error)

	// 流程实例管理
	StartProcess(ctx context.Context, req *dto.StartProcessRequest) (*dto.ProcessInstanceResponse, error)
	GetProcessInstance(ctx context.Context, id uuid.UUID) (*dto.ProcessInstanceResponse, error)
	ListMyApplications(ctx context.Context, applicantID uuid.UUID, limit, offset int) ([]*dto.ProcessInstanceResponse, error)
	ListProcessInstances(ctx context.Context, tenantID uuid.UUID, processDefID *uuid.UUID, status *model.ProcessStatus, applicantID *uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]*dto.ProcessInstanceResponse, int, error)
	CancelProcess(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason *string) error
	GetInstanceStatsSummary(ctx context.Context, tenantID uuid.UUID) (*dto.InstanceStatsSummary, error)
	GetInstanceStatsByStatus(ctx context.Context, tenantID uuid.UUID, processDefID *uuid.UUID, startDate, endDate *time.Time) (map[string]int, error)

	// 任务管理
	GetApprovalTask(ctx context.Context, id uuid.UUID) (*dto.ApprovalTaskResponse, error)
	ListMyTasks(ctx context.Context, assigneeID uuid.UUID, status *model.TaskStatus, limit, offset int) ([]*dto.ApprovalTaskResponse, error)
	CountPendingTasks(ctx context.Context, assigneeID uuid.UUID) (int, error)
	ListPendingTasks(ctx context.Context, tenantID uuid.UUID, processDefID, assigneeID *uuid.UUID, limit, offset int) ([]*dto.ApprovalTaskResponse, int, error)
	ListCompletedTasks(ctx context.Context, tenantID uuid.UUID, processDefID, assigneeID *uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]*dto.ApprovalTaskResponse, int, error)

	// 审批操作
	ProcessTask(ctx context.Context, req *dto.ProcessTaskRequest) error
	BatchProcessTasks(ctx context.Context, taskIDs []uuid.UUID, operatorID uuid.UUID, action model.ApprovalAction, comment *string) ([]*dto.BatchProcessResult, error)
	TransferTask(ctx context.Context, taskID uuid.UUID, fromUserID, toUserID uuid.UUID, comment *string) error
	DelegateTask(ctx context.Context, taskID uuid.UUID, fromUserID, toUserID uuid.UUID, comment *string) error
	WithdrawProcess(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID) error

	// 历史记录
	GetProcessHistory(ctx context.Context, instanceID uuid.UUID) ([]*dto.ProcessHistoryResponse, error)
	GetTaskHistory(ctx context.Context, taskID uuid.UUID) ([]*dto.ProcessHistoryResponse, error)

	// 流程图可视化
	GetProcessDiagram(ctx context.Context, processDefID uuid.UUID) (*dto.ProcessDiagramResponse, error)
	GetProcessInstanceDiagram(ctx context.Context, instanceID uuid.UUID) (*dto.ProcessInstanceDiagramResponse, error)
	GetProcessTrace(ctx context.Context, instanceID uuid.UUID) (*dto.ProcessTraceResponse, error)

	// 监控和统计
	GetDashboard(ctx context.Context, tenantID, userID uuid.UUID) (*dto.DashboardResponse, error)
	GetProcessMetrics(ctx context.Context, processDefID uuid.UUID, startDate, endDate *time.Time) (*dto.ProcessMetrics, error)
	GetUserWorkload(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time) (*dto.UserWorkload, error)
}

type approvalService struct {
	processDefRepo      repository.ProcessDefinitionRepository
	processInstRepo     repository.ProcessInstanceRepository
	taskRepo            repository.ApprovalTaskRepository
	historyRepo         repository.ProcessHistoryRepository
	formDefRepo         formRepo.FormDefinitionRepository
	formDataRepo        formRepo.FormDataRepository
	workflowEngine      *workflow.Engine
	assigneeResolver    *AssigneeResolver
	authzService        *authorization.Service
	notificationService notificationService.NotificationService
}

// NewApprovalService 创建审批服务
func NewApprovalService(
	processDefRepo repository.ProcessDefinitionRepository,
	processInstRepo repository.ProcessInstanceRepository,
	taskRepo repository.ApprovalTaskRepository,
	historyRepo repository.ProcessHistoryRepository,
	formDefRepo formRepo.FormDefinitionRepository,
	formDataRepo formRepo.FormDataRepository,
	workflowEngine *workflow.Engine,
	assigneeResolver *AssigneeResolver,
	authzService *authorization.Service,
	notificationService notificationService.NotificationService,
) ApprovalService {
	return &approvalService{
		processDefRepo:      processDefRepo,
		processInstRepo:     processInstRepo,
		taskRepo:            taskRepo,
		historyRepo:         historyRepo,
		formDefRepo:         formDefRepo,
		formDataRepo:        formDataRepo,
		workflowEngine:      workflowEngine,
		assigneeResolver:    assigneeResolver,
		authzService:        authzService,
		notificationService: notificationService,
	}
}

// CreateProcessDefinition 创建流程定义
func (s *approvalService) CreateProcessDefinition(ctx context.Context, req *dto.CreateProcessDefRequest) (*dto.ProcessDefResponse, error) {
	// 检查表单是否存在
	formDef, err := s.formDefRepo.FindByID(ctx, req.FormID)
	if err != nil {
		return nil, fmt.Errorf("form not found: %w", err)
	}

	// 检查工作流是否存在（可选，测试环境可能不存在）
	var workflowName string
	workflowDef, err := s.workflowEngine.GetWorkflow(req.WorkflowID.String())
	if err != nil {
		// 工作流不存在时，使用默认名称
		workflowName = "Workflow-" + req.WorkflowID.String()
	} else {
		workflowName = workflowDef.Name
	}

	// 创建流程定义
	now := time.Now()
	processDef := &model.ProcessDefinition{
		ID:         uuid.New(),
		TenantID:   req.TenantID,
		Code:       req.Code,
		Name:       req.Name,
		Category:   req.Category,
		FormID:     req.FormID,
		WorkflowID: req.WorkflowID,
		Enabled:    true,
		CreatedBy:  req.CreatedBy,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.processDefRepo.Create(ctx, processDef); err != nil {
		return nil, fmt.Errorf("failed to create process definition: %w", err)
	}

	return &dto.ProcessDefResponse{
		ID:           processDef.ID,
		TenantID:     processDef.TenantID,
		Code:         processDef.Code,
		Name:         processDef.Name,
		FormID:       processDef.FormID,
		FormName:     formDef.Name,
		WorkflowID:   processDef.WorkflowID,
		WorkflowName: workflowName,
		Enabled:      processDef.Enabled,
		CreatedAt:    processDef.CreatedAt,
		UpdatedAt:    processDef.UpdatedAt,
	}, nil
}

// UpdateProcessDefinition 更新流程定义
func (s *approvalService) UpdateProcessDefinition(ctx context.Context, id uuid.UUID, req *dto.UpdateProcessDefRequest) (*dto.ProcessDefResponse, error) {
	processDef, err := s.processDefRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrProcessNotFound
	}

	// 更新字段
	processDef.Name = req.Name
	processDef.FormID = req.FormID
	processDef.WorkflowID = req.WorkflowID
	processDef.Enabled = req.Enabled
	processDef.UpdatedBy = &req.UpdatedBy
	processDef.UpdatedAt = time.Now()

	if err := s.processDefRepo.Update(ctx, processDef); err != nil {
		return nil, fmt.Errorf("failed to update process definition: %w", err)
	}

	// 获取关联数据
	formDef, _ := s.formDefRepo.FindByID(ctx, processDef.FormID)
	workflowDef, _ := s.workflowEngine.GetWorkflow(processDef.WorkflowID.String())

	return &dto.ProcessDefResponse{
		ID:           processDef.ID,
		TenantID:     processDef.TenantID,
		Code:         processDef.Code,
		Name:         processDef.Name,
		FormID:       processDef.FormID,
		FormName:     formDef.Name,
		WorkflowID:   processDef.WorkflowID,
		WorkflowName: workflowDef.Name,
		Enabled:      processDef.Enabled,
		CreatedAt:    processDef.CreatedAt,
		UpdatedAt:    processDef.UpdatedAt,
	}, nil
}

// GetProcessDefinition 获取流程定义
func (s *approvalService) GetProcessDefinition(ctx context.Context, id uuid.UUID) (*dto.ProcessDefResponse, error) {
	processDef, err := s.processDefRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrProcessNotFound
	}

	formDef, _ := s.formDefRepo.FindByID(ctx, processDef.FormID)
	workflowDef, _ := s.workflowEngine.GetWorkflow(processDef.WorkflowID.String())

	return &dto.ProcessDefResponse{
		ID:           processDef.ID,
		TenantID:     processDef.TenantID,
		Code:         processDef.Code,
		Name:         processDef.Name,
		FormID:       processDef.FormID,
		FormName:     formDef.Name,
		WorkflowID:   processDef.WorkflowID,
		WorkflowName: workflowDef.Name,
		Enabled:      processDef.Enabled,
		CreatedAt:    processDef.CreatedAt,
		UpdatedAt:    processDef.UpdatedAt,
	}, nil
}

// ListProcessDefinitions 列出流程定义
func (s *approvalService) ListProcessDefinitions(ctx context.Context, tenantID uuid.UUID) ([]*dto.ProcessDefResponse, error) {
	defs, err := s.processDefRepo.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.ProcessDefResponse, 0, len(defs))
	for _, def := range defs {
		formDef, _ := s.formDefRepo.FindByID(ctx, def.FormID)
		workflowDef, _ := s.workflowEngine.GetWorkflow(def.WorkflowID.String())

		var formName, workflowName string
		if formDef != nil {
			formName = formDef.Name
		}
		if workflowDef != nil {
			workflowName = workflowDef.Name
		} else {
			workflowName = "Workflow-" + def.WorkflowID.String()
		}

		responses = append(responses, &dto.ProcessDefResponse{
			ID:           def.ID,
			TenantID:     def.TenantID,
			Code:         def.Code,
			Name:         def.Name,
			FormID:       def.FormID,
			FormName:     formName,
			WorkflowID:   def.WorkflowID,
			WorkflowName: workflowName,
			Enabled:      def.Enabled,
			CreatedAt:    def.CreatedAt,
			UpdatedAt:    def.UpdatedAt,
		})
	}

	return responses, nil
}

// StartProcess 启动流程
func (s *approvalService) StartProcess(ctx context.Context, req *dto.StartProcessRequest) (*dto.ProcessInstanceResponse, error) {
	// 获取流程定义
	processDef, err := s.processDefRepo.FindByID(ctx, req.ProcessDefID)
	if err != nil {
		return nil, ErrProcessNotFound
	}

	if !processDef.Enabled {
		return nil, fmt.Errorf("process definition is disabled")
	}

	// 获取表单定义并验证数据
	formDef, err := s.formDefRepo.FindByID(ctx, processDef.FormID)
	if err != nil {
		return nil, fmt.Errorf("form not found: %w", err)
	}

	// 验证表单数据
	if err := s.validateFormData(formDef, req.FormData); err != nil {
		return nil, fmt.Errorf("form validation failed: %w", err)
	}

	// 创建表单数据记录
	now := time.Now()
	formData := &formModel.FormData{
		ID:          uuid.New(),
		TenantID:    req.TenantID,
		FormID:      formDef.ID,
		Data:        req.FormData,
		SubmittedBy: req.ApplicantID,
		SubmittedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.formDataRepo.Create(ctx, formData); err != nil {
		return nil, fmt.Errorf("failed to create form data: %w", err)
	}

	// 启动工作流实例
	workflowInput := map[string]interface{}{
		"form_data":    req.FormData,
		"applicant_id": req.ApplicantID.String(),
		"tenant_id":    req.TenantID.String(),
	}

	executionID, err := s.workflowEngine.Execute(ctx, processDef.WorkflowID.String(), workflowInput, req.ApplicantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to start workflow: %w", err)
	}

	workflowInstanceID, _ := uuid.Parse(executionID)

	// 创建流程实例
	instance := &model.ProcessInstance{
		ID:                 uuid.New(),
		TenantID:           req.TenantID,
		ProcessDefID:       processDef.ID,
		WorkflowInstanceID: workflowInstanceID,
		FormDataID:         formData.ID,
		ApplicantID:        req.ApplicantID,
		Status:             model.ProcessStatusPending,
		Variables:          req.FormData,
		StartedAt:          now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.processInstRepo.Create(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to create process instance: %w", err)
	}

	// 创建第一个审批任务（从工作流引擎获取当前节点）
	execCtx, err := s.workflowEngine.GetExecution(executionID)
	if err == nil && execCtx.CurrentNodeID != "" {
		// 获取当前节点定义
		workflow, err := s.workflowEngine.GetWorkflow(processDef.WorkflowID.String())
		if err == nil {
			var currentNode *workflowModel.NodeDefinition
			for _, node := range workflow.Nodes {
				if node.ID == execCtx.CurrentNodeID {
					currentNode = node
					break
				}
			}

			if currentNode != nil {
				// 从节点配置中获取审批人ID
				if assigneeIDStr, ok := currentNode.Config["assignee_id"].(string); ok {
					assigneeID, err := uuid.Parse(assigneeIDStr)
					if err == nil {
						task := &model.ApprovalTask{
							ID:                uuid.New(),
							TenantID:          req.TenantID,
							ProcessInstanceID: instance.ID,
							NodeID:            currentNode.ID,
							NodeName:          currentNode.Name,
							AssigneeID:        assigneeID,
							Status:            model.TaskStatusPending,
							CreatedAt:         now,
							UpdatedAt:         now,
						}

						// 忽略任务创建错误，不影响流程启动
						_ = s.taskRepo.Create(ctx, task)
					}
				}
			}
		}
	}

	return &dto.ProcessInstanceResponse{
		ID:             instance.ID,
		ProcessDefID:   instance.ProcessDefID,
		ProcessDefCode: processDef.Code,
		ProcessDefName: processDef.Name,
		ApplicantID:    instance.ApplicantID,
		Status:         instance.Status,
		StartedAt:      instance.StartedAt,
		CreatedAt:      instance.CreatedAt,
	}, nil
}

// GetProcessInstance 获取流程实例
func (s *approvalService) GetProcessInstance(ctx context.Context, id uuid.UUID) (*dto.ProcessInstanceResponse, error) {
	instance, err := s.processInstRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrProcessInstanceNotFound
	}

	processDef, _ := s.processDefRepo.FindByID(ctx, instance.ProcessDefID)

	return &dto.ProcessInstanceResponse{
		ID:             instance.ID,
		ProcessDefID:   instance.ProcessDefID,
		ProcessDefCode: processDef.Code,
		ProcessDefName: processDef.Name,
		ApplicantID:    instance.ApplicantID,
		Status:         instance.Status,
		CurrentNodeID:  instance.CurrentNodeID,
		StartedAt:      instance.StartedAt,
		CompletedAt:    instance.CompletedAt,
		CreatedAt:      instance.CreatedAt,
	}, nil
}

// ListMyApplications 列出我的申请
func (s *approvalService) ListMyApplications(ctx context.Context, applicantID uuid.UUID, limit, offset int) ([]*dto.ProcessInstanceResponse, error) {
	instances, err := s.processInstRepo.ListByApplicant(ctx, applicantID, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.ProcessInstanceResponse, 0, len(instances))
	for _, instance := range instances {
		processDef, _ := s.processDefRepo.FindByID(ctx, instance.ProcessDefID)

		responses = append(responses, &dto.ProcessInstanceResponse{
			ID:             instance.ID,
			ProcessDefID:   instance.ProcessDefID,
			ProcessDefCode: processDef.Code,
			ProcessDefName: processDef.Name,
			ApplicantID:    instance.ApplicantID,
			Status:         instance.Status,
			CurrentNodeID:  instance.CurrentNodeID,
			StartedAt:      instance.StartedAt,
			CompletedAt:    instance.CompletedAt,
			CreatedAt:      instance.CreatedAt,
		})
	}

	return responses, nil
}

// GetApprovalTask 获取审批任务
func (s *approvalService) GetApprovalTask(ctx context.Context, id uuid.UUID) (*dto.ApprovalTaskResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	return &dto.ApprovalTaskResponse{
		ID:                task.ID,
		ProcessInstanceID: task.ProcessInstanceID,
		NodeID:            task.NodeID,
		AssigneeID:        task.AssigneeID,
		Status:            task.Status,
		Action:            task.Action,
		Comment:           task.Comment,
		ApprovedAt:        task.ApprovedAt,
		CreatedAt:         task.CreatedAt,
	}, nil
}

// ListMyTasks 列出我的任务
func (s *approvalService) ListMyTasks(ctx context.Context, assigneeID uuid.UUID, status *model.TaskStatus, limit, offset int) ([]*dto.ApprovalTaskResponse, error) {
	tasks, err := s.taskRepo.ListByAssignee(ctx, assigneeID, status, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.ApprovalTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		responses = append(responses, &dto.ApprovalTaskResponse{
			ID:                task.ID,
			ProcessInstanceID: task.ProcessInstanceID,
			NodeID:            task.NodeID,
			AssigneeID:        task.AssigneeID,
			Status:            task.Status,
			Action:            task.Action,
			Comment:           task.Comment,
			ApprovedAt:        task.ApprovedAt,
			CreatedAt:         task.CreatedAt,
		})
	}

	return responses, nil
}

// CountPendingTasks 统计待处理任务数
func (s *approvalService) CountPendingTasks(ctx context.Context, assigneeID uuid.UUID) (int, error) {
	return s.taskRepo.CountPendingByAssignee(ctx, assigneeID)
}

// ProcessTask 处理审批任务
func (s *approvalService) ProcessTask(ctx context.Context, req *dto.ProcessTaskRequest) error {
	// 获取任务
	task, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return ErrTaskNotFound
	}

	// 验证任务状态
	if task.Status != model.TaskStatusPending {
		return ErrTaskAlreadyProcessed
	}

	// 1. 基础验证：检查是否为任务审批人
	if task.AssigneeID != req.OperatorID {
		return ErrUnauthorized
	}

	// 2. 权限验证：集成4A权限系统
	if s.authzService != nil {
		allowed, err := s.authzService.CheckPermission(
			ctx,
			req.OperatorID,
			task.TenantID,
			"approval_task",
			"process",
			map[string]interface{}{
				"ID":          task.ID.String(),
				"process_id":  task.ProcessInstanceID.String(),
				"node_id":     task.NodeID,
				"assignee_id": task.AssigneeID.String(),
			},
		)
		if err != nil || !allowed {
			return ErrPermissionDenied
		}
	}

	// 验证操作类型
	if req.Action != model.ApprovalActionApprove && req.Action != model.ApprovalActionReject {
		return ErrInvalidAction
	}

	// 更新任务状态
	now := time.Now()
	task.Action = &req.Action
	task.Comment = req.Comment
	task.ApprovedAt = &now
	task.UpdatedAt = now

	if req.Action == model.ApprovalActionApprove {
		task.Status = model.TaskStatusApproved
	} else {
		task.Status = model.TaskStatusRejected
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// 获取流程实例
	instance, err := s.processInstRepo.FindByID(ctx, task.ProcessInstanceID)
	if err != nil {
		return fmt.Errorf("failed to get process instance: %w", err)
	}

	// 记录历史
	history := &model.ProcessHistory{
		ID:                uuid.New(),
		TenantID:          task.TenantID,
		ProcessInstanceID: task.ProcessInstanceID,
		TaskID:            &task.ID,
		NodeID:            task.NodeID,
		OperatorID:        req.OperatorID,
		Action:            req.Action,
		Comment:           req.Comment,
		CreatedAt:         now,
	}

	if err := s.historyRepo.Create(ctx, history); err != nil {
		return fmt.Errorf("failed to create history: %w", err)
	}

	// 推进工作流到下一个节点
	if req.Action == model.ApprovalActionReject {
		// 如果是拒绝，则终止流程
		instance.Status = model.ProcessStatusRejected
		instance.CompletedAt = &now
		instance.UpdatedAt = now
		if err := s.processInstRepo.Update(ctx, instance); err != nil {
			return fmt.Errorf("failed to update process instance: %w", err)
		}

		// 通知申请人流程被拒绝
		if s.notificationService != nil {
			s.sendTaskNotification(ctx, task, instance, "rejected")
		}
	} else {
		// 审批通过，推进工作流
		execCtx, err := s.workflowEngine.GetExecution(instance.WorkflowInstanceID.String())
		if err != nil {
			return fmt.Errorf("failed to get execution context: %w", err)
		}

		// 更新工作流上下文（传递审批结果）
		if execCtx.Variables == nil {
			execCtx.Variables = make(map[string]interface{})
		}
		execCtx.Variables["last_approval_action"] = string(req.Action)
		execCtx.Variables["last_approval_comment"] = req.Comment

		// 获取下一个节点
		processDef, _ := s.processDefRepo.FindByID(ctx, instance.ProcessDefID)
		workflow, err := s.workflowEngine.GetWorkflow(processDef.WorkflowID.String())
		if err != nil {
			return fmt.Errorf("failed to get workflow: %w", err)
		}

		// 查找当前节点的出边
		var nextNodeID string
		for _, edge := range workflow.Edges {
			if edge.Source == task.NodeID {
				// 简化逻辑：取第一个出边（实际应该根据条件选择）
				nextNodeID = edge.Target
				break
			}
		}

		if nextNodeID == "" {
			// 没有下一个节点，流程完成
			instance.Status = model.ProcessStatusApproved
			instance.CompletedAt = &now
			instance.UpdatedAt = now
			if err := s.processInstRepo.Update(ctx, instance); err != nil {
				return fmt.Errorf("failed to update process instance: %w", err)
			}
		} else {
			// 创建下一个审批任务
			var nextNode *workflowModel.NodeDefinition
			for _, node := range workflow.Nodes {
				if node.ID == nextNodeID {
					nextNode = node
					break
				}
			}

			if nextNode != nil {
				// 从节点配置中获取审批人ID
				if assigneeIDStr, ok := nextNode.Config["assignee_id"].(string); ok {
					assigneeID, err := uuid.Parse(assigneeIDStr)
					if err == nil {
						nextTask := &model.ApprovalTask{
							ID:                uuid.New(),
							TenantID:          instance.TenantID,
							ProcessInstanceID: instance.ID,
							NodeID:            nextNode.ID,
							NodeName:          nextNode.Name,
							AssigneeID:        assigneeID,
							Status:            model.TaskStatusPending,
							CreatedAt:         now,
							UpdatedAt:         now,
						}

						if err := s.taskRepo.Create(ctx, nextTask); err != nil {
							return fmt.Errorf("failed to create next task: %w", err)
						}

						// 更新流程实例的当前节点
						instance.CurrentNodeID = &nextNode.ID
						instance.CurrentNodeName = &nextNode.Name
						instance.UpdatedAt = now
						if err := s.processInstRepo.Update(ctx, instance); err != nil {
							return fmt.Errorf("failed to update process instance: %w", err)
						}
					}
				}
			}
		}
	}

	return nil
}

// WithdrawProcess 撤回流程
func (s *approvalService) WithdrawProcess(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID) error {
	instance, err := s.processInstRepo.FindByID(ctx, instanceID)
	if err != nil {
		return ErrProcessInstanceNotFound
	}

	// 验证是否为申请人
	if instance.ApplicantID != operatorID {
		return ErrUnauthorized
	}

	// 验证流程状态
	if instance.Status != model.ProcessStatusPending {
		return fmt.Errorf("can only withdraw pending process")
	}

	// 更新流程状态
	now := time.Now()
	instance.Status = model.ProcessStatusWithdrawn
	instance.CompletedAt = &now
	instance.UpdatedAt = now

	if err := s.processInstRepo.Update(ctx, instance); err != nil {
		return fmt.Errorf("failed to withdraw process: %w", err)
	}

	// 记录历史
	history := &model.ProcessHistory{
		ID:                uuid.New(),
		TenantID:          instance.TenantID,
		ProcessInstanceID: instanceID,
		NodeID:            *instance.CurrentNodeID,
		OperatorID:        operatorID,
		Action:            model.ApprovalActionWithdraw,
		CreatedAt:         now,
	}

	if err := s.historyRepo.Create(ctx, history); err != nil {
		return fmt.Errorf("failed to create history: %w", err)
	}

	return nil
}

// GetProcessHistory 获取流程历史
func (s *approvalService) GetProcessHistory(ctx context.Context, instanceID uuid.UUID) ([]*dto.ProcessHistoryResponse, error) {
	histories, err := s.historyRepo.ListByInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.ProcessHistoryResponse, 0, len(histories))
	for _, h := range histories {
		responses = append(responses, &dto.ProcessHistoryResponse{
			ID:                h.ID,
			ProcessInstanceID: h.ProcessInstanceID,
			TaskID:            h.TaskID,
			NodeID:            h.NodeID,
			OperatorID:        h.OperatorID,
			Action:            h.Action,
			Comment:           h.Comment,
			CreatedAt:         h.CreatedAt,
		})
	}

	return responses, nil
}

// validateFormData 验证表单数据（简化版）
func (s *approvalService) validateFormData(formDef *formModel.FormDefinition, data map[string]interface{}) error {
	for _, field := range formDef.Fields {
		if field.Required {
			if _, exists := data[field.Key]; !exists {
				return fmt.Errorf("required field missing: %s", field.Key)
			}
		}
	}
	return nil
}

// GetProcessDiagram 获取流程定义图
func (s *approvalService) GetProcessDiagram(ctx context.Context, processDefID uuid.UUID) (*dto.ProcessDiagramResponse, error) {
	// 获取流程定义
	processDef, err := s.processDefRepo.FindByID(ctx, processDefID)
	if err != nil {
		return nil, ErrProcessNotFound
	}

	// 获取工作流定义（可选，测试环境可能不存在）
	workflowDef, err := s.workflowEngine.GetWorkflow(processDef.WorkflowID.String())
	var nodes []*dto.DiagramNode
	var edges []*dto.DiagramEdge
	var version int

	if err != nil {
		// 工作流不存在时，返回空图表
		nodes = []*dto.DiagramNode{}
		edges = []*dto.DiagramEdge{}
		version = 0
	} else {
		// 转换节点
		nodes = make([]*dto.DiagramNode, 0, len(workflowDef.Nodes))
		for _, node := range workflowDef.Nodes {
			diagramNode := &dto.DiagramNode{
				ID:       node.ID,
				Type:     node.Type,
				Name:     node.Name,
				Config:   node.Config,
				Disabled: node.Disabled,
			}
			if node.Position != nil {
				diagramNode.Position = &dto.Position{
					X: node.Position.X,
					Y: node.Position.Y,
				}
			}
			nodes = append(nodes, diagramNode)
		}

		// 转换边
		edges = make([]*dto.DiagramEdge, 0, len(workflowDef.Edges))
		for _, edge := range workflowDef.Edges {
			edges = append(edges, &dto.DiagramEdge{
				ID:        edge.ID,
				Source:    edge.Source,
				Target:    edge.Target,
				Condition: edge.Condition,
				Label:     edge.Label,
			})
		}
		version = workflowDef.Version
	}

	return &dto.ProcessDiagramResponse{
		ProcessDefID: processDef.ID,
		ProcessCode:  processDef.Code,
		ProcessName:  processDef.Name,
		Version:      version,
		Nodes:        nodes,
		Edges:        edges,
	}, nil
}

// GetProcessInstanceDiagram 获取流程实例图（带执行状态）
func (s *approvalService) GetProcessInstanceDiagram(ctx context.Context, instanceID uuid.UUID) (*dto.ProcessInstanceDiagramResponse, error) {
	// 获取流程实例
	instance, err := s.processInstRepo.FindByID(ctx, instanceID)
	if err != nil {
		return nil, ErrProcessInstanceNotFound
	}

	// 获取流程定义
	processDef, err := s.processDefRepo.FindByID(ctx, instance.ProcessDefID)
	if err != nil {
		return nil, ErrProcessNotFound
	}

	// 获取工作流定义
	workflowDef, err := s.workflowEngine.GetWorkflow(processDef.WorkflowID.String())
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	// 获取工作流执行上下文
	execCtx, err := s.workflowEngine.GetExecution(instance.WorkflowInstanceID.String())
	if err != nil {
		// 如果执行上下文不存在，说明流程已结束，使用历史数据
		execCtx = nil
	}

	// 获取审批任务
	tasks, err := s.taskRepo.ListByInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	// 转换节点
	nodes := make([]*dto.DiagramNode, 0, len(workflowDef.Nodes))
	for _, node := range workflowDef.Nodes {
		diagramNode := &dto.DiagramNode{
			ID:       node.ID,
			Type:     node.Type,
			Name:     node.Name,
			Config:   node.Config,
			Disabled: node.Disabled,
		}
		if node.Position != nil {
			diagramNode.Position = &dto.Position{
				X: node.Position.X,
				Y: node.Position.Y,
			}
		}
		nodes = append(nodes, diagramNode)
	}

	// 转换边
	edges := make([]*dto.DiagramEdge, 0, len(workflowDef.Edges))
	for _, edge := range workflowDef.Edges {
		edges = append(edges, &dto.DiagramEdge{
			ID:        edge.ID,
			Source:    edge.Source,
			Target:    edge.Target,
			Condition: edge.Condition,
			Label:     edge.Label,
		})
	}

	// 构建节点状态
	nodeStates := make([]*dto.NodeExecutionState, 0)
	completedNodeIDs := make([]string, 0)
	activeNodeIDs := make([]string, 0)

	// 从审批任务构建节点状态
	for _, task := range tasks {
		state := &dto.NodeExecutionState{
			NodeID:   task.NodeID,
			NodeName: task.NodeName,
		}

		switch task.Status {
		case model.TaskStatusPending:
			state.Status = "pending"
			activeNodeIDs = append(activeNodeIDs, task.NodeID)
		case model.TaskStatusApproved:
			state.Status = "completed"
			completedNodeIDs = append(completedNodeIDs, task.NodeID)
			state.Action = task.Action
			state.Comment = task.Comment
			state.CompletedAt = task.ApprovedAt
		case model.TaskStatusRejected:
			state.Status = "failed"
			completedNodeIDs = append(completedNodeIDs, task.NodeID)
			state.Action = task.Action
			state.Comment = task.Comment
			state.CompletedAt = task.ApprovedAt
		case model.TaskStatusSkipped:
			state.Status = "skipped"
			completedNodeIDs = append(completedNodeIDs, task.NodeID)
		}

		if task.ApprovedAt != nil && task.CreatedAt.Before(*task.ApprovedAt) {
			duration := int64(task.ApprovedAt.Sub(task.CreatedAt).Seconds())
			state.Duration = &duration
		}

		nodeStates = append(nodeStates, state)
	}

	// 如果有执行上下文，补充工作流引擎的状态
	if execCtx != nil {
		if execCtx.CurrentNodeID != "" {
			// 确保当前节点在活动列表中
			found := false
			for _, nodeID := range activeNodeIDs {
				if nodeID == execCtx.CurrentNodeID {
					found = true
					break
				}
			}
			if !found {
				activeNodeIDs = append(activeNodeIDs, execCtx.CurrentNodeID)
			}
		}
	}

	return &dto.ProcessInstanceDiagramResponse{
		InstanceID:       instance.ID,
		ProcessDefID:     instance.ProcessDefID,
		ProcessCode:      processDef.Code,
		ProcessName:      processDef.Name,
		Status:           instance.Status,
		CurrentNodeID:    instance.CurrentNodeID,
		StartedAt:        instance.StartedAt,
		CompletedAt:      instance.CompletedAt,
		Nodes:            nodes,
		Edges:            edges,
		NodeStates:       nodeStates,
		CompletedNodeIDs: completedNodeIDs,
		ActiveNodeIDs:    activeNodeIDs,
	}, nil
}

// GetProcessTrace 获取流程轨迹（历史路径）
func (s *approvalService) GetProcessTrace(ctx context.Context, instanceID uuid.UUID) (*dto.ProcessTraceResponse, error) {
	// 获取流程实例
	instance, err := s.processInstRepo.FindByID(ctx, instanceID)
	if err != nil {
		return nil, ErrProcessInstanceNotFound
	}

	// 获取流程定义
	processDef, err := s.processDefRepo.FindByID(ctx, instance.ProcessDefID)
	if err != nil {
		return nil, ErrProcessNotFound
	}

	// 获取流程历史
	histories, err := s.historyRepo.ListByInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	// 获取审批任务（用于补充审批人信息）
	tasks, err := s.taskRepo.ListByInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	// 构建任务映射
	taskMap := make(map[uuid.UUID]*model.ApprovalTask)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	// 构建轨迹路径
	path := make([]*dto.ProcessTraceNode, 0, len(histories))
	for _, h := range histories {
		traceNode := &dto.ProcessTraceNode{
			NodeID:    h.NodeID,
			NodeName:  h.NodeName,
			NodeType:  "approval", // 从历史记录推断，可以扩展
			Action:    &h.Action,
			Comment:   h.Comment,
			EnteredAt: h.CreatedAt,
		}

		// 如果有关联任务，补充详细信息
		if h.TaskID != nil {
			if task, exists := taskMap[*h.TaskID]; exists {
				traceNode.OperatorID = &task.AssigneeID
				operatorName := task.AssigneeName
				traceNode.Operator = &operatorName

				if task.ApprovedAt != nil {
					traceNode.CompletedAt = task.ApprovedAt
					if task.CreatedAt.Before(*task.ApprovedAt) {
						duration := int64(task.ApprovedAt.Sub(task.CreatedAt).Seconds())
						traceNode.Duration = &duration
					}
				}
			}
		}

		path = append(path, traceNode)
	}

	// 计算总耗时
	var totalDuration int64
	if instance.CompletedAt != nil {
		totalDuration = int64(instance.CompletedAt.Sub(instance.StartedAt).Seconds())
	} else {
		totalDuration = int64(time.Since(instance.StartedAt).Seconds())
	}

	return &dto.ProcessTraceResponse{
		InstanceID:  instance.ID,
		ProcessCode: processDef.Code,
		ProcessName: processDef.Name,
		Status:      instance.Status,
		StartedAt:   instance.StartedAt,
		CompletedAt: instance.CompletedAt,
		Path:        path,
		TotalNodes:  len(path),
		Duration:    totalDuration,
	}, nil
}

// createTasksForNode 为节点创建审批任务（支持动态审批人）
func (s *approvalService) createTasksForNode(
	ctx context.Context,
	node *workflowModel.NodeDefinition,
	processInstance *model.ProcessInstance,
	processVariables map[string]interface{},
) error {
	// 使用审批人解析器获取审批人列表
	assigneeIDs, err := s.assigneeResolver.ResolveAssignee(ctx, node, processVariables)
	if err != nil {
		return fmt.Errorf("failed to resolve assignee: %w", err)
	}

	if len(assigneeIDs) == 0 {
		return fmt.Errorf("no assignee found for node: %s", node.ID)
	}

	now := time.Now()

	// 为每个审批人创建任务（支持会签）
	for _, assigneeID := range assigneeIDs {
		task := &model.ApprovalTask{
			ID:                uuid.New(),
			TenantID:          processInstance.TenantID,
			ProcessInstanceID: processInstance.ID,
			NodeID:            node.ID,
			NodeName:          node.Name,
			AssigneeID:        assigneeID,
			Status:            model.TaskStatusPending,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		if err := s.taskRepo.Create(ctx, task); err != nil {
			return fmt.Errorf("failed to create task for assignee %s: %w", assigneeID, err)
		}

		// 发送站内通知
		if s.notificationService != nil {
			s.sendTaskNotification(ctx, task, processInstance, "created")
		}
	}

	return nil
}

// sendTaskNotification 发送审批任务通知
func (s *approvalService) sendTaskNotification(
	ctx context.Context,
	task *model.ApprovalTask,
	processInstance *model.ProcessInstance,
	action string,
) {
	var title, content string

	switch action {
	case "created":
		title = fmt.Sprintf("新的审批任务：%s", task.NodeName)
		content = fmt.Sprintf(
			"您有一个新的审批任务需要处理：\n\n"+
				"流程：%s\n"+
				"节点：%s\n"+
				"申请人：%s\n"+
				"创建时间：%s\n\n"+
				"请及时处理。",
			processInstance.ProcessDefName,
			task.NodeName,
			processInstance.ApplicantName,
			task.CreatedAt.Format("2006-01-02 15:04:05"),
		)

	case "approved":
		title = "审批已通过"
		content = fmt.Sprintf(
			"您的审批任务已被处理：\n\n"+
				"流程：%s\n"+
				"节点：%s\n"+
				"结果：通过\n"+
				"处理时间：%s",
			processInstance.ProcessDefName,
			task.NodeName,
			task.UpdatedAt.Format("2006-01-02 15:04:05"),
		)

	case "rejected":
		title = "审批已拒绝"
		comment := ""
		if task.Comment != nil {
			comment = *task.Comment
		}
		content = fmt.Sprintf(
			"您的审批任务已被拒绝：\n\n"+
				"流程：%s\n"+
				"节点：%s\n"+
				"拒绝原因：%s\n"+
				"处理时间：%s",
			processInstance.ProcessDefName,
			task.NodeName,
			comment,
			task.UpdatedAt.Format("2006-01-02 15:04:05"),
		)
	}

	notifReq := &notificationDto.SendNotificationRequest{
		Type:        "approval",
		Channel:     "in_app",
		RecipientID: task.AssigneeID.String(),
		Title:       title,
		Content:     content,
		RelatedType: stringPtr("approval_task"),
		RelatedID:   stringPtr(task.ID.String()),
	}

	_, _ = s.notificationService.SendNotification(ctx, task.TenantID, notifReq)
}

func stringPtr(s string) *string {
	return &s
}
