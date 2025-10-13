# Database Migrations

此目录包含项目的所有数据库迁移文件。

## 文件命名规范

格式：`{序号}_{功能描述}.sql`

例如：
- `001_create_auth_tables.sql` - 创建认证授权相关表
- `002_create_organization_tables.sql` - 创建组织架构相关表
- `003_create_form_tables.sql` - 创建表单相关表

## 开发阶段规范

在开发阶段，**直接在具体的功能模块 SQL Schema 文件上进行修改**，无需创建 fix、add、update 等增量修改文件。

### ✅ 正确做法
直接编辑 `001_create_auth_tables.sql`，修改表结构

### ❌ 错误做法
创建 `008_fix_sessions_token.sql`、`009_add_column.sql` 等补丁文件

## 迁移文件列表

| 序号 | 文件名 | 说明 |
|------|--------|------|
| 001 | create_auth_tables.sql | 认证授权系统（4A 权限系统）|
| 002 | create_organization_tables.sql | 组织架构管理 |
| 003 | create_form_tables.sql | 动态表单系统 |
| 004 | create_notification_tables.sql | 通知系统 |
| 005 | create_approval_tables.sql | 审批流程系统 |
| 006 | create_file_tables.sql | 文件管理系统 |
| 007 | create_download_stats_table.sql | 文件下载统计 |

## 执行迁移

```bash
# 执行所有迁移
make migrate-up

# 回滚迁移
make migrate-down

# 查看迁移状态
make migrate-status
```

## 注意事项

1. **按序号执行** - 迁移文件必须按序号顺序执行
2. **幂等性** - 所有 SQL 语句应使用 `IF NOT EXISTS` 确保可重复执行
3. **外键约束** - 注意表之间的依赖关系，先创建被引用的表
4. **索引优化** - 为常用查询字段创建适当的索引
5. **默认数据** - 初始化数据使用 `ON CONFLICT DO NOTHING` 避免重复插入
