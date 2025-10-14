# SSE 通用模块 - 快速开始

## 目录

- [安装](#安装)
- [基础用法](#基础用法)
- [集成到项目](#集成到项目)
- [配置说明](#配置说明)
- [API 参考](#api-参考)
- [常见问题](#常见问题)

## 安装

SSE 模块已作为 `pkg/sse` 包内置在项目中，无需额外安装。

## 基础用法

### 1. 创建并启动 Broker

```go
package main

import (
    "context"
    "github.com/lk2023060901/go-next-erp/pkg/sse"
)

func main() {
    // 创建 Broker（使用默认配置）
    broker := sse.NewBroker(sse.DefaultBrokerConfig())
    
    // 启动 Broker
    ctx := context.Background()
    go broker.Start(ctx)
    
    // ... 业务逻辑
}
```

### 2. 创建 HTTP Handler

```go
// 创建 Handler（使用默认认证）
handler := sse.NewHandler(
    broker,
    sse.DefaultAuthenticator,   // 从查询参数获取 user_id
    sse.DefaultTopicResolver,   // 从查询参数获取 topics
)

// 注册路由
http.HandleFunc("/sse/stream", handler.ServeHTTP)
```

### 3. 发送消息

```go
import "github.com/google/uuid"

// 发送给指定用户
userID := uuid.MustParse("user-uuid-here")
broker.SendToUser(userID, "notification", `{"title":"新消息","content":"Hello"}`)

// 发送给主题订阅者
broker.SendToTopic("approval:pending", "task", `{"task_id":"123"}`)

// 广播给所有用户
broker.Broadcast("announcement", `{"message":"系统维护通知"}`)
```

### 4. 客户端连接

```javascript
// JavaScript EventSource API
const eventSource = new EventSource('/sse/stream?user_id=your-user-id');

eventSource.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log('收到消息:', data);
};

// 监听特定事件
eventSource.addEventListener('notification', (event) => {
    const data = JSON.parse(event.data);
    showNotification(data);
});
```

## 集成到项目

### 步骤 1: 修改 Wire 配置

在 `pkg/wire.go` 中添加 SSE Provider:

```go
var ProviderSet = wire.NewSet(
    // ... 现有 Providers
    sse.ProviderSet,  // 添加 SSE
)
```

### 步骤 2: 在 HTTP 服务器中注册路由

修改 `internal/server/http.go`:

```go
func NewHTTPServer(
    // ... 现有参数
    sseBroker *sse.Broker,
    sseHandler *sse.Handler,
) *http.Server {
    // 启动 SSE Broker
    go sseBroker.Start(context.Background())
    
    // 注册 SSE 路由
    srv.HandleFunc("/api/v1/sse/stream", sseHandler.ServeHTTP)
    
    return srv
}
```

### 步骤 3: 在业务服务中注入 Broker

以通知服务为例，修改 `internal/notification/service/notification_service.go`:

```go
type notificationService struct {
    repo      repository.NotificationRepository
    sseBroker *sse.Broker  // 注入 SSE Broker
}

func NewNotificationService(
    repo repository.NotificationRepository,
    sseBroker *sse.Broker,
) NotificationService {
    return &notificationService{
        repo:      repo,
        sseBroker: sseBroker,
    }
}
```

### 步骤 4: 发送通知时实时推送

```go
func (s *notificationService) SendNotification(ctx context.Context, req *dto.SendNotificationRequest) error {
    // 1. 创建通知记录
    notification := &model.Notification{...}
    s.repo.Create(ctx, notification)
    
    // 2. 通过 SSE 实时推送
    data, _ := json.Marshal(notification)
    s.sseBroker.SendToUser(notification.RecipientID, "notification", string(data))
    
    return nil
}
```

## 配置说明

### BrokerConfig 字段

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| ClientBufferSize | int | 256 | 客户端消息缓冲区大小 |
| HeartbeatInterval | Duration | 30s | 心跳间隔 |
| ClientTimeout | Duration | 5m | 客户端超时时间 |
| EnableTopics | bool | true | 是否启用主题订阅 |
| MaxConnections | int | 0 | 最大连接数（0=无限制） |
| HistorySize | int | 0 | 历史消息保留数量（暂未实现） |

### 自定义配置示例

```go
config := &sse.BrokerConfig{
    ClientBufferSize:  512,              // 高并发场景
    HeartbeatInterval: 30 * time.Second,
    ClientTimeout:     10 * time.Minute, // 移动网络适配
    EnableTopics:      true,
    MaxConnections:    10000,            // 限制最大连接数
    HistorySize:       0,
}

broker := sse.NewBroker(config)
```

## API 参考

### Broker 方法

#### SendToUser
```go
func (b *Broker) SendToUser(userID uuid.UUID, event, data string) error
```
发送消息给指定用户的所有连接。

**参数**:
- `userID`: 用户 UUID（不能为 Nil）
- `event`: 事件类型（如 "notification"）
- `data`: 消息数据（建议使用 JSON 字符串）

**返回**: 
- `error`: 如果 userID 无效或队列满则返回错误

#### SendToTopic
```go
func (b *Broker) SendToTopic(topic, event, data string) error
```
发送消息给订阅指定主题的所有客户端。

**参数**:
- `topic`: 主题名称（不能为空）
- `event`: 事件类型
- `data`: 消息数据

**返回**: 
- `error`: 如果主题功能未启用或主题为空则返回错误

#### Broadcast
```go
func (b *Broker) Broadcast(event, data string) error
```
广播消息给所有连接的客户端。

#### GetStats
```go
func (b *Broker) GetStats() *Stats
```
获取当前统计信息。

**返回**:
```go
type Stats struct {
    TotalClients  int  // 总客户端数
    TotalUsers    int  // 总用户数
    TotalTopics   int  // 总主题数
    IsRunning     bool // 是否运行中
    BufferSize    int  // 缓冲区大小
    MaxConnection int  // 最大连接数
}
```

#### IsUserOnline
```go
func (b *Broker) IsUserOnline(userID uuid.UUID) bool
```
检查用户是否在线。

#### GetOnlineUsers
```go
func (b *Broker) GetOnlineUsers() []uuid.UUID
```
获取所有在线用户的 UUID 列表。

### Handler 认证函数

#### 自定义 JWT 认证

```go
func JWTAuthenticator(jwtManager *jwt.Manager) sse.Authenticator {
    return func(r *http.Request) (uuid.UUID, uuid.UUID, error) {
        token := r.URL.Query().Get("token")
        if token == "" {
            return uuid.Nil, uuid.Nil, fmt.Errorf("missing token")
        }
        
        claims, err := jwtManager.ValidateToken(token)
        if err != nil {
            return uuid.Nil, uuid.Nil, err
        }
        
        return claims.UserID, claims.TenantID, nil
    }
}

// 使用
handler := sse.NewHandler(broker, JWTAuthenticator(jwtManager), sse.DefaultTopicResolver)
```

## 常见问题

### Q1: 消息发送后客户端收不到？

**A**: 检查以下几点：
1. Broker 是否已启动（`go broker.Start(ctx)`）
2. 用户是否已连接（`broker.IsUserOnline(userID)`）
3. Nginx 是否禁用了缓冲（见文档性能优化章节）
4. 浏览器控制台是否有错误

### Q2: 连接后立即断开？

**A**: 可能原因：
1. 认证失败（检查 token 或 user_id）
2. 服务器不支持 SSE（检查 Flusher）
3. 查看服务器日志获取详细错误

### Q3: 如何限制每个用户的连接数？

**A**: 当前一个用户可以有多个连接（多设备/多标签页）。如需限制，可在认证函数中实现：

```go
func LimitedAuthenticator(broker *sse.Broker, maxPerUser int) sse.Authenticator {
    return func(r *http.Request) (uuid.UUID, uuid.UUID, error) {
        userID, tenantID, err := sse.DefaultAuthenticator(r)
        if err != nil {
            return uuid.Nil, uuid.Nil, err
        }
        
        // 检查当前用户连接数
        if broker.GetUserConnectionCount(userID) >= maxPerUser {
            return uuid.Nil, uuid.Nil, fmt.Errorf("max connections exceeded")
        }
        
        return userID, tenantID, nil
    }
}
```

### Q4: 如何持久化历史消息？

**A**: 当前版本不支持历史消息持久化。建议方案：
1. 在数据库中保存消息记录
2. 客户端连接时通过 REST API 拉取历史消息
3. SSE 仅用于实时推送新消息

### Q5: 如何与 Redis 集成实现集群？

**A**: 建议扩展方案（需自行实现）：

```go
// 使用 Redis Pub/Sub
type RedisBroker struct {
    *sse.Broker
    redisClient *redis.Client
}

func (rb *RedisBroker) SendToUser(userID uuid.UUID, event, data string) error {
    // 1. 发送到本地 Broker
    rb.Broker.SendToUser(userID, event, data)
    
    // 2. 发布到 Redis（其他实例订阅）
    msg := map[string]interface{}{
        "user_id": userID,
        "event":   event,
        "data":    data,
    }
    rb.redisClient.Publish(ctx, "sse:user:"+userID.String(), msg)
    
    return nil
}
```

## 测试

### 运行单元测试

```bash
go test -v ./pkg/sse/...
```

### 性能基准测试

```bash
go test -bench=. ./pkg/sse/...
```

### 浏览器测试

```bash
# 启动服务器
go run cmd/server/main.go

# 打开测试页面
open test_sse.html
```

## 下一步

- 查看 [完整文档](docs/SSE_MODULE.md)
- 查看 [集成示例](pkg/sse/examples/integration.go)
- 查看 [测试代码](pkg/sse/broker_test.go)

## 支持

如有问题，请提交 Issue 或查看项目文档。
