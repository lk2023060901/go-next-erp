package sse

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Handler SSE HTTP 处理器
type Handler struct {
	broker *Broker

	// 认证函数（可选）
	authenticator Authenticator

	// 主题解析函数（可选，从请求中解析订阅主题）
	topicResolver TopicResolver
}

// Authenticator 认证函数类型
// 返回 (userID, tenantID, error)
type Authenticator func(r *http.Request) (uuid.UUID, uuid.UUID, error)

// TopicResolver 主题解析函数类型
// 从请求中解析订阅的主题列表
type TopicResolver func(r *http.Request) []string

// NewHandler 创建 SSE Handler
func NewHandler(broker *Broker, authenticator Authenticator, topicResolver TopicResolver) *Handler {
	return &Handler{
		broker:        broker,
		authenticator: authenticator,
		topicResolver: topicResolver,
	}
}

// ServeHTTP 处理 SSE 连接请求
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 检查是否支持 SSE
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// 2. 认证（如果配置了认证函数）
	var userID, tenantID uuid.UUID
	var err error

	if h.authenticator != nil {
		userID, tenantID, err = h.authenticator(r)
		if err != nil {
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}
	}

	// 3. 解析订阅主题（如果配置了主题解析函数）
	var topics []string
	if h.topicResolver != nil {
		topics = h.topicResolver(r)
	}

	// 4. 设置 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // 生产环境需配置
	w.Header().Set("X-Accel-Buffering", "no")          // 禁用 Nginx 缓冲

	// 5. 创建客户端并注册
	client := NewClient(h.broker, userID, tenantID, topics)
	h.broker.register <- client

	// 6. 发送欢迎消息
	welcomeMsg := NewMessage(EventMessage, fmt.Sprintf(`{"type":"welcome","client_id":"%s","timestamp":"%s"}`,
		client.ID, time.Now().Format(time.RFC3339)))
	h.writeMessage(w, flusher, welcomeMsg)

	// 7. 监听消息并推送
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			// 客户端断开连接
			h.broker.unregister <- client
			return

		case message, ok := <-client.send:
			if !ok {
				// 通道已关闭
				h.broker.unregister <- client
				return
			}

			// 写入消息
			if err := h.writeMessage(w, flusher, message); err != nil {
				h.broker.unregister <- client
				return
			}
		}
	}
}

// writeMessage 写入 SSE 消息
func (h *Handler) writeMessage(w http.ResponseWriter, flusher http.Flusher, message *Message) error {
	// 格式化消息
	formatted := message.Format()

	// 写入响应
	if _, err := fmt.Fprint(w, formatted); err != nil {
		return err
	}

	// 立即刷新
	flusher.Flush()
	return nil
}

// ServeHTTPWithContext 带上下文的 HTTP 处理（支持自定义超时）
func (h *Handler) ServeHTTPWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// 合并上下文
	r = r.WithContext(ctx)
	h.ServeHTTP(w, r)
}

// DefaultAuthenticator 默认认证函数（从查询参数获取 user_id）
func DefaultAuthenticator(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("missing user_id parameter")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid user_id: %w", err)
	}

	// 租户ID可选
	tenantIDStr := r.URL.Query().Get("tenant_id")
	var tenantID uuid.UUID
	if tenantIDStr != "" {
		tenantID, err = uuid.Parse(tenantIDStr)
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("invalid tenant_id: %w", err)
		}
	}

	return userID, tenantID, nil
}

// DefaultTopicResolver 默认主题解析函数（从查询参数获取 topics）
func DefaultTopicResolver(r *http.Request) []string {
	topicsParam := r.URL.Query().Get("topics")
	if topicsParam == "" {
		return nil
	}

	// 支持多个主题，逗号分隔
	topics := []string{}
	for _, topic := range splitByComma(topicsParam) {
		if topic != "" {
			topics = append(topics, topic)
		}
	}
	return topics
}

// splitByComma 按逗号分割字符串
func splitByComma(s string) []string {
	result := []string{}
	current := ""
	for _, c := range s {
		if c == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// WriteEvent 辅助函数：直接写入 SSE 事件（用于非 Broker 场景）
func WriteEvent(w http.ResponseWriter, event, data string) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("SSE not supported")
	}

	msg := NewMessage(EventType(event), data)
	if _, err := fmt.Fprint(w, msg.Format()); err != nil {
		return err
	}

	flusher.Flush()
	return nil
}

// WriteEventWithRetry 写入带重试时间的事件
func WriteEventWithRetry(w http.ResponseWriter, event, data string, retryMs int) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("SSE not supported")
	}

	msg := NewMessage(EventType(event), data)
	msg.Retry = retryMs

	if _, err := fmt.Fprint(w, msg.Format()); err != nil {
		return err
	}

	flusher.Flush()
	return nil
}

// SetSSEHeaders 设置 SSE 响应头（辅助函数）
func SetSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")
}

// ParseRetryParam 从请求中解析重试时间参数
func ParseRetryParam(r *http.Request, defaultValue int) int {
	retryStr := r.URL.Query().Get("retry")
	if retryStr == "" {
		return defaultValue
	}

	retry, err := strconv.Atoi(retryStr)
	if err != nil || retry < 0 {
		return defaultValue
	}

	return retry
}
