package sse
package sse_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/pkg/sse"
	"github.com/stretchr/testify/assert"
)

// TestBroker_Basic 基本功能测试
func TestBroker_Basic(t *testing.T) {
	broker := sse.NewBroker(sse.DefaultBrokerConfig())
	ctx := context.Background()
	go broker.Start(ctx)
	defer broker.Stop()

	// 等待 Broker 启动
	time.Sleep(100 * time.Millisecond)

	stats := broker.GetStats()
	assert.True(t, stats.IsRunning)
	assert.Equal(t, 0, stats.TotalClients)
}

// TestBroker_SendToUser 用户消息推送测试
func TestBroker_SendToUser(t *testing.T) {
	broker := sse.NewBroker(sse.DefaultBrokerConfig())
	ctx := context.Background()
	go broker.Start(ctx)
	defer broker.Stop()

	userID := uuid.New()
	client := sse.NewClient(broker, userID, uuid.Nil, nil)
	broker.register <- client

	// 等待注册完成
	time.Sleep(50 * time.Millisecond)

	// 发送消息
	err := broker.SendToUser(userID, "test_event", `{"message":"hello"}`)
	assert.NoError(t, err)

	// 接收消息
	select {
	case msg := <-client.send:
		assert.Equal(t, "test_event", string(msg.Event))
		assert.Equal(t, `{"message":"hello"}`, msg.Data)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

// TestBroker_SendToTopic 主题订阅测试
func TestBroker_SendToTopic(t *testing.T) {
	broker := sse.NewBroker(sse.DefaultBrokerConfig())
	ctx := context.Background()
	go broker.Start(ctx)
	defer broker.Stop()

	// 创建两个订阅同一主题的客户端
	topic := "test_topic"
	client1 := sse.NewClient(broker, uuid.New(), uuid.Nil, []string{topic})
	client2 := sse.NewClient(broker, uuid.New(), uuid.Nil, []string{topic})

	broker.Register(client1)
	broker.Register(client2)

	time.Sleep(50 * time.Millisecond)

	// 发送主题消息
	err := broker.SendToTopic(topic, "topic_event", "topic_data")
	assert.NoError(t, err)

	// 两个客户端都应该收到消息
	receivedCount := 0
	timeout := time.After(1 * time.Second)

	for receivedCount < 2 {
		select {
		case <-client1.Send():
			receivedCount++
		case <-client2.Send():
			receivedCount++
		case <-timeout:
			t.Fatalf("only received %d messages, expected 2", receivedCount)
		}
	}

	assert.Equal(t, 2, receivedCount)
}

// TestBroker_Broadcast 广播测试
func TestBroker_Broadcast(t *testing.T) {
	broker := sse.NewBroker(sse.DefaultBrokerConfig())
	ctx := context.Background()
	go broker.Start(ctx)
	defer broker.Stop()

	// 创建3个客户端
	client1 := sse.NewClient(broker, uuid.New(), uuid.Nil, nil)
	client2 := sse.NewClient(broker, uuid.New(), uuid.Nil, nil)
	client3 := sse.NewClient(broker, uuid.New(), uuid.Nil, nil)

	broker.Register(client1)
	broker.Register(client2)
	broker.Register(client3)

	time.Sleep(50 * time.Millisecond)

	// 广播消息
	err := broker.Broadcast("announcement", "system maintenance")
	assert.NoError(t, err)

	// 所有客户端都应该收到消息
	receivedCount := 0
	timeout := time.After(1 * time.Second)

	for receivedCount < 3 {
		select {
		case <-client1.Send():
			receivedCount++
		case <-client2.Send():
			receivedCount++
		case <-client3.Send():
			receivedCount++
		case <-timeout:
			t.Fatalf("only received %d messages, expected 3", receivedCount)
		}
	}

	assert.Equal(t, 3, receivedCount)
}

// TestBroker_IsUserOnline 在线状态测试
func TestBroker_IsUserOnline(t *testing.T) {
	broker := sse.NewBroker(sse.DefaultBrokerConfig())
	ctx := context.Background()
	go broker.Start(ctx)
	defer broker.Stop()

	userID := uuid.New()

	// 用户未连接
	assert.False(t, broker.IsUserOnline(userID))

	// 用户连接
	client := sse.NewClient(broker, userID, uuid.Nil, nil)
	broker.Register(client)
	time.Sleep(50 * time.Millisecond)

	assert.True(t, broker.IsUserOnline(userID))

	// 用户断开
	broker.Unregister(client)
	time.Sleep(50 * time.Millisecond)

	assert.False(t, broker.IsUserOnline(userID))
}

// TestBroker_MaxConnections 最大连接数限制测试
func TestBroker_MaxConnections(t *testing.T) {
	config := &sse.BrokerConfig{
		ClientBufferSize:  256,
		HeartbeatInterval: 30 * time.Second,
		ClientTimeout:     5 * time.Minute,
		EnableTopics:      true,
		MaxConnections:    2, // 限制最大2个连接
		HistorySize:       0,
	}

	broker := sse.NewBroker(config)
	ctx := context.Background()
	go broker.Start(ctx)
	defer broker.Stop()

	// 连接2个客户端（成功）
	client1 := sse.NewClient(broker, uuid.New(), uuid.Nil, nil)
	client2 := sse.NewClient(broker, uuid.New(), uuid.Nil, nil)

	broker.Register(client1)
	broker.Register(client2)
	time.Sleep(50 * time.Millisecond)

	stats := broker.GetStats()
	assert.Equal(t, 2, stats.TotalClients)

	// 尝试连接第3个客户端（应该失败）
	client3 := sse.NewClient(broker, uuid.New(), uuid.Nil, nil)
	broker.Register(client3)
	time.Sleep(50 * time.Millisecond)

	stats = broker.GetStats()
	assert.Equal(t, 2, stats.TotalClients) // 仍然是2个
}

// TestMessage_Format SSE 消息格式化测试
func TestMessage_Format(t *testing.T) {
	msg := sse.NewMessage(sse.EventMessage, "test data")
	msg.Retry = 5000

	formatted := msg.Format()

	assert.Contains(t, formatted, "id:")
	assert.Contains(t, formatted, "event: message")
	assert.Contains(t, formatted, "data: test data")
	assert.Contains(t, formatted, "\n\n") // SSE 格式以双换行结束
}

// BenchmarkBroker_SendToUser 性能基准测试
func BenchmarkBroker_SendToUser(b *testing.B) {
	broker := sse.NewBroker(sse.DefaultBrokerConfig())
	ctx := context.Background()
	go broker.Start(ctx)
	defer broker.Stop()

	userID := uuid.New()
	client := sse.NewClient(broker, userID, uuid.Nil, nil)
	broker.Register(client)
	time.Sleep(50 * time.Millisecond)

	// 启动接收 goroutine
	go func() {
		for range client.Send() {
			// 消费消息
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		broker.SendToUser(userID, "bench", "data")
	}
}

// BenchmarkBroker_Broadcast 广播性能测试
func BenchmarkBroker_Broadcast(b *testing.B) {
	broker := sse.NewBroker(sse.DefaultBrokerConfig())
	ctx := context.Background()
	go broker.Start(ctx)
	defer broker.Stop()

	// 创建100个客户端
	for i := 0; i < 100; i++ {
		client := sse.NewClient(broker, uuid.New(), uuid.Nil, nil)
		broker.Register(client)

		// 启动接收 goroutine
		go func(c *sse.Client) {
			for range c.Send() {
				// 消费消息
			}
		}(client)
	}

	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		broker.Broadcast("bench", "data")
	}
}
