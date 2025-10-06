package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"go.uber.org/zap"
)

// Cluster 数据库集群（主从）
type Cluster struct {
	master *pgxpool.Pool
	slaves []*SlaveNode

	config *Config
	logger *logger.Logger

	// 负载均衡器
	loadBalancer LoadBalancer

	mu sync.RWMutex
}

// SlaveNode 从库节点
type SlaveNode struct {
	pool   *pgxpool.Pool
	config *NodeConfig

	// 健康状态
	healthy      bool
	lastCheck    time.Time
	failureCount int

	mu sync.RWMutex
}

// NewCluster 创建数据库集群
func NewCluster(ctx context.Context, cfg *Config) (*Cluster, error) {
	cluster := &Cluster{
		config:       cfg,
		logger:       logger.GetLogger().With(zap.String("module", "database.cluster")),
		loadBalancer: NewLoadBalancer(cfg.LoadBalancePolicy),
	}

	// 创建主库连接池
	masterPool, err := createPool(ctx, cfg.Master)
	if err != nil {
		return nil, fmt.Errorf("failed to create master pool: %w", err)
	}
	cluster.master = masterPool

	cluster.logger.Info("Master database connected",
		zap.String("host", cfg.Master.Host),
		zap.Int("port", cfg.Master.Port),
	)

	// 创建从库连接池
	cluster.slaves = make([]*SlaveNode, 0, len(cfg.Slaves))
	for i, slaveCfg := range cfg.Slaves {
		slaveNode := &SlaveNode{
			config:  &slaveCfg,
			healthy: true,
		}

		slavePool, err := createPool(ctx, &slaveCfg)
		if err != nil {
			cluster.logger.Warn("Failed to create slave pool, skipping",
				zap.Int("index", i),
				zap.String("host", slaveCfg.Host),
				zap.Error(err),
			)
			continue
		}

		slaveNode.pool = slavePool
		cluster.slaves = append(cluster.slaves, slaveNode)

		cluster.logger.Info("Slave database connected",
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

// GetMaster 获取主库连接池
func (c *Cluster) GetMaster() *pgxpool.Pool {
	return c.master
}

// GetSlave 获取从库连接池（负载均衡）
func (c *Cluster) GetSlave() (*pgxpool.Pool, error) {
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

	return slave.pool, nil
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
		err := slave.pool.Ping(checkCtx)
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

// Close 关闭所有连接池
func (c *Cluster) Close() {
	c.logger.Info("Closing database cluster")

	// 关闭主库
	if c.master != nil {
		c.master.Close()
		c.logger.Info("Master pool closed")
	}

	// 关闭从库
	for i, slave := range c.slaves {
		if slave.pool != nil {
			slave.pool.Close()
			c.logger.Info("Slave pool closed",
				zap.Int("index", i),
				zap.String("host", slave.config.Host),
			)
		}
	}
}

// createPool 创建连接池
func createPool(ctx context.Context, cfg *NodeConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// 连接池配置
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	// 创建连接池
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Ping 测试连接
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return pool, nil
}
