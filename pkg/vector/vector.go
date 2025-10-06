package vector

import (
	"context"
	"fmt"

	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"go.uber.org/zap"
)

// Vector Milvus 向量数据库接口
type Vector interface {
	// 集合管理
	CreateCollection(ctx context.Context, collectionName string, schema *entity.Schema) error
	DropCollection(ctx context.Context, collectionName string) error
	HasCollection(ctx context.Context, collectionName string) (bool, error)
	ListCollections(ctx context.Context) ([]string, error)
	DescribeCollection(ctx context.Context, collectionName string) (*entity.Collection, error)
	LoadCollection(ctx context.Context, collectionName string) error
	ReleaseCollection(ctx context.Context, collectionName string) error

	// 分区管理
	CreatePartition(ctx context.Context, collectionName, partitionName string) error
	DropPartition(ctx context.Context, collectionName, partitionName string) error
	HasPartition(ctx context.Context, collectionName, partitionName string) (bool, error)
	ListPartitions(ctx context.Context, collectionName string) ([]string, error)

	// 索引管理
	CreateIndex(ctx context.Context, collectionName, fieldName string, idx entity.Index) error
	DropIndex(ctx context.Context, collectionName, fieldName string) error
	DescribeIndex(ctx context.Context, collectionName, fieldName string) ([]entity.Index, error)

	// 数据操作
	Insert(ctx context.Context, collectionName string, partitionName string, columns ...entity.Column) (entity.Column, error)
	Delete(ctx context.Context, collectionName string, partitionName string, expr string) error
	Flush(ctx context.Context, collectionName string) error

	// 向量搜索
	Search(ctx context.Context, collectionName string, partitions []string, expr string, outputFields []string, vectors []entity.Vector, vectorField string, metricType entity.MetricType, topK int, sp entity.SearchParam) ([]client.SearchResult, error)

	// 查询
	Query(ctx context.Context, collectionName string, partitions []string, expr string, outputFields []string) ([]entity.Column, error)

	// 关闭连接
	Close() error
}

// vector Milvus 实现
type vector struct {
	client client.Client
	config *Config
	logger *logger.Logger
}

// New 创建 Milvus 客户端
func New(ctx context.Context, opts ...Option) (Vector, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建 Milvus 客户端配置
	clientConfig := client.Config{
		Address: cfg.Endpoint,
		DBName:  cfg.Database,
	}

	// 如果提供了认证信息
	if cfg.Username != "" {
		clientConfig.Username = cfg.Username
		clientConfig.Password = cfg.Password
	}

	// 连接到 Milvus
	c, err := client.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to milvus: %w", err)
	}

	v := &vector{
		client: c,
		config: cfg,
		logger: logger.GetLogger().With(zap.String("module", "vector")),
	}

	v.logger.Info("Milvus client initialized",
		zap.String("endpoint", cfg.Endpoint),
		zap.String("database", cfg.Database),
	)

	return v, nil
}

// CreateCollection 创建集合
func (v *vector) CreateCollection(ctx context.Context, collectionName string, schema *entity.Schema) error {
	err := v.client.CreateCollection(ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		v.logger.Error("Failed to create collection",
			zap.String("collection", collectionName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create collection: %w", err)
	}

	v.logger.Info("Collection created successfully", zap.String("collection", collectionName))
	return nil
}

// DropCollection 删除集合
func (v *vector) DropCollection(ctx context.Context, collectionName string) error {
	err := v.client.DropCollection(ctx, collectionName)
	if err != nil {
		v.logger.Error("Failed to drop collection",
			zap.String("collection", collectionName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to drop collection: %w", err)
	}

	v.logger.Info("Collection dropped successfully", zap.String("collection", collectionName))
	return nil
}

// HasCollection 检查集合是否存在
func (v *vector) HasCollection(ctx context.Context, collectionName string) (bool, error) {
	exists, err := v.client.HasCollection(ctx, collectionName)
	if err != nil {
		return false, fmt.Errorf("failed to check collection existence: %w", err)
	}
	return exists, nil
}

// ListCollections 列出所有集合
func (v *vector) ListCollections(ctx context.Context) ([]string, error) {
	collections, err := v.client.ListCollections(ctx)
	if err != nil {
		v.logger.Error("Failed to list collections", zap.Error(err))
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	names := make([]string, len(collections))
	for i, coll := range collections {
		names[i] = coll.Name
	}
	return names, nil
}

// DescribeCollection 获取集合描述
func (v *vector) DescribeCollection(ctx context.Context, collectionName string) (*entity.Collection, error) {
	coll, err := v.client.DescribeCollection(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe collection: %w", err)
	}
	return coll, nil
}

// LoadCollection 加载集合到内存
func (v *vector) LoadCollection(ctx context.Context, collectionName string) error {
	err := v.client.LoadCollection(ctx, collectionName, false)
	if err != nil {
		v.logger.Error("Failed to load collection",
			zap.String("collection", collectionName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to load collection: %w", err)
	}

	v.logger.Debug("Collection loaded successfully", zap.String("collection", collectionName))
	return nil
}

// ReleaseCollection 从内存释放集合
func (v *vector) ReleaseCollection(ctx context.Context, collectionName string) error {
	err := v.client.ReleaseCollection(ctx, collectionName)
	if err != nil {
		v.logger.Error("Failed to release collection",
			zap.String("collection", collectionName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to release collection: %w", err)
	}

	v.logger.Debug("Collection released successfully", zap.String("collection", collectionName))
	return nil
}

// CreatePartition 创建分区
func (v *vector) CreatePartition(ctx context.Context, collectionName, partitionName string) error {
	err := v.client.CreatePartition(ctx, collectionName, partitionName)
	if err != nil {
		v.logger.Error("Failed to create partition",
			zap.String("collection", collectionName),
			zap.String("partition", partitionName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create partition: %w", err)
	}

	v.logger.Debug("Partition created successfully",
		zap.String("collection", collectionName),
		zap.String("partition", partitionName),
	)
	return nil
}

// DropPartition 删除分区
func (v *vector) DropPartition(ctx context.Context, collectionName, partitionName string) error {
	err := v.client.DropPartition(ctx, collectionName, partitionName)
	if err != nil {
		v.logger.Error("Failed to drop partition",
			zap.String("collection", collectionName),
			zap.String("partition", partitionName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to drop partition: %w", err)
	}

	v.logger.Debug("Partition dropped successfully",
		zap.String("collection", collectionName),
		zap.String("partition", partitionName),
	)
	return nil
}

// HasPartition 检查分区是否存在
func (v *vector) HasPartition(ctx context.Context, collectionName, partitionName string) (bool, error) {
	exists, err := v.client.HasPartition(ctx, collectionName, partitionName)
	if err != nil {
		return false, fmt.Errorf("failed to check partition existence: %w", err)
	}
	return exists, nil
}

// ListPartitions 列出集合的所有分区
func (v *vector) ListPartitions(ctx context.Context, collectionName string) ([]string, error) {
	partitions, err := v.client.ShowPartitions(ctx, collectionName)
	if err != nil {
		v.logger.Error("Failed to list partitions",
			zap.String("collection", collectionName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list partitions: %w", err)
	}

	names := make([]string, len(partitions))
	for i, p := range partitions {
		names[i] = p.Name
	}
	return names, nil
}

// CreateIndex 创建索引
func (v *vector) CreateIndex(ctx context.Context, collectionName, fieldName string, idx entity.Index) error {
	err := v.client.CreateIndex(ctx, collectionName, fieldName, idx, false)
	if err != nil {
		v.logger.Error("Failed to create index",
			zap.String("collection", collectionName),
			zap.String("field", fieldName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create index: %w", err)
	}

	v.logger.Info("Index created successfully",
		zap.String("collection", collectionName),
		zap.String("field", fieldName),
	)
	return nil
}

// DropIndex 删除索引
func (v *vector) DropIndex(ctx context.Context, collectionName, fieldName string) error {
	err := v.client.DropIndex(ctx, collectionName, fieldName)
	if err != nil {
		v.logger.Error("Failed to drop index",
			zap.String("collection", collectionName),
			zap.String("field", fieldName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to drop index: %w", err)
	}

	v.logger.Debug("Index dropped successfully",
		zap.String("collection", collectionName),
		zap.String("field", fieldName),
	)
	return nil
}

// DescribeIndex 获取索引描述
func (v *vector) DescribeIndex(ctx context.Context, collectionName, fieldName string) ([]entity.Index, error) {
	indexes, err := v.client.DescribeIndex(ctx, collectionName, fieldName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe index: %w", err)
	}
	return indexes, nil
}

// Insert 插入数据
func (v *vector) Insert(ctx context.Context, collectionName string, partitionName string, columns ...entity.Column) (entity.Column, error) {
	result, err := v.client.Insert(ctx, collectionName, partitionName, columns...)
	if err != nil {
		v.logger.Error("Failed to insert data",
			zap.String("collection", collectionName),
			zap.String("partition", partitionName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to insert data: %w", err)
	}

	v.logger.Debug("Data inserted successfully",
		zap.String("collection", collectionName),
		zap.String("partition", partitionName),
	)
	return result, nil
}

// Delete 删除数据
func (v *vector) Delete(ctx context.Context, collectionName string, partitionName string, expr string) error {
	err := v.client.Delete(ctx, collectionName, partitionName, expr)
	if err != nil {
		v.logger.Error("Failed to delete data",
			zap.String("collection", collectionName),
			zap.String("partition", partitionName),
			zap.String("expr", expr),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete data: %w", err)
	}

	v.logger.Debug("Data deleted successfully",
		zap.String("collection", collectionName),
		zap.String("partition", partitionName),
		zap.String("expr", expr),
	)
	return nil
}

// Flush 刷新数据到持久化存储
func (v *vector) Flush(ctx context.Context, collectionName string) error {
	err := v.client.Flush(ctx, collectionName, false)
	if err != nil {
		v.logger.Error("Failed to flush collection",
			zap.String("collection", collectionName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to flush collection: %w", err)
	}

	v.logger.Debug("Collection flushed successfully", zap.String("collection", collectionName))
	return nil
}

// Search 向量搜索
func (v *vector) Search(ctx context.Context, collectionName string, partitions []string, expr string, outputFields []string, vectors []entity.Vector, vectorField string, metricType entity.MetricType, topK int, sp entity.SearchParam) ([]client.SearchResult, error) {
	results, err := v.client.Search(ctx, collectionName, partitions, expr, outputFields, vectors, vectorField, metricType, topK, sp)
	if err != nil {
		v.logger.Error("Failed to search",
			zap.String("collection", collectionName),
			zap.Strings("partitions", partitions),
			zap.String("expr", expr),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	v.logger.Debug("Search completed successfully",
		zap.String("collection", collectionName),
		zap.Int("results", len(results)),
	)
	return results, nil
}

// Query 查询数据
func (v *vector) Query(ctx context.Context, collectionName string, partitions []string, expr string, outputFields []string) ([]entity.Column, error) {
	columns, err := v.client.Query(ctx, collectionName, partitions, expr, outputFields)
	if err != nil {
		v.logger.Error("Failed to query",
			zap.String("collection", collectionName),
			zap.Strings("partitions", partitions),
			zap.String("expr", expr),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	v.logger.Debug("Query completed successfully",
		zap.String("collection", collectionName),
		zap.Int("columns", len(columns)),
	)
	return columns, nil
}

// Close 关闭连接
func (v *vector) Close() error {
	if err := v.client.Close(); err != nil {
		v.logger.Error("Failed to close milvus client", zap.Error(err))
		return fmt.Errorf("failed to close milvus client: %w", err)
	}

	v.logger.Info("Milvus client closed")
	return nil
}
