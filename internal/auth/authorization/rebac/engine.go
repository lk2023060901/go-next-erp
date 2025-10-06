package rebac

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/model"
	"github.com/lk2023060901/go-next-erp/internal/auth/repository"
)

// Engine ReBAC 授权引擎（关系型授权，类似 Google Zanzibar）
type Engine struct {
	relationRepo repository.RelationRepository
}

// NewEngine 创建 ReBAC 引擎
func NewEngine(relationRepo repository.RelationRepository) *Engine {
	return &Engine{
		relationRepo: relationRepo,
	}
}

// Check 检查主体是否对客体拥有指定关系
func (e *Engine) Check(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) (bool, error) {
	return e.relationRepo.Check(ctx, tenantID, subject, relation, object)
}

// CheckTransitive 检查传递关系（支持继承）
// 例如：user:123 是否对 document:789 拥有 viewer 权限
// 如果 user:123 -> editor -> document:789，且 editor 继承 viewer，则返回 true
func (e *Engine) CheckTransitive(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) (bool, error) {
	// 1. 直接检查
	exists, err := e.relationRepo.Check(ctx, tenantID, subject, relation, object)
	if err != nil {
		return false, err
	}

	if exists {
		return true, nil
	}

	// 2. 检查继承关系
	// 定义关系继承：owner > editor > viewer
	inheritanceMap := map[string][]string{
		"owner":  {"editor", "viewer"},
		"editor": {"viewer"},
	}

	// 查找主体拥有的所有关系
	tuples, err := e.relationRepo.FindByRelation(ctx, tenantID, subject, object)
	if err != nil {
		return false, err
	}

	// 检查是否有继承关系
	for _, tuple := range tuples {
		if inherited, ok := inheritanceMap[tuple.Relation]; ok {
			for _, inheritedRelation := range inherited {
				if inheritedRelation == relation {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// Expand 展开所有拥有指定关系的主体
// 例如：谁是 document:789 的 viewer？
func (e *Engine) Expand(ctx context.Context, tenantID uuid.UUID, object, relation string) ([]string, error) {
	return e.relationRepo.Expand(ctx, tenantID, object, relation)
}

// ListUserObjects 列出用户可访问的所有对象
// 例如：user:123 作为 viewer 可以访问哪些 document？
func (e *Engine) ListUserObjects(ctx context.Context, tenantID uuid.UUID, userID, relation, objectType string) ([]string, error) {
	subject := fmt.Sprintf("user:%s", userID)

	// 查找用户的所有关系
	tuples, err := e.relationRepo.FindBySubject(ctx, tenantID, subject)
	if err != nil {
		return nil, err
	}

	var objects []string
	for _, tuple := range tuples {
		// 过滤关系和对象类型
		if tuple.Relation == relation {
			// 提取对象（格式：document:123）
			if objectType == "" || hasPrefix(tuple.Object, objectType+":") {
				objects = append(objects, tuple.Object)
			}
		}
	}

	return objects, nil
}

// Grant 授予关系
func (e *Engine) Grant(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) error {
	tuple := &model.RelationTuple{
		TenantID: tenantID,
		Subject:  subject,
		Relation: relation,
		Object:   object,
	}

	return e.relationRepo.Create(ctx, tuple)
}

// Revoke 撤销关系
func (e *Engine) Revoke(ctx context.Context, tenantID uuid.UUID, subject, relation, object string) error {
	return e.relationRepo.DeleteByTuple(ctx, tenantID, subject, relation, object)
}

// hasPrefix 检查字符串前缀
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
