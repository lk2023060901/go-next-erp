package service

import (
	"testing"

	"github.com/lk2023060901/go-next-erp/internal/form/model"
	"github.com/stretchr/testify/assert"
)

// TestValidateRule_Min tests minimum value/length validation
func TestValidateRule_Min(t *testing.T) {
	service := &formService{}

	tests := []struct {
		name    string
		field   model.FormField
		rule    model.ValidationRule
		value   interface{}
		wantErr bool
	}{
		{
			name:  "number - valid (above min)",
			field: model.FormField{Key: "age", Type: "number"},
			rule:  model.ValidationRule{Type: "min", Value: float64(18)},
			value: float64(20),
		},
		{
			name:  "number - valid (equals min)",
			field: model.FormField{Key: "age", Type: "number"},
			rule:  model.ValidationRule{Type: "min", Value: float64(18)},
			value: float64(18),
		},
		{
			name:    "number - invalid (below min)",
			field:   model.FormField{Key: "age", Type: "number"},
			rule:    model.ValidationRule{Type: "min", Value: float64(18)},
			value:   float64(15),
			wantErr: true,
		},
		{
			name:  "string - valid length",
			field: model.FormField{Key: "name", Type: "text"},
			rule:  model.ValidationRule{Type: "min", Value: float64(3)},
			value: "John",
		},
		{
			name:    "string - invalid length",
			field:   model.FormField{Key: "name", Type: "text"},
			rule:    model.ValidationRule{Type: "min", Value: float64(3)},
			value:   "Jo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateRule(tt.field, tt.value, tt.rule)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateRule_Max tests maximum value/length validation
func TestValidateRule_Max(t *testing.T) {
	service := &formService{}

	tests := []struct {
		name    string
		field   model.FormField
		rule    model.ValidationRule
		value   interface{}
		wantErr bool
	}{
		{
			name:  "number - valid (below max)",
			field: model.FormField{Key: "score", Type: "number"},
			rule:  model.ValidationRule{Type: "max", Value: float64(100)},
			value: float64(95),
		},
		{
			name:  "number - valid (equals max)",
			field: model.FormField{Key: "score", Type: "number"},
			rule:  model.ValidationRule{Type: "max", Value: float64(100)},
			value: float64(100),
		},
		{
			name:    "number - invalid (above max)",
			field:   model.FormField{Key: "score", Type: "number"},
			rule:    model.ValidationRule{Type: "max", Value: float64(100)},
			value:   float64(105),
			wantErr: true,
		},
		{
			name:  "string - valid length",
			field: model.FormField{Key: "code", Type: "text"},
			rule:  model.ValidationRule{Type: "max", Value: float64(5)},
			value: "ABC",
		},
		{
			name:    "string - invalid length",
			field:   model.FormField{Key: "code", Type: "text"},
			rule:    model.ValidationRule{Type: "max", Value: float64(5)},
			value:   "ABCDEF",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateRule(tt.field, tt.value, tt.rule)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateRule_Pattern tests pattern validation
func TestValidateRule_Pattern(t *testing.T) {
	service := &formService{}

	tests := []struct {
		name    string
		pattern string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "email - valid",
			pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			value:   "user@example.com",
		},
		{
			name:    "email - invalid",
			pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			value:   "invalid-email",
			wantErr: true,
		},
		{
			name:    "phone - valid",
			pattern: `^\d{10,11}$`,
			value:   "1234567890",
		},
		{
			name:    "phone - invalid",
			pattern: `^\d{10,11}$`,
			value:   "123-456-7890",
			wantErr: true,
		},
		{
			name:    "non-string value",
			pattern: `^\d+$`,
			value:   float64(123),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := model.FormField{Key: "test", Type: "text"}
			rule := model.ValidationRule{Type: "pattern", Value: tt.pattern}

			err := service.validateRule(field, tt.value, rule)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateFieldType tests field type validation
func TestValidateFieldType(t *testing.T) {
	service := &formService{}

	tests := []struct {
		name      string
		fieldType model.FieldType
		value     interface{}
		wantErr   bool
	}{
		{
			name:      "number - valid",
			fieldType: model.FieldTypeNumber,
			value:     float64(100),
		},
		{
			name:      "number - invalid (string)",
			fieldType: model.FieldTypeNumber,
			value:     "100",
			wantErr:   true,
		},
		{
			name:      "date - valid",
			fieldType: model.FieldTypeDate,
			value:     "2023-01-15",
		},
		{
			name:      "date - invalid (number)",
			fieldType: model.FieldTypeDate,
			value:     20230115,
			wantErr:   true,
		},
		{
			name:      "datetime - valid",
			fieldType: model.FieldTypeDateTime,
			value:     "2023-01-15T10:30:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := model.FormField{
				Key:   "test",
				Label: "Test",
				Type:  tt.fieldType,
			}

			err := service.validateFieldType(field, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFormServiceErrors tests error definitions
func TestFormServiceErrors(t *testing.T) {
	assert.NotNil(t, ErrFormDefinitionNotFound)
	assert.NotNil(t, ErrFormCodeExists)
	assert.NotNil(t, ErrInvalidFormData)

	assert.Contains(t, ErrFormDefinitionNotFound.Error(), "form definition")
	assert.Contains(t, ErrFormCodeExists.Error(), "code")
	assert.Contains(t, ErrInvalidFormData.Error(), "invalid")
}
