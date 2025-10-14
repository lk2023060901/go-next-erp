-- HRM Module Tables (Part 2)
-- 考勤记录、请假、加班、出差等表

-- =============================================================================
-- 7. 考勤记录表 (Attendance Records)
-- =============================================================================
-- 打卡记录表（按月分区以提升性能）
CREATE TABLE IF NOT EXISTS hrm_attendance_records (
    id UUID DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    
    -- 员工信息（关联 organization.employees）
    employee_id UUID NOT NULL REFERENCES employees(id),  -- 对应 organization.employees.id
    employee_name VARCHAR(100),       -- 冗余字段，便于查询
    department_id UUID REFERENCES organizations(id),  -- 冗余字段
    
    -- 打卡信息
    clock_time TIMESTAMP NOT NULL,    -- 打卡时间
    clock_type VARCHAR(20) NOT NULL,  -- check_in, check_out
    status VARCHAR(20) NOT NULL,      -- normal, late, early, absent, leave, overtime, trip
    
    -- 班次信息
    shift_id UUID REFERENCES hrm_shifts(id),
    shift_name VARCHAR(100),          -- 冗余字段
    
    -- 打卡方式和来源
    check_in_method VARCHAR(20) NOT NULL,  -- device, mobile, web, face, fingerprint, card, manual
    source_type VARCHAR(20) NOT NULL,  -- system, device, dingtalk, wecom, feishu, manual
    source_id VARCHAR(100),           -- 来源标识（设备ID或平台ID）
    
    -- 定位信息（支持地理围栏打卡）
    location JSONB,                   -- GPS定位 {latitude, longitude, accuracy}
    address VARCHAR(500),             -- 地址
    wifi_ssid VARCHAR(100),           -- WiFi名称
    wifi_mac VARCHAR(50),             -- WiFi MAC地址
    
    -- 生物识别信息
    photo_url VARCHAR(500),           -- 打卡照片
    face_score DECIMAL(5,4),          -- 人脸识别分数
    temperature DECIMAL(4,2),         -- 体温（疫情期间）
    
    -- 异常信息
    is_exception BOOLEAN DEFAULT FALSE,  -- 是否异常
    exception_reason VARCHAR(200),    -- 异常原因
    exception_type VARCHAR(50),       -- 异常类型（迟到、早退、缺卡等）
    
    -- 审批关联
    approval_id UUID,                 -- 补卡审批ID
    
    -- 原始数据（用于问题排查和审计）
    raw_data JSONB,
    
    -- 备注
    remark TEXT,
    
    -- 审计字段
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- 主键必须包含分区键
    PRIMARY KEY (id, clock_time)
) PARTITION BY RANGE (clock_time);

-- 创建月度分区（示例：2025年1月到12月）
CREATE TABLE IF NOT EXISTS hrm_attendance_records_2025_01 PARTITION OF hrm_attendance_records
    FOR VALUES FROM ('2025-01-01 00:00:00') TO ('2025-02-01 00:00:00');
CREATE TABLE IF NOT EXISTS hrm_attendance_records_2025_02 PARTITION OF hrm_attendance_records
    FOR VALUES FROM ('2025-02-01 00:00:00') TO ('2025-03-01 00:00:00');
CREATE TABLE IF NOT EXISTS hrm_attendance_records_2025_03 PARTITION OF hrm_attendance_records
    FOR VALUES FROM ('2025-03-01 00:00:00') TO ('2025-04-01 00:00:00');
CREATE TABLE IF NOT EXISTS hrm_attendance_records_2025_04 PARTITION OF hrm_attendance_records
    FOR VALUES FROM ('2025-04-01 00:00:00') TO ('2025-05-01 00:00:00');
CREATE TABLE IF NOT EXISTS hrm_attendance_records_2025_05 PARTITION OF hrm_attendance_records
    FOR VALUES FROM ('2025-05-01 00:00:00') TO ('2025-06-01 00:00:00');
CREATE TABLE IF NOT EXISTS hrm_attendance_records_2025_06 PARTITION OF hrm_attendance_records
    FOR VALUES FROM ('2025-06-01 00:00:00') TO ('2025-07-01 00:00:00');
CREATE TABLE IF NOT EXISTS hrm_attendance_records_2025_07 PARTITION OF hrm_attendance_records
    FOR VALUES FROM ('2025-07-01 00:00:00') TO ('2025-08-01 00:00:00');
CREATE TABLE IF NOT EXISTS hrm_attendance_records_2025_08 PARTITION OF hrm_attendance_records
    FOR VALUES FROM ('2025-08-01 00:00:00') TO ('2025-09-01 00:00:00');
CREATE TABLE IF NOT EXISTS hrm_attendance_records_2025_09 PARTITION OF hrm_attendance_records
    FOR VALUES FROM ('2025-09-01 00:00:00') TO ('2025-10-01 00:00:00');
CREATE TABLE IF NOT EXISTS hrm_attendance_records_2025_10 PARTITION OF hrm_attendance_records
    FOR VALUES FROM ('2025-10-01 00:00:00') TO ('2025-11-01 00:00:00');
CREATE TABLE IF NOT EXISTS hrm_attendance_records_2025_11 PARTITION OF hrm_attendance_records
    FOR VALUES FROM ('2025-11-01 00:00:00') TO ('2025-12-01 00:00:00');
CREATE TABLE IF NOT EXISTS hrm_attendance_records_2025_12 PARTITION OF hrm_attendance_records
    FOR VALUES FROM ('2025-12-01 00:00:00') TO ('2026-01-01 00:00:00');

-- 创建索引（在主表上创建，会自动应用到所有分区）
CREATE INDEX IF NOT EXISTS idx_attendance_records_tenant ON hrm_attendance_records(tenant_id);
CREATE INDEX IF NOT EXISTS idx_attendance_records_employee_date ON hrm_attendance_records(tenant_id, employee_id, clock_time);
CREATE INDEX IF NOT EXISTS idx_attendance_records_dept_date ON hrm_attendance_records(tenant_id, department_id, clock_time);
CREATE INDEX IF NOT EXISTS idx_attendance_records_status ON hrm_attendance_records(status);
CREATE INDEX IF NOT EXISTS idx_attendance_records_source ON hrm_attendance_records(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_attendance_records_exception ON hrm_attendance_records(is_exception) WHERE is_exception = TRUE;
CREATE INDEX IF NOT EXISTS idx_attendance_records_shift ON hrm_attendance_records(shift_id);

COMMENT ON TABLE hrm_attendance_records IS '考勤记录表（按月分区）';
COMMENT ON COLUMN hrm_attendance_records.clock_type IS '打卡类型: check_in(上班), check_out(下班)';
COMMENT ON COLUMN hrm_attendance_records.status IS '考勤状态: normal(正常), late(迟到), early(早退), absent(旷工), leave(请假), overtime(加班), trip(出差)';
COMMENT ON COLUMN hrm_attendance_records.check_in_method IS '打卡方式: device(考勤机), mobile(手机APP), web(网页), face(人脸), fingerprint(指纹), card(刷卡), manual(手动补卡)';
COMMENT ON COLUMN hrm_attendance_records.source_type IS '数据来源: system(系统), device(考勤机), dingtalk(钉钉), wecom(企微), feishu(飞书), manual(手动)';

-- =============================================================================
-- 8. 请假类型表 (Leave Types)
-- =============================================================================
CREATE TABLE IF NOT EXISTS hrm_leave_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    
    -- 类型信息
    code VARCHAR(50) NOT NULL,        -- 编码（唯一）
    name VARCHAR(100) NOT NULL,       -- 名称
    description TEXT,                 -- 描述
    
    -- 假期属性
    is_paid BOOLEAN DEFAULT TRUE,     -- 是否带薪
    need_approval BOOLEAN DEFAULT TRUE,  -- 是否需要审批
    need_attachment BOOLEAN DEFAULT FALSE,  -- 是否需要附件
    deduct_salary BOOLEAN DEFAULT FALSE,  -- 是否扣薪
    pay_rate DECIMAL(3,2) DEFAULT 1.0,  -- 薪资比例（0-1）
    
    -- 额度设置
    has_quota BOOLEAN DEFAULT TRUE,   -- 是否有额度限制
    quota_type VARCHAR(20),           -- annual, monthly, total
    quota_days DECIMAL(5,2),          -- 额度天数
    carry_forward BOOLEAN DEFAULT FALSE,  -- 是否可结转
    max_carry_days DECIMAL(5,2),      -- 最大结转天数
    
    -- 申请限制
    min_unit VARCHAR(20),             -- day, half_day, hour
    min_duration DECIMAL(5,2),        -- 最小时长
    max_duration DECIMAL(5,2),        -- 最大时长
    min_advance_days INTEGER,         -- 最少提前天数
    max_advance_days INTEGER,         -- 最多提前天数
    
    -- 适用范围
    apply_type VARCHAR(20) NOT NULL,  -- all, department, employee
    department_ids UUID[],
    employee_ids UUID[],
    
    -- 颜色标识
    color VARCHAR(20),
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    sort INTEGER DEFAULT 0,
    
    -- 审计字段
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_leave_types_code_tenant ON hrm_leave_types(tenant_id, code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_leave_types_tenant ON hrm_leave_types(tenant_id);
CREATE INDEX IF NOT EXISTS idx_leave_types_active ON hrm_leave_types(is_active);

COMMENT ON TABLE hrm_leave_types IS '请假类型表';
COMMENT ON COLUMN hrm_leave_types.is_paid IS '是否带薪，默认 true';
COMMENT ON COLUMN hrm_leave_types.quota_type IS '额度类型: annual(年度), monthly(月度), total(总额)';
COMMENT ON COLUMN hrm_leave_types.min_unit IS '最小单位: day(天), half_day(半天), hour(小时)';

-- =============================================================================
-- 9. 请假记录表 (Leave Records)
-- =============================================================================
CREATE TABLE IF NOT EXISTS hrm_leaves (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    
    -- 申请人信息（关联 organization.employees）
    employee_id UUID NOT NULL REFERENCES employees(id),  -- 对应 organization.employees.id
    employee_name VARCHAR(100),       -- 冗余
    department_id UUID REFERENCES organizations(id),  -- 冗余
    
    -- 请假类型
    leave_type_id UUID NOT NULL REFERENCES hrm_leave_types(id),
    leave_type_name VARCHAR(100),     -- 冗余
    
    -- 请假时间
    start_time TIMESTAMP NOT NULL,    -- 开始时间
    end_time TIMESTAMP NOT NULL,      -- 结束时间
    duration DECIMAL(5,2) NOT NULL,   -- 请假天数（支持小数）
    unit VARCHAR(20) NOT NULL,        -- day, hour
    
    -- 请假理由
    reason TEXT NOT NULL,
    attachment TEXT[],                -- 附件（证明材料）
    
    -- 审批信息
    approval_id UUID,                 -- 关联审批流程
    approval_status VARCHAR(20) DEFAULT 'pending',  -- pending, approved, rejected
    approved_by UUID,
    approved_at TIMESTAMP,
    reject_reason TEXT,
    
    -- 销假信息
    actual_end_time TIMESTAMP,        -- 实际结束时间
    is_canceled BOOLEAN DEFAULT FALSE,  -- 是否撤销
    
    -- 备注
    remark TEXT,
    
    -- 审计字段
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_leaves_tenant ON hrm_leaves(tenant_id);
CREATE INDEX IF NOT EXISTS idx_leaves_employee ON hrm_leaves(employee_id);
CREATE INDEX IF NOT EXISTS idx_leaves_type ON hrm_leaves(leave_type_id);
CREATE INDEX IF NOT EXISTS idx_leaves_status ON hrm_leaves(approval_status);
CREATE INDEX IF NOT EXISTS idx_leaves_time ON hrm_leaves(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_leaves_department ON hrm_leaves(department_id);

COMMENT ON TABLE hrm_leaves IS '请假记录表';
COMMENT ON COLUMN hrm_leaves.approval_status IS '审批状态: pending(待审批), approved(已批准), rejected(已拒绝)，默认 pending';
COMMENT ON COLUMN hrm_leaves.unit IS '单位: day(天), hour(小时)';

-- =============================================================================
-- 10. 请假额度表 (Leave Quotas)
-- =============================================================================
CREATE TABLE IF NOT EXISTS hrm_leave_quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    employee_id UUID NOT NULL REFERENCES employees(id),
    leave_type_id UUID NOT NULL REFERENCES hrm_leave_types(id),
    
    -- 额度信息
    year INTEGER NOT NULL,            -- 年份
    total_days DECIMAL(5,2) NOT NULL,  -- 总额度
    used_days DECIMAL(5,2) DEFAULT 0,  -- 已使用
    remaining_days DECIMAL(5,2) NOT NULL,  -- 剩余
    carried_days DECIMAL(5,2) DEFAULT 0,   -- 结转
    
    -- 审计字段
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_leave_quotas_unique ON hrm_leave_quotas(tenant_id, employee_id, leave_type_id, year);
CREATE INDEX IF NOT EXISTS idx_leave_quotas_employee ON hrm_leave_quotas(employee_id);
CREATE INDEX IF NOT EXISTS idx_leave_quotas_type ON hrm_leave_quotas(leave_type_id);

COMMENT ON TABLE hrm_leave_quotas IS '请假额度表';
COMMENT ON COLUMN hrm_leave_quotas.used_days IS '已使用天数，默认 0';
COMMENT ON COLUMN hrm_leave_quotas.carried_days IS '结转天数，默认 0';

-- =============================================================================
-- 11. 加班记录表 (Overtime Records)
-- =============================================================================
CREATE TABLE IF NOT EXISTS hrm_overtimes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    
    -- 申请人信息（关联 organization.employees）
    employee_id UUID NOT NULL REFERENCES employees(id),  -- 对应 organization.employees.id
    employee_name VARCHAR(100),       -- 冗余
    department_id UUID REFERENCES organizations(id),  -- 冗余
    
    -- 加班时间
    start_time TIMESTAMP NOT NULL,    -- 开始时间
    end_time TIMESTAMP NOT NULL,      -- 结束时间
    duration DECIMAL(5,2) NOT NULL,   -- 加班时长（小时）
    
    -- 加班类型
    overtime_type VARCHAR(20) NOT NULL,  -- workday, weekend, holiday
    pay_type VARCHAR(20) NOT NULL,    -- money, leave（调休）
    
    -- 加班倍率（根据劳动法）
    pay_rate DECIMAL(3,2) NOT NULL,   -- 工作日1.5倍，周末2倍，节假日3倍
    
    -- 加班原因
    reason TEXT NOT NULL,
    tasks TEXT[],                     -- 加班任务
    
    -- 审批信息
    approval_id UUID,                 -- 关联审批流程
    approval_status VARCHAR(20) DEFAULT 'pending',  -- pending, approved, rejected
    approved_by UUID,
    approved_at TIMESTAMP,
    reject_reason TEXT,
    
    -- 调休信息
    comp_off_days DECIMAL(5,2) DEFAULT 0,      -- 可调休天数
    comp_off_used DECIMAL(5,2) DEFAULT 0,      -- 已调休天数
    comp_off_expire_at TIMESTAMP,     -- 调休过期时间
    
    -- 备注
    remark TEXT,
    
    -- 审计字段
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_overtimes_tenant ON hrm_overtimes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_overtimes_employee ON hrm_overtimes(employee_id);
CREATE INDEX IF NOT EXISTS idx_overtimes_type ON hrm_overtimes(overtime_type);
CREATE INDEX IF NOT EXISTS idx_overtimes_status ON hrm_overtimes(approval_status);
CREATE INDEX IF NOT EXISTS idx_overtimes_time ON hrm_overtimes(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_overtimes_department ON hrm_overtimes(department_id);

COMMENT ON TABLE hrm_overtimes IS '加班记录表';
COMMENT ON COLUMN hrm_overtimes.overtime_type IS '加班类型: workday(工作日), weekend(周末), holiday(节假日)';
COMMENT ON COLUMN hrm_overtimes.pay_type IS '补偿方式: money(加班费), leave(调休)';
COMMENT ON COLUMN hrm_overtimes.pay_rate IS '加班倍率: 工作日1.5, 周末2.0, 节假日3.0';
COMMENT ON COLUMN hrm_overtimes.comp_off_days IS '可调休天数，默认 0';
COMMENT ON COLUMN hrm_overtimes.comp_off_used IS '已调休天数，默认 0';

-- 待续...（下一部分包含出差、考勤汇总、设备管理等表）
