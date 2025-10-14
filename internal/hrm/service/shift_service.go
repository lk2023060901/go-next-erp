package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
)

// ShiftService 班次服务接口
type ShiftService interface {
	// Create 创建班次
	Create(ctx context.Context, req *CreateShiftRequest) (*model.Shift, error)

	// GetByID 根据ID获取班次
	GetByID(ctx context.Context, id uuid.UUID) (*model.Shift, error)

	// Update 更新班次
	Update(ctx context.Context, id uuid.UUID, req *UpdateShiftRequest) (*model.Shift, error)

	// Delete 删除班次
	Delete(ctx context.Context, id uuid.UUID) error

	// List 列表查询
	List(ctx context.Context, tenantID uuid.UUID, filter *repository.ShiftFilter, offset, limit int) ([]*model.Shift, int, error)

	// ListActive 查询启用的班次
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.Shift, error)
}

// CreateShiftRequest 创建班次请求
type CreateShiftRequest struct {
	TenantID            uuid.UUID
	Code                string
	Name                string
	Description         string
	Type                model.ShiftType
	WorkStart           string
	WorkEnd             string
	FlexibleStart       string
	FlexibleEnd         string
	WorkDuration        int
	CheckInRequired     bool
	CheckOutRequired    bool
	LateGracePeriod     int
	EarlyGracePeriod    int
	RestPeriods         []model.RestPeriod
	IsCrossDays         bool
	AllowOvertime       bool
	OvertimeStartBuffer int
	OvertimeMinDuration int
	OvertimePayRate     float64
	WorkdayTypes        []string
	Color               string
	IsActive            bool
	Sort                int
	CreatedBy           uuid.UUID
}

// UpdateShiftRequest 更新班次请求
type UpdateShiftRequest struct {
	Code                *string
	Name                *string
	Description         *string
	Type                *model.ShiftType
	WorkStart           *string
	WorkEnd             *string
	FlexibleStart       *string
	FlexibleEnd         *string
	WorkDuration        *int
	CheckInRequired     *bool
	CheckOutRequired    *bool
	LateGracePeriod     *int
	EarlyGracePeriod    *int
	RestPeriods         *[]model.RestPeriod
	IsCrossDays         *bool
	AllowOvertime       *bool
	OvertimeStartBuffer *int
	OvertimeMinDuration *int
	OvertimePayRate     *float64
	WorkdayTypes        *[]string
	Color               *string
	IsActive            *bool
	Sort                *int
	UpdatedBy           uuid.UUID
}

type shiftService struct {
	shiftRepo repository.ShiftRepository
}

// NewShiftService 创建班次服务
func NewShiftService(shiftRepo repository.ShiftRepository) ShiftService {
	return &shiftService{
		shiftRepo: shiftRepo,
	}
}

func (s *shiftService) Create(ctx context.Context, req *CreateShiftRequest) (*model.Shift, error) {
	// 检查编码是否重复
	existing, err := s.shiftRepo.FindByCode(ctx, req.TenantID, req.Code)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("shift code already exists: %s", req.Code)
	}

	shift := &model.Shift{
		ID:                  uuid.New(),
		TenantID:            req.TenantID,
		Code:                req.Code,
		Name:                req.Name,
		Description:         req.Description,
		Type:                req.Type,
		WorkStart:           req.WorkStart,
		WorkEnd:             req.WorkEnd,
		FlexibleStart:       req.FlexibleStart,
		FlexibleEnd:         req.FlexibleEnd,
		WorkDuration:        req.WorkDuration,
		CheckInRequired:     req.CheckInRequired,
		CheckOutRequired:    req.CheckOutRequired,
		LateGracePeriod:     req.LateGracePeriod,
		EarlyGracePeriod:    req.EarlyGracePeriod,
		RestPeriods:         req.RestPeriods,
		IsCrossDays:         req.IsCrossDays,
		AllowOvertime:       req.AllowOvertime,
		OvertimeStartBuffer: req.OvertimeStartBuffer,
		OvertimeMinDuration: req.OvertimeMinDuration,
		OvertimePayRate:     req.OvertimePayRate,
		WorkdayTypes:        req.WorkdayTypes,
		Color:               req.Color,
		IsActive:            req.IsActive,
		Sort:                req.Sort,
		CreatedBy:           req.CreatedBy,
		UpdatedBy:           req.CreatedBy,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.shiftRepo.Create(ctx, shift); err != nil {
		return nil, fmt.Errorf("create shift failed: %w", err)
	}

	return shift, nil
}

func (s *shiftService) GetByID(ctx context.Context, id uuid.UUID) (*model.Shift, error) {
	return s.shiftRepo.FindByID(ctx, id)
}

func (s *shiftService) Update(ctx context.Context, id uuid.UUID, req *UpdateShiftRequest) (*model.Shift, error) {
	shift, err := s.shiftRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Code != nil {
		shift.Code = *req.Code
	}
	if req.Name != nil {
		shift.Name = *req.Name
	}
	if req.Description != nil {
		shift.Description = *req.Description
	}
	if req.Type != nil {
		shift.Type = *req.Type
	}
	if req.WorkStart != nil {
		shift.WorkStart = *req.WorkStart
	}
	if req.WorkEnd != nil {
		shift.WorkEnd = *req.WorkEnd
	}
	if req.FlexibleStart != nil {
		shift.FlexibleStart = *req.FlexibleStart
	}
	if req.FlexibleEnd != nil {
		shift.FlexibleEnd = *req.FlexibleEnd
	}
	if req.WorkDuration != nil {
		shift.WorkDuration = *req.WorkDuration
	}
	if req.CheckInRequired != nil {
		shift.CheckInRequired = *req.CheckInRequired
	}
	if req.CheckOutRequired != nil {
		shift.CheckOutRequired = *req.CheckOutRequired
	}
	if req.LateGracePeriod != nil {
		shift.LateGracePeriod = *req.LateGracePeriod
	}
	if req.EarlyGracePeriod != nil {
		shift.EarlyGracePeriod = *req.EarlyGracePeriod
	}
	if req.RestPeriods != nil {
		shift.RestPeriods = *req.RestPeriods
	}
	if req.IsCrossDays != nil {
		shift.IsCrossDays = *req.IsCrossDays
	}
	if req.AllowOvertime != nil {
		shift.AllowOvertime = *req.AllowOvertime
	}
	if req.OvertimeStartBuffer != nil {
		shift.OvertimeStartBuffer = *req.OvertimeStartBuffer
	}
	if req.OvertimeMinDuration != nil {
		shift.OvertimeMinDuration = *req.OvertimeMinDuration
	}
	if req.OvertimePayRate != nil {
		shift.OvertimePayRate = *req.OvertimePayRate
	}
	if req.WorkdayTypes != nil {
		shift.WorkdayTypes = *req.WorkdayTypes
	}
	if req.Color != nil {
		shift.Color = *req.Color
	}
	if req.IsActive != nil {
		shift.IsActive = *req.IsActive
	}
	if req.Sort != nil {
		shift.Sort = *req.Sort
	}

	shift.UpdatedBy = req.UpdatedBy
	shift.UpdatedAt = time.Now()

	if err := s.shiftRepo.Update(ctx, shift); err != nil {
		return nil, fmt.Errorf("update shift failed: %w", err)
	}

	return shift, nil
}

func (s *shiftService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.shiftRepo.Delete(ctx, id)
}

func (s *shiftService) List(ctx context.Context, tenantID uuid.UUID, filter *repository.ShiftFilter, offset, limit int) ([]*model.Shift, int, error) {
	return s.shiftRepo.List(ctx, tenantID, filter, offset, limit)
}

func (s *shiftService) ListActive(ctx context.Context, tenantID uuid.UUID) ([]*model.Shift, error) {
	filter := &repository.ShiftFilter{
		IsActive: boolPtr(true),
	}
	shifts, _, err := s.shiftRepo.List(ctx, tenantID, filter, 0, 1000)
	return shifts, err
}

func boolPtr(b bool) *bool {
	return &b
}
