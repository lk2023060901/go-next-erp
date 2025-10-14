package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
	"github.com/lk2023060901/go-next-erp/internal/file/service"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"go.uber.org/zap"
)

// FileHandler 文件处理器 - 整合所有文件相关服务
type FileHandler struct {
	// 核心服务
	uploadService   service.UploadService
	downloadService service.DownloadService
	fileManagement  service.FileManagementService
	multipartUpload service.MultipartUploadService

	// 配额管理
	quotaService      service.QuotaService
	quotaAlertService service.QuotaAlertService

	// 文件关系
	relationService service.FileRelationService

	// 版本管理
	versionService service.FileVersionService

	// 图片处理
	thumbnailService   service.ThumbnailService
	compressionService service.CompressionService

	// 清理服务
	cleanupService service.CleanupService

	logger *logger.Logger
}

// NewFileHandler 创建文件处理器
func NewFileHandler(
	uploadService service.UploadService,
	downloadService service.DownloadService,
	fileManagement service.FileManagementService,
	multipartUpload service.MultipartUploadService,
	quotaService service.QuotaService,
	quotaAlertService service.QuotaAlertService,
	relationService service.FileRelationService,
	versionService service.FileVersionService,
	thumbnailService service.ThumbnailService,
	compressionService service.CompressionService,
	cleanupService service.CleanupService,
	logger *logger.Logger,
) *FileHandler {
	return &FileHandler{
		uploadService:      uploadService,
		downloadService:    downloadService,
		fileManagement:     fileManagement,
		multipartUpload:    multipartUpload,
		quotaService:       quotaService,
		quotaAlertService:  quotaAlertService,
		relationService:    relationService,
		versionService:     versionService,
		thumbnailService:   thumbnailService,
		compressionService: compressionService,
		cleanupService:     cleanupService,
		logger:             logger.With(zap.String("handler", "file")),
	}
}

// UploadFile 上传文件
func (h *FileHandler) UploadFile(ctx context.Context, req *UploadFileRequest) (*UploadFileResponse, error) {
	h.logger.Info("Uploading file",
		zap.String("filename", req.Filename),
		zap.String("tenant_id", req.TenantID.String()),
	)

	// 1. 上传文件
	uploadReq := &service.UploadRequest{
		TenantID:    req.TenantID,
		UploadedBy:  req.UploadedBy,
		Filename:    req.Filename,
		Reader:      req.Reader,
		Size:        req.Size,
		ContentType: req.ContentType,
		IsTemporary: req.IsTemporary,
		ExpiresAt:   req.ExpiresAt,
		Category:    req.Category,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
		AccessLevel: req.AccessLevel,
	}

	file, err := h.uploadService.Upload(ctx, uploadReq)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// 2. 如果是图片，自动生成缩略图
	if h.thumbnailService.IsImageFile(req.Filename) && req.GenerateThumbnail {
		go func() {
			// 在后台生成缩略图
			if err := h.thumbnailService.GenerateThumbnail(context.Background(), file); err != nil {
				h.logger.Error("Failed to generate thumbnail",
					zap.String("file_id", file.ID.String()),
					zap.Error(err),
				)
			}
		}()
	}

	// 3. 如果需要压缩
	if req.Compress && h.compressionService.IsCompressible(req.Filename) {
		go func() {
			opts := service.DefaultCompressionOptions
			if _, err := h.compressionService.CompressImageInPlace(context.Background(), file.ID, opts); err != nil {
				h.logger.Error("Failed to compress file",
					zap.String("file_id", file.ID.String()),
					zap.Error(err),
				)
			}
		}()
	}

	return &UploadFileResponse{
		File: file,
	}, nil
}

// DownloadFile 下载文件
func (h *FileHandler) DownloadFile(ctx context.Context, req *DownloadFileRequest) (*DownloadFileResponse, error) {
	h.logger.Info("Downloading file",
		zap.String("file_id", req.FileID.String()),
		zap.String("user_id", req.UserID.String()),
	)

	// 生成下载URL
	url, err := h.downloadService.GetDownloadURL(ctx, req.FileID, req.UserID, req.TenantID, req.Expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to get download URL: %w", err)
	}

	// 记录下载
	if req.RecordDownload {
		go func() {
			recordReq := &service.RecordDownloadRequest{
				TenantID:     req.TenantID,
				FileID:       req.FileID,
				DownloadedBy: req.UserID,
				IPAddress:    req.IPAddress,
				UserAgent:    req.UserAgent,
			}
			if err := h.downloadService.RecordDownload(context.Background(), recordReq); err != nil {
				h.logger.Error("Failed to record download",
					zap.String("file_id", req.FileID.String()),
					zap.Error(err),
				)
			}
		}()
	}

	return &DownloadFileResponse{
		URL:       url,
		ExpiresAt: time.Now().Add(req.Expiry),
	}, nil
}

// GetFile 获取文件信息
func (h *FileHandler) GetFile(ctx context.Context, fileID uuid.UUID) (*model.File, error) {
	return h.fileManagement.GetFile(ctx, fileID)
}

// DeleteFile 删除文件
func (h *FileHandler) DeleteFile(ctx context.Context, req *DeleteFileRequest) error {
	return h.fileManagement.DeleteFile(ctx, req.FileID, req.UserID, req.Permanent)
}

// MoveFile 移动文件
func (h *FileHandler) MoveFile(ctx context.Context, req *MoveFileRequest) (*model.File, error) {
	return h.fileManagement.MoveFile(ctx, req.FileID, req.NewCategory, req.UserID)
}

// RenameFile 重命名文件
func (h *FileHandler) RenameFile(ctx context.Context, req *RenameFileRequest) (*model.File, error) {
	return h.fileManagement.RenameFile(ctx, req.FileID, req.NewFilename, req.UserID)
}

// CopyFile 复制文件
func (h *FileHandler) CopyFile(ctx context.Context, req *CopyFileRequest) (*model.File, error) {
	return h.fileManagement.CopyFile(ctx, req.FileID, req.UserID)
}

// ArchiveFile 归档文件
func (h *FileHandler) ArchiveFile(ctx context.Context, fileID, userID uuid.UUID) error {
	return h.fileManagement.ArchiveFile(ctx, fileID, userID)
}

// RestoreFile 恢复文件
func (h *FileHandler) RestoreFile(ctx context.Context, fileID, userID uuid.UUID) error {
	return h.fileManagement.RestoreFile(ctx, fileID, userID)
}

// BatchDelete 批量删除
func (h *FileHandler) BatchDelete(ctx context.Context, req *BatchDeleteRequest) (*BatchOperationResult, error) {
	return h.fileManagement.BatchDelete(ctx, req.FileIDs, req.UserID, req.Permanent)
}

// GetQuota 获取配额信息
func (h *FileHandler) GetQuota(ctx context.Context, tenantID uuid.UUID) (*model.Quota, error) {
	return h.quotaService.GetTenantQuota(ctx, tenantID)
}

// CheckQuotaAlert 检查配额预警
func (h *FileHandler) CheckQuotaAlert(ctx context.Context, tenantID uuid.UUID) ([]service.QuotaAlert, error) {
	return h.quotaAlertService.CheckTenantQuota(ctx, tenantID)
}

// AttachFileToEntity 关联文件到业务实体
func (h *FileHandler) AttachFileToEntity(ctx context.Context, req *AttachFileRequest) error {
	attachReq := &service.AttachFileRequest{
		FileID:     req.FileID,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		FieldName:  req.FieldName,
		AttachedBy: req.AttachedBy,
	}
	return h.relationService.AttachFileToEntity(ctx, attachReq)
}

// GetEntityFiles 获取实体的文件
func (h *FileHandler) GetEntityFiles(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) ([]*model.File, error) {
	return h.relationService.GetEntityFiles(ctx, entityType, entityID)
}

// CreateVersion 创建文件版本
func (h *FileHandler) CreateVersion(ctx context.Context, req *CreateVersionRequest) (*model.FileVersion, error) {
	// 读取文件内容
	data, err := io.ReadAll(req.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	versionReq := &service.CreateVersionRequest{
		FileID:   req.FileID,
		UserID:   req.UserID,
		TenantID: req.TenantID,
		Comment:  req.Comment,
		Reader:   bytes.NewReader(data),
		Size:     int64(len(data)),
		Checksum: req.Checksum,
	}

	return h.versionService.CreateVersion(ctx, versionReq)
}

// GetVersionHistory 获取版本历史
func (h *FileHandler) GetVersionHistory(ctx context.Context, fileID uuid.UUID) ([]*model.FileVersion, error) {
	return h.versionService.GetVersionHistory(ctx, fileID)
}

// RevertToVersion 回滚到指定版本
func (h *FileHandler) RevertToVersion(ctx context.Context, req *RevertVersionRequest) error {
	revertReq := &service.RevertVersionRequest{
		FileID:        req.FileID,
		VersionNumber: req.VersionNumber,
		UserID:        req.UserID,
		Comment:       req.Comment,
	}
	return h.versionService.RevertToVersion(ctx, revertReq)
}

// GenerateThumbnail 生成缩略图
func (h *FileHandler) GenerateThumbnail(ctx context.Context, fileID uuid.UUID) error {
	file, err := h.fileManagement.GetFile(ctx, fileID)
	if err != nil {
		return err
	}
	return h.thumbnailService.GenerateThumbnail(ctx, file)
}

// CompressFile 压缩文件
func (h *FileHandler) CompressFile(ctx context.Context, req *CompressFileRequest) (*service.CompressionResult, error) {
	opts := service.CompressionOptions{
		Quality:    req.Quality,
		MaxWidth:   req.MaxWidth,
		MaxHeight:  req.MaxHeight,
		KeepAspect: req.KeepAspect,
	}

	if req.InPlace {
		return h.compressionService.CompressImageInPlace(ctx, req.FileID, opts)
	}

	file, err := h.fileManagement.GetFile(ctx, req.FileID)
	if err != nil {
		return nil, err
	}

	return h.compressionService.CompressImage(ctx, file, opts)
}

// CleanExpiredFiles 清理过期文件
func (h *FileHandler) CleanExpiredFiles(ctx context.Context) (int64, error) {
	return h.cleanupService.CleanExpiredFiles(ctx)
}

// InitiateMultipartUpload 初始化分片上传
func (h *FileHandler) InitiateMultipartUpload(ctx context.Context, req *InitiateMultipartUploadRequest) (*MultipartUploadInfo, error) {
	initiateReq := &service.InitiateMultipartUploadRequest{
		TenantID:    req.TenantID,
		UserID:      req.UserID,
		Filename:    req.Filename,
		TotalSize:   req.TotalSize,
		PartSize:    req.PartSize,
		ContentType: req.ContentType,
	}

	uploadInfo, err := h.multipartUpload.InitiateUpload(ctx, initiateReq)
	if err != nil {
		return nil, err
	}

	return &MultipartUploadInfo{
		UploadID:   uploadInfo.UploadID,
		StorageKey: uploadInfo.StorageKey,
		PartSize:   uploadInfo.PartSize,
		TotalParts: uploadInfo.TotalParts,
	}, nil
}

// UploadPart 上传分片
func (h *FileHandler) UploadPart(ctx context.Context, req *UploadPartRequest) (*PartUploadResult, error) {
	uploadReq := &service.UploadPartRequest{
		UploadID:   req.UploadID,
		PartNumber: req.PartNumber,
		Reader:     req.Reader,
		Size:       req.Size,
	}

	part, err := h.multipartUpload.UploadPart(ctx, uploadReq)
	if err != nil {
		return nil, err
	}

	return &PartUploadResult{
		PartNumber: part.PartNumber,
		ETag:       part.ETag,
		Size:       part.Size,
	}, nil
}

// CompleteMultipartUpload 完成分片上传
func (h *FileHandler) CompleteMultipartUpload(ctx context.Context, req *CompleteMultipartUploadRequest) (*model.File, error) {
	parts := make([]service.UploadedPart, len(req.Parts))
	for i, p := range req.Parts {
		parts[i] = service.UploadedPart{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
		}
	}

	completeReq := &service.CompleteMultipartUploadRequest{
		UploadID: req.UploadID,
		Parts:    parts,
	}

	return h.multipartUpload.CompleteUpload(ctx, completeReq)
}

// AbortMultipartUpload 取消分片上传
func (h *FileHandler) AbortMultipartUpload(ctx context.Context, uploadID uuid.UUID) error {
	return h.multipartUpload.AbortUpload(ctx, uploadID)
}

// GetUploadProgress 获取上传进度
func (h *FileHandler) GetUploadProgress(ctx context.Context, uploadID uuid.UUID) (*service.UploadProgress, error) {
	return h.multipartUpload.GetUploadProgress(ctx, uploadID)
}
