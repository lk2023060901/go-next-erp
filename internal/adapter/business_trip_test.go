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

// MockBusinessTripService mocks the BusinessTripService interface
type MockBusinessTripService struct {
	mock.Mock
}

func (m *MockBusinessTripService) Create(ctx context.Context, trip *model.BusinessTrip) error {
	args := m.Called(ctx, trip)
	return args.Error(0)
}

func (m *MockBusinessTripService) Update(ctx context.Context, trip *model.BusinessTrip) error {
	args := m.Called(ctx, trip)
	return args.Error(0)
}

func (m *MockBusinessTripService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBusinessTripService) GetByID(ctx context.Context, id uuid.UUID) (*model.BusinessTrip, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.BusinessTrip), args.Error(1)
}

func (m *MockBusinessTripService) List(ctx context.Context, tenantID uuid.UUID, filter *repository.BusinessTripFilter, offset, limit int) ([]*model.BusinessTrip, int, error) {
	args := m.Called(ctx, tenantID, filter, offset, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.BusinessTrip), args.Int(1), args.Error(2)
}

func (m *MockBusinessTripService) ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.BusinessTrip, error) {
	args := m.Called(ctx, tenantID, employeeID, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.BusinessTrip), args.Error(1)
}

func (m *MockBusinessTripService) ListPending(ctx context.Context, tenantID uuid.UUID) ([]*model.BusinessTrip, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.BusinessTrip), args.Error(1)
}

func (m *MockBusinessTripService) Submit(ctx context.Context, tripID, submitterID uuid.UUID) error {
	args := m.Called(ctx, tripID, submitterID)
	return args.Error(0)
}

func (m *MockBusinessTripService) Approve(ctx context.Context, tripID, approverID uuid.UUID, comment string) error {
	args := m.Called(ctx, tripID, approverID, comment)
	return args.Error(0)
}

func (m *MockBusinessTripService) Reject(ctx context.Context, tripID, approverID uuid.UUID, reason string) error {
	args := m.Called(ctx, tripID, approverID, reason)
	return args.Error(0)
}

func (m *MockBusinessTripService) SubmitReport(ctx context.Context, tripID uuid.UUID, report string, actualCost float64) error {
	args := m.Called(ctx, tripID, report, actualCost)
	return args.Error(0)
}

func (m *MockBusinessTripService) CheckTimeConflict(ctx context.Context, tenantID, employeeID uuid.UUID, startTime, endTime time.Time, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, tenantID, employeeID, startTime, endTime, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockBusinessTripService) SumDaysByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	args := m.Called(ctx, tenantID, employeeID, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

// TestBusinessTripAdapter_CreateBusinessTrip 测试创建出差申请
func TestBusinessTripAdapter_CreateBusinessTrip(t *testing.T) {
	t.Run("创建成功", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		tenantID := uuid.New()
		employeeID := uuid.New()
		departmentID := uuid.New()

		mockService.On("Create", mock.Anything, mock.MatchedBy(func(trip *model.BusinessTrip) bool {
			return trip.TenantID == tenantID &&
				trip.EmployeeID == employeeID &&
				trip.Destination == "北京" &&
				trip.EstimatedCost == 5000.0
		})).Return(nil).Once()

		req := &hrmv1.CreateBusinessTripRequest{
			TenantId:       tenantID.String(),
			EmployeeId:     employeeID.String(),
			EmployeeName:   "张三",
			DepartmentId:   departmentID.String(),
			StartTime:      timestamppb.New(time.Now().Add(24 * time.Hour)),
			EndTime:        timestamppb.New(time.Now().Add(72 * time.Hour)),
			Destination:    "北京",
			Transportation: "高铁",
			Accommodation:  "商务酒店",
			Purpose:        "客户拜访",
			EstimatedCost:  5000.0,
		}

		resp, err := adapter.CreateBusinessTrip(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "北京", resp.Destination)
		mockService.AssertExpectations(t)
	})

	t.Run("创建失败_无效TenantID", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		req := &hrmv1.CreateBusinessTripRequest{
			TenantId:   "invalid-uuid",
			EmployeeId: uuid.New().String(),
		}

		resp, err := adapter.CreateBusinessTrip(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid tenant_id")
	})
}

// TestBusinessTripAdapter_UpdateBusinessTrip 测试更新出差申请
func TestBusinessTripAdapter_UpdateBusinessTrip(t *testing.T) {
	t.Run("更新成功", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		tripID := uuid.New()
		existingTrip := &model.BusinessTrip{
			ID:             tripID,
			TenantID:       uuid.New(),
			EmployeeID:     uuid.New(),
			Destination:    "北京",
			EstimatedCost:  5000.0,
			ApprovalStatus: "pending",
		}

		mockService.On("GetByID", mock.Anything, tripID).Return(existingTrip, nil).Once()
		mockService.On("Update", mock.Anything, mock.MatchedBy(func(trip *model.BusinessTrip) bool {
			return trip.ID == tripID && trip.EstimatedCost == 6000.0
		})).Return(nil).Once()

		req := &hrmv1.UpdateBusinessTripRequest{
			Id:            tripID.String(),
			EstimatedCost: 6000.0,
			Remark:        "预算调整",
		}

		resp, err := adapter.UpdateBusinessTrip(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockService.AssertExpectations(t)
	})

	t.Run("更新失败_无效ID", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		req := &hrmv1.UpdateBusinessTripRequest{
			Id: "invalid-uuid",
		}

		resp, err := adapter.UpdateBusinessTrip(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// TestBusinessTripAdapter_DeleteBusinessTrip 测试删除出差申请
func TestBusinessTripAdapter_DeleteBusinessTrip(t *testing.T) {
	t.Run("删除成功", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		tripID := uuid.New()

		mockService.On("Delete", mock.Anything, tripID).Return(nil).Once()

		req := &hrmv1.DeleteBusinessTripRequest{
			Id: tripID.String(),
		}

		resp, err := adapter.DeleteBusinessTrip(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Business trip deleted successfully", resp.Message)
		mockService.AssertExpectations(t)
	})

	t.Run("删除失败_服务错误", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		tripID := uuid.New()

		mockService.On("Delete", mock.Anything, tripID).Return(assert.AnError).Once()

		req := &hrmv1.DeleteBusinessTripRequest{
			Id: tripID.String(),
		}

		resp, err := adapter.DeleteBusinessTrip(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, resp.Success)
		mockService.AssertExpectations(t)
	})
}

// TestBusinessTripAdapter_GetBusinessTrip 测试获取出差详情
func TestBusinessTripAdapter_GetBusinessTrip(t *testing.T) {
	t.Run("获取成功", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		tripID := uuid.New()
		trip := &model.BusinessTrip{
			ID:             tripID,
			TenantID:       uuid.New(),
			EmployeeID:     uuid.New(),
			EmployeeName:   "张三",
			Destination:    "北京",
			Duration:       3.0,
			EstimatedCost:  5000.0,
			ApprovalStatus: "pending",
		}

		mockService.On("GetByID", mock.Anything, tripID).Return(trip, nil).Once()

		req := &hrmv1.GetBusinessTripRequest{
			Id: tripID.String(),
		}

		resp, err := adapter.GetBusinessTrip(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, tripID.String(), resp.Id)
		assert.Equal(t, "张三", resp.EmployeeName)
		assert.Equal(t, "北京", resp.Destination)
		assert.Equal(t, 5000.0, resp.EstimatedCost)
		mockService.AssertExpectations(t)
	})
}

// TestBusinessTripAdapter_ListBusinessTrips 测试列表查询
func TestBusinessTripAdapter_ListBusinessTrips(t *testing.T) {
	t.Run("列表查询成功", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		tenantID := uuid.New()
		trips := []*model.BusinessTrip{
			{
				ID:             uuid.New(),
				TenantID:       tenantID,
				EmployeeID:     uuid.New(),
				EmployeeName:   "张三",
				Destination:    "北京",
				ApprovalStatus: "approved",
			},
			{
				ID:             uuid.New(),
				TenantID:       tenantID,
				EmployeeID:     uuid.New(),
				EmployeeName:   "李四",
				Destination:    "上海",
				ApprovalStatus: "pending",
			},
		}

		mockService.On("List", mock.Anything, tenantID, mock.Anything, 0, 10).
			Return(trips, 2, nil).Once()

		req := &hrmv1.ListBusinessTripsRequest{
			TenantId: tenantID.String(),
			Page:     1,
			PageSize: 10,
		}

		resp, err := adapter.ListBusinessTrips(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Items, 2)
		assert.Equal(t, int64(2), resp.Total)
		mockService.AssertExpectations(t)
	})

	t.Run("列表查询_带过滤条件", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		tenantID := uuid.New()
		employeeID := uuid.New()

		mockService.On("List", mock.Anything, tenantID, mock.MatchedBy(func(filter *repository.BusinessTripFilter) bool {
			return filter.EmployeeID != nil && *filter.EmployeeID == employeeID &&
				filter.ApprovalStatus != nil && *filter.ApprovalStatus == "approved"
		}), 0, 20).Return([]*model.BusinessTrip{}, 0, nil).Once()

		req := &hrmv1.ListBusinessTripsRequest{
			TenantId:       tenantID.String(),
			EmployeeId:     employeeID.String(),
			ApprovalStatus: "approved",
			Page:           1,
			PageSize:       20,
		}

		resp, err := adapter.ListBusinessTrips(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockService.AssertExpectations(t)
	})
}

// TestBusinessTripAdapter_SubmitBusinessTrip 测试提交审批
func TestBusinessTripAdapter_SubmitBusinessTrip(t *testing.T) {
	t.Run("提交成功", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		tripID := uuid.New()
		submitterID := uuid.New()

		mockService.On("Submit", mock.Anything, tripID, submitterID).Return(nil).Once()

		req := &hrmv1.SubmitBusinessTripRequest{
			BusinessTripId: tripID.String(),
			SubmitterId:    submitterID.String(),
		}

		resp, err := adapter.SubmitBusinessTrip(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "successfully")
		mockService.AssertExpectations(t)
	})

	t.Run("提交失败_无效ID", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		req := &hrmv1.SubmitBusinessTripRequest{
			BusinessTripId: "invalid",
			SubmitterId:    uuid.New().String(),
		}

		resp, err := adapter.SubmitBusinessTrip(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// TestBusinessTripAdapter_ApproveBusinessTrip 测试批准出差
func TestBusinessTripAdapter_ApproveBusinessTrip(t *testing.T) {
	t.Run("批准成功", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		tripID := uuid.New()
		approverID := uuid.New()

		mockService.On("Approve", mock.Anything, tripID, approverID, "同意出差").Return(nil).Once()

		req := &hrmv1.ApproveBusinessTripRequest{
			BusinessTripId: tripID.String(),
			ApproverId:     approverID.String(),
			Comment:        "同意出差",
		}

		resp, err := adapter.ApproveBusinessTrip(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})
}

// TestBusinessTripAdapter_RejectBusinessTrip 测试拒绝出差
func TestBusinessTripAdapter_RejectBusinessTrip(t *testing.T) {
	t.Run("拒绝成功", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		tripID := uuid.New()
		approverID := uuid.New()

		mockService.On("Reject", mock.Anything, tripID, approverID, "时间冲突").Return(nil).Once()

		req := &hrmv1.RejectBusinessTripRequest{
			BusinessTripId: tripID.String(),
			ApproverId:     approverID.String(),
			Reason:         "时间冲突",
		}

		resp, err := adapter.RejectBusinessTrip(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})
}

// TestBusinessTripAdapter_SubmitTripReport 测试提交出差报告
func TestBusinessTripAdapter_SubmitTripReport(t *testing.T) {
	t.Run("提交报告成功", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		tripID := uuid.New()
		trip := &model.BusinessTrip{
			ID:             tripID,
			TenantID:       uuid.New(),
			EmployeeID:     uuid.New(),
			ApprovalStatus: "approved",
		}

		mockService.On("SubmitReport", mock.Anything, tripID, "已完成客户拜访", 4500.0).Return(nil).Once()
		mockService.On("GetByID", mock.Anything, tripID).Return(trip, nil).Once()

		req := &hrmv1.SubmitTripReportRequest{
			BusinessTripId: tripID.String(),
			Report:         "已完成客户拜访",
			ActualCost:     4500.0,
		}

		resp, err := adapter.SubmitTripReport(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockService.AssertExpectations(t)
	})
}

// TestBusinessTripAdapter_Security_BoundaryValues 测试边界值
func TestBusinessTripAdapter_Security_BoundaryValues(t *testing.T) {
	t.Run("EstimatedCost_负数", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		mockService.On("Create", mock.Anything, mock.MatchedBy(func(trip *model.BusinessTrip) bool {
			return trip.EstimatedCost == -1000.0
		})).Return(assert.AnError).Once()

		req := &hrmv1.CreateBusinessTripRequest{
			TenantId:      uuid.New().String(),
			EmployeeId:    uuid.New().String(),
			EmployeeName:  "测试",
			DepartmentId:  uuid.New().String(),
			StartTime:     timestamppb.New(time.Now().Add(24 * time.Hour)),
			EndTime:       timestamppb.New(time.Now().Add(48 * time.Hour)),
			Destination:   "测试",
			EstimatedCost: -1000.0, // 负数预算
		}

		resp, err := adapter.CreateBusinessTrip(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("Duration_零值", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		now := time.Now()
		mockService.On("Create", mock.Anything, mock.Anything).Return(assert.AnError).Once()

		req := &hrmv1.CreateBusinessTripRequest{
			TenantId:     uuid.New().String(),
			EmployeeId:   uuid.New().String(),
			EmployeeName: "测试",
			DepartmentId: uuid.New().String(),
			StartTime:    timestamppb.New(now),
			EndTime:      timestamppb.New(now), // 开始时间=结束时间
			Destination:  "测试",
		}

		resp, err := adapter.CreateBusinessTrip(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// TestBusinessTripAdapter_ListEmployeeBusinessTrips 测试查询员工出差记录
func TestBusinessTripAdapter_ListEmployeeBusinessTrips(t *testing.T) {
	t.Run("查询成功", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		employeeID := uuid.New()
		trips := []*model.BusinessTrip{
			{ID: uuid.New(), EmployeeID: employeeID, Destination: "北京"},
		}

		mockService.On("List", mock.Anything, employeeID, mock.Anything, 0, 20).
			Return(trips, 1, nil).Once()

		req := &hrmv1.ListEmployeeBusinessTripsRequest{
			EmployeeId: employeeID.String(),
			Page:       1,
			PageSize:   20,
		}

		resp, err := adapter.ListEmployeeBusinessTrips(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Items, 1)
		mockService.AssertExpectations(t)
	})
}

// TestBusinessTripAdapter_ListPendingBusinessTrips 测试查询待审批列表
func TestBusinessTripAdapter_ListPendingBusinessTrips(t *testing.T) {
	t.Run("查询成功", func(t *testing.T) {
		mockService := new(MockBusinessTripService)
		tripHandler := handler.NewBusinessTripHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, tripHandler, nil, nil)

		tenantID := uuid.New()
		pendingTrips := []*model.BusinessTrip{
			{ID: uuid.New(), TenantID: tenantID, ApprovalStatus: "pending"},
			{ID: uuid.New(), TenantID: tenantID, ApprovalStatus: "pending"},
		}

		mockService.On("List", mock.Anything, tenantID, mock.MatchedBy(func(filter *repository.BusinessTripFilter) bool {
			return filter.ApprovalStatus != nil && *filter.ApprovalStatus == "pending"
		}), 0, 20).Return(pendingTrips, 2, nil).Once()

		req := &hrmv1.ListPendingBusinessTripsRequest{
			TenantId: tenantID.String(),
			Page:     1,
			PageSize: 20,
		}

		resp, err := adapter.ListPendingBusinessTrips(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Items, 2)
		mockService.AssertExpectations(t)
	})
}
