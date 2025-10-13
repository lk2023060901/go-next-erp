package storage

import (
	"bytes"
	"context"
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

	// 获取底层 ObjectStorage
	store := s.GetObjectStorage()

	t.Run("BucketOperations", func(t *testing.T) {
		bucket := "test-bucket-ops"

		// 创建桶
		err := store.MakeBucket(ctx, bucket, MakeBucketOptions{
			Region: "us-east-1",
		})
		if err != nil {
			t.Errorf("MakeBucket() error = %v", err)
		}

		// 检查桶存在
		exists, err := store.BucketExists(ctx, bucket)
		if err != nil {
			t.Errorf("BucketExists() error = %v", err)
		}
		if !exists {
			t.Error("Bucket should exist")
		}

		// 列出桶
		buckets, err := store.ListBuckets(ctx)
		if err != nil {
			t.Errorf("ListBuckets() error = %v", err)
		}
		if len(buckets) == 0 {
			t.Error("Should have at least one bucket")
		}

		// 删除桶
		err = store.RemoveBucket(ctx, bucket)
		if err != nil {
			t.Errorf("RemoveBucket() error = %v", err)
		}
	})

	t.Run("ObjectOperations", func(t *testing.T) {
		bucket := "test-bucket"
		key := "test/file.txt"
		content := []byte("Hello, MinIO!")

		// 上传对象
		reader := bytes.NewReader(content)
		result, err := store.PutObject(ctx, bucket, key, reader, int64(len(content)), PutObjectOptions{
			ContentType: "text/plain",
		})
		if err != nil {
			t.Errorf("PutObject() error = %v", err)
		}
		if result == nil {
			t.Error("PutObject() result should not be nil")
		}

		// 检查对象信息
		info, err := store.StatObject(ctx, bucket, key)
		if err != nil {
			t.Errorf("StatObject() error = %v", err)
		}
		if info == nil || info.Size != int64(len(content)) {
			t.Error("Object info mismatch")
		}

		// 获取预签名 URL
		url, err := store.PresignedGetObject(ctx, bucket, key, 1*time.Hour, PresignedGetOptions{})
		if err != nil {
			t.Errorf("PresignedGetObject() error = %v", err)
		}
		if url == "" {
			t.Error("Presigned URL should not be empty")
		}

		// 删除对象
		err = store.RemoveObject(ctx, bucket, key)
		if err != nil {
			t.Errorf("RemoveObject() error = %v", err)
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
