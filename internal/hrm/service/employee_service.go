package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/dto"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	orgservice "github.com/lk2023060901/go-next-erp/internal/organization/service"
)

// HRMEmployeeService HRM员工服务接口
type HRMEmployeeService interface {
	// Create 创建HRM员工扩展信息
	Create(ctx context.Context, req *dto.CreateHRMEmployeeRequest) (*model.HRMEmployee, error)

	// Update 更新HRM员工扩展信息
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateHRMEmployeeRequest) (*model.HRMEmployee, error)

	// Delete 删除HRM员工扩展信息
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByID 根据ID获取
	GetByID(ctx context.Context, id uuid.UUID) (*model.HRMEmployee, error)

	// GetByEmployeeID 根据组织员工ID获取
	GetByEmployeeID(ctx context.Context, tenantID, employeeID uuid.UUID) (*model.HRMEmployee, error)

	// GetWithBase 获取员工完整信息（组织信息 + HRM扩展）
	GetWithBase(ctx context.Context, tenantID, employeeID uuid.UUID) (*dto.EmployeeWithHRM, error)

	// GetAttendanceInfo 获取员工考勤信息
	GetAttendanceInfo(ctx context.Context, tenantID, employeeID uuid.UUID) (*dto.EmployeeAttendanceInfo, error)

	// List 列表查询
	List(ctx context.Context, tenantID uuid.UUID, filter *repository.HRMEmployeeFilter, offset, limit int) ([]*model.HRMEmployee, int, error)

	// ListWithBase 列表查询（包含基础员工信息）
	ListWithBase(ctx context.Context, tenantID uuid.UUID, filter *repository.HRMEmployeeFilter, offset, limit int) ([]*dto.EmployeeWithHRM, int, error)

	// UpdateFaceData 更新人脸数据
	UpdateFaceData(ctx context.Context, id uuid.UUID, faceData string) error

	// UpdateFingerprint 更新指纹数据
	UpdateFingerprint(ctx context.Context, id uuid.UUID, fingerprint string) error

	// UpdateCardNo 更新考勤卡号
	UpdateCardNo(ctx context.Context, id uuid.UUID, cardNo string) error

	// UpdateThirdPartyID 更新第三方平台ID
	UpdateThirdPartyID(ctx context.Context, id uuid.UUID, platform model.PlatformType, platformID string) error

	// SyncFromThirdParty 从第三方平台同步员工信息
	SyncFromThirdParty(ctx context.Context, tenantID uuid.UUID, platform model.PlatformType) error

	// InitializeForEmployee 为员工初始化HRM信息
	InitializeForEmployee(ctx context.Context, tenantID, employeeID, operatorID uuid.UUID) (*model.HRMEmployee, error)

	// BatchInitialize 批量初始化HRM信息
	BatchInitialize(ctx context.Context, tenantID uuid.UUID, employeeIDs []uuid.UUID, operatorID uuid.UUID) error

	// Activate 启用员工考勤
	Activate(ctx context.Context, id uuid.UUID) error

	// Deactivate 停用员工考勤
	Deactivate(ctx context.Context, id uuid.UUID) error
}

type hrmEmployeeService struct {
	hrmEmpRepo      repository.HRMEmployeeRepository
	syncMappingRepo repository.EmployeeSyncMappingRepository
	orgEmpService   orgservice.EmployeeService // 组织模块的员工服务
}

// NewHRMEmployeeService 创建HRM员工服务
func NewHRMEmployeeService(
	hrmEmpRepo repository.HRMEmployeeRepository,
	syncMappingRepo repository.EmployeeSyncMappingRepository,
	orgEmpService orgservice.EmployeeService,
) HRMEmployeeService {
	return &hrmEmployeeService{
		hrmEmpRepo:      hrmEmpRepo,
		syncMappingRepo: syncMappingRepo,
		orgEmpService:   orgEmpService,
	}
}

func (s *hrmEmployeeService) Create(ctx context.Context, req *dto.CreateHRMEmployeeRequest) (*model.HRMEmployee, error) {
	// 验证组织员工是否存在
	_, err := s.orgEmpService.GetByID(ctx, req.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	// 检查是否已存在HRM信息
	tenantID := getTenantIDFromContext(ctx)
	exists, err := s.hrmEmpRepo.ExistsByEmployeeID(ctx, tenantID, req.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("check hrm employee exists failed: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("hrm employee already exists for employee %s", req.EmployeeID)
	}

	// 如果提供了考勤卡号，检查唯一性
	if req.CardNo != "" {
		existing, err := s.hrmEmpRepo.FindByCardNo(ctx, tenantID, req.CardNo)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("card no '%s' already exists", req.CardNo)
		}
	}

	// 创建HRM员工
	hrmEmp := &model.HRMEmployee{
		ID:                uuid.New(),
		TenantID:          tenantID,
		EmployeeID:        req.EmployeeID,
		IDCardNo:          req.IDCardNo,
		CardNo:            req.CardNo,
		FaceData:          req.FaceData,
		Fingerprint:       req.Fingerprint,
		DingTalkUserID:    req.DingTalkUserID,
		WeComUserID:       req.WeComUserID,
		FeishuUserID:      req.FeishuUserID,
		FeishuOpenID:      req.FeishuOpenID,
		WorkLocation:      req.WorkLocation,
		WorkScheduleType:  req.WorkScheduleType,
		AttendanceRuleID:  req.AttendanceRuleID,
		DefaultShiftID:    req.DefaultShiftID,
		AllowFieldWork:    req.AllowFieldWork,
		RequireFace:       req.RequireFace,
		RequireLocation:   req.RequireLocation,
		RequireWiFi:       req.RequireWiFi,
		EmergencyContact:  req.EmergencyContact,
		EmergencyPhone:    req.EmergencyPhone,
		EmergencyRelation: req.EmergencyRelation,
		IsActive:          true, // 默认启用
		Remark:            req.Remark,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.hrmEmpRepo.Create(ctx, hrmEmp); err != nil {
		return nil, fmt.Errorf("create hrm employee failed: %w", err)
	}

	// 如果提供了第三方平台ID，创建同步映射
	if err := s.createSyncMappings(ctx, hrmEmp); err != nil {
		// 记录日志，但不阻断主流程
		// TODO: 添加日志
	}

	return hrmEmp, nil
}

func (s *hrmEmployeeService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateHRMEmployeeRequest) (*model.HRMEmployee, error) {
	// 获取现有记录
	hrmEmp, err := s.hrmEmpRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("hrm employee not found: %w", err)
	}

	// 如果更新考勤卡号，检查唯一性
	if req.CardNo != nil && *req.CardNo != hrmEmp.CardNo {
		tenantID := getTenantIDFromContext(ctx)
		existing, err := s.hrmEmpRepo.FindByCardNo(ctx, tenantID, *req.CardNo)
		if err == nil && existing != nil && existing.ID != id {
			return nil, fmt.Errorf("card no '%s' already exists", *req.CardNo)
		}
		hrmEmp.CardNo = *req.CardNo
	}

	// 更新字段
	if req.FaceData != nil {
		hrmEmp.FaceData = *req.FaceData
	}
	if req.Fingerprint != nil {
		hrmEmp.Fingerprint = *req.Fingerprint
	}
	if req.DingTalkUserID != nil {
		hrmEmp.DingTalkUserID = *req.DingTalkUserID
	}
	if req.WeComUserID != nil {
		hrmEmp.WeComUserID = *req.WeComUserID
	}
	if req.FeishuUserID != nil {
		hrmEmp.FeishuUserID = *req.FeishuUserID
	}
	if req.FeishuOpenID != nil {
		hrmEmp.FeishuOpenID = *req.FeishuOpenID
	}
	if req.WorkLocation != nil {
		hrmEmp.WorkLocation = *req.WorkLocation
	}
	if req.WorkScheduleType != nil {
		hrmEmp.WorkScheduleType = *req.WorkScheduleType
	}
	if req.AttendanceRuleID != nil {
		hrmEmp.AttendanceRuleID = req.AttendanceRuleID
	}
	if req.DefaultShiftID != nil {
		hrmEmp.DefaultShiftID = req.DefaultShiftID
	}
	if req.AllowFieldWork != nil {
		hrmEmp.AllowFieldWork = *req.AllowFieldWork
	}
	if req.RequireFace != nil {
		hrmEmp.RequireFace = *req.RequireFace
	}
	if req.RequireLocation != nil {
		hrmEmp.RequireLocation = *req.RequireLocation
	}
	if req.RequireWiFi != nil {
		hrmEmp.RequireWiFi = *req.RequireWiFi
	}
	if req.EmergencyContact != nil {
		hrmEmp.EmergencyContact = *req.EmergencyContact
	}
	if req.EmergencyPhone != nil {
		hrmEmp.EmergencyPhone = *req.EmergencyPhone
	}
	if req.EmergencyRelation != nil {
		hrmEmp.EmergencyRelation = *req.EmergencyRelation
	}
	if req.IsActive != nil {
		hrmEmp.IsActive = *req.IsActive
	}
	if req.Remark != nil {
		hrmEmp.Remark = *req.Remark
	}

	hrmEmp.UpdatedAt = time.Now()

	if err := s.hrmEmpRepo.Update(ctx, hrmEmp); err != nil {
		return nil, fmt.Errorf("update hrm employee failed: %w", err)
	}

	return hrmEmp, nil
}

func (s *hrmEmployeeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.hrmEmpRepo.Delete(ctx, id)
}

func (s *hrmEmployeeService) GetByID(ctx context.Context, id uuid.UUID) (*model.HRMEmployee, error) {
	return s.hrmEmpRepo.FindByID(ctx, id)
}

func (s *hrmEmployeeService) GetByEmployeeID(ctx context.Context, tenantID, employeeID uuid.UUID) (*model.HRMEmployee, error) {
	return s.hrmEmpRepo.FindByEmployeeID(ctx, tenantID, employeeID)
}

func (s *hrmEmployeeService) GetWithBase(ctx context.Context, tenantID, employeeID uuid.UUID) (*dto.EmployeeWithHRM, error) {
	// 获取基础员工信息
	baseEmp, err := s.orgEmpService.GetByID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("get base employee failed: %w", err)
	}

	// 获取HRM扩展信息
	hrmEmp, err := s.hrmEmpRepo.FindByEmployeeID(ctx, tenantID, employeeID)
	if err != nil {
		// HRM信息可能不存在，不报错
		hrmEmp = nil
	}

	return &dto.EmployeeWithHRM{
		Employee: baseEmp,
		HRMInfo:  hrmEmp,
	}, nil
}

func (s *hrmEmployeeService) GetAttendanceInfo(ctx context.Context, tenantID, employeeID uuid.UUID) (*dto.EmployeeAttendanceInfo, error) {
	// 获取基础员工信息
	baseEmp, err := s.orgEmpService.GetByID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("get base employee failed: %w", err)
	}

	// 获取HRM扩展信息
	hrmEmp, err := s.hrmEmpRepo.FindByEmployeeID(ctx, tenantID, employeeID)
	if err != nil {
		hrmEmp = nil
	}

	// 组装考勤信息
	info := &dto.EmployeeAttendanceInfo{
		EmployeeID: baseEmp.ID,
		EmployeeNo: baseEmp.EmployeeNo,
		Name:       baseEmp.Name,
		Gender:     baseEmp.Gender,
		Mobile:     baseEmp.Mobile,
		Avatar:     baseEmp.Avatar,
		OrgID:      baseEmp.OrgID,
		JoinDate:   baseEmp.JoinDate,
		Status:     baseEmp.Status,
	}

	if hrmEmp != nil {
		info.CardNo = hrmEmp.CardNo
		info.WorkLocation = hrmEmp.WorkLocation
		info.AttendanceRuleID = hrmEmp.AttendanceRuleID
		info.DefaultShiftID = hrmEmp.DefaultShiftID
		info.DingTalkUserID = hrmEmp.DingTalkUserID
		info.WeComUserID = hrmEmp.WeComUserID
		info.FeishuUserID = hrmEmp.FeishuUserID
		info.AllowFieldWork = hrmEmp.AllowFieldWork
		info.RequireFace = hrmEmp.RequireFace
		info.RequireLocation = hrmEmp.RequireLocation
		info.RequireWiFi = hrmEmp.RequireWiFi
		info.HasFaceData = hrmEmp.HasFaceData()
		info.HasFingerprint = hrmEmp.HasFingerprint()
		info.IsAttendanceActive = hrmEmp.IsActive
	}

	return info, nil
}

func (s *hrmEmployeeService) List(ctx context.Context, tenantID uuid.UUID, filter *repository.HRMEmployeeFilter, offset, limit int) ([]*model.HRMEmployee, int, error) {
	return s.hrmEmpRepo.List(ctx, tenantID, filter, offset, limit)
}

func (s *hrmEmployeeService) ListWithBase(ctx context.Context, tenantID uuid.UUID, filter *repository.HRMEmployeeFilter, offset, limit int) ([]*dto.EmployeeWithHRM, int, error) {
	// 获取HRM员工列表
	hrmEmps, total, err := s.hrmEmpRepo.List(ctx, tenantID, filter, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// 批量获取基础员工信息
	result := make([]*dto.EmployeeWithHRM, 0, len(hrmEmps))
	for _, hrmEmp := range hrmEmps {
		baseEmp, err := s.orgEmpService.GetByID(ctx, hrmEmp.EmployeeID)
		if err != nil {
			// 记录日志，跳过
			continue
		}

		result = append(result, &dto.EmployeeWithHRM{
			Employee: baseEmp,
			HRMInfo:  hrmEmp,
		})
	}

	return result, total, nil
}

func (s *hrmEmployeeService) UpdateFaceData(ctx context.Context, id uuid.UUID, faceData string) error {
	return s.hrmEmpRepo.UpdateFaceData(ctx, id, faceData)
}

func (s *hrmEmployeeService) UpdateFingerprint(ctx context.Context, id uuid.UUID, fingerprint string) error {
	return s.hrmEmpRepo.UpdateFingerprint(ctx, id, fingerprint)
}

func (s *hrmEmployeeService) UpdateCardNo(ctx context.Context, id uuid.UUID, cardNo string) error {
	// 检查卡号唯一性
	tenantID := getTenantIDFromContext(ctx)
	existing, err := s.hrmEmpRepo.FindByCardNo(ctx, tenantID, cardNo)
	if err == nil && existing != nil && existing.ID != id {
		return fmt.Errorf("card no '%s' already exists", cardNo)
	}

	return s.hrmEmpRepo.UpdateCardNo(ctx, id, cardNo)
}

func (s *hrmEmployeeService) UpdateThirdPartyID(ctx context.Context, id uuid.UUID, platform model.PlatformType, platformID string) error {
	return s.hrmEmpRepo.UpdateThirdPartyID(ctx, id, platform, platformID)
}

func (s *hrmEmployeeService) SyncFromThirdParty(ctx context.Context, tenantID uuid.UUID, platform model.PlatformType) error {
	// TODO: 实现第三方平台同步逻辑
	// 1. 调用对应平台的适配器拉取员工列表
	// 2. 匹配内部员工（通过手机号或工号）
	// 3. 创建或更新同步映射
	// 4. 更新HRM员工的第三方ID
	return fmt.Errorf("not implemented")
}

func (s *hrmEmployeeService) InitializeForEmployee(ctx context.Context, tenantID, employeeID, operatorID uuid.UUID) (*model.HRMEmployee, error) {
	// 检查是否已存在
	existing, err := s.hrmEmpRepo.FindByEmployeeID(ctx, tenantID, employeeID)
	if err == nil && existing != nil {
		return existing, nil // 已存在，直接返回
	}

	// 创建默认的HRM信息
	hrmEmp := &model.HRMEmployee{
		ID:         uuid.New(),
		TenantID:   tenantID,
		EmployeeID: employeeID,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.hrmEmpRepo.Create(ctx, hrmEmp); err != nil {
		return nil, fmt.Errorf("initialize hrm employee failed: %w", err)
	}

	return hrmEmp, nil
}

func (s *hrmEmployeeService) BatchInitialize(ctx context.Context, tenantID uuid.UUID, employeeIDs []uuid.UUID, operatorID uuid.UUID) error {
	hrmEmps := make([]*model.HRMEmployee, 0, len(employeeIDs))

	for _, empID := range employeeIDs {
		// 检查是否已存在
		exists, err := s.hrmEmpRepo.ExistsByEmployeeID(ctx, tenantID, empID)
		if err != nil {
			return err
		}
		if exists {
			continue // 已存在，跳过
		}

		hrmEmps = append(hrmEmps, &model.HRMEmployee{
			ID:         uuid.New(),
			TenantID:   tenantID,
			EmployeeID: empID,
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		})
	}

	if len(hrmEmps) > 0 {
		if err := s.hrmEmpRepo.BatchCreate(ctx, hrmEmps); err != nil {
			return fmt.Errorf("batch initialize hrm employees failed: %w", err)
		}
	}

	return nil
}

func (s *hrmEmployeeService) Activate(ctx context.Context, id uuid.UUID) error {
	hrmEmp, err := s.hrmEmpRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	hrmEmp.IsActive = true
	hrmEmp.UpdatedAt = time.Now()

	return s.hrmEmpRepo.Update(ctx, hrmEmp)
}

func (s *hrmEmployeeService) Deactivate(ctx context.Context, id uuid.UUID) error {
	hrmEmp, err := s.hrmEmpRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	hrmEmp.IsActive = false
	hrmEmp.UpdatedAt = time.Now()

	return s.hrmEmpRepo.Update(ctx, hrmEmp)
}

// createSyncMappings 创建第三方平台同步映射
func (s *hrmEmployeeService) createSyncMappings(ctx context.Context, hrmEmp *model.HRMEmployee) error {
	mappings := make([]*model.EmployeeSyncMapping, 0)

	if hrmEmp.DingTalkUserID != "" {
		mappings = append(mappings, &model.EmployeeSyncMapping{
			ID:          uuid.New(),
			TenantID:    hrmEmp.TenantID,
			EmployeeID:  hrmEmp.EmployeeID,
			Platform:    model.PlatformDingTalk,
			PlatformID:  hrmEmp.DingTalkUserID,
			SyncEnabled: true,
			SyncStatus:  "success",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	if hrmEmp.WeComUserID != "" {
		mappings = append(mappings, &model.EmployeeSyncMapping{
			ID:          uuid.New(),
			TenantID:    hrmEmp.TenantID,
			EmployeeID:  hrmEmp.EmployeeID,
			Platform:    model.PlatformWeCom,
			PlatformID:  hrmEmp.WeComUserID,
			SyncEnabled: true,
			SyncStatus:  "success",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	if hrmEmp.FeishuUserID != "" {
		mappings = append(mappings, &model.EmployeeSyncMapping{
			ID:          uuid.New(),
			TenantID:    hrmEmp.TenantID,
			EmployeeID:  hrmEmp.EmployeeID,
			Platform:    model.PlatformFeishu,
			PlatformID:  hrmEmp.FeishuUserID,
			SyncEnabled: true,
			SyncStatus:  "success",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	if len(mappings) > 0 {
		return s.syncMappingRepo.BatchCreate(ctx, mappings)
	}

	return nil
}

// getTenantIDFromContext 从上下文获取租户ID
func getTenantIDFromContext(ctx context.Context) uuid.UUID {
	// TODO: 从context中获取tenant_id
	// 这里暂时返回空UUID，实际应该从认证信息中获取
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return tenantID
}
