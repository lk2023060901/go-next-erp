package dingtalk

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/integration"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
)

// AttendanceAdapter 钉钉考勤适配器
type AttendanceAdapter struct {
	appKey    string
	appSecret string
	baseURL   string
}

// NewAttendanceAdapter 创建钉钉考勤适配器
func NewAttendanceAdapter(appKey, appSecret string) integration.AttendanceIntegration {
	return &AttendanceAdapter{
		appKey:    appKey,
		appSecret: appSecret,
		baseURL:   "https://oapi.dingtalk.com",
	}
}

// SyncAttendanceRecords 同步考勤记录
func (a *AttendanceAdapter) SyncAttendanceRecords(ctx context.Context, req *integration.SyncAttendanceRequest) ([]*model.AttendanceRecord, error) {
	// 1. 获取 Access Token
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// 2. 调用钉钉考勤记录API
	records, err := a.fetchAttendanceRecords(ctx, token, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attendance records: %w", err)
	}

	// 3. 转换为系统内部格式
	result := make([]*model.AttendanceRecord, 0, len(records))
	for _, record := range records {
		converted, err := a.convertToAttendanceRecord(record, req.TenantID)
		if err != nil {
			// 记录错误但继续处理其他记录
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
			Platform:    model.PlatformDingTalk,
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
	// 钉钉不支持推送考勤规则,只能在钉钉后台配置
	return fmt.Errorf("dingtalk does not support pushing attendance rules")
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

	// 3. 调用钉钉API获取考勤记录
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

// getAccessToken 获取钉钉 Access Token
func (a *AttendanceAdapter) getAccessToken(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/gettoken?appkey=%s&appsecret=%s", a.baseURL, a.appKey, a.appSecret)

	// 这里应该调用 HTTP Client
	// 为了示例，返回模拟 token
	// TODO: 实现实际的 HTTP 请求
	_ = url // 避免未使用变量警告

	return "mock_access_token", nil
}

// fetchAttendanceRecords 获取钉钉考勤记录
func (a *AttendanceAdapter) fetchAttendanceRecords(ctx context.Context, token string, req *integration.SyncAttendanceRequest) ([]*DingTalkAttendanceRecord, error) {
	// 钉钉API: /attendance/listRecord
	// TODO: 实现实际的 HTTP 请求

	// 模拟返回数据
	return []*DingTalkAttendanceRecord{}, nil
}

// fetchDepartmentUsers 获取部门用户列表
func (a *AttendanceAdapter) fetchDepartmentUsers(ctx context.Context, token string, deptID string) ([]*DingTalkUser, error) {
	// 钉钉API: /user/listbypage
	// TODO: 实现实际的 HTTP 请求

	return []*DingTalkUser{}, nil
}

// convertToAttendanceRecord 转换钉钉考勤记录为系统格式
func (a *AttendanceAdapter) convertToAttendanceRecord(record *DingTalkAttendanceRecord, tenantID uuid.UUID) (*model.AttendanceRecord, error) {
	// 解析打卡类型
	var clockType model.AttendanceClockType
	switch record.CheckType {
	case "OnDuty":
		clockType = model.ClockTypeCheckIn
	case "OffDuty":
		clockType = model.ClockTypeCheckOut
	default:
		return nil, fmt.Errorf("unknown check type: %s", record.CheckType)
	}

	// 解析考勤状态
	var status model.AttendanceStatus
	switch record.TimeResult {
	case "Normal":
		status = model.AttendanceStatusNormal
	case "Late":
		status = model.AttendanceStatusLate
	case "Early":
		status = model.AttendanceStatusEarly
	case "Absenteeism":
		status = model.AttendanceStatusAbsent
	default:
		status = model.AttendanceStatusNormal
	}

	// 解析打卡方式
	var method model.AttendanceMethod
	switch record.LocationMethod {
	case "WiFi":
		method = model.MethodMobile
	case "Bluetooth":
		method = model.MethodMobile
	case "GPS":
		method = model.MethodMobile
	default:
		method = model.MethodDevice
	}

	// 构造考勤记录
	result := &model.AttendanceRecord{
		ClockTime:     record.UserCheckTime,
		ClockType:     clockType,
		Status:        status,
		CheckInMethod: method,
		SourceType:    model.SourceTypeDingTalk,
		SourceID:      record.ID,
		IsException:   record.TimeResult != "Normal",
		ExceptionType: record.TimeResult,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 处理位置信息
	if record.LocationResult == "Normal" && record.Latitude != 0 && record.Longitude != 0 {
		result.Location = &model.LocationInfo{
			Latitude:  record.Latitude,
			Longitude: record.Longitude,
		}
	}

	// 存储原始数据用于审计
	rawData := make(map[string]interface{})
	data, _ := json.Marshal(record)
	json.Unmarshal(data, &rawData)
	result.RawData = rawData

	return result, nil
}

// ========== 钉钉数据结构 ==========

// DingTalkAttendanceRecord 钉钉考勤记录
type DingTalkAttendanceRecord struct {
	ID             string    `json:"id"`
	UserID         string    `json:"userId"`
	WorkDate       int64     `json:"workDate"`
	UserCheckTime  time.Time `json:"userCheckTime"`
	CheckType      string    `json:"checkType"`      // OnDuty, OffDuty
	TimeResult     string    `json:"timeResult"`     // Normal, Late, Early, Absenteeism
	LocationResult string    `json:"locationResult"` // Normal, Abnormal
	LocationMethod string    `json:"locationMethod"` // WiFi, Bluetooth, GPS
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	BaseCheckTime  time.Time `json:"baseCheckTime"`
	SourceType     string    `json:"sourceType"`
	PlanID         int64     `json:"planId"`
	GroupID        int64     `json:"groupId"`
}

// DingTalkUser 钉钉用户
type DingTalkUser struct {
	UserID     string  `json:"userid"`
	Name       string  `json:"name"`
	Mobile     string  `json:"mobile"`
	Email      string  `json:"email"`
	DeptIDList []int64 `json:"dept_id_list"`
	Position   string  `json:"position"`
	JobNumber  string  `json:"job_number"`
	Active     bool    `json:"active"`
}

// toMap 转换为 map
func (u *DingTalkUser) toMap() map[string]interface{} {
	data := make(map[string]interface{})
	bytes, _ := json.Marshal(u)
	json.Unmarshal(bytes, &data)
	return data
}
