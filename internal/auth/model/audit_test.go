package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestAuditResult 测试审计结果常量
func TestAuditResult(t *testing.T) {
	tests := []struct {
		name   string
		result AuditResult
		want   string
	}{
		{
			name:   "成功结果",
			result: AuditResultSuccess,
			want:   "success",
		},
		{
			name:   "失败结果",
			result: AuditResultFailure,
			want:   "failure",
		},
		{
			name:   "拒绝结果",
			result: AuditResultDenied,
			want:   "denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.result))
		})
	}
}

// TestAuditAction 测试审计动作常量
func TestAuditAction(t *testing.T) {
	actions := map[string]string{
		// 认证相关
		"用户登录":    AuditActionLogin,
		"用户登出":    AuditActionLogout,
		"登录失败":    AuditActionLoginFailed,
		"密码重置":    AuditActionPasswordReset,
		// 用户管理
		"用户创建":    AuditActionUserCreate,
		"用户更新":    AuditActionUserUpdate,
		"用户删除":    AuditActionUserDelete,
		// 角色权限
		"角色分配":    AuditActionRoleAssign,
		"角色撤销":    AuditActionRoleRevoke,
		"权限授予":    AuditActionPermissionGrant,
		// 数据操作
		"数据读取":    AuditActionDataRead,
		"数据创建":    AuditActionDataCreate,
		"数据更新":    AuditActionDataUpdate,
		"数据删除":    AuditActionDataDelete,
	}

	for name, action := range actions {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, action)
			assert.Contains(t, action, ".")
		})
	}
}

// TestAuditLog_JSON 测试审计日志 JSON 序列化
func TestAuditLog_JSON(t *testing.T) {
	auditID := uuid.New()
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	beforeData := map[string]interface{}{"status": "active"}
	afterData := map[string]interface{}{"status": "inactive"}
	beforeJSON, _ := json.Marshal(beforeData)
	afterJSON, _ := json.Marshal(afterData)

	audit := &AuditLog{
		ID:         auditID,
		EventID:    "evt-123",
		TenantID:   tenantID,
		UserID:     userID,
		Action:     AuditActionUserUpdate,
		Resource:   "user",
		ResourceID: "user-456",
		BeforeData: beforeJSON,
		AfterData:  afterJSON,
		IPAddress:  "192.168.1.100",
		UserAgent:  "Mozilla/5.0",
		Result:     AuditResultSuccess,
		ErrorMsg:   "",
		Metadata: map[string]interface{}{
			"source": "web",
			"version": "1.0",
		},
		CreatedAt: now,
	}

	// 序列化
	data, err := json.Marshal(audit)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// 反序列化
	var decoded AuditLog
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	// 验证字段
	assert.Equal(t, auditID, decoded.ID)
	assert.Equal(t, "evt-123", decoded.EventID)
	assert.Equal(t, tenantID, decoded.TenantID)
	assert.Equal(t, userID, decoded.UserID)
	assert.Equal(t, AuditActionUserUpdate, decoded.Action)
	assert.Equal(t, "user", decoded.Resource)
	assert.Equal(t, "user-456", decoded.ResourceID)
	assert.Equal(t, "192.168.1.100", decoded.IPAddress)
	assert.Equal(t, "Mozilla/5.0", decoded.UserAgent)
	assert.Equal(t, AuditResultSuccess, decoded.Result)
	assert.Equal(t, now, decoded.CreatedAt)

	// 验证 BeforeData 和 AfterData
	var decodedBefore, decodedAfter map[string]interface{}
	json.Unmarshal(decoded.BeforeData, &decodedBefore)
	json.Unmarshal(decoded.AfterData, &decodedAfter)
	assert.Equal(t, "active", decodedBefore["status"])
	assert.Equal(t, "inactive", decodedAfter["status"])

	// 验证 Metadata
	assert.Equal(t, "web", decoded.Metadata["source"])
	assert.Equal(t, "1.0", decoded.Metadata["version"])
}

// TestAuditLog_MinimalFields 测试最小字段的审计日志
func TestAuditLog_MinimalFields(t *testing.T) {
	audit := &AuditLog{
		ID:        uuid.New(),
		EventID:   "evt-min",
		TenantID:  uuid.New(),
		UserID:    uuid.New(),
		Action:    AuditActionLogin,
		Resource:  "session",
		IPAddress: "127.0.0.1",
		UserAgent: "curl/7.0",
		Result:    AuditResultSuccess,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(audit)
	assert.NoError(t, err)

	var decoded AuditLog
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, audit.ID, decoded.ID)
	assert.Equal(t, audit.EventID, decoded.EventID)
	assert.Equal(t, audit.Action, decoded.Action)
	assert.Empty(t, decoded.ResourceID) // 可选字段应为空
	assert.Nil(t, decoded.BeforeData)   // 可选字段应为空
	assert.Nil(t, decoded.AfterData)    // 可选字段应为空
	assert.Empty(t, decoded.ErrorMsg)   // 可选字段应为空
}

// TestAuditLog_FailedOperation 测试失败操作的审计日志
func TestAuditLog_FailedOperation(t *testing.T) {
	audit := &AuditLog{
		ID:        uuid.New(),
		EventID:   "evt-fail",
		TenantID:  uuid.New(),
		UserID:    uuid.New(),
		Action:    AuditActionDataDelete,
		Resource:  "document",
		ResourceID: "doc-789",
		IPAddress: "10.0.0.1",
		UserAgent: "Mozilla/5.0",
		Result:    AuditResultFailure,
		ErrorMsg:  "permission denied",
		Metadata: map[string]interface{}{
			"attempt": 3,
			"reason":  "insufficient privileges",
		},
		CreatedAt: time.Now(),
	}

	assert.Equal(t, AuditResultFailure, audit.Result)
	assert.Equal(t, "permission denied", audit.ErrorMsg)
	assert.Equal(t, 3, audit.Metadata["attempt"])
	assert.Equal(t, "insufficient privileges", audit.Metadata["reason"])
}

// TestAuditLog_DeniedOperation 测试拒绝操作的审计日志
func TestAuditLog_DeniedOperation(t *testing.T) {
	audit := &AuditLog{
		ID:        uuid.New(),
		EventID:   "evt-deny",
		TenantID:  uuid.New(),
		UserID:    uuid.New(),
		Action:    AuditActionDataRead,
		Resource:  "confidential",
		ResourceID: "conf-123",
		IPAddress: "192.168.1.50",
		UserAgent: "PostmanRuntime/7.0",
		Result:    AuditResultDenied,
		ErrorMsg:  "access denied",
		CreatedAt: time.Now(),
	}

	assert.Equal(t, AuditResultDenied, audit.Result)
	assert.Equal(t, "access denied", audit.ErrorMsg)
}

// TestAuditLog_ZeroValues 测试零值情况
func TestAuditLog_ZeroValues(t *testing.T) {
	audit := &AuditLog{}

	// 零值 UUID 应该是有效的（全零）
	assert.Equal(t, uuid.UUID{}, audit.ID)
	assert.Equal(t, uuid.UUID{}, audit.TenantID)
	assert.Equal(t, uuid.UUID{}, audit.UserID)

	// 零值字符串应为空
	assert.Empty(t, audit.EventID)
	assert.Empty(t, audit.Action)
	assert.Empty(t, audit.Resource)
	assert.Empty(t, audit.ResourceID)
	assert.Empty(t, audit.IPAddress)
	assert.Empty(t, audit.UserAgent)
	assert.Empty(t, audit.ErrorMsg)

	// 零值 Result 应为空字符串
	assert.Empty(t, audit.Result)

	// 零值时间应为 zero time
	assert.True(t, audit.CreatedAt.IsZero())
}

// TestAuditLog_ComplexMetadata 测试复杂元数据
func TestAuditLog_ComplexMetadata(t *testing.T) {
	audit := &AuditLog{
		ID:       uuid.New(),
		EventID:  "evt-complex",
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Action:   AuditActionDataCreate,
		Resource: "report",
		Result:   AuditResultSuccess,
		Metadata: map[string]interface{}{
			"tags": []string{"finance", "quarterly"},
			"config": map[string]interface{}{
				"format":  "pdf",
				"language": "en",
			},
			"count": 42,
			"active": true,
		},
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(audit)
	assert.NoError(t, err)

	var decoded AuditLog
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	// 验证嵌套结构
	tags := decoded.Metadata["tags"].([]interface{})
	assert.Len(t, tags, 2)
	assert.Equal(t, "finance", tags[0])
	assert.Equal(t, "quarterly", tags[1])

	config := decoded.Metadata["config"].(map[string]interface{})
	assert.Equal(t, "pdf", config["format"])
	assert.Equal(t, "en", config["language"])

	assert.Equal(t, 42.0, decoded.Metadata["count"])
	assert.Equal(t, true, decoded.Metadata["active"])
}

// TestAuditLog_BeforeAfterData 测试 BeforeData 和 AfterData
func TestAuditLog_BeforeAfterData(t *testing.T) {
	before := map[string]interface{}{
		"name":   "John Doe",
		"email":  "john@example.com",
		"status": "active",
	}
	after := map[string]interface{}{
		"name":   "John Smith",
		"email":  "john.smith@example.com",
		"status": "active",
	}

	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)

	audit := &AuditLog{
		ID:         uuid.New(),
		EventID:    "evt-update",
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		Action:     AuditActionUserUpdate,
		Resource:   "user",
		ResourceID: "user-123",
		BeforeData: beforeJSON,
		AfterData:  afterJSON,
		IPAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		Result:     AuditResultSuccess,
		CreatedAt:  time.Now(),
	}

	// 验证数据可以正确序列化和反序列化
	data, err := json.Marshal(audit)
	assert.NoError(t, err)

	var decoded AuditLog
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	var decodedBefore, decodedAfter map[string]interface{}
	json.Unmarshal(decoded.BeforeData, &decodedBefore)
	json.Unmarshal(decoded.AfterData, &decodedAfter)

	// 验证 before 数据
	assert.Equal(t, "John Doe", decodedBefore["name"])
	assert.Equal(t, "john@example.com", decodedBefore["email"])

	// 验证 after 数据
	assert.Equal(t, "John Smith", decodedAfter["name"])
	assert.Equal(t, "john.smith@example.com", decodedAfter["email"])
}

// TestAuditResult_Validation 测试审计结果验证
func TestAuditResult_Validation(t *testing.T) {
	validResults := []AuditResult{
		AuditResultSuccess,
		AuditResultFailure,
		AuditResultDenied,
	}

	for _, result := range validResults {
		t.Run(string(result), func(t *testing.T) {
			assert.NotEmpty(t, result)
			assert.True(t, len(result) > 0)
		})
	}

	// 测试无效结果
	invalidResult := AuditResult("invalid")
	assert.NotEqual(t, AuditResultSuccess, invalidResult)
	assert.NotEqual(t, AuditResultFailure, invalidResult)
	assert.NotEqual(t, AuditResultDenied, invalidResult)
}
