package logger

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 日志配置
type Config struct {
	// Level 日志级别: debug/info/warn/error/fatal
	Level string `json:"level" yaml:"level"`

	// Format 日志格式: json/console
	Format string `json:"format" yaml:"format"`

	// File 文件输出配置
	File FileConfig `json:"file" yaml:"file"`

	// Console 控制台输出配置
	Console ConsoleConfig `json:"console" yaml:"console"`

	// EnableCaller 是否启用调用者信息 (文件名:行号)
	EnableCaller bool `json:"enable_caller" yaml:"enable_caller"`

	// EnableStacktrace 是否在 Error 级别启用堆栈跟踪
	EnableStacktrace bool `json:"enable_stacktrace" yaml:"enable_stacktrace"`
}

// FileConfig 文件输出配置
type FileConfig struct {
	// Enable 是否启用文件输出
	Enable bool `json:"enable" yaml:"enable"`

	// Filename 日志文件路径
	Filename string `json:"filename" yaml:"filename"`

	// MaxSize 单个日志文件最大大小 (MB)，超过此大小会自动轮换
	MaxSize int `json:"max_size" yaml:"max_size"`

	// MaxBackups 保留的旧日志文件最大数量
	MaxBackups int `json:"max_backups" yaml:"max_backups"`

	// MaxAge 保留旧日志文件的最大天数
	MaxAge int `json:"max_age" yaml:"max_age"`

	// Compress 是否压缩旧日志文件 (gzip)
	Compress bool `json:"compress" yaml:"compress"`
}

// ConsoleConfig 控制台输出配置
type ConsoleConfig struct {
	// Enable 是否启用控制台输出
	Enable bool `json:"enable" yaml:"enable"`

	// Color 是否启用彩色输出 (仅 console 格式生效)
	Color bool `json:"color" yaml:"color"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:  "info",
		Format: "console",
		File: FileConfig{
			Enable:     false,
			Filename:   "logs/app.log",
			MaxSize:    100,  // 100MB
			MaxBackups: 5,    // 保留 5 个备份
			MaxAge:     30,   // 保留 30 天
			Compress:   false,
		},
		Console: ConsoleConfig{
			Enable: true,
			Color:  true, // 默认启用彩色输出
		},
		EnableCaller:     true,
		EnableStacktrace: true,
	}
}

// ProductionConfig 返回生产环境推荐配置
func ProductionConfig() *Config {
	return &Config{
		Level:  "info",
		Format: "json",
		File: FileConfig{
			Enable:     true,
			Filename:   "/var/log/app/server.log",
			MaxSize:    100,
			MaxBackups: 10,
			MaxAge:     30,
			Compress:   true,
		},
		Console: ConsoleConfig{
			Enable: true,  // 同时输出到控制台
			Color:  false, // 生产环境禁用彩色
		},
		EnableCaller:     true,
		EnableStacktrace: true,
	}
}

// DevelopmentConfig 返回开发环境推荐配置
func DevelopmentConfig() *Config {
	return &Config{
		Level:  "debug",
		Format: "console",
		File: FileConfig{
			Enable:     false, // 开发环境可不启用文件
			Filename:   "logs/dev.log",
			MaxSize:    50,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   false,
		},
		Console: ConsoleConfig{
			Enable: true,
			Color:  true,
		},
		EnableCaller:     true,
		EnableStacktrace: true,
	}
}

// LoadFromFile 从 YAML 文件加载配置
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证日志级别
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLevels[c.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug/info/warn/error/fatal)", c.Level)
	}

	// 验证日志格式
	validFormats := map[string]bool{
		"json":    true,
		"console": true,
	}
	if !validFormats[c.Format] {
		return fmt.Errorf("invalid log format: %s (must be json/console)", c.Format)
	}

	// 验证文件配置
	if c.File.Enable {
		if c.File.Filename == "" {
			return fmt.Errorf("file output enabled but filename is empty")
		}
		if c.File.MaxSize <= 0 {
			return fmt.Errorf("file max_size must be positive")
		}
		if c.File.MaxBackups < 0 {
			return fmt.Errorf("file max_backups must be non-negative")
		}
		if c.File.MaxAge < 0 {
			return fmt.Errorf("file max_age must be non-negative")
		}
	}

	// 至少启用一个输出
	if !c.File.Enable && !c.Console.Enable {
		return fmt.Errorf("at least one output (file or console) must be enabled")
	}

	return nil
}
