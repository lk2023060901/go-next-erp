package handler

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/lk2023060901/go-next-erp/api/hrm/v1"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/internal/hrm/service"
)

// AttendanceRuleHandler 考勤规则处理器
type AttendanceRuleHandler struct {
	pb.UnimplementedAttendanceRuleServiceServer
	ruleService service.AttendanceRuleService
}

// NewAttendanceRuleHandler 创建考勤规则处理器
func NewAttendanceRuleHandler(ruleService service.AttendanceRuleService) *AttendanceRuleHandler {
	return &AttendanceRuleHandler{
		ruleService: ruleService,
	}
}

// CreateAttendanceRule 创建考勤规则
func (h *AttendanceRuleHandler) CreateAttendanceRule(ctx context.Context, req *pb.CreateAttendanceRuleRequest) (*pb.AttendanceRuleResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}

	// 转换允许地点
	allowedLocations := make([]model.AllowedLocation, 0, len(req.AllowedLocations))
	for _, loc := range req.AllowedLocations {
		allowedLocations = append(allowedLocations, model.AllowedLocation{
			Name:      loc.Name,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
			Radius:    int(loc.Radius),
			Address:   loc.Address,
		})
	}

	// 使用默认的创建者ID
	createdBy := uuid.New()

	svcReq := &service.CreateAttendanceRuleRequest{
		TenantID:         tenantID,
		Name:             req.Name,
		Description:      req.Description,
		LocationRequired: req.LocationRequired,
		AllowedLocations: allowedLocations,
		WiFiRequired:     req.WifiRequired,
		AllowedWiFi:      req.AllowedWifi,
		FaceRequired:     req.FaceRequired,
		FaceThreshold:    req.FaceThreshold,
		IsActive:         true, // 默认启用
		CreatedBy:        createdBy,
	}

	rule, err := h.ruleService.Create(ctx, svcReq)
	if err != nil {
		return nil, err
	}

	return convertAttendanceRuleToProto(rule), nil
}

// GetAttendanceRule 获取考勤规则
func (h *AttendanceRuleHandler) GetAttendanceRule(ctx context.Context, req *pb.GetAttendanceRuleRequest) (*pb.AttendanceRuleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	rule, err := h.ruleService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertAttendanceRuleToProto(rule), nil
}

// UpdateAttendanceRule 更新考勤规则
func (h *AttendanceRuleHandler) UpdateAttendanceRule(ctx context.Context, req *pb.UpdateAttendanceRuleRequest) (*pb.AttendanceRuleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	// 使用默认的更新者ID
	updatedBy := uuid.New()

	svcReq := &service.UpdateAttendanceRuleRequest{
		UpdatedBy: updatedBy,
	}

	if req.Name != "" {
		svcReq.Name = &req.Name
	}
	if req.Description != "" {
		svcReq.Description = &req.Description
	}
	svcReq.LocationRequired = &req.LocationRequired
	svcReq.WiFiRequired = &req.WifiRequired
	svcReq.FaceRequired = &req.FaceRequired
	svcReq.IsActive = &req.IsActive

	rule, err := h.ruleService.Update(ctx, id, svcReq)
	if err != nil {
		return nil, err
	}

	return convertAttendanceRuleToProto(rule), nil
}

// DeleteAttendanceRule 删除考勤规则
func (h *AttendanceRuleHandler) DeleteAttendanceRule(ctx context.Context, req *pb.DeleteAttendanceRuleRequest) (*pb.DeleteAttendanceRuleResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	err = h.ruleService.Delete(ctx, id)
	if err != nil {
		return &pb.DeleteAttendanceRuleResponse{
			Success: false,
		}, err
	}

	return &pb.DeleteAttendanceRuleResponse{
		Success: true,
	}, nil
}

// ListAttendanceRules 列出考勤规则
func (h *AttendanceRuleHandler) ListAttendanceRules(ctx context.Context, req *pb.ListAttendanceRulesRequest) (*pb.ListAttendanceRulesResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}

	filter := &repository.AttendanceRuleFilter{}
	filter.IsActive = &req.IsActive

	// 计算offset和limit
	offset := 0
	limit := 20
	if req.Page > 0 && req.PageSize > 0 {
		offset = int((req.Page - 1) * req.PageSize)
		limit = int(req.PageSize)
	}

	rules, total, err := h.ruleService.List(ctx, tenantID, filter, offset, limit)
	if err != nil {
		return nil, err
	}

	resp := &pb.ListAttendanceRulesResponse{
		Items: make([]*pb.AttendanceRuleResponse, 0, len(rules)),
		Total: int32(total),
	}

	for _, rule := range rules {
		resp.Items = append(resp.Items, convertAttendanceRuleToProto(rule))
	}

	return resp, nil
}

// convertAttendanceRuleToProto 转换考勤规则到Protobuf格式
func convertAttendanceRuleToProto(rule *model.AttendanceRule) *pb.AttendanceRuleResponse {
	resp := &pb.AttendanceRuleResponse{
		Id:               rule.ID.String(),
		TenantId:         rule.TenantID.String(),
		Name:             rule.Name,
		Description:      rule.Description,
		LocationRequired: rule.LocationRequired,
		WifiRequired:     rule.WiFiRequired,
		AllowedWifi:      rule.AllowedWiFi,
		FaceRequired:     rule.FaceRequired,
		FaceThreshold:    rule.FaceThreshold,
		IsActive:         rule.IsActive,
		CreatedAt:        rule.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        rule.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// 转换允许地点
	if len(rule.AllowedLocations) > 0 {
		resp.AllowedLocations = make([]*pb.AllowedLocationInfo, 0, len(rule.AllowedLocations))
		for _, loc := range rule.AllowedLocations {
			resp.AllowedLocations = append(resp.AllowedLocations, &pb.AllowedLocationInfo{
				Name:      loc.Name,
				Latitude:  loc.Latitude,
				Longitude: loc.Longitude,
				Radius:    int32(loc.Radius),
				Address:   loc.Address,
			})
		}
	}

	return resp
}
