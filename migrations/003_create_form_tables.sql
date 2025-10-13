-- 创建表单定义表
CREATE TABLE IF NOT EXISTS form_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    fields JSONB NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_by UUID NOT NULL,
    updated_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uk_form_definitions_tenant_code UNIQUE (tenant_id, code, deleted_at)
);

-- 创建索引
CREATE INDEX idx_form_definitions_tenant_id ON form_definitions(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_definitions_code ON form_definitions(code) WHERE deleted_at IS NULL;
CREATE INDEX idx_form_definitions_enabled ON form_definitions(enabled) WHERE deleted_at IS NULL;

-- 添加注释
COMMENT ON TABLE form_definitions IS '表单定义表';
COMMENT ON COLUMN form_definitions.id IS '主键';
COMMENT ON COLUMN form_definitions.tenant_id IS '租户ID';
COMMENT ON COLUMN form_definitions.code IS '表单编码（唯一标识）';
COMMENT ON COLUMN form_definitions.name IS '表单名称';
COMMENT ON COLUMN form_definitions.fields IS '字段列表（JSON）';
COMMENT ON COLUMN form_definitions.enabled IS '是否启用';
COMMENT ON COLUMN form_definitions.created_by IS '创建人';
COMMENT ON COLUMN form_definitions.updated_by IS '更新人';
COMMENT ON COLUMN form_definitions.created_at IS '创建时间';
COMMENT ON COLUMN form_definitions.updated_at IS '更新时间';
COMMENT ON COLUMN form_definitions.deleted_at IS '删除时间（软删除）';

-- 创建表单数据表
CREATE TABLE IF NOT EXISTS form_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    form_id UUID NOT NULL,
    data JSONB NOT NULL,
    submitted_by UUID NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    related_type VARCHAR(50),
    related_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_form_data_form_id FOREIGN KEY (form_id) REFERENCES form_definitions(id)
);

-- 创建索引
CREATE INDEX idx_form_data_tenant_id ON form_data(tenant_id);
CREATE INDEX idx_form_data_form_id ON form_data(form_id);
CREATE INDEX idx_form_data_submitted_by ON form_data(submitted_by);
CREATE INDEX idx_form_data_related ON form_data(related_type, related_id);
CREATE INDEX idx_form_data_submitted_at ON form_data(submitted_at DESC);

-- 添加注释
COMMENT ON TABLE form_data IS '表单数据表';
COMMENT ON COLUMN form_data.id IS '主键';
COMMENT ON COLUMN form_data.tenant_id IS '租户ID';
COMMENT ON COLUMN form_data.form_id IS '表单定义ID';
COMMENT ON COLUMN form_data.data IS '表单数据（JSON）';
COMMENT ON COLUMN form_data.submitted_by IS '提交人';
COMMENT ON COLUMN form_data.submitted_at IS '提交时间';
COMMENT ON COLUMN form_data.related_type IS '关联类型';
COMMENT ON COLUMN form_data.related_id IS '关联ID';
COMMENT ON COLUMN form_data.created_at IS '创建时间';
COMMENT ON COLUMN form_data.updated_at IS '更新时间';
