package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestRelationTuple_String 测试关系元组字符串表示
func TestRelationTuple_String(t *testing.T) {
	tests := []struct {
		name     string
		tuple    *RelationTuple
		expected string
	}{
		{
			name: "用户拥有文档",
			tuple: &RelationTuple{
				Subject:  "user:alice",
				Relation: "owner",
				Object:   "document:123",
			},
			expected: "(user:alice, owner, document:123)",
		},
		{
			name: "用户编辑文档",
			tuple: &RelationTuple{
				Subject:  "user:bob",
				Relation: "editor",
				Object:   "document:456",
			},
			expected: "(user:bob, editor, document:456)",
		},
		{
			name: "用户查看文档",
			tuple: &RelationTuple{
				Subject:  "user:charlie",
				Relation: "viewer",
				Object:   "document:789",
			},
			expected: "(user:charlie, viewer, document:789)",
		},
		{
			name: "组成员关系",
			tuple: &RelationTuple{
				Subject:  "user:123",
				Relation: "member",
				Object:   "group:sales",
			},
			expected: "(user:123, member, group:sales)",
		},
		{
			name: "父级关系",
			tuple: &RelationTuple{
				Subject:  "folder:abc",
				Relation: "parent",
				Object:   "document:xyz",
			},
			expected: "(folder:abc, parent, document:xyz)",
		},
		{
			name: "空值元组",
			tuple: &RelationTuple{
				Subject:  "",
				Relation: "",
				Object:   "",
			},
			expected: "(, , )",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tuple.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRelationConstants 测试关系常量
func TestRelationConstants(t *testing.T) {
	tests := []struct {
		name     string
		relation string
		expected string
	}{
		{"所有者", RelationOwner, "owner"},
		{"编辑者", RelationEditor, "editor"},
		{"查看者", RelationViewer, "viewer"},
		{"成员", RelationMember, "member"},
		{"管理员", RelationAdmin, "admin"},
		{"父级", RelationParent, "parent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.relation)
		})
	}
}

// TestRelationTuple_JSON 测试关系元组 JSON 序列化
func TestRelationTuple_JSON(t *testing.T) {
	tupleID := uuid.New()
	tenantID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	tuple := &RelationTuple{
		ID:        tupleID,
		TenantID:  tenantID,
		Subject:   "user:alice",
		Relation:  RelationOwner,
		Object:    "document:123",
		CreatedAt: now,
		DeletedAt: nil,
	}

	// 序列化
	data, err := json.Marshal(tuple)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// 反序列化
	var decoded RelationTuple
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	// 验证字段
	assert.Equal(t, tupleID, decoded.ID)
	assert.Equal(t, tenantID, decoded.TenantID)
	assert.Equal(t, "user:alice", decoded.Subject)
	assert.Equal(t, RelationOwner, decoded.Relation)
	assert.Equal(t, "document:123", decoded.Object)
	assert.Equal(t, now, decoded.CreatedAt)
	assert.Nil(t, decoded.DeletedAt)
}

// TestRelationTuple_SoftDelete 测试软删除
func TestRelationTuple_SoftDelete(t *testing.T) {
	deletedTime := time.Now().UTC()
	tuple := &RelationTuple{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Subject:   "user:bob",
		Relation:  RelationEditor,
		Object:    "document:456",
		CreatedAt: time.Now().Add(-time.Hour),
		DeletedAt: &deletedTime,
	}

	// 验证软删除字段
	assert.NotNil(t, tuple.DeletedAt)
	assert.Equal(t, deletedTime, *tuple.DeletedAt)

	// JSON 序列化时 DeletedAt 应该被忽略（json:"-"）
	data, err := json.Marshal(tuple)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	// deleted_at 不应该出现在 JSON 中
	_, exists := decoded["deleted_at"]
	assert.False(t, exists)
}

// TestRelationTuple_OwnerRelation 测试所有者关系
func TestRelationTuple_OwnerRelation(t *testing.T) {
	tuple := &RelationTuple{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Subject:   "user:alice",
		Relation:  RelationOwner,
		Object:    "document:report-2024",
		CreatedAt: time.Now(),
	}

	assert.Equal(t, RelationOwner, tuple.Relation)
	assert.Equal(t, "(user:alice, owner, document:report-2024)", tuple.String())
}

// TestRelationTuple_EditorRelation 测试编辑者关系
func TestRelationTuple_EditorRelation(t *testing.T) {
	tuple := &RelationTuple{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Subject:   "user:bob",
		Relation:  RelationEditor,
		Object:    "document:contract",
		CreatedAt: time.Now(),
	}

	assert.Equal(t, RelationEditor, tuple.Relation)
	assert.Equal(t, "(user:bob, editor, document:contract)", tuple.String())
}

// TestRelationTuple_ViewerRelation 测试查看者关系
func TestRelationTuple_ViewerRelation(t *testing.T) {
	tuple := &RelationTuple{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Subject:   "user:charlie",
		Relation:  RelationViewer,
		Object:    "document:presentation",
		CreatedAt: time.Now(),
	}

	assert.Equal(t, RelationViewer, tuple.Relation)
	assert.Equal(t, "(user:charlie, viewer, document:presentation)", tuple.String())
}

// TestRelationTuple_MemberRelation 测试成员关系
func TestRelationTuple_MemberRelation(t *testing.T) {
	tuple := &RelationTuple{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Subject:   "user:david",
		Relation:  RelationMember,
		Object:    "group:engineering",
		CreatedAt: time.Now(),
	}

	assert.Equal(t, RelationMember, tuple.Relation)
	assert.Equal(t, "(user:david, member, group:engineering)", tuple.String())
}

// TestRelationTuple_AdminRelation 测试管理员关系
func TestRelationTuple_AdminRelation(t *testing.T) {
	tuple := &RelationTuple{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Subject:   "user:eve",
		Relation:  RelationAdmin,
		Object:    "organization:acme",
		CreatedAt: time.Now(),
	}

	assert.Equal(t, RelationAdmin, tuple.Relation)
	assert.Equal(t, "(user:eve, admin, organization:acme)", tuple.String())
}

// TestRelationTuple_ParentRelation 测试父级关系
func TestRelationTuple_ParentRelation(t *testing.T) {
	tuple := &RelationTuple{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Subject:   "folder:projects",
		Relation:  RelationParent,
		Object:    "folder:2024",
		CreatedAt: time.Now(),
	}

	assert.Equal(t, RelationParent, tuple.Relation)
	assert.Equal(t, "(folder:projects, parent, folder:2024)", tuple.String())
}

// TestRelationTuple_ZeroValues 测试零值情况
func TestRelationTuple_ZeroValues(t *testing.T) {
	tuple := &RelationTuple{}

	// 零值 UUID 应该是有效的（全零）
	assert.Equal(t, uuid.UUID{}, tuple.ID)
	assert.Equal(t, uuid.UUID{}, tuple.TenantID)

	// 零值字符串应为空
	assert.Empty(t, tuple.Subject)
	assert.Empty(t, tuple.Relation)
	assert.Empty(t, tuple.Object)

	// 零值时间应为 zero time
	assert.True(t, tuple.CreatedAt.IsZero())

	// 零值指针应为 nil
	assert.Nil(t, tuple.DeletedAt)

	// String 方法应该能处理零值
	assert.Equal(t, "(, , )", tuple.String())
}

// TestRelationTuple_ComplexSubjects 测试复杂主体
func TestRelationTuple_ComplexSubjects(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		object  string
	}{
		{"用户到文档", "user:123", "document:456"},
		{"组到项目", "group:team-alpha", "project:phoenix"},
		{"角色到资源", "role:admin", "resource:database"},
		{"服务到服务", "service:api", "service:backend"},
		{"带UUID的用户", "user:" + uuid.New().String(), "document:789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tuple := &RelationTuple{
				ID:        uuid.New(),
				TenantID:  uuid.New(),
				Subject:   tt.subject,
				Relation:  RelationViewer,
				Object:    tt.object,
				CreatedAt: time.Now(),
			}

			assert.Equal(t, tt.subject, tuple.Subject)
			assert.Equal(t, tt.object, tuple.Object)
			assert.Contains(t, tuple.String(), tt.subject)
			assert.Contains(t, tuple.String(), tt.object)
		})
	}
}

// TestRelationTuple_Equality 测试元组相等性
func TestRelationTuple_Equality(t *testing.T) {
	tenantID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	tuple1 := &RelationTuple{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Subject:   "user:alice",
		Relation:  RelationOwner,
		Object:    "document:123",
		CreatedAt: now,
	}

	tuple2 := &RelationTuple{
		ID:        tuple1.ID, // 相同 ID
		TenantID:  tenantID,
		Subject:   "user:alice",
		Relation:  RelationOwner,
		Object:    "document:123",
		CreatedAt: now,
	}

	// 相同的字段应该相等
	assert.Equal(t, tuple1.ID, tuple2.ID)
	assert.Equal(t, tuple1.TenantID, tuple2.TenantID)
	assert.Equal(t, tuple1.Subject, tuple2.Subject)
	assert.Equal(t, tuple1.Relation, tuple2.Relation)
	assert.Equal(t, tuple1.Object, tuple2.Object)
	assert.Equal(t, tuple1.CreatedAt, tuple2.CreatedAt)
	assert.Equal(t, tuple1.String(), tuple2.String())
}

// TestRelationTuple_MultiTenant 测试多租户场景
func TestRelationTuple_MultiTenant(t *testing.T) {
	tenant1 := uuid.New()
	tenant2 := uuid.New()

	tuple1 := &RelationTuple{
		ID:        uuid.New(),
		TenantID:  tenant1,
		Subject:   "user:alice",
		Relation:  RelationOwner,
		Object:    "document:123",
		CreatedAt: time.Now(),
	}

	tuple2 := &RelationTuple{
		ID:        uuid.New(),
		TenantID:  tenant2,
		Subject:   "user:alice", // 相同用户，不同租户
		Relation:  RelationOwner,
		Object:    "document:123",
		CreatedAt: time.Now(),
	}

	// 不同租户的元组应该是独立的
	assert.NotEqual(t, tuple1.TenantID, tuple2.TenantID)
	assert.Equal(t, tuple1.Subject, tuple2.Subject) // 主体可以相同
	assert.Equal(t, tuple1.Object, tuple2.Object)   // 对象可以相同
	assert.NotEqual(t, tuple1.ID, tuple2.ID)        // 但 ID 必须不同
}

// TestRelationTuple_StringFormat 测试字符串格式的一致性
func TestRelationTuple_StringFormat(t *testing.T) {
	tuple := &RelationTuple{
		Subject:  "user:alice",
		Relation: "owner",
		Object:   "document:123",
	}

	str := tuple.String()

	// 验证格式：(subject, relation, object)
	assert.Contains(t, str, "(")
	assert.Contains(t, str, ")")
	assert.Contains(t, str, ",")
	assert.Contains(t, str, "user:alice")
	assert.Contains(t, str, "owner")
	assert.Contains(t, str, "document:123")

	// 验证格式顺序
	assert.Equal(t, "(user:alice, owner, document:123)", str)
}

// TestRelationTuple_SpecialCharacters 测试特殊字符
func TestRelationTuple_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		tuple    *RelationTuple
		expected string
	}{
		{
			name: "带空格的主体",
			tuple: &RelationTuple{
				Subject:  "user:John Doe",
				Relation: "owner",
				Object:   "document:123",
			},
			expected: "(user:John Doe, owner, document:123)",
		},
		{
			name: "带特殊字符的对象",
			tuple: &RelationTuple{
				Subject:  "user:alice",
				Relation: "editor",
				Object:   "document:report-2024_v1.0",
			},
			expected: "(user:alice, editor, document:report-2024_v1.0)",
		},
		{
			name: "Unicode 字符",
			tuple: &RelationTuple{
				Subject:  "user:张三",
				Relation: "viewer",
				Object:   "document:文档",
			},
			expected: "(user:张三, viewer, document:文档)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tuple.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
