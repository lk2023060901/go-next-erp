package handler

import (
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
)

// UploadFileRequest 上传文件请求
type UploadFileRequest struct {
	TenantID          uuid.UUID
	UploadedBy        uuid.UUID
	Filename          string
	Reader            io.Reader
	Size              int64
	ContentType       string
	IsTemporary       bool
	ExpiresAt         *time.Time
	Category          *string
	Tags              []string
	Metadata          map[string]interface{}
	AccessLevel       model.AccessLevel
	GenerateThumbnail bool // 是否自动生成缩略图
	Compress          bool // 是否自动压缩
}

// UploadFileResponse 上传文件响应
type UploadFileResponse struct {
	File *model.File
}

// DownloadFileRequest 下载文件请求
type DownloadFileRequest struct {
	FileID         uuid.UUID
	UserID         uuid.UUID
	TenantID       uuid.UUID
	Expiry         time.Duration
	IPAddress      string
	UserAgent      string
	RecordDownload bool // 是否记录下载
}

// DownloadFileResponse 下载文件响应
type DownloadFileResponse struct {
	URL       string
	ExpiresAt time.Time
}

// DeleteFileRequest 删除文件请求
type DeleteFileRequest struct {
	FileID    uuid.UUID
	UserID    uuid.UUID
	Permanent bool // 是否永久删除
}

// MoveFileRequest 移动文件请求
type MoveFileRequest struct {
	FileID      uuid.UUID
	NewCategory string
	UserID      uuid.UUID
}

// RenameFileRequest 重命名文件请求
type RenameFileRequest struct {
	FileID      uuid.UUID
	NewFilename string
	UserID      uuid.UUID
}

// CopyFileRequest 复制文件请求
type CopyFileRequest struct {
	FileID uuid.UUID
	UserID uuid.UUID
}

// BatchDeleteRequest 批量删除请求
type BatchDeleteRequest struct {
	FileIDs   []uuid.UUID
	UserID    uuid.UUID
	Permanent bool
}

// BatchOperationResult 批量操作结果
type BatchOperationResult struct {
	SuccessCount int
	FailedCount  int
	Errors       map[uuid.UUID]error
}

// AttachFileRequest 关联文件请求
type AttachFileRequest struct {
	FileID     uuid.UUID
	EntityType model.EntityType
	EntityID   uuid.UUID
	FieldName  string
	AttachedBy uuid.UUID
}

// CreateVersionRequest 创建版本请求
type CreateVersionRequest struct {
	FileID   uuid.UUID
	UserID   uuid.UUID
	TenantID uuid.UUID
	Comment  string
	Reader   io.Reader
	Checksum string
}

// RevertVersionRequest 回滚版本请求
type RevertVersionRequest struct {
	FileID        uuid.UUID
	VersionNumber int
	UserID        uuid.UUID
	Comment       string
}

// CompressFileRequest 压缩文件请求
type CompressFileRequest struct {
	FileID     uuid.UUID
	Quality    int // 1-100
	MaxWidth   int
	MaxHeight  int
	KeepAspect bool
	InPlace    bool // 是否就地压缩
}

// InitiateMultipartUploadRequest 初始化分片上传请求
type InitiateMultipartUploadRequest struct {
	TenantID    uuid.UUID
	UserID      uuid.UUID
	Filename    string
	TotalSize   int64
	PartSize    int64
	ContentType string
}

// MultipartUploadInfo 分片上传信息
type MultipartUploadInfo struct {
	UploadID   string
	StorageKey string
	PartSize   int64
	TotalParts int
}

// UploadPartRequest 上传分片请求
type UploadPartRequest struct {
	UploadID   string
	PartNumber int
	Reader     io.Reader
	Size       int64
}

// PartUploadResult 分片上传结果
type PartUploadResult struct {
	PartNumber int
	ETag       string
	Size       int64
}

// CompleteMultipartUploadRequest 完成分片上传请求
type CompleteMultipartUploadRequest struct {
	UploadID string
	Parts    []PartUploadResult
}
