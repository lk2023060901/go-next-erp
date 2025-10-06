# pkg/cache Redis 使用文档

## 目录结构

```
pkg/cache/
├── config.go         # 配置定义（支持单机/主从/哨兵/集群）
├── options.go        # 配置选项（Functional Options）
├── loadbalancer.go   # 负载均衡器（Random/RoundRobin/LeastConn/Weighted）
├── masterslave.go    # 主从集群管理
├── router.go         # 读写分离路由
├── redis.go          # Redis 核心封装
├── redis_test.go     # 单元测试
└── USAGE.md          # 使用文档
```

---

## 部署模式

### 1. 单机模式（Standalone）

**特点**：
- 单个 Redis 实例
- 无高可用
- 适用于开发环境、小规模应用

**使用示例**：

```go
package main

import (
    "context"
    "time"
    "github.com/lk2023060901/go-next-erp/pkg/cache"
)

func main() {
    ctx := context.Background()

    // 方式 1：使用默认配置
    r, err := cache.New(ctx)
    if err != nil {
        panic(err)
    }
    defer r.Close()

    // 方式 2：使用选项配置
    r, err = cache.New(ctx,
        cache.WithHost("localhost"),
        cache.WithPort(6379),
        cache.WithPassword("your_password"),
        cache.WithDB(0),
        cache.WithPoolSize(20),
    )

    // 基本操作
    r.Set(ctx, "key", "value", time.Minute)
    val, _ := r.Get(ctx, "key").Result()
    println(val) // "value"
}
```

---

### 2. 主从模式（Master-Slave）

**特点**：
- 1 个主库 + N 个从库
- 支持读写分离
- 手动故障转移
- 适用于读多写少场景

**使用示例**：

```go
package main

import (
    "context"
    "github.com/lk2023060901/go-next-erp/pkg/cache"
)

func main() {
    ctx := context.Background()

    // 方式 1：使用完整配置
    r, err := cache.New(ctx,
        cache.WithMasterSlave(
            cache.NodeConfig{
                Host:     "master.redis.com",
                Port:     6379,
                Password: "password",
                DB:       0,
                Weight:   1,
            },
            []cache.NodeConfig{
                {Host: "slave1.redis.com", Port: 6379, Password: "password", Weight: 1},
                {Host: "slave2.redis.com", Port: 6379, Password: "password", Weight: 2},
            },
        ),
        cache.WithReadPolicy(cache.ReadPolicySlaveFirst),          // 优先从库
        cache.WithLoadBalancePolicy(cache.LoadBalanceWeighted),   // 加权负载均衡
    )

    // 方式 2：使用快捷配置
    r, err = cache.New(ctx,
        cache.WithMasterSlaveConfig(
            "master.redis.com", 6379,
            []string{"slave1.redis.com", "slave2.redis.com"},
            "password",
        ),
    )

    // 写操作（自动路由到主库）
    r.Set(ctx, "user:1", "Alice", 0)

    // 读操作（自动路由到从库）
    val, _ := r.Get(ctx, "user:1").Result()

    // 显式主库操作
    master := r.Master()
    master.Set(ctx, "key", "value", 0)

    // 显式从库操作
    slave := r.Slave()
    val, _ = slave.Get(ctx, "key").Result()
}
```

**读写分离策略**：

```go
// ReadPolicyMaster - 总是读主库
cache.WithReadPolicy(cache.ReadPolicyMaster)

// ReadPolicySlave - 总是读从库（从库故障则失败）
cache.WithReadPolicy(cache.ReadPolicySlave)

// ReadPolicyMasterFirst - 优先主库，主库故障读从库
cache.WithReadPolicy(cache.ReadPolicyMasterFirst)

// ReadPolicySlaveFirst - 优先从库，从库故障读主库（推荐）
cache.WithReadPolicy(cache.ReadPolicySlaveFirst)
```

**负载均衡策略**：

```go
// LoadBalanceRandom - 随机（默认）
cache.WithLoadBalancePolicy(cache.LoadBalanceRandom)

// LoadBalanceRoundRobin - 轮询
cache.WithLoadBalancePolicy(cache.LoadBalanceRoundRobin)

// LoadBalanceLeastConn - 最少连接
cache.WithLoadBalancePolicy(cache.LoadBalanceLeastConn)

// LoadBalanceWeighted - 加权（基于 NodeConfig.Weight）
cache.WithLoadBalancePolicy(cache.LoadBalanceWeighted)
```

---

### 3. 哨兵模式（Sentinel）

**特点**：
- 主从模式 + 哨兵进程
- 自动故障转移
- 支持读写分离
- 适用于生产环境

**使用示例**：

```go
package main

import (
    "context"
    "github.com/lk2023060901/go-next-erp/pkg/cache"
)

func main() {
    ctx := context.Background()

    // 方式 1：使用完整配置
    r, err := cache.New(ctx,
        cache.WithSentinel(
            "mymaster",                        // 主节点名称
            []string{                          // 哨兵地址列表
                "sentinel1.redis.com:26379",
                "sentinel2.redis.com:26379",
                "sentinel3.redis.com:26379",
            },
            "redis_password",                  // Redis 密码
        ),
        cache.WithSentinelPassword("sentinel_password"), // 哨兵密码
        cache.WithDB(0),
    )

    // 方式 2：使用快捷配置
    r, err = cache.New(ctx,
        cache.WithSentinelConfig(
            "mymaster",
            []string{"sentinel1:26379", "sentinel2:26379", "sentinel3:26379"},
            "password",
        ),
    )

    // 自动故障转移
    // 当主库故障时，哨兵会自动提升从库为主库
    r.Set(ctx, "key", "value", 0)
    val, _ := r.Get(ctx, "key").Result()
}
```

**哨兵只读模式**（只读从库）：

```go
r, err := cache.New(ctx,
    cache.WithSentinel("mymaster", sentinelAddrs, "password"),
    cache.WithSlaveOnly(true), // 只读从库（不会自动切换到主库）
)
```

---

### 4. 集群模式（Cluster）

**特点**：
- 多主多从（16384 个哈希槽分片）
- 自动故障转移
- 支持读写分离
- 适用于大规模、高可用场景

**使用示例**：

```go
package main

import (
    "context"
    "github.com/lk2023060901/go-next-erp/pkg/cache"
)

func main() {
    ctx := context.Background()

    // 方式 1：使用完整配置
    r, err := cache.New(ctx,
        cache.WithCluster(
            []string{                          // 集群节点地址
                "node1.redis.com:7000",
                "node2.redis.com:7001",
                "node3.redis.com:7002",
                "node4.redis.com:7003",
                "node5.redis.com:7004",
                "node6.redis.com:7005",
            },
            "password",
        ),
        cache.WithClusterReadOnly(true),        // 从从节点读
        cache.WithClusterRouteRandomly(true),   // 随机路由
        cache.WithClusterRouteByLatency(false), // 按延迟路由
        cache.WithMaxRedirects(3),             // 最大重定向次数
    )

    // 方式 2：使用快捷配置
    r, err = cache.New(ctx,
        cache.WithClusterConfig(
            []string{"node1:7000", "node2:7001", "node3:7002"},
            "password",
            true, // readOnly
        ),
    )

    // 普通操作（自动分片）
    r.Set(ctx, "user:1", "Alice", 0)
    r.Set(ctx, "user:2", "Bob", 0)

    // 使用 Hash Tag 确保键在同一槽（支持事务）
    r.Set(ctx, "{user:1}:name", "Alice", 0)
    r.Set(ctx, "{user:1}:age", "25", 0)
}
```

---

## 通用操作

### 1. String 操作

```go
// Set
r.Set(ctx, "key", "value", time.Minute)

// Get
val, err := r.Get(ctx, "key").Result()

// Del
r.Del(ctx, "key1", "key2")

// Incr/Decr
r.Incr(ctx, "counter")
r.Decr(ctx, "counter")

// Expire
r.Expire(ctx, "key", time.Hour)

// TTL
ttl, _ := r.TTL(ctx, "key").Result()
```

### 2. Hash 操作

```go
// HSet
r.HSet(ctx, "user:1", "name", "Alice", "age", "25")

// HGet
name, _ := r.HGet(ctx, "user:1", "name").Result()

// HGetAll
user, _ := r.HGetAll(ctx, "user:1").Result()

// HDel
r.HDel(ctx, "user:1", "age")
```

### 3. List 操作

```go
// LPush
r.LPush(ctx, "queue", "task1", "task2")

// RPush
r.RPush(ctx, "queue", "task3")

// LPop
task, _ := r.LPop(ctx, "queue").Result()

// RPop
task, _ = r.RPop(ctx, "queue").Result()

// LRange
tasks, _ := r.LRange(ctx, "queue", 0, -1).Result()
```

### 4. Set 操作

```go
// SAdd
r.SAdd(ctx, "tags", "go", "redis", "cache")

// SMembers
tags, _ := r.SMembers(ctx, "tags").Result()

// SRem
r.SRem(ctx, "tags", "cache")
```

### 5. Sorted Set 操作

```go
// ZAdd
r.ZAdd(ctx, "leaderboard",
    redis.Z{Score: 100, Member: "Alice"},
    redis.Z{Score: 90, Member: "Bob"},
)

// ZRange
top10, _ := r.ZRange(ctx, "leaderboard", 0, 9).Result()

// ZRem
r.ZRem(ctx, "leaderboard", "Bob")
```

---

## 高级功能

### 1. 事务（Transaction）

```go
// 使用 Pipeline
pipe := r.TxPipeline()
pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
pipe.Incr(ctx, "counter")
_, err := pipe.Exec(ctx)
```

### 2. 管道（Pipeline）

```go
pipe := r.Pipeline()
pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
_, err := pipe.Exec(ctx)
```

### 3. 健康检查

```go
// 检查连接
err := r.Ping(ctx)

// 检查主库（仅主从模式）
err = r.PingMaster(ctx)

// 判断模式
if r.GetMode() == cache.ModeMasterSlave {
    println("Master-Slave mode enabled")
}
```

### 4. 健康检查配置

```go
r, err := cache.New(ctx,
    cache.WithMasterSlaveConfig(...),
    cache.WithHealthCheck(
        true,               // 启用
        30*time.Second,     // 检查间隔
        5*time.Second,      // 超时时间
    ),
)
```

---

## 配置文件（YAML）

### 单机模式

```yaml
host: localhost
port: 6379
password: ""
db: 0

pool_size: 10
min_idle_conns: 2
pool_timeout: 4s
idle_timeout: 5m

dial_timeout: 5s
read_timeout: 3s
write_timeout: 3s
```

### 主从模式

```yaml
master_slave:
  master:
    host: master.redis.com
    port: 6379
    password: password
  slaves:
    - host: slave1.redis.com
      port: 6379
      password: password
      weight: 1
    - host: slave2.redis.com
      port: 6379
      password: password
      weight: 2

read_policy: slave_first
load_balance_policy: weighted

pool_size: 20
min_idle_conns: 5

health_check:
  enable: true
  interval: 30s
  timeout: 5s
```

### 哨兵模式

```yaml
sentinel:
  master_name: mymaster
  sentinel_addrs:
    - sentinel1.redis.com:26379
    - sentinel2.redis.com:26379
    - sentinel3.redis.com:26379
  password: redis_password
  sentinel_password: sentinel_password
  db: 0
  slave_only: false

pool_size: 20
```

### 集群模式

```yaml
cluster:
  addrs:
    - node1.redis.com:7000
    - node2.redis.com:7001
    - node3.redis.com:7002
  password: password
  read_only: true
  route_randomly: true
  max_redirects: 3

pool_size: 30
```

### 加载配置文件

```go
cfg, err := cache.LoadFromFile("config.yaml")
if err != nil {
    panic(err)
}

r, err := cache.New(ctx, func(c *cache.Config) {
    *c = *cfg
})
```

---

## 最佳实践

### 1. 连接池配置

```go
cache.WithPoolSize(20),        // 根据并发量调整
cache.WithMinIdleConns(5),     // 保持一定空闲连接
cache.WithPoolTimeout(4*time.Second),
cache.WithIdleTimeout(5*time.Minute),
cache.WithMaxConnAge(1*time.Hour),
```

### 2. 超时配置

```go
cache.WithDialTimeout(5*time.Second),
cache.WithReadTimeout(3*time.Second),
cache.WithWriteTimeout(3*time.Second),
```

### 3. 读写分离

- 读多写少场景：`ReadPolicySlaveFirst` + `LoadBalanceWeighted`
- 主库压力大：`ReadPolicySlave`（从库故障会失败）
- 数据一致性要求高：`ReadPolicyMaster`

### 4. 高可用部署

- **开发环境**：单机模式
- **小规模生产**：主从模式 + 手动故障转移
- **中等规模生产**：哨兵模式（自动故障转移）
- **大规模生产**：集群模式（分片 + 自动故障转移）

---

## 常见问题

### 1. 主从同步延迟

**问题**：从库数据落后主库

**解决方案**：
- 使用 `ReadPolicyMasterFirst` 降低影响
- 关键数据显式读主库：`r.Master().Get(ctx, key)`

### 2. 集群 Hash Tag

**问题**：多键操作报错 `CROSSSLOT`

**解决方案**：
```go
// 使用 {} 确保键在同一槽
r.Set(ctx, "{user:1}:name", "Alice", 0)
r.Set(ctx, "{user:1}:age", "25", 0)
```

### 3. 健康检查频率

**建议**：
- 检查间隔：30s
- 超时时间：5s
- 失败阈值：3 次（配置中未暴露，可扩展）

---

## 性能优化

### 1. 使用 Pipeline 批量操作

```go
pipe := r.Pipeline()
for i := 0; i < 1000; i++ {
    pipe.Set(ctx, fmt.Sprintf("key:%d", i), i, 0)
}
pipe.Exec(ctx)
```

### 2. 合理配置连接池

- 根据并发量调整 `PoolSize`
- 保持适量 `MinIdleConns` 减少连接建立开销

### 3. 读写分离

- 将读请求分散到从库，降低主库压力
