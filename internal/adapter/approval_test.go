package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	approvalv1 "github.com/lk2023060901/go-next-erp/api/approval/v1"
	"github.com/lk2023060901/go-next-erp/internal/approval/dto"
	"github.com/lk2023060901/go-next-erp/internal/approval/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockApprovalService mocks the approval service
type MockApprovalService struct {
	mock.Mock
}

// ProcessDefinition methods
func (m *MockApprovalService) CreateProcessDefinition(ctx context.Context, req *dto.CreateProcessDefRequest) (*dto.ProcessDefResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProcessDefResponse), args.Error(1)
}

func (m *MockApprovalService) UpdateProcessDefinition(ctx context.Context, id uuid.UUID, req *dto.UpdateProcessDefRequest) (*dto.ProcessDefResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProcessDefResponse), args.Error(1)
}

func (m *MockApprovalService) GetProcessDefinition(ctx context.Context, id uuid.UUID) (*dto.ProcessDefResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProcessDefResponse), args.Error(1)
}

func (m *MockApprovalService) ListProcessDefinitions(ctx context.Context, tenantID uuid.UUID) ([]*dto.ProcessDefResponse, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.ProcessDefResponse), args.Error(1)
}

func (m *MockApprovalService) DeleteProcessDefinition(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockApprovalService) SetProcessDefinitionStatus(ctx context.Context, id uuid.UUID, enabled bool) error {
	args := m.Called(ctx, id, enabled)
	return args.Error(0)
}

func (m *MockApprovalService) GetProcessStats(ctx context.Context, processDefID uuid.UUID) (*dto.ProcessStatsResponse, error) {
	args := m.Called(ctx, processDefID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProcessStatsResponse), args.Error(1)
}

// ProcessInstance methods
func (m *MockApprovalService) StartProcess(ctx context.Context, req *dto.StartProcessRequest) (*dto.ProcessInstanceResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProcessInstanceResponse), args.Error(1)
}

func (m *MockApprovalService) GetProcessInstance(ctx context.Context, id uuid.UUID) (*dto.ProcessInstanceResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProcessInstanceResponse), args.Error(1)
}

func (m *MockApprovalService) ListMyApplications(ctx context.Context, applicantID uuid.UUID, limit, offset int) ([]*dto.ProcessInstanceResponse, error) {
	args := m.Called(ctx, applicantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.ProcessInstanceResponse), args.Error(1)
}

func (m *MockApprovalService) WithdrawProcess(ctx context.Context, instanceID, operatorID uuid.UUID) error {
	args := m.Called(ctx, instanceID, operatorID)
	return args.Error(0)
}

func (m *MockApprovalService) CancelProcess(ctx context.Context, instanceID, operatorID uuid.UUID, reason *string) error {
	args := m.Called(ctx, instanceID, operatorID, reason)
	return args.Error(0)
}

func (m *MockApprovalService) ListProcessInstances(ctx context.Context, tenantID uuid.UUID, processDefID *uuid.UUID, status *model.ProcessStatus, applicantID *uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]*dto.ProcessInstanceResponse, int, error) {
	args := m.Called(ctx, tenantID, processDefID, status, applicantID, startDate, endDate, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*dto.ProcessInstanceResponse), args.Get(1).(int), args.Error(2)
}

func (m *MockApprovalService) GetInstanceStatsSummary(ctx context.Context, tenantID uuid.UUID) (*dto.InstanceStatsSummary, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.InstanceStatsSummary), args.Error(1)
}

// ApprovalTask methods
func (m *MockApprovalService) GetApprovalTask(ctx context.Context, taskID uuid.UUID) (*dto.ApprovalTaskResponse, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ApprovalTaskResponse), args.Error(1)
}

func (m *MockApprovalService) ListMyTasks(ctx context.Context, assigneeID uuid.UUID, status *model.TaskStatus, limit, offset int) ([]*dto.ApprovalTaskResponse, error) {
	args := m.Called(ctx, assigneeID, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.ApprovalTaskResponse), args.Error(1)
}

func (m *MockApprovalService) CountPendingTasks(ctx context.Context, assigneeID uuid.UUID) (int, error) {
	args := m.Called(ctx, assigneeID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockApprovalService) ProcessTask(ctx context.Context, req *dto.ProcessTaskRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockApprovalService) BatchProcessTasks(ctx context.Context, taskIDs []uuid.UUID, operatorID uuid.UUID, action model.ApprovalAction, comment *string) ([]*dto.BatchProcessResult, error) {
	args := m.Called(ctx, taskIDs, operatorID, action, comment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.BatchProcessResult), args.Error(1)
}

func (m *MockApprovalService) TransferTask(ctx context.Context, taskID, fromUserID, toUserID uuid.UUID, comment *string) error {
	args := m.Called(ctx, taskID, fromUserID, toUserID, comment)
	return args.Error(0)
}

func (m *MockApprovalService) DelegateTask(ctx context.Context, taskID, fromUserID, toUserID uuid.UUID, comment *string) error {
	args := m.Called(ctx, taskID, fromUserID, toUserID, comment)
	return args.Error(0)
}

func (m *MockApprovalService) GetInstanceStatsByStatus(ctx context.Context, tenantID uuid.UUID, processDefID *uuid.UUID, startDate, endDate *time.Time) (map[string]int, error) {
	args := m.Called(ctx, tenantID, processDefID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *MockApprovalService) ListPendingTasks(ctx context.Context, tenantID uuid.UUID, processDefID, assigneeID *uuid.UUID, limit, offset int) ([]*dto.ApprovalTaskResponse, int, error) {
	args := m.Called(ctx, tenantID, processDefID, assigneeID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*dto.ApprovalTaskResponse), args.Get(1).(int), args.Error(2)
}

func (m *MockApprovalService) ListCompletedTasks(ctx context.Context, tenantID uuid.UUID, processDefID, assigneeID *uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]*dto.ApprovalTaskResponse, int, error) {
	args := m.Called(ctx, tenantID, processDefID, assigneeID, startDate, endDate, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*dto.ApprovalTaskResponse), args.Get(1).(int), args.Error(2)
}

func (m *MockApprovalService) GetProcessHistory(ctx context.Context, instanceID uuid.UUID) ([]*dto.ProcessHistoryResponse, error) {
	args := m.Called(ctx, instanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.ProcessHistoryResponse), args.Error(1)
}

func (m *MockApprovalService) GetTaskHistory(ctx context.Context, taskID uuid.UUID) ([]*dto.ProcessHistoryResponse, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.ProcessHistoryResponse), args.Error(1)
}

func (m *MockApprovalService) GetProcessDiagram(ctx context.Context, processDefID uuid.UUID) (*dto.ProcessDiagramResponse, error) {
	args := m.Called(ctx, processDefID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProcessDiagramResponse), args.Error(1)
}

func (m *MockApprovalService) GetProcessInstanceDiagram(ctx context.Context, instanceID uuid.UUID) (*dto.ProcessInstanceDiagramResponse, error) {
	args := m.Called(ctx, instanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProcessInstanceDiagramResponse), args.Error(1)
}

func (m *MockApprovalService) GetProcessTrace(ctx context.Context, instanceID uuid.UUID) (*dto.ProcessTraceResponse, error) {
	args := m.Called(ctx, instanceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProcessTraceResponse), args.Error(1)
}

func (m *MockApprovalService) GetDashboard(ctx context.Context, tenantID, userID uuid.UUID) (*dto.DashboardResponse, error) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DashboardResponse), args.Error(1)
}

func (m *MockApprovalService) GetProcessMetrics(ctx context.Context, processDefID uuid.UUID, startDate, endDate *time.Time) (*dto.ProcessMetrics, error) {
	args := m.Called(ctx, processDefID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProcessMetrics), args.Error(1)
}

func (m *MockApprovalService) GetUserWorkload(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time) (*dto.UserWorkload, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserWorkload), args.Error(1)
}

// TestApprovalAdapter_CreateProcessDefinition tests creating process definitions
func TestApprovalAdapter_CreateProcessDefinition(t *testing.T) {
	t.Run("CreateProcessDefinition successfully", func(t *testing.T) {
		mockService := new(MockApprovalService)
		adapter := NewApprovalAdapter(mockService)

		processDefID := uuid.New()
		tenantID := uuid.New()
		formID := uuid.New()
		workflowID := uuid.New()

		expectedProcessDef := &dto.ProcessDefResponse{
			ID:           processDefID,
			TenantID:     tenantID,
			Code:         "LEAVE_001",
			Name:         "请假审批",
			FormID:       formID,
			FormName:     "请假单",
			WorkflowID:   workflowID,
			WorkflowName: "请假流程",
			Enabled:      true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mockService.On("CreateProcessDefinition", mock.Anything, mock.AnythingOfType("*dto.CreateProcessDefRequest")).
			Return(expectedProcessDef, nil).Once()

		req := &approvalv1.CreateProcessDefinitionRequest{
			Code:       "LEAVE_001",
			Name:       "请假审批",
			Category:   "leave",
			FormId:     formID.String(),
			WorkflowId: workflowID.String(),
		}

		resp, err := adapter.CreateProcessDefinition(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, processDefID.String(), resp.Id)
		assert.Equal(t, "请假审批", resp.Name)
		mockService.AssertExpectations(t)
	})
}

// TestApprovalAdapter_GetProcessDefinition tests getting a process definition
func TestApprovalAdapter_GetProcessDefinition(t *testing.T) {
	t.Run("GetProcessDefinition successfully", func(t *testing.T) {
		mockService := new(MockApprovalService)
		adapter := NewApprovalAdapter(mockService)

		processDefID := uuid.New()

		expectedProcessDef := &dto.ProcessDefResponse{
			ID:        processDefID,
			TenantID:  uuid.New(),
			Code:      "LEAVE_001",
			Name:      "请假审批",
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("GetProcessDefinition", mock.Anything, processDefID).
			Return(expectedProcessDef, nil).Once()

		req := &approvalv1.GetProcessDefinitionRequest{
			Id: processDefID.String(),
		}

		resp, err := adapter.GetProcessDefinition(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, processDefID.String(), resp.Id)
		mockService.AssertExpectations(t)
	})
}

// TestApprovalAdapter_UpdateProcessDefinition tests updating process definitions
func TestApprovalAdapter_UpdateProcessDefinition(t *testing.T) {
	t.Run("UpdateProcessDefinition successfully", func(t *testing.T) {
		mockService := new(MockApprovalService)
		adapter := NewApprovalAdapter(mockService)

		processDefID := uuid.New()
		formID := uuid.New()
		workflowID := uuid.New()

		updatedProcessDef := &dto.ProcessDefResponse{
			ID:         processDefID,
			Name:       "请假审批-更新",
			FormID:     formID,
			WorkflowID: workflowID,
			Enabled:    false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockService.On("UpdateProcessDefinition", mock.Anything, processDefID, mock.AnythingOfType("*dto.UpdateProcessDefRequest")).
			Return(updatedProcessDef, nil).Once()

		req := &approvalv1.UpdateProcessDefinitionRequest{
			Id:         processDefID.String(),
			Name:       "请假审批-更新",
			FormId:     formID.String(),
			WorkflowId: workflowID.String(),
			Enabled:    false,
		}

		resp, err := adapter.UpdateProcessDefinition(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "请假审批-更新", resp.Name)
		mockService.AssertExpectations(t)
	})
}

// TestApprovalAdapter_DeleteProcessDefinition tests deleting process definitions
func TestApprovalAdapter_DeleteProcessDefinition(t *testing.T) {
	t.Run("DeleteProcessDefinition successfully", func(t *testing.T) {
		mockService := new(MockApprovalService)
		adapter := NewApprovalAdapter(mockService)

		processDefID := uuid.New()

		mockService.On("DeleteProcessDefinition", mock.Anything, processDefID).
			Return(nil).Once()

		req := &approvalv1.DeleteProcessDefinitionRequest{
			Id: processDefID.String(),
		}

		resp, err := adapter.DeleteProcessDefinition(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockService.AssertExpectations(t)
	})
}

// TestApprovalAdapter_StartProcess tests starting a process
func TestApprovalAdapter_StartProcess(t *testing.T) {
	t.Run("StartProcess successfully", func(t *testing.T) {
		mockService := new(MockApprovalService)
		adapter := NewApprovalAdapter(mockService)

		instanceID := uuid.New()
		processDefID := uuid.New()

		expectedInstance := &dto.ProcessInstanceResponse{
			ID:             instanceID,
			ProcessDefID:   processDefID,
			ProcessDefCode: "LEAVE_001",
			ProcessDefName: "请假审批",
			Status:         model.ProcessStatusPending,
			StartedAt:      time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		mockService.On("StartProcess", mock.Anything, mock.AnythingOfType("*dto.StartProcessRequest")).
			Return(expectedInstance, nil).Once()

		req := &approvalv1.StartProcessRequest{
			ProcessDefId: processDefID.String(),
			FormData: map[string]string{
				"days":   "3",
				"reason": "Personal leave",
			},
		}

		resp, err := adapter.StartProcess(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, instanceID.String(), resp.Id)
		mockService.AssertExpectations(t)
	})
}

// TestApprovalAdapter_ProcessTask tests processing an approval task
func TestApprovalAdapter_ProcessTask(t *testing.T) {
	t.Run("ProcessTask approve successfully", func(t *testing.T) {
		mockService := new(MockApprovalService)
		adapter := NewApprovalAdapter(mockService)

		taskID := uuid.New()

		mockService.On("ProcessTask", mock.Anything, mock.AnythingOfType("*dto.ProcessTaskRequest")).
			Return(nil).Once()

		req := &approvalv1.ProcessTaskRequest{
			Id:      taskID.String(),
			Action:  "approve",
			Comment: "Approved",
		}

		resp, err := adapter.ProcessTask(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockService.AssertExpectations(t)
	})

	t.Run("ProcessTask reject successfully", func(t *testing.T) {
		mockService := new(MockApprovalService)
		adapter := NewApprovalAdapter(mockService)

		taskID := uuid.New()

		mockService.On("ProcessTask", mock.Anything, mock.AnythingOfType("*dto.ProcessTaskRequest")).
			Return(nil).Once()

		req := &approvalv1.ProcessTaskRequest{
			Id:      taskID.String(),
			Action:  "reject",
			Comment: "Rejected due to insufficient information",
		}

		resp, err := adapter.ProcessTask(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockService.AssertExpectations(t)
	})
}

// TestApprovalAdapter_ListMyTasks tests listing user's tasks
func TestApprovalAdapter_ListMyTasks(t *testing.T) {
	t.Run("ListMyTasks successfully", func(t *testing.T) {
		mockService := new(MockApprovalService)
		adapter := NewApprovalAdapter(mockService)

		expectedTasks := []*dto.ApprovalTaskResponse{
			{
				ID:                uuid.New(),
				ProcessInstanceID: uuid.New(),
				NodeID:            "node_1",
				NodeName:          "Manager Approval",
				Status:            model.TaskStatusPending,
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			},
		}

		mockService.On("ListMyTasks", mock.Anything, mock.AnythingOfType("uuid.UUID"), (*model.TaskStatus)(nil), 20, 0).
			Return(expectedTasks, nil).Once()

		req := &approvalv1.ListMyTasksRequest{
			Limit:  20,
			Offset: 0,
		}

		resp, err := adapter.ListMyTasks(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Items, 1)
		mockService.AssertExpectations(t)
	})
}

// TestApprovalAdapter_CountPendingTasks tests counting pending tasks
func TestApprovalAdapter_CountPendingTasks(t *testing.T) {
	t.Run("CountPendingTasks successfully", func(t *testing.T) {
		mockService := new(MockApprovalService)
		adapter := NewApprovalAdapter(mockService)

		mockService.On("CountPendingTasks", mock.Anything, mock.AnythingOfType("uuid.UUID")).
			Return(5, nil).Once()

		req := &approvalv1.CountPendingTasksRequest{}

		resp, err := adapter.CountPendingTasks(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(5), resp.Count)
		mockService.AssertExpectations(t)
	})
}

// TestApprovalAdapter_BatchProcessTasks tests batch processing tasks
func TestApprovalAdapter_BatchProcessTasks(t *testing.T) {
	t.Run("BatchProcessTasks successfully", func(t *testing.T) {
		mockService := new(MockApprovalService)
		adapter := NewApprovalAdapter(mockService)

		taskID1 := uuid.New()
		taskID2 := uuid.New()

		expectedResults := []*dto.BatchProcessResult{
			{TaskID: taskID1, Success: true},
			{TaskID: taskID2, Success: true},
		}

		mockService.On("BatchProcessTasks", mock.Anything, mock.AnythingOfType("[]uuid.UUID"), mock.AnythingOfType("uuid.UUID"), model.ApprovalAction("approve"), (*string)(nil)).
			Return(expectedResults, nil).Once()

		req := &approvalv1.BatchProcessTasksRequest{
			TaskIds: []string{taskID1.String(), taskID2.String()},
			Action:  "approve",
		}

		resp, err := adapter.BatchProcessTasks(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Results, 2)
		mockService.AssertExpectations(t)
	})
}

// TestApprovalAdapter_TransferTask tests transferring a task
func TestApprovalAdapter_TransferTask(t *testing.T) {
	t.Run("TransferTask successfully", func(t *testing.T) {
		mockService := new(MockApprovalService)
		adapter := NewApprovalAdapter(mockService)

		taskID := uuid.New()
		toUserID := uuid.New()

		mockService.On("TransferTask", mock.Anything, taskID, mock.AnythingOfType("uuid.UUID"), toUserID, (*string)(nil)).
			Return(nil).Once()

		req := &approvalv1.TransferTaskRequest{
			Id:           taskID.String(),
			TransferToId: toUserID.String(),
		}

		resp, err := adapter.TransferTask(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockService.AssertExpectations(t)
	})
}

// TestApprovalAdapter_DelegateTask tests delegating a task
func TestApprovalAdapter_DelegateTask(t *testing.T) {
	t.Run("DelegateTask successfully", func(t *testing.T) {
		mockService := new(MockApprovalService)
		adapter := NewApprovalAdapter(mockService)

		taskID := uuid.New()
		toUserID := uuid.New()

		mockService.On("DelegateTask", mock.Anything, taskID, mock.AnythingOfType("uuid.UUID"), toUserID, (*string)(nil)).
			Return(nil).Once()

		req := &approvalv1.DelegateTaskRequest{
			Id:           taskID.String(),
			DelegateToId: toUserID.String(),
		}

		resp, err := adapter.DelegateTask(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.IsType(t, &emptypb.Empty{}, resp)
		mockService.AssertExpectations(t)
	})
}
