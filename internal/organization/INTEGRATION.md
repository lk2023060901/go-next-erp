# Organization Module - 与数据库模块集成指南

## 概述

组织架构模块需要使用 `pkg/database` 封装的数据库接口（基于 pgx/v5），而不是直接使用 GORM。

## 数据库模块接口

### 核心方法

```go
import "github.com/lk2023060901/go-next-erp/pkg/database"

// 数据库连接
db, err := database.New(ctx,
    database.WithHost("localhost"),
    database.WithPort(5432),
    database.WithDatabase("erp"),
    database.WithUsername("postgres"),
    database.WithPassword("password"),
)

// 查询（读操作，主从模式自动路由到从库）
rows, err := db.Query(ctx, "SELECT * FROM organizations WHERE tenant_id = $1", tenantID)

// 查询单行
row := db.QueryRow(ctx, "SELECT * FROM organizations WHERE id = $1", id)

// 执行命令（写操作，强制路由到主库）
tag, err := db.Exec(ctx, "INSERT INTO organizations (...) VALUES (...)", ...)

// 事务
err = db.Transaction(ctx, func(tx pgx.Tx) error {
    // 在事务中执行多个操作
    _, err := tx.Exec(ctx, "INSERT ...")
    return err
})
```

## Repository 层改造方案

### 方案 1：直接使用 pgx（推荐）

```go
package repository

import (
    "context"
    "github.com/google/uuid"
    "github.com/lk2023060901/go-next-erp/pkg/database"
    "github.com/lk2023060901/go-next-erp/internal/organization/model"
)

type OrganizationRepository struct {
    db *database.DB
}

func NewOrganizationRepository(db *database.DB) *OrganizationRepository {
    return &OrganizationRepository{db: db}
}

// GetByID 根据 ID 获取组织
func (r *OrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
    sql := `
        SELECT id, tenant_id, code, name, short_name, description,
               type_id, type_code, parent_id, level, path, path_names,
               ancestor_ids, is_leaf, leader_id, leader_name,
               legal_person, unified_code, register_date, register_addr,
               phone, email, address, employee_count, direct_emp_count,
               sort, status, tags, created_by, updated_by,
               created_at, updated_at, deleted_at
        FROM organizations
        WHERE id = $1 AND deleted_at IS NULL
    `

    org := &model.Organization{}
    err := r.db.QueryRow(ctx, sql, id).Scan(
        &org.ID, &org.TenantID, &org.Code, &org.Name, &org.ShortName, &org.Description,
        &org.TypeID, &org.TypeCode, &org.ParentID, &org.Level, &org.Path, &org.PathNames,
        &org.AncestorIDs, &org.IsLeaf, &org.LeaderID, &org.LeaderName,
        &org.LegalPerson, &org.UnifiedCode, &org.RegisterDate, &org.RegisterAddr,
        &org.Phone, &org.Email, &org.Address, &org.EmployeeCount, &org.DirectEmpCount,
        &org.Sort, &org.Status, &org.Tags, &org.CreatedBy, &org.UpdatedBy,
        &org.CreatedAt, &org.UpdatedAt, &org.DeletedAt,
    )

    if err != nil {
        return nil, err
    }

    return org, nil
}

// Create 创建组织
func (r *OrganizationRepository) Create(ctx context.Context, org *model.Organization) error {
    sql := `
        INSERT INTO organizations (
            id, tenant_id, code, name, short_name, description,
            type_id, type_code, parent_id, level, path, path_names,
            ancestor_ids, is_leaf, leader_id, leader_name,
            phone, email, address, sort, status, tags,
            created_by, updated_by, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6,
            $7, $8, $9, $10, $11, $12,
            $13, $14, $15, $16,
            $17, $18, $19, $20, $21, $22,
            $23, $24, $25, $26
        )
    `

    _, err := r.db.Exec(ctx, sql,
        org.ID, org.TenantID, org.Code, org.Name, org.ShortName, org.Description,
        org.TypeID, org.TypeCode, org.ParentID, org.Level, org.Path, org.PathNames,
        org.AncestorIDs, org.IsLeaf, org.LeaderID, org.LeaderName,
        org.Phone, org.Email, org.Address, org.Sort, org.Status, org.Tags,
        org.CreatedBy, org.UpdatedBy, org.CreatedAt, org.UpdatedAt,
    )

    return err
}

// List 列出组织
func (r *OrganizationRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*model.Organization, error) {
    sql := `
        SELECT id, tenant_id, code, name, type_code, parent_id, level,
               path, path_names, is_leaf, leader_name, status,
               employee_count, created_at
        FROM organizations
        WHERE tenant_id = $1 AND deleted_at IS NULL
        ORDER BY level ASC, sort ASC
        LIMIT $2 OFFSET $3
    `

    rows, err := r.db.Query(ctx, sql, tenantID, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orgs []*model.Organization
    for rows.Next() {
        org := &model.Organization{}
        err := rows.Scan(
            &org.ID, &org.TenantID, &org.Code, &org.Name, &org.TypeCode,
            &org.ParentID, &org.Level, &org.Path, &org.PathNames,
            &org.IsLeaf, &org.LeaderName, &org.Status,
            &org.EmployeeCount, &org.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        orgs = append(orgs, org)
    }

    if err = rows.Err(); err != nil {
        return nil, err
    }

    return orgs, nil
}

// Update 更新组织
func (r *OrganizationRepository) Update(ctx context.Context, org *model.Organization) error {
    sql := `
        UPDATE organizations SET
            name = $1, short_name = $2, description = $3,
            leader_id = $4, leader_name = $5,
            phone = $6, email = $7, address = $8,
            status = $9, tags = $10,
            updated_by = $11, updated_at = $12
        WHERE id = $13 AND tenant_id = $14
    `

    _, err := r.db.Exec(ctx, sql,
        org.Name, org.ShortName, org.Description,
        org.LeaderID, org.LeaderName,
        org.Phone, org.Email, org.Address,
        org.Status, org.Tags,
        org.UpdatedBy, time.Now(),
        org.ID, org.TenantID,
    )

    return err
}

// Delete 软删除
func (r *OrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
    sql := `UPDATE organizations SET deleted_at = $1 WHERE id = $2`
    _, err := r.db.Exec(ctx, sql, time.Now(), id)
    return err
}
```

### 方案 2：使用 sqlx（可选，如果需要 ORM 特性）

如果需要 ORM 特性，可以引入 `sqlx` 库包装 pgx：

```go
import (
    "github.com/jmoiron/sqlx"
    "github.com/jackc/pgx/v5/stdlib"
)

// 将 pgx 连接池转换为 sqlx.DB
pool := db.Pool() // 获取 *pgxpool.Pool
sqlxDB := sqlx.NewDb(stdlib.OpenDBFromPool(pool), "pgx")

// 使用 sqlx 查询
org := &model.Organization{}
err := sqlxDB.Get(org, "SELECT * FROM organizations WHERE id = $1", id)

// 批量查询
var orgs []model.Organization
err := sqlxDB.Select(&orgs, "SELECT * FROM organizations WHERE tenant_id = $1", tenantID)
```

## 事务处理

### 单表操作事务

```go
func (r *OrganizationRepository) CreateWithClosure(ctx context.Context, org *model.Organization) error {
    return r.db.Transaction(ctx, func(tx pgx.Tx) error {
        // 1. 插入组织
        _, err := tx.Exec(ctx, `INSERT INTO organizations (...) VALUES (...)`, ...)
        if err != nil {
            return err
        }

        // 2. 插入闭包表
        _, err = tx.Exec(ctx, `INSERT INTO organization_closures (...) VALUES (...)`, ...)
        if err != nil {
            return err
        }

        return nil
    })
}
```

### 跨 Repository 事务

```go
// service 层
func (s *OrganizationService) CreateOrganization(ctx context.Context, org *model.Organization) error {
    return s.db.Transaction(ctx, func(tx pgx.Tx) error {
        // 创建事务级别的 repository
        orgRepo := repository.NewOrganizationTxRepo(tx)
        closureRepo := repository.NewClosureTxRepo(tx)

        // 1. 创建组织
        if err := orgRepo.Create(ctx, org); err != nil {
            return err
        }

        // 2. 创建闭包关系
        if err := closureRepo.InsertClosure(ctx, org); err != nil {
            return err
        }

        return nil
    })
}

// 支持事务的 Repository
type OrganizationTxRepo struct {
    tx pgx.Tx
}

func NewOrganizationTxRepo(tx pgx.Tx) *OrganizationTxRepo {
    return &OrganizationTxRepo{tx: tx}
}

func (r *OrganizationTxRepo) Create(ctx context.Context, org *model.Organization) error {
    _, err := r.tx.Exec(ctx, `INSERT INTO organizations (...) VALUES (...)`, ...)
    return err
}
```

## 主从模式支持

### 自动路由

```go
// 读操作：自动路由到从库
orgs, err := r.db.Query(ctx, "SELECT * FROM organizations WHERE ...")

// 写操作：自动路由到主库
_, err := r.db.Exec(ctx, "INSERT INTO organizations (...) VALUES (...)", ...)
```

### 显式指定

```go
// 强制从主库读取（确保一致性）
orgs, err := r.db.Master().Query(ctx, "SELECT * FROM organizations WHERE ...")

// 强制从从库读取（减轻主库压力）
orgs, err := r.db.Slave().Query(ctx, "SELECT * FROM organizations WHERE ...")
```

## 性能优化

### 1. 批量插入

```go
import "github.com/jackc/pgx/v5"

func (r *OrganizationRepository) BatchCreate(ctx context.Context, orgs []*model.Organization) error {
    batch := &pgx.Batch{}

    for _, org := range orgs {
        batch.Queue(`INSERT INTO organizations (...) VALUES (...)`, ...)
    }

    br := r.db.SendBatch(ctx, batch)
    defer br.Close()

    for range orgs {
        if _, err := br.Exec(); err != nil {
            return err
        }
    }

    return nil
}
```

### 2. 预编译语句（Prepared Statements）

pgx 会自动缓存预编译语句，无需手动管理。

### 3. 连接池优化

```go
db, err := database.New(ctx,
    database.WithMaxConns(50),
    database.WithMinConns(10),
    database.WithMaxConnLifetime(time.Hour),
    database.WithMaxConnIdleTime(30*time.Minute),
)
```

## 迁移建议

### 阶段 1：移除 GORM 依赖

1. 移除所有 `gorm:` 标签
2. 移除 `TableName()` 方法（直接在 SQL 中指定表名）
3. 移除 GORM 特有的类型（如 `gorm.DeletedAt`）

### 阶段 2：实现 Repository

1. 创建基于 pgx 的 Repository 接口
2. 实现 CRUD 方法
3. 实现树形结构查询方法

### 阶段 3：集成测试

1. 编写 Repository 单元测试
2. 编写 Service 集成测试
3. 编写 API 集成测试

## 完整示例

参考项目中的其他模块实现（如果有），或者按照上述模式实现：

1. `internal/organization/repository/organization_repo.go`
2. `internal/organization/service/organization_service.go`
3. `internal/organization/handler/organization_handler.go`

## 下一步

1. 完成所有模型的 GORM 标签移除
2. 实现 Repository 层（使用 pgx）
3. 实现 Service 层（业务逻辑）
4. 实现 Handler 层（HTTP API）
