package siliconflow

import (
	"context"
	"testing"

	"github.com/lk2023060901/go-next-erp/internal/provider/ai"
)

func TestSiliconFlowProvider(t *testing.T) {
	// 跳过测试（需要真实 API Key）
	t.Skip("Skipping SiliconFlow provider tests (requires API key)")

	config := &ai.Config{
		BaseURL: "https://api.siliconflow.cn/v1",
		APIKey:  "test-api-key",
	}

	provider, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()

	t.Run("CreateCompletion", func(t *testing.T) {
		req := &ai.CompletionRequest{
			Model: "Qwen/Qwen2.5-7B-Instruct",
			Messages: []ai.Message{
				ai.NewUserTextMessage("你好"),
			},
		}

		resp, err := provider.CreateCompletion(ctx, req)
		if err != nil {
			t.Errorf("CreateCompletion() error = %v", err)
			return
		}

		if len(resp.Choices) == 0 {
			t.Error("Expected at least one choice")
		}
	})

	t.Run("GetCapabilities", func(t *testing.T) {
		caps := provider.GetCapabilities()

		if !caps.SupportText {
			t.Error("Expected text support")
		}

		if !caps.SupportVideoInput {
			t.Error("Expected video input support")
		}

		if !caps.SupportVideoGeneration {
			t.Error("Expected video generation support")
		}
	})

	t.Run("GetProviderName", func(t *testing.T) {
		name := provider.GetProviderName()
		if name != "siliconflow" {
			t.Errorf("Expected provider name 'siliconflow', got '%s'", name)
		}
	})
}

func TestConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *ai.Config
		wantErr bool
	}{
		{
			name: "Valid config",
			config: &ai.Config{
				BaseURL: "https://api.siliconflow.cn/v1",
				APIKey:  "sk-test",
			},
			wantErr: false,
		},
		{
			name: "Missing API key",
			config: &ai.Config{
				BaseURL: "https://api.siliconflow.cn/v1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
