package scheduler

import "time"

// Config 调度器配置
type Config struct {
	// Location 时区设置（默认：time.UTC）
	Location *time.Location

	// EnableSeconds 是否支持秒级精度（默认：false）
	// 启用后 Cron 表达式格式：秒 分 时 日 月 周
	// 禁用时 Cron 表达式格式：分 时 日 月 周
	EnableSeconds bool

	// PanicRecovery 是否启用 Panic 恢复（默认：true）
	// 启用后任务 panic 不会导致调度器崩溃
	PanicRecovery bool

	// MaxConcurrent 最大并发任务数（默认：0，不限制）
	// 设置大于 0 的值可以限制同时运行的任务数量
	MaxConcurrent int

	// ShutdownTimeout 优雅关闭超时时间（默认：30s）
	// 超过此时间仍有任务在运行，将强制关闭
	ShutdownTimeout time.Duration
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Location:        time.UTC,
		EnableSeconds:   false,
		PanicRecovery:   true,
		MaxConcurrent:   0,
		ShutdownTimeout: 30 * time.Second,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Location == nil {
		c.Location = time.UTC
	}

	if c.ShutdownTimeout <= 0 {
		c.ShutdownTimeout = 30 * time.Second
	}

	if c.MaxConcurrent < 0 {
		c.MaxConcurrent = 0
	}

	return nil
}
