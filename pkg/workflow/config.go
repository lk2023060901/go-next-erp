package workflow

import (
	"fmt"
	"time"
)

// Config 工作流引擎配置
type Config struct {
	// 执行配置
	DefaultExecutionTimeout time.Duration
	MaxConcurrentExecutions int

	// 重试配置
	DefaultMaxRetries  int
	DefaultRetryDelay  time.Duration
	DefaultBackoffRate float64

	// 持久化配置
	EnablePersistence  bool
	PersistenceBackend string // "postgres", "mongodb", "memory"

	// 监控配置
	EnableMetrics bool
	EnableTracing bool

	// 清理配置
	RetentionDays        int           // 执行记录保留天数
	CleanupInterval      time.Duration // 清理间隔
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		DefaultExecutionTimeout: 30 * time.Minute,
		MaxConcurrentExecutions: 100,

		DefaultMaxRetries:  3,
		DefaultRetryDelay:  5 * time.Second,
		DefaultBackoffRate: 2.0,

		EnablePersistence:  true,
		PersistenceBackend: "postgres",

		EnableMetrics: true,
		EnableTracing: false,

		RetentionDays:   30,
		CleanupInterval: 24 * time.Hour,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.DefaultExecutionTimeout <= 0 {
		return fmt.Errorf("invalid execution timeout: %v", c.DefaultExecutionTimeout)
	}

	if c.MaxConcurrentExecutions < 0 {
		return fmt.Errorf("max concurrent executions must be >= 0")
	}

	if c.DefaultMaxRetries < 0 {
		return fmt.Errorf("max retries must be >= 0")
	}

	if c.DefaultBackoffRate < 1.0 {
		return fmt.Errorf("backoff rate must be >= 1.0")
	}

	if c.EnablePersistence {
		if c.PersistenceBackend == "" {
			return fmt.Errorf("persistence backend not specified")
		}
	}

	if c.RetentionDays < 0 {
		return fmt.Errorf("retention days must be >= 0")
	}

	return nil
}
