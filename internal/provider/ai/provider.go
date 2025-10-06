package ai

import (
	"context"
	"io"
)

// Provider AI 服务提供商接口
type Provider interface {
	// 文本生成（支持多模态输入：文本、图像、音频、视频）
	CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	CreateCompletionStream(ctx context.Context, req *CompletionRequest) (io.ReadCloser, error)

	// 文本嵌入
	CreateEmbedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// 语音转文本
	CreateTranscription(ctx context.Context, req *TranscriptionRequest) (*TranscriptionResponse, error)

	// 视频理解/分析
	AnalyzeVideo(ctx context.Context, req *VideoAnalysisRequest) (*VideoAnalysisResponse, error)

	// 图像生成
	GenerateImage(ctx context.Context, req *ImageGenerationRequest) (*ImageGenerationResponse, error)

	// 视频生成
	GenerateVideo(ctx context.Context, req *VideoGenerationRequest) (*VideoGenerationResponse, error)

	// 模型管理
	ListModels(ctx context.Context) (*ListModelsResponse, error)

	// 工具方法
	GetProviderName() string
	GetCapabilities() *ProviderCapabilities
	Close() error
}

// ProviderCapabilities 提供商能力
type ProviderCapabilities struct {
	SupportText            bool // 支持文本生成
	SupportImageInput      bool // 支持图像输入
	SupportAudioInput      bool // 支持音频输入
	SupportVideoInput      bool // 支持视频输入
	SupportImageGeneration bool // 支持图像生成
	SupportVideoGeneration bool // 支持视频生成
	SupportEmbedding       bool // 支持文本嵌入
	SupportTranscription   bool // 支持语音转文本
	SupportStreaming       bool // 支持流式输出
}

// Config 通用配置
type Config struct {
	// 基础配置
	BaseURL string // API 基础地址
	APIKey  string // API Key

	// 可选配置
	Organization string // 组织 ID（某些提供商需要）
	Model        string // 默认模型名称

	// HTTP 配置
	Timeout    int // 超时时间（秒）
	MaxRetries int // 最大重试次数
}

// ProviderType 提供商类型
type ProviderType string

const (
	ProviderTypeOpenAI      ProviderType = "openai"      // OpenAI
	ProviderTypeSiliconFlow ProviderType = "siliconflow" // 硅基流动
	ProviderTypeAnthropic   ProviderType = "anthropic"   // Anthropic
)

// NewProvider 创建 AI Provider 工厂函数
type NewProvider func(config *Config) (Provider, error)

// registry 全局提供商注册表
var registry = make(map[ProviderType]NewProvider)

// Register 注册提供商
func Register(providerType ProviderType, constructor NewProvider) {
	registry[providerType] = constructor
}

// New 创建 Provider 实例
func New(providerType ProviderType, config *Config) (Provider, error) {
	constructor, exists := registry[providerType]
	if !exists {
		return nil, ErrProviderNotFound
	}
	return constructor(config)
}

// GetRegisteredProviders 获取所有已注册的提供商类型
func GetRegisteredProviders() []ProviderType {
	types := make([]ProviderType, 0, len(registry))
	for t := range registry {
		types = append(types, t)
	}
	return types
}
