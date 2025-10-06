-- 初始化数据

-- 1. 创建默认租户
INSERT INTO tenants (id, name, display_name, domain, status, max_users, max_storage, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'default',
    '默认租户',
    'default.local',
    'active',
    1000,
    107374182400, -- 100GB
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;

-- 2. 创建默认角色
WITH default_tenant AS (
    SELECT id FROM tenants WHERE name = 'default' LIMIT 1
)
INSERT INTO roles (id, name, display_name, description, tenant_id, created_at, updated_at)
SELECT
    gen_random_uuid(),
    role_data.name,
    role_data.display_name,
    role_data.description,
    default_tenant.id,
    NOW(),
    NOW()
FROM default_tenant, (
    VALUES
        ('admin', '系统管理员', '拥有系统所有权限'),
        ('manager', '部门经理', '拥有部门管理权限'),
        ('user', '普通用户', '基础用户权限'),
        ('guest', '访客', '只读权限')
) AS role_data(name, display_name, description)
ON CONFLICT DO NOTHING;

-- 3. 创建默认权限（CRUD）
WITH default_tenant AS (
    SELECT id FROM tenants WHERE name = 'default' LIMIT 1
)
INSERT INTO permissions (id, resource, action, display_name, description, tenant_id, created_at, updated_at)
SELECT
    gen_random_uuid(),
    perm_data.resource,
    perm_data.action,
    perm_data.display_name,
    perm_data.description,
    default_tenant.id,
    NOW(),
    NOW()
FROM default_tenant, (
    VALUES
        -- 用户权限
        ('user', 'create', '创建用户', '可以创建新用户'),
        ('user', 'read', '查看用户', '可以查看用户信息'),
        ('user', 'update', '更新用户', '可以更新用户信息'),
        ('user', 'delete', '删除用户', '可以删除用户'),

        -- 角色权限
        ('role', 'create', '创建角色', '可以创建新角色'),
        ('role', 'read', '查看角色', '可以查看角色信息'),
        ('role', 'update', '更新角色', '可以更新角色信息'),
        ('role', 'delete', '删除角色', '可以删除角色'),

        -- 权限管理
        ('permission', 'create', '创建权限', '可以创建新权限'),
        ('permission', 'read', '查看权限', '可以查看权限信息'),
        ('permission', 'update', '更新权限', '可以更新权限信息'),
        ('permission', 'delete', '删除权限', '可以删除权限'),

        -- 策略管理
        ('policy', 'create', '创建策略', '可以创建 ABAC 策略'),
        ('policy', 'read', '查看策略', '可以查看策略信息'),
        ('policy', 'update', '更新策略', '可以更新策略信息'),
        ('policy', 'delete', '删除策略', '可以删除策略'),

        -- 审计日志
        ('audit', 'read', '查看审计日志', '可以查看审计日志'),
        ('audit', 'export', '导出审计日志', '可以导出审计日志'),

        -- 通配符权限
        ('*', '*', '所有权限', '拥有所有资源的所有权限')
) AS perm_data(resource, action, display_name, description)
ON CONFLICT DO NOTHING;

-- 4. 为管理员角色分配所有权限
WITH default_tenant AS (
    SELECT id FROM tenants WHERE name = 'default' LIMIT 1
),
admin_role AS (
    SELECT id FROM roles WHERE name = 'admin' AND tenant_id = (SELECT id FROM default_tenant) LIMIT 1
),
all_permission AS (
    SELECT id FROM permissions WHERE resource = '*' AND action = '*' AND tenant_id = (SELECT id FROM default_tenant) LIMIT 1
)
INSERT INTO role_permissions (id, role_id, permission_id, tenant_id, created_at)
SELECT
    gen_random_uuid(),
    admin_role.id,
    all_permission.id,
    default_tenant.id,
    NOW()
FROM default_tenant, admin_role, all_permission
ON CONFLICT DO NOTHING;

-- 5. 为经理角色分配部分权限
WITH default_tenant AS (
    SELECT id FROM tenants WHERE name = 'default' LIMIT 1
),
manager_role AS (
    SELECT id FROM roles WHERE name = 'manager' AND tenant_id = (SELECT id FROM default_tenant) LIMIT 1
)
INSERT INTO role_permissions (id, role_id, permission_id, tenant_id, created_at)
SELECT
    gen_random_uuid(),
    manager_role.id,
    p.id,
    default_tenant.id,
    NOW()
FROM default_tenant, manager_role, permissions p
WHERE p.tenant_id = (SELECT id FROM default_tenant)
  AND p.resource IN ('user', 'role', 'audit')
  AND p.action IN ('read', 'create', 'update')
ON CONFLICT DO NOTHING;

-- 6. 为普通用户角色分配只读权限
WITH default_tenant AS (
    SELECT id FROM tenants WHERE name = 'default' LIMIT 1
),
user_role AS (
    SELECT id FROM roles WHERE name = 'user' AND tenant_id = (SELECT id FROM default_tenant) LIMIT 1
)
INSERT INTO role_permissions (id, role_id, permission_id, tenant_id, created_at)
SELECT
    gen_random_uuid(),
    user_role.id,
    p.id,
    default_tenant.id,
    NOW()
FROM default_tenant, user_role, permissions p
WHERE p.tenant_id = (SELECT id FROM default_tenant)
  AND p.action = 'read'
ON CONFLICT DO NOTHING;

-- 7. 创建默认 ABAC 策略
WITH default_tenant AS (
    SELECT id FROM tenants WHERE name = 'default' LIMIT 1
)
INSERT INTO policies (id, name, description, tenant_id, resource, action, expression, effect, priority, enabled, created_at, updated_at)
SELECT
    gen_random_uuid(),
    policy_data.name,
    policy_data.description,
    default_tenant.id,
    policy_data.resource,
    policy_data.action,
    policy_data.expression,
    policy_data.effect,
    policy_data.priority,
    true,
    NOW(),
    NOW()
FROM default_tenant, (
    VALUES
        -- 工作时间策略
        ('work_hours', '工作时间访问', '*', '*',
         'Time.Hour >= 9 && Time.Hour <= 18 && Time.Weekday >= 1 && Time.Weekday <= 5',
         'allow', 100),

        -- 同部门访问策略
        ('same_department', '同部门可访问', '*', 'read',
         'User.DepartmentID == Resource.DepartmentID',
         'allow', 50),

        -- 资源所有者策略
        ('resource_owner', '资源所有者可访问', '*', '*',
         'User.ID == Resource.OwnerID',
         'allow', 200),

        -- 默认拒绝策略
        ('default_deny', '默认拒绝', '*', '*',
         'true',
         'deny', 1)
) AS policy_data(name, description, resource, action, expression, effect, priority)
ON CONFLICT DO NOTHING;

-- 8. 创建超级管理员用户
-- 密码：Admin@123（Argon2 哈希，需要在应用中生成）
WITH default_tenant AS (
    SELECT id FROM tenants WHERE name = 'default' LIMIT 1
)
INSERT INTO users (id, username, email, password_hash, tenant_id, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    'admin',
    'admin@example.com',
    '$argon2id$v=19$m=65536,t=3,p=2$c29tZXJhbmRvbXNhbHQ$YW5vdGhlcmhhc2h2YWx1ZQ', -- 占位符，需替换
    default_tenant.id,
    'active',
    NOW(),
    NOW()
FROM default_tenant
ON CONFLICT DO NOTHING;

-- 9. 为超级管理员分配角色
WITH default_tenant AS (
    SELECT id FROM tenants WHERE name = 'default' LIMIT 1
),
admin_user AS (
    SELECT id FROM users WHERE username = 'admin' AND tenant_id = (SELECT id FROM default_tenant) LIMIT 1
),
admin_role AS (
    SELECT id FROM roles WHERE name = 'admin' AND tenant_id = (SELECT id FROM default_tenant) LIMIT 1
)
INSERT INTO user_roles (id, user_id, role_id, tenant_id, created_at)
SELECT
    gen_random_uuid(),
    admin_user.id,
    admin_role.id,
    default_tenant.id,
    NOW()
FROM default_tenant, admin_user, admin_role
ON CONFLICT DO NOTHING;

-- 提示
SELECT '✅ 初始数据插入完成！' AS status;
SELECT '默认管理员账号：admin' AS info;
SELECT '默认密码需要在应用中设置' AS warning;
