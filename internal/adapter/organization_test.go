package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	orgv1 "github.com/lk2023060901/go-next-erp/api/organization/v1"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
	"github.com/lk2023060901/go-next-erp/internal/organization/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrganizationService mocks the organization service
type MockOrganizationService struct {
	mock.Mock
}

func (m *MockOrganizationService) Create(ctx context.Context, req *service.CreateOrganizationRequest) (*model.Organization, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *MockOrganizationService) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *MockOrganizationService) Update(ctx context.Context, id uuid.UUID, req *service.UpdateOrganizationRequest) (*model.Organization, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *MockOrganizationService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrganizationService) List(ctx context.Context, tenantID uuid.UUID) ([]*model.Organization, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Organization), args.Error(1)
}

func (m *MockOrganizationService) GetTree(ctx context.Context, tenantID uuid.UUID, parentID *uuid.UUID) ([]*service.OrganizationTreeNode, error) {
	args := m.Called(ctx, tenantID, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*service.OrganizationTreeNode), args.Error(1)
}

func (m *MockOrganizationService) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.Organization, error) {
	args := m.Called(ctx, tenantID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Organization), args.Error(1)
}

func (m *MockOrganizationService) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*model.Organization, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Organization), args.Error(1)
}

func (m *MockOrganizationService) GetDescendants(ctx context.Context, orgID uuid.UUID) ([]*model.Organization, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Organization), args.Error(1)
}

func (m *MockOrganizationService) Move(ctx context.Context, orgID, newParentID uuid.UUID, operatorID uuid.UUID) error {
	args := m.Called(ctx, orgID, newParentID, operatorID)
	return args.Error(0)
}

func (m *MockOrganizationService) UpdateSort(ctx context.Context, orgID uuid.UUID, sort int) error {
	args := m.Called(ctx, orgID, sort)
	return args.Error(0)
}

func (m *MockOrganizationService) UpdateStatus(ctx context.Context, orgID uuid.UUID, status string) error {
	args := m.Called(ctx, orgID, status)
	return args.Error(0)
}

// MockEmployeeService mocks the employee service
type MockEmployeeService struct {
	mock.Mock
}

func (m *MockEmployeeService) Create(ctx context.Context, req *service.CreateEmployeeRequest) (*model.Employee, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Employee), args.Error(1)
}

func (m *MockEmployeeService) Update(ctx context.Context, id uuid.UUID, req *service.UpdateEmployeeRequest) (*model.Employee, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Employee), args.Error(1)
}

func (m *MockEmployeeService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEmployeeService) GetByID(ctx context.Context, id uuid.UUID) (*model.Employee, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Employee), args.Error(1)
}

func (m *MockEmployeeService) GetByUserID(ctx context.Context, tenantID, userID uuid.UUID) (*model.Employee, error) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Employee), args.Error(1)
}

func (m *MockEmployeeService) GetByEmployeeNo(ctx context.Context, tenantID uuid.UUID, employeeNo string) (*model.Employee, error) {
	args := m.Called(ctx, tenantID, employeeNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Employee), args.Error(1)
}

func (m *MockEmployeeService) List(ctx context.Context, tenantID uuid.UUID) ([]*model.Employee, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Employee), args.Error(1)
}

func (m *MockEmployeeService) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*model.Employee, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Employee), args.Error(1)
}

func (m *MockEmployeeService) ListByPosition(ctx context.Context, positionID uuid.UUID) ([]*model.Employee, error) {
	args := m.Called(ctx, positionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Employee), args.Error(1)
}

func (m *MockEmployeeService) ListByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*model.Employee, error) {
	args := m.Called(ctx, tenantID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Employee), args.Error(1)
}

func (m *MockEmployeeService) Transfer(ctx context.Context, empID, newOrgID uuid.UUID, newPositionID *uuid.UUID, operatorID uuid.UUID) error {
	args := m.Called(ctx, empID, newOrgID, newPositionID, operatorID)
	return args.Error(0)
}

func (m *MockEmployeeService) ChangePosition(ctx context.Context, empID uuid.UUID, newPositionID uuid.UUID, operatorID uuid.UUID) error {
	args := m.Called(ctx, empID, newPositionID, operatorID)
	return args.Error(0)
}

func (m *MockEmployeeService) ChangeLeader(ctx context.Context, empID, newLeaderID uuid.UUID, operatorID uuid.UUID) error {
	args := m.Called(ctx, empID, newLeaderID, operatorID)
	return args.Error(0)
}

func (m *MockEmployeeService) Regularize(ctx context.Context, empID uuid.UUID, formalDate time.Time, operatorID uuid.UUID) error {
	args := m.Called(ctx, empID, formalDate, operatorID)
	return args.Error(0)
}

func (m *MockEmployeeService) Resign(ctx context.Context, empID uuid.UUID, leaveDate time.Time, operatorID uuid.UUID) error {
	args := m.Called(ctx, empID, leaveDate, operatorID)
	return args.Error(0)
}

func (m *MockEmployeeService) Reinstate(ctx context.Context, empID uuid.UUID, operatorID uuid.UUID) error {
	args := m.Called(ctx, empID, operatorID)
	return args.Error(0)
}

// MockOrganizationTypeRepository mocks the organization type repository
type MockOrganizationTypeRepository struct {
	mock.Mock
}

func (m *MockOrganizationTypeRepository) Create(ctx context.Context, orgType *model.OrganizationType) error {
	args := m.Called(ctx, orgType)
	return args.Error(0)
}

func (m *MockOrganizationTypeRepository) Update(ctx context.Context, orgType *model.OrganizationType) error {
	args := m.Called(ctx, orgType)
	return args.Error(0)
}

func (m *MockOrganizationTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrganizationTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.OrganizationType, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrganizationType), args.Error(1)
}

func (m *MockOrganizationTypeRepository) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.OrganizationType, error) {
	args := m.Called(ctx, tenantID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrganizationType), args.Error(1)
}

func (m *MockOrganizationTypeRepository) List(ctx context.Context, tenantID uuid.UUID) ([]*model.OrganizationType, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.OrganizationType), args.Error(1)
}

func (m *MockOrganizationTypeRepository) ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.OrganizationType, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.OrganizationType), args.Error(1)
}

func (m *MockOrganizationTypeRepository) Exists(ctx context.Context, tenantID uuid.UUID, code string) (bool, error) {
	args := m.Called(ctx, tenantID, code)
	return args.Get(0).(bool), args.Error(1)
}

// TestOrganizationAdapter_CreateOrganization tests creating organizations
func TestOrganizationAdapter_CreateOrganization(t *testing.T) {
	t.Run("CreateOrganization successfully", func(t *testing.T) {
		mockService := new(MockOrganizationService)
		mockEmpService := new(MockEmployeeService)
		mockTypeRepo := new(MockOrganizationTypeRepository)
		adapter := NewOrganizationAdapter(mockService, mockEmpService, mockTypeRepo)

		tenantID := uuid.New()
		orgID := uuid.New()
		typeID := uuid.New()

		mockTypeRepo.On("GetByCode", mock.Anything, tenantID, "department").
			Return(&model.OrganizationType{
				ID:       typeID,
				TenantID: tenantID,
				Code:     "department",
				Name:     "部门",
			}, nil).Once()

		expectedOrg := &model.Organization{
			ID:        orgID,
			TenantID:  tenantID,
			TypeID:    typeID,
			Code:      "IT001",
			Name:      "IT部门",
			Path:      "/IT001",
			Level:     1,
			Sort:      1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("Create", mock.Anything, mock.AnythingOfType("*service.CreateOrganizationRequest")).
			Return(expectedOrg, nil).Once()

		req := &orgv1.CreateOrganizationRequest{
			TenantId: tenantID.String(),
			Code:     "IT001",
			Name:     "IT部门",
			Type:     "department",
			Sort:     1,
		}

		resp, err := adapter.CreateOrganization(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, orgID.String(), resp.Id)
		assert.Equal(t, "IT部门", resp.Name)
		assert.Equal(t, "IT001", resp.Code)
		mockService.AssertExpectations(t)
		mockTypeRepo.AssertExpectations(t)
	})

	t.Run("CreateOrganization with parent", func(t *testing.T) {
		mockService := new(MockOrganizationService)
		mockEmpService := new(MockEmployeeService)
		mockTypeRepo := new(MockOrganizationTypeRepository)
		adapter := NewOrganizationAdapter(mockService, mockEmpService, mockTypeRepo)

		tenantID := uuid.New()
		parentID := uuid.New()
		orgID := uuid.New()

		expectedOrg := &model.Organization{
			ID:        orgID,
			TenantID:  tenantID,
			Code:      "IT002",
			Name:      "开发组",
			ParentID:  &parentID,
			Path:      "/IT001/IT002",
			Level:     2,
			Sort:      1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("Create", mock.Anything, mock.AnythingOfType("*service.CreateOrganizationRequest")).
			Return(expectedOrg, nil).Once()

		req := &orgv1.CreateOrganizationRequest{
			TenantId: tenantID.String(),
			ParentId: parentID.String(),
			Code:     "IT002",
			Name:     "开发组",
			Sort:     1,
		}

		resp, err := adapter.CreateOrganization(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, orgID.String(), resp.Id)
		assert.Equal(t, parentID.String(), resp.ParentId)
		mockService.AssertExpectations(t)
	})
}

// TestOrganizationAdapter_GetOrganization tests getting an organization
func TestOrganizationAdapter_GetOrganization(t *testing.T) {
	t.Run("GetOrganization successfully", func(t *testing.T) {
		mockService := new(MockOrganizationService)
		mockEmpService := new(MockEmployeeService)
		mockTypeRepo := new(MockOrganizationTypeRepository)
		adapter := NewOrganizationAdapter(mockService, mockEmpService, mockTypeRepo)

		orgID := uuid.New()
		tenantID := uuid.New()

		expectedOrg := &model.Organization{
			ID:        orgID,
			TenantID:  tenantID,
			Code:      "IT001",
			Name:      "IT部门",
			Path:      "/IT001",
			Level:     1,
			Sort:      1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("GetByID", mock.Anything, orgID).
			Return(expectedOrg, nil).Once()

		req := &orgv1.GetOrganizationRequest{
			Id: orgID.String(),
		}

		resp, err := adapter.GetOrganization(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, orgID.String(), resp.Id)
		assert.Equal(t, "IT部门", resp.Name)
		mockService.AssertExpectations(t)
	})
}

// TestOrganizationAdapter_UpdateOrganization tests updating organizations
func TestOrganizationAdapter_UpdateOrganization(t *testing.T) {
	t.Run("UpdateOrganization successfully", func(t *testing.T) {
		mockService := new(MockOrganizationService)
		mockEmpService := new(MockEmployeeService)
		mockTypeRepo := new(MockOrganizationTypeRepository)
		adapter := NewOrganizationAdapter(mockService, mockEmpService, mockTypeRepo)

		orgID := uuid.New()
		tenantID := uuid.New()

		updatedOrg := &model.Organization{
			ID:        orgID,
			TenantID:  tenantID,
			Code:      "IT001",
			Name:      "IT部门-更新",
			Path:      "/IT001",
			Level:     1,
			Sort:      2,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("Update", mock.Anything, orgID, mock.AnythingOfType("*service.UpdateOrganizationRequest")).
			Return(updatedOrg, nil).Once()

		req := &orgv1.UpdateOrganizationRequest{
			Id:   orgID.String(),
			Name: "IT部门-更新",
			Sort: 2,
		}

		resp, err := adapter.UpdateOrganization(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, orgID.String(), resp.Id)
		assert.Equal(t, "IT部门-更新", resp.Name)
		assert.Equal(t, int32(2), resp.Sort)
		mockService.AssertExpectations(t)
	})
}

// TestOrganizationAdapter_DeleteOrganization tests deleting organizations
func TestOrganizationAdapter_DeleteOrganization(t *testing.T) {
	t.Run("DeleteOrganization successfully", func(t *testing.T) {
		mockService := new(MockOrganizationService)
		mockEmpService := new(MockEmployeeService)
		mockTypeRepo := new(MockOrganizationTypeRepository)
		adapter := NewOrganizationAdapter(mockService, mockEmpService, mockTypeRepo)

		orgID := uuid.New()

		mockService.On("Delete", mock.Anything, orgID).
			Return(nil).Once()

		req := &orgv1.DeleteOrganizationRequest{
			Id: orgID.String(),
		}

		resp, err := adapter.DeleteOrganization(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})
}

// TestOrganizationAdapter_ListOrganizations tests listing organizations
func TestOrganizationAdapter_ListOrganizations(t *testing.T) {
	t.Run("ListOrganizations successfully", func(t *testing.T) {
		mockService := new(MockOrganizationService)
		mockEmpService := new(MockEmployeeService)
		mockTypeRepo := new(MockOrganizationTypeRepository)
		adapter := NewOrganizationAdapter(mockService, mockEmpService, mockTypeRepo)

		tenantID := uuid.New()

		expectedOrgs := []*model.Organization{
			{
				ID:        uuid.New(),
				TenantID:  tenantID,
				Code:      "IT001",
				Name:      "IT部门",
				Path:      "/IT001",
				Level:     1,
				Sort:      1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:        uuid.New(),
				TenantID:  tenantID,
				Code:      "HR001",
				Name:      "HR部门",
				Path:      "/HR001",
				Level:     1,
				Sort:      2,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mockService.On("List", mock.Anything, tenantID).
			Return(expectedOrgs, nil).Once()

		req := &orgv1.ListOrganizationsRequest{
			TenantId: tenantID.String(),
		}

		resp, err := adapter.ListOrganizations(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Items, 2)
		assert.Equal(t, int32(2), resp.Total)
		mockService.AssertExpectations(t)
	})

	t.Run("ListOrganizations with parent filter", func(t *testing.T) {
		mockService := new(MockOrganizationService)
		mockEmpService := new(MockEmployeeService)
		mockTypeRepo := new(MockOrganizationTypeRepository)
		adapter := NewOrganizationAdapter(mockService, mockEmpService, mockTypeRepo)

		tenantID := uuid.New()
		parentID := uuid.New()

		expectedOrgs := []*model.Organization{
			{
				ID:        uuid.New(),
				TenantID:  tenantID,
				ParentID:  &parentID,
				Code:      "IT002",
				Name:      "开发组",
				Path:      "/IT001/IT002",
				Level:     2,
				Sort:      1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mockService.On("List", mock.Anything, tenantID).
			Return(expectedOrgs, nil).Once()

		req := &orgv1.ListOrganizationsRequest{
			TenantId: tenantID.String(),
			ParentId: parentID.String(),
		}

		resp, err := adapter.ListOrganizations(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Items, 1)
		mockService.AssertExpectations(t)
	})
}

// TestOrganizationAdapter_GetOrganizationTree tests getting organization tree
func TestOrganizationAdapter_GetOrganizationTree(t *testing.T) {
	t.Run("GetOrganizationTree successfully", func(t *testing.T) {
		mockService := new(MockOrganizationService)
		mockEmpService := new(MockEmployeeService)
		mockTypeRepo := new(MockOrganizationTypeRepository)
		adapter := NewOrganizationAdapter(mockService, mockEmpService, mockTypeRepo)

		tenantID := uuid.New()
		rootID := uuid.New()
		childID := uuid.New()

		expectedTree := []*service.OrganizationTreeNode{
			{
				Organization: &model.Organization{
					ID:        rootID,
					TenantID:  tenantID,
					Name:      "IT部门",
					Code:      "IT001",
					Path:      "/IT001",
					Level:     1,
					Sort:      1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Children: []*service.OrganizationTreeNode{
					{
						Organization: &model.Organization{
							ID:        childID,
							TenantID:  tenantID,
							Name:      "开发组",
							Code:      "IT002",
							Path:      "/IT001/IT002",
							Level:     2,
							Sort:      1,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						Children: []*service.OrganizationTreeNode{},
					},
				},
			},
		}

		mockService.On("GetTree", mock.Anything, tenantID, (*uuid.UUID)(nil)).
			Return(expectedTree, nil).Once()

		req := &orgv1.GetOrganizationTreeRequest{
			TenantId: tenantID.String(),
		}

		resp, err := adapter.GetOrganizationTree(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Nodes, 1)
		assert.Equal(t, "IT部门", resp.Nodes[0].Name)
		assert.Len(t, resp.Nodes[0].Children, 1)
		assert.Equal(t, "开发组", resp.Nodes[0].Children[0].Name)
		mockService.AssertExpectations(t)
	})
}

// TestOrganizationAdapter_CreateEmployee tests creating employees
func TestOrganizationAdapter_CreateEmployee(t *testing.T) {
	t.Run("CreateEmployee successfully", func(t *testing.T) {
		mockService := new(MockOrganizationService)
		mockEmpService := new(MockEmployeeService)
		mockTypeRepo := new(MockOrganizationTypeRepository)
		adapter := NewOrganizationAdapter(mockService, mockEmpService, mockTypeRepo)

		tenantID := uuid.New()
		userID := uuid.New()
		orgID := uuid.New()
		empID := uuid.New()

		expectedEmp := &model.Employee{
			ID:         empID,
			TenantID:   tenantID,
			UserID:     userID,
			OrgID:      orgID,
			EmployeeNo: "EMP001",
			Name:       "张三",
			Mobile:     "13800138000",
			Email:      "zhangsan@example.com",
			Status:     "active",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockEmpService.On("Create", mock.Anything, mock.AnythingOfType("*service.CreateEmployeeRequest")).
			Return(expectedEmp, nil).Once()

		req := &orgv1.CreateEmployeeRequest{
			TenantId:   tenantID.String(),
			UserId:     userID.String(),
			OrgId:      orgID.String(),
			EmployeeNo: "EMP001",
			Name:       "张三",
			Mobile:     "13800138000",
			Email:      "zhangsan@example.com",
			Status:     "active",
		}

		resp, err := adapter.CreateEmployee(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, empID.String(), resp.Id)
		assert.Equal(t, "张三", resp.Name)
		assert.Equal(t, "EMP001", resp.EmployeeNo)
		mockEmpService.AssertExpectations(t)
	})
}

// TestOrganizationAdapter_GetEmployee tests getting an employee
func TestOrganizationAdapter_GetEmployee(t *testing.T) {
	t.Run("GetEmployee successfully", func(t *testing.T) {
		mockService := new(MockOrganizationService)
		mockEmpService := new(MockEmployeeService)
		mockTypeRepo := new(MockOrganizationTypeRepository)
		adapter := NewOrganizationAdapter(mockService, mockEmpService, mockTypeRepo)

		empID := uuid.New()

		req := &orgv1.GetEmployeeRequest{
			Id: empID.String(),
		}

		resp, err := adapter.GetEmployee(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, empID.String(), resp.Id)
	})
}

// TestOrganizationAdapter_UpdateEmployee tests updating employees
func TestOrganizationAdapter_UpdateEmployee(t *testing.T) {
	t.Run("UpdateEmployee successfully", func(t *testing.T) {
		mockService := new(MockOrganizationService)
		mockEmpService := new(MockEmployeeService)
		mockTypeRepo := new(MockOrganizationTypeRepository)
		adapter := NewOrganizationAdapter(mockService, mockEmpService, mockTypeRepo)

		empID := uuid.New()

		req := &orgv1.UpdateEmployeeRequest{
			Id:     empID.String(),
			Name:   "李四",
			Mobile: "13900139000",
			Email:  "lisi@example.com",
			Status: "active",
		}

		resp, err := adapter.UpdateEmployee(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, empID.String(), resp.Id)
		assert.Equal(t, "李四", resp.Name)
	})
}

// TestOrganizationAdapter_DeleteEmployee tests deleting employees
func TestOrganizationAdapter_DeleteEmployee(t *testing.T) {
	t.Run("DeleteEmployee successfully", func(t *testing.T) {
		mockService := new(MockOrganizationService)
		mockEmpService := new(MockEmployeeService)
		mockTypeRepo := new(MockOrganizationTypeRepository)
		adapter := NewOrganizationAdapter(mockService, mockEmpService, mockTypeRepo)

		req := &orgv1.DeleteEmployeeRequest{
			Id: uuid.New().String(),
		}

		resp, err := adapter.DeleteEmployee(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)
	})
}

// TestOrganizationAdapter_ListEmployees tests listing employees
func TestOrganizationAdapter_ListEmployees(t *testing.T) {
	t.Run("ListEmployees successfully", func(t *testing.T) {
		mockService := new(MockOrganizationService)
		mockEmpService := new(MockEmployeeService)
		mockTypeRepo := new(MockOrganizationTypeRepository)
		adapter := NewOrganizationAdapter(mockService, mockEmpService, mockTypeRepo)

		req := &orgv1.ListEmployeesRequest{
			OrgId: uuid.New().String(),
		}

		resp, err := adapter.ListEmployees(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Empty(t, resp.Items)
		assert.Equal(t, int32(0), resp.Total)
	})
}
