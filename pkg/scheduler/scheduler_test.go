package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lk2023060901/go-next-erp/pkg/logger"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}

	if s.config == nil {
		t.Error("config is nil")
	}

	if s.logger == nil {
		t.Error("logger is nil")
	}

	if s.cron == nil {
		t.Error("cron is nil")
	}
}

func TestScheduler_StartStop(t *testing.T) {
	s := New()

	// 测试启动
	if err := s.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	if !s.IsRunning() {
		t.Error("IsRunning() should return true after Start()")
	}

	// 测试重复启动
	if err := s.Start(); err != ErrSchedulerAlreadyStarted {
		t.Errorf("Start() should return ErrSchedulerAlreadyStarted, got: %v", err)
	}

	// 测试停止
	if err := s.Stop(); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	if s.IsRunning() {
		t.Error("IsRunning() should return false after Stop()")
	}

	// 测试重复停止
	if err := s.Stop(); err != ErrSchedulerNotStarted {
		t.Errorf("Stop() should return ErrSchedulerNotStarted, got: %v", err)
	}
}

func TestScheduler_AddFunc(t *testing.T) {
	s := New()

	var executed atomic.Bool
	jobID, err := s.AddFunc("test-job", "* * * * *", func() {
		executed.Store(true)
	})

	if err != nil {
		t.Fatalf("AddFunc() failed: %v", err)
	}

	if jobID == "" {
		t.Error("jobID should not be empty")
	}

	// 检查任务是否添加
	meta, err := s.GetJob(jobID)
	if err != nil {
		t.Fatalf("GetJob() failed: %v", err)
	}

	if meta.Name != "test-job" {
		t.Errorf("job name = %s, want test-job", meta.Name)
	}

	if meta.Spec != "* * * * *" {
		t.Errorf("job spec = %s, want * * * * *", meta.Spec)
	}
}

func TestScheduler_AddFunc_InvalidName(t *testing.T) {
	s := New()

	_, err := s.AddFunc("", "* * * * *", func() {})
	if err != ErrInvalidJobName {
		t.Errorf("AddFunc() with empty name should return ErrInvalidJobName, got: %v", err)
	}
}

func TestScheduler_AddFunc_InvalidCronSpec(t *testing.T) {
	s := New()

	_, err := s.AddFunc("test", "invalid", func() {})
	if err == nil {
		t.Error("AddFunc() with invalid cron spec should return error")
	}
}

func TestScheduler_AddFunc_DuplicateName(t *testing.T) {
	s := New()

	// 添加第一个任务
	_, err := s.AddFunc("duplicate", "* * * * *", func() {})
	if err != nil {
		t.Fatalf("first AddFunc() failed: %v", err)
	}

	// 尝试添加同名任务
	_, err = s.AddFunc("duplicate", "* * * * *", func() {})
	if err != ErrJobAlreadyExists {
		t.Errorf("AddFunc() with duplicate name should return ErrJobAlreadyExists, got: %v", err)
	}
}

func TestScheduler_AddFunc_NilJob(t *testing.T) {
	s := New()

	_, err := s.AddJob("test", "* * * * *", nil)
	if err != ErrNilJob {
		t.Errorf("AddJob() with nil job should return ErrNilJob, got: %v", err)
	}
}

func TestScheduler_RemoveJob(t *testing.T) {
	s := New()

	jobID, err := s.AddFunc("test", "* * * * *", func() {})
	if err != nil {
		t.Fatalf("AddFunc() failed: %v", err)
	}

	// 移除任务
	if err := s.RemoveJob(jobID); err != nil {
		t.Fatalf("RemoveJob() failed: %v", err)
	}

	// 检查任务是否已移除
	_, err = s.GetJob(jobID)
	if err != ErrJobNotFound {
		t.Errorf("GetJob() after RemoveJob() should return ErrJobNotFound, got: %v", err)
	}
}

func TestScheduler_RemoveJob_NotFound(t *testing.T) {
	s := New()

	err := s.RemoveJob("non-existent-id")
	if err != ErrJobNotFound {
		t.Errorf("RemoveJob() with non-existent ID should return ErrJobNotFound, got: %v", err)
	}
}

func TestScheduler_GetJob(t *testing.T) {
	s := New()

	jobID, err := s.AddFunc("test", "* * * * *", func() {})
	if err != nil {
		t.Fatalf("AddFunc() failed: %v", err)
	}

	meta, err := s.GetJob(jobID)
	if err != nil {
		t.Fatalf("GetJob() failed: %v", err)
	}

	if meta.ID != jobID {
		t.Errorf("meta.ID = %s, want %s", meta.ID, jobID)
	}

	if meta.Name != "test" {
		t.Errorf("meta.Name = %s, want test", meta.Name)
	}
}

func TestScheduler_ListJobs(t *testing.T) {
	s := New()

	// 添加多个任务
	s.AddFunc("job1", "* * * * *", func() {})
	s.AddFunc("job2", "0 * * * *", func() {})
	s.AddFunc("job3", "0 0 * * *", func() {})

	jobs := s.ListJobs()
	if len(jobs) != 3 {
		t.Errorf("ListJobs() returned %d jobs, want 3", len(jobs))
	}
}

func TestScheduler_JobCount(t *testing.T) {
	s := New()

	if s.JobCount() != 0 {
		t.Errorf("JobCount() = %d, want 0", s.JobCount())
	}

	s.AddFunc("job1", "* * * * *", func() {})
	if s.JobCount() != 1 {
		t.Errorf("JobCount() = %d, want 1", s.JobCount())
	}

	s.AddFunc("job2", "* * * * *", func() {})
	if s.JobCount() != 2 {
		t.Errorf("JobCount() = %d, want 2", s.JobCount())
	}
}

func TestScheduler_WithSeconds(t *testing.T) {
	s := New(WithSeconds())

	// 测试秒级表达式
	_, err := s.AddFunc("every-5s", "*/5 * * * * *", func() {})
	if err != nil {
		t.Errorf("AddFunc() with seconds should succeed, got error: %v", err)
	}
}

func TestScheduler_WithLogger(t *testing.T) {
	customLogger, _ := logger.New(logger.WithLevel("debug"))
	s := New(WithLogger(customLogger))

	if s.logger != customLogger {
		t.Error("custom logger not set")
	}
}

func TestScheduler_WithLocation(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	s := New(WithLocation(loc))

	if s.config.Location != loc {
		t.Error("location not set")
	}
}

func TestScheduler_WithPanicRecovery(t *testing.T) {
	s := New(WithPanicRecovery(true))

	if !s.config.PanicRecovery {
		t.Error("panic recovery not enabled")
	}

	s = New(WithPanicRecovery(false))
	if s.config.PanicRecovery {
		t.Error("panic recovery should be disabled")
	}
}

func TestScheduler_Shutdown(t *testing.T) {
	s := New()
	s.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() failed: %v", err)
	}

	if s.IsRunning() {
		t.Error("IsRunning() should return false after Shutdown()")
	}
}

func TestScheduler_Shutdown_NotStarted(t *testing.T) {
	s := New()

	err := s.Shutdown(context.Background())
	if err != ErrSchedulerNotStarted {
		t.Errorf("Shutdown() on not started scheduler should return ErrSchedulerNotStarted, got: %v", err)
	}
}

func TestScheduler_JobExecution(t *testing.T) {
	s := New(WithSeconds())

	var executionCount atomic.Int64
	_, err := s.AddFunc("test", "* * * * * *", func() {
		executionCount.Add(1)
	})
	if err != nil {
		t.Fatalf("AddFunc() failed: %v", err)
	}

	s.Start()
	defer s.Stop()

	// 等待任务执行几次
	time.Sleep(3 * time.Second)

	count := executionCount.Load()
	if count < 2 {
		t.Errorf("job executed %d times, expected at least 2", count)
	}
}

func TestScheduler_Use(t *testing.T) {
	s := New()

	var middlewareExecuted atomic.Bool
	testMiddleware := func(next Job) Job {
		return JobFunc(func() {
			middlewareExecuted.Store(true)
			next.Run()
		})
	}

	s.Use(testMiddleware)

	if len(s.middlewares) != 1 {
		t.Errorf("middleware count = %d, want 1", len(s.middlewares))
	}
}

func TestJobMeta_IncRunCount(t *testing.T) {
	meta := NewJobMeta("id", "name", "spec", 1)

	meta.IncRunCount()
	if meta.RunCount.Load() != 1 {
		t.Errorf("RunCount = %d, want 1", meta.RunCount.Load())
	}

	meta.IncRunCount()
	if meta.RunCount.Load() != 2 {
		t.Errorf("RunCount = %d, want 2", meta.RunCount.Load())
	}
}

func TestJobMeta_IncFailCount(t *testing.T) {
	meta := NewJobMeta("id", "name", "spec", 1)

	meta.IncFailCount()
	if meta.FailCount.Load() != 1 {
		t.Errorf("FailCount = %d, want 1", meta.FailCount.Load())
	}
}

func TestJobMeta_SetStatus(t *testing.T) {
	meta := NewJobMeta("id", "name", "spec", 1)

	meta.SetStatus(JobStatusRunning)
	if meta.Status != JobStatusRunning {
		t.Errorf("Status = %v, want JobStatusRunning", meta.Status)
	}
}

func TestJobStatus_String(t *testing.T) {
	tests := []struct {
		status JobStatus
		want   string
	}{
		{JobStatusPending, "pending"},
		{JobStatusRunning, "running"},
		{JobStatusCompleted, "completed"},
		{JobStatusFailed, "failed"},
		{JobStatus(999), "unknown"},
	}

	for _, tt := range tests {
		got := tt.status.String()
		if got != tt.want {
			t.Errorf("JobStatus(%d).String() = %s, want %s", tt.status, got, tt.want)
		}
	}
}

func TestJobFunc(t *testing.T) {
	var executed atomic.Bool
	job := JobFunc(func() {
		executed.Store(true)
	})

	job.Run()

	if !executed.Load() {
		t.Error("JobFunc did not execute")
	}
}

func TestParseCronSpec(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		{"valid 5 fields", "0 * * * *", false},
		{"valid 6 fields", "0 0 * * * *", false},
		{"invalid 4 fields", "0 * * *", true},
		{"invalid 7 fields", "0 0 0 * * * *", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCronSpec(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCronSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestScheduler_Concurrent 测试并发安全性
func TestScheduler_Concurrent(t *testing.T) {
	s := New()
	s.Start()
	defer s.Stop()

	// 并发添加任务
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			_, err := s.AddFunc(
				time.Now().String()+string(rune(n)),
				"* * * * *",
				func() {},
			)
			if err != nil {
				t.Errorf("concurrent AddFunc() failed: %v", err)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证任务数量
	if s.JobCount() != 10 {
		t.Errorf("JobCount() = %d, want 10", s.JobCount())
	}
}

// BenchmarkScheduler_AddFunc 性能测试
func BenchmarkScheduler_AddFunc(b *testing.B) {
	s := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.AddFunc(time.Now().String(), "* * * * *", func() {})
	}
}

// BenchmarkScheduler_ListJobs 性能测试
func BenchmarkScheduler_ListJobs(b *testing.B) {
	s := New()

	// 添加100个任务
	for i := 0; i < 100; i++ {
		s.AddFunc(time.Now().String()+string(rune(i)), "* * * * *", func() {})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ListJobs()
	}
}
