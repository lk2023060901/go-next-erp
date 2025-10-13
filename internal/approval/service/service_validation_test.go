package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	formModel "github.com/lk2023060901/go-next-erp/internal/form/model"
	"github.com/stretchr/testify/assert"
)

// TestValidateFormData tests the form data validation function
func TestValidateFormData(t *testing.T) {
	service := &approvalService{}

	t.Run("valid data with all fields", func(t *testing.T) {
		formDef := &formModel.FormDefinition{
			ID:       uuid.New(),
			TenantID: uuid.New(),
			Code:     "test-form",
			Name:     "Test Form",
			Fields: []formModel.FormField{
				{
					Key:      "title",
					Label:    "Title",
					Type:     "text",
					Required: true,
				},
				{
					Key:      "amount",
					Label:    "Amount",
					Type:     "number",
					Required: true,
				},
				{
					Key:      "description",
					Label:    "Description",
					Type:     "textarea",
					Required: false,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		data := map[string]interface{}{
			"title":       "Test Title",
			"amount":      1000,
			"description": "Test Description",
		}

		err := service.validateFormData(formDef, data)
		assert.NoError(t, err)
	})

	t.Run("valid data without optional fields", func(t *testing.T) {
		formDef := &formModel.FormDefinition{
			Fields: []formModel.FormField{
				{Key: "title", Required: true},
				{Key: "amount", Required: true},
				{Key: "description", Required: false},
			},
		}

		data := map[string]interface{}{
			"title":  "Test Title",
			"amount": 1000,
		}

		err := service.validateFormData(formDef, data)
		assert.NoError(t, err)
	})

	t.Run("missing required field - title", func(t *testing.T) {
		formDef := &formModel.FormDefinition{
			Fields: []formModel.FormField{
				{Key: "title", Required: true},
				{Key: "amount", Required: true},
			},
		}

		data := map[string]interface{}{
			"amount": 1000,
			// missing "title"
		}

		err := service.validateFormData(formDef, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required field missing")
		assert.Contains(t, err.Error(), "title")
	})

	t.Run("missing required field - amount", func(t *testing.T) {
		formDef := &formModel.FormDefinition{
			Fields: []formModel.FormField{
				{Key: "title", Required: true},
				{Key: "amount", Required: true},
			},
		}

		data := map[string]interface{}{
			"title": "Test",
			// missing "amount"
		}

		err := service.validateFormData(formDef, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount")
	})

	t.Run("empty form data with required fields", func(t *testing.T) {
		formDef := &formModel.FormDefinition{
			Fields: []formModel.FormField{
				{Key: "title", Required: true},
			},
		}

		data := map[string]interface{}{}

		err := service.validateFormData(formDef, data)
		assert.Error(t, err)
	})

	t.Run("form with no required fields", func(t *testing.T) {
		formDef := &formModel.FormDefinition{
			Fields: []formModel.FormField{
				{Key: "description", Required: false},
				{Key: "notes", Required: false},
			},
		}

		data := map[string]interface{}{}

		err := service.validateFormData(formDef, data)
		assert.NoError(t, err)
	})

	t.Run("extra fields in data", func(t *testing.T) {
		formDef := &formModel.FormDefinition{
			Fields: []formModel.FormField{
				{Key: "title", Required: true},
			},
		}

		data := map[string]interface{}{
			"title":      "Test",
			"extra_field": "Should be ignored",
		}

		err := service.validateFormData(formDef, data)
		assert.NoError(t, err) // Extra fields are allowed
	})
}

// TestServiceErrors tests the error definitions
func TestServiceErrors(t *testing.T) {
	t.Run("error definitions exist", func(t *testing.T) {
		assert.NotNil(t, ErrProcessNotFound)
		assert.NotNil(t, ErrProcessInstanceNotFound)
		assert.NotNil(t, ErrTaskNotFound)
		assert.NotNil(t, ErrInvalidAction)
		assert.NotNil(t, ErrTaskAlreadyProcessed)
		assert.NotNil(t, ErrUnauthorized)
	})

	t.Run("error messages are descriptive", func(t *testing.T) {
		assert.Contains(t, ErrProcessNotFound.Error(), "process definition")
		assert.Contains(t, ErrProcessInstanceNotFound.Error(), "process instance")
		assert.Contains(t, ErrTaskNotFound.Error(), "task")
		assert.Contains(t, ErrInvalidAction.Error(), "invalid")
		assert.Contains(t, ErrTaskAlreadyProcessed.Error(), "already processed")
		assert.Contains(t, ErrUnauthorized.Error(), "unauthorized")
	})
}
