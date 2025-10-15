# HRM 加班服务快速启动指南

**版本**: v1.0  
**更新时间**: 2025-10-14  

---

## 🚀 快速启动

### 1. 启动基础设施

```bash
# 启动 PostgreSQL, Redis, MinIO 等服务
docker-compose up -d

# 检查服务状态
make docker-ps
```

### 2. 运行数据库迁移

```bash
# 构建迁移工具
make migrate-build

# 执行迁移
make migrate-up

# 检查迁移状态
make migrate-status
```

### 3. 启动应用服务

```bash
# 开发模式（热重载）
make dev

# 或者构建后运行
make build
./bin/server -conf ./configs/config.yaml
```

服务启动后：
- **HTTP 服务**: `http://localhost:8000`
- **gRPC 服务**: `localhost:9000`

---

## 📡 API 调用示例

### 使用 HTTP API

#### 1. 创建加班申请

```bash
curl -X POST http://localhost:8000/api/v1/hrm/overtimes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "employee_id": "660e8400-e29b-41d4-a716-446655440001",
    "employee_name": "张三",
    "department_id": "770e8400-e29b-41d4-a716-446655440002",
    "start_time": "2024-01-15T18:00:00Z",
    "end_time": "2024-01-15T21:00:00Z",
    "duration": 3.0,
    "overtime_type": "workday",
    "pay_type": "money",
    "reason": "项目紧急上线",
    "tasks": ["完成用户模块", "修复紧急bug"],
    "remark": "需要技术总监审批"
  }'
```

**响应示例**:
```json
{
  "id": "880e8400-e29b-41d4-a716-446655440003",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "employee_id": "660e8400-e29b-41d4-a716-446655440001",
  "employee_name": "张三",
  "department_id": "770e8400-e29b-41d4-a716-446655440002",
  "start_time": "2024-01-15T18:00:00Z",
  "end_time": "2024-01-15T21:00:00Z",
  "duration": 3.0,
  "overtime_type": "workday",
  "pay_type": "money",
  "pay_rate": 1.5,
  "reason": "项目紧急上线",
  "tasks": ["完成用户模块", "修复紧急bug"],
  "approval_status": "pending",
  "remark": "需要技术总监审批",
  "created_at": "2024-01-15T15:30:00Z",
  "updated_at": "2024-01-15T15:30:00Z"
}
```

---

#### 2. 查询加班记录列表

```bash
curl -X GET "http://localhost:8000/api/v1/hrm/overtimes?tenant_id=550e8400-e29b-41d4-a716-446655440000&page=1&page_size=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**查询参数**:
- `tenant_id`: 租户ID（必填）
- `employee_id`: 员工ID（可选）
- `department_id`: 部门ID（可选）
- `overtime_type`: 加班类型（可选，workday/weekend/holiday）
- `approval_status`: 审批状态（可选，pending/approved/rejected）
- `start_date`: 开始日期（可选）
- `end_date`: 结束日期（可选）
- `keyword`: 关键词搜索（可选）
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认10）

**响应示例**:
```json
{
  "items": [
    {
      "id": "880e8400-e29b-41d4-a716-446655440003",
      "employee_name": "张三",
      "duration": 3.0,
      "overtime_type": "workday",
      "approval_status": "pending",
      "created_at": "2024-01-15T15:30:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10
}
```

---

#### 3. 获取加班详情

```bash
curl -X GET "http://localhost:8000/api/v1/hrm/overtimes/880e8400-e29b-41d4-a716-446655440003" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

#### 4. 更新加班申请

```bash
curl -X PUT http://localhost:8000/api/v1/hrm/overtimes/880e8400-e29b-41d4-a716-446655440003 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "duration": 4.0,
    "reason": "项目紧急上线，需要延长加班时间"
  }'
```

---

#### 5. 查询员工加班记录

```bash
curl -X GET "http://localhost:8000/api/v1/hrm/employees/660e8400-e29b-41d4-a716-446655440001/overtimes?tenant_id=550e8400-e29b-41d4-a716-446655440000&year=2024" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

#### 6. 查询待审批的加班

```bash
curl -X GET "http://localhost:8000/api/v1/hrm/overtimes/pending?tenant_id=550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

#### 7. 提交加班审批

```bash
curl -X POST http://localhost:8000/api/v1/hrm/overtimes/880e8400-e29b-41d4-a716-446655440003/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "submitter_id": "990e8400-e29b-41d4-a716-446655440004"
  }'
```

**响应示例**:
```json
{
  "success": true,
  "message": "加班申请已提交审批"
}
```

---

#### 8. 批准加班

```bash
curl -X POST http://localhost:8000/api/v1/hrm/overtimes/880e8400-e29b-41d4-a716-446655440003/approve \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "approver_id": "aa0e8400-e29b-41d4-a716-446655440005"
  }'
```

**响应示例**:
```json
{
  "success": true,
  "message": "加班申请已批准"
}
```

---

#### 9. 拒绝加班

```bash
curl -X POST http://localhost:8000/api/v1/hrm/overtimes/880e8400-e29b-41d4-a716-446655440003/reject \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "approver_id": "aa0e8400-e29b-41d4-a716-446655440005",
    "reason": "加班时间过长，建议分两天完成"
  }'
```

**响应示例**:
```json
{
  "success": true,
  "message": "加班申请已拒绝"
}
```

---

#### 10. 统计加班时长

```bash
curl -X GET "http://localhost:8000/api/v1/hrm/employees/660e8400-e29b-41d4-a716-446655440001/overtime-hours?tenant_id=550e8400-e29b-41d4-a716-446655440000&start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**响应示例**:
```json
{
  "total_hours": 24.5
}
```

---

#### 11. 查询可调休天数

```bash
curl -X GET "http://localhost:8000/api/v1/hrm/employees/660e8400-e29b-41d4-a716-446655440001/comp-off-days?tenant_id=550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**响应示例**:
```json
{
  "available_days": 3.5
}
```

---

#### 12. 使用调休

```bash
curl -X POST http://localhost:8000/api/v1/hrm/employees/660e8400-e29b-41d4-a716-446655440001/comp-off-days/use \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "days": 1.0
  }'
```

**响应示例**:
```json
{
  "success": true,
  "message": "调休使用成功",
  "remaining_days": 2.5
}
```

---

#### 13. 删除加班申请

```bash
curl -X DELETE "http://localhost:8000/api/v1/hrm/overtimes/880e8400-e29b-41d4-a716-446655440003" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**响应示例**:
```json
{
  "success": true,
  "message": "Overtime deleted successfully"
}
```

---

### 使用 gRPC API

#### 安装 grpcurl

```bash
# macOS
brew install grpcurl

# Linux
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

#### 列出所有服务

```bash
grpcurl -plaintext localhost:9000 list
```

#### 列出加班服务的方法

```bash
grpcurl -plaintext localhost:9000 list api.hrm.v1.OvertimeService
```

#### 创建加班申请（gRPC）

```bash
grpcurl -plaintext \
  -d '{
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "employee_id": "660e8400-e29b-41d4-a716-446655440001",
    "employee_name": "张三",
    "department_id": "770e8400-e29b-41d4-a716-446655440002",
    "start_time": "2024-01-15T18:00:00Z",
    "end_time": "2024-01-15T21:00:00Z",
    "duration": 3.0,
    "overtime_type": "workday",
    "pay_type": "money",
    "reason": "项目紧急上线"
  }' \
  localhost:9000 \
  api.hrm.v1.OvertimeService/CreateOvertime
```

#### 查询加班列表（gRPC）

```bash
grpcurl -plaintext \
  -d '{
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "page": 1,
    "page_size": 10
  }' \
  localhost:9000 \
  api.hrm.v1.OvertimeService/ListOvertimes
```

---

## 🔐 认证说明

所有 API（除了登录、注册等公开接口）都需要 JWT Token 认证。

### 获取 Token

```bash
# 登录获取 Token
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your_password"
  }'
```

**响应**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

### 使用 Token

在请求头中添加：
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## 📊 加班类型说明

### overtime_type（加班类型）

| 值 | 说明 | 默认倍率 |
|----|------|---------|
| `workday` | 工作日加班 | 1.5x |
| `weekend` | 周末加班 | 2.0x |
| `holiday` | 法定节假日加班 | 3.0x |

### pay_type（补偿类型）

| 值 | 说明 |
|----|------|
| `money` | 加班费 |
| `leave` | 调休 |

### approval_status（审批状态）

| 值 | 说明 |
|----|------|
| `pending` | 待审批 |
| `approved` | 已批准 |
| `rejected` | 已拒绝 |

---

## 🧪 测试数据

### 测试租户
```json
{
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "tenant_name": "测试公司"
}
```

### 测试员工
```json
{
  "employee_id": "660e8400-e29b-41d4-a716-446655440001",
  "employee_name": "张三",
  "department_id": "770e8400-e29b-41d4-a716-446655440002"
}
```

---

## 🐛 常见问题

### 1. 连接被拒绝
**问题**: `connection refused`  
**解决**: 确保服务已启动，检查端口是否正确

```bash
# 检查服务状态
lsof -i :8000  # HTTP
lsof -i :9000  # gRPC
```

### 2. 认证失败
**问题**: `401 Unauthorized`  
**解决**: 检查 JWT Token 是否过期，重新登录获取新 Token

### 3. 无效的 UUID
**问题**: `invalid UUID`  
**解决**: 确保所有 ID 参数都是有效的 UUID 格式

```bash
# 有效的 UUID 格式
550e8400-e29b-41d4-a716-446655440000

# 无效的 UUID 格式
123456
abc-def-ghi
```

### 4. 数据库连接失败
**问题**: `failed to connect to database`  
**解决**: 确保 PostgreSQL 服务已启动

```bash
# 启动数据库
docker-compose up -d erp-postgres

# 检查连接
psql -h localhost -p 15000 -U postgres -d erp
```

---

## 📝 开发建议

### 1. 使用 Postman 集合

创建 Postman Collection 方便测试：

```json
{
  "info": {
    "name": "HRM Overtime API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Create Overtime",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Authorization",
            "value": "Bearer {{token}}"
          }
        ],
        "url": "{{baseUrl}}/api/v1/hrm/overtimes"
      }
    }
  ]
}
```

### 2. 环境变量配置

```bash
# .env
BASE_URL=http://localhost:8000
GRPC_URL=localhost:9000
JWT_TOKEN=your_token_here
TENANT_ID=550e8400-e29b-41d4-a716-446655440000
```

### 3. 日志查看

```bash
# 查看应用日志
tail -f logs/app.log

# 查看错误日志
tail -f logs/error.log

# Docker 日志
docker-compose logs -f erp-server
```

---

## 🔗 相关文档

- [HRM 加班模块迁移完成报告](./hrm_overtime_kratos_migration_complete.md)
- [加班模块测试总结报告](../test_reports/overtime_module_test_summary.md)
- [优化完成报告](../optimization/overtime_optimization_complete.md)
- [Kratos 官方文档](https://go-kratos.dev/)
- [Protocol Buffers 文档](https://protobuf.dev/)

---

**文档维护**: AI Assistant  
**最后更新**: 2025-10-14  
**版本**: v1.0  
