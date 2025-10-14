package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	orgv1 "github.com/lk2023060901/go-next-erp/api/organization/v1"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
	"github.com/lk2023060901/go-next-erp/internal/organization/repository"
	"github.com/lk2023060901/go-next-erp/internal/organization/service"
)

// OrganizationAdapter 组织适配器
type OrganizationAdapter struct {
	orgv1.UnimplementedOrganizationServiceServer
	orgv1.UnimplementedEmployeeServiceServer
	orgService service.OrganizationService
	empService service.EmployeeService
	typeRepo   repository.OrganizationTypeRepository
}

// NewOrganizationAdapter 创建组织适配器
func NewOrganizationAdapter(
	orgService service.OrganizationService,
	empService service.EmployeeService,
	typeRepo repository.OrganizationTypeRepository,
) *OrganizationAdapter {
	return &OrganizationAdapter{
		orgService: orgService,
		empService: empService,
		typeRepo:   typeRepo,
	}
}

// CreateOrganization 创建组织
func (a *OrganizationAdapter) CreateOrganization(ctx context.Context, req *orgv1.CreateOrganizationRequest) (*orgv1.OrganizationResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)
	var parentID *uuid.UUID
	if req.ParentId != "" {
		id, _ := uuid.Parse(req.ParentId)
		parentID = &id
	}

	// TODO: 需要从 context 获取当前用户ID
	createdBy := tenantID

	createReq := &service.CreateOrganizationRequest{
		TenantID:  tenantID,
		Code:      req.Code,
		Name:      req.Name,
		ParentID:  parentID,
		Sort:      int(req.Sort),
		Status:    "active",
		CreatedBy: createdBy,
	}

	// 设置 TypeID（根据 type 字符串查找对应的 TypeID）
	if req.Type != "" {
		orgType, err := a.typeRepo.GetByCode(ctx, tenantID, req.Type)
		if err != nil {
			return nil, fmt.Errorf("get organization type failed: %w", err)
		}
		createReq.TypeID = orgType.ID
	}

	org, err := a.orgService.Create(ctx, createReq)
	if err != nil {
		return nil, err
	}

	return a.orgToProto(org), nil
}

// GetOrganization 获取组织
func (a *OrganizationAdapter) GetOrganization(ctx context.Context, req *orgv1.GetOrganizationRequest) (*orgv1.OrganizationResponse, error) {
	id, _ := uuid.Parse(req.Id)
	org, err := a.orgService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return a.orgToProto(org), nil
}

// UpdateOrganization 更新组织
func (a *OrganizationAdapter) UpdateOrganization(ctx context.Context, req *orgv1.UpdateOrganizationRequest) (*orgv1.OrganizationResponse, error) {
	id, _ := uuid.Parse(req.Id)

	// TODO: 需要从 context 获取当前用户ID
	updatedBy := uuid.New()

	updateReq := &service.UpdateOrganizationRequest{
		Name:      req.Name,
		Sort:      int(req.Sort),
		Status:    "active",
		UpdatedBy: updatedBy,
	}

	org, err := a.orgService.Update(ctx, id, updateReq)
	if err != nil {
		return nil, err
	}

	return a.orgToProto(org), nil
}

// DeleteOrganization 删除组织
func (a *OrganizationAdapter) DeleteOrganization(ctx context.Context, req *orgv1.DeleteOrganizationRequest) (*orgv1.DeleteOrganizationResponse, error) {
	id, _ := uuid.Parse(req.Id)
	if err := a.orgService.Delete(ctx, id); err != nil {
		return nil, err
	}
	return &orgv1.DeleteOrganizationResponse{Success: true}, nil
}

// ListOrganizations 列出组织
func (a *OrganizationAdapter) ListOrganizations(ctx context.Context, req *orgv1.ListOrganizationsRequest) (*orgv1.ListOrganizationsResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)

	orgs, err := a.orgService.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 如果指定了 parent_id，过滤结果
	if req.ParentId != "" {
		parentID, _ := uuid.Parse(req.ParentId)
		filtered := make([]*orgv1.OrganizationResponse, 0)
		for _, org := range orgs {
			if org.ParentID != nil && *org.ParentID == parentID {
				filtered = append(filtered, a.orgToProto(org))
			}
		}
		return &orgv1.ListOrganizationsResponse{
			Items: filtered,
			Total: int32(len(filtered)),
		}, nil
	}

	items := make([]*orgv1.OrganizationResponse, 0, len(orgs))
	for _, org := range orgs {
		items = append(items, a.orgToProto(org))
	}

	return &orgv1.ListOrganizationsResponse{
		Items: items,
		Total: int32(len(items)),
	}, nil
}

// GetOrganizationTree 获取组织树
func (a *OrganizationAdapter) GetOrganizationTree(ctx context.Context, req *orgv1.GetOrganizationTreeRequest) (*orgv1.OrganizationTreeResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)

	treeNodes, err := a.orgService.GetTree(ctx, tenantID, nil)
	if err != nil {
		return nil, err
	}

	nodes := a.buildTreeProtoNodes(treeNodes)
	return &orgv1.OrganizationTreeResponse{Nodes: nodes}, nil
}

// CreateEmployee 创建员工
func (a *OrganizationAdapter) CreateEmployee(ctx context.Context, req *orgv1.CreateEmployeeRequest) (*orgv1.EmployeeResponse, error) {
	tenantID, _ := uuid.Parse(req.TenantId)
	orgID, _ := uuid.Parse(req.OrgId)

	// TODO: 需要从 context 获取当前用户ID
	createdBy := tenantID

	// 创建请求
	createReq := &service.CreateEmployeeRequest{
		TenantID:   tenantID,
		EmployeeNo: req.EmployeeNo,
		Name:       req.Name,
		Mobile:     req.Mobile,
		Email:      req.Email,
		OrgID:      orgID,
		Status:     "active", // 默认状态
		CreatedBy:  createdBy,
	}

	// 如果指定了 UserID 且非空，则使用它
	if req.UserId != "" {
		userID, err := uuid.Parse(req.UserId)
		if err == nil {
			createReq.UserID = userID
		} else {
			// 如果 UserID 解析失败，使用 UUID Nil
			createReq.UserID = uuid.Nil
		}
	} else {
		// 如果没有指定 UserID，使用 UUID Nil （表示还没有关联用户）
		createReq.UserID = uuid.Nil
	}

	// 如果指定了 PositionID
	if req.PositionId != "" {
		positionID, _ := uuid.Parse(req.PositionId)
		createReq.PositionID = &positionID
	}

	// 如果指定了 Status
	if req.Status != "" {
		createReq.Status = req.Status
	}

	// 调用 service 创建员工
	emp, err := a.empService.Create(ctx, createReq)
	if err != nil {
		return nil, err
	}

	return a.empToProto(emp), nil
}

// empToProto 将 model.Employee 转换为 orgv1.EmployeeResponse
func (a *OrganizationAdapter) empToProto(emp *model.Employee) *orgv1.EmployeeResponse {
	resp := &orgv1.EmployeeResponse{
		Id:         emp.ID.String(),
		TenantId:   emp.TenantID.String(),
		UserId:     emp.UserID.String(),
		OrgId:      emp.OrgID.String(),
		EmployeeNo: emp.EmployeeNo,
		Name:       emp.Name,
		Mobile:     emp.Mobile,
		Email:      emp.Email,
		Status:     emp.Status,
		CreatedAt:  emp.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  emp.UpdatedAt.Format(time.RFC3339),
	}

	if emp.PositionID != nil {
		resp.PositionId = emp.PositionID.String()
	}

	return resp
}

// GetEmployee 获取员工（简化实现）
func (a *OrganizationAdapter) GetEmployee(ctx context.Context, req *orgv1.GetEmployeeRequest) (*orgv1.EmployeeResponse, error) {
	// TODO: 实现员工查询逻辑
	return &orgv1.EmployeeResponse{
		Id:        req.Id,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// UpdateEmployee 更新员工（简化实现）
func (a *OrganizationAdapter) UpdateEmployee(ctx context.Context, req *orgv1.UpdateEmployeeRequest) (*orgv1.EmployeeResponse, error) {
	// TODO: 实现员工更新逻辑
	return &orgv1.EmployeeResponse{
		Id:        req.Id,
		Name:      req.Name,
		Mobile:    req.Mobile,
		Email:     req.Email,
		Status:    req.Status,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// DeleteEmployee 删除员工（简化实现）
func (a *OrganizationAdapter) DeleteEmployee(ctx context.Context, req *orgv1.DeleteEmployeeRequest) (*orgv1.DeleteEmployeeResponse, error) {
	// TODO: 实现员工删除逻辑
	return &orgv1.DeleteEmployeeResponse{Success: true}, nil
}

// ListEmployees 列出员工（简化实现）
func (a *OrganizationAdapter) ListEmployees(ctx context.Context, req *orgv1.ListEmployeesRequest) (*orgv1.ListEmployeesResponse, error) {
	// TODO: 实现员工列表查询逻辑
	return &orgv1.ListEmployeesResponse{
		Items: []*orgv1.EmployeeResponse{},
		Total: 0,
	}, nil
}

// 辅助方法
func (a *OrganizationAdapter) orgToProto(org *model.Organization) *orgv1.OrganizationResponse {
	resp := &orgv1.OrganizationResponse{
		Id:        org.ID.String(),
		TenantId:  org.TenantID.String(),
		Name:      org.Name,
		Code:      org.Code,
		Path:      org.Path,
		Level:     int32(org.Level),
		Sort:      int32(org.Sort),
		CreatedAt: org.CreatedAt.Format(time.RFC3339),
		UpdatedAt: org.UpdatedAt.Format(time.RFC3339),
	}

	if org.ParentID != nil {
		resp.ParentId = org.ParentID.String()
	}

	return resp
}

func (a *OrganizationAdapter) buildTreeProtoNodes(treeNodes []*service.OrganizationTreeNode) []*orgv1.OrganizationTreeNode {
	nodes := make([]*orgv1.OrganizationTreeNode, 0, len(treeNodes))
	for _, tn := range treeNodes {
		node := &orgv1.OrganizationTreeNode{
			Id:   tn.ID.String(),
			Name: tn.Name,
			Code: tn.Code,
		}
		if len(tn.Children) > 0 {
			node.Children = a.buildTreeProtoNodes(tn.Children)
		}
		nodes = append(nodes, node)
	}
	return nodes
}
