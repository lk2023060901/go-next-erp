package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/pkg/workflow"
)

// OvertimeWorkflowEngine 加班工作流引擎
type OvertimeWorkflowEngine struct {
	engine *workflow.Engine
}

// NewOvertimeWorkflowEngine 创建加班工作流引擎
func NewOvertimeWorkflowEngine(engine *workflow.Engine) *OvertimeWorkflowEngine {
	return &OvertimeWorkflowEngine{
		engine: engine,
	}
}

// CreateOvertimeApprovalWorkflow 创建加班审批工作流
// 支持串行审批和组织架构审批人配置
func (e *OvertimeWorkflowEngine) CreateOvertimeApprovalWorkflow(
	tenantID uuid.UUID,
	approvalConfig *OvertimeApprovalConfig,
) (*workflow.WorkflowDefinition, error) {
	workflowID := fmt.Sprintf("overtime-approval-%s", tenantID.String())

	nodes := make([]*workflow.NodeDefinition, 0)
	edges := make([]*workflow.Edge, 0)

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

	// 2. 验证节点
	validateNode := &workflow.NodeDefinition{
		ID:   "validate",
		Name: "验证加班申请",
		Type: "validation",
		Config: map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"field":    "duration",
					"operator": ">",
					"value":    0.0,
					"message":  "加班时长必须大于0",
				},
				{
					"field":    "overtime_type",
					"operator": "in",
					"value":    []string{"workday", "weekend", "holiday"},
					"message":  "无效的加班类型",
				},
			},
		},
	}
	nodes = append(nodes, validateNode)
	edges = append(edges, &workflow.Edge{
		ID:     "start-validate",
		Source: "start",
		Target: "validate",
	})

	lastNodeID := "validate"

	// 3. 创建审批链（串行执行）
	if approvalConfig != nil && len(approvalConfig.ApprovalChain) > 0 {
		for i, approver := range approvalConfig.ApprovalChain {
			nodeID := fmt.Sprintf("approval-level-%d", i+1)

			approverConfig := map[string]interface{}{
				"level":         i + 1,
				"approver_type": approver.Type, // direct_manager, dept_manager, hr等
				"required":      true,
				"timeout":       "48h", // 48小时超时（加班审批比请假快）
			}

			// 如果是指定审批人
			if approver.Type == "custom" && approver.UserID != nil {
				approverConfig["approver_id"] = approver.UserID.String()
			} else if approver.Type == "department" && approver.DeptID != nil {
				approverConfig["dept_id"] = approver.DeptID.String()
			}

			// 创建审批节点
			node := &workflow.NodeDefinition{
				ID:      nodeID,
				Name:    fmt.Sprintf("第%d级-%s", i+1, getOvertimeApproverTypeName(approver.Type)),
				Type:    "approval",
				Config:  approverConfig,
				Timeout: 48 * time.Hour,
				RetryPolicy: &workflow.RetryPolicy{
					MaxAttempts: 2,
					Delay:       30 * time.Minute,
					BackoffRate: 1.5,
				},
			}
			nodes = append(nodes, node)

			// 串行连接
			edges = append(edges, &workflow.Edge{
				ID:     fmt.Sprintf("%s-%s", lastNodeID, nodeID),
				Source: lastNodeID,
				Target: nodeID,
				Label:  fmt.Sprintf("第%d级审批", i+1),
			})

			lastNodeID = nodeID
		}
	}

	// 4. 计算加班费/调休
	calculateNode := &workflow.NodeDefinition{
		ID:   "calculate",
		Name: "计算加班补偿",
		Type: "calculation",
		Config: map[string]interface{}{
			"pay_type":      "{{.pay_type}}", // money或leave
			"overtime_type": "{{.overtime_type}}",
			"duration":      "{{.duration}}",
			"rules": map[string]interface{}{
				"workday": 1.5, // 工作日加班1.5倍
				"weekend": 2.0, // 周末加班2倍
				"holiday": 3.0, // 节假日加班3倍
			},
		},
	}
	nodes = append(nodes, calculateNode)
	edges = append(edges, &workflow.Edge{
		ID:     fmt.Sprintf("%s-calculate", lastNodeID),
		Source: lastNodeID,
		Target: "calculate",
	})

	// 5. 如果是调休，更新调休余额
	updateBalanceNode := &workflow.NodeDefinition{
		ID:   "update-balance",
		Name: "更新调休余额",
		Type: "update-balance",
		Config: map[string]interface{}{
			"condition": "pay_type == 'leave'",
			"field":     "comp_off_days",
		},
	}
	nodes = append(nodes, updateBalanceNode)
	edges = append(edges, &workflow.Edge{
		ID:        "calculate-balance",
		Source:    "calculate",
		Target:    "update-balance",
		Condition: "pay_type == 'leave'",
		Label:     "调休方式",
	})

	// 6. 通知节点
	notifyNode := &workflow.NodeDefinition{
		ID:   "notify",
		Name: "发送通知",
		Type: "notification",
		Config: map[string]interface{}{
			"template": "overtime-approved",
			"channels": []string{"email", "system"},
		},
	}
	nodes = append(nodes, notifyNode)
	edges = append(edges, &workflow.Edge{
		ID:     "balance-notify",
		Source: "update-balance",
		Target: "notify",
	})
	// 如果不是调休，直接跳到通知
	edges = append(edges, &workflow.Edge{
		ID:        "calculate-notify",
		Source:    "calculate",
		Target:    "notify",
		Condition: "pay_type == 'money'",
		Label:     "加班费方式",
	})

	// 7. 结束节点
	endNode := &workflow.NodeDefinition{
		ID:   "end",
		Name: "结束",
		Type: "end",
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
		Name:        "加班审批流程",
		Description: "加班申请审批工作流（支持串行审批、组织架构审批人配置、自动计算加班费/调休）",
		Version:     1,
		Status:      workflow.WorkflowStatusActive,
		Nodes:       nodes,
		Edges:       edges,
		Settings: &workflow.WorkflowSettings{
			ExecutionTimeout: 120 * time.Hour, // 5天总超时
			MaxRetries:       2,
			RetryDelay:       10 * time.Minute,
			OnError:          "stop",
		},
	}

	return def, nil
}

// ExecuteOvertimeApproval 执行加班审批
func (e *OvertimeWorkflowEngine) ExecuteOvertimeApproval(
	ctx context.Context,
	workflowID string,
	overtime *model.Overtime,
	triggerBy string,
) (string, error) {
	input := map[string]interface{}{
		"overtime_id":   overtime.ID.String(),
		"employee_id":   overtime.EmployeeID.String(),
		"start_time":    overtime.StartTime,
		"end_time":      overtime.EndTime,
		"duration":      overtime.Duration,
		"overtime_type": string(overtime.OvertimeType),
		"pay_type":      overtime.PayType,
		"reason":        overtime.Reason,
	}

	executionID, err := e.engine.Execute(ctx, workflowID, input, triggerBy)
	if err != nil {
		return "", fmt.Errorf("failed to execute workflow: %w", err)
	}

	return executionID, nil
}

// HandleApproval 处理审批操作
func (e *OvertimeWorkflowEngine) HandleApproval(
	ctx context.Context,
	executionID string,
	nodeID string,
	approved bool,
	approverID uuid.UUID,
	comment string,
) error {
	execCtx, err := e.engine.GetExecution(executionID)
	if err != nil {
		return fmt.Errorf("failed to get execution context: %w", err)
	}

	nodeState := execCtx.NodeStates[nodeID]
	if nodeState == nil {
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
		return nil
	} else {
		nodeState.Status = workflow.NodeStatusFailed
		nodeState.Error = "审批被拒绝"
		return e.engine.CancelExecution(executionID)
	}
}

// GetExecutionStatus 获取执行状态
func (e *OvertimeWorkflowEngine) GetExecutionStatus(executionID string) (*workflow.ExecutionContext, error) {
	return e.engine.GetExecution(executionID)
}

// OvertimeApprovalConfig 加班审批配置
type OvertimeApprovalConfig struct {
	ApprovalChain []*ApproverConfig `json:"approval_chain"`
}

// ApproverConfig 审批人配置
type ApproverConfig struct {
	Type   string     `json:"type"` // direct_manager, dept_manager, hr, general_manager, custom, department
	UserID *uuid.UUID `json:"user_id,omitempty"`
	DeptID *uuid.UUID `json:"dept_id,omitempty"`
}

// getOvertimeApproverTypeName 获取审批人类型名称
func getOvertimeApproverTypeName(approverType string) string {
	names := map[string]string{
		"direct_manager":  "直接主管",
		"dept_manager":    "部门经理",
		"hr":              "人事",
		"general_manager": "总经理",
		"custom":          "指定审批人",
		"department":      "指定部门",
	}

	if name, ok := names[approverType]; ok {
		return name
	}
	return "审批"
}
