package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
)

// CreateOrganizationRequest 创建组织请求
type CreateOrganizationRequest struct {
	Code         string     `json:"code" binding:"required,max=50"`
	Name         string     `json:"name" binding:"required,max=200"`
	ShortName    string     `json:"short_name" binding:"max=100"`
	Description  string     `json:"description"`
	TypeID       string     `json:"type_id" binding:"required,uuid"`
	ParentID     *string    `json:"parent_id" binding:"omitempty,uuid"`
	LeaderID     *string    `json:"leader_id" binding:"omitempty,uuid"`
	LeaderName   string     `json:"leader_name" binding:"max=100"`
	LegalPerson  string     `json:"legal_person" binding:"max=100"`
	UnifiedCode  string     `json:"unified_code" binding:"max=50"`
	RegisterDate *time.Time `json:"register_date"`
	RegisterAddr string     `json:"register_addr" binding:"max=500"`
	Phone        string     `json:"phone" binding:"max=50"`
	Email        string     `json:"email" binding:"omitempty,email,max=100"`
	Address      string     `json:"address" binding:"max=500"`
	Sort         int        `json:"sort"`
	Status       string     `json:"status" binding:"required,oneof=active inactive disbanded"`
	Tags         []string   `json:"tags"`
}

// UpdateOrganizationRequest 更新组织请求
type UpdateOrganizationRequest struct {
	Name         string     `json:"name" binding:"required,max=200"`
	ShortName    string     `json:"short_name" binding:"max=100"`
	Description  string     `json:"description"`
	LeaderID     *string    `json:"leader_id" binding:"omitempty,uuid"`
	LeaderName   string     `json:"leader_name" binding:"max=100"`
	LegalPerson  string     `json:"legal_person" binding:"max=100"`
	UnifiedCode  string     `json:"unified_code" binding:"max=50"`
	RegisterDate *time.Time `json:"register_date"`
	RegisterAddr string     `json:"register_addr" binding:"max=500"`
	Phone        string     `json:"phone" binding:"max=50"`
	Email        string     `json:"email" binding:"omitempty,email,max=100"`
	Address      string     `json:"address" binding:"max=500"`
	Sort         int        `json:"sort"`
	Status       string     `json:"status" binding:"required,oneof=active inactive disbanded"`
	Tags         []string   `json:"tags"`
}

// MoveOrganizationRequest 移动组织请求
type MoveOrganizationRequest struct {
	NewParentID string `json:"new_parent_id" binding:"required,uuid"`
}

// OrganizationResponse 组织响应
type OrganizationResponse struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	ShortName      string     `json:"short_name,omitempty"`
	Description    string     `json:"description,omitempty"`
	TypeID         uuid.UUID  `json:"type_id"`
	TypeCode       string     `json:"type_code"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	Level          int        `json:"level"`
	Path           string     `json:"path"`
	PathNames      string     `json:"path_names"`
	IsLeaf         bool       `json:"is_leaf"`
	LeaderID       *uuid.UUID `json:"leader_id,omitempty"`
	LeaderName     string     `json:"leader_name,omitempty"`
	LegalPerson    string     `json:"legal_person,omitempty"`
	UnifiedCode    string     `json:"unified_code,omitempty"`
	RegisterDate   *time.Time `json:"register_date,omitempty"`
	RegisterAddr   string     `json:"register_addr,omitempty"`
	Phone          string     `json:"phone,omitempty"`
	Email          string     `json:"email,omitempty"`
	Address        string     `json:"address,omitempty"`
	EmployeeCount  int        `json:"employee_count"`
	DirectEmpCount int        `json:"direct_emp_count"`
	Sort           int        `json:"sort"`
	Status         string     `json:"status"`
	Tags           []string   `json:"tags,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// OrganizationTreeResponse 组织树响应
type OrganizationTreeResponse struct {
	*OrganizationResponse
	Children []*OrganizationTreeResponse `json:"children,omitempty"`
}

// ToOrganizationResponse 转换为响应对象
func ToOrganizationResponse(org *model.Organization) *OrganizationResponse {
	return &OrganizationResponse{
		ID:             org.ID,
		TenantID:       org.TenantID,
		Code:           org.Code,
		Name:           org.Name,
		ShortName:      org.ShortName,
		Description:    org.Description,
		TypeID:         org.TypeID,
		TypeCode:       org.TypeCode,
		ParentID:       org.ParentID,
		Level:          org.Level,
		Path:           org.Path,
		PathNames:      org.PathNames,
		IsLeaf:         org.IsLeaf,
		LeaderID:       org.LeaderID,
		LeaderName:     org.LeaderName,
		LegalPerson:    org.LegalPerson,
		UnifiedCode:    org.UnifiedCode,
		RegisterDate:   org.RegisterDate,
		RegisterAddr:   org.RegisterAddr,
		Phone:          org.Phone,
		Email:          org.Email,
		Address:        org.Address,
		EmployeeCount:  org.EmployeeCount,
		DirectEmpCount: org.DirectEmpCount,
		Sort:           org.Sort,
		Status:         org.Status,
		Tags:           org.Tags,
		CreatedAt:      org.CreatedAt,
		UpdatedAt:      org.UpdatedAt,
	}
}

// ToOrganizationResponseList 批量转换
func ToOrganizationResponseList(orgs []*model.Organization) []*OrganizationResponse {
	result := make([]*OrganizationResponse, len(orgs))
	for i, org := range orgs {
		result[i] = ToOrganizationResponse(org)
	}
	return result
}
