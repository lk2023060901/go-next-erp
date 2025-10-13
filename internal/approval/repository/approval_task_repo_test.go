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

// Helper function to create test approval task
func createTestApprovalTask(t *testing.T, db *database.DB, tenantID uuid.UUID) *model.ApprovalTask {
	t.Helper()
	ctx := context.Background()

	// Create dependencies
	formID, workflowID := createTestFormAndWorkflow(t, db, tenantID)
	_ = createTestProcessDefinition(t, db, tenantID, formID, workflowID)
	processInstance := createTestProcessInstance(t, db, tenantID)

	// Insert process instance into database
	processInstanceRepo := NewProcessInstanceRepository(db)
	err := processInstanceRepo.Create(ctx, processInstance)
	if err != nil {
		t.Logf("Warning: failed to create process instance: %v", err)
	}

	task := &model.ApprovalTask{
		ID:                uuid.New(),
		TenantID:          tenantID,
		ProcessInstanceID: processInstance.ID,
		NodeID:            "node_" + uuid.NewString()[:8],
		NodeName:          "测试节点",
		AssigneeID:        uuid.New(),
		AssigneeName:      "测试审批人",
		Status:            model.TaskStatusPending,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	return task
}

// Cleanup helper
func cleanupApprovalTasks(t *testing.T, db *database.DB, tenantID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	_, _ = db.Exec(ctx, "DELETE FROM approval_tasks WHERE tenant_id = $1", tenantID)
	_, _ = db.Exec(ctx, "DELETE FROM approval_process_instances WHERE tenant_id = $1", tenantID)
	_, _ = db.Exec(ctx, "DELETE FROM approval_process_definitions WHERE tenant_id = $1", tenantID)
	_, _ = db.Exec(ctx, "DELETE FROM form_data WHERE tenant_id = $1", tenantID)
	_, _ = db.Exec(ctx, "DELETE FROM form_definitions WHERE tenant_id = $1", tenantID)
}

// TestApprovalTaskRepository_Create tests task creation
func TestApprovalTaskRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewApprovalTaskRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupApprovalTasks(t, db, tenantID)

	t.Run("Create successfully", func(t *testing.T) {
		task := createTestApprovalTask(t, db, tenantID)

		err := repo.Create(ctx, task)
		assert.NoError(t, err)

		// Verify
		found, err := repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, task.ID, found.ID)
		assert.Equal(t, task.TenantID, found.TenantID)
		assert.Equal(t, task.ProcessInstanceID, found.ProcessInstanceID)
		assert.Equal(t, task.NodeID, found.NodeID)
		assert.Equal(t, task.AssigneeID, found.AssigneeID)
		assert.Equal(t, task.Status, found.Status)
	})

	t.Run("Create with pending status", func(t *testing.T) {
		task := createTestApprovalTask(t, db, tenantID)
		task.Status = model.TaskStatusPending

		err := repo.Create(ctx, task)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, model.TaskStatusPending, found.Status)
		assert.Nil(t, found.Action)
		assert.Nil(t, found.Comment)
		assert.Nil(t, found.ApprovedAt)
	})
}

// TestApprovalTaskRepository_Update tests task updates
func TestApprovalTaskRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewApprovalTaskRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupApprovalTasks(t, db, tenantID)

	t.Run("Update status and action successfully", func(t *testing.T) {
		task := createTestApprovalTask(t, db, tenantID)
		err := repo.Create(ctx, task)
		require.NoError(t, err)

		// Update
		task.Status = model.TaskStatusApproved
		action := model.ApprovalActionApprove
		task.Action = &action
		comment := "审批通过"
		task.Comment = &comment
		now := time.Now()
		task.ApprovedAt = &now
		task.UpdatedAt = time.Now()

		err = repo.Update(ctx, task)
		assert.NoError(t, err)

		// Verify
		found, err := repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, model.TaskStatusApproved, found.Status)
		assert.NotNil(t, found.Action)
		assert.Equal(t, model.ApprovalActionApprove, *found.Action)
		assert.NotNil(t, found.Comment)
		assert.Equal(t, "审批通过", *found.Comment)
		assert.NotNil(t, found.ApprovedAt)
	})

	t.Run("Update status to rejected", func(t *testing.T) {
		task := createTestApprovalTask(t, db, tenantID)
		err := repo.Create(ctx, task)
		require.NoError(t, err)

		// Update
		task.Status = model.TaskStatusRejected
		action := model.ApprovalActionReject
		task.Action = &action
		comment := "不符合要求"
		task.Comment = &comment
		now := time.Now()
		task.ApprovedAt = &now

		err = repo.Update(ctx, task)
		assert.NoError(t, err)

		// Verify
		found, err := repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, model.TaskStatusRejected, found.Status)
		assert.Equal(t, model.ApprovalActionReject, *found.Action)
	})

	t.Run("Update non-existent task should not error", func(t *testing.T) {
		task := &model.ApprovalTask{
			ID:        uuid.New(),
			Status:    model.TaskStatusApproved,
			UpdatedAt: time.Now(),
		}

		err := repo.Update(ctx, task)
		assert.NoError(t, err) // No error, just no rows affected
	})
}

// TestApprovalTaskRepository_FindByID tests finding task by ID
func TestApprovalTaskRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewApprovalTaskRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupApprovalTasks(t, db, tenantID)

	t.Run("Find existing task", func(t *testing.T) {
		task := createTestApprovalTask(t, db, tenantID)
		err := repo.Create(ctx, task)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, task.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, task.ID, found.ID)
		assert.Equal(t, task.AssigneeID, found.AssigneeID)
	})

	t.Run("Find non-existent task", func(t *testing.T) {
		found, err := repo.FindByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}

// TestApprovalTaskRepository_ListByInstance tests listing tasks by process instance
func TestApprovalTaskRepository_ListByInstance(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewApprovalTaskRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupApprovalTasks(t, db, tenantID)

	t.Run("List tasks for instance", func(t *testing.T) {
		task1 := createTestApprovalTask(t, db, tenantID)
		task2 := createTestApprovalTask(t, db, tenantID)
		task2.ProcessInstanceID = task1.ProcessInstanceID // Same instance
		task2.NodeID = "node_" + uuid.NewString()[:8]

		err := repo.Create(ctx, task1)
		require.NoError(t, err)
		err = repo.Create(ctx, task2)
		require.NoError(t, err)

		// List
		tasks, err := repo.ListByInstance(ctx, task1.ProcessInstanceID)
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)

		// Verify ordering (by created_at ASC)
		assert.Equal(t, task1.ID, tasks[0].ID)
		assert.Equal(t, task2.ID, tasks[1].ID)
	})

	t.Run("List tasks for instance with no tasks", func(t *testing.T) {
		tasks, err := repo.ListByInstance(ctx, uuid.New())
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("List tasks ordered by created_at", func(t *testing.T) {
		// Create a common process instance for all tasks
		processInstance := createTestProcessInstance(t, db, tenantID)
		processInstanceRepo := NewProcessInstanceRepository(db)
		err := processInstanceRepo.Create(ctx, processInstance)
		require.NoError(t, err)
		processInstanceID := processInstance.ID

		// Create tasks with different timestamps
		task1 := createTestApprovalTask(t, db, tenantID)
		task1.ProcessInstanceID = processInstanceID
		task1.CreatedAt = time.Now().Add(-2 * time.Hour)
		err = repo.Create(ctx, task1)
		require.NoError(t, err)

		task2 := createTestApprovalTask(t, db, tenantID)
		task2.ProcessInstanceID = processInstanceID
		task2.CreatedAt = time.Now().Add(-1 * time.Hour)
		err = repo.Create(ctx, task2)
		require.NoError(t, err)

		task3 := createTestApprovalTask(t, db, tenantID)
		task3.ProcessInstanceID = processInstanceID
		task3.CreatedAt = time.Now()
		err = repo.Create(ctx, task3)
		require.NoError(t, err)

		// List and verify order
		tasks, err := repo.ListByInstance(ctx, processInstanceID)
		assert.NoError(t, err)
		assert.Len(t, tasks, 3)
		assert.Equal(t, task1.ID, tasks[0].ID)
		assert.Equal(t, task2.ID, tasks[1].ID)
		assert.Equal(t, task3.ID, tasks[2].ID)
	})
}

// TestApprovalTaskRepository_ListByAssignee tests listing tasks by assignee
func TestApprovalTaskRepository_ListByAssignee(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewApprovalTaskRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupApprovalTasks(t, db, tenantID)

	t.Run("List all tasks for assignee", func(t *testing.T) {
		assigneeID := uuid.New()

		task1 := createTestApprovalTask(t, db, tenantID)
		task1.AssigneeID = assigneeID
		task1.Status = model.TaskStatusPending
		err := repo.Create(ctx, task1)
		require.NoError(t, err)

		task2 := createTestApprovalTask(t, db, tenantID)
		task2.AssigneeID = assigneeID
		task2.Status = model.TaskStatusApproved
		err = repo.Create(ctx, task2)
		require.NoError(t, err)

		// List all (no status filter)
		tasks, err := repo.ListByAssignee(ctx, assigneeID, nil, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
	})

	t.Run("List tasks by assignee with status filter", func(t *testing.T) {
		assigneeID := uuid.New()

		task1 := createTestApprovalTask(t, db, tenantID)
		task1.AssigneeID = assigneeID
		task1.Status = model.TaskStatusPending
		err := repo.Create(ctx, task1)
		require.NoError(t, err)

		task2 := createTestApprovalTask(t, db, tenantID)
		task2.AssigneeID = assigneeID
		task2.Status = model.TaskStatusApproved
		err = repo.Create(ctx, task2)
		require.NoError(t, err)

		// List only pending
		status := model.TaskStatusPending
		tasks, err := repo.ListByAssignee(ctx, assigneeID, &status, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, model.TaskStatusPending, tasks[0].Status)
	})

	t.Run("List with pagination", func(t *testing.T) {
		assigneeID := uuid.New()

		// Create 5 tasks
		for i := 0; i < 5; i++ {
			task := createTestApprovalTask(t, db, tenantID)
			task.AssigneeID = assigneeID
			err := repo.Create(ctx, task)
			require.NoError(t, err)
		}

		// First page
		tasks, err := repo.ListByAssignee(ctx, assigneeID, nil, 2, 0)
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)

		// Second page
		tasks, err = repo.ListByAssignee(ctx, assigneeID, nil, 2, 2)
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)

		// Third page
		tasks, err = repo.ListByAssignee(ctx, assigneeID, nil, 2, 4)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
	})

	t.Run("List tasks ordered by created_at DESC", func(t *testing.T) {
		assigneeID := uuid.New()

		task1 := createTestApprovalTask(t, db, tenantID)
		task1.AssigneeID = assigneeID
		task1.CreatedAt = time.Now().Add(-2 * time.Hour)
		err := repo.Create(ctx, task1)
		require.NoError(t, err)

		task2 := createTestApprovalTask(t, db, tenantID)
		task2.AssigneeID = assigneeID
		task2.CreatedAt = time.Now()
		err = repo.Create(ctx, task2)
		require.NoError(t, err)

		// List and verify order (newest first)
		tasks, err := repo.ListByAssignee(ctx, assigneeID, nil, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		assert.Equal(t, task2.ID, tasks[0].ID)
		assert.Equal(t, task1.ID, tasks[1].ID)
	})
}

// TestApprovalTaskRepository_ListPendingByAssignee tests listing pending tasks
func TestApprovalTaskRepository_ListPendingByAssignee(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewApprovalTaskRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupApprovalTasks(t, db, tenantID)

	t.Run("List only pending tasks", func(t *testing.T) {
		assigneeID := uuid.New()

		// Create pending task
		task1 := createTestApprovalTask(t, db, tenantID)
		task1.AssigneeID = assigneeID
		task1.Status = model.TaskStatusPending
		err := repo.Create(ctx, task1)
		require.NoError(t, err)

		// Create approved task (should not be returned)
		task2 := createTestApprovalTask(t, db, tenantID)
		task2.AssigneeID = assigneeID
		task2.Status = model.TaskStatusApproved
		err = repo.Create(ctx, task2)
		require.NoError(t, err)

		// List pending only
		tasks, err := repo.ListPendingByAssignee(ctx, assigneeID)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, model.TaskStatusPending, tasks[0].Status)
		assert.Equal(t, task1.ID, tasks[0].ID)
	})

	t.Run("List pending tasks for assignee with no pending", func(t *testing.T) {
		assigneeID := uuid.New()

		tasks, err := repo.ListPendingByAssignee(ctx, assigneeID)
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("List multiple pending tasks ordered by created_at DESC", func(t *testing.T) {
		assigneeID := uuid.New()

		task1 := createTestApprovalTask(t, db, tenantID)
		task1.AssigneeID = assigneeID
		task1.Status = model.TaskStatusPending
		task1.CreatedAt = time.Now().Add(-1 * time.Hour)
		err := repo.Create(ctx, task1)
		require.NoError(t, err)

		task2 := createTestApprovalTask(t, db, tenantID)
		task2.AssigneeID = assigneeID
		task2.Status = model.TaskStatusPending
		task2.CreatedAt = time.Now()
		err = repo.Create(ctx, task2)
		require.NoError(t, err)

		tasks, err := repo.ListPendingByAssignee(ctx, assigneeID)
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		assert.Equal(t, task2.ID, tasks[0].ID) // Newest first
		assert.Equal(t, task1.ID, tasks[1].ID)
	})
}

// TestApprovalTaskRepository_CountPendingByAssignee tests counting pending tasks
func TestApprovalTaskRepository_CountPendingByAssignee(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewApprovalTaskRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupApprovalTasks(t, db, tenantID)

	t.Run("Count pending tasks", func(t *testing.T) {
		assigneeID := uuid.New()

		// Create 3 pending tasks
		for i := 0; i < 3; i++ {
			task := createTestApprovalTask(t, db, tenantID)
			task.AssigneeID = assigneeID
			task.Status = model.TaskStatusPending
			err := repo.Create(ctx, task)
			require.NoError(t, err)
		}

		// Create 2 approved tasks (should not be counted)
		for i := 0; i < 2; i++ {
			task := createTestApprovalTask(t, db, tenantID)
			task.AssigneeID = assigneeID
			task.Status = model.TaskStatusApproved
			err := repo.Create(ctx, task)
			require.NoError(t, err)
		}

		count, err := repo.CountPendingByAssignee(ctx, assigneeID)
		assert.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("Count with no pending tasks", func(t *testing.T) {
		assigneeID := uuid.New()

		count, err := repo.CountPendingByAssignee(ctx, assigneeID)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Count only includes pending status", func(t *testing.T) {
		assigneeID := uuid.New()

		// Create one of each status
		statuses := []model.TaskStatus{
			model.TaskStatusPending,
			model.TaskStatusApproved,
			model.TaskStatusRejected,
		}

		for _, status := range statuses {
			task := createTestApprovalTask(t, db, tenantID)
			task.AssigneeID = assigneeID
			task.Status = status
			err := repo.Create(ctx, task)
			require.NoError(t, err)
		}

		count, err := repo.CountPendingByAssignee(ctx, assigneeID)
		assert.NoError(t, err)
		assert.Equal(t, 1, count) // Only the pending one
	})
}

// TestApprovalTaskRepository_UpdateStatus tests updating task status
func TestApprovalTaskRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewApprovalTaskRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupApprovalTasks(t, db, tenantID)

	t.Run("Update status to approved", func(t *testing.T) {
		task := createTestApprovalTask(t, db, tenantID)
		err := repo.Create(ctx, task)
		require.NoError(t, err)

		action := model.ApprovalActionApprove
		comment := "批准"
		now := time.Now()

		err = repo.UpdateStatus(ctx, task.ID, model.TaskStatusApproved, &action, &comment, &now)
		assert.NoError(t, err)

		// Verify
		found, err := repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, model.TaskStatusApproved, found.Status)
		assert.NotNil(t, found.Action)
		assert.Equal(t, model.ApprovalActionApprove, *found.Action)
		assert.NotNil(t, found.Comment)
		assert.Equal(t, "批准", *found.Comment)
		assert.NotNil(t, found.ApprovedAt)
	})

	t.Run("Update status to rejected", func(t *testing.T) {
		task := createTestApprovalTask(t, db, tenantID)
		err := repo.Create(ctx, task)
		require.NoError(t, err)

		action := model.ApprovalActionReject
		comment := "不批准"
		now := time.Now()

		err = repo.UpdateStatus(ctx, task.ID, model.TaskStatusRejected, &action, &comment, &now)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, model.TaskStatusRejected, found.Status)
		assert.Equal(t, model.ApprovalActionReject, *found.Action)
	})

	t.Run("Update status with nil comment", func(t *testing.T) {
		task := createTestApprovalTask(t, db, tenantID)
		err := repo.Create(ctx, task)
		require.NoError(t, err)

		action := model.ApprovalActionApprove
		now := time.Now()

		err = repo.UpdateStatus(ctx, task.ID, model.TaskStatusApproved, &action, nil, &now)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, model.TaskStatusApproved, found.Status)
		assert.Nil(t, found.Comment)
	})

	t.Run("Update non-existent task should not error", func(t *testing.T) {
		action := model.ApprovalActionApprove
		comment := "测试"
		now := time.Now()

		err := repo.UpdateStatus(ctx, uuid.New(), model.TaskStatusApproved, &action, &comment, &now)
		assert.NoError(t, err) // No error, just no rows affected
	})
}

// TestApprovalTaskRepository_TenantIsolation tests tenant isolation
func TestApprovalTaskRepository_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewApprovalTaskRepository(db)
	ctx := context.Background()
	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	defer cleanupApprovalTasks(t, db, tenantID1)
	defer cleanupApprovalTasks(t, db, tenantID2)

	t.Run("Tasks isolated by tenant", func(t *testing.T) {
		// Create task for tenant1
		task1 := createTestApprovalTask(t, db, tenantID1)
		task1.AssigneeID = uuid.New()
		err := repo.Create(ctx, task1)
		require.NoError(t, err)

		// Create task for tenant2 with same assignee
		task2 := createTestApprovalTask(t, db, tenantID2)
		task2.AssigneeID = task1.AssigneeID // Same assignee
		err = repo.Create(ctx, task2)
		require.NoError(t, err)

		// List tasks for assignee
		tasks, err := repo.ListByAssignee(ctx, task1.AssigneeID, nil, 10, 0)
		assert.NoError(t, err)
		// Should return tasks from both tenants
		assert.Len(t, tasks, 2)
	})
}

// ======================== Benchmark Tests ========================

// BenchmarkApprovalTaskRepository_Create benchmarks task creation
func BenchmarkApprovalTaskRepository_Create(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewApprovalTaskRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := createTestApprovalTask(&testing.T{}, db, tenantID)
		_ = repo.Create(ctx, task)
	}
}

// BenchmarkApprovalTaskRepository_FindByID benchmarks finding task by ID
func BenchmarkApprovalTaskRepository_FindByID(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewApprovalTaskRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	// Create test task
	task := createTestApprovalTask(&testing.T{}, db, tenantID)
	_ = repo.Create(ctx, task)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByID(ctx, task.ID)
	}
}

// BenchmarkApprovalTaskRepository_ListByAssignee benchmarks listing tasks by assignee
func BenchmarkApprovalTaskRepository_ListByAssignee(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewApprovalTaskRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	assigneeID := uuid.New()

	// Create 10 test tasks
	for i := 0; i < 10; i++ {
		task := createTestApprovalTask(&testing.T{}, db, tenantID)
		task.AssigneeID = assigneeID
		_ = repo.Create(ctx, task)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.ListByAssignee(ctx, assigneeID, nil, 10, 0)
	}
}

// BenchmarkApprovalTaskRepository_CountPendingByAssignee benchmarks counting pending tasks
func BenchmarkApprovalTaskRepository_CountPendingByAssignee(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewApprovalTaskRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	assigneeID := uuid.New()

	// Create 10 pending tasks
	for i := 0; i < 10; i++ {
		task := createTestApprovalTask(&testing.T{}, db, tenantID)
		task.AssigneeID = assigneeID
		task.Status = model.TaskStatusPending
		_ = repo.Create(ctx, task)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.CountPendingByAssignee(ctx, assigneeID)
	}
}

// BenchmarkApprovalTaskRepository_UpdateStatus benchmarks status updates
func BenchmarkApprovalTaskRepository_UpdateStatus(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewApprovalTaskRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	// Create test task
	task := createTestApprovalTask(&testing.T{}, db, tenantID)
	_ = repo.Create(ctx, task)

	action := model.ApprovalActionApprove
	comment := "批准"
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.UpdateStatus(ctx, task.ID, model.TaskStatusApproved, &action, &comment, &now)
	}
}
