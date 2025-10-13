package model

import (
	"time"

	"github.com/google/uuid"
)

// DownloadStats 下载统计记录
type DownloadStats struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`
	FileID   uuid.UUID `json:"file_id"`

	// 下载信息
	DownloadedBy uuid.UUID `json:"downloaded_by"` // 下载者
	IPAddress    string    `json:"ip_address"`    // IP 地址
	UserAgent    string    `json:"user_agent"`    // User-Agent

	// 统计信息
	BytesDownloaded int64         `json:"bytes_downloaded"` // 下载字节数（支持断点续传）
	DownloadTime    time.Duration `json:"download_time"`    // 下载耗时（毫秒）
	IsComplete      bool          `json:"is_complete"`      // 是否完整下载
	IsResumed       bool          `json:"is_resumed"`       // 是否断点续传

	// 时间戳
	DownloadedAt time.Time `json:"downloaded_at"`
}

// FileDownloadSummary 文件下载汇总
type FileDownloadSummary struct {
	FileID            uuid.UUID  `json:"file_id"`
	TotalDownloads    int64      `json:"total_downloads"`    // 总下载次数
	UniqueDownloads   int64      `json:"unique_downloads"`   // 唯一下载用户数
	TotalBytes        int64      `json:"total_bytes"`        // 总下载流量
	CompleteDownloads int64      `json:"complete_downloads"` // 完整下载次数
	ResumedDownloads  int64      `json:"resumed_downloads"`  // 断点续传次数
	LastDownloadedAt  *time.Time `json:"last_downloaded_at"` // 最后下载时间
}

// TenantDownloadSummary 租户下载汇总
type TenantDownloadSummary struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	TotalDownloads int64     `json:"total_downloads"` // 总下载次数
	TotalBytes     int64     `json:"total_bytes"`     // 总下载流量
	TotalFiles     int64     `json:"total_files"`     // 被下载的文件数
	ActiveUsers    int64     `json:"active_users"`    // 活跃下载用户数
	Period         string    `json:"period"`          // 统计周期（day/week/month）
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
}

// UserDownloadSummary 用户下载汇总
type UserDownloadSummary struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	UserID         uuid.UUID `json:"user_id"`
	TotalDownloads int64     `json:"total_downloads"` // 总下载次数
	TotalBytes     int64     `json:"total_bytes"`     // 总下载流量
	TotalFiles     int64     `json:"total_files"`     // 下载的文件数
	Period         string    `json:"period"`          // 统计周期
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
}
