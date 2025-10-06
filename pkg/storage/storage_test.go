package storage

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"
)

func TestStorage(t *testing.T) {
	ctx := context.Background()

	// 创建存储客户端
	s, err := New(ctx,
		WithEndpoint("localhost:15002"),
		WithCredentials("minioadmin", "minioadmin123"),
		WithSSL(false),
		WithBucket("test-bucket"),
	)
	if err != nil {
		t.Skipf("MinIO not available: %v", err)
	}
	defer s.Close()

	t.Run("BucketOperations", func(t *testing.T) {
		bucket := "test-bucket-ops"

		// 创建桶
		err := s.CreateBucket(ctx, bucket)
		if err != nil {
			t.Errorf("CreateBucket() error = %v", err)
		}

		// 检查桶存在
		exists, err := s.BucketExists(ctx, bucket)
		if err != nil {
			t.Errorf("BucketExists() error = %v", err)
		}
		if !exists {
			t.Error("Bucket should exist")
		}

		// 列出桶
		buckets, err := s.ListBuckets(ctx)
		if err != nil {
			t.Errorf("ListBuckets() error = %v", err)
		}
		if len(buckets) == 0 {
			t.Error("Should have at least one bucket")
		}

		// 删除桶
		err = s.DeleteBucket(ctx, bucket)
		if err != nil {
			t.Errorf("DeleteBucket() error = %v", err)
		}
	})

	t.Run("FileOperations", func(t *testing.T) {
		bucket := "test-bucket"
		key := "test/file.txt"
		content := []byte("Hello, MinIO!")

		// 上传文件
		reader := bytes.NewReader(content)
		err := s.UploadFile(ctx, bucket, key, reader, int64(len(content)), "text/plain")
		if err != nil {
			t.Errorf("UploadFile() error = %v", err)
		}

		// 检查文件存在
		exists, err := s.FileExists(ctx, bucket, key)
		if err != nil {
			t.Errorf("FileExists() error = %v", err)
		}
		if !exists {
			t.Error("File should exist")
		}

		// 获取文件信息
		info, err := s.GetFileInfo(ctx, bucket, key)
		if err != nil {
			t.Errorf("GetFileInfo() error = %v", err)
		}
		if info.Size != int64(len(content)) {
			t.Errorf("File size = %d, want %d", info.Size, len(content))
		}

		// 下载文件
		tmpFile := "/tmp/test-download.txt"
		err = s.DownloadFile(ctx, bucket, key, tmpFile)
		if err != nil {
			t.Errorf("DownloadFile() error = %v", err)
		}
		defer os.Remove(tmpFile)

		// 验证下载内容
		downloaded, err := os.ReadFile(tmpFile)
		if err != nil {
			t.Errorf("ReadFile() error = %v", err)
		}
		if !bytes.Equal(downloaded, content) {
			t.Errorf("Downloaded content = %s, want %s", downloaded, content)
		}

		// 获取预签名 URL
		url, err := s.GetPresignedURL(ctx, bucket, key, 1*time.Hour)
		if err != nil {
			t.Errorf("GetPresignedURL() error = %v", err)
		}
		if url == "" {
			t.Error("Presigned URL should not be empty")
		}

		// 列出文件
		files, err := s.ListFiles(ctx, bucket, "test/")
		if err != nil {
			t.Errorf("ListFiles() error = %v", err)
		}
		if len(files) == 0 {
			t.Error("Should have at least one file")
		}

		// 删除文件
		err = s.DeleteFile(ctx, bucket, key)
		if err != nil {
			t.Errorf("DeleteFile() error = %v", err)
		}

		// 验证文件已删除
		exists, err = s.FileExists(ctx, bucket, key)
		if err != nil {
			t.Errorf("FileExists() error = %v", err)
		}
		if exists {
			t.Error("File should not exist after deletion")
		}
	})
}

func TestConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "Valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "Missing endpoint",
			config: &Config{
				AccessKeyID:     "key",
				SecretAccessKey: "secret",
				BucketName:      "bucket",
			},
			wantErr: true,
		},
		{
			name: "Missing access key",
			config: &Config{
				Endpoint:        "localhost:9000",
				SecretAccessKey: "secret",
				BucketName:      "bucket",
			},
			wantErr: true,
		},
		{
			name: "Missing bucket name",
			config: &Config{
				Endpoint:        "localhost:9000",
				AccessKeyID:     "key",
				SecretAccessKey: "secret",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
