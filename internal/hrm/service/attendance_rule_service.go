package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
)

// AttendanceRuleService 考勤规则服务接口
type AttendanceRuleService interface {
	// Create 创建考勤规则
	Create(ctx context.Context, req *CreateAttendanceRuleRequest) (*model.AttendanceRule, error)

	// GetByID 根据ID获取考勤规则
	GetByID(ctx context.Context, id uuid.UUID) (*model.AttendanceRule, error)

	// Update 更新考勤规则
	Update(ctx context.Context, id uuid.UUID, req *UpdateAttendanceRuleRequest) (*model.AttendanceRule, error)

	// Delete 删除考勤规则
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询
	List(ctx context.Context, tenantID uuid.UUID, filter *repository.AttendanceRuleFilter, offset, limit int) ([]*model.AttendanceRule, int, error)
}

// CreateAttendanceRuleRequest 创建考勤规则请求
type CreateAttendanceRuleRequest struct {
	TenantID         uuid.UUID
	Name             string
	Description      string
	LocationRequired bool
	AllowedLocations []model.AllowedLocation
	WiFiRequired     bool
	AllowedWiFi      []string
	FaceRequired     bool
	FaceThreshold    float64
	IsActive         bool
	CreatedBy        uuid.UUID
}

// UpdateAttendanceRuleRequest 更新考勤规则请求
type UpdateAttendanceRuleRequest struct {
	Name             *string
	Description      *string
	LocationRequired *bool
	AllowedLocations *[]model.AllowedLocation
	WiFiRequired     *bool
	AllowedWiFi      *[]string
	FaceRequired     *bool
	FaceThreshold    *float64
	IsActive         *bool
	UpdatedBy        uuid.UUID
}

type attendanceRuleService struct {
	ruleRepo repository.AttendanceRuleRepository
}

// NewAttendanceRuleService 创建考勤规则服务
func NewAttendanceRuleService(ruleRepo repository.AttendanceRuleRepository) AttendanceRuleService {
	return &attendanceRuleService{
		ruleRepo: ruleRepo,
	}
}

func (s *attendanceRuleService) Create(ctx context.Context, req *CreateAttendanceRuleRequest) (*model.AttendanceRule, error) {
	rule := &model.AttendanceRule{
		ID:               uuid.New(),
		TenantID:         req.TenantID,
		Name:             req.Name,
		Description:      req.Description,
		LocationRequired: req.LocationRequired,
		AllowedLocations: req.AllowedLocations,
		WiFiRequired:     req.WiFiRequired,
		AllowedWiFi:      req.AllowedWiFi,
		FaceRequired:     req.FaceRequired,
		FaceThreshold:    req.FaceThreshold,
		IsActive:         req.IsActive,
		CreatedBy:        req.CreatedBy,
		UpdatedBy:        req.CreatedBy,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("create attendance rule failed: %w", err)
	}

	return rule, nil
}

func (s *attendanceRuleService) GetByID(ctx context.Context, id uuid.UUID) (*model.AttendanceRule, error) {
	return s.ruleRepo.FindByID(ctx, id)
}

func (s *attendanceRuleService) Update(ctx context.Context, id uuid.UUID, req *UpdateAttendanceRuleRequest) (*model.AttendanceRule, error) {
	rule, err := s.ruleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Description != nil {
		rule.Description = *req.Description
	}
	if req.LocationRequired != nil {
		rule.LocationRequired = *req.LocationRequired
	}
	if req.AllowedLocations != nil {
		rule.AllowedLocations = *req.AllowedLocations
	}
	if req.WiFiRequired != nil {
		rule.WiFiRequired = *req.WiFiRequired
	}
	if req.AllowedWiFi != nil {
		rule.AllowedWiFi = *req.AllowedWiFi
	}
	if req.FaceRequired != nil {
		rule.FaceRequired = *req.FaceRequired
	}
	if req.FaceThreshold != nil {
		rule.FaceThreshold = *req.FaceThreshold
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}

	rule.UpdatedBy = req.UpdatedBy
	rule.UpdatedAt = time.Now()

	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("update attendance rule failed: %w", err)
	}

	return rule, nil
}

func (s *attendanceRuleService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.ruleRepo.Delete(ctx, id)
}

func (s *attendanceRuleService) List(ctx context.Context, tenantID uuid.UUID, filter *repository.AttendanceRuleFilter, offset, limit int) ([]*model.AttendanceRule, int, error) {
	return s.ruleRepo.List(ctx, tenantID, filter, offset, limit)
}
