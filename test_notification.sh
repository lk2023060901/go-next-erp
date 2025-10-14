#!/bin/bash

# 测试通知模块 API
BASE_URL="http://localhost:15006"

echo "==================== 通知模块 API 测试 ===================="

# 1. 登录获取 Token
echo -e "\n【1. 登录获取 Token】"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser1","password":"Password@123"}')

echo "登录响应: $LOGIN_RESPONSE"

TOKEN=$(echo $LOGIN_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['access_token'])" 2>/dev/null)

if [ -z "$TOKEN" ]; then
  echo "❌ 登录失败，无法获取 Token"
  echo "尝试注册新用户..."
  
  # 注册
  REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d '{
      "username": "notifytest",
      "email": "notify@test.com",
      "password": "Test@123456"
    }')
  
  echo "注册响应: $REGISTER_RESPONSE"
  
  # 再次登录
  LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"notifytest","password":"Test@123456"}')
  
  echo "登录响应: $LOGIN_RESPONSE"
  TOKEN=$(echo $LOGIN_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['access_token'])" 2>/dev/null)
fi

if [ -z "$TOKEN" ]; then
  echo "❌ 无法获取 Token，测试终止"
  exit 1
fi

echo "✅ Token 获取成功: ${TOKEN:0:50}..."

# 2. 发送通知
echo -e "\n【2. 发送站内通知】"
SEND_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/notifications" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "type": "system",
    "title": "系统通知测试",
    "content": "这是一条通过 WebSocket 推送的系统通知",
    "priority": "high"
  }')

echo "响应: $SEND_RESPONSE"

NOTIF_ID=$(echo $SEND_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])" 2>/dev/null)
echo "通知 ID: $NOTIF_ID"

# 3. 获取未读数量
echo -e "\n【3. 获取未读数量】"
UNREAD_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/notifications/unread-count" \
  -H "Authorization: Bearer $TOKEN")

echo "未读数量: $UNREAD_RESPONSE"

# 4. 列出通知
echo -e "\n【4. 列出通知】"
LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/notifications?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN")

echo "通知列表: $LIST_RESPONSE"

# 5. 获取单个通知
if [ ! -z "$NOTIF_ID" ]; then
  echo -e "\n【5. 获取单个通知】"
  GET_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/notifications/$NOTIF_ID" \
    -H "Authorization: Bearer $TOKEN")
  
  echo "通知详情: $GET_RESPONSE"
fi

# 6. 标记为已读
if [ ! -z "$NOTIF_ID" ]; then
  echo -e "\n【6. 标记为已读】"
  READ_RESPONSE=$(curl -s -X PUT "$BASE_URL/api/v1/notifications/$NOTIF_ID/read" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{}')
  
  echo "标记已读响应: $READ_RESPONSE"
fi

# 7. 再次获取未读数量
echo -e "\n【7. 再次获取未读数量】"
UNREAD_RESPONSE2=$(curl -s -X GET "$BASE_URL/api/v1/notifications/unread-count" \
  -H "Authorization: Bearer $TOKEN")

echo "未读数量: $UNREAD_RESPONSE2"

echo -e "\n==================== 测试完成 ===================="
