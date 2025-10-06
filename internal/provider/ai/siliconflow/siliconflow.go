package siliconflow

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

// Provider 硅基流动提供商实现
type Provider struct {
	config     *ai.Config
	httpClient *http.Client
	logger     *logger.Logger
}

// init 注册硅基流动提供商
func init() {
	ai.Register(ai.ProviderTypeSiliconFlow, New)
}

// New 创建硅基流动 Provider
func New(config *ai.Config) (ai.Provider, error) {
	if err := ai.ValidateConfig(config); err != nil {
		return nil, err
	}

	// 设置默认值
	if config.BaseURL == "" {
		config.BaseURL = "https://api.siliconflow.cn/v1"
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
			zap.String("provider", "siliconflow"),
		),
	}

	p.logger.Info("SiliconFlow provider initialized",
		zap.String("base_url", config.BaseURL),
	)

	return p, nil
}

// CreateCompletion 创建文本补全
func (p *Provider) CreateCompletion(ctx context.Context, req *ai.CompletionRequest) (*ai.CompletionResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", p.config.BaseURL)

	var resp ai.CompletionResponse
	if err := p.doRequest(ctx, "POST", url, req, &resp); err != nil {
		return nil, err
	}

	p.logger.Debug("Completion created",
		zap.String("model", resp.Model),
		zap.Int("total_tokens", resp.Usage.TotalTokens),
	)

	return &resp, nil
}

// CreateCompletionStream 创建流式文本补全
func (p *Provider) CreateCompletionStream(ctx context.Context, req *ai.CompletionRequest) (io.ReadCloser, error) {
	req.Stream = true
	url := fmt.Sprintf("%s/chat/completions", p.config.BaseURL)

	body, err := json.Marshal(req)
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
	url := fmt.Sprintf("%s/embeddings", p.config.BaseURL)

	var resp ai.EmbeddingResponse
	if err := p.doRequest(ctx, "POST", url, req, &resp); err != nil {
		return nil, err
	}

	p.logger.Debug("Embedding created",
		zap.String("model", resp.Model),
		zap.Int("input_count", len(req.Input)),
	)

	return &resp, nil
}

// CreateTranscription 创建语音转文本
func (p *Provider) CreateTranscription(ctx context.Context, req *ai.TranscriptionRequest) (*ai.TranscriptionResponse, error) {
	// 硅基流动支持语音转文本，使用 multipart/form-data
	// 实际项目中需要实现 multipart 上传
	return nil, ai.ErrUnsupportedFeature
}

// AnalyzeVideo 分析视频
func (p *Provider) AnalyzeVideo(ctx context.Context, req *ai.VideoAnalysisRequest) (*ai.VideoAnalysisResponse, error) {
	// 硅基流动支持视频理解
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

	// 调用 chat/completions 接口
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
	url := fmt.Sprintf("%s/images/generations", p.config.BaseURL)

	var resp ai.ImageGenerationResponse
	if err := p.doRequest(ctx, "POST", url, req, &resp); err != nil {
		return nil, err
	}

	p.logger.Debug("Image generated",
		zap.String("model", req.Model),
		zap.Int("count", len(resp.Data)),
	)

	return &resp, nil
}

// GenerateVideo 生成视频
func (p *Provider) GenerateVideo(ctx context.Context, req *ai.VideoGenerationRequest) (*ai.VideoGenerationResponse, error) {
	// 硅基流动支持视频生成
	url := fmt.Sprintf("%s/videos/generations", p.config.BaseURL)

	var resp ai.VideoGenerationResponse
	if err := p.doRequest(ctx, "POST", url, req, &resp); err != nil {
		return nil, err
	}

	p.logger.Debug("Video generated",
		zap.String("model", req.Model),
		zap.Int("count", len(resp.Data)),
	)

	return &resp, nil
}

// ListModels 列出模型
func (p *Provider) ListModels(ctx context.Context) (*ai.ListModelsResponse, error) {
	url := fmt.Sprintf("%s/models", p.config.BaseURL)

	var resp ai.ListModelsResponse
	if err := p.doRequest(ctx, "GET", url, nil, &resp); err != nil {
		return nil, err
	}

	p.logger.Debug("Models listed",
		zap.Int("count", len(resp.Data)),
	)

	return &resp, nil
}

// GetProviderName 获取提供商名称
func (p *Provider) GetProviderName() string {
	return "siliconflow"
}

// GetCapabilities 获取提供商能力
func (p *Provider) GetCapabilities() *ai.ProviderCapabilities {
	return &ai.ProviderCapabilities{
		SupportText:            true,
		SupportImageInput:      true,
		SupportAudioInput:      true,
		SupportVideoInput:      true,
		SupportImageGeneration: true,
		SupportVideoGeneration: true,
		SupportEmbedding:       true,
		SupportTranscription:   true,
		SupportStreaming:       true,
	}
}

// Close 关闭提供商
func (p *Provider) Close() error {
	p.logger.Info("SiliconFlow provider closed")
	return nil
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.config.APIKey))
}

// handleErrorResponse 处理错误响应
func (p *Provider) handleErrorResponse(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp ai.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil {
		p.logger.Error("API error",
			zap.Int("status", resp.StatusCode),
			zap.String("type", errResp.Error.Type),
			zap.String("message", errResp.Error.Message),
		)

		// 根据错误类型返回特定错误
		switch errResp.Error.Type {
		case "insufficient_quota":
			return ai.ErrInsufficientQuota
		case "invalid_request_error":
			return ai.ErrInvalidRequest
		case "rate_limit_exceeded":
			return ai.ErrRateLimitExceeded
		}

		return fmt.Errorf("%w: %s", ai.ErrAPIError, errResp.Error.Message)
	}

	return fmt.Errorf("HTTP error: %d, body: %s", resp.StatusCode, string(body))
}
