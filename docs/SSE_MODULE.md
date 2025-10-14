# SSE (Server-Sent Events) 通用模块

## 概述

本模块提供了一个**独立、可复用的 SSE 实时推送解决方案**，支持多用户、主题订阅、心跳保持等完整功能。其他业务模块（通知、审批流程、任务进度等）可直接使用此模块实现实时推送。

## 技术架构

```
┌─────────────────────┐
│  HTTP Client        │
│  (EventSource API)  │
└──────────┬──────────┘
           │
           │ GET /sse/stream?user_id=xxx&topics=a,b
           │
           ↓
┌─────────────────────────────────────────┐
│         SSE Handler                     │
│  (认证 + 连接管理 + 消息推送)           │
└──────────┬──────────────────────────────┘
           │
           ↓
┌─────────────────────────────────────────┐
│          SSE Broker                     │
│  (连接池 + 主题订阅 + 消息广播)         │
└──────────┬──────────────────────────────┘
           │
           ↓
    业务模块调用
  (SendToUser/SendToTopic/Broadcast)
```

## 核心组件

### 1. Broker (消息代理)

**文件**: `pkg/sse/broker.go`

**职责**:
- 管理所有 SSE 客户端连接
- 按用户ID和主题分组管理
- 消息路由与广播
- 心跳保持与超时检测
- 统计信息收集

**关键方法**:

| 方法 | 参数 | 返回值 | 说明 | 默认值/验证规则 |
|------|------|--------|------|----------------|
| `NewBroker(config)` | config: BrokerConfig | *Broker | 创建新的 Broker | config=nil 时使用默认配置 |
| `Start(ctx)` | ctx: Context | void | 启动 Broker（需在 goroutine 中） | - |
| `Stop()` | - | void | 停止 Broker | - |
| `SendToUser(userID, event, data)` | userID: UUID, event: string, data: string | error | 发送消息给指定用户 | userID 不能为 Nil |
| `SendToTopic(topic, event, data)` | topic: string, event: string, data: string | error | 发送消息给主题订阅者 | topic 不能为空，需启用主题 |
| `Broadcast(event, data)` | event: string, data: string | error | 广播消息给所有客户端 | - |
| `GetStats()` | - | *Stats | 获取统计信息 | - |
| `IsUserOnline(userID)` | userID: UUID | bool | 检查用户是否在线 | - |
| `GetOnlineUsers()` | - | []UUID | 获取在线用户列表 | - |

### 2. BrokerConfig (配置)

**字段默认值表**:

| 字段 | 类型 | 默认值 | 说明 | 验证规则 | 性能建议 |
|------|------|--------|------|---------|---------|
| `ClientBufferSize` | int | 256 | 客户端消息缓冲区大小 | ≥ 1 | 高并发场景建议 512-1024 |
| `HeartbeatInterval` | Duration | 30s | 心跳间隔 | ≥ 5s | 平衡心跳频率与服务器负载 |
| `ClientTimeout` | Duration | 5m | 客户端超时时间 | ≥ HeartbeatInterval×2 | 移动网络建议 10m |
| `EnableTopics` | bool | true | 是否启用主题订阅 | - | 不需要主题功能可禁用 |
| `MaxConnections` | int | 0 | 最大连接数（0=无限制） | ≥ 0 | 根据服务器资源设置 |
| `HistorySize` | int | 0 | 历史消息保留数量 | ≥ 0 | 暂未实现，预留字段 |

**配置示例**:

```go
// 默认配置（推荐用于开发环境）
config := sse.DefaultBrokerConfig()

// 生产环境配置
config := &sse.BrokerConfig{
    ClientBufferSize:  512,           // 增大缓冲区
    HeartbeatInterval: 30 * time.Second,
    ClientTimeout:     10 * time.Minute, // 适应移动网络
    EnableTopics:      true,
    MaxConnections:    10000,         // 限制最大连接数
    HistorySize:       0,
}
```

### 3. Handler (HTTP 处理器)

**文件**: `pkg/sse/handler.go`

**职责**:
- 处理 SSE 连接请求
- 客户端认证
- 设置 SSE 响应头
- 消息推送与刷新

**关键方法**:

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `NewHandler(broker, auth, resolver)` | broker: *Broker, auth: Authenticator, resolver: TopicResolver | *Handler | 创建 Handler |
| `ServeHTTP(w, r)` | w: ResponseWriter, r: *Request | void | 处理 HTTP 请求 |
| `WriteEvent(w, event, data)` | w: ResponseWriter, event: string, data: string | error | 直接写入事件（非 Broker） |

**认证函数类型**:

```go
// Authenticator 认证函数
// 返回 (userID, tenantID, error)
type Authenticator func(r *http.Request) (uuid.UUID, uuid.UUID, error)
```

**主题解析函数类型**:

```go
// TopicResolver 主题解析函数
// 从请求中解析订阅的主题列表
type TopicResolver func(r *http.Request) []string
```

**默认实现**:

| 函数 | 说明 | 参数来源 | 验证规则 |
|------|------|---------|---------|
| `DefaultAuthenticator` | 从查询参数获取用户信息 | `?user_id=xxx&tenant_id=xxx` | user_id 必需且为有效 UUID |
| `DefaultTopicResolver` | 从查询参数获取主题列表 | `?topics=topic1,topic2` | 逗号分隔，可选 |

### 4. Message (消息)

**字段默认值表**:

| 字段 | 类型 | 默认值 | 说明 | 是否必需 |
|------|------|--------|------|---------|
| `ID` | string | UUID | 消息唯一标识 | 自动生成 |
| `Target` | TargetType | - | 目标类型（user/topic/broadcast） | 是 |
| `UserID` | UUID | Nil | 目标用户ID（Target=user时） | 条件必需 |
| `Topic` | string | "" | 目标主题（Target=topic时） | 条件必需 |
| `Event` | EventType | "message" | 事件类型 | 否 |
| `Data` | string | "" | 消息数据（JSON字符串） | 是 |
| `Retry` | int | 3000 | 重试时间（毫秒） | 否 |
| `CreatedAt` | Time | Now() | 创建时间 | 自动生成 |

**事件类型**:

| 类型 | 值 | 说明 | 使用场景 |
|------|-----|------|---------|
| `EventMessage` | "message" | 普通消息 | 业务数据推送 |
| `EventHeartbeat` | "heartbeat" | 心跳 | 保持连接活跃 |
| `EventError` | "error" | 错误 | 错误通知 |
| `EventClose` | "close" | 关闭 | 主动关闭连接 |

### 5. Client (客户端连接)

**字段说明**:

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `ID` | UUID | 客户端唯一标识 | 自动生成 |
| `UserID` | UUID | 用户ID | 来自认证 |
| `TenantID` | UUID | 租户ID | 来自认证 |
| `Topics` | []string | 订阅的主题列表 | 来自请求 |
| `send` | chan *Message | 发送消息通道 | 容量=ClientBufferSize |
| `lastSeen` | Time | 最后活跃时间 | 自动更新 |
| `Metadata` | map[string]interface{} | 自定义元数据 | 空map |

## 使用方法

### 基本使用（单例模式）

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/lk2023060901/go-next-erp/pkg/sse"
)

func main() {
    // 1. 创建 Broker
    broker := sse.NewBroker(sse.DefaultBrokerConfig())
    
    // 2. 启动 Broker
    ctx := context.Background()
    go broker.Start(ctx)
    
    // 3. 创建 Handler（使用默认认证和主题解析）
    handler := sse.NewHandler(broker, sse.DefaultAuthenticator, sse.DefaultTopicResolver)
    
    // 4. 注册 HTTP 路由
    http.HandleFunc("/sse/stream", handler.ServeHTTP)
    
    // 5. 启动 HTTP 服务器
    http.ListenAndServe(":8080", nil)
}
```

### 自定义认证（JWT Token）

```go
import (
    "fmt"
    "net/http"
    
    "github.com/google/uuid"
    "github.com/lk2023060901/go-next-erp/internal/auth/authentication/jwt"
    "github.com/lk2023060901/go-next-erp/pkg/sse"
)

// JWTAuthenticator 基于 JWT 的认证函数
func JWTAuthenticator(jwtManager *jwt.Manager) sse.Authenticator {
    return func(r *http.Request) (uuid.UUID, uuid.UUID, error) {
        // 1. 从查询参数或 Header 获取 token
        token := r.URL.Query().Get("token")
        if token == "" {
            token = r.Header.Get("Authorization")
            if len(token) > 7 && token[:7] == "Bearer " {
                token = token[7:]
            }
        }
        
        if token == "" {
            return uuid.Nil, uuid.Nil, fmt.Errorf("missing token")
        }
        
        // 2. 验证 token
        claims, err := jwtManager.ValidateToken(token)
        if err != nil {
            return uuid.Nil, uuid.Nil, fmt.Errorf("invalid token: %w", err)
        }
        
        // 3. 返回用户ID和租户ID
        return claims.UserID, claims.TenantID, nil
    }
}

// 使用自定义认证
func NewCustomHandler(broker *sse.Broker, jwtManager *jwt.Manager) *sse.Handler {
    return sse.NewHandler(
        broker,
        JWTAuthenticator(jwtManager),
        sse.DefaultTopicResolver,
    )
}
```

### 发送消息

```go
// 1. 发送给指定用户
userID := uuid.MustParse("user-uuid-here")
err := broker.SendToUser(userID, "notification", `{"title":"新消息","content":"您有一条新通知"}`)

// 2. 发送给主题订阅者
err := broker.SendToTopic("approval:pending", "task", `{"task_id":"123","type":"approval"}`)

// 3. 广播给所有客户端
err := broker.Broadcast("announcement", `{"message":"系统维护通知"}`)
```

### 客户端连接（JavaScript）

```javascript
// 1. 基本连接
const eventSource = new EventSource('/sse/stream?user_id=your-user-id');

eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('收到消息:', data);
};

eventSource.onerror = (error) => {
  console.error('SSE 错误:', error);
  eventSource.close();
};

// 2. 监听特定事件
eventSource.addEventListener('notification', (event) => {
  const data = JSON.parse(event.data);
  showNotification(data);
});

eventSource.addEventListener('heartbeat', (event) => {
  console.log('心跳:', event.data);
});

// 3. 订阅主题
const eventSource = new EventSource('/sse/stream?user_id=xxx&topics=approval:pending,task:assigned');

// 4. 带 JWT Token
const token = 'your-jwt-token';
const eventSource = new EventSource(`/sse/stream?token=${token}`);
```

## 集成到业务模块

### 示例：通知模块集成

**步骤1**: 在 NotificationService 中注入 SSE Broker

```go
// internal/notification/service/notification_service.go
type notificationService struct {
    repo      repository.NotificationRepository
    sseBroker *sse.Broker  // 添加 SSE Broker
}

func NewNotificationService(repo repository.NotificationRepository, sseBroker *sse.Broker) NotificationService {
    return &notificationService{
        repo:      repo,
        sseBroker: sseBroker,
    }
}
```

**步骤2**: 创建通知时实时推送

```go
func (s *notificationService) SendNotification(ctx context.Context, req *dto.SendNotificationRequest) error {
    // 1. 创建通知记录
    notification := &model.Notification{...}
    if err := s.repo.Create(ctx, notification); err != nil {
        return err
    }
    
    // 2. 通过 SSE 实时推送
    data, _ := json.Marshal(notification)
    s.sseBroker.SendToUser(notification.RecipientID, "notification", string(data))
    
    return nil
}
```

**步骤3**: 注册 SSE 路由

```go
// internal/server/http.go
func NewHTTPServer(..., sseBroker *sse.Broker, sseHandler *sse.Handler) *http.Server {
    // ... 其他配置
    
    // 启动 SSE Broker
    go sseBroker.Start(context.Background())
    
    // 注册 SSE 路由
    srv.HandleFunc("/api/v1/notifications/stream", sseHandler.ServeHTTP)
    
    return srv
}
```

### 示例：审批流程实时推送

```go
// 审批状态变更时推送
func (s *approvalService) UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status string) error {
    // 1. 更新审批任务
    task, err := s.repo.UpdateStatus(ctx, taskID, status)
    if err != nil {
        return err
    }
    
    // 2. 推送给审批人
    data, _ := json.Marshal(map[string]interface{}{
        "task_id": taskID,
        "status":  status,
        "updated_at": time.Now(),
    })
    
    s.sseBroker.SendToUser(task.ApproverID, "approval:status_changed", string(data))
    
    // 3. 推送给主题订阅者（如：所有关注该审批流程的人）
    topic := fmt.Sprintf("approval:process:%s", task.ProcessID)
    s.sseBroker.SendToTopic(topic, "approval:status_changed", string(data))
    
    return nil
}
```

## 错误处理

### 错误类型

| 错误 | 场景 | 处理方式 |
|------|------|---------|
| "invalid user ID" | UserID 为 Nil | 返回 400 Bad Request |
| "authentication failed" | Token 无效或过期 | 返回 401 Unauthorized |
| "topics not enabled" | 配置禁用主题但调用 SendToTopic | 返回 500 Internal Error |
| "invalid topic" | Topic 为空字符串 | 返回 400 Bad Request |
| "broadcast channel full" | 消息队列满 | 丢弃消息，记录日志 |
| "SSE not supported" | 浏览器不支持 Flusher | 返回 500 Internal Error |

### 错误处理示例

```go
// 业务代码中的错误处理
if err := broker.SendToUser(userID, "event", data); err != nil {
    log.Error("Failed to send SSE message",
        zap.String("user_id", userID.String()),
        zap.Error(err),
    )
    // 不影响主流程，继续执行
}
```

## 性能优化

### 1. 缓冲区调优

```go
// 高并发场景
config := &sse.BrokerConfig{
    ClientBufferSize: 1024,  // 增大缓冲区，减少阻塞
}
```

### 2. 心跳优化

```go
// 移动网络场景（延长心跳间隔）
config := &sse.BrokerConfig{
    HeartbeatInterval: 60 * time.Second,
    ClientTimeout:     10 * time.Minute,
}
```

### 3. 连接数限制

```go
// 保护服务器资源
config := &sse.BrokerConfig{
    MaxConnections: 10000,  // 根据服务器配置设置
}
```

### 4. Nginx 配置

```nginx
location /sse/ {
    proxy_pass http://backend;
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    
    # 禁用缓冲
    proxy_buffering off;
    proxy_cache off;
    
    # 超时设置
    proxy_read_timeout 24h;
    proxy_send_timeout 24h;
    
    # 关闭 gzip
    gzip off;
}
```

## 监控与日志

### 统计信息

```go
stats := broker.GetStats()
fmt.Printf("总客户端数: %d\n", stats.TotalClients)
fmt.Printf("总用户数: %d\n", stats.TotalUsers)
fmt.Printf("总主题数: %d\n", stats.TotalTopics)
fmt.Printf("是否运行: %v\n", stats.IsRunning)
```

### 日志记录（建议）

```go
// 在业务代码中记录关键操作
log.Info("SSE message sent",
    zap.String("target", "user"),
    zap.String("user_id", userID.String()),
    zap.String("event", "notification"),
    zap.Int("data_size", len(data)),
)
```

### Prometheus 指标（建议扩展）

```go
// 建议添加的监控指标
// - sse_active_connections_total
// - sse_messages_sent_total
// - sse_messages_dropped_total
// - sse_client_connect_total
// - sse_client_disconnect_total
```

## 测试

### 单元测试示例

```go
func TestBroker_SendToUser(t *testing.T) {
    broker := sse.NewBroker(sse.DefaultBrokerConfig())
    ctx := context.Background()
    go broker.Start(ctx)
    defer broker.Stop()
    
    userID := uuid.New()
    client := sse.NewClient(broker, userID, uuid.Nil, nil)
    broker.register <- client
    
    // 发送消息
    err := broker.SendToUser(userID, "test", "hello")
    assert.NoError(t, err)
    
    // 接收消息
    select {
    case msg := <-client.send:
        assert.Equal(t, "test", string(msg.Event))
        assert.Equal(t, "hello", msg.Data)
    case <-time.After(1 * time.Second):
        t.Fatal("timeout waiting for message")
    }
}
```

### 集成测试

```bash
# 启动服务器
go run cmd/server/main.go

# 测试连接
curl -N -H "Accept: text/event-stream" \
  "http://localhost:8080/sse/stream?user_id=your-uuid"

# 发送测试消息（另一个终端）
curl -X POST http://localhost:8080/api/test/send \
  -d '{"user_id":"your-uuid","event":"test","data":"hello"}'
```

## 安全性

### 1. 认证保护

- **必须**: 实现认证函数，验证客户端身份
- **建议**: 使用 JWT Token 或 Session 认证
- **禁止**: 使用明文传输敏感信息

### 2. 租户隔离

```go
// 多租户场景下验证租户归属
func (b *Broker) SendToUser(userID uuid.UUID, event, data string) error {
    // 验证 userID 和 tenantID 关联
    // ...
}
```

### 3. CORS 配置

```go
// 生产环境需配置正确的 Origin
w.Header().Set("Access-Control-Allow-Origin", "https://your-domain.com")
```

### 4. 速率限制

```go
// 建议在 Handler 层实现速率限制
// 防止恶意客户端频繁连接
```

## 故障排查

### 问题1: 客户端连接后立即断开

**原因**: 认证失败或不支持 SSE

**解决**:
1. 检查认证参数是否正确
2. 检查浏览器是否支持 EventSource
3. 查看服务器日志

### 问题2: 消息无法接收

**原因**: Nginx 缓冲或客户端过滤

**解决**:
1. 配置 Nginx 禁用缓冲（见性能优化章节）
2. 检查客户端事件监听是否正确
3. 确认消息已成功发送（查看 GetStats）

### 问题3: 内存占用过高

**原因**: 连接数过多或缓冲区过大

**解决**:
1. 设置 `MaxConnections` 限制
2. 减小 `ClientBufferSize`
3. 实现连接数监控和告警

## 与 WebSocket 对比

| 特性 | SSE | WebSocket |
|------|-----|-----------|
| 通信方向 | 单向（服务器→客户端） | 双向 |
| 协议 | HTTP | WS/WSS |
| 浏览器支持 | 广泛（除IE） | 广泛 |
| 实现复杂度 | 简单 | 中等 |
| 重连机制 | 自动 | 需手动实现 |
| 二进制数据 | 不支持 | 支持 |
| 代理友好度 | 高 | 中 |
| 使用场景 | 实时推送、通知、进度 | 聊天、游戏、协作 |

**选择建议**:
- **SSE**: 通知推送、任务进度、实时监控等**单向推送场景**
- **WebSocket**: 聊天、在线协作、游戏等**双向交互场景**

## 总结

✅ **已实现功能**:
- 独立的 SSE 模块（pkg/sse）
- 连接管理（Broker）
- 用户级消息推送
- 主题订阅与发布
- 心跳保持
- 超时检测
- 统计信息
- 可扩展的认证机制
- Wire 依赖注入支持

✅ **特性**:
- 零依赖（仅使用标准库 + UUID）
- 并发安全
- 配置灵活
- 易于集成
- 完整文档

✅ **适用场景**:
- 通知推送
- 审批流程状态推送
- 任务进度实时更新
- 系统公告广播
- 数据监控大屏
- 日志实时流

🎯 **后续优化方向**:
- 历史消息重放
- Redis Pub/Sub 集群支持
- Prometheus 监控指标
- 速率限制
- 消息持久化
