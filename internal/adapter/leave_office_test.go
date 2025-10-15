package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	hrmv1 "github.com/lk2023060901/go-next-erp/api/hrm/v1"
	"github.com/lk2023060901/go-next-erp/internal/hrm/handler"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MockLeaveOfficeService mocks the LeaveOfficeService interface
type MockLeaveOfficeService struct {
	mock.Mock
}

func (m *MockLeaveOfficeService) Create(ctx context.Context, leaveOffice *model.LeaveOffice) error {
	args := m.Called(ctx, leaveOffice)
	return args.Error(0)
}

func (m *MockLeaveOfficeService) Update(ctx context.Context, leaveOffice *model.LeaveOffice) error {
	args := m.Called(ctx, leaveOffice)
	return args.Error(0)
}

func (m *MockLeaveOfficeService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockLeaveOfficeService) GetByID(ctx context.Context, id uuid.UUID) (*model.LeaveOffice, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.LeaveOffice), args.Error(1)
}

func (m *MockLeaveOfficeService) List(ctx context.Context, tenantID uuid.UUID, filter *repository.LeaveOfficeFilter, offset, limit int) ([]*model.LeaveOffice, int, error) {
	args := m.Called(ctx, tenantID, filter, offset, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.LeaveOffice), args.Int(1), args.Error(2)
}

func (m *MockLeaveOfficeService) ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.LeaveOffice, error) {
	args := m.Called(ctx, tenantID, employeeID, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.LeaveOffice), args.Error(1)
}

func (m *MockLeaveOfficeService) Submit(ctx context.Context, leaveOfficeID, submitterID uuid.UUID) error {
	args := m.Called(ctx, leaveOfficeID, submitterID)
	return args.Error(0)
}

func (m *MockLeaveOfficeService) Approve(ctx context.Context, leaveOfficeID, approverID uuid.UUID, comment string) error {
	args := m.Called(ctx, leaveOfficeID, approverID, comment)
	return args.Error(0)
}

func (m *MockLeaveOfficeService) Reject(ctx context.Context, leaveOfficeID, approverID uuid.UUID, reason string) error {
	args := m.Called(ctx, leaveOfficeID, approverID, reason)
	return args.Error(0)
}

func (m *MockLeaveOfficeService) CheckTimeConflict(ctx context.Context, tenantID, employeeID uuid.UUID, startTime, endTime time.Time, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, tenantID, employeeID, startTime, endTime, excludeID)
	return args.Bool(0), args.Error(1)
}

// TestLeaveOfficeAdapter_CreateLeaveOffice 测试创建外出申请
func TestLeaveOfficeAdapter_CreateLeaveOffice(t *testing.T) {
	t.Run("创建成功", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		tenantID := uuid.New()
		employeeID := uuid.New()
		departmentID := uuid.New()

		mockService.On("Create", mock.Anything, mock.MatchedBy(func(lo *model.LeaveOffice) bool {
			return lo.TenantID == tenantID &&
				lo.EmployeeID == employeeID &&
				lo.Destination == "客户公司" &&
				lo.Purpose == "拜访客户"
		})).Return(nil).Once()

		req := &hrmv1.CreateLeaveOfficeRequest{
			TenantId:     tenantID.String(),
			EmployeeId:   employeeID.String(),
			EmployeeName: "张三",
			DepartmentId: departmentID.String(),
			StartTime:    timestamppb.New(time.Now().Add(2 * time.Hour)),
			EndTime:      timestamppb.New(time.Now().Add(5 * time.Hour)),
			Destination:  "客户公司",
			Purpose:      "拜访客户",
			Contact:      "13800138000",
		}

		resp, err := adapter.CreateLeaveOffice(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "客户公司", resp.Destination)
		assert.Equal(t, "拜访客户", resp.Purpose)
		mockService.AssertExpectations(t)
	})

	t.Run("创建失败_无效EmployeeID", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		req := &hrmv1.CreateLeaveOfficeRequest{
			TenantId:   uuid.New().String(),
			EmployeeId: "invalid-uuid",
		}

		resp, err := adapter.CreateLeaveOffice(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid employee_id")
	})
}

// TestLeaveOfficeAdapter_UpdateLeaveOffice 测试更新外出申请
func TestLeaveOfficeAdapter_UpdateLeaveOffice(t *testing.T) {
	t.Run("更新成功", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		leaveOfficeID := uuid.New()
		existingLeaveOffice := &model.LeaveOffice{
			ID:             leaveOfficeID,
			TenantID:       uuid.New(),
			EmployeeID:     uuid.New(),
			Destination:    "客户公司A",
			ApprovalStatus: "pending",
		}

		mockService.On("GetByID", mock.Anything, leaveOfficeID).Return(existingLeaveOffice, nil).Once()
		mockService.On("Update", mock.Anything, mock.MatchedBy(func(lo *model.LeaveOffice) bool {
			return lo.ID == leaveOfficeID && lo.Destination == "客户公司B"
		})).Return(nil).Once()

		req := &hrmv1.UpdateLeaveOfficeRequest{
			Id:          leaveOfficeID.String(),
			Destination: "客户公司B",
			Purpose:     "变更拜访地点",
		}

		resp, err := adapter.UpdateLeaveOffice(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockService.AssertExpectations(t)
	})

	t.Run("更新失败_无效ID", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		req := &hrmv1.UpdateLeaveOfficeRequest{
			Id: "invalid-uuid",
		}

		resp, err := adapter.UpdateLeaveOffice(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// TestLeaveOfficeAdapter_DeleteLeaveOffice 测试删除外出申请
func TestLeaveOfficeAdapter_DeleteLeaveOffice(t *testing.T) {
	t.Run("删除成功", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		leaveOfficeID := uuid.New()

		mockService.On("Delete", mock.Anything, leaveOfficeID).Return(nil).Once()

		req := &hrmv1.DeleteLeaveOfficeRequest{
			Id: leaveOfficeID.String(),
		}

		resp, err := adapter.DeleteLeaveOffice(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Leave office deleted successfully", resp.Message)
		mockService.AssertExpectations(t)
	})

	t.Run("删除失败_服务错误", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		leaveOfficeID := uuid.New()

		mockService.On("Delete", mock.Anything, leaveOfficeID).Return(assert.AnError).Once()

		req := &hrmv1.DeleteLeaveOfficeRequest{
			Id: leaveOfficeID.String(),
		}

		resp, err := adapter.DeleteLeaveOffice(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, resp.Success)
		mockService.AssertExpectations(t)
	})
}

// TestLeaveOfficeAdapter_GetLeaveOffice 测试获取外出详情
func TestLeaveOfficeAdapter_GetLeaveOffice(t *testing.T) {
	t.Run("获取成功", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		leaveOfficeID := uuid.New()
		leaveOffice := &model.LeaveOffice{
			ID:             leaveOfficeID,
			TenantID:       uuid.New(),
			EmployeeID:     uuid.New(),
			EmployeeName:   "李四",
			Destination:    "银行",
			Duration:       2.5,
			Purpose:        "办理业务",
			ApprovalStatus: "pending",
		}

		mockService.On("GetByID", mock.Anything, leaveOfficeID).Return(leaveOffice, nil).Once()

		req := &hrmv1.GetLeaveOfficeRequest{
			Id: leaveOfficeID.String(),
		}

		resp, err := adapter.GetLeaveOffice(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, leaveOfficeID.String(), resp.Id)
		assert.Equal(t, "李四", resp.EmployeeName)
		assert.Equal(t, "银行", resp.Destination)
		assert.Equal(t, 2.5, resp.Duration)
		mockService.AssertExpectations(t)
	})
}

// TestLeaveOfficeAdapter_ListLeaveOffices 测试列表查询
func TestLeaveOfficeAdapter_ListLeaveOffices(t *testing.T) {
	t.Run("列表查询成功", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		tenantID := uuid.New()
		leaveOffices := []*model.LeaveOffice{
			{
				ID:             uuid.New(),
				TenantID:       tenantID,
				EmployeeID:     uuid.New(),
				EmployeeName:   "张三",
				Destination:    "客户公司",
				ApprovalStatus: "approved",
			},
			{
				ID:             uuid.New(),
				TenantID:       tenantID,
				EmployeeID:     uuid.New(),
				EmployeeName:   "李四",
				Destination:    "银行",
				ApprovalStatus: "pending",
			},
		}

		mockService.On("List", mock.Anything, tenantID, mock.Anything, 0, 10).
			Return(leaveOffices, 2, nil).Once()

		req := &hrmv1.ListLeaveOfficesRequest{
			TenantId: tenantID.String(),
			Page:     1,
			PageSize: 10,
		}

		resp, err := adapter.ListLeaveOffices(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Items, 2)
		assert.Equal(t, int64(2), resp.Total)
		mockService.AssertExpectations(t)
	})

	t.Run("列表查询_带过滤条件", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		tenantID := uuid.New()
		departmentID := uuid.New()

		mockService.On("List", mock.Anything, tenantID, mock.MatchedBy(func(filter *repository.LeaveOfficeFilter) bool {
			return filter.DepartmentID != nil && *filter.DepartmentID == departmentID &&
				filter.ApprovalStatus != nil && *filter.ApprovalStatus == "pending"
		}), 0, 20).Return([]*model.LeaveOffice{}, 0, nil).Once()

		req := &hrmv1.ListLeaveOfficesRequest{
			TenantId:       tenantID.String(),
			DepartmentId:   departmentID.String(),
			ApprovalStatus: "pending",
			Page:           1,
			PageSize:       20,
		}

		resp, err := adapter.ListLeaveOffices(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockService.AssertExpectations(t)
	})
}

// TestLeaveOfficeAdapter_SubmitLeaveOffice 测试提交审批
func TestLeaveOfficeAdapter_SubmitLeaveOffice(t *testing.T) {
	t.Run("提交成功", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		leaveOfficeID := uuid.New()
		submitterID := uuid.New()

		mockService.On("Submit", mock.Anything, leaveOfficeID, submitterID).Return(nil).Once()

		req := &hrmv1.SubmitLeaveOfficeRequest{
			LeaveOfficeId: leaveOfficeID.String(),
			SubmitterId:   submitterID.String(),
		}

		resp, err := adapter.SubmitLeaveOffice(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "successfully")
		mockService.AssertExpectations(t)
	})

	t.Run("提交失败_无效ID", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		req := &hrmv1.SubmitLeaveOfficeRequest{
			LeaveOfficeId: "invalid",
			SubmitterId:   uuid.New().String(),
		}

		resp, err := adapter.SubmitLeaveOffice(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// TestLeaveOfficeAdapter_ApproveLeaveOffice 测试批准外出
func TestLeaveOfficeAdapter_ApproveLeaveOffice(t *testing.T) {
	t.Run("批准成功", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		leaveOfficeID := uuid.New()
		approverID := uuid.New()

		mockService.On("Approve", mock.Anything, leaveOfficeID, approverID, "同意外出").Return(nil).Once()

		req := &hrmv1.ApproveLeaveOfficeRequest{
			LeaveOfficeId: leaveOfficeID.String(),
			ApproverId:    approverID.String(),
			Comment:       "同意外出",
		}

		resp, err := adapter.ApproveLeaveOffice(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})
}

// TestLeaveOfficeAdapter_RejectLeaveOffice 测试拒绝外出
func TestLeaveOfficeAdapter_RejectLeaveOffice(t *testing.T) {
	t.Run("拒绝成功", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		leaveOfficeID := uuid.New()
		approverID := uuid.New()

		mockService.On("Reject", mock.Anything, leaveOfficeID, approverID, "工作繁忙").Return(nil).Once()

		req := &hrmv1.RejectLeaveOfficeRequest{
			LeaveOfficeId: leaveOfficeID.String(),
			ApproverId:    approverID.String(),
			Reason:        "工作繁忙",
		}

		resp, err := adapter.RejectLeaveOffice(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})
}

// TestLeaveOfficeAdapter_Security_DurationLimit 测试时长限制
func TestLeaveOfficeAdapter_Security_DurationLimit(t *testing.T) {
	t.Run("Duration_超过24小时", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		now := time.Now()
		mockService.On("Create", mock.Anything, mock.Anything).Return(assert.AnError).Once()

		req := &hrmv1.CreateLeaveOfficeRequest{
			TenantId:     uuid.New().String(),
			EmployeeId:   uuid.New().String(),
			EmployeeName: "测试",
			DepartmentId: uuid.New().String(),
			StartTime:    timestamppb.New(now),
			EndTime:      timestamppb.New(now.Add(25 * time.Hour)), // 超过24小时
			Destination:  "测试",
			Purpose:      "测试",
		}

		resp, err := adapter.CreateLeaveOffice(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("Duration_正常范围", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		now := time.Now()
		mockService.On("Create", mock.Anything, mock.MatchedBy(func(lo *model.LeaveOffice) bool {
			return lo.Destination == "测试" && lo.Purpose == "测试"
		})).Return(nil).Once()

		req := &hrmv1.CreateLeaveOfficeRequest{
			TenantId:     uuid.New().String(),
			EmployeeId:   uuid.New().String(),
			EmployeeName: "测试",
			DepartmentId: uuid.New().String(),
			StartTime:    timestamppb.New(now),
			EndTime:      timestamppb.New(now.Add(3 * time.Hour)), // 3小时，正常
			Destination:  "测试",
			Purpose:      "测试",
		}

		resp, err := adapter.CreateLeaveOffice(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockService.AssertExpectations(t)
	})
}

// TestLeaveOfficeAdapter_ListEmployeeLeaveOffices 测试查询员工外出记录
func TestLeaveOfficeAdapter_ListEmployeeLeaveOffices(t *testing.T) {
	t.Run("查询成功", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		employeeID := uuid.New()
		leaveOffices := []*model.LeaveOffice{
			{ID: uuid.New(), EmployeeID: employeeID, Destination: "银行"},
		}

		mockService.On("List", mock.Anything, employeeID, mock.Anything, 0, 20).
			Return(leaveOffices, 1, nil).Once()

		req := &hrmv1.ListEmployeeLeaveOfficesRequest{
			EmployeeId: employeeID.String(),
			Page:       1,
			PageSize:   20,
		}

		resp, err := adapter.ListEmployeeLeaveOffices(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Items, 1)
		mockService.AssertExpectations(t)
	})
}

// TestLeaveOfficeAdapter_ListPendingLeaveOffices 测试查询待审批列表
func TestLeaveOfficeAdapter_ListPendingLeaveOffices(t *testing.T) {
	t.Run("查询成功", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		tenantID := uuid.New()
		pendingLeaveOffices := []*model.LeaveOffice{
			{ID: uuid.New(), TenantID: tenantID, ApprovalStatus: "pending"},
			{ID: uuid.New(), TenantID: tenantID, ApprovalStatus: "pending"},
			{ID: uuid.New(), TenantID: tenantID, ApprovalStatus: "pending"},
		}

		mockService.On("List", mock.Anything, tenantID, mock.MatchedBy(func(filter *repository.LeaveOfficeFilter) bool {
			return filter.ApprovalStatus != nil && *filter.ApprovalStatus == "pending"
		}), 0, 20).Return(pendingLeaveOffices, 3, nil).Once()

		req := &hrmv1.ListPendingLeaveOfficesRequest{
			TenantId: tenantID.String(),
			Page:     1,
			PageSize: 20,
		}

		resp, err := adapter.ListPendingLeaveOffices(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Items, 3)
		mockService.AssertExpectations(t)
	})
}

// TestLeaveOfficeAdapter_Security_InvalidUUID 测试无效UUID处理
func TestLeaveOfficeAdapter_Security_InvalidUUID(t *testing.T) {
	t.Run("CreateLeaveOffice_无效DepartmentID", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		req := &hrmv1.CreateLeaveOfficeRequest{
			TenantId:     uuid.New().String(),
			EmployeeId:   uuid.New().String(),
			DepartmentId: "invalid-uuid",
		}

		resp, err := adapter.CreateLeaveOffice(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid department_id")
	})

	t.Run("GetLeaveOffice_无效ID", func(t *testing.T) {
		mockService := new(MockLeaveOfficeService)
		leaveOfficeHandler := handler.NewLeaveOfficeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, leaveOfficeHandler, nil)

		req := &hrmv1.GetLeaveOfficeRequest{
			Id: "not-a-uuid",
		}

		resp, err := adapter.GetLeaveOffice(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
