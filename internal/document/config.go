package document

import (
	"errors"
	"os"
	"time"
)

// Config MinerU 配置
type Config struct {
	// API 配置
	BaseURL string // API 基础地址，默认: https://mineru.net/api/v4
	APIKey  string // API Key (Token)

	// 超时配置
	Timeout       time.Duration // HTTP 请求超时时间，默认: 30s
	PollInterval  time.Duration // 轮询间隔，默认: 5s
	PollTimeout   time.Duration // 轮询总超时时间，默认: 30m
	UploadTimeout time.Duration // 文件上传超时时间，默认: 10m

	// 重试配置
	MaxRetries int // 最大重试次数，默认: 3
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		BaseURL:       getEnv("MINERU_BASE_URL", "https://mineru.net/api/v4"),
		APIKey:        getEnv("MINERU_API_KEY", ""),
		Timeout:       30 * time.Second,
		PollInterval:  5 * time.Second,
		PollTimeout:   30 * time.Minute,
		UploadTimeout: 10 * time.Minute,
		MaxRetries:    3,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.BaseURL == "" {
		return errors.New("base URL is required")
	}
	if c.APIKey == "" {
		return errors.New("API key is required")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}
	if c.PollInterval <= 0 {
		return errors.New("poll interval must be greater than 0")
	}
	if c.PollTimeout <= 0 {
		return errors.New("poll timeout must be greater than 0")
	}
	if c.UploadTimeout <= 0 {
		return errors.New("upload timeout must be greater than 0")
	}
	if c.MaxRetries < 0 {
		return errors.New("max retries must be non-negative")
	}
	return nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
