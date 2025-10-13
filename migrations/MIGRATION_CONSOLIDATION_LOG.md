# Migrations 整理日志

## 整理时间
2025-10-13

## 整理目的
1. 统一管理所有数据库迁移文件，移动到根目录 `migrations/`
2. 消除无意义的补丁文件名（fix、old、new 等）
3. 开发阶段直接在功能模块 SQL 上修改，避免创建增量修改文件

## 整理前的目录结构

```
cmd/migrate/migrations/
├── 001_create_auth_tables.sql
├── 002_create_organization_tables.sql
├── 003_fix_sessions_token_column.sql        ❌ 补丁文件
├── 004_add_sessions_updated_at.sql          ❌ 补丁文件
├── 005_increase_sessions_token_length.sql   ❌ 补丁文件
├── 006_create_form_tables.sql
├── 007_create_notification_tables.sql
├── 008_create_approval_tables.sql
├── 009_create_file_tables.sql
└── 010_create_download_stats_table.sql

internal/auth/migrations/
└── (已清空)

internal/organization/migrations/
└── (已清空)
```

## 整理后的目录结构

```
migrations/
├── 001_create_auth_tables.sql               ✅ 已合并所有 sessions 表修复
├── 002_create_organization_tables.sql
├── 003_create_form_tables.sql
├── 004_create_notification_tables.sql
├── 005_create_approval_tables.sql
├── 006_create_file_tables.sql
├── 007_create_download_stats_table.sql
└── README.md                                ✅ 添加说明文档
```

## 合并的补丁文件详情

### sessions 表修复合并

将以下三个补丁文件的内容合并到 `001_create_auth_tables.sql`：

1. **003_fix_sessions_token_column.sql**
   - 修改：token_id VARCHAR(255) → token TEXT
   - 修改：refresh_token VARCHAR(255) → refresh_token TEXT

2. **004_add_sessions_updated_at.sql**
   - 添加：updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP

3. **005_increase_sessions_token_length.sql**
   - 已包含在第一个修复中（TEXT 类型）

### 合并后的 sessions 表结构

```sql
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    token TEXT NOT NULL,                    -- ✅ 直接使用 TEXT
    refresh_token TEXT,                      -- ✅ 直接使用 TEXT
    ip_address VARCHAR(50),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- ✅ 直接添加
    revoked_at TIMESTAMP
);
```

## 文件重新编号

| 原编号 | 新编号 | 文件名 | 说明 |
|--------|--------|--------|------|
| 001 | 001 | create_auth_tables.sql | 合并了 003/004/005 补丁 |
| 002 | 002 | create_organization_tables.sql | 保持不变 |
| ~~003~~ | - | ~~fix_sessions_token_column.sql~~ | ❌ 已合并到 001 |
| ~~004~~ | - | ~~add_sessions_updated_at.sql~~ | ❌ 已合并到 001 |
| ~~005~~ | - | ~~increase_sessions_token_length.sql~~ | ❌ 已合并到 001 |
| 006 | 003 | create_form_tables.sql | 重新编号 |
| 007 | 004 | create_notification_tables.sql | 重新编号 |
| 008 | 005 | create_approval_tables.sql | 重新编号 |
| 009 | 006 | create_file_tables.sql | 重新编号 |
| 010 | 007 | create_download_stats_table.sql | 重新编号 |

## 代码修改

### 1. pkg/migrate/migrate.go
**添加**: `LoadFromDir` 方法，支持从文件系统目录加载迁移文件

```go
// LoadFromDir 从文件系统目录加载迁移
func (m *Migrator) LoadFromDir(dir string) error {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return fmt.Errorf("read migration dir: %w", err)
    }

    for _, entry := range entries {
        if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
            continue
        }

        content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
        if err != nil {
            return fmt.Errorf("read migration file %s: %w", entry.Name(), err)
        }

        // 解析文件名
        parts := strings.SplitN(entry.Name(), "_", 2)
        if len(parts) < 2 {
            return fmt.Errorf("invalid migration filename: %s", entry.Name())
        }

        version := parts[0]
        description := strings.TrimSuffix(parts[1], ".sql")

        m.migrations = append(m.migrations, Migration{
            Version:     version,
            Description: description,
            SQL:         string(content),
        })
    }

    sort.Slice(m.migrations, func(i, j int) bool {
        return m.migrations[i].Version < m.migrations[j].Version
    })

    return nil
}
```

### 2. cmd/migrate/main.go
**修改**: 移除 `//go:embed` 指令，使用文件系统加载

**添加**: `getMigrationsDir()` 函数，自动定位 migrations 目录

```go
// getMigrationsDir 获取 migrations 目录路径
func getMigrationsDir() string {
    // 优先使用环境变量
    if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
        return dir
    }
    
    // 默认使用根目录下的 migrations
    execPath, err := os.Executable()
    if err != nil {
        log.Printf("Warning: failed to get executable path: %v", err)
        return "./migrations"
    }
    
    // 获取项目根目录（假设 bin/migrate 结构）
    rootDir := filepath.Dir(filepath.Dir(execPath))
    migrationsPath := filepath.Join(rootDir, "migrations")
    
    // 检查目录是否存在
    if _, err := os.Stat(migrationsPath); err == nil {
        return migrationsPath
    }
    
    // 开发环境兜底
    if _, err := os.Stat("./migrations"); err == nil {
        return "./migrations"
    }
    
    if _, err := os.Stat("../../migrations"); err == nil {
        return "../../migrations"
    }
    
    log.Fatalf("migrations directory not found. Please set MIGRATIONS_DIR environment variable")
    return ""
}
```

## 删除的文件和目录

- ❌ `cmd/migrate/migrations/003_fix_sessions_token_column.sql`
- ❌ `cmd/migrate/migrations/004_add_sessions_updated_at.sql`
- ❌ `cmd/migrate/migrations/005_increase_sessions_token_length.sql`
- ❌ `internal/auth/migrations/` 整个目录
- ❌ `internal/organization/migrations/` 整个目录
- ❌ `cmd/migrate/migrations/` 整个目录

## 新增的文件

- ✅ `migrations/README.md` - 使用说明文档
- ✅ `migrations/MIGRATION_CONSOLIDATION_LOG.md` - 本整理日志

## 开发规范

### ✅ 正确做法
开发阶段直接修改功能模块的 SQL 文件：
```bash
# 需要修改 sessions 表？直接编辑
vim migrations/001_create_auth_tables.sql
```

### ❌ 错误做法
创建补丁文件：
```bash
# ❌ 不要这样做
vim migrations/011_fix_sessions.sql
vim migrations/012_add_column.sql
vim migrations/013_update_index.sql
```

## 验证

### 编译验证
```bash
$ make migrate-build
✅ 编译成功
```

### 文件结构验证
```bash
$ ls -1 migrations/*.sql
001_create_auth_tables.sql
002_create_organization_tables.sql
003_create_form_tables.sql
004_create_notification_tables.sql
005_create_approval_tables.sql
006_create_file_tables.sql
007_create_download_stats_table.sql
✅ 7个文件，编号连续，无补丁文件
```

### sessions 表验证
```bash
$ grep -A 5 "token TEXT" migrations/001_create_auth_tables.sql
    token TEXT NOT NULL,                    -- Changed from token_id VARCHAR(255)
    refresh_token TEXT,                      -- Changed from VARCHAR(255) to TEXT
✅ token 和 refresh_token 已正确设置为 TEXT 类型
```

## 注意事项

1. **生产环境不适用**: 这种直接修改 SQL 文件的方式仅适用于开发阶段。生产环境必须创建新的迁移文件。

2. **数据库状态重置**: 如果已经执行过旧的迁移，需要重置数据库：
   ```bash
   make db-reset
   ```

3. **schema_migrations 表**: 迁移系统使用 `schema_migrations` 表追踪已执行的迁移，重置数据库会清空此表。

4. **环境变量**: 可以通过 `MIGRATIONS_DIR` 环境变量指定 migrations 目录位置：
   ```bash
   export MIGRATIONS_DIR=/path/to/migrations
   ./bin/migrate up
   ```

## 总结

✅ **已完成的工作**:
- 统一 migrations 目录到根目录
- 合并所有无意义的补丁文件
- 重新编号迁移文件（001-007）
- 添加文档说明
- 更新代码以支持文件系统加载
- 清理所有旧的 migrations 目录

✅ **达成的目标**:
- 目录结构清晰，易于管理
- 文件命名规范，无冗余补丁
- 开发流程简化，直接修改 SQL
- 代码可维护性提升

✅ **遵循的原则**:
- 开发阶段灵活修改
- 生产环境严格版本控制
- 文件命名语义化
- 目录结构扁平化
