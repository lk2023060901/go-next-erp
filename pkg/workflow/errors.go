package workflow

import "errors"

var (
	// 工作流状态错误
	ErrWorkflowNotFound      = errors.New("workflow not found")
	ErrWorkflowAlreadyExists = errors.New("workflow already exists")
	ErrWorkflowInvalidState  = errors.New("workflow in invalid state")

	// 执行错误
	ErrExecutionNotFound     = errors.New("execution not found")
	ErrExecutionAlreadyDone  = errors.New("execution already completed or failed")
	ErrExecutionTimeout      = errors.New("execution timeout")
	ErrExecutionCancelled    = errors.New("execution cancelled")

	// 节点错误
	ErrNodeNotFound          = errors.New("node not found")
	ErrNodeTypeNotRegistered = errors.New("node type not registered")
	ErrNodeExecutionFailed   = errors.New("node execution failed")
	ErrInvalidNodeConfig     = errors.New("invalid node configuration")

	// 连接错误
	ErrInvalidEdge           = errors.New("invalid edge definition")
	ErrCyclicDependency      = errors.New("cyclic dependency detected")
	ErrDisconnectedGraph     = errors.New("disconnected workflow graph")

	// 验证错误
	ErrInvalidWorkflowDef    = errors.New("invalid workflow definition")
	ErrMissingTriggerNode    = errors.New("workflow must have at least one trigger node")
	ErrInvalidCondition      = errors.New("invalid condition expression")

	// 持久化错误
	ErrPersistenceFailed     = errors.New("persistence operation failed")
)
