package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lk2023060901/go-next-erp/pkg/logger"
	"go.uber.org/zap"
)

// DB PostgreSQL 数据库封装（支持单机/主从）
type DB struct {
	// 单机模式
	pool *pgxpool.Pool

	// 主从模式（可选）
	cluster *Cluster
	router  *Router

	config *Config
	logger *logger.Logger

	// 模式标识
	isMasterSlave bool
}

// New 创建数据库连接（自动检测单机/主从）
func New(ctx context.Context, opts ...Option) (*DB, error) {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	db := &DB{
		config:        cfg,
		logger:        logger.GetLogger().With(zap.String("module", "database")),
		isMasterSlave: cfg.IsMasterSlaveMode(),
	}

	// 根据配置模式初始化
	if db.isMasterSlave {
		// 主从模式
		cluster, err := NewCluster(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create cluster: %w", err)
		}
		db.cluster = cluster
		db.router = NewRouter(cluster, cfg)

		db.logger.Info("Database initialized in master-slave mode",
			zap.Int("slaves", len(cfg.Slaves)),
			zap.String("read_policy", string(cfg.ReadPolicy)),
			zap.String("load_balance_policy", string(cfg.LoadBalancePolicy)),
		)
	} else {
		// 单机模式
		pool, err := createPool(ctx, cfg.ToNodeConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to create pool: %w", err)
		}
		db.pool = pool

		db.logger.Info("Database initialized in standalone mode",
			zap.String("host", cfg.Host),
			zap.Int("port", cfg.Port),
			zap.String("database", cfg.Database),
		)
	}

	return db, nil
}

// ============ 查询接口（自动适配单机/主从）============

// Query 执行查询（返回多行）
// 单机：使用唯一连接池
// 主从：自动路由到从库
func (db *DB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	start := time.Now()

	var rows pgx.Rows
	var err error

	if db.isMasterSlave {
		pool, routeErr := db.router.RouteRead(ctx)
		if routeErr != nil {
			return nil, routeErr
		}
		rows, err = pool.Query(ctx, sql, args...)
	} else {
		rows, err = db.pool.Query(ctx, sql, args...)
	}

	db.logQuery(ctx, "Query", sql, time.Since(start), err)
	return rows, err
}

// QueryRow 执行查询（返回单行）
// 单机：使用唯一连接池
// 主从：自动路由到从库
func (db *DB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	start := time.Now()

	var row pgx.Row

	if db.isMasterSlave {
		pool, _ := db.router.RouteRead(ctx)
		row = pool.QueryRow(ctx, sql, args...)
	} else {
		row = db.pool.QueryRow(ctx, sql, args...)
	}

	db.logQuery(ctx, "QueryRow", sql, time.Since(start), nil)
	return row
}

// Exec 执行命令（INSERT/UPDATE/DELETE）
// 单机：使用唯一连接池
// 主从：强制路由到主库
func (db *DB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	start := time.Now()

	var tag pgconn.CommandTag
	var err error

	if db.isMasterSlave {
		pool := db.router.RouteMaster()
		tag, err = pool.Exec(ctx, sql, args...)
	} else {
		tag, err = db.pool.Exec(ctx, sql, args...)
	}

	db.logQuery(ctx, "Exec", sql, time.Since(start), err)
	return tag, err
}

// ============ 批量操作 ============

// SendBatch 发送批量操作
func (db *DB) SendBatch(ctx context.Context, batch *pgx.Batch) pgx.BatchResults {
	if db.isMasterSlave {
		return db.router.RouteMaster().SendBatch(ctx, batch)
	}
	return db.pool.SendBatch(ctx, batch)
}

// ============ 主从模式专用方法 ============

// Master 获取主库操作器
func (db *DB) Master() *MasterDB {
	if !db.isMasterSlave {
		return &MasterDB{pool: db.pool, logger: db.logger}
	}
	return &MasterDB{pool: db.cluster.GetMaster(), logger: db.logger}
}

// Slave 获取从库操作器
func (db *DB) Slave() *SlaveDB {
	if !db.isMasterSlave {
		return &SlaveDB{pool: db.pool, logger: db.logger}
	}

	pool, _ := db.cluster.GetSlave()
	if pool == nil {
		// 回退到主库
		pool = db.cluster.GetMaster()
	}
	return &SlaveDB{pool: pool, logger: db.logger}
}

// IsMasterSlaveMode 判断当前是否为主从模式
func (db *DB) IsMasterSlaveMode() bool {
	return db.isMasterSlave
}

// ============ 事务（单机/主从通用，强制主库）============

// Begin 开始事务
func (db *DB) Begin(ctx context.Context) (pgx.Tx, error) {
	if db.isMasterSlave {
		return db.cluster.GetMaster().Begin(ctx)
	}
	return db.pool.Begin(ctx)
}

// BeginTx 开始事务（带选项）
func (db *DB) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	if db.isMasterSlave {
		return db.cluster.GetMaster().BeginTx(ctx, txOptions)
	}
	return db.pool.BeginTx(ctx, txOptions)
}

// Transaction 执行事务（自动管理提交/回滚）
func (db *DB) Transaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // 重新抛出 panic
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	err = fn(tx)
	return err
}

// ============ 健康检查 ============

// Ping 检查连接
func (db *DB) Ping(ctx context.Context) error {
	if db.isMasterSlave {
		// 检查主库
		if err := db.cluster.GetMaster().Ping(ctx); err != nil {
			return fmt.Errorf("master ping failed: %w", err)
		}

		// 检查从库
		slaves := db.cluster.GetHealthySlaves()
		if len(db.config.Slaves) > 0 && len(slaves) == 0 {
			db.logger.Warn("No healthy slaves available")
		}

		return nil
	}

	return db.pool.Ping(ctx)
}

// PingMaster 检查主库（仅主从模式）
func (db *DB) PingMaster(ctx context.Context) error {
	if !db.isMasterSlave {
		return db.pool.Ping(ctx)
	}
	return db.cluster.GetMaster().Ping(ctx)
}

// PingSlaves 检查所有从库（仅主从模式）
func (db *DB) PingSlaves(ctx context.Context) error {
	if !db.isMasterSlave {
		return nil
	}

	slaves := db.cluster.GetHealthySlaves()
	if len(slaves) == 0 {
		return fmt.Errorf("no healthy slaves")
	}

	for _, slave := range slaves {
		if err := slave.pool.Ping(ctx); err != nil {
			return fmt.Errorf("slave %s ping failed: %w", slave.config.Host, err)
		}
	}

	return nil
}

// ============ 工具方法 ============

// Close 关闭连接池
func (db *DB) Close() {
	if db.isMasterSlave {
		db.cluster.Close()
	} else {
		db.pool.Close()
	}
	db.logger.Info("Database closed")
}

// Pool 获取底层连接池（单机模式）
func (db *DB) Pool() *pgxpool.Pool {
	if db.isMasterSlave {
		db.logger.Warn("Pool() called in master-slave mode, returning master pool")
		return db.cluster.GetMaster()
	}
	return db.pool
}

// Stats 获取连接池统计
func (db *DB) Stats() *pgxpool.Stat {
	if db.isMasterSlave {
		return db.cluster.GetMaster().Stat()
	}
	return db.pool.Stat()
}

// logQuery 记录查询日志
func (db *DB) logQuery(ctx context.Context, method, sql string, duration time.Duration, err error) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.Duration("duration", duration),
	}

	// 提取 Context 字段
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		db.logger.Error("Query failed", fields...)
	} else if duration > 1*time.Second {
		// 慢查询告警
		fields = append(fields, zap.String("sql", truncateQuery(sql)))
		db.logger.Warn("Slow query detected", fields...)
	} else {
		// 正常日志（仅 debug 级别）
		db.logger.Debug("Query executed", fields...)
	}
}

// ============ 辅助类型 ============

// MasterDB 主库操作器（显式主库操作）
type MasterDB struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

// Query 查询（主库）
func (m *MasterDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return m.pool.Query(ctx, sql, args...)
}

// QueryRow 查询单行（主库）
func (m *MasterDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return m.pool.QueryRow(ctx, sql, args...)
}

// Exec 执行命令（主库）
func (m *MasterDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return m.pool.Exec(ctx, sql, args...)
}

// SlaveDB 从库操作器（显式从库操作）
type SlaveDB struct {
	pool   *pgxpool.Pool
	logger *logger.Logger
}

// Query 查询（从库）
func (s *SlaveDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return s.pool.Query(ctx, sql, args...)
}

// QueryRow 查询单行（从库）
func (s *SlaveDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return s.pool.QueryRow(ctx, sql, args...)
}
