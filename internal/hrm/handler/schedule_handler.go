package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
	pb "github.com/lk2023060901/go-next-erp/api/hrm/v1"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/service"
)

// ScheduleHandler 排班处理器
type ScheduleHandler struct {
	pb.UnimplementedScheduleServiceServer
	scheduleService service.ScheduleService
}

// NewScheduleHandler 创建排班处理器
func NewScheduleHandler(scheduleService service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleService: scheduleService,
	}
}

// CreateSchedule 创建排班
func (h *ScheduleHandler) CreateSchedule(ctx context.Context, req *pb.CreateScheduleRequest) (*pb.ScheduleResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}
	employeeID, err := uuid.Parse(req.EmployeeId)
	if err != nil {
		return nil, err
	}
	shiftID, err := uuid.Parse(req.ShiftId)
	if err != nil {
		return nil, err
	}

	// 解析日期
	scheduleDate, err := time.Parse("2006-01-02", req.ScheduleDate)
	if err != nil {
		return nil, err
	}

	// 使用默认的创建者ID
	createdBy := uuid.New()

	svcReq := &service.CreateScheduleRequest{
		TenantID:     tenantID,
		EmployeeID:   employeeID,
		ShiftID:      shiftID,
		ScheduleDate: scheduleDate,
		WorkdayType:  req.WorkdayType,
		Remark:       req.Remark,
		Status:       "published",
		CreatedBy:    createdBy,
	}

	schedule, err := h.scheduleService.Create(ctx, svcReq)
	if err != nil {
		return nil, err
	}

	return convertScheduleToProto(schedule), nil
}

// BatchCreateSchedules 批量创建排班
func (h *ScheduleHandler) BatchCreateSchedules(ctx context.Context, req *pb.BatchCreateSchedulesRequest) (*pb.BatchCreateSchedulesResponse, error) {
	var successCount, failedCount int32
	errorMessages := make([]string, 0)

	for _, scheduleReq := range req.Schedules {
		_, err := h.CreateSchedule(ctx, scheduleReq)
		if err != nil {
			failedCount++
			errorMessages = append(errorMessages, err.Error())
		} else {
			successCount++
		}
	}

	return &pb.BatchCreateSchedulesResponse{
		SuccessCount:  successCount,
		FailedCount:   failedCount,
		ErrorMessages: errorMessages,
	}, nil
}

// GetSchedule 获取排班
func (h *ScheduleHandler) GetSchedule(ctx context.Context, req *pb.GetScheduleRequest) (*pb.ScheduleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	schedule, err := h.scheduleService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertScheduleToProto(schedule), nil
}

// UpdateSchedule 更新排班
func (h *ScheduleHandler) UpdateSchedule(ctx context.Context, req *pb.UpdateScheduleRequest) (*pb.ScheduleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	// 使用默认的更新者ID
	updatedBy := uuid.New()

	svcReq := &service.UpdateScheduleRequest{
		UpdatedBy: updatedBy,
	}

	if req.ShiftId != "" {
		shiftID, err := uuid.Parse(req.ShiftId)
		if err != nil {
			return nil, err
		}
		svcReq.ShiftID = &shiftID
	}
	if req.WorkdayType != "" {
		svcReq.WorkdayType = &req.WorkdayType
	}
	if req.Status != "" {
		svcReq.Status = &req.Status
	}
	if req.Remark != "" {
		svcReq.Remark = &req.Remark
	}

	schedule, err := h.scheduleService.Update(ctx, id, svcReq)
	if err != nil {
		return nil, err
	}

	return convertScheduleToProto(schedule), nil
}

// DeleteSchedule 删除排班
func (h *ScheduleHandler) DeleteSchedule(ctx context.Context, req *pb.DeleteScheduleRequest) (*pb.DeleteScheduleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	err = h.scheduleService.Delete(ctx, id)
	if err != nil {
		return &pb.DeleteScheduleResponse{
			Success: false,
		}, err
	}

	return &pb.DeleteScheduleResponse{
		Success: true,
	}, nil
}

// ListEmployeeSchedules 查询员工排班
func (h *ScheduleHandler) ListEmployeeSchedules(ctx context.Context, req *pb.ListEmployeeSchedulesRequest) (*pb.ListSchedulesResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}
	employeeID, err := uuid.Parse(req.EmployeeId)
	if err != nil {
		return nil, err
	}

	schedules, err := h.scheduleService.ListEmployeeSchedules(ctx, tenantID, employeeID, req.Month)
	if err != nil {
		return nil, err
	}

	resp := &pb.ListSchedulesResponse{
		Items: make([]*pb.ScheduleResponse, 0, len(schedules)),
		Total: int32(len(schedules)),
	}

	for _, schedule := range schedules {
		resp.Items = append(resp.Items, convertScheduleToProto(schedule))
	}

	return resp, nil
}

// ListDepartmentSchedules 查询部门排班
func (h *ScheduleHandler) ListDepartmentSchedules(ctx context.Context, req *pb.ListDepartmentSchedulesRequest) (*pb.ListSchedulesResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}
	departmentID, err := uuid.Parse(req.DepartmentId)
	if err != nil {
		return nil, err
	}

	schedules, err := h.scheduleService.ListDepartmentSchedules(ctx, tenantID, departmentID, req.Month)
	if err != nil {
		return nil, err
	}

	resp := &pb.ListSchedulesResponse{
		Items: make([]*pb.ScheduleResponse, 0, len(schedules)),
		Total: int32(len(schedules)),
	}

	for _, schedule := range schedules {
		resp.Items = append(resp.Items, convertScheduleToProto(schedule))
	}

	return resp, nil
}

// convertScheduleToProto 转换排班到Protobuf格式
func convertScheduleToProto(schedule *model.Schedule) *pb.ScheduleResponse {
	return &pb.ScheduleResponse{
		Id:           schedule.ID.String(),
		TenantId:     schedule.TenantID.String(),
		EmployeeId:   schedule.EmployeeID.String(),
		EmployeeName: schedule.EmployeeName,
		DepartmentId: schedule.DepartmentID.String(),
		ShiftId:      schedule.ShiftID.String(),
		ShiftName:    schedule.ShiftName,
		ScheduleDate: schedule.ScheduleDate.Format("2006-01-02"),
		WorkdayType:  schedule.WorkdayType,
		Status:       schedule.Status,
		Remark:       schedule.Remark,
		CreatedAt:    schedule.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
