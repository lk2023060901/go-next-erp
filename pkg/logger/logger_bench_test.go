package logger

import (
	"context"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

// BenchmarkStructuredLog 基准测试：结构化日志（零开销）
func BenchmarkStructuredLog(b *testing.B) {
	l, _ := New(
		WithLevel("info"),
		WithConsole(false),
		WithFile("/tmp/bench_structured.log", 100, 5, 30, false),
	)
	defer l.Sync()
	defer os.Remove("/tmp/bench_structured.log")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("benchmark message",
				zap.String("key", "value"),
				zap.Int("count", 42),
				zap.Duration("elapsed", 123*time.Millisecond),
			)
		}
	})
}

// BenchmarkSugarLog 基准测试：Sugar 键值对风格
func BenchmarkSugarLog(b *testing.B) {
	l, _ := New(
		WithLevel("info"),
		WithConsole(false),
		WithFile("/tmp/bench_sugar.log", 100, 5, 30, false),
	)
	defer l.Sync()
	defer os.Remove("/tmp/bench_sugar.log")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Infow("benchmark message",
				"key", "value",
				"count", 42,
				"elapsed", 123*time.Millisecond,
			)
		}
	})
}

// BenchmarkFormattedLog 基准测试：格式化日志
func BenchmarkFormattedLog(b *testing.B) {
	l, _ := New(
		WithLevel("info"),
		WithConsole(false),
		WithFile("/tmp/bench_formatted.log", 100, 5, 30, false),
	)
	defer l.Sync()
	defer os.Remove("/tmp/bench_formatted.log")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Infof("benchmark message: key=%s count=%d elapsed=%v",
				"value", 42, 123*time.Millisecond,
			)
		}
	})
}

// BenchmarkWithFields 基准测试：带预设字段
func BenchmarkWithFields(b *testing.B) {
	l, _ := New(
		WithLevel("info"),
		WithConsole(false),
		WithFile("/tmp/bench_with_fields.log", 100, 5, 30, false),
	)
	defer l.Sync()
	defer os.Remove("/tmp/bench_with_fields.log")

	// 预设字段
	logger := l.With(
		zap.String("module", "benchmark"),
		zap.String("version", "v1.0.0"),
		zap.String("env", "test"),
	)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("benchmark message",
				zap.String("key", "value"),
				zap.Int("count", 42),
			)
		}
	})
}

// BenchmarkWithContext 基准测试：Context 字段提取
func BenchmarkWithContext(b *testing.B) {
	l, _ := New(
		WithLevel("info"),
		WithConsole(false),
		WithFile("/tmp/bench_context.log", 100, 5, 30, false),
	)
	defer l.Sync()
	defer os.Remove("/tmp/bench_context.log")

	ctx := context.WithValue(context.Background(), "trace_id", "abc123")
	ctx = context.WithValue(ctx, "user_id", int64(1001))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger := l.WithContext(ctx)
			logger.Info("benchmark message",
				zap.String("key", "value"),
			)
		}
	})
}

// BenchmarkJSONEncoding 基准测试：JSON 编码
func BenchmarkJSONEncoding(b *testing.B) {
	l, _ := New(
		WithLevel("info"),
		WithFormat("json"),
		WithConsole(false),
		WithFile("/tmp/bench_json.log", 100, 5, 30, false),
	)
	defer l.Sync()
	defer os.Remove("/tmp/bench_json.log")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("benchmark message",
				zap.String("key", "value"),
				zap.Int("count", 42),
				zap.Duration("elapsed", 123*time.Millisecond),
			)
		}
	})
}

// BenchmarkConsoleEncoding 基准测试：Console 编码
func BenchmarkConsoleEncoding(b *testing.B) {
	l, _ := New(
		WithLevel("info"),
		WithFormat("console"),
		WithConsole(false),
		WithFile("/tmp/bench_console.log", 100, 5, 30, false),
	)
	defer l.Sync()
	defer os.Remove("/tmp/bench_console.log")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("benchmark message",
				zap.String("key", "value"),
				zap.Int("count", 42),
				zap.Duration("elapsed", 123*time.Millisecond),
			)
		}
	})
}

// BenchmarkComplexFields 基准测试：复杂字段类型
func BenchmarkComplexFields(b *testing.B) {
	l, _ := New(
		WithLevel("info"),
		WithConsole(false),
		WithFile("/tmp/bench_complex.log", 100, 5, 30, false),
	)
	defer l.Sync()
	defer os.Remove("/tmp/bench_complex.log")

	type User struct {
		ID    int64
		Email string
		Role  string
	}

	user := User{
		ID:    12345,
		Email: "user@example.com",
		Role:  "admin",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("benchmark message",
				zap.String("action", "create_user"),
				zap.Any("user", user),
				zap.Time("timestamp", time.Now()),
				zap.Bool("success", true),
				zap.Float64("score", 98.5),
			)
		}
	})
}

// BenchmarkDifferentLevels 基准测试：不同日志级别
func BenchmarkDifferentLevels(b *testing.B) {
	l, _ := New(
		WithLevel("info"),
		WithConsole(false),
		WithFile("/tmp/bench_levels.log", 100, 5, 30, false),
	)
	defer l.Sync()
	defer os.Remove("/tmp/bench_levels.log")

	b.Run("Debug-Disabled", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Debug("debug message", zap.String("key", "value"))
			}
		})
	})

	b.Run("Info-Enabled", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("info message", zap.String("key", "value"))
			}
		})
	})

	b.Run("Warn-Enabled", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Warn("warn message", zap.String("key", "value"))
			}
		})
	})

	b.Run("Error-Enabled", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Error("error message", zap.String("key", "value"))
			}
		})
	})
}

// BenchmarkCaller 基准测试：调用者信息开销
func BenchmarkCaller(b *testing.B) {
	b.Run("WithCaller", func(b *testing.B) {
		l, _ := New(
			WithLevel("info"),
			WithConsole(false),
			WithFile("/tmp/bench_caller_on.log", 100, 5, 30, false),
			WithCaller(true),
		)
		defer l.Sync()
		defer os.Remove("/tmp/bench_caller_on.log")

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("benchmark message", zap.String("key", "value"))
			}
		})
	})

	b.Run("WithoutCaller", func(b *testing.B) {
		l, _ := New(
			WithLevel("info"),
			WithConsole(false),
			WithFile("/tmp/bench_caller_off.log", 100, 5, 30, false),
			WithCaller(false),
		)
		defer l.Sync()
		defer os.Remove("/tmp/bench_caller_off.log")

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("benchmark message", zap.String("key", "value"))
			}
		})
	})
}

// BenchmarkMultipleOutputs 基准测试：多输出性能
func BenchmarkMultipleOutputs(b *testing.B) {
	b.Run("FileOnly", func(b *testing.B) {
		l, _ := New(
			WithLevel("info"),
			WithFormat("json"),
			WithConsole(false),
			WithFile("/tmp/bench_file_only.log", 100, 5, 30, false),
		)
		defer l.Sync()
		defer os.Remove("/tmp/bench_file_only.log")

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("benchmark message", zap.String("key", "value"))
			}
		})
	})

	b.Run("FileAndConsole", func(b *testing.B) {
		l, _ := New(
			WithLevel("info"),
			WithFormat("json"),
			WithConsole(true),
			WithFile("/tmp/bench_file_console.log", 100, 5, 30, false),
		)
		defer l.Sync()
		defer os.Remove("/tmp/bench_file_console.log")

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("benchmark message", zap.String("key", "value"))
			}
		})
	})
}

// BenchmarkGlobalLogger 基准测试：全局 Logger
func BenchmarkGlobalLogger(b *testing.B) {
	InitGlobal(
		WithLevel("info"),
		WithConsole(false),
		WithFile("/tmp/bench_global.log", 100, 5, 30, false),
	)
	defer Sync()
	defer os.Remove("/tmp/bench_global.log")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Info("benchmark message",
				zap.String("key", "value"),
				zap.Int("count", 42),
			)
		}
	})
}

// BenchmarkMemoryAllocation 基准测试：内存分配
func BenchmarkMemoryAllocation(b *testing.B) {
	l, _ := New(
		WithLevel("info"),
		WithConsole(false),
		WithFile("/tmp/bench_alloc.log", 100, 5, 30, false),
	)
	defer l.Sync()
	defer os.Remove("/tmp/bench_alloc.log")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("benchmark message",
			zap.String("key", "value"),
			zap.Int("count", i),
			zap.Duration("elapsed", 123*time.Millisecond),
		)
	}
}

// BenchmarkComparison 基准测试：对比不同风格
func BenchmarkComparison(b *testing.B) {
	l, _ := New(
		WithLevel("info"),
		WithConsole(false),
		WithFile("/tmp/bench_comparison.log", 100, 5, 30, false),
	)
	defer l.Sync()
	defer os.Remove("/tmp/bench_comparison.log")

	b.Run("Structured-ZeroAlloc", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("benchmark message",
					zap.String("key", "value"),
					zap.Int("count", 42),
					zap.Bool("success", true),
				)
			}
		})
	})

	b.Run("Sugar-KeyValue", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Infow("benchmark message",
					"key", "value",
					"count", 42,
					"success", true,
				)
			}
		})
	})

	b.Run("Formatted-Printf", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Infof("benchmark message: key=%s count=%d success=%t",
					"value", 42, true,
				)
			}
		})
	})
}
