package database

import (
	"fmt"
	"math/rand"
	"sync/atomic"
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	// Select 选择一个从库节点
	Select(slaves []*SlaveNode) (*SlaveNode, error)
}

// RandomBalancer 随机负载均衡
type RandomBalancer struct{}

// NewRandomBalancer 创建随机负载均衡器
func NewRandomBalancer() *RandomBalancer {
	return &RandomBalancer{}
}

// Select 随机选择一个从库
func (r *RandomBalancer) Select(slaves []*SlaveNode) (*SlaveNode, error) {
	if len(slaves) == 0 {
		return nil, fmt.Errorf("no available slaves")
	}

	idx := rand.Intn(len(slaves))
	return slaves[idx], nil
}

// RoundRobinBalancer 轮询负载均衡
type RoundRobinBalancer struct {
	counter uint64
}

// NewRoundRobinBalancer 创建轮询负载均衡器
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{}
}

// Select 轮询选择一个从库
func (r *RoundRobinBalancer) Select(slaves []*SlaveNode) (*SlaveNode, error) {
	if len(slaves) == 0 {
		return nil, fmt.Errorf("no available slaves")
	}

	count := atomic.AddUint64(&r.counter, 1)
	idx := int(count-1) % len(slaves)
	return slaves[idx], nil
}

// LeastConnBalancer 最少连接负载均衡
type LeastConnBalancer struct{}

// NewLeastConnBalancer 创建最少连接负载均衡器
func NewLeastConnBalancer() *LeastConnBalancer {
	return &LeastConnBalancer{}
}

// Select 选择连接数最少的从库
func (l *LeastConnBalancer) Select(slaves []*SlaveNode) (*SlaveNode, error) {
	if len(slaves) == 0 {
		return nil, fmt.Errorf("no available slaves")
	}

	var selected *SlaveNode
	var minAcquired int32 = -1

	for _, slave := range slaves {
		stats := slave.pool.Stat()
		acquired := stats.AcquiredConns()

		if minAcquired == -1 || acquired < minAcquired {
			minAcquired = acquired
			selected = slave
		}
	}

	return selected, nil
}

// WeightedBalancer 加权负载均衡
type WeightedBalancer struct {
	counter uint64
}

// NewWeightedBalancer 创建加权负载均衡器
func NewWeightedBalancer() *WeightedBalancer {
	return &WeightedBalancer{}
}

// Select 根据权重选择从库
func (w *WeightedBalancer) Select(slaves []*SlaveNode) (*SlaveNode, error) {
	if len(slaves) == 0 {
		return nil, fmt.Errorf("no available slaves")
	}

	// 计算总权重
	totalWeight := 0
	for _, slave := range slaves {
		weight := slave.config.Weight
		if weight <= 0 {
			weight = 1
		}
		totalWeight += weight
	}

	if totalWeight == 0 {
		totalWeight = len(slaves) // 默认所有权重为 1
	}

	// 根据权重选择
	count := atomic.AddUint64(&w.counter, 1)
	offset := int(count) % totalWeight

	current := 0
	for _, slave := range slaves {
		weight := slave.config.Weight
		if weight <= 0 {
			weight = 1
		}

		current += weight
		if offset < current {
			return slave, nil
		}
	}

	// 回退到第一个
	return slaves[0], nil
}

// NewLoadBalancer 根据策略创建负载均衡器
func NewLoadBalancer(policy LoadBalancePolicy) LoadBalancer {
	switch policy {
	case LoadBalanceRoundRobin:
		return NewRoundRobinBalancer()
	case LoadBalanceLeastConn:
		return NewLeastConnBalancer()
	case LoadBalanceWeighted:
		return NewWeightedBalancer()
	default:
		return NewRandomBalancer()
	}
}
