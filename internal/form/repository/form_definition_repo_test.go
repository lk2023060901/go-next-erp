package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/form/model"
	"github.com/lk2023060901/go-next-erp/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
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
		t.Skipf("Skipping test: database not available: %v", err)
		return nil
	}

	return db
}

// Helper function to create test form definition
func createTestFormDefinition(t *testing.T, tenantID uuid.UUID) *model.FormDefinition {
	t.Helper()

	return &model.FormDefinition{
		ID:       uuid.New(),
		TenantID: tenantID,
		Code:     "TEST_FORM_" + uuid.NewString()[:8],
		Name:     "测试表单",
		Fields: []model.FormField{
			{
				Key:      "name",
				Label:    "姓名",
				Type:     model.FieldTypeText,
				Required: true,
				Sort:     1,
			},
			{
				Key:      "age",
				Label:    "年龄",
				Type:     model.FieldTypeNumber,
				Required: false,
				Sort:     2,
			},
		},
		Enabled:   true,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Cleanup helper
func cleanupFormDefinitions(t *testing.T, db *database.DB, tenantID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	_, _ = db.Exec(ctx, "DELETE FROM form_definitions WHERE tenant_id = $1", tenantID)
}

// TestFormDefinitionRepository_Create tests form definition creation
func TestFormDefinitionRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewFormDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupFormDefinitions(t, db, tenantID)

	t.Run("Create successfully", func(t *testing.T) {
		form := createTestFormDefinition(t, tenantID)

		err := repo.Create(ctx, form)
		assert.NoError(t, err)

		// Verify
		found, err := repo.FindByID(ctx, form.ID)
		require.NoError(t, err)
		assert.Equal(t, form.ID, found.ID)
		assert.Equal(t, form.Code, found.Code)
		assert.Equal(t, form.Name, found.Name)
		assert.Len(t, found.Fields, 2)
		assert.Equal(t, form.Enabled, found.Enabled)
	})

	t.Run("Create with complex fields", func(t *testing.T) {
		form := createTestFormDefinition(t, tenantID)
		form.Fields = []model.FormField{
			{
				Key:      "department",
				Label:    "部门",
				Type:     model.FieldTypeDept,
				Required: true,
				Sort:     1,
			},
			{
				Key:   "gender",
				Label: "性别",
				Type:  model.FieldTypeSelect,
				Options: []model.FieldOption{
					{Label: "男", Value: "male"},
					{Label: "女", Value: "female"},
				},
				Required: true,
				Sort:     2,
			},
			{
				Key:   "hobbies",
				Label: "爱好",
				Type:  model.FieldTypeCheckbox,
				Options: []model.FieldOption{
					{Label: "阅读", Value: "reading"},
					{Label: "运动", Value: "sports"},
					{Label: "音乐", Value: "music"},
				},
				Required: false,
				Sort:     3,
			},
		}

		err := repo.Create(ctx, form)
		assert.NoError(t, err)

		// Verify complex fields
		found, err := repo.FindByID(ctx, form.ID)
		require.NoError(t, err)
		assert.Len(t, found.Fields, 3)
		assert.Equal(t, model.FieldTypeDept, found.Fields[0].Type)
		assert.Len(t, found.Fields[1].Options, 2)
		assert.Len(t, found.Fields[2].Options, 3)
	})

	t.Run("Create with validation rules", func(t *testing.T) {
		form := createTestFormDefinition(t, tenantID)
		form.Fields = []model.FormField{
			{
				Key:      "email",
				Label:    "邮箱",
				Type:     model.FieldTypeText,
				Required: true,
				Rules: []model.ValidationRule{
					{
						Type:    "pattern",
						Value:   "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
						Message: "请输入有效的邮箱地址",
					},
				},
				Sort: 1,
			},
			{
				Key:      "age",
				Label:    "年龄",
				Type:     model.FieldTypeNumber,
				Required: true,
				Rules: []model.ValidationRule{
					{Type: "min", Value: float64(18), Message: "年龄不能小于18"},
					{Type: "max", Value: float64(65), Message: "年龄不能大于65"},
				},
				Sort: 2,
			},
		}

		err := repo.Create(ctx, form)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, form.ID)
		require.NoError(t, err)
		assert.Len(t, found.Fields[0].Rules, 1)
		assert.Len(t, found.Fields[1].Rules, 2)
	})

	t.Run("Create disabled form", func(t *testing.T) {
		form := createTestFormDefinition(t, tenantID)
		form.Enabled = false

		err := repo.Create(ctx, form)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, form.ID)
		require.NoError(t, err)
		assert.False(t, found.Enabled)
	})
}

// TestFormDefinitionRepository_Update tests form definition updates
func TestFormDefinitionRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewFormDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupFormDefinitions(t, db, tenantID)

	t.Run("Update successfully", func(t *testing.T) {
		form := createTestFormDefinition(t, tenantID)
		err := repo.Create(ctx, form)
		require.NoError(t, err)

		// Update
		form.Name = "更新后的表单"
		form.Enabled = false
		form.Fields = append(form.Fields, model.FormField{
			Key:      "email",
			Label:    "邮箱",
			Type:     model.FieldTypeText,
			Required: true,
			Sort:     3,
		})
		updatedBy := uuid.New()
		form.UpdatedBy = &updatedBy
		form.UpdatedAt = time.Now()

		err = repo.Update(ctx, form)
		assert.NoError(t, err)

		// Verify
		found, err := repo.FindByID(ctx, form.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新后的表单", found.Name)
		assert.False(t, found.Enabled)
		assert.Len(t, found.Fields, 3)
		assert.NotNil(t, found.UpdatedBy)
	})

	t.Run("Update non-existent form", func(t *testing.T) {
		form := createTestFormDefinition(t, tenantID)
		form.ID = uuid.New() // Non-existent ID

		err := repo.Update(ctx, form)
		assert.NoError(t, err) // No error, just no rows affected
	})

	t.Run("Update fields only", func(t *testing.T) {
		form := createTestFormDefinition(t, tenantID)
		err := repo.Create(ctx, form)
		require.NoError(t, err)

		originalName := form.Name
		form.Fields = []model.FormField{
			{
				Key:      "new_field",
				Label:    "新字段",
				Type:     model.FieldTypeText,
				Required: false,
				Sort:     1,
			},
		}
		form.UpdatedAt = time.Now()

		err = repo.Update(ctx, form)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, form.ID)
		require.NoError(t, err)
		assert.Equal(t, originalName, found.Name) // Name unchanged
		assert.Len(t, found.Fields, 1)           // Fields updated
	})
}

// TestFormDefinitionRepository_Delete tests form definition soft deletion
func TestFormDefinitionRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewFormDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupFormDefinitions(t, db, tenantID)

	t.Run("Delete successfully", func(t *testing.T) {
		form := createTestFormDefinition(t, tenantID)
		err := repo.Create(ctx, form)
		require.NoError(t, err)

		// Delete
		err = repo.Delete(ctx, form.ID)
		assert.NoError(t, err)

		// Verify soft delete (should not be found)
		found, err := repo.FindByID(ctx, form.ID)
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("Delete non-existent form", func(t *testing.T) {
		err := repo.Delete(ctx, uuid.New())
		assert.NoError(t, err) // No error, just no rows affected
	})

	t.Run("Delete does not appear in list", func(t *testing.T) {
		form := createTestFormDefinition(t, tenantID)
		err := repo.Create(ctx, form)
		require.NoError(t, err)

		// List before delete
		forms, err := repo.List(ctx, tenantID)
		require.NoError(t, err)
		initialCount := len(forms)

		// Delete
		err = repo.Delete(ctx, form.ID)
		require.NoError(t, err)

		// List after delete
		forms, err = repo.List(ctx, tenantID)
		assert.NoError(t, err)
		assert.Equal(t, initialCount-1, len(forms))
	})
}

// TestFormDefinitionRepository_FindByID tests finding form by ID
func TestFormDefinitionRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewFormDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupFormDefinitions(t, db, tenantID)

	t.Run("Find existing form", func(t *testing.T) {
		form := createTestFormDefinition(t, tenantID)
		err := repo.Create(ctx, form)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, form.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, form.ID, found.ID)
		assert.Equal(t, form.Code, found.Code)
	})

	t.Run("Find non-existent form", func(t *testing.T) {
		found, err := repo.FindByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("Find deleted form should fail", func(t *testing.T) {
		form := createTestFormDefinition(t, tenantID)
		err := repo.Create(ctx, form)
		require.NoError(t, err)

		// Delete
		err = repo.Delete(ctx, form.ID)
		require.NoError(t, err)

		// Try to find
		found, err := repo.FindByID(ctx, form.ID)
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}

// TestFormDefinitionRepository_FindByCode tests finding form by code
func TestFormDefinitionRepository_FindByCode(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewFormDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupFormDefinitions(t, db, tenantID)

	t.Run("Find by code successfully", func(t *testing.T) {
		form := createTestFormDefinition(t, tenantID)
		err := repo.Create(ctx, form)
		require.NoError(t, err)

		found, err := repo.FindByCode(ctx, tenantID, form.Code)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, form.ID, found.ID)
		assert.Equal(t, form.Code, found.Code)
	})

	t.Run("Find by non-existent code", func(t *testing.T) {
		found, err := repo.FindByCode(ctx, tenantID, "NON_EXISTENT_CODE")
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("Tenant isolation", func(t *testing.T) {
		tenantID1 := uuid.New()
		tenantID2 := uuid.New()
		defer cleanupFormDefinitions(t, db, tenantID1)
		defer cleanupFormDefinitions(t, db, tenantID2)

		form := createTestFormDefinition(t, tenantID1)
		err := repo.Create(ctx, form)
		require.NoError(t, err)

		// Try to find with different tenant
		found, err := repo.FindByCode(ctx, tenantID2, form.Code)
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}

// TestFormDefinitionRepository_List tests listing forms
func TestFormDefinitionRepository_List(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewFormDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupFormDefinitions(t, db, tenantID)

	t.Run("List forms", func(t *testing.T) {
		// Create multiple forms
		for i := 0; i < 3; i++ {
			form := createTestFormDefinition(t, tenantID)
			err := repo.Create(ctx, form)
			require.NoError(t, err)
		}

		forms, err := repo.List(ctx, tenantID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(forms), 3)
	})

	t.Run("List empty", func(t *testing.T) {
		emptyTenant := uuid.New()
		forms, err := repo.List(ctx, emptyTenant)
		assert.NoError(t, err)
		assert.Empty(t, forms)
	})

	t.Run("List ordered by created_at DESC", func(t *testing.T) {
		tenantID := uuid.New()
		defer cleanupFormDefinitions(t, db, tenantID)

		// Create forms with different timestamps
		form1 := createTestFormDefinition(t, tenantID)
		form1.CreatedAt = time.Now().Add(-2 * time.Hour)
		err := repo.Create(ctx, form1)
		require.NoError(t, err)

		form2 := createTestFormDefinition(t, tenantID)
		form2.CreatedAt = time.Now().Add(-1 * time.Hour)
		err = repo.Create(ctx, form2)
		require.NoError(t, err)

		form3 := createTestFormDefinition(t, tenantID)
		form3.CreatedAt = time.Now()
		err = repo.Create(ctx, form3)
		require.NoError(t, err)

		forms, err := repo.List(ctx, tenantID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(forms), 3)

		// Verify order (newest first)
		for i := 0; i < len(forms)-1; i++ {
			assert.True(t, forms[i].CreatedAt.After(forms[i+1].CreatedAt) ||
				forms[i].CreatedAt.Equal(forms[i+1].CreatedAt))
		}
	})

	t.Run("List excludes deleted forms", func(t *testing.T) {
		tenantID := uuid.New()
		defer cleanupFormDefinitions(t, db, tenantID)

		form1 := createTestFormDefinition(t, tenantID)
		err := repo.Create(ctx, form1)
		require.NoError(t, err)

		form2 := createTestFormDefinition(t, tenantID)
		err = repo.Create(ctx, form2)
		require.NoError(t, err)

		// Delete one
		err = repo.Delete(ctx, form1.ID)
		require.NoError(t, err)

		// List should only show non-deleted
		forms, err := repo.List(ctx, tenantID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(forms))
		assert.Equal(t, form2.ID, forms[0].ID)
	})
}

// TestFormDefinitionRepository_ListEnabled tests listing enabled forms
func TestFormDefinitionRepository_ListEnabled(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewFormDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupFormDefinitions(t, db, tenantID)

	t.Run("List only enabled forms", func(t *testing.T) {
		// Create enabled form
		form1 := createTestFormDefinition(t, tenantID)
		form1.Enabled = true
		err := repo.Create(ctx, form1)
		require.NoError(t, err)

		// Create disabled form
		form2 := createTestFormDefinition(t, tenantID)
		form2.Enabled = false
		err = repo.Create(ctx, form2)
		require.NoError(t, err)

		// List enabled only
		forms, err := repo.ListEnabled(ctx, tenantID)
		assert.NoError(t, err)

		// Verify all are enabled
		for _, form := range forms {
			assert.True(t, form.Enabled)
		}
	})

	t.Run("List enabled excludes deleted", func(t *testing.T) {
		tenantID := uuid.New()
		defer cleanupFormDefinitions(t, db, tenantID)

		form := createTestFormDefinition(t, tenantID)
		form.Enabled = true
		err := repo.Create(ctx, form)
		require.NoError(t, err)

		// Delete
		err = repo.Delete(ctx, form.ID)
		require.NoError(t, err)

		// Should not appear in enabled list
		forms, err := repo.ListEnabled(ctx, tenantID)
		assert.NoError(t, err)

		for _, f := range forms {
			assert.NotEqual(t, form.ID, f.ID)
		}
	})
}

// ======================== Benchmark Tests ========================

// BenchmarkFormDefinitionRepository_Create benchmarks form creation
func BenchmarkFormDefinitionRepository_Create(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewFormDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		form := createTestFormDefinition(&testing.T{}, tenantID)
		_ = repo.Create(ctx, form)
	}
}

// BenchmarkFormDefinitionRepository_FindByID benchmarks finding by ID
func BenchmarkFormDefinitionRepository_FindByID(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewFormDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	form := createTestFormDefinition(&testing.T{}, tenantID)
	_ = repo.Create(ctx, form)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByID(ctx, form.ID)
	}
}

// BenchmarkFormDefinitionRepository_List benchmarks listing forms
func BenchmarkFormDefinitionRepository_List(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewFormDefinitionRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	// Create test forms
	for i := 0; i < 10; i++ {
		form := createTestFormDefinition(&testing.T{}, tenantID)
		_ = repo.Create(ctx, form)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.List(ctx, tenantID)
	}
}
