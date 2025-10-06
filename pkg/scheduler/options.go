package scheduler

import (
	"time"

	"github.com/lk2023060901/go-next-erp/pkg/logger"
)

// Option 调度器选项函数
type Option func(*Scheduler)

// WithConfig 使用自定义配置
func WithConfig(config *Config) Option {
	return func(s *Scheduler) {
		s.config = config
	}
}

// WithLogger 使用自定义日志器
func WithLogger(log *logger.Logger) Option {
	return func(s *Scheduler) {
		s.logger = log
	}
}

// WithLocation 设置时区
//
// 示例:
//
//	shanghai, _ := time.LoadLocation("Asia/Shanghai")
//	scheduler.New(scheduler.WithLocation(shanghai))
func WithLocation(loc *time.Location) Option {
	return func(s *Scheduler) {
		s.config.Location = loc
	}
}

// WithSeconds 启用秒级精度
//
// 启用后 Cron 表达式格式变为：秒 分 时 日 月 周
// 示例: "*/5 * * * * *" 表示每5秒执行一次
func WithSeconds() Option {
	return func(s *Scheduler) {
		s.config.EnableSeconds = true
	}
}

// WithPanicRecovery 配置 Panic 恢复
//
// 启用后任务 panic 不会导致调度器崩溃
// 默认启用，建议生产环境保持开启
func WithPanicRecovery(enabled bool) Option {
	return func(s *Scheduler) {
		s.config.PanicRecovery = enabled
	}
}

// WithMaxConcurrent 设置最大并发任务数
//
// 限制同时运行的任务数量，0 表示不限制
// 适用于需要控制资源消耗的场景
func WithMaxConcurrent(max int) Option {
	return func(s *Scheduler) {
		s.config.MaxConcurrent = max
	}
}

// WithShutdownTimeout 设置优雅关闭超时时间
//
// 超过此时间仍有任务在运行，将强制关闭
// 默认 30 秒
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(s *Scheduler) {
		s.config.ShutdownTimeout = timeout
	}
}

// WithMiddlewares 设置全局中间件
//
// 中间件会应用到所有任务
// 执行顺序：先注册的先执行
func WithMiddlewares(middlewares ...Middleware) Option {
	return func(s *Scheduler) {
		s.middlewares = append(s.middlewares, middlewares...)
	}
}
