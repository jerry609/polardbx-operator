#!/bin/bash
# 监控模块修复验证脚本

set -e

echo "======================================"
echo "监控模块修复验证脚本"
echo "======================================"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查函数
check_pass() {
    echo -e "${GREEN}✓${NC} $1"
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

echo "1. 检查硬编码路径修复..."
if grep -q "defaultTemplatePath" backend/pkg/api/grafana/templates.go; then
    check_pass "已移除硬编码路径，使用 defaultTemplatePath 常量"
else
    check_fail "未找到 defaultTemplatePath 常量"
fi

if grep -q "POLARDBX_GRAFANA_TEMPLATES_DIR" backend/pkg/api/grafana/templates.go; then
    check_pass "已添加环境变量支持"
else
    check_fail "未找到环境变量支持"
fi

if grep -q "searchPaths :=" backend/pkg/api/grafana/templates.go; then
    check_pass "已实现多路径搜索机制"
else
    check_fail "未找到多路径搜索逻辑"
fi

echo ""
echo "2. 检查预览日志增强..."
if grep -q "\[预览\]" polardbx-ui/src/app/components/monitoring-dashboard-templates/monitoring-dashboard-templates.component.ts; then
    check_pass "已添加预览流程日志（标签：[预览]）"
else
    check_fail "未找到预览日志"
fi

if grep -q "\[获取模板\]" polardbx-ui/src/app/components/monitoring-dashboard-templates/monitoring-dashboard-templates.component.ts; then
    check_pass "已添加模板获取日志（标签：[获取模板]）"
else
    check_fail "未找到模板获取日志"
fi

echo ""
echo "3. 检查健康检查修复..."
if grep -q "monitoringNamespaceFallback" polardbx-ui/src/app/components/monitoring-health/monitoring-health.component.ts; then
    check_pass "已添加命名空间回退常量"
else
    check_fail "未找到命名空间回退常量"
fi

if grep -q "normalizeNamespace" polardbx-ui/src/app/components/monitoring-health/monitoring-health.component.ts; then
    check_pass "已实现命名空间归一化逻辑"
else
    check_fail "未找到命名空间归一化函数"
fi

if grep -q "\[监控健康\]" polardbx-ui/src/app/components/monitoring-health/monitoring-health.component.ts; then
    check_pass "已添加健康检查日志（标签：[监控健康]）"
else
    check_fail "未找到健康检查日志"
fi

echo ""
echo "4. 编译验证..."

# 前端编译
echo -n "正在编译前端... "
if cd polardbx-ui && npm run build > /tmp/frontend-build.log 2>&1; then
    check_pass "前端编译成功"
    cd ..
else
    check_fail "前端编译失败，详情见 /tmp/frontend-build.log"
    cd ..
    exit 1
fi

# 后端编译
echo -n "正在编译后端... "
if cd backend && go build -o /tmp/backend-test main.go > /tmp/backend-build.log 2>&1; then
    check_pass "后端编译成功"
    cd ..
else
    check_fail "后端编译失败，详情见 /tmp/backend-build.log"
    cd ..
    exit 1
fi

echo ""
echo "======================================"
echo "验证完成！"
echo "======================================"
echo ""
echo "下一步操作："
echo "1. 启动后端服务器:"
echo "   cd backend && go run main.go"
echo ""
echo "2. 启动前端开发服务器:"
echo "   cd polardbx-ui && npm start"
echo ""
echo "3. 在浏览器中测试:"
echo "   - 访问 http://localhost:4200/operations/monitoring/dashboards"
echo "   - 点击任意模板的"预览"按钮"
echo "   - 打开浏览器控制台查看详细日志（[预览] 标签）"
echo ""
echo "4. 测试健康检查:"
echo "   - 访问 http://localhost:4200/operations/monitoring/health"
echo "   - 点击"刷新状态"按钮"
echo "   - 查看控制台日志确认命名空间归一化（[监控健康] 标签）"
echo ""
echo "5. 测试自定义模板目录（可选）:"
echo "   export POLARDBX_GRAFANA_TEMPLATES_DIR=/your/custom/path"
echo "   cd backend && go run main.go"
echo ""
