package sse

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Broker SSE 消息代理（连接管理器）
type Broker struct {
	// 客户端连接池（按用户ID分组）
	clients map[uuid.UUID]map[*Client]bool

	// 主题订阅（主题 -> 客户端列表）
	topics map[string]map[*Client]bool

	// 新客户端注册通道
	register chan *Client

	// 客户端注销通道
	unregister chan *Client

	// 广播消息通道
	broadcast chan *Message

	// 配置
	config *BrokerConfig

	// 互斥锁
	mu sync.RWMutex

	// 停止信号
	stop chan struct{}

	// 运行状态
	running bool
}

// BrokerConfig Broker 配置
type BrokerConfig struct {
	// 客户端缓冲区大小（默认：256）
	ClientBufferSize int

	// 心跳间隔（默认：30秒）
	HeartbeatInterval time.Duration

	// 客户端超时时间（默认：5分钟）
	ClientTimeout time.Duration

	// 是否启用主题订阅（默认：true）
	EnableTopics bool

	// 最大连接数（0表示无限制，默认：0）
	MaxConnections int

	// 历史消息保留数量（默认：0，不保留）
	HistorySize int
}

// DefaultBrokerConfig 默认配置
func DefaultBrokerConfig() *BrokerConfig {
	return &BrokerConfig{
		ClientBufferSize:  256,
		HeartbeatInterval: 30 * time.Second,
		ClientTimeout:     5 * time.Minute,
		EnableTopics:      true,
		MaxConnections:    0,
		HistorySize:       0,
	}
}

// NewBroker 创建新的 SSE Broker
func NewBroker(config *BrokerConfig) *Broker {
	if config == nil {
		config = DefaultBrokerConfig()
	}

	return &Broker{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		topics:     make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, config.ClientBufferSize),
		config:     config,
		stop:       make(chan struct{}),
		running:    false,
	}
}

// Start 启动 Broker（在 goroutine 中运行）
func (b *Broker) Start(ctx context.Context) {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return
	}
	b.running = true
	b.mu.Unlock()

	// 启动心跳定时器
	heartbeatTicker := time.NewTicker(b.config.HeartbeatInterval)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			b.shutdown()
			return

		case <-b.stop:
			b.shutdown()
			return

		case client := <-b.register:
			b.registerClient(client)

		case client := <-b.unregister:
			b.unregisterClient(client)

		case message := <-b.broadcast:
			b.broadcastMessage(message)

		case <-heartbeatTicker.C:
			b.sendHeartbeat()
		}
	}
}

// Stop 停止 Broker
func (b *Broker) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return
	}

	close(b.stop)
	b.running = false
}

// registerClient 注册客户端
func (b *Broker) registerClient(client *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 检查最大连接数限制
	if b.config.MaxConnections > 0 && b.getTotalConnections() >= b.config.MaxConnections {
		client.close()
		return
	}

	// 按用户ID分组
	if client.UserID != uuid.Nil {
		if b.clients[client.UserID] == nil {
			b.clients[client.UserID] = make(map[*Client]bool)
		}
		b.clients[client.UserID][client] = true
	}

	// 订阅主题
	if b.config.EnableTopics {
		for _, topic := range client.Topics {
			if b.topics[topic] == nil {
				b.topics[topic] = make(map[*Client]bool)
			}
			b.topics[topic][client] = true
		}
	}

	// 更新最后活跃时间
	client.lastSeen = time.Now()
}

// unregisterClient 注销客户端
func (b *Broker) unregisterClient(client *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 从用户组移除
	if clients, ok := b.clients[client.UserID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(b.clients, client.UserID)
		}
	}

	// 从主题移除
	for _, topic := range client.Topics {
		if clients, ok := b.topics[topic]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(b.topics, topic)
			}
		}
	}

	client.close()
}

// broadcastMessage 广播消息
func (b *Broker) broadcastMessage(message *Message) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var targets []*Client

	// 根据目标类型选择客户端
	switch message.Target {
	case TargetUser:
		// 发送给指定用户
		if clients, ok := b.clients[message.UserID]; ok {
			for client := range clients {
				targets = append(targets, client)
			}
		}

	case TargetTopic:
		// 发送给订阅主题的客户端
		if clients, ok := b.topics[message.Topic]; ok {
			for client := range clients {
				targets = append(targets, client)
			}
		}

	case TargetBroadcast:
		// 广播给所有客户端
		for _, clients := range b.clients {
			for client := range clients {
				targets = append(targets, client)
			}
		}
	}

	// 发送消息
	for _, client := range targets {
		select {
		case client.send <- message:
			client.lastSeen = time.Now()
		default:
			// 发送失败，断开连接
			b.unregister <- client
		}
	}
}

// sendHeartbeat 发送心跳
func (b *Broker) sendHeartbeat() {
	b.mu.RLock()
	defer b.mu.RUnlock()

	now := time.Now()
	heartbeat := &Message{
		Event: EventHeartbeat,
		Data:  "ping",
	}

	for _, clients := range b.clients {
		for client := range clients {
			// 检查超时
			if now.Sub(client.lastSeen) > b.config.ClientTimeout {
				b.unregister <- client
				continue
			}

			// 发送心跳
			select {
			case client.send <- heartbeat:
				// 心跳发送成功
			default:
				// 发送失败，断开连接
				b.unregister <- client
			}
		}
	}
}

// shutdown 关闭所有连接
func (b *Broker) shutdown() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 关闭所有客户端
	for _, clients := range b.clients {
		for client := range clients {
			client.close()
		}
	}

	// 清空映射
	b.clients = make(map[uuid.UUID]map[*Client]bool)
	b.topics = make(map[string]map[*Client]bool)
}

// SendToUser 发送消息给指定用户
func (b *Broker) SendToUser(userID uuid.UUID, event, data string) error {
	if userID == uuid.Nil {
		return fmt.Errorf("invalid user ID")
	}

	message := &Message{
		Target: TargetUser,
		UserID: userID,
		Event:  EventType(event),
		Data:   data,
	}

	select {
	case b.broadcast <- message:
		return nil
	default:
		return fmt.Errorf("broadcast channel full")
	}
}

// SendToTopic 发送消息给主题订阅者
func (b *Broker) SendToTopic(topic, event, data string) error {
	if !b.config.EnableTopics {
		return fmt.Errorf("topics not enabled")
	}

	if topic == "" {
		return fmt.Errorf("invalid topic")
	}

	message := &Message{
		Target: TargetTopic,
		Topic:  topic,
		Event:  EventType(event),
		Data:   data,
	}

	select {
	case b.broadcast <- message:
		return nil
	default:
		return fmt.Errorf("broadcast channel full")
	}
}

// Broadcast 广播消息给所有客户端
func (b *Broker) Broadcast(event, data string) error {
	message := &Message{
		Target: TargetBroadcast,
		Event:  EventType(event),
		Data:   data,
	}

	select {
	case b.broadcast <- message:
		return nil
	default:
		return fmt.Errorf("broadcast channel full")
	}
}

// GetStats 获取统计信息
func (b *Broker) GetStats() *Stats {
	b.mu.RLock()
	defer b.mu.RUnlock()

	totalClients := 0
	for _, clients := range b.clients {
		totalClients += len(clients)
	}

	return &Stats{
		TotalClients:  totalClients,
		TotalUsers:    len(b.clients),
		TotalTopics:   len(b.topics),
		IsRunning:     b.running,
		BufferSize:    b.config.ClientBufferSize,
		MaxConnection: b.config.MaxConnections,
	}
}

// getTotalConnections 获取总连接数（内部使用，需已加锁）
func (b *Broker) getTotalConnections() int {
	total := 0
	for _, clients := range b.clients {
		total += len(clients)
	}
	return total
}

// IsUserOnline 检查用户是否在线
func (b *Broker) IsUserOnline(userID uuid.UUID) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	_, exists := b.clients[userID]
	return exists
}

// GetOnlineUsers 获取在线用户列表
func (b *Broker) GetOnlineUsers() []uuid.UUID {
	b.mu.RLock()
	defer b.mu.RUnlock()

	users := make([]uuid.UUID, 0, len(b.clients))
	for userID := range b.clients {
		if userID != uuid.Nil {
			users = append(users, userID)
		}
	}
	return users
}

// Stats 统计信息
type Stats struct {
	TotalClients  int  // 总客户端数
	TotalUsers    int  // 总用户数
	TotalTopics   int  // 总主题数
	IsRunning     bool // 是否运行中
	BufferSize    int  // 缓冲区大小
	MaxConnection int  // 最大连接数
}
