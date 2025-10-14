package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	formv1 "github.com/lk2023060901/go-next-erp/api/form/v1"
	"github.com/lk2023060901/go-next-erp/internal/form/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFormDefinitionRepository mocks the form definition repository
type MockFormDefinitionRepository struct {
	mock.Mock
}

func (m *MockFormDefinitionRepository) Create(ctx context.Context, formDef *model.FormDefinition) error {
	args := m.Called(ctx, formDef)
	return args.Error(0)
}

func (m *MockFormDefinitionRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.FormDefinition, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FormDefinition), args.Error(1)
}

func (m *MockFormDefinitionRepository) Update(ctx context.Context, formDef *model.FormDefinition) error {
	args := m.Called(ctx, formDef)
	return args.Error(0)
}

func (m *MockFormDefinitionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFormDefinitionRepository) List(ctx context.Context, tenantID uuid.UUID) ([]*model.FormDefinition, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.FormDefinition), args.Error(1)
}

func (m *MockFormDefinitionRepository) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.FormDefinition, error) {
	args := m.Called(ctx, tenantID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FormDefinition), args.Error(1)
}

func (m *MockFormDefinitionRepository) ListEnabled(ctx context.Context, tenantID uuid.UUID) ([]*model.FormDefinition, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.FormDefinition), args.Error(1)
}

// MockFormDataRepository mocks the form data repository
type MockFormDataRepository struct {
	mock.Mock
}

func (m *MockFormDataRepository) Create(ctx context.Context, formData *model.FormData) error {
	args := m.Called(ctx, formData)
	return args.Error(0)
}

func (m *MockFormDataRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.FormData, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FormData), args.Error(1)
}

func (m *MockFormDataRepository) Update(ctx context.Context, formData *model.FormData) error {
	args := m.Called(ctx, formData)
	return args.Error(0)
}

func (m *MockFormDataRepository) ListByForm(ctx context.Context, formID uuid.UUID) ([]*model.FormData, error) {
	args := m.Called(ctx, formID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.FormData), args.Error(1)
}

func (m *MockFormDataRepository) FindByRelated(ctx context.Context, relatedType string, relatedID uuid.UUID) (*model.FormData, error) {
	args := m.Called(ctx, relatedType, relatedID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FormData), args.Error(1)
}

func (m *MockFormDataRepository) ListBySubmitter(ctx context.Context, submitterID uuid.UUID) ([]*model.FormData, error) {
	args := m.Called(ctx, submitterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.FormData), args.Error(1)
}

// TestFormAdapter_CreateFormDefinition tests creating form definitions
func TestFormAdapter_CreateFormDefinition(t *testing.T) {
	t.Run("CreateFormDefinition successfully", func(t *testing.T) {
		mockFormDefRepo := new(MockFormDefinitionRepository)
		mockFormDataRepo := new(MockFormDataRepository)
		adapter := NewFormAdapter(mockFormDefRepo, mockFormDataRepo)

		tenantID := uuid.New()

		mockFormDefRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.FormDefinition")).
			Return(nil).Once()

		req := &formv1.CreateFormDefinitionRequest{
			TenantId: tenantID.String(),
			Code:     "LEAVE_FORM",
			Name:     "请假单",
			Schema:   `[{"name":"days","type":"number","label":"请假天数"}]`,
		}

		resp, err := adapter.CreateFormDefinition(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "LEAVE_FORM", resp.Code)
		assert.Equal(t, "请假单", resp.Name)
		mockFormDefRepo.AssertExpectations(t)
	})
}

// TestFormAdapter_GetFormDefinition tests getting a form definition
func TestFormAdapter_GetFormDefinition(t *testing.T) {
	t.Run("GetFormDefinition successfully", func(t *testing.T) {
		mockFormDefRepo := new(MockFormDefinitionRepository)
		mockFormDataRepo := new(MockFormDataRepository)
		adapter := NewFormAdapter(mockFormDefRepo, mockFormDataRepo)

		formID := uuid.New()
		tenantID := uuid.New()

		expectedFormDef := &model.FormDefinition{
			ID:        formID,
			TenantID:  tenantID,
			Code:      "LEAVE_FORM",
			Name:      "请假单",
			Fields:    []model.FormField{},
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockFormDefRepo.On("FindByID", mock.Anything, formID).
			Return(expectedFormDef, nil).Once()

		req := &formv1.GetFormDefinitionRequest{
			Id: formID.String(),
		}

		resp, err := adapter.GetFormDefinition(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, formID.String(), resp.Id)
		assert.Equal(t, "LEAVE_FORM", resp.Code)
		assert.Equal(t, "请假单", resp.Name)
		mockFormDefRepo.AssertExpectations(t)
	})
}

// TestFormAdapter_UpdateFormDefinition tests updating form definitions
func TestFormAdapter_UpdateFormDefinition(t *testing.T) {
	t.Run("UpdateFormDefinition successfully", func(t *testing.T) {
		mockFormDefRepo := new(MockFormDefinitionRepository)
		mockFormDataRepo := new(MockFormDataRepository)
		adapter := NewFormAdapter(mockFormDefRepo, mockFormDataRepo)

		formID := uuid.New()
		tenantID := uuid.New()

		existingFormDef := &model.FormDefinition{
			ID:        formID,
			TenantID:  tenantID,
			Code:      "LEAVE_FORM",
			Name:      "请假单",
			Fields:    []model.FormField{},
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockFormDefRepo.On("FindByID", mock.Anything, formID).
			Return(existingFormDef, nil).Once()
		mockFormDefRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.FormDefinition")).
			Return(nil).Once()

		req := &formv1.UpdateFormDefinitionRequest{
			Id:     formID.String(),
			Name:   "请假单-更新",
			Schema: `[{"name":"days","type":"number","label":"请假天数"},{"name":"reason","type":"text","label":"请假原因"}]`,
		}

		resp, err := adapter.UpdateFormDefinition(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, formID.String(), resp.Id)
		assert.Equal(t, "请假单-更新", resp.Name)
		mockFormDefRepo.AssertExpectations(t)
	})
}

// TestFormAdapter_DeleteFormDefinition tests deleting form definitions
func TestFormAdapter_DeleteFormDefinition(t *testing.T) {
	t.Run("DeleteFormDefinition successfully", func(t *testing.T) {
		mockFormDefRepo := new(MockFormDefinitionRepository)
		mockFormDataRepo := new(MockFormDataRepository)
		adapter := NewFormAdapter(mockFormDefRepo, mockFormDataRepo)

		formID := uuid.New()

		mockFormDefRepo.On("Delete", mock.Anything, formID).
			Return(nil).Once()

		req := &formv1.DeleteFormDefinitionRequest{
			Id: formID.String(),
		}

		resp, err := adapter.DeleteFormDefinition(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)
		mockFormDefRepo.AssertExpectations(t)
	})
}

// TestFormAdapter_ListFormDefinitions tests listing form definitions
func TestFormAdapter_ListFormDefinitions(t *testing.T) {
	t.Run("ListFormDefinitions successfully", func(t *testing.T) {
		mockFormDefRepo := new(MockFormDefinitionRepository)
		mockFormDataRepo := new(MockFormDataRepository)
		adapter := NewFormAdapter(mockFormDefRepo, mockFormDataRepo)

		tenantID := uuid.New()

		expectedFormDefs := []*model.FormDefinition{
			{
				ID:        uuid.New(),
				TenantID:  tenantID,
				Code:      "LEAVE_FORM",
				Name:      "请假单",
				Fields:    []model.FormField{},
				Enabled:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:        uuid.New(),
				TenantID:  tenantID,
				Code:      "EXPENSE_FORM",
				Name:      "报销单",
				Fields:    []model.FormField{},
				Enabled:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mockFormDefRepo.On("List", mock.Anything, tenantID).
			Return(expectedFormDefs, nil).Once()

		req := &formv1.ListFormDefinitionsRequest{
			TenantId: tenantID.String(),
		}

		resp, err := adapter.ListFormDefinitions(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Items, 2)
		assert.Equal(t, int32(2), resp.Total)
		mockFormDefRepo.AssertExpectations(t)
	})
}

// TestFormAdapter_SubmitFormData tests submitting form data
func TestFormAdapter_SubmitFormData(t *testing.T) {
	t.Run("SubmitFormData successfully", func(t *testing.T) {
		mockFormDefRepo := new(MockFormDefinitionRepository)
		mockFormDataRepo := new(MockFormDataRepository)
		adapter := NewFormAdapter(mockFormDefRepo, mockFormDataRepo)

		formID := uuid.New()
		tenantID := uuid.New()
		userID := uuid.New()

		mockFormDataRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.FormData")).
			Return(nil).Once()

		req := &formv1.SubmitFormDataRequest{
			FormId:   formID.String(),
			TenantId: tenantID.String(),
			UserId:   userID.String(),
			Data:     `{"days":3,"reason":"Personal leave"}`,
		}

		resp, err := adapter.SubmitFormData(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, formID.String(), resp.FormId)
		assert.Equal(t, tenantID.String(), resp.TenantId)
		mockFormDataRepo.AssertExpectations(t)
	})
}

// TestFormAdapter_GetFormData tests getting form data
func TestFormAdapter_GetFormData(t *testing.T) {
	t.Run("GetFormData successfully", func(t *testing.T) {
		mockFormDefRepo := new(MockFormDefinitionRepository)
		mockFormDataRepo := new(MockFormDataRepository)
		adapter := NewFormAdapter(mockFormDefRepo, mockFormDataRepo)

		dataID := uuid.New()
		formID := uuid.New()
		tenantID := uuid.New()
		userID := uuid.New()

		expectedFormData := &model.FormData{
			ID:          dataID,
			FormID:      formID,
			TenantID:    tenantID,
			Data:        map[string]interface{}{"days": 3, "reason": "Personal leave"},
			SubmittedBy: userID,
			SubmittedAt: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockFormDataRepo.On("FindByID", mock.Anything, dataID).
			Return(expectedFormData, nil).Once()

		req := &formv1.GetFormDataRequest{
			Id: dataID.String(),
		}

		resp, err := adapter.GetFormData(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, dataID.String(), resp.Id)
		assert.Equal(t, formID.String(), resp.FormId)
		mockFormDataRepo.AssertExpectations(t)
	})
}

// TestFormAdapter_UpdateFormData tests updating form data
func TestFormAdapter_UpdateFormData(t *testing.T) {
	t.Run("UpdateFormData successfully", func(t *testing.T) {
		mockFormDefRepo := new(MockFormDefinitionRepository)
		mockFormDataRepo := new(MockFormDataRepository)
		adapter := NewFormAdapter(mockFormDefRepo, mockFormDataRepo)

		dataID := uuid.New()
		formID := uuid.New()
		tenantID := uuid.New()
		userID := uuid.New()

		existingFormData := &model.FormData{
			ID:          dataID,
			FormID:      formID,
			TenantID:    tenantID,
			Data:        map[string]interface{}{"days": 3},
			SubmittedBy: userID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockFormDataRepo.On("FindByID", mock.Anything, dataID).
			Return(existingFormData, nil).Once()
		mockFormDataRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.FormData")).
			Return(nil).Once()

		req := &formv1.UpdateFormDataRequest{
			Id:   dataID.String(),
			Data: `{"days":5,"reason":"Updated reason"}`,
		}

		resp, err := adapter.UpdateFormData(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, dataID.String(), resp.Id)
		mockFormDataRepo.AssertExpectations(t)
	})
}

// TestFormAdapter_ListFormData tests listing form data
func TestFormAdapter_ListFormData(t *testing.T) {
	t.Run("ListFormData successfully", func(t *testing.T) {
		mockFormDefRepo := new(MockFormDefinitionRepository)
		mockFormDataRepo := new(MockFormDataRepository)
		adapter := NewFormAdapter(mockFormDefRepo, mockFormDataRepo)

		formID := uuid.New()
		tenantID := uuid.New()
		userID := uuid.New()

		expectedFormDataList := []*model.FormData{
			{
				ID:          uuid.New(),
				FormID:      formID,
				TenantID:    tenantID,
				Data:        map[string]interface{}{"days": 3},
				SubmittedBy: userID,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          uuid.New(),
				FormID:      formID,
				TenantID:    tenantID,
				Data:        map[string]interface{}{"days": 5},
				SubmittedBy: userID,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		mockFormDataRepo.On("ListByForm", mock.Anything, formID).
			Return(expectedFormDataList, nil).Once()

		req := &formv1.ListFormDataRequest{
			FormId:   formID.String(),
			TenantId: tenantID.String(),
		}

		resp, err := adapter.ListFormData(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Items, 2)
		assert.Equal(t, int32(2), resp.Total)
		mockFormDataRepo.AssertExpectations(t)
	})
}
