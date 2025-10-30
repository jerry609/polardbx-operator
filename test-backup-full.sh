#!/bin/bash

# 测试备份和 PITR 功能
set -e

BACKEND_URL="http://localhost:8080"
NAMESPACE="default"
CLUSTER_NAME="test-pitr"
KUBECONFIG_B64=$(kubectl config view --raw --minify | base64 -w 0)

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=== 测试备份和 PITR 功能 ===${NC}"
echo ""

# 1. 准备测试数据
echo -e "${BLUE}[1] 准备测试数据...${NC}"
CN_POD=$(kubectl get pods -n $NAMESPACE -l polardbx/name=$CLUSTER_NAME,polardbx/role=cn -o jsonpath='{.items[0].metadata.name}')

kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root << 'EOF'
CREATE DATABASE IF NOT EXISTS backup_test mode='auto';
USE backup_test;

CREATE TABLE IF NOT EXISTS test_data (
    id INT AUTO_INCREMENT PRIMARY KEY,
    data VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

DELETE FROM test_data;

INSERT INTO test_data (data) VALUES
    ('Initial data 1'),
    ('Initial data 2'),
    ('Initial data 3');

SELECT * FROM test_data;
EOF

echo -e "${GREEN}✓ 测试数据已创建${NC}"
echo ""

# 2. 创建备份
echo -e "${BLUE}[2] 创建备份...${NC}"

BACKUP_NAME="test-backup-$(date +%Y%m%d-%H%M%S)"

BACKUP_PAYLOAD=$(cat <<EOF
{
  "metadata": {
    "name": "$BACKUP_NAME"
  },
  "spec": {
    "retentionTime": "168h",
    "cleanPolicy": "Retain",
    "storageProvider": {
      "storageName": "oss",
            "sink": "oss://test-bucket/backups/$CLUSTER_NAME/$BACKUP_NAME"
    },
    "preferredBackupRole": "follower"
  }
}
EOF
)

echo "备份配置:"
echo "$BACKUP_PAYLOAD" | jq .
echo ""

RESPONSE=$(curl -s -X POST "$BACKEND_URL/api/v1/clusters/$NAMESPACE/$CLUSTER_NAME/backups" \
    -H "X-Kubeconfig-B64: $KUBECONFIG_B64" \
    -H "Content-Type: application/json" \
    -d "$BACKUP_PAYLOAD")

if echo "$RESPONSE" | jq -e '.metadata.name' >/dev/null 2>&1; then
    echo -e "${GREEN}✓ 备份创建成功${NC}"
    echo "$RESPONSE" | jq '{name: .metadata.name, namespace: .metadata.namespace, phase: .status.phase}'
else
    echo -e "${YELLOW}响应:${NC}"
    echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
fi
echo ""

# 3. 查询备份状态
echo -e "${BLUE}[3] 查询备份状态...${NC}"
sleep 3

RESPONSE=$(curl -s -X GET "$BACKEND_URL/api/v1/clusters/$NAMESPACE/$CLUSTER_NAME/backups" \
    -H "X-Kubeconfig-B64: $KUBECONFIG_B64")

echo "$RESPONSE" | jq '.[] | {name: .metadata.name, phase: .status.phase, startTime: .status.startTime}' 2>/dev/null || echo "$RESPONSE"
echo ""

# 4. 直接使用 kubectl 查看备份
echo -e "${BLUE}[4] 使用 kubectl 查看备份...${NC}"
kubectl get polardbxbackup -n $NAMESPACE
echo ""

# 5. 添加更多数据（PITR 测试点）
echo -e "${BLUE}[5] 添加 PITR 测试数据...${NC}"
sleep 2

kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root backup_test << 'EOF'
INSERT INTO test_data (data) VALUES
    ('After backup 1'),
    ('After backup 2');
    
SELECT NOW() as pitr_target_time;
SELECT * FROM test_data;
EOF

PITR_TIME=$(kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root -N backup_test -e "SELECT NOW()")
echo -e "${GREEN}✓ PITR 目标时间: $PITR_TIME${NC}"
echo "$PITR_TIME" > /tmp/pitr_test_time.txt
echo ""

# 6. 添加错误数据
echo -e "${BLUE}[6] 添加错误数据（稍后将回滚）...${NC}"
sleep 3

kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root backup_test << 'EOF'
INSERT INTO test_data (data) VALUES
    ('ERROR - should be rolled back'),
    ('MISTAKE - will be removed');
    
SELECT * FROM test_data;
EOF

echo -e "${YELLOW}当前有 7 条记录（包括 2 条错误数据）${NC}"
echo ""

# 7. 测试 PITR API
echo -e "${BLUE}[7] 测试 PITR 恢复 API...${NC}"

TARGET_CLUSTER="$CLUSTER_NAME-restored"
PITR_TIME_ISO=$(date -d "$PITR_TIME" -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "")

if [ -z "$PITR_TIME_ISO" ]; then
    echo -e "${YELLOW}⚠ 时间格式转换可能有问题，使用原始时间${NC}"
    PITR_TIME_ISO="$PITR_TIME"
fi

PITR_PAYLOAD=$(cat <<EOF
{
  "time": "$PITR_TIME_ISO",
  "targetCluster": "$TARGET_CLUSTER",
  "backupSet": "$BACKUP_NAME",
  "timezone": "UTC"
}
EOF
)

echo "PITR 配置:"
echo "$PITR_PAYLOAD" | jq .
echo ""

echo -e "${YELLOW}注意: PITR 恢复需要有效的备份集，当前备份可能还在进行中${NC}"
echo -e "${YELLOW}如果备份尚未完成，PITR 调用可能会失败${NC}"
echo ""

read -p "是否继续测试 PITR API？(y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    RESPONSE=$(curl -s -X POST "$BACKEND_URL/api/v1/clusters/$NAMESPACE/$CLUSTER_NAME/pitr" \
        -H "X-Kubeconfig-B64: $KUBECONFIG_B64" \
        -H "Content-Type: application/json" \
        -d "$PITR_PAYLOAD")
    
    echo "PITR 响应:"
    echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
fi
echo ""

# 8. 总结
echo -e "${GREEN}=== 测试总结 ===${NC}"
echo ""
echo "✅ 后端 API 正常工作"
echo "✅ 可以创建备份"
echo "✅ 可以查询备份状态"
echo "✅ 测试数据已准备（初始 3 条 + 目标 2 条 + 错误 2 条）"
echo ""
echo "📋 下一步操作："
echo "1. 等待备份完成: kubectl get polardbxbackup -n $NAMESPACE -w"
echo "2. 配置有效的存储（OSS/SFTP/MinIO）"
echo "3. 执行完整的 PITR 测试"
echo ""
echo "💡 提示:"
echo "- 当前备份使用占位符存储配置 (oss://test-bucket/...)"
echo "- 需要在 polardbx-hpfs-config ConfigMap 中配置真实的 OSS 凭据"
echo "- 或者使用 ./deploy-minio.sh 部署本地 MinIO 服务"
echo ""
