package ai

// ContentType 内容类型
type ContentType string

const (
	ContentTypeText  ContentType = "text"  // 文本
	ContentTypeImage ContentType = "image" // 图像
	ContentTypeAudio ContentType = "audio" // 音频
	ContentTypeVideo ContentType = "video" // 视频
)

// Role 角色类型
type Role string

const (
	RoleSystem    Role = "system"    // 系统消息
	RoleUser      Role = "user"      // 用户消息
	RoleAssistant Role = "assistant" // 助手消息
)

// Content 多模态内容
type Content struct {
	Type ContentType `json:"type"`           // 内容类型
	Text string      `json:"text,omitempty"` // 文本内容

	// 媒体内容（图像/音频/视频）
	URL    string `json:"url,omitempty"`     // 媒体 URL
	Base64 string `json:"base64,omitempty"`  // Base64 编码的媒体数据
	Detail string `json:"detail,omitempty"`  // 图像细节级别: low, high, auto
}

// Message 消息
type Message struct {
	Role    Role      `json:"role"`    // 角色
	Content []Content `json:"content"` // 内容列表（支持多模态）
}

// CompletionRequest 完成请求
type CompletionRequest struct {
	Model       string    `json:"model"`                  // 模型名称
	Messages    []Message `json:"messages"`               // 消息列表
	MaxTokens   int       `json:"max_tokens,omitempty"`   // 最大生成 token 数
	Temperature float64   `json:"temperature,omitempty"`  // 温度参数 (0-2)
	TopP        float64   `json:"top_p,omitempty"`        // Top-p 采样
	Stream      bool      `json:"stream,omitempty"`       // 是否流式输出
	Stop        []string  `json:"stop,omitempty"`         // 停止词
	N           int       `json:"n,omitempty"`            // 生成数量
	User        string    `json:"user,omitempty"`         // 用户标识
	Metadata    map[string]interface{} `json:"metadata,omitempty"` // 额外元数据
}

// Choice 选择
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"` // stop, length, content_filter
}

// Usage 使用统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CompletionResponse 完成响应
type CompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// StreamChunk 流式响应块
type StreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    Role      `json:"role,omitempty"`
			Content []Content `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

// EmbeddingRequest 嵌入请求
type EmbeddingRequest struct {
	Model      string   `json:"model"`                 // 模型名称
	Input      []string `json:"input"`                 // 输入文本列表
	User       string   `json:"user,omitempty"`        // 用户标识
	Dimensions int      `json:"dimensions,omitempty"`  // 向量维度
}

// Embedding 嵌入结果
type Embedding struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

// EmbeddingResponse 嵌入响应
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
}

// TranscriptionRequest 语音转文本请求
type TranscriptionRequest struct {
	Model    string `json:"model"`              // 模型名称
	File     []byte `json:"file"`               // 音频文件数据
	Language string `json:"language,omitempty"` // 语言代码
	Prompt   string `json:"prompt,omitempty"`   // 提示文本
}

// TranscriptionResponse 语音转文本响应
type TranscriptionResponse struct {
	Text string `json:"text"`
}

// ImageGenerationRequest 图像生成请求
type ImageGenerationRequest struct {
	Prompt         string `json:"prompt"`                    // 提示词
	Model          string `json:"model,omitempty"`           // 模型名称
	N              int    `json:"n,omitempty"`               // 生成数量
	Size           string `json:"size,omitempty"`            // 尺寸: 256x256, 512x512, 1024x1024
	Quality        string `json:"quality,omitempty"`         // 质量: standard, hd
	Style          string `json:"style,omitempty"`           // 风格: vivid, natural
	ResponseFormat string `json:"response_format,omitempty"` // 响应格式: url, b64_json
	User           string `json:"user,omitempty"`            // 用户标识
}

// ImageData 图像数据
type ImageData struct {
	URL           string `json:"url,omitempty"`
	Base64        string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// ImageGenerationResponse 图像生成响应
type ImageGenerationResponse struct {
	Created int64       `json:"created"`
	Data    []ImageData `json:"data"`
}

// VideoAnalysisRequest 视频分析请求
type VideoAnalysisRequest struct {
	Model       string   `json:"model"`                 // 模型名称
	VideoURL    string   `json:"video_url,omitempty"`   // 视频 URL
	VideoBase64 string   `json:"video_base64,omitempty"` // Base64 编码的视频数据
	Prompt      string   `json:"prompt"`                // 分析提示词
	MaxTokens   int      `json:"max_tokens,omitempty"`  // 最大生成 token 数
	Temperature float64  `json:"temperature,omitempty"` // 温度参数
	FrameRate   int      `json:"frame_rate,omitempty"`  // 采样帧率（每秒帧数）
	MaxFrames   int      `json:"max_frames,omitempty"`  // 最大帧数
}

// VideoAnalysisResponse 视频分析响应
type VideoAnalysisResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`

	// 分析结果
	Analysis string `json:"analysis"` // 文本分析结果

	// 帧级别信息（可选）
	Frames []VideoFrame `json:"frames,omitempty"`

	Usage Usage `json:"usage"`
}

// VideoFrame 视频帧信息
type VideoFrame struct {
	Timestamp   float64 `json:"timestamp"`   // 时间戳（秒）
	FrameIndex  int     `json:"frame_index"` // 帧索引
	Description string  `json:"description"` // 帧描述
	Objects     []DetectedObject `json:"objects,omitempty"` // 检测到的对象
}

// DetectedObject 检测到的对象
type DetectedObject struct {
	Label      string  `json:"label"`      // 对象标签
	Confidence float64 `json:"confidence"` // 置信度
	BoundingBox *BoundingBox `json:"bounding_box,omitempty"` // 边界框
}

// BoundingBox 边界框
type BoundingBox struct {
	X      int `json:"x"`      // X 坐标
	Y      int `json:"y"`      // Y 坐标
	Width  int `json:"width"`  // 宽度
	Height int `json:"height"` // 高度
}

// VideoGenerationRequest 视频生成请求
type VideoGenerationRequest struct {
	Prompt         string  `json:"prompt"`                    // 提示词
	Model          string  `json:"model,omitempty"`           // 模型名称
	Duration       int     `json:"duration,omitempty"`        // 视频时长（秒）
	FPS            int     `json:"fps,omitempty"`             // 帧率
	Resolution     string  `json:"resolution,omitempty"`      // 分辨率: 720p, 1080p
	AspectRatio    string  `json:"aspect_ratio,omitempty"`    // 宽高比: 16:9, 9:16, 1:1
	N              int     `json:"n,omitempty"`               // 生成数量
	Seed           int64   `json:"seed,omitempty"`            // 随机种子
	ResponseFormat string  `json:"response_format,omitempty"` // 响应格式: url, base64
	User           string  `json:"user,omitempty"`            // 用户标识
}

// VideoData 视频数据
type VideoData struct {
	URL           string `json:"url,omitempty"`
	Base64        string `json:"base64,omitempty"`
	Duration      int    `json:"duration,omitempty"`       // 视频时长（秒）
	FPS           int    `json:"fps,omitempty"`            // 帧率
	Resolution    string `json:"resolution,omitempty"`     // 分辨率
	RevisedPrompt string `json:"revised_prompt,omitempty"` // 修订后的提示词
}

// VideoGenerationResponse 视频生成响应
type VideoGenerationResponse struct {
	Created int64       `json:"created"`
	Data    []VideoData `json:"data"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	ID         string   `json:"id"`
	Object     string   `json:"object"`
	Created    int64    `json:"created"`
	OwnedBy    string   `json:"owned_by"`
	Modalities []string `json:"modalities,omitempty"` // 支持的模态: text, image, audio, video
}

// ListModelsResponse 模型列表响应
type ListModelsResponse struct {
	Object string      `json:"object"`
	Data   []ModelInfo `json:"data"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param,omitempty"`
		Code    string `json:"code,omitempty"`
	} `json:"error"`
}
