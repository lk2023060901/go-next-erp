package workflow

import (
	"sync"
	"time"
)

// NewExecutionContext 创建新的执行上下文
func NewExecutionContext(workflowID, executionID, triggerBy string, input map[string]interface{}) *ExecutionContext {
	return &ExecutionContext{
		ID:         executionID,
		WorkflowID: workflowID,
		Status:     ExecutionStatusPending,
		Input:      input,
		Output:     make(map[string]interface{}),
		Variables:  make(map[string]interface{}),
		NodeStates: make(map[string]*NodeState),
		TriggerBy:  triggerBy,
		StartedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}
}

// SetVariable 设置上下文变量
func (ctx *ExecutionContext) SetVariable(key string, value interface{}) {
	if ctx.Variables == nil {
		ctx.Variables = make(map[string]interface{})
	}
	ctx.Variables[key] = value
}

// GetVariable 获取上下文变量
func (ctx *ExecutionContext) GetVariable(key string) (interface{}, bool) {
	if ctx.Variables == nil {
		return nil, false
	}
	val, ok := ctx.Variables[key]
	return val, ok
}

// SetOutput 设置输出数据
func (ctx *ExecutionContext) SetOutput(key string, value interface{}) {
	if ctx.Output == nil {
		ctx.Output = make(map[string]interface{})
	}
	ctx.Output[key] = value
}

// GetNodeState 获取节点状态
func (ctx *ExecutionContext) GetNodeState(nodeID string) (*NodeState, bool) {
	state, ok := ctx.NodeStates[nodeID]
	return state, ok
}

// SetNodeState 设置节点状态
func (ctx *ExecutionContext) SetNodeState(nodeID string, state *NodeState) {
	if ctx.NodeStates == nil {
		ctx.NodeStates = make(map[string]*NodeState)
	}
	ctx.NodeStates[nodeID] = state
}

// MarkCompleted 标记执行完成
func (ctx *ExecutionContext) MarkCompleted() {
	ctx.Status = ExecutionStatusCompleted
	now := time.Now()
	ctx.CompletedAt = &now
}

// MarkFailed 标记执行失败
func (ctx *ExecutionContext) MarkFailed(err error) {
	ctx.Status = ExecutionStatusFailed
	ctx.Error = err.Error()
	now := time.Now()
	ctx.CompletedAt = &now
}

// MarkCancelled 标记执行取消
func (ctx *ExecutionContext) MarkCancelled() {
	ctx.Status = ExecutionStatusCancelled
	now := time.Now()
	ctx.CompletedAt = &now
}

// Duration 获取执行耗时
func (ctx *ExecutionContext) Duration() time.Duration {
	if ctx.CompletedAt == nil {
		return time.Since(ctx.StartedAt)
	}
	return ctx.CompletedAt.Sub(ctx.StartedAt)
}

// Clone 克隆执行上下文（用于子工作流）
func (ctx *ExecutionContext) Clone() *ExecutionContext {
	cloned := &ExecutionContext{
		ID:            ctx.ID + "-clone",
		WorkflowID:    ctx.WorkflowID,
		Status:        ExecutionStatusPending,
		Input:         make(map[string]interface{}),
		Output:        make(map[string]interface{}),
		Variables:     make(map[string]interface{}),
		NodeStates:    make(map[string]*NodeState),
		TriggerBy:     ctx.TriggerBy,
		StartedAt:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}

	// 深拷贝变量
	for k, v := range ctx.Variables {
		cloned.Variables[k] = v
	}

	return cloned
}

// ContextManager 上下文管理器
// 管理所有正在执行的工作流上下文
type ContextManager struct {
	mu       sync.RWMutex
	contexts map[string]*ExecutionContext // executionID -> context
}

// NewContextManager 创建上下文管理器
func NewContextManager() *ContextManager {
	return &ContextManager{
		contexts: make(map[string]*ExecutionContext),
	}
}

// Store 存储执行上下文
func (cm *ContextManager) Store(ctx *ExecutionContext) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.contexts[ctx.ID] = ctx
}

// Load 加载执行上下文
func (cm *ContextManager) Load(executionID string) (*ExecutionContext, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	ctx, ok := cm.contexts[executionID]
	return ctx, ok
}

// Delete 删除执行上下文
func (cm *ContextManager) Delete(executionID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.contexts, executionID)
}

// Count 获取当前执行数量
func (cm *ContextManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.contexts)
}

// List 列出所有执行上下文
func (cm *ContextManager) List() []*ExecutionContext {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	contexts := make([]*ExecutionContext, 0, len(cm.contexts))
	for _, ctx := range cm.contexts {
		contexts = append(contexts, ctx)
	}

	return contexts
}

// Cleanup 清理已完成的上下文
func (cm *ContextManager) Cleanup(retentionDuration time.Duration) int {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	threshold := time.Now().Add(-retentionDuration)
	cleaned := 0

	for id, ctx := range cm.contexts {
		// 只清理已完成的且超过保留期的
		if ctx.CompletedAt != nil && ctx.CompletedAt.Before(threshold) {
			delete(cm.contexts, id)
			cleaned++
		}
	}

	return cleaned
}
