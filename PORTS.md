# 服务端口说明

## 容器服务端口映射（宿主机 → 容器）

所有容器服务使用连续端口 **15000-15005**，避免与现有服务冲突。

| 服务 | 容器名称 | 宿主机端口 | 容器端口 | 说明 |
|------|---------|-----------|---------|------|
| PostgreSQL | erp-postgres | 15000 | 5432 | 主数据库 |
| Redis | erp-redis | 15001 | 6379 | 缓存服务 |
| MinIO API | erp-minio | 15002 | 9000 | 对象存储 API |
| MinIO Console | erp-minio | 15003 | 9001 | MinIO 控制台 |
| Milvus API | erp-milvus | 15004 | 19530 | 向量数据库 gRPC |
| Milvus Metrics | erp-milvus | 15005 | 9091 | Milvus 监控指标 |

## 服务访问地址

### 1. PostgreSQL 数据库
```bash
# 连接字符串
postgresql://postgres:postgres123@localhost:15000/erp

# psql 命令行
psql -h localhost -p 15000 -U postgres -d erp

# 或使用 make 命令
make docker-exec-postgres
```

**环境变量配置**：
```env
DB_HOST=localhost
DB_PORT=15000
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=erp
```

### 2. Redis 缓存
```bash
# redis-cli 连接
redis-cli -h localhost -p 15001 -a redis123

# 或使用 make 命令
make docker-exec-redis
```

**环境变量配置**：
```env
REDIS_HOST=localhost
REDIS_PORT=15001
REDIS_PASSWORD=redis123
```

### 3. MinIO 对象存储
- **API 地址**: http://localhost:15002
- **控制台**: http://localhost:15003
- **用户名**: minioadmin
- **密码**: minioadmin123

**环境变量配置**：
```env
MINIO_ENDPOINT=localhost:15002
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_USE_SSL=false
```

### 4. Milvus 向量数据库
- **gRPC 地址**: localhost:15004
- **Metrics 地址**: http://localhost:15005/metrics

**环境变量配置**：
```env
MILVUS_HOST=localhost
MILVUS_PORT=15004
```

## Go 应用端口

**开发环境**：应用在宿主机运行，**不使用容器部署**

- 默认端口 8080 已被占用，建议使用其他端口（如 8081）
- 配置方式：
  ```env
  APP_PORT=8081
  ```

## Docker 命令

### 启动所有服务
```bash
make docker-up
# 或
docker-compose up -d
```

### 停止所有服务
```bash
make docker-down
# 或
docker-compose down
```

### 查看服务状态
```bash
make docker-ps
# 或
docker-compose ps
```

### 查看日志
```bash
# 所有服务日志
make docker-logs

# PostgreSQL 日志
make docker-logs-postgres

# Redis 日志
make docker-logs-redis

# MinIO 日志
make docker-logs-minio

# Milvus 日志
make docker-logs-milvus
```

### 进入容器
```bash
# PostgreSQL
make docker-exec-postgres

# Redis
make docker-exec-redis

# MinIO
make docker-exec-minio
```

## 开发配置示例

### config/app.yml
```yaml
server:
  host: 0.0.0.0
  port: 8081

database:
  host: localhost
  port: 15000
  user: postgres
  password: postgres123
  database: erp
  max_conns: 20
  min_conns: 5

redis:
  host: localhost
  port: 15001
  password: redis123
  db: 0

minio:
  endpoint: localhost:15002
  access_key: minioadmin
  secret_key: minioadmin123
  use_ssl: false

milvus:
  host: localhost
  port: 15004
```

### .env 文件
```env
# 应用配置
APP_ENV=development
APP_PORT=8081

# 数据库配置
DB_HOST=localhost
DB_PORT=15000
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=erp

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=15001
REDIS_PASSWORD=redis123

# MinIO 配置
MINIO_ENDPOINT=localhost:15002
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123

# Milvus 配置
MILVUS_HOST=localhost
MILVUS_PORT=15004
```

## 注意事项

1. **端口冲突检查**：
   ```bash
   # 检查端口占用
   lsof -i :15000
   ```

2. **容器网络**：
   - 所有容器在 `erp-network` 网络中
   - 容器间通信使用服务名（如 `erp-postgres:5432`）

3. **数据持久化**：
   - 所有数据存储在 Docker volumes 中
   - 删除容器不会丢失数据
   - 清理数据需要手动删除 volumes：
     ```bash
     make clean-docker
     ```

4. **健康检查**：
   - 所有服务都配置了健康检查
   - 启动时会自动等待依赖服务就绪
