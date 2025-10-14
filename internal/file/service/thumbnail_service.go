package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
	"github.com/lk2023060901/go-next-erp/internal/file/repository"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"github.com/lk2023060901/go-next-erp/pkg/storage"
	"go.uber.org/zap"
)

// ThumbnailSize 缩略图尺寸配置
type ThumbnailSize struct {
	Width  int
	Height int
	Suffix string
}

var (
	// DefaultThumbnailSizes 默认缩略图尺寸
	DefaultThumbnailSizes = []ThumbnailSize{
		{Width: 150, Height: 150, Suffix: "small"},  // 小图
		{Width: 300, Height: 300, Suffix: "medium"}, // 中图
		{Width: 800, Height: 600, Suffix: "large"},  // 大图
	}

	// SupportedImageFormats 支持的图片格式
	SupportedImageFormats = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
)

// ThumbnailService 缩略图服务接口
type ThumbnailService interface {
	// GenerateThumbnail 生成缩略图
	GenerateThumbnail(ctx context.Context, file *model.File, sizes ...ThumbnailSize) error

	// GenerateSingleThumbnail 生成单个尺寸的缩略图
	GenerateSingleThumbnail(ctx context.Context, file *model.File, width, height int) (string, error)

	// RegenerateThumbnail 重新生成缩略图
	RegenerateThumbnail(ctx context.Context, fileID uuid.UUID) error

	// DeleteThumbnail 删除缩略图
	DeleteThumbnail(ctx context.Context, file *model.File) error

	// IsImageFile 检查是否为图片文件
	IsImageFile(filename string) bool
}

type thumbnailService struct {
	storage  storage.Storage
	fileRepo repository.FileRepository
	logger   *logger.Logger
}

// NewThumbnailService 创建缩略图服务
func NewThumbnailService(
	storage storage.Storage,
	fileRepo repository.FileRepository,
	logger *logger.Logger,
) ThumbnailService {
	return &thumbnailService{
		storage:  storage,
		fileRepo: fileRepo,
		logger:   logger,
	}
}

// GenerateThumbnail 生成缩略图
func (s *thumbnailService) GenerateThumbnail(ctx context.Context, file *model.File, sizes ...ThumbnailSize) error {
	// 检查是否为图片文件
	if !s.IsImageFile(file.Filename) {
		return fmt.Errorf("file is not an image: %s", file.Filename)
	}

	// 如果未指定尺寸，使用默认尺寸
	if len(sizes) == 0 {
		sizes = DefaultThumbnailSizes
	}

	// 从存储中获取原始图片
	objectStorage := s.storage.GetObjectStorage()
	reader, _, err := objectStorage.GetObject(ctx, file.Bucket, file.StorageKey, storage.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get file from storage: %w", err)
	}
	defer reader.Close()

	// 读取图片数据
	imgData, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read image data: %w", err)
	}

	// 解码图片
	img, format, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	s.logger.Info("Generating thumbnails",
		zap.String("file_id", file.ID.String()),
		zap.String("format", format),
		zap.Int("sizes", len(sizes)),
	)

	// 生成各个尺寸的缩略图
	for _, size := range sizes {
		thumbnailKey, err := s.generateAndUploadThumbnail(ctx, img, file, size, format)
		if err != nil {
			s.logger.Error("Failed to generate thumbnail",
				zap.String("file_id", file.ID.String()),
				zap.String("size", size.Suffix),
				zap.Error(err),
			)
			continue
		}

		// 更新文件的缩略图键（使用最后一个作为主缩略图）
		file.ThumbnailKey = &thumbnailKey
	}

	// 更新文件记录
	if file.ThumbnailKey != nil {
		if err := s.fileRepo.Update(ctx, file); err != nil {
			return fmt.Errorf("failed to update file thumbnail key: %w", err)
		}
	}

	return nil
}

// GenerateSingleThumbnail 生成单个尺寸的缩略图
func (s *thumbnailService) GenerateSingleThumbnail(ctx context.Context, file *model.File, width, height int) (string, error) {
	if !s.IsImageFile(file.Filename) {
		return "", fmt.Errorf("file is not an image: %s", file.Filename)
	}

	// 从存储中获取原始图片
	objectStorage := s.storage.GetObjectStorage()
	reader, _, err := objectStorage.GetObject(ctx, file.Bucket, file.StorageKey, storage.GetObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get file from storage: %w", err)
	}
	defer reader.Close()

	// 读取并解码图片
	imgData, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	img, format, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// 生成缩略图
	size := ThumbnailSize{Width: width, Height: height, Suffix: fmt.Sprintf("%dx%d", width, height)}
	thumbnailKey, err := s.generateAndUploadThumbnail(ctx, img, file, size, format)
	if err != nil {
		return "", err
	}

	return thumbnailKey, nil
}

// generateAndUploadThumbnail 生成并上传缩略图
func (s *thumbnailService) generateAndUploadThumbnail(
	ctx context.Context,
	img image.Image,
	file *model.File,
	size ThumbnailSize,
	format string,
) (string, error) {
	// 生成缩略图
	thumbnail := imaging.Fit(img, size.Width, size.Height, imaging.Lanczos)

	// 编码缩略图
	var buf bytes.Buffer
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 85}); err != nil {
			return "", fmt.Errorf("failed to encode jpeg thumbnail: %w", err)
		}
	case "png":
		if err := png.Encode(&buf, thumbnail); err != nil {
			return "", fmt.Errorf("failed to encode png thumbnail: %w", err)
		}
	default:
		// 默认使用 JPEG
		if err := jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 85}); err != nil {
			return "", fmt.Errorf("failed to encode thumbnail: %w", err)
		}
	}

	// 生成缩略图存储键
	ext := filepath.Ext(file.Filename)
	baseName := strings.TrimSuffix(file.Filename, ext)
	thumbnailKey := fmt.Sprintf("thumbnails/%s/%s_%s%s",
		file.TenantID.String(),
		baseName,
		size.Suffix,
		ext,
	)

	// 上传缩略图
	objectStorage := s.storage.GetObjectStorage()
	_, err := objectStorage.PutObject(ctx, file.Bucket, thumbnailKey, &buf, int64(buf.Len()), storage.PutObjectOptions{
		ContentType: fmt.Sprintf("image/%s", format),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	s.logger.Info("Thumbnail generated",
		zap.String("file_id", file.ID.String()),
		zap.String("size", size.Suffix),
		zap.String("key", thumbnailKey),
	)

	return thumbnailKey, nil
}

// RegenerateThumbnail 重新生成缩略图
func (s *thumbnailService) RegenerateThumbnail(ctx context.Context, fileID uuid.UUID) error {
	// 获取文件
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to find file: %w", err)
	}

	// 删除旧缩略图
	if err := s.DeleteThumbnail(ctx, file); err != nil {
		s.logger.Warn("Failed to delete old thumbnail",
			zap.String("file_id", fileID.String()),
			zap.Error(err),
		)
	}

	// 生成新缩略图
	return s.GenerateThumbnail(ctx, file)
}

// DeleteThumbnail 删除缩略图
func (s *thumbnailService) DeleteThumbnail(ctx context.Context, file *model.File) error {
	if file.ThumbnailKey == nil || *file.ThumbnailKey == "" {
		return nil // 没有缩略图，无需删除
	}

	objectStorage := s.storage.GetObjectStorage()

	// 删除所有尺寸的缩略图
	for _, size := range DefaultThumbnailSizes {
		ext := filepath.Ext(file.Filename)
		baseName := strings.TrimSuffix(file.Filename, ext)
		thumbnailKey := fmt.Sprintf("thumbnails/%s/%s_%s%s",
			file.TenantID.String(),
			baseName,
			size.Suffix,
			ext,
		)

		if err := objectStorage.RemoveObject(ctx, file.Bucket, thumbnailKey); err != nil {
			s.logger.Warn("Failed to delete thumbnail",
				zap.String("file_id", file.ID.String()),
				zap.String("key", thumbnailKey),
				zap.Error(err),
			)
		}
	}

	// 清除文件的缩略图键
	file.ThumbnailKey = nil
	if err := s.fileRepo.Update(ctx, file); err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	return nil
}

// IsImageFile 检查是否为图片文件
func (s *thumbnailService) IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, supportedExt := range SupportedImageFormats {
		if ext == supportedExt {
			return true
		}
	}
	return false
}
