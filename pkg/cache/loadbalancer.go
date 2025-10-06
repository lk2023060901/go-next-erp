package cache

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	Select(nodes []*SlaveNode) (*SlaveNode, error)
}

// RandomLoadBalancer 随机负载均衡
type RandomLoadBalancer struct {
	rand *rand.Rand
	mu   sync.Mutex
}

// NewRandomLoadBalancer 创建随机负载均衡器
func NewRandomLoadBalancer() *RandomLoadBalancer {
	return &RandomLoadBalancer{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 随机选择节点
func (lb *RandomLoadBalancer) Select(nodes []*SlaveNode) (*SlaveNode, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	lb.mu.Lock()
	idx := lb.rand.Intn(len(nodes))
	lb.mu.Unlock()

	return nodes[idx], nil
}

// RoundRobinLoadBalancer 轮询负载均衡
type RoundRobinLoadBalancer struct {
	counter uint64
}

// NewRoundRobinLoadBalancer 创建轮询负载均衡器
func NewRoundRobinLoadBalancer() *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{}
}

// Select 轮询选择节点
func (lb *RoundRobinLoadBalancer) Select(nodes []*SlaveNode) (*SlaveNode, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	idx := atomic.AddUint64(&lb.counter, 1) % uint64(len(nodes))
	return nodes[idx], nil
}

// LeastConnLoadBalancer 最少连接负载均衡
type LeastConnLoadBalancer struct{}

// NewLeastConnLoadBalancer 创建最少连接负载均衡器
func NewLeastConnLoadBalancer() *LeastConnLoadBalancer {
	return &LeastConnLoadBalancer{}
}

// Select 选择连接数最少的节点
func (lb *LeastConnLoadBalancer) Select(nodes []*SlaveNode) (*SlaveNode, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	// 使用 Redis 的 PoolStats 获取连接数
	minConns := int32(^uint32(0) >> 1) // MaxInt32
	var selected *SlaveNode

	for _, node := range nodes {
		if client, ok := node.client.(*redis.Client); ok {
			stats := client.PoolStats()
			conns := int32(stats.TotalConns)

			if conns < minConns {
				minConns = conns
				selected = node
			}
		}
	}

	if selected == nil {
		return nodes[0], nil
	}

	return selected, nil
}

// WeightedLoadBalancer 加权负载均衡
type WeightedLoadBalancer struct {
	rand *rand.Rand
	mu   sync.Mutex
}

// NewWeightedLoadBalancer 创建加权负载均衡器
func NewWeightedLoadBalancer() *WeightedLoadBalancer {
	return &WeightedLoadBalancer{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Select 加权随机选择节点
func (lb *WeightedLoadBalancer) Select(nodes []*SlaveNode) (*SlaveNode, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	// 计算总权重
	totalWeight := 0
	for _, node := range nodes {
		if node.config.Weight <= 0 {
			node.config.Weight = 1
		}
		totalWeight += node.config.Weight
	}

	// 加权随机选择
	lb.mu.Lock()
	r := lb.rand.Intn(totalWeight)
	lb.mu.Unlock()

	for _, node := range nodes {
		r -= node.config.Weight
		if r < 0 {
			return node, nil
		}
	}

	// 理论上不会到这里
	return nodes[0], nil
}

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(policy LoadBalancePolicy) LoadBalancer {
	switch policy {
	case LoadBalanceRoundRobin:
		return NewRoundRobinLoadBalancer()
	case LoadBalanceLeastConn:
		return NewLeastConnLoadBalancer()
	case LoadBalanceWeighted:
		return NewWeightedLoadBalancer()
	default:
		return NewRandomLoadBalancer()
	}
}
