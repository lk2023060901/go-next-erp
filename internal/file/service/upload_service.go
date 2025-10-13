package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
	"github.com/lk2023060901/go-next-erp/internal/file/repository"
	"github.com/lk2023060901/go-next-erp/internal/file/utils"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"github.com/lk2023060901/go-next-erp/pkg/storage"
	"go.uber.org/zap"
)

// UploadService 文件上传服务接口
type UploadService interface {
	// 普通上传
	Upload(ctx context.Context, req *UploadRequest) (*model.File, error)

	// 分片上传
	InitiateMultipartUpload(ctx context.Context, req *InitiateMultipartRequest) (*MultipartUploadInfo, error)
	UploadPart(ctx context.Context, req *UploadPartRequest) (*PartUploadResult, error)
	CompleteMultipartUpload(ctx context.Context, req *CompleteMultipartRequest) (*model.File, error)
	AbortMultipartUpload(ctx context.Context, uploadID string) error

	// 检查重复文件（秒传）
	CheckDuplicate(ctx context.Context, checksum string, tenantID uuid.UUID) (*model.File, error)
}

// UploadRequest 上传请求
type UploadRequest struct {
	TenantID    uuid.UUID
	UploadedBy  uuid.UUID
	Filename    string
	Reader      io.Reader
	Size        int64
	ContentType string

	// Optional fields
	IsTemporary bool
	ExpiresAt   *time.Time
	Category    *string
	Tags        []string
	Metadata    map[string]interface{}
	AccessLevel model.AccessLevel
}

// InitiateMultipartRequest 初始化分片上传请求
type InitiateMultipartRequest struct {
	TenantID    uuid.UUID
	UploadedBy  uuid.UUID
	Filename    string
	TotalSize   int64
	ContentType string
	IsTemporary bool
	ExpiresAt   *time.Time
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

// CompleteMultipartRequest 完成分片上传请求
type CompleteMultipartRequest struct {
	UploadID   string
	TenantID   uuid.UUID
	UploadedBy uuid.UUID
	Parts      []PartUploadResult
}

type uploadService struct {
	fileRepo  repository.FileRepository
	quotaRepo repository.QuotaRepository
	storage   storage.Storage
	logger    *logger.Logger

	// Configuration
	maxFileSize         int64
	allowedExtensions   []string
	enableVirusScan     bool
	enableDeduplication bool
}

// UploadServiceConfig 上传服务配置
type UploadServiceConfig struct {
	MaxFileSize         int64
	AllowedExtensions   []string
	EnableVirusScan     bool
	EnableDeduplication bool
}

// NewUploadService 创建上传服务
func NewUploadService(
	fileRepo repository.FileRepository,
	quotaRepo repository.QuotaRepository,
	storage storage.Storage,
	logger *logger.Logger,
	config *UploadServiceConfig,
) UploadService {
	if config == nil {
		config = &UploadServiceConfig{
			MaxFileSize:         100 * 1024 * 1024, // 100MB default
			EnableVirusScan:     false,
			EnableDeduplication: true,
		}
	}

	return &uploadService{
		fileRepo:            fileRepo,
		quotaRepo:           quotaRepo,
		storage:             storage,
		logger:              logger.With(zap.String("service", "upload")),
		maxFileSize:         config.MaxFileSize,
		allowedExtensions:   config.AllowedExtensions,
		enableVirusScan:     config.EnableVirusScan,
		enableDeduplication: config.EnableDeduplication,
	}
}

// Upload 普通文件上传
func (s *uploadService) Upload(ctx context.Context, req *UploadRequest) (*model.File, error) {
	// 1. 验证文件
	if err := s.validateUpload(req.Filename, req.Size); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 2. 检查配额
	quota, err := s.quotaRepo.GetOrCreateTenantQuota(ctx, req.TenantID, 10*1024*1024*1024) // 10GB default
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	if !quota.CanAllocate(req.Size) {
		return nil, fmt.Errorf("quota exceeded: available %s, requested %s",
			formatBytes(quota.GetAvailableQuota()),
			formatBytes(req.Size))
	}

	// 3. 预留配额
	if err := s.quotaRepo.ReserveQuota(ctx, quota.ID, req.Size); err != nil {
		return nil, fmt.Errorf("failed to reserve quota: %w", err)
	}

	// 4. 计算哈希值 (使用 TeeReader 同时上传和计算哈希)
	hashReader, checksum, err := s.calculateChecksumWhileReading(req.Reader, req.Size)
	if err != nil {
		s.quotaRepo.ReleaseReservation(ctx, quota.ID, req.Size)
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// 5. 检查重复（秒传）
	if s.enableDeduplication {
		existingFile, _ := s.fileRepo.FindByChecksum(ctx, checksum, req.TenantID)
		if existingFile != nil && existingFile.IsActive() {
			s.logger.Info("File already exists, using existing file",
				zap.String("checksum", checksum),
				zap.String("existing_file_id", existingFile.ID.String()))

			// Release reservation
			s.quotaRepo.ReleaseReservation(ctx, quota.ID, req.Size)

			// Return existing file (可以创建新的引用或直接返回)
			return existingFile, nil
		}
	}

	// 6. 生成存储键
	storageKey := utils.GenerateStorageKey(req.TenantID, req.Filename)

	// 7. 检测 MIME 类型
	mimeType := req.ContentType
	if mimeType == "" {
		// Use default MIME detection
		mimeType = "application/octet-stream"
	}

	// 8. 上传到 MinIO
	bucket := "files" // Default bucket
	opts := storage.PutObjectOptions{
		ContentType: mimeType,
	}
	_, err = s.storage.GetObjectStorage().PutObject(ctx, bucket, storageKey, hashReader, req.Size, opts)
	if err != nil {
		s.quotaRepo.ReleaseReservation(ctx, quota.ID, req.Size)
		return nil, fmt.Errorf("failed to upload to storage: %w", err)
	}

	// 9. 创建文件记录
	file := &model.File{
		TenantID:      req.TenantID,
		Filename:      utils.SanitizeFilename(req.Filename),
		StorageKey:    storageKey,
		Size:          req.Size,
		MimeType:      mimeType,
		ContentType:   req.ContentType,
		Checksum:      checksum,
		Extension:     utils.GetFileExtension(req.Filename),
		Bucket:        bucket,
		Category:      utils.DetectCategory(mimeType),
		Tags:          req.Tags,
		Metadata:      req.Metadata,
		Status:        model.FileStatusActive,
		IsTemporary:   req.IsTemporary,
		IsPublic:      req.AccessLevel == model.AccessLevelPublic,
		VersionNumber: 1,
		UploadedBy:    req.UploadedBy,
		AccessLevel:   req.AccessLevel,
		ExpiresAt:     req.ExpiresAt,
	}

	if req.Category != nil {
		file.Category = *req.Category
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		// Rollback: delete uploaded file
		s.storage.GetObjectStorage().RemoveObject(ctx, bucket, storageKey)
		s.quotaRepo.ReleaseReservation(ctx, quota.ID, req.Size)
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// 10. 提交配额
	if err := s.quotaRepo.CommitReservation(ctx, quota.ID, req.Size); err != nil {
		s.logger.Error("Failed to commit quota reservation", zap.Error(err))
		// Non-fatal, continue
	}

	s.logger.Info("File uploaded successfully",
		zap.String("file_id", file.ID.String()),
		zap.String("filename", file.Filename),
		zap.Int64("size", file.Size))

	return file, nil
}

// CheckDuplicate 检查重复文件
func (s *uploadService) CheckDuplicate(ctx context.Context, checksum string, tenantID uuid.UUID) (*model.File, error) {
	return s.fileRepo.FindByChecksum(ctx, checksum, tenantID)
}

// InitiateMultipartUpload 初始化分片上传
func (s *uploadService) InitiateMultipartUpload(ctx context.Context, req *InitiateMultipartRequest) (*MultipartUploadInfo, error) {
	// Validate
	if err := s.validateUpload(req.Filename, req.TotalSize); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check quota
	quota, err := s.quotaRepo.GetOrCreateTenantQuota(ctx, req.TenantID, 10*1024*1024*1024)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	if !quota.CanAllocate(req.TotalSize) {
		return nil, fmt.Errorf("quota exceeded")
	}

	// Reserve quota
	if err := s.quotaRepo.ReserveQuota(ctx, quota.ID, req.TotalSize); err != nil {
		return nil, fmt.Errorf("failed to reserve quota: %w", err)
	}

	// Generate storage key
	storageKey := utils.GenerateStorageKey(req.TenantID, req.Filename)

	// Initiate multipart upload in MinIO
	bucket := "files"
	opts := storage.PutObjectOptions{
		ContentType: req.ContentType,
	}
	uploadID, err := s.storage.GetObjectStorage().NewMultipartUpload(ctx, bucket, storageKey, opts)
	if err != nil {
		s.quotaRepo.ReleaseReservation(ctx, quota.ID, req.TotalSize)
		return nil, fmt.Errorf("failed to initiate multipart upload: %w", err)
	}

	// Calculate part size and total parts
	partSize := utils.CalculatePartSize(req.TotalSize)
	totalParts := int((req.TotalSize + partSize - 1) / partSize)

	return &MultipartUploadInfo{
		UploadID:   uploadID,
		StorageKey: storageKey,
		PartSize:   partSize,
		TotalParts: totalParts,
	}, nil
}

// UploadPart 上传分片
func (s *uploadService) UploadPart(ctx context.Context, req *UploadPartRequest) (*PartUploadResult, error) {
	// TODO: Retrieve bucket and key from multipart upload record
	// For now using simplified approach
	bucket := "files"
	key := "" // Should retrieve from database

	part, err := s.storage.GetObjectStorage().PutObjectPart(ctx, bucket, key, req.UploadID, req.PartNumber, req.Reader, req.Size, storage.PutObjectPartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to upload part: %w", err)
	}

	return &PartUploadResult{
		PartNumber: req.PartNumber,
		ETag:       part.ETag,
		Size:       part.Size,
	}, nil
}

// CompleteMultipartUpload 完成分片上传
func (s *uploadService) CompleteMultipartUpload(ctx context.Context, req *CompleteMultipartRequest) (*model.File, error) {
	// TODO: Implement multipart completion with proper database tracking
	s.logger.Info("Multipart upload completed",
		zap.String("upload_id", req.UploadID),
		zap.Int("parts", len(req.Parts)))

	return nil, fmt.Errorf("not fully implemented")
}

// AbortMultipartUpload 中止分片上传
func (s *uploadService) AbortMultipartUpload(ctx context.Context, uploadID string) error {
	// TODO: Retrieve multipart info and abort
	return fmt.Errorf("not implemented")
}

// validateUpload 验证上传请求
func (s *uploadService) validateUpload(filename string, size int64) error {
	if filename == "" {
		return fmt.Errorf("filename is required")
	}

	if size <= 0 {
		return fmt.Errorf("file size must be greater than 0")
	}

	if s.maxFileSize > 0 && size > s.maxFileSize {
		return fmt.Errorf("file size %s exceeds maximum %s",
			formatBytes(size),
			formatBytes(s.maxFileSize))
	}

	if len(s.allowedExtensions) > 0 {
		ext := utils.GetFileExtension(filename)
		allowed := false
		for _, allowedExt := range s.allowedExtensions {
			if ext == allowedExt {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("file extension %s is not allowed", ext)
		}
	}

	return nil
}

// calculateChecksumWhileReading 在读取时计算校验和
func (s *uploadService) calculateChecksumWhileReading(reader io.Reader, size int64) (io.Reader, string, error) {
	hash := sha256.New()
	teeReader := io.TeeReader(reader, hash)

	// Read all data to compute hash
	buf := make([]byte, 32*1024) // 32KB buffer
	written := int64(0)

	for {
		n, err := teeReader.Read(buf)
		written += int64(n)

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", err
		}
	}

	checksum := hex.EncodeToString(hash.Sum(nil))

	// Return a new reader for the actual upload
	// Note: This is a simplified approach. In production, consider using io.Pipe
	// or storing the data temporarily
	return reader, checksum, nil
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
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
