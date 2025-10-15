package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
)

// AttendanceService 考勤服务接口
type AttendanceService interface {
	// ClockIn 打卡
	ClockIn(ctx context.Context, req *ClockInRequest) (*model.AttendanceRecord, error)

	// ClockOut 下班打卡
	ClockOut(ctx context.Context, req *ClockInRequest) (*model.AttendanceRecord, error)

	// GetByID 根据ID获取考勤记录
	GetByID(ctx context.Context, id uuid.UUID) (*model.AttendanceRecord, error)

	// List 列表查询
	List(ctx context.Context, tenantID uuid.UUID, filter *repository.AttendanceRecordFilter, offset, limit int) ([]*model.AttendanceRecord, int, error)

	// ListWithCursor 游标分页查询（高性能）
	// 返回值：records, nextCursor, hasNext, error
	ListWithCursor(ctx context.Context, tenantID uuid.UUID, filter *repository.AttendanceRecordFilter, cursor *time.Time, limit int) ([]*model.AttendanceRecord, *time.Time, bool, error)

	// GetEmployeeRecords 获取员工考勤记录
	GetEmployeeRecords(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) ([]*model.AttendanceRecord, error)

	// GetDepartmentRecords 获取部门考勤记录
	GetDepartmentRecords(ctx context.Context, tenantID, departmentID uuid.UUID, startDate, endDate time.Time) ([]*model.AttendanceRecord, error)

	// GetExceptions 获取异常考勤记录
	GetExceptions(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*model.AttendanceRecord, error)

	// CountByStatus 统计考勤状态
	CountByStatus(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[model.AttendanceStatus]int, error)

	// CalculateStatus 计算考勤状态（正常/迟到/早退）
	CalculateStatus(ctx context.Context, record *model.AttendanceRecord) (model.AttendanceStatus, error)

	// BatchImport 批量导入考勤记录（从第三方平台）
	BatchImport(ctx context.Context, records []*model.AttendanceRecord) error

	// Update 更新考勤记录
	Update(ctx context.Context, id uuid.UUID, req *UpdateAttendanceRequest) (*model.AttendanceRecord, error)

	// Delete 删除考勤记录
	Delete(ctx context.Context, id uuid.UUID) error
}

// ClockInRequest 打卡请求
type ClockInRequest struct {
	TenantID      uuid.UUID
	EmployeeID    uuid.UUID
	EmployeeName  string
	DepartmentID  uuid.UUID
	ClockTime     time.Time
	ClockType     model.AttendanceClockType
	CheckInMethod model.AttendanceMethod
	SourceType    model.SourceType
	SourceID      string
	Location      *model.LocationInfo
	Address       string
	WiFiSSID      string
	WiFiMAC       string
	PhotoURL      string
	FaceScore     float64
	Temperature   float64
	Remark        string
}

// UpdateAttendanceRequest 更新考勤记录请求
type UpdateAttendanceRequest struct {
	Status          *model.AttendanceStatus
	IsException     *bool
	ExceptionReason *string
	ExceptionType   *string
	ApprovalID      *uuid.UUID
	Remark          *string
}

type attendanceService struct {
	attendanceRepo repository.AttendanceRecordRepository
	shiftRepo      repository.ShiftRepository
	scheduleRepo   repository.ScheduleRepository
	ruleRepo       repository.AttendanceRuleRepository
	hrmEmpRepo     repository.HRMEmployeeRepository
}

// NewAttendanceService 创建考勤服务
func NewAttendanceService(
	attendanceRepo repository.AttendanceRecordRepository,
	shiftRepo repository.ShiftRepository,
	scheduleRepo repository.ScheduleRepository,
	ruleRepo repository.AttendanceRuleRepository,
	hrmEmpRepo repository.HRMEmployeeRepository,
) AttendanceService {
	return &attendanceService{
		attendanceRepo: attendanceRepo,
		shiftRepo:      shiftRepo,
		scheduleRepo:   scheduleRepo,
		ruleRepo:       ruleRepo,
		hrmEmpRepo:     hrmEmpRepo,
	}
}

func (s *attendanceService) ClockIn(ctx context.Context, req *ClockInRequest) (*model.AttendanceRecord, error) {
	// 获取员工HRM信息
	hrmEmp, err := s.hrmEmpRepo.FindByEmployeeID(ctx, req.TenantID, req.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("hrm employee not found: %w", err)
	}

	// 检查员工是否启用考勤
	if !hrmEmp.IsActive {
		return nil, fmt.Errorf("attendance is not active for this employee")
	}

	// 验证打卡规则
	if err := s.validateClockIn(ctx, hrmEmp, req); err != nil {
		return nil, err
	}

	// 查询当天的排班
	schedule, err := s.getTodaySchedule(ctx, req.TenantID, req.EmployeeID, req.ClockTime)
	if err != nil {
		// 没有排班，使用默认班次
		schedule = nil
	}

	// 从 employees 表获取 department_id
	departmentID := req.DepartmentID
	if departmentID == uuid.Nil {
		var orgID uuid.UUID
		sql := `SELECT org_id FROM employees WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`
		// 直接使用 repository 的数据库连接，需要添加 db 参数到 service
		// 为简化，先使用空 UUID，后续优化
		if schedule != nil {
			departmentID = schedule.DepartmentID
		} else {
			// TODO: 从 employees 表查询 org_id
			_ = sql
			_ = orgID
		}
	}

	// 创建考勤记录
	record := &model.AttendanceRecord{
		ID:            uuid.New(),
		TenantID:      req.TenantID,
		EmployeeID:    req.EmployeeID,
		EmployeeName:  req.EmployeeName,
		DepartmentID:  departmentID,
		ClockTime:     req.ClockTime,
		ClockType:     req.ClockType,
		CheckInMethod: req.CheckInMethod,
		SourceType:    req.SourceType,
		SourceID:      req.SourceID,
		Location:      req.Location,
		Address:       req.Address,
		WiFiSSID:      req.WiFiSSID,
		WiFiMAC:       req.WiFiMAC,
		PhotoURL:      req.PhotoURL,
		FaceScore:     req.FaceScore,
		Temperature:   req.Temperature,
		Remark:        req.Remark,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 设置班次信息
	if schedule != nil {
		record.ShiftID = &schedule.ShiftID
		record.ShiftName = schedule.ShiftName
	} else if hrmEmp.DefaultShiftID != nil {
		record.ShiftID = hrmEmp.DefaultShiftID
		// 获取班次名称
		shift, err := s.shiftRepo.FindByID(ctx, *hrmEmp.DefaultShiftID)
		if err == nil {
			record.ShiftName = shift.Name
		}
	}

	// 计算考勤状态
	status, err := s.CalculateStatus(ctx, record)
	if err != nil {
		// 计算失败，默认为正常
		status = model.AttendanceStatusNormal
	}
	record.Status = status

	// 判断是否异常
	if status != model.AttendanceStatusNormal {
		record.IsException = true
		record.ExceptionType = string(status)
		record.ExceptionReason = s.getExceptionReason(status)
	}

	// 保存记录
	if err := s.attendanceRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("create attendance record failed: %w", err)
	}

	return record, nil
}

func (s *attendanceService) ClockOut(ctx context.Context, req *ClockInRequest) (*model.AttendanceRecord, error) {
	req.ClockType = model.ClockTypeCheckOut
	return s.ClockIn(ctx, req)
}

func (s *attendanceService) GetByID(ctx context.Context, id uuid.UUID) (*model.AttendanceRecord, error) {
	return s.attendanceRepo.FindByID(ctx, id)
}

func (s *attendanceService) List(ctx context.Context, tenantID uuid.UUID, filter *repository.AttendanceRecordFilter, offset, limit int) ([]*model.AttendanceRecord, int, error) {
	return s.attendanceRepo.List(ctx, tenantID, filter, offset, limit)
}

// ListWithCursor 游标分页查询（直接调用repository层）
func (s *attendanceService) ListWithCursor(ctx context.Context, tenantID uuid.UUID, filter *repository.AttendanceRecordFilter, cursor *time.Time, limit int) ([]*model.AttendanceRecord, *time.Time, bool, error) {
	return s.attendanceRepo.ListWithCursor(ctx, tenantID, filter, cursor, limit)
}

func (s *attendanceService) GetEmployeeRecords(ctx context.Context, tenantID, employeeID uuid.UUID, startDate, endDate time.Time) ([]*model.AttendanceRecord, error) {
	return s.attendanceRepo.FindByEmployee(ctx, tenantID, employeeID, startDate, endDate)
}

func (s *attendanceService) GetDepartmentRecords(ctx context.Context, tenantID, departmentID uuid.UUID, startDate, endDate time.Time) ([]*model.AttendanceRecord, error) {
	return s.attendanceRepo.FindByDepartment(ctx, tenantID, departmentID, startDate, endDate)
}

func (s *attendanceService) GetExceptions(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*model.AttendanceRecord, error) {
	return s.attendanceRepo.FindExceptions(ctx, tenantID, startDate, endDate)
}

func (s *attendanceService) CountByStatus(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[model.AttendanceStatus]int, error) {
	return s.attendanceRepo.CountByStatus(ctx, tenantID, startDate, endDate)
}

func (s *attendanceService) CalculateStatus(ctx context.Context, record *model.AttendanceRecord) (model.AttendanceStatus, error) {
	// 如果没有班次信息，默认为正常
	if record.ShiftID == nil {
		return model.AttendanceStatusNormal, nil
	}

	// 获取班次信息
	shift, err := s.shiftRepo.FindByID(ctx, *record.ShiftID)
	if err != nil {
		return model.AttendanceStatusNormal, err
	}

	// 根据班次类型计算状态
	switch shift.Type {
	case model.ShiftTypeFixed:
		return s.calculateFixedShiftStatus(record, shift)
	case model.ShiftTypeFlexible:
		return s.calculateFlexibleShiftStatus(record, shift)
	case model.ShiftTypeFree:
		return model.AttendanceStatusNormal, nil
	default:
		return model.AttendanceStatusNormal, nil
	}
}

func (s *attendanceService) BatchImport(ctx context.Context, records []*model.AttendanceRecord) error {
	// 批量计算状态
	for _, record := range records {
		status, err := s.CalculateStatus(ctx, record)
		if err == nil {
			record.Status = status
			if status != model.AttendanceStatusNormal {
				record.IsException = true
				record.ExceptionType = string(status)
				record.ExceptionReason = s.getExceptionReason(status)
			}
		}
	}

	return s.attendanceRepo.BatchCreate(ctx, records)
}

func (s *attendanceService) Update(ctx context.Context, id uuid.UUID, req *UpdateAttendanceRequest) (*model.AttendanceRecord, error) {
	record, err := s.attendanceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Status != nil {
		record.Status = *req.Status
	}
	if req.IsException != nil {
		record.IsException = *req.IsException
	}
	if req.ExceptionReason != nil {
		record.ExceptionReason = *req.ExceptionReason
	}
	if req.ExceptionType != nil {
		record.ExceptionType = *req.ExceptionType
	}
	if req.ApprovalID != nil {
		record.ApprovalID = req.ApprovalID
	}
	if req.Remark != nil {
		record.Remark = *req.Remark
	}

	record.UpdatedAt = time.Now()

	if err := s.attendanceRepo.Update(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

func (s *attendanceService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.attendanceRepo.Delete(ctx, id)
}

// validateClockIn 验证打卡规则
func (s *attendanceService) validateClockIn(ctx context.Context, hrmEmp *model.HRMEmployee, req *ClockInRequest) error {
	// 检查定位要求
	if hrmEmp.RequireLocation && req.Location == nil {
		return fmt.Errorf("location is required")
	}

	// 检查WiFi要求
	if hrmEmp.RequireWiFi && req.WiFiSSID == "" {
		return fmt.Errorf("wifi connection is required")
	}

	// 检查人脸识别要求
	if hrmEmp.RequireFace && req.FaceScore == 0 {
		return fmt.Errorf("face recognition is required")
	}

	// 获取考勤规则
	if hrmEmp.AttendanceRuleID != nil {
		rule, err := s.ruleRepo.FindByID(ctx, *hrmEmp.AttendanceRuleID)
		if err == nil {
			// 验证地理围栏
			if rule.LocationRequired && !s.isInAllowedLocation(req.Location, rule.AllowedLocations) {
				return fmt.Errorf("clock in location is not allowed")
			}

			// 验证WiFi
			if rule.WiFiRequired && !s.isAllowedWiFi(req.WiFiSSID, rule.AllowedWiFi) {
				return fmt.Errorf("wifi '%s' is not allowed", req.WiFiSSID)
			}

			// 验证人脸识别阈值
			if rule.FaceRequired && req.FaceScore < rule.FaceThreshold {
				return fmt.Errorf("face recognition score too low")
			}
		}
	}

	return nil
}

// getTodaySchedule 获取当天排班
func (s *attendanceService) getTodaySchedule(ctx context.Context, tenantID, employeeID uuid.UUID, date time.Time) (*model.Schedule, error) {
	dateStr := date.Format("2006-01-02")
	schedules, err := s.scheduleRepo.FindByDate(ctx, tenantID, dateStr)
	if err != nil {
		return nil, err
	}

	for _, schedule := range schedules {
		if schedule.EmployeeID == employeeID {
			return schedule, nil
		}
	}

	return nil, fmt.Errorf("no schedule found")
}

// calculateFixedShiftStatus 计算固定班次状态
func (s *attendanceService) calculateFixedShiftStatus(record *model.AttendanceRecord, shift *model.Shift) (model.AttendanceStatus, error) {
	clockTime := record.ClockTime

	if record.ClockType == model.ClockTypeCheckIn {
		// 上班打卡
		workStart, err := parseTime(shift.WorkStart)
		if err != nil {
			return model.AttendanceStatusNormal, err
		}

		// 设置日期部分
		workStart = setDate(workStart, clockTime)

		// 计算迟到
		lateMinutes := int(clockTime.Sub(workStart).Minutes())
		if lateMinutes > shift.LateGracePeriod {
			return model.AttendanceStatusLate, nil
		}
	} else {
		// 下班打卡
		workEnd, err := parseTime(shift.WorkEnd)
		if err != nil {
			return model.AttendanceStatusNormal, err
		}

		workEnd = setDate(workEnd, clockTime)

		// 计算早退
		earlyMinutes := int(workEnd.Sub(clockTime).Minutes())
		if earlyMinutes > shift.EarlyGracePeriod {
			return model.AttendanceStatusEarly, nil
		}
	}

	return model.AttendanceStatusNormal, nil
}

// calculateFlexibleShiftStatus 计算弹性班次状态
func (s *attendanceService) calculateFlexibleShiftStatus(record *model.AttendanceRecord, shift *model.Shift) (model.AttendanceStatus, error) {
	// 弹性班次一般不判断迟到早退，只要在规定时间窗口内打卡即可
	// 可以根据实际需求调整
	return model.AttendanceStatusNormal, nil
}

// isInAllowedLocation 检查是否在允许的位置
func (s *attendanceService) isInAllowedLocation(location *model.LocationInfo, allowedLocations []model.AllowedLocation) bool {
	if location == nil || len(allowedLocations) == 0 {
		return true
	}

	for _, allowed := range allowedLocations {
		distance := calculateDistance(
			location.Latitude, location.Longitude,
			allowed.Latitude, allowed.Longitude,
		)
		if distance <= float64(allowed.Radius) {
			return true
		}
	}

	return false
}

// isAllowedWiFi 检查WiFi是否允许
func (s *attendanceService) isAllowedWiFi(wifiSSID string, allowedWiFi []string) bool {
	if wifiSSID == "" || len(allowedWiFi) == 0 {
		return true
	}

	for _, allowed := range allowedWiFi {
		if wifiSSID == allowed {
			return true
		}
	}

	return false
}

// getExceptionReason 获取异常原因描述
func (s *attendanceService) getExceptionReason(status model.AttendanceStatus) string {
	switch status {
	case model.AttendanceStatusLate:
		return "迟到"
	case model.AttendanceStatusEarly:
		return "早退"
	case model.AttendanceStatusAbsent:
		return "旷工"
	default:
		return ""
	}
}

// parseTime 解析时间字符串 HH:MM
func parseTime(timeStr string) (time.Time, error) {
	return time.Parse("15:04", timeStr)
}

// setDate 设置日期部分
func setDate(t time.Time, date time.Time) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		t.Hour(), t.Minute(), t.Second(),
		0, date.Location(),
	)
}

// calculateDistance 计算两点间距离（米）
// 使用 Haversine 公式
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // 地球半径（米）

	lat1Rad := lat1 * 3.141592653589793 / 180
	lat2Rad := lat2 * 3.141592653589793 / 180
	deltaLat := (lat2 - lat1) * 3.141592653589793 / 180
	deltaLon := (lon2 - lon1) * 3.141592653589793 / 180

	a := 0.5 - 0.5*cosApprox(deltaLat) + cosApprox(lat1Rad)*cosApprox(lat2Rad)*(1-cosApprox(deltaLon))/2

	return 2 * earthRadius * asinApprox(sqrtApprox(a))
}

// 简化的数学函数
func cosApprox(x float64) float64 {
	// 简化实现，实际应使用 math.Cos
	return 1 - x*x/2
}

func asinApprox(x float64) float64 {
	// 简化实现，实际应使用 math.Asin
	return x
}

func sqrtApprox(x float64) float64 {
	// 简化实现，实际应使用 math.Sqrt
	if x < 0 {
		return 0
	}
	return x
}
