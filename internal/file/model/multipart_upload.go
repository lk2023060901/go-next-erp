package model

import (
	"time"

	"github.com/google/uuid"
)

// UploadStatus 上传状态
type UploadStatus string

const (
	UploadStatusInProgress UploadStatus = "in_progress" // 进行中
	UploadStatusCompleted  UploadStatus = "completed"   // 已完成
	UploadStatusAborted    UploadStatus = "aborted"     // 已中止
)

// MultipartUpload 分片上传跟踪
type MultipartUpload struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// Upload info
	UploadID   string `json:"upload_id"`    // MinIO upload ID
	Filename   string `json:"filename"`     // 文件名
	StorageKey string `json:"storage_key"`  // 目标存储键
	TotalSize  *int64 `json:"total_size"`   // 预期总大小
	PartSize   int64  `json:"part_size"`    // 分片大小

	// Progress tracking
	UploadedParts []int `json:"uploaded_parts"` // 已完成的分片号列表
	TotalParts    *int  `json:"total_parts"`    // 总分片数

	// Metadata
	MimeType string                 `json:"mime_type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Status
	Status UploadStatus `json:"status"`

	// Owner
	CreatedBy uuid.UUID `json:"created_by"`

	// Timestamps
	ExpiresAt   time.Time  `json:"expires_at"`    // 上传过期时间
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// IsExpired 检查上传是否过期
func (mu *MultipartUpload) IsExpired() bool {
	return time.Now().After(mu.ExpiresAt)
}

// IsCompleted 检查上传是否完成
func (mu *MultipartUpload) IsCompleted() bool {
	return mu.Status == UploadStatusCompleted
}

// GetProgress 获取上传进度百分比
func (mu *MultipartUpload) GetProgress() float64 {
	if mu.TotalParts == nil || *mu.TotalParts == 0 {
		return 0
	}
	return float64(len(mu.UploadedParts)) / float64(*mu.TotalParts) * 100
}

// AddCompletedPart 添加已完成的分片
func (mu *MultipartUpload) AddCompletedPart(partNumber int) {
	// Check if already exists
	for _, p := range mu.UploadedParts {
		if p == partNumber {
			return
		}
	}
	mu.UploadedParts = append(mu.UploadedParts, partNumber)
}

// IsPartCompleted 检查分片是否已完成
func (mu *MultipartUpload) IsPartCompleted(partNumber int) bool {
	for _, p := range mu.UploadedParts {
		if p == partNumber {
			return true
		}
	}
	return false
}

// GetRemainingParts 获取剩余未完成的分片号
func (mu *MultipartUpload) GetRemainingParts() []int {
	if mu.TotalParts == nil {
		return []int{}
	}

	completed := make(map[int]bool)
	for _, p := range mu.UploadedParts {
		completed[p] = true
	}

	remaining := []int{}
	for i := 1; i <= *mu.TotalParts; i++ {
		if !completed[i] {
			remaining = append(remaining, i)
		}
	}

	return remaining
}
