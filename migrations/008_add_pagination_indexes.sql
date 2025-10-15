-- ==========================================
-- 分页查询性能优化索引
-- ==========================================
-- 目的：优化各模块的分页查询性能
-- 包括：游标分页索引、覆盖索引、部分索引
-- ==========================================

-- ==========================================
-- 1. HRM模块 - 考勤记录优化
-- ==========================================

-- 1.1 游标分页索引（推荐用于大数据量分页）
-- 用途：支持 clock_time 游标分页，性能稳定不受偏移量影响
CREATE INDEX IF NOT EXISTS idx_attendance_cursor 
ON hrm_attendance_records(tenant_id, clock_time DESC, id DESC) 
WHERE deleted_at IS NULL;

-- 1.2 部门考勤查询优化索引
CREATE INDEX IF NOT EXISTS idx_attendance_dept_time 
ON hrm_attendance_records(tenant_id, department_id, clock_time DESC) 
WHERE deleted_at IS NULL;

-- 1.3 员工考勤查询优化索引
CREATE INDEX IF NOT EXISTS idx_attendance_emp_time 
ON hrm_attendance_records(tenant_id, employee_id, clock_time DESC) 
WHERE deleted_at IS NULL;

-- 1.4 异常考勤查询优化索引（部分索引，只索引异常记录）
CREATE INDEX IF NOT EXISTS idx_attendance_exceptions 
ON hrm_attendance_records(tenant_id, clock_time DESC) 
WHERE deleted_at IS NULL AND is_exception = TRUE;

-- 1.5 考勤状态统计优化索引
CREATE INDEX IF NOT EXISTS idx_attendance_status_time 
ON hrm_attendance_records(tenant_id, status, clock_time) 
WHERE deleted_at IS NULL;

-- ==========================================
-- 2. HRM模块 - 班次管理优化
-- ==========================================

-- 2.1 班次列表游标分页索引
CREATE INDEX IF NOT EXISTS idx_shifts_cursor 
ON hrm_shifts(tenant_id, created_at DESC, id DESC) 
WHERE deleted_at IS NULL;

-- 2.2 有效班次查询优化（部分索引）
CREATE INDEX IF NOT EXISTS idx_shifts_active 
ON hrm_shifts(tenant_id, type, sort ASC) 
WHERE deleted_at IS NULL AND is_active = TRUE;

-- ==========================================
-- 3. HRM模块 - 排班管理优化
-- ==========================================

-- 3.1 员工排班查询优化索引
CREATE INDEX IF NOT EXISTS idx_schedules_employee 
ON hrm_schedules(tenant_id, employee_id, schedule_date DESC) 
WHERE deleted_at IS NULL;

-- 3.2 部门排班查询优化索引
CREATE INDEX IF NOT EXISTS idx_schedules_department 
ON hrm_schedules(tenant_id, department_id, schedule_date DESC) 
WHERE deleted_at IS NULL;

-- 3.3 排班状态查询优化
CREATE INDEX IF NOT EXISTS idx_schedules_status 
ON hrm_schedules(tenant_id, status, schedule_date DESC) 
WHERE deleted_at IS NULL;

-- ==========================================
-- 4. HRM模块 - 员工管理优化
-- ==========================================

-- 4.1 HRM员工列表游标分页索引
CREATE INDEX IF NOT EXISTS idx_hrm_employees_cursor 
ON hrm_employees(tenant_id, created_at DESC, id DESC) 
WHERE deleted_at IS NULL;

-- 4.2 部门员工查询优化
CREATE INDEX IF NOT EXISTS idx_hrm_employees_dept 
ON hrm_employees(tenant_id, department_id, employee_status) 
WHERE deleted_at IS NULL;

-- 4.3 在职员工查询优化（部分索引）
CREATE INDEX IF NOT EXISTS idx_hrm_employees_active 
ON hrm_employees(tenant_id, department_id, created_at DESC) 
WHERE deleted_at IS NULL AND employee_status = 'active';

-- ==========================================
-- 5. 文件模块优化
-- ==========================================

-- 5.1 文件列表游标分页索引
CREATE INDEX IF NOT EXISTS idx_files_cursor 
ON files(tenant_id, created_at DESC, id DESC) 
WHERE deleted_at IS NULL;

-- 5.2 文件分类查询优化（覆盖索引）
CREATE INDEX IF NOT EXISTS idx_files_category_covering 
ON files(tenant_id, category, created_at DESC) 
INCLUDE (id, filename, size, mime_type, status)
WHERE deleted_at IS NULL;

-- 5.3 用户文件查询优化
CREATE INDEX IF NOT EXISTS idx_files_uploader 
ON files(tenant_id, uploaded_by, created_at DESC) 
WHERE deleted_at IS NULL;

-- 5.4 临时文件清理优化（部分索引）
CREATE INDEX IF NOT EXISTS idx_files_temporary 
ON files(expires_at ASC) 
WHERE deleted_at IS NULL AND is_temporary = TRUE AND expires_at IS NOT NULL;

-- ==========================================
-- 6. 认证模块优化
-- ==========================================

-- 6.1 用户列表游标分页索引（覆盖索引）
CREATE INDEX IF NOT EXISTS idx_users_cursor_covering 
ON users(tenant_id, created_at DESC, id DESC) 
INCLUDE (username, email, status)
WHERE deleted_at IS NULL;

-- 6.2 有效用户查询优化（部分索引）
CREATE INDEX IF NOT EXISTS idx_users_active 
ON users(tenant_id, status, created_at DESC) 
WHERE deleted_at IS NULL AND status = 'active';

-- 6.3 审计日志时间范围查询优化
CREATE INDEX IF NOT EXISTS idx_audit_logs_time_range 
ON audit_logs(tenant_id, created_at DESC, id DESC);

-- 6.4 用户审计日志查询优化
CREATE INDEX IF NOT EXISTS idx_audit_logs_user 
ON audit_logs(tenant_id, user_id, created_at DESC);

-- ==========================================
-- 7. 组织架构模块优化
-- ==========================================

-- 7.1 组织列表游标分页索引
CREATE INDEX IF NOT EXISTS idx_organizations_cursor 
ON organizations(tenant_id, created_at DESC, id DESC) 
WHERE deleted_at IS NULL;

-- 7.2 组织树查询优化（层级+排序）
CREATE INDEX IF NOT EXISTS idx_organizations_tree 
ON organizations(tenant_id, parent_id, level ASC, sort ASC) 
WHERE deleted_at IS NULL;

-- 7.3 有效组织查询（部分索引）
CREATE INDEX IF NOT EXISTS idx_organizations_active 
ON organizations(tenant_id, type_id, status) 
WHERE deleted_at IS NULL AND status = 'active';

-- 7.4 员工列表游标分页索引
CREATE INDEX IF NOT EXISTS idx_employees_cursor 
ON employees(tenant_id, created_at DESC, id DESC) 
WHERE deleted_at IS NULL;

-- 7.5 部门员工查询优化
CREATE INDEX IF NOT EXISTS idx_employees_org 
ON employees(tenant_id, org_id, status, created_at DESC) 
WHERE deleted_at IS NULL;

-- ==========================================
-- 8. 审批模块优化
-- ==========================================

-- 8.1 审批实例列表游标分页
CREATE INDEX IF NOT EXISTS idx_process_instances_cursor 
ON process_instances(tenant_id, created_at DESC, id DESC) 
WHERE deleted_at IS NULL;

-- 8.2 申请人审批查询优化
CREATE INDEX IF NOT EXISTS idx_process_instances_applicant 
ON process_instances(tenant_id, applicant_id, status, created_at DESC) 
WHERE deleted_at IS NULL;

-- 8.3 审批任务待办查询优化（部分索引）
CREATE INDEX IF NOT EXISTS idx_approval_tasks_pending 
ON approval_tasks(tenant_id, assignee_id, due_date ASC) 
WHERE deleted_at IS NULL AND status = 'pending';

-- 8.4 审批任务历史查询
CREATE INDEX IF NOT EXISTS idx_approval_tasks_history 
ON approval_tasks(tenant_id, assignee_id, completed_at DESC) 
WHERE deleted_at IS NULL AND status IN ('approved', 'rejected');

-- ==========================================
-- 9. 通知模块优化
-- ==========================================

-- 9.1 用户通知列表游标分页
CREATE INDEX IF NOT EXISTS idx_notifications_cursor 
ON notifications(recipient_id, created_at DESC, id DESC) 
WHERE deleted_at IS NULL;

-- 9.2 未读通知查询优化（部分索引）
CREATE INDEX IF NOT EXISTS idx_notifications_unread 
ON notifications(recipient_id, created_at DESC) 
WHERE deleted_at IS NULL AND read_at IS NULL;

-- 9.3 通知类型查询优化
CREATE INDEX IF NOT EXISTS idx_notifications_type 
ON notifications(tenant_id, type, created_at DESC) 
WHERE deleted_at IS NULL;

-- ==========================================
-- 10. 索引维护建议
-- ==========================================

-- 10.1 定期分析表统计信息（建议每天执行）
-- ANALYZE hrm_attendance_records;
-- ANALYZE files;
-- ANALYZE users;
-- ANALYZE organizations;

-- 10.2 定期清理无用索引（执行前请备份）
-- 查询未使用的索引：
-- SELECT schemaname, tablename, indexname, idx_scan
-- FROM pg_stat_user_indexes
-- WHERE idx_scan = 0
--   AND indexrelname NOT LIKE 'pg_toast%'
-- ORDER BY pg_relation_size(indexrelid) DESC;

-- 10.3 定期重建索引（可选，通常不需要）
-- REINDEX INDEX CONCURRENTLY idx_attendance_cursor;

-- ==========================================
-- 注释说明
-- ==========================================

COMMENT ON INDEX idx_attendance_cursor IS '考勤记录游标分页索引，用于高性能分页查询';
COMMENT ON INDEX idx_attendance_exceptions IS '异常考勤查询优化索引（部分索引）';
COMMENT ON INDEX idx_files_cursor IS '文件列表游标分页索引';
COMMENT ON INDEX idx_users_cursor_covering IS '用户列表覆盖索引，避免回表查询';
COMMENT ON INDEX idx_notifications_unread IS '未读通知查询优化索引（部分索引）';
