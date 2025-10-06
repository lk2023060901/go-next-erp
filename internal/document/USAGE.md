## internal/document - MinerU 文档解析使用文档

### 一、快速开始

#### 1.1 初始化客户端

```go
package main

import (
    "context"
    "github.com/lk2023060901/go-next-erp/internal/document"
)

func main() {
    ctx := context.Background()

    // 方式 1：使用默认配置
    c, err := document.New(
        document.WithAPIKey("your-api-key"),
    )
    if err != nil {
        panic(err)
    }
    defer c.Close()

    // 方式 2：自定义配置
    c, err = document.New(
        document.WithBaseURL("https://mineru.net/api/v4"),
        document.WithAPIKey("your-api-key"),
        document.WithPollInterval(3*time.Second),
    )
}
```

#### 1.2 环境变量配置

```env
MINERU_BASE_URL=https://mineru.net/api/v4
MINERU_API_KEY=your-api-key
```

---

### 二、单个文件解析

#### 2.1 创建解析任务

```go
// 创建任务请求
req := &document.CreateTaskRequest{
    URL:           "https://example.com/document.pdf",
    IsOCR:         true,
    EnableFormula: true,
    EnableTable:   true,
    Language:      "ch",
}

// 创建任务
taskID, err := c.CreateTask(ctx, req)
if err != nil {
    panic(err)
}
fmt.Printf("Task created: %s\n", taskID)
```

#### 2.2 查询任务状态

```go
// 获取任务结果
result, err := c.GetTaskResult(ctx, taskID)
if err != nil {
    panic(err)
}

switch result.State {
case document.TaskStateDone:
    fmt.Printf("Task completed! Download URL: %s\n", result.FullZipURL)
case document.TaskStateRunning:
    fmt.Println("Task is running...")
case document.TaskStateFailed:
    fmt.Printf("Task failed: %s\n", result.ErrorMsg)
}
```

#### 2.3 等待任务完成

```go
// 自动轮询等待任务完成
result, err := c.WaitForTask(ctx, taskID)
if err != nil {
    panic(err)
}

fmt.Printf("Task completed! Download URL: %s\n", result.FullZipURL)
```

#### 2.4 完整示例

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/lk2023060901/go-next-erp/internal/document"
)

func main() {
    ctx := context.Background()

    // 初始化客户端
    c, err := document.New(
        document.WithAPIKey("your-api-key"),
    )
    if err != nil {
        panic(err)
    }
    defer c.Close()

    // 创建任务
    req := &document.CreateTaskRequest{
        URL:           "https://cdn-mineru.openxlab.org.cn/demo/example.pdf",
        IsOCR:         true,
        EnableFormula: false,
        EnableTable:   true,
        Language:      "ch",
        ExtraFormats:  []string{"docx", "html"},
    }

    taskID, err := c.CreateTask(ctx, req)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Task created: %s\n", taskID)

    // 等待完成
    result, err := c.WaitForTask(ctx, taskID)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Download URL: %s\n", result.FullZipURL)
}
```

---

### 三、批量 URL 解析

#### 3.1 批量提交任务

```go
req := &document.CreateBatchURLRequest{
    EnableFormula: true,
    Language:      "ch",
    EnableTable:   true,
    Files: []document.BatchURLItem{
        {
            URL:    "https://example.com/doc1.pdf",
            IsOCR:  true,
            DataID: "doc-1",
        },
        {
            URL:    "https://example.com/doc2.pdf",
            IsOCR:  false,
            DataID: "doc-2",
        },
    },
}

batchID, err := c.CreateBatchURL(ctx, req)
if err != nil {
    panic(err)
}
fmt.Printf("Batch created: %s\n", batchID)
```

#### 3.2 查询批量结果

```go
results, err := c.GetBatchResult(ctx, batchID)
if err != nil {
    panic(err)
}

for _, result := range results {
    fmt.Printf("File: %s, State: %s\n", result.FileName, result.State)
    if result.State == document.TaskStateDone {
        fmt.Printf("  Download: %s\n", result.FullZipURL)
    }
    if result.ExtractProgress != nil {
        fmt.Printf("  Progress: %d/%d pages\n",
            result.ExtractProgress.ExtractedPages,
            result.ExtractProgress.TotalPages,
        )
    }
}
```

#### 3.3 等待批量完成

```go
results, err := c.WaitForBatch(ctx, batchID)
if err != nil {
    panic(err)
}

for _, result := range results {
    if result.State == document.TaskStateDone {
        fmt.Printf("%s completed: %s\n", result.FileName, result.FullZipURL)
    } else {
        fmt.Printf("%s failed: %s\n", result.FileName, result.ErrorMsg)
    }
}
```

---

### 四、批量文件上传解析

#### 4.1 申请上传 URL

```go
req := &document.CreateBatchUploadRequest{
    EnableFormula: true,
    Language:      "ch",
    EnableTable:   true,
    Files: []document.BatchFileItem{
        {
            Name:   "document1.pdf",
            IsOCR:  true,
            DataID: "doc-1",
        },
        {
            Name:   "document2.pdf",
            IsOCR:  false,
            DataID: "doc-2",
        },
    },
}

batchID, uploadURLs, err := c.CreateBatchUpload(ctx, req)
if err != nil {
    panic(err)
}

fmt.Printf("Batch ID: %s\n", batchID)
fmt.Printf("Upload URLs: %v\n", uploadURLs)
```

#### 4.2 上传文件

```go
filePaths := []string{
    "/path/to/document1.pdf",
    "/path/to/document2.pdf",
}

for i, uploadURL := range uploadURLs {
    err := c.UploadFile(ctx, uploadURL, filePaths[i])
    if err != nil {
        fmt.Printf("Upload failed for %s: %v\n", filePaths[i], err)
        continue
    }
    fmt.Printf("Uploaded: %s\n", filePaths[i])
}
```

#### 4.3 等待解析完成

```go
results, err := c.WaitForBatch(ctx, batchID)
if err != nil {
    panic(err)
}

for _, result := range results {
    fmt.Printf("%s: %s\n", result.FileName, result.State)
    if result.State == document.TaskStateDone {
        fmt.Printf("  Download: %s\n", result.FullZipURL)
    }
}
```

---

### 五、高级选项

#### 5.1 请求参数说明

```go
req := &document.CreateTaskRequest{
    URL:           "https://example.com/doc.pdf",  // 文件 URL (必填)
    IsOCR:         true,                           // 启用 OCR
    EnableFormula: true,                           // 启用公式识别
    EnableTable:   true,                           // 启用表格识别
    Language:      "ch",                           // 文档语言: ch, en
    DataID:        "custom-id",                    // 自定义数据 ID
    Callback:      "https://api.example.com/cb",   // 回调 URL
    Seed:          "random-seed",                  // 签名随机字符串
    ExtraFormats:  []string{"docx", "html"},       // 额外导出格式
    PageRanges:    "1-10,20-30",                   // 指定页码范围
    ModelVersion:  "vlm",                          // 模型版本: pipeline, vlm
}
```

#### 5.2 支持的文件格式

- PDF: `.pdf`
- Word: `.doc`, `.docx`
- PowerPoint: `.ppt`, `.pptx`
- 图片: `.png`, `.jpg`, `.jpeg`

#### 5.3 额外导出格式

```go
ExtraFormats: []string{
    "docx",  // Word 文档
    "html",  // HTML 网页
    "latex", // LaTeX 文档
}
```

---

### 六、错误处理

#### 6.1 常见错误

```go
_, err := c.CreateTask(ctx, req)
if err != nil {
    // 检查错误类型
    switch {
    case strings.Contains(err.Error(), "A0202"):
        fmt.Println("Token 错误，请检查 API Key")
    case strings.Contains(err.Error(), "A0211"):
        fmt.Println("Token 过期，请更新 API Key")
    case strings.Contains(err.Error(), "-60005"):
        fmt.Println("文件大小超出限制 (最大 200MB)")
    case strings.Contains(err.Error(), "-60006"):
        fmt.Println("文件页数超过限制 (最大 600 页)")
    default:
        fmt.Printf("其他错误: %v\n", err)
    }
}
```

#### 6.2 错误码说明

| 错误码 | 说明 | 解决方法 |
|--------|------|----------|
| A0202 | Token 错误 | 检查 API Key |
| A0211 | Token 过期 | 更新 API Key |
| -60005 | 文件大小超限 | 最大 200MB |
| -60006 | 页数超限 | 最大 600 页 |
| -60008 | 文件读取超时 | 检查 URL 可访问性 |
| -60010 | 解析失败 | 稍后重试 |

---

### 七、配置参考

#### 7.1 客户端配置

```go
c, err := document.New(
    document.WithBaseURL("https://mineru.net/api/v4"),
    document.WithAPIKey("your-api-key"),
    document.WithTimeout(30*time.Second),        // HTTP 请求超时
    document.WithPollInterval(5*time.Second),    // 轮询间隔
    document.WithPollTimeout(30*time.Minute),    // 轮询总超时
    document.WithUploadTimeout(10*time.Minute),  // 上传超时
    document.WithMaxRetries(3),                  // 最大重试次数
)
```

#### 7.2 配置项说明

| 选项 | 说明 | 默认值 |
|------|------|--------|
| BaseURL | API 基础地址 | https://mineru.net/api/v4 |
| APIKey | API Key (Token) | "" |
| Timeout | HTTP 请求超时 | 30s |
| PollInterval | 轮询间隔 | 5s |
| PollTimeout | 轮询总超时 | 30m |
| UploadTimeout | 上传超时 | 10m |
| MaxRetries | 最大重试次数 | 3 |

---

### 八、最佳实践

#### 8.1 使用上下文控制

```go
// 设置超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result, err := c.WaitForTask(ctx, taskID)
```

#### 8.2 批量处理优化

```go
// 分批处理大量文件
const batchSize = 10

for i := 0; i < len(files); i += batchSize {
    end := i + batchSize
    if end > len(files) {
        end = len(files)
    }

    batch := files[i:end]
    // 提交批次
    batchID, err := c.CreateBatchURL(ctx, &document.CreateBatchURLRequest{
        Files: batch,
    })
    // 处理批次结果...
}
```

#### 8.3 回调通知

```go
req := &document.CreateTaskRequest{
    URL:      "https://example.com/doc.pdf",
    Callback: "https://api.example.com/mineru/callback",
    Seed:     "random-seed-for-signature",
}

// 在回调服务中验证签名并处理结果
```

---

### 九、限制说明

- **文件大小**：单个文件不超过 200MB
- **页数限制**：单个文件不超过 600 页
- **每日额度**：每个账号每天 2000 页最高优先级额度
- **网络限制**：GitHub、AWS 等国外 URL 可能超时
- **上传方式**：仅支持 URL 提交或预签名 URL 上传，不支持直接上传
