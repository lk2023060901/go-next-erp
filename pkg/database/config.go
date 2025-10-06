package database

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 数据库配置（支持单机/主从模式）
type Config struct {
	// ============ 单机模式配置 ============
	// 当 Master 为 nil 时，使用单机模式
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Database string `json:"database" yaml:"database"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	SSLMode  string `json:"ssl_mode" yaml:"ssl_mode"` // disable/require/verify-ca/verify-full

	// 连接池配置
	MaxConns          int32         `json:"max_conns" yaml:"max_conns"`                     // 最大连接数
	MinConns          int32         `json:"min_conns" yaml:"min_conns"`                     // 最小连接数
	MaxConnLifetime   time.Duration `json:"max_conn_lifetime" yaml:"max_conn_lifetime"`     // 连接最大生命周期
	MaxConnIdleTime   time.Duration `json:"max_conn_idle_time" yaml:"max_conn_idle_time"`   // 连接最大空闲时间
	HealthCheckPeriod time.Duration `json:"health_check_period" yaml:"health_check_period"` // 健康检查周期
	ConnectTimeout    time.Duration `json:"connect_timeout" yaml:"connect_timeout"`         // 连接超时

	// 查询配置
	DefaultQueryTimeout time.Duration `json:"default_query_timeout" yaml:"default_query_timeout"` // 默认查询超时

	// 日志配置
	LogLevel string `json:"log_level" yaml:"log_level"` // trace/debug/info/warn/error/none

	// ============ 主从模式配置（可选）============
	// 配置 Master 则启用主从模式
	Master *NodeConfig  `json:"master,omitempty" yaml:"master,omitempty"`
	Slaves []NodeConfig `json:"slaves,omitempty" yaml:"slaves,omitempty"`

	// 读写分离策略（仅主从模式）
	ReadPolicy        ReadPolicy        `json:"read_policy,omitempty" yaml:"read_policy,omitempty"`
	LoadBalancePolicy LoadBalancePolicy `json:"load_balance_policy,omitempty" yaml:"load_balance_policy,omitempty"`

	// 故障转移配置（可选）
	Failover *FailoverConfig `json:"failover,omitempty" yaml:"failover,omitempty"`

	// 健康检查配置（可选）
	HealthCheck *HealthCheckConfig `json:"health_check,omitempty" yaml:"health_check,omitempty"`
}

// NodeConfig 节点配置
type NodeConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Database string `json:"database" yaml:"database"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	SSLMode  string `json:"ssl_mode" yaml:"ssl_mode"`

	// 连接池配置
	MaxConns          int32         `json:"max_conns" yaml:"max_conns"`
	MinConns          int32         `json:"min_conns" yaml:"min_conns"`
	MaxConnLifetime   time.Duration `json:"max_conn_lifetime" yaml:"max_conn_lifetime"`
	MaxConnIdleTime   time.Duration `json:"max_conn_idle_time" yaml:"max_conn_idle_time"`
	HealthCheckPeriod time.Duration `json:"health_check_period" yaml:"health_check_period"`
	ConnectTimeout    time.Duration `json:"connect_timeout" yaml:"connect_timeout"`

	// 查询配置
	DefaultQueryTimeout time.Duration `json:"default_query_timeout" yaml:"default_query_timeout"`

	// 日志配置
	LogLevel string `json:"log_level" yaml:"log_level"`

	// 节点权重（用于加权负载均衡，默认 1）
	Weight int `json:"weight,omitempty" yaml:"weight,omitempty"`
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

// FailoverConfig 故障转移配置
type FailoverConfig struct {
	Enable              bool          `json:"enable" yaml:"enable"`
	MaxRetries          int           `json:"max_retries" yaml:"max_retries"`                   // 最大重试次数
	RetryInterval       time.Duration `json:"retry_interval" yaml:"retry_interval"`             // 重试间隔
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"` // 健康检查间隔
	FailureThreshold    int           `json:"failure_threshold" yaml:"failure_threshold"`       // 故障阈值
	RecoveryThreshold   int           `json:"recovery_threshold" yaml:"recovery_threshold"`     // 恢复阈值
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enable   bool          `json:"enable" yaml:"enable"`
	Interval time.Duration `json:"interval" yaml:"interval"`
	Timeout  time.Duration `json:"timeout" yaml:"timeout"`
}

// DefaultConfig 返回默认单机配置
func DefaultConfig() *Config {
	return &Config{
		Host:                "localhost",
		Port:                5432,
		Database:            "postgres",
		Username:            "postgres",
		Password:            "",
		SSLMode:             "disable",
		MaxConns:            25,
		MinConns:            5,
		MaxConnLifetime:     1 * time.Hour,
		MaxConnIdleTime:     30 * time.Minute,
		HealthCheckPeriod:   1 * time.Minute,
		ConnectTimeout:      10 * time.Second,
		DefaultQueryTimeout: 30 * time.Second,
		LogLevel:            "info",
	}
}

// ProductionConfig 返回生产环境推荐配置（单机）
func ProductionConfig() *Config {
	return &Config{
		Host:                "localhost",
		Port:                5432,
		Database:            "myapp",
		Username:            "postgres",
		Password:            "",
		SSLMode:             "require",
		MaxConns:            50,
		MinConns:            10,
		MaxConnLifetime:     1 * time.Hour,
		MaxConnIdleTime:     30 * time.Minute,
		HealthCheckPeriod:   1 * time.Minute,
		ConnectTimeout:      10 * time.Second,
		DefaultQueryTimeout: 30 * time.Second,
		LogLevel:            "info",
	}
}

// DefaultMasterSlaveConfig 返回默认主从配置
func DefaultMasterSlaveConfig() *Config {
	cfg := DefaultConfig()

	// 主库配置
	cfg.Master = &NodeConfig{
		Host:                cfg.Host,
		Port:                cfg.Port,
		Database:            cfg.Database,
		Username:            cfg.Username,
		Password:            cfg.Password,
		SSLMode:             cfg.SSLMode,
		MaxConns:            50,
		MinConns:            10,
		MaxConnLifetime:     cfg.MaxConnLifetime,
		MaxConnIdleTime:     cfg.MaxConnIdleTime,
		HealthCheckPeriod:   cfg.HealthCheckPeriod,
		ConnectTimeout:      cfg.ConnectTimeout,
		DefaultQueryTimeout: cfg.DefaultQueryTimeout,
		LogLevel:            cfg.LogLevel,
	}

	// 读策略
	cfg.ReadPolicy = ReadPolicySlaveFirst
	cfg.LoadBalancePolicy = LoadBalanceRandom

	// 故障转移
	cfg.Failover = &FailoverConfig{
		Enable:              true,
		MaxRetries:          3,
		RetryInterval:       1 * time.Second,
		HealthCheckInterval: 10 * time.Second,
		FailureThreshold:    3,
		RecoveryThreshold:   2,
	}

	// 健康检查
	cfg.HealthCheck = &HealthCheckConfig{
		Enable:   true,
		Interval: 30 * time.Second,
		Timeout:  5 * time.Second,
	}

	return cfg
}

// IsMasterSlaveMode 判断是否为主从模式
func (c *Config) IsMasterSlaveMode() bool {
	return c.Master != nil
}

// ToNodeConfig 将单机配置转换为 NodeConfig
func (c *Config) ToNodeConfig() *NodeConfig {
	return &NodeConfig{
		Host:                c.Host,
		Port:                c.Port,
		Database:            c.Database,
		Username:            c.Username,
		Password:            c.Password,
		SSLMode:             c.SSLMode,
		MaxConns:            c.MaxConns,
		MinConns:            c.MinConns,
		MaxConnLifetime:     c.MaxConnLifetime,
		MaxConnIdleTime:     c.MaxConnIdleTime,
		HealthCheckPeriod:   c.HealthCheckPeriod,
		ConnectTimeout:      c.ConnectTimeout,
		DefaultQueryTimeout: c.DefaultQueryTimeout,
		LogLevel:            c.LogLevel,
		Weight:              1,
	}
}

// DSN 返回 PostgreSQL 连接字符串
func (n *NodeConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		n.Host, n.Port, n.Database, n.Username, n.Password, n.SSLMode,
	)
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.IsMasterSlaveMode() {
		// 主从模式验证
		if c.Master == nil {
			return fmt.Errorf("master config is required in master-slave mode")
		}
		if err := c.Master.Validate(); err != nil {
			return fmt.Errorf("master config: %w", err)
		}

		// 验证从库配置
		for i, slave := range c.Slaves {
			if err := slave.Validate(); err != nil {
				return fmt.Errorf("slave[%d] config: %w", i, err)
			}
		}

		// 验证读策略
		validReadPolicies := map[ReadPolicy]bool{
			ReadPolicyMaster:      true,
			ReadPolicySlave:       true,
			ReadPolicyMasterFirst: true,
			ReadPolicySlaveFirst:  true,
		}
		if c.ReadPolicy != "" && !validReadPolicies[c.ReadPolicy] {
			return fmt.Errorf("invalid read_policy: %s", c.ReadPolicy)
		}

		// 如果没有从库但读策略要求从库
		if len(c.Slaves) == 0 && c.ReadPolicy == ReadPolicySlave {
			return fmt.Errorf("read_policy is 'slave' but no slaves configured")
		}
	} else {
		// 单机模式验证
		if c.Host == "" {
			return fmt.Errorf("host is required")
		}
		if c.Port <= 0 || c.Port > 65535 {
			return fmt.Errorf("invalid port: %d", c.Port)
		}
		if c.Database == "" {
			return fmt.Errorf("database is required")
		}
		if c.Username == "" {
			return fmt.Errorf("username is required")
		}
	}

	// 通用验证
	if c.MaxConns <= 0 && !c.IsMasterSlaveMode() {
		return fmt.Errorf("max_conns must be positive")
	}
	if c.MinConns < 0 && !c.IsMasterSlaveMode() {
		return fmt.Errorf("min_conns must be non-negative")
	}
	if c.MinConns > c.MaxConns && !c.IsMasterSlaveMode() {
		return fmt.Errorf("min_conns (%d) cannot exceed max_conns (%d)", c.MinConns, c.MaxConns)
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
	if n.Database == "" {
		return fmt.Errorf("database is required")
	}
	if n.Username == "" {
		return fmt.Errorf("username is required")
	}
	if n.MaxConns <= 0 {
		return fmt.Errorf("max_conns must be positive")
	}
	if n.MinConns < 0 {
		return fmt.Errorf("min_conns must be non-negative")
	}
	if n.MinConns > n.MaxConns {
		return fmt.Errorf("min_conns (%d) cannot exceed max_conns (%d)", n.MinConns, n.MaxConns)
	}
	if n.Weight < 0 {
		n.Weight = 1 // 默认权重
	}

	return nil
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
