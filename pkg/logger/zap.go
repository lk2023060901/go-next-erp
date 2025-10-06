package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewZapLogger 创建 Zap Logger
func NewZapLogger(cfg *Config) (*zap.Logger, error) {
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建核心
	core := newZapCore(cfg)

	// 创建 Logger 选项
	opts := []zap.Option{
		zap.AddCallerSkip(1), // 跳过一层封装
	}

	// 添加调用者信息
	if cfg.EnableCaller {
		opts = append(opts, zap.AddCaller())
	}

	// 添加堆栈跟踪
	if cfg.EnableStacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	return zap.New(core, opts...), nil
}

// newZapCore 创建 zapcore.Core（支持多输出）
func newZapCore(cfg *Config) zapcore.Core {
	// 获取日志级别
	level := getLogLevel(cfg.Level)

	// 获取编码器
	encoder := getEncoder(cfg.Format, cfg.EnableCaller, cfg.Console.Color)

	// 收集所有核心
	var cores []zapcore.Core

	// 文件输出核心
	if cfg.File.Enable {
		fileWriter := getFileWriter(&cfg.File)
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.AddSync(fileWriter),
			level,
		))
	}

	// 控制台输出核心
	if cfg.Console.Enable {
		consoleWriter := zapcore.Lock(os.Stdout)
		cores = append(cores, zapcore.NewCore(
			encoder,
			consoleWriter,
			level,
		))
	}

	// 合并多个核心
	return zapcore.NewTee(cores...)
}

// getLogLevel 解析日志级别
func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// getEncoder 获取编码器
func getEncoder(format string, enableCaller bool, enableColor bool) zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Console 格式支持彩色输出
	if format == "console" && enableColor {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 不启用 Caller 时移除 CallerKey
	if !enableCaller {
		encoderConfig.CallerKey = zapcore.OmitKey
	}

	if format == "json" {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// getFileWriter 获取文件写入器（使用 Lumberjack 轮换）
func getFileWriter(cfg *FileConfig) zapcore.WriteSyncer {
	// 确保日志目录存在
	logDir := filepath.Dir(cfg.Filename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// 如果创建目录失败，回退到当前目录
		cfg.Filename = filepath.Base(cfg.Filename)
	}

	// Lumberjack 配置
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,    // MB
		MaxBackups: cfg.MaxBackups, // 保留数量
		MaxAge:     cfg.MaxAge,     // 天数
		Compress:   cfg.Compress,   // 压缩
		LocalTime:  true,           // 使用本地时间
	})
}
