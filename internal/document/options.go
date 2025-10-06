package document

import "time"

// Option 配置选项
type Option func(*Config)

// WithBaseURL 设置 API 基础地址
func WithBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

// WithAPIKey 设置 API Key
func WithAPIKey(apiKey string) Option {
	return func(c *Config) {
		c.APIKey = apiKey
	}
}

// WithTimeout 设置 HTTP 请求超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithPollInterval 设置轮询间隔
func WithPollInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.PollInterval = interval
	}
}

// WithPollTimeout 设置轮询总超时时间
func WithPollTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.PollTimeout = timeout
	}
}

// WithUploadTimeout 设置文件上传超时时间
func WithUploadTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.UploadTimeout = timeout
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) Option {
	return func(c *Config) {
		c.MaxRetries = maxRetries
	}
}
