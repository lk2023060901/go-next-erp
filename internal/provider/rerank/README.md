# internal/provider/rerank - Rerank 服务提供商接口（预留）

## 概述

Rerank（重排序）服务用于优化搜索结果的相关性排序。通过语义理解对初步检索的文档进行重新排序，提升搜索质量。

## 核心接口

```go
type Provider interface {
    // Rerank 重排序
    Rerank(ctx context.Context, req *RerankRequest) (*RerankResponse, error)

    // 工具方法
    GetProviderName() string
    Close() error
}
```

## 使用场景

### 1. 搜索结果优化

```go
// 初步检索（例如向量搜索）得到候选文档
candidates := []string{
    "文档1：关于人工智能的介绍",
    "文档2：机器学习基础知识",
    "文档3：深度学习应用实例",
    "文档4：自然语言处理技术",
    "文档5：计算机视觉算法",
}

// 使用 Rerank 优化排序
provider, _ := rerank.New(rerank.ProviderTypeCohere, &rerank.Config{
    BaseURL: "https://api.cohere.ai/v1",
    APIKey:  "your-api-key",
})

req := &rerank.RerankRequest{
    Model:     "rerank-multilingual-v2.0",
    Query:     "深度学习在 NLP 中的应用",
    Documents: candidates,
    TopN:      3,
    ReturnDocuments: true,
}

resp, _ := provider.Rerank(ctx, req)

// 输出重排序后的结果
for _, result := range resp.Results {
    fmt.Printf("分数: %.4f, 索引: %d, 文档: %s\n",
        result.RelevanceScore,
        result.Index,
        result.Document,
    )
}
```

### 2. RAG 系统集成

```go
// 步骤 1: 向量检索
vectorResults := vectorDB.Search(query, topK=100)

// 步骤 2: Rerank 精排
documents := extractDocuments(vectorResults)
rerankReq := &rerank.RerankRequest{
    Query:     query,
    Documents: documents,
    TopN:      10,
}

rerankResp, _ := rerankProvider.Rerank(ctx, rerankReq)

// 步骤 3: 使用重排序后的文档生成回答
topDocuments := getTopDocuments(rerankResp.Results)
answer := llm.Generate(query, topDocuments)
```

### 3. 多路召回融合

```go
// 从多个来源召回候选
keywordResults := keywordSearch(query, 50)
vectorResults := vectorSearch(query, 50)
bm25Results := bm25Search(query, 50)

// 合并去重
allCandidates := merge(keywordResults, vectorResults, bm25Results)

// Rerank 统一排序
rerankResp, _ := provider.Rerank(ctx, &rerank.RerankRequest{
    Query:     query,
    Documents: allCandidates,
    TopN:      20,
})
```

## 支持的提供商

### Cohere

```go
provider, _ := rerank.New(rerank.ProviderTypeCohere, &rerank.Config{
    BaseURL: "https://api.cohere.ai/v1",
    APIKey:  "your-cohere-api-key",
    Model:   "rerank-multilingual-v2.0",
})
```

**模型：**
- `rerank-english-v2.0` - 英文优化
- `rerank-multilingual-v2.0` - 多语言支持

### Jina AI

```go
provider, _ := rerank.New(rerank.ProviderTypeJina, &rerank.Config{
    BaseURL: "https://api.jina.ai/v1",
    APIKey:  "your-jina-api-key",
    Model:   "jina-reranker-v1-base-en",
})
```

**模型：**
- `jina-reranker-v1-base-en` - 英文基础模型
- `jina-reranker-v1-turbo-en` - 英文高速模型

### Voyage AI

```go
provider, _ := rerank.New(rerank.ProviderTypeVoyage, &rerank.Config{
    BaseURL: "https://api.voyageai.com/v1",
    APIKey:  "your-voyage-api-key",
    Model:   "rerank-lite-1",
})
```

**模型：**
- `rerank-lite-1` - 轻量级模型
- `rerank-1` - 标准模型

### 自定义实现

```go
// 实现自己的 Rerank 服务
type CustomReranker struct {
    // ...
}

func (c *CustomReranker) Rerank(ctx context.Context, req *rerank.RerankRequest) (*rerank.RerankResponse, error) {
    // 自定义重排序逻辑
    // 例如：基于 BM25、TF-IDF 或自训练模型
}

// 注册
func init() {
    rerank.Register(rerank.ProviderTypeCustom, func(config *rerank.Config) (rerank.Provider, error) {
        return NewCustomReranker(config)
    })
}
```

## 请求参数

```go
type RerankRequest struct {
    Model           string   // 模型名称
    Query           string   // 查询文本
    Documents       []string // 文档列表（候选集）
    TopN            int      // 返回前 N 个结果
    ReturnDocuments bool     // 是否返回文档内容
}
```

## 响应格式

```go
type RerankResponse struct {
    ID      string         // 请求 ID
    Model   string         // 使用的模型
    Results []RerankResult // 重排序结果
    Usage   Usage          // 使用统计
}

type RerankResult struct {
    Index          int     // 原始文档索引
    RelevanceScore float64 // 相关性分数 (0-1)
    Document       string  // 文档内容（可选）
}
```

## 最佳实践

### 1. 两阶段检索

```go
// 第一阶段：快速粗排（向量检索）
candidates := vectorDB.Search(query, topK=100) // 召回 100 个

// 第二阶段：精确重排
finalResults := reranker.Rerank(query, candidates, topN=10) // 精排前 10
```

### 2. 批量处理

```go
// 批处理多个查询
queries := []string{"查询1", "查询2", "查询3"}
documents := getSharedDocuments() // 共享文档集

for _, query := range queries {
    resp, _ := provider.Rerank(ctx, &rerank.RerankRequest{
        Query:     query,
        Documents: documents,
        TopN:      5,
    })
    // 处理结果...
}
```

### 3. 性能优化

```go
// 控制候选集大小
maxCandidates := 100
if len(candidates) > maxCandidates {
    candidates = candidates[:maxCandidates]
}

// 并发处理
var wg sync.WaitGroup
results := make(chan *rerank.RerankResponse, len(queries))

for _, query := range queries {
    wg.Add(1)
    go func(q string) {
        defer wg.Done()
        resp, _ := provider.Rerank(ctx, &rerank.RerankRequest{
            Query: q,
            Documents: documents,
            TopN: 10,
        })
        results <- resp
    }(query)
}
```

### 4. 错误处理

```go
resp, err := provider.Rerank(ctx, req)
if err != nil {
    switch {
    case errors.Is(err, rerank.ErrRateLimitExceeded):
        // 处理速率限制
        time.Sleep(time.Second)
        // 重试...
    case errors.Is(err, rerank.ErrInsufficientQuota):
        // 处理配额不足
    default:
        // 其他错误
    }
}
```

## 集成示例

### RAG 完整流程

```go
package main

import (
    "context"
    "fmt"

    "github.com/lk2023060901/go-next-erp/internal/provider/ai"
    "github.com/lk2023060901/go-next-erp/internal/provider/rerank"
    "github.com/lk2023060901/go-next-erp/pkg/vector"
)

func main() {
    ctx := context.Background()

    // 1. 初始化组件
    vectorDB, _ := vector.New(ctx)
    rerankProvider, _ := rerank.New(rerank.ProviderTypeCohere, &rerank.Config{
        APIKey: "your-api-key",
    })
    llm, _ := ai.New(ai.ProviderTypeOpenAI, &ai.Config{
        APIKey: "your-api-key",
    })

    // 2. 用户查询
    query := "什么是深度学习？"

    // 3. 向量检索（粗排）
    vectorResults, _ := vectorDB.Search(ctx, "knowledge_base", []string{}, query, []string{"content"}, nil, "embedding", nil, 100, nil)

    // 4. 提取候选文档
    candidates := extractDocuments(vectorResults)

    // 5. Rerank 重排序（精排）
    rerankResp, _ := rerankProvider.Rerank(ctx, &rerank.RerankRequest{
        Query:     query,
        Documents: candidates,
        TopN:      5,
        ReturnDocuments: true,
    })

    // 6. 构建上下文
    context := buildContext(rerankResp.Results)

    // 7. LLM 生成回答
    answer, _ := llm.CreateCompletion(ctx, &ai.CompletionRequest{
        Model: "gpt-4",
        Messages: []ai.Message{
            ai.NewSystemMessage("你是一个专业的AI助手，请基于提供的上下文回答问题。"),
            ai.NewUserTextMessage(fmt.Sprintf("上下文：\n%s\n\n问题：%s", context, query)),
        },
    })

    fmt.Println(ai.GetTextFromResponse(answer))
}
```

## 注意事项

1. **候选集大小**：通常建议 50-200 个候选文档
2. **TopN 选择**：根据下游任务需求，通常取 3-20 个
3. **成本控制**：Rerank 按 token 计费，控制文档长度
4. **延迟权衡**：Rerank 增加延迟，需平衡精度和速度
5. **模型选择**：根据语言和场景选择合适的模型

## 未来扩展

- [ ] 支持批量 Rerank
- [ ] 支持自定义评分函数
- [ ] 支持多字段重排序
- [ ] 支持跨语言重排序
- [ ] 集成本地模型部署
