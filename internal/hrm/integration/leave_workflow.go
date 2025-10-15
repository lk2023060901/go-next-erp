package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/pkg/workflow"
)

// LeaveWorkflowEngine 请假工作流引擎
type LeaveWorkflowEngine struct {
	engine *workflow.Engine
}

// NewLeaveWorkflowEngine 创建请假工作流引擎
func NewLeaveWorkflowEngine(engine *workflow.Engine) *LeaveWorkflowEngine {
	return &LeaveWorkflowEngine{
		engine: engine,
	}
}

// CreateLeaveApprovalWorkflow 创建请假审批工作流
// 支持：
// 1. 串行执行（通过Edges定义顺序）
// 2. 通过组织架构指定审批人（approver_type支持：direct_manager, dept_manager, hr等）
// 3. 可自定义多个审批步骤（通过ApprovalRules配置）
func (e *LeaveWorkflowEngine) CreateLeaveApprovalWorkflow(
	leaveType *model.LeaveType,
) (*workflow.WorkflowDefinition, error) {
	workflowID := fmt.Sprintf("leave-approval-%s", leaveType.ID.String())

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

	// 2. 验证节点 - 验证请假申请的基本信息
	validateNode := &workflow.NodeDefinition{
		ID:   "validate",
		Name: "验证请假申请",
		Type: "validation",
		Config: map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"field":    "duration",
					"operator": ">=",
					"value":    leaveType.MinDuration,
					"message":  fmt.Sprintf("请假时长不能少于%.1f%s", leaveType.MinDuration, leaveType.Unit),
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

	// 3. 创建审批节点链 - 串行执行
	// 支持按组织架构配置审批人
	if leaveType.ApprovalRules != nil {
		// 3.1 默认审批链（串行执行）
		if len(leaveType.ApprovalRules.DefaultChain) > 0 {
			for i, approvalNode := range leaveType.ApprovalRules.DefaultChain {
				nodeID := fmt.Sprintf("approval-level-%d", approvalNode.Level)

				// 根据审批人类型设置不同的配置
				approverConfig := map[string]interface{}{
					"level":         approvalNode.Level,
					"approver_type": approvalNode.ApproverType, // direct_manager, dept_manager, hr, general_manager, custom
					"required":      approvalNode.Required,
					"timeout":       "72h", // 72小时超时
				}

				// 如果是自定义审批人，添加具体的审批人ID
				if approvalNode.ApproverType == "custom" && approvalNode.ApproverID != nil && *approvalNode.ApproverID != "" {
					approverConfig["approver_id"] = *approvalNode.ApproverID
				}

				// 创建审批节点
				node := &workflow.NodeDefinition{
					ID:      nodeID,
					Name:    getApproverTypeName(string(approvalNode.ApproverType), int32(approvalNode.Level)),
					Type:    "approval",
					Config:  approverConfig,
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
		}

		// 3.2 基于天数的审批规则（支持不同请假时长走不同审批流程）
		if len(leaveType.ApprovalRules.DurationRules) > 0 {
			// 创建条件分支节点
			conditionNode := &workflow.NodeDefinition{
				ID:   "duration-check",
				Name: "请假时长判断",
				Type: "condition",
				Config: map[string]interface{}{
					"rules": buildDurationRules(leaveType.ApprovalRules.DurationRules),
				},
			}
			nodes = append(nodes, conditionNode)
			edges = append(edges, &workflow.Edge{
				ID:     fmt.Sprintf("%s-condition", lastNodeID),
				Source: lastNodeID,
				Target: "duration-check",
			})

			// 为每个天数规则创建独立的审批链
			for ruleIdx, durationRule := range leaveType.ApprovalRules.DurationRules {
				branchLastNodeID := "duration-check"

				for i, approvalNode := range durationRule.ApprovalChain {
					nodeID := fmt.Sprintf("approval-rule%d-level-%d", ruleIdx, approvalNode.Level)

					node := &workflow.NodeDefinition{
						ID:   nodeID,
						Name: fmt.Sprintf("规则%d-第%d级审批", ruleIdx+1, i+1),
						Type: "approval",
						Config: map[string]interface{}{
							"level":         approvalNode.Level,
							"approver_type": approvalNode.ApproverType,
							"approver_id":   approvalNode.ApproverID,
							"required":      approvalNode.Required,
							"timeout":       "72h",
						},
						Timeout: 72 * time.Hour,
					}
					nodes = append(nodes, node)

					// 第一个节点从条件节点连接
					if i == 0 {
						maxDur := "∞"
						if durationRule.MaxDuration != nil {
							maxDur = fmt.Sprintf("%.1f", *durationRule.MaxDuration)
						}
						edges = append(edges, &workflow.Edge{
							ID:        fmt.Sprintf("condition-rule%d", ruleIdx),
							Source:    "duration-check",
							Target:    nodeID,
							Condition: fmt.Sprintf("duration >= %.1f && duration <= %s", durationRule.MinDuration, maxDur),
							Label:     fmt.Sprintf("%.1f-%s天", durationRule.MinDuration, maxDur),
						})
					} else {
						edges = append(edges, &workflow.Edge{
							ID:     fmt.Sprintf("%s-%s", branchLastNodeID, nodeID),
							Source: branchLastNodeID,
							Target: nodeID,
						})
					}

					branchLastNodeID = nodeID
				}

				// 所有分支最后汇聚到额度扣减节点
				lastNodeID = branchLastNodeID
			}
		}

		// 4. 额度扣减节点（如果需要扣减额度）
		if leaveType.DeductQuota {
			deductNode := &workflow.NodeDefinition{
				ID:   "deduct-quota",
				Name: "扣减请假额度",
				Type: "quota-deduct",
				Config: map[string]interface{}{
					"leave_type_id": leaveType.ID.String(),
					"deduct_field":  "duration",
				},
			}
			nodes = append(nodes, deductNode)
			edges = append(edges, &workflow.Edge{
				ID:     fmt.Sprintf("%s-deduct", lastNodeID),
				Source: lastNodeID,
				Target: "deduct-quota",
			})
			lastNodeID = "deduct-quota"
		}

		// 5. 通知节点 - 通知申请人审批结果
		notifyNode := &workflow.NodeDefinition{
			ID:   "notify",
			Name: "发送审批通知",
			Type: "notification",
			Config: map[string]interface{}{
				"template": "leave-approved",
				"channels": []string{"email", "system", "sms"},
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

		// 6. 结束节点
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
	}

	// 创建工作流定义
	def := &workflow.WorkflowDefinition{
		ID:          workflowID,
		Name:        fmt.Sprintf("%s审批流程", leaveType.Name),
		Description: fmt.Sprintf("请假类型: %s 的审批工作流（支持串行审批、组织架构审批人配置）", leaveType.Name),
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

// ExecuteLeaveApproval 执行请假审批
func (e *LeaveWorkflowEngine) ExecuteLeaveApproval(
	ctx context.Context,
	workflowID string,
	request *model.LeaveRequest,
	triggerBy string,
) (string, error) {
	input := map[string]interface{}{
		"request_id":    request.ID.String(),
		"employee_id":   request.EmployeeID.String(),
		"leave_type_id": request.LeaveTypeID.String(),
		"start_time":    request.StartTime,
		"end_time":      request.EndTime,
		"duration":      request.Duration,
		"reason":        request.Reason,
	}

	executionID, err := e.engine.Execute(ctx, workflowID, input, triggerBy)
	if err != nil {
		return "", fmt.Errorf("failed to execute workflow: %w", err)
	}

	return executionID, nil
}

// HandleApproval 处理审批操作
func (e *LeaveWorkflowEngine) HandleApproval(
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
func (e *LeaveWorkflowEngine) GetExecutionStatus(executionID string) (*workflow.ExecutionContext, error) {
	return e.engine.GetExecution(executionID)
}

// getApproverTypeName 获取审批人类型名称
func getApproverTypeName(approverType string, level int32) string {
	names := map[string]string{
		"direct_manager":  "直接主管审批",
		"dept_manager":    "部门经理审批",
		"hr":              "人事审批",
		"general_manager": "总经理审批",
		"custom":          "指定审批人",
	}

	if name, ok := names[approverType]; ok {
		return fmt.Sprintf("第%d级-%s", level, name)
	}
	return fmt.Sprintf("第%d级审批", level)
}

// buildDurationRules 构建天数规则
func buildDurationRules(rules []*model.DurationRule) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rules))
	for _, rule := range rules {
		maxDur := "∞"
		if rule.MaxDuration != nil {
			maxDur = fmt.Sprintf("%.1f", *rule.MaxDuration)
		}
		result = append(result, map[string]interface{}{
			"min_duration": rule.MinDuration,
			"max_duration": maxDur,
			"condition":    fmt.Sprintf("duration >= %.1f && duration <= %s", rule.MinDuration, maxDur),
		})
	}
	return result
}
