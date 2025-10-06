package search

import "context"

// Provider 搜索引擎服务提供商接口（预留）
type Provider interface {
	// Search 搜索
	Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error)

	// 工具方法
	GetProviderName() string
	Close() error
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query      string   `json:"query"`                 // 查询文本
	NumResults int      `json:"num_results,omitempty"` // 返回结果数，默认 10
	Language   string   `json:"language,omitempty"`    // 语言偏好
	Country    string   `json:"country,omitempty"`     // 国家/地区
	SafeSearch bool     `json:"safe_search,omitempty"` // 安全搜索
	Freshness  string   `json:"freshness,omitempty"`   // 时效性: day, week, month, year
	Sites      []string `json:"sites,omitempty"`       // 限定站点
}

// SearchResult 搜索结果
type SearchResult struct {
	Title       string            `json:"title"`       // 标题
	URL         string            `json:"url"`         // URL
	Description string            `json:"description"` // 描述
	Content     string            `json:"content,omitempty"` // 网页内容（可选）
	PublishedDate string          `json:"published_date,omitempty"` // 发布日期
	Author      string            `json:"author,omitempty"`        // 作者
	Score       float64           `json:"score,omitempty"`         // 相关性分数
	Metadata    map[string]string `json:"metadata,omitempty"`      // 额外元数据
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Total   int            `json:"total,omitempty"` // 总结果数（估计）
}

// Config 通用配置
type Config struct {
	// 基础配置
	BaseURL string // API 基础地址
	APIKey  string // API Key

	// 可选配置
	Timeout int // 超时时间（秒）
}

// ProviderType 提供商类型
type ProviderType string

const (
	ProviderTypeSerper    ProviderType = "serper"    // Serper (Google Search)
	ProviderTypeTavily    ProviderType = "tavily"    // Tavily AI Search
	ProviderTypeBing      ProviderType = "bing"      // Bing Search
	ProviderTypeSerpAPI   ProviderType = "serpapi"   // SerpAPI
	ProviderTypeBrave     ProviderType = "brave"     // Brave Search
	ProviderTypeCustom    ProviderType = "custom"    // 自定义实现
)

// NewProvider 创建 Search Provider 工厂函数
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
