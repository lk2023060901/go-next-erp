package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
)

// CreateOrganizationTypeRequest 创建组织类型请求
type CreateOrganizationTypeRequest struct {
	Code               string   `json:"code" binding:"required,max=50"`
	Name               string   `json:"name" binding:"required,max=100"`
	Icon               string   `json:"icon" binding:"max=100"`
	Level              int      `json:"level" binding:"required,min=1"`
	MaxLevel           int      `json:"max_level" binding:"min=0"`
	AllowRoot          bool     `json:"allow_root"`
	AllowMulti         bool     `json:"allow_multi"`
	AllowedParentTypes []string `json:"allowed_parent_types"`
	AllowedChildTypes  []string `json:"allowed_child_types"`
	EnableLeader       bool     `json:"enable_leader"`
	EnableLegalInfo    bool     `json:"enable_legal_info"`
	EnableAddress      bool     `json:"enable_address"`
	Sort               int      `json:"sort"`
	Status             string   `json:"status" binding:"required,oneof=active inactive"`
}

// UpdateOrganizationTypeRequest 更新组织类型请求
type UpdateOrganizationTypeRequest struct {
	Name               string   `json:"name" binding:"required,max=100"`
	Icon               string   `json:"icon" binding:"max=100"`
	MaxLevel           int      `json:"max_level" binding:"min=0"`
	AllowRoot          bool     `json:"allow_root"`
	AllowMulti         bool     `json:"allow_multi"`
	AllowedParentTypes []string `json:"allowed_parent_types"`
	AllowedChildTypes  []string `json:"allowed_child_types"`
	EnableLeader       bool     `json:"enable_leader"`
	EnableLegalInfo    bool     `json:"enable_legal_info"`
	EnableAddress      bool     `json:"enable_address"`
	Sort               int      `json:"sort"`
	Status             string   `json:"status" binding:"required,oneof=active inactive"`
}

// OrganizationTypeResponse 组织类型响应
type OrganizationTypeResponse struct {
	ID                 uuid.UUID `json:"id"`
	TenantID           uuid.UUID `json:"tenant_id"`
	Code               string    `json:"code"`
	Name               string    `json:"name"`
	Icon               string    `json:"icon,omitempty"`
	Level              int       `json:"level"`
	MaxLevel           int       `json:"max_level"`
	AllowRoot          bool      `json:"allow_root"`
	AllowMulti         bool      `json:"allow_multi"`
	AllowedParentTypes []string  `json:"allowed_parent_types,omitempty"`
	AllowedChildTypes  []string  `json:"allowed_child_types,omitempty"`
	EnableLeader       bool      `json:"enable_leader"`
	EnableLegalInfo    bool      `json:"enable_legal_info"`
	EnableAddress      bool      `json:"enable_address"`
	Sort               int       `json:"sort"`
	Status             string    `json:"status"`
	IsSystem           bool      `json:"is_system"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ToOrganizationTypeResponse 转换为响应对象
func ToOrganizationTypeResponse(orgType *model.OrganizationType) *OrganizationTypeResponse {
	return &OrganizationTypeResponse{
		ID:                 orgType.ID,
		TenantID:           orgType.TenantID,
		Code:               orgType.Code,
		Name:               orgType.Name,
		Icon:               orgType.Icon,
		Level:              orgType.Level,
		MaxLevel:           orgType.MaxLevel,
		AllowRoot:          orgType.AllowRoot,
		AllowMulti:         orgType.AllowMulti,
		AllowedParentTypes: orgType.AllowedParentTypes,
		AllowedChildTypes:  orgType.AllowedChildTypes,
		EnableLeader:       orgType.EnableLeader,
		EnableLegalInfo:    orgType.EnableLegalInfo,
		EnableAddress:      orgType.EnableAddress,
		Sort:               orgType.Sort,
		Status:             orgType.Status,
		IsSystem:           orgType.IsSystem,
		CreatedAt:          orgType.CreatedAt,
		UpdatedAt:          orgType.UpdatedAt,
	}
}

// ToOrganizationTypeResponseList 批量转换
func ToOrganizationTypeResponseList(orgTypes []*model.OrganizationType) []*OrganizationTypeResponse {
	result := make([]*OrganizationTypeResponse, len(orgTypes))
	for i, orgType := range orgTypes {
		result[i] = ToOrganizationTypeResponse(orgType)
	}
	return result
}
