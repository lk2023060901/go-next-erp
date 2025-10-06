package logger

import (
	"context"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Level != "info" {
		t.Errorf("expected level info, got %s", cfg.Level)
	}
	if cfg.Format != "console" {
		t.Errorf("expected format console, got %s", cfg.Format)
	}
	if !cfg.Console.Enable {
		t.Error("expected console enabled")
	}
}

func TestProductionConfig(t *testing.T) {
	cfg := ProductionConfig()
	if cfg.Level != "info" {
		t.Errorf("expected level info, got %s", cfg.Level)
	}
	if cfg.Format != "json" {
		t.Errorf("expected format json, got %s", cfg.Format)
	}
	if !cfg.File.Enable {
		t.Error("expected file output enabled")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			cfg:     DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid level",
			cfg: &Config{
				Level:  "invalid",
				Format: "json",
				Console: ConsoleConfig{
					Enable: true,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			cfg: &Config{
				Level:  "info",
				Format: "invalid",
				Console: ConsoleConfig{
					Enable: true,
				},
			},
			wantErr: true,
		},
		{
			name: "no output enabled",
			cfg: &Config{
				Level:  "info",
				Format: "json",
				File: FileConfig{
					Enable: false,
				},
				Console: ConsoleConfig{
					Enable: false,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	l, err := New(
		WithLevel("debug"),
		WithFormat("console"),
		WithConsole(true),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if l == nil {
		t.Fatal("expected logger, got nil")
	}
}

func TestLoggerStructured(t *testing.T) {
	l, err := New(
		WithLevel("debug"),
		WithConsole(true),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// 测试结构化日志
	l.Debug("debug message", zap.String("key", "value"))
	l.Info("info message", zap.Int("count", 42))
	l.Warn("warn message", zap.Duration("elapsed", 100*time.Millisecond))
	l.Error("error message", zap.Bool("success", false))

	// Ignore "sync /dev/stdout: bad file descriptor" error in tests
	_ = l.Sync()
}

func TestLoggerSugar(t *testing.T) {
	l, err := New(
		WithLevel("debug"),
		WithConsole(true),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// 测试键值对风格
	l.Debugw("debug message", "key", "value")
	l.Infow("info message", "count", 42)
	l.Warnw("warn message", "elapsed", 100*time.Millisecond)
	l.Errorw("error message", "success", false)

	// Ignore "sync /dev/stdout: bad file descriptor" error in tests
	_ = l.Sync()
}

func TestLoggerFormatted(t *testing.T) {
	l, err := New(
		WithLevel("debug"),
		WithConsole(true),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// 测试格式化日志
	l.Debugf("debug: %s", "message")
	l.Infof("info: %d", 42)
	l.Warnf("warn: %v", time.Now())
	l.Errorf("error: %t", false)

	// Ignore "sync /dev/stdout: bad file descriptor" error in tests
	_ = l.Sync()
}

func TestLoggerWith(t *testing.T) {
	l, err := New(
		WithLevel("info"),
		WithConsole(true),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// 添加字段
	l2 := l.With(
		zap.String("module", "test"),
		zap.String("version", "v1.0.0"),
	)

	l2.Info("message with fields")

	// Ignore sync errors in tests
	_ = l2.Sync()
}

func TestLoggerWithContext(t *testing.T) {
	l, err := New(
		WithLevel("info"),
		WithConsole(true),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// 创建带字段的 context
	ctx := context.WithValue(context.Background(), "trace_id", "abc123")
	ctx = context.WithValue(ctx, "user_id", int64(1001))

	l2 := l.WithContext(ctx)
	l2.Info("message with context fields")

	// Ignore sync errors in tests
	_ = l2.Sync()
}

func TestGlobalLogger(t *testing.T) {
	// 初始化全局 logger
	err := InitGlobal(
		WithLevel("debug"),
		WithConsole(true),
	)
	if err != nil {
		t.Fatalf("InitGlobal() error = %v", err)
	}

	// 测试全局方法
	Debug("debug", zap.String("key", "value"))
	Info("info", zap.Int("count", 42))
	Warn("warn")
	Error("error")

	Debugw("debugw", "key", "value")
	Infow("infow", "count", 42)
	Warnw("warnw")
	Errorw("errorw")

	Debugf("debugf: %s", "msg")
	Infof("infof: %d", 42)
	Warnf("warnf")
	Errorf("errorf")

	// Ignore sync errors in tests
	_ = Sync()
}

func TestFileOutput(t *testing.T) {
	tmpFile := "/tmp/test_logger.log"
	defer os.Remove(tmpFile)

	l, err := New(
		WithLevel("info"),
		WithFormat("json"),
		WithFile(tmpFile, 10, 3, 7, false),
		WithConsole(false),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	l.Info("test file output", zap.String("key", "value"))

	// Ignore "sync /dev/stdout: bad file descriptor" error in tests
	_ = l.Sync()

	// 检查文件是否创建
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("log file was not created")
	}
}
