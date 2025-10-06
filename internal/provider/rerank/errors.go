package rerank

import "errors"

var (
	// ErrProviderNotFound 提供商未找到
	ErrProviderNotFound = errors.New("provider not found")

	// ErrInvalidConfig 无效配置
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrInvalidRequest 无效请求
	ErrInvalidRequest = errors.New("invalid request")

	// ErrAPIError API 错误
	ErrAPIError = errors.New("API error")

	// ErrNetworkError 网络错误
	ErrNetworkError = errors.New("network error")

	// ErrRateLimitExceeded 速率限制超出
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrInsufficientQuota 配额不足
	ErrInsufficientQuota = errors.New("insufficient quota")
)
