#!/bin/bash

# PolarDB-X 完整备份和 PITR 测试
# 使用 MinIO 作为存储后端

set -e

BACKEND_URL="http://localhost:8080"
NAMESPACE="default"
CLUSTER_NAME="test-pitr"
KUBECONFIG_B64=$(kubectl config view --raw --minify | base64 -w 0)
MINIO_ENDPOINT="http://minio.default.svc.cluster.local:9000"
MINIO_SINK_NAME=${MINIO_SINK_NAME:-s3}

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}╔═══════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  PolarDB-X 备份和 PITR 完整测试              ║${NC}"
echo -e "${BLUE}║  存储: MinIO (S3 Compatible)                 ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════╝${NC}"
echo ""

# 步骤 1: 准备测试数据
echo -e "${BLUE}═══ 步骤 1/8: 准备测试数据 ═══${NC}"
CN_POD=$(kubectl get pods -n $NAMESPACE -l polardbx/name=$CLUSTER_NAME,polardbx/role=cn -o jsonpath='{.items[0].metadata.name}')
echo "CN Pod: $CN_POD"

kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root << 'EOF'
CREATE DATABASE IF NOT EXISTS backup_test mode='auto';
USE backup_test;

DROP TABLE IF EXISTS test_data;

CREATE TABLE test_data (
    id INT AUTO_INCREMENT PRIMARY KEY,
    data VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

INSERT INTO test_data (data) VALUES
    ('Initial-1: Before backup'),
    ('Initial-2: Before backup'),
    ('Initial-3: Before backup');

SELECT '=== 初始数据 ===' as stage;
SELECT * FROM test_data;
EOF

echo -e "${GREEN}✓ 测试数据已创建（3条初始记录）${NC}"
echo ""

# 步骤 2: 创建全量备份
echo -e "${BLUE}═══ 步骤 2/8: 创建全量备份 ═══${NC}"
BACKUP_NAME="full-backup-$(date +%Y%m%d-%H%M%S)"
echo "备份名称: $BACKUP_NAME"

BACKUP_PAYLOAD=$(cat <<EOF
{
  "metadata": {
    "name": "$BACKUP_NAME"
  },
    "spec": {
        "retentionTime": "168h",
        "cleanPolicy": "Retain",
        "storageProvider": {
            "storageName": "s3",
            "sink": "$MINIO_SINK_NAME"
        },
    "preferredBackupRole": "follower"
  }
}
EOF
)

RESPONSE=$(curl -s -X POST "$BACKEND_URL/api/v1/clusters/$NAMESPACE/$CLUSTER_NAME/backups" \
    -H "X-Kubeconfig-B64: $KUBECONFIG_B64" \
    -H "Content-Type: application/json" \
    -d "$BACKUP_PAYLOAD")

if echo "$RESPONSE" | jq -e '.metadata.name' >/dev/null 2>&1; then
    echo -e "${GREEN}✓ 备份创建成功${NC}"
    echo "$RESPONSE" | jq -r '"\(.metadata.name) - Phase: \(.status.phase // "Pending")"'
else
    echo -e "${RED}✗ 备份创建失败${NC}"
    echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
    exit 1
fi
echo ""

# 步骤 3: 等待备份完成
echo -e "${BLUE}═══ 步骤 3/8: 等待备份完成 ═══${NC}"
echo "等待备份完成（最长10分钟）..."

MAX_WAIT=600
ELAPSED=0
while [ $ELAPSED -lt $MAX_WAIT ]; do
    PHASE=$(kubectl get polardbxbackup $BACKUP_NAME -n $NAMESPACE -o jsonpath='{.status.phase}' 2>/dev/null || echo "Unknown")
    
    case $PHASE in
        "Completed"|"Finished")
            echo -e "\n${GREEN}✓ 备份完成！${NC}"
            kubectl get polardbxbackup $BACKUP_NAME -n $NAMESPACE
            break
            ;;
        "Failed")
            echo -e "\n${RED}✗ 备份失败${NC}"
            kubectl describe polardbxbackup $BACKUP_NAME -n $NAMESPACE | tail -30
            exit 1
            ;;
        *)
            echo -n "."
            sleep 10
            ELAPSED=$((ELAPSED + 10))
            ;;
    esac
done

if [ $ELAPSED -ge $MAX_WAIT ]; then
    echo -e "\n${YELLOW}⚠ 备份超时，当前状态：${NC}"
    kubectl get polardbxbackup $BACKUP_NAME -n $NAMESPACE
    kubectl describe polardbxbackup $BACKUP_NAME -n $NAMESPACE | tail -30
    echo ""
    read -p "备份可能仍在进行中，是否继续？(y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi
echo ""

# 步骤 4: 创建 Binlog 备份
echo -e "${BLUE}═══ 步骤 4/8: 启动 Binlog 备份 ═══${NC}"
BINLOG_NAME="binlog-backup-$CLUSTER_NAME"
CLUSTER_UID=$(kubectl get polardbxcluster $CLUSTER_NAME -n $NAMESPACE -o jsonpath='{.metadata.uid}')

BINLOG_PAYLOAD=$(cat <<EOF
{
  "metadata": {
    "name": "$BINLOG_NAME"
  },
    "spec": {
        "pxcName": "$CLUSTER_NAME",
        "pxcUid": "$CLUSTER_UID",
        "storageProvider": {
            "storageName": "s3",
            "sink": "$MINIO_SINK_NAME"
        }
  }
}
EOF
)

# 检查是否已存在
if kubectl get polardbxbackupbinlog $BINLOG_NAME -n $NAMESPACE >/dev/null 2>&1; then
    echo -e "${YELLOW}Binlog 备份已存在，跳过创建${NC}"
else
    kubectl apply -f - <<EOF
apiVersion: polardbx.aliyun.com/v1
kind: PolarDBXBackupBinlog
metadata:
  name: $BINLOG_NAME
  namespace: $NAMESPACE
spec:
    pxcName: $CLUSTER_NAME
    pxcUid: $CLUSTER_UID
    storageProvider:
        storageName: s3
        sink: $MINIO_SINK_NAME
EOF
    echo -e "${GREEN}✓ Binlog 备份已启动${NC}"
fi
echo ""

# 步骤 5: 添加更多数据（PITR 目标点）
echo -e "${BLUE}═══ 步骤 5/8: 添加 PITR 目标数据 ═══${NC}"
sleep 5

kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root backup_test << 'EOF'
INSERT INTO test_data (data) VALUES
    ('Target-1: At PITR point'),
    ('Target-2: At PITR point'),
    ('Target-3: At PITR point');

SELECT '=== PITR 目标点数据 ===' as stage;
SELECT * FROM test_data;
SELECT NOW() as pitr_target_time;
EOF

# 记录 PITR 时间
PITR_TIME=$(kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root -N backup_test -e "SELECT NOW()")
PITR_TIME_ISO=$(date -d "$PITR_TIME" -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "$PITR_TIME")

echo -e "${GREEN}✓ PITR 目标时间: $PITR_TIME_ISO${NC}"
echo "$PITR_TIME_ISO" > /tmp/pitr_test_time.txt
echo ""

# 步骤 6: 添加错误数据
echo -e "${BLUE}═══ 步骤 6/8: 添加错误数据（将被回滚）═══${NC}"
sleep 5

kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root backup_test << 'EOF'
INSERT INTO test_data (data) VALUES
    ('ERROR-1: Should be rolled back'),
    ('ERROR-2: Should be rolled back'),
    ('ERROR-3: Should be rolled back');

SELECT '=== 当前所有数据（包含错误）===' as stage;
SELECT * FROM test_data;
SELECT COUNT(*) as total FROM test_data;
EOF

echo -e "${YELLOW}⚠ 当前有 9 条记录（包含 3 条需要回滚的错误数据）${NC}"
echo ""

# 步骤 7: 执行 PITR 恢复
echo -e "${BLUE}═══ 步骤 7/8: 执行 PITR 恢复 ═══${NC}"
TARGET_CLUSTER="${CLUSTER_NAME}-restored"
echo "目标集群: $TARGET_CLUSTER"
echo "恢复到时间点: $PITR_TIME_ISO"

PITR_PAYLOAD=$(cat <<EOF
{
  "time": "$PITR_TIME_ISO",
  "targetCluster": "$TARGET_CLUSTER",
  "backupSet": "$BACKUP_NAME",
  "timezone": "UTC"
}
EOF
)

# 检查目标集群是否已存在
if kubectl get polardbxcluster $TARGET_CLUSTER -n $NAMESPACE >/dev/null 2>&1; then
    echo -e "${YELLOW}目标集群已存在，先删除...${NC}"
    kubectl delete polardbxcluster $TARGET_CLUSTER -n $NAMESPACE --wait=false
    sleep 10
fi

RESPONSE=$(curl -s -X POST "$BACKEND_URL/api/v1/clusters/$NAMESPACE/$CLUSTER_NAME/pitr" \
    -H "X-Kubeconfig-B64: $KUBECONFIG_B64" \
    -H "Content-Type: application/json" \
    -d "$PITR_PAYLOAD")

if echo "$RESPONSE" | jq -e '.targetCluster' >/dev/null 2>&1; then
    echo -e "${GREEN}✓ PITR 恢复已启动${NC}"
    echo "$RESPONSE" | jq '{targetCluster, phase, stage}'
else
    echo -e "${RED}✗ PITR 启动失败${NC}"
    echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
    exit 1
fi
echo ""

# 步骤 8: 等待 PITR 完成并验证
echo -e "${BLUE}═══ 步骤 8/8: 等待恢复完成并验证数据 ═══${NC}"
echo "等待集群恢复（最长15分钟）..."

MAX_WAIT=900
ELAPSED=0
while [ $ELAPSED -lt $MAX_WAIT ]; do
    if kubectl get polardbxcluster $TARGET_CLUSTER -n $NAMESPACE >/dev/null 2>&1; then
        PHASE=$(kubectl get polardbxcluster $TARGET_CLUSTER -n $NAMESPACE -o jsonpath='{.status.phase}')
        
        case $PHASE in
            "Running")
                echo -e "\n${GREEN}✓ 集群恢复完成！${NC}"
                kubectl get polardbxcluster $TARGET_CLUSTER -n $NAMESPACE
                break
                ;;
            "Failed")
                echo -e "\n${RED}✗ 集群恢复失败${NC}"
                kubectl describe polardbxcluster $TARGET_CLUSTER -n $NAMESPACE | tail -30
                exit 1
                ;;
            *)
                echo -n "."
                sleep 15
                ELAPSED=$((ELAPSED + 15))
                ;;
        esac
    else
        echo -n "."
        sleep 15
        ELAPSED=$((ELAPSED + 15))
    fi
done

if [ $ELAPSED -ge $MAX_WAIT ]; then
    echo -e "\n${YELLOW}⚠ 恢复超时${NC}"
    exit 1
fi
echo ""

# 验证数据
echo -e "${BLUE}═══ 验证恢复的数据 ═══${NC}"
echo "等待 CN Pod 就绪..."
kubectl wait --for=condition=Ready pod -l polardbx/name=$TARGET_CLUSTER,polardbx/role=cn -n $NAMESPACE --timeout=300s || true

RESTORED_CN_POD=$(kubectl get pods -n $NAMESPACE -l polardbx/name=$TARGET_CLUSTER,polardbx/role=cn -o jsonpath='{.items[0].metadata.name}')
echo "恢复集群 CN Pod: $RESTORED_CN_POD"
echo ""

if [ -z "$RESTORED_CN_POD" ]; then
    echo -e "${RED}✗ 未找到恢复集群的 CN Pod${NC}"
    exit 1
fi

echo -e "${BLUE}查询恢复的数据：${NC}"
kubectl exec -n $NAMESPACE $RESTORED_CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root backup_test -e "
    SELECT '=== 恢复后的数据 ===' as result;
    SELECT * FROM test_data ORDER BY id;
    SELECT '---' as separator;
    SELECT COUNT(*) as total_records FROM test_data;
" || {
    echo -e "${YELLOW}⚠ 数据库查询失败，可能还在初始化中${NC}"
    exit 1
}

# 验证记录数
RECORD_COUNT=$(kubectl exec -n $NAMESPACE $RESTORED_CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root -N backup_test -e "SELECT COUNT(*) FROM test_data;" 2>/dev/null || echo "0")

echo ""
echo -e "${BLUE}═══ 测试结果 ═══${NC}"
echo "预期记录数: 6 (3条初始 + 3条目标点)"
echo "实际记录数: $RECORD_COUNT"

if [ "$RECORD_COUNT" -eq 6 ]; then
    # 检查错误数据
    ERROR_COUNT=$(kubectl exec -n $NAMESPACE $RESTORED_CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root -N backup_test -e "SELECT COUNT(*) FROM test_data WHERE data LIKE 'ERROR%';" 2>/dev/null || echo "0")
    
    if [ "$ERROR_COUNT" -eq 0 ]; then
        echo -e "${GREEN}✅ 测试通过！${NC}"
        echo -e "${GREEN}✓ 数据恢复到正确的时间点${NC}"
        echo -e "${GREEN}✓ 错误数据已成功回滚${NC}"
    else
        echo -e "${RED}✗ 测试失败：发现 $ERROR_COUNT 条错误数据${NC}"
    fi
else
    echo -e "${RED}✗ 测试失败：记录数不匹配${NC}"
fi

echo ""
echo -e "${BLUE}═══ 测试总结 ═══${NC}"
echo "原始集群: $CLUSTER_NAME (9条记录)"
echo "恢复集群: $TARGET_CLUSTER ($RECORD_COUNT条记录)"
echo "备份名称: $BACKUP_NAME"
echo "PITR 时间: $PITR_TIME_ISO"
echo ""
echo -e "${BLUE}MinIO 控制台: http://$(minikube ip):30901${NC}"
echo "   用户名: minioadmin"
echo "   密码: minioadmin123"
echo ""

read -p "是否清理测试资源？(y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "清理恢复的集群..."
    kubectl delete polardbxcluster $TARGET_CLUSTER -n $NAMESPACE --wait=false || true
    echo "清理备份..."
    kubectl delete polardbxbackup $BACKUP_NAME -n $NAMESPACE --wait=false || true
    echo -e "${GREEN}✓ 清理完成${NC}"
fi

echo ""
echo -e "${GREEN}测试脚本执行完毕！${NC}"
