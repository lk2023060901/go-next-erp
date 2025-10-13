-- 创建通知表
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    recipient_id UUID NOT NULL,
    recipient_email VARCHAR(255),
    recipient_phone VARCHAR(20),
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    data JSONB,
    priority VARCHAR(10) NOT NULL DEFAULT 'normal',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    sent_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    read_at TIMESTAMPTZ,
    related_type VARCHAR(50),
    related_id UUID,
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_notifications_tenant_id ON notifications(tenant_id);
CREATE INDEX idx_notifications_recipient_id ON notifications(recipient_id);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_read_at ON notifications(read_at) WHERE read_at IS NULL;
CREATE INDEX idx_notifications_related ON notifications(related_type, related_id);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);

-- 添加注释
COMMENT ON TABLE notifications IS '通知表';
COMMENT ON COLUMN notifications.id IS '主键';
COMMENT ON COLUMN notifications.tenant_id IS '租户ID';
COMMENT ON COLUMN notifications.type IS '通知类型';
COMMENT ON COLUMN notifications.channel IS '通知渠道';
COMMENT ON COLUMN notifications.recipient_id IS '接收人ID';
COMMENT ON COLUMN notifications.recipient_email IS '接收人邮箱';
COMMENT ON COLUMN notifications.recipient_phone IS '接收人手机';
COMMENT ON COLUMN notifications.title IS '通知标题';
COMMENT ON COLUMN notifications.content IS '通知内容';
COMMENT ON COLUMN notifications.data IS '附加数据（JSON）';
COMMENT ON COLUMN notifications.priority IS '优先级';
COMMENT ON COLUMN notifications.status IS '状态';
COMMENT ON COLUMN notifications.sent_at IS '发送时间';
COMMENT ON COLUMN notifications.delivered_at IS '送达时间';
COMMENT ON COLUMN notifications.read_at IS '已读时间';
COMMENT ON COLUMN notifications.related_type IS '关联类型';
COMMENT ON COLUMN notifications.related_id IS '关联ID';
COMMENT ON COLUMN notifications.error_message IS '错误信息';
COMMENT ON COLUMN notifications.retry_count IS '重试次数';

-- 创建通知模板表
CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    subject VARCHAR(200),
    template TEXT NOT NULL,
    variables JSONB,
    enabled BOOLEAN DEFAULT true,
    created_by UUID NOT NULL,
    updated_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uk_notification_templates_tenant_code UNIQUE (tenant_id, code, deleted_at)
);

-- 创建索引
CREATE INDEX idx_notification_templates_tenant_id ON notification_templates(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_notification_templates_code ON notification_templates(code) WHERE deleted_at IS NULL;
CREATE INDEX idx_notification_templates_enabled ON notification_templates(enabled) WHERE deleted_at IS NULL;

-- 添加注释
COMMENT ON TABLE notification_templates IS '通知模板表';
COMMENT ON COLUMN notification_templates.id IS '主键';
COMMENT ON COLUMN notification_templates.tenant_id IS '租户ID';
COMMENT ON COLUMN notification_templates.code IS '模板编码';
COMMENT ON COLUMN notification_templates.name IS '模板名称';
COMMENT ON COLUMN notification_templates.type IS '通知类型';
COMMENT ON COLUMN notification_templates.channel IS '通知渠道';
COMMENT ON COLUMN notification_templates.subject IS '主题';
COMMENT ON COLUMN notification_templates.template IS '模板内容';
COMMENT ON COLUMN notification_templates.variables IS '变量列表';
COMMENT ON COLUMN notification_templates.enabled IS '是否启用';
COMMENT ON COLUMN notification_templates.created_by IS '创建人';
COMMENT ON COLUMN notification_templates.updated_by IS '更新人';
COMMENT ON COLUMN notification_templates.deleted_at IS '删除时间（软删除）';
