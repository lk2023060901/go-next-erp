-- 009_create_leave_tables.sql
-- 请假管理模块相关表

-- 1. 请假类型表
CREATE TABLE IF NOT EXISTS hrm_leave_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_paid BOOLEAN NOT NULL DEFAULT true,
    requires_approval BOOLEAN NOT NULL DEFAULT true,
    requires_proof BOOLEAN NOT NULL DEFAULT false,
    deduct_quota BOOLEAN NOT NULL DEFAULT true,
    unit VARCHAR(20) NOT NULL DEFAULT 'day',
    min_duration DECIMAL(10,2) DEFAULT 0.5,
    max_duration DECIMAL(10,2),
    advance_days INT DEFAULT 0,
    approval_rules JSONB,
    color VARCHAR(20) DEFAULT '#1890ff',
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort INT DEFAULT 0,
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT uk_leave_type_code UNIQUE (tenant_id, code, deleted_at)
);

COMMENT ON TABLE hrm_leave_types IS '请假类型表';
COMMENT ON COLUMN hrm_leave_types.code IS '类型编码：annual_leave, sick_leave, personal_leave, etc.';
COMMENT ON COLUMN hrm_leave_types.is_paid IS '是否带薪';
COMMENT ON COLUMN hrm_leave_types.requires_approval IS '是否需要审批';
COMMENT ON COLUMN hrm_leave_types.requires_proof IS '是否需要证明材料';
COMMENT ON COLUMN hrm_leave_types.deduct_quota IS '是否扣除额度';
COMMENT ON COLUMN hrm_leave_types.unit IS '最小单位：day/half_day/hour';
COMMENT ON COLUMN hrm_leave_types.advance_days IS '需要提前申请的天数';
COMMENT ON COLUMN hrm_leave_types.approval_rules IS '审批规则配置（JSON格式），支持基于天数的动态审批链';

-- 审批规则JSON示例：
-- {
--   "default_chain": [  // 默认审批链（天数未匹配时使用）
--     {"level": 1, "approver_type": "direct_manager", "required": true},
--     {"level": 2, "approver_type": "hr", "required": false}
--   ],
--   "duration_rules": [  // 基于请假天数的规则
--     {
--       "min_duration": 0,
--       "max_duration": 3,
--       "approval_chain": [{"level": 1, "approver_type": "direct_manager", "required": true}]
--     },
--     {
--       "min_duration": 3,
--       "max_duration": 7,
--       "approval_chain": [
--         {"level": 1, "approver_type": "direct_manager", "required": true},
--         {"level": 2, "approver_type": "dept_manager", "required": true}
--       ]
--     },
--     {
--       "min_duration": 7,
--       "max_duration": null,
--       "approval_chain": [
--         {"level": 1, "approver_type": "direct_manager", "required": true},
--         {"level": 2, "approver_type": "dept_manager", "required": true},
--         {"level": 3, "approver_type": "hr", "required": true}
--       ]
--     }
--   ]
-- }

-- 2. 请假额度表
CREATE TABLE IF NOT EXISTS hrm_leave_quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    leave_type_id UUID NOT NULL REFERENCES hrm_leave_types(id),
    year INT NOT NULL,
    total_quota DECIMAL(10,2) NOT NULL DEFAULT 0,
    used_quota DECIMAL(10,2) NOT NULL DEFAULT 0,
    pending_quota DECIMAL(10,2) NOT NULL DEFAULT 0,
    expired_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_leave_quota UNIQUE (tenant_id, employee_id, leave_type_id, year)
);

COMMENT ON TABLE hrm_leave_quotas IS '请假额度表';
COMMENT ON COLUMN hrm_leave_quotas.total_quota IS '总额度（天或小时）';
COMMENT ON COLUMN hrm_leave_quotas.used_quota IS '已使用额度';
COMMENT ON COLUMN hrm_leave_quotas.pending_quota IS '待审批额度';

-- 3. 请假申请表
CREATE TABLE IF NOT EXISTS hrm_leave_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    employee_name VARCHAR(100) NOT NULL,
    department_id UUID,
    leave_type_id UUID NOT NULL REFERENCES hrm_leave_types(id),
    leave_type_name VARCHAR(100) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    duration DECIMAL(10,2) NOT NULL,
    unit VARCHAR(20) NOT NULL DEFAULT 'day',
    reason TEXT NOT NULL,
    proof_urls JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    current_approver_id UUID,
    submitted_at TIMESTAMP,
    approved_at TIMESTAMP,
    rejected_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    remark TEXT,
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

COMMENT ON TABLE hrm_leave_requests IS '请假申请表';
COMMENT ON COLUMN hrm_leave_requests.duration IS '请假时长（天或小时）';
COMMENT ON COLUMN hrm_leave_requests.proof_urls IS '证明材料附件URL数组（JSON格式）';
COMMENT ON COLUMN hrm_leave_requests.status IS '状态：draft/pending/approved/rejected/withdrawn/cancelled';

-- 4. 请假审批记录表
CREATE TABLE IF NOT EXISTS hrm_leave_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    leave_request_id UUID NOT NULL REFERENCES hrm_leave_requests(id),
    approver_id UUID NOT NULL,
    approver_name VARCHAR(100) NOT NULL,
    level INT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    action VARCHAR(20),
    comment TEXT,
    approved_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE hrm_leave_approvals IS '请假审批记录表';
COMMENT ON COLUMN hrm_leave_approvals.level IS '审批层级，1表示第一级';
COMMENT ON COLUMN hrm_leave_approvals.status IS '审批状态：pending/approved/rejected/skipped';
COMMENT ON COLUMN hrm_leave_approvals.action IS '审批动作：approve/reject';

-- 创建索引
CREATE INDEX idx_leave_types_tenant ON hrm_leave_types(tenant_id, is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_leave_types_cursor ON hrm_leave_types(tenant_id, created_at DESC, id DESC) WHERE deleted_at IS NULL;

CREATE INDEX idx_leave_quotas_employee ON hrm_leave_quotas(tenant_id, employee_id, year);
CREATE INDEX idx_leave_quotas_type ON hrm_leave_quotas(tenant_id, leave_type_id, year);

CREATE INDEX idx_leave_requests_employee ON hrm_leave_requests(tenant_id, employee_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_leave_requests_status ON hrm_leave_requests(tenant_id, status, start_time) WHERE deleted_at IS NULL;
CREATE INDEX idx_leave_requests_approver ON hrm_leave_requests(tenant_id, current_approver_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_leave_requests_time ON hrm_leave_requests(tenant_id, start_time, end_time) WHERE deleted_at IS NULL;
CREATE INDEX idx_leave_requests_cursor ON hrm_leave_requests(tenant_id, created_at DESC, id DESC) WHERE deleted_at IS NULL;

CREATE INDEX idx_leave_approvals_request ON hrm_leave_approvals(leave_request_id, level);
CREATE INDEX idx_leave_approvals_approver ON hrm_leave_approvals(tenant_id, approver_id, status);

-- 插入默认请假类型数据
INSERT INTO hrm_leave_types (tenant_id, code, name, description, is_paid, requires_approval, deduct_quota, unit, min_duration, max_duration, advance_days, color, sort) VALUES
('00000000-0000-0000-0000-000000000000', 'annual_leave', '年假', '带薪年假，根据工龄计算', true, true, true, 'day', 0.5, NULL, 3, '#52c41a', 1),
('00000000-0000-0000-0000-000000000000', 'sick_leave', '病假', '因病请假，需要医院证明', true, true, true, 'day', 0.5, NULL, 0, '#faad14', 2),
('00000000-0000-0000-0000-000000000000', 'personal_leave', '事假', '个人事务请假，不带薪', false, true, false, 'day', 0.5, NULL, 1, '#1890ff', 3),
('00000000-0000-0000-0000-000000000000', 'compensatory_leave', '调休', '加班调休', true, true, true, 'day', 0.5, NULL, 0, '#722ed1', 4),
('00000000-0000-0000-0000-000000000000', 'marriage_leave', '婚假', '结婚假期', true, true, true, 'day', 1, 3, 7, '#eb2f96', 5),
('00000000-0000-0000-0000-000000000000', 'maternity_leave', '产假', '女性产假', true, true, true, 'day', 1, 158, 30, '#f5222d', 6),
('00000000-0000-0000-0000-000000000000', 'paternity_leave', '陪产假', '男性陪产假', true, true, true, 'day', 1, 15, 7, '#13c2c2', 7),
('00000000-0000-0000-0000-000000000000', 'bereavement_leave', '丧假', '直系亲属丧假', true, true, false, 'day', 1, 3, 0, '#595959', 8);
