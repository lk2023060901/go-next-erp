package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/jwt"
	"github.com/lk2023060901/go-next-erp/internal/notification/dto"
	"github.com/lk2023060901/go-next-erp/internal/notification/service"
	"github.com/lk2023060901/go-next-erp/pkg/sse"
)

// ============================================
// 示例 1: 通知模块集成 SSE
// ============================================

// NotificationServiceWithSSE 集成 SSE 的通知服务
type NotificationServiceWithSSE struct {
	service.NotificationService
	sseBroker *sse.Broker
}

// SendNotification 发送通知（重写，添加 SSE 推送）
func (s *NotificationServiceWithSSE) SendNotification(ctx context.Context, tenantID uuid.UUID, req *dto.SendNotificationRequest) (*dto.NotificationResponse, error) {
	// 1. 调用原有服务创建通知
	notif, err := s.NotificationService.SendNotification(ctx, tenantID, req)
	if err != nil {
		return nil, err
	}

	// 2. 通过 SSE 实时推送
	recipientID, _ := uuid.Parse(req.RecipientID)
	data, _ := json.Marshal(map[string]interface{}{
		"id":       notif.ID,
		"type":     notif.Type,
		"title":    notif.Title,
		"content":  notif.Content,
		"priority": notif.Priority,
	})

	s.sseBroker.SendToUser(recipientID, "notification", string(data))

	return notif, nil
}

// ============================================
// 示例 2: SSE 服务器设置
// ============================================

// SetupSSEServer 设置 SSE 服务器
func SetupSSEServer(jwtManager *jwt.Manager, notifService service.NotificationService) *http.ServeMux {
	// 1. 创建 SSE Broker
	config := &sse.BrokerConfig{
		ClientBufferSize:  512,
		HeartbeatInterval: 30 * time.Second,
		ClientTimeout:     10 * time.Minute,
		EnableTopics:      true,
		MaxConnections:    10000,
		HistorySize:       0,
	}
	broker := sse.NewBroker(config)

	// 2. 启动 Broker
	ctx := context.Background()
	go broker.Start(ctx)

	// 3. 创建自定义认证函数
	authenticator := func(r *http.Request) (uuid.UUID, uuid.UUID, error) {
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

	// 4. 创建 SSE Handler
	handler := sse.NewHandler(broker, authenticator, sse.DefaultTopicResolver)

	// 5. 集成通知服务与 SSE
	_ = &NotificationServiceWithSSE{
		NotificationService: notifService,
		sseBroker:           broker,
	}

	// 6. 注册路由
	mux := http.NewServeMux()

	// SSE 流端点
	mux.HandleFunc("/api/v1/sse/stream", handler.ServeHTTP)

	// 测试发送端点
	mux.HandleFunc("/api/v1/sse/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			TargetType string `json:"target_type"` // user, topic, broadcast
			UserID     string `json:"user_id,omitempty"`
			Topic      string `json:"topic,omitempty"`
			Event      string `json:"event"`
			Data       string `json:"data"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var err error
		switch req.TargetType {
		case "user":
			userID, _ := uuid.Parse(req.UserID)
			err = broker.SendToUser(userID, req.Event, req.Data)
		case "topic":
			err = broker.SendToTopic(req.Topic, req.Event, req.Data)
		case "broadcast":
			err = broker.Broadcast(req.Event, req.Data)
		default:
			http.Error(w, "invalid target_type", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Message sent successfully",
		})
	})

	// 统计信息端点
	mux.HandleFunc("/api/v1/sse/stats", func(w http.ResponseWriter, r *http.Request) {
		stats := broker.GetStats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	return mux
}

// ============================================
// 示例 3: 审批流程实时推送
// ============================================

// ApprovalService 审批服务示例
type ApprovalService struct {
	sseBroker *sse.Broker
}

// UpdateTaskStatus 更新审批任务状态（带 SSE 推送）
func (s *ApprovalService) UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, approverID uuid.UUID, status string) error {
	// 1. 更新数据库状态（省略具体实现）
	// ...

	// 2. 推送给审批人
	data, _ := json.Marshal(map[string]interface{}{
		"task_id":    taskID,
		"status":     status,
		"updated_at": time.Now(),
	})

	s.sseBroker.SendToUser(approverID, "approval:status_changed", string(data))

	// 3. 推送给主题订阅者（如：流程关注者）
	topic := fmt.Sprintf("approval:task:%s", taskID)
	s.sseBroker.SendToTopic(topic, "approval:status_changed", string(data))

	return nil
}

// ============================================
// 示例 4: 任务进度实时推送
// ============================================

// TaskProgressTracker 任务进度追踪器
type TaskProgressTracker struct {
	sseBroker *sse.Broker
}

// UpdateProgress 更新任务进度
func (t *TaskProgressTracker) UpdateProgress(userID uuid.UUID, taskID string, progress int, message string) error {
	data, _ := json.Marshal(map[string]interface{}{
		"task_id":   taskID,
		"progress":  progress,
		"message":   message,
		"timestamp": time.Now(),
	})

	return t.sseBroker.SendToUser(userID, "task:progress", string(data))
}

// ============================================
// 示例 5: 系统公告广播
// ============================================

// AnnouncementService 公告服务
type AnnouncementService struct {
	sseBroker *sse.Broker
}

// BroadcastAnnouncement 广播系统公告
func (s *AnnouncementService) BroadcastAnnouncement(title, content string, priority string) error {
	data, _ := json.Marshal(map[string]interface{}{
		"title":    title,
		"content":  content,
		"priority": priority,
		"time":     time.Now(),
	})

	return s.sseBroker.Broadcast("announcement", string(data))
}

// ============================================
// 示例 6: 多主题订阅
// ============================================

// MultiTopicExample 多主题订阅示例
func MultiTopicExample() {
	// 客户端连接时订阅多个主题
	// URL: /api/v1/sse/stream?user_id=xxx&topics=notification,approval:pending,task:assigned

	// 服务端向不同主题发送消息
	broker := sse.NewBroker(sse.DefaultBrokerConfig())
	go broker.Start(context.Background())

	// 发送通知主题消息
	broker.SendToTopic("notification", "new_message", `{"count":5}`)

	// 发送审批主题消息
	broker.SendToTopic("approval:pending", "task_assigned", `{"task_id":"123"}`)

	// 发送任务主题消息
	broker.SendToTopic("task:assigned", "task_update", `{"task_id":"456","status":"in_progress"}`)
}

// ============================================
// 示例 7: 简单的独立 SSE 服务
// ============================================

// SimpleSSEServer 简单的独立 SSE 服务器
func SimpleSSEServer() {
	// 1. 创建 Broker
	broker := sse.NewBroker(sse.DefaultBrokerConfig())
	go broker.Start(context.Background())

	// 2. 创建 Handler（使用默认认证）
	handler := sse.NewHandler(broker, sse.DefaultAuthenticator, sse.DefaultTopicResolver)

	// 3. 注册路由
	http.HandleFunc("/sse", handler.ServeHTTP)

	// 4. 测试推送端点
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		userID, _ := uuid.Parse(r.URL.Query().Get("user_id"))
		message := r.URL.Query().Get("message")

		broker.SendToUser(userID, "message", message)
		fmt.Fprintf(w, "Message sent to user %s", userID)
	})

	// 5. 启动服务器
	fmt.Println("SSE Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}
