package scheduler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"github.com/robfig/cron/v3"
)

// Scheduler 任务调度器
type Scheduler struct {
	cron        *cron.Cron
	config      *Config
	logger      *logger.Logger
	middlewares []Middleware

	// 任务注册表 (jobID -> *JobMeta)
	jobs sync.Map

	// 运行状态
	running atomic.Bool
	mu      sync.RWMutex
}

// New 创建新的调度器
func New(opts ...Option) *Scheduler {
	// 创建默认日志器
	defaultLogger, _ := logger.New()

	s := &Scheduler{
		config: DefaultConfig(),
		logger: defaultLogger,
	}

	// 应用选项
	for _, opt := range opts {
		opt(s)
	}

	// 验证配置
	if err := s.config.Validate(); err != nil {
		s.logger.Warnw("config validation failed, using defaults", "error", err)
	}

	// 创建 cron 实例
	s.initCron()

	return s
}

// initCron 初始化 cron 实例
func (s *Scheduler) initCron() {
	cronOpts := []cron.Option{
		cron.WithLocation(s.config.Location),
		cron.WithLogger(&cronLoggerAdapter{s.logger}),
	}

	// 秒级精度
	if s.config.EnableSeconds {
		cronOpts = append(cronOpts, cron.WithSeconds())
	}

	// Panic 恢复
	if s.config.PanicRecovery {
		cronOpts = append(cronOpts, cron.WithChain(
			cron.Recover(&cronLoggerAdapter{s.logger}),
		))
	}

	s.cron = cron.New(cronOpts...)
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	if s.running.Load() {
		return ErrSchedulerAlreadyStarted
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cron.Start()
	s.running.Store(true)

	s.logger.Infow("scheduler started",
		"jobs", s.JobCount(),
		"timezone", s.config.Location.String(),
		"seconds_enabled", s.config.EnableSeconds,
	)

	return nil
}

// Stop 停止调度器（不等待任务完成）
func (s *Scheduler) Stop() error {
	if !s.running.Load() {
		return ErrSchedulerNotStarted
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cron.Stop()
	s.running.Store(false)

	s.logger.Infow("scheduler stopped")

	return nil
}

// Shutdown 优雅关闭调度器（等待任务完成）
func (s *Scheduler) Shutdown(ctx context.Context) error {
	if !s.running.Load() {
		return ErrSchedulerNotStarted
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 设置超时上下文
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()
	}

	// 停止接收新任务
	cronCtx := s.cron.Stop()

	// 等待任务完成或超时
	select {
	case <-cronCtx.Done():
		s.running.Store(false)
		s.logger.Infow("scheduler shutdown gracefully")
		return nil
	case <-ctx.Done():
		s.running.Store(false)
		s.logger.Warnw("scheduler shutdown timeout")
		return ErrShutdownTimeout
	}
}

// AddJob 添加任务
//
// name: 任务名称（必须唯一）
// spec: Cron 表达式
// job: 任务实例
//
// 返回任务 ID
func (s *Scheduler) AddJob(name, spec string, job Job) (string, error) {
	if name == "" {
		return "", ErrInvalidJobName
	}

	if job == nil {
		return "", ErrNilJob
	}

	// 检查任务名称是否已存在
	if s.hasJobByName(name) {
		return "", ErrJobAlreadyExists
	}

	// 应用中间件
	wrappedJob := s.applyMiddlewares(job)

	// 添加到 cron
	entryID, err := s.cron.AddJob(spec, wrappedJob)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidCronSpec, err)
	}

	// 生成任务 ID
	jobID := uuid.New().String()

	// 创建元数据
	meta := NewJobMeta(jobID, name, spec, entryID)

	// 更新执行时间
	entry := s.cron.Entry(entryID)
	meta.UpdateRunTime(entry.Next, entry.Prev)

	// 保存到注册表
	s.jobs.Store(jobID, meta)

	s.logger.Infow("job added",
		"job_id", jobID,
		"job_name", name,
		"spec", spec,
		"next_run", meta.NextRun.Format(time.RFC3339),
	)

	return jobID, nil
}

// AddFunc 添加函数任务（便捷方法）
func (s *Scheduler) AddFunc(name, spec string, cmd func()) (string, error) {
	return s.AddJob(name, spec, JobFunc(cmd))
}

// RemoveJob 移除任务
func (s *Scheduler) RemoveJob(jobID string) error {
	value, ok := s.jobs.Load(jobID)
	if !ok {
		return ErrJobNotFound
	}

	meta := value.(*JobMeta)

	// 从 cron 中移除
	s.cron.Remove(meta.EntryID)

	// 从注册表中删除
	s.jobs.Delete(jobID)

	s.logger.Infow("job removed",
		"job_id", jobID,
		"job_name", meta.Name,
	)

	return nil
}

// GetJob 获取任务元数据
func (s *Scheduler) GetJob(jobID string) (*JobMeta, error) {
	value, ok := s.jobs.Load(jobID)
	if !ok {
		return nil, ErrJobNotFound
	}

	meta := value.(*JobMeta)

	// 更新执行时间
	entry := s.cron.Entry(meta.EntryID)
	meta.UpdateRunTime(entry.Next, entry.Prev)

	return meta, nil
}

// ListJobs 列出所有任务
func (s *Scheduler) ListJobs() []*JobMeta {
	var jobs []*JobMeta

	s.jobs.Range(func(key, value interface{}) bool {
		meta := value.(*JobMeta)

		// 更新执行时间
		entry := s.cron.Entry(meta.EntryID)
		meta.UpdateRunTime(entry.Next, entry.Prev)

		jobs = append(jobs, meta)
		return true
	})

	return jobs
}

// JobCount 获取任务数量
func (s *Scheduler) JobCount() int {
	count := 0
	s.jobs.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// IsRunning 检查调度器是否运行中
func (s *Scheduler) IsRunning() bool {
	return s.running.Load()
}

// Use 添加全局中间件
func (s *Scheduler) Use(middlewares ...Middleware) {
	s.middlewares = append(s.middlewares, middlewares...)
}

// applyMiddlewares 应用中间件
func (s *Scheduler) applyMiddlewares(job Job) cron.Job {
	// 从右到左应用中间件（洋葱模型）
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		job = s.middlewares[i](job)
	}

	return job
}

// hasJobByName 检查任务名称是否已存在
func (s *Scheduler) hasJobByName(name string) bool {
	found := false
	s.jobs.Range(func(key, value interface{}) bool {
		meta := value.(*JobMeta)
		if meta.Name == name {
			found = true
			return false // 停止遍历
		}
		return true
	})
	return found
}

// cronLoggerAdapter 适配项目日志到 cron.Logger 接口
type cronLoggerAdapter struct {
	logger *logger.Logger
}

// Info 实现 cron.Logger 接口
func (a *cronLoggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	a.logger.Infow(msg, keysAndValues...)
}

// Error 实现 cron.Logger 接口
func (a *cronLoggerAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	fields := append(keysAndValues, "error", err)
	a.logger.Errorw(msg, fields...)
}

// ParseCronSpec 解析 Cron 表达式（工具方法）
//
// 返回字段含义（当 EnableSeconds=true）:
// - 秒 分 时 日 月 周
//
// 返回字段含义（当 EnableSeconds=false）:
// - 分 时 日 月 周
func ParseCronSpec(spec string) (fields []string, err error) {
	fields = strings.Fields(spec)

	// 基本校验
	if len(fields) < 5 || len(fields) > 6 {
		return nil, fmt.Errorf("invalid cron spec: must have 5 or 6 fields")
	}

	return fields, nil
}
