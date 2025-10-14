-- HRM (Human Resource Management) Module Tables
-- 人力资源管理模块 - 考勤系统

-- =============================================================================
-- 1. HRM 员工扩展表 (HRM Employee Extensions)
-- =============================================================================
-- 扩展 organization.employees 表，添加考勤系统所需的专属字段
CREATE TABLE IF NOT EXISTS hrm_employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    
    -- 身份信息
    id_card_no VARCHAR(50),  -- 身份证号（加密存储）
    
    -- 考勤设备信息
    card_no VARCHAR(50),     -- 考勤卡号
    face_data TEXT,          -- 人脸特征数据（加密存储）
    fingerprint TEXT,        -- 指纹数据（加密存储）
    
    -- 第三方平台映射
    dingtalk_user_id VARCHAR(100),   -- 钉钉 UserID
    wecom_user_id VARCHAR(100),      -- 企业微信 UserID
    feishu_user_id VARCHAR(100),     -- 飞书 UserID
    feishu_open_id VARCHAR(100),     -- 飞书 OpenID
    
    -- 工作信息
    work_location VARCHAR(200),      -- 工作地点
    work_schedule_type VARCHAR(50),  -- 工作时间表类型：weekday, shift, flexible
    attendance_rule_id UUID REFERENCES hrm_attendance_rules(id),  -- 关联考勤规则
    default_shift_id UUID REFERENCES hrm_shifts(id),              -- 默认班次
    
    -- 考勤设置
    allow_field_work BOOLEAN DEFAULT FALSE,   -- 是否允许外勤打卡
    require_face BOOLEAN DEFAULT FALSE,       -- 是否必须人脸识别
    require_location BOOLEAN DEFAULT FALSE,   -- 是否必须定位
    require_wifi BOOLEAN DEFAULT FALSE,       -- 是否必须WiFi
    
    -- 紧急联系人
    emergency_contact VARCHAR(100),    -- 紧急联系人姓名
    emergency_phone VARCHAR(20),       -- 紧急联系人电话
    emergency_relation VARCHAR(50),    -- 与紧急联系人关系
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,    -- 是否启用考勤
    
    -- 备注
    remark TEXT,
    
    -- 审计字段
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_hrm_employees_employee ON hrm_employees(tenant_id, employee_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_hrm_employees_card_no ON hrm_employees(tenant_id, card_no) WHERE deleted_at IS NULL AND card_no IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_hrm_employees_tenant ON hrm_employees(tenant_id);
CREATE INDEX IF NOT EXISTS idx_hrm_employees_rule ON hrm_employees(attendance_rule_id);
CREATE INDEX IF NOT EXISTS idx_hrm_employees_shift ON hrm_employees(default_shift_id);
CREATE INDEX IF NOT EXISTS idx_hrm_employees_active ON hrm_employees(is_active);
CREATE INDEX IF NOT EXISTS idx_hrm_employees_dingtalk ON hrm_employees(dingtalk_user_id) WHERE dingtalk_user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_hrm_employees_wecom ON hrm_employees(wecom_user_id) WHERE wecom_user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_hrm_employees_feishu ON hrm_employees(feishu_user_id) WHERE feishu_user_id IS NOT NULL;

COMMENT ON TABLE hrm_employees IS 'HRM员工扩展信息表';
COMMENT ON COLUMN hrm_employees.employee_id IS '关联 employees 表的员工ID';
COMMENT ON COLUMN hrm_employees.card_no IS '考勤卡号，全局唯一';
COMMENT ON COLUMN hrm_employees.is_active IS '是否启用考勤，默认 true';

-- =============================================================================
-- 2. 员工第三方平台同步映射表 (Employee Sync Mappings)
-- =============================================================================
-- 记录员工在各个第三方平台的账号映射关系和同步状态
CREATE TABLE IF NOT EXISTS hrm_employee_sync_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    
    -- 平台信息
    platform VARCHAR(50) NOT NULL,   -- dingtalk, wecom, feishu
    platform_id VARCHAR(100) NOT NULL,  -- 第三方平台的用户ID
    
    -- 同步信息
    sync_enabled BOOLEAN DEFAULT TRUE,  -- 是否启用同步
    last_sync_at TIMESTAMP,            -- 最后同步时间
    sync_status VARCHAR(20),           -- success, failed
    sync_error TEXT,                   -- 同步错误信息
    
    -- 映射数据（原始数据，JSON格式）
    raw_data JSONB,
    
    -- 审计字段
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_employee_sync_platform ON hrm_employee_sync_mappings(tenant_id, employee_id, platform);
CREATE UNIQUE INDEX IF NOT EXISTS idx_employee_sync_platform_id ON hrm_employee_sync_mappings(tenant_id, platform, platform_id);
CREATE INDEX IF NOT EXISTS idx_employee_sync_employee ON hrm_employee_sync_mappings(employee_id);
CREATE INDEX IF NOT EXISTS idx_employee_sync_tenant ON hrm_employee_sync_mappings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_employee_sync_enabled ON hrm_employee_sync_mappings(sync_enabled);

COMMENT ON TABLE hrm_employee_sync_mappings IS '员工第三方平台同步映射表';
COMMENT ON COLUMN hrm_employee_sync_mappings.platform IS '平台类型: dingtalk, wecom, feishu';

-- =============================================================================
-- 3. 员工工作时间表 (Employee Work Schedules)
-- =============================================================================
-- 记录员工的工作安排（如固定班次、轮班等）
CREATE TABLE IF NOT EXISTS hrm_employee_work_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    
    -- 时间表类型
    schedule_type VARCHAR(50) NOT NULL,  -- weekday, shift, flexible, custom
    
    -- 工作日配置（适用于 weekday 类型）
    work_days INTEGER[],      -- 工作日：1-7 (1=周一)
    work_hours INTEGER,       -- 每日工作小时数
    work_start VARCHAR(5),    -- 标准上班时间 HH:MM
    work_end VARCHAR(5),      -- 标准下班时间 HH:MM
    
    -- 轮班配置（适用于 shift 类型）
    shift_cycle INTEGER,      -- 轮班周期（天）
    shift_pattern VARCHAR(100),  -- 轮班模式（如：早中晚休）
    
    -- 生效时间
    effective_from TIMESTAMP NOT NULL,  -- 生效开始时间
    effective_to TIMESTAMP,             -- 生效结束时间
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,  -- 是否启用
    
    -- 备注
    remark TEXT,
    
    -- 审计字段
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_employee_schedules_employee ON hrm_employee_work_schedules(tenant_id, employee_id);
CREATE INDEX IF NOT EXISTS idx_employee_schedules_active ON hrm_employee_work_schedules(employee_id, is_active);
CREATE INDEX IF NOT EXISTS idx_employee_schedules_effective ON hrm_employee_work_schedules(effective_from, effective_to);

COMMENT ON TABLE hrm_employee_work_schedules IS '员工工作时间表';
COMMENT ON COLUMN hrm_employee_work_schedules.schedule_type IS '时间表类型: weekday(标准), shift(轮班), flexible(弹性), custom(自定义)';

-- =============================================================================
-- 4. 班次表 (Shifts)
-- =============================================================================
-- 定义工作时间段（固定班次、弹性班次、自由班次）
CREATE TABLE IF NOT EXISTS hrm_shifts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    
    -- 班次基本信息
    code VARCHAR(50) NOT NULL,        -- 班次编码（唯一）
    name VARCHAR(100) NOT NULL,       -- 班次名称
    description TEXT,                 -- 描述
    
    -- 班次类型
    type VARCHAR(20) NOT NULL,        -- fixed, flexible, free
    
    -- 固定班次配置
    work_start VARCHAR(5),            -- 上班时间 HH:MM
    work_end VARCHAR(5),              -- 下班时间 HH:MM
    
    -- 弹性班次配置
    flexible_start VARCHAR(5),        -- 弹性上班开始时间
    flexible_end VARCHAR(5),          -- 弹性上班结束时间
    work_duration INTEGER,            -- 工作时长（分钟）
    
    -- 打卡规则
    check_in_required BOOLEAN DEFAULT TRUE,   -- 是否必须上班打卡
    check_out_required BOOLEAN DEFAULT TRUE,  -- 是否必须下班打卡
    
    -- 迟到早退规则
    late_grace_period INTEGER DEFAULT 0,      -- 迟到宽限时间（分钟）
    early_grace_period INTEGER DEFAULT 0,     -- 早退宽限时间（分钟）
    
    -- 休息时间
    rest_periods JSONB,               -- 休息时间段数组
    
    -- 跨天标识
    is_cross_days BOOLEAN DEFAULT FALSE,  -- 是否跨天班次（如夜班）
    
    -- 加班规则
    allow_overtime BOOLEAN DEFAULT TRUE,      -- 是否允许加班
    overtime_start_buffer INTEGER DEFAULT 0,  -- 加班开始缓冲（分钟）
    overtime_min_duration INTEGER DEFAULT 60, -- 最小加班时长（分钟）
    overtime_pay_rate DECIMAL(4,2) DEFAULT 1.5,  -- 加班倍率
    
    -- 工作日类型
    workday_types TEXT[],             -- workday, weekend, holiday
    
    -- 颜色标识（用于排班表显示）
    color VARCHAR(20),
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,   -- 是否启用
    sort INTEGER DEFAULT 0,           -- 排序
    
    -- 审计字段
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_shifts_code_tenant ON hrm_shifts(tenant_id, code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_shifts_tenant ON hrm_shifts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_shifts_type ON hrm_shifts(type);
CREATE INDEX IF NOT EXISTS idx_shifts_active ON hrm_shifts(is_active);

COMMENT ON TABLE hrm_shifts IS '班次表';
COMMENT ON COLUMN hrm_shifts.type IS '班次类型: fixed(固定), flexible(弹性), free(自由)';
COMMENT ON COLUMN hrm_shifts.late_grace_period IS '迟到宽限时间（分钟），默认 0';
COMMENT ON COLUMN hrm_shifts.early_grace_period IS '早退宽限时间（分钟），默认 0';

-- =============================================================================
-- 5. 排班表 (Schedules)
-- =============================================================================
-- 员工排班记录
CREATE TABLE IF NOT EXISTS hrm_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    
    -- 员工信息
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    employee_name VARCHAR(100),       -- 冗余
    department_id UUID REFERENCES organizations(id),  -- 冗余
    
    -- 班次信息
    shift_id UUID NOT NULL REFERENCES hrm_shifts(id),
    shift_name VARCHAR(100),          -- 冗余
    
    -- 排班日期
    schedule_date DATE NOT NULL,      -- 排班日期
    workday_type VARCHAR(20),         -- workday, weekend, holiday
    
    -- 状态
    status VARCHAR(20) DEFAULT 'draft',  -- draft, published, executed
    
    -- 备注
    remark TEXT,
    
    -- 审计字段
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_schedules_employee_date ON hrm_schedules(tenant_id, employee_id, schedule_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_schedules_tenant ON hrm_schedules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_schedules_employee ON hrm_schedules(employee_id);
CREATE INDEX IF NOT EXISTS idx_schedules_shift ON hrm_schedules(shift_id);
CREATE INDEX IF NOT EXISTS idx_schedules_department ON hrm_schedules(department_id);
CREATE INDEX IF NOT EXISTS idx_schedules_date ON hrm_schedules(schedule_date);
CREATE INDEX IF NOT EXISTS idx_schedules_status ON hrm_schedules(status);

COMMENT ON TABLE hrm_schedules IS '排班表';
COMMENT ON COLUMN hrm_schedules.status IS '状态: draft(草稿), published(已发布), executed(已执行)，默认 draft';

-- =============================================================================
-- 6. 考勤规则表 (Attendance Rules / 考勤组)
-- =============================================================================
-- 定义考勤规则（考勤组）
CREATE TABLE IF NOT EXISTS hrm_attendance_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    
    -- 规则基本信息
    code VARCHAR(50) NOT NULL,        -- 规则编码
    name VARCHAR(100) NOT NULL,       -- 规则名称
    description TEXT,                 -- 描述
    
    -- 适用范围
    apply_type VARCHAR(20) NOT NULL,  -- all, department, employee
    department_ids UUID[],            -- 适用部门
    employee_ids UUID[],              -- 适用员工
    
    -- 工作制
    workday_type VARCHAR(20) NOT NULL,  -- five_day, six_day, custom
    weekend_days INTEGER[],           -- 0=周日, 1=周一, ..., 6=周六
    
    -- 默认班次
    default_shift_id UUID REFERENCES hrm_shifts(id),
    
    -- 打卡位置限制
    location_required BOOLEAN DEFAULT FALSE,  -- 是否必须定位
    allowed_locations JSONB,          -- 允许的打卡位置数组
    
    -- WiFi限制
    wifi_required BOOLEAN DEFAULT FALSE,  -- 是否必须连接指定WiFi
    allowed_wifi TEXT[],              -- 允许的WiFi列表（SSID或MAC）
    
    -- 人脸识别
    face_required BOOLEAN DEFAULT FALSE,      -- 是否必须人脸识别
    face_threshold DECIMAL(3,2) DEFAULT 0.80,  -- 人脸识别阈值
    face_anti_spoofing BOOLEAN DEFAULT FALSE,  -- 是否开启活体检测
    
    -- 外勤打卡
    allow_field_work BOOLEAN DEFAULT FALSE,  -- 是否允许外勤打卡
    
    -- 节假日设置
    holiday_calendar_id UUID,         -- 关联假期日历
    
    -- 审批设置
    require_approval_for_late BOOLEAN DEFAULT FALSE,   -- 迟到需要审批
    require_approval_for_early BOOLEAN DEFAULT FALSE,  -- 早退需要审批
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,   -- 是否启用
    priority INTEGER DEFAULT 0,       -- 优先级，数字越大优先级越高
    
    -- 审计字段
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_attendance_rules_code_tenant ON hrm_attendance_rules(tenant_id, code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_attendance_rules_tenant ON hrm_attendance_rules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_attendance_rules_active ON hrm_attendance_rules(is_active);
CREATE INDEX IF NOT EXISTS idx_attendance_rules_priority ON hrm_attendance_rules(priority DESC);

COMMENT ON TABLE hrm_attendance_rules IS '考勤规则表（考勤组）';
COMMENT ON COLUMN hrm_attendance_rules.apply_type IS '适用类型: all(全员), department(按部门), employee(按员工)';
COMMENT ON COLUMN hrm_attendance_rules.workday_type IS '工作制: five_day(标准5天), six_day(大小周), custom(自定义)';
COMMENT ON COLUMN hrm_attendance_rules.face_threshold IS '人脸识别阈值，默认 0.80';
COMMENT ON COLUMN hrm_attendance_rules.priority IS '优先级，数字越大优先级越高，默认 0';

-- 待续...（下一部分包含考勤记录表等）
