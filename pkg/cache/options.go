package cache

import (
	"time"
)

// Option Redis 配置选项
type Option func(*Config)

// ============ 单机模式 ============

// WithHost 设置主机
func WithHost(host string) Option {
	return func(c *Config) {
		c.Host = host
	}
}

// WithPort 设置端口
func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithPassword 设置密码
func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

// WithDB 设置数据库编号（0-15，集群模式无效）
func WithDB(db int) Option {
	return func(c *Config) {
		c.DB = db
	}
}

// ============ 主从模式 ============

// WithMasterSlave 启用主从模式
func WithMasterSlave(master NodeConfig, slaves []NodeConfig) Option {
	return func(c *Config) {
		c.MasterSlave = &MasterSlaveConfig{
			Master: master,
			Slaves: slaves,
		}
	}
}

// ============ 哨兵模式 ============

// WithSentinel 启用哨兵模式
func WithSentinel(masterName string, sentinelAddrs []string, password string) Option {
	return func(c *Config) {
		c.Sentinel = &SentinelConfig{
			MasterName:    masterName,
			SentinelAddrs: sentinelAddrs,
			Password:      password,
		}
	}
}

// WithSentinelPassword 设置哨兵密码
func WithSentinelPassword(password string) Option {
	return func(c *Config) {
		if c.Sentinel != nil {
			c.Sentinel.SentinelPassword = password
		}
	}
}

// WithSlaveOnly 设置只读从库（仅哨兵模式）
func WithSlaveOnly(slaveOnly bool) Option {
	return func(c *Config) {
		if c.Sentinel != nil {
			c.Sentinel.SlaveOnly = slaveOnly
		}
	}
}

// ============ 集群模式 ============

// WithCluster 启用集群模式
func WithCluster(addrs []string, password string) Option {
	return func(c *Config) {
		c.Cluster = &ClusterConfig{
			Addrs:        addrs,
			Password:     password,
			MaxRedirects: 3,
		}
	}
}

// WithClusterReadOnly 设置从从节点读（集群模式）
func WithClusterReadOnly(readOnly bool) Option {
	return func(c *Config) {
		if c.Cluster != nil {
			c.Cluster.ReadOnly = readOnly
		}
	}
}

// WithClusterRouteByLatency 启用按延迟路由（集群模式）
func WithClusterRouteByLatency(enable bool) Option {
	return func(c *Config) {
		if c.Cluster != nil {
			c.Cluster.RouteByLatency = enable
		}
	}
}

// WithClusterRouteRandomly 启用随机路由（集群模式）
func WithClusterRouteRandomly(enable bool) Option {
	return func(c *Config) {
		if c.Cluster != nil {
			c.Cluster.RouteRandomly = enable
		}
	}
}

// WithMaxRedirects 设置最大重定向次数（集群模式）
func WithMaxRedirects(max int) Option {
	return func(c *Config) {
		if c.Cluster != nil {
			c.Cluster.MaxRedirects = max
		}
	}
}

// ============ 读写分离配置 ============

// WithReadPolicy 设置读策略
func WithReadPolicy(policy ReadPolicy) Option {
	return func(c *Config) {
		c.ReadPolicy = policy
	}
}

// WithLoadBalancePolicy 设置负载均衡策略
func WithLoadBalancePolicy(policy LoadBalancePolicy) Option {
	return func(c *Config) {
		c.LoadBalancePolicy = policy
	}
}

// ============ 连接池配置 ============

// WithPoolSize 设置连接池大小
func WithPoolSize(size int) Option {
	return func(c *Config) {
		c.PoolSize = size
	}
}

// WithMinIdleConns 设置最小空闲连接
func WithMinIdleConns(min int) Option {
	return func(c *Config) {
		c.MinIdleConns = min
	}
}

// WithMaxConnAge 设置连接最大生命周期
func WithMaxConnAge(age time.Duration) Option {
	return func(c *Config) {
		c.MaxConnAge = age
	}
}

// WithPoolTimeout 设置获取连接超时
func WithPoolTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.PoolTimeout = timeout
	}
}

// WithIdleTimeout 设置空闲连接超时
func WithIdleTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.IdleTimeout = timeout
	}
}

// ============ 超时配置 ============

// WithDialTimeout 设置拨号超时
func WithDialTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.DialTimeout = timeout
	}
}

// WithReadTimeout 设置读超时
func WithReadTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = timeout
	}
}

// WithWriteTimeout 设置写超时
func WithWriteTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.WriteTimeout = timeout
	}
}

// ============ 健康检查 ============

// WithHealthCheck 启用健康检查
func WithHealthCheck(enable bool, interval, timeout time.Duration) Option {
	return func(c *Config) {
		c.HealthCheck = &HealthCheckConfig{
			Enable:   enable,
			Interval: interval,
			Timeout:  timeout,
		}
	}
}

// ============ 快捷配置 ============

// WithStandaloneConfig 快捷配置单机模式
func WithStandaloneConfig(host string, port int, password string, db int) Option {
	return func(c *Config) {
		c.Host = host
		c.Port = port
		c.Password = password
		c.DB = db
		c.MasterSlave = nil
		c.Sentinel = nil
		c.Cluster = nil
	}
}

// WithMasterSlaveConfig 快捷配置主从模式
func WithMasterSlaveConfig(masterHost string, masterPort int, slaveHosts []string, password string) Option {
	return func(c *Config) {
		// 设置主库
		c.MasterSlave = &MasterSlaveConfig{
			Master: NodeConfig{
				Host:     masterHost,
				Port:     masterPort,
				Password: password,
			},
		}

		// 设置从库
		slaves := make([]NodeConfig, 0, len(slaveHosts))
		for _, addr := range slaveHosts {
			host := addr
			port := 6379 // 默认端口
			slaves = append(slaves, NodeConfig{
				Host:     host,
				Port:     port,
				Password: password,
			})
		}
		c.MasterSlave.Slaves = slaves

		// 默认读写分离策略
		c.ReadPolicy = ReadPolicySlaveFirst
		c.LoadBalancePolicy = LoadBalanceRandom
	}
}

// WithSentinelConfig 快捷配置哨兵模式
func WithSentinelConfig(masterName string, sentinelAddrs []string, password string) Option {
	return func(c *Config) {
		c.Sentinel = &SentinelConfig{
			MasterName:    masterName,
			SentinelAddrs: sentinelAddrs,
			Password:      password,
			DB:            c.DB,
		}

		// 默认读写分离策略
		c.ReadPolicy = ReadPolicySlaveFirst
		c.LoadBalancePolicy = LoadBalanceRandom
	}
}

// WithClusterConfig 快捷配置集群模式
func WithClusterConfig(addrs []string, password string, readOnly bool) Option {
	return func(c *Config) {
		c.Cluster = &ClusterConfig{
			Addrs:          addrs,
			Password:       password,
			MaxRedirects:   3,
			ReadOnly:       readOnly,
			RouteRandomly:  true,
			RouteByLatency: false,
		}
	}
}
