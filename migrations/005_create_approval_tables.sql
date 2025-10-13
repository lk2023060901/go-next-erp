-- 创建流程定义表
CREATE TABLE IF NOT EXISTS approval_process_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,
    form_id UUID NOT NULL REFERENCES form_definitions(id),
    workflow_id UUID NOT NULL,
    icon VARCHAR(100),
    description TEXT,
    enabled BOOLEAN DEFAULT true,
    sort INTEGER DEFAULT 0,
    created_by UUID NOT NULL,
    updated_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- 使用部分唯一索引代替 UNIQUE 约束，只对未删除的记录生效
CREATE UNIQUE INDEX IF NOT EXISTS uk_approval_process_defs_tenant_code
    ON approval_process_definitions(tenant_id, code)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_approval_process_defs_tenant ON approval_process_definitions(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_approval_process_defs_enabled ON approval_process_definitions(enabled) WHERE deleted_at IS NULL;

-- 创建流程实例表
CREATE TABLE IF NOT EXISTS approval_process_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    process_def_id UUID NOT NULL REFERENCES approval_process_definitions(id),
    process_def_code VARCHAR(50) NOT NULL,
    process_def_name VARCHAR(100) NOT NULL,
    workflow_instance_id UUID NOT NULL,
    form_data_id UUID NOT NULL REFERENCES form_data(id),
    applicant_id UUID NOT NULL,
    applicant_name VARCHAR(100) NOT NULL,
    title VARCHAR(200) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    current_node_id VARCHAR(50),
    current_node_name VARCHAR(100),
    variables JSONB,
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_approval_process_instances_tenant ON approval_process_instances(tenant_id);
CREATE INDEX idx_approval_process_instances_applicant ON approval_process_instances(applicant_id);
CREATE INDEX idx_approval_process_instances_status ON approval_process_instances(status);
CREATE INDEX idx_approval_process_instances_def ON approval_process_instances(process_def_id);

-- 创建审批任务表
CREATE TABLE IF NOT EXISTS approval_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    process_instance_id UUID NOT NULL REFERENCES approval_process_instances(id),
    node_id VARCHAR(50) NOT NULL,
    node_name VARCHAR(100) NOT NULL,
    assignee_id UUID NOT NULL,
    assignee_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    action VARCHAR(20),
    comment TEXT,
    attachments JSONB,
    transfer_to_id UUID,
    transfer_to_name VARCHAR(100),
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_approval_tasks_tenant ON approval_tasks(tenant_id);
CREATE INDEX idx_approval_tasks_assignee ON approval_tasks(assignee_id);
CREATE INDEX idx_approval_tasks_status ON approval_tasks(status);
CREATE INDEX idx_approval_tasks_process ON approval_tasks(process_instance_id);

-- 创建流程历史表
CREATE TABLE IF NOT EXISTS approval_process_histories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    process_instance_id UUID NOT NULL REFERENCES approval_process_instances(id),
    task_id UUID,
    node_id VARCHAR(50) NOT NULL,
    node_name VARCHAR(100) NOT NULL,
    operator_id UUID NOT NULL,
    operator_name VARCHAR(100) NOT NULL,
    action VARCHAR(20) NOT NULL,
    comment TEXT,
    from_status VARCHAR(20),
    to_status VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_approval_histories_process ON approval_process_histories(process_instance_id);
CREATE INDEX idx_approval_histories_operator ON approval_process_histories(operator_id);
CREATE INDEX idx_approval_histories_created ON approval_process_histories(created_at DESC);

-- 添加注释
COMMENT ON TABLE approval_process_definitions IS '审批流程定义表';
COMMENT ON TABLE approval_process_instances IS '审批流程实例表';
COMMENT ON TABLE approval_tasks IS '审批任务表';
COMMENT ON TABLE approval_process_histories IS '审批流程历史表';
