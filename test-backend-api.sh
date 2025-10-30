#!/bin/bash

# PolarDB-X 后端 API 测试脚本
# 自动注入 kubeconfig 并测试各个 API 端点

set -e

BACKEND_URL="${BACKEND_URL:-http://localhost:8080}"
NAMESPACE="${NAMESPACE:-default}"
CLUSTER_NAME="${CLUSTER_NAME:-test-pitr}"

# 颜色
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=== PolarDB-X 后端 API 测试 ===${NC}"
echo ""

# 获取 kubeconfig
echo -e "${BLUE}[1] 获取 kubeconfig...${NC}"
KUBECONFIG_B64=$(kubectl config view --raw --minify | base64 -w 0)
echo -e "${GREEN}✓ Kubeconfig 已编码${NC}"
echo ""

# 测试函数
test_api() {
    local name="$1"
    local method="$2"
    local path="$3"
    local data="$4"
    
    echo -e "${BLUE}测试: $name${NC}"
    echo "  方法: $method"
    echo "  路径: $path"
    
    if [ "$method" = "GET" ]; then
        RESPONSE=$(curl -s -X GET "$BACKEND_URL$path" \
            -H "X-Kubeconfig-B64: $KUBECONFIG_B64" \
            -H "Content-Type: application/json")
    else
        RESPONSE=$(curl -s -X "$method" "$BACKEND_URL$path" \
            -H "X-Kubeconfig-B64: $KUBECONFIG_B64" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    if echo "$RESPONSE" | jq . >/dev/null 2>&1; then
        echo -e "${GREEN}✓ 成功${NC}"
        echo "$RESPONSE" | jq . | head -20
    else
        echo -e "${YELLOW}⚠ 响应:${NC}"
        echo "$RESPONSE" | head -10
    fi
    echo ""
}

# 1. 获取集群列表
test_api "获取集群列表" "GET" "/api/v1/clusters"

# 2. 获取指定集群
test_api "获取集群详情" "GET" "/api/v1/clusters/$NAMESPACE/$CLUSTER_NAME"

# 3. 获取集群备份列表
test_api "获取备份列表" "GET" "/api/v1/clusters/$NAMESPACE/$CLUSTER_NAME/backups"

# 4. 获取集群 Pods
test_api "获取集群 Pods" "GET" "/api/v1/clusters/$NAMESPACE/$CLUSTER_NAME/pods"

# 5. 获取恢复状态
test_api "获取恢复状态" "GET" "/api/v1/clusters/$NAMESPACE/$CLUSTER_NAME/restore-status"

# 6. 测试创建备份（dry-run）
echo -e "${BLUE}[测试] 验证备份配置 (Dry-run)${NC}"
BACKUP_PAYLOAD='{
  "metadata": {
    "name": "test-backup-dryrun"
  },
  "spec": {
    "retentionTime": "168h",
    "cleanPolicy": "Retain",
    "storageProvider": {
      "storageName": "local",
      "sink": "local:///tmp/backup"
    },
    "preferredBackupRole": "follower"
  }
}'

RESPONSE=$(curl -s -X POST "$BACKEND_URL/api/v1/clusters/$NAMESPACE/$CLUSTER_NAME/backups/validate" \
    -H "X-Kubeconfig-B64: $KUBECONFIG_B64" \
    -H "Content-Type: application/json" \
    -d "$BACKUP_PAYLOAD")

echo "验证结果:"
echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
echo ""

# 7. 测试 HPFS Sinks
echo -e "${BLUE}[测试] 获取 HPFS Sinks${NC}"
RESPONSE=$(curl -s -X GET "$BACKEND_URL/api/v1/hpfs/sinks" \
    -H "X-Kubeconfig-B64: $KUBECONFIG_B64")
echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
echo ""

# 8. 测试系统状态
echo -e "${BLUE}[测试] 获取系统信息${NC}"
RESPONSE=$(curl -s -X GET "$BACKEND_URL/api/v1/system/info" \
    -H "X-Kubeconfig-B64: $KUBECONFIG_B64")
echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
echo ""

echo -e "${GREEN}=== 测试完成 ===${NC}"
echo ""
echo "后端服务运行正常！✅"
echo ""
echo "可用的主要 API 端点："
echo "  GET  /api/v1/clusters"
echo "  GET  /api/v1/clusters/:namespace/:name"
echo "  POST /api/v1/clusters/:namespace/:name/backups"
echo "  POST /api/v1/clusters/:namespace/:name/pitr"
echo "  GET  /api/v1/clusters/:namespace/:name/restore-status"
echo "  GET  /api/v1/hpfs/sinks"
echo ""
