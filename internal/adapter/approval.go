package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"

	approvalv1 "github.com/lk2023060901/go-next-erp/api/approval/v1"
	"github.com/lk2023060901/go-next-erp/internal/approval/dto"
	"github.com/lk2023060901/go-next-erp/internal/approval/model"
	"github.com/lk2023060901/go-next-erp/internal/approval/service"
)

// ApprovalAdapter 审批模块适配器（实现三个服务接口）
type ApprovalAdapter struct {
	approvalv1.UnimplementedProcessDefinitionServiceServer
	approvalv1.UnimplementedProcessInstanceServiceServer
	approvalv1.UnimplementedApprovalTaskServiceServer
	approvalService service.ApprovalService
}

// NewApprovalAdapter 创建审批适配器
func NewApprovalAdapter(approvalService service.ApprovalService) *ApprovalAdapter {
	return &ApprovalAdapter{
		approvalService: approvalService,
	}
}

// ========== ProcessDefinitionService 实现 ==========

func (a *ApprovalAdapter) CreateProcessDefinition(ctx context.Context, req *approvalv1.CreateProcessDefinitionRequest) (*approvalv1.ProcessDefinitionResponse, error) {
	formID, _ := uuid.Parse(req.FormId)
	workflowID, _ := uuid.Parse(req.WorkflowId)

	createReq := &dto.CreateProcessDefRequest{
		Code:       req.Code,
		Name:       req.Name,
		Category:   req.Category,
		FormID:     formID,
		WorkflowID: workflowID,
	}

	processDef, err := a.approvalService.CreateProcessDefinition(ctx, createReq)
	if err != nil {
		return nil, err
	}

	return toProcessDefinitionResponse(processDef), nil
}

func (a *ApprovalAdapter) UpdateProcessDefinition(ctx context.Context, req *approvalv1.UpdateProcessDefinitionRequest) (*approvalv1.ProcessDefinitionResponse, error) {
	id, _ := uuid.Parse(req.Id)
	formID, _ := uuid.Parse(req.FormId)
	workflowID, _ := uuid.Parse(req.WorkflowId)

	updateReq := &dto.UpdateProcessDefRequest{
		Name:       req.Name,
		FormID:     formID,
		WorkflowID: workflowID,
		Enabled:    req.Enabled,
	}

	processDef, err := a.approvalService.UpdateProcessDefinition(ctx, id, updateReq)
	if err != nil {
		return nil, err
	}

	return toProcessDefinitionResponse(processDef), nil
}

func (a *ApprovalAdapter) GetProcessDefinition(ctx context.Context, req *approvalv1.GetProcessDefinitionRequest) (*approvalv1.ProcessDefinitionResponse, error) {
	id, _ := uuid.Parse(req.Id)

	processDef, err := a.approvalService.GetProcessDefinition(ctx, id)
	if err != nil {
		return nil, err
	}

	return toProcessDefinitionResponse(processDef), nil
}

func (a *ApprovalAdapter) ListProcessDefinitions(ctx context.Context, req *approvalv1.ListProcessDefinitionsRequest) (*approvalv1.ListProcessDefinitionsResponse, error) {
	// TODO: 从 context 获取 tenantID
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	processDefsDto, err := a.approvalService.ListProcessDefinitions(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	items := make([]*approvalv1.ProcessDefinitionResponse, len(processDefsDto))
	for i, pd := range processDefsDto {
		items[i] = toProcessDefinitionResponse(pd)
	}

	return &approvalv1.ListProcessDefinitionsResponse{
		Items: items,
		Total: int32(len(items)),
	}, nil
}

func (a *ApprovalAdapter) DeleteProcessDefinition(ctx context.Context, req *approvalv1.DeleteProcessDefinitionRequest) (*emptypb.Empty, error) {
	id, _ := uuid.Parse(req.Id)

	err := a.approvalService.DeleteProcessDefinition(ctx, id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (a *ApprovalAdapter) EnableProcessDefinition(ctx context.Context, req *approvalv1.EnableProcessDefinitionRequest) (*emptypb.Empty, error) {
	id, _ := uuid.Parse(req.Id)

	err := a.approvalService.SetProcessDefinitionStatus(ctx, id, true)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (a *ApprovalAdapter) DisableProcessDefinition(ctx context.Context, req *approvalv1.DisableProcessDefinitionRequest) (*emptypb.Empty, error) {
	id, _ := uuid.Parse(req.Id)

	err := a.approvalService.SetProcessDefinitionStatus(ctx, id, false)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (a *ApprovalAdapter) GetProcessStats(ctx context.Context, req *approvalv1.GetProcessStatsRequest) (*approvalv1.ProcessStatsResponse, error) {
	id, _ := uuid.Parse(req.Id)

	stats, err := a.approvalService.GetProcessStats(ctx, id)
	if err != nil {
		return nil, err
	}

	return &approvalv1.ProcessStatsResponse{
		ProcessDefId:      stats.ProcessDefID.String(),
		ProcessCode:       stats.ProcessCode,
		ProcessName:       stats.ProcessName,
		TotalInstances:    int32(stats.TotalInstances),
		PendingInstances:  int32(stats.PendingInstances),
		ApprovedInstances: int32(stats.ApprovedInstances),
		RejectedInstances: int32(stats.RejectedInstances),
		AvgDuration:       stats.AvgDuration,
	}, nil
}

// ========== ProcessInstanceService 实现 ==========

func (a *ApprovalAdapter) StartProcess(ctx context.Context, req *approvalv1.StartProcessRequest) (*approvalv1.ProcessInstanceResponse, error) {
	processDefID, _ := uuid.Parse(req.ProcessDefId)
	// TODO: 从 context 获取 applicantID
	applicantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// 转换 FormData 从 map[string]string 到 map[string]interface{}
	formDataMap := make(map[string]interface{})
	for k, v := range req.FormData {
		formDataMap[k] = v
	}

	startReq := &dto.StartProcessRequest{
		ProcessDefID: processDefID,
		ApplicantID:  applicantID,
		FormData:     formDataMap,
	}

	instance, err := a.approvalService.StartProcess(ctx, startReq)
	if err != nil {
		return nil, err
	}

	return toProcessInstanceResponse(instance), nil
}

func (a *ApprovalAdapter) GetProcessInstance(ctx context.Context, req *approvalv1.GetProcessInstanceRequest) (*approvalv1.ProcessInstanceResponse, error) {
	id, _ := uuid.Parse(req.Id)

	instance, err := a.approvalService.GetProcessInstance(ctx, id)
	if err != nil {
		return nil, err
	}

	return toProcessInstanceResponse(instance), nil
}

func (a *ApprovalAdapter) ListMyApplications(ctx context.Context, req *approvalv1.ListMyApplicationsRequest) (*approvalv1.ListProcessInstancesResponse, error) {
	// TODO: 从 context 获取 applicantID
	applicantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	instances, err := a.approvalService.ListMyApplications(ctx, applicantID, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}

	items := make([]*approvalv1.ProcessInstanceResponse, len(instances))
	for i, inst := range instances {
		items[i] = toProcessInstanceResponse(inst)
	}

	return &approvalv1.ListProcessInstancesResponse{
		Items: items,
		Total: int32(len(items)),
	}, nil
}

func (a *ApprovalAdapter) WithdrawProcess(ctx context.Context, req *approvalv1.WithdrawProcessRequest) (*emptypb.Empty, error) {
	id, _ := uuid.Parse(req.Id)
	// TODO: 从 context 获取 operatorID
	operatorID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	err := a.approvalService.WithdrawProcess(ctx, id, operatorID)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (a *ApprovalAdapter) CancelProcess(ctx context.Context, req *approvalv1.CancelProcessRequest) (*emptypb.Empty, error) {
	id, _ := uuid.Parse(req.Id)
	// TODO: 从 context 获取 operatorID
	operatorID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	var reason *string
	if req.Reason != "" {
		reason = &req.Reason
	}

	err := a.approvalService.CancelProcess(ctx, id, operatorID, reason)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (a *ApprovalAdapter) ListProcessInstances(ctx context.Context, req *approvalv1.ListProcessInstancesRequest) (*approvalv1.ListProcessInstancesResponse, error) {
	// TODO: 从 context 获取 tenantID
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	var processDefID *uuid.UUID
	if req.ProcessDefId != "" {
		id, _ := uuid.Parse(req.ProcessDefId)
		processDefID = &id
	}

	var status *model.ProcessStatus
	if req.Status != "" {
		s := model.ProcessStatus(req.Status)
		status = &s
	}

	var applicantID *uuid.UUID
	if req.ApplicantId != "" {
		id, _ := uuid.Parse(req.ApplicantId)
		applicantID = &id
	}

	var startDate, endDate *time.Time
	if req.StartDate != "" {
		t, _ := time.Parse(time.RFC3339, req.StartDate)
		startDate = &t
	}
	if req.EndDate != "" {
		t, _ := time.Parse(time.RFC3339, req.EndDate)
		endDate = &t
	}

	instances, total, err := a.approvalService.ListProcessInstances(ctx, tenantID, processDefID, status, applicantID, startDate, endDate, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}

	items := make([]*approvalv1.ProcessInstanceResponse, len(instances))
	for i, inst := range instances {
		items[i] = toProcessInstanceResponse(inst)
	}

	return &approvalv1.ListProcessInstancesResponse{
		Items: items,
		Total: int32(total),
	}, nil
}

func (a *ApprovalAdapter) GetInstanceStatsSummary(ctx context.Context, req *approvalv1.GetInstanceStatsSummaryRequest) (*approvalv1.InstanceStatsSummaryResponse, error) {
	// TODO: 从 context 获取 tenantID
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	summary, err := a.approvalService.GetInstanceStatsSummary(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return &approvalv1.InstanceStatsSummaryResponse{
		Total:     int32(summary.Total),
		Pending:   int32(summary.Pending),
		Approved:  int32(summary.Approved),
		Rejected:  int32(summary.Rejected),
		Withdrawn: int32(summary.Withdrawn),
		Cancelled: int32(summary.Cancelled),
		ByStatus:  convertByStatus(summary.ByStatus),
	}, nil
}

// ========== ApprovalTaskService 实现 ==========

func (a *ApprovalAdapter) GetApprovalTask(ctx context.Context, req *approvalv1.GetApprovalTaskRequest) (*approvalv1.ApprovalTaskResponse, error) {
	id, _ := uuid.Parse(req.Id)

	task, err := a.approvalService.GetApprovalTask(ctx, id)
	if err != nil {
		return nil, err
	}

	return toApprovalTaskResponse(task), nil
}

func (a *ApprovalAdapter) ListMyTasks(ctx context.Context, req *approvalv1.ListMyTasksRequest) (*approvalv1.ListApprovalTasksResponse, error) {
	// TODO: 从 context 获取 assigneeID
	assigneeID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	var status *model.TaskStatus
	if req.Status != "" {
		s := model.TaskStatus(req.Status)
		status = &s
	}

	tasks, err := a.approvalService.ListMyTasks(ctx, assigneeID, status, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}

	items := make([]*approvalv1.ApprovalTaskResponse, len(tasks))
	for i, task := range tasks {
		items[i] = toApprovalTaskResponse(task)
	}

	return &approvalv1.ListApprovalTasksResponse{
		Items: items,
		Total: int32(len(items)),
	}, nil
}

func (a *ApprovalAdapter) CountPendingTasks(ctx context.Context, req *approvalv1.CountPendingTasksRequest) (*approvalv1.CountPendingTasksResponse, error) {
	// TODO: 从 context 获取 assigneeID
	assigneeID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	count, err := a.approvalService.CountPendingTasks(ctx, assigneeID)
	if err != nil {
		return nil, err
	}

	return &approvalv1.CountPendingTasksResponse{
		Count: int32(count),
	}, nil
}

func (a *ApprovalAdapter) ProcessTask(ctx context.Context, req *approvalv1.ProcessTaskRequest) (*emptypb.Empty, error) {
	taskID, _ := uuid.Parse(req.Id)
	// TODO: 从 context 获取 operatorID
	operatorID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	action := model.ApprovalAction(req.Action)

	var comment *string
	if req.Comment != "" {
		comment = &req.Comment
	}

	processReq := &dto.ProcessTaskRequest{
		TaskID:     taskID,
		OperatorID: operatorID,
		Action:     action,
		Comment:    comment,
	}

	err := a.approvalService.ProcessTask(ctx, processReq)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (a *ApprovalAdapter) BatchProcessTasks(ctx context.Context, req *approvalv1.BatchProcessTasksRequest) (*approvalv1.BatchProcessTasksResponse, error) {
	// TODO: 从 context 获取 operatorID
	operatorID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	taskIDs := make([]uuid.UUID, len(req.TaskIds))
	for i, idStr := range req.TaskIds {
		taskIDs[i], _ = uuid.Parse(idStr)
	}

	action := model.ApprovalAction(req.Action)

	var comment *string
	if req.Comment != "" {
		comment = &req.Comment
	}

	results, err := a.approvalService.BatchProcessTasks(ctx, taskIDs, operatorID, action, comment)
	if err != nil {
		return nil, err
	}

	protoResults := make([]*approvalv1.BatchProcessResult, len(results))
	for i, r := range results {
		errMsg := ""
		if r.Error != nil {
			errMsg = *r.Error
		}
		protoResults[i] = &approvalv1.BatchProcessResult{
			TaskId:  r.TaskID.String(),
			Success: r.Success,
			Error:   errMsg,
		}
	}

	return &approvalv1.BatchProcessTasksResponse{
		Results: protoResults,
	}, nil
}

func (a *ApprovalAdapter) TransferTask(ctx context.Context, req *approvalv1.TransferTaskRequest) (*emptypb.Empty, error) {
	taskID, _ := uuid.Parse(req.Id)
	toUserID, _ := uuid.Parse(req.TransferToId)
	// TODO: 从 context 获取 fromUserID
	fromUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	var comment *string
	if req.Comment != "" {
		comment = &req.Comment
	}

	err := a.approvalService.TransferTask(ctx, taskID, fromUserID, toUserID, comment)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (a *ApprovalAdapter) DelegateTask(ctx context.Context, req *approvalv1.DelegateTaskRequest) (*emptypb.Empty, error) {
	taskID, _ := uuid.Parse(req.Id)
	toUserID, _ := uuid.Parse(req.DelegateToId)
	// TODO: 从 context 获取 fromUserID
	fromUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	var comment *string
	if req.Comment != "" {
		comment = &req.Comment
	}

	err := a.approvalService.DelegateTask(ctx, taskID, fromUserID, toUserID, comment)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ========== 辅助转换函数 ==========

func toProcessDefinitionResponse(dto *dto.ProcessDefResponse) *approvalv1.ProcessDefinitionResponse {
	return &approvalv1.ProcessDefinitionResponse{
		Id:           dto.ID.String(),
		TenantId:     dto.TenantID.String(),
		Code:         dto.Code,
		Name:         dto.Name,
		FormId:       dto.FormID.String(),
		FormName:     dto.FormName,
		WorkflowId:   dto.WorkflowID.String(),
		WorkflowName: dto.WorkflowName,
		Enabled:      dto.Enabled,
		CreatedAt:    dto.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    dto.UpdatedAt.Format(time.RFC3339),
	}
}

func toProcessInstanceResponse(dto *dto.ProcessInstanceResponse) *approvalv1.ProcessInstanceResponse {
	resp := &approvalv1.ProcessInstanceResponse{
		Id:                 dto.ID.String(),
		TenantId:           dto.TenantID.String(),
		ProcessDefId:       dto.ProcessDefID.String(),
		ProcessDefCode:     dto.ProcessDefCode,
		ProcessDefName:     dto.ProcessDefName,
		WorkflowInstanceId: dto.WorkflowInstanceID.String(),
		FormDataId:         dto.FormDataID.String(),
		ApplicantId:        dto.ApplicantID.String(),
		ApplicantName:      dto.ApplicantName,
		Title:              dto.Title,
		Status:             string(dto.Status),
		StartedAt:          dto.StartedAt.Format(time.RFC3339),
		CreatedAt:          dto.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          dto.UpdatedAt.Format(time.RFC3339),
	}

	if dto.CurrentNodeID != nil {
		resp.CurrentNodeId = *dto.CurrentNodeID
	}
	if dto.CurrentNodeName != nil {
		resp.CurrentNodeName = *dto.CurrentNodeName
	}
	if dto.CompletedAt != nil {
		resp.CompletedAt = dto.CompletedAt.Format(time.RFC3339)
	}

	return resp
}

func toApprovalTaskResponse(dto *dto.ApprovalTaskResponse) *approvalv1.ApprovalTaskResponse {
	resp := &approvalv1.ApprovalTaskResponse{
		Id:                dto.ID.String(),
		TenantId:          dto.TenantID.String(),
		ProcessInstanceId: dto.ProcessInstanceID.String(),
		NodeId:            dto.NodeID,
		NodeName:          dto.NodeName,
		AssigneeId:        dto.AssigneeID.String(),
		AssigneeName:      dto.AssigneeName,
		Status:            string(dto.Status),
		CreatedAt:         dto.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         dto.UpdatedAt.Format(time.RFC3339),
	}

	if dto.Action != nil {
		resp.Action = string(*dto.Action)
	}
	if dto.Comment != nil {
		resp.Comment = *dto.Comment
	}
	if dto.ApprovedAt != nil {
		resp.ApprovedAt = dto.ApprovedAt.Format(time.RFC3339)
	}

	return resp
}

func convertByStatus(byStatus map[string]int) map[string]int32 {
	result := make(map[string]int32)
	for k, v := range byStatus {
		result[k] = int32(v)
	}
	return result
}
