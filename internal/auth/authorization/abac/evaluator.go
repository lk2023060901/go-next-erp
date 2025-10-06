package abac

import (
	"fmt"
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// EvaluationContext 评估上下文
type EvaluationContext struct {
	User        map[string]interface{} `expr:"User"`        // 用户属性
	Resource    map[string]interface{} `expr:"Resource"`    // 资源属性
	Environment map[string]interface{} `expr:"Environment"` // 环境属性
	Time        map[string]interface{} `expr:"Time"`        // 时间属性
}

// Evaluator 表达式评估器（基于 Expr）
type Evaluator struct {
	// 程序缓存（提升性能）
	programCache sync.Map // map[string]*vm.Program
}

// NewEvaluator 创建评估器
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate 评估表达式
func (e *Evaluator) Evaluate(expression string, ctx *EvaluationContext) (bool, error) {
	// 1. 尝试从缓存获取编译后的程序
	if cached, ok := e.programCache.Load(expression); ok {
		program := cached.(*vm.Program)
		return e.runProgram(program, ctx)
	}

	// 2. 编译表达式
	program, err := expr.Compile(
		expression,
		expr.Env(EvaluationContext{}),
		expr.AsBool(), // 结果必须是布尔值
	)
	if err != nil {
		return false, fmt.Errorf("compile expression failed: %w", err)
	}

	// 3. 缓存编译结果
	e.programCache.Store(expression, program)

	// 4. 执行程序
	return e.runProgram(program, ctx)
}

// runProgram 执行编译后的程序
func (e *Evaluator) runProgram(program *vm.Program, ctx *EvaluationContext) (bool, error) {
	output, err := expr.Run(program, ctx)
	if err != nil {
		return false, fmt.Errorf("execute expression failed: %w", err)
	}

	result, ok := output.(bool)
	if !ok {
		return false, fmt.Errorf("expression result is not boolean: %T", output)
	}

	return result, nil
}

// ValidateExpression 验证表达式语法
func (e *Evaluator) ValidateExpression(expression string) error {
	_, err := expr.Compile(
		expression,
		expr.Env(EvaluationContext{}),
		expr.AsBool(),
	)
	return err
}

// ClearCache 清除程序缓存
func (e *Evaluator) ClearCache() {
	e.programCache.Range(func(key, value interface{}) bool {
		e.programCache.Delete(key)
		return true
	})
}
