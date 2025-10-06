package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lk2023060901/go-next-erp/internal/provider/ai"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"go.uber.org/zap"
)

// Provider Anthropic 提供商实现
type Provider struct {
	config     *ai.Config
	httpClient *http.Client
	logger     *logger.Logger
}

// init 注册 Anthropic 提供商
func init() {
	ai.Register(ai.ProviderTypeAnthropic, New)
}

// New 创建 Anthropic Provider
func New(config *ai.Config) (ai.Provider, error) {
	if err := ai.ValidateConfig(config); err != nil {
		return nil, err
	}

	// 设置默认值
	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com/v1"
	}
	if config.Timeout == 0 {
		config.Timeout = 60
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	p := &Provider{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		logger: logger.GetLogger().With(
			zap.String("module", "provider"),
			zap.String("provider", "anthropic"),
		),
	}

	p.logger.Info("Anthropic provider initialized",
		zap.String("base_url", config.BaseURL),
	)

	return p, nil
}

// CreateCompletion 创建文本补全
func (p *Provider) CreateCompletion(ctx context.Context, req *ai.CompletionRequest) (*ai.CompletionResponse, error) {
	url := fmt.Sprintf("%s/messages", p.config.BaseURL)

	// 转换为 Anthropic 格式
	anthropicReq := p.convertToAnthropicFormat(req)

	var anthropicResp anthropicResponse
	if err := p.doRequest(ctx, "POST", url, anthropicReq, &anthropicResp); err != nil {
		return nil, err
	}

	// 转换回标准格式
	resp := p.convertFromAnthropicFormat(&anthropicResp)

	p.logger.Debug("Completion created",
		zap.String("model", resp.Model),
		zap.Int("total_tokens", resp.Usage.TotalTokens),
	)

	return resp, nil
}

// CreateCompletionStream 创建流式文本补全
func (p *Provider) CreateCompletionStream(ctx context.Context, req *ai.CompletionRequest) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/messages", p.config.BaseURL)

	// 转换为 Anthropic 格式
	anthropicReq := p.convertToAnthropicFormat(req)
	anthropicReq["stream"] = true

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	p.setHeaders(httpReq)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		p.logger.Error("HTTP request failed",
			zap.String("url", url),
			zap.Error(err),
		)
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, p.handleErrorResponse(resp)
	}

	p.logger.Debug("Stream created",
		zap.String("model", req.Model),
	)

	return resp.Body, nil
}

// CreateEmbedding 创建文本嵌入
func (p *Provider) CreateEmbedding(ctx context.Context, req *ai.EmbeddingRequest) (*ai.EmbeddingResponse, error) {
	// Anthropic 不提供嵌入功能
	return nil, ai.ErrUnsupportedFeature
}

// CreateTranscription 创建语音转文本
func (p *Provider) CreateTranscription(ctx context.Context, req *ai.TranscriptionRequest) (*ai.TranscriptionResponse, error) {
	// Anthropic 不提供语音转文本功能
	return nil, ai.ErrUnsupportedFeature
}

// AnalyzeVideo 分析视频
func (p *Provider) AnalyzeVideo(ctx context.Context, req *ai.VideoAnalysisRequest) (*ai.VideoAnalysisResponse, error) {
	// Anthropic Claude 3.5 支持视频理解
	// 将视频转换为多模态消息格式
	messages := []ai.Message{
		{
			Role: ai.RoleUser,
			Content: []ai.Content{
				{
					Type: ai.ContentTypeText,
					Text: req.Prompt,
				},
			},
		},
	}

	// 添加视频内容
	if req.VideoURL != "" {
		messages[0].Content = append(messages[0].Content, ai.Content{
			Type: ai.ContentTypeVideo,
			URL:  req.VideoURL,
		})
	} else if req.VideoBase64 != "" {
		messages[0].Content = append(messages[0].Content, ai.Content{
			Type:   ai.ContentTypeVideo,
			Base64: req.VideoBase64,
		})
	}

	// 调用 messages 接口
	compReq := &ai.CompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	compResp, err := p.CreateCompletion(ctx, compReq)
	if err != nil {
		return nil, err
	}

	// 转换响应
	resp := &ai.VideoAnalysisResponse{
		ID:      compResp.ID,
		Object:  compResp.Object,
		Created: compResp.Created,
		Model:   compResp.Model,
		Usage:   compResp.Usage,
	}

	if len(compResp.Choices) > 0 {
		resp.Analysis = ai.GetTextFromResponse(compResp)
	}

	return resp, nil
}

// GenerateImage 生成图像
func (p *Provider) GenerateImage(ctx context.Context, req *ai.ImageGenerationRequest) (*ai.ImageGenerationResponse, error) {
	// Anthropic 不提供图像生成功能
	return nil, ai.ErrUnsupportedFeature
}

// GenerateVideo 生成视频
func (p *Provider) GenerateVideo(ctx context.Context, req *ai.VideoGenerationRequest) (*ai.VideoGenerationResponse, error) {
	// Anthropic 不提供视频生成功能
	return nil, ai.ErrUnsupportedFeature
}

// ListModels 列出模型
func (p *Provider) ListModels(ctx context.Context) (*ai.ListModelsResponse, error) {
	// Anthropic API 不提供模型列表接口，返回已知模型
	return &ai.ListModelsResponse{
		Object: "list",
		Data: []ai.ModelInfo{
			{
				ID:         "claude-3-5-sonnet-20241022",
				Object:     "model",
				OwnedBy:    "anthropic",
				Modalities: []string{"text", "image", "video"},
			},
			{
				ID:         "claude-3-opus-20240229",
				Object:     "model",
				OwnedBy:    "anthropic",
				Modalities: []string{"text", "image"},
			},
			{
				ID:         "claude-3-sonnet-20240229",
				Object:     "model",
				OwnedBy:    "anthropic",
				Modalities: []string{"text", "image"},
			},
			{
				ID:         "claude-3-haiku-20240307",
				Object:     "model",
				OwnedBy:    "anthropic",
				Modalities: []string{"text", "image"},
			},
		},
	}, nil
}

// GetProviderName 获取提供商名称
func (p *Provider) GetProviderName() string {
	return "anthropic"
}

// GetCapabilities 获取提供商能力
func (p *Provider) GetCapabilities() *ai.ProviderCapabilities {
	return &ai.ProviderCapabilities{
		SupportText:            true,
		SupportImageInput:      true,
		SupportAudioInput:      false,
		SupportVideoInput:      true,
		SupportImageGeneration: false,
		SupportVideoGeneration: false,
		SupportEmbedding:       false,
		SupportTranscription:   false,
		SupportStreaming:       true,
	}
}

// Close 关闭提供商
func (p *Provider) Close() error {
	p.logger.Info("Anthropic provider closed")
	return nil
}

// convertToAnthropicFormat 转换为 Anthropic 格式
func (p *Provider) convertToAnthropicFormat(req *ai.CompletionRequest) map[string]interface{} {
	anthropicReq := map[string]interface{}{
		"model":      req.Model,
		"max_tokens": req.MaxTokens,
	}

	if req.Temperature > 0 {
		anthropicReq["temperature"] = req.Temperature
	}

	if req.TopP > 0 {
		anthropicReq["top_p"] = req.TopP
	}

	// 转换消息
	var system string
	var messages []map[string]interface{}

	for _, msg := range req.Messages {
		if msg.Role == ai.RoleSystem {
			// Anthropic 的 system 是单独的字段
			for _, content := range msg.Content {
				if content.Type == ai.ContentTypeText {
					system = content.Text
					break
				}
			}
		} else {
			// 转换内容
			var contents []map[string]interface{}
			for _, content := range msg.Content {
				c := map[string]interface{}{
					"type": string(content.Type),
				}

				switch content.Type {
				case ai.ContentTypeText:
					c["text"] = content.Text
				case ai.ContentTypeImage:
					if content.URL != "" {
						c["source"] = map[string]interface{}{
							"type": "url",
							"url":  content.URL,
						}
					} else if content.Base64 != "" {
						c["source"] = map[string]interface{}{
							"type":       "base64",
							"media_type": "image/jpeg",
							"data":       content.Base64,
						}
					}
				case ai.ContentTypeVideo:
					if content.URL != "" {
						c["source"] = map[string]interface{}{
							"type": "url",
							"url":  content.URL,
						}
					} else if content.Base64 != "" {
						c["source"] = map[string]interface{}{
							"type":       "base64",
							"media_type": "video/mp4",
							"data":       content.Base64,
						}
					}
				}

				contents = append(contents, c)
			}

			messages = append(messages, map[string]interface{}{
				"role":    string(msg.Role),
				"content": contents,
			})
		}
	}

	if system != "" {
		anthropicReq["system"] = system
	}
	anthropicReq["messages"] = messages

	return anthropicReq
}

// anthropicResponse Anthropic 响应格式
type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// convertFromAnthropicFormat 从 Anthropic 格式转换
func (p *Provider) convertFromAnthropicFormat(resp *anthropicResponse) *ai.CompletionResponse {
	var contents []ai.Content
	for _, c := range resp.Content {
		contents = append(contents, ai.Content{
			Type: ai.ContentTypeText,
			Text: c.Text,
		})
	}

	return &ai.CompletionResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   resp.Model,
		Choices: []ai.Choice{
			{
				Index: 0,
				Message: ai.Message{
					Role:    ai.RoleAssistant,
					Content: contents,
				},
				FinishReason: resp.StopReason,
			},
		},
		Usage: ai.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}
}

// doRequest 执行 HTTP 请求
func (p *Provider) doRequest(ctx context.Context, method, url string, body, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	p.setHeaders(req)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.logger.Error("HTTP request failed",
			zap.String("method", method),
			zap.String("url", url),
			zap.Error(err),
		)
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return p.handleErrorResponse(resp)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// setHeaders 设置请求头
func (p *Provider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
}

// handleErrorResponse 处理错误响应
func (p *Provider) handleErrorResponse(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp struct {
		Type  string `json:"type"`
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &errResp); err == nil {
		p.logger.Error("API error",
			zap.Int("status", resp.StatusCode),
			zap.String("type", errResp.Error.Type),
			zap.String("message", errResp.Error.Message),
		)

		// 根据错误类型返回特定错误
		switch errResp.Error.Type {
		case "rate_limit_error":
			return ai.ErrRateLimitExceeded
		case "invalid_request_error":
			return ai.ErrInvalidRequest
		}

		return fmt.Errorf("%w: %s", ai.ErrAPIError, errResp.Error.Message)
	}

	return fmt.Errorf("HTTP error: %d, body: %s", resp.StatusCode, string(body))
}
