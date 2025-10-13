# Organization - 组织架构模块

## 概述

组织架构模块提供完整的企业组织管理功能，支持多租户、树形结构、灵活的组织类型定义，以及员工与职位管理。

## 核心特性

- ✅ **多租户隔离**：每个租户独立的组织架构
- ✅ **可配置组织类型**：支持租户自定义组织层级（集团/公司/部门/小组等）
- ✅ **树形结构**：路径枚举 + 闭包表混合存储，支持高效查询
- ✅ **员工管理**：员工信息、职位关联、汇报关系
- ✅ **一人多职**：支持员工兼任多个职位
- ✅ **统计维护**：自动维护员工数量统计
- ✅ **数据权限**：基于组织路径的数据权限过滤

## 数据模型

### 1. 组织类型（OrganizationType）

定义组织的层级类型，如：集团、公司、部门、小组等。

**关键字段**：
- `code`: 类型编码（如 "group", "company", "department"）
- `level`: 建议层级（1=根节点）
- `allowed_parent_types`: 允许的父类型
- `allowed_child_types`: 允许的子类型

**预设类型**（互联网公司）：
```
集团 (group)
  └── 公司 (company)
        ├── 事业部 (division)
        └── 部门 (department)
              └── 小组 (team)
```

### 2. 组织（Organization）

企业的组织实体，支持树形结构。

**关键字段**：
- `code`: 组织编码
- `name`: 组织名称
- `type_id`: 组织类型
- `parent_id`: 父组织
- `path`: 组织路径（`/uuid1/uuid2/uuid3/`）
- `level`: 层级深度
- `leader_id`: 负责人

**树形结构存储**：
```go
type Organization struct {
    Path        string     // "/uuid1/uuid2/uuid3/"
    PathNames   string     // "/集团/公司A/部门B/"
    AncestorIDs []string   // 所有祖先ID数组
    Level       int        // 层级深度
    IsLeaf      bool       // 是否叶子节点
}
```

### 3. 组织闭包表（OrganizationClosure）

用于高效查询祖先和后代关系。

**关键字段**：
- `ancestor_id`: 祖先节点
- `descendant_id`: 后代节点
- `depth`: 距离（0=自己，1=直接子节点）

**用途**：
- 查询所有子组织：`WHERE ancestor_id = ? AND depth > 0`
- 查询所有父组织：`WHERE descendant_id = ? AND depth > 0`
- 查询直接子节点：`WHERE ancestor_id = ? AND depth = 1`

### 4. 职位（Position）

定义组织内的职位。

**关键字段**：
- `code`: 职位编码
- `name`: 职位名称
- `org_id`: 所属组织（NULL 表示全局职位）
- `level`: 职级（数字越大级别越高）
- `category`: 职位类别（management/technical/sales/support）

### 5. 员工（Employee）

员工信息及组织关系。

**关键字段**：
- `user_id`: 关联用户表
- `employee_no`: 工号
- `org_id`: 所属组织
- `org_path`: 组织路径（用于数据权限）
- `position_id`: 主职位
- `direct_leader_id`: 直接上级
- `status`: 状态（probation/active/resigned）

### 6. 员工职位关联（EmployeePosition）

支持一人多职。

**关键字段**：
- `employee_id`: 员工ID
- `position_id`: 职位ID
- `org_id`: 该职位所在组织
- `is_primary`: 是否主职位
- `start_date`, `end_date`: 生效时间

## 数据库表结构

```
organization_types       -- 组织类型配置
  ├── organizations      -- 组织实体（树形结构）
  │     ├── employees    -- 员工
  │     │     └── employee_positions  -- 员工职位关联
  │     └── positions    -- 职位
  └── organization_closures  -- 组织闭包表
```

## 使用示例

### 1. 创建组织类型（仅需一次）

```sql
-- 系统已预设互联网公司的组织类型
-- 租户可以根据需要自定义
INSERT INTO organization_types (tenant_id, code, name, level, allowed_child_types)
VALUES (
    'tenant-uuid',
    'workshop',
    '车间',
    4,
    ARRAY['team']
);
```

### 2. 创建组织

```go
org := &model.Organization{
    TenantID: tenantID,
    Code:     "TECH001",
    Name:     "技术部",
    TypeID:   departmentTypeID,
    TypeCode: "department",
    ParentID: &companyID,
    Level:    3,
    Path:     "/group-id/company-id/dept-id/",
    PathNames: "/集团/公司A/技术部/",
    LeaderID: &leaderUserID,
    Status:   "active",
}
```

### 3. 创建员工

```go
employee := &model.Employee{
    TenantID:   tenantID,
    UserID:     userID,
    EmployeeNo: "EMP001",
    Name:       "张三",
    OrgID:      deptID,
    OrgPath:    "/group-id/company-id/dept-id/",
    PositionID: &engineerPositionID,
    Status:     "active",
}
```

### 4. 查询子组织（使用闭包表）

```sql
-- 查询某组织的所有子组织
SELECT o.*
FROM organizations o
INNER JOIN organization_closures c ON o.id = c.descendant_id
WHERE c.ancestor_id = 'parent-org-id'
AND c.depth > 0
AND c.tenant_id = 'tenant-uuid';
```

### 5. 查询组织下的所有员工（包含子组织）

```sql
-- 使用路径前缀匹配
SELECT *
FROM employees
WHERE org_path LIKE '/parent-org-id/%'
AND status IN ('active', 'probation')
AND tenant_id = 'tenant-uuid';
```

## 树形结构操作

### 1. 路径枚举（Path Enumeration）

**优点**：查询祖先快、插入简单
**用途**：数据权限过滤、面包屑导航

```sql
-- 查询某节点的所有祖先
SELECT * FROM organizations
WHERE id = ANY(string_to_array('/uuid1/uuid2/uuid3/', '/'));

-- 查询某节点的所有后代
SELECT * FROM organizations
WHERE path LIKE '/parent-uuid/%';
```

### 2. 闭包表（Closure Table）

**优点**：查询后代快、支持复杂查询
**用途**：组织树展示、层级统计

```sql
-- 查询直接子节点
SELECT o.* FROM organizations o
INNER JOIN organization_closures c ON o.id = c.descendant_id
WHERE c.ancestor_id = 'parent-id' AND c.depth = 1;

-- 查询指定层级的后代
SELECT o.* FROM organizations o
INNER JOIN organization_closures c ON o.id = c.descendant_id
WHERE c.ancestor_id = 'parent-id' AND c.depth = 3;
```

## 数据权限设计

基于组织路径的数据权限过滤：

```go
// 用户只能查看本部门及子部门的数据
func (s *Service) GetEmployees(ctx context.Context, userOrgPath string) ([]*Employee, error) {
    var employees []*Employee
    err := s.db.Where("org_path LIKE ?", userOrgPath + "%").
        Find(&employees).Error
    return employees, err
}
```

## 性能优化

1. **索引策略**
   - 组织路径：GIN 索引（支持前缀匹配）
   - 闭包表：ancestor_id, descendant_id, depth 索引
   - 员工查询：org_id, org_path 索引

2. **统计字段冗余**
   - `employee_count`: 包含子组织的总员工数
   - `direct_emp_count`: 直属员工数
   - 触发器自动维护

3. **查询优化**
   - 使用路径前缀查询代替递归查询
   - 闭包表支持一次查询获取所有后代

## 扩展点

### 1. 组织变更历史

```go
type OrganizationChange struct {
    ID       uuid.UUID
    OrgID    uuid.UUID
    ChangeType string  // create, update, move, delete
    Before   map[string]interface{} // 变更前数据
    After    map[string]interface{} // 变更后数据
    ChangedBy uuid.UUID
    ChangedAt time.Time
}
```

### 2. 数据同步

支持与第三方平台同步组织架构：
- 钉钉
- 企业微信
- 飞书

### 3. 组织权限

基于组织的角色权限分配：

```go
type OrganizationRole struct {
    OrgID    uuid.UUID
    RoleID   uuid.UUID
    UserID   uuid.UUID
    Scope    string  // org_only, include_children
}
```

## 后续计划

- [ ] 实现组织 Repository 层
- [ ] 实现组织 Service 层（树操作、移动节点、统计）
- [ ] 实现员工 Service 层（入职、离职、调岗）
- [ ] 实现 HTTP API 层
- [ ] 实现数据同步服务（钉钉/企业微信）
- [ ] 实现组织变更审计日志

## 目录结构

```
internal/organization/
├── model/                      # 数据模型
│   ├── organization_type.go
│   ├── organization.go
│   ├── organization_closure.go
│   ├── position.go
│   ├── employee.go
│   └── employee_position.go
├── repository/                 # 数据访问层
├── service/                    # 业务逻辑层
├── handler/                    # HTTP 接口层
├── dto/                        # 数据传输对象
└── migrations/                 # 数据库迁移
    ├── 001_create_organization_types.sql
    ├── 002_create_organizations.sql
    ├── 003_create_organization_closures.sql
    ├── 004_create_positions.sql
    ├── 005_create_employees.sql
    ├── 006_create_employee_positions.sql
    └── 007_create_triggers.sql
```
