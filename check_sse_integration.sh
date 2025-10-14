#!/bin/bash

# SSE 模块集成检查脚本
# 用于验证 SSE 模块是否正确集成到项目中

echo "=========================================="
echo "  SSE 模块集成检查"
echo "=========================================="
echo ""

# 1. 检查核心文件是否存在
echo "【1. 检查核心文件】"
FILES=(
    "pkg/sse/broker.go"
    "pkg/sse/types.go"
    "pkg/sse/handler.go"
    "pkg/sse/wire.go"
)

ALL_EXISTS=true
for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "  ✅ $file"
    else
        echo "  ❌ $file (缺失)"
        ALL_EXISTS=false
    fi
done

if [ "$ALL_EXISTS" = false ]; then
    echo ""
    echo "❌ 部分核心文件缺失，请检查"
    exit 1
fi

echo ""

# 2. 检查文档文件
echo "【2. 检查文档文件】"
DOCS=(
    "docs/SSE_MODULE.md"
    "pkg/sse/README.md"
    "docs/SSE_IMPLEMENTATION_SUMMARY.md"
)

for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        echo "  ✅ $doc"
    else
        echo "  ⚠️  $doc (建议添加)"
    fi
done

echo ""

# 3. 检查示例文件
echo "【3. 检查示例文件】"
if [ -f "pkg/sse/examples/integration.go" ]; then
    echo "  ✅ 集成示例存在"
else
    echo "  ⚠️  集成示例缺失"
fi

if [ -f "test_sse.html" ]; then
    echo "  ✅ 浏览器测试页面存在"
else
    echo "  ⚠️  测试页面缺失"
fi

echo ""

# 4. 检查代码语法
echo "【4. 检查代码语法】"
if command -v go &> /dev/null; then
    if go build -o /dev/null ./pkg/sse/... 2>&1 | grep -q "error"; then
        echo "  ❌ SSE 模块编译失败"
        go build ./pkg/sse/... 2>&1 | head -20
    else
        echo "  ✅ SSE 模块编译通过"
    fi
else
    echo "  ⚠️  Go 未安装，跳过编译检查"
fi

echo ""

# 5. 统计信息
echo "【5. 代码统计】"
if command -v wc &> /dev/null; then
    BROKER_LINES=$(wc -l < pkg/sse/broker.go 2>/dev/null || echo "0")
    TYPES_LINES=$(wc -l < pkg/sse/types.go 2>/dev/null || echo "0")
    HANDLER_LINES=$(wc -l < pkg/sse/handler.go 2>/dev/null || echo "0")
    TOTAL=$((BROKER_LINES + TYPES_LINES + HANDLER_LINES))
    
    echo "  📊 broker.go:  $BROKER_LINES 行"
    echo "  📊 types.go:   $TYPES_LINES 行"
    echo "  📊 handler.go: $HANDLER_LINES 行"
    echo "  📊 总计:       $TOTAL 行"
fi

echo ""

# 6. 集成建议
echo "【6. 集成步骤】"
echo ""
echo "步骤1: 添加到 Wire Provider"
echo "  编辑 pkg/wire.go，添加:"
echo "  import \"github.com/lk2023060901/go-next-erp/pkg/sse\""
echo "  var ProviderSet = wire.NewSet("
echo "      // ... 现有 Providers"
echo "      sse.ProviderSet,  // 添加这一行"
echo "  )"
echo ""

echo "步骤2: 注册 HTTP 路由"
echo "  编辑 internal/server/http.go，在 NewHTTPServer 中添加:"
echo "  sseBroker *sse.Broker,"
echo "  sseHandler *sse.Handler,"
echo ""
echo "  然后在函数体中:"
echo "  go sseBroker.Start(context.Background())"
echo "  srv.HandleFunc(\"/api/v1/sse/stream\", sseHandler.ServeHTTP)"
echo ""

echo "步骤3: 在业务代码中使用"
echo "  在需要实时推送的服务中注入 SSE Broker:"
echo "  type YourService struct {"
echo "      sseBroker *sse.Broker"
echo "  }"
echo ""
echo "  发送推送:"
echo "  s.sseBroker.SendToUser(userID, \"event\", data)"
echo ""

# 7. 测试建议
echo "【7. 测试方式】"
echo ""
echo "单元测试:"
echo "  cd pkg/sse && go test -v"
echo ""
echo "性能测试:"
echo "  cd pkg/sse && go test -bench=. -benchmem"
echo ""
echo "浏览器测试:"
echo "  1. 启动服务器: go run cmd/server/main.go"
echo "  2. 打开页面: open test_sse.html"
echo ""

# 8. 完成提示
echo "=========================================="
echo "  ✅ SSE 模块集成检查完成"
echo "=========================================="
echo ""
echo "📚 详细文档:"
echo "  - 完整技术文档: docs/SSE_MODULE.md"
echo "  - 快速开始: pkg/sse/README.md"
echo "  - 实现总结: docs/SSE_IMPLEMENTATION_SUMMARY.md"
echo ""
echo "🎯 下一步:"
echo "  根据上述步骤将 SSE 模块集成到您的业务代码中"
echo ""
