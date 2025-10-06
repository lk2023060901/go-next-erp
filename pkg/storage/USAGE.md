## pkg/storage - MinIO 对象存储使用文档

### 一、快速开始

#### 1.1 初始化客户端

```go
package main

import (
    "context"
    "github.com/lk2023060901/go-next-erp/pkg/storage"
)

func main() {
    ctx := context.Background()

    // 方式 1：使用默认配置
    s, err := storage.New(ctx)
    if err != nil {
        panic(err)
    }
    defer s.Close()

    // 方式 2：使用选项配置
    s, err = storage.New(ctx,
        storage.WithEndpoint("localhost:15002"),
        storage.WithCredentials("minioadmin", "minioadmin123"),
        storage.WithSSL(false),
        storage.WithBucket("my-bucket"),
    )
}
```

#### 1.2 环境变量配置

```env
MINIO_ENDPOINT=localhost:15002
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_USE_SSL=false
MINIO_BUCKET=documents
```

---

### 二、文件操作

#### 2.1 上传文件

```go
// 从文件上传
file, err := os.Open("example.pdf")
if err != nil {
    panic(err)
}
defer file.Close()

stat, _ := file.Stat()
err = s.UploadFile(ctx, "docs", "example.pdf", file, stat.Size(), "application/pdf")

// 从字节流上传
content := []byte("Hello, MinIO!")
reader := bytes.NewReader(content)
err = s.UploadFile(ctx, "docs", "hello.txt", reader, int64(len(content)), "text/plain")
```

#### 2.2 下载文件

```go
// 下载到本地文件
err := s.DownloadFile(ctx, "docs", "example.pdf", "/tmp/example.pdf")
```

#### 2.3 删除文件

```go
err := s.DeleteFile(ctx, "docs", "example.pdf")
```

#### 2.4 检查文件是否存在

```go
exists, err := s.FileExists(ctx, "docs", "example.pdf")
if exists {
    fmt.Println("File exists")
}
```

#### 2.5 获取文件信息

```go
info, err := s.GetFileInfo(ctx, "docs", "example.pdf")
fmt.Printf("Size: %d bytes\n", info.Size)
fmt.Printf("ETag: %s\n", info.ETag)
fmt.Printf("ContentType: %s\n", info.ContentType)
fmt.Printf("LastModified: %s\n", info.LastModified)
```

#### 2.6 列出文件

```go
// 列出指定前缀的所有文件
files, err := s.ListFiles(ctx, "docs", "pdf/")
for _, file := range files {
    fmt.Printf("%s - %d bytes\n", file.Key, file.Size)
}
```

---

### 三、预签名 URL

#### 3.1 生成下载 URL

```go
// 生成有效期 7 天的下载 URL
url, err := s.GetPresignedURL(ctx, "docs", "example.pdf", 7*24*time.Hour)
fmt.Println("Download URL:", url)
```

#### 3.2 使用场景

- 前端直接上传文件
- 临时分享文件链接
- 第三方服务访问私有文件

---

### 四、存储桶管理

#### 4.1 创建存储桶

```go
err := s.CreateBucket(ctx, "my-new-bucket")
```

#### 4.2 删除存储桶

```go
err := s.DeleteBucket(ctx, "my-old-bucket")
```

#### 4.3 检查存储桶是否存在

```go
exists, err := s.BucketExists(ctx, "my-bucket")
```

#### 4.4 列出所有存储桶

```go
buckets, err := s.ListBuckets(ctx)
for _, bucket := range buckets {
    fmt.Printf("%s - Created: %s\n", bucket.Name, bucket.CreationDate)
}
```

---

### 五、完整示例

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/lk2023060901/go-next-erp/pkg/storage"
)

func main() {
    ctx := context.Background()

    // 初始化
    s, err := storage.New(ctx,
        storage.WithEndpoint("localhost:15002"),
        storage.WithCredentials("minioadmin", "minioadmin123"),
        storage.WithBucket("documents"),
    )
    if err != nil {
        panic(err)
    }
    defer s.Close()

    // 上传文件
    file, _ := os.Open("example.pdf")
    defer file.Close()

    stat, _ := file.Stat()
    err = s.UploadFile(ctx, "", "docs/example.pdf", file, stat.Size(), "application/pdf")
    if err != nil {
        panic(err)
    }
    fmt.Println("File uploaded successfully")

    // 获取预签名 URL
    url, _ := s.GetPresignedURL(ctx, "", "docs/example.pdf", 7*24*time.Hour)
    fmt.Println("Download URL:", url)

    // 列出文件
    files, _ := s.ListFiles(ctx, "", "docs/")
    for _, file := range files {
        fmt.Printf("- %s (%d bytes)\n", file.Key, file.Size)
    }

    // 下载文件
    err = s.DownloadFile(ctx, "", "docs/example.pdf", "/tmp/example.pdf")
    if err != nil {
        panic(err)
    }
    fmt.Println("File downloaded successfully")

    // 删除文件
    err = s.DeleteFile(ctx, "", "docs/example.pdf")
    if err != nil {
        panic(err)
    }
    fmt.Println("File deleted successfully")
}
```

---

### 六、错误处理

```go
import "github.com/minio/minio-go/v7"

// 检查特定错误
_, err := s.GetFileInfo(ctx, "bucket", "key")
if err != nil {
    errResponse := minio.ToErrorResponse(err)
    if errResponse.Code == "NoSuchKey" {
        fmt.Println("File not found")
    }
}
```

---

### 七、最佳实践

#### 7.1 使用默认桶

```go
// 传空字符串使用默认桶
s.UploadFile(ctx, "", "file.txt", reader, size, "text/plain")
```

#### 7.2 文件命名规范

```go
// 推荐：使用路径分隔符组织文件
s.UploadFile(ctx, "docs", "2024/01/report.pdf", ...)

// 避免：使用特殊字符
// s.UploadFile(ctx, "docs", "报告@2024.pdf", ...)
```

#### 7.3 批量操作

```go
// 批量上传
files := []string{"file1.txt", "file2.txt", "file3.txt"}
for _, filename := range files {
    file, _ := os.Open(filename)
    stat, _ := file.Stat()
    s.UploadFile(ctx, "batch", filename, file, stat.Size(), "text/plain")
    file.Close()
}
```

---

### 八、配置参考

| 选项 | 说明 | 默认值 |
|------|------|--------|
| Endpoint | MinIO 地址 | localhost:9000 |
| AccessKeyID | 访问密钥 | minioadmin |
| SecretAccessKey | 密钥 | minioadmin |
| UseSSL | 是否使用 SSL | false |
| BucketName | 默认存储桶 | default |
| Region | 区域 | us-east-1 |
| MaxRetries | 最大重试次数 | 3 |
| Timeout | 超时时间 | 30s |
