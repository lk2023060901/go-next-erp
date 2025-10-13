package rebac

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/stretchr/testify/mock"
)

// BenchmarkEngine_Check 基准测试直接关系检查
func BenchmarkEngine_Check(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	repo := new(MockRelationRepository)
	repo.On("Check", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(true, nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Check(ctx, tenantID, "user:alice", "viewer", "document:123")
	}
}

// BenchmarkEngine_Check_NotFound 基准测试关系不存在的情况
func BenchmarkEngine_Check_NotFound(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	repo := new(MockRelationRepository)
	repo.On("Check", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(false, nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Check(ctx, tenantID, "user:alice", "viewer", "document:123")
	}
}

// BenchmarkEngine_CheckTransitive_Direct 基准测试传递检查（直接关系）
func BenchmarkEngine_CheckTransitive_Direct(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	repo := new(MockRelationRepository)
	repo.On("Check", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(true, nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.CheckTransitive(ctx, tenantID, "user:alice", "viewer", "document:123")
	}
}

// BenchmarkEngine_CheckTransitive_Inheritance 基准测试传递检查（继承关系）
func BenchmarkEngine_CheckTransitive_Inheritance(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	repo := new(MockRelationRepository)
	// 直接关系不存在
	repo.On("Check", mock.Anything, mock.Anything, "user:alice", "viewer", "document:123").
		Return(false, nil)
	// 返回owner关系（继承viewer）
	tuples := []*model.RelationTuple{
		{
			Subject:  "user:alice",
			Relation: "owner",
			Object:   "document:123",
		},
	}
	repo.On("FindByRelation", mock.Anything, mock.Anything, "user:alice", "document:123").
		Return(tuples, nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.CheckTransitive(ctx, tenantID, "user:alice", "viewer", "document:123")
	}
}

// BenchmarkEngine_CheckTransitive_MultiLevel 基准测试传递检查（多层继承）
func BenchmarkEngine_CheckTransitive_MultiLevel(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	repo := new(MockRelationRepository)
	// 直接关系不存在
	repo.On("Check", mock.Anything, mock.Anything, "user:alice", "viewer", "document:123").
		Return(false, nil)
	// 返回editor关系（editor继承viewer）
	tuples := []*model.RelationTuple{
		{
			Subject:  "user:alice",
			Relation: "editor",
			Object:   "document:123",
		},
	}
	repo.On("FindByRelation", mock.Anything, mock.Anything, "user:alice", "document:123").
		Return(tuples, nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.CheckTransitive(ctx, tenantID, "user:alice", "viewer", "document:123")
	}
}

// BenchmarkEngine_Expand 基准测试展开关系
func BenchmarkEngine_Expand(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	repo := new(MockRelationRepository)
	subjects := []string{"user:alice", "user:bob", "user:charlie"}
	repo.On("Expand", mock.Anything, mock.Anything, "document:123", "viewer").
		Return(subjects, nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Expand(ctx, tenantID, "document:123", "viewer")
	}
}

// BenchmarkEngine_Expand_ManySubjects 基准测试展开大量主体
func BenchmarkEngine_Expand_ManySubjects(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	// 生成100个主体
	subjects := make([]string, 100)
	for i := 0; i < 100; i++ {
		subjects[i] = fmt.Sprintf("user:user%d", i)
	}

	repo := new(MockRelationRepository)
	repo.On("Expand", mock.Anything, mock.Anything, "document:123", "viewer").
		Return(subjects, nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Expand(ctx, tenantID, "document:123", "viewer")
	}
}

// BenchmarkEngine_ListUserObjects 基准测试列出用户对象
func BenchmarkEngine_ListUserObjects(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()
	userID := "alice"

	repo := new(MockRelationRepository)
	tuples := []*model.RelationTuple{
		{Object: "document:123"},
		{Object: "document:456"},
		{Object: "document:789"},
	}
	repo.On("FindBySubject", mock.Anything, mock.Anything, "user:alice").
		Return(tuples, nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.ListUserObjects(ctx, tenantID, userID, "viewer", "document")
	}
}

// BenchmarkEngine_ListUserObjects_ManyObjects 基准测试列出大量对象
func BenchmarkEngine_ListUserObjects_ManyObjects(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()
	userID := "alice"

	// 生成100个对象
	tuples := make([]*model.RelationTuple, 100)
	for i := 0; i < 100; i++ {
		tuples[i] = &model.RelationTuple{
			Object: fmt.Sprintf("document:%d", i),
		}
	}

	repo := new(MockRelationRepository)
	repo.On("FindBySubject", mock.Anything, mock.Anything, "user:alice").
		Return(tuples, nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.ListUserObjects(ctx, tenantID, userID, "viewer", "document")
	}
}

// BenchmarkEngine_Grant 基准测试授予关系
func BenchmarkEngine_Grant(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	repo := new(MockRelationRepository)
	repo.On("Create", mock.Anything, mock.Anything).
		Return(nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.Grant(ctx, tenantID, "user:alice", "viewer", "document:123")
	}
}

// BenchmarkEngine_Revoke 基准测试撤销关系
func BenchmarkEngine_Revoke(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	repo := new(MockRelationRepository)
	repo.On("DeleteByTuple", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.Revoke(ctx, tenantID, "user:alice", "viewer", "document:123")
	}
}

// BenchmarkEngine_Check_Parallel 并发基准测试直接检查
func BenchmarkEngine_Check_Parallel(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	repo := new(MockRelationRepository)
	repo.On("Check", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(true, nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = engine.Check(ctx, tenantID, "user:alice", "viewer", "document:123")
		}
	})
}

// BenchmarkEngine_CheckTransitive_Parallel 并发基准测试传递检查
func BenchmarkEngine_CheckTransitive_Parallel(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	repo := new(MockRelationRepository)
	repo.On("Check", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(false, nil)
	tuples := []*model.RelationTuple{
		{
			Subject:  "user:alice",
			Relation: "owner",
			Object:   "document:123",
		},
	}
	repo.On("FindByRelation", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(tuples, nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = engine.CheckTransitive(ctx, tenantID, "user:alice", "viewer", "document:123")
		}
	})
}

// BenchmarkEngine_Expand_Parallel 并发基准测试展开关系
func BenchmarkEngine_Expand_Parallel(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	repo := new(MockRelationRepository)
	subjects := []string{"user:alice", "user:bob", "user:charlie"}
	repo.On("Expand", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(subjects, nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = engine.Expand(ctx, tenantID, "document:123", "viewer")
		}
	})
}

// BenchmarkEngine_ListUserObjects_Parallel 并发基准测试列出用户对象
func BenchmarkEngine_ListUserObjects_Parallel(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()
	userID := "alice"

	repo := new(MockRelationRepository)
	tuples := []*model.RelationTuple{
		{Object: "document:123"},
		{Object: "document:456"},
		{Object: "document:789"},
	}
	repo.On("FindBySubject", mock.Anything, mock.Anything, mock.Anything).
		Return(tuples, nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = engine.ListUserObjects(ctx, tenantID, userID, "viewer", "document")
		}
	})
}

// BenchmarkEngine_Grant_Parallel 并发基准测试授予关系
func BenchmarkEngine_Grant_Parallel(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	repo := new(MockRelationRepository)
	repo.On("Create", mock.Anything, mock.Anything).
		Return(nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = engine.Grant(ctx, tenantID, "user:alice", "viewer", "document:123")
		}
	})
}

// BenchmarkEngine_Revoke_Parallel 并发基准测试撤销关系
func BenchmarkEngine_Revoke_Parallel(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	repo := new(MockRelationRepository)
	repo.On("DeleteByTuple", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = engine.Revoke(ctx, tenantID, "user:alice", "viewer", "document:123")
		}
	})
}

// BenchmarkEngine_ComplexScenario 复杂场景基准测试
func BenchmarkEngine_ComplexScenario(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()

	repo := new(MockRelationRepository)

	// Grant操作
	repo.On("Create", mock.Anything, mock.MatchedBy(func(tuple *model.RelationTuple) bool {
		return tuple.Relation == "owner"
	})).Return(nil)

	// CheckTransitive操作
	repo.On("Check", mock.Anything, mock.Anything, "user:alice", "viewer", "document:123").
		Return(false, nil)
	ownerTuples := []*model.RelationTuple{
		{Subject: "user:alice", Relation: "owner", Object: "document:123"},
	}
	repo.On("FindByRelation", mock.Anything, mock.Anything, "user:alice", "document:123").
		Return(ownerTuples, nil)

	// Expand操作
	repo.On("Expand", mock.Anything, mock.Anything, "document:123", "viewer").
		Return([]string{"user:alice", "user:bob"}, nil)

	// ListUserObjects操作
	userTuples := []*model.RelationTuple{
		{Object: "document:123"},
		{Object: "document:456"},
	}
	repo.On("FindBySubject", mock.Anything, mock.Anything, "user:alice").
		Return(userTuples, nil)

	engine := NewEngine(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 1. 授予owner权限
		_ = engine.Grant(ctx, tenantID, "user:alice", "owner", "document:123")

		// 2. 检查传递的viewer权限
		_, _ = engine.CheckTransitive(ctx, tenantID, "user:alice", "viewer", "document:123")

		// 3. 展开有viewer权限的用户
		_, _ = engine.Expand(ctx, tenantID, "document:123", "viewer")

		// 4. 列出用户可访问的对象
		_, _ = engine.ListUserObjects(ctx, tenantID, "alice", "viewer", "document")
	}
}
