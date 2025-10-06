# internal/provider/search - 搜索引擎服务提供商接口（预留）

## 概述

搜索引擎服务提供商接口，用于集成外部搜索 API，为 AI Agent 提供实时互联网搜索能力。

## 核心接口

```go
type Provider interface {
    // Search 搜索
    Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error)

    // 工具方法
    GetProviderName() string
    Close() error
}
```

## 使用场景

### 1. AI Agent 实时搜索

```go
provider, _ := search.New(search.ProviderTypeSerper, &search.Config{
    APIKey: "your-api-key",
})

req := &search.SearchRequest{
    Query:      "2025年人工智能最新进展",
    NumResults: 5,
    Freshness:  "month", // 最近一个月
    Language:   "zh",
}

resp, _ := provider.Search(ctx, req)

for _, result := range resp.Results {
    fmt.Printf("标题: %s\n", result.Title)
    fmt.Printf("链接: %s\n", result.URL)
    fmt.Printf("摘要: %s\n\n", result.Description)
}
```

### 2. RAG 系统增强

```go
// 1. 搜索最新信息
searchResults, _ := searchProvider.Search(ctx, &search.SearchRequest{
    Query:      userQuery,
    NumResults: 10,
    Freshness:  "week",
})

// 2. 提取网页内容
documents := extractContent(searchResults)

// 3. 向量化并存储
embeddings := embeddingProvider.CreateEmbedding(ctx, documents)
vectorDB.Insert(embeddings)

// 4. 结合本地知识库回答
localResults := vectorDB.Search(userQuery)
allContext := merge(searchResults, localResults)
answer := llm.Generate(userQuery, allContext)
```

### 3. 多引擎聚合

```go
// 并发调用多个搜索引擎
var wg sync.WaitGroup
results := make(chan *search.SearchResponse, 3)

engines := []search.Provider{
    serperProvider,
    tavilyProvider,
    bingProvider,
}

for _, engine := range engines {
    wg.Add(1)
    go func(e search.Provider) {
        defer wg.Done()
        resp, _ := e.Search(ctx, &search.SearchRequest{
            Query: "AI news",
            NumResults: 5,
        })
        results <- resp
    }(engine)
}

wg.Wait()
close(results)

// 聚合去重
aggregated := aggregateResults(results)
```

## 支持的提供商

### Serper (Google Search)

```go
provider, _ := search.New(search.ProviderTypeSerper, &search.Config{
    BaseURL: "https://google.serper.dev",
    APIKey:  "your-serper-api-key",
})
```

**特点：**
- 基于 Google 搜索
- 结果质量高
- 支持多语言
- 低延迟

### Tavily AI Search

```go
provider, _ := search.New(search.ProviderTypeTavily, &search.Config{
    BaseURL: "https://api.tavily.com",
    APIKey:  "your-tavily-api-key",
})
```

**特点：**
- AI 优化的搜索
- 自动提取关键内容
- 支持深度搜索
- RAG 友好

### Bing Search

```go
provider, _ := search.New(search.ProviderTypeBing, &search.Config{
    BaseURL: "https://api.bing.microsoft.com/v7.0",
    APIKey:  "your-bing-api-key",
})
```

**特点：**
- Microsoft 官方 API
- 覆盖面广
- 企业级可靠性
- 丰富的过滤选项

### SerpAPI

```go
provider, _ := search.New(search.ProviderTypeSerpAPI, &search.Config{
    BaseURL: "https://serpapi.com",
    APIKey:  "your-serpapi-key",
})
```

**特点：**
- 支持多个搜索引擎
- 结构化数据提取
- 丰富的元数据
- 支持图片、新闻等多种搜索

### Brave Search

```go
provider, _ := search.New(search.ProviderTypeBrave, &search.Config{
    BaseURL: "https://api.search.brave.com/res/v1",
    APIKey:  "your-brave-api-key",
})
```

**特点：**
- 隐私保护
- 独立索引
- 无广告
- 快速响应

## 请求参数

```go
type SearchRequest struct {
    Query      string   // 查询文本（必填）
    NumResults int      // 返回结果数，默认 10
    Language   string   // 语言: zh, en, ja 等
    Country    string   // 国家/地区: CN, US, JP 等
    SafeSearch bool     // 安全搜索过滤
    Freshness  string   // 时效性: day, week, month, year
    Sites      []string // 限定站点，如 ["wikipedia.org", "github.com"]
}
```

## 响应格式

```go
type SearchResponse struct {
    Query   string         // 查询文本
    Results []SearchResult // 搜索结果
    Total   int            // 总结果数（估计）
}

type SearchResult struct {
    Title         string            // 标题
    URL           string            // URL
    Description   string            // 描述/摘要
    Content       string            // 网页内容（可选）
    PublishedDate string            // 发布日期
    Author        string            // 作者
    Score         float64           // 相关性分数
    Metadata      map[string]string // 额外元数据
}
```

## 最佳实践

### 1. 站点限定搜索

```go
req := &search.SearchRequest{
    Query: "机器学习教程",
    Sites: []string{
        "github.com",
        "arxiv.org",
        "medium.com",
    },
    NumResults: 10,
}

resp, _ := provider.Search(ctx, req)
```

### 2. 时效性控制

```go
// 最新资讯
newsReq := &search.SearchRequest{
    Query:     "AI breakthrough",
    Freshness: "day", // 最近一天
    NumResults: 5,
}

// 学术论文
paperReq := &search.SearchRequest{
    Query:     "deep learning survey",
    Freshness: "year", // 最近一年
    Sites:     []string{"arxiv.org"},
}
```

### 3. 多语言搜索

```go
// 中文搜索
zhResp, _ := provider.Search(ctx, &search.SearchRequest{
    Query:    "人工智能",
    Language: "zh",
    Country:  "CN",
})

// 英文搜索
enResp, _ := provider.Search(ctx, &search.SearchRequest{
    Query:    "artificial intelligence",
    Language: "en",
    Country:  "US",
})
```

### 4. 内容提取

```go
import "net/http"

func extractContent(result search.SearchResult) string {
    // 如果 API 返回了内容，直接使用
    if result.Content != "" {
        return result.Content
    }

    // 否则，爬取网页内容
    resp, _ := http.Get(result.URL)
    defer resp.Body.Close()

    // 解析 HTML，提取正文
    content := parseHTML(resp.Body)
    return content
}
```

### 5. 错误处理和重试

```go
func searchWithRetry(provider search.Provider, req *search.SearchRequest, maxRetries int) (*search.SearchResponse, error) {
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        resp, err := provider.Search(ctx, req)
        if err == nil {
            return resp, nil
        }

        lastErr = err

        // 处理特定错误
        if errors.Is(err, search.ErrRateLimitExceeded) {
            // 等待后重试
            time.Sleep(time.Second * time.Duration(i+1))
            continue
        }

        // 其他错误不重试
        break
    }

    return nil, lastErr
}
```

## 集成示例

### AI Agent 搜索工具

```go
package main

import (
    "context"
    "fmt"

    "github.com/lk2023060901/go-next-erp/internal/provider/ai"
    "github.com/lk2023060901/go-next-erp/internal/provider/search"
)

func main() {
    ctx := context.Background()

    // 初始化搜索和 LLM
    searchProvider, _ := search.New(search.ProviderTypeSerper, &search.Config{
        APIKey: "your-search-api-key",
    })

    llm, _ := ai.New(ai.ProviderTypeOpenAI, &ai.Config{
        APIKey: "your-openai-api-key",
    })

    // 用户问题
    question := "2025年最新的 AI 模型有哪些？"

    // 1. 搜索最新信息
    searchResp, _ := searchProvider.Search(ctx, &search.SearchRequest{
        Query:      question,
        NumResults: 5,
        Freshness:  "month",
        Language:   "zh",
    })

    // 2. 构建上下文
    context := buildSearchContext(searchResp.Results)

    // 3. LLM 生成回答
    answer, _ := llm.CreateCompletion(ctx, &ai.CompletionRequest{
        Model: "gpt-4",
        Messages: []ai.Message{
            ai.NewSystemMessage("你是一个AI助手，请基于搜索结果回答问题，并注明信息来源。"),
            ai.NewUserTextMessage(fmt.Sprintf("搜索结果：\n%s\n\n问题：%s", context, question)),
        },
    })

    fmt.Println(ai.GetTextFromResponse(answer))

    // 4. 附上来源链接
    fmt.Println("\n参考来源：")
    for i, result := range searchResp.Results {
        fmt.Printf("%d. %s - %s\n", i+1, result.Title, result.URL)
    }
}

func buildSearchContext(results []search.SearchResult) string {
    var context string
    for i, result := range results {
        context += fmt.Sprintf("[%d] %s\n%s\n\n", i+1, result.Title, result.Description)
    }
    return context
}
```

### 实时知识库更新

```go
// 定期搜索并更新知识库
func updateKnowledgeBase() {
    topics := []string{
        "AI breakthroughs 2025",
        "machine learning trends",
        "deep learning applications",
    }

    for _, topic := range topics {
        // 搜索最新内容
        results, _ := searchProvider.Search(ctx, &search.SearchRequest{
            Query:      topic,
            Freshness:  "day",
            NumResults: 10,
        })

        // 提取内容
        for _, result := range results.Results {
            content := extractContent(result)

            // 生成嵌入
            embedding, _ := embeddingProvider.CreateEmbedding(ctx, &ai.EmbeddingRequest{
                Input: []string{content},
            })

            // 存入向量数据库
            vectorDB.Insert(ctx, "knowledge_base", "",
                entity.NewColumnInt64("id", []int64{generateID()}),
                entity.NewColumnVarChar("content", []string{content}),
                entity.NewColumnFloatVector("embedding", 768, embedding.Data[0].Embedding),
            )
        }
    }
}
```

## 注意事项

1. **API 配额**：控制搜索频率，避免超出限制
2. **内容过滤**：使用 SafeSearch 过滤不当内容
3. **延迟优化**：缓存常见查询结果
4. **成本控制**：合理设置 NumResults
5. **合规性**：遵守各搜索引擎的使用条款

## 未来扩展

- [ ] 支持图片搜索
- [ ] 支持新闻搜索
- [ ] 支持学术论文搜索
- [ ] 支持视频搜索
- [ ] 集成网页爬虫
- [ ] 支持本地搜索引擎（Elasticsearch）
