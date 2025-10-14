package sse

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TargetType 消息目标类型
type TargetType string

const (
	TargetUser      TargetType = "user"      // 指定用户
	TargetTopic     TargetType = "topic"     // 指定主题
	TargetBroadcast TargetType = "broadcast" // 广播
)

// EventType SSE 事件类型
type EventType string

const (
	EventMessage   EventType = "message"   // 普通消息
	EventHeartbeat EventType = "heartbeat" // 心跳
	EventError     EventType = "error"     // 错误
	EventClose     EventType = "close"     // 关闭连接
)

// Client SSE 客户端连接
type Client struct {
	// 客户端ID（自动生成）
	ID uuid.UUID

	// 用户ID（可选，用于用户级消息推送）
	UserID uuid.UUID

	// 租户ID（可选，用于多租户隔离）
	TenantID uuid.UUID

	// 订阅的主题列表
	Topics []string

	// 发送消息通道
	send chan *Message

	// Broker 引用
	broker *Broker

	// 最后活跃时间
	lastSeen time.Time

	// 元数据（自定义字段）
	Metadata map[string]interface{}
}

// NewClient 创建新客户端
func NewClient(broker *Broker, userID, tenantID uuid.UUID, topics []string) *Client {
	return &Client{
		ID:       uuid.New(),
		UserID:   userID,
		TenantID: tenantID,
		Topics:   topics,
		send:     make(chan *Message, broker.config.ClientBufferSize),
		broker:   broker,
		lastSeen: time.Now(),
		Metadata: make(map[string]interface{}),
	}
}

// close 关闭客户端连接
func (c *Client) close() {
	select {
	case <-c.send:
		// 通道已关闭
	default:
		close(c.send)
	}
}

// Message SSE 消息
type Message struct {
	// 消息ID（自动生成）
	ID string

	// 目标类型
	Target TargetType

	// 目标用户ID（当 Target = TargetUser 时使用）
	UserID uuid.UUID

	// 目标主题（当 Target = TargetTopic 时使用）
	Topic string

	// 事件类型（默认：message）
	Event EventType

	// 消息数据
	Data string

	// 重试时间（毫秒，默认：3000）
	Retry int

	// 创建时间
	CreatedAt time.Time
}

// NewMessage 创建新消息
func NewMessage(event EventType, data string) *Message {
	return &Message{
		ID:        uuid.New().String(),
		Event:     event,
		Data:      data,
		Retry:     3000,
		CreatedAt: time.Now(),
	}
}

// Format 格式化为 SSE 格式
func (m *Message) Format() string {
	result := ""

	if m.ID != "" {
		result += "id: " + m.ID + "\n"
	}

	if m.Event != "" {
		result += "event: " + string(m.Event) + "\n"
	}

	if m.Retry > 0 {
		result += fmt.Sprintf("retry: %d\n", m.Retry)
	}

	if m.Data != "" {
		result += "data: " + m.Data + "\n"
	}

	result += "\n"
	return result
}
