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

// PunchCardSupplementHandler 补卡申请处理器
type PunchCardSupplementHandler struct {
	pb.UnimplementedPunchCardSupplementServiceServer
	supplementService service.PunchCardSupplementService
}

// NewPunchCardSupplementHandler 创建补卡申请处理器
func NewPunchCardSupplementHandler(supplementService service.PunchCardSupplementService) *PunchCardSupplementHandler {
	return &PunchCardSupplementHandler{
		supplementService: supplementService,
	}
}

// CreatePunchCardSupplement 创建补卡申请
func (h *PunchCardSupplementHandler) CreatePunchCardSupplement(ctx context.Context, req *pb.CreatePunchCardSupplementRequest) (*pb.PunchCardSupplementResponse, error) {
	// 1. 参数验证
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

	// 2. 转换证明材料
	evidence := make([]model.SupplementEvidence, 0, len(req.Evidence))
	for _, e := range req.Evidence {
		evidence = append(evidence, model.SupplementEvidence{
			Type:        e.Type,
			URL:         e.Url,
			Description: e.Description,
		})
	}

	// 3. 构造补卡申请对象
	supplement := &model.PunchCardSupplement{
		TenantID:       tenantID,
		EmployeeID:     employeeID,
		EmployeeName:   req.EmployeeName,
		DepartmentID:   departmentID,
		SupplementDate: req.SupplementDate.AsTime(),
		SupplementType: model.SupplementType(req.SupplementType),
		SupplementTime: req.SupplementTime.AsTime(),
		MissingType:    model.PunchCardMissingType(req.MissingType),
		Reason:         req.Reason,
		Evidence:       evidence,
		Remark:         req.Remark,
	}

	// 4. 创建补卡申请
	if err := h.supplementService.Create(ctx, supplement); err != nil {
		return nil, err
	}

	// 5. 返回响应
	return h.modelToProto(supplement), nil
}

// UpdatePunchCardSupplement 更新补卡申请
func (h *PunchCardSupplementHandler) UpdatePunchCardSupplement(ctx context.Context, req *pb.UpdatePunchCardSupplementRequest) (*pb.PunchCardSupplementResponse, error) {
	// 1. 参数验证
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	// 2. 获取原记录
	supplement, err := h.supplementService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. 更新字段
	if req.SupplementDate != nil {
		supplement.SupplementDate = req.SupplementDate.AsTime()
	}
	if req.SupplementType != "" {
		supplement.SupplementType = model.SupplementType(req.SupplementType)
	}
	if req.SupplementTime != nil {
		supplement.SupplementTime = req.SupplementTime.AsTime()
	}
	if req.MissingType != "" {
		supplement.MissingType = model.PunchCardMissingType(req.MissingType)
	}
	if req.Reason != "" {
		supplement.Reason = req.Reason
	}
	if len(req.Evidence) > 0 {
		evidence := make([]model.SupplementEvidence, 0, len(req.Evidence))
		for _, e := range req.Evidence {
			evidence = append(evidence, model.SupplementEvidence{
				Type:        e.Type,
				URL:         e.Url,
				Description: e.Description,
			})
		}
		supplement.Evidence = evidence
	}
	if req.Remark != "" {
		supplement.Remark = req.Remark
	}

	// 4. 更新补卡申请
	if err := h.supplementService.Update(ctx, supplement); err != nil {
		return nil, err
	}

	// 5. 返回响应
	return h.modelToProto(supplement), nil
}

// DeletePunchCardSupplement 删除补卡申请
func (h *PunchCardSupplementHandler) DeletePunchCardSupplement(ctx context.Context, req *pb.DeletePunchCardSupplementRequest) (*pb.DeletePunchCardSupplementResponse, error) {
	// 1. 参数验证
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return &pb.DeletePunchCardSupplementResponse{
			Success: false,
			Message: fmt.Sprintf("invalid id: %v", err),
		}, nil
	}

	// 2. 删除补卡申请
	if err := h.supplementService.Delete(ctx, id); err != nil {
		return &pb.DeletePunchCardSupplementResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 3. 返回响应
	return &pb.DeletePunchCardSupplementResponse{
		Success: true,
		Message: "Punch card supplement deleted successfully",
	}, nil
}

// GetPunchCardSupplement 获取补卡申请详情
func (h *PunchCardSupplementHandler) GetPunchCardSupplement(ctx context.Context, req *pb.GetPunchCardSupplementRequest) (*pb.PunchCardSupplementResponse, error) {
	// 1. 参数验证
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %w", err)
	}

	// 2. 获取补卡申请
	supplement, err := h.supplementService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. 返回响应
	return h.modelToProto(supplement), nil
}

// ListPunchCardSupplements 列表查询补卡申请
func (h *PunchCardSupplementHandler) ListPunchCardSupplements(ctx context.Context, req *pb.ListPunchCardSupplementsRequest) (*pb.ListPunchCardSupplementsResponse, error) {
	// 1. 参数验证
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	// 2. 构建过滤器
	filter := &repository.PunchCardSupplementFilter{
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

	if req.SupplementType != "" {
		suppType := model.SupplementType(req.SupplementType)
		filter.SupplementType = &suppType
	}

	if req.MissingType != "" {
		missType := model.PunchCardMissingType(req.MissingType)
		filter.MissingType = &missType
	}

	if req.ApprovalStatus != "" {
		filter.ApprovalStatus = &req.ApprovalStatus
	}

	if req.ProcessStatus != "" {
		filter.ProcessStatus = &req.ProcessStatus
	}

	if req.StartDate != nil {
		startDate := req.StartDate.AsTime()
		filter.StartDate = &startDate
	}

	if req.EndDate != nil {
		endDate := req.EndDate.AsTime()
		filter.EndDate = &endDate
	}

	// 3. 分页参数
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 4. 查询补卡申请
	supplements, total, err := h.supplementService.List(ctx, tenantID, filter, int(offset), int(pageSize))
	if err != nil {
		return nil, err
	}

	// 5. 转换为proto格式
	items := make([]*pb.PunchCardSupplementResponse, 0, len(supplements))
	for _, s := range supplements {
		items = append(items, h.modelToProto(s))
	}

	// 6. 返回响应
	return &pb.ListPunchCardSupplementsResponse{
		Items:    items,
		Total:    int64(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ListEmployeePunchCardSupplements 查询员工补卡申请记录
func (h *PunchCardSupplementHandler) ListEmployeePunchCardSupplements(ctx context.Context, req *pb.ListEmployeePunchCardSupplementsRequest) (*pb.ListPunchCardSupplementsResponse, error) {
	// 1. 参数验证
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	employeeID, err := uuid.Parse(req.EmployeeId)
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}

	// 2. 查询员工补卡申请
	year := int(req.Year)
	if year == 0 {
		year = time.Now().Year()
	}

	supplements, err := h.supplementService.ListByEmployee(ctx, tenantID, employeeID, year)
	if err != nil {
		return nil, err
	}

	// 3. 转换为proto格式
	items := make([]*pb.PunchCardSupplementResponse, 0, len(supplements))
	for _, s := range supplements {
		items = append(items, h.modelToProto(s))
	}

	// 4. 返回响应
	return &pb.ListPunchCardSupplementsResponse{
		Items:    items,
		Total:    int64(len(items)),
		Page:     1,
		PageSize: int32(len(items)),
	}, nil
}

// ListPendingPunchCardSupplements 查询待审批的补卡申请
func (h *PunchCardSupplementHandler) ListPendingPunchCardSupplements(ctx context.Context, req *pb.ListPendingPunchCardSupplementsRequest) (*pb.ListPunchCardSupplementsResponse, error) {
	// 1. 参数验证
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	// 2. 查询待审批的补卡申请
	supplements, err := h.supplementService.ListPending(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 3. 转换为proto格式
	items := make([]*pb.PunchCardSupplementResponse, 0, len(supplements))
	for _, s := range supplements {
		items = append(items, h.modelToProto(s))
	}

	// 4. 返回响应
	return &pb.ListPunchCardSupplementsResponse{
		Items:    items,
		Total:    int64(len(items)),
		Page:     1,
		PageSize: int32(len(items)),
	}, nil
}

// SubmitPunchCardSupplement 提交补卡申请审批
func (h *PunchCardSupplementHandler) SubmitPunchCardSupplement(ctx context.Context, req *pb.SubmitPunchCardSupplementRequest) (*pb.SubmitPunchCardSupplementResponse, error) {
	// 1. 参数验证
	supplementID, err := uuid.Parse(req.SupplementId)
	if err != nil {
		return &pb.SubmitPunchCardSupplementResponse{
			Success: false,
			Message: fmt.Sprintf("invalid supplement_id: %v", err),
		}, nil
	}

	submitterID, err := uuid.Parse(req.SubmitterId)
	if err != nil {
		return &pb.SubmitPunchCardSupplementResponse{
			Success: false,
			Message: fmt.Sprintf("invalid submitter_id: %v", err),
		}, nil
	}

	// 2. 提交审批
	if err := h.supplementService.Submit(ctx, supplementID, submitterID); err != nil {
		return &pb.SubmitPunchCardSupplementResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 3. 返回响应
	return &pb.SubmitPunchCardSupplementResponse{
		Success: true,
		Message: "Punch card supplement submitted successfully",
	}, nil
}

// ApprovePunchCardSupplement 批准补卡申请
func (h *PunchCardSupplementHandler) ApprovePunchCardSupplement(ctx context.Context, req *pb.ApprovePunchCardSupplementRequest) (*pb.ApprovePunchCardSupplementResponse, error) {
	// 1. 参数验证
	supplementID, err := uuid.Parse(req.SupplementId)
	if err != nil {
		return &pb.ApprovePunchCardSupplementResponse{
			Success: false,
			Message: fmt.Sprintf("invalid supplement_id: %v", err),
		}, nil
	}

	approverID, err := uuid.Parse(req.ApproverId)
	if err != nil {
		return &pb.ApprovePunchCardSupplementResponse{
			Success: false,
			Message: fmt.Sprintf("invalid approver_id: %v", err),
		}, nil
	}

	// 2. 批准补卡申请
	if err := h.supplementService.Approve(ctx, supplementID, approverID, req.Comment); err != nil {
		return &pb.ApprovePunchCardSupplementResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 3. 返回响应
	return &pb.ApprovePunchCardSupplementResponse{
		Success: true,
		Message: "Punch card supplement approved successfully",
	}, nil
}

// RejectPunchCardSupplement 拒绝补卡申请
func (h *PunchCardSupplementHandler) RejectPunchCardSupplement(ctx context.Context, req *pb.RejectPunchCardSupplementRequest) (*pb.RejectPunchCardSupplementResponse, error) {
	// 1. 参数验证
	supplementID, err := uuid.Parse(req.SupplementId)
	if err != nil {
		return &pb.RejectPunchCardSupplementResponse{
			Success: false,
			Message: fmt.Sprintf("invalid supplement_id: %v", err),
		}, nil
	}

	approverID, err := uuid.Parse(req.ApproverId)
	if err != nil {
		return &pb.RejectPunchCardSupplementResponse{
			Success: false,
			Message: fmt.Sprintf("invalid approver_id: %v", err),
		}, nil
	}

	// 2. 拒绝补卡申请
	if err := h.supplementService.Reject(ctx, supplementID, approverID, req.Reason); err != nil {
		return &pb.RejectPunchCardSupplementResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 3. 返回响应
	return &pb.RejectPunchCardSupplementResponse{
		Success: true,
		Message: "Punch card supplement rejected successfully",
	}, nil
}

// ProcessPunchCardSupplement 处理补卡（生成补卡记录）
func (h *PunchCardSupplementHandler) ProcessPunchCardSupplement(ctx context.Context, req *pb.ProcessPunchCardSupplementRequest) (*pb.ProcessPunchCardSupplementResponse, error) {
	// 1. 参数验证
	supplementID, err := uuid.Parse(req.SupplementId)
	if err != nil {
		return &pb.ProcessPunchCardSupplementResponse{
			Success: false,
			Message: fmt.Sprintf("invalid supplement_id: %v", err),
		}, nil
	}

	processorID, err := uuid.Parse(req.ProcessorId)
	if err != nil {
		return &pb.ProcessPunchCardSupplementResponse{
			Success: false,
			Message: fmt.Sprintf("invalid processor_id: %v", err),
		}, nil
	}

	// 2. 处理补卡
	if err := h.supplementService.Process(ctx, supplementID, processorID); err != nil {
		return &pb.ProcessPunchCardSupplementResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 3. 返回响应
	return &pb.ProcessPunchCardSupplementResponse{
		Success: true,
		Message: "Punch card supplement processed successfully",
	}, nil
}

// CancelPunchCardSupplement 取消补卡申请
func (h *PunchCardSupplementHandler) CancelPunchCardSupplement(ctx context.Context, req *pb.CancelPunchCardSupplementRequest) (*pb.CancelPunchCardSupplementResponse, error) {
	// 1. 参数验证
	supplementID, err := uuid.Parse(req.SupplementId)
	if err != nil {
		return &pb.CancelPunchCardSupplementResponse{
			Success: false,
			Message: fmt.Sprintf("invalid supplement_id: %v", err),
		}, nil
	}

	// 2. 取消补卡申请
	if err := h.supplementService.Cancel(ctx, supplementID); err != nil {
		return &pb.CancelPunchCardSupplementResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 3. 返回响应
	return &pb.CancelPunchCardSupplementResponse{
		Success: true,
		Message: "Punch card supplement cancelled successfully",
	}, nil
}

// modelToProto 将model转换为proto格式
func (h *PunchCardSupplementHandler) modelToProto(supplement *model.PunchCardSupplement) *pb.PunchCardSupplementResponse {
	resp := &pb.PunchCardSupplementResponse{
		Id:             supplement.ID.String(),
		TenantId:       supplement.TenantID.String(),
		EmployeeId:     supplement.EmployeeID.String(),
		EmployeeName:   supplement.EmployeeName,
		DepartmentId:   supplement.DepartmentID.String(),
		SupplementDate: timestamppb.New(supplement.SupplementDate),
		SupplementType: string(supplement.SupplementType),
		SupplementTime: timestamppb.New(supplement.SupplementTime),
		MissingType:    string(supplement.MissingType),
		Reason:         supplement.Reason,
		ApprovalStatus: supplement.ApprovalStatus,
		ProcessStatus:  supplement.ProcessStatus,
		Remark:         supplement.Remark,
		CreatedAt:      timestamppb.New(supplement.CreatedAt),
		UpdatedAt:      timestamppb.New(supplement.UpdatedAt),
	}

	// 转换证明材料
	if len(supplement.Evidence) > 0 {
		evidence := make([]*pb.SupplementEvidence, 0, len(supplement.Evidence))
		for _, e := range supplement.Evidence {
			evidence = append(evidence, &pb.SupplementEvidence{
				Type:        e.Type,
				Url:         e.URL,
				Description: e.Description,
			})
		}
		resp.Evidence = evidence
	}

	// 可选字段
	if supplement.AttendanceRecordID != nil {
		resp.AttendanceRecordId = supplement.AttendanceRecordID.String()
	}

	if supplement.ApprovalID != nil {
		resp.ApprovalId = supplement.ApprovalID.String()
	}

	if supplement.ApprovedBy != nil {
		resp.ApprovedBy = supplement.ApprovedBy.String()
	}

	if supplement.ApprovedAt != nil {
		resp.ApprovedAt = timestamppb.New(*supplement.ApprovedAt)
	}

	if supplement.RejectReason != "" {
		resp.RejectReason = supplement.RejectReason
	}

	if supplement.ProcessedAt != nil {
		resp.ProcessedAt = timestamppb.New(*supplement.ProcessedAt)
	}

	if supplement.ProcessedBy != nil {
		resp.ProcessedBy = supplement.ProcessedBy.String()
	}

	return resp
}
