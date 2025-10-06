package cache

import (
	"context"
	"testing"
	"time"
)

// TestStandaloneMode 测试单机模式
func TestStandaloneMode(t *testing.T) {
	ctx := context.Background()

	// 创建单机模式 Redis
	r, err := New(ctx,
		WithHost("localhost"),
		WithPort(6379),
		WithDB(0),
		WithPoolSize(10),
	)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer r.Close()

	// 验证模式
	if r.GetMode() != ModeStandalone {
		t.Errorf("Expected standalone mode, got %v", r.GetMode())
	}

	// 基本操作
	t.Run("Set and Get", func(t *testing.T) {
		key := "test:standalone:key"
		value := "test_value"

		// Set
		if err := r.Set(ctx, key, value, time.Minute).Err(); err != nil {
			t.Errorf("Set() error = %v", err)
		}

		// Get
		result, err := r.Get(ctx, key).Result()
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}
		if result != value {
			t.Errorf("Get() = %v, want %v", result, value)
		}

		// Del
		if err := r.Del(ctx, key).Err(); err != nil {
			t.Errorf("Del() error = %v", err)
		}
	})

	// Hash 操作
	t.Run("Hash Operations", func(t *testing.T) {
		key := "test:standalone:hash"

		// HSet
		if err := r.HSet(ctx, key, "field1", "value1", "field2", "value2").Err(); err != nil {
			t.Errorf("HSet() error = %v", err)
		}

		// HGet
		result, err := r.HGet(ctx, key, "field1").Result()
		if err != nil {
			t.Errorf("HGet() error = %v", err)
		}
		if result != "value1" {
			t.Errorf("HGet() = %v, want value1", result)
		}

		// HGetAll
		all, err := r.HGetAll(ctx, key).Result()
		if err != nil {
			t.Errorf("HGetAll() error = %v", err)
		}
		if len(all) != 2 {
			t.Errorf("HGetAll() returned %v fields, want 2", len(all))
		}

		// HDel
		if err := r.HDel(ctx, key, "field1").Err(); err != nil {
			t.Errorf("HDel() error = %v", err)
		}

		// Cleanup
		r.Del(ctx, key)
	})

	// List 操作
	t.Run("List Operations", func(t *testing.T) {
		key := "test:standalone:list"

		// LPush
		if err := r.LPush(ctx, key, "item1", "item2", "item3").Err(); err != nil {
			t.Errorf("LPush() error = %v", err)
		}

		// LRange
		items, err := r.LRange(ctx, key, 0, -1).Result()
		if err != nil {
			t.Errorf("LRange() error = %v", err)
		}
		if len(items) != 3 {
			t.Errorf("LRange() returned %v items, want 3", len(items))
		}

		// LPop
		item, err := r.LPop(ctx, key).Result()
		if err != nil {
			t.Errorf("LPop() error = %v", err)
		}
		if item != "item3" {
			t.Errorf("LPop() = %v, want item3", item)
		}

		// Cleanup
		r.Del(ctx, key)
	})

	// Set 操作
	t.Run("Set Operations", func(t *testing.T) {
		key := "test:standalone:set"

		// SAdd
		if err := r.SAdd(ctx, key, "member1", "member2", "member3").Err(); err != nil {
			t.Errorf("SAdd() error = %v", err)
		}

		// SMembers
		members, err := r.SMembers(ctx, key).Result()
		if err != nil {
			t.Errorf("SMembers() error = %v", err)
		}
		if len(members) != 3 {
			t.Errorf("SMembers() returned %v members, want 3", len(members))
		}

		// SRem
		if err := r.SRem(ctx, key, "member1").Err(); err != nil {
			t.Errorf("SRem() error = %v", err)
		}

		// Cleanup
		r.Del(ctx, key)
	})

	// 过期时间
	t.Run("Expiration", func(t *testing.T) {
		key := "test:standalone:expire"

		// Set with expiration
		if err := r.Set(ctx, key, "value", 2*time.Second).Err(); err != nil {
			t.Errorf("Set() error = %v", err)
		}

		// TTL
		ttl, err := r.TTL(ctx, key).Result()
		if err != nil {
			t.Errorf("TTL() error = %v", err)
		}
		if ttl <= 0 || ttl > 2*time.Second {
			t.Errorf("TTL() = %v, want <= 2s", ttl)
		}

		// Expire
		if err := r.Expire(ctx, key, 5*time.Second).Err(); err != nil {
			t.Errorf("Expire() error = %v", err)
		}

		// Cleanup
		r.Del(ctx, key)
	})

	// Ping
	t.Run("Ping", func(t *testing.T) {
		if err := r.Ping(ctx); err != nil {
			t.Errorf("Ping() error = %v", err)
		}
	})
}

// TestMasterSlaveMode 测试主从模式
func TestMasterSlaveMode(t *testing.T) {
	ctx := context.Background()

	// 创建主从模式 Redis
	r, err := New(ctx,
		WithMasterSlave(
			NodeConfig{Host: "localhost", Port: 6379},
			[]NodeConfig{
				{Host: "localhost", Port: 6380},
				{Host: "localhost", Port: 6381},
			},
		),
		WithReadPolicy(ReadPolicySlaveFirst),
		WithLoadBalancePolicy(LoadBalanceRandom),
	)
	if err != nil {
		t.Skipf("Redis master-slave not available: %v", err)
	}
	defer r.Close()

	// 验证模式
	if r.GetMode() != ModeMasterSlave {
		t.Errorf("Expected master-slave mode, got %v", r.GetMode())
	}

	// 写操作（应该路由到主库）
	t.Run("Write to Master", func(t *testing.T) {
		key := "test:masterslave:key"
		value := "test_value"

		if err := r.Set(ctx, key, value, time.Minute).Err(); err != nil {
			t.Errorf("Set() error = %v", err)
		}

		// 等待主从同步
		time.Sleep(100 * time.Millisecond)

		// 读操作（应该路由到从库）
		result, err := r.Get(ctx, key).Result()
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}
		if result != value {
			t.Errorf("Get() = %v, want %v", result, value)
		}

		// Cleanup
		r.Del(ctx, key)
	})

	// 显式主库操作
	t.Run("Explicit Master", func(t *testing.T) {
		key := "test:masterslave:master"

		master := r.Master()
		if err := master.Set(ctx, key, "master_value", time.Minute).Err(); err != nil {
			t.Errorf("Master.Set() error = %v", err)
		}

		// Cleanup
		r.Del(ctx, key)
	})

	// 显式从库操作
	t.Run("Explicit Slave", func(t *testing.T) {
		key := "test:masterslave:slave"

		// 先写入主库
		r.Set(ctx, key, "slave_value", time.Minute)
		time.Sleep(100 * time.Millisecond)

		// 从从库读取
		slave := r.Slave()
		result, err := slave.Get(ctx, key).Result()
		if err != nil {
			t.Errorf("Slave.Get() error = %v", err)
		}
		if result != "slave_value" {
			t.Errorf("Slave.Get() = %v, want slave_value", result)
		}

		// Cleanup
		r.Del(ctx, key)
	})

	// Ping
	t.Run("Ping", func(t *testing.T) {
		if err := r.Ping(ctx); err != nil {
			t.Errorf("Ping() error = %v", err)
		}

		if err := r.PingMaster(ctx); err != nil {
			t.Errorf("PingMaster() error = %v", err)
		}
	})
}

// TestSentinelMode 测试哨兵模式
func TestSentinelMode(t *testing.T) {
	ctx := context.Background()

	// 创建哨兵模式 Redis
	r, err := New(ctx,
		WithSentinel("mymaster", []string{"localhost:26379", "localhost:26380", "localhost:26381"}, ""),
		WithDB(0),
	)
	if err != nil {
		t.Skipf("Redis sentinel not available: %v", err)
	}
	defer r.Close()

	// 验证模式
	if r.GetMode() != ModeSentinel {
		t.Errorf("Expected sentinel mode, got %v", r.GetMode())
	}

	// 基本操作
	t.Run("Basic Operations", func(t *testing.T) {
		key := "test:sentinel:key"
		value := "test_value"

		if err := r.Set(ctx, key, value, time.Minute).Err(); err != nil {
			t.Errorf("Set() error = %v", err)
		}

		result, err := r.Get(ctx, key).Result()
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}
		if result != value {
			t.Errorf("Get() = %v, want %v", result, value)
		}

		r.Del(ctx, key)
	})

	// Ping
	t.Run("Ping", func(t *testing.T) {
		if err := r.Ping(ctx); err != nil {
			t.Errorf("Ping() error = %v", err)
		}
	})
}

// TestClusterMode 测试集群模式
func TestClusterMode(t *testing.T) {
	ctx := context.Background()

	// 创建集群模式 Redis
	r, err := New(ctx,
		WithCluster(
			[]string{"localhost:7000", "localhost:7001", "localhost:7002"},
			"",
		),
		WithClusterReadOnly(true),
		WithClusterRouteRandomly(true),
	)
	if err != nil {
		t.Skipf("Redis cluster not available: %v", err)
	}
	defer r.Close()

	// 验证模式
	if r.GetMode() != ModeCluster {
		t.Errorf("Expected cluster mode, got %v", r.GetMode())
	}

	// 基本操作
	t.Run("Basic Operations", func(t *testing.T) {
		key := "test:cluster:key"
		value := "test_value"

		if err := r.Set(ctx, key, value, time.Minute).Err(); err != nil {
			t.Errorf("Set() error = %v", err)
		}

		result, err := r.Get(ctx, key).Result()
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}
		if result != value {
			t.Errorf("Get() = %v, want %v", result, value)
		}

		r.Del(ctx, key)
	})

	// 多键操作（集群模式下可能需要相同的 slot）
	t.Run("Multi Key Operations", func(t *testing.T) {
		// 使用 hash tag 确保键在同一个 slot
		key1 := "{user:1}:name"
		key2 := "{user:1}:age"

		r.Set(ctx, key1, "Alice", time.Minute)
		r.Set(ctx, key2, "25", time.Minute)

		name, _ := r.Get(ctx, key1).Result()
		age, _ := r.Get(ctx, key2).Result()

		if name != "Alice" {
			t.Errorf("Get(%v) = %v, want Alice", key1, name)
		}
		if age != "25" {
			t.Errorf("Get(%v) = %v, want 25", key2, age)
		}

		r.Del(ctx, key1, key2)
	})

	// Ping
	t.Run("Ping", func(t *testing.T) {
		if err := r.Ping(ctx); err != nil {
			t.Errorf("Ping() error = %v", err)
		}
	})
}

// TestConfig 测试配置验证
func TestConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "Valid standalone config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "Invalid standalone config (empty host)",
			config: &Config{
				Host: "",
				Port: 6379,
			},
			wantErr: true,
		},
		{
			name: "Invalid standalone config (invalid port)",
			config: &Config{
				Host: "localhost",
				Port: 0,
			},
			wantErr: true,
		},
		{
			name:    "Valid master-slave config",
			config:  DefaultMasterSlaveConfig(),
			wantErr: false,
		},
		{
			name:    "Valid sentinel config",
			config:  DefaultSentinelConfig(),
			wantErr: false,
		},
		{
			name:    "Valid cluster config",
			config:  DefaultClusterConfig(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
