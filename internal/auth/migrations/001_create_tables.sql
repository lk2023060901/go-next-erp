-- 4A 权限系统数据库表结构
-- 使用 UUID v7 作为主键

-- 1. 租户表
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    domain VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'active',

    -- 配额
    max_users INT NOT NULL DEFAULT 100,
    max_storage BIGINT NOT NULL DEFAULT 10737418240, -- 10GB

    -- 配置
    settings JSONB,

    -- 时间戳
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- 索引
    UNIQUE(domain)
);

CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at);

-- 2. 用户表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active',

    -- MFA
    mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_secret VARCHAR(255),

    -- 安全
    last_login_at TIMESTAMP,
    last_login_ip VARCHAR(45),
    login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMP,

    -- 元数据
    metadata JSONB,

    -- 时间戳
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- 索引
    UNIQUE(tenant_id, username),
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- 3. 角色表
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    parent_id UUID REFERENCES roles(id), -- 角色继承

    -- 时间戳
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- 索引
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_roles_parent_id ON roles(parent_id);
CREATE INDEX idx_roles_deleted_at ON roles(deleted_at);

-- 4. 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    -- 时间戳
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- 索引
    UNIQUE(tenant_id, resource, action)
);

CREATE INDEX idx_permissions_tenant_id ON permissions(tenant_id);
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_deleted_at ON permissions(deleted_at);

-- 5. 用户-角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- 索引
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_tenant_id ON user_roles(tenant_id);

-- 6. 角色-权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- 索引
    UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- 7. ABAC 策略表
CREATE TABLE IF NOT EXISTS policies (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    expression TEXT NOT NULL, -- Expr 表达式
    effect VARCHAR(10) NOT NULL, -- allow, deny
    priority INT NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,

    -- 时间戳
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- 索引
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_policies_tenant_id ON policies(tenant_id);
CREATE INDEX idx_policies_resource_action ON policies(resource, action);
CREATE INDEX idx_policies_priority ON policies(priority DESC);
CREATE INDEX idx_policies_enabled ON policies(enabled);
CREATE INDEX idx_policies_deleted_at ON policies(deleted_at);

-- 8. ReBAC 关系元组表
CREATE TABLE IF NOT EXISTS relation_tuples (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    subject VARCHAR(255) NOT NULL, -- user:123, group:456
    relation VARCHAR(50) NOT NULL, -- owner, editor, viewer
    object VARCHAR(255) NOT NULL,  -- document:789, folder:abc

    -- 时间戳
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    -- 索引
    UNIQUE(tenant_id, subject, relation, object)
);

CREATE INDEX idx_relation_tuples_tenant_id ON relation_tuples(tenant_id);
CREATE INDEX idx_relation_tuples_subject ON relation_tuples(subject);
CREATE INDEX idx_relation_tuples_object ON relation_tuples(object);
CREATE INDEX idx_relation_tuples_relation ON relation_tuples(relation);
CREATE INDEX idx_relation_tuples_deleted_at ON relation_tuples(deleted_at);

-- 9. 会话表
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    token TEXT NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- 时间控制
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP,

    -- 索引
    UNIQUE(token)
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_revoked_at ON sessions(revoked_at);

-- 10. 审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY,
    event_id VARCHAR(255) NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100),
    resource_id VARCHAR(255),

    -- 操作详情
    before_data JSONB,
    after_data JSONB,

    -- 请求信息
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- 结果
    result VARCHAR(20) NOT NULL, -- success, failure, denied
    error_msg TEXT,

    -- 元数据
    metadata JSONB,

    -- 时间戳
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- 索引
    UNIQUE(event_id)
);

CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX idx_audit_logs_result ON audit_logs(result);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- 创建审计日志分区（按月）
-- CREATE TABLE audit_logs_2025_01 PARTITION OF audit_logs
--     FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

-- 注释
COMMENT ON TABLE tenants IS '租户表';
COMMENT ON TABLE users IS '用户表';
COMMENT ON TABLE roles IS '角色表';
COMMENT ON TABLE permissions IS '权限表';
COMMENT ON TABLE user_roles IS '用户-角色关联表';
COMMENT ON TABLE role_permissions IS '角色-权限关联表';
COMMENT ON TABLE policies IS 'ABAC 策略表';
COMMENT ON TABLE relation_tuples IS 'ReBAC 关系元组表';
COMMENT ON TABLE sessions IS '会话表';
COMMENT ON TABLE audit_logs IS '审计日志表';
