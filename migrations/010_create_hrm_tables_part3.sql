-- HRM Module Tables (Part 3)
-- 出差、考勤汇总、设备管理、第三方集成等表

-- =============================================================================
-- 12. 出差记录表 (Business Trip Records)
-- =============================================================================
CREATE TABLE IF NOT EXISTS hrm_business_trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    
    -- 申请人信息（关联 organization.employees）
    employee_id UUID NOT NULL REFERENCES employees(id),  -- 对应 organization.employees.id
    employee_name VARCHAR(100),       -- 冗余
    department_id UUID REFERENCES organizations(id),  -- 冗余
    
    -- 出差时间
    start_time TIMESTAMP NOT NULL,    -- 开始时间
    end_time TIMESTAMP NOT NULL,      -- 结束时间
    duration DECIMAL(5,2) NOT NULL,   -- 出差天数
    
    -- 出差地点
    destination VARCHAR(200) NOT NULL,  -- 目的地
    transportation VARCHAR(100),      -- 交通方式
    accommodation VARCHAR(200),       -- 住宿安排
    companions UUID[],                -- 同行人员ID数组
    
    -- 出差原因
    purpose TEXT NOT NULL,
    tasks TEXT,
    
    -- 预算信息
    estimated_cost DECIMAL(10,2) DEFAULT 0,  -- 预计费用
    actual_cost DECIMAL(10,2) DEFAULT 0,     -- 实际费用
    
    -- 审批信息
    approval_id UUID,                 -- 关联审批流程
    approval_status VARCHAR(20) DEFAULT 'pending',  -- pending, approved, rejected
    approved_by UUID,
    approved_at TIMESTAMP,
    reject_reason TEXT,
    
    -- 出差报告
    report TEXT,
    report_at TIMESTAMP,
    
    -- 备注
    remark TEXT,
    
    -- 审计字段
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_business_trips_tenant ON hrm_business_trips(tenant_id);
CREATE INDEX IF NOT EXISTS idx_business_trips_employee ON hrm_business_trips(employee_id);
CREATE INDEX IF NOT EXISTS idx_business_trips_status ON hrm_business_trips(approval_status);
CREATE INDEX IF NOT EXISTS idx_business_trips_time ON hrm_business_trips(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_business_trips_department ON hrm_business_trips(department_id);

COMMENT ON TABLE hrm_business_trips IS '出差记录表';
COMMENT ON COLUMN hrm_business_trips.estimated_cost IS '预计费用，默认 0';
COMMENT ON COLUMN hrm_business_trips.actual_cost IS '实际费用，默认 0';

-- =============================================================================
-- 13. 考勤汇总表 (Attendance Summaries)
-- =============================================================================
-- 按员工按月统计考勤数据
CREATE TABLE IF NOT EXISTS hrm_attendance_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    
    -- 员工信息（关联 organization.employees）
    employee_id UUID NOT NULL REFERENCES employees(id),  -- 对应 organization.employees.id
    employee_name VARCHAR(100),       -- 冗余
    department_id UUID REFERENCES organizations(id),  -- 冗余
    
    -- 统计周期
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    
    -- 出勤统计
    work_days INTEGER DEFAULT 0,      -- 应出勤天数
    actual_days INTEGER DEFAULT 0,    -- 实际出勤天数
    late_count INTEGER DEFAULT 0,     -- 迟到次数
    late_duration INTEGER DEFAULT 0,  -- 迟到总时长（分钟）
    early_count INTEGER DEFAULT 0,    -- 早退次数
    early_duration INTEGER DEFAULT 0, -- 早退总时长（分钟）
    absent_count INTEGER DEFAULT 0,   -- 旷工次数
    absent_days DECIMAL(5,2) DEFAULT 0,  -- 旷工天数
    missing_count INTEGER DEFAULT 0,  -- 缺卡次数
    
    -- 请假统计
    leave_count INTEGER DEFAULT 0,    -- 请假次数
    leave_days DECIMAL(5,2) DEFAULT 0,  -- 请假天数
    
    -- 加班统计
    overtime_count INTEGER DEFAULT 0,    -- 加班次数
    overtime_hours DECIMAL(6,2) DEFAULT 0,  -- 加班小时数
    weekend_ot_hours DECIMAL(6,2) DEFAULT 0,  -- 周末加班小时
    holiday_ot_hours DECIMAL(6,2) DEFAULT 0,  -- 节假日加班小时
    comp_off_days DECIMAL(5,2) DEFAULT 0,     -- 可调休天数
    
    -- 出差统计
    trip_count INTEGER DEFAULT 0,     -- 出差次数
    trip_days DECIMAL(5,2) DEFAULT 0,   -- 出差天数
    
    -- 工时统计
    work_hours DECIMAL(8,2) DEFAULT 0,        -- 工作总时长
    standard_work_hours DECIMAL(8,2) DEFAULT 0,  -- 标准工时
    
    -- 状态
    status VARCHAR(20) DEFAULT 'draft',  -- draft, confirmed, locked
    
    -- 审计字段
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    confirmed_at TIMESTAMP,
    confirmed_by UUID
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_attendance_summaries_unique ON hrm_attendance_summaries(tenant_id, employee_id, year, month);
CREATE INDEX IF NOT EXISTS idx_attendance_summaries_tenant ON hrm_attendance_summaries(tenant_id);
CREATE INDEX IF NOT EXISTS idx_attendance_summaries_employee ON hrm_attendance_summaries(employee_id);
CREATE INDEX IF NOT EXISTS idx_attendance_summaries_period ON hrm_attendance_summaries(year, month);
CREATE INDEX IF NOT EXISTS idx_attendance_summaries_status ON hrm_attendance_summaries(status);
CREATE INDEX IF NOT EXISTS idx_attendance_summaries_department ON hrm_attendance_summaries(department_id);

COMMENT ON TABLE hrm_attendance_summaries IS '考勤汇总表（按员工按月统计）';
COMMENT ON COLUMN hrm_attendance_summaries.status IS '状态: draft(草稿), confirmed(已确认), locked(已锁定)，默认 draft';
COMMENT ON COLUMN hrm_attendance_summaries.work_days IS '应出勤天数，默认 0';
COMMENT ON COLUMN hrm_attendance_summaries.actual_days IS '实际出勤天数，默认 0';

-- =============================================================================
-- 14. 考勤设备表 (Attendance Devices)
-- =============================================================================
-- 考勤机设备管理
CREATE TABLE IF NOT EXISTS hrm_attendance_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    
    -- 设备信息
    device_type VARCHAR(50) NOT NULL,  -- zkteco, dingtalk_m2, deli, hikvision, dahua, other
    device_sn VARCHAR(100) NOT NULL,  -- 设备序列号（唯一标识）
    device_name VARCHAR(100) NOT NULL,  -- 设备名称
    device_model VARCHAR(100),        -- 设备型号
    
    -- 网络信息
    ip_address VARCHAR(50),           -- IP地址
    port INTEGER,                     -- 端口号
    mac_address VARCHAR(50),          -- MAC地址
    
    -- 位置信息
    location JSONB,                   -- 设备位置 {latitude, longitude}
    install_address VARCHAR(500),     -- 安装地址
    department_id UUID REFERENCES organizations(id),  -- 关联部门
    
    -- 认证信息
    auth_type VARCHAR(50),            -- 认证方式：password, apikey, certificate
    username VARCHAR(100),
    password TEXT,                    -- 加密存储
    api_key TEXT,
    secret_key TEXT,
    
    -- 同步配置
    sync_enabled BOOLEAN DEFAULT TRUE,  -- 是否启用同步
    sync_interval INTEGER DEFAULT 15,   -- 同步间隔（分钟）
    sync_mode VARCHAR(20) DEFAULT 'pull',  -- push, pull
    last_sync_at TIMESTAMP,           -- 最后同步时间
    
    -- 功能支持
    support_face BOOLEAN DEFAULT FALSE,         -- 支持人脸识别
    support_fingerprint BOOLEAN DEFAULT FALSE,  -- 支持指纹
    support_card BOOLEAN DEFAULT FALSE,         -- 支持刷卡
    support_temperature BOOLEAN DEFAULT FALSE,  -- 支持体温检测
    
    -- 设备状态
    status VARCHAR(20) DEFAULT 'offline',  -- online, offline, error
    is_active BOOLEAN DEFAULT TRUE,   -- 是否启用
    last_heartbeat TIMESTAMP,         -- 最后心跳时间
    error_message TEXT,               -- 错误信息
    
    -- 统计信息
    total_records INTEGER DEFAULT 0,  -- 总记录数
    today_records INTEGER DEFAULT 0,  -- 今日记录数
    
    -- 备注
    remark TEXT,
    
    -- 审计字段
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_sn_tenant ON hrm_attendance_devices(tenant_id, device_sn) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_devices_tenant ON hrm_attendance_devices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_devices_type ON hrm_attendance_devices(device_type);
CREATE INDEX IF NOT EXISTS idx_devices_status ON hrm_attendance_devices(status);
CREATE INDEX IF NOT EXISTS idx_devices_active ON hrm_attendance_devices(is_active);
CREATE INDEX IF NOT EXISTS idx_devices_department ON hrm_attendance_devices(department_id);

COMMENT ON TABLE hrm_attendance_devices IS '考勤设备表（考勤机）';
COMMENT ON COLUMN hrm_attendance_devices.device_type IS '设备类型: zkteco(中控智慧), dingtalk_m2(钉钉M2), deli(得力), hikvision(海康), dahua(大华), other(其他)';
COMMENT ON COLUMN hrm_attendance_devices.sync_enabled IS '是否启用同步，默认 true';
COMMENT ON COLUMN hrm_attendance_devices.sync_interval IS '同步间隔（分钟），默认 15';
COMMENT ON COLUMN hrm_attendance_devices.sync_mode IS '同步模式: push(推送), pull(拉取)，默认 pull';
COMMENT ON COLUMN hrm_attendance_devices.status IS '设备状态: online(在线), offline(离线), error(异常)，默认 offline';

-- =============================================================================
-- 15. 第三方集成配置表 (Third Party Integrations)
-- =============================================================================
-- 钉钉、企业微信、飞书等平台集成配置
CREATE TABLE IF NOT EXISTS hrm_third_party_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    
    -- 平台信息
    platform VARCHAR(50) NOT NULL,    -- dingtalk, wecom, feishu
    app_name VARCHAR(100) NOT NULL,   -- 应用名称
    app_id VARCHAR(100) NOT NULL,     -- 应用ID
    app_key VARCHAR(200) NOT NULL,    -- AppKey
    app_secret TEXT NOT NULL,         -- AppSecret（加密存储）
    
    -- 企业信息
    corp_id VARCHAR(100),             -- 企业ID
    agent_id VARCHAR(100),            -- 应用AgentID（企微）
    suite_key VARCHAR(100),           -- 套件Key（飞书）
    suite_secret TEXT,                -- 套件Secret（飞书）
    
    -- Webhook配置
    webhook_url VARCHAR(500),         -- Webhook接收地址
    webhook_token VARCHAR(200),       -- Webhook验证Token
    webhook_secret TEXT,              -- Webhook加密Secret
    
    -- 同步配置
    sync_enabled BOOLEAN DEFAULT TRUE,        -- 是否启用同步
    sync_attendance BOOLEAN DEFAULT TRUE,     -- 同步考勤记录
    sync_employee BOOLEAN DEFAULT FALSE,      -- 同步员工信息
    sync_department BOOLEAN DEFAULT FALSE,    -- 同步部门信息
    sync_schedule BOOLEAN DEFAULT FALSE,      -- 同步排班
    sync_interval INTEGER DEFAULT 30,         -- 同步间隔（分钟）
    sync_direction VARCHAR(20) DEFAULT 'pull',  -- both, pull, push
    last_sync_at TIMESTAMP,           -- 最后同步时间
    
    -- 字段映射配置
    field_mapping JSONB,              -- 字段映射关系 JSON
    
    -- 状态
    status VARCHAR(20) DEFAULT 'inactive',  -- active, inactive, error
    is_active BOOLEAN DEFAULT TRUE,   -- 是否启用
    error_message TEXT,               -- 错误信息
    
    -- 统计信息
    total_sync_count INTEGER DEFAULT 0,  -- 总同步次数
    last_sync_count INTEGER DEFAULT 0,   -- 最后一次同步数量
    
    -- 备注
    remark TEXT,
    
    -- 审计字段
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_integrations_platform_tenant ON hrm_third_party_integrations(tenant_id, platform) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_integrations_tenant ON hrm_third_party_integrations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_integrations_platform ON hrm_third_party_integrations(platform);
CREATE INDEX IF NOT EXISTS idx_integrations_status ON hrm_third_party_integrations(status);
CREATE INDEX IF NOT EXISTS idx_integrations_active ON hrm_third_party_integrations(is_active);

COMMENT ON TABLE hrm_third_party_integrations IS '第三方平台集成配置表';
COMMENT ON COLUMN hrm_third_party_integrations.platform IS '平台类型: dingtalk(钉钉), wecom(企业微信), feishu(飞书)';
COMMENT ON COLUMN hrm_third_party_integrations.sync_enabled IS '是否启用同步，默认 true';
COMMENT ON COLUMN hrm_third_party_integrations.sync_interval IS '同步间隔（分钟），默认 30';
COMMENT ON COLUMN hrm_third_party_integrations.sync_direction IS '同步方向: both(双向), pull(拉取), push(推送)，默认 pull';
COMMENT ON COLUMN hrm_third_party_integrations.status IS '状态: active(正常), inactive(未激活), error(异常)，默认 inactive';

-- =============================================================================
-- 16. 同步日志表 (Sync Logs)
-- =============================================================================
-- 记录设备和第三方平台的数据同步日志
CREATE TABLE IF NOT EXISTS hrm_sync_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    
    -- 同步来源
    source_type VARCHAR(50) NOT NULL,  -- device, dingtalk, wecom, feishu
    source_id UUID NOT NULL,           -- 设备或集成配置ID
    
    -- 同步信息
    sync_type VARCHAR(50) NOT NULL,    -- attendance, employee, department
    sync_direction VARCHAR(20) NOT NULL,  -- pull, push
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    duration INTEGER NOT NULL,         -- 毫秒
    
    -- 同步结果
    status VARCHAR(20) NOT NULL,       -- success, failed, partial
    total_count INTEGER DEFAULT 0,     -- 总数
    success_count INTEGER DEFAULT 0,   -- 成功数
    failed_count INTEGER DEFAULT 0,    -- 失败数
    error_message TEXT,                -- 错误信息
    
    -- 详细信息
    details JSONB,
    
    -- 审计字段
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sync_logs_tenant ON hrm_sync_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sync_logs_source ON hrm_sync_logs(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_sync_logs_type ON hrm_sync_logs(sync_type);
CREATE INDEX IF NOT EXISTS idx_sync_logs_status ON hrm_sync_logs(status);
CREATE INDEX IF NOT EXISTS idx_sync_logs_time ON hrm_sync_logs(start_time DESC);

COMMENT ON TABLE hrm_sync_logs IS '同步日志表';
COMMENT ON COLUMN hrm_sync_logs.source_type IS '同步来源: device(设备), dingtalk(钉钉), wecom(企微), feishu(飞书)';
COMMENT ON COLUMN hrm_sync_logs.sync_type IS '同步类型: attendance(考勤记录), employee(员工信息), department(部门信息)';
COMMENT ON COLUMN hrm_sync_logs.sync_direction IS '同步方向: pull(拉取), push(推送)';
COMMENT ON COLUMN hrm_sync_logs.status IS '同步状态: success(成功), failed(失败), partial(部分成功)';
COMMENT ON COLUMN hrm_sync_logs.total_count IS '总数，默认 0';
COMMENT ON COLUMN hrm_sync_logs.success_count IS '成功数，默认 0';
COMMENT ON COLUMN hrm_sync_logs.failed_count IS '失败数，默认 0';

-- =============================================================================
-- 创建视图和函数
-- =============================================================================

-- 创建视图：员工完整考勤信息（组织信息 + HRM扩展）
CREATE OR REPLACE VIEW v_hrm_employee_full AS
SELECT 
    e.id AS employee_id,
    e.tenant_id,
    e.user_id,
    e.employee_no,
    e.name,
    e.gender,
    e.mobile,
    e.email,
    e.avatar,
    e.org_id,
    e.org_name,
    e.position_id,
    e.position_name,
    e.status AS employment_status,
    e.entry_date,
    h.id AS hrm_id,
    h.card_no,
    h.work_location,
    h.attendance_rule_id,
    h.default_shift_id,
    h.dingtalk_user_id,
    h.wecom_user_id,
    h.feishu_user_id,
    h.allow_field_work,
    h.require_face,
    h.require_location,
    h.require_wifi,
    h.is_active AS attendance_active,
    CASE WHEN h.face_data IS NOT NULL AND h.face_data != '' THEN TRUE ELSE FALSE END AS has_face_data,
    CASE WHEN h.fingerprint IS NOT NULL AND h.fingerprint != '' THEN TRUE ELSE FALSE END AS has_fingerprint
FROM employees e
LEFT JOIN hrm_employees h ON e.id = h.employee_id AND h.deleted_at IS NULL
WHERE e.deleted_at IS NULL;

COMMENT ON VIEW v_hrm_employee_full IS '员工完整考勤信息视图（组织信息 + HRM扩展）';

-- 创建函数：自动更新 updated_at 字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为所有表创建触发器
CREATE TRIGGER update_hrm_employees_updated_at BEFORE UPDATE ON hrm_employees
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hrm_shifts_updated_at BEFORE UPDATE ON hrm_shifts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hrm_schedules_updated_at BEFORE UPDATE ON hrm_schedules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hrm_attendance_rules_updated_at BEFORE UPDATE ON hrm_attendance_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hrm_attendance_records_updated_at BEFORE UPDATE ON hrm_attendance_records
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hrm_leave_types_updated_at BEFORE UPDATE ON hrm_leave_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hrm_leaves_updated_at BEFORE UPDATE ON hrm_leaves
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hrm_overtimes_updated_at BEFORE UPDATE ON hrm_overtimes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hrm_business_trips_updated_at BEFORE UPDATE ON hrm_business_trips
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hrm_attendance_summaries_updated_at BEFORE UPDATE ON hrm_attendance_summaries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hrm_attendance_devices_updated_at BEFORE UPDATE ON hrm_attendance_devices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_hrm_third_party_integrations_updated_at BEFORE UPDATE ON hrm_third_party_integrations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- 迁移完成
-- =============================================================================
