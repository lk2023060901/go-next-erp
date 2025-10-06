# internal/provider/ai - 多模态 AI Provider 统一接口

## 概述

提供统一的 AI 服务提供商接口，支持多模态输入输出（文本、图像、音频、视频）。

## 支持的模态

- **文本 (Text)**: 文本生成、对话、嵌入
- **图像 (Image)**: 图像理解、图像生成
- **音频 (Audio)**: 语音转文本
- **视频 (Video)**: 视频理解、视频生成

## 核心接口

### Provider 接口

```go
type Provider interface {
    // 文本生成（支持多模态输入）
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
```

## 使用示例

### 1. 文本对话

```go
provider, _ := ai.New(ai.ProviderTypeOpenAI, &ai.Config{
    BaseURL: "https://api.openai.com/v1",
    APIKey:  "your-api-key",
})

req := &ai.CompletionRequest{
    Model: "gpt-4",
    Messages: []ai.Message{
        ai.NewUserTextMessage("你好，请介绍一下自己"),
    },
}

resp, _ := provider.CreateCompletion(ctx, req)
fmt.Println(ai.GetTextFromResponse(resp))
```

### 2. 图像理解

```go
req := &ai.CompletionRequest{
    Model: "gpt-4-vision-preview",
    Messages: []ai.Message{
        ai.NewUserMessage(
            ai.NewTextContent("这张图片里有什么？"),
            ai.NewImageContentFromURL("https://example.com/image.jpg", "high"),
        ),
    },
}

resp, _ := provider.CreateCompletion(ctx, req)
fmt.Println(ai.GetTextFromResponse(resp))
```

### 3. 视频理解

```go
req := &ai.VideoAnalysisRequest{
    Model:    "gpt-4-video",
    VideoURL: "https://example.com/video.mp4",
    Prompt:   "描述这个视频的主要内容",
    MaxFrames: 10,
}

resp, _ := provider.AnalyzeVideo(ctx, req)
fmt.Println(resp.Analysis)

// 查看帧级别信息
for _, frame := range resp.Frames {
    fmt.Printf("时间 %.2fs: %s\n", frame.Timestamp, frame.Description)
}
```

### 4. 音频转文本

```go
audioData, _ := os.ReadFile("speech.mp3")

req := &ai.TranscriptionRequest{
    Model:    "whisper-1",
    File:     audioData,
    Language: "zh",
}

resp, _ := provider.CreateTranscription(ctx, req)
fmt.Println(resp.Text)
```

### 5. 图像生成

```go
req := &ai.ImageGenerationRequest{
    Prompt:  "一只在太空中的猫",
    Model:   "dall-e-3",
    Size:    "1024x1024",
    Quality: "hd",
    N:       1,
}

resp, _ := provider.GenerateImage(ctx, req)
fmt.Println(resp.Data[0].URL)
```

### 6. 视频生成

```go
req := &ai.VideoGenerationRequest{
    Prompt:      "一只猫在草地上奔跑",
    Duration:    5,
    FPS:         30,
    Resolution:  "1080p",
    AspectRatio: "16:9",
}

resp, _ := provider.GenerateVideo(ctx, req)
fmt.Println(resp.Data[0].URL)
```

### 7. 多模态组合

```go
// 文本 + 图像 + 音频输入
req := &ai.CompletionRequest{
    Model: "gpt-4-omni",
    Messages: []ai.Message{
        ai.NewUserMessage(
            ai.NewTextContent("请分析这张图片和这段音频的关联性"),
            ai.NewImageContentFromURL("https://example.com/scene.jpg"),
            ai.NewAudioContentFromURL("https://example.com/sound.mp3"),
        ),
    },
}

resp, _ := provider.CreateCompletion(ctx, req)
```

### 8. 文本嵌入

```go
req := &ai.EmbeddingRequest{
    Model: "text-embedding-3-small",
    Input: []string{
        "人工智能技术的发展",
        "机器学习算法原理",
    },
}

resp, _ := provider.CreateEmbedding(ctx, req)
for _, emb := range resp.Data {
    fmt.Printf("向量维度: %d\n", len(emb.Embedding))
}
```

### 9. 流式输出

```go
req := &ai.CompletionRequest{
    Model:  "gpt-4",
    Stream: true,
    Messages: []ai.Message{
        ai.NewUserTextMessage("写一首关于春天的诗"),
    },
}

stream, _ := provider.CreateCompletionStream(ctx, req)
defer stream.Close()

// 读取流式数据（需要自行解析 SSE 格式）
```

### 10. 检查提供商能力

```go
capabilities := provider.GetCapabilities()

if capabilities.SupportVideoInput {
    fmt.Println("支持视频输入")
}

if capabilities.SupportVideoGeneration {
    fmt.Println("支持视频生成")
}
```

## 内容类型

### ContentType

- `ContentTypeText`: 文本内容
- `ContentTypeImage`: 图像内容
- `ContentTypeAudio`: 音频内容
- `ContentTypeVideo`: 视频内容

### 创建内容辅助函数

```go
// 文本
ai.NewTextContent("文本内容")

// 图像（URL）
ai.NewImageContentFromURL("https://example.com/image.jpg", "high")

// 图像（Base64）
ai.NewImageContentFromBase64("base64-encoded-data", "auto")

// 音频（URL）
ai.NewAudioContentFromURL("https://example.com/audio.mp3")

// 音频（Base64）
ai.NewAudioContentFromBase64("base64-encoded-data")

// 视频（URL）
ai.NewVideoContentFromURL("https://example.com/video.mp4")

// 视频（Base64）
ai.NewVideoContentFromBase64("base64-encoded-data")
```

## 消息角色

- `RoleSystem`: 系统消息（设置行为）
- `RoleUser`: 用户消息
- `RoleAssistant`: 助手消息（历史回复）

## 提供商类型

- `ProviderTypeOpenAI`: OpenAI
- `ProviderTypeSiliconFlow`: 硅基流动
- `ProviderTypeAnthropic`: Anthropic

## 错误处理

```go
resp, err := provider.CreateCompletion(ctx, req)
if err != nil {
    switch {
    case errors.Is(err, ai.ErrRateLimitExceeded):
        // 处理速率限制
    case errors.Is(err, ai.ErrInsufficientQuota):
        // 处理配额不足
    case errors.Is(err, ai.ErrContextLengthExceeded):
        // 处理上下文长度超出
    case errors.Is(err, ai.ErrUnsupportedFeature):
        // 处理不支持的功能
    default:
        // 其他错误
    }
}
```

## 提供商能力

```go
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
```

## 注册自定义提供商

```go
func init() {
    ai.Register(ai.ProviderType("custom"), func(config *ai.Config) (ai.Provider, error) {
        return NewCustomProvider(config)
    })
}
```

## 最佳实践

1. **使用上下文超时控制**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   ```

2. **检查提供商能力**
   ```go
   if !provider.GetCapabilities().SupportVideoInput {
       return errors.New("provider does not support video input")
   }
   ```

3. **正确处理流式输出**
   ```go
   stream, _ := provider.CreateCompletionStream(ctx, req)
   defer stream.Close()
   ```

4. **多模态内容组合**
   ```go
   // 先文本，后图像
   contents := []ai.Content{
       ai.NewTextContent("描述"),
       ai.NewImageContentFromURL("..."),
   }
   ```

5. **错误重试机制**
   ```go
   for i := 0; i < 3; i++ {
       resp, err := provider.CreateCompletion(ctx, req)
       if err == nil {
           break
       }
       if errors.Is(err, ai.ErrRateLimitExceeded) {
           time.Sleep(time.Second * time.Duration(i+1))
           continue
       }
       return err
   }
   ```
