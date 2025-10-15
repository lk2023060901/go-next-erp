package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	pb "github.com/lk2023060901/go-next-erp/api/hrm/v1"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/internal/hrm/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OvertimeHandler 加班处理器
type OvertimeHandler struct {
	pb.UnimplementedOvertimeServiceServer
	overtimeService service.OvertimeService
}

// NewOvertimeHandler 创建加班处理器
func NewOvertimeHandler(overtimeService service.OvertimeService) *OvertimeHandler {
	return &OvertimeHandler{
		overtimeService: overtimeService,
	}
}

// CreateOvertime 创建加班申请
func (h *OvertimeHandler) CreateOvertime(ctx context.Context, req *pb.CreateOvertimeRequest) (*pb.OvertimeResponse, error) {
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

	overtime := &model.Overtime{
		TenantID:     tenantID,
		EmployeeID:   employeeID,
		EmployeeName: req.EmployeeName,
		DepartmentID: departmentID,
		StartTime:    req.StartTime.AsTime(),
		EndTime:      req.EndTime.AsTime(),
		Duration:     req.Duration,
		OvertimeType: model.OvertimeType(req.OvertimeType),
		PayType:      req.PayType,
		Reason:       req.Reason,
		Tasks:        req.Tasks,
		Remark:       req.Remark,
	}

	if err := h.overtimeService.Create(ctx, overtime); err != nil {
		return nil, err
	}

	return h.modelToProto(overtime), nil
}

// UpdateOvertime 更新加班申请
func (h *OvertimeHandler) UpdateOvertime(ctx context.Context, req *pb.UpdateOvertimeRequest) (*pb.OvertimeResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	// 获取原记录
	overtime, err := h.overtimeService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.StartTime != nil {
		overtime.StartTime = req.StartTime.AsTime()
	}
	if req.EndTime != nil {
		overtime.EndTime = req.EndTime.AsTime()
	}
	if req.Duration > 0 {
		overtime.Duration = req.Duration
	}
	if req.OvertimeType != "" {
		overtime.OvertimeType = model.OvertimeType(req.OvertimeType)
	}
	if req.PayType != "" {
		overtime.PayType = req.PayType
	}
	if req.Reason != "" {
		overtime.Reason = req.Reason
	}
	if req.Tasks != nil {
		overtime.Tasks = req.Tasks
	}
	if req.Remark != "" {
		overtime.Remark = req.Remark
	}

	if err := h.overtimeService.Update(ctx, overtime); err != nil {
		return nil, err
	}

	return h.modelToProto(overtime), nil
}

// DeleteOvertime 删除加班申请
func (h *OvertimeHandler) DeleteOvertime(ctx context.Context, req *pb.DeleteOvertimeRequest) (*pb.DeleteOvertimeResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return &pb.DeleteOvertimeResponse{
			Success: false,
			Message: fmt.Sprintf("invalid id: %v", err),
		}, nil
	}

	if err := h.overtimeService.Delete(ctx, id); err != nil {
		return &pb.DeleteOvertimeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.DeleteOvertimeResponse{
		Success: true,
		Message: "Overtime deleted successfully",
	}, nil
}

// GetOvertime 获取加班详情
func (h *OvertimeHandler) GetOvertime(ctx context.Context, req *pb.GetOvertimeRequest) (*pb.OvertimeResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	overtime, err := h.overtimeService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return h.modelToProto(overtime), nil
}

// ListOvertimes 列表查询加班记录
func (h *OvertimeHandler) ListOvertimes(ctx context.Context, req *pb.ListOvertimesRequest) (*pb.ListOvertimesResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	filter := &repository.OvertimeFilter{
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

	if req.OvertimeType != "" {
		overtimeType := model.OvertimeType(req.OvertimeType)
		filter.OvertimeType = &overtimeType
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
		pageSize = 10
	}

	offset := int((page - 1) * pageSize)
	limit := int(pageSize)

	overtimes, total, err := h.overtimeService.List(ctx, tenantID, filter, offset, limit)
	if err != nil {
		return nil, err
	}

	items := make([]*pb.OvertimeResponse, 0, len(overtimes))
	for _, overtime := range overtimes {
		items = append(items, h.modelToProto(overtime))
	}

	return &pb.ListOvertimesResponse{
		Items:    items,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ListEmployeeOvertimes 查询员工加班记录
func (h *OvertimeHandler) ListEmployeeOvertimes(ctx context.Context, req *pb.ListEmployeeOvertimesRequest) (*pb.ListOvertimesResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	employeeID, err := uuid.Parse(req.EmployeeId)
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}

	year := int(req.Year)
	if year == 0 {
		year = time.Now().Year()
	}

	overtimes, err := h.overtimeService.ListByEmployee(ctx, tenantID, employeeID, year)
	if err != nil {
		return nil, err
	}

	items := make([]*pb.OvertimeResponse, 0, len(overtimes))
	for _, overtime := range overtimes {
		items = append(items, h.modelToProto(overtime))
	}

	return &pb.ListOvertimesResponse{
		Items:    items,
		Total:    int32(len(overtimes)),
		Page:     1,
		PageSize: int32(len(overtimes)),
	}, nil
}

// ListPendingOvertimes 查询待审批的加班
func (h *OvertimeHandler) ListPendingOvertimes(ctx context.Context, req *pb.ListPendingOvertimesRequest) (*pb.ListOvertimesResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	overtimes, err := h.overtimeService.ListPending(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	items := make([]*pb.OvertimeResponse, 0, len(overtimes))
	for _, overtime := range overtimes {
		items = append(items, h.modelToProto(overtime))
	}

	return &pb.ListOvertimesResponse{
		Items:    items,
		Total:    int32(len(overtimes)),
		Page:     1,
		PageSize: int32(len(overtimes)),
	}, nil
}

// SubmitOvertime 提交加班审批
func (h *OvertimeHandler) SubmitOvertime(ctx context.Context, req *pb.SubmitOvertimeRequest) (*pb.SubmitOvertimeResponse, error) {
	overtimeID, err := uuid.Parse(req.OvertimeId)
	if err != nil {
		return &pb.SubmitOvertimeResponse{
			Success: false,
			Message: fmt.Sprintf("invalid overtime_id: %v", err),
		}, nil
	}

	submitterID, err := uuid.Parse(req.SubmitterId)
	if err != nil {
		return &pb.SubmitOvertimeResponse{
			Success: false,
			Message: fmt.Sprintf("invalid submitter_id: %v", err),
		}, nil
	}

	if err := h.overtimeService.Submit(ctx, overtimeID, submitterID); err != nil {
		return &pb.SubmitOvertimeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.SubmitOvertimeResponse{
		Success: true,
		Message: "Overtime submitted successfully",
	}, nil
}

// ApproveOvertime 批准加班
func (h *OvertimeHandler) ApproveOvertime(ctx context.Context, req *pb.ApproveOvertimeRequest) (*pb.ApproveOvertimeResponse, error) {
	overtimeID, err := uuid.Parse(req.OvertimeId)
	if err != nil {
		return &pb.ApproveOvertimeResponse{
			Success: false,
			Message: fmt.Sprintf("invalid overtime_id: %v", err),
		}, nil
	}

	approverID, err := uuid.Parse(req.ApproverId)
	if err != nil {
		return &pb.ApproveOvertimeResponse{
			Success: false,
			Message: fmt.Sprintf("invalid approver_id: %v", err),
		}, nil
	}

	if err := h.overtimeService.Approve(ctx, overtimeID, approverID); err != nil {
		return &pb.ApproveOvertimeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.ApproveOvertimeResponse{
		Success: true,
		Message: "Overtime approved successfully",
	}, nil
}

// RejectOvertime 拒绝加班
func (h *OvertimeHandler) RejectOvertime(ctx context.Context, req *pb.RejectOvertimeRequest) (*pb.RejectOvertimeResponse, error) {
	overtimeID, err := uuid.Parse(req.OvertimeId)
	if err != nil {
		return &pb.RejectOvertimeResponse{
			Success: false,
			Message: fmt.Sprintf("invalid overtime_id: %v", err),
		}, nil
	}

	approverID, err := uuid.Parse(req.ApproverId)
	if err != nil {
		return &pb.RejectOvertimeResponse{
			Success: false,
			Message: fmt.Sprintf("invalid approver_id: %v", err),
		}, nil
	}

	if err := h.overtimeService.Reject(ctx, overtimeID, approverID, req.Reason); err != nil {
		return &pb.RejectOvertimeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.RejectOvertimeResponse{
		Success: true,
		Message: "Overtime rejected successfully",
	}, nil
}

// SumOvertimeHours 统计员工加班时长
func (h *OvertimeHandler) SumOvertimeHours(ctx context.Context, req *pb.SumOvertimeHoursRequest) (*pb.SumOvertimeHoursResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	employeeID, err := uuid.Parse(req.EmployeeId)
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}

	startDate := req.StartDate.AsTime()
	endDate := req.EndDate.AsTime()

	totalHours, err := h.overtimeService.SumHoursByEmployee(ctx, tenantID, employeeID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &pb.SumOvertimeHoursResponse{
		TotalHours: totalHours,
	}, nil
}

// GetCompOffDays 统计可调休天数
func (h *OvertimeHandler) GetCompOffDays(ctx context.Context, req *pb.GetCompOffDaysRequest) (*pb.GetCompOffDaysResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	employeeID, err := uuid.Parse(req.EmployeeId)
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}

	availableDays, err := h.overtimeService.SumCompOffDays(ctx, tenantID, employeeID)
	if err != nil {
		return nil, err
	}

	return &pb.GetCompOffDaysResponse{
		AvailableDays: availableDays,
	}, nil
}

// UseCompOffDays 使用调休
func (h *OvertimeHandler) UseCompOffDays(ctx context.Context, req *pb.UseCompOffDaysRequest) (*pb.UseCompOffDaysResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return &pb.UseCompOffDaysResponse{
			Success: false,
			Message: fmt.Sprintf("invalid tenant_id: %v", err),
		}, nil
	}

	employeeID, err := uuid.Parse(req.EmployeeId)
	if err != nil {
		return &pb.UseCompOffDaysResponse{
			Success: false,
			Message: fmt.Sprintf("invalid employee_id: %v", err),
		}, nil
	}

	if err := h.overtimeService.UseCompOffDays(ctx, tenantID, employeeID, req.Days); err != nil {
		return &pb.UseCompOffDaysResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 查询剩余天数
	remainingDays, err := h.overtimeService.SumCompOffDays(ctx, tenantID, employeeID)
	if err != nil {
		remainingDays = 0
	}

	return &pb.UseCompOffDaysResponse{
		Success:       true,
		Message:       "Comp off days used successfully",
		RemainingDays: remainingDays,
	}, nil
}

// modelToProto 将Model转换为Proto
func (h *OvertimeHandler) modelToProto(overtime *model.Overtime) *pb.OvertimeResponse {
	resp := &pb.OvertimeResponse{
		Id:             overtime.ID.String(),
		TenantId:       overtime.TenantID.String(),
		EmployeeId:     overtime.EmployeeID.String(),
		EmployeeName:   overtime.EmployeeName,
		DepartmentId:   overtime.DepartmentID.String(),
		StartTime:      timestamppb.New(overtime.StartTime),
		EndTime:        timestamppb.New(overtime.EndTime),
		Duration:       overtime.Duration,
		OvertimeType:   string(overtime.OvertimeType),
		PayType:        overtime.PayType,
		PayRate:        overtime.PayRate,
		Reason:         overtime.Reason,
		Tasks:          overtime.Tasks,
		ApprovalStatus: overtime.ApprovalStatus,
		RejectReason:   overtime.RejectReason,
		CompOffDays:    overtime.CompOffDays,
		CompOffUsed:    overtime.CompOffUsed,
		Remark:         overtime.Remark,
		CreatedAt:      timestamppb.New(overtime.CreatedAt),
		UpdatedAt:      timestamppb.New(overtime.UpdatedAt),
	}

	if overtime.ApprovedBy != nil {
		resp.ApprovedBy = overtime.ApprovedBy.String()
	}

	if overtime.ApprovedAt != nil {
		resp.ApprovedAt = timestamppb.New(*overtime.ApprovedAt)
	}

	if overtime.CompOffExpireAt != nil {
		resp.CompOffExpireAt = timestamppb.New(*overtime.CompOffExpireAt)
	}

	return resp
}
