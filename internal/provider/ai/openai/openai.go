package openai

import (
	"bufio"
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

// Provider OpenAI 提供商实现
type Provider struct {
	config     *ai.Config
	httpClient *http.Client
	logger     *logger.Logger
}

// init 注册 OpenAI 提供商
func init() {
	ai.Register(ai.ProviderTypeOpenAI, New)
}

// New 创建 OpenAI Provider
func New(config *ai.Config) (ai.Provider, error) {
	if err := ai.ValidateConfig(config); err != nil {
		return nil, err
	}

	// 设置默认值
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
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
			zap.String("provider", "openai"),
		),
	}

	p.logger.Info("OpenAI provider initialized",
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
	// OpenAI 的语音转文本需要使用 multipart/form-data
	// 这里返回未实现错误，实际项目中需要实现 multipart 上传
	return nil, ai.ErrUnsupportedFeature
}

// AnalyzeVideo 分析视频
func (p *Provider) AnalyzeVideo(ctx context.Context, req *ai.VideoAnalysisRequest) (*ai.VideoAnalysisResponse, error) {
	// OpenAI GPT-4 Vision 目前不直接支持视频，需要先提取帧
	// 这里返回未实现错误
	return nil, ai.ErrUnsupportedFeature
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
	// OpenAI 目前不支持视频生成
	return nil, ai.ErrUnsupportedFeature
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
	return "openai"
}

// GetCapabilities 获取提供商能力
func (p *Provider) GetCapabilities() *ai.ProviderCapabilities {
	return &ai.ProviderCapabilities{
		SupportText:            true,
		SupportImageInput:      true,
		SupportAudioInput:      false,
		SupportVideoInput:      false,
		SupportImageGeneration: true,
		SupportVideoGeneration: false,
		SupportEmbedding:       true,
		SupportTranscription:   true,
		SupportStreaming:       true,
	}
}

// Close 关闭提供商
func (p *Provider) Close() error {
	p.logger.Info("OpenAI provider closed")
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

	if p.config.Organization != "" {
		req.Header.Set("OpenAI-Organization", p.config.Organization)
	}
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

// ReadStream 读取流式响应（辅助函数）
func ReadStream(stream io.ReadCloser, callback func(*ai.StreamChunk) error) error {
	defer stream.Close()

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()

		// SSE 格式: data: {...}
		if len(line) < 6 || line[:5] != "data:" {
			continue
		}

		data := line[6:]
		if data == "[DONE]" {
			break
		}

		var chunk ai.StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return fmt.Errorf("failed to parse stream chunk: %w", err)
		}

		if err := callback(&chunk); err != nil {
			return err
		}
	}

	return scanner.Err()
}
