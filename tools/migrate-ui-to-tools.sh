#!/bin/bash
# 迁移脚本：将 backend 和 polardbx-ui 迁移到 tools 目录下

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "=== 开始迁移 backend 和 polardbx-ui 到 tools 目录 ==="
echo ""

# 检查源目录是否存在
if [ ! -d "backend" ]; then
    echo "❌ 错误: backend 目录不存在"
    exit 1
fi

if [ ! -d "polardbx-ui" ]; then
    echo "❌ 错误: polardbx-ui 目录不存在"
    exit 1
fi

# 检查目标目录是否已存在
if [ -d "tools/ui-backend" ] || [ -d "tools/ui-frontend" ]; then
    echo "⚠️  警告: 目标目录已存在"
    read -p "是否继续？这将覆盖现有目录 (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "取消迁移"
        exit 1
    fi
fi

echo "步骤 1: 创建目标目录..."
mkdir -p tools/ui-backend
mkdir -p tools/ui-frontend

echo "步骤 2: 移动 backend 到 tools/ui-backend..."
cp -r backend/* tools/ui-backend/
# 保留隐藏文件
cp -r backend/.[!.]* tools/ui-backend/ 2>/dev/null || true

echo "步骤 3: 移动 polardbx-ui 到 tools/ui-frontend..."
cp -r polardbx-ui/* tools/ui-frontend/
# 保留隐藏文件
cp -r polardbx-ui/.[!.]* tools/ui-frontend/ 2>/dev/null || true

echo "步骤 4: 更新 tools/ui-backend/go.mod..."
if [ -f "tools/ui-backend/go.mod" ]; then
    # 更新 replace 指令：从 ../ 改为 ../../
    sed -i 's|replace github.com/alibaba/polardbx-operator => \.\./|replace github.com/alibaba/polardbx-operator => ../..|g' tools/ui-backend/go.mod
    echo "✓ go.mod 已更新"
else
    echo "⚠️  警告: go.mod 不存在"
fi

echo "步骤 5: 更新 Dockerfile..."
if [ -f "tools/ui-all-in-one.Dockerfile" ]; then
    # 备份原文件
    cp tools/ui-all-in-one.Dockerfile tools/ui-all-in-one.Dockerfile.bak
    
    # 更新路径
    sed -i 's|COPY polardbx-ui/|COPY tools/ui-frontend/|g' tools/ui-all-in-one.Dockerfile
    sed -i 's|COPY backend/|COPY tools/ui-backend/|g' tools/ui-all-in-one.Dockerfile
    sed -i 's|/workspace/polardbx-ui|/workspace/ui-frontend|g' tools/ui-all-in-one.Dockerfile
    
    echo "✓ Dockerfile 已更新（备份: tools/ui-all-in-one.Dockerfile.bak）"
else
    echo "⚠️  警告: Dockerfile 不存在"
fi

echo "步骤 6: 更新部署脚本..."
# 更新 deploy-all-in-one.sh
if [ -f "tools/deploy-all-in-one.sh" ]; then
    cp tools/deploy-all-in-one.sh tools/deploy-all-in-one.sh.bak
    sed -i 's|backend/|tools/ui-backend/|g' tools/deploy-all-in-one.sh
    sed -i 's|polardbx-ui/|tools/ui-frontend/|g' tools/deploy-all-in-one.sh
    echo "✓ deploy-all-in-one.sh 已更新"
fi

# 更新 test-all-in-one.sh
if [ -f "tools/test-all-in-one.sh" ]; then
    cp tools/test-all-in-one.sh tools/test-all-in-one.sh.bak
    sed -i 's|backend/|tools/ui-backend/|g' tools/test-all-in-one.sh
    sed -i 's|polardbx-ui/|tools/ui-frontend/|g' tools/test-all-in-one.sh
    echo "✓ test-all-in-one.sh 已更新"
fi

echo ""
echo "=== 迁移完成 ==="
echo ""
echo "下一步："
echo "1. 检查 tools/ui-backend/go.mod 中的 replace 路径是否正确"
echo "2. 测试构建: docker build -f tools/ui-all-in-one.Dockerfile -t test ."
echo "3. 如果一切正常，可以删除原目录:"
echo "   rm -rf backend polardbx-ui"
echo ""
echo "备份文件:"
echo "  - tools/ui-all-in-one.Dockerfile.bak"
echo "  - tools/deploy-all-in-one.sh.bak"
echo "  - tools/test-all-in-one.sh.bak"
echo ""
echo "⚠️  注意: 原目录 backend/ 和 polardbx-ui/ 仍然存在，请验证后再删除"

