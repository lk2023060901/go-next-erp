package handler

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/lk2023060901/go-next-erp/api/hrm/v1"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/internal/hrm/service"
)

// ShiftHandler 班次处理器
type ShiftHandler struct {
	pb.UnimplementedShiftServiceServer
	shiftService service.ShiftService
}

// NewShiftHandler 创建班次处理器
func NewShiftHandler(shiftService service.ShiftService) *ShiftHandler {
	return &ShiftHandler{
		shiftService: shiftService,
	}
}

// CreateShift 创建班次
func (h *ShiftHandler) CreateShift(ctx context.Context, req *pb.CreateShiftRequest) (*pb.ShiftResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}

	// 使用默认的创建者ID (在实际使用中应从context中获取)
	createdBy := uuid.New()

	svcReq := &service.CreateShiftRequest{
		TenantID:         tenantID,
		Code:             req.Code,
		Name:             req.Name,
		Description:      req.Description,
		Type:             model.ShiftType(req.Type),
		WorkStart:        req.WorkStart,
		WorkEnd:          req.WorkEnd,
		FlexibleStart:    req.FlexibleStart,
		FlexibleEnd:      req.FlexibleEnd,
		WorkDuration:     int(req.WorkDuration),
		CheckInRequired:  req.CheckInRequired,
		CheckOutRequired: req.CheckOutRequired,
		LateGracePeriod:  int(req.LateGracePeriod),
		EarlyGracePeriod: int(req.EarlyGracePeriod),
		IsCrossDays:      req.IsCrossDays,
		AllowOvertime:    req.AllowOvertime,
		IsActive:         true, // 默认启用
		Sort:             int(req.Sort),
		CreatedBy:        createdBy,
	}

	shift, err := h.shiftService.Create(ctx, svcReq)
	if err != nil {
		return nil, err
	}

	return convertShiftToProto(shift), nil
}

// GetShift 获取班次
func (h *ShiftHandler) GetShift(ctx context.Context, req *pb.GetShiftRequest) (*pb.ShiftResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	shift, err := h.shiftService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertShiftToProto(shift), nil
}

// UpdateShift 更新班次
func (h *ShiftHandler) UpdateShift(ctx context.Context, req *pb.UpdateShiftRequest) (*pb.ShiftResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	// 使用默认的更新者ID (在实际使用中应从context中获取)
	updatedBy := uuid.New()

	svcReq := &service.UpdateShiftRequest{
		UpdatedBy: updatedBy,
	}

	if req.Name != "" {
		svcReq.Name = &req.Name
	}
	if req.Description != "" {
		svcReq.Description = &req.Description
	}
	if req.WorkStart != "" {
		svcReq.WorkStart = &req.WorkStart
	}
	if req.WorkEnd != "" {
		svcReq.WorkEnd = &req.WorkEnd
	}
	if req.LateGracePeriod > 0 {
		late := int(req.LateGracePeriod)
		svcReq.LateGracePeriod = &late
	}
	if req.EarlyGracePeriod > 0 {
		early := int(req.EarlyGracePeriod)
		svcReq.EarlyGracePeriod = &early
	}
	svcReq.IsActive = &req.IsActive
	if req.Sort > 0 {
		sort := int(req.Sort)
		svcReq.Sort = &sort
	}

	shift, err := h.shiftService.Update(ctx, id, svcReq)
	if err != nil {
		return nil, err
	}

	return convertShiftToProto(shift), nil
}

// DeleteShift 删除班次
func (h *ShiftHandler) DeleteShift(ctx context.Context, req *pb.DeleteShiftRequest) (*pb.DeleteShiftResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	err = h.shiftService.Delete(ctx, id)
	if err != nil {
		return &pb.DeleteShiftResponse{
			Success: false,
		}, err
	}

	return &pb.DeleteShiftResponse{
		Success: true,
	}, nil
}

// ListShifts 列出班次
func (h *ShiftHandler) ListShifts(ctx context.Context, req *pb.ListShiftsRequest) (*pb.ListShiftsResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}

	filter := &repository.ShiftFilter{}
	if req.Type != "" {
		shiftType := model.ShiftType(req.Type)
		filter.Type = &shiftType
	}
	filter.IsActive = &req.IsActive

	// 计算offset和limit
	offset := 0
	limit := 20
	if req.Page > 0 && req.PageSize > 0 {
		offset = int((req.Page - 1) * req.PageSize)
		limit = int(req.PageSize)
	}

	shifts, total, err := h.shiftService.List(ctx, tenantID, filter, offset, limit)
	if err != nil {
		return nil, err
	}

	resp := &pb.ListShiftsResponse{
		Items: make([]*pb.ShiftResponse, 0, len(shifts)),
		Total: int32(total),
	}

	for _, shift := range shifts {
		resp.Items = append(resp.Items, convertShiftToProto(shift))
	}

	return resp, nil
}

// ListActiveShifts 列出启用的班次
func (h *ShiftHandler) ListActiveShifts(ctx context.Context, req *pb.ListActiveShiftsRequest) (*pb.ListShiftsResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}

	shifts, err := h.shiftService.ListActive(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	resp := &pb.ListShiftsResponse{
		Items: make([]*pb.ShiftResponse, 0, len(shifts)),
		Total: int32(len(shifts)),
	}

	for _, shift := range shifts {
		resp.Items = append(resp.Items, convertShiftToProto(shift))
	}

	return resp, nil
}

// convertShiftToProto 转换班次到Protobuf格式
func convertShiftToProto(shift *model.Shift) *pb.ShiftResponse {
	return &pb.ShiftResponse{
		Id:               shift.ID.String(),
		TenantId:         shift.TenantID.String(),
		Code:             shift.Code,
		Name:             shift.Name,
		Description:      shift.Description,
		Type:             string(shift.Type),
		WorkStart:        shift.WorkStart,
		WorkEnd:          shift.WorkEnd,
		FlexibleStart:    shift.FlexibleStart,
		FlexibleEnd:      shift.FlexibleEnd,
		WorkDuration:     int32(shift.WorkDuration),
		CheckInRequired:  shift.CheckInRequired,
		CheckOutRequired: shift.CheckOutRequired,
		LateGracePeriod:  int32(shift.LateGracePeriod),
		EarlyGracePeriod: int32(shift.EarlyGracePeriod),
		IsCrossDays:      shift.IsCrossDays,
		AllowOvertime:    shift.AllowOvertime,
		IsActive:         shift.IsActive,
		Sort:             int32(shift.Sort),
		CreatedAt:        shift.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        shift.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
