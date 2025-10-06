package workflow

import (
	"fmt"
	"sync"
)

// Registry 节点类型注册表
// 管理所有可用的节点类型及其工厂函数
type Registry struct {
	mu        sync.RWMutex
	factories map[string]NodeFactory
}

// NewRegistry 创建节点注册表
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]NodeFactory),
	}
}

// Register 注册节点类型
func (r *Registry) Register(nodeType string, factory NodeFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[nodeType]; exists {
		return fmt.Errorf("node type %s already registered", nodeType)
	}

	r.factories[nodeType] = factory
	return nil
}

// Unregister 注销节点类型
func (r *Registry) Unregister(nodeType string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.factories, nodeType)
}

// Create 创建节点实例
func (r *Registry) Create(def *NodeDefinition) (Node, error) {
	r.mu.RLock()
	factory, exists := r.factories[def.Type]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrNodeTypeNotRegistered, def.Type)
	}

	node, err := factory(def)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	// 验证节点配置
	if err := node.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidNodeConfig, err)
	}

	return node, nil
}

// ListTypes 列出所有已注册的节点类型
func (r *Registry) ListTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.factories))
	for typ := range r.factories {
		types = append(types, typ)
	}

	return types
}

// HasType 检查节点类型是否已注册
func (r *Registry) HasType(nodeType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.factories[nodeType]
	return exists
}

// globalRegistry 全局节点注册表
var globalRegistry = NewRegistry()

// RegisterNodeType 向全局注册表注册节点类型（便捷方法）
func RegisterNodeType(nodeType string, factory NodeFactory) error {
	return globalRegistry.Register(nodeType, factory)
}

// CreateNode 从全局注册表创建节点（便捷方法）
func CreateNode(def *NodeDefinition) (Node, error) {
	return globalRegistry.Create(def)
}
