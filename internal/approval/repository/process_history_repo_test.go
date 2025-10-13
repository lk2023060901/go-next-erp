package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/approval/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create test process history
func createTestProcessHistory(t *testing.T, db *database.DB, tenantID uuid.UUID) *model.ProcessHistory {
	t.Helper()
	ctx := context.Background()

	// Create process instance first and insert into DB
	processInstance := createTestProcessInstance(t, db, tenantID)

	// Insert the process instance into database
	processInstanceRepo := NewProcessInstanceRepository(db)
	err := processInstanceRepo.Create(ctx, processInstance)
	if err != nil {
		t.Logf("Warning: failed to create process instance: %v", err)
	}

	fromStatus := model.ProcessStatusPending
	history := &model.ProcessHistory{
		ID:                uuid.New(),
		TenantID:          tenantID,
		ProcessInstanceID: processInstance.ID,
		TaskID:            nil, // Optional
		NodeID:            "node_" + uuid.NewString()[:8],
		NodeName:          "测试节点",
		OperatorID:        uuid.New(),
		OperatorName:      "测试操作员",
		Action:            model.ApprovalActionApprove,
		Comment:           nil,
		FromStatus:        &fromStatus,
		ToStatus:          model.ProcessStatusApproved,
		CreatedAt:         time.Now(),
	}

	return history
}

// Cleanup helper
func cleanupProcessHistories(t *testing.T, db *database.DB, tenantID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	_, _ = db.Exec(ctx, "DELETE FROM approval_process_histories WHERE tenant_id = $1", tenantID)
	_, _ = db.Exec(ctx, "DELETE FROM approval_tasks WHERE tenant_id = $1", tenantID)
	_, _ = db.Exec(ctx, "DELETE FROM approval_process_instances WHERE tenant_id = $1", tenantID)
	_, _ = db.Exec(ctx, "DELETE FROM approval_process_definitions WHERE tenant_id = $1", tenantID)
	// workflows表不存在，无需删除
	_, _ = db.Exec(ctx, "DELETE FROM form_data WHERE tenant_id = $1", tenantID)
	_, _ = db.Exec(ctx, "DELETE FROM form_definitions WHERE tenant_id = $1", tenantID)
}

// TestProcessHistoryRepository_Create tests history creation
func TestProcessHistoryRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessHistoryRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessHistories(t, db, tenantID)

	t.Run("Create successfully", func(t *testing.T) {
		history := createTestProcessHistory(t, db, tenantID)

		err := repo.Create(ctx, history)
		assert.NoError(t, err)

		// Verify by listing
		histories, err := repo.ListByInstance(ctx, history.ProcessInstanceID)
		require.NoError(t, err)
		assert.Len(t, histories, 1)
		assert.Equal(t, history.ID, histories[0].ID)
		assert.Equal(t, history.Action, histories[0].Action)
	})

	t.Run("Create with approve action", func(t *testing.T) {
		history := createTestProcessHistory(t, db, tenantID)
		history.Action = model.ApprovalActionApprove
		comment := "同意"
		history.Comment = &comment

		err := repo.Create(ctx, history)
		assert.NoError(t, err)

		histories, err := repo.ListByInstance(ctx, history.ProcessInstanceID)
		require.NoError(t, err)
		assert.NotEmpty(t, histories)
		found := histories[len(histories)-1] // Get the last one
		assert.Equal(t, model.ApprovalActionApprove, found.Action)
		assert.NotNil(t, found.Comment)
		assert.Equal(t, "同意", *found.Comment)
	})

	t.Run("Create with reject action", func(t *testing.T) {
		history := createTestProcessHistory(t, db, tenantID)
		history.Action = model.ApprovalActionReject
		comment := "拒绝"
		history.Comment = &comment

		err := repo.Create(ctx, history)
		assert.NoError(t, err)

		histories, err := repo.ListByInstance(ctx, history.ProcessInstanceID)
		require.NoError(t, err)
		assert.NotEmpty(t, histories)
		found := histories[len(histories)-1]
		assert.Equal(t, model.ApprovalActionReject, found.Action)
	})

	t.Run("Create with transfer action", func(t *testing.T) {
		history := createTestProcessHistory(t, db, tenantID)
		history.Action = model.ApprovalActionTransfer
		comment := "转交给其他人"
		history.Comment = &comment

		err := repo.Create(ctx, history)
		assert.NoError(t, err)

		histories, err := repo.ListByInstance(ctx, history.ProcessInstanceID)
		require.NoError(t, err)
		assert.NotEmpty(t, histories)
		found := histories[len(histories)-1]
		assert.Equal(t, model.ApprovalActionTransfer, found.Action)
	})

	t.Run("Create with task ID", func(t *testing.T) {
		history := createTestProcessHistory(t, db, tenantID)
		taskID := uuid.New()
		history.TaskID = &taskID

		err := repo.Create(ctx, history)
		assert.NoError(t, err)

		histories, err := repo.ListByTaskID(ctx, taskID)
		require.NoError(t, err)
		assert.Len(t, histories, 1)
		assert.NotNil(t, histories[0].TaskID)
		assert.Equal(t, taskID, *histories[0].TaskID)
	})

	t.Run("Create without task ID", func(t *testing.T) {
		history := createTestProcessHistory(t, db, tenantID)
		history.TaskID = nil

		err := repo.Create(ctx, history)
		assert.NoError(t, err)

		histories, err := repo.ListByInstance(ctx, history.ProcessInstanceID)
		require.NoError(t, err)
		assert.NotEmpty(t, histories)
		found := histories[len(histories)-1]
		assert.Nil(t, found.TaskID)
	})

	t.Run("Create with nil comment", func(t *testing.T) {
		history := createTestProcessHistory(t, db, tenantID)
		history.Comment = nil

		err := repo.Create(ctx, history)
		assert.NoError(t, err)

		histories, err := repo.ListByInstance(ctx, history.ProcessInstanceID)
		require.NoError(t, err)
		assert.NotEmpty(t, histories)
		found := histories[len(histories)-1]
		assert.Nil(t, found.Comment)
	})
}

// TestProcessHistoryRepository_ListByInstance tests listing by process instance
func TestProcessHistoryRepository_ListByInstance(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessHistoryRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessHistories(t, db, tenantID)

	t.Run("List histories for instance", func(t *testing.T) {
		// Create a process instance and insert it
		processInstance := createTestProcessInstance(t, db, tenantID)
		processInstanceRepo := NewProcessInstanceRepository(db)
		err := processInstanceRepo.Create(ctx, processInstance)
		require.NoError(t, err)

		fromStatus := model.ProcessStatusPending
		// Create multiple history records
		history1 := &model.ProcessHistory{
			ID:                uuid.New(),
			TenantID:          tenantID,
			ProcessInstanceID: processInstance.ID,
			NodeID:            "node_1",
			NodeName:          "测试节点1",
			OperatorID:        uuid.New(),
			OperatorName:      "测试操作员1",
			Action:            model.ApprovalActionApprove,
			FromStatus:        &fromStatus,
			ToStatus:          model.ProcessStatusApproved,
			CreatedAt:         time.Now().Add(-2 * time.Hour),
		}
		err = repo.Create(ctx, history1)
		require.NoError(t, err)

		history2 := &model.ProcessHistory{
			ID:                uuid.New(),
			TenantID:          tenantID,
			ProcessInstanceID: processInstance.ID,
			NodeID:            "node_2",
			NodeName:          "测试节点2",
			OperatorID:        uuid.New(),
			OperatorName:      "测试操作员2",
			Action:            model.ApprovalActionTransfer,
			FromStatus:        &fromStatus,
			ToStatus:          model.ProcessStatusApproved,
			CreatedAt:         time.Now().Add(-1 * time.Hour),
		}
		err = repo.Create(ctx, history2)
		require.NoError(t, err)

		history3 := &model.ProcessHistory{
			ID:                uuid.New(),
			TenantID:          tenantID,
			ProcessInstanceID: processInstance.ID,
			NodeID:            "node_3",
			NodeName:          "测试节点3",
			OperatorID:        uuid.New(),
			OperatorName:      "测试操作员3",
			Action:            model.ApprovalActionApprove,
			FromStatus:        &fromStatus,
			ToStatus:          model.ProcessStatusApproved,
			CreatedAt:         time.Now(),
		}
		err = repo.Create(ctx, history3)
		require.NoError(t, err)

		// List
		histories, err := repo.ListByInstance(ctx, processInstance.ID)
		assert.NoError(t, err)
		assert.Len(t, histories, 3)

		// Verify ordering (by created_at ASC)
		assert.Equal(t, history1.ID, histories[0].ID)
		assert.Equal(t, history2.ID, histories[1].ID)
		assert.Equal(t, history3.ID, histories[2].ID)

		// Verify actions
		assert.Equal(t, model.ApprovalActionApprove, histories[0].Action)
		assert.Equal(t, model.ApprovalActionTransfer, histories[1].Action)
		assert.Equal(t, model.ApprovalActionApprove, histories[2].Action)
	})

	t.Run("List histories for instance with no history", func(t *testing.T) {
		histories, err := repo.ListByInstance(ctx, uuid.New())
		assert.NoError(t, err)
		assert.Empty(t, histories)
	})

	t.Run("List histories shows approval trail", func(t *testing.T) {
		processInstance := createTestProcessInstance(t, db, tenantID)
		processInstanceRepo := NewProcessInstanceRepository(db)
		err := processInstanceRepo.Create(ctx, processInstance)
		require.NoError(t, err)

		// Simulate approval workflow
		actions := []struct {
			action  model.ApprovalAction
			comment string
			delay   time.Duration
		}{
			{model.ApprovalActionApprove, "第一级审批通过", -4 * time.Hour},
			{model.ApprovalActionApprove, "第二级审批通过", -3 * time.Hour},
			{model.ApprovalActionTransfer, "转交给主管", -2 * time.Hour},
			{model.ApprovalActionApprove, "主管审批通过", -1 * time.Hour},
		}

		fromStatus := model.ProcessStatusPending
		for _, a := range actions {
			comment := a.comment
			history := &model.ProcessHistory{
				ID:                uuid.New(),
				TenantID:          tenantID,
				ProcessInstanceID: processInstance.ID,
				NodeID:            "node_" + uuid.NewString()[:8],
				NodeName:          "测试节点",
				OperatorID:        uuid.New(),
				OperatorName:      "测试操作员",
				Action:            a.action,
				Comment:           &comment,
				FromStatus:        &fromStatus,
				ToStatus:          model.ProcessStatusApproved,
				CreatedAt:         time.Now().Add(a.delay),
			}
			err := repo.Create(ctx, history)
			require.NoError(t, err)
		}

		// List and verify trail
		histories, err := repo.ListByInstance(ctx, processInstance.ID)
		assert.NoError(t, err)
		assert.Len(t, histories, 4)

		// Verify chronological order
		for i, expected := range actions {
			assert.Equal(t, expected.action, histories[i].Action)
			assert.NotNil(t, histories[i].Comment)
			assert.Equal(t, expected.comment, *histories[i].Comment)
		}
	})

	t.Run("List histories only for specific instance", func(t *testing.T) {
		// Create two instances and insert them
		instance1 := createTestProcessInstance(t, db, tenantID)
		processInstanceRepo := NewProcessInstanceRepository(db)
		err := processInstanceRepo.Create(ctx, instance1)
		require.NoError(t, err)

		instance2 := createTestProcessInstance(t, db, tenantID)
		err = processInstanceRepo.Create(ctx, instance2)
		require.NoError(t, err)

		fromStatus := model.ProcessStatusPending
		// Create histories for instance1
		for i := 0; i < 3; i++ {
			history := &model.ProcessHistory{
				ID:                uuid.New(),
				TenantID:          tenantID,
				ProcessInstanceID: instance1.ID,
				NodeID:            "node_" + uuid.NewString()[:8],
				NodeName:          "测试节点",
				OperatorID:        uuid.New(),
				OperatorName:      "测试操作员",
				Action:            model.ApprovalActionApprove,
				FromStatus:        &fromStatus,
				ToStatus:          model.ProcessStatusApproved,
				CreatedAt:         time.Now(),
			}
			err := repo.Create(ctx, history)
			require.NoError(t, err)
		}

		// Create histories for instance2
		for i := 0; i < 2; i++ {
			history := &model.ProcessHistory{
				ID:                uuid.New(),
				TenantID:          tenantID,
				ProcessInstanceID: instance2.ID,
				NodeID:            "node_" + uuid.NewString()[:8],
				NodeName:          "测试节点",
				OperatorID:        uuid.New(),
				OperatorName:      "测试操作员",
				Action:            model.ApprovalActionApprove,
				FromStatus:        &fromStatus,
				ToStatus:          model.ProcessStatusApproved,
				CreatedAt:         time.Now(),
			}
			err := repo.Create(ctx, history)
			require.NoError(t, err)
		}

		// Verify isolation
		histories1, err := repo.ListByInstance(ctx, instance1.ID)
		assert.NoError(t, err)
		assert.Len(t, histories1, 3)

		histories2, err := repo.ListByInstance(ctx, instance2.ID)
		assert.NoError(t, err)
		assert.Len(t, histories2, 2)
	})
}

// TestProcessHistoryRepository_ListByTaskID tests listing by task ID
func TestProcessHistoryRepository_ListByTaskID(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessHistoryRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessHistories(t, db, tenantID)

	t.Run("List histories for task", func(t *testing.T) {
		taskID := uuid.New()

		// Create multiple history records for the same task
		history1 := createTestProcessHistory(t, db, tenantID)
		history1.TaskID = &taskID
		history1.Action = model.ApprovalActionApprove
		history1.CreatedAt = time.Now().Add(-1 * time.Hour)
		err := repo.Create(ctx, history1)
		require.NoError(t, err)

		history2 := createTestProcessHistory(t, db, tenantID)
		history2.TaskID = &taskID
		history2.Action = model.ApprovalActionApprove
		history2.CreatedAt = time.Now()
		err = repo.Create(ctx, history2)
		require.NoError(t, err)

		// List
		histories, err := repo.ListByTaskID(ctx, taskID)
		assert.NoError(t, err)
		assert.Len(t, histories, 2)

		// Verify all have the task ID
		for _, h := range histories {
			assert.NotNil(t, h.TaskID)
			assert.Equal(t, taskID, *h.TaskID)
		}

		// Verify ordering
		assert.Equal(t, history1.ID, histories[0].ID)
		assert.Equal(t, history2.ID, histories[1].ID)
	})

	t.Run("List histories for task with no history", func(t *testing.T) {
		histories, err := repo.ListByTaskID(ctx, uuid.New())
		assert.NoError(t, err)
		assert.Empty(t, histories)
	})

	t.Run("List histories only for specific task", func(t *testing.T) {
		taskID1 := uuid.New()
		taskID2 := uuid.New()

		// Create histories for task1
		for i := 0; i < 3; i++ {
			history := createTestProcessHistory(t, db, tenantID)
			history.TaskID = &taskID1
			err := repo.Create(ctx, history)
			require.NoError(t, err)
		}

		// Create histories for task2
		for i := 0; i < 2; i++ {
			history := createTestProcessHistory(t, db, tenantID)
			history.TaskID = &taskID2
			err := repo.Create(ctx, history)
			require.NoError(t, err)
		}

		// Verify isolation
		histories1, err := repo.ListByTaskID(ctx, taskID1)
		assert.NoError(t, err)
		assert.Len(t, histories1, 3)

		histories2, err := repo.ListByTaskID(ctx, taskID2)
		assert.NoError(t, err)
		assert.Len(t, histories2, 2)
	})

	t.Run("List task histories in chronological order", func(t *testing.T) {
		taskID := uuid.New()

		// Create histories with different timestamps
		times := []time.Duration{-3 * time.Hour, -2 * time.Hour, -1 * time.Hour}
		for _, delay := range times {
			history := createTestProcessHistory(t, db, tenantID)
			history.TaskID = &taskID
			history.CreatedAt = time.Now().Add(delay)
			err := repo.Create(ctx, history)
			require.NoError(t, err)
		}

		histories, err := repo.ListByTaskID(ctx, taskID)
		assert.NoError(t, err)
		assert.Len(t, histories, 3)

		// Verify chronological order (oldest first)
		for i := 0; i < len(histories)-1; i++ {
			assert.True(t, histories[i].CreatedAt.Before(histories[i+1].CreatedAt))
		}
	})
}

// TestProcessHistoryRepository_ApprovalTrail tests full approval trail scenario
func TestProcessHistoryRepository_ApprovalTrail(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessHistoryRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessHistories(t, db, tenantID)

	t.Run("Complete approval trail", func(t *testing.T) {
		processInstance := createTestProcessInstance(t, db, tenantID)
		processInstanceRepo := NewProcessInstanceRepository(db)
		err := processInstanceRepo.Create(ctx, processInstance)
		require.NoError(t, err)

		// Simulate complete approval workflow
		workflow := []struct {
			action   model.ApprovalAction
			operator string
			comment  string
			nodeID   string
		}{
			{model.ApprovalActionApprove, "申请人", "提交申请", "start"},
			{model.ApprovalActionApprove, "部门主管", "部门审批通过", "dept_manager"},
			{model.ApprovalActionTransfer, "部门主管", "转交给财务", "dept_manager"},
			{model.ApprovalActionApprove, "财务主管", "财务审批通过", "finance_manager"},
			{model.ApprovalActionApprove, "总经理", "最终批准", "ceo"},
		}

		fromStatus := model.ProcessStatusPending
		for i, step := range workflow {
			history := &model.ProcessHistory{
				ID:                uuid.New(),
				TenantID:          tenantID,
				ProcessInstanceID: processInstance.ID,
				NodeID:            step.nodeID,
				NodeName:          step.operator + "节点",
				Action:            step.action,
				OperatorID:        uuid.New(),
				OperatorName:      step.operator,
				Comment:           &step.comment,
				FromStatus:        &fromStatus,
				ToStatus:          model.ProcessStatusApproved,
				CreatedAt:         time.Now().Add(time.Duration(-len(workflow)+i) * time.Hour),
			}
			err := repo.Create(ctx, history)
			require.NoError(t, err)
		}

		// Retrieve and verify complete trail
		histories, err := repo.ListByInstance(ctx, processInstance.ID)
		assert.NoError(t, err)
		assert.Len(t, histories, 5)

		// Verify trail integrity
		for i, step := range workflow {
			assert.Equal(t, step.action, histories[i].Action)
			assert.Equal(t, step.nodeID, histories[i].NodeID)
			assert.NotNil(t, histories[i].Comment)
			assert.Equal(t, step.comment, *histories[i].Comment)
		}
	})
}

// TestProcessHistoryRepository_TenantIsolation tests tenant isolation
func TestProcessHistoryRepository_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessHistoryRepository(db)
	ctx := context.Background()
	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	defer cleanupProcessHistories(t, db, tenantID1)
	defer cleanupProcessHistories(t, db, tenantID2)

	t.Run("Histories isolated by tenant", func(t *testing.T) {
		// Create instances for different tenants and insert them
		instance1 := createTestProcessInstance(t, db, tenantID1)
		processInstanceRepo := NewProcessInstanceRepository(db)
		err := processInstanceRepo.Create(ctx, instance1)
		require.NoError(t, err)

		instance2 := createTestProcessInstance(t, db, tenantID2)
		err = processInstanceRepo.Create(ctx, instance2)
		require.NoError(t, err)

		fromStatus := model.ProcessStatusPending
		// Create histories for tenant1
		for i := 0; i < 3; i++ {
			history := &model.ProcessHistory{
				ID:                uuid.New(),
				TenantID:          tenantID1,
				ProcessInstanceID: instance1.ID,
				NodeID:            "node_" + uuid.NewString()[:8],
				NodeName:          "测试节点",
				OperatorID:        uuid.New(),
				OperatorName:      "测试操作员",
				Action:            model.ApprovalActionApprove,
				FromStatus:        &fromStatus,
				ToStatus:          model.ProcessStatusApproved,
				CreatedAt:         time.Now(),
			}
			err := repo.Create(ctx, history)
			require.NoError(t, err)
		}

		// Create histories for tenant2
		for i := 0; i < 2; i++ {
			history := &model.ProcessHistory{
				ID:                uuid.New(),
				TenantID:          tenantID2,
				ProcessInstanceID: instance2.ID,
				NodeID:            "node_" + uuid.NewString()[:8],
				NodeName:          "测试节点",
				OperatorID:        uuid.New(),
				OperatorName:      "测试操作员",
				Action:            model.ApprovalActionApprove,
				FromStatus:        &fromStatus,
				ToStatus:          model.ProcessStatusApproved,
				CreatedAt:         time.Now(),
			}
			err := repo.Create(ctx, history)
			require.NoError(t, err)
		}

		// Verify tenant isolation
		histories1, err := repo.ListByInstance(ctx, instance1.ID)
		assert.NoError(t, err)
		assert.Len(t, histories1, 3)

		histories2, err := repo.ListByInstance(ctx, instance2.ID)
		assert.NoError(t, err)
		assert.Len(t, histories2, 2)

		// Verify all histories belong to correct tenant
		for _, h := range histories1 {
			assert.Equal(t, tenantID1, h.TenantID)
		}
		for _, h := range histories2 {
			assert.Equal(t, tenantID2, h.TenantID)
		}
	})
}

// ======================== Benchmark Tests ========================

// BenchmarkProcessHistoryRepository_Create benchmarks history creation
func BenchmarkProcessHistoryRepository_Create(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewProcessHistoryRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		history := createTestProcessHistory(&testing.T{}, db, tenantID)
		_ = repo.Create(ctx, history)
	}
}

// BenchmarkProcessHistoryRepository_ListByInstance benchmarks listing by instance
func BenchmarkProcessHistoryRepository_ListByInstance(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewProcessHistoryRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	// Create test process instance and histories
	processInstance := createTestProcessInstance(&testing.T{}, db, tenantID)
	for i := 0; i < 10; i++ {
		history := createTestProcessHistory(&testing.T{}, db, tenantID)
		history.ProcessInstanceID = processInstance.ID
		_ = repo.Create(ctx, history)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.ListByInstance(ctx, processInstance.ID)
	}
}

// BenchmarkProcessHistoryRepository_ListByTaskID benchmarks listing by task ID
func BenchmarkProcessHistoryRepository_ListByTaskID(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewProcessHistoryRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	taskID := uuid.New()

	// Create test histories
	for i := 0; i < 10; i++ {
		history := createTestProcessHistory(&testing.T{}, db, tenantID)
		history.TaskID = &taskID
		_ = repo.Create(ctx, history)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.ListByTaskID(ctx, taskID)
	}
}
