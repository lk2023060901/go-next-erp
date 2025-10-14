package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
	"github.com/lk2023060901/go-next-erp/internal/file/repository"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"github.com/lk2023060901/go-next-erp/pkg/storage"
	"go.uber.org/zap"
)

// CreateVersionRequest 创建版本请求
type CreateVersionRequest struct {
	FileID   uuid.UUID
	UserID   uuid.UUID
	TenantID uuid.UUID
	Comment  string
	Reader   io.Reader
	Size     int64
	Checksum string
}

// RevertVersionRequest 回滚版本请求
type RevertVersionRequest struct {
	FileID        uuid.UUID
	VersionNumber int
	UserID        uuid.UUID
	Comment       string
}

// VersionCompareResult 版本比较结果
type VersionCompareResult struct {
	Version1      *model.FileVersion
	Version2      *model.FileVersion
	SizeDiff      int64
	TimeDiff      time.Duration
	IsSameContent bool // 基于 checksum
}

// FileVersionService 文件版本服务接口
type FileVersionService interface {
	// CreateVersion 创建新版本
	CreateVersion(ctx context.Context, req *CreateVersionRequest) (*model.FileVersion, error)

	// GetVersionHistory 获取版本历史
	GetVersionHistory(ctx context.Context, fileID uuid.UUID) ([]*model.FileVersion, error)

	// GetVersion 获取指定版本
	GetVersion(ctx context.Context, fileID uuid.UUID, versionNumber int) (*model.FileVersion, error)

	// RevertToVersion 回滚到指定版本
	RevertToVersion(ctx context.Context, req *RevertVersionRequest) error

	// CompareVersions 比较两个版本
	CompareVersions(ctx context.Context, fileID uuid.UUID, version1, version2 int) (*VersionCompareResult, error)

	// DeleteVersion 删除指定版本
	DeleteVersion(ctx context.Context, fileID uuid.UUID, versionNumber int) error

	// GetLatestVersion 获取最新版本
	GetLatestVersion(ctx context.Context, fileID uuid.UUID) (*model.FileVersion, error)

	// DownloadVersion 下载指定版本
	DownloadVersion(ctx context.Context, fileID uuid.UUID, versionNumber int) (io.ReadCloser, *model.FileVersion, error)
}

type fileVersionService struct {
	storage     storage.Storage
	fileRepo    repository.FileRepository
	versionRepo repository.VersionRepository
	logger      *logger.Logger
}

// NewFileVersionService 创建文件版本服务
func NewFileVersionService(
	storage storage.Storage,
	fileRepo repository.FileRepository,
	versionRepo repository.VersionRepository,
	logger *logger.Logger,
) FileVersionService {
	return &fileVersionService{
		storage:     storage,
		fileRepo:    fileRepo,
		versionRepo: versionRepo,
		logger:      logger,
	}
}

// CreateVersion 创建新版本
func (s *fileVersionService) CreateVersion(ctx context.Context, req *CreateVersionRequest) (*model.FileVersion, error) {
	// 1. 获取当前文件
	file, err := s.fileRepo.FindByID(ctx, req.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to find file: %w", err)
	}

	// 2. 检查权限
	if file.TenantID != req.TenantID {
		return nil, fmt.Errorf("access denied: tenant mismatch")
	}

	// 3. 获取当前版本号并递增
	newVersionNumber := file.VersionNumber + 1

	// 4. 生成新的存储键
	newStorageKey := fmt.Sprintf("versions/%s/%s/v%d",
		file.TenantID.String(),
		file.ID.String(),
		newVersionNumber,
	)

	// 5. 上传新版本到存储
	objectStorage := s.storage.GetObjectStorage()
	_, err = objectStorage.PutObject(ctx, file.Bucket, newStorageKey, req.Reader, req.Size, storage.PutObjectOptions{
		ContentType: file.ContentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload new version: %w", err)
	}

	// 6. 创建版本记录
	version := &model.FileVersion{
		FileID:        req.FileID,
		TenantID:      req.TenantID,
		VersionNumber: newVersionNumber,
		StorageKey:    newStorageKey,
		Size:          req.Size,
		Checksum:      req.Checksum,
		Filename:      file.Filename,
		MimeType:      file.MimeType,
		Comment:       req.Comment,
		ChangedBy:     req.UserID,
		ChangeType:    model.ChangeTypeUpdate,
	}

	if err := s.versionRepo.Create(ctx, version); err != nil {
		// 删除已上传的文件
		_ = objectStorage.RemoveObject(ctx, file.Bucket, newStorageKey)
		return nil, fmt.Errorf("failed to create version record: %w", err)
	}

	// 7. 更新文件的当前版本号
	file.VersionNumber = newVersionNumber
	if err := s.fileRepo.Update(ctx, file); err != nil {
		s.logger.Error("Failed to update file version number",
			zap.String("file_id", req.FileID.String()),
			zap.Error(err),
		)
	}

	s.logger.Info("Version created",
		zap.String("file_id", req.FileID.String()),
		zap.Int("version", newVersionNumber),
		zap.String("user_id", req.UserID.String()),
	)

	return version, nil
}

// GetVersionHistory 获取版本历史
func (s *fileVersionService) GetVersionHistory(ctx context.Context, fileID uuid.UUID) ([]*model.FileVersion, error) {
	versions, err := s.versionRepo.ListByFile(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get version history: %w", err)
	}

	s.logger.Info("Retrieved version history",
		zap.String("file_id", fileID.String()),
		zap.Int("count", len(versions)),
	)

	return versions, nil
}

// GetVersion 获取指定版本
func (s *fileVersionService) GetVersion(ctx context.Context, fileID uuid.UUID, versionNumber int) (*model.FileVersion, error) {
	version, err := s.versionRepo.FindByFileAndVersion(ctx, fileID, versionNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to find version: %w", err)
	}
	return version, nil
}

// RevertToVersion 回滚到指定版本
func (s *fileVersionService) RevertToVersion(ctx context.Context, req *RevertVersionRequest) error {
	// 1. 获取文件
	file, err := s.fileRepo.FindByID(ctx, req.FileID)
	if err != nil {
		return fmt.Errorf("failed to find file: %w", err)
	}

	// 2. 获取目标版本
	targetVersion, err := s.versionRepo.FindByFileAndVersion(ctx, req.FileID, req.VersionNumber)
	if err != nil {
		return fmt.Errorf("failed to find target version: %w", err)
	}

	// 3. 从存储中获取目标版本的文件
	objectStorage := s.storage.GetObjectStorage()
	reader, _, err := objectStorage.GetObject(ctx, file.Bucket, targetVersion.StorageKey, storage.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get version from storage: %w", err)
	}
	defer reader.Close()

	// 4. 创建新版本（回滚作为新版本）
	newVersionNumber := file.VersionNumber + 1
	newStorageKey := fmt.Sprintf("versions/%s/%s/v%d",
		file.TenantID.String(),
		file.ID.String(),
		newVersionNumber,
	)

	// 5. 上传回滚后的文件
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read version data: %w", err)
	}

	_, err = objectStorage.PutObject(ctx, file.Bucket, newStorageKey, io.NopCloser(bytes.NewReader(data)), int64(len(data)), storage.PutObjectOptions{
		ContentType: file.ContentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload reverted version: %w", err)
	}

	// 6. 创建回滚版本记录
	revertVersion := &model.FileVersion{
		FileID:        req.FileID,
		TenantID:      file.TenantID,
		VersionNumber: newVersionNumber,
		StorageKey:    newStorageKey,
		Size:          targetVersion.Size,
		Checksum:      targetVersion.Checksum,
		Filename:      file.Filename,
		MimeType:      file.MimeType,
		Comment:       fmt.Sprintf("Reverted to version %d. %s", req.VersionNumber, req.Comment),
		ChangedBy:     req.UserID,
		ChangeType:    model.ChangeTypeRevert,
	}

	if err := s.versionRepo.Create(ctx, revertVersion); err != nil {
		_ = objectStorage.RemoveObject(ctx, file.Bucket, newStorageKey)
		return fmt.Errorf("failed to create revert version record: %w", err)
	}

	// 7. 更新文件的当前版本
	file.VersionNumber = newVersionNumber
	file.StorageKey = newStorageKey
	file.Size = targetVersion.Size
	file.Checksum = targetVersion.Checksum

	if err := s.fileRepo.Update(ctx, file); err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	s.logger.Info("File reverted to version",
		zap.String("file_id", req.FileID.String()),
		zap.Int("target_version", req.VersionNumber),
		zap.Int("new_version", newVersionNumber),
		zap.String("user_id", req.UserID.String()),
	)

	return nil
}

// CompareVersions 比较两个版本
func (s *fileVersionService) CompareVersions(ctx context.Context, fileID uuid.UUID, version1, version2 int) (*VersionCompareResult, error) {
	// 获取两个版本
	v1, err := s.versionRepo.FindByFileAndVersion(ctx, fileID, version1)
	if err != nil {
		return nil, fmt.Errorf("failed to find version %d: %w", version1, err)
	}

	v2, err := s.versionRepo.FindByFileAndVersion(ctx, fileID, version2)
	if err != nil {
		return nil, fmt.Errorf("failed to find version %d: %w", version2, err)
	}

	// 比较
	result := &VersionCompareResult{
		Version1:      v1,
		Version2:      v2,
		SizeDiff:      v2.Size - v1.Size,
		TimeDiff:      v2.CreatedAt.Sub(v1.CreatedAt),
		IsSameContent: v1.Checksum == v2.Checksum,
	}

	return result, nil
}

// DeleteVersion 删除指定版本
func (s *fileVersionService) DeleteVersion(ctx context.Context, fileID uuid.UUID, versionNumber int) error {
	// 1. 获取文件
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to find file: %w", err)
	}

	// 2. 不能删除当前版本
	if file.VersionNumber == versionNumber {
		return fmt.Errorf("cannot delete current version")
	}

	// 3. 获取要删除的版本
	version, err := s.versionRepo.FindByFileAndVersion(ctx, fileID, versionNumber)
	if err != nil {
		return fmt.Errorf("failed to find version: %w", err)
	}

	// 4. 从存储中删除文件
	objectStorage := s.storage.GetObjectStorage()
	if err := objectStorage.RemoveObject(ctx, file.Bucket, version.StorageKey); err != nil {
		s.logger.Warn("Failed to delete version from storage",
			zap.String("file_id", fileID.String()),
			zap.Int("version", versionNumber),
			zap.Error(err),
		)
	}

	// 5. 删除版本记录（需要在 repository 中实现 Delete 方法）
	// TODO: 实现 versionRepo.Delete 方法

	s.logger.Info("Version deleted",
		zap.String("file_id", fileID.String()),
		zap.Int("version", versionNumber),
	)

	return nil
}

// GetLatestVersion 获取最新版本
func (s *fileVersionService) GetLatestVersion(ctx context.Context, fileID uuid.UUID) (*model.FileVersion, error) {
	version, err := s.versionRepo.GetLatestVersion(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}
	return version, nil
}

// DownloadVersion 下载指定版本
func (s *fileVersionService) DownloadVersion(ctx context.Context, fileID uuid.UUID, versionNumber int) (io.ReadCloser, *model.FileVersion, error) {
	// 1. 获取版本
	version, err := s.versionRepo.FindByFileAndVersion(ctx, fileID, versionNumber)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find version: %w", err)
	}

	// 2. 获取文件以获取 bucket 信息
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find file: %w", err)
	}

	// 3. 从存储中获取文件
	objectStorage := s.storage.GetObjectStorage()
	reader, _, err := objectStorage.GetObject(ctx, file.Bucket, version.StorageKey, storage.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get version from storage: %w", err)
	}

	s.logger.Info("Version downloaded",
		zap.String("file_id", fileID.String()),
		zap.Int("version", versionNumber),
	)

	return reader, version, nil
}
