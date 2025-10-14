package websocket

import (
	"sync"

	"github.com/google/uuid"
)

// Client 代表一个 WebSocket 客户端连接
type Client struct {
	UserID   uuid.UUID   // 用户ID
	TenantID uuid.UUID   // 租户ID
	Send     chan []byte // 发送消息通道
	Hub      *Hub        // 所属Hub
}

// Hub 管理所有活跃的 WebSocket 连接
type Hub struct {
	// 用户ID -> 客户端连接映射（一个用户可能有多个连接）
	clients map[uuid.UUID]map[*Client]bool

	// 注册新客户端
	register chan *Client

	// 注销客户端
	unregister chan *Client

	// 广播消息到特定用户
	broadcast chan *BroadcastMessage

	// 互斥锁
	mu sync.RWMutex
}

// BroadcastMessage 广播消息
type BroadcastMessage struct {
	UserID  uuid.UUID
	Message []byte
}

// NewHub 创建新的 Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
	}
}

// Run 启动 Hub（需要在 goroutine 中运行）
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[message.UserID]
			h.mu.RUnlock()

			for client := range clients {
				select {
				case client.Send <- message.Message:
				default:
					// 发送失败，关闭连接
					h.mu.Lock()
					close(client.Send)
					delete(h.clients[client.UserID], client)
					if len(h.clients[client.UserID]) == 0 {
						delete(h.clients, client.UserID)
					}
					h.mu.Unlock()
				}
			}
		}
	}
}

// SendToUser 发送消息给指定用户的所有连接
func (h *Hub) SendToUser(userID uuid.UUID, message []byte) {
	h.broadcast <- &BroadcastMessage{
		UserID:  userID,
		Message: message,
	}
}

// GetOnlineUsers 获取在线用户列表
func (h *Hub) GetOnlineUsers() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]uuid.UUID, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// IsUserOnline 检查用户是否在线
func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.clients[userID]
	return exists
}

// GetUserConnectionCount 获取用户的连接数
func (h *Hub) GetUserConnectionCount(userID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID])
}
