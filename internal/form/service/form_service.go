package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/form/dto"
	"github.com/lk2023060901/go-next-erp/internal/form/model"
	"github.com/lk2023060901/go-next-erp/internal/form/repository"
)

var (
	ErrFormDefinitionNotFound = errors.New("form definition not found")
	ErrFormCodeExists         = errors.New("form code already exists")
	ErrInvalidFormData        = errors.New("invalid form data")
)

// FormService 表单服务接口
type FormService interface {
	// 表单定义管理
	CreateFormDefinition(ctx context.Context, req *CreateFormDefinitionRequest) (*model.FormDefinition, error)
	UpdateFormDefinition(ctx context.Context, id uuid.UUID, req *UpdateFormDefinitionRequest) (*model.FormDefinition, error)
	DeleteFormDefinition(ctx context.Context, id uuid.UUID) error
	GetFormDefinition(ctx context.Context, id uuid.UUID) (*model.FormDefinition, error)
	GetFormDefinitionByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.FormDefinition, error)
	ListFormDefinitions(ctx context.Context, tenantID uuid.UUID) ([]*model.FormDefinition, error)
	ListEnabledFormDefinitions(ctx context.Context, tenantID uuid.UUID) ([]*model.FormDefinition, error)

	// 表单数据管理
	SubmitFormData(ctx context.Context, req *SubmitFormDataRequest) (*model.FormData, error)
	ValidateFormData(ctx context.Context, formID uuid.UUID, data map[string]interface{}) ([]dto.ValidationError, error)
	GetFormData(ctx context.Context, id uuid.UUID) (*model.FormData, error)
	GetFormDataByRelated(ctx context.Context, relatedType string, relatedID uuid.UUID) (*model.FormData, error)
}

type formService struct {
	formDefRepo  repository.FormDefinitionRepository
	formDataRepo repository.FormDataRepository
}

// NewFormService 创建表单服务
func NewFormService(
	formDefRepo repository.FormDefinitionRepository,
	formDataRepo repository.FormDataRepository,
) FormService {
	return &formService{
		formDefRepo:  formDefRepo,
		formDataRepo: formDataRepo,
	}
}

// CreateFormDefinitionRequest 创建表单定义请求
type CreateFormDefinitionRequest struct {
	TenantID  uuid.UUID
	Code      string
	Name      string
	Fields    []model.FormField
	CreatedBy uuid.UUID
}

func (s *formService) CreateFormDefinition(ctx context.Context, req *CreateFormDefinitionRequest) (*model.FormDefinition, error) {
	// 检查编码是否已存在
	existing, _ := s.formDefRepo.FindByCode(ctx, req.TenantID, req.Code)
	if existing != nil {
		return nil, ErrFormCodeExists
	}

	now := time.Now()
	form := &model.FormDefinition{
		ID:        uuid.New(),
		TenantID:  req.TenantID,
		Code:      req.Code,
		Name:      req.Name,
		Fields:    req.Fields,
		Enabled:   true,
		CreatedBy: req.CreatedBy,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.formDefRepo.Create(ctx, form); err != nil {
		return nil, err
	}

	return form, nil
}

// UpdateFormDefinitionRequest 更新表单定义请求
type UpdateFormDefinitionRequest struct {
	Name      *string
	Fields    *[]model.FormField
	Enabled   *bool
	UpdatedBy uuid.UUID
}

func (s *formService) UpdateFormDefinition(ctx context.Context, id uuid.UUID, req *UpdateFormDefinitionRequest) (*model.FormDefinition, error) {
	form, err := s.formDefRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrFormDefinitionNotFound
	}

	if req.Name != nil {
		form.Name = *req.Name
	}
	if req.Fields != nil {
		form.Fields = *req.Fields
	}
	if req.Enabled != nil {
		form.Enabled = *req.Enabled
	}

	form.UpdatedBy = &req.UpdatedBy
	form.UpdatedAt = time.Now()

	if err := s.formDefRepo.Update(ctx, form); err != nil {
		return nil, err
	}

	return form, nil
}

func (s *formService) DeleteFormDefinition(ctx context.Context, id uuid.UUID) error {
	return s.formDefRepo.Delete(ctx, id)
}

func (s *formService) GetFormDefinition(ctx context.Context, id uuid.UUID) (*model.FormDefinition, error) {
	return s.formDefRepo.FindByID(ctx, id)
}

func (s *formService) GetFormDefinitionByCode(ctx context.Context, tenantID uuid.UUID, code string) (*model.FormDefinition, error) {
	return s.formDefRepo.FindByCode(ctx, tenantID, code)
}

func (s *formService) ListFormDefinitions(ctx context.Context, tenantID uuid.UUID) ([]*model.FormDefinition, error) {
	return s.formDefRepo.List(ctx, tenantID)
}

func (s *formService) ListEnabledFormDefinitions(ctx context.Context, tenantID uuid.UUID) ([]*model.FormDefinition, error) {
	return s.formDefRepo.ListEnabled(ctx, tenantID)
}

// SubmitFormDataRequest 提交表单数据请求
type SubmitFormDataRequest struct {
	TenantID    uuid.UUID
	FormID      uuid.UUID
	Data        map[string]interface{}
	SubmittedBy uuid.UUID
	RelatedType *string
	RelatedID   *uuid.UUID
}

func (s *formService) SubmitFormData(ctx context.Context, req *SubmitFormDataRequest) (*model.FormData, error) {
	// 验证表单数据
	validationErrors, err := s.ValidateFormData(ctx, req.FormID, req.Data)
	if err != nil {
		return nil, err
	}

	if len(validationErrors) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrInvalidFormData, validationErrors)
	}

	now := time.Now()
	formData := &model.FormData{
		ID:          uuid.New(),
		TenantID:    req.TenantID,
		FormID:      req.FormID,
		Data:        req.Data,
		SubmittedBy: req.SubmittedBy,
		SubmittedAt: now,
		RelatedType: req.RelatedType,
		RelatedID:   req.RelatedID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.formDataRepo.Create(ctx, formData); err != nil {
		return nil, err
	}

	return formData, nil
}

func (s *formService) ValidateFormData(ctx context.Context, formID uuid.UUID, data map[string]interface{}) ([]dto.ValidationError, error) {
	// 获取表单定义
	formDef, err := s.formDefRepo.FindByID(ctx, formID)
	if err != nil {
		return nil, ErrFormDefinitionNotFound
	}

	var validationErrors []dto.ValidationError

	// 验证每个字段
	for _, field := range formDef.Fields {
		value, exists := data[field.Key]

		// 必填验证
		if field.Required && (!exists || value == nil || value == "") {
			validationErrors = append(validationErrors, dto.ValidationError{
				Field:   field.Key,
				Message: fmt.Sprintf("%s is required", field.Label),
			})
			continue
		}

		// 如果字段不存在且非必填，跳过后续验证
		if !exists {
			continue
		}

		// 执行自定义验证规则
		for _, rule := range field.Rules {
			if err := s.validateRule(field, value, rule); err != nil {
				validationErrors = append(validationErrors, dto.ValidationError{
					Field:   field.Key,
					Message: rule.Message,
				})
			}
		}

		// 类型验证
		if err := s.validateFieldType(field, value); err != nil {
			validationErrors = append(validationErrors, dto.ValidationError{
				Field:   field.Key,
				Message: err.Error(),
			})
		}
	}

	return validationErrors, nil
}

func (s *formService) validateRule(field model.FormField, value interface{}, rule model.ValidationRule) error {
	switch rule.Type {
	case "min":
		minVal, ok := rule.Value.(float64)
		if !ok {
			return nil
		}
		if numVal, ok := value.(float64); ok && numVal < minVal {
			return fmt.Errorf("value must be at least %v", minVal)
		}
		if strVal, ok := value.(string); ok && float64(len(strVal)) < minVal {
			return fmt.Errorf("length must be at least %v", minVal)
		}

	case "max":
		maxVal, ok := rule.Value.(float64)
		if !ok {
			return nil
		}
		if numVal, ok := value.(float64); ok && numVal > maxVal {
			return fmt.Errorf("value must be at most %v", maxVal)
		}
		if strVal, ok := value.(string); ok && float64(len(strVal)) > maxVal {
			return fmt.Errorf("length must be at most %v", maxVal)
		}

	case "pattern":
		pattern, ok := rule.Value.(string)
		if !ok {
			return nil
		}
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("value must be a string")
		}
		matched, err := regexp.MatchString(pattern, strVal)
		if err != nil || !matched {
			return fmt.Errorf("value does not match pattern")
		}
	}

	return nil
}

func (s *formService) validateFieldType(field model.FormField, value interface{}) error {
	switch field.Type {
	case model.FieldTypeNumber:
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("%s must be a number", field.Label)
		}
	case model.FieldTypeDate, model.FieldTypeDateTime:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%s must be a date string", field.Label)
		}
	}
	return nil
}

func (s *formService) GetFormData(ctx context.Context, id uuid.UUID) (*model.FormData, error) {
	return s.formDataRepo.FindByID(ctx, id)
}

func (s *formService) GetFormDataByRelated(ctx context.Context, relatedType string, relatedID uuid.UUID) (*model.FormData, error) {
	return s.formDataRepo.FindByRelated(ctx, relatedType, relatedID)
}
