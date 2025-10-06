package scheduler

import (
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
)

// Job 任务接口 - 所有任务必须实现此接口
type Job interface {
	Run()
}

// NamedJob 带名称的任务接口
type NamedJob interface {
	Job
	Name() string
}

// JobFunc 函数类型适配器，使普通函数实现 Job 接口
type JobFunc func()

// Run 实现 Job 接口
func (f JobFunc) Run() {
	f()
}

// JobStatus 任务状态
type JobStatus int

const (
	// JobStatusPending 待执行
	JobStatusPending JobStatus = iota
	// JobStatusRunning 执行中
	JobStatusRunning
	// JobStatusCompleted 执行完成
	JobStatusCompleted
	// JobStatusFailed 执行失败
	JobStatusFailed
)

// String 实现 Stringer 接口
func (s JobStatus) String() string {
	switch s {
	case JobStatusPending:
		return "pending"
	case JobStatusRunning:
		return "running"
	case JobStatusCompleted:
		return "completed"
	case JobStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// JobMeta 任务元数据
type JobMeta struct {
	// ID 任务唯一标识
	ID string

	// Name 任务名称
	Name string

	// Spec Cron 表达式
	Spec string

	// EntryID cron 内部 ID
	EntryID cron.EntryID

	// Status 任务状态
	Status JobStatus

	// NextRun 下次执行时间
	NextRun time.Time

	// PrevRun 上次执行时间
	PrevRun time.Time

	// RunCount 执行次数
	RunCount atomic.Int64

	// FailCount 失败次数
	FailCount atomic.Int64

	// LastError 最后一次错误
	LastError error

	// CreatedAt 创建时间
	CreatedAt time.Time

	// UpdatedAt 更新时间
	UpdatedAt time.Time
}

// NewJobMeta 创建任务元数据
func NewJobMeta(id, name, spec string, entryID cron.EntryID) *JobMeta {
	now := time.Now()
	return &JobMeta{
		ID:        id,
		Name:      name,
		Spec:      spec,
		EntryID:   entryID,
		Status:    JobStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IncRunCount 增加执行次数
func (m *JobMeta) IncRunCount() {
	m.RunCount.Add(1)
	m.UpdatedAt = time.Now()
}

// IncFailCount 增加失败次数
func (m *JobMeta) IncFailCount() {
	m.FailCount.Add(1)
	m.UpdatedAt = time.Now()
}

// SetStatus 设置任务状态
func (m *JobMeta) SetStatus(status JobStatus) {
	m.Status = status
	m.UpdatedAt = time.Now()
}

// SetLastError 设置最后错误
func (m *JobMeta) SetLastError(err error) {
	m.LastError = err
	m.UpdatedAt = time.Now()
}

// UpdateRunTime 更新执行时间
func (m *JobMeta) UpdateRunTime(next, prev time.Time) {
	m.NextRun = next
	m.PrevRun = prev
	m.UpdatedAt = time.Now()
}

// wrappedJob 包装后的任务（用于追踪执行状态）
type wrappedJob struct {
	job  Job
	meta *JobMeta
}

// Run 实现 cron.Job 接口
func (w *wrappedJob) Run() {
	w.meta.SetStatus(JobStatusRunning)
	w.meta.PrevRun = time.Now()
	w.meta.IncRunCount()

	defer func() {
		if r := recover(); r != nil {
			w.meta.IncFailCount()
			w.meta.SetStatus(JobStatusFailed)
		} else {
			w.meta.SetStatus(JobStatusCompleted)
		}
	}()

	w.job.Run()
}
