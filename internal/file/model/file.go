package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FileStatus 文件状态
type FileStatus string

const (
	FileStatusActive   FileStatus = "active"   // 活跃
	FileStatusArchived FileStatus = "archived" // 已归档
	FileStatusDeleted  FileStatus = "deleted"  // 已删除
)

// VirusScanResult 病毒扫描结果
type VirusScanResult string

const (
	VirusScanClean    VirusScanResult = "clean"    // 干净
	VirusScanInfected VirusScanResult = "infected" // 感染
	VirusScanError    VirusScanResult = "error"    // 扫描错误
)

// AccessLevel 访问级别
type AccessLevel string

const (
	AccessLevelPrivate AccessLevel = "private" // 私有（仅上传者）
	AccessLevelTenant  AccessLevel = "tenant"  // 租户内可见
	AccessLevelPublic  AccessLevel = "public"  // 公开
)

// File 文件元数据
type File struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// Basic info
	Filename    string `json:"filename"`      // 原始文件名
	StorageKey  string `json:"storage_key"`   // MinIO 存储键
	Size        int64  `json:"size"`          // 文件大小（字节）
	MimeType    string `json:"mime_type"`     // MIME 类型
	ContentType string `json:"content_type"`  // Content-Type

	// Security & integrity
	Checksum         string           `json:"checksum"`           // SHA-256 哈希
	VirusScanned     bool             `json:"virus_scanned"`      // 是否已扫描
	VirusScanResult  *VirusScanResult `json:"virus_scan_result"`  // 扫描结果
	VirusScannedAt   *time.Time       `json:"virus_scanned_at"`   // 扫描时间

	// Metadata
	Extension string   `json:"extension"` // 文件扩展名
	Bucket    string   `json:"bucket"`    // MinIO 存储桶
	Category  string   `json:"category"`  // 文件分类
	Tags      []string `json:"tags"`      // 标签

	// Status & flags
	Status      FileStatus  `json:"status"`        // 文件状态
	IsTemporary bool        `json:"is_temporary"`  // 临时文件标志
	IsPublic    bool        `json:"is_public"`     // 公开访问标志

	// Version control
	VersionNumber int        `json:"version_number"`  // 当前版本号
	ParentFileID  *uuid.UUID `json:"parent_file_id"`  // 父文件 ID（版本）

	// Compression & watermark
	IsCompressed  bool    `json:"is_compressed"`   // 是否压缩
	HasWatermark  bool    `json:"has_watermark"`   // 是否有水印
	WatermarkText *string `json:"watermark_text"`  // 水印文本

	// Access control
	UploadedBy  uuid.UUID   `json:"uploaded_by"`   // 上传者
	AccessLevel AccessLevel `json:"access_level"`  // 访问级别

	// Preview & thumbnail
	ThumbnailKey     *string    `json:"thumbnail_key"`       // 缩略图存储键
	PreviewURL       *string    `json:"preview_url"`         // 预览 URL
	PreviewExpiresAt *time.Time `json:"preview_expires_at"`  // 预览 URL 过期时间

	// Lifecycle
	ExpiresAt  *time.Time `json:"expires_at"`   // 过期时间（临时文件）
	ArchivedAt *time.Time `json:"archived_at"`  // 归档时间

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`

	// Additional metadata (JSON)
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// IsActive 检查文件是否活跃
func (f *File) IsActive() bool {
	return f.Status == FileStatusActive && f.DeletedAt == nil
}

// IsExpired 检查文件是否过期
func (f *File) IsExpired() bool {
	if f.ExpiresAt == nil {
		return false
	}
	return f.ExpiresAt.Before(time.Now())
}

// IsVirusFree 检查文件是否通过病毒扫描
func (f *File) IsVirusFree() bool {
	if !f.VirusScanned {
		return false // 未扫描视为不安全
	}
	return f.VirusScanResult != nil && *f.VirusScanResult == VirusScanClean
}

// CanAccess 检查用户是否可以访问文件
func (f *File) CanAccess(userID uuid.UUID, tenantID uuid.UUID) bool {
	// Deleted files cannot be accessed
	if f.DeletedAt != nil {
		return false
	}

	// Different tenant
	if f.TenantID != tenantID {
		return false
	}

	// Public files
	if f.AccessLevel == AccessLevelPublic {
		return true
	}

	// Tenant level access
	if f.AccessLevel == AccessLevelTenant && f.TenantID == tenantID {
		return true
	}

	// Private - only uploader
	if f.AccessLevel == AccessLevelPrivate {
		return f.UploadedBy == userID
	}

	return false
}

// GetFormattedSize 获取格式化的文件大小
func (f *File) GetFormattedSize() string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	size := float64(f.Size)
	switch {
	case f.Size >= TB:
		return formatSize(size/TB, "TB")
	case f.Size >= GB:
		return formatSize(size/GB, "GB")
	case f.Size >= MB:
		return formatSize(size/MB, "MB")
	case f.Size >= KB:
		return formatSize(size/KB, "KB")
	default:
		return formatSize(size, "B")
	}
}

func formatSize(size float64, unit string) string {
	return fmt.Sprintf("%.2f %s", size, unit)
}
