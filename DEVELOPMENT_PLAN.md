# 基础服务层封装开发计划

## 一、架构概览

```
┌─────────────────────────────────────────────────────────┐
│                   基础服务封装层                          │
├──────────────┬──────────────┬──────────────┬───────────┤
│  对象存储     │  向量数据库   │  文档处理    │  AI服务   │
│  (MinIO)     │  (Milvus)    │ (MinerU API)│ (Provider)│
└──────────────┴──────────────┴──────────────┴───────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│              Provider 接口层 (统一多模态)                 │
├──────────────┬──────────────┬──────────────────────────┤
│ AI Provider  │  Rerank      │  Search Engine           │
│ (多模态)      │  Provider    │  Provider (预留)          │
│              │  (预留接口)   │                          │
└──────────────┴──────────────┴──────────────────────────┘
```

---

## 二、核心模块

### 2.1 pkg/storage - MinIO 对象存储

**目录结构**：
```
pkg/storage/
├── config.go
├── options.go
├── storage.go
├── storage_test.go
└── USAGE.md
```

**核心接口**：
```go
type Storage interface {
    UploadFile(ctx context.Context, bucket, key string, reader io.Reader, size int64) error
    DownloadFile(ctx context.Context, bucket, key, filePath string) error
    DeleteFile(ctx context.Context, bucket, key string) error
    GetPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
    ListFiles(ctx context.Context, bucket, prefix string) ([]FileInfo, error)
    CreateBucket(ctx context.Context, bucket string) error
    Close() error
}
```

---

### 2.2 pkg/vector - Milvus 向量数据库

**目录结构**：
```
pkg/vector/
├── config.go
├── options.go
├── vector.go
├── vector_test.go
└── USAGE.md
```

**核心接口**：
```go
type Vector interface {
    CreateCollection(ctx context.Context, name string, dim int) error
    Insert(ctx context.Context, collection string, vectors [][]float32, ids []int64) ([]int64, error)
    Search(ctx context.Context, collection string, vectors [][]float32, topK int) ([]SearchResult, error)
    Delete(ctx context.Context, collection string, ids []int64) error
    CreateIndex(ctx context.Context, collection, field, indexType string) error
    Close() error
}
```

---

### 2.3 internal/document - 文档处理（MinerU HTTP API）

**目录结构**：
```
internal/document/
├── parser/
│   ├── mineru_client.go    # MinerU HTTP 客户端
│   ├── parser.go            # 解析接口
│   └── types.go
├── chunker/
│   ├── chunker.go
│   └── types.go
└── service.go
```

**MinerU HTTP 客户端**：
```go
// MinerU HTTP API 客户端
type MinerUClient struct {
    baseURL string
    apiKey  string
    client  *http.Client
    logger  *logger.Logger
}

// MinerU API 请求
type ParseRequest struct {
    FileURL  string `json:"file_url"`   // 文件 URL
    FileType string `json:"file_type"`  // pdf, docx, pptx
    Options  ParseOptions `json:"options,omitempty"`
}

type ParseOptions struct {
    ExtractImages bool `json:"extract_images"` // 是否提取图片
    ExtractTables bool `json:"extract_tables"` // 是否提取表格
    OCR           bool `json:"ocr"`            // 是否启用 OCR
}

// MinerU API 响应
type ParseResponse struct {
    TaskID   string `json:"task_id"`
    Status   string `json:"status"`   // pending, processing, completed, failed
    Text     string `json:"text"`
    Markdown string `json:"markdown"`
    Images   []ImageInfo `json:"images,omitempty"`
    Tables   []TableInfo `json:"tables,omitempty"`
}

type ImageInfo struct {
    Index int    `json:"index"`
    URL   string `json:"url"`
    Page  int    `json:"page"`
}

type TableInfo struct {
    Index   int      `json:"index"`
    Content string   `json:"content"`  // Markdown 格式
    Headers []string `json:"headers"`
    Rows    [][]string `json:"rows"`
    Page    int      `json:"page"`
}

// MinerU Client 实现
func NewMinerUClient(baseURL, apiKey string) Parser {
    return &MinerUClient{
        baseURL: baseURL,
        apiKey:  apiKey,
        client:  &http.Client{Timeout: 300 * time.Second}, // 5分钟超时
        logger:  logger.GetLogger().With(zap.String("client", "mineru")),
    }
}

// 解析文档（同步）
func (c *MinerUClient) Parse(ctx context.Context, fileURL string) (*ParseResult, error) {
    // 1. 提交解析任务
    req := &ParseRequest{
        FileURL:  fileURL,
        FileType: detectFileType(fileURL),
        Options: ParseOptions{
            ExtractImages: true,
            ExtractTables: true,
            OCR:           true,
        },
    }

    taskID, err := c.submitTask(ctx, req)
    if err != nil {
        return nil, err
    }

    // 2. 轮询任务状态
    for {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(2 * time.Second):
            resp, err := c.getTaskStatus(ctx, taskID)
            if err != nil {
                return nil, err
            }

            switch resp.Status {
            case "completed":
                return c.convertResponse(resp), nil
            case "failed":
                return nil, fmt.Errorf("parse failed")
            case "processing", "pending":
                continue
            }
        }
    }
}

// 提交解析任务
func (c *MinerUClient) submitTask(ctx context.Context, req *ParseRequest) (string, error) {
    url := fmt.Sprintf("%s/api/v1/parse", c.baseURL)

    body, _ := json.Marshal(req)
    httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

    resp, err := c.client.Do(httpReq)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        TaskID string `json:"task_id"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    return result.TaskID, nil
}

// 获取任务状态
func (c *MinerUClient) getTaskStatus(ctx context.Context, taskID string) (*ParseResponse, error) {
    url := fmt.Sprintf("%s/api/v1/tasks/%s", c.baseURL, taskID)

    httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

    resp, err := c.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result ParseResponse
    json.NewDecoder(resp.Body).Decode(&result)

    return &result, nil
}

// 支持的文件类型
func (c *MinerUClient) SupportedTypes() []string {
    return []string{"pdf", "docx", "pptx"}
}
```

**Parser 接口**：
```go
// Parser 文档解析接口
type Parser interface {
    Parse(ctx context.Context, fileURL string) (*ParseResult, error)
    SupportedTypes() []string
}

// ParseResult 解析结果
type ParseResult struct {
    Text     string      // 纯文本
    Markdown string      // Markdown 格式
    Images   []ImageInfo // 图片列表
    Tables   []TableInfo // 表格列表
}
```

**文本分块器**：
```go
type Chunker interface {
    Chunk(ctx context.Context, text string, size int, overlap int) ([]Chunk, error)
}

type Chunk struct {
    Index   int
    Content string
    Start   int
    End     int
}

// 固定大小分块实现
type FixedSizeChunker struct{}

func (c *FixedSizeChunker) Chunk(ctx context.Context, text string, size int, overlap int) ([]Chunk, error) {
    chunks := []Chunk{}
    runes := []rune(text)
    totalLen := len(runes)

    for i := 0; i < totalLen; i += (size - overlap) {
        end := i + size
        if end > totalLen {
            end = totalLen
        }

        chunks = append(chunks, Chunk{
            Index:   len(chunks),
            Content: string(runes[i:end]),
            Start:   i,
            End:     end,
        })

        if end >= totalLen {
            break
        }
    }

    return chunks, nil
}
```

---

### 2.4 internal/provider/ai - 多模态 AI Provider

**（与之前设计保持一致，支持文本、图像、音频、视频多模态）**

**目录结构**：
```
internal/provider/ai/
├── provider.go         # 统一接口定义
├── types.go            # 多模态类型定义
├── openai.go           # OpenAI 实现
├── siliconflow.go      # 硅基流动实现
├── anthropic.go        # Anthropic 实现
├── factory.go          # Provider 工厂
└── README.md
```

**核心类型**（简化版）：
```go
// 内容类型
type ContentType string

const (
    ContentTypeText  ContentType = "text"
    ContentTypeImage ContentType = "image"
    ContentTypeAudio ContentType = "audio"
    ContentTypeVideo ContentType = "video"
)

// 统一内容结构
type Content struct {
    Type     ContentType `json:"type"`
    Text     string      `json:"text,omitempty"`
    ImageURL string      `json:"image_url,omitempty"`
    AudioURL string      `json:"audio_url,omitempty"`
    VideoURL string      `json:"video_url,omitempty"`
}

// 消息结构
type Message struct {
    Role    string    `json:"role"`
    Content []Content `json:"content"`
}

// AI Provider 接口
type AIProvider interface {
    // 文本嵌入
    CreateEmbedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

    // 聊天补全（支持多模态）
    ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

    // 语音识别（可选）
    SpeechToText(ctx context.Context, req *SpeechToTextRequest) (*SpeechToTextResponse, error)

    // 语音合成（可选）
    TextToSpeech(ctx context.Context, req *TextToSpeechRequest) (*TextToSpeechResponse, error)

    Name() string
    SupportedModalities() []ContentType
}
```

---

### 2.5 internal/provider/rerank - Rerank Provider（预留）

```go
type RerankProvider interface {
    Rerank(ctx context.Context, req *RerankRequest) (*RerankResponse, error)
    Name() string
}

type RerankRequest struct {
    Query     string   `json:"query"`
    Documents []string `json:"documents"`
    TopN      int      `json:"top_n"`
}

type RerankResponse struct {
    Results []RerankResult `json:"results"`
}

type RerankResult struct {
    Index int     `json:"index"`
    Score float32 `json:"score"`
}

// 支持外部服务商（Cohere、Jina）和自实现
type RerankConfig struct {
    Provider string // "cohere", "jina", "internal", "custom"
    APIKey   string
    BaseURL  string
}
```

---

### 2.6 internal/provider/search - Search Engine Provider（预留）

```go
type SearchProvider interface {
    Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
    Name() string
}

type SearchRequest struct {
    Query     string                 `json:"query"`
    TopK      int                    `json:"top_k"`
    Filters   map[string]interface{} `json:"filters,omitempty"`
    Embedding []float32              `json:"embedding,omitempty"`
}

type SearchResponse struct {
    Results []SearchResult `json:"results"`
}

type SearchResult struct {
    ID      string  `json:"id"`
    Score   float32 `json:"score"`
    Content string  `json:"content"`
}

// 支持外部服务商和自实现
type SearchConfig struct {
    Provider string // "elasticsearch", "meilisearch", "internal"
    APIKey   string
    BaseURL  string
}
```

---

## 三、环境配置

### 3.1 服务端口
```
MinIO:         15002 (API), 15003 (Console)
Milvus:        15004
PostgreSQL:    15000
Redis:         15001
MinerU API:    外部服务（需配置 BaseURL）
```

### 3.2 环境变量
```env
# MinIO
MINIO_ENDPOINT=localhost:15002
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123

# Milvus
MILVUS_HOST=localhost
MILVUS_PORT=15004

# MinerU API
MINERU_BASE_URL=https://api.mineru.com
MINERU_API_KEY=your_api_key

# AI Provider
AI_PROVIDER=openai
OPENAI_API_KEY=sk-xxx
SILICONFLOW_API_KEY=sk-xxx
ANTHROPIC_API_KEY=sk-xxx
```

---

## 四、使用示例

### 4.1 MinerU 文档解析
```go
// 初始化 MinerU 客户端
mineruClient := document.NewMinerUClient(
    "https://api.mineru.com",
    "your_api_key",
)

// 解析文档
result, err := mineruClient.Parse(ctx, "https://example.com/document.pdf")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Text:", result.Text)
fmt.Println("Markdown:", result.Markdown)
fmt.Println("Images:", len(result.Images))
fmt.Println("Tables:", len(result.Tables))
```

### 4.2 多模态 AI Provider
```go
// 初始化 OpenAI Provider
provider, _ := ai.NewAIProvider(&ai.ProviderConfig{
    Type:   "openai",
    APIKey: "sk-xxx",
})

// 文本 + 图像多模态对话
req := &ai.ChatRequest{
    Model: "gpt-4-vision-preview",
    Messages: []ai.Message{
        {
            Role: "user",
            Content: []ai.Content{
                {Type: ai.ContentTypeText, Text: "这张图片里有什么？"},
                {Type: ai.ContentTypeImage, ImageURL: "https://example.com/image.jpg"},
            },
        },
    },
}

resp, _ := provider.ChatCompletion(ctx, req)
fmt.Println(resp.Choices[0].Message.Content[0].Text)
```

---

## 五、开发任务清单

### 5.1 基础服务封装
- [ ] pkg/storage - MinIO 客户端
- [ ] pkg/vector - Milvus 客户端

### 5.2 文档处理
- [ ] internal/document/parser - MinerU HTTP 客户端
- [ ] internal/document/chunker - 文本分块器

### 5.3 AI Provider
- [ ] 定义多模态统一接口
- [ ] 实现 OpenAI Provider
- [ ] 实现硅基流动 Provider
- [ ] 实现 Anthropic Provider

### 5.4 预留接口
- [ ] Rerank Provider 接口定义
- [ ] Search Engine Provider 接口定义

---

## 六、依赖安装

```bash
# MinIO SDK
go get github.com/minio/minio-go/v7

# Milvus SDK
go get github.com/milvus-io/milvus-sdk-go/v2
```

---

## 七、交付标准

1. ✅ 代码实现完整
2. ✅ 单元测试覆盖率 > 80%
3. ✅ 接口设计统一、可扩展
4. ✅ 支持多模态（文本、图像、音频、视频）
5. ✅ 使用文档清晰
