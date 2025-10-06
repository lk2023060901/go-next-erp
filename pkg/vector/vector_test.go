package vector

import (
	"context"
	"testing"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

func TestVector(t *testing.T) {
	ctx := context.Background()

	// 创建向量客户端
	v, err := New(ctx,
		WithEndpoint("localhost:15004"),
		WithDatabase("default"),
	)
	if err != nil {
		t.Skipf("Milvus not available: %v", err)
	}
	defer v.Close()

	collectionName := "test_collection"

	t.Run("CollectionOperations", func(t *testing.T) {
		// 创建 Schema
		schema := &entity.Schema{
			CollectionName: collectionName,
			AutoID:         true,
			Fields: []*entity.Field{
				{
					Name:       "id",
					DataType:   entity.FieldTypeInt64,
					PrimaryKey: true,
					AutoID:     true,
				},
				{
					Name:     "embedding",
					DataType: entity.FieldTypeFloatVector,
					TypeParams: map[string]string{
						"dim": "128",
					},
				},
				{
					Name:     "content",
					DataType: entity.FieldTypeVarChar,
					TypeParams: map[string]string{
						"max_length": "512",
					},
				},
			},
		}

		// 创建集合
		err := v.CreateCollection(ctx, collectionName, schema)
		if err != nil {
			t.Errorf("CreateCollection() error = %v", err)
		}

		// 检查集合存在
		exists, err := v.HasCollection(ctx, collectionName)
		if err != nil {
			t.Errorf("HasCollection() error = %v", err)
		}
		if !exists {
			t.Error("Collection should exist")
		}

		// 列出集合
		collections, err := v.ListCollections(ctx)
		if err != nil {
			t.Errorf("ListCollections() error = %v", err)
		}
		if len(collections) == 0 {
			t.Error("Should have at least one collection")
		}

		// 描述集合
		coll, err := v.DescribeCollection(ctx, collectionName)
		if err != nil {
			t.Errorf("DescribeCollection() error = %v", err)
		}
		if coll.Name != collectionName {
			t.Errorf("Collection name = %s, want %s", coll.Name, collectionName)
		}

		// 删除集合
		err = v.DropCollection(ctx, collectionName)
		if err != nil {
			t.Errorf("DropCollection() error = %v", err)
		}
	})

	t.Run("PartitionOperations", func(t *testing.T) {
		// 先创建集合
		schema := &entity.Schema{
			CollectionName: collectionName,
			Fields: []*entity.Field{
				{
					Name:       "id",
					DataType:   entity.FieldTypeInt64,
					PrimaryKey: true,
					AutoID:     true,
				},
				{
					Name:     "embedding",
					DataType: entity.FieldTypeFloatVector,
					TypeParams: map[string]string{
						"dim": "128",
					},
				},
			},
		}

		if err := v.CreateCollection(ctx, collectionName, schema); err != nil {
			t.Skipf("Failed to create collection: %v", err)
		}
		defer v.DropCollection(ctx, collectionName)

		partitionName := "test_partition"

		// 创建分区
		err := v.CreatePartition(ctx, collectionName, partitionName)
		if err != nil {
			t.Errorf("CreatePartition() error = %v", err)
		}

		// 检查分区存在
		exists, err := v.HasPartition(ctx, collectionName, partitionName)
		if err != nil {
			t.Errorf("HasPartition() error = %v", err)
		}
		if !exists {
			t.Error("Partition should exist")
		}

		// 列出分区
		partitions, err := v.ListPartitions(ctx, collectionName)
		if err != nil {
			t.Errorf("ListPartitions() error = %v", err)
		}
		if len(partitions) == 0 {
			t.Error("Should have at least one partition")
		}

		// 删除分区
		err = v.DropPartition(ctx, collectionName, partitionName)
		if err != nil {
			t.Errorf("DropPartition() error = %v", err)
		}
	})

	t.Run("IndexOperations", func(t *testing.T) {
		// 先创建集合
		schema := &entity.Schema{
			CollectionName: collectionName,
			Fields: []*entity.Field{
				{
					Name:       "id",
					DataType:   entity.FieldTypeInt64,
					PrimaryKey: true,
					AutoID:     true,
				},
				{
					Name:     "embedding",
					DataType: entity.FieldTypeFloatVector,
					TypeParams: map[string]string{
						"dim": "128",
					},
				},
			},
		}

		if err := v.CreateCollection(ctx, collectionName, schema); err != nil {
			t.Skipf("Failed to create collection: %v", err)
		}
		defer v.DropCollection(ctx, collectionName)

		// 创建索引
		idx, err := entity.NewIndexIvfFlat(entity.L2, 128)
		if err != nil {
			t.Errorf("NewIndexIvfFlat() error = %v", err)
		}

		err = v.CreateIndex(ctx, collectionName, "embedding", idx)
		if err != nil {
			t.Errorf("CreateIndex() error = %v", err)
		}

		// 描述索引
		indexes, err := v.DescribeIndex(ctx, collectionName, "embedding")
		if err != nil {
			t.Errorf("DescribeIndex() error = %v", err)
		}
		if len(indexes) == 0 {
			t.Error("Should have at least one index")
		}

		// 删除索引
		err = v.DropIndex(ctx, collectionName, "embedding")
		if err != nil {
			t.Errorf("DropIndex() error = %v", err)
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
				Database:     "default",
				MaxIdleConns: 10,
				MaxOpenConns: 50,
				Timeout:      30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "Missing database",
			config: &Config{
				Endpoint:     "localhost:19530",
				MaxIdleConns: 10,
				MaxOpenConns: 50,
				Timeout:      30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "Invalid max idle conns",
			config: &Config{
				Endpoint:     "localhost:19530",
				Database:     "default",
				MaxIdleConns: 0,
				MaxOpenConns: 50,
				Timeout:      30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "Invalid timeout",
			config: &Config{
				Endpoint:     "localhost:19530",
				Database:     "default",
				MaxIdleConns: 10,
				MaxOpenConns: 50,
				Timeout:      0,
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
