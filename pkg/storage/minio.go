package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/encrypt"
	"go.uber.org/zap"
)

// minioStorage MinIO 存储实现
type minioStorage struct {
	client *minio.Client
	core   *minio.Core
	config *Config
	logger *logger.Logger
}

// NewMinIOStorage 创建 MinIO 存储客户端
func NewMinIOStorage(ctx context.Context, config *Config, log *logger.Logger) (ObjectStorage, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 初始化 MinIO 客户端
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	s := &minioStorage{
		client: client,
		core:   &minio.Core{Client: client},
		config: config,
		logger: log.With(zap.String("module", "storage-minio")),
	}

	// 确保默认桶存在
	if config.BucketName != "" {
		exists, err := s.BucketExists(ctx, config.BucketName)
		if err != nil {
			return nil, fmt.Errorf("failed to check bucket existence: %w", err)
		}

		if !exists {
			opts := MakeBucketOptions{Region: config.Region}
			if err := s.MakeBucket(ctx, config.BucketName, opts); err != nil {
				return nil, fmt.Errorf("failed to create default bucket: %w", err)
			}
		}
	}

	s.logger.Info("MinIO storage initialized",
		zap.String("endpoint", config.Endpoint),
		zap.String("bucket", config.BucketName),
		zap.Bool("ssl", config.UseSSL),
	)

	return s, nil
}

// ====================== Bucket 操作实现 ======================

// MakeBucket 创建存储桶
func (s *minioStorage) MakeBucket(ctx context.Context, bucketName string, opts MakeBucketOptions) error {
	err := s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		Region:        opts.Region,
		ObjectLocking: opts.ObjectLocking,
	})
	if err != nil {
		// 检查是否已存在
		exists, errExists := s.BucketExists(ctx, bucketName)
		if errExists == nil && exists {
			s.logger.Debug("Bucket already exists", zap.String("bucket", bucketName))
			return nil
		}

		s.logger.Error("Failed to create bucket",
			zap.String("bucket", bucketName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	s.logger.Info("Bucket created successfully", zap.String("bucket", bucketName))
	return nil
}

// RemoveBucket 删除存储桶
func (s *minioStorage) RemoveBucket(ctx context.Context, bucketName string) error {
	err := s.client.RemoveBucket(ctx, bucketName)
	if err != nil {
		s.logger.Error("Failed to remove bucket",
			zap.String("bucket", bucketName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to remove bucket: %w", err)
	}

	s.logger.Info("Bucket removed successfully", zap.String("bucket", bucketName))
	return nil
}

// BucketExists 检查存储桶是否存在
func (s *minioStorage) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	return exists, nil
}

// ListBuckets 列出所有存储桶
func (s *minioStorage) ListBuckets(ctx context.Context) ([]BucketInfo, error) {
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

// SetBucketPolicy 设置存储桶策略
func (s *minioStorage) SetBucketPolicy(ctx context.Context, bucketName string, policy string) error {
	err := s.client.SetBucketPolicy(ctx, bucketName, policy)
	if err != nil {
		s.logger.Error("Failed to set bucket policy",
			zap.String("bucket", bucketName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}

	s.logger.Info("Bucket policy set successfully", zap.String("bucket", bucketName))
	return nil
}

// GetBucketPolicy 获取存储桶策略
func (s *minioStorage) GetBucketPolicy(ctx context.Context, bucketName string) (string, error) {
	policy, err := s.client.GetBucketPolicy(ctx, bucketName)
	if err != nil {
		return "", fmt.Errorf("failed to get bucket policy: %w", err)
	}
	return policy, nil
}

// ====================== Object 基础操作实现 ======================

// PutObject 上传对象
func (s *minioStorage) PutObject(ctx context.Context, bucketName string, objectName string, reader io.Reader, size int64, opts PutObjectOptions) (*PutObjectResult, error) {
	minioOpts := minio.PutObjectOptions{
		ContentType:        opts.ContentType,
		ContentEncoding:    opts.ContentEncoding,
		ContentDisposition: opts.ContentDisposition,
		CacheControl:       opts.CacheControl,
		UserMetadata:       opts.UserMetadata,
		StorageClass:       opts.StorageClass,
		UserTags:           opts.Tags,
	}

	// 设置服务端加密
	if opts.ServerSideEncryption != nil {
		minioOpts.ServerSideEncryption = s.convertSSE(opts.ServerSideEncryption)
	}

	info, err := s.client.PutObject(ctx, bucketName, objectName, reader, size, minioOpts)
	if err != nil {
		s.logger.Error("Failed to put object",
			zap.String("bucket", bucketName),
			zap.String("object", objectName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to put object: %w", err)
	}

	return &PutObjectResult{
		Bucket:       info.Bucket,
		Key:          info.Key,
		ETag:         info.ETag,
		Size:         info.Size,
		LastModified: info.LastModified,
		VersionID:    info.VersionID,
	}, nil
}

// GetObject 获取对象
func (s *minioStorage) GetObject(ctx context.Context, bucketName string, objectName string, opts GetObjectOptions) (io.ReadCloser, *ObjectInfo, error) {
	minioOpts := minio.GetObjectOptions{}

	// 设置范围
	if opts.Range != nil {
		if opts.Range.End == -1 {
			minioOpts.SetRange(opts.Range.Start, 0)
		} else {
			minioOpts.SetRange(opts.Range.Start, opts.Range.End)
		}
	}

	// 设置条件
	if opts.IfModifiedSince != nil {
		minioOpts.SetModified(*opts.IfModifiedSince)
	}
	if opts.IfUnmodifiedSince != nil {
		minioOpts.SetUnmodified(*opts.IfUnmodifiedSince)
	}
	if opts.IfMatch != "" {
		minioOpts.SetMatchETag(opts.IfMatch)
	}
	if opts.IfNoneMatch != "" {
		minioOpts.SetMatchETagExcept(opts.IfNoneMatch)
	}

	object, err := s.client.GetObject(ctx, bucketName, objectName, minioOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object: %w", err)
	}

	// 获取对象信息
	stat, err := object.Stat()
	if err != nil {
		object.Close()
		return nil, nil, fmt.Errorf("failed to stat object: %w", err)
	}

	info := &ObjectInfo{
		Key:          stat.Key,
		Size:         stat.Size,
		ETag:         stat.ETag,
		ContentType:  stat.ContentType,
		LastModified: stat.LastModified,
		Metadata:     stat.UserMetadata,
		VersionID:    stat.VersionID,
		StorageClass: stat.StorageClass,
	}

	return object, info, nil
}

// StatObject 获取对象信息
func (s *minioStorage) StatObject(ctx context.Context, bucketName string, objectName string) (*ObjectInfo, error) {
	stat, err := s.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}

	return &ObjectInfo{
		Key:          stat.Key,
		Size:         stat.Size,
		ETag:         stat.ETag,
		ContentType:  stat.ContentType,
		LastModified: stat.LastModified,
		Metadata:     stat.UserMetadata,
		VersionID:    stat.VersionID,
		StorageClass: stat.StorageClass,
	}, nil
}

// RemoveObject 删除对象
func (s *minioStorage) RemoveObject(ctx context.Context, bucketName string, objectName string) error {
	err := s.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		s.logger.Error("Failed to remove object",
			zap.String("bucket", bucketName),
			zap.String("object", objectName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to remove object: %w", err)
	}

	s.logger.Debug("Object removed successfully",
		zap.String("bucket", bucketName),
		zap.String("object", objectName))

	return nil
}

// RemoveObjects 批量删除对象
func (s *minioStorage) RemoveObjects(ctx context.Context, bucketName string, objectNames []string) ([]RemoveObjectError, error) {
	// 创建对象通道
	objectsCh := make(chan minio.ObjectInfo, len(objectNames))
	for _, name := range objectNames {
		objectsCh <- minio.ObjectInfo{Key: name}
	}
	close(objectsCh)

	// 批量删除
	opts := minio.RemoveObjectsOptions{}
	errorsCh := s.client.RemoveObjects(ctx, bucketName, objectsCh, opts)

	var errors []RemoveObjectError
	for err := range errorsCh {
		if err.Err != nil {
			errors = append(errors, RemoveObjectError{
				ObjectName: err.ObjectName,
				Error:      err.Err,
			})
		}
	}

	if len(errors) > 0 {
		s.logger.Warn("Some objects failed to delete",
			zap.String("bucket", bucketName),
			zap.Int("failed_count", len(errors)))
	}

	return errors, nil
}

// CopyObject 复制对象
func (s *minioStorage) CopyObject(ctx context.Context, destBucket, destObject, srcBucket, srcObject string, opts CopyObjectOptions) (*CopyObjectResult, error) {
	srcOpts := minio.CopySrcOptions{
		Bucket:    srcBucket,
		Object:    srcObject,
		VersionID: opts.SourceVersionID,
	}

	destOpts := minio.CopyDestOptions{
		Bucket: destBucket,
		Object: destObject,
	}

	if opts.ReplaceMetadata {
		destOpts.UserMetadata = opts.UserMetadata
		destOpts.ReplaceMetadata = true
	}

	if opts.ServerSideEncryption != nil {
		destOpts.Encryption = s.convertSSE(opts.ServerSideEncryption)
	}

	info, err := s.client.CopyObject(ctx, destOpts, srcOpts)
	if err != nil {
		s.logger.Error("Failed to copy object",
			zap.String("src_bucket", srcBucket),
			zap.String("src_object", srcObject),
			zap.String("dest_bucket", destBucket),
			zap.String("dest_object", destObject),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to copy object: %w", err)
	}

	return &CopyObjectResult{
		ETag:         info.ETag,
		LastModified: info.LastModified,
		VersionID:    info.VersionID,
	}, nil
}

// ListObjects 列出对象
func (s *minioStorage) ListObjects(ctx context.Context, bucketName string, opts ListObjectsOptions) ([]ObjectInfo, error) {
	listOpts := minio.ListObjectsOptions{
		Prefix:    opts.Prefix,
		Recursive: opts.Recursive,
		MaxKeys:   opts.MaxKeys,
	}

	var objects []ObjectInfo
	for object := range s.client.ListObjects(ctx, bucketName, listOpts) {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}

		objects = append(objects, ObjectInfo{
			Key:          object.Key,
			Size:         object.Size,
			ETag:         object.ETag,
			ContentType:  object.ContentType,
			LastModified: object.LastModified,
			StorageClass: object.StorageClass,
			IsDir:        object.Key[len(object.Key)-1] == '/',
		})
	}

	return objects, nil
}

// ====================== Multipart 分片上传实现 ======================

// NewMultipartUpload 初始化分片上传
func (s *minioStorage) NewMultipartUpload(ctx context.Context, bucketName string, objectName string, opts PutObjectOptions) (string, error) {
	minioOpts := minio.PutObjectOptions{
		ContentType:  opts.ContentType,
		UserMetadata: opts.UserMetadata,
		StorageClass: opts.StorageClass,
		UserTags:     opts.Tags,
	}

	if opts.ServerSideEncryption != nil {
		minioOpts.ServerSideEncryption = s.convertSSE(opts.ServerSideEncryption)
	}

	uploadID, err := s.core.NewMultipartUpload(ctx, bucketName, objectName, minioOpts)
	if err != nil {
		s.logger.Error("Failed to initiate multipart upload",
			zap.String("bucket", bucketName),
			zap.String("object", objectName),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to initiate multipart upload: %w", err)
	}

	s.logger.Info("Multipart upload initiated",
		zap.String("bucket", bucketName),
		zap.String("object", objectName),
		zap.String("upload_id", uploadID))

	return uploadID, nil
}

// PutObjectPart 上传分片
func (s *minioStorage) PutObjectPart(ctx context.Context, bucketName string, objectName string, uploadID string, partNumber int, reader io.Reader, size int64, opts PutObjectPartOptions) (*ObjectPart, error) {
	minioOpts := minio.PutObjectPartOptions{}

	if opts.ServerSideEncryption != nil {
		minioOpts.SSE = s.convertSSE(opts.ServerSideEncryption)
	}

	part, err := s.core.PutObjectPart(ctx, bucketName, objectName, uploadID, partNumber, reader, size, minioOpts)
	if err != nil {
		s.logger.Error("Failed to upload part",
			zap.String("bucket", bucketName),
			zap.String("object", objectName),
			zap.String("upload_id", uploadID),
			zap.Int("part_number", partNumber),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to upload part: %w", err)
	}

	return &ObjectPart{
		PartNumber: partNumber,
		ETag:       part.ETag,
		Size:       part.Size,
	}, nil
}

// CompleteMultipartUpload 完成分片上传
func (s *minioStorage) CompleteMultipartUpload(ctx context.Context, bucketName string, objectName string, uploadID string, parts []CompletePart) (*CompleteMultipartResult, error) {
	minioParts := make([]minio.CompletePart, len(parts))
	for i, part := range parts {
		minioParts[i] = minio.CompletePart{
			PartNumber: part.PartNumber,
			ETag:       part.ETag,
		}
	}

	upload, err := s.core.CompleteMultipartUpload(ctx, bucketName, objectName, uploadID, minioParts, minio.PutObjectOptions{})
	if err != nil {
		s.logger.Error("Failed to complete multipart upload",
			zap.String("bucket", bucketName),
			zap.String("object", objectName),
			zap.String("upload_id", uploadID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	s.logger.Info("Multipart upload completed",
		zap.String("bucket", bucketName),
		zap.String("object", objectName),
		zap.String("upload_id", uploadID))

	return &CompleteMultipartResult{
		Bucket:    upload.Bucket,
		Key:       upload.Key,
		ETag:      upload.ETag,
		VersionID: upload.VersionID,
	}, nil
}

// AbortMultipartUpload 中止分片上传
func (s *minioStorage) AbortMultipartUpload(ctx context.Context, bucketName string, objectName string, uploadID string) error {
	err := s.core.AbortMultipartUpload(ctx, bucketName, objectName, uploadID)
	if err != nil {
		s.logger.Error("Failed to abort multipart upload",
			zap.String("bucket", bucketName),
			zap.String("object", objectName),
			zap.String("upload_id", uploadID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to abort multipart upload: %w", err)
	}

	s.logger.Info("Multipart upload aborted",
		zap.String("bucket", bucketName),
		zap.String("object", objectName),
		zap.String("upload_id", uploadID))

	return nil
}

// ListMultipartUploads 列出正在进行的分片上传
func (s *minioStorage) ListMultipartUploads(ctx context.Context, bucketName string, opts ListMultipartUploadsOptions) ([]MultipartUploadInfo, error) {
	result, err := s.core.ListMultipartUploads(ctx, bucketName, opts.Prefix, opts.KeyMarker, opts.UploadIDMarker, "", opts.MaxUploads)
	if err != nil {
		return nil, fmt.Errorf("failed to list multipart uploads: %w", err)
	}

	var uploads []MultipartUploadInfo
	for _, upload := range result.Uploads {
		uploads = append(uploads, MultipartUploadInfo{
			Key:          upload.Key,
			UploadID:     upload.UploadID,
			Initiated:    upload.Initiated,
			StorageClass: upload.StorageClass,
		})
	}

	return uploads, nil
}

// ListObjectParts 列出已上传的分片
func (s *minioStorage) ListObjectParts(ctx context.Context, bucketName string, objectName string, uploadID string, opts ListObjectPartsOptions) ([]ObjectPart, error) {
	info, err := s.core.ListObjectParts(ctx, bucketName, objectName, uploadID, opts.PartNumberMarker, opts.MaxParts)
	if err != nil {
		return nil, fmt.Errorf("failed to list object parts: %w", err)
	}

	parts := make([]ObjectPart, len(info.ObjectParts))
	for i, part := range info.ObjectParts {
		parts[i] = ObjectPart{
			PartNumber:   part.PartNumber,
			ETag:         part.ETag,
			Size:         part.Size,
			LastModified: part.LastModified,
		}
	}

	return parts, nil
}

// ====================== Presigned URL 操作实现 ======================

// PresignedGetObject 生成下载预签名 URL
func (s *minioStorage) PresignedGetObject(ctx context.Context, bucketName string, objectName string, expires time.Duration, opts PresignedGetOptions) (string, error) {
	reqParams := url.Values{}
	if opts.RequestParams != nil {
		for k, v := range opts.RequestParams {
			reqParams.Set(k, v)
		}
	}

	url, err := s.client.PresignedGetObject(ctx, bucketName, objectName, expires, reqParams)
	if err != nil {
		s.logger.Error("Failed to generate presigned GET URL",
			zap.String("bucket", bucketName),
			zap.String("object", objectName),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to generate presigned GET URL: %w", err)
	}

	return url.String(), nil
}

// PresignedPutObject 生成上传预签名 URL
func (s *minioStorage) PresignedPutObject(ctx context.Context, bucketName string, objectName string, expires time.Duration) (string, error) {
	url, err := s.client.PresignedPutObject(ctx, bucketName, objectName, expires)
	if err != nil {
		s.logger.Error("Failed to generate presigned PUT URL",
			zap.String("bucket", bucketName),
			zap.String("object", objectName),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to generate presigned PUT URL: %w", err)
	}

	return url.String(), nil
}

// PresignedPostPolicy 生成表单上传预签名策略
func (s *minioStorage) PresignedPostPolicy(ctx context.Context, policy *PostPolicy) (map[string]string, error) {
	minioPolicy := minio.NewPostPolicy()
	minioPolicy.SetBucket(policy.BucketName)
	minioPolicy.SetKey(policy.ObjectName)
	minioPolicy.SetExpires(policy.Expiration)

	// TODO: 根据实际需求添加更多条件

	url, formData, err := s.client.PresignedPostPolicy(ctx, minioPolicy)
	if err != nil {
		s.logger.Error("Failed to generate presigned POST policy",
			zap.String("bucket", policy.BucketName),
			zap.String("object", policy.ObjectName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to generate presigned POST policy: %w", err)
	}

	result := make(map[string]string)
	result["url"] = url.String()
	for k, v := range formData {
		result[k] = v
	}

	return result, nil
}

// ====================== 辅助方法 ======================

// Close 关闭连接
func (s *minioStorage) Close() error {
	s.logger.Info("MinIO storage closed")
	return nil
}

// convertSSE 转换服务端加密配置
func (s *minioStorage) convertSSE(sse *ServerSideEncryption) encrypt.ServerSide {
	if sse == nil {
		return nil
	}

	switch sse.Type {
	case "AES256":
		return encrypt.NewSSE()
	case "aws:kms":
		// KMS 加密需要上下文，这里简化处理
		sse, _ := encrypt.NewSSEKMS(sse.KeyID, nil)
		return sse
	default:
		return nil
	}
}
