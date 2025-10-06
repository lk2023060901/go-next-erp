package workflow

import (
	"context"
)

// Node 节点接口
// 所有自定义节点必须实现此接口
type Node interface {
	// Execute 执行节点逻辑
	// ctx: 执行上下文
	// input: 节点输入数据
	// 返回: 输出数据和错误
	Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)

	// Type 返回节点类型
	Type() string

	// Validate 验证节点配置
	Validate() error
}

// NamedNode 命名节点接口（可选）
type NamedNode interface {
	Node
	Name() string
}

// BaseNode 基础节点实现
// 其他节点可以嵌入此结构体来获得基础功能
type BaseNode struct {
	id     string
	name   string
	typ    string
	config map[string]interface{}
}

// NewBaseNode 创建基础节点
func NewBaseNode(id, name, typ string, config map[string]interface{}) *BaseNode {
	return &BaseNode{
		id:     id,
		name:   name,
		typ:    typ,
		config: config,
	}
}

// ID 返回节点 ID
func (n *BaseNode) ID() string {
	return n.id
}

// Name 返回节点名称
func (n *BaseNode) Name() string {
	return n.name
}

// Type 返回节点类型
func (n *BaseNode) Type() string {
	return n.typ
}

// Config 返回节点配置
func (n *BaseNode) Config() map[string]interface{} {
	return n.config
}

// Validate 默认验证（空实现）
func (n *BaseNode) Validate() error {
	return nil
}

// NodeFactory 节点工厂函数
// 根据节点定义创建节点实例
type NodeFactory func(def *NodeDefinition) (Node, error)
