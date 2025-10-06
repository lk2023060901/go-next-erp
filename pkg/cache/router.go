package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Router 读写分离路由器
type Router struct {
	cluster *Cluster
	config  *Config
	logger  *logger.Logger
}

// NewRouter 创建路由器
func NewRouter(cluster *Cluster, cfg *Config) *Router {
	return &Router{
		cluster: cluster,
		config:  cfg,
		logger:  logger.GetLogger().With(zap.String("module", "cache.router")),
	}
}

// Route 路由请求（自动判断读写）
func (r *Router) Route(ctx context.Context, cmd string) (redis.UniversalClient, error) {
	if r.isWriteCommand(cmd) {
		r.logger.Debug("Routing to master (write command)", zap.String("cmd", cmd))
		return r.cluster.GetMaster(), nil
	}

	return r.RouteRead(ctx)
}

// RouteRead 路由读请求
func (r *Router) RouteRead(ctx context.Context) (redis.UniversalClient, error) {
	policy := r.config.ReadPolicy
	if policy == "" {
		policy = ReadPolicyMasterFirst // 默认策略
	}

	switch policy {
	case ReadPolicyMaster:
		// 总是读主库
		r.logger.Debug("Routing to master (policy: master)")
		return r.cluster.GetMaster(), nil

	case ReadPolicySlave:
		// 总是读从库
		client, err := r.cluster.GetSlave()
		if err != nil {
			r.logger.Warn("Failed to get slave, no fallback", zap.Error(err))
			return nil, fmt.Errorf("no slaves available: %w", err)
		}
		r.logger.Debug("Routing to slave (policy: slave)")
		return client, nil

	case ReadPolicyMasterFirst:
		// 优先主库，主库故障读从库
		if r.isMasterHealthy() {
			r.logger.Debug("Routing to master (policy: master_first)")
			return r.cluster.GetMaster(), nil
		}

		client, err := r.cluster.GetSlave()
		if err != nil {
			r.logger.Warn("Master and slaves unavailable", zap.Error(err))
			// 回退到主库
			return r.cluster.GetMaster(), nil
		}
		r.logger.Warn("Routing to slave (master unavailable)")
		return client, nil

	case ReadPolicySlaveFirst:
		// 优先从库，从库故障读主库
		client, err := r.cluster.GetSlave()
		if err != nil {
			r.logger.Debug("Routing to master (slaves unavailable)", zap.Error(err))
			return r.cluster.GetMaster(), nil
		}
		r.logger.Debug("Routing to slave (policy: slave_first)")
		return client, nil

	default:
		r.logger.Warn("Unknown read policy, using master", zap.String("policy", string(policy)))
		return r.cluster.GetMaster(), nil
	}
}

// RouteMaster 强制路由到主库
func (r *Router) RouteMaster() redis.UniversalClient {
	r.logger.Debug("Routing to master (forced)")
	return r.cluster.GetMaster()
}

// isWriteCommand 判断是否为写命令
func (r *Router) isWriteCommand(cmd string) bool {
	cmd = strings.TrimSpace(strings.ToUpper(cmd))

	writeCommands := []string{
		// String 类型
		"SET", "SETEX", "SETNX", "SETRANGE", "MSET", "MSETNX",
		"APPEND", "INCR", "INCRBY", "INCRBYFLOAT", "DECR", "DECRBY",
		"GETSET", "GETDEL",

		// Hash 类型
		"HSET", "HSETNX", "HMSET", "HINCRBY", "HINCRBYFLOAT", "HDEL",

		// List 类型
		"LPUSH", "LPUSHX", "RPUSH", "RPUSHX", "LPOP", "RPOP",
		"LINSERT", "LSET", "LREM", "LTRIM", "RPOPLPUSH",
		"BLPOP", "BRPOP", "BRPOPLPUSH",

		// Set 类型
		"SADD", "SREM", "SPOP", "SMOVE", "SINTERSTORE", "SUNIONSTORE", "SDIFFSTORE",

		// Sorted Set 类型
		"ZADD", "ZREM", "ZINCRBY", "ZREMRANGEBYRANK", "ZREMRANGEBYSCORE",
		"ZREMRANGEBYLEX", "ZPOPMIN", "ZPOPMAX", "BZPOPMIN", "BZPOPMAX",
		"ZINTERSTORE", "ZUNIONSTORE",

		// Key 管理
		"DEL", "UNLINK", "EXPIRE", "EXPIREAT", "PEXPIRE", "PEXPIREAT",
		"PERSIST", "RENAME", "RENAMENX", "MOVE", "COPY",
		"RESTORE", "MIGRATE",

		// 其他
		"FLUSHDB", "FLUSHALL", "SELECT", "SWAPDB",
		"PUBLISH", // 发布订阅
		"SCRIPT", "EVAL", "EVALSHA", // Lua 脚本（可能有写操作）
	}

	for _, writeCmd := range writeCommands {
		if strings.HasPrefix(cmd, writeCmd) {
			return true
		}
	}

	return false
}

// isMasterHealthy 检查主库是否健康
func (r *Router) isMasterHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.cluster.GetMaster().Ping(ctx).Err()
	return err == nil
}
