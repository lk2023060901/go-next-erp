package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/jwt"
	"github.com/lk2023060901/go-next-erp/internal/notification/service"
)

const (
	// 允许向对等方写入消息的时间
	writeWait = 10 * time.Second

	// 允许从对等方读取下一条消息的时间
	pongWait = 60 * time.Second

	// 在此期间向对等方发送 ping
	pingPeriod = (pongWait * 9) / 10

	// 消息的最大大小
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 在生产环境中应该检查 Origin
		return true
	},
}

// Handler WebSocket 处理器
type Handler struct {
	hub           *Hub
	jwtManager    *jwt.Manager
	notifyService service.NotificationService
}

// NewHandler 创建 WebSocket 处理器
func NewHandler(hub *Hub, jwtManager *jwt.Manager, notifyService service.NotificationService) *Handler {
	return &Handler{
		hub:           hub,
		jwtManager:    jwtManager,
		notifyService: notifyService,
	}
}

// ServeHTTP 处理 WebSocket 连接请求
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 从查询参数或 Header 获取 token
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	// 验证 token
	claims, err := h.jwtManager.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// 用户ID和租户ID已经在 claims 中
	userID := claims.UserID
	tenantID := claims.TenantID

	// 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// 创建客户端
	client := &Client{
		UserID:   userID,
		TenantID: tenantID,
		Send:     make(chan []byte, 256),
		Hub:      h.hub,
	}

	// 注册客户端
	h.hub.register <- client

	// 启动读写协程
	go h.writePump(client, conn)
	go h.readPump(client, conn)

	// 发送欢迎消息和未读数量
	h.sendWelcomeMessage(client)
}

// readPump 从 WebSocket 连接读取消息
func (h *Handler) readPump(client *Client, conn *websocket.Conn) {
	defer func() {
		h.hub.unregister <- client
		conn.Close()
	}()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log error
			}
			break
		}

		// 处理客户端消息（如心跳、已读确认等）
		h.handleClientMessage(client, message)
	}
}

// writePump 向 WebSocket 连接写入消息
func (h *Handler) writePump(client *Client, conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 关闭了通道
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 将排队的消息添加到当前 WebSocket 消息中
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// sendWelcomeMessage 发送欢迎消息
func (h *Handler) sendWelcomeMessage(client *Client) {
	ctx := context.Background()

	// 获取未读数量
	unreadCount, err := h.notifyService.CountUnread(ctx, client.UserID)
	if err != nil {
		unreadCount = 0
	}

	message := map[string]interface{}{
		"type":         "welcome",
		"message":      "Connected to notification service",
		"unread_count": unreadCount,
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	data, _ := json.Marshal(message)
	select {
	case client.Send <- data:
	default:
	}
}

// handleClientMessage 处理客户端消息
func (h *Handler) handleClientMessage(client *Client, message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "ping":
		// 响应心跳
		pong := map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now().Format(time.RFC3339),
		}
		data, _ := json.Marshal(pong)
		select {
		case client.Send <- data:
		default:
		}

	case "mark_read":
		// 处理标记已读
		// TODO: 实现标记已读逻辑
	}
}

// SendNotificationToUser 发送通知给用户
func (h *Handler) SendNotificationToUser(userID uuid.UUID, notification map[string]interface{}) error {
	notification["type"] = "notification"
	notification["timestamp"] = time.Now().Format(time.RFC3339)

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	h.hub.SendToUser(userID, data)
	return nil
}
