package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Cluster Redis 集群（主从）
type Cluster struct {
	master redis.UniversalClient
	slaves []*SlaveNode

	config *Config
	logger *logger.Logger

	// 负载均衡器
	loadBalancer LoadBalancer

	mu sync.RWMutex
}

// SlaveNode 从库节点
type SlaveNode struct {
	client redis.UniversalClient
	config *NodeConfig

	// 健康状态
	healthy      bool
	lastCheck    time.Time
	failureCount int

	mu sync.RWMutex
}

// NewCluster 创建主从集群
func NewCluster(ctx context.Context, cfg *Config) (*Cluster, error) {
	cluster := &Cluster{
		config:       cfg,
		logger:       logger.GetLogger().With(zap.String("module", "cache.cluster")),
		loadBalancer: NewLoadBalancer(cfg.LoadBalancePolicy),
	}

	// 创建主库连接
	masterClient, err := createClient(ctx, cfg.MasterSlave.Master)
	if err != nil {
		return nil, fmt.Errorf("failed to create master client: %w", err)
	}
	cluster.master = masterClient

	cluster.logger.Info("Master Redis connected",
		zap.String("host", cfg.MasterSlave.Master.Host),
		zap.Int("port", cfg.MasterSlave.Master.Port),
	)

	// 创建从库连接
	cluster.slaves = make([]*SlaveNode, 0, len(cfg.MasterSlave.Slaves))
	for i, slaveCfg := range cfg.MasterSlave.Slaves {
		slaveNode := &SlaveNode{
			config:  &slaveCfg,
			healthy: true,
		}

		slaveClient, err := createClient(ctx, slaveCfg)
		if err != nil {
			cluster.logger.Warn("Failed to create slave client, skipping",
				zap.Int("index", i),
				zap.String("host", slaveCfg.Host),
				zap.Error(err),
			)
			continue
		}

		slaveNode.client = slaveClient
		cluster.slaves = append(cluster.slaves, slaveNode)

		cluster.logger.Info("Slave Redis connected",
			zap.Int("index", i),
			zap.String("host", slaveCfg.Host),
			zap.Int("port", slaveCfg.Port),
		)
	}

	if len(cluster.slaves) == 0 {
		cluster.logger.Warn("No slaves connected, all queries will use master")
	}

	// 启动健康检查
	if cfg.HealthCheck != nil && cfg.HealthCheck.Enable {
		go cluster.startHealthCheck(ctx)
	}

	return cluster, nil
}

// GetMaster 获取主库连接
func (c *Cluster) GetMaster() redis.UniversalClient {
	return c.master
}

// GetSlave 获取从库连接（负载均衡）
func (c *Cluster) GetSlave() (redis.UniversalClient, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	healthySlaves := c.getHealthySlaves()
	if len(healthySlaves) == 0 {
		return nil, fmt.Errorf("no healthy slaves available")
	}

	slave, err := c.loadBalancer.Select(healthySlaves)
	if err != nil {
		return nil, err
	}

	return slave.client, nil
}

// GetHealthySlaves 获取健康的从库列表
func (c *Cluster) GetHealthySlaves() []*SlaveNode {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.getHealthySlaves()
}

// getHealthySlaves 获取健康的从库列表（内部方法，不加锁）
func (c *Cluster) getHealthySlaves() []*SlaveNode {
	healthy := make([]*SlaveNode, 0, len(c.slaves))
	for _, slave := range c.slaves {
		slave.mu.RLock()
		if slave.healthy {
			healthy = append(healthy, slave)
		}
		slave.mu.RUnlock()
	}
	return healthy
}

// MarkSlaveUnhealthy 标记从库为不健康
func (c *Cluster) MarkSlaveUnhealthy(slave *SlaveNode) {
	slave.mu.Lock()
	defer slave.mu.Unlock()

	if slave.healthy {
		slave.healthy = false
		slave.failureCount++

		c.logger.Warn("Slave marked as unhealthy",
			zap.String("host", slave.config.Host),
			zap.Int("failure_count", slave.failureCount),
		)
	}
}

// MarkSlaveHealthy 标记从库为健康
func (c *Cluster) MarkSlaveHealthy(slave *SlaveNode) {
	slave.mu.Lock()
	defer slave.mu.Unlock()

	if !slave.healthy {
		slave.healthy = true
		slave.failureCount = 0

		c.logger.Info("Slave marked as healthy",
			zap.String("host", slave.config.Host),
		)
	}
}

// startHealthCheck 启动健康检查
func (c *Cluster) startHealthCheck(ctx context.Context) {
	interval := c.config.HealthCheck.Interval
	if interval == 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	c.logger.Info("Health check started", zap.Duration("interval", interval))

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Health check stopped")
			return
		case <-ticker.C:
			c.checkSlaveHealth(ctx)
		}
	}
}

// checkSlaveHealth 检查从库健康状态
func (c *Cluster) checkSlaveHealth(ctx context.Context) {
	timeout := c.config.HealthCheck.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for _, slave := range c.slaves {
		err := slave.client.Ping(checkCtx).Err()
		if err != nil {
			c.MarkSlaveUnhealthy(slave)
			c.logger.Error("Slave health check failed",
				zap.String("host", slave.config.Host),
				zap.Error(err),
			)
		} else {
			c.MarkSlaveHealthy(slave)
		}
	}
}

// Close 关闭所有连接
func (c *Cluster) Close() error {
	c.logger.Info("Closing Redis cluster")

	// 关闭主库
	if c.master != nil {
		if err := c.master.Close(); err != nil {
			c.logger.Error("Failed to close master", zap.Error(err))
		} else {
			c.logger.Info("Master closed")
		}
	}

	// 关闭从库
	for i, slave := range c.slaves {
		if slave.client != nil {
			if err := slave.client.Close(); err != nil {
				c.logger.Error("Failed to close slave",
					zap.Int("index", i),
					zap.String("host", slave.config.Host),
					zap.Error(err),
				)
			} else {
				c.logger.Info("Slave closed",
					zap.Int("index", i),
					zap.String("host", slave.config.Host),
				)
			}
		}
	}

	return nil
}

// createClient 创建 Redis 客户端
func createClient(ctx context.Context, cfg NodeConfig) (redis.UniversalClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,

		// 连接池配置（使用全局配置）
		PoolSize:     10,
		MinIdleConns: 2,

		// 超时配置
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Ping 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return client, nil
}
