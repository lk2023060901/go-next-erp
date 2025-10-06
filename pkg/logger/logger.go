package logger

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// Logger 结构化日志器（高性能封装）
type Logger struct {
	zap   *zap.Logger
	sugar *zap.SugaredLogger
}

var (
	globalLogger *Logger
	once         sync.Once
)

// New 创建 Logger 实例
func New(opts ...Option) (*Logger, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	zapLogger, err := NewZapLogger(cfg)
	if err != nil {
		return nil, err
	}

	return &Logger{
		zap:   zapLogger,
		sugar: zapLogger.Sugar(),
	}, nil
}

// ============ 结构化 API（零开销，强类型）============

// Debug 调试日志（强类型字段）
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// Info 信息日志
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Warn 警告日志
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Error 错误日志
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Fatal 致命错误（会调用 os.Exit(1)）
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// ============ 便捷 API（Sugar 风格，格式化）============

// Debugf 格式化调试
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.sugar.Debugf(template, args...)
}

// Infof 格式化信息
func (l *Logger) Infof(template string, args ...interface{}) {
	l.sugar.Infof(template, args...)
}

// Warnf 格式化警告
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.sugar.Warnf(template, args...)
}

// Errorf 格式化错误
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.sugar.Errorf(template, args...)
}

// Fatalf 格式化致命错误
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.sugar.Fatalf(template, args...)
}

// ============ 键值对 API（Sugar 风格）============

// Debugw 调试日志（键值对）
func (l *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	l.sugar.Debugw(msg, keysAndValues...)
}

// Infow 信息日志（键值对）
func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.sugar.Infow(msg, keysAndValues...)
}

// Warnw 警告日志（键值对）
func (l *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	l.sugar.Warnw(msg, keysAndValues...)
}

// Errorw 错误日志（键值对）
func (l *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	l.sugar.Errorw(msg, keysAndValues...)
}

// Fatalw 致命错误（键值对）
func (l *Logger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.sugar.Fatalw(msg, keysAndValues...)
}

// ============ 上下文相关 ============

// WithContext 从 context 中提取字段（TraceID/UserID/RequestID）
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := extractFieldsFromContext(ctx)
	if len(fields) == 0 {
		return l
	}
	return &Logger{
		zap:   l.zap.With(fields...),
		sugar: l.zap.With(fields...).Sugar(),
	}
}

// With 添加字段（返回新 Logger）
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		zap:   l.zap.With(fields...),
		sugar: l.zap.With(fields...).Sugar(),
	}
}

// WithFields 添加多个字段（便捷方法，键值对）
func (l *Logger) WithFields(keysAndValues ...interface{}) *Logger {
	sugar := l.sugar.With(keysAndValues...)
	return &Logger{
		zap:   sugar.Desugar(),
		sugar: sugar,
	}
}

// ============ 工具方法 ============

// Sync 刷新缓冲
func (l *Logger) Sync() error {
	_ = l.sugar.Sync()
	return l.zap.Sync()
}

// Clone 克隆 Logger
func (l *Logger) Clone() *Logger {
	return &Logger{
		zap:   l.zap,
		sugar: l.sugar,
	}
}

// Zap 获取底层 Zap Logger（高级用途）
func (l *Logger) Zap() *zap.Logger {
	return l.zap
}

// Sugar 获取 SugaredLogger（高级用途）
func (l *Logger) Sugar() *zap.SugaredLogger {
	return l.sugar
}

// ============ Context 字段提取 ============

func extractFieldsFromContext(ctx context.Context) []zap.Field {
	fields := make([]zap.Field, 0, 4)

	// 提取 TraceID（OpenTelemetry 风格）
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, zap.Any("trace_id", traceID))
	}

	// 提取 SpanID
	if spanID := ctx.Value("span_id"); spanID != nil {
		fields = append(fields, zap.Any("span_id", spanID))
	}

	// 提取 UserID
	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, zap.Any("user_id", userID))
	}

	// 提取 RequestID
	if reqID := ctx.Value("request_id"); reqID != nil {
		fields = append(fields, zap.Any("request_id", reqID))
	}

	return fields
}

// ============ 全局 Logger 管理 ============

// InitGlobal 初始化全局 Logger
func InitGlobal(opts ...Option) error {
	l, err := New(opts...)
	if err != nil {
		return err
	}
	globalLogger = l
	return nil
}

// GetLogger 获取全局 Logger
func GetLogger() *Logger {
	if globalLogger == nil {
		// 使用默认配置初始化
		once.Do(func() {
			_ = InitGlobal()
		})
	}
	return globalLogger
}

// SetLogger 设置全局 Logger
func SetLogger(l *Logger) {
	globalLogger = l
}

// Sync 刷新全局 Logger
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// ============ 全局便捷方法（结构化）============

// Debug 全局调试日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 全局信息日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 全局警告日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 全局错误日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 全局致命错误
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// ============ 全局便捷方法（键值对）============

// Debugw 全局调试日志（键值对）
func Debugw(msg string, keysAndValues ...interface{}) {
	GetLogger().Debugw(msg, keysAndValues...)
}

// Infow 全局信息日志（键值对）
func Infow(msg string, keysAndValues ...interface{}) {
	GetLogger().Infow(msg, keysAndValues...)
}

// Warnw 全局警告日志（键值对）
func Warnw(msg string, keysAndValues ...interface{}) {
	GetLogger().Warnw(msg, keysAndValues...)
}

// Errorw 全局错误日志（键值对）
func Errorw(msg string, keysAndValues ...interface{}) {
	GetLogger().Errorw(msg, keysAndValues...)
}

// Fatalw 全局致命错误（键值对）
func Fatalw(msg string, keysAndValues ...interface{}) {
	GetLogger().Fatalw(msg, keysAndValues...)
}

// ============ 全局便捷方法（格式化）============

// Debugf 全局格式化调试
func Debugf(template string, args ...interface{}) {
	GetLogger().Debugf(template, args...)
}

// Infof 全局格式化信息
func Infof(template string, args ...interface{}) {
	GetLogger().Infof(template, args...)
}

// Warnf 全局格式化警告
func Warnf(template string, args ...interface{}) {
	GetLogger().Warnf(template, args...)
}

// Errorf 全局格式化错误
func Errorf(template string, args ...interface{}) {
	GetLogger().Errorf(template, args...)
}

// Fatalf 全局格式化致命错误
func Fatalf(template string, args ...interface{}) {
	GetLogger().Fatalf(template, args...)
}
