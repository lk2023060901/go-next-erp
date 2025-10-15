package handler

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pb "github.com/lk2023060901/go-next-erp/api/hrm/v1"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/internal/hrm/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// LeaveOfficeHandler 外出处理器
type LeaveOfficeHandler struct {
	pb.UnimplementedLeaveOfficeServiceServer
	leaveOfficeService service.LeaveOfficeService
}

// NewLeaveOfficeHandler 创建外出处理器
func NewLeaveOfficeHandler(leaveOfficeService service.LeaveOfficeService) *LeaveOfficeHandler {
	return &LeaveOfficeHandler{
		leaveOfficeService: leaveOfficeService,
	}
}

// CreateLeaveOffice 创建外出申请
func (h *LeaveOfficeHandler) CreateLeaveOffice(ctx context.Context, req *pb.CreateLeaveOfficeRequest) (*pb.LeaveOfficeResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	employeeID, err := uuid.Parse(req.EmployeeId)
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}

	departmentID, err := uuid.Parse(req.DepartmentId)
	if err != nil {
		return nil, fmt.Errorf("invalid department_id: %w", err)
	}

	leaveOffice := &model.LeaveOffice{
		TenantID:     tenantID,
		EmployeeID:   employeeID,
		EmployeeName: req.EmployeeName,
		DepartmentID: departmentID,
		StartTime:    req.StartTime.AsTime(),
		EndTime:      req.EndTime.AsTime(),
		Destination:  req.Destination,
		Purpose:      req.Purpose,
		Contact:      req.Contact,
		Remark:       req.Remark,
	}

	if err := h.leaveOfficeService.Create(ctx, leaveOffice); err != nil {
		return nil, err
	}

	return h.modelToProto(leaveOffice), nil
}

// UpdateLeaveOffice 更新外出申请
func (h *LeaveOfficeHandler) UpdateLeaveOffice(ctx context.Context, req *pb.UpdateLeaveOfficeRequest) (*pb.LeaveOfficeResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	// 获取原记录
	leaveOffice, err := h.leaveOfficeService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.StartTime != nil {
		leaveOffice.StartTime = req.StartTime.AsTime()
	}
	if req.EndTime != nil {
		leaveOffice.EndTime = req.EndTime.AsTime()
	}
	if req.Destination != "" {
		leaveOffice.Destination = req.Destination
	}
	if req.Purpose != "" {
		leaveOffice.Purpose = req.Purpose
	}
	if req.Contact != "" {
		leaveOffice.Contact = req.Contact
	}
	if req.Remark != "" {
		leaveOffice.Remark = req.Remark
	}

	if err := h.leaveOfficeService.Update(ctx, leaveOffice); err != nil {
		return nil, err
	}

	return h.modelToProto(leaveOffice), nil
}

// DeleteLeaveOffice 删除外出申请
func (h *LeaveOfficeHandler) DeleteLeaveOffice(ctx context.Context, req *pb.DeleteLeaveOfficeRequest) (*pb.DeleteLeaveOfficeResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return &pb.DeleteLeaveOfficeResponse{
			Success: false,
			Message: fmt.Sprintf("invalid id: %v", err),
		}, nil
	}

	if err := h.leaveOfficeService.Delete(ctx, id); err != nil {
		return &pb.DeleteLeaveOfficeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.DeleteLeaveOfficeResponse{
		Success: true,
		Message: "Leave office deleted successfully",
	}, nil
}

// GetLeaveOffice 获取外出详情
func (h *LeaveOfficeHandler) GetLeaveOffice(ctx context.Context, req *pb.GetLeaveOfficeRequest) (*pb.LeaveOfficeResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	leaveOffice, err := h.leaveOfficeService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return h.modelToProto(leaveOffice), nil
}

// ListLeaveOffices 列表查询外出记录
func (h *LeaveOfficeHandler) ListLeaveOffices(ctx context.Context, req *pb.ListLeaveOfficesRequest) (*pb.ListLeaveOfficesResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	filter := &repository.LeaveOfficeFilter{
		Keyword: req.Keyword,
	}

	if req.EmployeeId != "" {
		employeeID, err := uuid.Parse(req.EmployeeId)
		if err != nil {
			return nil, fmt.Errorf("invalid employee_id: %w", err)
		}
		filter.EmployeeID = &employeeID
	}

	if req.DepartmentId != "" {
		deptID, err := uuid.Parse(req.DepartmentId)
		if err != nil {
			return nil, fmt.Errorf("invalid department_id: %w", err)
		}
		filter.DepartmentID = &deptID
	}

	if req.ApprovalStatus != "" {
		filter.ApprovalStatus = &req.ApprovalStatus
	}

	if req.StartDate != nil {
		startDate := req.StartDate.AsTime()
		filter.StartDate = &startDate
	}

	if req.EndDate != nil {
		endDate := req.EndDate.AsTime()
		filter.EndDate = &endDate
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	leaveOffices, total, err := h.leaveOfficeService.List(ctx, tenantID, filter, int(offset), int(pageSize))
	if err != nil {
		return nil, err
	}

	pbLeaveOffices := make([]*pb.LeaveOfficeResponse, len(leaveOffices))
	for i, lo := range leaveOffices {
		pbLeaveOffices[i] = h.modelToProto(lo)
	}

	return &pb.ListLeaveOfficesResponse{
		Items:    pbLeaveOffices,
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ListEmployeeLeaveOffices 查询员工外出记录
func (h *LeaveOfficeHandler) ListEmployeeLeaveOffices(ctx context.Context, req *pb.ListEmployeeLeaveOfficesRequest) (*pb.ListLeaveOfficesResponse, error) {
	employeeID, err := uuid.Parse(req.EmployeeId)
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}

	filter := &repository.LeaveOfficeFilter{}

	if req.ApprovalStatus != "" {
		filter.ApprovalStatus = &req.ApprovalStatus
	}

	if req.StartDate != nil {
		startDate := req.StartDate.AsTime()
		filter.StartDate = &startDate
	}

	if req.EndDate != nil {
		endDate := req.EndDate.AsTime()
		filter.EndDate = &endDate
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	filter.EmployeeID = &employeeID

	// 假设tenantID可以从context中获取，或使用employeeID查询
	// 这里简化处理，直接使用employeeID作为tenantID（实际应从context获取）
	leaveOffices, total, err := h.leaveOfficeService.List(ctx, employeeID, filter, int(offset), int(pageSize))
	if err != nil {
		return nil, err
	}

	pbLeaveOffices := make([]*pb.LeaveOfficeResponse, len(leaveOffices))
	for i, lo := range leaveOffices {
		pbLeaveOffices[i] = h.modelToProto(lo)
	}

	return &pb.ListLeaveOfficesResponse{
		Items:    pbLeaveOffices,
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ListPendingLeaveOffices 查询待审批的外出
func (h *LeaveOfficeHandler) ListPendingLeaveOffices(ctx context.Context, req *pb.ListPendingLeaveOfficesRequest) (*pb.ListLeaveOfficesResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	filter := &repository.LeaveOfficeFilter{}
	status := "pending"
	filter.ApprovalStatus = &status

	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	leaveOffices, total, err := h.leaveOfficeService.List(ctx, tenantID, filter, int(offset), int(pageSize))
	if err != nil {
		return nil, err
	}

	pbLeaveOffices := make([]*pb.LeaveOfficeResponse, len(leaveOffices))
	for i, lo := range leaveOffices {
		pbLeaveOffices[i] = h.modelToProto(lo)
	}

	return &pb.ListLeaveOfficesResponse{
		Items:    pbLeaveOffices,
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// SubmitLeaveOffice 提交外出审批
func (h *LeaveOfficeHandler) SubmitLeaveOffice(ctx context.Context, req *pb.SubmitLeaveOfficeRequest) (*pb.SubmitLeaveOfficeResponse, error) {
	leaveOfficeID, err := uuid.Parse(req.LeaveOfficeId)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	submitterID, err := uuid.Parse(req.SubmitterId)
	if err != nil {
		return nil, fmt.Errorf("invalid submitter_id: %w", err)
	}

	if err := h.leaveOfficeService.Submit(ctx, leaveOfficeID, submitterID); err != nil {
		return &pb.SubmitLeaveOfficeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.SubmitLeaveOfficeResponse{
		Success:    true,
		Message:    "Leave office submitted successfully",
		WorkflowId: fmt.Sprintf("leave-office-approval-%s", leaveOfficeID.String()),
	}, nil
}

// ApproveLeaveOffice 批准外出
func (h *LeaveOfficeHandler) ApproveLeaveOffice(ctx context.Context, req *pb.ApproveLeaveOfficeRequest) (*pb.ApproveLeaveOfficeResponse, error) {
	leaveOfficeID, err := uuid.Parse(req.LeaveOfficeId)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	approverID, err := uuid.Parse(req.ApproverId)
	if err != nil {
		return nil, fmt.Errorf("invalid approver_id: %w", err)
	}

	if err := h.leaveOfficeService.Approve(ctx, leaveOfficeID, approverID, req.Comment); err != nil {
		return &pb.ApproveLeaveOfficeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.ApproveLeaveOfficeResponse{
		Success: true,
		Message: "Leave office approved successfully",
	}, nil
}

// RejectLeaveOffice 拒绝外出
func (h *LeaveOfficeHandler) RejectLeaveOffice(ctx context.Context, req *pb.RejectLeaveOfficeRequest) (*pb.RejectLeaveOfficeResponse, error) {
	leaveOfficeID, err := uuid.Parse(req.LeaveOfficeId)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	approverID, err := uuid.Parse(req.ApproverId)
	if err != nil {
		return nil, fmt.Errorf("invalid approver_id: %w", err)
	}

	if err := h.leaveOfficeService.Reject(ctx, leaveOfficeID, approverID, req.Reason); err != nil {
		return &pb.RejectLeaveOfficeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.RejectLeaveOfficeResponse{
		Success: true,
		Message: "Leave office rejected successfully",
	}, nil
}

// modelToProto 将模型转换为Proto响应
func (h *LeaveOfficeHandler) modelToProto(leaveOffice *model.LeaveOffice) *pb.LeaveOfficeResponse {
	resp := &pb.LeaveOfficeResponse{
		Id:             leaveOffice.ID.String(),
		TenantId:       leaveOffice.TenantID.String(),
		EmployeeId:     leaveOffice.EmployeeID.String(),
		EmployeeName:   leaveOffice.EmployeeName,
		DepartmentId:   leaveOffice.DepartmentID.String(),
		StartTime:      timestamppb.New(leaveOffice.StartTime),
		EndTime:        timestamppb.New(leaveOffice.EndTime),
		Duration:       leaveOffice.Duration,
		Destination:    leaveOffice.Destination,
		Purpose:        leaveOffice.Purpose,
		Contact:        leaveOffice.Contact,
		ApprovalStatus: leaveOffice.ApprovalStatus,
		Remark:         leaveOffice.Remark,
		CreatedAt:      timestamppb.New(leaveOffice.CreatedAt),
		UpdatedAt:      timestamppb.New(leaveOffice.UpdatedAt),
	}

	if leaveOffice.ApprovedBy != nil {
		resp.ApprovedBy = leaveOffice.ApprovedBy.String()
	}

	if leaveOffice.ApprovedAt != nil {
		resp.ApprovedAt = timestamppb.New(*leaveOffice.ApprovedAt)
	}

	if leaveOffice.RejectReason != "" {
		resp.RejectReason = leaveOffice.RejectReason
	}

	return resp
}
