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

// MultipartUploadService 分片上传服务接口
type MultipartUploadService interface {
	// 初始化分片上传
	InitiateUpload(ctx context.Context, req *InitiateMultipartUploadRequest) (*MultipartUploadResponse, error)

	// 上传分片
	UploadPart(ctx context.Context, req *MultipartUploadPartRequest) (*UploadPartResponse, error)

	// 完成分片上传
	CompleteUpload(ctx context.Context, req *CompleteMultipartUploadRequest) (*model.File, error)

	// 中止分片上传
	AbortUpload(ctx context.Context, uploadID string, tenantID uuid.UUID) error

	// 获取上传进度
	GetUploadProgress(ctx context.Context, uploadID string) (*UploadProgressResponse, error)

	// 列出已上传的分片
	ListUploadedParts(ctx context.Context, uploadID string) ([]UploadedPartInfo, error)

	// 续传（获取剩余分片）
	GetRemainingParts(ctx context.Context, uploadID string) ([]int, error)

	// 清理过期的上传
	CleanExpiredUploads(ctx context.Context) (int64, error)
}

// InitiateMultipartUploadRequest 初始化分片上传请求
type InitiateMultipartUploadRequest struct {
	TenantID    uuid.UUID
	UploadedBy  uuid.UUID
	Filename    string
	TotalSize   int64
	ContentType string
	Checksum    string // 可选：整个文件的 checksum
	IsTemporary bool
	ExpiresAt   *time.Time
	Category    *string
	Tags        []string
	Metadata    map[string]interface{}
	AccessLevel model.AccessLevel
}

// MultipartUploadResponse 分片上传响应
type MultipartUploadResponse struct {
	UploadID   string    `json:"upload_id"`
	RecordID   uuid.UUID `json:"record_id"`
	StorageKey string    `json:"storage_key"`
	PartSize   int64     `json:"part_size"`
	TotalParts int       `json:"total_parts"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// MultipartUploadPartRequest 上传分片请求
type MultipartUploadPartRequest struct {
	UploadID   string
	PartNumber int
	Reader     io.Reader
	Size       int64
	Checksum   string // 可选：分片的 checksum
}

// UploadPartResponse 上传分片响应
type UploadPartResponse struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
	Size       int64  `json:"size"`
	Uploaded   bool   `json:"uploaded"`
}

// CompleteMultipartUploadRequest 完成分片上传请求
type CompleteMultipartUploadRequest struct {
	UploadID string
	TenantID uuid.UUID
	Parts    []CompletedPartInfo // 可选：如果为空则自动从数据库获取
}

// CompletedPartInfo 已完成的分片信息
type CompletedPartInfo struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

// UploadProgressResponse 上传进度响应
type UploadProgressResponse struct {
	UploadID      string    `json:"upload_id"`
	Filename      string    `json:"filename"`
	TotalSize     int64     `json:"total_size"`
	TotalParts    int       `json:"total_parts"`
	UploadedParts []int     `json:"uploaded_parts"`
	UploadedCount int       `json:"uploaded_count"`
	Progress      float64   `json:"progress"`
	Status        string    `json:"status"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UploadedPartInfo 已上传分片信息
type UploadedPartInfo struct {
	PartNumber   int       `json:"part_number"`
	Size         int64     `json:"size"`
	ETag         string    `json:"etag"`
	LastModified time.Time `json:"last_modified"`
}

type multipartUploadService struct {
	fileRepo      repository.FileRepository
	quotaRepo     repository.QuotaRepository
	multipartRepo repository.MultipartUploadRepository
	storageClient storage.Storage
	logger        *logger.Logger

	// Configuration
	partSize          int64         // 默认分片大小
	maxPartSize       int64         // 最大分片大小
	minPartSize       int64         // 最小分片大小
	uploadExpiration  time.Duration // 上传过期时间
	maxPartsPerUpload int           // 单次上传最大分片数
}

// MultipartUploadServiceConfig 分片上传服务配置
type MultipartUploadServiceConfig struct {
	PartSize          int64         // 默认 5MB
	MaxPartSize       int64         // 默认 5GB
	MinPartSize       int64         // 默认 5MB
	UploadExpiration  time.Duration // 默认 7天
	MaxPartsPerUpload int           // 默认 10000
}

// NewMultipartUploadService 创建分片上传服务
func NewMultipartUploadService(
	fileRepo repository.FileRepository,
	quotaRepo repository.QuotaRepository,
	multipartRepo repository.MultipartUploadRepository,
	storageClient storage.Storage,
	logger *logger.Logger,
	config *MultipartUploadServiceConfig,
) MultipartUploadService {
	if config == nil {
		config = &MultipartUploadServiceConfig{
			PartSize:          5 * 1024 * 1024,        // 5MB
			MaxPartSize:       5 * 1024 * 1024 * 1024, // 5GB
			MinPartSize:       5 * 1024 * 1024,        // 5MB
			UploadExpiration:  7 * 24 * time.Hour,     // 7 days
			MaxPartsPerUpload: 10000,
		}
	}

	return &multipartUploadService{
		fileRepo:          fileRepo,
		quotaRepo:         quotaRepo,
		multipartRepo:     multipartRepo,
		storageClient:     storageClient,
		logger:            logger.With(zap.String("service", "multipart_upload")),
		partSize:          config.PartSize,
		maxPartSize:       config.MaxPartSize,
		minPartSize:       config.MinPartSize,
		uploadExpiration:  config.UploadExpiration,
		maxPartsPerUpload: config.MaxPartsPerUpload,
	}
}

// InitiateUpload 初始化分片上传
func (s *multipartUploadService) InitiateUpload(ctx context.Context, req *InitiateMultipartUploadRequest) (*MultipartUploadResponse, error) {
	// 1. 验证请求
	if req.Filename == "" {
		return nil, fmt.Errorf("filename is required")
	}
	if req.TotalSize <= 0 {
		return nil, fmt.Errorf("total size must be greater than 0")
	}

	// 2. 检查配额
	quota, err := s.quotaRepo.GetOrCreateTenantQuota(ctx, req.TenantID, 10*1024*1024*1024) // 10GB default
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	if !quota.CanAllocate(req.TotalSize) {
		return nil, fmt.Errorf("quota exceeded: available %s, requested %s",
			formatBytes(quota.GetAvailableQuota()),
			formatBytes(req.TotalSize))
	}

	// 3. 预留配额
	if err := s.quotaRepo.ReserveQuota(ctx, quota.ID, req.TotalSize); err != nil {
		return nil, fmt.Errorf("failed to reserve quota: %w", err)
	}

	// 4. 计算分片大小和数量
	partSize := s.calculatePartSize(req.TotalSize)
	totalParts := int((req.TotalSize + partSize - 1) / partSize)

	if totalParts > s.maxPartsPerUpload {
		s.quotaRepo.ReleaseReservation(ctx, quota.ID, req.TotalSize)
		return nil, fmt.Errorf("too many parts: %d (max: %d)", totalParts, s.maxPartsPerUpload)
	}

	// 5. 生成存储键
	storageKey := utils.GenerateStorageKey(req.TenantID, req.Filename)
	bucket := "files" // 默认 bucket

	// 6. 在对象存储中初始化分片上传
	objectStorage := s.storageClient.GetObjectStorage()
	uploadID, err := objectStorage.NewMultipartUpload(ctx, bucket, storageKey, storage.PutObjectOptions{
		ContentType:  req.ContentType,
		UserMetadata: map[string]string{"original_filename": req.Filename},
	})
	if err != nil {
		s.quotaRepo.ReleaseReservation(ctx, quota.ID, req.TotalSize)
		return nil, fmt.Errorf("failed to initiate multipart upload: %w", err)
	}

	// 7. 创建数据库记录
	expiresAt := time.Now().Add(s.uploadExpiration)
	upload := &model.MultipartUpload{
		TenantID:      req.TenantID,
		UploadID:      uploadID,
		Filename:      utils.SanitizeFilename(req.Filename),
		StorageKey:    storageKey,
		TotalSize:     &req.TotalSize,
		PartSize:      partSize,
		UploadedParts: []int{},
		TotalParts:    &totalParts,
		MimeType:      req.ContentType,
		Metadata:      req.Metadata,
		Status:        model.UploadStatusInProgress,
		CreatedBy:     req.UploadedBy,
		ExpiresAt:     expiresAt,
	}

	if err := s.multipartRepo.Create(ctx, upload); err != nil {
		// 回滚：中止对象存储中的上传
		objectStorage.AbortMultipartUpload(ctx, bucket, storageKey, uploadID)
		s.quotaRepo.ReleaseReservation(ctx, quota.ID, req.TotalSize)
		return nil, fmt.Errorf("failed to create upload record: %w", err)
	}

	s.logger.Info("Multipart upload initiated",
		zap.String("upload_id", uploadID),
		zap.String("record_id", upload.ID.String()),
		zap.String("filename", req.Filename),
		zap.Int64("total_size", req.TotalSize),
		zap.Int("total_parts", totalParts))

	return &MultipartUploadResponse{
		UploadID:   uploadID,
		RecordID:   upload.ID,
		StorageKey: storageKey,
		PartSize:   partSize,
		TotalParts: totalParts,
		ExpiresAt:  expiresAt,
	}, nil
}

// UploadPart 上传分片
func (s *multipartUploadService) UploadPart(ctx context.Context, req *MultipartUploadPartRequest) (*UploadPartResponse, error) {
	// 1. 查找上传记录
	upload, err := s.multipartRepo.FindByUploadID(ctx, req.UploadID)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %w", err)
	}

	// 2. 验证状态
	if upload.Status != model.UploadStatusInProgress {
		return nil, fmt.Errorf("upload is not in progress (status: %s)", upload.Status)
	}

	// 3. 检查是否过期
	if upload.IsExpired() {
		return nil, fmt.Errorf("upload has expired")
	}

	// 4. 验证分片编号
	if upload.TotalParts != nil && req.PartNumber > *upload.TotalParts {
		return nil, fmt.Errorf("part number %d exceeds total parts %d", req.PartNumber, *upload.TotalParts)
	}

	// 5. 检查分片是否已上传
	if upload.IsPartCompleted(req.PartNumber) {
		s.logger.Info("Part already uploaded",
			zap.String("upload_id", req.UploadID),
			zap.Int("part_number", req.PartNumber))
		// 返回已上传状态
		return &UploadPartResponse{
			PartNumber: req.PartNumber,
			Uploaded:   true,
		}, nil
	}

	// 6. 上传分片到对象存储
	bucket := "files"
	objectStorage := s.storageClient.GetObjectStorage()

	part, err := objectStorage.PutObjectPart(ctx, bucket, upload.StorageKey, req.UploadID, req.PartNumber, req.Reader, req.Size, storage.PutObjectPartOptions{})
	if err != nil {
		s.logger.Error("Failed to upload part",
			zap.String("upload_id", req.UploadID),
			zap.Int("part_number", req.PartNumber),
			zap.Error(err))
		return nil, fmt.Errorf("failed to upload part: %w", err)
	}

	// 7. 更新数据库记录
	if err := s.multipartRepo.AddCompletedPart(ctx, upload.ID, req.PartNumber); err != nil {
		s.logger.Error("Failed to update completed part",
			zap.String("upload_id", req.UploadID),
			zap.Int("part_number", req.PartNumber),
			zap.Error(err))
		// 非致命错误，继续
	}

	s.logger.Info("Part uploaded successfully",
		zap.String("upload_id", req.UploadID),
		zap.Int("part_number", req.PartNumber),
		zap.Int64("size", part.Size))

	return &UploadPartResponse{
		PartNumber: req.PartNumber,
		ETag:       part.ETag,
		Size:       part.Size,
		Uploaded:   true,
	}, nil
}

// CompleteUpload 完成分片上传
func (s *multipartUploadService) CompleteUpload(ctx context.Context, req *CompleteMultipartUploadRequest) (*model.File, error) {
	// 1. 查找上传记录
	upload, err := s.multipartRepo.FindByUploadID(ctx, req.UploadID)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %w", err)
	}

	// 2. 验证状态
	if upload.Status != model.UploadStatusInProgress {
		return nil, fmt.Errorf("upload is not in progress (status: %s)", upload.Status)
	}

	// 3. 获取已上传的分片列表
	bucket := "files"
	objectStorage := s.storageClient.GetObjectStorage()

	uploadedParts, err := objectStorage.ListObjectParts(ctx, bucket, upload.StorageKey, req.UploadID, storage.ListObjectPartsOptions{
		MaxParts: s.maxPartsPerUpload,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list parts: %w", err)
	}

	// 4. 构建完成分片列表
	var parts []storage.CompletePart
	if len(req.Parts) > 0 {
		// 使用请求中的分片列表
		for _, p := range req.Parts {
			parts = append(parts, storage.CompletePart{
				PartNumber: p.PartNumber,
				ETag:       p.ETag,
			})
		}
	} else {
		// 使用从对象存储获取的分片列表
		for _, p := range uploadedParts {
			parts = append(parts, storage.CompletePart{
				PartNumber: p.PartNumber,
				ETag:       p.ETag,
			})
		}
	}

	// 5. 验证所有分片都已上传
	if upload.TotalParts != nil && len(parts) != *upload.TotalParts {
		return nil, fmt.Errorf("incomplete upload: expected %d parts, got %d", *upload.TotalParts, len(parts))
	}

	// 6. 完成对象存储中的分片上传
	result, err := objectStorage.CompleteMultipartUpload(ctx, bucket, upload.StorageKey, req.UploadID, parts)
	if err != nil {
		s.logger.Error("Failed to complete multipart upload",
			zap.String("upload_id", req.UploadID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	// 7. 计算文件 checksum（如果需要）
	// 注意：对于分片上传，完整的 checksum 需要在客户端计算或在这里重新下载计算
	checksum := result.ETag // 临时使用 ETag

	// 8. 创建文件记录
	totalSize := int64(0)
	if upload.TotalSize != nil {
		totalSize = *upload.TotalSize
	}

	file := &model.File{
		TenantID:      upload.TenantID,
		Filename:      upload.Filename,
		StorageKey:    upload.StorageKey,
		Size:          totalSize,
		MimeType:      upload.MimeType,
		ContentType:   upload.MimeType,
		Checksum:      checksum,
		Extension:     utils.GetFileExtension(upload.Filename),
		Bucket:        bucket,
		Category:      utils.DetectCategory(upload.MimeType),
		Metadata:      upload.Metadata,
		Status:        model.FileStatusActive,
		IsTemporary:   false,
		IsPublic:      false,
		VersionNumber: 1,
		UploadedBy:    upload.CreatedBy,
		AccessLevel:   model.AccessLevelPrivate,
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		s.logger.Error("Failed to create file record",
			zap.String("upload_id", req.UploadID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// 9. 标记上传为已完成
	if err := s.multipartRepo.MarkAsCompleted(ctx, upload.ID); err != nil {
		s.logger.Error("Failed to mark upload as completed",
			zap.String("upload_id", req.UploadID),
			zap.Error(err))
		// 非致命错误
	}

	// 10. 提交配额
	quota, _ := s.quotaRepo.GetOrCreateTenantQuota(ctx, req.TenantID, 10*1024*1024*1024)
	if quota != nil {
		if err := s.quotaRepo.CommitReservation(ctx, quota.ID, totalSize); err != nil {
			s.logger.Error("Failed to commit quota reservation", zap.Error(err))
			// 非致命错误
		}
	}

	s.logger.Info("Multipart upload completed",
		zap.String("upload_id", req.UploadID),
		zap.String("file_id", file.ID.String()),
		zap.String("filename", file.Filename))

	return file, nil
}

// AbortUpload 中止分片上传
func (s *multipartUploadService) AbortUpload(ctx context.Context, uploadID string, tenantID uuid.UUID) error {
	// 1. 查找上传记录
	upload, err := s.multipartRepo.FindByUploadID(ctx, uploadID)
	if err != nil {
		return fmt.Errorf("upload not found: %w", err)
	}

	// 2. 验证租户
	if upload.TenantID != tenantID {
		return fmt.Errorf("permission denied")
	}

	// 3. 中止对象存储中的上传
	bucket := "files"
	objectStorage := s.storageClient.GetObjectStorage()
	if err := objectStorage.AbortMultipartUpload(ctx, bucket, upload.StorageKey, uploadID); err != nil {
		s.logger.Error("Failed to abort multipart upload in storage",
			zap.String("upload_id", uploadID),
			zap.Error(err))
		// 继续执行，清理数据库记录
	}

	// 4. 标记为已中止
	if err := s.multipartRepo.MarkAsAborted(ctx, upload.ID); err != nil {
		return fmt.Errorf("failed to mark as aborted: %w", err)
	}

	// 5. 释放预留配额
	if upload.TotalSize != nil {
		quota, _ := s.quotaRepo.GetOrCreateTenantQuota(ctx, tenantID, 10*1024*1024*1024)
		if quota != nil {
			s.quotaRepo.ReleaseReservation(ctx, quota.ID, *upload.TotalSize)
		}
	}

	s.logger.Info("Multipart upload aborted",
		zap.String("upload_id", uploadID))

	return nil
}

// GetUploadProgress 获取上传进度
func (s *multipartUploadService) GetUploadProgress(ctx context.Context, uploadID string) (*UploadProgressResponse, error) {
	upload, err := s.multipartRepo.FindByUploadID(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %w", err)
	}

	totalSize := int64(0)
	if upload.TotalSize != nil {
		totalSize = *upload.TotalSize
	}

	totalParts := 0
	if upload.TotalParts != nil {
		totalParts = *upload.TotalParts
	}

	return &UploadProgressResponse{
		UploadID:      upload.UploadID,
		Filename:      upload.Filename,
		TotalSize:     totalSize,
		TotalParts:    totalParts,
		UploadedParts: upload.UploadedParts,
		UploadedCount: len(upload.UploadedParts),
		Progress:      upload.GetProgress(),
		Status:        string(upload.Status),
		ExpiresAt:     upload.ExpiresAt,
		CreatedAt:     upload.CreatedAt,
		UpdatedAt:     upload.UpdatedAt,
	}, nil
}

// ListUploadedParts 列出已上传的分片
func (s *multipartUploadService) ListUploadedParts(ctx context.Context, uploadID string) ([]UploadedPartInfo, error) {
	upload, err := s.multipartRepo.FindByUploadID(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %w", err)
	}

	bucket := "files"
	objectStorage := s.storageClient.GetObjectStorage()

	parts, err := objectStorage.ListObjectParts(ctx, bucket, upload.StorageKey, uploadID, storage.ListObjectPartsOptions{
		MaxParts: s.maxPartsPerUpload,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list parts: %w", err)
	}

	result := make([]UploadedPartInfo, len(parts))
	for i, part := range parts {
		result[i] = UploadedPartInfo{
			PartNumber:   part.PartNumber,
			Size:         part.Size,
			ETag:         part.ETag,
			LastModified: part.LastModified,
		}
	}

	return result, nil
}

// GetRemainingParts 获取剩余分片
func (s *multipartUploadService) GetRemainingParts(ctx context.Context, uploadID string) ([]int, error) {
	upload, err := s.multipartRepo.FindByUploadID(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %w", err)
	}

	return upload.GetRemainingParts(), nil
}

// CleanExpiredUploads 清理过期的上传
func (s *multipartUploadService) CleanExpiredUploads(ctx context.Context) (int64, error) {
	// 1. 获取过期的上传列表
	expired, err := s.multipartRepo.ListExpiredUploads(ctx, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to list expired uploads: %w", err)
	}

	bucket := "files"
	objectStorage := s.storageClient.GetObjectStorage()
	cleaned := int64(0)

	// 2. 逐个中止并清理
	for _, upload := range expired {
		// 中止对象存储中的上传
		if err := objectStorage.AbortMultipartUpload(ctx, bucket, upload.StorageKey, upload.UploadID); err != nil {
			s.logger.Error("Failed to abort expired upload",
				zap.String("upload_id", upload.UploadID),
				zap.Error(err))
			// 继续下一个
			continue
		}

		// 释放配额
		if upload.TotalSize != nil {
			quota, _ := s.quotaRepo.GetOrCreateTenantQuota(ctx, upload.TenantID, 10*1024*1024*1024)
			if quota != nil {
				s.quotaRepo.ReleaseReservation(ctx, quota.ID, *upload.TotalSize)
			}
		}

		cleaned++
	}

	// 3. 清理数据库记录
	count, err := s.multipartRepo.CleanExpiredUploads(ctx, time.Now())
	if err != nil {
		return cleaned, fmt.Errorf("failed to clean database records: %w", err)
	}

	s.logger.Info("Cleaned expired uploads",
		zap.Int64("storage_cleaned", cleaned),
		zap.Int64("db_cleaned", count))

	return count, nil
}

// calculatePartSize 计算分片大小
func (s *multipartUploadService) calculatePartSize(totalSize int64) int64 {
	// 如果文件小于默认分片大小，使用文件大小
	if totalSize <= s.partSize {
		return totalSize
	}

	// 计算需要的分片数
	parts := (totalSize + s.partSize - 1) / s.partSize

	// 如果分片数超过最大限制，增加分片大小
	if parts > int64(s.maxPartsPerUpload) {
		partSize := (totalSize + int64(s.maxPartsPerUpload) - 1) / int64(s.maxPartsPerUpload)
		// 确保不超过最大分片大小
		if partSize > s.maxPartSize {
			return s.maxPartSize
		}
		return partSize
	}

	return s.partSize
}

// calculateChecksum 计算文件 checksum
func (s *multipartUploadService) calculateChecksum(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
