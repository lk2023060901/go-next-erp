## pkg/vector - Milvus 向量数据库使用文档

### 一、快速开始

#### 1.1 初始化客户端

```go
package main

import (
    "context"
    "github.com/lk2023060901/go-next-erp/pkg/vector"
)

func main() {
    ctx := context.Background()

    // 方式 1：使用默认配置
    v, err := vector.New(ctx)
    if err != nil {
        panic(err)
    }
    defer v.Close()

    // 方式 2：使用选项配置
    v, err = vector.New(ctx,
        vector.WithEndpoint("localhost:15004"),
        vector.WithDatabase("default"),
        vector.WithCredentials("username", "password"), // 可选
    )
}
```

#### 1.2 环境变量配置

```env
MILVUS_ENDPOINT=localhost:15004
MILVUS_DATABASE=default
MILVUS_USERNAME=username
MILVUS_PASSWORD=password
MILVUS_MAX_IDLE_CONNS=10
MILVUS_MAX_OPEN_CONNS=50
MILVUS_TIMEOUT=30s
MILVUS_MAX_RETRIES=3
```

---

### 二、集合管理

#### 2.1 创建集合

```go
import "github.com/milvus-io/milvus-sdk-go/v2/entity"

// 定义 Schema
schema := &entity.Schema{
    CollectionName: "documents",
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
                "dim": "768", // 向量维度
            },
        },
        {
            Name:     "content",
            DataType: entity.FieldTypeVarChar,
            TypeParams: map[string]string{
                "max_length": "1024",
            },
        },
        {
            Name:     "metadata",
            DataType: entity.FieldTypeJSON,
        },
    },
}

// 创建集合
err := v.CreateCollection(ctx, "documents", schema)
```

#### 2.2 检查集合是否存在

```go
exists, err := v.HasCollection(ctx, "documents")
if exists {
    fmt.Println("Collection exists")
}
```

#### 2.3 列出所有集合

```go
collections, err := v.ListCollections(ctx)
for _, coll := range collections {
    fmt.Println("Collection:", coll)
}
```

#### 2.4 获取集合描述

```go
coll, err := v.DescribeCollection(ctx, "documents")
fmt.Printf("Name: %s, Schema: %+v\n", coll.Name, coll.Schema)
```

#### 2.5 加载/释放集合

```go
// 加载集合到内存（搜索前必须加载）
err := v.LoadCollection(ctx, "documents")

// 从内存释放集合
err = v.ReleaseCollection(ctx, "documents")
```

#### 2.6 删除集合

```go
err := v.DropCollection(ctx, "documents")
```

---

### 三、索引管理

#### 3.1 创建索引

```go
// IVF_FLAT 索引（适合中小规模数据）
idx, err := entity.NewIndexIvfFlat(entity.L2, 1024)
if err != nil {
    panic(err)
}
err = v.CreateIndex(ctx, "documents", "embedding", idx)

// HNSW 索引（适合大规模数据）
idx, err = entity.NewIndexHNSW(entity.L2, 16, 200)
if err != nil {
    panic(err)
}
err = v.CreateIndex(ctx, "documents", "embedding", idx)
```

#### 3.2 索引类型

| 索引类型 | 适用场景 | 特点 |
|---------|---------|------|
| IVF_FLAT | 中小规模 | 精度高，速度中等 |
| IVF_SQ8 | 中等规模 | 内存占用小，速度快 |
| IVF_PQ | 大规模 | 内存占用最小，精度略低 |
| HNSW | 大规模高性能 | 速度最快，内存占用较大 |

#### 3.3 距离度量

```go
entity.L2          // 欧氏距离（默认）
entity.IP          // 内积
entity.COSINE      // 余弦相似度
```

#### 3.4 描述索引

```go
indexes, err := v.DescribeIndex(ctx, "documents", "embedding")
for _, idx := range indexes {
    fmt.Printf("Index: %+v\n", idx)
}
```

#### 3.5 删除索引

```go
err := v.DropIndex(ctx, "documents", "embedding")
```

---

### 四、数据操作

#### 4.1 插入数据

```go
import "github.com/milvus-io/milvus-sdk-go/v2/entity"

// 准备数据
embeddings := [][]float32{
    {0.1, 0.2, 0.3, ..., 0.768}, // 768 维向量
    {0.2, 0.3, 0.4, ..., 0.768},
}

contents := []string{
    "这是第一段内容",
    "这是第二段内容",
}

// 构建列数据
embeddingCol := entity.NewColumnFloatVector("embedding", 768, embeddings)
contentCol := entity.NewColumnVarChar("content", contents)

// 插入数据
ids, err := v.Insert(ctx, "documents", "", embeddingCol, contentCol)
if err != nil {
    panic(err)
}

fmt.Printf("Inserted %d vectors\n", ids.Len())
```

#### 4.2 刷新数据

```go
// 将数据持久化到磁盘
err := v.Flush(ctx, "documents")
```

#### 4.3 删除数据

```go
// 通过表达式删除
err := v.Delete(ctx, "documents", "", "id in [1, 2, 3]")
```

---

### 五、向量搜索

#### 5.1 基本搜索

```go
// 搜索向量
queryVector := []float32{0.1, 0.2, 0.3, ..., 0.768}

// 构建搜索参数
sp, _ := entity.NewIndexIvfFlatSearchParam(16) // nprobe=16

// 执行搜索
results, err := v.Search(
    ctx,
    "documents",                    // 集合名
    []string{},                     // 分区（空=所有分区）
    "",                             // 过滤表达式
    []string{"id", "content"},      // 输出字段
    []entity.Vector{entity.FloatVector(queryVector)}, // 查询向量
    "embedding",                    // 向量字段
    entity.L2,                      // 距离度量
    10,                             // TopK
    sp,                             // 搜索参数
)

// 处理结果
for _, result := range results {
    for i := 0; i < result.ResultCount; i++ {
        id, _ := result.IDs.GetAsInt64(i)
        score := result.Scores[i]
        fmt.Printf("ID: %d, Score: %f\n", id, score)
    }
}
```

#### 5.2 过滤搜索

```go
// 带过滤条件的搜索
results, err := v.Search(
    ctx,
    "documents",
    []string{},
    "content like '%关键词%'", // 过滤表达式
    []string{"id", "content"},
    []entity.Vector{entity.FloatVector(queryVector)},
    "embedding",
    entity.L2,
    10,
    sp,
)
```

#### 5.3 多向量搜索

```go
// 批量搜索
queryVectors := []entity.Vector{
    entity.FloatVector(vector1),
    entity.FloatVector(vector2),
}

results, err := v.Search(
    ctx,
    "documents",
    []string{},
    "",
    []string{"id", "content"},
    queryVectors,
    "embedding",
    entity.L2,
    10,
    sp,
)
```

---

### 六、查询数据

#### 6.1 按表达式查询

```go
// 查询特定 ID 的数据
columns, err := v.Query(
    ctx,
    "documents",
    []string{},                     // 分区
    "id in [1, 2, 3]",              // 查询表达式
    []string{"id", "content"},      // 输出字段
)

// 处理结果
for _, col := range columns {
    switch col.Name() {
    case "id":
        if idCol, ok := col.(*entity.ColumnInt64); ok {
            fmt.Println("IDs:", idCol.Data())
        }
    case "content":
        if contentCol, ok := col.(*entity.ColumnVarChar); ok {
            fmt.Println("Contents:", contentCol.Data())
        }
    }
}
```

---

### 七、分区管理

#### 7.1 创建分区

```go
err := v.CreatePartition(ctx, "documents", "2024_01")
```

#### 7.2 检查分区是否存在

```go
exists, err := v.HasPartition(ctx, "documents", "2024_01")
```

#### 7.3 列出分区

```go
partitions, err := v.ListPartitions(ctx, "documents")
for _, partition := range partitions {
    fmt.Println("Partition:", partition)
}
```

#### 7.4 删除分区

```go
err := v.DropPartition(ctx, "documents", "2024_01")
```

---

### 八、完整示例

```go
package main

import (
    "context"
    "fmt"

    "github.com/lk2023060901/go-next-erp/pkg/vector"
    "github.com/milvus-io/milvus-sdk-go/v2/entity"
)

func main() {
    ctx := context.Background()

    // 初始化
    v, err := vector.New(ctx,
        vector.WithEndpoint("localhost:15004"),
        vector.WithDatabase("default"),
    )
    if err != nil {
        panic(err)
    }
    defer v.Close()

    collectionName := "demo_collection"

    // 1. 创建集合
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

    err = v.CreateCollection(ctx, collectionName, schema)
    if err != nil {
        panic(err)
    }
    fmt.Println("Collection created")

    // 2. 创建索引
    idx, _ := entity.NewIndexIvfFlat(entity.L2, 128)
    err = v.CreateIndex(ctx, collectionName, "embedding", idx)
    if err != nil {
        panic(err)
    }
    fmt.Println("Index created")

    // 3. 加载集合
    err = v.LoadCollection(ctx, collectionName)
    if err != nil {
        panic(err)
    }
    fmt.Println("Collection loaded")

    // 4. 插入数据
    embeddings := make([][]float32, 10)
    contents := make([]string, 10)
    for i := 0; i < 10; i++ {
        // 生成随机向量
        embedding := make([]float32, 128)
        for j := range embedding {
            embedding[j] = float32(i + j)
        }
        embeddings[i] = embedding
        contents[i] = fmt.Sprintf("Content %d", i)
    }

    embeddingCol := entity.NewColumnFloatVector("embedding", 128, embeddings)
    contentCol := entity.NewColumnVarChar("content", contents)

    ids, err := v.Insert(ctx, collectionName, "", embeddingCol, contentCol)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Inserted %d vectors\n", ids.Len())

    // 5. 刷新数据
    err = v.Flush(ctx, collectionName)
    if err != nil {
        panic(err)
    }
    fmt.Println("Data flushed")

    // 6. 搜索
    queryVector := make([]float32, 128)
    for i := range queryVector {
        queryVector[i] = float32(i)
    }

    sp, _ := entity.NewIndexIvfFlatSearchParam(16)
    results, err := v.Search(
        ctx,
        collectionName,
        []string{},
        "",
        []string{"id", "content"},
        []entity.Vector{entity.FloatVector(queryVector)},
        "embedding",
        entity.L2,
        5,
        sp,
    )
    if err != nil {
        panic(err)
    }

    fmt.Println("Search results:")
    for _, result := range results {
        for i := 0; i < result.ResultCount; i++ {
            id, _ := result.IDs.GetAsInt64(i)
            score := result.Scores[i]
            fmt.Printf("  ID: %d, Score: %f\n", id, score)
        }
    }

    // 7. 清理
    err = v.DropCollection(ctx, collectionName)
    if err != nil {
        panic(err)
    }
    fmt.Println("Collection dropped")
}
```

---

### 九、最佳实践

#### 9.1 向量维度选择

- **小规模 (< 10万)**: 128-384 维
- **中等规模 (10万-100万)**: 384-768 维
- **大规模 (> 100万)**: 768-1536 维

#### 9.2 索引选择

```go
// 小规模：IVF_FLAT
idx, _ := entity.NewIndexIvfFlat(entity.L2, 1024)

// 中等规模：IVF_SQ8
idx, _ := entity.NewIndexIvfSQ8(entity.L2, 1024)

// 大规模：HNSW
idx, _ := entity.NewIndexHNSW(entity.L2, 16, 200)
```

#### 9.3 批量插入

```go
// 分批插入，每批 1000-10000 条
batchSize := 5000
for i := 0; i < totalCount; i += batchSize {
    end := i + batchSize
    if end > totalCount {
        end = totalCount
    }

    batch := embeddings[i:end]
    // 插入批次数据
    v.Insert(ctx, collectionName, "", batch...)
}
```

#### 9.4 搜索参数调优

```go
// nprobe 越大，精度越高，但速度越慢
// 建议值：nlist 的 10%-30%

// 创建索引时的 nlist
idx, _ := entity.NewIndexIvfFlat(entity.L2, 1024) // nlist=1024

// 搜索时的 nprobe
sp, _ := entity.NewIndexIvfFlatSearchParam(256) // nprobe=256 (25%)
```

---

### 十、配置参考

| 选项 | 说明 | 默认值 |
|------|------|--------|
| Endpoint | Milvus 地址 | localhost:19530 |
| Database | 数据库名称 | default |
| Username | 用户名 | "" |
| Password | 密码 | "" |
| MaxIdleConns | 最大空闲连接数 | 10 |
| MaxOpenConns | 最大打开连接数 | 50 |
| Timeout | 超时时间 | 30s |
| MaxRetries | 最大重试次数 | 3 |
