package rerank

import "context"

// Provider Rerank 服务提供商接口（预留）
type Provider interface {
	// Rerank 重排序
	Rerank(ctx context.Context, req *RerankRequest) (*RerankResponse, error)

	// 工具方法
	GetProviderName() string
	Close() error
}

// RerankRequest 重排序请求
type RerankRequest struct {
	Model     string   `json:"model"`               // 模型名称
	Query     string   `json:"query"`               // 查询文本
	Documents []string `json:"documents"`           // 文档列表
	TopN      int      `json:"top_n,omitempty"`     // 返回前 N 个结果
	ReturnDocuments bool `json:"return_documents,omitempty"` // 是否返回文档内容
}

// RerankResult 重排序结果
type RerankResult struct {
	Index          int     `json:"index"`           // 原始文档索引
	RelevanceScore float64 `json:"relevance_score"` // 相关性分数
	Document       string  `json:"document,omitempty"` // 文档内容（可选）
}

// RerankResponse 重排序响应
type RerankResponse struct {
	ID      string         `json:"id"`
	Model   string         `json:"model"`
	Results []RerankResult `json:"results"`
	Usage   Usage          `json:"usage,omitempty"`
}

// Usage 使用统计
type Usage struct {
	TotalTokens int `json:"total_tokens"`
}

// Config 通用配置
type Config struct {
	// 基础配置
	BaseURL string // API 基础地址
	APIKey  string // API Key

	// 可选配置
	Model   string // 默认模型名称
	Timeout int    // 超时时间（秒）
}

// ProviderType 提供商类型
type ProviderType string

const (
	ProviderTypeCohere   ProviderType = "cohere"   // Cohere
	ProviderTypeJina     ProviderType = "jina"     // Jina AI
	ProviderTypeVoyage   ProviderType = "voyage"   // Voyage AI
	ProviderTypeCustom   ProviderType = "custom"   // 自定义实现
)

// NewProvider 创建 Rerank Provider 工厂函数
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
