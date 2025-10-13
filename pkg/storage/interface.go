package storage

import (
	"context"
	"io"
	"time"
)

// ObjectStorage 对象存储抽象接口
// 符合 S3 兼容协议，支持多云存储后端（MinIO、AWS S3、阿里云 OSS、腾讯云 COS 等）
type ObjectStorage interface {
	// Bucket 操作
	BucketOperations

	// Object 基础操作
	ObjectOperations

	// Multipart 分片上传操作
	MultipartOperations

	// Presigned URL 操作
	PresignedOperations

	// 生命周期管理
	Close() error
}

// BucketOperations 存储桶操作接口
type BucketOperations interface {
	// MakeBucket 创建存储桶
	MakeBucket(ctx context.Context, bucketName string, opts MakeBucketOptions) error

	// RemoveBucket 删除存储桶
	RemoveBucket(ctx context.Context, bucketName string) error

	// BucketExists 检查存储桶是否存在
	BucketExists(ctx context.Context, bucketName string) (bool, error)

	// ListBuckets 列出所有存储桶
	ListBuckets(ctx context.Context) ([]BucketInfo, error)

	// SetBucketPolicy 设置存储桶策略
	SetBucketPolicy(ctx context.Context, bucketName string, policy string) error

	// GetBucketPolicy 获取存储桶策略
	GetBucketPolicy(ctx context.Context, bucketName string) (string, error)
}

// ObjectOperations 对象基础操作接口
type ObjectOperations interface {
	// PutObject 上传对象
	PutObject(ctx context.Context, bucketName string, objectName string, reader io.Reader, size int64, opts PutObjectOptions) (*PutObjectResult, error)

	// GetObject 获取对象
	GetObject(ctx context.Context, bucketName string, objectName string, opts GetObjectOptions) (io.ReadCloser, *ObjectInfo, error)

	// StatObject 获取对象信息
	StatObject(ctx context.Context, bucketName string, objectName string) (*ObjectInfo, error)

	// RemoveObject 删除对象
	RemoveObject(ctx context.Context, bucketName string, objectName string) error

	// RemoveObjects 批量删除对象
	RemoveObjects(ctx context.Context, bucketName string, objectNames []string) ([]RemoveObjectError, error)

	// CopyObject 复制对象
	CopyObject(ctx context.Context, destBucket, destObject, srcBucket, srcObject string, opts CopyObjectOptions) (*CopyObjectResult, error)

	// ListObjects 列出对象
	ListObjects(ctx context.Context, bucketName string, opts ListObjectsOptions) ([]ObjectInfo, error)
}

// MultipartOperations 分片上传操作接口
type MultipartOperations interface {
	// NewMultipartUpload 初始化分片上传
	NewMultipartUpload(ctx context.Context, bucketName string, objectName string, opts PutObjectOptions) (string, error)

	// PutObjectPart 上传分片
	PutObjectPart(ctx context.Context, bucketName string, objectName string, uploadID string, partNumber int, reader io.Reader, size int64, opts PutObjectPartOptions) (*ObjectPart, error)

	// CompleteMultipartUpload 完成分片上传
	CompleteMultipartUpload(ctx context.Context, bucketName string, objectName string, uploadID string, parts []CompletePart) (*CompleteMultipartResult, error)

	// AbortMultipartUpload 中止分片上传
	AbortMultipartUpload(ctx context.Context, bucketName string, objectName string, uploadID string) error

	// ListMultipartUploads 列出正在进行的分片上传
	ListMultipartUploads(ctx context.Context, bucketName string, opts ListMultipartUploadsOptions) ([]MultipartUploadInfo, error)

	// ListObjectParts 列出已上传的分片
	ListObjectParts(ctx context.Context, bucketName string, objectName string, uploadID string, opts ListObjectPartsOptions) ([]ObjectPart, error)
}

// PresignedOperations 预签名 URL 操作接口
type PresignedOperations interface {
	// PresignedGetObject 生成下载预签名 URL
	PresignedGetObject(ctx context.Context, bucketName string, objectName string, expires time.Duration, opts PresignedGetOptions) (string, error)

	// PresignedPutObject 生成上传预签名 URL
	PresignedPutObject(ctx context.Context, bucketName string, objectName string, expires time.Duration) (string, error)

	// PresignedPostPolicy 生成表单上传预签名策略
	PresignedPostPolicy(ctx context.Context, policy *PostPolicy) (map[string]string, error)
}

// ====================== 选项和参数定义 ======================

// MakeBucketOptions 创建存储桶选项
type MakeBucketOptions struct {
	Region        string            // 区域
	ObjectLocking bool              // 是否启用对象锁定
	Tags          map[string]string // 标签
}

// PutObjectOptions 上传对象选项
type PutObjectOptions struct {
	ContentType          string                // Content-Type
	ContentEncoding      string                // Content-Encoding
	ContentDisposition   string                // Content-Disposition
	CacheControl         string                // Cache-Control
	UserMetadata         map[string]string     // 用户自定义元数据
	ServerSideEncryption *ServerSideEncryption // 服务端加密
	StorageClass         string                // 存储类型
	Tags                 map[string]string     // 对象标签
}

// GetObjectOptions 获取对象选项
type GetObjectOptions struct {
	Range             *RangeOption // 字节范围
	VersionID         string       // 版本 ID
	IfModifiedSince   *time.Time   // 仅在修改后返回
	IfUnmodifiedSince *time.Time   // 仅在未修改时返回
	IfMatch           string       // ETag 匹配
	IfNoneMatch       string       // ETag 不匹配
}

// RangeOption 字节范围选项
type RangeOption struct {
	Start int64 // 起始字节（包含）
	End   int64 // 结束字节（包含），-1 表示到文件末尾
}

// CopyObjectOptions 复制对象选项
type CopyObjectOptions struct {
	SourceVersionID      string                // 源版本 ID
	UserMetadata         map[string]string     // 新的用户元数据
	ReplaceMetadata      bool                  // 是否替换元数据
	ServerSideEncryption *ServerSideEncryption // 服务端加密
}

// ListObjectsOptions 列出对象选项
type ListObjectsOptions struct {
	Prefix          string // 前缀过滤
	Delimiter       string // 分隔符
	MaxKeys         int    // 最大返回数量
	Marker          string // 起始标记（分页）
	Recursive       bool   // 是否递归列出
	StartAfter      string // 从指定键之后开始
	VersionIDMarker string // 版本 ID 标记
}

// PutObjectPartOptions 上传分片选项
type PutObjectPartOptions struct {
	ServerSideEncryption *ServerSideEncryption
	ContentMD5           string // MD5 校验
}

// CompletePart 完成分片信息
type CompletePart struct {
	PartNumber int    // 分片编号
	ETag       string // ETag
}

// ListMultipartUploadsOptions 列出分片上传选项
type ListMultipartUploadsOptions struct {
	Prefix         string // 前缀过滤
	KeyMarker      string // 键标记
	UploadIDMarker string // 上传 ID 标记
	MaxUploads     int    // 最大返回数量
}

// ListObjectPartsOptions 列出分片选项
type ListObjectPartsOptions struct {
	MaxParts         int // 最大返回数量
	PartNumberMarker int // 分片编号标记
}

// PresignedGetOptions 预签名下载选项
type PresignedGetOptions struct {
	RequestParams map[string]string // 请求参数
}

// ServerSideEncryption 服务端加密配置
type ServerSideEncryption struct {
	Type      string // 加密类型: "AES256", "aws:kms"
	KeyID     string // KMS 密钥 ID
	Algorithm string // 加密算法
}

// PostPolicy 表单上传策略
type PostPolicy struct {
	BucketName string
	ObjectName string
	Expiration time.Time
	Conditions []interface{}
}

// ====================== 结果类型定义 ======================

// PutObjectResult 上传对象结果
type PutObjectResult struct {
	Bucket       string    // 存储桶名
	Key          string    // 对象键
	ETag         string    // ETag
	Size         int64     // 大小
	LastModified time.Time // 最后修改时间
	VersionID    string    // 版本 ID
}

// ObjectInfo 对象信息
type ObjectInfo struct {
	Key          string            // 对象键
	Size         int64             // 大小
	ETag         string            // ETag
	ContentType  string            // Content-Type
	LastModified time.Time         // 最后修改时间
	Metadata     map[string]string // 元数据
	VersionID    string            // 版本 ID
	StorageClass string            // 存储类型
	IsDir        bool              // 是否为目录
}

// ObjectPart 对象分片信息
type ObjectPart struct {
	PartNumber   int       // 分片编号
	ETag         string    // ETag
	Size         int64     // 大小
	LastModified time.Time // 最后修改时间
}

// CompleteMultipartResult 完成分片上传结果
type CompleteMultipartResult struct {
	Bucket    string // 存储桶名
	Key       string // 对象键
	ETag      string // ETag
	VersionID string // 版本 ID
}

// CopyObjectResult 复制对象结果
type CopyObjectResult struct {
	ETag         string    // ETag
	LastModified time.Time // 最后修改时间
	VersionID    string    // 版本 ID
}

// MultipartUploadInfo 分片上传信息
type MultipartUploadInfo struct {
	Key          string    // 对象键
	UploadID     string    // 上传 ID
	Initiated    time.Time // 发起时间
	StorageClass string    // 存储类型
}

// RemoveObjectError 删除对象错误
type RemoveObjectError struct {
	ObjectName string // 对象名
	Error      error  // 错误信息
}

// BucketInfo 存储桶信息
type BucketInfo struct {
	Name         string    // 名称
	CreationDate time.Time // 创建时间
	Region       string    // 区域
}
