package document

import (
	"context"
	"testing"
	"time"
)

func TestDocument(t *testing.T) {
	ctx := context.Background()

	// 创建客户端
	c, err := New(
		WithBaseURL("https://mineru.net/api/v4"),
		WithAPIKey("test-api-key"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	t.Run("CreateTask", func(t *testing.T) {
		req := &CreateTaskRequest{
			URL:           "https://cdn-mineru.openxlab.org.cn/demo/example.pdf",
			IsOCR:         true,
			EnableFormula: false,
			EnableTable:   true,
		}

		// 注意：这个测试需要有效的 API Key 才能通过
		taskID, err := c.CreateTask(ctx, req)
		if err != nil {
			t.Skipf("CreateTask failed (expected without valid API key): %v", err)
		}

		if taskID == "" {
			t.Error("TaskID should not be empty")
		}
	})

	t.Run("GetTaskResult", func(t *testing.T) {
		// 使用一个虚拟的 task ID
		taskID := "test-task-id"

		_, err := c.GetTaskResult(ctx, taskID)
		if err == nil {
			t.Skip("GetTaskResult succeeded unexpectedly")
		}
		// 预期会失败，因为没有有效的 task ID
	})

	t.Run("CreateBatchURL", func(t *testing.T) {
		req := &CreateBatchURLRequest{
			EnableFormula: true,
			Language:      "ch",
			EnableTable:   true,
			Files: []BatchURLItem{
				{
					URL:    "https://cdn-mineru.openxlab.org.cn/demo/example.pdf",
					IsOCR:  true,
					DataID: "test-data-1",
				},
			},
		}

		batchID, err := c.CreateBatchURL(ctx, req)
		if err != nil {
			t.Skipf("CreateBatchURL failed (expected without valid API key): %v", err)
		}

		if batchID == "" {
			t.Error("BatchID should not be empty")
		}
	})
}

func TestConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "Valid config",
			config: &Config{
				BaseURL:       "https://mineru.net/api/v4",
				APIKey:        "test-key",
				Timeout:       30 * time.Second,
				PollInterval:  5 * time.Second,
				PollTimeout:   30 * time.Minute,
				UploadTimeout: 10 * time.Minute,
				MaxRetries:    3,
			},
			wantErr: false,
		},
		{
			name: "Missing base URL",
			config: &Config{
				APIKey:        "test-key",
				Timeout:       30 * time.Second,
				PollInterval:  5 * time.Second,
				PollTimeout:   30 * time.Minute,
				UploadTimeout: 10 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "Missing API key",
			config: &Config{
				BaseURL:       "https://mineru.net/api/v4",
				Timeout:       30 * time.Second,
				PollInterval:  5 * time.Second,
				PollTimeout:   30 * time.Minute,
				UploadTimeout: 10 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "Invalid timeout",
			config: &Config{
				BaseURL:       "https://mineru.net/api/v4",
				APIKey:        "test-key",
				Timeout:       0,
				PollInterval:  5 * time.Second,
				PollTimeout:   30 * time.Minute,
				UploadTimeout: 10 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "Invalid poll interval",
			config: &Config{
				BaseURL:       "https://mineru.net/api/v4",
				APIKey:        "test-key",
				Timeout:       30 * time.Second,
				PollInterval:  0,
				PollTimeout:   30 * time.Minute,
				UploadTimeout: 10 * time.Minute,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskState(t *testing.T) {
	states := []TaskState{
		TaskStatePending,
		TaskStateRunning,
		TaskStateDone,
		TaskStateFailed,
	}

	for _, state := range states {
		if string(state) == "" {
			t.Errorf("TaskState %v should not be empty", state)
		}
	}
}
