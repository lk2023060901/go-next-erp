# Scheduler 使用指南

基于 `robfig/cron/v3` 封装的企业级任务调度器。

## 特性

- ✅ 标准 Cron 表达式支持
- ✅ 可选秒级精度
- ✅ 中间件系统（日志、恢复、监控）
- ✅ 任务元数据管理
- ✅ 优雅关闭
- ✅ 线程安全
- ✅ 集成项目日志系统

## 快速开始

### 1. 基础使用

```go
package main

import (
    "fmt"
    "github.com/lk2023060901/go-next-erp/pkg/scheduler"
)

func main() {
    // 创建调度器
    s := scheduler.New()

    // 添加任务
    s.AddFunc("daily-backup", "0 2 * * *", func() {
        fmt.Println("执行每日备份")
    })

    // 启动调度器
    s.Start()

    // 程序退出时优雅关闭
    defer s.Shutdown(context.Background())

    // 保持运行
    select {}
}
```

### 2. 秒级精度

```go
s := scheduler.New(
    scheduler.WithSeconds(), // 启用秒级精度
)

// Cron 表达式变为 6 个字段：秒 分 时 日 月 周
s.AddFunc("every-5s", "*/5 * * * * *", func() {
    fmt.Println("每5秒执行一次")
})
```

### 3. 自定义配置

```go
shanghai, _ := time.LoadLocation("Asia/Shanghai")

s := scheduler.New(
    scheduler.WithLocation(shanghai),       // 设置时区
    scheduler.WithSeconds(),                // 启用秒级
    scheduler.WithPanicRecovery(true),      // 启用 panic 恢复
    scheduler.WithMaxConcurrent(10),        // 最大并发10个任务
    scheduler.WithShutdownTimeout(60*time.Second), // 关闭超时60秒
)
```

### 4. 使用项目日志

```go
import (
    "github.com/lk2023060901/go-next-erp/pkg/logger"
    "github.com/lk2023060901/go-next-erp/pkg/scheduler"
)

// 创建自定义日志器
log := logger.New(logger.WithLevel("debug"))

s := scheduler.New(
    scheduler.WithLogger(log),
)
```

## Cron 表达式语法

### 标准格式（5 个字段）

```
┌─────────── 分钟 (0 - 59)
│ ┌─────────── 小时 (0 - 23)
│ │ ┌─────────── 日期 (1 - 31)
│ │ │ ┌─────────── 月份 (1 - 12)
│ │ │ │ ┌─────────── 星期 (0 - 6) (0 = 周日)
│ │ │ │ │
* * * * *
```

### 秒级格式（6 个字段）

```
┌─────────── 秒 (0 - 59)
│ ┌─────────── 分钟 (0 - 59)
│ │ ┌─────────── 小时 (0 - 23)
│ │ │ ┌─────────── 日期 (1 - 31)
│ │ │ │ ┌─────────── 月份 (1 - 12)
│ │ │ │ │ ┌─────────── 星期 (0 - 6)
│ │ │ │ │ │
* * * * * *
```

### 常用示例

| 表达式 | 说明 |
|--------|------|
| `0 0 * * *` | 每天午夜执行 |
| `0 */2 * * *` | 每2小时执行 |
| `0 9-17 * * 1-5` | 工作日9点到17点每小时执行 |
| `0 0 1 * *` | 每月1号执行 |
| `0 0 * * 0` | 每周日执行 |
| `*/5 * * * * *` | 每5秒执行（需启用秒级） |

### 特殊字符

- `*` : 任意值
- `,` : 列表分隔，如 `1,3,5`
- `-` : 范围，如 `1-5`
- `/` : 步长，如 `*/10`
- `@yearly` : 每年 (`0 0 1 1 *`)
- `@monthly` : 每月 (`0 0 1 * *`)
- `@weekly` : 每周 (`0 0 * * 0`)
- `@daily` : 每天 (`0 0 * * *`)
- `@hourly` : 每小时 (`0 * * * *`)

## 中间件使用

### 全局中间件

```go
s := scheduler.New()

// 应用全局中间件（对所有任务生效）
s.Use(
    scheduler.RecoveryMiddleware(log),  // Panic 恢复
    scheduler.LoggingMiddleware(log),   // 日志记录
    scheduler.MetricsMiddleware(log),   // 指标收集
)
```

### 中间件链

```go
// 组合多个中间件
middleware := scheduler.Chain(
    scheduler.RecoveryMiddleware(log),
    scheduler.LoggingMiddleware(log),
    scheduler.TimeoutMiddleware(5*time.Minute, log),
    scheduler.SkipIfRunningMiddleware(log),
)

s.Use(middleware)
```

### 自定义中间件

```go
func CustomMiddleware(log *logger.Logger) scheduler.Middleware {
    return func(next scheduler.Job) scheduler.Job {
        return scheduler.JobFunc(func() {
            // 执行前
            log.Info("before job")

            // 执行任务
            next.Run()

            // 执行后
            log.Info("after job")
        })
    }
}

s.Use(CustomMiddleware(log))
```

## 任务管理

### 添加任务

```go
// 方法1: 添加函数
jobID, err := s.AddFunc("my-task", "0 * * * *", func() {
    // 任务逻辑
})

// 方法2: 添加任务对象
type MyJob struct {
    name string
}

func (j *MyJob) Name() string { return j.name }
func (j *MyJob) Run() { /* 任务逻辑 */ }

jobID, err := s.AddJob("my-task", "0 * * * *", &MyJob{name: "my-task"})
```

### 移除任务

```go
err := s.RemoveJob(jobID)
```

### 查询任务

```go
// 获取单个任务元数据
meta, err := s.GetJob(jobID)
fmt.Printf("任务: %s, 下次执行: %s\n", meta.Name, meta.NextRun)

// 列出所有任务
jobs := s.ListJobs()
for _, job := range jobs {
    fmt.Printf("ID: %s, 名称: %s, 表达式: %s\n",
        job.ID, job.Name, job.Spec)
}

// 获取任务数量
count := s.JobCount()
```

## 生命周期管理

### 启动和停止

```go
// 启动调度器
if err := s.Start(); err != nil {
    log.Fatal(err)
}

// 检查运行状态
if s.IsRunning() {
    fmt.Println("调度器正在运行")
}

// 立即停止（不等待任务完成）
s.Stop()

// 优雅关闭（等待任务完成，带超时）
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
s.Shutdown(ctx)
```

## 完整示例

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/lk2023060901/go-next-erp/pkg/logger"
    "github.com/lk2023060901/go-next-erp/pkg/scheduler"
)

func main() {
    // 创建日志器
    log := logger.New(logger.WithLevel("info"))

    // 创建调度器
    s := scheduler.New(
        scheduler.WithLogger(log),
        scheduler.WithSeconds(),
        scheduler.WithLocation(time.UTC),
    )

    // 应用中间件
    s.Use(
        scheduler.RecoveryMiddleware(log),
        scheduler.LoggingMiddleware(log),
    )

    // 添加任务
    s.AddFunc("every-minute", "0 * * * *", func() {
        fmt.Println("每分钟执行")
    })

    s.AddFunc("every-5-seconds", "*/5 * * * * *", func() {
        fmt.Println("每5秒执行")
    })

    s.AddFunc("daily-report", "0 9 * * *", func() {
        fmt.Println("生成每日报表")
    })

    // 启动
    if err := s.Start(); err != nil {
        log.Fatal("failed to start scheduler", "error", err)
    }

    // 优雅关闭
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        if err := s.Shutdown(ctx); err != nil {
            log.Error("shutdown error", "error", err)
        }
    }()

    // 保持运行
    select {}
}
```

## 最佳实践

1. **生产环境建议启用 Panic 恢复**
   ```go
   scheduler.WithPanicRecovery(true)
   ```

2. **长任务使用跳过中间件**
   ```go
   s.Use(scheduler.SkipIfRunningMiddleware(log))
   ```

3. **设置合理的关闭超时**
   ```go
   scheduler.WithShutdownTimeout(60*time.Second)
   ```

4. **使用有意义的任务名称**
   ```go
   s.AddFunc("user-report-daily", ...) // ✅
   s.AddFunc("task1", ...)             // ❌
   ```

5. **监控任务执行状态**
   ```go
   meta, _ := s.GetJob(jobID)
   log.Infow("job stats",
       "run_count", meta.RunCount.Load(),
       "fail_count", meta.FailCount.Load(),
   )
   ```

## 常见问题

### 1. 任务重复执行？

使用 `SkipIfRunningMiddleware` 中间件：
```go
s.Use(scheduler.SkipIfRunningMiddleware(log))
```

### 2. 如何设置时区？

```go
loc, _ := time.LoadLocation("Asia/Shanghai")
scheduler.New(scheduler.WithLocation(loc))
```

### 3. 任务 panic 导致调度器停止？

确保启用 Panic 恢复：
```go
scheduler.New(scheduler.WithPanicRecovery(true))
```

### 4. 如何调试 Cron 表达式？

使用在线工具或项目日志查看 `next_run` 时间。

## 错误码

| 错误 | 说明 |
|------|------|
| `ErrSchedulerNotStarted` | 调度器未启动 |
| `ErrSchedulerAlreadyStarted` | 调度器已经启动 |
| `ErrJobNotFound` | 任务不存在 |
| `ErrJobAlreadyExists` | 任务名称已存在 |
| `ErrInvalidCronSpec` | 无效的 Cron 表达式 |
| `ErrShutdownTimeout` | 关闭超时 |
