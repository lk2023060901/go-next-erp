package scheduler

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/lk2023060901/go-next-erp/pkg/logger"
)

// Middleware 中间件函数签名
// 接收一个 Job 并返回包装后的 Job
type Middleware func(Job) Job

// LoggingMiddleware 日志中间件
// 记录任务的开始、完成和执行时长
func LoggingMiddleware(log *logger.Logger) Middleware {
	return func(next Job) Job {
		return JobFunc(func() {
			name := getJobName(next)
			start := time.Now()

			log.Infow("job started", "job", name)

			next.Run()

			duration := time.Since(start)
			log.Infow("job completed",
				"job", name,
				"duration", duration.String(),
			)
		})
	}
}

// RecoveryMiddleware Panic 恢复中间件
// 捕获任务执行中的 panic，避免调度器崩溃
func RecoveryMiddleware(log *logger.Logger) Middleware {
	return func(next Job) Job {
		return JobFunc(func() {
			defer func() {
				if r := recover(); r != nil {
					name := getJobName(next)
					stack := string(debug.Stack())

					log.Errorw("job panicked",
						"job", name,
						"panic", r,
						"stack", stack,
					)
				}
			}()

			next.Run()
		})
	}
}

// MetricsMiddleware 指标收集中间件
// 记录任务执行次数、耗时等指标
func MetricsMiddleware(log *logger.Logger) Middleware {
	return func(next Job) Job {
		return JobFunc(func() {
			name := getJobName(next)
			start := time.Now()

			// 记录执行开始
			log.Debugw("job metrics",
				"job", name,
				"event", "start",
			)

			// 执行任务
			defer func() {
				duration := time.Since(start)

				// 记录执行结果
				status := "success"
				if r := recover(); r != nil {
					status = "failed"
					panic(r) // 重新抛出 panic，让 RecoveryMiddleware 处理
				}

				log.Debugw("job metrics",
					"job", name,
					"event", "complete",
					"status", status,
					"duration_ms", duration.Milliseconds(),
				)
			}()

			next.Run()
		})
	}
}

// TimeoutMiddleware 超时控制中间件
// 限制任务的最大执行时间
//
// 注意：此中间件不会强制中断任务，只会记录超时日志
// 如需真正的超时中断，任务内部需要支持 context.Context
func TimeoutMiddleware(timeout time.Duration, log *logger.Logger) Middleware {
	return func(next Job) Job {
		return JobFunc(func() {
			name := getJobName(next)
			done := make(chan struct{})

			go func() {
				next.Run()
				close(done)
			}()

			select {
			case <-done:
				// 正常完成
			case <-time.After(timeout):
				// 超时
				log.Warnw("job timeout",
					"job", name,
					"timeout", timeout.String(),
				)
			}
		})
	}
}

// SkipIfRunningMiddleware 跳过正在执行的任务
// 如果任务正在执行，跳过本次调度
//
// 适用于执行时间可能超过调度间隔的长任务
func SkipIfRunningMiddleware(log *logger.Logger) Middleware {
	running := make(map[string]bool)

	return func(next Job) Job {
		return JobFunc(func() {
			name := getJobName(next)

			// 检查是否正在运行
			if running[name] {
				log.Warnw("job already running, skipped", "job", name)
				return
			}

			// 标记为运行中
			running[name] = true
			defer func() {
				running[name] = false
			}()

			next.Run()
		})
	}
}

// Chain 链式组合多个中间件
// 中间件从右到左执行（类似洋葱模型）
//
// 示例:
//
//	middleware := Chain(
//	    RecoveryMiddleware(log),
//	    LoggingMiddleware(log),
//	    MetricsMiddleware(log),
//	)
//
// 执行顺序：Recovery -> Logging -> Metrics -> Job -> Metrics -> Logging -> Recovery
func Chain(middlewares ...Middleware) Middleware {
	return func(job Job) Job {
		// 从右到左应用中间件
		for i := len(middlewares) - 1; i >= 0; i-- {
			job = middlewares[i](job)
		}
		return job
	}
}

// getJobName 获取任务名称
// 优先使用 NamedJob.Name()，否则返回默认值
func getJobName(job Job) string {
	if namedJob, ok := job.(NamedJob); ok {
		return namedJob.Name()
	}
	return fmt.Sprintf("%T", job)
}
