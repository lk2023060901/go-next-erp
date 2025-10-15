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

// MockOvertimeService mocks the OvertimeService interface
type MockOvertimeService struct {
	mock.Mock
}

func (m *MockOvertimeService) Create(ctx context.Context, overtime *model.Overtime) error {
	args := m.Called(ctx, overtime)
	return args.Error(0)
}

func (m *MockOvertimeService) Update(ctx context.Context, overtime *model.Overtime) error {
	args := m.Called(ctx, overtime)
	return args.Error(0)
}

func (m *MockOvertimeService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOvertimeService) GetByID(ctx context.Context, id uuid.UUID) (*model.Overtime, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Overtime), args.Error(1)
}

func (m *MockOvertimeService) List(ctx context.Context, tenantID uuid.UUID, filter *repository.OvertimeFilter, offset, limit int) ([]*model.Overtime, int, error) {
	args := m.Called(ctx, tenantID, filter, offset, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Overtime), args.Int(1), args.Error(2)
}

func (m *MockOvertimeService) ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.Overtime, error) {
	args := m.Called(ctx, tenantID, employeeID, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Overtime), args.Error(1)
}

func (m *MockOvertimeService) ListPending(ctx context.Context, tenantID uuid.UUID) ([]*model.Overtime, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Overtime), args.Error(1)
}

func (m *MockOvertimeService) Submit(ctx context.Context, overtimeID, submitterID uuid.UUID) error {
	args := m.Called(ctx, overtimeID, submitterID)
	return args.Error(0)
}

func (m *MockOvertimeService) Approve(ctx context.Context, overtimeID, approverID uuid.UUID) error {
	args := m.Called(ctx, overtimeID, approverID)
	return args.Error(0)
}

func (m *MockOvertimeService) Reject(ctx context.Context, overtimeID, approverID uuid.UUID, reason string) error {
	args := m.Called(ctx, overtimeID, approverID, reason)
	return args.Error(0)
}

func (m *MockOvertimeService) SumHoursByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	args := m.Called(ctx, tenantID, employeeID, startDate, endDate)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockOvertimeService) SumCompOffDays(ctx context.Context, tenantID, employeeID uuid.UUID) (float64, error) {
	args := m.Called(ctx, tenantID, employeeID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockOvertimeService) UseCompOffDays(ctx context.Context, tenantID, employeeID uuid.UUID, days float64) error {
	args := m.Called(ctx, tenantID, employeeID, days)
	return args.Error(0)
}

// TestOvertimeAdapter_CreateOvertime 测试创建加班申请
func TestOvertimeAdapter_CreateOvertime(t *testing.T) {
	t.Run("创建成功", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		tenantID := uuid.New()
		employeeID := uuid.New()

		mockService.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		req := &hrmv1.CreateOvertimeRequest{
			TenantId:     tenantID.String(),
			EmployeeId:   employeeID.String(),
			EmployeeName: "张三",
			DepartmentId: uuid.New().String(),
			StartTime:    timestamppb.New(time.Now()),
			EndTime:      timestamppb.New(time.Now().Add(3 * time.Hour)),
			Duration:     3.0,
			OvertimeType: "workday",
			PayType:      "money",
			Reason:       "项目紧急",
		}

		resp, err := adapter.CreateOvertime(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockService.AssertExpectations(t)
	})
}

// TestOvertimeAdapter_ApproveOvertime 测试批准加班
func TestOvertimeAdapter_ApproveOvertime(t *testing.T) {
	t.Run("批准成功", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		overtimeID := uuid.New()
		approverID := uuid.New()

		mockService.On("Approve", mock.Anything, overtimeID, approverID).Return(nil).Once()

		req := &hrmv1.ApproveOvertimeRequest{
			OvertimeId: overtimeID.String(),
			ApproverId: approverID.String(),
		}

		resp, err := adapter.ApproveOvertime(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})
}

// TestOvertimeAdapter_UseCompOffDays 测试使用调休
func TestOvertimeAdapter_UseCompOffDays(t *testing.T) {
	t.Run("使用调休成功", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		tenantID := uuid.New()
		employeeID := uuid.New()

		mockService.On("UseCompOffDays", mock.Anything, tenantID, employeeID, 1.5).Return(nil).Once()
		mockService.On("SumCompOffDays", mock.Anything, tenantID, employeeID).Return(2.0, nil).Once()

		req := &hrmv1.UseCompOffDaysRequest{
			TenantId:   tenantID.String(),
			EmployeeId: employeeID.String(),
			Days:       1.5,
		}

		resp, err := adapter.UseCompOffDays(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, 2.0, resp.RemainingDays)
		mockService.AssertExpectations(t)
	})
}

// TestOvertimeAdapter_ListOvertimes 测试列表查询
func TestOvertimeAdapter_ListOvertimes(t *testing.T) {
	t.Run("列表查询成功", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		tenantID := uuid.New()
		overtimes := []*model.Overtime{
			{
				ID:             uuid.New(),
				TenantID:       tenantID,
				EmployeeID:     uuid.New(),
				EmployeeName:   "张三",
				Duration:       3.0,
				OvertimeType:   model.OvertimeTypeWorkday,
				ApprovalStatus: "approved",
			},
		}

		mockService.On("List", mock.Anything, tenantID, mock.Anything, 0, 10).
			Return(overtimes, 1, nil).Once()

		req := &hrmv1.ListOvertimesRequest{
			TenantId: tenantID.String(),
			Page:     1,
			PageSize: 10,
		}

		resp, err := adapter.ListOvertimes(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Items, 1)
		assert.Equal(t, int32(1), resp.Total)
		mockService.AssertExpectations(t)
	})
}

// TestOvertimeAdapter_GetOvertime 测试获取详情
func TestOvertimeAdapter_GetOvertime(t *testing.T) {
	t.Run("获取详情成功", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		overtimeID := uuid.New()
		overtime := &model.Overtime{
			ID:             overtimeID,
			TenantID:       uuid.New(),
			EmployeeID:     uuid.New(),
			EmployeeName:   "张三",
			Duration:       3.0,
			OvertimeType:   model.OvertimeTypeWorkday,
			PayRate:        1.5,
			ApprovalStatus: "pending",
		}

		mockService.On("GetByID", mock.Anything, overtimeID).Return(overtime, nil).Once()

		req := &hrmv1.GetOvertimeRequest{
			Id: overtimeID.String(),
		}

		resp, err := adapter.GetOvertime(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, overtimeID.String(), resp.Id)
		assert.Equal(t, "张三", resp.EmployeeName)
		assert.Equal(t, 1.5, resp.PayRate)
		mockService.AssertExpectations(t)
	})
}

// TestOvertimeAdapter_RejectOvertime 测试拒绝加班
func TestOvertimeAdapter_RejectOvertime(t *testing.T) {
	t.Run("拒绝成功", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		overtimeID := uuid.New()
		approverID := uuid.New()

		mockService.On("Reject", mock.Anything, overtimeID, approverID, "不符合条件").Return(nil).Once()

		req := &hrmv1.RejectOvertimeRequest{
			OvertimeId: overtimeID.String(),
			ApproverId: approverID.String(),
			Reason:     "不符合条件",
		}

		resp, err := adapter.RejectOvertime(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})
}

// TestOvertimeAdapter_Security_InvalidUUID 测试无效UUID处理
func TestOvertimeAdapter_Security_InvalidUUID(t *testing.T) {
	t.Run("CreateOvertime_无效TenantID", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		req := &hrmv1.CreateOvertimeRequest{
			TenantId:   "invalid-uuid",
			EmployeeId: uuid.New().String(),
		}

		resp, err := adapter.CreateOvertime(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid UUID")
	})

	t.Run("GetOvertime_无效ID", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		req := &hrmv1.GetOvertimeRequest{
			Id: "not-a-uuid",
		}

		resp, err := adapter.GetOvertime(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// TestOvertimeAdapter_Security_BoundaryValues 测试边界值
func TestOvertimeAdapter_Security_BoundaryValues(t *testing.T) {
	t.Run("UseCompOffDays_负数天数", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		tenantID := uuid.New()
		employeeID := uuid.New()

		mockService.On("UseCompOffDays", mock.Anything, tenantID, employeeID, -1.0).Return(assert.AnError).Once()
		mockService.On("SumCompOffDays", mock.Anything, tenantID, employeeID).Return(3.0, nil).Maybe()

		req := &hrmv1.UseCompOffDaysRequest{
			TenantId:   tenantID.String(),
			EmployeeId: employeeID.String(),
			Days:       -1.0,
		}

		resp, err := adapter.UseCompOffDays(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, resp.Success)
	})
}

// TestOvertimeAdapter_UpdateOvertime 测试更新加班申请
func TestOvertimeAdapter_UpdateOvertime(t *testing.T) {
	t.Run("更新成功", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		overtimeID := uuid.New()
		existingOvertime := &model.Overtime{
			ID:           overtimeID,
			TenantID:     uuid.New(),
			EmployeeID:   uuid.New(),
			EmployeeName: "张三",
			Duration:     3.0,
			OvertimeType: model.OvertimeTypeWorkday,
		}

		mockService.On("GetByID", mock.Anything, overtimeID).Return(existingOvertime, nil).Once()
		mockService.On("Update", mock.Anything, mock.MatchedBy(func(o *model.Overtime) bool {
			return o.ID == overtimeID && o.Duration == 4.0
		})).Return(nil).Once()

		req := &hrmv1.UpdateOvertimeRequest{
			Id:       overtimeID.String(),
			Duration: 4.0,
			Reason:   "延长加班时间",
		}

		resp, err := adapter.UpdateOvertime(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 4.0, resp.Duration)
		mockService.AssertExpectations(t)
	})

	t.Run("更新失败_无效ID", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		req := &hrmv1.UpdateOvertimeRequest{
			Id:       "invalid-uuid",
			Duration: 4.0,
		}

		resp, err := adapter.UpdateOvertime(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// TestOvertimeAdapter_DeleteOvertime 测试删除加班申请
func TestOvertimeAdapter_DeleteOvertime(t *testing.T) {
	t.Run("删除成功", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		overtimeID := uuid.New()

		mockService.On("Delete", mock.Anything, overtimeID).Return(nil).Once()

		req := &hrmv1.DeleteOvertimeRequest{
			Id: overtimeID.String(),
		}

		resp, err := adapter.DeleteOvertime(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Overtime deleted successfully", resp.Message)
		mockService.AssertExpectations(t)
	})

	t.Run("删除失败_无效ID", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		req := &hrmv1.DeleteOvertimeRequest{
			Id: "invalid-uuid",
		}

		resp, err := adapter.DeleteOvertime(context.Background(), req)

		assert.NoError(t, err) // Handler返回错误响应而非error
		assert.False(t, resp.Success)
	})
}

// TestOvertimeAdapter_ListEmployeeOvertimes 测试查询员工加班记录
func TestOvertimeAdapter_ListEmployeeOvertimes(t *testing.T) {
	t.Run("查询成功", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		tenantID := uuid.New()
		employeeID := uuid.New()
		year := 2024

		overtimes := []*model.Overtime{
			{ID: uuid.New(), TenantID: tenantID, EmployeeID: employeeID, Duration: 3.0},
		}

		mockService.On("ListByEmployee", mock.Anything, tenantID, employeeID, year).Return(overtimes, nil).Once()

		req := &hrmv1.ListEmployeeOvertimesRequest{
			TenantId:   tenantID.String(),
			EmployeeId: employeeID.String(),
			Year:       int32(year),
		}

		resp, err := adapter.ListEmployeeOvertimes(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Items, 1)
		mockService.AssertExpectations(t)
	})
}

// TestOvertimeAdapter_ListPendingOvertimes 测试查询待审批列表
func TestOvertimeAdapter_ListPendingOvertimes(t *testing.T) {
	t.Run("查询成功", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		tenantID := uuid.New()
		pendingOvertimes := []*model.Overtime{
			{ID: uuid.New(), TenantID: tenantID, ApprovalStatus: "pending"},
			{ID: uuid.New(), TenantID: tenantID, ApprovalStatus: "pending"},
		}

		mockService.On("ListPending", mock.Anything, tenantID).Return(pendingOvertimes, nil).Once()

		req := &hrmv1.ListPendingOvertimesRequest{
			TenantId: tenantID.String(),
		}

		resp, err := adapter.ListPendingOvertimes(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Items, 2)
		mockService.AssertExpectations(t)
	})
}

// TestOvertimeAdapter_SubmitOvertime 测试提交审批
func TestOvertimeAdapter_SubmitOvertime(t *testing.T) {
	t.Run("提交成功", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		overtimeID := uuid.New()
		submitterID := uuid.New()

		mockService.On("Submit", mock.Anything, overtimeID, submitterID).Return(nil).Once()

		req := &hrmv1.SubmitOvertimeRequest{
			OvertimeId:  overtimeID.String(),
			SubmitterId: submitterID.String(),
		}

		resp, err := adapter.SubmitOvertime(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})

	t.Run("提交失败_无效OvertimeID", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		req := &hrmv1.SubmitOvertimeRequest{
			OvertimeId:  "invalid",
			SubmitterId: uuid.New().String(),
		}

		resp, err := adapter.SubmitOvertime(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "invalid overtime_id")
	})
}

// TestOvertimeAdapter_SumOvertimeHours 测试统计加班时长
func TestOvertimeAdapter_SumOvertimeHours(t *testing.T) {
	t.Run("统计成功", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		tenantID := uuid.New()
		employeeID := uuid.New()
		startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
		endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.Local)

		mockService.On("SumHoursByEmployee", mock.Anything, tenantID, employeeID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(24.5, nil).Once()

		req := &hrmv1.SumOvertimeHoursRequest{
			TenantId:   tenantID.String(),
			EmployeeId: employeeID.String(),
			StartDate:  timestamppb.New(startDate),
			EndDate:    timestamppb.New(endDate),
		}

		resp, err := adapter.SumOvertimeHours(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, 24.5, resp.TotalHours)
		mockService.AssertExpectations(t)
	})
}

// TestOvertimeAdapter_GetCompOffDays 测试统计可调休天数
func TestOvertimeAdapter_GetCompOffDays(t *testing.T) {
	t.Run("统计成功", func(t *testing.T) {
		mockService := new(MockOvertimeService)
		overtimeHandler := handler.NewOvertimeHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, overtimeHandler, nil, nil, nil)

		tenantID := uuid.New()
		employeeID := uuid.New()

		mockService.On("SumCompOffDays", mock.Anything, tenantID, employeeID).Return(3.5, nil).Once()

		req := &hrmv1.GetCompOffDaysRequest{
			TenantId:   tenantID.String(),
			EmployeeId: employeeID.String(),
		}

		resp, err := adapter.GetCompOffDays(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, 3.5, resp.AvailableDays)
		mockService.AssertExpectations(t)
	})
}
