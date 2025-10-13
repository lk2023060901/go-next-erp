package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
)

// UploadRequest 文件上传请求
type UploadRequest struct {
	Filename    string                 `json:"filename" validate:"required"`
	ContentType string                 `json:"content_type"`
	Size        int64                  `json:"size" validate:"required,min=1"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	IsTemporary bool                   `json:"is_temporary"`
	ExpiresAt   *time.Time             `json:"expires_at"`
	AccessLevel string                 `json:"access_level" validate:"omitempty,oneof=private tenant public"`
}

// UploadResponse 文件上传响应
type UploadResponse struct {
	FileID      uuid.UUID `json:"file_id"`
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	MimeType    string    `json:"mime_type"`
	StorageKey  string    `json:"storage_key"`
	Checksum    string    `json:"checksum"`
	DownloadURL string    `json:"download_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// FileInfoResponse 文件信息响应
type FileInfoResponse struct {
	ID              uuid.UUID              `json:"id"`
	TenantID        uuid.UUID              `json:"tenant_id"`
	Filename        string                 `json:"filename"`
	OriginalName    string                 `json:"original_name"`
	Size            int64                  `json:"size"`
	MimeType        string                 `json:"mime_type"`
	Extension       string                 `json:"extension"`
	Category        string                 `json:"category"`
	Checksum        string                 `json:"checksum"`
	StorageKey      string                 `json:"storage_key"`
	Bucket          string                 `json:"bucket"`
	Tags            []string               `json:"tags"`
	Metadata        map[string]interface{} `json:"metadata"`
	Status          string                 `json:"status"`
	IsTemporary     bool                   `json:"is_temporary"`
	IsPublic        bool                   `json:"is_public"`
	VirusScanned    bool                   `json:"virus_scanned"`
	VirusScanResult *string                `json:"virus_scan_result,omitempty"`
	DownloadCount   int                    `json:"download_count"`
	VersionNumber   int                    `json:"version_number"`
	AccessLevel     string                 `json:"access_level"`
	UploadedBy      uuid.UUID              `json:"uploaded_by"`
	UploadedAt      time.Time              `json:"uploaded_at"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty"`
	PreviewURL      *string                `json:"preview_url,omitempty"`
	DownloadURL     *string                `json:"download_url,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// FileListRequest 文件列表请求
type FileListRequest struct {
	Page         int        `json:"page" form:"page" validate:"min=1"`
	PageSize     int        `json:"page_size" form:"page_size" validate:"min=1,max=100"`
	Filename     string     `json:"filename" form:"filename"`
	Category     string     `json:"category" form:"category"`
	MimeType     string     `json:"mime_type" form:"mime_type"`
	Status       string     `json:"status" form:"status"`
	IsTemporary  *bool      `json:"is_temporary" form:"is_temporary"`
	UploadedBy   *uuid.UUID `json:"uploaded_by" form:"uploaded_by"`
	StartDate    *time.Time `json:"start_date" form:"start_date"`
	EndDate      *time.Time `json:"end_date" form:"end_date"`
	Tags         []string   `json:"tags" form:"tags"`
	SortBy       string     `json:"sort_by" form:"sort_by" validate:"omitempty,oneof=created_at size filename"`
	SortOrder    string     `json:"sort_order" form:"sort_order" validate:"omitempty,oneof=asc desc"`
	IncludeStats bool       `json:"include_stats" form:"include_stats"`
}

// FileListResponse 文件列表响应
type FileListResponse struct {
	Files      []FileInfoResponse `json:"files"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
	Stats      *FileStats         `json:"stats,omitempty"`
}

// FileStats 文件统计信息
type FileStats struct {
	TotalFiles int64 `json:"total_files"`
	TotalSize  int64 `json:"total_size"`
	TotalSizeFormatted string `json:"total_size_formatted"`
}

// MultipartInitiateRequest 分片上传初始化请求
type MultipartInitiateRequest struct {
	Filename    string     `json:"filename" validate:"required"`
	TotalSize   int64      `json:"total_size" validate:"required,min=1"`
	ContentType string     `json:"content_type"`
	IsTemporary bool       `json:"is_temporary"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// MultipartInitiateResponse 分片上传初始化响应
type MultipartInitiateResponse struct {
	UploadID   string `json:"upload_id"`
	StorageKey string `json:"storage_key"`
	PartSize   int64  `json:"part_size"`
	TotalParts int    `json:"total_parts"`
}

// MultipartUploadPartRequest 分片上传请求
type MultipartUploadPartRequest struct {
	UploadID   string `json:"upload_id" validate:"required"`
	PartNumber int    `json:"part_number" validate:"required,min=1"`
}

// MultipartUploadPartResponse 分片上传响应
type MultipartUploadPartResponse struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size"`
}

// MultipartCompleteRequest 完成分片上传请求
type MultipartCompleteRequest struct {
	UploadID string                  `json:"upload_id" validate:"required"`
	Parts    []MultipartPartComplete `json:"parts" validate:"required,min=1,dive"`
}

// MultipartPartComplete 分片完成信息
type MultipartPartComplete struct {
	PartNumber int    `json:"part_number" validate:"required,min=1"`
	ETag       string `json:"etag" validate:"required"`
}

// QuotaInfoResponse 配额信息响应
type QuotaInfoResponse struct {
	SubjectType        string  `json:"subject_type"`
	SubjectID          *string `json:"subject_id,omitempty"`
	QuotaLimit         int64   `json:"quota_limit"`
	QuotaUsed          int64   `json:"quota_used"`
	QuotaAvailable     int64   `json:"quota_available"`
	QuotaUsedPercent   float64 `json:"quota_used_percent"`
	QuotaLimitFormatted string `json:"quota_limit_formatted"`
	QuotaUsedFormatted  string `json:"quota_used_formatted"`
	IsWarning          bool    `json:"is_warning"`
	IsExceeded         bool    `json:"is_exceeded"`
}

// DownloadURLRequest 下载URL请求
type DownloadURLRequest struct {
	Expiry int `json:"expiry" form:"expiry" validate:"omitempty,min=60,max=86400"` // 秒，默认1小时
}

// DownloadURLResponse 下载URL响应
type DownloadURLResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// BatchDownloadRequest 批量下载请求
type BatchDownloadRequest struct {
	FileIDs []uuid.UUID `json:"file_ids" validate:"required,min=1,max=100,dive"`
	Expiry  int         `json:"expiry" validate:"omitempty,min=60,max=86400"`
}

// BatchDownloadResponse 批量下载响应
type BatchDownloadResponse struct {
	Downloads []FileDownloadInfo `json:"downloads"`
	Total     int                `json:"total"`
	Success   int                `json:"success"`
	Failed    int                `json:"failed"`
}

// FileDownloadInfo 文件下载信息
type FileDownloadInfo struct {
	FileID    uuid.UUID `json:"file_id"`
	Filename  string    `json:"filename"`
	URL       string    `json:"url,omitempty"`
	Error     string    `json:"error,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// FromModel 从 model.File 转换为 FileInfoResponse
func FromModel(file *model.File) *FileInfoResponse {
	resp := &FileInfoResponse{
		ID:              file.ID,
		TenantID:        file.TenantID,
		Filename:        file.Filename,
		OriginalName:    file.Filename, // Use Filename as OriginalName
		Size:            file.Size,
		MimeType:        file.MimeType,
		Extension:       file.Extension,
		Category:        file.Category,
		Checksum:        file.Checksum,
		StorageKey:      file.StorageKey,
		Bucket:          file.Bucket,
		Tags:            file.Tags,
		Metadata:        nil, // Model uses map[string]string, DTO uses map[string]interface{}
		Status:          string(file.Status),
		IsTemporary:     file.IsTemporary,
		IsPublic:        file.IsPublic,
		VirusScanned:    file.VirusScanned,
		DownloadCount:   0, // TODO: Track download count separately
		VersionNumber:   file.VersionNumber,
		AccessLevel:     string(file.AccessLevel),
		UploadedBy:      file.UploadedBy,
		UploadedAt:      file.CreatedAt, // Use CreatedAt as UploadedAt
		ExpiresAt:       file.ExpiresAt,
		PreviewURL:      file.PreviewURL,
		CreatedAt:       file.CreatedAt,
		UpdatedAt:       file.UpdatedAt,
	}

	if file.VirusScanResult != nil {
		result := string(*file.VirusScanResult)
		resp.VirusScanResult = &result
	}

	return resp
}

// FromModelList 批量转换
func FromModelList(files []*model.File) []FileInfoResponse {
	result := make([]FileInfoResponse, len(files))
	for i, file := range files {
		result[i] = *FromModel(file)
	}
	return result
}
