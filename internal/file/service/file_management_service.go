package service

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
	"github.com/lk2023060901/go-next-erp/internal/file/repository"
	"github.com/lk2023060901/go-next-erp/internal/file/utils"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"github.com/lk2023060901/go-next-erp/pkg/storage"
	"go.uber.org/zap"
)

// FileManagementService 文件管理服务接口
type FileManagementService interface {
	// 批量删除文件
	BatchDelete(ctx context.Context, req *BatchDeleteRequest) (*BatchDeleteResult, error)

	// 移动文件（更改分类/路径）
	MoveFile(ctx context.Context, req *MoveFileRequest) (*model.File, error)

	// 重命名文件
	RenameFile(ctx context.Context, req *RenameFileRequest) (*model.File, error)

	// 复制文件
	CopyFile(ctx context.Context, req *CopyFileRequest) (*model.File, error)

	// 批量移动文件
	BatchMove(ctx context.Context, req *BatchMoveRequest) (*BatchMoveResult, error)

	// 归档文件
	ArchiveFile(ctx context.Context, fileID uuid.UUID, tenantID uuid.UUID) error

	// 恢复已删除的文件
	RestoreFile(ctx context.Context, fileID uuid.UUID, tenantID uuid.UUID) error
}

// BatchDeleteRequest 批量删除请求
type BatchDeleteRequest struct {
	FileIDs    []uuid.UUID
	TenantID   uuid.UUID
	UserID     uuid.UUID
	HardDelete bool // 是否物理删除（从对象存储中删除）
}

// BatchDeleteResult 批量删除结果
type BatchDeleteResult struct {
	TotalCount   int                  `json:"total_count"`
	SuccessCount int                  `json:"success_count"`
	FailedCount  int                  `json:"failed_count"`
	Errors       map[uuid.UUID]string `json:"errors,omitempty"`
	DeletedFiles []uuid.UUID          `json:"deleted_files"`
	FreedSpace   int64                `json:"freed_space"` // 释放的空间（字节）
}

// MoveFileRequest 移动文件请求
type MoveFileRequest struct {
	FileID      uuid.UUID
	TenantID    uuid.UUID
	UserID      uuid.UUID
	NewCategory string // 新分类
	NewPath     string // 新路径（可选）
}

// RenameFileRequest 重命名文件请求
type RenameFileRequest struct {
	FileID      uuid.UUID
	TenantID    uuid.UUID
	UserID      uuid.UUID
	NewFilename string
}

// CopyFileRequest 复制文件请求
type CopyFileRequest struct {
	FileID   uuid.UUID
	TenantID uuid.UUID
	UserID   uuid.UUID
	NewName  string // 新文件名（可选，默认为 "原文件名_copy"）
}

// BatchMoveRequest 批量移动请求
type BatchMoveRequest struct {
	FileIDs     []uuid.UUID
	TenantID    uuid.UUID
	UserID      uuid.UUID
	NewCategory string
}

// BatchMoveResult 批量移动结果
type BatchMoveResult struct {
	TotalCount   int                  `json:"total_count"`
	SuccessCount int                  `json:"success_count"`
	FailedCount  int                  `json:"failed_count"`
	Errors       map[uuid.UUID]string `json:"errors,omitempty"`
	MovedFiles   []uuid.UUID          `json:"moved_files"`
}

type fileManagementService struct {
	fileRepo  repository.FileRepository
	quotaRepo repository.QuotaRepository
	storage   storage.Storage
	logger    *logger.Logger
}

// NewFileManagementService 创建文件管理服务
func NewFileManagementService(
	fileRepo repository.FileRepository,
	quotaRepo repository.QuotaRepository,
	storage storage.Storage,
	logger *logger.Logger,
) FileManagementService {
	return &fileManagementService{
		fileRepo:  fileRepo,
		quotaRepo: quotaRepo,
		storage:   storage,
		logger:    logger.With(zap.String("service", "file_management")),
	}
}

// BatchDelete 批量删除文件
func (s *fileManagementService) BatchDelete(ctx context.Context, req *BatchDeleteRequest) (*BatchDeleteResult, error) {
	result := &BatchDeleteResult{
		TotalCount:   len(req.FileIDs),
		SuccessCount: 0,
		FailedCount:  0,
		Errors:       make(map[uuid.UUID]string),
		DeletedFiles: make([]uuid.UUID, 0),
		FreedSpace:   0,
	}

	objectStorage := s.storage.GetObjectStorage()

	for _, fileID := range req.FileIDs {
		// 1. 查找文件
		file, err := s.fileRepo.FindByID(ctx, fileID)
		if err != nil {
			result.FailedCount++
			result.Errors[fileID] = fmt.Sprintf("file not found: %v", err)
			continue
		}

		// 2. 检查权限
		if !file.CanAccess(req.UserID, req.TenantID) {
			result.FailedCount++
			result.Errors[fileID] = "access denied"
			continue
		}

		// 3. 软删除或硬删除
		if req.HardDelete {
			// 物理删除：从对象存储中删除
			if err := objectStorage.RemoveObject(ctx, file.Bucket, file.StorageKey); err != nil {
				s.logger.Error("Failed to delete object from storage",
					zap.String("file_id", fileID.String()),
					zap.String("storage_key", file.StorageKey),
					zap.Error(err))
				result.FailedCount++
				result.Errors[fileID] = fmt.Sprintf("failed to delete from storage: %v", err)
				continue
			}

			// 删除缩略图（如果存在）
			if file.ThumbnailKey != nil {
				objectStorage.RemoveObject(ctx, file.Bucket, *file.ThumbnailKey)
			}

			// 从数据库中删除
			if err := s.fileRepo.Delete(ctx, fileID); err != nil {
				s.logger.Error("Failed to delete file record",
					zap.String("file_id", fileID.String()),
					zap.Error(err))
				result.FailedCount++
				result.Errors[fileID] = fmt.Sprintf("failed to delete record: %v", err)
				continue
			}
		} else {
			// 软删除：只标记为已删除
			if err := s.fileRepo.SoftDelete(ctx, fileID); err != nil {
				s.logger.Error("Failed to soft delete file",
					zap.String("file_id", fileID.String()),
					zap.Error(err))
				result.FailedCount++
				result.Errors[fileID] = fmt.Sprintf("failed to soft delete: %v", err)
				continue
			}
		}

		// 4. 释放配额
		quota, _ := s.quotaRepo.GetOrCreateTenantQuota(ctx, req.TenantID, 10*1024*1024*1024)
		if quota != nil {
			quota.ReleaseUsage(file.Size)
			s.quotaRepo.Update(ctx, quota)
		}

		result.SuccessCount++
		result.DeletedFiles = append(result.DeletedFiles, fileID)
		result.FreedSpace += file.Size

		s.logger.Info("File deleted",
			zap.String("file_id", fileID.String()),
			zap.String("filename", file.Filename),
			zap.Bool("hard_delete", req.HardDelete))
	}

	return result, nil
}

// MoveFile 移动文件
func (s *fileManagementService) MoveFile(ctx context.Context, req *MoveFileRequest) (*model.File, error) {
	// 1. 查找文件
	file, err := s.fileRepo.FindByID(ctx, req.FileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// 2. 检查权限
	if !file.CanAccess(req.UserID, req.TenantID) {
		return nil, fmt.Errorf("access denied")
	}

	// 3. 如果需要移动到新路径（更改存储键）
	if req.NewPath != "" {
		oldStorageKey := file.StorageKey
		newStorageKey := req.NewPath + "/" + file.Filename

		objectStorage := s.storage.GetObjectStorage()

		// 复制对象到新路径
		_, err := objectStorage.CopyObject(ctx, file.Bucket, newStorageKey, file.Bucket, oldStorageKey, storage.CopyObjectOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to copy object: %w", err)
		}

		// 删除旧对象
		if err := objectStorage.RemoveObject(ctx, file.Bucket, oldStorageKey); err != nil {
			s.logger.Error("Failed to delete old object",
				zap.String("storage_key", oldStorageKey),
				zap.Error(err))
			// 非致命错误，继续
		}

		file.StorageKey = newStorageKey
	}

	// 4. 更新分类
	if req.NewCategory != "" {
		file.Category = req.NewCategory
	}

	// 5. 更新数据库记录
	if err := s.fileRepo.Update(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to update file: %w", err)
	}

	s.logger.Info("File moved",
		zap.String("file_id", req.FileID.String()),
		zap.String("new_category", req.NewCategory),
		zap.String("new_path", req.NewPath))

	return file, nil
}

// RenameFile 重命名文件
func (s *fileManagementService) RenameFile(ctx context.Context, req *RenameFileRequest) (*model.File, error) {
	// 1. 查找文件
	file, err := s.fileRepo.FindByID(ctx, req.FileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// 2. 检查权限
	if !file.CanAccess(req.UserID, req.TenantID) {
		return nil, fmt.Errorf("access denied")
	}

	// 3. 验证新文件名
	newFilename := utils.SanitizeFilename(req.NewFilename)
	if newFilename == "" {
		return nil, fmt.Errorf("invalid filename")
	}

	// 4. 生成新的存储键
	oldStorageKey := file.StorageKey
	dir := filepath.Dir(oldStorageKey)
	newExt := utils.GetFileExtension(newFilename)
	newStorageKey := filepath.Join(dir, newFilename)

	objectStorage := s.storage.GetObjectStorage()

	// 5. 复制对象到新键
	_, err = objectStorage.CopyObject(ctx, file.Bucket, newStorageKey, file.Bucket, oldStorageKey, storage.CopyObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to copy object: %w", err)
	}

	// 6. 删除旧对象
	if err := objectStorage.RemoveObject(ctx, file.Bucket, oldStorageKey); err != nil {
		s.logger.Error("Failed to delete old object",
			zap.String("storage_key", oldStorageKey),
			zap.Error(err))
		// 非致命错误，继续
	}

	// 7. 更新文件信息
	file.Filename = newFilename
	file.StorageKey = newStorageKey
	file.Extension = newExt

	if err := s.fileRepo.Update(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to update file: %w", err)
	}

	s.logger.Info("File renamed",
		zap.String("file_id", req.FileID.String()),
		zap.String("old_filename", file.Filename),
		zap.String("new_filename", newFilename))

	return file, nil
}

// CopyFile 复制文件
func (s *fileManagementService) CopyFile(ctx context.Context, req *CopyFileRequest) (*model.File, error) {
	// 1. 查找原文件
	originalFile, err := s.fileRepo.FindByID(ctx, req.FileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// 2. 检查权限
	if !originalFile.CanAccess(req.UserID, req.TenantID) {
		return nil, fmt.Errorf("access denied")
	}

	// 3. 检查配额
	quota, err := s.quotaRepo.GetOrCreateTenantQuota(ctx, req.TenantID, 10*1024*1024*1024)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	if !quota.CanAllocate(originalFile.Size) {
		return nil, fmt.Errorf("quota exceeded")
	}

	// 4. 生成新文件名
	newFilename := req.NewName
	if newFilename == "" {
		ext := originalFile.Extension
		baseName := originalFile.Filename[:len(originalFile.Filename)-len(ext)]
		newFilename = fmt.Sprintf("%s_copy%s", baseName, ext)
	}

	// 5. 生成新的存储键
	newStorageKey := utils.GenerateStorageKey(req.TenantID, newFilename)

	objectStorage := s.storage.GetObjectStorage()

	// 6. 复制对象
	_, err = objectStorage.CopyObject(ctx, originalFile.Bucket, newStorageKey, originalFile.Bucket, originalFile.StorageKey, storage.CopyObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to copy object: %w", err)
	}

	// 7. 创建新文件记录
	newFile := &model.File{
		TenantID:      req.TenantID,
		Filename:      newFilename,
		StorageKey:    newStorageKey,
		Size:          originalFile.Size,
		MimeType:      originalFile.MimeType,
		ContentType:   originalFile.ContentType,
		Checksum:      originalFile.Checksum,
		Extension:     originalFile.Extension,
		Bucket:        originalFile.Bucket,
		Category:      originalFile.Category,
		Tags:          originalFile.Tags,
		Metadata:      originalFile.Metadata,
		Status:        model.FileStatusActive,
		IsTemporary:   originalFile.IsTemporary,
		IsPublic:      originalFile.IsPublic,
		VersionNumber: 1,
		UploadedBy:    req.UserID,
		AccessLevel:   originalFile.AccessLevel,
	}

	if err := s.fileRepo.Create(ctx, newFile); err != nil {
		// 回滚：删除已复制的对象
		objectStorage.RemoveObject(ctx, newFile.Bucket, newStorageKey)
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// 8. 更新配额
	quota.QuotaUsed += originalFile.Size
	quota.FileCountUsed++
	s.quotaRepo.Update(ctx, quota)

	s.logger.Info("File copied",
		zap.String("original_file_id", req.FileID.String()),
		zap.String("new_file_id", newFile.ID.String()),
		zap.String("new_filename", newFilename))

	return newFile, nil
}

// BatchMove 批量移动文件
func (s *fileManagementService) BatchMove(ctx context.Context, req *BatchMoveRequest) (*BatchMoveResult, error) {
	result := &BatchMoveResult{
		TotalCount:   len(req.FileIDs),
		SuccessCount: 0,
		FailedCount:  0,
		Errors:       make(map[uuid.UUID]string),
		MovedFiles:   make([]uuid.UUID, 0),
	}

	for _, fileID := range req.FileIDs {
		moveReq := &MoveFileRequest{
			FileID:      fileID,
			TenantID:    req.TenantID,
			UserID:      req.UserID,
			NewCategory: req.NewCategory,
		}

		_, err := s.MoveFile(ctx, moveReq)
		if err != nil {
			result.FailedCount++
			result.Errors[fileID] = err.Error()
			continue
		}

		result.SuccessCount++
		result.MovedFiles = append(result.MovedFiles, fileID)
	}

	return result, nil
}

// ArchiveFile 归档文件
func (s *fileManagementService) ArchiveFile(ctx context.Context, fileID uuid.UUID, tenantID uuid.UUID) error {
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	if file.TenantID != tenantID {
		return fmt.Errorf("access denied")
	}

	now := time.Now()
	file.Status = model.FileStatusArchived
	file.ArchivedAt = &now

	if err := s.fileRepo.Update(ctx, file); err != nil {
		return fmt.Errorf("failed to archive file: %w", err)
	}

	s.logger.Info("File archived",
		zap.String("file_id", fileID.String()))

	return nil
}

// RestoreFile 恢复已删除的文件
func (s *fileManagementService) RestoreFile(ctx context.Context, fileID uuid.UUID, tenantID uuid.UUID) error {
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	if file.TenantID != tenantID {
		return fmt.Errorf("access denied")
	}

	if file.DeletedAt == nil {
		return fmt.Errorf("file is not deleted")
	}

	file.Status = model.FileStatusActive
	file.DeletedAt = nil

	if err := s.fileRepo.Update(ctx, file); err != nil {
		return fmt.Errorf("failed to restore file: %w", err)
	}

	s.logger.Info("File restored",
		zap.String("file_id", fileID.String()))

	return nil
}
