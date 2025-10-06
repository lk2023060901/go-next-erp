package openai

import (
	"context"
	"testing"

	"github.com/lk2023060901/go-next-erp/internal/provider/ai"
)

func TestOpenAIProvider(t *testing.T) {
	// 跳过测试（需要真实 API Key）
	t.Skip("Skipping OpenAI provider tests (requires API key)")

	config := &ai.Config{
		BaseURL: "https://api.openai.com/v1",
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
			Model: "gpt-3.5-turbo",
			Messages: []ai.Message{
				ai.NewUserTextMessage("Hello, how are you?"),
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

		if !caps.SupportImageInput {
			t.Error("Expected image input support")
		}

		if !caps.SupportStreaming {
			t.Error("Expected streaming support")
		}
	})

	t.Run("GetProviderName", func(t *testing.T) {
		name := provider.GetProviderName()
		if name != "openai" {
			t.Errorf("Expected provider name 'openai', got '%s'", name)
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
				BaseURL: "https://api.openai.com/v1",
				APIKey:  "sk-test",
			},
			wantErr: false,
		},
		{
			name: "Missing API key",
			config: &ai.Config{
				BaseURL: "https://api.openai.com/v1",
			},
			wantErr: true,
		},
		{
			name: "Missing base URL",
			config: &ai.Config{
				APIKey: "sk-test",
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
