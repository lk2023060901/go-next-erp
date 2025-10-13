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

// setupTestDB 创建测试数据库连接（需要实际数据库）
// 注意：这些测试需要运行中的 PostgreSQL 数据库
func setupTestDB(t *testing.T) *database.DB {
	t.Helper()

	ctx := context.Background()

	// 使用环境变量配置的测试数据库
	db, err := database.New(ctx,
		database.WithHost("localhost"),
		database.WithPort(15000),
		database.WithDatabase("erp_test"),
		database.WithUsername("postgres"),
		database.WithPassword("postgres123"),
		database.WithSSLMode("disable"),
	)

	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return nil
	}

	return db
}

// cleanupProcessDefinitions 清理测试数据
func cleanupProcessDefinitions(t *testing.T, db *database.DB, tenantID uuid.UUID) {
	t.Helper()
	_, _ = db.Exec(context.Background(),
		"DELETE FROM approval_process_definitions WHERE tenant_id = $1", tenantID)
}

// TB 是testing.T和testing.B的共同接口
type TB interface {
	Helper()
	Logf(format string, args ...interface{})
}

// createTestFormAndWorkflow 创建测试所需的表单和工作流
func createTestFormAndWorkflow(t TB, db *database.DB, tenantID uuid.UUID) (formID uuid.UUID, workflowID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	// 创建测试表单 - 使用正确的字段名 fields 而不是 schema
	formID = uuid.New()
	_, err := db.Exec(ctx, `
		INSERT INTO form_definitions (id, tenant_id, code, name, fields, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO NOTHING
	`, formID, tenantID, uuid.NewString(), "测试表单", `{"fields":[]}`, uuid.New(), time.Now(), time.Now())

	if err != nil {
		t.Logf("Warning: failed to create test form: %v", err)
	}

	// 工作流ID - 不需要创建实际记录，因为没有workflows表，只需要一个UUID
	workflowID = uuid.New()

	return formID, workflowID
}

// createValidProcessDefinition 创建一个有效的ProcessDefinition用于测试
func createValidProcessDefinition(t *testing.T, db *database.DB, tenantID uuid.UUID, code string) *model.ProcessDefinition {
	t.Helper()
	formID, workflowID := createTestFormAndWorkflow(t, db, tenantID)
	return &model.ProcessDefinition{
		ID:         uuid.New(),
		TenantID:   tenantID,
		Code:       code,
		Name:       "测试流程",
		Category:   "测试分类",
		FormID:     formID,
		WorkflowID: workflowID,
		Enabled:    true,
		CreatedBy:  uuid.New(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func TestProcessDefinitionRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessDefinitions(t, db, tenantID)

	t.Run("Create process definition successfully", func(t *testing.T) {
		// 先创建必需的表单和工作流（或使用已存在的）
		formID, workflowID := createTestFormAndWorkflow(t, db, tenantID)

		def := &model.ProcessDefinition{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Code:        "LEAVE_REQUEST",
			Name:        "请假申请",
			Category:    "人事管理",
			FormID:      formID,
			WorkflowID:  workflowID,
			Enabled:     true,
			CreatedBy:   uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.Create(ctx, def)
		assert.NoError(t, err)

		// 验证是否成功插入
		found, err := repo.FindByID(ctx, def.ID)
		require.NoError(t, err)
		assert.Equal(t, def.Code, found.Code)
		assert.Equal(t, def.Name, found.Name)
		assert.Equal(t, def.Category, found.Category)
		assert.True(t, found.Enabled)
	})

	t.Run("Create with duplicate code should fail", func(t *testing.T) {
		code := "DUPLICATE_TEST"
		// 创建有效的form和workflow
		formID, workflowID := createTestFormAndWorkflow(t, db, tenantID)

		def1 := &model.ProcessDefinition{
			ID:         uuid.New(),
			TenantID:   tenantID,
			Code:       code,
			Name:       "测试流程1",
			Category:   "测试",
			FormID:     formID,
			WorkflowID: workflowID,
			Enabled:    true,
			CreatedBy:  uuid.New(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		err := repo.Create(ctx, def1)
		require.NoError(t, err)

		// 尝试创建相同 code 的流程定义
		def2 := &model.ProcessDefinition{
			ID:         uuid.New(),
			TenantID:   tenantID,
			Code:       code, // 相同的 code
			Name:       "测试流程2",
			Category:   "测试",
			FormID:     formID,
			WorkflowID: workflowID,
			Enabled:    true,
			CreatedBy:  uuid.New(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		err = repo.Create(ctx, def2)
		assert.Error(t, err) // 应该失败（unique constraint）
	})
}

func TestProcessDefinitionRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessDefinitions(t, db, tenantID)

	t.Run("Update process definition successfully", func(t *testing.T) {
		// 先创建
		def := createValidProcessDefinition(t, db, tenantID, "UPDATE_TEST")
		def.Name = "原始名称"
		def.Category = "原始分类"

		err := repo.Create(ctx, def)
		require.NoError(t, err)

		// 更新 - 创建新的form和workflow用于更新
		newFormID, newWorkflowID := createTestFormAndWorkflow(t, db, tenantID)
		updatedBy := uuid.New()

		def.Name = "更新后名称"
		def.FormID = newFormID
		def.WorkflowID = newWorkflowID
		def.Enabled = false
		def.UpdatedBy = &updatedBy
		def.UpdatedAt = time.Now()

		err = repo.Update(ctx, def)
		assert.NoError(t, err)

		// 验证更新
		found, err := repo.FindByID(ctx, def.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新后名称", found.Name)
		assert.Equal(t, newFormID, found.FormID)
		assert.Equal(t, newWorkflowID, found.WorkflowID)
		assert.False(t, found.Enabled)
		assert.NotNil(t, found.UpdatedBy)
		assert.Equal(t, updatedBy, *found.UpdatedBy)
	})

	t.Run("Update non-existent definition", func(t *testing.T) {
		formID, workflowID := createTestFormAndWorkflow(t, db, tenantID)
		def := &model.ProcessDefinition{
			ID:         uuid.New(), // 不存在的ID
			Name:       "不存在的流程",
			FormID:     formID,
			WorkflowID: workflowID,
			UpdatedAt:  time.Now(),
		}

		err := repo.Update(ctx, def)
		// 不会报错，但不会更新任何行
		assert.NoError(t, err)
	})
}

func TestProcessDefinitionRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessDefinitions(t, db, tenantID)

	t.Run("Soft delete process definition", func(t *testing.T) {
		// 创建
		formID, workflowID := createTestFormAndWorkflow(t, db, tenantID)
		def := &model.ProcessDefinition{
			ID:         uuid.New(),
			TenantID:   tenantID,
			Code:       "DELETE_TEST",
			Name:       "待删除流程",
			Category:   "测试",
			FormID:     formID,
			WorkflowID: workflowID,
			Enabled:    true,
			CreatedBy:  uuid.New(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		err := repo.Create(ctx, def)
		require.NoError(t, err)

		// 删除
		err = repo.Delete(ctx, def.ID)
		assert.NoError(t, err)

		// 验证软删除：FindByID 应该找不到
		_, err = repo.FindByID(ctx, def.ID)
		assert.Error(t, err)
		assert.Equal(t, pgx.ErrNoRows, err)

		// 验证数据库中仍存在（deleted_at不为空）
		var deletedAt *time.Time
		err = db.QueryRow(ctx,
			"SELECT deleted_at FROM approval_process_definitions WHERE id = $1",
			def.ID).Scan(&deletedAt)
		assert.NoError(t, err)
		assert.NotNil(t, deletedAt)
	})
}

func TestProcessDefinitionRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessDefinitions(t, db, tenantID)

	t.Run("Find existing definition", func(t *testing.T) {
		formID, workflowID := createTestFormAndWorkflow(t, db, tenantID)
		def := &model.ProcessDefinition{
			ID:         uuid.New(),
			TenantID:   tenantID,
			Code:       "FIND_TEST",
			Name:       "查询测试",
			Category:   "测试",
			FormID:     formID,
			WorkflowID: workflowID,
			Enabled:    true,
			CreatedBy:  uuid.New(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		err := repo.Create(ctx, def)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, def.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, def.ID, found.ID)
		assert.Equal(t, def.Code, found.Code)
		assert.Equal(t, def.Name, found.Name)
	})

	t.Run("Find non-existent definition", func(t *testing.T) {
		_, err := repo.FindByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Equal(t, pgx.ErrNoRows, err)
	})
}

func TestProcessDefinitionRepository_FindByCode(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessDefinitions(t, db, tenantID)

	t.Run("Find by code successfully", func(t *testing.T) {
		code := "FIND_BY_CODE_TEST"
		formID, workflowID := createTestFormAndWorkflow(t, db, tenantID)

		def := &model.ProcessDefinition{
			ID:         uuid.New(),
			TenantID:   tenantID,
			Code:       code,
			Name:       "按编码查询测试",
			Category:   "测试",
			FormID:     formID,
			WorkflowID: workflowID,
			Enabled:    true,
			CreatedBy:  uuid.New(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		err := repo.Create(ctx, def)
		require.NoError(t, err)

		found, err := repo.FindByCode(ctx, tenantID, code)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, code, found.Code)
		assert.Equal(t, tenantID, found.TenantID)
	})

	t.Run("Find by non-existent code", func(t *testing.T) {
		_, err := repo.FindByCode(ctx, tenantID, "NON_EXISTENT_CODE")
		assert.Error(t, err)
		assert.Equal(t, pgx.ErrNoRows, err)
	})

	t.Run("Find by code in different tenant", func(t *testing.T) {
		// 在当前租户创建流程
		code := "TENANT_TEST"
		formID, workflowID := createTestFormAndWorkflow(t, db, tenantID)
		def := &model.ProcessDefinition{
			ID:         uuid.New(),
			TenantID:   tenantID,
			Code:       code,
			Name:       "租户隔离测试",
			Category:   "测试",
			FormID:     formID,
			WorkflowID: workflowID,
			Enabled:    true,
			CreatedBy:  uuid.New(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		err := repo.Create(ctx, def)
		require.NoError(t, err)

		// 使用不同租户ID查询
		differentTenantID := uuid.New()
		_, err = repo.FindByCode(ctx, differentTenantID, code)
		assert.Error(t, err)
		assert.Equal(t, pgx.ErrNoRows, err)
	})
}

func TestProcessDefinitionRepository_List(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessDefinitions(t, db, tenantID)

	t.Run("List all definitions for tenant", func(t *testing.T) {
		// 创建多个流程定义
		formID, workflowID := createTestFormAndWorkflow(t, db, tenantID)
		for i := 1; i <= 3; i++ {
			def := &model.ProcessDefinition{
				ID:         uuid.New(),
				TenantID:   tenantID,
				Code:       uuid.NewString(),
				Name:       uuid.NewString(),
				Category:   "测试",
				FormID:     formID,
				WorkflowID: workflowID,
				Enabled:    i%2 == 0, // 交替启用/禁用
				CreatedBy:  uuid.New(),
				CreatedAt:  time.Now().Add(time.Duration(i) * time.Second),
				UpdatedAt:  time.Now(),
			}

			err := repo.Create(ctx, def)
			require.NoError(t, err)
		}

		// 查询所有
		defs, err := repo.List(ctx, tenantID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(defs), 3)

		// 验证排序（按创建时间降序）
		for i := 0; i < len(defs)-1; i++ {
			assert.True(t, defs[i].CreatedAt.After(defs[i+1].CreatedAt) ||
				defs[i].CreatedAt.Equal(defs[i+1].CreatedAt))
		}
	})

	t.Run("List empty result for new tenant", func(t *testing.T) {
		newTenantID := uuid.New()
		defs, err := repo.List(ctx, newTenantID)
		assert.NoError(t, err)
		assert.Empty(t, defs)
	})
}

func TestProcessDefinitionRepository_ListEnabled(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupProcessDefinitions(t, db, tenantID)

	t.Run("List only enabled definitions", func(t *testing.T) {
		// 创建启用和禁用的流程定义
		formID, workflowID := createTestFormAndWorkflow(t, db, tenantID)
		enabledCount := 0
		for i := 1; i <= 5; i++ {
			enabled := i%2 == 0
			if enabled {
				enabledCount++
			}

			def := &model.ProcessDefinition{
				ID:         uuid.New(),
				TenantID:   tenantID,
				Code:       uuid.NewString(),
				Name:       uuid.NewString(),
				Category:   "测试",
				FormID:     formID,
				WorkflowID: workflowID,
				Enabled:    enabled,
				CreatedBy:  uuid.New(),
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			err := repo.Create(ctx, def)
			require.NoError(t, err)
		}

		// 仅查询启用的
		defs, err := repo.ListEnabled(ctx, tenantID)
		assert.NoError(t, err)
		assert.Equal(t, enabledCount, len(defs))

		// 验证所有返回的都是启用状态
		for _, def := range defs {
			assert.True(t, def.Enabled)
		}
	})
}

// ===== Benchmark Tests =====

func BenchmarkProcessDefinitionRepository_Create(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	formID, workflowID := createTestFormAndWorkflow(b, db, tenantID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		def := &model.ProcessDefinition{
			ID:         uuid.New(),
			TenantID:   tenantID,
			Code:       uuid.NewString(),
			Name:       "基准测试流程",
			Category:   "测试",
			FormID:     formID,
			WorkflowID: workflowID,
			Enabled:    true,
			CreatedBy:  uuid.New(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		_ = repo.Create(ctx, def)
	}
}

func BenchmarkProcessDefinitionRepository_FindByID(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	// 准备测试数据
	formID, workflowID := createTestFormAndWorkflow(b, db, tenantID)
	def := &model.ProcessDefinition{
		ID:         uuid.New(),
		TenantID:   tenantID,
		Code:       "BENCH_FIND",
		Name:       "查询基准测试",
		Category:   "测试",
		FormID:     formID,
		WorkflowID: workflowID,
		Enabled:    true,
		CreatedBy:  uuid.New(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_ = repo.Create(ctx, def)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByID(ctx, def.ID)
	}
}

func BenchmarkProcessDefinitionRepository_List(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewProcessDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	// 准备10条测试数据
	formID, workflowID := createTestFormAndWorkflow(b, db, tenantID)
	for i := 0; i < 10; i++ {
		def := &model.ProcessDefinition{
			ID:         uuid.New(),
			TenantID:   tenantID,
			Code:       uuid.NewString(),
			Name:       "基准测试流程",
			Category:   "测试",
			FormID:     formID,
			WorkflowID: workflowID,
			Enabled:    true,
			CreatedBy:  uuid.New(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		_ = repo.Create(ctx, def)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.List(ctx, tenantID)
	}
}

// setupBenchDB 基准测试数据库设置
func setupBenchDB(b *testing.B) *database.DB {
	b.Helper()

	ctx := context.Background()

	db, err := database.New(ctx,
		database.WithHost("localhost"),
		database.WithPort(15000),
		database.WithDatabase("erp_test"),
		database.WithUsername("postgres"),
		database.WithPassword("postgres123"),
		database.WithSSLMode("disable"),
	)

	if err != nil {
		b.Skipf("Skipping benchmark: database not available: %v", err)
		return nil
	}

	return db
}
