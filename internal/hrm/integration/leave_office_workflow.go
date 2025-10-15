package integration

import (
	"context"
	"fmt"

	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/pkg/workflow"
)

// LeaveOfficeWorkflowEngine 外出工作流引擎
type LeaveOfficeWorkflowEngine struct {
	engine *workflow.Engine
}

// NewLeaveOfficeWorkflowEngine 创建外出工作流引擎
func NewLeaveOfficeWorkflowEngine(engine *workflow.Engine) *LeaveOfficeWorkflowEngine {
	return &LeaveOfficeWorkflowEngine{
		engine: engine,
	}
}

// LeaveOfficeApprovalConfig 外出审批配置
type LeaveOfficeApprovalConfig struct {
	TenantID   string
	EmployeeID string
	Duration   float64 // 外出时长（小时）
}

// CreateLeaveOfficeApprovalWorkflow 创建外出审批工作流
// 外出审批规则：通常只需直属上级审批，流程简化
func (e *LeaveOfficeWorkflowEngine) CreateLeaveOfficeApprovalWorkflow(
	config *LeaveOfficeApprovalConfig,
) (*workflow.WorkflowDefinition, error) {
	workflowID := fmt.Sprintf("leave-office-approval-%s", config.TenantID)

	nodes := make([]*workflow.NodeDefinition, 0)
	edges := make([]*workflow.Edge, 0)

	// 1. 开始节点
	startNode := &workflow.NodeDefinition{
		ID:   "start",
		Name: "开始",
		Type: "start",
	}
	nodes = append(nodes, startNode)

	// 2. 验证节点
	validationNode := &workflow.NodeDefinition{
		ID:   "validation",
		Name: "验证外出申请",
		Type: "validation",
		Config: map[string]interface{}{
			"rules": []string{
				"check_time_conflict",  // 检查时间冲突
				"check_duration_limit", // 检查时长限制（建议≤24小时）
				"check_not_past_time",  // 检查不能是过去时间
			},
		},
	}
	nodes = append(nodes, validationNode)
	edges = append(edges, &workflow.Edge{Source: "start", Target: "validation"})

	// 3. 直属上级审批（外出通常只需一级审批）
	directManagerNode := &workflow.NodeDefinition{
		ID:   "approval-direct_manager",
		Name: "直属上级审批",
		Type: "approval",
		Config: map[string]interface{}{
			"approver_type": "direct_manager",
			"employee_id":   config.EmployeeID,
			"timeout":       "24h", // 24小时超时
		},
	}
	nodes = append(nodes, directManagerNode)
	edges = append(edges, &workflow.Edge{Source: "validation", Target: "approval-direct_manager"})

	// 4. 通知节点
	notifyNode := &workflow.NodeDefinition{
		ID:   "notify",
		Name: "发送通知",
		Type: "notification",
		Config: map[string]interface{}{
			"notify_employee": true,  // 通知申请人
			"notify_approver": false, // 不通知审批人
		},
	}
	nodes = append(nodes, notifyNode)
	edges = append(edges, &workflow.Edge{Source: "approval-direct_manager", Target: "notify"})

	// 5. 结束节点
	endNode := &workflow.NodeDefinition{
		ID:   "end",
		Name: "结束",
		Type: "end",
	}
	nodes = append(nodes, endNode)
	edges = append(edges, &workflow.Edge{Source: "notify", Target: "end"})

	workflowDef := &workflow.WorkflowDefinition{
		ID:          workflowID,
		Name:        "外出审批流程",
		Description: "简化的外出审批流程，只需直属上级审批",
		Version:     1,
		Nodes:       nodes,
		Edges:       edges,
	}

	return workflowDef, nil
}

// ExecuteLeaveOfficeApproval 执行外出审批
func (e *LeaveOfficeWorkflowEngine) ExecuteLeaveOfficeApproval(
	ctx context.Context,
	workflowID string,
	leaveOffice *model.LeaveOffice,
	submitterID string,
) (string, error) {
	// 1. 创建工作流配置
	config := &LeaveOfficeApprovalConfig{
		TenantID:   leaveOffice.TenantID.String(),
		EmployeeID: leaveOffice.EmployeeID.String(),
		Duration:   leaveOffice.Duration,
	}

	// 2. 创建工作流定义
	workflowDef, err := e.CreateLeaveOfficeApprovalWorkflow(config)
	if err != nil {
		return "", fmt.Errorf("failed to create workflow definition: %w", err)
	}

	// 3. 注册工作流（如果尚未注册）
	if err := e.engine.CreateWorkflow(workflowDef); err != nil {
		// 忽略已存在错误
		if err != workflow.ErrWorkflowAlreadyExists {
			return "", fmt.Errorf("failed to create workflow: %w", err)
		}
	}

	// 4. 准备执行参数
	params := map[string]interface{}{
		"leave_office_id": leaveOffice.ID.String(),
		"tenant_id":       leaveOffice.TenantID.String(),
		"employee_id":     leaveOffice.EmployeeID.String(),
		"employee_name":   leaveOffice.EmployeeName,
		"start_time":      leaveOffice.StartTime.Format("2006-01-02 15:04:05"),
		"end_time":        leaveOffice.EndTime.Format("2006-01-02 15:04:05"),
		"duration":        leaveOffice.Duration,
		"destination":     leaveOffice.Destination,
		"purpose":         leaveOffice.Purpose,
		"submitter_id":    submitterID,
	}

	// 5. 执行工作流
	executionID, err := e.engine.Execute(ctx, workflowID, params, submitterID)
	if err != nil {
		return "", fmt.Errorf("failed to execute workflow: %w", err)
	}

	return executionID, nil
}
