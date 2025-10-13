package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "postgres", cfg.Database)
	assert.Equal(t, "postgres", cfg.Username)
	assert.Equal(t, "disable", cfg.SSLMode)
	assert.Equal(t, int32(25), cfg.MaxConns)
	assert.Equal(t, int32(5), cfg.MinConns)
	assert.Equal(t, 1*time.Hour, cfg.MaxConnLifetime)
	assert.Equal(t, 30*time.Minute, cfg.MaxConnIdleTime)
}

func TestProductionConfig(t *testing.T) {
	cfg := ProductionConfig()

	assert.Equal(t, int32(50), cfg.MaxConns)
	assert.Equal(t, int32(10), cfg.MinConns)
	assert.Equal(t, 1*time.Hour, cfg.MaxConnLifetime)
	assert.Equal(t, 30*time.Minute, cfg.MaxConnIdleTime)
	assert.Equal(t, 10*time.Second, cfg.ConnectTimeout)
	assert.Equal(t, 30*time.Second, cfg.DefaultQueryTimeout)
	assert.Equal(t, "require", cfg.SSLMode)
}

func TestConfigValidation(t *testing.T) {
	t.Run("valid single node config", func(t *testing.T) {
		cfg := DefaultConfig()
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing host", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Host = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "host")
	})

	t.Run("invalid port", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Port = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "port")
	})

	t.Run("invalid port range", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Port = 70000
		err := cfg.Validate()
		assert.Error(t, err)
	})

	t.Run("missing database name", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Database = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database")
	})

	t.Run("missing username", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Username = ""
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "username")
	})

	t.Run("invalid max connections", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.MaxConns = 0
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max_conns")
	})

	t.Run("invalid min connections", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.MinConns = -1
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "min_conns")
	})

	t.Run("min_conns greater than max_conns", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.MinConns = 50
		cfg.MaxConns = 10
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot exceed")
	})
}

func TestConfigMethods(t *testing.T) {
	t.Run("IsMasterSlaveMode - false for single node", func(t *testing.T) {
		cfg := DefaultConfig()
		assert.False(t, cfg.IsMasterSlaveMode())
	})

	t.Run("IsMasterSlaveMode - true with master and slaves", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Master = &NodeConfig{Host: "master", Port: 5432}
		cfg.Slaves = []NodeConfig{
			{Host: "slave1", Port: 5432},
		}
		assert.True(t, cfg.IsMasterSlaveMode())
	})

	t.Run("ToNodeConfig", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Host = "testhost"
		cfg.Port = 5433
		cfg.Database = "testdb"

		nodeConfig := cfg.ToNodeConfig()
		assert.Equal(t, "testhost", nodeConfig.Host)
		assert.Equal(t, 5433, nodeConfig.Port)
		assert.Equal(t, "testdb", nodeConfig.Database)
	})
}

func TestOptions(t *testing.T) {
	t.Run("WithHost", func(t *testing.T) {
		cfg := DefaultConfig()
		WithHost("newhost")(cfg)
		assert.Equal(t, "newhost", cfg.Host)
	})

	t.Run("WithPort", func(t *testing.T) {
		cfg := DefaultConfig()
		WithPort(5433)(cfg)
		assert.Equal(t, 5433, cfg.Port)
	})

	t.Run("WithDatabase", func(t *testing.T) {
		cfg := DefaultConfig()
		WithDatabase("mydb")(cfg)
		assert.Equal(t, "mydb", cfg.Database)
	})

	t.Run("WithUsername", func(t *testing.T) {
		cfg := DefaultConfig()
		WithUsername("admin")(cfg)
		assert.Equal(t, "admin", cfg.Username)
	})

	t.Run("WithPassword", func(t *testing.T) {
		cfg := DefaultConfig()
		WithPassword("secret")(cfg)
		assert.Equal(t, "secret", cfg.Password)
	})

	t.Run("WithMaxConns", func(t *testing.T) {
		cfg := DefaultConfig()
		WithMaxConns(50)(cfg)
		assert.Equal(t, int32(50), cfg.MaxConns)
	})

	t.Run("WithMinConns", func(t *testing.T) {
		cfg := DefaultConfig()
		WithMinConns(10)(cfg)
		assert.Equal(t, int32(10), cfg.MinConns)
	})

	t.Run("WithConnectTimeout", func(t *testing.T) {
		cfg := DefaultConfig()
		WithConnectTimeout(15 * time.Second)(cfg)
		assert.Equal(t, 15*time.Second, cfg.ConnectTimeout)
	})

	t.Run("multiple options", func(t *testing.T) {
		cfg := DefaultConfig()
		opts := []Option{
			WithHost("prodhost"),
			WithPort(5433),
			WithDatabase("proddb"),
			WithMaxConns(100),
		}

		for _, opt := range opts {
			opt(cfg)
		}

		assert.Equal(t, "prodhost", cfg.Host)
		assert.Equal(t, 5433, cfg.Port)
		assert.Equal(t, "proddb", cfg.Database)
		assert.Equal(t, int32(100), cfg.MaxConns)
	})
}

func TestWithMasterSlaveMode(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Database = "testdb"
	cfg.Username = "admin"
	cfg.Password = "secret"

	WithMasterSlaveMode("master.db", "slave1.db", "slave2.db")(cfg)

	// 验证主库配置
	require.NotNil(t, cfg.Master)
	assert.Equal(t, "master.db", cfg.Master.Host)
	assert.Equal(t, "testdb", cfg.Master.Database)
	assert.Equal(t, "admin", cfg.Master.Username)

	// 验证从库配置
	require.Len(t, cfg.Slaves, 2)
	assert.Equal(t, "slave1.db", cfg.Slaves[0].Host)
	assert.Equal(t, "slave2.db", cfg.Slaves[1].Host)

	// 验证默认策略
	assert.Equal(t, ReadPolicySlaveFirst, cfg.ReadPolicy)
	assert.Equal(t, LoadBalanceRandom, cfg.LoadBalancePolicy)

	// 验证是主从模式
	assert.True(t, cfg.IsMasterSlaveMode())
}

func TestNodeConfigValidation(t *testing.T) {
	t.Run("valid node config", func(t *testing.T) {
		node := NodeConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			Username: "admin",
			MaxConns: 10,
			MinConns: 2,
		}
		err := node.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing host", func(t *testing.T) {
		node := NodeConfig{
			Port:     5432,
			Database: "testdb",
			Username: "admin",
		}
		err := node.Validate()
		assert.Error(t, err)
	})

	t.Run("negative weight gets corrected", func(t *testing.T) {
		node := NodeConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			Username: "admin",
			MaxConns: 10,
			MinConns: 2,
			Weight:   -1,
		}
		err := node.Validate()
		assert.NoError(t, err)
		assert.Equal(t, 1, node.Weight) // Auto-corrected to 1
	})

	t.Run("min > max connections", func(t *testing.T) {
		node := NodeConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			Username: "admin",
			MaxConns: 5,
			MinConns: 10,
		}
		err := node.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot exceed")
	})
}
