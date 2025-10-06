package cache

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config Redis 配置（支持单机/主从/哨兵/集群）
type Config struct {
	// ============ 单机模式配置 ============
	// 当 MasterSlave/Sentinel/Cluster 都为 nil 时，使用单机模式
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"` // 数据库编号（0-15，集群模式无效）

	// ============ 主从模式配置（可选）============
	// 配置了 MasterSlave 则启用主从模式
	MasterSlave *MasterSlaveConfig `json:"master_slave,omitempty" yaml:"master_slave,omitempty"`

	// ============ 哨兵模式配置（可选）============
	// 配置了 Sentinel 则启用哨兵模式
	Sentinel *SentinelConfig `json:"sentinel,omitempty" yaml:"sentinel,omitempty"`

	// ============ 集群模式配置（可选）============
	// 配置了 Cluster 则启用集群模式
	Cluster *ClusterConfig `json:"cluster,omitempty" yaml:"cluster,omitempty"`

	// ============ 读写分离配置（主从/哨兵/集群）============
	ReadPolicy        ReadPolicy        `json:"read_policy,omitempty" yaml:"read_policy,omitempty"`
	LoadBalancePolicy LoadBalancePolicy `json:"load_balance_policy,omitempty" yaml:"load_balance_policy,omitempty"`

	// ============ 连接池配置 ============
	PoolSize     int           `json:"pool_size" yaml:"pool_size"`         // 连接池大小
	MinIdleConns int           `json:"min_idle_conns" yaml:"min_idle_conns"` // 最小空闲连接
	MaxConnAge   time.Duration `json:"max_conn_age" yaml:"max_conn_age"`   // 连接最大生命周期
	PoolTimeout  time.Duration `json:"pool_timeout" yaml:"pool_timeout"`   // 获取连接超时
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`   // 空闲连接超时

	// ============ 超时配置 ============
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`   // 拨号超时
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`   // 读超时
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"` // 写超时

	// ============ 健康检查（可选）============
	HealthCheck *HealthCheckConfig `json:"health_check,omitempty" yaml:"health_check,omitempty"`
}

// MasterSlaveConfig 主从配置
type MasterSlaveConfig struct {
	Master NodeConfig   `json:"master" yaml:"master"`
	Slaves []NodeConfig `json:"slaves" yaml:"slaves"`
}

// SentinelConfig 哨兵配置
type SentinelConfig struct {
	MasterName       string   `json:"master_name" yaml:"master_name"`           // 主节点名称
	SentinelAddrs    []string `json:"sentinel_addrs" yaml:"sentinel_addrs"`     // 哨兵地址列表
	Password         string   `json:"password" yaml:"password"`                 // Redis 密码
	DB               int      `json:"db" yaml:"db"`                             // 数据库编号
	SentinelPassword string   `json:"sentinel_password,omitempty" yaml:"sentinel_password,omitempty"` // 哨兵密码

	// 哨兵支持读写分离（读从库）
	SlaveOnly bool `json:"slave_only,omitempty" yaml:"slave_only,omitempty"` // 是否只读从库
}

// ClusterConfig 集群配置
type ClusterConfig struct {
	Addrs    []string `json:"addrs" yaml:"addrs"`       // 集群节点地址列表
	Password string   `json:"password" yaml:"password"` // 密码

	// 集群读写分离配置
	ReadOnly       bool `json:"read_only,omitempty" yaml:"read_only,omitempty"`             // 是否从从节点读
	RouteByLatency bool `json:"route_by_latency,omitempty" yaml:"route_by_latency,omitempty"` // 按延迟路由
	RouteRandomly  bool `json:"route_randomly,omitempty" yaml:"route_randomly,omitempty"`   // 随机路由

	// 集群配置
	MaxRedirects int `json:"max_redirects,omitempty" yaml:"max_redirects,omitempty"` // 最大重定向次数（默认 3）
}

// NodeConfig 节点配置
type NodeConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	DB       int    `json:"db,omitempty" yaml:"db,omitempty"`
	Weight   int    `json:"weight,omitempty" yaml:"weight,omitempty"` // 权重（用于加权负载均衡）
}

// ReadPolicy 读策略
type ReadPolicy string

const (
	ReadPolicyMaster      ReadPolicy = "master"       // 总是读主库
	ReadPolicySlave       ReadPolicy = "slave"        // 总是读从库
	ReadPolicyMasterFirst ReadPolicy = "master_first" // 优先主库，主库故障读从库
	ReadPolicySlaveFirst  ReadPolicy = "slave_first"  // 优先从库，从库故障读主库（推荐）
)

// LoadBalancePolicy 负载均衡策略
type LoadBalancePolicy string

const (
	LoadBalanceRandom     LoadBalancePolicy = "random"      // 随机（默认）
	LoadBalanceRoundRobin LoadBalancePolicy = "round_robin" // 轮询
	LoadBalanceLeastConn  LoadBalancePolicy = "least_conn"  // 最少连接
	LoadBalanceWeighted   LoadBalancePolicy = "weighted"    // 加权
)

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enable   bool          `json:"enable" yaml:"enable"`
	Interval time.Duration `json:"interval" yaml:"interval"`
	Timeout  time.Duration `json:"timeout" yaml:"timeout"`
}

// RedisMode Redis 部署模式
type RedisMode string

const (
	ModeStandalone  RedisMode = "standalone"   // 单机模式
	ModeMasterSlave RedisMode = "master_slave" // 主从模式
	ModeSentinel    RedisMode = "sentinel"     // 哨兵模式
	ModeCluster     RedisMode = "cluster"      // 集群模式
)

// DefaultConfig 返回默认单机配置
func DefaultConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,

		PoolSize:     10,
		MinIdleConns: 2,
		MaxConnAge:   0, // 0 表示不限制
		PoolTimeout:  4 * time.Second,
		IdleTimeout:  5 * time.Minute,

		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// DefaultMasterSlaveConfig 返回默认主从配置
func DefaultMasterSlaveConfig() *Config {
	cfg := DefaultConfig()

	cfg.MasterSlave = &MasterSlaveConfig{
		Master: NodeConfig{
			Host:     cfg.Host,
			Port:     cfg.Port,
			Password: cfg.Password,
			DB:       cfg.DB,
		},
	}

	cfg.ReadPolicy = ReadPolicySlaveFirst
	cfg.LoadBalancePolicy = LoadBalanceRandom

	return cfg
}

// DefaultSentinelConfig 返回默认哨兵配置
func DefaultSentinelConfig() *Config {
	cfg := DefaultConfig()

	cfg.Sentinel = &SentinelConfig{
		MasterName:    "mymaster",
		SentinelAddrs: []string{"localhost:26379"},
		Password:      "",
		DB:            0,
		SlaveOnly:     false,
	}

	cfg.ReadPolicy = ReadPolicySlaveFirst
	cfg.LoadBalancePolicy = LoadBalanceRandom

	return cfg
}

// DefaultClusterConfig 返回默认集群配置
func DefaultClusterConfig() *Config {
	cfg := DefaultConfig()

	cfg.Cluster = &ClusterConfig{
		Addrs:          []string{"localhost:7000", "localhost:7001", "localhost:7002"},
		Password:       "",
		MaxRedirects:   3,
		ReadOnly:       false,
		RouteByLatency: false,
		RouteRandomly:  true,
	}

	return cfg
}

// GetMode 获取部署模式
func (c *Config) GetMode() RedisMode {
	if c.Cluster != nil {
		return ModeCluster
	}
	if c.Sentinel != nil {
		return ModeSentinel
	}
	if c.MasterSlave != nil {
		return ModeMasterSlave
	}
	return ModeStandalone
}

// Validate 验证配置
func (c *Config) Validate() error {
	mode := c.GetMode()

	switch mode {
	case ModeStandalone:
		if c.Host == "" {
			return fmt.Errorf("host is required in standalone mode")
		}
		if c.Port <= 0 || c.Port > 65535 {
			return fmt.Errorf("invalid port: %d", c.Port)
		}

	case ModeMasterSlave:
		if c.MasterSlave == nil {
			return fmt.Errorf("master_slave config is required")
		}
		if err := c.MasterSlave.Master.Validate(); err != nil {
			return fmt.Errorf("master config: %w", err)
		}
		for i, slave := range c.MasterSlave.Slaves {
			if err := slave.Validate(); err != nil {
				return fmt.Errorf("slave[%d] config: %w", i, err)
			}
		}

	case ModeSentinel:
		if c.Sentinel == nil {
			return fmt.Errorf("sentinel config is required")
		}
		if c.Sentinel.MasterName == "" {
			return fmt.Errorf("sentinel master_name is required")
		}
		if len(c.Sentinel.SentinelAddrs) == 0 {
			return fmt.Errorf("sentinel_addrs is required")
		}

	case ModeCluster:
		if c.Cluster == nil {
			return fmt.Errorf("cluster config is required")
		}
		if len(c.Cluster.Addrs) == 0 {
			return fmt.Errorf("cluster addrs is required")
		}
	}

	// 通用验证
	if c.PoolSize <= 0 {
		return fmt.Errorf("pool_size must be positive")
	}
	if c.MinIdleConns < 0 {
		return fmt.Errorf("min_idle_conns must be non-negative")
	}

	return nil
}

// Validate 验证节点配置
func (n *NodeConfig) Validate() error {
	if n.Host == "" {
		return fmt.Errorf("host is required")
	}
	if n.Port <= 0 || n.Port > 65535 {
		return fmt.Errorf("invalid port: %d", n.Port)
	}
	if n.Weight < 0 {
		n.Weight = 1 // 默认权重
	}
	return nil
}

// Addr 返回节点地址
func (n *NodeConfig) Addr() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
}

// LoadFromFile 从 YAML 文件加载配置
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
