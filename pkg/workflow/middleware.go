package workflow

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/lk2023060901/go-next-erp/pkg/logger"
)

// Middleware 中间件函数签名
// 接收一个 Node 并返回包装后的 Node
type Middleware func(Node) Node

// NodeWrapper 节点包装器（实现 Node 接口）
type NodeWrapper struct {
	wrapped Node
	execute func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
}

func (nw *NodeWrapper) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	return nw.execute(ctx, input)
}

func (nw *NodeWrapper) Type() string {
	return nw.wrapped.Type()
}

func (nw *NodeWrapper) Validate() error {
	return nw.wrapped.Validate()
}

// LoggingMiddleware 日志中间件
// 记录节点执行的开始、完成和耗时
func LoggingMiddleware(log *logger.Logger) Middleware {
	return func(next Node) Node {
		return &NodeWrapper{
			wrapped: next,
			execute: func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
				nodeName := getNodeName(next)
				nodeType := next.Type()

				log.Infow("node execution started",
					"node", nodeName,
					"type", nodeType,
				)

				start := time.Now()

				// 执行节点
				output, err := next.Execute(ctx, input)

				duration := time.Since(start)

				if err != nil {
					log.Errorw("node execution failed",
						"node", nodeName,
						"type", nodeType,
						"duration", duration,
						"error", err,
					)
				} else {
					log.Infow("node execution completed",
						"node", nodeName,
						"type", nodeType,
						"duration", duration,
					)
				}

				return output, err
			},
		}
	}
}

// RecoveryMiddleware Panic 恢复中间件
// 捕获节点执行中的 panic，避免工作流崩溃
func RecoveryMiddleware(log *logger.Logger) Middleware {
	return func(next Node) Node {
		return &NodeWrapper{
			wrapped: next,
			execute: func(ctx context.Context, input map[string]interface{}) (output map[string]interface{}, err error) {
				defer func() {
					if r := recover(); r != nil {
						nodeName := getNodeName(next)
						stack := string(debug.Stack())

						log.Errorw("node panicked",
							"node", nodeName,
							"panic", r,
							"stack", stack,
						)

						// 将 panic 转换为错误
						err = fmt.Errorf("node panicked: %v", r)
						output = nil
					}
				}()

				return next.Execute(ctx, input)
			},
		}
	}
}

// MetricsMiddleware 指标收集中间件
// 记录节点执行次数、成功率、耗时分布等指标
func MetricsMiddleware(log *logger.Logger) Middleware {
	return func(next Node) Node {
		return &NodeWrapper{
			wrapped: next,
			execute: func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
				nodeName := getNodeName(next)
				nodeType := next.Type()

				// 记录执行开始
				start := time.Now()

				// 执行节点
				output, err := next.Execute(ctx, input)

				// 计算耗时
				duration := time.Since(start)

				// 记录指标
				status := "success"
				if err != nil {
					status = "failed"
				}

				log.Debugw("node metrics collected",
					"node", nodeName,
					"type", nodeType,
					"status", status,
					"duration_ms", duration.Milliseconds(),
				)

				// TODO: 发送到 Prometheus/StatsD
				// metrics.RecordNodeExecution(nodeType, status, duration)

				return output, err
			},
		}
	}
}

// TimeoutMiddleware 超时控制中间件
// 为节点执行设置超时时间，防止长时间阻塞
func TimeoutMiddleware(timeout time.Duration, log *logger.Logger) Middleware {
	return func(next Node) Node {
		return &NodeWrapper{
			wrapped: next,
			execute: func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
				nodeName := getNodeName(next)

				// 创建带超时的上下文
				timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				// 使用 channel 接收执行结果
				type result struct {
					output map[string]interface{}
					err    error
				}

				resultChan := make(chan result, 1)

				// 在 goroutine 中执行节点
				go func() {
					output, err := next.Execute(timeoutCtx, input)
					resultChan <- result{output, err}
				}()

				// 等待结果或超时
				select {
				case res := <-resultChan:
					return res.output, res.err

				case <-timeoutCtx.Done():
					log.Warnw("node execution timeout",
						"node", nodeName,
						"timeout", timeout,
					)
					return nil, fmt.Errorf("%w: node %s timeout after %v", ErrExecutionTimeout, nodeName, timeout)
				}
			},
		}
	}
}

// RetryMiddleware 重试中间件
// 在节点失败时自动重试（指数退避）
func RetryMiddleware(maxRetries int, initialDelay time.Duration, maxDelay time.Duration, log *logger.Logger) Middleware {
	return func(next Node) Node {
		return &NodeWrapper{
			wrapped: next,
			execute: func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
				nodeName := getNodeName(next)
				var lastErr error

				for attempt := 1; attempt <= maxRetries; attempt++ {
					// 执行节点
					output, err := next.Execute(ctx, input)

					if err == nil {
						// 成功
						if attempt > 1 {
							log.Infow("node execution succeeded after retry",
								"node", nodeName,
								"attempt", attempt,
							)
						}
						return output, nil
					}

					// 失败
					lastErr = err

					log.Warnw("node execution failed",
						"node", nodeName,
						"attempt", attempt,
						"max_retries", maxRetries,
						"error", err,
					)

					// 如果还有重试机会，等待后重试
					if attempt < maxRetries {
						// 指数退避
						delay := initialDelay * time.Duration(1<<uint(attempt-1))
						if delay > maxDelay {
							delay = maxDelay
						}

						log.Debugw("retrying node after delay",
							"node", nodeName,
							"delay", delay,
							"next_attempt", attempt+1,
						)

						// 可取消的延迟
						select {
						case <-time.After(delay):
							// 继续重试
						case <-ctx.Done():
							// 上下文取消
							return nil, ctx.Err()
						}
					}
				}

				// 所有重试都失败
				return nil, fmt.Errorf("node %s failed after %d attempts: %w", nodeName, maxRetries, lastErr)
			},
		}
	}
}

// RateLimitMiddleware 限流中间件
// 控制节点的执行频率，防止过载
func RateLimitMiddleware(maxConcurrent int, log *logger.Logger) Middleware {
	semaphore := make(chan struct{}, maxConcurrent)

	return func(next Node) Node {
		return &NodeWrapper{
			wrapped: next,
			execute: func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
				nodeName := getNodeName(next)

				// 获取信号量
				select {
				case semaphore <- struct{}{}:
					defer func() { <-semaphore }()

				case <-ctx.Done():
					return nil, ctx.Err()
				}

				log.Debugw("node execution acquired slot",
					"node", nodeName,
					"max_concurrent", maxConcurrent,
				)

				return next.Execute(ctx, input)
			},
		}
	}
}

// CachingMiddleware 缓存中间件
// 缓存节点的执行结果，避免重复计算
type CachingMiddleware struct {
	cache      sync.Map // 缓存: inputHash -> result
	ttl        time.Duration
	log        *logger.Logger
	keyBuilder func(map[string]interface{}) string
}

// NewCachingMiddleware 创建缓存中间件
func NewCachingMiddleware(ttl time.Duration, log *logger.Logger) *CachingMiddleware {
	return &CachingMiddleware{
		ttl: ttl,
		log: log,
		keyBuilder: func(input map[string]interface{}) string {
			// 简单实现：将输入转换为字符串作为 key
			return fmt.Sprintf("%v", input)
		},
	}
}

type cachedResult struct {
	output    map[string]interface{}
	err       error
	timestamp time.Time
}

func (cm *CachingMiddleware) Middleware() Middleware {
	return func(next Node) Node {
		return &NodeWrapper{
			wrapped: next,
			execute: func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
				nodeName := getNodeName(next)
				cacheKey := cm.keyBuilder(input)

				// 检查缓存
				if cached, ok := cm.cache.Load(cacheKey); ok {
					result := cached.(*cachedResult)

					// 检查是否过期
					if time.Since(result.timestamp) < cm.ttl {
						cm.log.Debugw("cache hit",
							"node", nodeName,
							"cache_age", time.Since(result.timestamp),
						)
						return result.output, result.err
					}

					// 过期，删除缓存
					cm.cache.Delete(cacheKey)
				}

				// 执行节点
				cm.log.Debugw("cache miss, executing node",
					"node", nodeName,
				)

				output, err := next.Execute(ctx, input)

				// 缓存结果
				cm.cache.Store(cacheKey, &cachedResult{
					output:    output,
					err:       err,
					timestamp: time.Now(),
				})

				return output, err
			},
		}
	}
}

// TracingMiddleware 分布式追踪中间件
// 集成 OpenTelemetry/Jaeger 进行链路追踪
func TracingMiddleware(log *logger.Logger) Middleware {
	return func(next Node) Node {
		return &NodeWrapper{
			wrapped: next,
			execute: func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
				nodeName := getNodeName(next)
				nodeType := next.Type()

				// TODO: 创建 span
				// tracer := otel.Tracer("workflow")
				// ctx, span := tracer.Start(ctx, fmt.Sprintf("node.%s", nodeType))
				// defer span.End()

				// span.SetAttributes(
				//     attribute.String("node.name", nodeName),
				//     attribute.String("node.type", nodeType),
				// )

				log.Debugw("tracing node execution",
					"node", nodeName,
					"type", nodeType,
				)

				output, err := next.Execute(ctx, input)

				if err != nil {
					// span.RecordError(err)
					// span.SetStatus(codes.Error, err.Error())
					log.Debugw("node execution traced (error)",
						"node", nodeName,
						"error", err,
					)
				} else {
					// span.SetStatus(codes.Ok, "success")
					log.Debugw("node execution traced (success)",
						"node", nodeName,
					)
				}

				return output, err
			},
		}
	}
}

// ValidationMiddleware 验证中间件
// 验证节点输入和输出的数据格式
type ValidationMiddleware struct {
	validateInput  func(map[string]interface{}) error
	validateOutput func(map[string]interface{}) error
	log            *logger.Logger
}

// NewValidationMiddleware 创建验证中间件
func NewValidationMiddleware(
	validateInput func(map[string]interface{}) error,
	validateOutput func(map[string]interface{}) error,
	log *logger.Logger,
) *ValidationMiddleware {
	return &ValidationMiddleware{
		validateInput:  validateInput,
		validateOutput: validateOutput,
		log:            log,
	}
}

func (vm *ValidationMiddleware) Middleware() Middleware {
	return func(next Node) Node {
		return &NodeWrapper{
			wrapped: next,
			execute: func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
				nodeName := getNodeName(next)

				// 验证输入
				if vm.validateInput != nil {
					if err := vm.validateInput(input); err != nil {
						vm.log.Errorw("node input validation failed",
							"node", nodeName,
							"error", err,
						)
						return nil, fmt.Errorf("input validation failed: %w", err)
					}
				}

				// 执行节点
				output, err := next.Execute(ctx, input)
				if err != nil {
					return nil, err
				}

				// 验证输出
				if vm.validateOutput != nil {
					if err := vm.validateOutput(output); err != nil {
						vm.log.Errorw("node output validation failed",
							"node", nodeName,
							"error", err,
						)
						return nil, fmt.Errorf("output validation failed: %w", err)
					}
				}

				return output, nil
			},
		}
	}
}

// Chain 链式组合多个中间件
// 中间件从右到左执行（类似洋葱模型）
//
// 示例:
//
//	middleware := Chain(
//	    RecoveryMiddleware(log),      // 最外层
//	    LoggingMiddleware(log),
//	    MetricsMiddleware(log),       // 最内层
//	)
//
// 执行顺序：Recovery -> Logging -> Metrics -> Node -> Metrics -> Logging -> Recovery
func Chain(middlewares ...Middleware) Middleware {
	return func(node Node) Node {
		// 从右到左应用中间件
		for i := len(middlewares) - 1; i >= 0; i-- {
			node = middlewares[i](node)
		}
		return node
	}
}

// getNodeName 获取节点名称
// 优先使用 NamedNode.Name()，否则返回类型名
func getNodeName(node Node) string {
	if namedNode, ok := node.(NamedNode); ok {
		return namedNode.Name()
	}

	return node.Type()
}
