package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	filev1 "github.com/lk2023060901/go-next-erp/api/file/v1"
	"github.com/lk2023060901/go-next-erp/internal/file/model"
	"github.com/lk2023060901/go-next-erp/internal/file/repository"
	"github.com/lk2023060901/go-next-erp/internal/file/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockFileRepository mocks the file repository
type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) Create(ctx context.Context, file *model.File) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockFileRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.File, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.File), args.Error(1)
}

func (m *MockFileRepository) FindByStorageKey(ctx context.Context, storageKey string) (*model.File, error) {
	args := m.Called(ctx, storageKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.File), args.Error(1)
}

func (m *MockFileRepository) FindByChecksum(ctx context.Context, checksum string, tenantID uuid.UUID) (*model.File, error) {
	args := m.Called(ctx, checksum, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.File), args.Error(1)
}

func (m *MockFileRepository) Update(ctx context.Context, file *model.File) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockFileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFileRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFileRepository) List(ctx context.Context, filter *repository.FileFilter) ([]*model.File, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.File), args.Get(1).(int64), args.Error(2)
}

func (m *MockFileRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.File, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.File), args.Error(1)
}

func (m *MockFileRepository) ListByUploader(ctx context.Context, uploaderID uuid.UUID, limit, offset int) ([]*model.File, error) {
	args := m.Called(ctx, uploaderID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.File), args.Error(1)
}

func (m *MockFileRepository) UpdateVirusScanResult(ctx context.Context, id uuid.UUID, result model.VirusScanResult) error {
	args := m.Called(ctx, id, result)
	return args.Error(0)
}

func (m *MockFileRepository) MarkAsExpired(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFileRepository) CleanTemporaryFiles(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFileRepository) GetTotalSize(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFileRepository) GetFileCount(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(int64), args.Error(1)
}

// MockDownloadService mocks the download service
type MockDownloadService struct {
	mock.Mock
}

func (m *MockDownloadService) GetDownloadURL(ctx context.Context, fileID, userID, tenantID uuid.UUID, expiry time.Duration) (string, error) {
	args := m.Called(ctx, fileID, userID, tenantID, expiry)
	return args.String(0), args.Error(1)
}

func (m *MockDownloadService) GetPreviewURL(ctx context.Context, fileID, userID, tenantID uuid.UUID, expiry time.Duration) (string, error) {
	args := m.Called(ctx, fileID, userID, tenantID, expiry)
	return args.String(0), args.Error(1)
}

func (m *MockDownloadService) CheckAccess(ctx context.Context, fileID, userID, tenantID uuid.UUID) (bool, error) {
	args := m.Called(ctx, fileID, userID, tenantID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDownloadService) GetBatchDownloadURLs(ctx context.Context, fileIDs []uuid.UUID, userID, tenantID uuid.UUID, expiry time.Duration) (map[uuid.UUID]string, error) {
	args := m.Called(ctx, fileIDs, userID, tenantID, expiry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]string), args.Error(1)
}

func (m *MockDownloadService) RecordDownload(ctx context.Context, req *service.RecordDownloadRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockDownloadService) GetFileDownloadStats(ctx context.Context, fileID uuid.UUID) (*model.FileDownloadSummary, error) {
	args := m.Called(ctx, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FileDownloadSummary), args.Error(1)
}

func (m *MockDownloadService) GetTenantDownloadStats(ctx context.Context, tenantID uuid.UUID, period string) (*model.TenantDownloadSummary, error) {
	args := m.Called(ctx, tenantID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TenantDownloadSummary), args.Error(1)
}

func (m *MockDownloadService) GetUserDownloadStats(ctx context.Context, tenantID, userID uuid.UUID, period string) (*model.UserDownloadSummary, error) {
	args := m.Called(ctx, tenantID, userID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserDownloadSummary), args.Error(1)
}

// MockQuotaService mocks the quota service
type MockQuotaService struct {
	mock.Mock
}

func (m *MockQuotaService) GetTenantQuota(ctx context.Context, tenantID uuid.UUID) (*model.StorageQuota, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StorageQuota), args.Error(1)
}

func (m *MockQuotaService) GetUserQuota(ctx context.Context, tenantID, userID uuid.UUID) (*model.StorageQuota, error) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.StorageQuota), args.Error(1)
}

func (m *MockQuotaService) UpdateQuotaLimit(ctx context.Context, quotaID uuid.UUID, newLimit int64) error {
	args := m.Called(ctx, quotaID, newLimit)
	return args.Error(0)
}

func (m *MockQuotaService) CheckQuota(ctx context.Context, tenantID uuid.UUID, size int64) (bool, error) {
	args := m.Called(ctx, tenantID, size)
	return args.Bool(0), args.Error(1)
}

func (m *MockQuotaService) CheckUserQuota(ctx context.Context, tenantID, userID uuid.UUID, size int64) (bool, error) {
	args := m.Called(ctx, tenantID, userID, size)
	return args.Bool(0), args.Error(1)
}

func (m *MockQuotaService) GetQuotaUsage(ctx context.Context, tenantID uuid.UUID) (*service.QuotaUsageInfo, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.QuotaUsageInfo), args.Error(1)
}

func (m *MockQuotaService) GetQuotaList(ctx context.Context, tenantID uuid.UUID) ([]*service.QuotaInfo, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*service.QuotaInfo), args.Error(1)
}

func (m *MockQuotaService) CheckQuotaWarning(ctx context.Context, tenantID uuid.UUID, threshold float64) (bool, string, error) {
	args := m.Called(ctx, tenantID, threshold)
	return args.Bool(0), args.String(1), args.Error(2)
}

// MockMultipartUploadService mocks the multipart upload service
type MockMultipartUploadService struct {
	mock.Mock
}

func (m *MockMultipartUploadService) InitiateUpload(ctx context.Context, req *service.InitiateMultipartUploadRequest) (*service.MultipartUploadResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.MultipartUploadResponse), args.Error(1)
}

func (m *MockMultipartUploadService) UploadPart(ctx context.Context, req *service.MultipartUploadPartRequest) (*service.UploadPartResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.UploadPartResponse), args.Error(1)
}

func (m *MockMultipartUploadService) CompleteUpload(ctx context.Context, req *service.CompleteMultipartUploadRequest) (*model.File, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.File), args.Error(1)
}

func (m *MockMultipartUploadService) AbortUpload(ctx context.Context, uploadID string, tenantID uuid.UUID) error {
	args := m.Called(ctx, uploadID, tenantID)
	return args.Error(0)
}

func (m *MockMultipartUploadService) GetUploadProgress(ctx context.Context, uploadID string) (*service.UploadProgressResponse, error) {
	args := m.Called(ctx, uploadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.UploadProgressResponse), args.Error(1)
}

func (m *MockMultipartUploadService) ListUploadedParts(ctx context.Context, uploadID string) ([]service.UploadedPartInfo, error) {
	args := m.Called(ctx, uploadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.UploadedPartInfo), args.Error(1)
}

func (m *MockMultipartUploadService) GetRemainingParts(ctx context.Context, uploadID string) ([]int, error) {
	args := m.Called(ctx, uploadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockMultipartUploadService) CleanExpiredUploads(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// TestFileAdapter_GetFile tests getting a file
func TestFileAdapter_GetFile(t *testing.T) {
	t.Run("GetFile successfully", func(t *testing.T) {
		mockFileRepo := new(MockFileRepository)
		mockDownloadSvc := new(MockDownloadService)
		mockQuotaSvc := new(MockQuotaService)
		mockMultipartSvc := new(MockMultipartUploadService)

		adapter := NewFileAdapter(mockFileRepo, nil, mockDownloadSvc, mockQuotaSvc, mockMultipartSvc)

		fileID := uuid.New()
		tenantID := uuid.New()
		userID := uuid.New()

		expectedFile := &model.File{
			ID:          fileID,
			TenantID:    tenantID,
			Filename:    "test.pdf",
			Size:        1024,
			MimeType:    "application/pdf",
			Extension:   "pdf",
			Category:    "document",
			Checksum:    "abc123",
			StorageKey:  "/files/test.pdf",
			Bucket:      "default",
			Status:      model.FileStatusActive,
			AccessLevel: model.AccessLevelPrivate,
			UploadedBy:  userID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockDownloadSvc.On("CheckAccess", mock.Anything, fileID, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).
			Return(true, nil).Once()
		mockFileRepo.On("FindByID", mock.Anything, fileID).
			Return(expectedFile, nil).Once()
		mockDownloadSvc.On("GetFileDownloadStats", mock.Anything, fileID).
			Return(&model.FileDownloadSummary{TotalDownloads: 10}, nil).Once()

		req := &filev1.GetFileRequest{
			Id: fileID.String(),
		}

		resp, err := adapter.GetFile(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, fileID.String(), resp.Id)
		assert.Equal(t, "test.pdf", resp.Filename)
		assert.Equal(t, int64(1024), resp.Size)
		assert.Equal(t, int32(10), resp.DownloadCount)
		mockFileRepo.AssertExpectations(t)
		mockDownloadSvc.AssertExpectations(t)
	})

	t.Run("GetFile access denied", func(t *testing.T) {
		mockFileRepo := new(MockFileRepository)
		mockDownloadSvc := new(MockDownloadService)
		mockQuotaSvc := new(MockQuotaService)
		mockMultipartSvc := new(MockMultipartUploadService)

		adapter := NewFileAdapter(mockFileRepo, nil, mockDownloadSvc, mockQuotaSvc, mockMultipartSvc)

		fileID := uuid.New()

		mockDownloadSvc.On("CheckAccess", mock.Anything, fileID, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).
			Return(false, nil).Once()

		req := &filev1.GetFileRequest{
			Id: fileID.String(),
		}

		resp, err := adapter.GetFile(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "access denied")
		mockDownloadSvc.AssertExpectations(t)
	})
}

// TestFileAdapter_ListFiles tests listing files
func TestFileAdapter_ListFiles(t *testing.T) {
	t.Run("ListFiles successfully", func(t *testing.T) {
		mockFileRepo := new(MockFileRepository)
		mockDownloadSvc := new(MockDownloadService)
		mockQuotaSvc := new(MockQuotaService)
		mockMultipartSvc := new(MockMultipartUploadService)

		adapter := NewFileAdapter(mockFileRepo, nil, mockDownloadSvc, mockQuotaSvc, mockMultipartSvc)

		tenantID := uuid.New()
		userID := uuid.New()

		expectedFiles := []*model.File{
			{
				ID:          uuid.New(),
				TenantID:    tenantID,
				Filename:    "file1.pdf",
				Size:        1024,
				MimeType:    "application/pdf",
				Status:      model.FileStatusActive,
				UploadedBy:  userID,
				AccessLevel: model.AccessLevelPrivate,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          uuid.New(),
				TenantID:    tenantID,
				Filename:    "file2.doc",
				Size:        2048,
				MimeType:    "application/msword",
				Status:      model.FileStatusActive,
				UploadedBy:  userID,
				AccessLevel: model.AccessLevelPrivate,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		mockFileRepo.On("List", mock.Anything, mock.AnythingOfType("*repository.FileFilter")).
			Return(expectedFiles, int64(2), nil).Once()

		req := &filev1.ListFilesRequest{
			Page:     1,
			PageSize: 20,
		}

		resp, err := adapter.ListFiles(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Files, 2)
		assert.Equal(t, int64(2), resp.Total)
		mockFileRepo.AssertExpectations(t)
	})
}

// TestFileAdapter_DeleteFile tests deleting a file
func TestFileAdapter_DeleteFile(t *testing.T) {
	t.Run("DeleteFile successfully", func(t *testing.T) {
		mockFileRepo := new(MockFileRepository)
		mockDownloadSvc := new(MockDownloadService)
		mockQuotaSvc := new(MockQuotaService)
		mockMultipartSvc := new(MockMultipartUploadService)

		adapter := NewFileAdapter(mockFileRepo, nil, mockDownloadSvc, mockQuotaSvc, mockMultipartSvc)

		fileID := uuid.New()
		// 使用与 getAuthInfoFromContext 返回的相同的 UUID
		tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

		file := &model.File{
			ID:          fileID,
			TenantID:    tenantID,
			Filename:    "test.pdf",
			UploadedBy:  userID,
			AccessLevel: model.AccessLevelPrivate,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockFileRepo.On("FindByID", mock.Anything, fileID).
			Return(file, nil).Once()
		mockFileRepo.On("SoftDelete", mock.Anything, fileID).
			Return(nil).Once()

		req := &filev1.DeleteFileRequest{
			Id: fileID.String(),
		}

		resp, err := adapter.DeleteFile(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockFileRepo.AssertExpectations(t)
	})
}

// TestFileAdapter_GetDownloadURL tests getting a download URL
func TestFileAdapter_GetDownloadURL(t *testing.T) {
	t.Run("GetDownloadURL successfully", func(t *testing.T) {
		mockFileRepo := new(MockFileRepository)
		mockDownloadSvc := new(MockDownloadService)
		mockQuotaSvc := new(MockQuotaService)
		mockMultipartSvc := new(MockMultipartUploadService)

		adapter := NewFileAdapter(mockFileRepo, nil, mockDownloadSvc, mockQuotaSvc, mockMultipartSvc)

		fileID := uuid.New()
		expectedURL := "https://example.com/download/test.pdf?token=abc123"

		mockDownloadSvc.On("GetDownloadURL", mock.Anything, fileID, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).
			Return(expectedURL, nil).Once()

		req := &filev1.GetDownloadURLRequest{
			Id:     fileID.String(),
			Expiry: 3600,
		}

		resp, err := adapter.GetDownloadURL(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedURL, resp.Url)
		assert.NotNil(t, resp.ExpiresAt)
		mockDownloadSvc.AssertExpectations(t)
	})
}

// TestFileAdapter_GetQuota tests getting quota information
func TestFileAdapter_GetQuota(t *testing.T) {
	t.Run("GetQuota successfully", func(t *testing.T) {
		mockFileRepo := new(MockFileRepository)
		mockDownloadSvc := new(MockDownloadService)
		mockQuotaSvc := new(MockQuotaService)
		mockMultipartSvc := new(MockMultipartUploadService)

		adapter := NewFileAdapter(mockFileRepo, nil, mockDownloadSvc, mockQuotaSvc, mockMultipartSvc)

		expectedQuota := &service.QuotaUsageInfo{
			QuotaLimit:     10737418240,
			QuotaUsed:      5368709120,
			QuotaAvailable: 5368709120,
			UsagePercent:   50.0,
			IsNearLimit:    false,
			IsExceeded:     false,
		}

		mockQuotaSvc.On("GetQuotaUsage", mock.Anything, mock.AnythingOfType("uuid.UUID")).
			Return(expectedQuota, nil).Once()

		resp, err := adapter.GetQuota(context.Background(), &emptypb.Empty{})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(10737418240), resp.QuotaLimit)
		assert.Equal(t, int64(5368709120), resp.QuotaUsed)
		assert.False(t, resp.IsExceeded)
		mockQuotaSvc.AssertExpectations(t)
	})
}

// TestFileAdapter_InitiateMultipartUpload tests initiating multipart upload
func TestFileAdapter_InitiateMultipartUpload(t *testing.T) {
	t.Run("InitiateMultipartUpload successfully", func(t *testing.T) {
		mockFileRepo := new(MockFileRepository)
		mockDownloadSvc := new(MockDownloadService)
		mockQuotaSvc := new(MockQuotaService)
		mockMultipartSvc := new(MockMultipartUploadService)

		adapter := NewFileAdapter(mockFileRepo, nil, mockDownloadSvc, mockQuotaSvc, mockMultipartSvc)

		expectedResp := &service.MultipartUploadResponse{
			UploadID:   "upload-123",
			StorageKey: "/uploads/large-file.zip",
			PartSize:   5242880,
			TotalParts: 10,
		}

		mockMultipartSvc.On("InitiateUpload", mock.Anything, mock.AnythingOfType("*service.InitiateMultipartUploadRequest")).
			Return(expectedResp, nil).Once()

		req := &filev1.InitiateMultipartRequest{
			Filename:    "large-file.zip",
			TotalSize:   52428800,
			ContentType: "application/zip",
		}

		resp, err := adapter.InitiateMultipartUpload(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "upload-123", resp.UploadId)
		assert.Equal(t, int64(5242880), resp.PartSize)
		assert.Equal(t, int32(10), resp.TotalParts)
		mockMultipartSvc.AssertExpectations(t)
	})
}

// TestFileAdapter_CompleteMultipartUpload tests completing multipart upload
func TestFileAdapter_CompleteMultipartUpload(t *testing.T) {
	t.Run("CompleteMultipartUpload successfully", func(t *testing.T) {
		mockFileRepo := new(MockFileRepository)
		mockDownloadSvc := new(MockDownloadService)
		mockQuotaSvc := new(MockQuotaService)
		mockMultipartSvc := new(MockMultipartUploadService)

		adapter := NewFileAdapter(mockFileRepo, nil, mockDownloadSvc, mockQuotaSvc, mockMultipartSvc)

		fileID := uuid.New()
		tenantID := uuid.New()
		userID := uuid.New()

		completedFile := &model.File{
			ID:          fileID,
			TenantID:    tenantID,
			Filename:    "large-file.zip",
			Size:        52428800,
			MimeType:    "application/zip",
			StorageKey:  "/files/large-file.zip",
			Checksum:    "def456",
			UploadedBy:  userID,
			AccessLevel: model.AccessLevelPrivate,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockMultipartSvc.On("CompleteUpload", mock.Anything, mock.AnythingOfType("*service.CompleteMultipartUploadRequest")).
			Return(completedFile, nil).Once()
		mockDownloadSvc.On("GetDownloadURL", mock.Anything, fileID, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("time.Duration")).
			Return("https://example.com/download/large-file.zip", nil).Once()

		req := &filev1.CompleteMultipartRequest{
			UploadId: "upload-123",
			Parts: []*filev1.PartInfo{
				{PartNumber: 1, Etag: "etag1"},
				{PartNumber: 2, Etag: "etag2"},
			},
		}

		resp, err := adapter.CompleteMultipartUpload(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, fileID.String(), resp.FileId)
		assert.Equal(t, "large-file.zip", resp.Filename)
		assert.Equal(t, int64(52428800), resp.Size)
		mockMultipartSvc.AssertExpectations(t)
		mockDownloadSvc.AssertExpectations(t)
	})
}
