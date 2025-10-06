package workflow

import (
	"time"

	"github.com/lk2023060901/go-next-erp/pkg/logger"
)

// Option 配置函数
type Option func(*Engine)

// WithConfig 设置配置
func WithConfig(cfg *Config) Option {
	return func(e *Engine) {
		e.config = cfg
	}
}

// WithLogger 设置日志器
func WithLogger(log *logger.Logger) Option {
	return func(e *Engine) {
		e.logger = log
	}
}

// WithExecutionTimeout 设置执行超时时间
func WithExecutionTimeout(timeout time.Duration) Option {
	return func(e *Engine) {
		e.config.DefaultExecutionTimeout = timeout
	}
}

// WithMaxConcurrent 设置最大并发执行数
func WithMaxConcurrent(max int) Option {
	return func(e *Engine) {
		e.config.MaxConcurrentExecutions = max
	}
}

// WithRetryPolicy 设置默认重试策略
func WithRetryPolicy(maxRetries int, delay time.Duration, backoffRate float64) Option {
	return func(e *Engine) {
		e.config.DefaultMaxRetries = maxRetries
		e.config.DefaultRetryDelay = delay
		e.config.DefaultBackoffRate = backoffRate
	}
}

// WithPersistence 启用持久化
func WithPersistence(enabled bool, backend string) Option {
	return func(e *Engine) {
		e.config.EnablePersistence = enabled
		e.config.PersistenceBackend = backend
	}
}

// WithMetrics 启用指标收集
func WithMetrics(enabled bool) Option {
	return func(e *Engine) {
		e.config.EnableMetrics = enabled
	}
}

// WithTracing 启用链路追踪
func WithTracing(enabled bool) Option {
	return func(e *Engine) {
		e.config.EnableTracing = enabled
	}
}

// WithRetention 设置执行记录保留策略
func WithRetention(days int, cleanupInterval time.Duration) Option {
	return func(e *Engine) {
		e.config.RetentionDays = days
		e.config.CleanupInterval = cleanupInterval
	}
}

// WithPersistenceProvider 设置持久化提供者
func WithPersistenceProvider(provider PersistenceProvider) Option {
	return func(e *Engine) {
		e.persistence = provider
	}
}
