package storage

import (
	"fmt"
	"time"
)

// Config MinIO 配置
type Config struct {
	// 连接配置
	Endpoint        string // MinIO 地址 (例如: localhost:9000)
	AccessKeyID     string // Access Key
	SecretAccessKey string // Secret Key
	UseSSL          bool   // 是否使用 SSL

	// 默认存储桶
	BucketName string // 默认存储桶名称
	Region     string // 区域 (例如: us-east-1)

	// 连接池配置
	MaxRetries int           // 最大重试次数
	Timeout    time.Duration // 超时时间
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		UseSSL:          false,
		BucketName:      "default",
		Region:          "us-east-1",
		MaxRetries:      3,
		Timeout:         30 * time.Second,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	if c.AccessKeyID == "" {
		return fmt.Errorf("access_key_id is required")
	}
	if c.SecretAccessKey == "" {
		return fmt.Errorf("secret_access_key is required")
	}
	if c.BucketName == "" {
		return fmt.Errorf("bucket_name is required")
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be >= 0")
	}
	if c.Timeout <= 0 {
		c.Timeout = 30 * time.Second
	}
	return nil
}
