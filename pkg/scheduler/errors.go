package scheduler

import "errors"

var (
	// ErrSchedulerNotStarted 调度器未启动
	ErrSchedulerNotStarted = errors.New("scheduler not started")

	// ErrSchedulerAlreadyStarted 调度器已经启动
	ErrSchedulerAlreadyStarted = errors.New("scheduler already started")

	// ErrSchedulerStopped 调度器已停止
	ErrSchedulerStopped = errors.New("scheduler stopped")

	// ErrJobNotFound 任务不存在
	ErrJobNotFound = errors.New("job not found")

	// ErrJobAlreadyExists 任务已存在
	ErrJobAlreadyExists = errors.New("job already exists")

	// ErrInvalidCronSpec 无效的 Cron 表达式
	ErrInvalidCronSpec = errors.New("invalid cron spec")

	// ErrInvalidJobName 无效的任务名称
	ErrInvalidJobName = errors.New("invalid job name: must not be empty")

	// ErrNilJob 任务为空
	ErrNilJob = errors.New("job is nil")

	// ErrShutdownTimeout 关闭超时
	ErrShutdownTimeout = errors.New("shutdown timeout")
)
