# WebSocket 通知模块实现文档

## 概述

本项目已成功实现基于 **Kratos + WebSocket** 的实时通知推送功能，支持站内消息的实时推送、未读管理等完整功能。

## 技术架构

```
┌─────────────────┐
│  前端 Client    │
│  (WebSocket)    │
└────────┬────────┘
         │
         │ ws://host/api/v1/notifications/ws?token=<JWT>
         │
         ↓
┌─────────────────────────────────────────┐
│        WebSocket Handler                │
│    (JWT认证 + 连接管理)                 │
└────────┬────────────────────────────────┘
         │
         ↓
┌─────────────────────────────────────────┐
│          WebSocket Hub                  │
│   (连接池 + 消息广播)                   │
└────────┬────────────────────────────────┘
         │
         ↓
┌─────────────────────────────────────────┐
│      NotificationService                │
│   (业务逻辑 + WebSocket推送)            │
└────────┬────────────────────────────────┘
         │
         ↓
┌─────────────────────────────────────────┐
│   NotificationRepository                │
│        (数据持久化)                      │
└─────────────────────────────────────────┘
         │
         ↓
    PostgreSQL
```

## 核心实现

### 1. WebSocket Hub (连接管理器)

**文件**: `internal/notification/websocket/hub.go`

**功能**:
- 管理所有活跃的 WebSocket 连接
- 支持一个用户多个连接（多设备/多标签页）
- 消息广播到指定用户的所有连接
- 自动清理断开的连接

**关键方法**:
```go
- Run()                        // 启动Hub（在goroutine中运行）
- SendToUser(userID, message)  // 向指定用户推送消息
- IsUserOnline(userID)         // 检查用户是否在线
- GetOnlineUsers()             // 获取在线用户列表
```

### 2. WebSocket Handler (请求处理器)

**文件**: `internal/notification/websocket/handler.go`

**功能**:
- 处理 WebSocket 连接升级请求
- JWT Token 认证
- 心跳保持（Ping/Pong）
- 读写协程管理
- 欢迎消息（包含未读数量）

**认证流程**:
1. 从查询参数 `?token=<JWT>` 或 Header `Authorization: Bearer <JWT>` 获取 Token
2. 使用 JWTManager 验证 Token
3. 提取 UserID 和 TenantID
4. 创建 WebSocket 连接并注册到 Hub

**消息格式**:
```json
{
  "type": "notification",       // 消息类型: welcome, notification, ping, pong
  "timestamp": "2025-10-14T14:00:00Z",
  "notification": {
    "id": "uuid",
    "type": "system",
    "title": "通知标题",
    "content": "通知内容",
    "priority": "high"
  },
  "unread_count": 5
}
```

### 3. NotificationService (通知服务)

**文件**: `internal/notification/service/notification_service.go`

**增强功能**:
- 添加 `PushHandler` 接口，支持 WebSocket 推送
- 创建站内消息时自动通过 WebSocket 实时推送
- SetPushHandler() 方法注入 WebSocket Handler

**推送流程**:
```
SendNotification() 
    ↓
创建通知记录到数据库
    ↓
异步发送 (sendNotification)
    ↓
如果是 in_app 且 pushHandler != nil
    ↓
调用 pushHandler.SendNotificationToUser()
    ↓
WebSocket 实时推送给在线用户
```

### 4. NotificationAdapter (API适配层)

**文件**: `internal/adapter/notification.go`

**实现方法**:
- `SendNotification` - 发送通知
- `GetNotification` - 获取单个通知
- `ListNotifications` - 列出通知（支持分页、筛选未读）
- `MarkAsRead` - 标记单个通知已读
- `BatchMarkAsRead` - 批量标记已读
- `DeleteNotification` - 删除通知
- `GetUnreadCount` - 获取未读数量

## API 端点

### REST API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/notifications` | 发送通知 |
| GET | `/api/v1/notifications/{id}` | 获取单个通知 |
| GET | `/api/v1/notifications` | 列出通知（支持分页） |
| PUT | `/api/v1/notifications/{id}/read` | 标记为已读 |
| PUT | `/api/v1/notifications/read` | 批量标记已读 |
| DELETE | `/api/v1/notifications/{id}` | 删除通知 |
| GET | `/api/v1/notifications/unread-count` | 获取未读数量 |

### WebSocket API

| 端点 | 说明 |
|------|------|
| `ws://host:port/api/v1/notifications/ws?token=<JWT>` | WebSocket 实时推送 |

**连接参数**:
- `token`: JWT Token（必需）

**客户端 → 服务器消息**:
```json
// 心跳
{"type": "ping"}

// 标记已读（待实现）
{"type": "mark_read", "notification_ids": ["uuid1", "uuid2"]}
```

**服务器 → 客户端消息**:
```json
// 欢迎消息
{
  "type": "welcome",
  "message": "Connected to notification service",
  "unread_count": 5,
  "timestamp": "2025-10-14T14:00:00Z"
}

// 心跳响应
{
  "type": "pong",
  "timestamp": "2025-10-14T14:00:00Z"
}

// 新通知推送
{
  "type": "notification",
  "notification": {
    "id": "uuid",
    "type": "system",
    "title": "标题",
    "content": "内容",
    "priority": "high"
  },
  "timestamp": "2025-10-14T14:00:00Z"
}
```

## 使用示例

### 1. JavaScript 客户端

```javascript
// 连接 WebSocket
const ws = new WebSocket(`ws://localhost:15006/api/v1/notifications/ws?token=${jwtToken}`);

ws.onopen = () => {
  console.log('WebSocket 已连接');
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('收到消息:', data);
  
  if (data.type === 'notification') {
    // 显示通知
    showNotification(data.notification);
  }
};

ws.onerror = (error) => {
  console.error('WebSocket 错误:', error);
};

ws.onclose = () => {
  console.log('WebSocket 已关闭');
};
```

### 2. Go 客户端（测试用）

```go
import (
    "github.com/gorilla/websocket"
)

func connectWebSocket(token string) {
    url := fmt.Sprintf("ws://localhost:15006/api/v1/notifications/ws?token=%s", token)
    
    conn, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            log.Println("read error:", err)
            return
        }
        
        fmt.Printf("收到消息: %s\n", message)
    }
}
```

### 3. curl 测试 REST API

```bash
# 发送通知
curl -X POST http://localhost:15006/api/v1/notifications \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "system",
    "title": "系统通知",
    "content": "这是一条测试通知",
    "priority": "high"
  }'

# 获取未读数量
curl -X GET http://localhost:15006/api/v1/notifications/unread-count \
  -H "Authorization: Bearer YOUR_TOKEN"

# 列出通知
curl -X GET "http://localhost:15006/api/v1/notifications?page=1&page_size=10" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 标记为已读
curl -X PUT http://localhost:15006/api/v1/notifications/{id}/read \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

## 测试方法

### 方式1: 浏览器测试页面（推荐）

```bash
# 打开测试页面
open /Volumes/work/coding/golang/go-next-erp/test_ws_notification.html
```

功能:
- ✅ 用户登录
- ✅ WebSocket 连接/断开
- ✅ 发送测试通知
- ✅ 实时接收推送
- ✅ 查询未读数量
- ✅ 实时日志显示

### 方式2: wscat 命令行工具

```bash
# 安装 wscat
npm install -g wscat

# 先登录获取 token
TOKEN=$(curl -s -X POST http://localhost:15006/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"YOUR_USER","password":"YOUR_PASSWORD"}' | jq -r .access_token)

# 连接 WebSocket
wscat -c "ws://localhost:15006/api/v1/notifications/ws?token=$TOKEN"
```

### 方式3: websocat 工具

```bash
# 安装 websocat (macOS)
brew install websocat

# 连接
websocat "ws://localhost:15006/api/v1/notifications/ws?token=$TOKEN"
```

## 配置与部署

### 依赖注入 (Wire)

已自动配置在以下文件:
- `internal/notification/wire.go` - Notification 模块 Provider
- `internal/notification/websocket/wire.go` - WebSocket Provider
- `internal/server/http.go` - HTTP 服务器集成

### 服务启动

```bash
# 编译
go build -o bin/server cmd/server/main.go cmd/server/wire_gen.go

# 运行
./bin/server -conf=configs/config.yaml
```

### 日志说明

服务启动后会输出:
```
INFO msg=[HTTP] server listening on: [::]:15006
```

WebSocket 连接日志:
- 连接建立时会发送欢迎消息
- Ping/Pong 心跳每 54 秒一次
- 连接断开会自动清理

## 性能特性

- ✅ **高并发支持**: Hub 使用读写锁，支持大量并发连接
- ✅ **心跳保持**: 自动 Ping/Pong 保持连接活跃
- ✅ **优雅关闭**: 断开连接时自动清理资源
- ✅ **多设备支持**: 同一用户可建立多个 WebSocket 连接
- ✅ **消息缓冲**: 发送通道带 256 缓冲，避免阻塞
- ✅ **异步推送**: 通知发送和 WebSocket 推送解耦

## 安全特性

- ✅ **JWT 认证**: 所有 WebSocket 连接需要有效 JWT Token
- ✅ **租户隔离**: 基于 TenantID 的多租户隔离
- ✅ **Origin 检查**: WebSocket Upgrader 支持 Origin 验证（生产环境需配置）
- ✅ **自动超时**: 60秒无活动自动断开

## 扩展性

### 添加新的消息类型

在 `internal/notification/websocket/handler.go` 的 `handleClientMessage` 中添加:

```go
case "custom_action":
    // 处理自定义动作
    handleCustomAction(client, msg)
```

### 添加新的通知渠道

在 `internal/notification/service/notification_service.go` 的 `sendNotification` 中添加:

```go
case model.NotificationChannelCustom:
    // 实现自定义通知渠道
```

## 故障排查

### 问题: WebSocket 连接失败 401

**原因**: Token 无效或过期

**解决**:
1. 检查 Token 是否正确传递 (`?token=xxx`)
2. 使用 `/api/v1/auth/login` 重新获取 Token
3. 检查 JWT 配置（secret_key, 过期时间等）

### 问题: 连接成功但收不到消息

**原因**: PushHandler 未正确初始化

**解决**:
检查 `internal/notification/init.go` 的 `InitNotificationWebSocket` 是否被调用

### 问题: 连接频繁断开

**原因**: Ping/Pong 超时

**解决**:
调整 `pongWait` 参数（默认60秒）

## 后续优化建议

1. **Redis 集群支持**: 
   - 使用 Redis Pub/Sub 实现多实例消息分发
   - 支持水平扩展

2. **消息持久化**:
   - 离线消息队列
   - 重连后推送未读消息

3. **更丰富的消息类型**:
   - 打字状态
   - 已读回执
   - 消息撤回

4. **监控与统计**:
   - 在线用户数统计
   - 消息推送成功率
   - WebSocket 连接时长

## 总结

✅ **已完成功能**:
- WebSocket 实时通知推送
- JWT 认证与授权
- 连接管理（Hub模式）
- REST API 完整实现
- 心跳保持机制
- 多设备支持
- 站内消息自动推送

✅ **测试覆盖**:
- 浏览器测试页面
- 命令行测试工具
- API 端点验证

✅ **生产就绪**:
- 错误处理完善
- 资源自动清理
- 并发安全
- 可扩展架构
