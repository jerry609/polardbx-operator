#!/bin/bash
# 简化版 HPFS Bug 测试脚本
# 适用于已有集群的快速测试

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

echo "=============================================="
echo "HPFS Bug 快速测试"
echo "=============================================="
echo ""

# 检查环境
log_info "检查环境..."
if ! kubectl cluster-info &> /dev/null; then
    log_error "无法连接到 Kubernetes 集群"
    exit 1
fi

# 查找 HPFS pod
HPFS_POD=$(kubectl get pod -n polardbx-operator-system -l app.kubernetes.io/component=polardbx-hpfs -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
if [ -z "$HPFS_POD" ]; then
    log_error "HPFS pod 未找到"
    kubectl get pods -n polardbx-operator-system
    exit 1
fi

log_info "HPFS Pod: $HPFS_POD"

# 获取初始重启次数
INITIAL_RESTARTS=$(kubectl get pod $HPFS_POD -n polardbx-operator-system -o jsonpath='{.status.containerStatuses[0].restartCount}')
log_info "当前重启次数: $INITIAL_RESTARTS"

if [ "$INITIAL_RESTARTS" -gt "0" ]; then
    log_warn "⚠️  HPFS 已经重启过 $INITIAL_RESTARTS 次"
    log_info "检查上次崩溃原因..."
    
    if kubectl logs $HPFS_POD -n polardbx-operator-system --previous 2>/dev/null | grep -q "panic"; then
        log_error "✗ 在上次日志中发现 panic！"
        echo ""
        echo "Panic 详情:"
        kubectl logs $HPFS_POD -n polardbx-operator-system --previous 2>/dev/null | grep -A 10 "panic"
        echo ""
        log_error "Bug 已触发过！详见上面的 panic stack trace"
        exit 1
    fi
fi

echo ""

# 检查现有集群
log_info "查找现有 PolarDB-X 集群..."
CLUSTERS=$(kubectl get polardbxcluster -n default -o jsonpath='{.items[*].metadata.name}' 2>/dev/null)

if [ -z "$CLUSTERS" ]; then
    log_warn "未找到现有集群"
    log_info "请先创建一个集群，然后创建备份来测试"
    exit 0
fi

log_info "找到集群: $CLUSTERS"
CLUSTER_NAME=$(echo $CLUSTERS | awk '{print $1}')
log_info "将使用集群: $CLUSTER_NAME"

echo ""

# 检查集群状态
CLUSTER_PHASE=$(kubectl get polardbxcluster $CLUSTER_NAME -n default -o jsonpath='{.status.phase}' 2>/dev/null)
if [ "$CLUSTER_PHASE" != "Running" ]; then
    log_warn "集群状态: $CLUSTER_PHASE (不是 Running)"
    log_info "等待集群就绪..."
fi

echo ""

# 提示用户创建备份
log_warn "准备测试 Bug..."
log_info ""
log_info "现在将创建一个备份来触发并发上传"
log_info "这可能会导致 HPFS 崩溃（这正是我们要验证的）"
log_info ""
read -p "按 Enter 继续，或 Ctrl+C 取消..."

echo ""

# 创建备份
BACKUP_NAME="test-bug-$(date +%s)"
log_info "创建备份: $BACKUP_NAME"

kubectl apply -f - <<EOF
apiVersion: polardbx.aliyun.com/v1
kind: PolarDBXBackup
metadata:
  name: $BACKUP_NAME
  namespace: default
spec:
  cluster:
    name: $CLUSTER_NAME
  storageProvider:
    storageName: s3
    sink: s3
EOF

echo ""
log_info "备份已创建，监控 HPFS 状态（30秒）..."

# 监控 HPFS 重启
BUG_TRIGGERED=false
for i in {1..30}; do
    CURRENT_RESTARTS=$(kubectl get pod $HPFS_POD -n polardbx-operator-system -o jsonpath='{.status.containerStatuses[0].restartCount}' 2>/dev/null || echo "$INITIAL_RESTARTS")
    
    if [ "$CURRENT_RESTARTS" -gt "$INITIAL_RESTARTS" ]; then
        log_error "⚠️⚠️⚠️  HPFS Pod 重启了！"
        log_error "初始: $INITIAL_RESTARTS, 当前: $CURRENT_RESTARTS"
        log_error "增加: $((CURRENT_RESTARTS - INITIAL_RESTARTS)) 次"
        BUG_TRIGGERED=true
        break
    fi
    
    echo -n "."
    sleep 1
done

echo ""
echo ""

# 检查结果
if [ "$BUG_TRIGGERED" = true ]; then
    log_error "=========================================="
    log_error "  ✗ Bug 已触发！HPFS 重启了"
    log_error "=========================================="
    echo ""
    
    log_info "等待 pod 重新启动..."
    sleep 5
    
    # 获取新的 pod 名称（如果 pod 名称改变）
    NEW_HPFS_POD=$(kubectl get pod -n polardbx-operator-system -l app.kubernetes.io/component=polardbx-hpfs -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    
    log_info "检查崩溃日志..."
    echo "---"
    kubectl logs $NEW_HPFS_POD -n polardbx-operator-system --previous 2>/dev/null | tail -30 || \
        kubectl logs $HPFS_POD -n polardbx-operator-system --previous 2>/dev/null | tail -30
    echo "---"
    
    echo ""
    log_error "这证实了 HPFS Flow Control Bug 存在！"
    log_info "详细分析请查看:"
    log_info "  - HPFS_FLOW_CONTROL_BUG_ANALYSIS.md"
    log_info "  - HPFS_BUG_FIX_GUIDE.md"
    
else
    log_info "=========================================="
    log_info "  ✓ Bug 未在30秒内触发"
    log_info "=========================================="
    echo ""
    
    # 检查备份状态
    BACKUP_PHASE=$(kubectl get polardbxbackup $BACKUP_NAME -n default -o jsonpath='{.status.phase}' 2>/dev/null)
    log_info "备份状态: $BACKUP_PHASE"
    
    if [ "$BACKUP_PHASE" = "Failed" ]; then
        log_warn "备份失败了，但 HPFS 未重启"
        log_info "检查备份 job 日志..."
        
        BACKUP_JOBS=$(kubectl get pods -n default -o jsonpath='{.items[?(@.metadata.labels.polardbx/backup=="'$BACKUP_NAME'")].metadata.name}' 2>/dev/null)
        for job in $BACKUP_JOBS; do
            log_info "Job: $job"
            kubectl logs $job -n default --tail=20 2>/dev/null || true
        done
    fi
    
    log_info ""
    log_info "可能的原因:"
    log_info "  1. 时间不够（可能需要等待更久）"
    log_info "  2. 数据量太小，任务完成太快"
    log_info "  3. Bug 只在特定条件下触发"
    log_info "  4. 已经被修复（检查代码版本）"
fi

echo ""
log_info "测试完成！"
log_info "备份名称: $BACKUP_NAME"
log_info ""
log_info "清理测试备份: kubectl delete polardbxbackup $BACKUP_NAME -n default"
