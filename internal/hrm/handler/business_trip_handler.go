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

// BusinessTripHandler 出差处理器
type BusinessTripHandler struct {
	pb.UnimplementedBusinessTripServiceServer
	tripService service.BusinessTripService
}

// NewBusinessTripHandler 创建出差处理器
func NewBusinessTripHandler(tripService service.BusinessTripService) *BusinessTripHandler {
	return &BusinessTripHandler{
		tripService: tripService,
	}
}

// CreateBusinessTrip 创建出差申请
func (h *BusinessTripHandler) CreateBusinessTrip(ctx context.Context, req *pb.CreateBusinessTripRequest) (*pb.BusinessTripResponse, error) {
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

	// 解析同行人员ID
	companions := make([]uuid.UUID, 0, len(req.Companions))
	for _, companionID := range req.Companions {
		id, err := uuid.Parse(companionID)
		if err != nil {
			return nil, fmt.Errorf("invalid companion_id: %w", err)
		}
		companions = append(companions, id)
	}

	trip := &model.BusinessTrip{
		TenantID:       tenantID,
		EmployeeID:     employeeID,
		EmployeeName:   req.EmployeeName,
		DepartmentID:   departmentID,
		StartTime:      req.StartTime.AsTime(),
		EndTime:        req.EndTime.AsTime(),
		Destination:    req.Destination,
		Transportation: req.Transportation,
		Accommodation:  req.Accommodation,
		Companions:     companions,
		Purpose:        req.Purpose,
		Tasks:          req.Tasks,
		EstimatedCost:  req.EstimatedCost,
		Remark:         req.Remark,
	}

	if err := h.tripService.Create(ctx, trip); err != nil {
		return nil, err
	}

	return h.modelToProto(trip), nil
}

// UpdateBusinessTrip 更新出差申请
func (h *BusinessTripHandler) UpdateBusinessTrip(ctx context.Context, req *pb.UpdateBusinessTripRequest) (*pb.BusinessTripResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	// 获取原记录
	trip, err := h.tripService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.StartTime != nil {
		trip.StartTime = req.StartTime.AsTime()
	}
	if req.EndTime != nil {
		trip.EndTime = req.EndTime.AsTime()
	}
	if req.Destination != "" {
		trip.Destination = req.Destination
	}
	if req.Transportation != "" {
		trip.Transportation = req.Transportation
	}
	if req.Accommodation != "" {
		trip.Accommodation = req.Accommodation
	}
	if req.Companions != nil {
		companions := make([]uuid.UUID, 0, len(req.Companions))
		for _, companionID := range req.Companions {
			id, err := uuid.Parse(companionID)
			if err != nil {
				return nil, fmt.Errorf("invalid companion_id: %w", err)
			}
			companions = append(companions, id)
		}
		trip.Companions = companions
	}
	if req.Purpose != "" {
		trip.Purpose = req.Purpose
	}
	if req.Tasks != "" {
		trip.Tasks = req.Tasks
	}
	if req.EstimatedCost > 0 {
		trip.EstimatedCost = req.EstimatedCost
	}
	if req.Remark != "" {
		trip.Remark = req.Remark
	}

	if err := h.tripService.Update(ctx, trip); err != nil {
		return nil, err
	}

	return h.modelToProto(trip), nil
}

// DeleteBusinessTrip 删除出差申请
func (h *BusinessTripHandler) DeleteBusinessTrip(ctx context.Context, req *pb.DeleteBusinessTripRequest) (*pb.DeleteBusinessTripResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return &pb.DeleteBusinessTripResponse{
			Success: false,
			Message: fmt.Sprintf("invalid id: %v", err),
		}, nil
	}

	if err := h.tripService.Delete(ctx, id); err != nil {
		return &pb.DeleteBusinessTripResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.DeleteBusinessTripResponse{
		Success: true,
		Message: "Business trip deleted successfully",
	}, nil
}

// GetBusinessTrip 获取出差详情
func (h *BusinessTripHandler) GetBusinessTrip(ctx context.Context, req *pb.GetBusinessTripRequest) (*pb.BusinessTripResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	trip, err := h.tripService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return h.modelToProto(trip), nil
}

// ListBusinessTrips 列表查询出差记录
func (h *BusinessTripHandler) ListBusinessTrips(ctx context.Context, req *pb.ListBusinessTripsRequest) (*pb.ListBusinessTripsResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	filter := &repository.BusinessTripFilter{
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

	trips, total, err := h.tripService.List(ctx, tenantID, filter, int(offset), int(pageSize))
	if err != nil {
		return nil, err
	}

	pbTrips := make([]*pb.BusinessTripResponse, len(trips))
	for i, trip := range trips {
		pbTrips[i] = h.modelToProto(trip)
	}

	return &pb.ListBusinessTripsResponse{
		Items:    pbTrips,
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// SubmitBusinessTrip 提交出差审批
func (h *BusinessTripHandler) SubmitBusinessTrip(ctx context.Context, req *pb.SubmitBusinessTripRequest) (*pb.SubmitBusinessTripResponse, error) {
	tripID, err := uuid.Parse(req.BusinessTripId)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	submitterID, err := uuid.Parse(req.SubmitterId)
	if err != nil {
		return nil, fmt.Errorf("invalid submitter_id: %w", err)
	}

	if err := h.tripService.Submit(ctx, tripID, submitterID); err != nil {
		return &pb.SubmitBusinessTripResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.SubmitBusinessTripResponse{
		Success:    true,
		Message:    "Business trip submitted successfully",
		WorkflowId: fmt.Sprintf("business-trip-approval-%s", tripID.String()),
	}, nil
}

// ApproveBusinessTrip 批准出差
func (h *BusinessTripHandler) ApproveBusinessTrip(ctx context.Context, req *pb.ApproveBusinessTripRequest) (*pb.ApproveBusinessTripResponse, error) {
	tripID, err := uuid.Parse(req.BusinessTripId)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	approverID, err := uuid.Parse(req.ApproverId)
	if err != nil {
		return nil, fmt.Errorf("invalid approver_id: %w", err)
	}

	if err := h.tripService.Approve(ctx, tripID, approverID, req.Comment); err != nil {
		return &pb.ApproveBusinessTripResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.ApproveBusinessTripResponse{
		Success: true,
		Message: "Business trip approved successfully",
	}, nil
}

// RejectBusinessTrip 拒绝出差
func (h *BusinessTripHandler) RejectBusinessTrip(ctx context.Context, req *pb.RejectBusinessTripRequest) (*pb.RejectBusinessTripResponse, error) {
	tripID, err := uuid.Parse(req.BusinessTripId)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	approverID, err := uuid.Parse(req.ApproverId)
	if err != nil {
		return nil, fmt.Errorf("invalid approver_id: %w", err)
	}

	if err := h.tripService.Reject(ctx, tripID, approverID, req.Reason); err != nil {
		return &pb.RejectBusinessTripResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.RejectBusinessTripResponse{
		Success: true,
		Message: "Business trip rejected successfully",
	}, nil
}

// SubmitTripReport 提交出差报告
func (h *BusinessTripHandler) SubmitTripReport(ctx context.Context, req *pb.SubmitTripReportRequest) (*pb.BusinessTripResponse, error) {
	tripID, err := uuid.Parse(req.BusinessTripId)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	if err := h.tripService.SubmitReport(ctx, tripID, req.Report, req.ActualCost); err != nil {
		return nil, err
	}

	trip, err := h.tripService.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	return h.modelToProto(trip), nil
}

// ListEmployeeBusinessTrips 查询员工出差记录
func (h *BusinessTripHandler) ListEmployeeBusinessTrips(ctx context.Context, req *pb.ListEmployeeBusinessTripsRequest) (*pb.ListBusinessTripsResponse, error) {
	employeeID, err := uuid.Parse(req.EmployeeId)
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}

	filter := &repository.BusinessTripFilter{}

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

	trips, total, err := h.tripService.List(ctx, employeeID, filter, int(offset), int(pageSize))
	if err != nil {
		return nil, err
	}

	pbTrips := make([]*pb.BusinessTripResponse, len(trips))
	for i, trip := range trips {
		pbTrips[i] = h.modelToProto(trip)
	}

	return &pb.ListBusinessTripsResponse{
		Items:    pbTrips,
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ListPendingBusinessTrips 查询待审批的出差
func (h *BusinessTripHandler) ListPendingBusinessTrips(ctx context.Context, req *pb.ListPendingBusinessTripsRequest) (*pb.ListBusinessTripsResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	filter := &repository.BusinessTripFilter{}
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

	trips, total, err := h.tripService.List(ctx, tenantID, filter, int(offset), int(pageSize))
	if err != nil {
		return nil, err
	}

	pbTrips := make([]*pb.BusinessTripResponse, len(trips))
	for i, trip := range trips {
		pbTrips[i] = h.modelToProto(trip)
	}

	return &pb.ListBusinessTripsResponse{
		Items:    pbTrips,
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// modelToProto 将模型转换为Proto响应
func (h *BusinessTripHandler) modelToProto(trip *model.BusinessTrip) *pb.BusinessTripResponse {
	resp := &pb.BusinessTripResponse{
		Id:             trip.ID.String(),
		TenantId:       trip.TenantID.String(),
		EmployeeId:     trip.EmployeeID.String(),
		EmployeeName:   trip.EmployeeName,
		DepartmentId:   trip.DepartmentID.String(),
		StartTime:      timestamppb.New(trip.StartTime),
		EndTime:        timestamppb.New(trip.EndTime),
		Duration:       trip.Duration,
		Destination:    trip.Destination,
		Transportation: trip.Transportation,
		Accommodation:  trip.Accommodation,
		Purpose:        trip.Purpose,
		Tasks:          trip.Tasks,
		EstimatedCost:  trip.EstimatedCost,
		ActualCost:     trip.ActualCost,
		ApprovalStatus: trip.ApprovalStatus,
		Remark:         trip.Remark,
		CreatedAt:      timestamppb.New(trip.CreatedAt),
		UpdatedAt:      timestamppb.New(trip.UpdatedAt),
	}

	// 转换同行人员
	if len(trip.Companions) > 0 {
		companions := make([]string, len(trip.Companions))
		for i, id := range trip.Companions {
			companions[i] = id.String()
		}
		resp.Companions = companions
	}

	// ApprovalId字段未在Proto中定义，暂时省略

	if trip.ApprovedBy != nil {
		resp.ApprovedBy = trip.ApprovedBy.String()
	}

	if trip.ApprovedAt != nil {
		resp.ApprovedAt = timestamppb.New(*trip.ApprovedAt)
	}

	if trip.RejectReason != "" {
		resp.RejectReason = trip.RejectReason
	}

	if trip.Report != "" {
		resp.Report = trip.Report
	}

	if trip.ReportAt != nil {
		resp.ReportAt = timestamppb.New(*trip.ReportAt)
	}

	return resp
}
