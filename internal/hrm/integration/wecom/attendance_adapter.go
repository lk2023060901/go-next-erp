package wecom

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/integration"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
)

// AttendanceAdapter 企业微信考勤适配器
type AttendanceAdapter struct {
	corpID     string
	corpSecret string
	baseURL    string
}

// NewAttendanceAdapter 创建企业微信考勤适配器
func NewAttendanceAdapter(corpID, corpSecret string) integration.AttendanceIntegration {
	return &AttendanceAdapter{
		corpID:     corpID,
		corpSecret: corpSecret,
		baseURL:    "https://qyapi.weixin.qq.com",
	}
}

// SyncAttendanceRecords 同步考勤记录
func (a *AttendanceAdapter) SyncAttendanceRecords(ctx context.Context, req *integration.SyncAttendanceRequest) ([]*model.AttendanceRecord, error) {
	// 1. 获取 Access Token
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// 2. 调用企业微信考勤记录API
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
	// 1. 获取 Access Token
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
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
			Platform:    model.PlatformWeCom,
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
	// 企业微信不支持推送考勤规则,只能在企业微信后台配置
	return fmt.Errorf("wecom does not support pushing attendance rules")
}

// GetEmployeeAttendance 获取员工考勤数据
func (a *AttendanceAdapter) GetEmployeeAttendance(ctx context.Context, req *integration.GetEmployeeAttendanceRequest) ([]*model.AttendanceRecord, error) {
	// 1. 获取 Access Token
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// 2. 构造请求参数
	syncReq := &integration.SyncAttendanceRequest{
		TenantID:   req.TenantID,
		EmployeeID: req.EmployeeID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
	}

	// 3. 调用企业微信API获取考勤记录
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

// getAccessToken 获取企业微信 Access Token
func (a *AttendanceAdapter) getAccessToken(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/cgi-bin/gettoken?corpid=%s&corpsecret=%s", a.baseURL, a.corpID, a.corpSecret)

	// TODO: 实现实际的 HTTP 请求
	_ = url // 避免未使用变量警告
	return "mock_access_token", nil
}

// fetchAttendanceRecords 获取企业微信考勤记录
func (a *AttendanceAdapter) fetchAttendanceRecords(ctx context.Context, token string, req *integration.SyncAttendanceRequest) ([]*WeComAttendanceRecord, error) {
	// 企业微信API: /cgi-bin/checkin/getcheckindata
	// TODO: 实现实际的 HTTP 请求

	return []*WeComAttendanceRecord{}, nil
}

// fetchDepartmentUsers 获取部门用户列表
func (a *AttendanceAdapter) fetchDepartmentUsers(ctx context.Context, token string, deptID string) ([]*WeComUser, error) {
	// 企业微信API: /cgi-bin/user/list
	// TODO: 实现实际的 HTTP 请求

	return []*WeComUser{}, nil
}

// convertToAttendanceRecord 转换企业微信考勤记录为系统格式
func (a *AttendanceAdapter) convertToAttendanceRecord(record *WeComAttendanceRecord, tenantID uuid.UUID) (*model.AttendanceRecord, error) {
	// 解析打卡类型 (企业微信没有明确的上下班类型,根据时间判断)
	var clockType model.AttendanceClockType
	hour := record.CheckinTime.Hour()
	if hour < 12 {
		clockType = model.ClockTypeCheckIn
	} else {
		clockType = model.ClockTypeCheckOut
	}

	// 解析考勤状态
	var status model.AttendanceStatus
	switch record.ExceptionType {
	case "":
		status = model.AttendanceStatusNormal
	case "时间异常":
		if hour < 12 {
			status = model.AttendanceStatusLate
		} else {
			status = model.AttendanceStatusEarly
		}
	case "地点异常":
		status = model.AttendanceStatusNormal // 地点异常单独标记
	default:
		status = model.AttendanceStatusNormal
	}

	// 解析打卡方式
	var method model.AttendanceMethod
	switch record.CheckinType {
	case "WiFi打卡":
		method = model.MethodMobile
	case "GPS打卡":
		method = model.MethodMobile
	case "摄像头打卡":
		method = model.MethodFace
	default:
		method = model.MethodMobile
	}

	// 构造考勤记录
	result := &model.AttendanceRecord{
		TenantID:      tenantID,
		ClockTime:     record.CheckinTime,
		ClockType:     clockType,
		Status:        status,
		CheckInMethod: method,
		SourceType:    model.SourceTypeWeCom,
		SourceID:      fmt.Sprintf("%s_%d", record.UserID, record.CheckinTime.Unix()),
		IsException:   record.ExceptionType != "",
		ExceptionType: record.ExceptionType,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 处理位置信息
	if record.Lat != 0 && record.Lng != 0 {
		result.Location = &model.LocationInfo{
			Latitude:  record.Lat,
			Longitude: record.Lng,
		}
		result.Address = record.LocationTitle
	}

	// WiFi 信息
	if record.WiFiName != "" {
		result.WiFiSSID = record.WiFiName
		result.WiFiMAC = record.WiFiMAC
	}

	// 存储原始数据
	rawData := make(map[string]interface{})
	data, _ := json.Marshal(record)
	json.Unmarshal(data, &rawData)
	result.RawData = rawData

	return result, nil
}

// ========== 企业微信数据结构 ==========

// WeComAttendanceRecord 企业微信考勤记录
type WeComAttendanceRecord struct {
	UserID         string    `json:"userid"`
	GroupName      string    `json:"groupname"`
	CheckinType    string    `json:"checkin_type"`   // WiFi打卡、GPS打卡、摄像头打卡
	ExceptionType  string    `json:"exception_type"` // 时间异常、地点异常
	CheckinTime    time.Time `json:"checkin_time"`
	LocationTitle  string    `json:"location_title"`
	LocationDetail string    `json:"location_detail"`
	WiFiName       string    `json:"wifiname"`
	WiFiMAC        string    `json:"wifimac"`
	Notes          string    `json:"notes"`
	MediaIDs       []string  `json:"mediaids"`
	Lat            float64   `json:"lat"`
	Lng            float64   `json:"lng"`
}

// WeComUser 企业微信用户
type WeComUser struct {
	UserID     string  `json:"userid"`
	Name       string  `json:"name"`
	Mobile     string  `json:"mobile"`
	Email      string  `json:"email"`
	Department []int64 `json:"department"`
	Position   string  `json:"position"`
	Status     int     `json:"status"`
	Enable     int     `json:"enable"`
}

// toMap 转换为 map
func (u *WeComUser) toMap() map[string]interface{} {
	data := make(map[string]interface{})
	bytes, _ := json.Marshal(u)
	json.Unmarshal(bytes, &data)
	return data
}
