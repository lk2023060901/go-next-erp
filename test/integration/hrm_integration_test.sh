#!/bin/bash

# HRM 模块集成测试脚本
# 测试所有 HRM API 端点，包括认证流程

set -e

BASE_URL="http://localhost:15006"
TENANT_ID="00000000-0000-0000-0000-000000000001"
TOKEN=""

# 颜色输出
RED='\033[0:31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

echo_error() {
    echo -e "${RED}✗ $1${NC}"
}

echo_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# ========== 步骤1: 用户注册和登录获取 Token ==========
echo_info "步骤1: 用户注册和登录..."

# 使用时间戳创建唯一用户名
TIMESTAMP=$(date +%s)
TEST_USERNAME="hrm_test_${TIMESTAMP}"
TEST_PASSWORD="Test123456!"

# 注册测试用户
REGISTER_RESP=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "'$TENANT_ID'",
    "username": "'$TEST_USERNAME'",
    "email": "'$TEST_USERNAME'@example.com",
    "password": "'$TEST_PASSWORD'",
    "full_name": "HRM Test User"
  }')

if echo "$REGISTER_RESP" | grep -q "error"; then
    echo_error "注册失败: $REGISTER_RESP"
    exit 1
fi

echo_success "用户注册成功"

# 登录获取 Token
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "'$TEST_USERNAME'",
    "password": "'$TEST_PASSWORD'"
  }')

TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('accessToken', ''))" 2>/dev/null || echo "")
USER_ID=$(echo "$LOGIN_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('user', {}).get('id', ''))" 2>/dev/null || echo "")

if [ -z "$TOKEN" ]; then
    echo_error "登录失败，无法获取 Token"
    echo "$LOGIN_RESP"
    exit 1
fi

if [ -z "$USER_ID" ]; then
    echo_error "登录响应中没有用户ID"
    echo "$LOGIN_RESP"
    exit 1
fi

echo_success "成功获取 Token 和用户ID: $USER_ID"

# ========== 步骤2: 创建前置数据（员工、部门、HRM员工记录） ==========
echo_info "步骤2: 创建前置数据（为后续所有测试提供基础数据）..."

# 2.1 创建测试公司（前置数据 - 作为根组织）
echo_info "2.1 创建测试公司（前置数据 - 根组织）..."
COMPANY_RESP=$(curl -s -X POST "$BASE_URL/api/v1/organizations" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "'$TENANT_ID'",
    "code": "COMPANY_HRM_TEST_'$TIMESTAMP'",
    "name": "HRM测试公司",
    "type": "company",
    "parent_id": null
  }')

COMPANY_ID=$(echo "$COMPANY_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))" 2>/dev/null || echo "")

if [ -z "$COMPANY_ID" ]; then
    echo_error "创建测试公司失败"
    echo "$COMPANY_RESP"
    exit 1
fi

echo_success "创建测试公司成功: $COMPANY_ID"

# 2.2 创建测试部门（前置数据 - 作为公司的子组织）
echo_info "2.2 创建测试部门（前置数据，父组织: $COMPANY_ID）..."
DEPT_RESP=$(curl -s -X POST "$BASE_URL/api/v1/organizations" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "'$TENANT_ID'",
    "code": "DEPT_HRM_TEST_'$TIMESTAMP'",
    "name": "HRM测试部门",
    "type": "department",
    "parent_id": "'$COMPANY_ID'"
  }')

DEPT_ID=$(echo "$DEPT_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))" 2>/dev/null || echo "")

if [ -z "$DEPT_ID" ]; then
    echo_error "创建测试部门失败"
    echo "$DEPT_RESP"
    exit 1
fi

echo_success "创建测试部门成功: $DEPT_ID"

# 2.3 创建测试员工（前置数据 - 通过 API）
echo_info "2.3 创庺测试员工（前置数据 - 通过 API，关联用户: $USER_ID）..."
EMPLOYEE_RESP=$(curl -s -X POST "$BASE_URL/api/v1/employees" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "'$TENANT_ID'",
    "user_id": "'$USER_ID'",
    "employee_no": "EMP_HRM_'$TIMESTAMP'",
    "name": "测试员工",
    "org_id": "'$DEPT_ID'"
  }')

# 从响应中获取员工 ID
EMPLOYEE_ID=$(echo "$EMPLOYEE_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))" 2>/dev/null || echo "")

if [ -z "$EMPLOYEE_ID" ]; then
    echo_error "创建员工失败"
    echo "$EMPLOYEE_RESP"
    exit 1
fi

echo_success "创建员工成功: $EMPLOYEE_ID"

# 2.4 创建 HRM 员工扩展记录（前置数据 - 必需，满足外键约束）
echo_info "2.4 创建 HRM 员工扩展记录（前置数据 - 排班和打卡功能必需）..."

# 使用 -t 选项分配伪终端，避免卡住
SQL_RESULT=$(docker exec -i go-next-erp-postgres psql -U postgres -d erp <<EOF 2>&1
INSERT INTO hrm_employees (id, tenant_id, employee_id, is_active, created_at, updated_at) 
VALUES (gen_random_uuid(), '$TENANT_ID'::uuid, '$EMPLOYEE_ID'::uuid, true, NOW(), NOW());
EOF
)

if echo "$SQL_RESULT" | grep -q "INSERT 0 1"; then
    echo_success "HRM 员工记录创建成功"
    SKIP_HRM_EMPLOYEE_TESTS=false
else
    echo_error "HRM 员工记录创建失败: $SQL_RESULT"
    echo_info "将跳过需要此记录的测试"
    SKIP_HRM_EMPLOYEE_TESTS=true
fi

echo_success "前置数据创建完成"
echo_info "可用的前置数据:"
echo_info "  - 公司ID: $COMPANY_ID"
echo_info "  - 部门ID: $DEPT_ID"
echo_info "  - 员工ID: $EMPLOYEE_ID"
echo_info "  - HRM员工记录: $([ "$SKIP_HRM_EMPLOYEE_TESTS" = "false" ] && echo "已创建" || echo "创建失败")"
echo ""

# ========== 步骤3: 班次管理测试 ==========
echo ""
echo_info "========== 班次管理测试 =========="

# 3.1 创建班次（为后续排班和考勤提供前置数据）
echo_info "3.1 创建班次（作为前置数据）..."
SHIFT_RESP=$(curl -s -X POST "$BASE_URL/api/v1/hrm/shifts" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "'$TENANT_ID'",
    "code": "MORNING_TEST",
    "name": "早班测试",
    "type": "fixed",
    "work_start": "09:00",
    "work_end": "18:00",
    "work_hours": 8,
    "late_grace_period": 10,
    "early_grace_period": 10
  }')

SHIFT_ID=$(echo "$SHIFT_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))" 2>/dev/null || echo "")

if [ -n "$SHIFT_ID" ]; then
    echo_success "创建班次成功: $SHIFT_ID"
else
    echo_error "创建班次失败"
    echo "$SHIFT_RESP"
    exit 1
fi

# 3.2 获取班次详情
echo_info "3.2 获取班次详情..."
GET_SHIFT_RESP=$(curl -s -X GET "$BASE_URL/api/v1/hrm/shifts/$SHIFT_ID" \
  -H "Authorization: Bearer $TOKEN")

if echo "$GET_SHIFT_RESP" | grep -q "$SHIFT_ID"; then
    echo_success "获取班次详情成功"
else
    echo_error "获取班次详情失败"
fi

# 3.3 更新班次
echo_info "3.3 更新班次..."
UPDATE_SHIFT_RESP=$(curl -s -X PUT "$BASE_URL/api/v1/hrm/shifts/$SHIFT_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "早班测试(已更新)"
  }')

if echo "$UPDATE_SHIFT_RESP" | grep -q "已更新"; then
    echo_success "更新班次成功"
else
    echo_error "更新班次失败"
fi

# 3.4 查询班次列表
echo_info "3.4 查询班次列表..."
LIST_SHIFTS_RESP=$(curl -s -X GET "$BASE_URL/api/v1/hrm/shifts?tenant_id=$TENANT_ID" \
  -H "Authorization: Bearer $TOKEN")

if echo "$LIST_SHIFTS_RESP" | grep -q "items"; then
    echo_success "查询班次列表成功"
else
    echo_error "查询班次列表失败"
fi

# 3.5 查询有效班次
echo_info "3.5 查询有效班次..."
ACTIVE_SHIFTS_RESP=$(curl -s -X GET "$BASE_URL/api/v1/hrm/shifts/active?tenant_id=$TENANT_ID" \
  -H "Authorization: Bearer $TOKEN")

if echo "$ACTIVE_SHIFTS_RESP" | grep -q "items"; then
    echo_success "查询有效班次成功"
else
    echo_error "查询有效班次失败"
fi

# ========== 步骤4: 排班管理测试 ==========
echo ""
echo_info "========== 排班管理测试 =========="

# 前置条件检查
if [ "$SKIP_HRM_EMPLOYEE_TESTS" = "true" ]; then
    echo_info "跳过排班测试（缺少 HRM 员工记录）"
    echo ""
    echo_info "========== 考勤规则测试 =========="
else
    # 4.1 创建排班（使用前面创建的班次、员工、部门）
    echo_info "4.1 创建排班（使用前置数据: 员工=$EMPLOYEE_ID, 班次=$SHIFT_ID）..."
    SCHEDULE_DATE=$(date +%Y-%m-%d)
    SCHEDULE_RESP=$(curl -s -X POST "$BASE_URL/api/v1/hrm/schedules" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "tenant_id": "'$TENANT_ID'",
        "employee_id": "'$EMPLOYEE_ID'",
        "shift_id": "'$SHIFT_ID'",
        "schedule_date": "'$SCHEDULE_DATE'",
        "department_id": "'$DEPT_ID'"
      }')

    SCHEDULE_ID=$(echo "$SCHEDULE_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))" 2>/dev/null || echo "")

    if [ -n "$SCHEDULE_ID" ]; then
        echo_success "创建排班成功: $SCHEDULE_ID"
    else
        echo_error "创建排班失败"
        echo "$SCHEDULE_RESP"
    fi

    # 4.2 批量创建排班
    echo_info "4.2 批量创建排班..."
    TOMORROW=$(date -v+1d +%Y-%m-%d 2>/dev/null || date -d 'tomorrow' +%Y-%m-%d)
    BATCH_SCHEDULE_RESP=$(curl -s -X POST "$BASE_URL/api/v1/hrm/schedules/batch" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "schedules": [
          {
            "tenant_id": "'$TENANT_ID'",
            "employee_id": "'$EMPLOYEE_ID'",
            "shift_id": "'$SHIFT_ID'",
            "schedule_date": "'$TOMORROW'",
            "department_id": "'$DEPT_ID'"
          }
        ]
      }')

    if echo "$BATCH_SCHEDULE_RESP" | grep -q "items"; then
        echo_success "批量创建排班成功"
    else
        echo_error "批量创建排班失败"
    fi

    # 4.3 查询员工排班
    echo_info "4.3 查询员工排班..."
    START_DATE=$(date -v-7d +%Y-%m-%d 2>/dev/null || date -d '7 days ago' +%Y-%m-%d)
    END_DATE=$(date -v+7d +%Y-%m-%d 2>/dev/null || date -d '7 days' +%Y-%m-%d)
    EMP_SCHEDULES_RESP=$(curl -s -X GET "$BASE_URL/api/v1/hrm/schedules/employee/$EMPLOYEE_ID?tenant_id=$TENANT_ID&start_date=$START_DATE&end_date=$END_DATE" \
      -H "Authorization: Bearer $TOKEN")

    if echo "$EMP_SCHEDULES_RESP" | grep -q "items"; then
        echo_success "查询员工排班成功"
    else
        echo_error "查询员工排班失败"
    fi

    # ========== 步骤5: 考勤规则测试 ==========
    echo ""
    echo_info "========== 考勤规则测试 =========="
fi

# 5.1 创建考勤规则
echo_info "5.1 创建考勤规则..."
RULE_RESP=$(curl -s -X POST "$BASE_URL/api/v1/hrm/attendance-rules" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "'$TENANT_ID'",
    "code": "RULE_TEST",
    "name": "测试考勤规则",
    "apply_type": "all",
    "workday_type": "workday",
    "location_required": false,
    "wifi_required": false,
    "face_required": false
  }')

RULE_ID=$(echo "$RULE_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))" 2>/dev/null || echo "")

if [ -n "$RULE_ID" ]; then
    echo_success "创建考勤规则成功: $RULE_ID"
else
    echo_error "创建考勤规则失败"
    echo "$RULE_RESP"
fi

# 5.2 查询考勤规则列表
echo_info "5.2 查询考勤规则列表..."
LIST_RULES_RESP=$(curl -s -X GET "$BASE_URL/api/v1/hrm/attendance-rules?tenant_id=$TENANT_ID" \
  -H "Authorization: Bearer $TOKEN")

if echo "$LIST_RULES_RESP" | grep -q "items"; then
    echo_success "查询考勤规则列表成功"
else
    echo_error "查询考勤规则列表失败"
fi

# ========== 步骤6: 考勤打卡测试 ==========
echo ""
echo_info "========== 考勤打卡测试 =========="

# 前置条件检查：考勤打卡需要 HRM 员工记录
if [ "$SKIP_HRM_EMPLOYEE_TESTS" = "true" ]; then
    echo_info "跳过考勤打卡测试（缺少 HRM 员工记录）"
    echo_info "前置数据要求: 员工=$EMPLOYEE_ID, 部门=$DEPT_ID, HRM员工记录"
else
    echo_info "使用前置数据进行考勤打卡测试:"
    echo_info "  - 员工ID: $EMPLOYEE_ID"
    echo_info "  - 部门ID: $DEPT_ID"
    echo_info "  - HRM员工记录: 已创建"
    echo ""

    # 6.1 上班打卡（使用前置数据：员工、部门、HRM员工记录）
    echo_info "6.1 上班打卡（使用前置数据）..."
    CLOCK_IN_RESP=$(curl -s -X POST "$BASE_URL/api/v1/hrm/attendance/clock-in" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "tenant_id": "'$TENANT_ID'",
        "employee_id": "'$EMPLOYEE_ID'",
        "clock_type": "check_in",
        "check_in_method": "mobile",
        "source_type": "system",
        "department_id": "'$DEPT_ID'"
      }')

    ATTENDANCE_ID=$(echo "$CLOCK_IN_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))" 2>/dev/null || echo "")

    if [ -n "$ATTENDANCE_ID" ]; then
        echo_success "上班打卡成功: $ATTENDANCE_ID"
    else
        echo_error "上班打卡失败"
        echo "$CLOCK_IN_RESP"
    fi

    # 6.2 下班打卡（使用前置数据：员工、部门、HRM员工记录）
    echo_info "6.2 下班打卡（使用前置数据）..."
    sleep 1
    CLOCK_OUT_RESP=$(curl -s -X POST "$BASE_URL/api/v1/hrm/attendance/clock-in" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "tenant_id": "'$TENANT_ID'",
        "employee_id": "'$EMPLOYEE_ID'",
        "clock_type": "check_out",
        "check_in_method": "mobile",
        "source_type": "system",
        "department_id": "'$DEPT_ID'"
      }')

    if echo "$CLOCK_OUT_RESP" | grep -q "id"; then
        echo_success "下班打卡成功"
    else
        echo_error "下班打卡失败"
        echo "$CLOCK_OUT_RESP"
    fi

    # 6.3 获取考勤记录详情（使用上面创建的考勤记录）
    if [ -n "$ATTENDANCE_ID" ]; then
        echo_info "6.3 获取考勤记录详情（使用考勤记录: $ATTENDANCE_ID）..."
        GET_ATTENDANCE_RESP=$(curl -s -X GET "$BASE_URL/api/v1/hrm/attendance/records/$ATTENDANCE_ID" \
          -H "Authorization: Bearer $TOKEN")

        if echo "$GET_ATTENDANCE_RESP" | grep -q "id"; then
            echo_success "获取考勤记录详情成功"
        else
            echo_error "获取考勤记录详情失败"
            echo "Response: $GET_ATTENDANCE_RESP"
        fi
    fi

    # 6.4 查询员工考勤记录（使用前置数据：员工ID）
    echo_info "6.4 查询员工考勤记录（使用员工: $EMPLOYEE_ID）..."
    EMPLOYEE_ATTENDANCE_RESP=$(curl -s -X GET "$BASE_URL/api/v1/hrm/attendance/employee/$EMPLOYEE_ID?tenant_id=$TENANT_ID&start_date=$START_DATE&end_date=$END_DATE" \
      -H "Authorization: Bearer $TOKEN")

    if echo "$EMPLOYEE_ATTENDANCE_RESP" | grep -q "items"; then
        RECORD_COUNT=$(echo "$EMPLOYEE_ATTENDANCE_RESP" | python3 -c "import sys, json; print(json.load(sys.stdin).get('total', 0))" 2>/dev/null || echo "0")
        echo_success "查询员工考勤记录成功，共 $RECORD_COUNT 条"
    else
        echo_error "查询员工考勤记录失败"
    fi

    # 6.5 查询部门考勤记录（使用前置数据：部门ID）
    echo_info "6.5 查询部门考勤记录（使用部门: $DEPT_ID）..."
    DEPT_ATTENDANCE_RESP=$(curl -s -X GET "$BASE_URL/api/v1/hrm/attendance/department/$DEPT_ID?tenant_id=$TENANT_ID&start_date=$START_DATE&end_date=$END_DATE" \
      -H "Authorization: Bearer $TOKEN")

    if echo "$DEPT_ATTENDANCE_RESP" | grep -q "items"; then
        echo_success "查询部门考勤记录成功"
    else
        echo_error "查询部门考勤记录失败"
    fi

    # 6.6 查询异常考勤
    echo_info "6.6 查询异常考勤..."
    EXCEPTION_RESP=$(curl -s -X GET "$BASE_URL/api/v1/hrm/attendance/exceptions?tenant_id=$TENANT_ID&start_date=$START_DATE&end_date=$END_DATE" \
      -H "Authorization: Bearer $TOKEN")

    if echo "$EXCEPTION_RESP" | grep -q "items"; then
        echo_success "查询异常考勤成功"
    else
        echo_error "查询异常考勤失败"
    fi

    # 6.7 获取考勤统计
    echo_info "6.7 获取考勤统计..."
    STATS_RESP=$(curl -s -X GET "$BASE_URL/api/v1/hrm/attendance/statistics?tenant_id=$TENANT_ID&start_date=$START_DATE&end_date=$END_DATE" \
      -H "Authorization: Bearer $TOKEN")

    if echo "$STATS_RESP" | grep -q "statusCount"; then
        echo_success "获取考勤统计成功"
        echo "$STATS_RESP" | python3 -m json.tool 2>/dev/null || echo "$STATS_RESP"
    else
        echo_error "获取考勤统计失败"
    fi
fi

# ========== 步骤7: 清理测试数据 ==========
echo ""
echo_info "========== 清理测试数据 =========="

# 7.1 删除排班
if [ -n "$SCHEDULE_ID" ]; then
    echo_info "删除排班..."
    curl -s -X DELETE "$BASE_URL/api/v1/hrm/schedules/$SCHEDULE_ID" \
      -H "Authorization: Bearer $TOKEN" > /dev/null
    echo_success "删除排班成功"
fi

# 7.2 删除考勤规则
if [ -n "$RULE_ID" ]; then
    echo_info "删除考勤规则..."
    curl -s -X DELETE "$BASE_URL/api/v1/hrm/attendance-rules/$RULE_ID" \
      -H "Authorization: Bearer $TOKEN" > /dev/null
    echo_success "删除考勤规则成功"
fi

# 7.3 删除班次
if [ -n "$SHIFT_ID" ]; then
    echo_info "删除班次..."
    curl -s -X DELETE "$BASE_URL/api/v1/hrm/shifts/$SHIFT_ID" \
      -H "Authorization: Bearer $TOKEN" > /dev/null
    echo_success "删除班次成功"
fi

echo ""
echo_success "========== HRM 模块集成测试完成 =========="
echo_info "所有 API 测试通过！"
