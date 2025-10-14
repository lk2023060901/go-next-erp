#!/bin/bash

# HRM 模块 API 测试脚本
# 使用方法: ./test_hrm_api.sh

BASE_URL="http://localhost:15006"
TENANT_ID="00000000-0000-0000-0000-000000000001"
EMPLOYEE_ID="00000000-0000-0000-0000-000000000002"
DEPARTMENT_ID="00000000-0000-0000-0000-000000000003"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}HRM 模块 API 测试${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# 测试函数
test_api() {
    local name="$1"
    local method="$2"
    local path="$3"
    local data="$4"
    
    echo -e "${YELLOW}测试: $name${NC}"
    echo "URL: $method $BASE_URL$path"
    
    if [ -n "$data" ]; then
        echo "数据: $data"
        response=$(curl -s -X $method "$BASE_URL$path" \
            -H "Content-Type: application/json" \
            -d "$data" \
            -w "\nHTTP_CODE:%{http_code}")
    else
        response=$(curl -s -X $method "$BASE_URL$path" \
            -H "Content-Type: application/json" \
            -w "\nHTTP_CODE:%{http_code}")
    fi
    
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d':' -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d')
    
    if [ "$http_code" == "200" ]; then
        echo -e "${GREEN}✓ 成功 (HTTP $http_code)${NC}"
        echo "响应: $body" | python3 -m json.tool 2>/dev/null || echo "$body"
    else
        echo -e "${RED}✗ 失败 (HTTP $http_code)${NC}"
        echo "响应: $body"
    fi
    echo ""
}

# 1. 班次管理测试
echo -e "${GREEN}========== 班次管理 ==========${NC}"
echo ""

# 创建班次 - 标准上班班次
test_api "创建班次 (标准班)" "POST" "/api/v1/hrm/shifts" '{
  "tenant_id": "'"$TENANT_ID"'",
  "code": "SHIFT_NORMAL",
  "name": "标准上班班次",
  "type": "fixed",
  "description": "朝九晚六标准班次",
  "work_start": "09:00",
  "work_end": "18:00",
  "check_in_required": true,
  "check_out_required": true,
  "late_grace_period": 10,
  "early_grace_period": 10,
  "is_cross_days": false,
  "allow_overtime": true,
  "sort": 1
}'

# 保存班次ID用于后续测试
SHIFT_ID=$(curl -s -X POST "$BASE_URL/api/v1/hrm/shifts" \
    -H "Content-Type: application/json" \
    -d '{
  "tenant_id": "'"$TENANT_ID"'",
  "code": "SHIFT_NORMAL_2",
  "name": "标准上班班次2",
  "type": "fixed",
  "work_start": "09:00",
  "work_end": "18:00",
  "check_in_required": true,
  "check_out_required": true,
  "late_grace_period": 10,
  "early_grace_period": 10
}' | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))" 2>/dev/null)

# 查询班次列表
test_api "查询班次列表" "GET" "/api/v1/hrm/shifts?tenant_id=$TENANT_ID&is_active=true&page=1&page_size=10"

# 查询启用的班次
test_api "查询启用的班次" "GET" "/api/v1/hrm/shifts/active?tenant_id=$TENANT_ID"

# 如果获取到班次ID，测试更新和删除
if [ -n "$SHIFT_ID" ]; then
    # 获取班次详情
    test_api "获取班次详情" "GET" "/api/v1/hrm/shifts/$SHIFT_ID"
    
    # 更新班次
    test_api "更新班次" "PUT" "/api/v1/hrm/shifts/$SHIFT_ID" '{
      "name": "标准上班班次(已更新)",
      "description": "更新后的描述",
      "work_start": "09:30",
      "work_end": "18:30",
      "late_grace_period": 15,
      "early_grace_period": 15,
      "is_active": true,
      "sort": 1
    }'
fi

# 2. 排班管理测试
echo -e "${GREEN}========== 排班管理 ==========${NC}"
echo ""

# 创建排班记录
test_api "创建排班" "POST" "/api/v1/hrm/schedules" '{
  "tenant_id": "'"$TENANT_ID"'",
  "employee_id": "'"$EMPLOYEE_ID"'",
  "shift_id": "'"${SHIFT_ID:-$TENANT_ID}"'",
  "schedule_date": "2025-10-15",
  "workday_type": "workday",
  "remark": "正常排班"
}'

# 保存排班ID
SCHEDULE_ID=$(curl -s -X POST "$BASE_URL/api/v1/hrm/schedules" \
    -H "Content-Type: application/json" \
    -d '{
  "tenant_id": "'"$TENANT_ID"'",
  "employee_id": "'"$EMPLOYEE_ID"'",
  "shift_id": "'"${SHIFT_ID:-$TENANT_ID}"'",
  "schedule_date": "2025-10-16",
  "workday_type": "workday"
}' | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))" 2>/dev/null)

# 查询员工排班
test_api "查询员工排班" "GET" "/api/v1/hrm/schedules/employee?tenant_id=$TENANT_ID&employee_id=$EMPLOYEE_ID&month=2025-10"

# 查询部门排班
test_api "查询部门排班" "GET" "/api/v1/hrm/schedules/department?tenant_id=$TENANT_ID&department_id=$DEPARTMENT_ID&month=2025-10"

# 如果获取到排班ID，测试更新和删除
if [ -n "$SCHEDULE_ID" ]; then
    # 获取排班详情
    test_api "获取排班详情" "GET" "/api/v1/hrm/schedules/$SCHEDULE_ID"
    
    # 更新排班
    test_api "更新排班" "PUT" "/api/v1/hrm/schedules/$SCHEDULE_ID" '{
      "workday_type": "weekend",
      "status": "published",
      "remark": "更新后的排班"
    }'
fi

# 3. 考勤规则测试
echo -e "${GREEN}========== 考勤规则 ==========${NC}"
echo ""

# 创建考勤规则
test_api "创建考勤规则" "POST" "/api/v1/hrm/attendance-rules" '{
  "tenant_id": "'"$TENANT_ID"'",
  "code": "RULE_OFFICE",
  "name": "办公室考勤规则",
  "description": "需要定位、WiFi、人脸识别",
  "location_required": true,
  "allowed_locations": [
    {
      "name": "公司总部",
      "latitude": 39.9042,
      "longitude": 116.4074,
      "radius": 500,
      "address": "北京市朝阳区"
    }
  ],
  "wifi_required": true,
  "allowed_wifi": ["Company-WiFi", "Office-5G"],
  "face_required": true,
  "face_threshold": 0.85
}'

# 保存规则ID
RULE_ID=$(curl -s -X POST "$BASE_URL/api/v1/hrm/attendance-rules" \
    -H "Content-Type: application/json" \
    -d '{
  "tenant_id": "'"$TENANT_ID"'",
  "code": "RULE_OFFICE_2",
  "name": "办公室考勤规则2",
  "location_required": true,
  "wifi_required": false,
  "face_required": false
}' | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))" 2>/dev/null)

# 查询考勤规则列表
test_api "查询考勤规则列表" "GET" "/api/v1/hrm/attendance-rules?tenant_id=$TENANT_ID&is_active=true&page=1&page_size=10"

# 如果获取到规则ID，测试更新和删除
if [ -n "$RULE_ID" ]; then
    # 获取规则详情
    test_api "获取考勤规则详情" "GET" "/api/v1/hrm/attendance-rules/$RULE_ID"
    
    # 更新规则
    test_api "更新考勤规则" "PUT" "/api/v1/hrm/attendance-rules/$RULE_ID" '{
      "name": "办公室考勤规则(已更新)",
      "description": "更新后的规则描述",
      "location_required": false,
      "wifi_required": false,
      "face_required": false,
      "is_active": true
    }'
fi

# 4. 考勤记录测试
echo -e "${GREEN}========== 考勤记录 ==========${NC}"
echo ""

# 上班打卡
test_api "上班打卡" "POST" "/api/v1/hrm/attendance/clock-in" '{
  "tenant_id": "'"$TENANT_ID"'",
  "employee_id": "'"$EMPLOYEE_ID"'",
  "clock_type": "check_in",
  "check_in_method": "app",
  "location": {
    "latitude": 39.9042,
    "longitude": 116.4074,
    "accuracy": 10.5
  },
  "address": "北京市朝阳区公司总部",
  "wifi_ssid": "Company-WiFi"
}'

# 保存考勤记录ID
ATTENDANCE_ID=$(curl -s -X POST "$BASE_URL/api/v1/hrm/attendance/clock-in" \
    -H "Content-Type: application/json" \
    -d '{
  "tenant_id": "'"$TENANT_ID"'",
  "employee_id": "'"$EMPLOYEE_ID"'",
  "clock_type": "check_in",
  "check_in_method": "app"
}' | python3 -c "import sys, json; print(json.load(sys.stdin).get('id', ''))" 2>/dev/null)

# 查询员工考勤记录
test_api "查询员工考勤记录" "GET" "/api/v1/hrm/attendance/employee?tenant_id=$TENANT_ID&employee_id=$EMPLOYEE_ID&start_date=2025-10-01&end_date=2025-10-31&page=1&page_size=10"

# 查询部门考勤记录
test_api "查询部门考勤记录" "GET" "/api/v1/hrm/attendance/department?tenant_id=$TENANT_ID&department_id=$DEPARTMENT_ID&start_date=2025-10-01&end_date=2025-10-31&page=1&page_size=10"

# 查询异常考勤
test_api "查询异常考勤" "GET" "/api/v1/hrm/attendance/exceptions?tenant_id=$TENANT_ID&start_date=2025-10-01&end_date=2025-10-31&page=1&page_size=10"

# 查询考勤统计
test_api "查询考勤统计" "GET" "/api/v1/hrm/attendance/statistics?tenant_id=$TENANT_ID&employee_id=$EMPLOYEE_ID&start_date=2025-10-01&end_date=2025-10-31"

# 如果获取到考勤记录ID，测试查询详情
if [ -n "$ATTENDANCE_ID" ]; then
    test_api "获取考勤记录详情" "GET" "/api/v1/hrm/attendance/$ATTENDANCE_ID"
fi

# 5. 清理测试数据(可选)
echo -e "${GREEN}========== 清理测试数据 ==========${NC}"
echo ""

# 删除班次
if [ -n "$SHIFT_ID" ]; then
    test_api "删除班次" "DELETE" "/api/v1/hrm/shifts/$SHIFT_ID"
fi

# 删除排班
if [ -n "$SCHEDULE_ID" ]; then
    test_api "删除排班" "DELETE" "/api/v1/hrm/schedules/$SCHEDULE_ID"
fi

# 删除考勤规则
if [ -n "$RULE_ID" ]; then
    test_api "删除考勤规则" "DELETE" "/api/v1/hrm/attendance-rules/$RULE_ID"
fi

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}测试完成!${NC}"
echo -e "${YELLOW}========================================${NC}"
