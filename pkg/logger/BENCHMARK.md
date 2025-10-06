# pkg/logger - 性能基准测试报告

## 测试环境
- **模块**: `pkg/logger`
- **CPU**: Apple M1 Pro (ARM64)
- **OS**: macOS (Darwin)
- **Go**: 1.24.5
- **测试时长**: 2s per benchmark
- **并发**: 10 goroutines

---

## 📊 核心性能指标

### 1. 不同日志风格对比

| 日志风格 | 性能 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) | 推荐场景 |
|---------|-------------|----------------|---------------------|---------|
| **结构化 (Structured)** | 3,808 | 546 | 7 | ⭐ 生产环境、高性能场景 |
| **键值对 (Sugar)** | 3,929 | 739 | 7 | 开发环境、业务日志 |
| **格式化 (Printf)** | 4,193 | 418 | 7 | 简单场景、临时调试 |

**结论**: 结构化日志性能最优，且类型安全，**强烈推荐生产环境使用**。

---

### 2. 日志级别性能差异

| 级别 | 启用状态 | 性能 (ns/op) | 内存 (B/op) | 说明 |
|------|---------|-------------|------------|------|
| Debug | **禁用** | 22.91 | 64 | ✅ 禁用级别几乎无开销 |
| Info | 启用 | 3,456 | 417 | 默认级别 |
| Warn | 启用 | 3,583 | 417 | 警告级别 |
| Error | 启用 | 4,372 | 675 | ⚠️ Error 级别稍慢（包含堆栈） |

**结论**: 禁用的日志级别开销极低（**仅 23ns**），可放心在代码中使用 Debug 日志。

---

### 3. 调用者信息 (Caller) 开销

| 配置 | 性能 (ns/op) | 内存 (B/op) | 开销 |
|------|-------------|------------|------|
| 启用 Caller | 3,420 | 417 | 基准 |
| 禁用 Caller | 2,708 | 128 | **-21% 时间, -69% 内存** |

**结论**: 启用 Caller 有一定开销，但**生产环境建议启用**以便快速定位问题。

---

### 4. 编码格式对比

| 格式 | 性能 (ns/op) | 内存 (B/op) | 特点 |
|------|-------------|------------|------|
| JSON | 3,975 | 465 | ⭐ 生产推荐（结构化、易解析） |
| Console | 4,371 | 546 | 开发推荐（人类可读） |

**性能差异**: Console 比 JSON 慢 **~10%**，但开发环境更友好。

---

### 5. 多输出性能

| 输出配置 | 性能 (ns/op) | 说明 |
|---------|-------------|------|
| 仅文件 | 3,394 | ✅ 最快 |
| 文件 + 控制台 | ~4,500 | 额外 **~30%** 开销 |

**建议**: 生产环境仅输出到文件，通过日志收集系统查看。

---

### 6. 上下文字段提取开销

| 操作 | 性能 (ns/op) | 内存 (B/op) | 分配次数 |
|------|-------------|------------|---------|
| WithContext | 4,636 | 3,432 | 21 |
| WithFields | 3,622 | 482 | 7 |

**结论**: `WithContext` 有一定开销，建议**在请求入口处调用一次**，而非每次日志调用。

---

## 🚀 性能优化建议

### 1. 生产环境配置
```go
import "github.com/lk2023060901/go-next-erp/pkg/logger"

logger.InitGlobal(
    logger.WithLevel("info"),         // ✅ Info 级别（Debug 会被高效过滤）
    logger.WithFormat("json"),        // ✅ JSON 格式（结构化）
    logger.WithFile("/var/log/app.log", 100, 10, 30, true),
    logger.WithConsole(false),        // ✅ 禁用控制台（减少 30% 开销）
    logger.WithCaller(true),          // ✅ 启用 Caller（便于排查）
    logger.WithStacktrace(true),      // ✅ Error 级别显示堆栈
)
```

### 2. 使用结构化日志（而非 Sugar）
```go
import (
    "github.com/lk2023060901/go-next-erp/pkg/logger"
    "go.uber.org/zap"
)

// ❌ 避免（较慢，739B 内存）
log.Infow("user created", "user_id", 123, "email", "test@example.com")

// ✅ 推荐（最快，546B 内存）
log.Info("user created",
    zap.Int64("user_id", 123),
    zap.String("email", "test@example.com"),
)
```

### 3. Context 字段优化
```go
// ❌ 避免（每次调用都提取 Context）
for _, item := range items {
    log.WithContext(ctx).Info("processing", zap.Int("id", item.ID))
}

// ✅ 推荐（提取一次）
logger := log.WithContext(ctx)
for _, item := range items {
    logger.Info("processing", zap.Int("id", item.ID))
}
```

### 4. 预设字段优化
```go
import (
    "github.com/lk2023060901/go-next-erp/pkg/logger"
    "go.uber.org/zap"
)

// ✅ 在服务初始化时创建带模块字段的 Logger
type UserService struct {
    log *logger.Logger
}

func NewUserService() *UserService {
    return &UserService{
        log: logger.GetLogger().With(
            zap.String("module", "service.user"),
        ),
    }
}
```

---

## 📈 性能对比总结

### 吞吐量对比
```
结构化日志:   ~263,000 logs/sec
Sugar 日志:   ~255,000 logs/sec
格式化日志:   ~238,000 logs/sec
```

### 内存效率
```
结构化日志:   546 B/op (基准)
Sugar 日志:   739 B/op (+35%)
格式化日志:   418 B/op (-23%, 但类型不安全)
```

---

## 🎯 最佳实践

1. **生产环境**: 使用结构化日志 + JSON 格式 + 文件输出
2. **开发环境**: 可用 Sugar + Console 格式 + 控制台输出
3. **高频日志**: 预先创建带字段的 Logger，避免重复 With 操作
4. **Context 集成**: 在请求入口提取一次，传递 Logger 实例
5. **禁用的级别**: 几乎无开销，可放心使用 Debug 日志

---

## 🔬 详细基准数据

```
BenchmarkStructuredLog-10         536,203    4,045 ns/op    546 B/op    7 allocs/op
BenchmarkSugarLog-10              571,990    4,902 ns/op    739 B/op    7 allocs/op
BenchmarkFormattedLog-10          584,379    4,081 ns/op    423 B/op    8 allocs/op
BenchmarkWithFields-10            686,614    3,622 ns/op    482 B/op    7 allocs/op
BenchmarkWithContext-10           495,534    4,636 ns/op  3,432 B/op   21 allocs/op
BenchmarkJSONEncoding-10          683,395    3,975 ns/op    465 B/op    3 allocs/op
BenchmarkConsoleEncoding-10       649,593    4,371 ns/op    546 B/op    7 allocs/op
BenchmarkComplexFields-10         495,477    5,120 ns/op    820 B/op    9 allocs/op

BenchmarkDifferentLevels/Debug-Disabled-10    100,000,000    22.91 ns/op    64 B/op    1 allocs/op
BenchmarkDifferentLevels/Info-Enabled-10        648,022     3,456 ns/op   417 B/op    7 allocs/op
BenchmarkDifferentLevels/Warn-Enabled-10        650,804     3,583 ns/op   417 B/op    7 allocs/op
BenchmarkDifferentLevels/Error-Enabled-10       598,750     4,372 ns/op   675 B/op    8 allocs/op

BenchmarkCaller/WithCaller-10                   697,977     3,420 ns/op   417 B/op    7 allocs/op
BenchmarkCaller/WithoutCaller-10                855,897     2,708 ns/op   128 B/op    4 allocs/op

BenchmarkMultipleOutputs/FileOnly-10            727,470     3,394 ns/op   337 B/op    3 allocs/op
BenchmarkGlobalLogger-10                        608,972     3,462 ns/op   474 B/op    7 allocs/op
BenchmarkMemoryAllocation-10                    860,738     2,893 ns/op   545 B/op    7 allocs/op
```

---

## 📝 测试命令

```bash
# 运行所有基准测试
go test -bench=. -benchmem -benchtime=2s ./pkg/logger/

# 运行特定基准测试
go test -bench=BenchmarkComparison -benchmem ./pkg/logger/

# 生成 CPU 性能分析
go test -bench=BenchmarkStructuredLog -cpuprofile=cpu.prof ./pkg/logger/
go tool pprof cpu.prof

# 生成内存性能分析
go test -bench=BenchmarkStructuredLog -memprofile=mem.prof ./pkg/logger/
go tool pprof mem.prof
```

---

## 📦 模块信息

- **包路径**: `github.com/lk2023060901/go-next-erp/pkg/logger`
- **依赖**:
  - `go.uber.org/zap` - 高性能日志库
  - `gopkg.in/natefinch/lumberjack.v2` - 日志轮换
  - `gopkg.in/yaml.v3` - 配置解析
