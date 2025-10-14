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

// CompressionOptions 压缩选项
type CompressionOptions struct {
	Quality    int  // 压缩质量 (1-100)
	MaxWidth   int  // 最大宽度（像素）
	MaxHeight  int  // 最大高度（像素）
	KeepAspect bool // 保持宽高比
}

// DefaultCompressionOptions 默认压缩选项
var DefaultCompressionOptions = CompressionOptions{
	Quality:    85,
	MaxWidth:   1920,
	MaxHeight:  1080,
	KeepAspect: true,
}

// CompressionResult 压缩结果
type CompressionResult struct {
	OriginalSize    int64   // 原始大小
	CompressedSize  int64   // 压缩后大小
	CompressionRate float64 // 压缩率
	StorageKey      string  // 新的存储键
}

// CompressionService 文件压缩服务接口
type CompressionService interface {
	// CompressImage 压缩图片
	CompressImage(ctx context.Context, file *model.File, opts CompressionOptions) (*CompressionResult, error)

	// CompressImageInPlace 就地压缩图片（替换原文件）
	CompressImageInPlace(ctx context.Context, fileID uuid.UUID, opts CompressionOptions) (*CompressionResult, error)

	// BatchCompressImages 批量压缩图片
	BatchCompressImages(ctx context.Context, fileIDs []uuid.UUID, opts CompressionOptions) (map[uuid.UUID]*CompressionResult, error)

	// IsCompressible 检查文件是否可压缩
	IsCompressible(filename string) bool

	// EstimateCompression 估算压缩大小
	EstimateCompression(ctx context.Context, fileID uuid.UUID, opts CompressionOptions) (int64, error)
}

type compressionService struct {
	storage  storage.Storage
	fileRepo repository.FileRepository
	logger   *logger.Logger
}

// NewCompressionService 创建压缩服务
func NewCompressionService(
	storage storage.Storage,
	fileRepo repository.FileRepository,
	logger *logger.Logger,
) CompressionService {
	return &compressionService{
		storage:  storage,
		fileRepo: fileRepo,
		logger:   logger,
	}
}

// CompressImage 压缩图片
func (s *compressionService) CompressImage(ctx context.Context, file *model.File, opts CompressionOptions) (*CompressionResult, error) {
	// 检查是否为图片文件
	if !s.IsCompressible(file.Filename) {
		return nil, fmt.Errorf("file is not compressible: %s", file.Filename)
	}

	// 从存储中获取原始图片
	objectStorage := s.storage.GetObjectStorage()
	reader, _, err := objectStorage.GetObject(ctx, file.Bucket, file.StorageKey, storage.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file from storage: %w", err)
	}
	defer reader.Close()

	// 读取图片数据
	imgData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	originalSize := int64(len(imgData))

	// 解码图片
	img, format, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	s.logger.Info("Compressing image",
		zap.String("file_id", file.ID.String()),
		zap.String("format", format),
		zap.Int64("original_size", originalSize),
	)

	// 如果图片尺寸超过最大限制，进行缩放
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if (opts.MaxWidth > 0 && width > opts.MaxWidth) || (opts.MaxHeight > 0 && height > opts.MaxHeight) {
		if opts.KeepAspect {
			img = imaging.Fit(img, opts.MaxWidth, opts.MaxHeight, imaging.Lanczos)
		} else {
			img = imaging.Resize(img, opts.MaxWidth, opts.MaxHeight, imaging.Lanczos)
		}
		s.logger.Info("Image resized",
			zap.Int("original_width", width),
			zap.Int("original_height", height),
			zap.Int("new_width", img.Bounds().Dx()),
			zap.Int("new_height", img.Bounds().Dy()),
		)
	}

	// 压缩图片
	var buf bytes.Buffer
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: opts.Quality}); err != nil {
			return nil, fmt.Errorf("failed to encode jpeg: %w", err)
		}
	case "png":
		// PNG 使用默认压缩
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("failed to encode png: %w", err)
		}
	default:
		// 默认转换为 JPEG
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: opts.Quality}); err != nil {
			return nil, fmt.Errorf("failed to encode image: %w", err)
		}
	}

	compressedSize := int64(buf.Len())
	compressionRate := float64(originalSize-compressedSize) / float64(originalSize) * 100

	// 生成新的存储键
	ext := filepath.Ext(file.Filename)
	baseName := strings.TrimSuffix(file.StorageKey, ext)
	newStorageKey := fmt.Sprintf("%s_compressed%s", baseName, ext)

	// 上传压缩后的文件
	_, err = objectStorage.PutObject(ctx, file.Bucket, newStorageKey, &buf, compressedSize, storage.PutObjectOptions{
		ContentType: fmt.Sprintf("image/%s", format),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload compressed file: %w", err)
	}

	result := &CompressionResult{
		OriginalSize:    originalSize,
		CompressedSize:  compressedSize,
		CompressionRate: compressionRate,
		StorageKey:      newStorageKey,
	}

	s.logger.Info("Image compressed successfully",
		zap.String("file_id", file.ID.String()),
		zap.Int64("original_size", originalSize),
		zap.Int64("compressed_size", compressedSize),
		zap.Float64("compression_rate", compressionRate),
	)

	return result, nil
}

// CompressImageInPlace 就地压缩图片（替换原文件）
func (s *compressionService) CompressImageInPlace(ctx context.Context, fileID uuid.UUID, opts CompressionOptions) (*CompressionResult, error) {
	// 获取文件
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to find file: %w", err)
	}

	// 压缩图片
	result, err := s.CompressImage(ctx, file, opts)
	if err != nil {
		return nil, err
	}

	objectStorage := s.storage.GetObjectStorage()

	// 删除旧文件
	oldKey := file.StorageKey
	if err := objectStorage.RemoveObject(ctx, file.Bucket, oldKey); err != nil {
		s.logger.Warn("Failed to delete old file",
			zap.String("file_id", fileID.String()),
			zap.String("old_key", oldKey),
			zap.Error(err),
		)
	}

	// 更新文件记录
	file.StorageKey = result.StorageKey
	file.Size = result.CompressedSize
	file.IsCompressed = true

	if err := s.fileRepo.Update(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to update file record: %w", err)
	}

	s.logger.Info("File compressed in place",
		zap.String("file_id", fileID.String()),
		zap.String("new_key", result.StorageKey),
	)

	return result, nil
}

// BatchCompressImages 批量压缩图片
func (s *compressionService) BatchCompressImages(ctx context.Context, fileIDs []uuid.UUID, opts CompressionOptions) (map[uuid.UUID]*CompressionResult, error) {
	results := make(map[uuid.UUID]*CompressionResult)

	for _, fileID := range fileIDs {
		result, err := s.CompressImageInPlace(ctx, fileID, opts)
		if err != nil {
			s.logger.Error("Failed to compress file",
				zap.String("file_id", fileID.String()),
				zap.Error(err),
			)
			continue
		}
		results[fileID] = result
	}

	s.logger.Info("Batch compression completed",
		zap.Int("total", len(fileIDs)),
		zap.Int("success", len(results)),
		zap.Int("failed", len(fileIDs)-len(results)),
	)

	return results, nil
}

// IsCompressible 检查文件是否可压缩
func (s *compressionService) IsCompressible(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	compressibleFormats := []string{".jpg", ".jpeg", ".png", ".bmp", ".webp"}
	for _, format := range compressibleFormats {
		if ext == format {
			return true
		}
	}
	return false
}

// EstimateCompression 估算压缩大小
func (s *compressionService) EstimateCompression(ctx context.Context, fileID uuid.UUID, opts CompressionOptions) (int64, error) {
	// 获取文件
	file, err := s.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return 0, fmt.Errorf("failed to find file: %w", err)
	}

	// 简单估算：根据质量参数估算
	// 这只是粗略估算，实际压缩率取决于图片内容
	qualityFactor := float64(opts.Quality) / 100.0
	estimatedSize := int64(float64(file.Size) * qualityFactor)

	// 如果需要缩放，进一步减小估算大小
	if opts.MaxWidth > 0 || opts.MaxHeight > 0 {
		// 假设缩放后大小减少一半（粗略估算）
		estimatedSize = estimatedSize / 2
	}

	return estimatedSize, nil
}
