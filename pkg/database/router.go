package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
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
		logger:  logger.GetLogger().With(zap.String("module", "database.router")),
	}
}

// Route 路由请求（自动判断读写）
func (r *Router) Route(ctx context.Context, query string) (*pgxpool.Pool, error) {
	if r.isWriteQuery(query) {
		r.logger.Debug("Routing to master (write query)", zap.String("query", truncateQuery(query)))
		return r.cluster.GetMaster(), nil
	}

	return r.RouteRead(ctx)
}

// RouteRead 路由读请求
func (r *Router) RouteRead(ctx context.Context) (*pgxpool.Pool, error) {
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
		pool, err := r.cluster.GetSlave()
		if err != nil {
			r.logger.Warn("Failed to get slave, no fallback", zap.Error(err))
			return nil, fmt.Errorf("no slaves available: %w", err)
		}
		r.logger.Debug("Routing to slave (policy: slave)")
		return pool, nil

	case ReadPolicyMasterFirst:
		// 优先主库，主库故障读从库
		if r.isMasterHealthy() {
			r.logger.Debug("Routing to master (policy: master_first)")
			return r.cluster.GetMaster(), nil
		}

		pool, err := r.cluster.GetSlave()
		if err != nil {
			r.logger.Warn("Master and slaves unavailable", zap.Error(err))
			// 回退到主库
			return r.cluster.GetMaster(), nil
		}
		r.logger.Warn("Routing to slave (master unavailable)")
		return pool, nil

	case ReadPolicySlaveFirst:
		// 优先从库，从库故障读主库
		pool, err := r.cluster.GetSlave()
		if err != nil {
			r.logger.Debug("Routing to master (slaves unavailable)", zap.Error(err))
			return r.cluster.GetMaster(), nil
		}
		r.logger.Debug("Routing to slave (policy: slave_first)")
		return pool, nil

	default:
		r.logger.Warn("Unknown read policy, using master", zap.String("policy", string(policy)))
		return r.cluster.GetMaster(), nil
	}
}

// RouteMaster 强制路由到主库
func (r *Router) RouteMaster() *pgxpool.Pool {
	r.logger.Debug("Routing to master (forced)")
	return r.cluster.GetMaster()
}

// isWriteQuery 判断是否为写查询
func (r *Router) isWriteQuery(query string) bool {
	query = strings.TrimSpace(strings.ToUpper(query))

	writeKeywords := []string{
		"INSERT", "UPDATE", "DELETE",
		"CREATE", "DROP", "ALTER",
		"TRUNCATE", "REPLACE",
		"GRANT", "REVOKE",
	}

	for _, keyword := range writeKeywords {
		if strings.HasPrefix(query, keyword) {
			return true
		}
	}

	return false
}

// isMasterHealthy 检查主库是否健康
func (r *Router) isMasterHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.cluster.GetMaster().Ping(ctx)
	return err == nil
}

// truncateQuery 截断查询字符串（用于日志）
func truncateQuery(query string) string {
	const maxLen = 100
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "..."
}
