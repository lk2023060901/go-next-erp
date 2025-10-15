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

// MockPunchCardSupplementService mocks the PunchCardSupplementService interface
type MockPunchCardSupplementService struct {
	mock.Mock
}

func (m *MockPunchCardSupplementService) Create(ctx context.Context, supplement *model.PunchCardSupplement) error {
	args := m.Called(ctx, supplement)
	return args.Error(0)
}

func (m *MockPunchCardSupplementService) Update(ctx context.Context, supplement *model.PunchCardSupplement) error {
	args := m.Called(ctx, supplement)
	return args.Error(0)
}

func (m *MockPunchCardSupplementService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPunchCardSupplementService) GetByID(ctx context.Context, id uuid.UUID) (*model.PunchCardSupplement, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PunchCardSupplement), args.Error(1)
}

func (m *MockPunchCardSupplementService) List(ctx context.Context, tenantID uuid.UUID, filter *repository.PunchCardSupplementFilter, offset, limit int) ([]*model.PunchCardSupplement, int, error) {
	args := m.Called(ctx, tenantID, filter, offset, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.PunchCardSupplement), args.Int(1), args.Error(2)
}

func (m *MockPunchCardSupplementService) ListByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.PunchCardSupplement, error) {
	args := m.Called(ctx, tenantID, employeeID, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.PunchCardSupplement), args.Error(1)
}

func (m *MockPunchCardSupplementService) ListPending(ctx context.Context, tenantID uuid.UUID) ([]*model.PunchCardSupplement, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.PunchCardSupplement), args.Error(1)
}

func (m *MockPunchCardSupplementService) Submit(ctx context.Context, supplementID, submitterID uuid.UUID) error {
	args := m.Called(ctx, supplementID, submitterID)
	return args.Error(0)
}

func (m *MockPunchCardSupplementService) Approve(ctx context.Context, supplementID, approverID uuid.UUID, comment string) error {
	args := m.Called(ctx, supplementID, approverID, comment)
	return args.Error(0)
}

func (m *MockPunchCardSupplementService) Reject(ctx context.Context, supplementID, approverID uuid.UUID, reason string) error {
	args := m.Called(ctx, supplementID, approverID, reason)
	return args.Error(0)
}

func (m *MockPunchCardSupplementService) Process(ctx context.Context, supplementID, processorID uuid.UUID) error {
	args := m.Called(ctx, supplementID, processorID)
	return args.Error(0)
}

func (m *MockPunchCardSupplementService) Cancel(ctx context.Context, supplementID uuid.UUID) error {
	args := m.Called(ctx, supplementID)
	return args.Error(0)
}

func (m *MockPunchCardSupplementService) ValidateSupplement(ctx context.Context, supplement *model.PunchCardSupplement) error {
	args := m.Called(ctx, supplement)
	return args.Error(0)
}

func (m *MockPunchCardSupplementService) CheckDuplicate(ctx context.Context, tenantID, employeeID uuid.UUID, date time.Time, supplementType model.SupplementType) (bool, error) {
	args := m.Called(ctx, tenantID, employeeID, date, supplementType)
	return args.Bool(0), args.Error(1)
}

func (m *MockPunchCardSupplementService) CountByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) (int, error) {
	args := m.Called(ctx, tenantID, employeeID, startDate, endDate)
	return args.Int(0), args.Error(1)
}

// TestPunchCardSupplementAdapter_CreatePunchCardSupplement 测试创建补卡申请
func TestPunchCardSupplementAdapter_CreatePunchCardSupplement(t *testing.T) {
	t.Run("创建成功", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		tenantID := uuid.New()
		employeeID := uuid.New()
		departmentID := uuid.New()
		now := time.Now()

		mockService.On("Create", mock.Anything, mock.MatchedBy(func(s *model.PunchCardSupplement) bool {
			return s.TenantID == tenantID &&
				s.EmployeeID == employeeID &&
				s.SupplementType == model.SupplementTypeCheckIn &&
				s.MissingType == model.MissingTypeForgot
		})).Return(nil).Once()

		req := &hrmv1.CreatePunchCardSupplementRequest{
			TenantId:       tenantID.String(),
			EmployeeId:     employeeID.String(),
			EmployeeName:   "张三",
			DepartmentId:   departmentID.String(),
			SupplementDate: timestamppb.New(now),
			SupplementType: "checkin",
			SupplementTime: timestamppb.New(now.Add(time.Hour)),
			MissingType:    "forgot",
			Reason:         "忘记打卡",
			Evidence: []*hrmv1.SupplementEvidence{
				{
					Type:        "image",
					Url:         "https://example.com/proof.jpg",
					Description: "工作照片证明",
				},
			},
		}

		resp, err := adapter.CreatePunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "张三", resp.EmployeeName)
		mockService.AssertExpectations(t)
	})

	t.Run("创建失败_无效TenantID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.CreatePunchCardSupplementRequest{
			TenantId:   "invalid-uuid",
			EmployeeId: uuid.New().String(),
		}

		resp, err := adapter.CreatePunchCardSupplement(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid tenant_id")
	})
}

// TestPunchCardSupplementAdapter_UpdatePunchCardSupplement 测试更新补卡申请
func TestPunchCardSupplementAdapter_UpdatePunchCardSupplement(t *testing.T) {
	t.Run("更新成功", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		supplementID := uuid.New()
		existingSupplement := &model.PunchCardSupplement{
			ID:             supplementID,
			TenantID:       uuid.New(),
			EmployeeID:     uuid.New(),
			SupplementType: model.SupplementTypeCheckIn,
			MissingType:    model.MissingTypeForgot,
			Reason:         "原始原因",
			ApprovalStatus: "pending",
		}

		mockService.On("GetByID", mock.Anything, supplementID).Return(existingSupplement, nil).Once()
		mockService.On("Update", mock.Anything, mock.MatchedBy(func(s *model.PunchCardSupplement) bool {
			return s.ID == supplementID && s.Reason == "更新后的原因"
		})).Return(nil).Once()

		req := &hrmv1.UpdatePunchCardSupplementRequest{
			Id:     supplementID.String(),
			Reason: "更新后的原因",
		}

		resp, err := adapter.UpdatePunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockService.AssertExpectations(t)
	})

	t.Run("更新失败_无效ID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.UpdatePunchCardSupplementRequest{
			Id: "invalid-uuid",
		}

		resp, err := adapter.UpdatePunchCardSupplement(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// TestPunchCardSupplementAdapter_DeletePunchCardSupplement 测试删除补卡申请
func TestPunchCardSupplementAdapter_DeletePunchCardSupplement(t *testing.T) {
	t.Run("删除成功", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		supplementID := uuid.New()

		mockService.On("Delete", mock.Anything, supplementID).Return(nil).Once()

		req := &hrmv1.DeletePunchCardSupplementRequest{
			Id: supplementID.String(),
		}

		resp, err := adapter.DeletePunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})

	t.Run("删除失败_服务错误", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		supplementID := uuid.New()

		mockService.On("Delete", mock.Anything, supplementID).Return(assert.AnError).Once()

		req := &hrmv1.DeletePunchCardSupplementRequest{
			Id: supplementID.String(),
		}

		resp, err := adapter.DeletePunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, resp.Success)
		mockService.AssertExpectations(t)
	})

	t.Run("删除失败_无效ID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.DeletePunchCardSupplementRequest{
			Id: "invalid-uuid",
		}

		resp, err := adapter.DeletePunchCardSupplement(context.Background(), req)

		assert.NoError(t, err) // No error from handler, but success will be false
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "invalid id")
	})
}

// TestPunchCardSupplementAdapter_GetPunchCardSupplement 测试获取详情
func TestPunchCardSupplementAdapter_GetPunchCardSupplement(t *testing.T) {
	t.Run("获取成功", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		supplementID := uuid.New()
		supplement := &model.PunchCardSupplement{
			ID:             supplementID,
			TenantID:       uuid.New(),
			EmployeeID:     uuid.New(),
			EmployeeName:   "张三",
			SupplementType: model.SupplementTypeCheckIn,
			MissingType:    model.MissingTypeForgot,
			Reason:         "忘记打卡",
			ApprovalStatus: "pending",
		}

		mockService.On("GetByID", mock.Anything, supplementID).Return(supplement, nil).Once()

		req := &hrmv1.GetPunchCardSupplementRequest{
			Id: supplementID.String(),
		}

		resp, err := adapter.GetPunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, supplementID.String(), resp.Id)
		assert.Equal(t, "张三", resp.EmployeeName)
		mockService.AssertExpectations(t)
	})

	t.Run("获取失败_无效ID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.GetPunchCardSupplementRequest{
			Id: "invalid-uuid",
		}

		resp, err := adapter.GetPunchCardSupplement(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid id")
	})
}

// TestPunchCardSupplementAdapter_ListPunchCardSupplements 测试列表查询
func TestPunchCardSupplementAdapter_ListPunchCardSupplements(t *testing.T) {
	t.Run("列表查询成功", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		tenantID := uuid.New()
		supplements := []*model.PunchCardSupplement{
			{
				ID:             uuid.New(),
				TenantID:       tenantID,
				EmployeeID:     uuid.New(),
				EmployeeName:   "张三",
				SupplementType: model.SupplementTypeCheckIn,
				MissingType:    model.MissingTypeForgot,
				ApprovalStatus: "pending",
			},
		}

		mockService.On("List", mock.Anything, tenantID, mock.Anything, 0, 20).
			Return(supplements, 1, nil).Once()

		req := &hrmv1.ListPunchCardSupplementsRequest{
			TenantId: tenantID.String(),
			Page:     1,
			PageSize: 20,
		}

		resp, err := adapter.ListPunchCardSupplements(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Items, 1)
		assert.Equal(t, int64(1), resp.Total)
		mockService.AssertExpectations(t)
	})

	t.Run("无效TenantID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.ListPunchCardSupplementsRequest{
			TenantId: "invalid-uuid",
		}

		resp, err := adapter.ListPunchCardSupplements(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid tenant_id")
	})

	t.Run("无效EmployeeID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.ListPunchCardSupplementsRequest{
			TenantId:   uuid.New().String(),
			EmployeeId: "invalid-uuid",
		}

		resp, err := adapter.ListPunchCardSupplements(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid employee_id")
	})

	t.Run("无效DepartmentID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.ListPunchCardSupplementsRequest{
			TenantId:     uuid.New().String(),
			DepartmentId: "invalid-uuid",
		}

		resp, err := adapter.ListPunchCardSupplements(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid department_id")
	})

	t.Run("无效StartDate", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.ListPunchCardSupplementsRequest{
			TenantId:  uuid.New().String(),
			StartDate: nil, // This should be handled gracefully
		}

		// We'll just check that it doesn't panic or error with nil timestamp
		mockService.On("List", mock.Anything, mock.Anything, mock.Anything, 0, 20).
			Return([]*model.PunchCardSupplement{}, 0, nil).Once()

		resp, err := adapter.ListPunchCardSupplements(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockService.AssertExpectations(t)
	})
}

// TestPunchCardSupplementAdapter_ApprovePunchCardSupplement 测试批准申请
func TestPunchCardSupplementAdapter_ApprovePunchCardSupplement(t *testing.T) {
	t.Run("批准成功", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		supplementID := uuid.New()
		approverID := uuid.New()

		mockService.On("Approve", mock.Anything, supplementID, approverID, "同意补卡").Return(nil).Once()

		req := &hrmv1.ApprovePunchCardSupplementRequest{
			SupplementId: supplementID.String(),
			ApproverId:   approverID.String(),
			Comment:      "同意补卡",
		}

		resp, err := adapter.ApprovePunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})
}

// TestPunchCardSupplementAdapter_RejectPunchCardSupplement 测试拒绝申请
func TestPunchCardSupplementAdapter_RejectPunchCardSupplement(t *testing.T) {
	t.Run("拒绝成功", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		supplementID := uuid.New()
		approverID := uuid.New()

		mockService.On("Reject", mock.Anything, supplementID, approverID, "不符合条件").Return(nil).Once()

		req := &hrmv1.RejectPunchCardSupplementRequest{
			SupplementId: supplementID.String(),
			ApproverId:   approverID.String(),
			Reason:       "不符合条件",
		}

		resp, err := adapter.RejectPunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})
}

// TestPunchCardSupplementAdapter_ListPendingPunchCardSupplements 测试查询待审批列表
func TestPunchCardSupplementAdapter_ListPendingPunchCardSupplements(t *testing.T) {
	t.Run("查询成功", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		tenantID := uuid.New()
		pendingSupplements := []*model.PunchCardSupplement{
			{ID: uuid.New(), TenantID: tenantID, ApprovalStatus: "pending"},
			{ID: uuid.New(), TenantID: tenantID, ApprovalStatus: "pending"},
		}

		mockService.On("ListPending", mock.Anything, tenantID).Return(pendingSupplements, nil).Once()

		req := &hrmv1.ListPendingPunchCardSupplementsRequest{
			TenantId: tenantID.String(),
			Page:     1,
			PageSize: 20,
		}

		resp, err := adapter.ListPendingPunchCardSupplements(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Items, 2)
		mockService.AssertExpectations(t)
	})
}

// TestPunchCardSupplementAdapter_ListEmployeePunchCardSupplements 测试查询员工补卡申请记录
func TestPunchCardSupplementAdapter_ListEmployeePunchCardSupplements(t *testing.T) {
	t.Run("查询成功", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		tenantID := uuid.New()
		employeeID := uuid.New()
		supplements := []*model.PunchCardSupplement{
			{
				ID:             uuid.New(),
				TenantID:       tenantID,
				EmployeeID:     employeeID,
				EmployeeName:   "张三",
				SupplementType: model.SupplementTypeCheckIn,
				MissingType:    model.MissingTypeForgot,
				ApprovalStatus: "pending",
			},
		}

		mockService.On("ListByEmployee", mock.Anything, tenantID, employeeID, 2023).
			Return(supplements, nil).Once()

		req := &hrmv1.ListEmployeePunchCardSupplementsRequest{
			TenantId:   tenantID.String(),
			EmployeeId: employeeID.String(),
			Year:       2023,
		}

		resp, err := adapter.ListEmployeePunchCardSupplements(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, resp.Items, 1)
		assert.Equal(t, int64(1), resp.Total)
		mockService.AssertExpectations(t)
	})

	t.Run("无效TenantID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.ListEmployeePunchCardSupplementsRequest{
			TenantId:   "invalid-uuid",
			EmployeeId: uuid.New().String(),
		}

		resp, err := adapter.ListEmployeePunchCardSupplements(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid tenant_id")
	})

	t.Run("无效EmployeeID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.ListEmployeePunchCardSupplementsRequest{
			TenantId:   uuid.New().String(),
			EmployeeId: "invalid-uuid",
		}

		resp, err := adapter.ListEmployeePunchCardSupplements(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid employee_id")
	})
}

// TestPunchCardSupplementAdapter_SubmitPunchCardSupplement 测试提交补卡申请
func TestPunchCardSupplementAdapter_SubmitPunchCardSupplement(t *testing.T) {
	t.Run("提交成功", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		supplementID := uuid.New()
		submitterID := uuid.New()

		mockService.On("Submit", mock.Anything, supplementID, submitterID).Return(nil).Once()

		req := &hrmv1.SubmitPunchCardSupplementRequest{
			SupplementId: supplementID.String(),
			SubmitterId:  submitterID.String(),
		}

		resp, err := adapter.SubmitPunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})

	t.Run("无效SupplementID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.SubmitPunchCardSupplementRequest{
			SupplementId: "invalid-uuid",
			SubmitterId:  uuid.New().String(),
		}

		resp, err := adapter.SubmitPunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "invalid supplement_id")
	})

	t.Run("无效SubmitterID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.SubmitPunchCardSupplementRequest{
			SupplementId: uuid.New().String(),
			SubmitterId:  "invalid-uuid",
		}

		resp, err := adapter.SubmitPunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "invalid submitter_id")
	})
}

// TestPunchCardSupplementAdapter_ProcessPunchCardSupplement 测试处理补卡
func TestPunchCardSupplementAdapter_ProcessPunchCardSupplement(t *testing.T) {
	t.Run("处理成功", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		supplementID := uuid.New()
		processorID := uuid.New()

		mockService.On("Process", mock.Anything, supplementID, processorID).Return(nil).Once()

		req := &hrmv1.ProcessPunchCardSupplementRequest{
			SupplementId: supplementID.String(),
			ProcessorId:  processorID.String(),
		}

		resp, err := adapter.ProcessPunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})

	t.Run("无效SupplementID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.ProcessPunchCardSupplementRequest{
			SupplementId: "invalid-uuid",
			ProcessorId:  uuid.New().String(),
		}

		resp, err := adapter.ProcessPunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "invalid supplement_id")
	})

	t.Run("无效ProcessorID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.ProcessPunchCardSupplementRequest{
			SupplementId: uuid.New().String(),
			ProcessorId:  "invalid-uuid",
		}

		resp, err := adapter.ProcessPunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "invalid processor_id")
	})
}

// TestPunchCardSupplementAdapter_CancelPunchCardSupplement 测试取消补卡申请
func TestPunchCardSupplementAdapter_CancelPunchCardSupplement(t *testing.T) {
	t.Run("取消成功", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		supplementID := uuid.New()

		mockService.On("Cancel", mock.Anything, supplementID).Return(nil).Once()

		req := &hrmv1.CancelPunchCardSupplementRequest{
			SupplementId: supplementID.String(),
		}

		resp, err := adapter.CancelPunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		mockService.AssertExpectations(t)
	})

	t.Run("无效SupplementID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.CancelPunchCardSupplementRequest{
			SupplementId: "invalid-uuid",
		}

		resp, err := adapter.CancelPunchCardSupplement(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "invalid supplement_id")
	})
}

// TestPunchCardSupplementAdapter_Security_InvalidUUID 测试UUID验证
func TestPunchCardSupplementAdapter_Security_InvalidUUID(t *testing.T) {
	t.Run("CreatePunchCardSupplement_无效DepartmentID", func(t *testing.T) {
		mockService := new(MockPunchCardSupplementService)
		supplementHandler := handler.NewPunchCardSupplementHandler(mockService)
		adapter := NewHRMAdapter(nil, nil, nil, nil, nil, nil, nil, nil, supplementHandler)

		req := &hrmv1.CreatePunchCardSupplementRequest{
			TenantId:     uuid.New().String(),
			EmployeeId:   uuid.New().String(),
			DepartmentId: "invalid-uuid",
		}

		resp, err := adapter.CreatePunchCardSupplement(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid department_id")
	})
}
