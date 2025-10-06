package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// Storage MinIO 对象存储接口
type Storage interface {
	// 文件操作
	UploadFile(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error
	DownloadFile(ctx context.Context, bucket, key, filePath string) error
	DeleteFile(ctx context.Context, bucket, key string) error
	GetPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
	ListFiles(ctx context.Context, bucket, prefix string) ([]FileInfo, error)
	FileExists(ctx context.Context, bucket, key string) (bool, error)
	GetFileInfo(ctx context.Context, bucket, key string) (*FileInfo, error)

	// 桶管理
	CreateBucket(ctx context.Context, bucket string) error
	DeleteBucket(ctx context.Context, bucket string) error
	BucketExists(ctx context.Context, bucket string) (bool, error)
	ListBuckets(ctx context.Context) ([]BucketInfo, error)

	// 关闭连接
	Close() error
}

// FileInfo 文件信息
type FileInfo struct {
	Key          string
	Size         int64
	ETag         string
	ContentType  string
	LastModified time.Time
}

// BucketInfo 存储桶信息
type BucketInfo struct {
	Name         string
	CreationDate time.Time
}

// storage MinIO 存储实现
type storage struct {
	client *minio.Client
	config *Config
	logger *logger.Logger
}

// New 创建 MinIO 存储客户端
func New(ctx context.Context, opts ...Option) (Storage, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 初始化 MinIO 客户端
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	s := &storage{
		client: client,
		config: cfg,
		logger: logger.GetLogger().With(zap.String("module", "storage")),
	}

	// 确保默认桶存在
	exists, err := s.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		if err := s.CreateBucket(ctx, cfg.BucketName); err != nil {
			return nil, fmt.Errorf("failed to create default bucket: %w", err)
		}
	}

	s.logger.Info("MinIO storage initialized",
		zap.String("endpoint", cfg.Endpoint),
		zap.String("bucket", cfg.BucketName),
		zap.Bool("ssl", cfg.UseSSL),
	)

	return s, nil
}

// UploadFile 上传文件
func (s *storage) UploadFile(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	if bucket == "" {
		bucket = s.config.BucketName
	}

	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	info, err := s.client.PutObject(ctx, bucket, key, reader, size, opts)
	if err != nil {
		s.logger.Error("Failed to upload file",
			zap.String("bucket", bucket),
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to upload file: %w", err)
	}

	s.logger.Debug("File uploaded successfully",
		zap.String("bucket", bucket),
		zap.String("key", key),
		zap.Int64("size", info.Size),
	)

	return nil
}

// DownloadFile 下载文件到本地
func (s *storage) DownloadFile(ctx context.Context, bucket, key, filePath string) error {
	if bucket == "" {
		bucket = s.config.BucketName
	}

	err := s.client.FGetObject(ctx, bucket, key, filePath, minio.GetObjectOptions{})
	if err != nil {
		s.logger.Error("Failed to download file",
			zap.String("bucket", bucket),
			zap.String("key", key),
			zap.String("file_path", filePath),
			zap.Error(err),
		)
		return fmt.Errorf("failed to download file: %w", err)
	}

	s.logger.Debug("File downloaded successfully",
		zap.String("bucket", bucket),
		zap.String("key", key),
		zap.String("file_path", filePath),
	)

	return nil
}

// DeleteFile 删除文件
func (s *storage) DeleteFile(ctx context.Context, bucket, key string) error {
	if bucket == "" {
		bucket = s.config.BucketName
	}

	err := s.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		s.logger.Error("Failed to delete file",
			zap.String("bucket", bucket),
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	s.logger.Debug("File deleted successfully",
		zap.String("bucket", bucket),
		zap.String("key", key),
	)

	return nil
}

// GetPresignedURL 获取预签名 URL
func (s *storage) GetPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	if bucket == "" {
		bucket = s.config.BucketName
	}

	url, err := s.client.PresignedGetObject(ctx, bucket, key, expiry, nil)
	if err != nil {
		s.logger.Error("Failed to generate presigned URL",
			zap.String("bucket", bucket),
			zap.String("key", key),
			zap.Duration("expiry", expiry),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// ListFiles 列出文件
func (s *storage) ListFiles(ctx context.Context, bucket, prefix string) ([]FileInfo, error) {
	if bucket == "" {
		bucket = s.config.BucketName
	}

	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}

	files := []FileInfo{}
	for object := range s.client.ListObjects(ctx, bucket, opts) {
		if object.Err != nil {
			s.logger.Error("Error listing objects",
				zap.String("bucket", bucket),
				zap.Error(object.Err),
			)
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}

		files = append(files, FileInfo{
			Key:          object.Key,
			Size:         object.Size,
			ETag:         object.ETag,
			ContentType:  object.ContentType,
			LastModified: object.LastModified,
		})
	}

	return files, nil
}

// FileExists 检查文件是否存在
func (s *storage) FileExists(ctx context.Context, bucket, key string) (bool, error) {
	if bucket == "" {
		bucket = s.config.BucketName
	}

	_, err := s.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// GetFileInfo 获取文件信息
func (s *storage) GetFileInfo(ctx context.Context, bucket, key string) (*FileInfo, error) {
	if bucket == "" {
		bucket = s.config.BucketName
	}

	stat, err := s.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &FileInfo{
		Key:          key,
		Size:         stat.Size,
		ETag:         stat.ETag,
		ContentType:  stat.ContentType,
		LastModified: stat.LastModified,
	}, nil
}

// CreateBucket 创建存储桶
func (s *storage) CreateBucket(ctx context.Context, bucket string) error {
	err := s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{
		Region: s.config.Region,
	})
	if err != nil {
		exists, errBucketExists := s.BucketExists(ctx, bucket)
		if errBucketExists == nil && exists {
			s.logger.Debug("Bucket already exists", zap.String("bucket", bucket))
			return nil
		}

		s.logger.Error("Failed to create bucket",
			zap.String("bucket", bucket),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	s.logger.Info("Bucket created successfully", zap.String("bucket", bucket))
	return nil
}

// DeleteBucket 删除存储桶
func (s *storage) DeleteBucket(ctx context.Context, bucket string) error {
	err := s.client.RemoveBucket(ctx, bucket)
	if err != nil {
		s.logger.Error("Failed to delete bucket",
			zap.String("bucket", bucket),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	s.logger.Info("Bucket deleted successfully", zap.String("bucket", bucket))
	return nil
}

// BucketExists 检查存储桶是否存在
func (s *storage) BucketExists(ctx context.Context, bucket string) (bool, error) {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	return exists, nil
}

// ListBuckets 列出所有存储桶
func (s *storage) ListBuckets(ctx context.Context) ([]BucketInfo, error) {
	buckets, err := s.client.ListBuckets(ctx)
	if err != nil {
		s.logger.Error("Failed to list buckets", zap.Error(err))
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	result := make([]BucketInfo, len(buckets))
	for i, bucket := range buckets {
		result[i] = BucketInfo{
			Name:         bucket.Name,
			CreationDate: bucket.CreationDate,
		}
	}

	return result, nil
}

// Close 关闭连接
func (s *storage) Close() error {
	s.logger.Info("MinIO storage closed")
	return nil
}
