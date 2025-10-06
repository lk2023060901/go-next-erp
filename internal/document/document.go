package document

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"go.uber.org/zap"
)

// Client MinerU 文档解析客户端接口
type Client interface {
	// 单个文件解析
	CreateTask(ctx context.Context, req *CreateTaskRequest) (string, error)
	GetTaskResult(ctx context.Context, taskID string) (*TaskResult, error)
	WaitForTask(ctx context.Context, taskID string) (*TaskResult, error)

	// 批量文件上传解析
	CreateBatchUpload(ctx context.Context, req *CreateBatchUploadRequest) (batchID string, uploadURLs []string, err error)
	UploadFile(ctx context.Context, uploadURL string, filePath string) error

	// 批量 URL 解析
	CreateBatchURL(ctx context.Context, req *CreateBatchURLRequest) (string, error)

	// 批量获取结果
	GetBatchResult(ctx context.Context, batchID string) ([]BatchExtractResult, error)
	WaitForBatch(ctx context.Context, batchID string) ([]BatchExtractResult, error)

	// 关闭客户端
	Close() error
}

// client MinerU 客户端实现
type client struct {
	config     *Config
	httpClient *http.Client
	logger     *logger.Logger
}

// New 创建 MinerU 客户端
func New(opts ...Option) (Client, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	c := &client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger.GetLogger().With(zap.String("module", "document")),
	}

	c.logger.Info("MinerU client initialized",
		zap.String("base_url", cfg.BaseURL),
	)

	return c, nil
}

// CreateTask 创建单个解析任务
func (c *client) CreateTask(ctx context.Context, req *CreateTaskRequest) (string, error) {
	url := fmt.Sprintf("%s/extract/task", c.config.BaseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	var resp CreateTaskResponse
	if err := c.doRequest(ctx, "POST", url, body, &resp); err != nil {
		return "", err
	}

	if resp.Code != 0 {
		c.logger.Error("Failed to create task",
			zap.Int("code", resp.Code),
			zap.String("message", resp.Message),
		)
		return "", fmt.Errorf("API error (code: %d): %s", resp.Code, resp.Message)
	}

	c.logger.Info("Task created successfully",
		zap.String("task_id", resp.Data.TaskID),
		zap.String("url", req.URL),
	)

	return resp.Data.TaskID, nil
}

// GetTaskResult 获取任务结果
func (c *client) GetTaskResult(ctx context.Context, taskID string) (*TaskResult, error) {
	url := fmt.Sprintf("%s/extract/task/%s", c.config.BaseURL, taskID)

	var resp GetTaskResultResponse
	if err := c.doRequest(ctx, "GET", url, nil, &resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error (code: %d): %s", resp.Code, resp.Message)
	}

	return &resp.Data, nil
}

// WaitForTask 等待任务完成
func (c *client) WaitForTask(ctx context.Context, taskID string) (*TaskResult, error) {
	ticker := time.NewTicker(c.config.PollInterval)
	defer ticker.Stop()

	timeout := time.After(c.config.PollTimeout)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("task timeout after %v", c.config.PollTimeout)
		case <-ticker.C:
			result, err := c.GetTaskResult(ctx, taskID)
			if err != nil {
				return nil, err
			}

			switch result.State {
			case TaskStateDone:
				c.logger.Info("Task completed",
					zap.String("task_id", taskID),
					zap.String("zip_url", result.FullZipURL),
				)
				return result, nil
			case TaskStateFailed:
				c.logger.Error("Task failed",
					zap.String("task_id", taskID),
					zap.String("error", result.ErrorMsg),
				)
				return nil, fmt.Errorf("task failed: %s", result.ErrorMsg)
			case TaskStatePending, TaskStateRunning:
				c.logger.Debug("Task in progress",
					zap.String("task_id", taskID),
					zap.String("state", string(result.State)),
				)
				// 继续轮询
			default:
				return nil, fmt.Errorf("unknown task state: %s", result.State)
			}
		}
	}
}

// CreateBatchUpload 批量上传解析
func (c *client) CreateBatchUpload(ctx context.Context, req *CreateBatchUploadRequest) (string, []string, error) {
	url := fmt.Sprintf("%s/file-urls/batch", c.config.BaseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var resp CreateBatchUploadResponse
	if err := c.doRequest(ctx, "POST", url, body, &resp); err != nil {
		return "", nil, err
	}

	if resp.Code != 0 {
		c.logger.Error("Failed to create batch upload",
			zap.Int("code", resp.Code),
			zap.String("message", resp.Message),
		)
		return "", nil, fmt.Errorf("API error (code: %d): %s", resp.Code, resp.Message)
	}

	c.logger.Info("Batch upload created successfully",
		zap.String("batch_id", resp.Data.BatchID),
		zap.Int("file_count", len(resp.Data.FileURLs)),
	)

	return resp.Data.BatchID, resp.Data.FileURLs, nil
}

// UploadFile 上传文件到预签名 URL
func (c *client) UploadFile(ctx context.Context, uploadURL string, filePath string) error {
	file, err := http.DefaultClient.Get(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	defer file.Body.Close()

	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, file.Body)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	uploadClient := &http.Client{
		Timeout: c.config.UploadTimeout,
	}

	resp, err := uploadClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed with status: %d", resp.StatusCode)
	}

	c.logger.Debug("File uploaded successfully",
		zap.String("file_path", filePath),
	)

	return nil
}

// CreateBatchURL 批量 URL 解析
func (c *client) CreateBatchURL(ctx context.Context, req *CreateBatchURLRequest) (string, error) {
	url := fmt.Sprintf("%s/extract/task/batch", c.config.BaseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	var resp CreateBatchURLResponse
	if err := c.doRequest(ctx, "POST", url, body, &resp); err != nil {
		return "", err
	}

	if resp.Code != 0 {
		c.logger.Error("Failed to create batch URL task",
			zap.Int("code", resp.Code),
			zap.String("message", resp.Message),
		)
		return "", fmt.Errorf("API error (code: %d): %s", resp.Code, resp.Message)
	}

	c.logger.Info("Batch URL task created successfully",
		zap.String("batch_id", resp.Data.BatchID),
	)

	return resp.Data.BatchID, nil
}

// GetBatchResult 批量获取结果
func (c *client) GetBatchResult(ctx context.Context, batchID string) ([]BatchExtractResult, error) {
	url := fmt.Sprintf("%s/extract-results/batch/%s", c.config.BaseURL, batchID)

	var resp GetBatchResultResponse
	if err := c.doRequest(ctx, "GET", url, nil, &resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error (code: %d): %s", resp.Code, resp.Message)
	}

	return resp.Data.ExtractResult, nil
}

// WaitForBatch 等待批量任务完成
func (c *client) WaitForBatch(ctx context.Context, batchID string) ([]BatchExtractResult, error) {
	ticker := time.NewTicker(c.config.PollInterval)
	defer ticker.Stop()

	timeout := time.After(c.config.PollTimeout)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("batch timeout after %v", c.config.PollTimeout)
		case <-ticker.C:
			results, err := c.GetBatchResult(ctx, batchID)
			if err != nil {
				return nil, err
			}

			allDone := true
			hasFailure := false
			for _, result := range results {
				if result.State == TaskStateFailed {
					hasFailure = true
				}
				if result.State == TaskStatePending || result.State == TaskStateRunning {
					allDone = false
				}
			}

			if allDone {
				if hasFailure {
					c.logger.Warn("Batch completed with failures",
						zap.String("batch_id", batchID),
					)
				} else {
					c.logger.Info("Batch completed successfully",
						zap.String("batch_id", batchID),
					)
				}
				return results, nil
			}

			c.logger.Debug("Batch in progress",
				zap.String("batch_id", batchID),
			)
		}
	}
}

// doRequest 执行 HTTP 请求
func (c *client) doRequest(ctx context.Context, method, url string, body []byte, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("HTTP request failed",
			zap.String("method", method),
			zap.String("url", url),
			zap.Error(err),
		)
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			c.logger.Error("API error",
				zap.Int("status", resp.StatusCode),
				zap.Int("code", errResp.Code),
				zap.String("message", errResp.Message),
			)
			return fmt.Errorf("API error (status: %d, code: %d): %s", resp.StatusCode, errResp.Code, errResp.Message)
		}
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// Close 关闭客户端
func (c *client) Close() error {
	c.logger.Info("MinerU client closed")
	return nil
}
