package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/pkg/workflow"
)

// BusinessTripWorkflowEngine 出差工作流引擎
type BusinessTripWorkflowEngine struct {
	engine *workflow.Engine
}

// NewBusinessTripWorkflowEngine 创建出差工作流引擎
func NewBusinessTripWorkflowEngine(engine *workflow.Engine) *BusinessTripWorkflowEngine {
	return &BusinessTripWorkflowEngine{
		engine: engine,
	}
}

// BusinessTripApprovalConfig 出差审批配置
type BusinessTripApprovalConfig struct {
	TenantID      uuid.UUID
	Duration      float64 // 出差天数
	EstimatedCost float64 // 预计费用
}

// CreateBusinessTripApprovalWorkflow 创建出差审批工作流
// 审批规则：
// - ≤3天且费用≤5000: 直属上级审批
// - 3-7天或费用5000-10000: 直属上级 → 部门经理
// - >7天或费用>10000: 直属上级 → 部门经理 → 总经理
// - 费用>5000: 需要财务审批
func (e *BusinessTripWorkflowEngine) CreateBusinessTripApprovalWorkflow(
	config *BusinessTripApprovalConfig,
) (*workflow.WorkflowDefinition, error) {
	workflowID := fmt.Sprintf("business-trip-approval-%s", config.TenantID.String())

	var nodes []*workflow.NodeDefinition
	var edges []*workflow.Edge

	// 1. 开始节点
	startNode := &workflow.NodeDefinition{
		ID:   "start",
		Name: "开始",
		Type: "start",
		Config: map[string]interface{}{
			"trigger": "manual",
		},
	}
	nodes = append(nodes, startNode)

	// 2. 验证节点 - 检查时间冲突、提前申请天数等
	validationNode := &workflow.NodeDefinition{
		ID:   "validation",
		Name: "验证出差申请",
		Type: "validation",
		Config: map[string]interface{}{
			"rules": []string{
				"check_time_conflict",
				"check_advance_days",
				"validate_dates",
			},
		},
	}
	nodes = append(nodes, validationNode)
	edges = append(edges, &workflow.Edge{
		ID:     "start-validation",
		Source: "start",
		Target: "validation",
	})

	lastNodeID := "validation"

	// 3. 根据出差天数和费用确定审批链
	var approvalChain []string

	if config.Duration <= 3 && config.EstimatedCost <= 5000 {
		// 简单出差：只需直属上级审批
		approvalChain = []string{"direct_manager"}
	} else if config.Duration <= 7 && config.EstimatedCost <= 10000 {
		// 中等出差：直属上级 + 部门经理
		approvalChain = []string{"direct_manager", "dept_manager"}
	} else {
		// 复杂出差：直属上级 + 部门经理 + 总经理
		approvalChain = []string{"direct_manager", "dept_manager", "general_manager"}
	}

	// 3.1 创建审批节点（串行）
	for i, approverType := range approvalChain {
		nodeID := fmt.Sprintf("approval-%s", approverType)

		node := &workflow.NodeDefinition{
			ID:   nodeID,
			Name: getBusinessTripApproverTypeName(approverType),
			Type: "approval",
			Config: map[string]interface{}{
				"level":         i + 1,
				"approver_type": approverType,
				"required":      true,
				"timeout":       "72h",
			},
			Timeout: 72 * time.Hour,
			RetryPolicy: &workflow.RetryPolicy{
				MaxAttempts: 3,
				Delay:       1 * time.Hour,
				BackoffRate: 1.5,
			},
		}
		nodes = append(nodes, node)

		// 串行连接：上一个节点 -> 当前审批节点
		edges = append(edges, &workflow.Edge{
			ID:     fmt.Sprintf("%s-%s", lastNodeID, nodeID),
			Source: lastNodeID,
			Target: nodeID,
			Label:  fmt.Sprintf("第%d级审批", i+1),
		})

		lastNodeID = nodeID
	}

	// 3.2 如果费用超过5000，需要财务审批
	if config.EstimatedCost > 5000 {
		financeNode := &workflow.NodeDefinition{
			ID:   "approval-finance",
			Name: "财务审批",
			Type: "approval",
			Config: map[string]interface{}{
				"level":         len(approvalChain) + 1,
				"approver_type": "finance",
				"required":      true,
				"timeout":       "48h",
			},
			Timeout: 48 * time.Hour,
		}
		nodes = append(nodes, financeNode)

		edges = append(edges, &workflow.Edge{
			ID:     fmt.Sprintf("%s-finance", lastNodeID),
			Source: lastNodeID,
			Target: "approval-finance",
			Label:  "财务审批",
		})

		lastNodeID = "approval-finance"
	}

	// 4. 通知节点 - 通知申请人审批结果
	notifyNode := &workflow.NodeDefinition{
		ID:   "notify",
		Name: "发送审批通知",
		Type: "notification",
		Config: map[string]interface{}{
			"template": "business-trip-approved",
			"channels": []string{"email", "system"},
			"recipients": map[string]interface{}{
				"employee": "{{.employee_id}}",
				"cc":       []string{"hr", "manager"},
			},
		},
	}
	nodes = append(nodes, notifyNode)
	edges = append(edges, &workflow.Edge{
		ID:     fmt.Sprintf("%s-notify", lastNodeID),
		Source: lastNodeID,
		Target: "notify",
	})

	// 5. 结束节点
	endNode := &workflow.NodeDefinition{
		ID:   "end",
		Name: "结束",
		Type: "end",
		Config: map[string]interface{}{
			"final_status": "approved",
		},
	}
	nodes = append(nodes, endNode)
	edges = append(edges, &workflow.Edge{
		ID:     "notify-end",
		Source: "notify",
		Target: "end",
	})

	// 创建工作流定义
	def := &workflow.WorkflowDefinition{
		ID:          workflowID,
		Name:        "出差审批流程",
		Description: fmt.Sprintf("出差审批工作流（天数: %.1f天, 预算: %.2f元）", config.Duration, config.EstimatedCost),
		Version:     1,
		Status:      workflow.WorkflowStatusActive,
		Nodes:       nodes,
		Edges:       edges,
		Settings: &workflow.WorkflowSettings{
			ExecutionTimeout: 168 * time.Hour, // 7天总超时
			MaxRetries:       3,
			RetryDelay:       5 * time.Minute,
			OnError:          "stop", // 遇到错误停止执行
		},
	}

	return def, nil
}

// ExecuteBusinessTripApproval 执行出差审批
func (e *BusinessTripWorkflowEngine) ExecuteBusinessTripApproval(
	ctx context.Context,
	workflowID string,
	trip *model.BusinessTrip,
	triggerBy string,
) (string, error) {
	input := map[string]interface{}{
		"trip_id":        trip.ID.String(),
		"employee_id":    trip.EmployeeID.String(),
		"start_time":     trip.StartTime,
		"end_time":       trip.EndTime,
		"duration":       trip.Duration,
		"destination":    trip.Destination,
		"purpose":        trip.Purpose,
		"estimated_cost": trip.EstimatedCost,
	}

	executionID, err := e.engine.Execute(ctx, workflowID, input, triggerBy)
	if err != nil {
		return "", fmt.Errorf("failed to execute workflow: %w", err)
	}

	return executionID, nil
}

// HandleApproval 处理审批操作
func (e *BusinessTripWorkflowEngine) HandleApproval(
	ctx context.Context,
	executionID string,
	nodeID string,
	approved bool,
	approverID uuid.UUID,
	comment string,
) error {
	// 获取执行上下文
	execCtx, err := e.engine.GetExecution(executionID)
	if err != nil {
		return fmt.Errorf("failed to get execution context: %w", err)
	}

	// 更新节点状态
	nodeState, ok := execCtx.GetNodeState(nodeID)
	if !ok {
		return fmt.Errorf("node state not found: %s", nodeID)
	}

	nodeState.Output = map[string]interface{}{
		"approved":    approved,
		"approver_id": approverID.String(),
		"comment":     comment,
		"approved_at": time.Now(),
	}

	if approved {
		nodeState.Status = workflow.NodeStatusCompleted
		// 审批通过，工作流引擎会自动继续执行下一个串行节点
		// TODO: 实现继续执行逻辑
		return nil
	} else {
		nodeState.Status = workflow.NodeStatusFailed
		nodeState.Error = "审批被拒绝"
		// 终止工作流
		return e.engine.CancelExecution(executionID)
	}
}

// GetExecutionStatus 获取执行状态
func (e *BusinessTripWorkflowEngine) GetExecutionStatus(executionID string) (*workflow.ExecutionContext, error) {
	return e.engine.GetExecution(executionID)
}

// getApproverTypeName 获取审批人类型名称
func getBusinessTripApproverTypeName(approverType string) string {
	names := map[string]string{
		"direct_manager":  "直接主管审批",
		"dept_manager":    "部门经理审批",
		"general_manager": "总经理审批",
		"finance":         "财务审批",
		"hr":              "人事审批",
	}

	if name, ok := names[approverType]; ok {
		return name
	}
	return "审批"
}
