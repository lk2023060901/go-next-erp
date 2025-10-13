package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	filev1 "github.com/lk2023060901/go-next-erp/api/file/v1"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
	"github.com/lk2023060901/go-next-erp/internal/file/repository"
	"github.com/lk2023060901/go-next-erp/internal/file/service"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FileAdapter 文件适配器
type FileAdapter struct {
	filev1.UnimplementedFileServiceServer

	fileRepo         repository.FileRepository
	uploadService    service.UploadService
	downloadService  service.DownloadService
	quotaService     service.QuotaService
	multipartService service.MultipartUploadService
}

// NewFileAdapter 创建文件适配器
func NewFileAdapter(
	fileRepo repository.FileRepository,
	uploadService service.UploadService,
	downloadService service.DownloadService,
	quotaService service.QuotaService,
	multipartService service.MultipartUploadService,
) *FileAdapter {
	return &FileAdapter{
		fileRepo:         fileRepo,
		uploadService:    uploadService,
		downloadService:  downloadService,
		quotaService:     quotaService,
		multipartService: multipartService,
	}
}

// Upload 上传文件
func (a *FileAdapter) Upload(ctx context.Context, req *filev1.UploadRequest) (*filev1.UploadResponse, error) {
	// 从上下文获取租户ID和用户ID（应该由认证中间件注入）
	_, _, err := getAuthInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: 实际实现需要从 stream 或 multipart 获取文件内容
	// 这里暂时返回错误提示
	return nil, fmt.Errorf("file upload via gRPC requires streaming, use HTTP multipart upload instead")
}

// GetFile 获取文件信息
func (a *FileAdapter) GetFile(ctx context.Context, req *filev1.GetFileRequest) (*filev1.FileInfo, error) {
	tenantID, userID, err := getAuthInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	fileID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID: %w", err)
	}

	// 检查访问权限
	canAccess, err := a.downloadService.CheckAccess(ctx, fileID, userID, tenantID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, fmt.Errorf("access denied")
	}

	// 获取文件信息
	file, err := a.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	// 获取下载统计（如果需要）
	stats, _ := a.downloadService.GetFileDownloadStats(ctx, fileID)
	downloadCount := int32(0)
	if stats != nil {
		downloadCount = int32(stats.TotalDownloads)
	}

	return a.fileToProto(file, downloadCount), nil
}

// ListFiles 获取文件列表
func (a *FileAdapter) ListFiles(ctx context.Context, req *filev1.ListFilesRequest) (*filev1.ListFilesResponse, error) {
	tenantID, _, err := getAuthInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 构建筛选条件
	filter := &repository.FileFilter{
		TenantID: tenantID,
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	if req.Filename != "" {
		filter.SearchQuery = &req.Filename
	}
	if req.Category != "" {
		filter.Category = &req.Category
	}
	if req.MimeType != "" {
		filter.MimeType = &req.MimeType
	}
	if req.Status != "" {
		status := model.FileStatus(req.Status)
		filter.Status = &status
	}
	if req.UploadedBy != "" {
		userID, err := uuid.Parse(req.UploadedBy)
		if err == nil {
			filter.UploadedBy = &userID
		}
	}

	// 查询文件列表
	files, total, err := a.fileRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 转换为 Proto
	items := make([]*filev1.FileInfo, 0, len(files))
	for _, file := range files {
		items = append(items, a.fileToProto(file, 0))
	}

	// 计算总页数
	totalPages := int32((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	response := &filev1.ListFilesResponse{
		Files:      items,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}

	// 如果请求包含统计信息
	if req.IncludeStats {
		totalSize, _ := a.fileRepo.GetTotalSize(ctx, tenantID)
		fileCount, _ := a.fileRepo.GetFileCount(ctx, tenantID)

		response.Stats = &filev1.FileStats{
			TotalFiles:         fileCount,
			TotalSize:          totalSize,
			TotalSizeFormatted: formatBytes(totalSize),
		}
	}

	return response, nil
}

// DeleteFile 删除文件
func (a *FileAdapter) DeleteFile(ctx context.Context, req *filev1.DeleteFileRequest) (*emptypb.Empty, error) {
	tenantID, userID, err := getAuthInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	fileID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID: %w", err)
	}

	// 检查权限
	file, err := a.fileRepo.FindByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	if !file.CanAccess(userID, tenantID) {
		return nil, fmt.Errorf("access denied")
	}

	// 软删除
	if err := a.fileRepo.SoftDelete(ctx, fileID); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// GetDownloadURL 获取下载URL
func (a *FileAdapter) GetDownloadURL(ctx context.Context, req *filev1.GetDownloadURLRequest) (*filev1.DownloadURLResponse, error) {
	tenantID, userID, err := getAuthInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	fileID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID: %w", err)
	}

	// 设置有效期
	expiry := 7 * 24 * time.Hour // 默认7天
	if req.Expiry > 0 {
		expiry = time.Duration(req.Expiry) * time.Second
	}

	// 生成下载URL
	url, err := a.downloadService.GetDownloadURL(ctx, fileID, userID, tenantID, expiry)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(expiry)

	return &filev1.DownloadURLResponse{
		Url:       url,
		ExpiresAt: timestamppb.New(expiresAt),
	}, nil
}

// GetQuota 获取配额信息
func (a *FileAdapter) GetQuota(ctx context.Context, _ *emptypb.Empty) (*filev1.QuotaInfo, error) {
	tenantID, _, err := getAuthInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 获取配额使用信息
	usage, err := a.quotaService.GetQuotaUsage(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return &filev1.QuotaInfo{
		SubjectType:         "tenant",
		SubjectId:           tenantID.String(),
		QuotaLimit:          usage.QuotaLimit,
		QuotaUsed:           usage.QuotaUsed,
		QuotaAvailable:      usage.QuotaAvailable,
		QuotaUsedPercent:    usage.UsagePercent,
		QuotaLimitFormatted: formatBytes(usage.QuotaLimit),
		QuotaUsedFormatted:  formatBytes(usage.QuotaUsed),
		IsWarning:           usage.IsNearLimit,
		IsExceeded:          usage.IsExceeded,
	}, nil
}

// InitiateMultipartUpload 初始化分片上传
func (a *FileAdapter) InitiateMultipartUpload(ctx context.Context, req *filev1.InitiateMultipartRequest) (*filev1.InitiateMultipartResponse, error) {
	tenantID, userID, err := getAuthInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 构建初始化请求
	initiateReq := &service.InitiateMultipartUploadRequest{
		TenantID:    tenantID,
		UploadedBy:  userID,
		Filename:    req.Filename,
		TotalSize:   req.TotalSize,
		ContentType: req.ContentType,
		IsTemporary: req.IsTemporary,
	}

	if req.ExpiresAt != nil {
		expiresAt := req.ExpiresAt.AsTime()
		initiateReq.ExpiresAt = &expiresAt
	}

	// 初始化分片上传
	response, err := a.multipartService.InitiateUpload(ctx, initiateReq)
	if err != nil {
		return nil, err
	}

	return &filev1.InitiateMultipartResponse{
		UploadId:   response.UploadID,
		StorageKey: response.StorageKey,
		PartSize:   response.PartSize,
		TotalParts: int32(response.TotalParts),
	}, nil
}

// CompleteMultipartUpload 完成分片上传
func (a *FileAdapter) CompleteMultipartUpload(ctx context.Context, req *filev1.CompleteMultipartRequest) (*filev1.UploadResponse, error) {
	tenantID, userID, err := getAuthInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 转换分片信息
	parts := make([]service.CompletedPartInfo, len(req.Parts))
	for i, part := range req.Parts {
		parts[i] = service.CompletedPartInfo{
			PartNumber: int(part.PartNumber),
			ETag:       part.Etag,
		}
	}

	// 构建完成请求
	completeReq := &service.CompleteMultipartUploadRequest{
		UploadID: req.UploadId,
		TenantID: tenantID,
		Parts:    parts,
	}

	// 完成分片上传
	file, err := a.multipartService.CompleteUpload(ctx, completeReq)
	if err != nil {
		return nil, err
	}

	// 生成下载URL
	downloadURL, _ := a.downloadService.GetDownloadURL(ctx, file.ID, userID, tenantID, 7*24*time.Hour)

	return &filev1.UploadResponse{
		FileId:      file.ID.String(),
		Filename:    file.Filename,
		Size:        file.Size,
		MimeType:    file.MimeType,
		StorageKey:  file.StorageKey,
		Checksum:    file.Checksum,
		DownloadUrl: downloadURL,
		CreatedAt:   timestamppb.New(file.CreatedAt),
	}, nil
}

// ====================== 辅助方法 ======================

// fileToProto 转换文件模型到 Proto
func (a *FileAdapter) fileToProto(file *model.File, downloadCount int32) *filev1.FileInfo {
	info := &filev1.FileInfo{
		Id:            file.ID.String(),
		TenantId:      file.TenantID.String(),
		Filename:      file.Filename,
		OriginalName:  file.Filename,
		Size:          file.Size,
		MimeType:      file.MimeType,
		Extension:     file.Extension,
		Category:      file.Category,
		Checksum:      file.Checksum,
		StorageKey:    file.StorageKey,
		Bucket:        file.Bucket,
		Tags:          file.Tags,
		Status:        string(file.Status),
		IsTemporary:   file.IsTemporary,
		IsPublic:      file.IsPublic,
		VirusScanned:  file.VirusScanned,
		DownloadCount: downloadCount,
		VersionNumber: int32(file.VersionNumber),
		AccessLevel:   string(file.AccessLevel),
		UploadedBy:    file.UploadedBy.String(),
		UploadedAt:    timestamppb.New(file.CreatedAt),
		CreatedAt:     timestamppb.New(file.CreatedAt),
		UpdatedAt:     timestamppb.New(file.UpdatedAt),
	}

	if file.VirusScanResult != nil {
		info.VirusScanResult = string(*file.VirusScanResult)
	}

	if file.ExpiresAt != nil {
		info.ExpiresAt = timestamppb.New(*file.ExpiresAt)
	}

	if file.PreviewURL != nil {
		info.PreviewUrl = *file.PreviewURL
	}

	// 转换 metadata
	if file.Metadata != nil {
		info.Metadata = make(map[string]string)
		for k, v := range file.Metadata {
			if str, ok := v.(string); ok {
				info.Metadata[k] = str
			}
		}
	}

	return info
}

// getAuthInfoFromContext 从上下文获取认证信息
func getAuthInfoFromContext(ctx context.Context) (tenantID uuid.UUID, userID uuid.UUID, err error) {
	// TODO: 从上下文中获取实际的认证信息
	// 这里应该由认证中间件注入到 context 中

	// 临时返回测试用的 UUID
	tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID = uuid.MustParse("00000000-0000-0000-0000-000000000002")

	return tenantID, userID, nil
}

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
