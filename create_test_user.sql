-- 创建测试数据用于通知功能测试

-- 创建测试租户
INSERT INTO tenants (id, name, code, status, created_at, updated_at)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    '测试租户',
    'test_tenant',
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

-- 创建测试用户
INSERT INTO users (id, tenant_id, username, email, password_hash, status, created_at, updated_at)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    'notifyuser',
    'notify@test.com',
    '$2a$10$6LJzmH8qKJOFQCY.Pz5kbuZ.bI8Y9YWx4qmVr1qIo6vK5PJ8Xz7iy',  -- Password: Test@123
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

SELECT '测试用户创建成功！' AS message;
SELECT 'Username: notifyuser' AS info;
SELECT 'Password: Test@123' AS info;
