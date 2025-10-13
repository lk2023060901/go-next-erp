package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lk2023060901/go-next-erp/internal/approval/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// cleanupProcessInstances 清理测试数据
func cleanupProcessInstances(t *testing.T, db *database.DB, tenantID uuid.UUID) {
	t.Helper()
	_, _ = db.Exec(context.Background(),
		"DELETE FROM approval_process_instances WHERE tenant_id = $1", tenantID)
}

// createTestProcessInstance 创建测试流程实例（不插入数据库）
func createTestProcessInstance(t *testing.T, db *database.DB, tenantID uuid.UUID) *model.ProcessInstance {
	t.Helper()
	ctx := context.Background()

	// 创建必需的前置数据
	formID, workflowID := createTestFormAndWorkflow(t, db, tenantID)
	processDefID := createTestProcessDefinition(t, db, tenantID, formID, workflowID)

	// 先创建form_data记录（因为process_instance需要form_data_id外键）
	formDataID := uuid.New()
	_, err := db.Exec(ctx, `
		INSERT INTO form_data (id, tenant_id, form_id, data, submitted_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`, formDataID, tenantID, formID, `{"test":"data"}`, uuid.New(), time.Now(), time.Now())
	if err != nil {
		t.Logf("Warning: failed to create test form_data: %v", err)
	}

	instance := &model.ProcessInstance{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		ProcessDefID:       processDefID,
		ProcessDefCode:     "TEST_PROCESS",
		ProcessDefName:     "测试流程",
		WorkflowInstanceID: uuid.New(),
		FormDataID:         formDataID,
		ApplicantID:        uuid.New(),
		ApplicantName:      "测试申请人",
		Title:              "测试流程实例",
		Status:             model.ProcessStatusPending,
		Variables: map[string]interface{}{
			"amount": 1000,
			"days":   3,
		},
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return instance
}

// createTestProcessDefinition 创建测试流程定义
func createTestProcessDefinition(t *testing.T, db *database.DB, tenantID, formID, workflowID uuid.UUID) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	processDefID := uuid.New()
	_, err := db.Exec(ctx, `
		INSERT INTO approval_process_definitions (
			id, tenant_id, code, name, category, form_id, workflow_id,
			enabled, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO NOTHING
	`, processDefID, tenantID, uuid.NewString(), "测试流程定义", "测试",
	   formID, workflowID, true, uuid.New(), time.Now(), time.Now())

	if err != nil {
		t.Logf("Warning: failed to create test process definition: %v", err)
	}

	return processDefID
}

func TestProcessInstanceRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessInstanceRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessInstances(t, db, tenantID)

	t.Run("Create process instance successfully", func(t *testing.T) {
		instance := createTestProcessInstance(t, db, tenantID)

		err := repo.Create(ctx, instance)
		assert.NoError(t, err)

		// 验证是否成功插入
		found, err := repo.FindByID(ctx, instance.ID)
		require.NoError(t, err)
		assert.Equal(t, instance.ID, found.ID)
		assert.Equal(t, instance.Status, found.Status)
		assert.Equal(t, instance.ProcessDefID, found.ProcessDefID)
		assert.Equal(t, instance.ApplicantID, found.ApplicantID)

		// 验证Variables正确序列化
		assert.NotNil(t, found.Variables)
		assert.Equal(t, 1000, int(found.Variables["amount"].(float64)))
		assert.Equal(t, 3, int(found.Variables["days"].(float64)))
	})

	t.Run("Create with empty variables", func(t *testing.T) {
		instance := createTestProcessInstance(t, db, tenantID)
		instance.Variables = map[string]interface{}{}

		err := repo.Create(ctx, instance)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, instance.ID)
		require.NoError(t, err)
		assert.NotNil(t, found.Variables)
		assert.Empty(t, found.Variables)
	})

	t.Run("Create with complex variables", func(t *testing.T) {
		instance := createTestProcessInstance(t, db, tenantID)
		instance.Variables = map[string]interface{}{
			"nested": map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
			"array": []string{"item1", "item2"},
			"bool":  true,
		}

		err := repo.Create(ctx, instance)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, instance.ID)
		require.NoError(t, err)
		assert.NotNil(t, found.Variables["nested"])
		assert.NotNil(t, found.Variables["array"])
		assert.True(t, found.Variables["bool"].(bool))
	})
}

func TestProcessInstanceRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessInstanceRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessInstances(t, db, tenantID)

	t.Run("Update process instance status", func(t *testing.T) {
		// 创建
		instance := createTestProcessInstance(t, db, tenantID)
		err := repo.Create(ctx, instance)
		require.NoError(t, err)

		// 更新状态
		instance.Status = model.ProcessStatusApproved
		completedAt := time.Now()
		instance.CompletedAt = &completedAt
		instance.Variables["approved_by"] = "manager"
		instance.UpdatedAt = time.Now()

		err = repo.Update(ctx, instance)
		assert.NoError(t, err)

		// 验证更新
		found, err := repo.FindByID(ctx, instance.ID)
		require.NoError(t, err)
		assert.Equal(t, model.ProcessStatusApproved, found.Status)
		assert.NotNil(t, found.CompletedAt)
		assert.NotNil(t, found.Variables["approved_by"])
	})

	t.Run("Update current node", func(t *testing.T) {
		instance := createTestProcessInstance(t, db, tenantID)
		err := repo.Create(ctx, instance)
		require.NoError(t, err)

		// 更新当前节点
		currentNodeID := "node_2"
		instance.CurrentNodeID = &currentNodeID
		instance.UpdatedAt = time.Now()

		err = repo.Update(ctx, instance)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, instance.ID)
		require.NoError(t, err)
		require.NotNil(t, found.CurrentNodeID)
		assert.Equal(t, currentNodeID, *found.CurrentNodeID)
	})
}

func TestProcessInstanceRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessInstanceRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessInstances(t, db, tenantID)

	t.Run("Find existing instance", func(t *testing.T) {
		instance := createTestProcessInstance(t, db, tenantID)
		err := repo.Create(ctx, instance)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, instance.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, instance.ID, found.ID)
	})

	t.Run("Find non-existent instance", func(t *testing.T) {
		_, err := repo.FindByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Equal(t, pgx.ErrNoRows, err)
	})
}

func TestProcessInstanceRepository_FindByWorkflowInstanceID(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessInstanceRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessInstances(t, db, tenantID)

	t.Run("Find by workflow instance ID", func(t *testing.T) {
		instance := createTestProcessInstance(t, db, tenantID)
		err := repo.Create(ctx, instance)
		require.NoError(t, err)

		found, err := repo.FindByWorkflowInstanceID(ctx, instance.WorkflowInstanceID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, instance.ID, found.ID)
		assert.Equal(t, instance.WorkflowInstanceID, found.WorkflowInstanceID)
	})

	t.Run("Find by non-existent workflow instance ID", func(t *testing.T) {
		_, err := repo.FindByWorkflowInstanceID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Equal(t, pgx.ErrNoRows, err)
	})
}

func TestProcessInstanceRepository_ListByApplicant(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessInstanceRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessInstances(t, db, tenantID)

	t.Run("List instances by applicant", func(t *testing.T) {
		applicantID := uuid.New()

		// 创建多个实例
		for i := 0; i < 3; i++ {
			instance := createTestProcessInstance(t, db, tenantID)
			instance.ApplicantID = applicantID
			instance.StartedAt = time.Now().Add(time.Duration(i) * time.Second)
			err := repo.Create(ctx, instance)
			require.NoError(t, err)
		}

		// 查询
		instances, err := repo.ListByApplicant(ctx, applicantID, 10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(instances), 3)

		// 验证排序（按开始时间降序）
		for i := 0; i < len(instances)-1; i++ {
			assert.True(t, instances[i].StartedAt.After(instances[i+1].StartedAt) ||
				instances[i].StartedAt.Equal(instances[i+1].StartedAt))
		}
	})

	t.Run("List with pagination", func(t *testing.T) {
		applicantID := uuid.New()

		// 创建5个实例
		for i := 0; i < 5; i++ {
			instance := createTestProcessInstance(t, db, tenantID)
			instance.ApplicantID = applicantID
			err := repo.Create(ctx, instance)
			require.NoError(t, err)
		}

		// 第一页
		page1, err := repo.ListByApplicant(ctx, applicantID, 2, 0)
		assert.NoError(t, err)
		assert.Len(t, page1, 2)

		// 第二页
		page2, err := repo.ListByApplicant(ctx, applicantID, 2, 2)
		assert.NoError(t, err)
		assert.Len(t, page2, 2)

		// 验证不重复
		assert.NotEqual(t, page1[0].ID, page2[0].ID)
	})
}

func TestProcessInstanceRepository_ListByStatus(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessInstanceRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessInstances(t, db, tenantID)

	t.Run("List by status", func(t *testing.T) {
		// 创建不同状态的实例
		statuses := []model.ProcessStatus{
			model.ProcessStatusPending,
			model.ProcessStatusPending,
			model.ProcessStatusApproved,
			model.ProcessStatusRejected,
		}

		for _, status := range statuses {
			instance := createTestProcessInstance(t, db, tenantID)
			instance.Status = status
			err := repo.Create(ctx, instance)
			require.NoError(t, err)
		}

		// 查询待审批的实例
		pendingInstances, err := repo.ListByStatus(ctx, tenantID, model.ProcessStatusPending, 10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(pendingInstances), 2)

		// 验证状态
		for _, inst := range pendingInstances {
			assert.Equal(t, model.ProcessStatusPending, inst.Status)
		}
	})

	t.Run("Tenant isolation", func(t *testing.T) {
		// 在当前租户创建实例
		instance1 := createTestProcessInstance(t, db, tenantID)
		instance1.Status = model.ProcessStatusPending
		err := repo.Create(ctx, instance1)
		require.NoError(t, err)

		// 在不同租户查询
		differentTenantID := uuid.New()
		instances, err := repo.ListByStatus(ctx, differentTenantID, model.ProcessStatusPending, 10, 0)
		assert.NoError(t, err)
		assert.Empty(t, instances)
	})
}

func TestProcessInstanceRepository_ListByProcessDef(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessInstanceRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessInstances(t, db, tenantID)

	t.Run("List by process definition", func(t *testing.T) {
		formID, workflowID := createTestFormAndWorkflow(t, db, tenantID)
		processDefID := createTestProcessDefinition(t, db, tenantID, formID, workflowID)

		// 创建多个实例
		for i := 0; i < 3; i++ {
			instance := createTestProcessInstance(t, db, tenantID)
			instance.ProcessDefID = processDefID
			err := repo.Create(ctx, instance)
			require.NoError(t, err)
		}

		// 查询
		instances, err := repo.ListByProcessDef(ctx, processDefID, 10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(instances), 3)

		// 验证都属于同一流程定义
		for _, inst := range instances {
			assert.Equal(t, processDefID, inst.ProcessDefID)
		}
	})
}

func TestProcessInstanceRepository_CountByStatus(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessInstanceRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessInstances(t, db, tenantID)

	t.Run("Count by status", func(t *testing.T) {
		// 创建3个待审批实例
		for i := 0; i < 3; i++ {
			instance := createTestProcessInstance(t, db, tenantID)
			instance.Status = model.ProcessStatusPending
			err := repo.Create(ctx, instance)
			require.NoError(t, err)
		}

		// 创建2个已批准实例
		for i := 0; i < 2; i++ {
			instance := createTestProcessInstance(t, db, tenantID)
			instance.Status = model.ProcessStatusApproved
			err := repo.Create(ctx, instance)
			require.NoError(t, err)
		}

		// 统计待审批数量
		count, err := repo.CountByStatus(ctx, tenantID, model.ProcessStatusPending)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3)

		// 统计已批准数量
		count, err = repo.CountByStatus(ctx, tenantID, model.ProcessStatusApproved)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, 2)
	})
}

// ===== Benchmark Tests =====

func BenchmarkProcessInstanceRepository_Create(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessInstanceRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	// 准备前置数据
	formID, workflowID := createTestFormAndWorkflow(&testing.T{}, db, tenantID)
	processDefID := createTestProcessDefinition(&testing.T{}, db, tenantID, formID, workflowID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		instance := &model.ProcessInstance{
			ID:                 uuid.New(),
			TenantID:           tenantID,
			ProcessDefID:       processDefID,
			WorkflowInstanceID: uuid.New(),
			FormDataID:         uuid.New(),
			ApplicantID:        uuid.New(),
			Status:             model.ProcessStatusPending,
			Variables:          map[string]interface{}{"test": "data"},
			StartedAt:          time.Now(),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		_ = repo.Create(ctx, instance)
	}
}

func BenchmarkProcessInstanceRepository_FindByID(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessInstanceRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	// 准备测试数据
	instance := createTestProcessInstance(&testing.T{}, db, tenantID)
	_ = repo.Create(ctx, instance)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByID(ctx, instance.ID)
	}
}

func BenchmarkProcessInstanceRepository_ListByApplicant(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessInstanceRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	applicantID := uuid.New()

	// 准备10条测试数据
	for i := 0; i < 10; i++ {
		instance := createTestProcessInstance(&testing.T{}, db, tenantID)
		instance.ApplicantID = applicantID
		_ = repo.Create(ctx, instance)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.ListByApplicant(ctx, applicantID, 10, 0)
	}
}

func BenchmarkProcessInstanceRepository_CountByStatus(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessInstanceRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	// 准备测试数据
	for i := 0; i < 20; i++ {
		instance := createTestProcessInstance(&testing.T{}, db, tenantID)
		instance.Status = model.ProcessStatusPending
		_ = repo.Create(ctx, instance)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.CountByStatus(ctx, tenantID, model.ProcessStatusPending)
	}
}
