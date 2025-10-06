package vector

import (
	"errors"
	"os"
	"strconv"
	"time"
)

// Config Milvus 配置
type Config struct {
	// 连接配置
	Endpoint string // Milvus 地址，例如: localhost:19530
	Username string // 用户名（可选）
	Password string // 密码（可选）

	// 数据库配置
	Database string // 数据库名称，默认: default

	// 连接池配置
	MaxIdleConns int           // 最大空闲连接数，默认: 10
	MaxOpenConns int           // 最大打开连接数，默认: 50
	Timeout      time.Duration // 超时时间，默认: 30s

	// 重试配置
	MaxRetries int // 最大重试次数，默认: 3
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Endpoint:     getEnv("MILVUS_ENDPOINT", "localhost:19530"),
		Username:     getEnv("MILVUS_USERNAME", ""),
		Password:     getEnv("MILVUS_PASSWORD", ""),
		Database:     getEnv("MILVUS_DATABASE", "default"),
		MaxIdleConns: getEnvAsInt("MILVUS_MAX_IDLE_CONNS", 10),
		MaxOpenConns: getEnvAsInt("MILVUS_MAX_OPEN_CONNS", 50),
		Timeout:      getEnvAsDuration("MILVUS_TIMEOUT", 30*time.Second),
		MaxRetries:   getEnvAsInt("MILVUS_MAX_RETRIES", 3),
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	if c.Database == "" {
		return errors.New("database is required")
	}
	if c.MaxIdleConns <= 0 {
		return errors.New("max idle conns must be greater than 0")
	}
	if c.MaxOpenConns <= 0 {
		return errors.New("max open conns must be greater than 0")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
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

// getEnvAsInt 获取整型环境变量
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsDuration 获取时间段环境变量
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
