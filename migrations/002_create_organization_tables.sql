-- Organization Module Tables

-- Organization Types table
CREATE TABLE IF NOT EXISTS organization_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    code VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    icon VARCHAR(100),
    level INT NOT NULL,
    max_level INT DEFAULT 1,
    allow_root BOOLEAN DEFAULT FALSE,
    allow_multi BOOLEAN DEFAULT TRUE,
    allowed_parent_types TEXT[],
    allowed_child_types TEXT[],
    enable_leader BOOLEAN DEFAULT TRUE,
    enable_legal_info BOOLEAN DEFAULT FALSE,
    enable_address BOOLEAN DEFAULT FALSE,
    sort INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    is_system BOOLEAN DEFAULT FALSE,
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_organization_types_code_tenant ON organization_types(code, tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_organization_types_tenant ON organization_types(tenant_id);
CREATE INDEX IF NOT EXISTS idx_organization_types_status ON organization_types(status);

-- Organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    code VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    short_name VARCHAR(100),
    description TEXT,
    type_id UUID NOT NULL REFERENCES organization_types(id),
    type_code VARCHAR(50),
    parent_id UUID REFERENCES organizations(id),
    level INT NOT NULL,
    path VARCHAR(1000),
    path_names VARCHAR(1000),
    ancestor_ids UUID[],
    is_leaf BOOLEAN DEFAULT TRUE,
    leader_id UUID REFERENCES users(id),
    leader_name VARCHAR(100),
    legal_person VARCHAR(100),
    unified_code VARCHAR(50),
    register_date DATE,
    register_addr VARCHAR(500),
    phone VARCHAR(50),
    email VARCHAR(100),
    address VARCHAR(500),
    employee_count INT DEFAULT 0,
    direct_emp_count INT DEFAULT 0,
    sort INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    tags TEXT[],
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_code_tenant ON organizations(code, tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_organizations_tenant ON organizations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_organizations_type ON organizations(type_id);
CREATE INDEX IF NOT EXISTS idx_organizations_parent ON organizations(parent_id);
CREATE INDEX IF NOT EXISTS idx_organizations_leader ON organizations(leader_id);
CREATE INDEX IF NOT EXISTS idx_organizations_status ON organizations(status);

-- Positions table
CREATE TABLE IF NOT EXISTS positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) NOT NULL,
    description TEXT,
    organization_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    level INT,
    sort_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_positions_code_tenant ON positions(code, tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_positions_tenant ON positions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_positions_organization ON positions(organization_id);
CREATE INDEX IF NOT EXISTS idx_positions_active ON positions(is_active);

-- Employees table (员工表)
CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL REFERENCES users(id),
    employee_no VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    gender VARCHAR(10),
    birth_date DATE,
    id_card_no VARCHAR(50),
    mobile VARCHAR(20),
    email VARCHAR(100),
    avatar VARCHAR(500),
    org_id UUID NOT NULL REFERENCES organizations(id),
    org_path VARCHAR(1000),
    org_name VARCHAR(255),
    position_id UUID REFERENCES positions(id),
    position_name VARCHAR(100),
    job_level VARCHAR(50),
    entry_date DATE,
    probation_end_date DATE,
    regular_date DATE,
    contract_start_date DATE,
    contract_end_date DATE,
    status VARCHAR(20) DEFAULT 'probation',
    is_leader BOOLEAN DEFAULT FALSE,
    superior_id UUID REFERENCES employees(id),
    superior_name VARCHAR(100),
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_employees_no_tenant ON employees(employee_no, tenant_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_employees_user_tenant ON employees(user_id, tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employees_tenant ON employees(tenant_id);
CREATE INDEX IF NOT EXISTS idx_employees_org ON employees(org_id);
CREATE INDEX IF NOT EXISTS idx_employees_position ON employees(position_id);
CREATE INDEX IF NOT EXISTS idx_employees_status ON employees(status);
CREATE INDEX IF NOT EXISTS idx_employees_superior ON employees(superior_id);
CREATE INDEX IF NOT EXISTS idx_employees_mobile ON employees(mobile) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email) WHERE deleted_at IS NULL;

-- User Positions junction table
CREATE TABLE IF NOT EXISTS user_positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    position_id UUID NOT NULL REFERENCES positions(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    is_primary BOOLEAN DEFAULT FALSE,
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_positions_unique ON user_positions(user_id, position_id, organization_id);
CREATE INDEX IF NOT EXISTS idx_user_positions_user ON user_positions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_positions_position ON user_positions(position_id);
CREATE INDEX IF NOT EXISTS idx_user_positions_organization ON user_positions(organization_id);
CREATE INDEX IF NOT EXISTS idx_user_positions_tenant ON user_positions(tenant_id);

-- Organization Closures table (闭包表，用于组织树查询优化)
CREATE TABLE IF NOT EXISTS organization_closures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    ancestor_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    descendant_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    depth INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_org_closures_unique ON organization_closures(tenant_id, ancestor_id, descendant_id);
CREATE INDEX IF NOT EXISTS idx_org_closures_ancestor ON organization_closures(ancestor_id);
CREATE INDEX IF NOT EXISTS idx_org_closures_descendant ON organization_closures(descendant_id);
CREATE INDEX IF NOT EXISTS idx_org_closures_depth ON organization_closures(depth);
CREATE INDEX IF NOT EXISTS idx_org_closures_tenant ON organization_closures(tenant_id);
