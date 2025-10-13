package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenerateStorageKey 生成存储键
// 格式: {tenant_id}/{year}/{month}/{uuid}.{ext}
func GenerateStorageKey(tenantID uuid.UUID, filename string) string {
	now := time.Now()
	ext := filepath.Ext(filename)
	fileUUID := uuid.New()

	return fmt.Sprintf("%s/%d/%02d/%s%s",
		tenantID.String(),
		now.Year(),
		now.Month(),
		fileUUID.String(),
		ext,
	)
}

// CalculateChecksum 计算文件 SHA-256 哈希值
func CalculateChecksum(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// DetectCategory 根据 MIME 类型检测文件分类
func DetectCategory(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	case strings.HasPrefix(mimeType, "application/pdf"):
		return "document"
	case strings.Contains(mimeType, "word"), strings.Contains(mimeType, "document"):
		return "document"
	case strings.Contains(mimeType, "spreadsheet"), strings.Contains(mimeType, "excel"):
		return "spreadsheet"
	case strings.Contains(mimeType, "presentation"), strings.Contains(mimeType, "powerpoint"):
		return "presentation"
	case strings.HasPrefix(mimeType, "text/"):
		return "text"
	case strings.Contains(mimeType, "zip"), strings.Contains(mimeType, "archive"), strings.Contains(mimeType, "compressed"):
		return "archive"
	default:
		return "other"
	}
}

// SanitizeFilename 清理文件名，移除非法字符
func SanitizeFilename(filename string) string {
	// Replace invalid characters with underscore
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(filename)
}

// GetFileExtension 获取文件扩展名（小写）
func GetFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}

// FormatFileSize 格式化文件大小为人类可读格式
func FormatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/TB)
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// CalculatePartSize 计算分片大小
// 根据文件总大小自动调整分片大小
func CalculatePartSize(totalSize int64) int64 {
	const (
		MinPartSize     = 5 * 1024 * 1024      // 5MB minimum
		MaxPartSize     = 100 * 1024 * 1024    // 100MB maximum
		MaxParts        = 10000                // S3/MinIO limit
		DefaultPartSize = 10 * 1024 * 1024     // 10MB default
	)

	if totalSize <= 0 {
		return DefaultPartSize
	}

	// Calculate ideal part size
	partSize := totalSize / MaxParts

	// Adjust to min/max bounds
	if partSize < MinPartSize {
		partSize = MinPartSize
	} else if partSize > MaxPartSize {
		partSize = MaxPartSize
	}

	return partSize
}

// IsImageFile 检查是否为图片文件
func IsImageFile(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

// IsPDFFile 检查是否为 PDF 文件
func IsPDFFile(mimeType string) bool {
	return mimeType == "application/pdf"
}

// IsVideoFile 检查是否为视频文件
func IsVideoFile(mimeType string) bool {
	return strings.HasPrefix(mimeType, "video/")
}

// IsCompressibleFile 检查文件是否适合压缩
func IsCompressibleFile(mimeType string) bool {
	// 已压缩的格式不再压缩
	uncompressible := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"video/",
		"audio/",
		"application/zip",
		"application/gzip",
		"application/x-7z-compressed",
		"application/x-rar-compressed",
	}

	for _, prefix := range uncompressible {
		if strings.HasPrefix(mimeType, prefix) {
			return false
		}
	}

	return true
}

// ValidateFileSize 验证文件大小
func ValidateFileSize(size int64, maxSize int64) error {
	if size <= 0 {
		return fmt.Errorf("file size must be greater than 0")
	}
	if maxSize > 0 && size > maxSize {
		return fmt.Errorf("file size %s exceeds maximum allowed size %s",
			FormatFileSize(size),
			FormatFileSize(maxSize),
		)
	}
	return nil
}

// ValidateFileExtension 验证文件扩展名
func ValidateFileExtension(filename string, allowedExtensions []string) error {
	if len(allowedExtensions) == 0 {
		return nil // No restrictions
	}

	ext := GetFileExtension(filename)
	if ext == "" {
		return fmt.Errorf("file has no extension")
	}

	for _, allowed := range allowedExtensions {
		if strings.EqualFold(ext, allowed) {
			return nil
		}
	}

	return fmt.Errorf("file extension %s is not allowed. Allowed: %v", ext, allowedExtensions)
}
