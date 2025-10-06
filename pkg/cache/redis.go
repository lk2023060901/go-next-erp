package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Cache 是 Redis 的类型别名（为了向后兼容）
type Cache = Redis

// Redis 封装（支持单机/主从/哨兵/集群）
type Redis struct {
	// 单机模式
	client redis.UniversalClient

	// 主从模式（可选）
	cluster *Cluster
	router  *Router

	config *Config
	logger *logger.Logger

	// 模式标识
	mode RedisMode
}

// New 创建 Redis 连接（自动检测模式）
func New(ctx context.Context, opts ...Option) (*Redis, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	r := &Redis{
		config: cfg,
		logger: logger.GetLogger().With(zap.String("module", "cache")),
		mode:   cfg.GetMode(),
	}

	// 根据配置模式初始化
	switch r.mode {
	case ModeCluster:
		return r.initCluster(ctx)

	case ModeSentinel:
		return r.initSentinel(ctx)

	case ModeMasterSlave:
		return r.initMasterSlave(ctx)

	default: // ModeStandalone
		return r.initStandalone(ctx)
	}
}

// ============ 初始化方法 ============

// initStandalone 初始化单机模式
func (r *Redis) initStandalone(ctx context.Context) (*Redis, error) {
	r.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", r.config.Host, r.config.Port),
		Password: r.config.Password,
		DB:       r.config.DB,

		PoolSize:        r.config.PoolSize,
		MinIdleConns:    r.config.MinIdleConns,
		ConnMaxLifetime: r.config.MaxConnAge,
		PoolTimeout:     r.config.PoolTimeout,
		ConnMaxIdleTime: r.config.IdleTimeout,

		DialTimeout:  r.config.DialTimeout,
		ReadTimeout:  r.config.ReadTimeout,
		WriteTimeout: r.config.WriteTimeout,
	})

	if err := r.client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("standalone ping failed: %w", err)
	}

	r.logger.Info("Redis initialized in standalone mode",
		zap.String("host", r.config.Host),
		zap.Int("port", r.config.Port),
		zap.Int("db", r.config.DB),
	)

	return r, nil
}

// initMasterSlave 初始化主从模式
func (r *Redis) initMasterSlave(ctx context.Context) (*Redis, error) {
	cluster, err := NewCluster(ctx, r.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	r.cluster = cluster
	r.router = NewRouter(cluster, r.config)

	r.logger.Info("Redis initialized in master-slave mode",
		zap.Int("slaves", len(r.config.MasterSlave.Slaves)),
		zap.String("read_policy", string(r.config.ReadPolicy)),
		zap.String("load_balance_policy", string(r.config.LoadBalancePolicy)),
	)

	return r, nil
}

// initSentinel 初始化哨兵模式
func (r *Redis) initSentinel(ctx context.Context) (*Redis, error) {
	r.client = redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:       r.config.Sentinel.MasterName,
		SentinelAddrs:    r.config.Sentinel.SentinelAddrs,
		Password:         r.config.Sentinel.Password,
		DB:               r.config.Sentinel.DB,
		SentinelPassword: r.config.Sentinel.SentinelPassword,

		// 从库只读（支持读写分离）
		ReplicaOnly: r.config.Sentinel.SlaveOnly,

		PoolSize:        r.config.PoolSize,
		MinIdleConns:    r.config.MinIdleConns,
		ConnMaxLifetime: r.config.MaxConnAge,
		PoolTimeout:     r.config.PoolTimeout,
		ConnMaxIdleTime: r.config.IdleTimeout,

		DialTimeout:  r.config.DialTimeout,
		ReadTimeout:  r.config.ReadTimeout,
		WriteTimeout: r.config.WriteTimeout,
	})

	if err := r.client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("sentinel ping failed: %w", err)
	}

	r.logger.Info("Redis initialized in sentinel mode",
		zap.String("master_name", r.config.Sentinel.MasterName),
		zap.Int("sentinels", len(r.config.Sentinel.SentinelAddrs)),
		zap.Bool("slave_only", r.config.Sentinel.SlaveOnly),
	)

	return r, nil
}

// initCluster 初始化集群模式
func (r *Redis) initCluster(ctx context.Context) (*Redis, error) {
	r.client = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    r.config.Cluster.Addrs,
		Password: r.config.Cluster.Password,

		// 集群读写分离配置
		ReadOnly:       r.config.Cluster.ReadOnly,
		RouteByLatency: r.config.Cluster.RouteByLatency,
		RouteRandomly:  r.config.Cluster.RouteRandomly,

		MaxRedirects: r.config.Cluster.MaxRedirects,

		PoolSize:        r.config.PoolSize,
		MinIdleConns:    r.config.MinIdleConns,
		ConnMaxLifetime: r.config.MaxConnAge,
		PoolTimeout:     r.config.PoolTimeout,
		ConnMaxIdleTime: r.config.IdleTimeout,

		DialTimeout:  r.config.DialTimeout,
		ReadTimeout:  r.config.ReadTimeout,
		WriteTimeout: r.config.WriteTimeout,
	})

	if err := r.client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cluster ping failed: %w", err)
	}

	r.logger.Info("Redis initialized in cluster mode",
		zap.Int("nodes", len(r.config.Cluster.Addrs)),
		zap.Bool("read_only", r.config.Cluster.ReadOnly),
		zap.Bool("route_by_latency", r.config.Cluster.RouteByLatency),
		zap.Bool("route_randomly", r.config.Cluster.RouteRandomly),
	)

	return r, nil
}

// ============ 通用接口（自动适配单机/主从/哨兵/集群）============

// GetRaw 获取原始值（返回 Redis 命令）
func (r *Redis) GetRaw(ctx context.Context, key string) *redis.StringCmd {
	if r.cluster != nil {
		client, _ := r.router.RouteRead(ctx)
		return client.Get(ctx, key)
	}
	return r.client.Get(ctx, key)
}

// SetRaw 设置原始值（返回 Redis 命令）
func (r *Redis) SetRaw(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().Set(ctx, key, value, expiration)
	}
	return r.client.Set(ctx, key, value, expiration)
}

// Get 获取缓存值（自动反序列化 JSON）
// 用法: cache.Get(ctx, "key", &value)
func (r *Redis) Get(ctx context.Context, key string, value interface{}) error {
	var result string
	var err error

	if r.cluster != nil {
		client, routeErr := r.router.RouteRead(ctx)
		if routeErr != nil {
			return routeErr
		}
		result, err = client.Get(ctx, key).Result()
	} else {
		result, err = r.client.Get(ctx, key).Result()
	}

	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found")
		}
		return err
	}

	// JSON 反序列化
	if err := json.Unmarshal([]byte(result), value); err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	return nil
}

// Set 设置缓存值（自动序列化 JSON）
// 用法: cache.Set(ctx, "key", value, 300) // 300秒过期
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expirationSeconds int) error {
	// JSON 序列化
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	expiration := time.Duration(expirationSeconds) * time.Second

	if r.cluster != nil {
		return r.router.RouteMaster().Set(ctx, key, data, expiration).Err()
	}
	return r.client.Set(ctx, key, data, expiration).Err()
}

// Del 删除键
func (r *Redis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().Del(ctx, keys...)
	}
	return r.client.Del(ctx, keys...)
}

// Delete 删除缓存键（兼容仓储层接口）
// 用法: cache.Delete(ctx, "key")
func (r *Redis) Delete(ctx context.Context, key string) error {
	if r.cluster != nil {
		return r.router.RouteMaster().Del(ctx, key).Err()
	}
	return r.client.Del(ctx, key).Err()
}

// Exists 判断键是否存在
func (r *Redis) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	if r.cluster != nil {
		client, _ := r.router.RouteRead(ctx)
		return client.Exists(ctx, keys...)
	}
	return r.client.Exists(ctx, keys...)
}

// Expire 设置过期时间
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().Expire(ctx, key, expiration)
	}
	return r.client.Expire(ctx, key, expiration)
}

// TTL 获取剩余过期时间
func (r *Redis) TTL(ctx context.Context, key string) *redis.DurationCmd {
	if r.cluster != nil {
		client, _ := r.router.RouteRead(ctx)
		return client.TTL(ctx, key)
	}
	return r.client.TTL(ctx, key)
}

// Incr 自增
func (r *Redis) Incr(ctx context.Context, key string) *redis.IntCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().Incr(ctx, key)
	}
	return r.client.Incr(ctx, key)
}

// Decr 自减
func (r *Redis) Decr(ctx context.Context, key string) *redis.IntCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().Decr(ctx, key)
	}
	return r.client.Decr(ctx, key)
}

// ============ Hash 操作 ============

// HGet 获取哈希字段值
func (r *Redis) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	if r.cluster != nil {
		client, _ := r.router.RouteRead(ctx)
		return client.HGet(ctx, key, field)
	}
	return r.client.HGet(ctx, key, field)
}

// HSet 设置哈希字段值
func (r *Redis) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().HSet(ctx, key, values...)
	}
	return r.client.HSet(ctx, key, values...)
}

// HGetAll 获取所有哈希字段
func (r *Redis) HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd {
	if r.cluster != nil {
		client, _ := r.router.RouteRead(ctx)
		return client.HGetAll(ctx, key)
	}
	return r.client.HGetAll(ctx, key)
}

// HDel 删除哈希字段
func (r *Redis) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().HDel(ctx, key, fields...)
	}
	return r.client.HDel(ctx, key, fields...)
}

// ============ List 操作 ============

// LPush 从左侧推入列表
func (r *Redis) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().LPush(ctx, key, values...)
	}
	return r.client.LPush(ctx, key, values...)
}

// RPush 从右侧推入列表
func (r *Redis) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().RPush(ctx, key, values...)
	}
	return r.client.RPush(ctx, key, values...)
}

// LPop 从左侧弹出列表
func (r *Redis) LPop(ctx context.Context, key string) *redis.StringCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().LPop(ctx, key)
	}
	return r.client.LPop(ctx, key)
}

// RPop 从右侧弹出列表
func (r *Redis) RPop(ctx context.Context, key string) *redis.StringCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().RPop(ctx, key)
	}
	return r.client.RPop(ctx, key)
}

// LRange 获取列表范围
func (r *Redis) LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	if r.cluster != nil {
		client, _ := r.router.RouteRead(ctx)
		return client.LRange(ctx, key, start, stop)
	}
	return r.client.LRange(ctx, key, start, stop)
}

// ============ Set 操作 ============

// SAdd 添加集合成员
func (r *Redis) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().SAdd(ctx, key, members...)
	}
	return r.client.SAdd(ctx, key, members...)
}

// SMembers 获取所有集合成员
func (r *Redis) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	if r.cluster != nil {
		client, _ := r.router.RouteRead(ctx)
		return client.SMembers(ctx, key)
	}
	return r.client.SMembers(ctx, key)
}

// SRem 删除集合成员
func (r *Redis) SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().SRem(ctx, key, members...)
	}
	return r.client.SRem(ctx, key, members...)
}

// ============ Sorted Set 操作 ============

// ZAdd 添加有序集合成员
func (r *Redis) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().ZAdd(ctx, key, members...)
	}
	return r.client.ZAdd(ctx, key, members...)
}

// ZRange 获取有序集合范围
func (r *Redis) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	if r.cluster != nil {
		client, _ := r.router.RouteRead(ctx)
		return client.ZRange(ctx, key, start, stop)
	}
	return r.client.ZRange(ctx, key, start, stop)
}

// ZRem 删除有序集合成员
func (r *Redis) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	if r.cluster != nil {
		return r.router.RouteMaster().ZRem(ctx, key, members...)
	}
	return r.client.ZRem(ctx, key, members...)
}

// ============ 事务/管道 ============

// TxPipeline 创建事务管道
func (r *Redis) TxPipeline() redis.Pipeliner {
	if r.cluster != nil {
		return r.router.RouteMaster().TxPipeline()
	}
	return r.client.TxPipeline()
}

// Pipeline 创建管道
func (r *Redis) Pipeline() redis.Pipeliner {
	if r.cluster != nil {
		return r.router.RouteMaster().Pipeline()
	}
	return r.client.Pipeline()
}

// ============ 主从模式专用方法 ============

// Master 获取主库操作器
func (r *Redis) Master() redis.UniversalClient {
	if r.cluster != nil {
		return r.cluster.GetMaster()
	}
	return r.client
}

// Slave 获取从库操作器
func (r *Redis) Slave() redis.UniversalClient {
	if r.cluster != nil {
		client, _ := r.cluster.GetSlave()
		if client == nil {
			// 回退到主库
			return r.cluster.GetMaster()
		}
		return client
	}
	return r.client
}

// IsMasterSlaveMode 判断当前是否为主从模式
func (r *Redis) IsMasterSlaveMode() bool {
	return r.mode == ModeMasterSlave
}

// GetMode 获取当前部署模式
func (r *Redis) GetMode() RedisMode {
	return r.mode
}

// ============ 健康检查 ============

// Ping 检查连接
func (r *Redis) Ping(ctx context.Context) error {
	if r.cluster != nil {
		// 检查主库
		if err := r.cluster.GetMaster().Ping(ctx).Err(); err != nil {
			return fmt.Errorf("master ping failed: %w", err)
		}

		// 检查从库
		slaves := r.cluster.GetHealthySlaves()
		if len(r.config.MasterSlave.Slaves) > 0 && len(slaves) == 0 {
			r.logger.Warn("No healthy slaves available")
		}

		return nil
	}

	return r.client.Ping(ctx).Err()
}

// PingMaster 检查主库（仅主从模式）
func (r *Redis) PingMaster(ctx context.Context) error {
	if r.cluster == nil {
		return r.client.Ping(ctx).Err()
	}
	return r.cluster.GetMaster().Ping(ctx).Err()
}

// ============ 工具方法 ============

// Close 关闭连接
func (r *Redis) Close() error {
	if r.cluster != nil {
		return r.cluster.Close()
	}
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// Client 获取底层客户端（通用模式）
func (r *Redis) Client() redis.UniversalClient {
	if r.cluster != nil {
		r.logger.Warn("Client() called in master-slave mode, returning master client")
		return r.cluster.GetMaster()
	}
	return r.client
}
