package workflow

import (
	"context"
	"fmt"
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// ConditionEvaluator 条件表达式求值器
// 基于 expr-lang/expr 实现，与 ABAC 模块共用同一表达式引擎
type ConditionEvaluator struct {
	programCache sync.Map // 编译后的程序缓存
}

// NewConditionEvaluator 创建条件求值器
func NewConditionEvaluator() *ConditionEvaluator {
	return &ConditionEvaluator{}
}

// Evaluate 求值条件表达式
//
// expression: 表达式字符串，如 "output.status == 'success'"
// ctx: 执行上下文，包含所有可访问的变量
//
// 返回: 布尔值结果和错误
func (e *ConditionEvaluator) Evaluate(ctx context.Context, expression string, execCtx *ExecutionContext) (bool, error) {
	if expression == "" {
		return true, nil // 空表达式默认为 true
	}

	// 尝试从缓存获取编译后的程序
	if cached, ok := e.programCache.Load(expression); ok {
		program := cached.(*vm.Program)
		return e.runProgram(program, execCtx)
	}

	// 编译表达式
	env := e.buildEnv(execCtx)
	program, err := expr.Compile(
		expression,
		expr.Env(env),
		expr.AsBool(), // 确保返回布尔值
	)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrInvalidCondition, err)
	}

	// 缓存编译结果
	e.programCache.Store(expression, program)

	// 执行程序
	return e.runProgram(program, execCtx)
}

// runProgram 执行已编译的程序
func (e *ConditionEvaluator) runProgram(program *vm.Program, execCtx *ExecutionContext) (bool, error) {
	env := e.buildEnv(execCtx)

	result, err := expr.Run(program, env)
	if err != nil {
		return false, fmt.Errorf("expression execution failed: %w", err)
	}

	boolResult, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("expression did not return boolean: got %T", result)
	}

	return boolResult, nil
}

// buildEnv 构建表达式执行环境
// 提供所有可在表达式中访问的变量
func (e *ConditionEvaluator) buildEnv(execCtx *ExecutionContext) map[string]interface{} {
	env := make(map[string]interface{})

	// 基础上下文
	env["input"] = execCtx.Input
	env["output"] = execCtx.Output
	env["variables"] = execCtx.Variables
	env["status"] = execCtx.Status
	env["workflow_id"] = execCtx.WorkflowID
	env["execution_id"] = execCtx.ID

	// 节点状态
	env["nodes"] = make(map[string]interface{})
	for nodeID, state := range execCtx.NodeStates {
		env["nodes"].(map[string]interface{})[nodeID] = map[string]interface{}{
			"status": state.Status,
			"input":  state.Input,
			"output": state.Output,
			"error":  state.Error,
		}
	}

	// 元数据
	if execCtx.Metadata != nil {
		env["metadata"] = execCtx.Metadata
	}

	return env
}

// ValidateExpression 验证表达式语法
// 在工作流定义阶段使用，提前发现语法错误
func (e *ConditionEvaluator) ValidateExpression(expression string) error {
	if expression == "" {
		return nil
	}

	// 使用空环境编译，只检查语法
	_, err := expr.Compile(expression, expr.AsBool())
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCondition, err)
	}

	return nil
}

// ClearCache 清空程序缓存
// 在表达式更新后调用
func (e *ConditionEvaluator) ClearCache() {
	e.programCache = sync.Map{}
}

// 常用条件表达式示例（文档）:
//
// 1. 基于输出状态:
//    "output.status == 'success'"
//    "output.code >= 200 && output.code < 300"
//
// 2. 基于节点输出:
//    "nodes['http_request'].output.status_code == 200"
//    "nodes['db_query'].output.row_count > 0"
//
// 3. 基于变量:
//    "variables.user_role == 'admin'"
//    "variables.amount > 1000"
//
// 4. 组合条件:
//    "(output.status == 'success') && (variables.retry_count < 3)"
//    "nodes['step1'].status == 'completed' || variables.skip_validation"
//
// 5. 字符串操作:
//    "variables.email contains '@example.com'"
//    "len(output.items) > 0"
//
// 6. 数组操作:
//    "output.tags contains 'important'"
//    "len(output.results) >= 10"
