#!/bin/bash
# 手动测试步骤 - 逐步执行

echo "=== 手动 HPFS Bug 测试步骤 ==="
echo ""
echo "步骤 1: 记录 HPFS 初始状态"
echo "----------------------------"
echo "命令: kubectl get pods -n polardbx-operator-system | grep hpfs"
kubectl get pods -n polardbx-operator-system | grep hpfs
echo ""

HPFS_POD=$(kubectl get pod -n polardbx-operator-system -l app.kubernetes.io/component=polardbx-hpfs -o jsonpath='{.items[0].metadata.name}')
INITIAL_RESTARTS=$(kubectl get pod $HPFS_POD -n polardbx-operator-system -o jsonpath='{.status.containerStatuses[0].restartCount}')

echo "HPFS Pod: $HPFS_POD"
echo "初始重启次数: $INITIAL_RESTARTS"
echo ""
echo "按 Enter 继续到下一步..."
read

echo ""
echo "步骤 2: 在另一个终端监控 HPFS 日志"
echo "----------------------------"
echo "打开新终端，运行:"
echo "  kubectl logs -f -n polardbx-operator-system $HPFS_POD"
echo ""
echo "准备好后按 Enter 继续..."
read

echo ""
echo "步骤 3: 创建备份（触发并发上传）"
echo "----------------------------"

BACKUP_NAME="manual-test-$(date +%s)"
echo "备份名称: $BACKUP_NAME"
echo ""
echo "将创建以下备份:"
cat <<EOF
apiVersion: polardbx.aliyun.com/v1
kind: PolarDBXBackup
metadata:
  name: $BACKUP_NAME
  namespace: default
spec:
  cluster:
    name: test-pitr
  storageProvider:
    storageName: s3
    sink: s3
EOF

echo ""
echo "按 Enter 创建备份..."
read

kubectl apply -f - <<EOF
apiVersion: polardbx.aliyun.com/v1
kind: PolarDBXBackup
metadata:
  name: $BACKUP_NAME
  namespace: default
spec:
  cluster:
    name: test-pitr
  storageProvider:
    storageName: s3
    sink: s3
EOF

echo ""
echo "✓ 备份已创建"
echo ""

echo "步骤 4: 观察现象"
echo "----------------------------"
echo "在你的日志终端中，观察:"
echo "  1. 是否出现两个 upload 请求（GMS 和 DN）"
echo "  2. 是否出现 'panic: send on closed channel'"
echo "  3. Pod 是否重启"
echo ""
echo "同时监控备份 Job:"
echo "  watch kubectl get polardbxbackup,xstorebackup,jobs,pods -n default"
echo ""
echo "按 Enter 检查 HPFS 状态..."
read

echo ""
echo "步骤 5: 检查结果"
echo "----------------------------"

CURRENT_RESTARTS=$(kubectl get pod -n polardbx-operator-system -l app.kubernetes.io/component=polardbx-hpfs -o jsonpath='{.items[0].status.containerStatuses[0].restartCount}')

echo "当前重启次数: $CURRENT_RESTARTS"
echo "初始重启次数: $INITIAL_RESTARTS"

if [ "$CURRENT_RESTARTS" -gt "$INITIAL_RESTARTS" ]; then
    echo ""
    echo "❌ HPFS Pod 重启了 $((CURRENT_RESTARTS - INITIAL_RESTARTS)) 次！"
    echo ""
    echo "获取崩溃日志:"
    kubectl logs $HPFS_POD -n polardbx-operator-system --previous 2>/dev/null | tail -30
    echo ""
    echo "✗ Bug 已触发！"
else
    echo ""
    echo "✓ HPFS 未重启（但可能需要等待更长时间）"
fi

echo ""
echo "步骤 6: 检查备份状态"
echo "----------------------------"
kubectl get polardbxbackup $BACKUP_NAME -n default
echo ""

echo "步骤 7: 清理"
echo "----------------------------"
echo "删除测试备份:"
echo "  kubectl delete polardbxbackup $BACKUP_NAME -n default"
echo ""
echo "测试完成！"
