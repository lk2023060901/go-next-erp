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

// Helper function to create test form data
func createTestFormData(t *testing.T, db *database.DB, tenantID uuid.UUID) *model.FormData {
	t.Helper()
	ctx := context.Background()

	// Create a form definition first
	formDef := createTestFormDefinition(t, tenantID)
	formDefRepo := NewFormDefinitionRepository(db)
	err := formDefRepo.Create(ctx, formDef)
	require.NoError(t, err)

	return &model.FormData{
		ID:       uuid.New(),
		TenantID: tenantID,
		FormID:   formDef.ID,
		Data: map[string]interface{}{
			"name": "张三",
			"age":  30,
		},
		SubmittedBy: uuid.New(),
		SubmittedAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Cleanup helper
func cleanupFormData(t *testing.T, db *database.DB, tenantID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	_, _ = db.Exec(ctx, "DELETE FROM form_data WHERE tenant_id = $1", tenantID)
	_, _ = db.Exec(ctx, "DELETE FROM form_definitions WHERE tenant_id = $1", tenantID)
}

// TestFormDataRepository_Create tests form data creation
func TestFormDataRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewFormDataRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupFormData(t, db, tenantID)

	t.Run("Create successfully", func(t *testing.T) {
		formData := createTestFormData(t, db, tenantID)

		err := repo.Create(ctx, formData)
		assert.NoError(t, err)

		// Verify
		found, err := repo.FindByID(ctx, formData.ID)
		require.NoError(t, err)
		assert.Equal(t, formData.ID, found.ID)
		assert.Equal(t, formData.FormID, found.FormID)
		assert.Equal(t, formData.SubmittedBy, found.SubmittedBy)
		assert.Equal(t, "张三", found.Data["name"])
		assert.Equal(t, float64(30), found.Data["age"])
	})

	t.Run("Create with simple data", func(t *testing.T) {
		formData := createTestFormData(t, db, tenantID)
		formData.Data = map[string]interface{}{
			"field1": "value1",
			"field2": 123,
		}

		err := repo.Create(ctx, formData)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, formData.ID)
		require.NoError(t, err)
		assert.Equal(t, "value1", found.Data["field1"])
		assert.Equal(t, float64(123), found.Data["field2"])
	})

	t.Run("Create with complex nested data", func(t *testing.T) {
		formData := createTestFormData(t, db, tenantID)
		formData.Data = map[string]interface{}{
			"name": "李四",
			"address": map[string]interface{}{
				"province": "广东省",
				"city":     "深圳市",
				"district": "南山区",
			},
			"skills": []string{"Go", "Python", "JavaScript"},
			"metadata": map[string]interface{}{
				"tags": []string{"developer", "backend"},
				"score": float64(95.5),
			},
		}

		err := repo.Create(ctx, formData)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, formData.ID)
		require.NoError(t, err)
		assert.NotNil(t, found.Data["address"])
		assert.NotNil(t, found.Data["skills"])
		assert.NotNil(t, found.Data["metadata"])
	})

	t.Run("Create with empty data", func(t *testing.T) {
		formData := createTestFormData(t, db, tenantID)
		formData.Data = map[string]interface{}{}

		err := repo.Create(ctx, formData)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, formData.ID)
		require.NoError(t, err)
		assert.Empty(t, found.Data)
	})

	t.Run("Create with related type and ID", func(t *testing.T) {
		formData := createTestFormData(t, db, tenantID)
		relatedType := "approval_process"
		relatedID := uuid.New()
		formData.RelatedType = &relatedType
		formData.RelatedID = &relatedID

		err := repo.Create(ctx, formData)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, formData.ID)
		require.NoError(t, err)
		assert.NotNil(t, found.RelatedType)
		assert.Equal(t, "approval_process", *found.RelatedType)
		assert.NotNil(t, found.RelatedID)
		assert.Equal(t, relatedID, *found.RelatedID)
	})

	t.Run("Create without related type and ID", func(t *testing.T) {
		formData := createTestFormData(t, db, tenantID)
		formData.RelatedType = nil
		formData.RelatedID = nil

		err := repo.Create(ctx, formData)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, formData.ID)
		require.NoError(t, err)
		assert.Nil(t, found.RelatedType)
		assert.Nil(t, found.RelatedID)
	})
}

// TestFormDataRepository_Update tests form data updates
func TestFormDataRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewFormDataRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupFormData(t, db, tenantID)

	t.Run("Update successfully", func(t *testing.T) {
		formData := createTestFormData(t, db, tenantID)
		err := repo.Create(ctx, formData)
		require.NoError(t, err)

		// Update
		formData.Data = map[string]interface{}{
			"name":  "王五",
			"age":   35,
			"email": "wangwu@example.com",
		}
		formData.UpdatedAt = time.Now()

		err = repo.Update(ctx, formData)
		assert.NoError(t, err)

		// Verify
		found, err := repo.FindByID(ctx, formData.ID)
		require.NoError(t, err)
		assert.Equal(t, "王五", found.Data["name"])
		assert.Equal(t, float64(35), found.Data["age"])
		assert.Equal(t, "wangwu@example.com", found.Data["email"])
	})

	t.Run("Update with complex data", func(t *testing.T) {
		formData := createTestFormData(t, db, tenantID)
		err := repo.Create(ctx, formData)
		require.NoError(t, err)

		// Update with nested data
		formData.Data = map[string]interface{}{
			"contact": map[string]interface{}{
				"phone":  "13800138000",
				"wechat": "wechat123",
			},
			"preferences": []string{"option1", "option2"},
		}
		formData.UpdatedAt = time.Now()

		err = repo.Update(ctx, formData)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, formData.ID)
		require.NoError(t, err)
		assert.NotNil(t, found.Data["contact"])
		assert.NotNil(t, found.Data["preferences"])
	})

	t.Run("Update non-existent form data", func(t *testing.T) {
		formData := createTestFormData(t, db, tenantID)
		formData.ID = uuid.New() // Non-existent ID

		err := repo.Update(ctx, formData)
		assert.NoError(t, err) // No error, just no rows affected
	})
}

// TestFormDataRepository_FindByID tests finding form data by ID
func TestFormDataRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewFormDataRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupFormData(t, db, tenantID)

	t.Run("Find existing form data", func(t *testing.T) {
		formData := createTestFormData(t, db, tenantID)
		err := repo.Create(ctx, formData)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, formData.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, formData.ID, found.ID)
	})

	t.Run("Find non-existent form data", func(t *testing.T) {
		found, err := repo.FindByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}

// TestFormDataRepository_FindByRelated tests finding form data by related entity
func TestFormDataRepository_FindByRelated(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewFormDataRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupFormData(t, db, tenantID)

	t.Run("Find by related type and ID", func(t *testing.T) {
		formData := createTestFormData(t, db, tenantID)
		relatedType := "approval_process"
		relatedID := uuid.New()
		formData.RelatedType = &relatedType
		formData.RelatedID = &relatedID

		err := repo.Create(ctx, formData)
		require.NoError(t, err)

		// Find by related
		found, err := repo.FindByRelated(ctx, relatedType, relatedID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, formData.ID, found.ID)
		assert.Equal(t, relatedID, *found.RelatedID)
	})

	t.Run("Find by non-existent related", func(t *testing.T) {
		found, err := repo.FindByRelated(ctx, "non_existent_type", uuid.New())
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("Different related types are isolated", func(t *testing.T) {
		relatedID := uuid.New()

		// Create form data with type1
		formData1 := createTestFormData(t, db, tenantID)
		type1 := "approval_process"
		formData1.RelatedType = &type1
		formData1.RelatedID = &relatedID
		err := repo.Create(ctx, formData1)
		require.NoError(t, err)

		// Create form data with type2 but same ID
		formData2 := createTestFormData(t, db, tenantID)
		type2 := "workflow_instance"
		formData2.RelatedType = &type2
		formData2.RelatedID = &relatedID
		err = repo.Create(ctx, formData2)
		require.NoError(t, err)

		// Find by type1 should only return formData1
		found, err := repo.FindByRelated(ctx, "approval_process", relatedID)
		assert.NoError(t, err)
		assert.Equal(t, formData1.ID, found.ID)

		// Find by type2 should only return formData2
		found, err = repo.FindByRelated(ctx, "workflow_instance", relatedID)
		assert.NoError(t, err)
		assert.Equal(t, formData2.ID, found.ID)
	})
}

// TestFormDataRepository_ListByForm tests listing form data by form definition
func TestFormDataRepository_ListByForm(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewFormDataRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupFormData(t, db, tenantID)

	t.Run("List form data by form ID", func(t *testing.T) {
		// Create form data for same form
		formData1 := createTestFormData(t, db, tenantID)
		formID := formData1.FormID
		formData1.SubmittedAt = time.Now().Add(-2 * time.Hour)
		err := repo.Create(ctx, formData1)
		require.NoError(t, err)

		formData2 := createTestFormData(t, db, tenantID)
		formData2.FormID = formID
		formData2.SubmittedAt = time.Now().Add(-1 * time.Hour)
		err = repo.Create(ctx, formData2)
		require.NoError(t, err)

		formData3 := createTestFormData(t, db, tenantID)
		formData3.FormID = formID
		formData3.SubmittedAt = time.Now()
		err = repo.Create(ctx, formData3)
		require.NoError(t, err)

		// List
		list, err := repo.ListByForm(ctx, formID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 3)

		// Verify ordering (by submitted_at DESC)
		for i := 0; i < len(list)-1; i++ {
			assert.True(t, list[i].SubmittedAt.After(list[i+1].SubmittedAt) ||
				list[i].SubmittedAt.Equal(list[i+1].SubmittedAt))
		}
	})

	t.Run("List empty for non-existent form", func(t *testing.T) {
		list, err := repo.ListByForm(ctx, uuid.New())
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("List isolates different forms", func(t *testing.T) {
		// Create data for form1
		formData1 := createTestFormData(t, db, tenantID)
		err := repo.Create(ctx, formData1)
		require.NoError(t, err)

		// Create data for form2
		formData2 := createTestFormData(t, db, tenantID)
		err = repo.Create(ctx, formData2)
		require.NoError(t, err)

		// List for form1 should not include form2 data
		list1, err := repo.ListByForm(ctx, formData1.FormID)
		assert.NoError(t, err)

		for _, data := range list1 {
			assert.Equal(t, formData1.FormID, data.FormID)
		}
	})
}

// TestFormDataRepository_ListBySubmitter tests listing form data by submitter
func TestFormDataRepository_ListBySubmitter(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewFormDataRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	defer cleanupFormData(t, db, tenantID)

	t.Run("List form data by submitter", func(t *testing.T) {
		submitterID := uuid.New()

		// Create multiple submissions by same submitter
		for i := 0; i < 3; i++ {
			formData := createTestFormData(t, db, tenantID)
			formData.SubmittedBy = submitterID
			formData.SubmittedAt = time.Now().Add(time.Duration(-i) * time.Hour)
			err := repo.Create(ctx, formData)
			require.NoError(t, err)
		}

		// List
		list, err := repo.ListBySubmitter(ctx, submitterID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 3)

		// Verify all belong to submitter
		for _, data := range list {
			assert.Equal(t, submitterID, data.SubmittedBy)
		}

		// Verify ordering (by submitted_at DESC)
		for i := 0; i < len(list)-1; i++ {
			assert.True(t, list[i].SubmittedAt.After(list[i+1].SubmittedAt) ||
				list[i].SubmittedAt.Equal(list[i+1].SubmittedAt))
		}
	})

	t.Run("List empty for submitter with no submissions", func(t *testing.T) {
		list, err := repo.ListBySubmitter(ctx, uuid.New())
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("List isolates different submitters", func(t *testing.T) {
		submitter1 := uuid.New()
		submitter2 := uuid.New()

		// Create for submitter1
		formData1 := createTestFormData(t, db, tenantID)
		formData1.SubmittedBy = submitter1
		err := repo.Create(ctx, formData1)
		require.NoError(t, err)

		// Create for submitter2
		formData2 := createTestFormData(t, db, tenantID)
		formData2.SubmittedBy = submitter2
		err = repo.Create(ctx, formData2)
		require.NoError(t, err)

		// List for submitter1 should not include submitter2 data
		list1, err := repo.ListBySubmitter(ctx, submitter1)
		assert.NoError(t, err)

		for _, data := range list1 {
			assert.Equal(t, submitter1, data.SubmittedBy)
		}
	})
}

// TestFormDataRepository_TenantIsolation tests tenant isolation
func TestFormDataRepository_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewFormDataRepository(db)
	ctx := context.Background()
	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	defer cleanupFormData(t, db, tenantID1)
	defer cleanupFormData(t, db, tenantID2)

	t.Run("Form data isolated by tenant", func(t *testing.T) {
		// Create data for tenant1
		formData1 := createTestFormData(t, db, tenantID1)
		err := repo.Create(ctx, formData1)
		require.NoError(t, err)

		// Create data for tenant2
		formData2 := createTestFormData(t, db, tenantID2)
		err = repo.Create(ctx, formData2)
		require.NoError(t, err)

		// Verify they can be found by their own IDs
		found1, err := repo.FindByID(ctx, formData1.ID)
		assert.NoError(t, err)
		assert.Equal(t, tenantID1, found1.TenantID)

		found2, err := repo.FindByID(ctx, formData2.ID)
		assert.NoError(t, err)
		assert.Equal(t, tenantID2, found2.TenantID)
	})
}

// ======================== Benchmark Tests ========================

// BenchmarkFormDataRepository_Create benchmarks form data creation
func BenchmarkFormDataRepository_Create(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewFormDataRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formData := createTestFormData(&testing.T{}, db, tenantID)
		_ = repo.Create(ctx, formData)
	}
}

// BenchmarkFormDataRepository_FindByID benchmarks finding by ID
func BenchmarkFormDataRepository_FindByID(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewFormDataRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	formData := createTestFormData(&testing.T{}, db, tenantID)
	_ = repo.Create(ctx, formData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByID(ctx, formData.ID)
	}
}

// BenchmarkFormDataRepository_ListByForm benchmarks listing by form
func BenchmarkFormDataRepository_ListByForm(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewFormDataRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()

	formData := createTestFormData(&testing.T{}, db, tenantID)
	formID := formData.FormID

	// Create 10 form data entries
	for i := 0; i < 10; i++ {
		data := createTestFormData(&testing.T{}, db, tenantID)
		data.FormID = formID
		_ = repo.Create(ctx, data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.ListByForm(ctx, formID)
	}
}

// BenchmarkFormDataRepository_ListBySubmitter benchmarks listing by submitter
func BenchmarkFormDataRepository_ListBySubmitter(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available")
		return
	}
	defer db.Close()

	repo := NewFormDataRepository(db)
	ctx := context.Background()
	tenantID := uuid.New()
	submitterID := uuid.New()

	// Create 10 form data entries
	for i := 0; i < 10; i++ {
		data := createTestFormData(&testing.T{}, db, tenantID)
		data.SubmittedBy = submitterID
		_ = repo.Create(ctx, data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.ListBySubmitter(ctx, submitterID)
	}
}
