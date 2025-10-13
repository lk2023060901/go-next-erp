package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SubjectType 配额主体类型
type SubjectType string

const (
	SubjectTypeTenant     SubjectType = "tenant"     // 租户级配额
	SubjectTypeUser       SubjectType = "user"       // 用户级配额
	SubjectTypeDepartment SubjectType = "department" // 部门级配额
)

// StorageQuota 存储配额
type StorageQuota struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// Quota subject
	SubjectType SubjectType `json:"subject_type"` // 主体类型
	SubjectID   *uuid.UUID  `json:"subject_id"`   // 主体 ID (NULL for tenant-level)

	// Quota limits (bytes)
	QuotaLimit    int64 `json:"quota_limit"`    // 配额限制
	QuotaUsed     int64 `json:"quota_used"`     // 已使用
	QuotaReserved int64 `json:"quota_reserved"` // 预留空间

	// File count limits
	FileCountLimit *int `json:"file_count_limit"` // 文件数量限制
	FileCountUsed  int  `json:"file_count_used"`  // 已使用数量

	// Settings
	Settings map[string]interface{} `json:"settings,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetAvailableQuota 获取可用配额
func (sq *StorageQuota) GetAvailableQuota() int64 {
	return sq.QuotaLimit - sq.QuotaUsed - sq.QuotaReserved
}

// GetUsagePercentage 获取使用百分比
func (sq *StorageQuota) GetUsagePercentage() float64 {
	if sq.QuotaLimit == 0 {
		return 0
	}
	return float64(sq.QuotaUsed) / float64(sq.QuotaLimit) * 100
}

// CanAllocate 检查是否可以分配指定大小
func (sq *StorageQuota) CanAllocate(size int64) bool {
	return sq.GetAvailableQuota() >= size
}

// IsQuotaExceeded 检查配额是否超限
func (sq *StorageQuota) IsQuotaExceeded() bool {
	return sq.QuotaUsed > sq.QuotaLimit
}

// IsNearLimit 检查是否接近配额限制
func (sq *StorageQuota) IsNearLimit(threshold float64) bool {
	return sq.GetUsagePercentage() >= threshold
}

// Reserve 预留空间
func (sq *StorageQuota) Reserve(size int64) error {
	if !sq.CanAllocate(size) {
		return fmt.Errorf("insufficient quota: available %d bytes, requested %d bytes",
			sq.GetAvailableQuota(), size)
	}
	sq.QuotaReserved += size
	return nil
}

// CommitReservation 提交预留（转为已使用）
func (sq *StorageQuota) CommitReservation(size int64) {
	if sq.QuotaReserved >= size {
		sq.QuotaReserved -= size
	}
	sq.QuotaUsed += size
	sq.FileCountUsed++
}

// ReleaseReservation 释放预留
func (sq *StorageQuota) ReleaseReservation(size int64) {
	sq.QuotaReserved -= size
	if sq.QuotaReserved < 0 {
		sq.QuotaReserved = 0
	}
}

// ReleaseUsage 释放已使用空间
func (sq *StorageQuota) ReleaseUsage(size int64) {
	sq.QuotaUsed -= size
	if sq.QuotaUsed < 0 {
		sq.QuotaUsed = 0
	}
	sq.FileCountUsed--
	if sq.FileCountUsed < 0 {
		sq.FileCountUsed = 0
	}
}

// FormatQuota 格式化配额信息
func (sq *StorageQuota) FormatQuota() string {
	return fmt.Sprintf("Used: %s / Limit: %s (%.1f%%)",
		formatBytes(sq.QuotaUsed),
		formatBytes(sq.QuotaLimit),
		sq.GetUsagePercentage(),
	)
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
