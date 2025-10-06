# 4A 权限系统

完整的企业级 4A 权限系统实现（Authentication, Authorization, Accounting, Audit）

## 特性

✅ **多租户隔离** - 完整的租户数据隔离
✅ **三种授权模型** - RBAC + ABAC + ReBAC
✅ **会话管理** - JWT + 数据库会话双重管理
✅ **审计日志** - 完整的操作记录和合规性支持
✅ **密码安全** - Argon2 哈希 + 强度验证
✅ **高性能** - 表达式缓存 + Redis 缓存
✅ **主从分离** - 自动读写分离路由

## 快速开始

### 1. 运行数据库迁移

```bash
# 创建表结构
psql -U postgres -d go_next_erp -f internal/auth/migrations/001_create_tables.sql

# 导入初始数据
psql -U postgres -d go_next_erp -f internal/auth/migrations/002_seed_data.sql
```

### 2. 运行测试

```bash
# 认证模块测试
go test -v ./internal/auth/authentication/...

# 授权模块测试
go test -v ./internal/auth/authorization/...

# 运行所有测试
go test -v ./internal/auth/...
```

### 3. 使用示例

#### 初始化服务

```go
import (
    "github.com/lk2023060901/go-next-erp/internal/auth/authentication"
    "github.com/lk2023060901/go-next-erp/internal/auth/authorization"
    "github.com/lk2023060901/go-next-erp/internal/auth/repository"
    "github.com/lk2023060901/go-next-erp/pkg/database"
    "github.com/lk2023060901/go-next-erp/pkg/cache"
)

// 初始化数据库和缓存
db, _ := database.New(ctx)
cacheClient, _ := cache.New(ctx)

// 初始化 Repository
userRepo := repository.NewUserRepository(db, cacheClient)
roleRepo := repository.NewRoleRepository(db, cacheClient)
permissionRepo := repository.NewPermissionRepository(db, cacheClient)
// ... 其他 repo

// 初始化认证服务
authService := authentication.NewService(
    userRepo, sessionRepo, auditRepo,
    &jwt.Config{
        SecretKey:       "your-secret-key",
        AccessTokenTTL:  24 * time.Hour,
        RefreshTokenTTL: 7 * 24 * time.Hour,
        Issuer:          "go-next-erp",
    },
)

// 初始化授权服务
authzService := authorization.NewService(
    roleRepo, permissionRepo, policyRepo,
    userRepo, relationRepo, auditRepo, cacheClient,
)
```

#### 用户注册和登录

```go
// 注册
user, err := authService.Register(ctx, "john_doe", "john@example.com", "SecurePass@123", tenantID)

// 登录
loginResp, err := authService.Login(ctx, &authentication.LoginRequest{
    Username:  "john_doe",
    Password:  "SecurePass@123",
    IPAddress: "192.168.1.1",
    UserAgent: "Mozilla/5.0",
})

fmt.Printf("Access Token: %s\n", loginResp.AccessToken)
```

#### RBAC 授权

```go
// 创建角色
role := &model.Role{
    Name:        "editor",
    DisplayName: "编辑者",
    TenantID:    tenantID,
}
roleRepo.Create(ctx, role)

// 创建权限
perm := &model.Permission{
    Resource:    "document",
    Action:      "update",
    DisplayName: "编辑文档",
    TenantID:    tenantID,
}
permissionRepo.Create(ctx, perm)

// 分配权限给角色
permissionRepo.AssignPermissionToRole(ctx, role.ID, perm.ID, tenantID)

// 分配角色给用户
roleRepo.AssignRoleToUser(ctx, userID, role.ID, tenantID)

// 检查权限
allowed, _ := authzService.CheckPermission(ctx, userID, tenantID, "document", "update", nil)
```

#### ABAC 策略

```go
// 创建策略
policy := &model.Policy{
    Name:        "same_dept_read",
    Description: "同部门用户可读文档",
    TenantID:    tenantID,
    Resource:    "document",
    Action:      "read",
    Expression:  "User.DepartmentID == Resource.DepartmentID",
    Effect:      model.PolicyEffectAllow,
    Priority:    100,
    Enabled:     true,
}
policyRepo.Create(ctx, policy)

// 检查权限（带资源属性）
resourceAttrs := map[string]interface{}{
    "DepartmentID": "dept-001",
    "OwnerID":      userID.String(),
}

allowed, _ := authzService.CheckPermission(ctx, userID, tenantID, "document", "read", resourceAttrs)
```

#### ReBAC 关系型授权

```go
// 建立关系
subject := "user:" + userID.String()
object := "document:" + docID

authzService.GrantRelation(ctx, tenantID, subject, "owner", object)

// 检查关系
resourceAttrs := map[string]interface{}{"ID": docID}
allowed, _ := authzService.CheckPermission(ctx, userID, tenantID, "document", "owner", resourceAttrs)
```

## 架构设计

### 模块结构

```
internal/auth/
├── model/              # 数据模型
│   ├── user.go
│   ├── role.go
│   ├── permission.go
│   ├── policy.go
│   ├── tenant.go
│   ├── session.go
│   ├── audit.go
│   └── relation.go
├── repository/         # 数据访问层
│   ├── user_repo.go
│   ├── role_repo.go
│   ├── permission_repo.go
│   ├── policy_repo.go
│   ├── tenant_repo.go
│   ├── session_repo.go
│   ├── audit_repo.go
│   └── relation_repo.go
├── authentication/     # 认证模块
│   ├── password/       # 密码认证
│   ├── jwt/            # JWT 管理
│   └── service.go      # 认证服务
├── authorization/      # 授权模块
│   ├── rbac/           # RBAC 引擎
│   ├── abac/           # ABAC 引擎
│   ├── rebac/          # ReBAC 引擎
│   └── service.go      # 授权服务
└── migrations/         # 数据库迁移
    ├── 001_create_tables.sql
    └── 002_seed_data.sql
```

### 数据流

```
┌─────────────┐
│  HTTP 请求  │
└──────┬──────┘
       │
┌──────▼──────┐
│  认证中间件  │ ← JWT 验证
└──────┬──────┘
       │
┌──────▼──────┐
│  授权检查    │ ← RBAC/ABAC/ReBAC
└──────┬──────┘
       │
┌──────▼──────┐
│  业务逻辑    │
└──────┬──────┘
       │
┌──────▼──────┐
│  审计日志    │
└─────────────┘
```

## ABAC 表达式示例

支持基于 Expr 的强大表达式语法：

```javascript
// 同部门访问
User.DepartmentID == Resource.DepartmentID

// 级别检查
User.Level >= 3

// 时间限制
Time.Hour >= 9 && Time.Hour <= 18

// 组合条件
User.DepartmentID == Resource.DepartmentID &&
User.Level >= 3 &&
Time.Weekday >= 1 && Time.Weekday <= 5

// 所有者或管理员
User.ID == Resource.OwnerID || "admin" in User.Roles
```

更多示例见：[ABAC 表达式示例](authorization/abac/EXAMPLES.md)

## 数据库设计

### 核心表

- `tenants` - 租户表
- `users` - 用户表
- `roles` - 角色表（支持继承）
- `permissions` - 权限表
- `user_roles` - 用户-角色关联
- `role_permissions` - 角色-权限关联
- `policies` - ABAC 策略表
- `relation_tuples` - ReBAC 关系表
- `sessions` - 会话表
- `audit_logs` - 审计日志表

### 索引优化

所有表都有完善的索引设计：
- 主键使用 UUID v7（时间有序）
- 外键索引
- 查询字段索引
- 软删除索引

## 性能优化

### 缓存策略

1. **用户权限缓存** - 10分钟 TTL
2. **角色权限缓存** - 10分钟 TTL
3. **策略缓存** - 5分钟 TTL
4. **会话缓存** - 到过期时间
5. **ABAC 表达式编译缓存** - 永久（内存）

### 主从分离

- **读操作** - 自动路由到从库
- **写操作** - 强制路由到主库
- **事务** - 强制使用主库

## 测试

### 单元测试

```bash
# 认证模块
go test -v ./internal/auth/authentication/... -run TestAll_Success

# 授权模块
go test -v ./internal/auth/authorization/... -run TestAll_Success
```

### 集成测试

```bash
# 需要先启动数据库和 Redis
docker-compose up -d postgres redis

# 运行集成测试
go test -v ./internal/auth/... -tags=integration
```

## 默认账号

初始化数据后，系统会创建默认管理员账号：

- **用户名**: `admin`
- **密码**: 需要在应用中设置（未预设）
- **角色**: 系统管理员（拥有所有权限）

## 依赖

```go
require (
    github.com/google/uuid v1.6.0
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/expr-lang/expr v1.16.0
    github.com/stretchr/testify v1.8.4
    golang.org/x/crypto v0.17.0
)
```

## 许可证

MIT License
