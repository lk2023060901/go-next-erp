package handler

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/lk2023060901/go-next-erp/api/hrm/v1"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/internal/hrm/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// LeaveHandler 请假处理器
type LeaveHandler struct {
	pb.UnimplementedLeaveTypeServiceServer
	pb.UnimplementedLeaveRequestServiceServer
	pb.UnimplementedLeaveQuotaServiceServer
	leaveService service.LeaveService
}

// NewLeaveHandler 创建请假处理器
func NewLeaveHandler(leaveService service.LeaveService) *LeaveHandler {
	return &LeaveHandler{
		leaveService: leaveService,
	}
}

// ==================== 请假类型管理 ====================

// CreateLeaveType 创建请假类型
func (h *LeaveHandler) CreateLeaveType(ctx context.Context, req *pb.CreateLeaveTypeRequest) (*pb.LeaveTypeResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)

	leaveType := &model.LeaveType{
		TenantID:         tenantID,
		Code:             req.Code,
		Name:             req.Name,
		Description:      req.Description,
		IsPaid:           req.IsPaid,
		RequiresApproval: req.RequiresApproval,
		RequiresProof:    req.RequiresProof,
		DeductQuota:      req.DeductQuota,
		Unit:             model.LeaveUnit(req.Unit),
		MinDuration:      req.MinDuration,
		AdvanceDays:      int(req.AdvanceDays),
		Color:            req.Color,
		IsActive:         true,
		Sort:             int(req.Sort),
	}

	if req.MaxDuration > 0 {
		maxDur := req.MaxDuration
		leaveType.MaxDuration = &maxDur
	}

	// 转换审批规则
	if req.ApprovalRules != nil {
		leaveType.ApprovalRules = h.toModelApprovalRules(req.ApprovalRules)
	}

	if err := h.leaveService.CreateLeaveType(ctx, leaveType); err != nil {
		return nil, err
	}

	return h.toLeaveTypeResponse(leaveType), nil
}

// UpdateLeaveType 更新请假类型
func (h *LeaveHandler) UpdateLeaveType(ctx context.Context, req *pb.UpdateLeaveTypeRequest) (*pb.LeaveTypeResponse, error) {
	id, _ := uuid.Parse(req.Id)

	leaveType, err := h.leaveService.GetLeaveType(ctx, id)
	if err != nil {
		return nil, err
	}

	leaveType.Name = req.Name
	leaveType.Description = req.Description
	leaveType.IsPaid = req.IsPaid
	leaveType.RequiresApproval = req.RequiresApproval
	leaveType.RequiresProof = req.RequiresProof
	leaveType.DeductQuota = req.DeductQuota
	leaveType.Unit = model.LeaveUnit(req.Unit)
	leaveType.MinDuration = req.MinDuration
	leaveType.AdvanceDays = int(req.AdvanceDays)
	leaveType.Color = req.Color
	leaveType.IsActive = req.IsActive
	leaveType.Sort = int(req.Sort)

	if req.MaxDuration > 0 {
		maxDur := req.MaxDuration
		leaveType.MaxDuration = &maxDur
	}

	// 转换审批规则
	if req.ApprovalRules != nil {
		leaveType.ApprovalRules = h.toModelApprovalRules(req.ApprovalRules)
	}

	if err := h.leaveService.UpdateLeaveType(ctx, leaveType); err != nil {
		return nil, err
	}

	return h.toLeaveTypeResponse(leaveType), nil
}

// DeleteLeaveType 删除请假类型
func (h *LeaveHandler) DeleteLeaveType(ctx context.Context, req *pb.DeleteLeaveTypeRequest) (*pb.DeleteLeaveTypeResponse, error) {
	id, _ := uuid.Parse(req.Id)

	if err := h.leaveService.DeleteLeaveType(ctx, id); err != nil {
		return &pb.DeleteLeaveTypeResponse{Success: false}, err
	}

	return &pb.DeleteLeaveTypeResponse{Success: true}, nil
}

// GetLeaveType 获取请假类型详情
func (h *LeaveHandler) GetLeaveType(ctx context.Context, req *pb.GetLeaveTypeRequest) (*pb.LeaveTypeResponse, error) {
	id, _ := uuid.Parse(req.Id)

	leaveType, err := h.leaveService.GetLeaveType(ctx, id)
	if err != nil {
		return nil, err
	}

	return h.toLeaveTypeResponse(leaveType), nil
}

// ListLeaveTypes 列表查询请假类型
func (h *LeaveHandler) ListLeaveTypes(ctx context.Context, req *pb.ListLeaveTypesRequest) (*pb.ListLeaveTypesResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)

	filter := &repository.LeaveTypeFilter{
		Keyword: req.Keyword,
	}

	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	leaveTypes, total, err := h.leaveService.ListLeaveTypes(ctx, tenantID, filter, offset, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*pb.LeaveTypeResponse, 0, len(leaveTypes))
	for _, lt := range leaveTypes {
		items = append(items, h.toLeaveTypeResponse(lt))
	}

	return &pb.ListLeaveTypesResponse{
		Items:    items,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// ListActiveLeaveTypes 获取启用的请假类型
func (h *LeaveHandler) ListActiveLeaveTypes(ctx context.Context, req *pb.ListActiveLeaveTypesRequest) (*pb.ListActiveLeaveTypesResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)

	leaveTypes, err := h.leaveService.ListActiveLeaveTypes(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	items := make([]*pb.LeaveTypeResponse, 0, len(leaveTypes))
	for _, lt := range leaveTypes {
		items = append(items, h.toLeaveTypeResponse(lt))
	}

	return &pb.ListActiveLeaveTypesResponse{Items: items}, nil
}

// ==================== 请假申请管理 ====================

// CreateLeaveRequest 创建请假申请
func (h *LeaveHandler) CreateLeaveRequest(ctx context.Context, req *pb.CreateLeaveRequestRequest) (*pb.LeaveRequestResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)
	employeeID, _ := uuid.Parse(req.EmployeeId)
	leaveTypeID, _ := uuid.Parse(req.LeaveTypeId)

	var deptID *uuid.UUID
	if req.DepartmentId != "" {
		id, _ := uuid.Parse(req.DepartmentId)
		deptID = &id
	}

	request := &model.LeaveRequest{
		TenantID:     tenantID,
		EmployeeID:   employeeID,
		EmployeeName: req.EmployeeName,
		DepartmentID: deptID,
		LeaveTypeID:  leaveTypeID,
		StartTime:    req.StartTime.AsTime(),
		EndTime:      req.EndTime.AsTime(),
		Duration:     req.Duration,
		Reason:       req.Reason,
		ProofURLs:    req.ProofUrls,
	}

	if err := h.leaveService.CreateLeaveRequest(ctx, request); err != nil {
		return nil, err
	}

	return h.toLeaveRequestResponse(request), nil
}

// UpdateLeaveRequest 更新请假申请
func (h *LeaveHandler) UpdateLeaveRequest(ctx context.Context, req *pb.UpdateLeaveRequestRequest) (*pb.LeaveRequestResponse, error) {
	id, _ := uuid.Parse(req.Id)

	request, err := h.leaveService.GetLeaveRequest(ctx, id)
	if err != nil {
		return nil, err
	}

	request.StartTime = req.StartTime.AsTime()
	request.EndTime = req.EndTime.AsTime()
	request.Duration = req.Duration
	request.Reason = req.Reason
	request.ProofURLs = req.ProofUrls

	if err := h.leaveService.UpdateLeaveRequest(ctx, &request.LeaveRequest); err != nil {
		return nil, err
	}

	return h.toLeaveRequestResponse(&request.LeaveRequest), nil
}

// SubmitLeaveRequest 提交请假申请
func (h *LeaveHandler) SubmitLeaveRequest(ctx context.Context, req *pb.SubmitLeaveRequestRequest) (*pb.SubmitLeaveRequestResponse, error) {
	requestID, _ := uuid.Parse(req.RequestId)
	submitterID, _ := uuid.Parse(req.SubmitterId)

	if err := h.leaveService.SubmitLeaveRequest(ctx, requestID, submitterID); err != nil {
		return &pb.SubmitLeaveRequestResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.SubmitLeaveRequestResponse{
		Success: true,
		Message: "请假申请已提交",
	}, nil
}

// WithdrawLeaveRequest 撤回请假申请
func (h *LeaveHandler) WithdrawLeaveRequest(ctx context.Context, req *pb.WithdrawLeaveRequestRequest) (*pb.WithdrawLeaveRequestResponse, error) {
	requestID, _ := uuid.Parse(req.RequestId)
	operatorID, _ := uuid.Parse(req.OperatorId)

	if err := h.leaveService.WithdrawLeaveRequest(ctx, requestID, operatorID); err != nil {
		return &pb.WithdrawLeaveRequestResponse{Success: false}, err
	}

	return &pb.WithdrawLeaveRequestResponse{Success: true}, nil
}

// CancelLeaveRequest 取消请假
func (h *LeaveHandler) CancelLeaveRequest(ctx context.Context, req *pb.CancelLeaveRequestRequest) (*pb.CancelLeaveRequestResponse, error) {
	requestID, _ := uuid.Parse(req.RequestId)
	operatorID, _ := uuid.Parse(req.OperatorId)

	if err := h.leaveService.CancelLeaveRequest(ctx, requestID, operatorID, req.Reason); err != nil {
		return &pb.CancelLeaveRequestResponse{Success: false}, err
	}

	return &pb.CancelLeaveRequestResponse{Success: true}, nil
}

// GetLeaveRequest 获取请假详情
func (h *LeaveHandler) GetLeaveRequest(ctx context.Context, req *pb.GetLeaveRequestRequest) (*pb.LeaveRequestDetailResponse, error) {
	requestID, _ := uuid.Parse(req.RequestId)

	request, err := h.leaveService.GetLeaveRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}

	approvals := make([]*pb.ApprovalResponse, 0, len(request.Approvals))
	for _, ap := range request.Approvals {
		approvals = append(approvals, h.toApprovalResponse(ap))
	}

	return &pb.LeaveRequestDetailResponse{
		Request:   h.toLeaveRequestResponse(&request.LeaveRequest),
		Approvals: approvals,
	}, nil
}

// ListMyLeaveRequests 查询我的请假记录
func (h *LeaveHandler) ListMyLeaveRequests(ctx context.Context, req *pb.ListMyLeaveRequestsRequest) (*pb.ListLeaveRequestsResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)
	employeeID, _ := uuid.Parse(req.EmployeeId)

	filter := &repository.LeaveRequestFilter{}
	if req.LeaveTypeId != "" {
		ltID, _ := uuid.Parse(req.LeaveTypeId)
		filter.LeaveTypeID = &ltID
	}
	if req.Status != "" {
		status := model.LeaveRequestStatus(req.Status)
		filter.Status = &status
	}
	if req.StartDate != nil {
		startDate := req.StartDate.AsTime()
		filter.StartDate = &startDate
	}
	if req.EndDate != nil {
		endDate := req.EndDate.AsTime()
		filter.EndDate = &endDate
	}

	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	requests, total, err := h.leaveService.ListMyLeaveRequests(ctx, tenantID, employeeID, filter, offset, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*pb.LeaveRequestResponse, 0, len(requests))
	for _, r := range requests {
		items = append(items, h.toLeaveRequestResponse(r))
	}

	return &pb.ListLeaveRequestsResponse{
		Items:    items,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// ListLeaveRequests 查询请假记录（管理员）
func (h *LeaveHandler) ListLeaveRequests(ctx context.Context, req *pb.ListLeaveRequestsRequest) (*pb.ListLeaveRequestsResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)

	filter := &repository.LeaveRequestFilter{
		Keyword: req.Keyword,
	}
	if req.LeaveTypeId != "" {
		ltID, _ := uuid.Parse(req.LeaveTypeId)
		filter.LeaveTypeID = &ltID
	}
	if req.DepartmentId != "" {
		deptID, _ := uuid.Parse(req.DepartmentId)
		filter.DepartmentID = &deptID
	}
	if req.Status != "" {
		status := model.LeaveRequestStatus(req.Status)
		filter.Status = &status
	}
	if req.StartDate != nil {
		startDate := req.StartDate.AsTime()
		filter.StartDate = &startDate
	}
	if req.EndDate != nil {
		endDate := req.EndDate.AsTime()
		filter.EndDate = &endDate
	}

	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	requests, total, err := h.leaveService.ListLeaveRequests(ctx, tenantID, filter, offset, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*pb.LeaveRequestResponse, 0, len(requests))
	for _, r := range requests {
		items = append(items, h.toLeaveRequestResponse(r))
	}

	return &pb.ListLeaveRequestsResponse{
		Items:    items,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// ListPendingApprovals 查询待我审批的请假
func (h *LeaveHandler) ListPendingApprovals(ctx context.Context, req *pb.ListPendingApprovalsRequest) (*pb.ListLeaveRequestsResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)
	approverID, _ := uuid.Parse(req.ApproverId)

	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	requests, total, err := h.leaveService.ListPendingApprovals(ctx, tenantID, approverID, offset, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*pb.LeaveRequestResponse, 0, len(requests))
	for _, r := range requests {
		items = append(items, h.toLeaveRequestResponse(r))
	}

	return &pb.ListLeaveRequestsResponse{
		Items:    items,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// ApproveLeaveRequest 批准请假
func (h *LeaveHandler) ApproveLeaveRequest(ctx context.Context, req *pb.ApproveLeaveRequestRequest) (*pb.ApproveLeaveRequestResponse, error) {
	requestID, _ := uuid.Parse(req.RequestId)
	approverID, _ := uuid.Parse(req.ApproverId)

	if err := h.leaveService.ApproveLeaveRequest(ctx, requestID, approverID, req.Comment); err != nil {
		return &pb.ApproveLeaveRequestResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.ApproveLeaveRequestResponse{
		Success: true,
		Message: "已批准请假申请",
	}, nil
}

// RejectLeaveRequest 拒绝请假
func (h *LeaveHandler) RejectLeaveRequest(ctx context.Context, req *pb.RejectLeaveRequestRequest) (*pb.RejectLeaveRequestResponse, error) {
	requestID, _ := uuid.Parse(req.RequestId)
	approverID, _ := uuid.Parse(req.ApproverId)

	if err := h.leaveService.RejectLeaveRequest(ctx, requestID, approverID, req.Comment); err != nil {
		return &pb.RejectLeaveRequestResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.RejectLeaveRequestResponse{
		Success: true,
		Message: "已拒绝请假申请",
	}, nil
}

// ==================== 请假额度管理 ====================

// InitEmployeeQuota 初始化员工额度
func (h *LeaveHandler) InitEmployeeQuota(ctx context.Context, req *pb.InitEmployeeQuotaRequest) (*pb.InitEmployeeQuotaResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)
	employeeID, _ := uuid.Parse(req.EmployeeId)

	if err := h.leaveService.InitEmployeeQuota(ctx, tenantID, employeeID, int(req.Year)); err != nil {
		return &pb.InitEmployeeQuotaResponse{Success: false}, err
	}

	return &pb.InitEmployeeQuotaResponse{Success: true}, nil
}

// UpdateQuota 更新额度
func (h *LeaveHandler) UpdateQuota(ctx context.Context, req *pb.UpdateQuotaRequest) (*pb.QuotaResponse, error) {
	id, _ := uuid.Parse(req.Id)

	quota := &model.LeaveQuota{
		ID:         id,
		TotalQuota: req.TotalQuota,
	}

	if err := h.leaveService.UpdateQuota(ctx, quota); err != nil {
		return nil, err
	}

	return h.toQuotaResponse(&model.LeaveQuotaWithType{LeaveQuota: *quota}), nil
}

// GetEmployeeQuotas 获取员工额度
func (h *LeaveHandler) GetEmployeeQuotas(ctx context.Context, req *pb.GetEmployeeQuotasRequest) (*pb.GetEmployeeQuotasResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)
	employeeID, _ := uuid.Parse(req.EmployeeId)

	quotas, err := h.leaveService.GetEmployeeQuotas(ctx, tenantID, employeeID, int(req.Year))
	if err != nil {
		return nil, err
	}

	items := make([]*pb.QuotaResponse, 0, len(quotas))
	for _, q := range quotas {
		items = append(items, h.toQuotaResponse(q))
	}

	return &pb.GetEmployeeQuotasResponse{Items: items}, nil
}

// ==================== 辅助方法 ====================

func (h *LeaveHandler) toLeaveTypeResponse(lt *model.LeaveType) *pb.LeaveTypeResponse {
	resp := &pb.LeaveTypeResponse{
		Id:               lt.ID.String(),
		TenantId:         lt.TenantID.String(),
		Code:             lt.Code,
		Name:             lt.Name,
		Description:      lt.Description,
		IsPaid:           lt.IsPaid,
		RequiresApproval: lt.RequiresApproval,
		RequiresProof:    lt.RequiresProof,
		DeductQuota:      lt.DeductQuota,
		Unit:             string(lt.Unit),
		MinDuration:      lt.MinDuration,
		AdvanceDays:      int32(lt.AdvanceDays),
		Color:            lt.Color,
		IsActive:         lt.IsActive,
		Sort:             int32(lt.Sort),
		CreatedAt:        timestamppb.New(lt.CreatedAt),
	}

	if lt.MaxDuration != nil {
		resp.MaxDuration = *lt.MaxDuration
	}

	// 转换审批规则
	if lt.ApprovalRules != nil {
		resp.ApprovalRules = h.toPbApprovalRules(lt.ApprovalRules)
	}

	return resp
}

func (h *LeaveHandler) toLeaveRequestResponse(r *model.LeaveRequest) *pb.LeaveRequestResponse {
	resp := &pb.LeaveRequestResponse{
		Id:            r.ID.String(),
		TenantId:      r.TenantID.String(),
		EmployeeId:    r.EmployeeID.String(),
		EmployeeName:  r.EmployeeName,
		LeaveTypeId:   r.LeaveTypeID.String(),
		LeaveTypeName: r.LeaveTypeName,
		StartTime:     timestamppb.New(r.StartTime),
		EndTime:       timestamppb.New(r.EndTime),
		Duration:      r.Duration,
		Unit:          string(r.Unit),
		Reason:        r.Reason,
		ProofUrls:     r.ProofURLs,
		Status:        string(r.Status),
		CreatedAt:     timestamppb.New(r.CreatedAt),
	}

	if r.DepartmentID != nil {
		resp.DepartmentId = r.DepartmentID.String()
	}
	if r.CurrentApproverID != nil {
		resp.CurrentApproverId = r.CurrentApproverID.String()
	}
	if r.SubmittedAt != nil {
		resp.SubmittedAt = timestamppb.New(*r.SubmittedAt)
	}

	return resp
}

func (h *LeaveHandler) toApprovalResponse(a *model.LeaveApproval) *pb.ApprovalResponse {
	resp := &pb.ApprovalResponse{
		Id:             a.ID.String(),
		LeaveRequestId: a.LeaveRequestID.String(),
		ApproverId:     a.ApproverID.String(),
		ApproverName:   a.ApproverName,
		Level:          int32(a.Level),
		Status:         string(a.Status),
		Comment:        a.Comment,
		CreatedAt:      timestamppb.New(a.CreatedAt),
	}

	if a.Action != nil {
		resp.Action = string(*a.Action)
	}
	if a.ApprovedAt != nil {
		resp.ApprovedAt = timestamppb.New(*a.ApprovedAt)
	}

	return resp
}

func (h *LeaveHandler) toQuotaResponse(q *model.LeaveQuotaWithType) *pb.QuotaResponse {
	resp := &pb.QuotaResponse{
		Id:             q.ID.String(),
		EmployeeId:     q.EmployeeID.String(),
		LeaveTypeId:    q.LeaveTypeID.String(),
		Year:           int32(q.Year),
		TotalQuota:     q.TotalQuota,
		UsedQuota:      q.UsedQuota,
		PendingQuota:   q.PendingQuota,
		RemainingQuota: q.RemainingQuota(),
	}

	if q.LeaveType != nil {
		resp.LeaveTypeName = q.LeaveType.Name
		resp.LeaveTypeCode = q.LeaveType.Code
		resp.Unit = string(q.LeaveType.Unit)
		resp.Color = q.LeaveType.Color
	}

	if q.ExpiredAt != nil {
		resp.ExpiredAt = timestamppb.New(*q.ExpiredAt)
	}

	return resp
}

// toModelApprovalRules 将 Proto 审批规则转换为 Model 审批规则
func (h *LeaveHandler) toModelApprovalRules(pbRules *pb.ApprovalRules) *model.ApprovalRules {
	if pbRules == nil {
		return nil
	}

	rules := &model.ApprovalRules{
		DefaultChain:  make([]*model.ApprovalNode, 0, len(pbRules.DefaultChain)),
		DurationRules: make([]*model.DurationRule, 0, len(pbRules.DurationRules)),
	}

	// 转换默认审批链
	for _, node := range pbRules.DefaultChain {
		modelNode := &model.ApprovalNode{
			Level:        int(node.Level),
			ApproverType: model.ApproverType(node.ApproverType),
			Required:     node.Required,
		}
		if node.ApproverId != "" {
			modelNode.ApproverID = &node.ApproverId
		}
		rules.DefaultChain = append(rules.DefaultChain, modelNode)
	}

	// 转换天数规则
	for _, rule := range pbRules.DurationRules {
		modelRule := &model.DurationRule{
			MinDuration:   rule.MinDuration,
			ApprovalChain: make([]*model.ApprovalNode, 0, len(rule.ApprovalChain)),
		}
		if rule.MaxDuration > 0 {
			modelRule.MaxDuration = &rule.MaxDuration
		}
		for _, node := range rule.ApprovalChain {
			modelNode := &model.ApprovalNode{
				Level:        int(node.Level),
				ApproverType: model.ApproverType(node.ApproverType),
				Required:     node.Required,
			}
			if node.ApproverId != "" {
				modelNode.ApproverID = &node.ApproverId
			}
			modelRule.ApprovalChain = append(modelRule.ApprovalChain, modelNode)
		}
		rules.DurationRules = append(rules.DurationRules, modelRule)
	}

	return rules
}

// toPbApprovalRules 将 Model 审批规则转换为 Proto 审批规则
func (h *LeaveHandler) toPbApprovalRules(modelRules *model.ApprovalRules) *pb.ApprovalRules {
	if modelRules == nil {
		return nil
	}

	rules := &pb.ApprovalRules{
		DefaultChain:  make([]*pb.ApprovalNode, 0, len(modelRules.DefaultChain)),
		DurationRules: make([]*pb.DurationRule, 0, len(modelRules.DurationRules)),
	}

	// 转换默认审批链
	for _, node := range modelRules.DefaultChain {
		pbNode := &pb.ApprovalNode{
			Level:        int32(node.Level),
			ApproverType: string(node.ApproverType),
			Required:     node.Required,
		}
		if node.ApproverID != nil {
			pbNode.ApproverId = *node.ApproverID
		}
		rules.DefaultChain = append(rules.DefaultChain, pbNode)
	}

	// 转换天数规则
	for _, rule := range modelRules.DurationRules {
		pbRule := &pb.DurationRule{
			MinDuration:   rule.MinDuration,
			ApprovalChain: make([]*pb.ApprovalNode, 0, len(rule.ApprovalChain)),
		}
		if rule.MaxDuration != nil {
			pbRule.MaxDuration = *rule.MaxDuration
		}
		for _, node := range rule.ApprovalChain {
			pbNode := &pb.ApprovalNode{
				Level:        int32(node.Level),
				ApproverType: string(node.ApproverType),
				Required:     node.Required,
			}
			if node.ApproverID != nil {
				pbNode.ApproverId = *node.ApproverID
			}
			pbRule.ApprovalChain = append(pbRule.ApprovalChain, pbNode)
		}
		rules.DurationRules = append(rules.DurationRules, pbRule)
	}

	return rules
}
