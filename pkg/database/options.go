package database

import "time"

// Option 配置选项函数
type Option func(*Config)

// ============ 单机模式选项 ============

// WithHost 设置主机地址
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

// WithDatabase 设置数据库名
func WithDatabase(database string) Option {
	return func(c *Config) {
		c.Database = database
	}
}

// WithUsername 设置用户名
func WithUsername(username string) Option {
	return func(c *Config) {
		c.Username = username
	}
}

// WithPassword 设置密码
func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

// WithSSLMode 设置 SSL 模式
func WithSSLMode(sslMode string) Option {
	return func(c *Config) {
		c.SSLMode = sslMode
	}
}

// WithMaxConns 设置最大连接数
func WithMaxConns(maxConns int32) Option {
	return func(c *Config) {
		c.MaxConns = maxConns
	}
}

// WithMinConns 设置最小连接数
func WithMinConns(minConns int32) Option {
	return func(c *Config) {
		c.MinConns = minConns
	}
}

// WithMaxConnLifetime 设置连接最大生命周期
func WithMaxConnLifetime(d time.Duration) Option {
	return func(c *Config) {
		c.MaxConnLifetime = d
	}
}

// WithMaxConnIdleTime 设置连接最大空闲时间
func WithMaxConnIdleTime(d time.Duration) Option {
	return func(c *Config) {
		c.MaxConnIdleTime = d
	}
}

// WithConnectTimeout 设置连接超时
func WithConnectTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.ConnectTimeout = d
	}
}

// WithDefaultQueryTimeout 设置默认查询超时
func WithDefaultQueryTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.DefaultQueryTimeout = d
	}
}

// WithLogLevel 设置日志级别
func WithLogLevel(level string) Option {
	return func(c *Config) {
		c.LogLevel = level
	}
}

// ============ 主从模式选项 ============

// WithMaster 设置主库配置
func WithMaster(master NodeConfig) Option {
	return func(c *Config) {
		c.Master = &master
	}
}

// WithSlaves 设置从库配置
func WithSlaves(slaves ...NodeConfig) Option {
	return func(c *Config) {
		c.Slaves = slaves
	}
}

// AddSlave 添加从库
func AddSlave(slave NodeConfig) Option {
	return func(c *Config) {
		c.Slaves = append(c.Slaves, slave)
	}
}

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

// WithFailover 设置故障转移配置
func WithFailover(failover FailoverConfig) Option {
	return func(c *Config) {
		c.Failover = &failover
	}
}

// WithHealthCheck 设置健康检查配置
func WithHealthCheck(healthCheck HealthCheckConfig) Option {
	return func(c *Config) {
		c.HealthCheck = &healthCheck
	}
}

// ============ 快捷选项 ============

// WithConfig 直接使用配置对象
func WithConfig(cfg *Config) Option {
	return func(c *Config) {
		*c = *cfg
	}
}

// WithProductionConfig 使用生产环境配置
func WithProductionConfig() Option {
	return func(c *Config) {
		*c = *ProductionConfig()
	}
}

// WithMasterSlaveMode 启用主从模式（快捷方式）
func WithMasterSlaveMode(masterHost string, slaveHosts ...string) Option {
	return func(c *Config) {
		// 主库配置
		c.Master = &NodeConfig{
			Host:                masterHost,
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
		}

		// 从库配置
		c.Slaves = make([]NodeConfig, 0, len(slaveHosts))
		for _, host := range slaveHosts {
			c.Slaves = append(c.Slaves, NodeConfig{
				Host:                host,
				Port:                c.Port,
				Database:            c.Database,
				Username:            c.Username,
				Password:            c.Password,
				SSLMode:             c.SSLMode,
				MaxConns:            c.MaxConns * 2, // 从库连接数通常更多
				MinConns:            c.MinConns * 2,
				MaxConnLifetime:     c.MaxConnLifetime,
				MaxConnIdleTime:     c.MaxConnIdleTime,
				HealthCheckPeriod:   c.HealthCheckPeriod,
				ConnectTimeout:      c.ConnectTimeout,
				DefaultQueryTimeout: c.DefaultQueryTimeout,
				LogLevel:            c.LogLevel,
				Weight:              1,
			})
		}

		// 默认读策略
		if c.ReadPolicy == "" {
			c.ReadPolicy = ReadPolicySlaveFirst
		}
		if c.LoadBalancePolicy == "" {
			c.LoadBalancePolicy = LoadBalanceRandom
		}
	}
}
