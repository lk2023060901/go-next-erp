package storage

import "time"

// Option 配置选项
type Option func(*Config)

// WithEndpoint 设置 MinIO 地址
func WithEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.Endpoint = endpoint
	}
}

// WithCredentials 设置访问凭证
func WithCredentials(accessKeyID, secretAccessKey string) Option {
	return func(c *Config) {
		c.AccessKeyID = accessKeyID
		c.SecretAccessKey = secretAccessKey
	}
}

// WithSSL 设置是否使用 SSL
func WithSSL(useSSL bool) Option {
	return func(c *Config) {
		c.UseSSL = useSSL
	}
}

// WithBucket 设置默认存储桶
func WithBucket(bucketName string) Option {
	return func(c *Config) {
		c.BucketName = bucketName
	}
}

// WithRegion 设置区域
func WithRegion(region string) Option {
	return func(c *Config) {
		c.Region = region
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) Option {
	return func(c *Config) {
		c.MaxRetries = maxRetries
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}
