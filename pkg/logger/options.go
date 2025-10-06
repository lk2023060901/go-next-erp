package logger

// Option 配置选项函数
type Option func(*Config)

// WithLevel 设置日志级别
func WithLevel(level string) Option {
	return func(c *Config) {
		c.Level = level
	}
}

// WithFormat 设置日志格式 (json/console)
func WithFormat(format string) Option {
	return func(c *Config) {
		c.Format = format
	}
}

// WithFile 启用文件输出
func WithFile(filename string, maxSize, maxBackups, maxAge int, compress bool) Option {
	return func(c *Config) {
		c.File.Enable = true
		c.File.Filename = filename
		c.File.MaxSize = maxSize
		c.File.MaxBackups = maxBackups
		c.File.MaxAge = maxAge
		c.File.Compress = compress
	}
}

// WithConsole 启用/禁用控制台输出
func WithConsole(enable bool) Option {
	return func(c *Config) {
		c.Console.Enable = enable
	}
}

// WithConsoleColor 启用/禁用控制台彩色输出
func WithConsoleColor(enable bool) Option {
	return func(c *Config) {
		c.Console.Color = enable
	}
}

// WithCaller 启用/禁用调用者信息
func WithCaller(enable bool) Option {
	return func(c *Config) {
		c.EnableCaller = enable
	}
}

// WithStacktrace 启用/禁用堆栈跟踪
func WithStacktrace(enable bool) Option {
	return func(c *Config) {
		c.EnableStacktrace = enable
	}
}

// WithConfig 直接使用配置对象
func WithConfig(cfg *Config) Option {
	return func(c *Config) {
		*c = *cfg
	}
}

// WithProductionConfig 使用生产环境配置
func WithProductionConfig() Option {
	return func(c *Config) {
		*c = *ProductionConfig()
	}
}

// WithDevelopmentConfig 使用开发环境配置
func WithDevelopmentConfig() Option {
	return func(c *Config) {
		*c = *DevelopmentConfig()
	}
}
