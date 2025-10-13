package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
)

// CreatePositionRequest 创建职位请求
type CreatePositionRequest struct {
	Code        string  `json:"code" binding:"required,max=50"`
	Name        string  `json:"name" binding:"required,max=100"`
	Description string  `json:"description"`
	OrgID       *string `json:"org_id" binding:"omitempty,uuid"`
	Level       int     `json:"level" binding:"min=0"`
	Category    string  `json:"category" binding:"omitempty,oneof=management technical sales support general"`
	Sort        int     `json:"sort"`
	Status      string  `json:"status" binding:"required,oneof=active inactive"`
}

// UpdatePositionRequest 更新职位请求
type UpdatePositionRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description"`
	Level       int    `json:"level" binding:"min=0"`
	Category    string `json:"category" binding:"omitempty,oneof=management technical sales support general"`
	Sort        int    `json:"sort"`
	Status      string `json:"status" binding:"required,oneof=active inactive"`
}

// PositionResponse 职位响应
type PositionResponse struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	OrgID       *uuid.UUID `json:"org_id,omitempty"`
	Level       int        `json:"level"`
	Category    string     `json:"category,omitempty"`
	Sort        int        `json:"sort"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToPositionResponse 转换为响应对象
func ToPositionResponse(pos *model.Position) *PositionResponse {
	return &PositionResponse{
		ID:          pos.ID,
		TenantID:    pos.TenantID,
		Code:        pos.Code,
		Name:        pos.Name,
		Description: pos.Description,
		OrgID:       pos.OrgID,
		Level:       pos.Level,
		Category:    pos.Category,
		Sort:        pos.Sort,
		Status:      pos.Status,
		CreatedAt:   pos.CreatedAt,
		UpdatedAt:   pos.UpdatedAt,
	}
}

// ToPositionResponseList 批量转换
func ToPositionResponseList(positions []*model.Position) []*PositionResponse {
	result := make([]*PositionResponse, len(positions))
	for i, pos := range positions {
		result[i] = ToPositionResponse(pos)
	}
	return result
}
