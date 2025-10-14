package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/integration"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
)

// AttendanceAdapter 飞书考勤适配器
type AttendanceAdapter struct {
	appID     string
	appSecret string
	baseURL   string
}

// NewAttendanceAdapter 创建飞书考勤适配器
func NewAttendanceAdapter(appID, appSecret string) integration.AttendanceIntegration {
	return &AttendanceAdapter{
		appID:     appID,
		appSecret: appSecret,
		baseURL:   "https://open.feishu.cn",
	}
}

// SyncAttendanceRecords 同步考勤记录
func (a *AttendanceAdapter) SyncAttendanceRecords(ctx context.Context, req *integration.SyncAttendanceRequest) ([]*model.AttendanceRecord, error) {
	// 1. 获取 Tenant Access Token
	token, err := a.getTenantAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant access token: %w", err)
	}

	// 2. 调用飞书考勤记录API
	records, err := a.fetchAttendanceRecords(ctx, token, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attendance records: %w", err)
	}

	// 3. 转换为系统内部格式
	result := make([]*model.AttendanceRecord, 0, len(records))
	for _, record := range records {
		converted, err := a.convertToAttendanceRecord(record, req.TenantID)
		if err != nil {
			continue
		}
		result = append(result, converted)
	}

	return result, nil
}

// SyncEmployees 同步员工信息
func (a *AttendanceAdapter) SyncEmployees(ctx context.Context, req *integration.SyncEmployeeRequest) ([]*model.EmployeeSyncMapping, error) {
	// 1. 获取 Tenant Access Token
	token, err := a.getTenantAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant access token: %w", err)
	}

	// 2. 获取部门用户列表
	users, err := a.fetchDepartmentUsers(ctx, token, req.DepartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	// 3. 转换为员工同步映射
	result := make([]*model.EmployeeSyncMapping, 0, len(users))
	for _, user := range users {
		mapping := &model.EmployeeSyncMapping{
			TenantID:    req.TenantID,
			Platform:    model.PlatformFeishu,
			PlatformID:  user.UserID,
			SyncEnabled: true,
			SyncStatus:  "success",
			RawData:     user.toMap(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		result = append(result, mapping)
	}

	return result, nil
}

// PushAttendanceRule 推送考勤规则
func (a *AttendanceAdapter) PushAttendanceRule(ctx context.Context, req *integration.PushAttendanceRuleRequest) error {
	// 飞书支持通过API创建考勤组
	// 1. 获取 Tenant Access Token
	token, err := a.getTenantAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant access token: %w", err)
	}

	// 2. 构造考勤组参数
	// TODO: 实现实际的 HTTP 请求到 /open-apis/attendance/v1/groups
	_ = token

	return nil
}

// GetEmployeeAttendance 获取员工考勤数据
func (a *AttendanceAdapter) GetEmployeeAttendance(ctx context.Context, req *integration.GetEmployeeAttendanceRequest) ([]*model.AttendanceRecord, error) {
	// 1. 获取 Tenant Access Token
	token, err := a.getTenantAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant access token: %w", err)
	}

	// 2. 构造请求参数
	syncReq := &integration.SyncAttendanceRequest{
		TenantID:   req.TenantID,
		EmployeeID: req.EmployeeID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
	}

	// 3. 调用飞书API获取考勤记录
	records, err := a.fetchAttendanceRecords(ctx, token, syncReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attendance records: %w", err)
	}

	// 4. 转换为系统内部格式
	result := make([]*model.AttendanceRecord, 0, len(records))
	for _, record := range records {
		converted, err := a.convertToAttendanceRecord(record, req.TenantID)
		if err != nil {
			continue
		}
		result = append(result, converted)
	}

	return result, nil
}

// ========== 内部辅助方法 ==========

// getTenantAccessToken 获取飞书 Tenant Access Token
func (a *AttendanceAdapter) getTenantAccessToken(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/open-apis/auth/v3/tenant_access_token/internal", a.baseURL)

	// TODO: 实现实际的 HTTP 请求
	_ = url // 避免未使用变量警告
	return "mock_tenant_access_token", nil
}

// fetchAttendanceRecords 获取飞书考勤记录
func (a *AttendanceAdapter) fetchAttendanceRecords(ctx context.Context, token string, req *integration.SyncAttendanceRequest) ([]*FeishuAttendanceRecord, error) {
	// 飞书API: /open-apis/attendance/v1/user_flows/query
	// TODO: 实现实际的 HTTP 请求

	return []*FeishuAttendanceRecord{}, nil
}

// fetchDepartmentUsers 获取部门用户列表
func (a *AttendanceAdapter) fetchDepartmentUsers(ctx context.Context, token string, deptID string) ([]*FeishuUser, error) {
	// 飞书API: /open-apis/contact/v3/users
	// TODO: 实现实际的 HTTP 请求

	return []*FeishuUser{}, nil
}

// convertToAttendanceRecord 转换飞书考勤记录为系统格式
func (a *AttendanceAdapter) convertToAttendanceRecord(record *FeishuAttendanceRecord, tenantID uuid.UUID) (*model.AttendanceRecord, error) {
	// 解析打卡类型
	var clockType model.AttendanceClockType
	switch record.CheckType {
	case "OnDuty":
		clockType = model.ClockTypeCheckIn
	case "OffDuty":
		clockType = model.ClockTypeCheckOut
	default:
		clockType = model.ClockTypeCheckIn
	}

	// 解析考勤状态
	var status model.AttendanceStatus
	switch record.Result {
	case "Normal":
		status = model.AttendanceStatusNormal
	case "Late":
		status = model.AttendanceStatusLate
	case "Early":
		status = model.AttendanceStatusEarly
	case "Lack": // 缺卡
		status = model.AttendanceStatusAbsent
	case "SeriousLate": // 严重迟到
		status = model.AttendanceStatusAbsent
	default:
		status = model.AttendanceStatusNormal
	}

	// 解析打卡方式
	var method model.AttendanceMethod
	switch record.CheckInMethod {
	case "WiFi":
		method = model.MethodMobile
	case "GPS":
		method = model.MethodMobile
	case "Face":
		method = model.MethodFace
	case "Device":
		method = model.MethodDevice
	default:
		method = model.MethodMobile
	}

	// 构造考勤记录
	result := &model.AttendanceRecord{
		TenantID:      tenantID,
		ClockTime:     time.Unix(record.CheckTime, 0),
		ClockType:     clockType,
		Status:        status,
		CheckInMethod: method,
		SourceType:    model.SourceTypeFeishu,
		SourceID:      record.RecordID,
		IsException:   record.Result != "Normal",
		ExceptionType: record.Result,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 处理位置信息
	if record.Latitude != 0 && record.Longitude != 0 {
		result.Location = &model.LocationInfo{
			Latitude:  record.Latitude,
			Longitude: record.Longitude,
		}
		result.Address = record.LocationName
	}

	// WiFi 信息
	if record.WiFiSSID != "" {
		result.WiFiSSID = record.WiFiSSID
		result.WiFiMAC = record.WiFiBSSID
	}

	// 照片信息
	if len(record.PhotoURLs) > 0 {
		result.PhotoURL = record.PhotoURLs[0]
	}

	// 存储原始数据
	rawData := make(map[string]interface{})
	data, _ := json.Marshal(record)
	json.Unmarshal(data, &rawData)
	result.RawData = rawData

	return result, nil
}

// ========== 飞书数据结构 ==========

// FeishuAttendanceRecord 飞书考勤记录
type FeishuAttendanceRecord struct {
	RecordID        string   `json:"record_id"`
	UserID          string   `json:"user_id"`
	CreatorID       string   `json:"creator_id"`
	LocationName    string   `json:"location_name"`
	CheckTime       int64    `json:"check_time"`
	Comment         string   `json:"comment"`
	RecordType      string   `json:"record_type"`
	CheckType       string   `json:"check_type"`      // OnDuty, OffDuty
	Result          string   `json:"result"`          // Normal, Late, Early, Lack, SeriousLate
	CheckInMethod   string   `json:"check_in_method"` // WiFi, GPS, Face, Device
	PhotoURLs       []string `json:"photo_urls"`
	Latitude        float64  `json:"latitude"`
	Longitude       float64  `json:"longitude"`
	WiFiSSID        string   `json:"ssid"`
	WiFiBSSID       string   `json:"bssid"`
	IsOutOfRange    bool     `json:"is_out_of_range"`
	DeviceSN        string   `json:"device_sn"`
	ShiftID         string   `json:"shift_id"`
	GroupID         string   `json:"group_id"`
	TimeZone        string   `json:"time_zone"`
	CheckTimeFormat string   `json:"check_time_format"`
}

// FeishuUser 飞书用户
type FeishuUser struct {
	UserID        string   `json:"user_id"`
	OpenID        string   `json:"open_id"`
	UnionID       string   `json:"union_id"`
	Name          string   `json:"name"`
	EnName        string   `json:"en_name"`
	Nickname      string   `json:"nickname"`
	Email         string   `json:"email"`
	Mobile        string   `json:"mobile"`
	Gender        int      `json:"gender"`
	Avatar        Avatar   `json:"avatar"`
	Status        Status   `json:"status"`
	DepartmentIDs []string `json:"department_ids"`
	LeaderUserID  string   `json:"leader_user_id"`
	City          string   `json:"city"`
	Country       string   `json:"country"`
	WorkStation   string   `json:"work_station"`
	JoinTime      int64    `json:"join_time"`
	EmployeeNo    string   `json:"employee_no"`
	EmployeeType  int      `json:"employee_type"`
	Orders        []Order  `json:"orders"`
	JobTitle      string   `json:"job_title"`
}

// Avatar 头像
type Avatar struct {
	Avatar72     string `json:"avatar_72"`
	Avatar240    string `json:"avatar_240"`
	Avatar640    string `json:"avatar_640"`
	AvatarOrigin string `json:"avatar_origin"`
}

// Status 用户状态
type Status struct {
	IsFrozen    bool `json:"is_frozen"`
	IsResigned  bool `json:"is_resigned"`
	IsActivated bool `json:"is_activated"`
	IsExited    bool `json:"is_exited"`
	IsUnjoin    bool `json:"is_unjoin"`
}

// Order 用户部门排序
type Order struct {
	DepartmentID    string `json:"department_id"`
	UserOrder       int    `json:"user_order"`
	DepartmentOrder int    `json:"department_order"`
}

// toMap 转换为 map
func (u *FeishuUser) toMap() map[string]interface{} {
	data := make(map[string]interface{})
	bytes, _ := json.Marshal(u)
	json.Unmarshal(bytes, &data)
	return data
}
