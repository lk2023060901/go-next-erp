package adapter

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	formv1 "github.com/lk2023060901/go-next-erp/api/form/v1"
	"github.com/lk2023060901/go-next-erp/internal/form/model"
	"github.com/lk2023060901/go-next-erp/internal/form/repository"
)

// FormAdapter 表单适配器
type FormAdapter struct {
	formv1.UnimplementedFormDefinitionServiceServer
	formv1.UnimplementedFormDataServiceServer
	formDefRepo  repository.FormDefinitionRepository
	formDataRepo repository.FormDataRepository
}

// NewFormAdapter 创建表单适配器
func NewFormAdapter(
	formDefRepo repository.FormDefinitionRepository,
	formDataRepo repository.FormDataRepository,
) *FormAdapter {
	return &FormAdapter{
		formDefRepo:  formDefRepo,
		formDataRepo: formDataRepo,
	}
}

// CreateFormDefinition 创建表单定义
func (a *FormAdapter) CreateFormDefinition(ctx context.Context, req *formv1.CreateFormDefinitionRequest) (*formv1.FormDefinitionResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}

	// 解析 schema 为 Fields
	var fields []model.FormField
	if req.Schema != "" {
		if err := json.Unmarshal([]byte(req.Schema), &fields); err != nil {
			return nil, err
		}
	}

	formDef := &model.FormDefinition{
		ID:        uuid.Must(uuid.NewV7()),
		TenantID:  tenantID,
		Code:      req.Code,
		Name:      req.Name,
		Fields:    fields,
		Enabled:   true,
		CreatedBy: tenantID, // TODO: 从上下文获取当前用户ID
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := a.formDefRepo.Create(ctx, formDef); err != nil {
		return nil, err
	}

	return a.formDefinitionToProto(formDef), nil
}

// GetFormDefinition 获取表单定义
func (a *FormAdapter) GetFormDefinition(ctx context.Context, req *formv1.GetFormDefinitionRequest) (*formv1.FormDefinitionResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	formDef, err := a.formDefRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return a.formDefinitionToProto(formDef), nil
}

// UpdateFormDefinition 更新表单定义
func (a *FormAdapter) UpdateFormDefinition(ctx context.Context, req *formv1.UpdateFormDefinitionRequest) (*formv1.FormDefinitionResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	formDef, err := a.formDefRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	formDef.Name = req.Name
	if req.Schema != "" {
		var fields []model.FormField
		if err := json.Unmarshal([]byte(req.Schema), &fields); err != nil {
			return nil, err
		}
		formDef.Fields = fields
	}
	formDef.UpdatedAt = time.Now()

	if err := a.formDefRepo.Update(ctx, formDef); err != nil {
		return nil, err
	}

	return a.formDefinitionToProto(formDef), nil
}

// DeleteFormDefinition 删除表单定义
func (a *FormAdapter) DeleteFormDefinition(ctx context.Context, req *formv1.DeleteFormDefinitionRequest) (*formv1.DeleteFormDefinitionResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	if err := a.formDefRepo.Delete(ctx, id); err != nil {
		return nil, err
	}

	return &formv1.DeleteFormDefinitionResponse{Success: true}, nil
}

// ListFormDefinitions 列出表单定义
func (a *FormAdapter) ListFormDefinitions(ctx context.Context, req *formv1.ListFormDefinitionsRequest) (*formv1.ListFormDefinitionsResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}

	formDefs, err := a.formDefRepo.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	items := make([]*formv1.FormDefinitionResponse, 0, len(formDefs))
	for _, formDef := range formDefs {
		items = append(items, a.formDefinitionToProto(formDef))
	}

	return &formv1.ListFormDefinitionsResponse{
		Items: items,
		Total: int32(len(items)),
	}, nil
}

// SubmitFormData 提交表单数据
func (a *FormAdapter) SubmitFormData(ctx context.Context, req *formv1.SubmitFormDataRequest) (*formv1.FormDataResponse, error) {
	formID, err := uuid.Parse(req.FormId)
	if err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	// 解析数据
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(req.Data), &data); err != nil {
		return nil, err
	}

	formData := &model.FormData{
		ID:          uuid.Must(uuid.NewV7()),
		TenantID:    tenantID,
		FormID:      formID,
		Data:        data,
		SubmittedBy: userID,
		SubmittedAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := a.formDataRepo.Create(ctx, formData); err != nil {
		return nil, err
	}

	return a.formDataToProto(formData), nil
}

// GetFormData 获取表单数据
func (a *FormAdapter) GetFormData(ctx context.Context, req *formv1.GetFormDataRequest) (*formv1.FormDataResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	formData, err := a.formDataRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return a.formDataToProto(formData), nil
}

// UpdateFormData 更新表单数据
func (a *FormAdapter) UpdateFormData(ctx context.Context, req *formv1.UpdateFormDataRequest) (*formv1.FormDataResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	formData, err := a.formDataRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新数据
	if req.Data != "" {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(req.Data), &data); err != nil {
			return nil, err
		}
		formData.Data = data
	}
	formData.UpdatedAt = time.Now()

	if err := a.formDataRepo.Update(ctx, formData); err != nil {
		return nil, err
	}

	return a.formDataToProto(formData), nil
}

// ListFormData 查询表单数据列表
func (a *FormAdapter) ListFormData(ctx context.Context, req *formv1.ListFormDataRequest) (*formv1.ListFormDataResponse, error) {
	formID, err := uuid.Parse(req.FormId)
	if err != nil {
		return nil, err
	}

	// 验证 tenantID 但不使用（仅用于参数校验）
	if _, err := uuid.Parse(req.TenantId); err != nil {
		return nil, err
	}

	formDataList, err := a.formDataRepo.ListByForm(ctx, formID)
	if err != nil {
		return nil, err
	}

	items := make([]*formv1.FormDataResponse, 0, len(formDataList))
	for _, formData := range formDataList {
		items = append(items, a.formDataToProto(formData))
	}

	return &formv1.ListFormDataResponse{
		Items: items,
		Total: int32(len(items)),
	}, nil
}

// 辅助方法：将 model.FormDefinition 转换为 proto
func (a *FormAdapter) formDefinitionToProto(formDef *model.FormDefinition) *formv1.FormDefinitionResponse {
	// 将 Fields 序列化为 JSON
	schemaBytes, _ := json.Marshal(formDef.Fields)

	return &formv1.FormDefinitionResponse{
		Id:          formDef.ID.String(),
		TenantId:    formDef.TenantID.String(),
		Name:        formDef.Name,
		Code:        formDef.Code,
		Description: "",
		Category:    "",
		Schema:      string(schemaBytes),
		UiSchema:    "",
		Version:     1,
		CreatedAt:   formDef.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   formDef.UpdatedAt.Format(time.RFC3339),
	}
}

// 辅助方法：将 model.FormData 转换为 proto
func (a *FormAdapter) formDataToProto(formData *model.FormData) *formv1.FormDataResponse {
	// 将 Data 序列化为 JSON
	dataBytes, _ := json.Marshal(formData.Data)

	return &formv1.FormDataResponse{
		Id:        formData.ID.String(),
		FormId:    formData.FormID.String(),
		TenantId:  formData.TenantID.String(),
		UserId:    formData.SubmittedBy.String(),
		Data:      string(dataBytes),
		CreatedAt: formData.CreatedAt.Format(time.RFC3339),
		UpdatedAt: formData.UpdatedAt.Format(time.RFC3339),
	}
}
