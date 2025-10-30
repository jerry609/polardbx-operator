#!/bin/bash
# HPFS Flow Control Bug 自动复现脚本
# 用途: 验证 Bug 是否存在，并收集诊断信息

set -e

echo "=============================================="
echo "HPFS Flow Control Bug 复现测试"
echo "====================================# 启动日志监控
echo "步骤 3: 启动日志监控"
echo "----------------------------"

LOG_FILE="/tmp/hpfs-bug-test-$(date +%s).log"
log_info "HPFS 日志将保存到: $LOG_FILE"

# 在后台监控 HPFS 日志
kubectl logs -f -n $NAMESPACE_OPERATOR $HPFS_POD --tail=100 > $LOG_FILE 2>&1 &
LOG_PID=$!
log_info "日志监控进程: $LOG_PID"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
NAMESPACE_OPERATOR="polardbx-operator-system"
NAMESPACE_TEST="default"
CLUSTER_NAME="test-bug-reproduction"
BACKUP_NAME="test-concurrent-backup-$(date +%s)"

# 函数: 打印带颜色的消息
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 函数: 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "需要 $1 命令，请先安装"
        exit 1
    fi
}

# 函数: 等待资源就绪
wait_for_resource() {
    local resource=$1
    local name=$2
    local namespace=$3
    local timeout=${4:-300}
    
    log_info "等待 $resource/$name 就绪..."
    kubectl wait --for=condition=ready $resource/$name -n $namespace --timeout=${timeout}s || {
        log_error "$resource/$name 未就绪"
        return 1
    }
    log_info "$resource/$name 已就绪"
}

# 步骤 0: 检查前置条件
echo "步骤 0: 检查环境"
echo "----------------------------"

check_command kubectl
check_command minikube

# 检查集群连接
if ! kubectl cluster-info &> /dev/null; then
    log_error "无法连接到 Kubernetes 集群"
    exit 1
fi
log_info "Kubernetes 集群连接正常"

# 检查 Operator 是否运行
if ! kubectl get deployment polardbx-controller-manager -n $NAMESPACE_OPERATOR &> /dev/null; then
    log_error "PolarDB-X Operator 未安装"
    exit 1
fi
log_info "PolarDB-X Operator 已安装"

# 检查 HPFS 初始状态
HPFS_POD=$(kubectl get pod -n $NAMESPACE_OPERATOR -l app.kubernetes.io/component=polardbx-hpfs -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
if [ -z "$HPFS_POD" ]; then
    log_error "HPFS pod 未找到"
    log_info "尝试查找所有 HPFS 相关 pods..."
    kubectl get pods -n $NAMESPACE_OPERATOR | grep hpfs || true
    exit 1
fi

INITIAL_RESTARTS=$(kubectl get pod $HPFS_POD -n $NAMESPACE_OPERATOR -o jsonpath='{.status.containerStatuses[0].restartCount}')
log_info "HPFS pod: $HPFS_POD (当前重启次数: $INITIAL_RESTARTS)"

echo ""

# 步骤 1: 检查 MinIO
echo "步骤 1: 检查 MinIO"
echo "----------------------------"

if kubectl get svc minio -n default &> /dev/null; then
    log_info "MinIO 服务已存在"
else
    log_warn "MinIO 服务不存在，请先部署 MinIO"
    log_info "参考命令: kubectl apply -f minio-deployment.yaml"
    exit 1
fi

# 测试 MinIO 连接
log_info "测试 MinIO 连接..."
if kubectl run test-minio-conn --image=curlimages/curl:latest --rm -i --restart=Never -- \
    curl -s -o /dev/null -w "%{http_code}" http://minio.default.svc.cluster.local:9000 | grep -q "403\|200"; then
    log_info "MinIO 可访问"
else
    log_error "MinIO 不可访问"
    exit 1
fi

echo ""

# 步骤 2: 创建测试集群
echo "步骤 2: 创建测试集群"
echo "----------------------------"

if kubectl get polardbxcluster $CLUSTER_NAME -n $NAMESPACE_TEST &> /dev/null; then
    log_warn "测试集群已存在，删除旧集群..."
    kubectl delete polardbxcluster $CLUSTER_NAME -n $NAMESPACE_TEST --wait=false
    sleep 10
fi

log_info "创建测试集群: $CLUSTER_NAME"
kubectl apply -f - <<EOF
apiVersion: polardbx.aliyun.com/v1
kind: PolarDBXCluster
metadata:
  name: $CLUSTER_NAME
  namespace: $NAMESPACE_TEST
spec:
  topology:
    nodes:
      gms:
        replicas: 1
        template:
          imagePullPolicy: IfNotPresent
          resources:
            requests:
              cpu: 500m
              memory: 1Gi
            limits:
              cpu: 1000m
              memory: 2Gi
      dn:
        replicas: 1
        template:
          imagePullPolicy: IfNotPresent
          resources:
            requests:
              cpu: 500m
              memory: 1Gi
            limits:
              cpu: 1000m
              memory: 2Gi
      cn:
        replicas: 1
        template:
          resources:
            requests:
              cpu: 250m
              memory: 512Mi
EOF

# 等待集群就绪
log_info "等待集群就绪（可能需要 5-10 分钟）..."
sleep 30

# 检查 GMS 和 DN pods
GMS_POD=$(kubectl get pod -n $NAMESPACE_TEST -l polardbx/name=$CLUSTER_NAME,polardbx/role=gms -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
DN_POD=$(kubectl get pod -n $NAMESPACE_TEST -l polardbx/name=$CLUSTER_NAME,polardbx/role=dn -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

if [ -z "$GMS_POD" ] || [ -z "$DN_POD" ]; then
    log_error "集群 Pod 未创建"
    kubectl get pods -n $NAMESPACE_TEST | grep $CLUSTER_NAME
    exit 1
fi

log_info "GMS Pod: $GMS_POD"
log_info "DN Pod: $DN_POD"

# 等待 pods 就绪
for i in {1..60}; do
    GMS_READY=$(kubectl get pod $GMS_POD -n $NAMESPACE_TEST -o jsonpath='{.status.containerStatuses[?(@.name=="engine")].ready}' 2>/dev/null || echo "false")
    DN_READY=$(kubectl get pod $DN_POD -n $NAMESPACE_TEST -o jsonpath='{.status.containerStatuses[?(@.name=="engine")].ready}' 2>/dev/null || echo "false")
    
    if [ "$GMS_READY" = "true" ] && [ "$DN_READY" = "true" ]; then
        log_info "集群 Pods 已就绪"
        break
    fi
    
    echo -n "."
    sleep 5
    
    if [ $i -eq 60 ]; then
        log_error "集群 Pods 未在预期时间内就绪"
        exit 1
    fi
done

echo ""
echo ""

# 步骤 3: 启动日志监控
echo "步骤 3: 启动日志监控"
echo "----------------------------"

LOG_FILE="/tmp/hpfs-bug-test-$(date +%s).log"
log_info "HPFS 日志将保存到: $LOG_FILE"

# 在后台监控 HPFS 日志
kubectl logs -f -n $NAMESPACE_OPERATOR $HPFS_POD > $LOG_FILE 2>&1 &
LOG_PID=$!
log_info "日志监控进程: $LOG_PID"

# 确保退出时停止日志监控
trap "kill $LOG_PID 2>/dev/null || true" EXIT

echo ""

# 步骤 4: 触发并发备份
echo "步骤 4: 触发并发备份 (关键步骤)"
echo "----------------------------"

log_warn "即将创建备份，这将触发 GMS 和 DN 并发上传..."
sleep 2

log_info "创建备份: $BACKUP_NAME"
kubectl apply -f - <<EOF
apiVersion: polardbx.aliyun.com/v1
kind: PolarDBXBackup
metadata:
  name: $BACKUP_NAME
  namespace: $NAMESPACE_TEST
spec:
  cluster:
    name: $CLUSTER_NAME
  storageProvider:
    storageName: s3
    sink: s3
EOF

echo ""

# 步骤 5: 监控 Bug 触发
echo "步骤 5: 监控 Bug 触发"
echo "----------------------------"

log_info "监控备份进度和 HPFS 状态（60秒）..."

BUG_TRIGGERED=false
for i in {1..60}; do
    # 检查 HPFS 重启次数
    CURRENT_RESTARTS=$(kubectl get pod $HPFS_POD -n $NAMESPACE_OPERATOR -o jsonpath='{.status.containerStatuses[0].restartCount}' 2>/dev/null || echo "$INITIAL_RESTARTS")
    
    if [ "$CURRENT_RESTARTS" -gt "$INITIAL_RESTARTS" ]; then
        log_error "⚠️  检测到 HPFS pod 重启！"
        log_error "   初始重启次数: $INITIAL_RESTARTS"
        log_error "   当前重启次数: $CURRENT_RESTARTS"
        log_error "   增加: $((CURRENT_RESTARTS - INITIAL_RESTARTS)) 次"
        BUG_TRIGGERED=true
        break
    fi
    
    # 检查备份状态
    BACKUP_PHASE=$(kubectl get polardbxbackup $BACKUP_NAME -n $NAMESPACE_TEST -o jsonpath='{.status.phase}' 2>/dev/null || echo "Unknown")
    
    if [ "$BACKUP_PHASE" = "Failed" ]; then
        log_error "备份失败！"
        BUG_TRIGGERED=true
        break
    fi
    
    if [ "$BACKUP_PHASE" = "Finished" ]; then
        log_info "备份成功完成"
        break
    fi
    
    echo -n "."
    sleep 1
done

echo ""
echo ""

# 步骤 6: 收集诊断信息
echo "步骤 6: 收集诊断信息"
echo "=============================================="

# 停止日志监控
kill $LOG_PID 2>/dev/null || true
sleep 1

# 检查日志中的 panic
log_info "检查 HPFS 日志中的 panic..."
if grep -i "panic" $LOG_FILE; then
    log_error "✗ 发现 panic 日志！"
    echo ""
    log_info "完整 panic stack trace:"
    grep -A 20 "panic" $LOG_FILE
    BUG_TRIGGERED=true
else
    log_info "✓ 未发现 panic（但不代表 bug 未触发）"
fi

echo ""

# 检查备份任务状态
log_info "备份任务最终状态:"
kubectl get polardbxbackup $BACKUP_NAME -n $NAMESPACE_TEST

echo ""

# 检查 XStore 备份状态
log_info "XStore 备份状态:"
kubectl get xstorebackup -n $NAMESPACE_TEST | grep $BACKUP_NAME || echo "无 XStoreBackup"

echo ""

# 检查备份 Job Pods
log_info "备份 Job Pods:"
kubectl get pods -n $NAMESPACE_TEST | grep backup-job || echo "无 backup job pods"

echo ""

# 检查失败的 Job 日志
FAILED_JOBS=$(kubectl get pods -n $NAMESPACE_TEST -o jsonpath='{range .items[?(@.status.phase=="Failed")]}{.metadata.name}{"\n"}{end}' | grep backup-job || true)
if [ ! -z "$FAILED_JOBS" ]; then
    log_error "发现失败的备份 Job:"
    echo "$FAILED_JOBS"
    
    for job in $FAILED_JOBS; do
        log_info "获取 $job 的日志..."
        echo "---"
        kubectl logs $job -n $NAMESPACE_TEST --tail=30 || true
        echo "---"
    done
fi

echo ""

# 步骤 7: 生成报告
echo "步骤 7: 测试报告"
echo "=============================================="

REPORT_FILE="/tmp/hpfs-bug-report-$(date +%s).txt"

cat > $REPORT_FILE <<REPORT_EOF
HPFS Flow Control Bug 测试报告
========================================

测试时间: $(date)
测试环境:
  - Kubernetes: $(kubectl version --short 2>/dev/null | head -2 || echo "unknown")
  - PolarDB-X Operator: $(kubectl get deployment polardbx-controller-manager -n $NAMESPACE_OPERATOR -o jsonpath='{.spec.template.spec.containers[0].image}')
  - HPFS Pod: $HPFS_POD

测试集群:
  - 名称: $CLUSTER_NAME
  - GMS Pod: $GMS_POD
  - DN Pod: $DN_POD

备份任务:
  - 名称: $BACKUP_NAME
  - 最终状态: $(kubectl get polardbxbackup $BACKUP_NAME -n $NAMESPACE_TEST -o jsonpath='{.status.phase}' 2>/dev/null || echo "Unknown")

HPFS 状态:
  - 初始重启次数: $INITIAL_RESTARTS
  - 最终重启次数: $CURRENT_RESTARTS
  - 重启增加: $((CURRENT_RESTARTS - INITIAL_RESTARTS))

Bug 触发情况:
  - $([ "$BUG_TRIGGERED" = true ] && echo "✗ Bug 已触发" || echo "✓ Bug 未触发（或测试时间不够）")

日志文件:
  - HPFS 日志: $LOG_FILE
  - 本报告: $REPORT_FILE

========================================
REPORT_EOF

cat $REPORT_FILE

if [ "$BUG_TRIGGERED" = true ]; then
    echo ""
    log_error "=========================================="
    log_error "  ✗ Bug 复现成功！"
    log_error "=========================================="
    log_error ""
    log_error "此环境中确认存在 HPFS Flow Control 并发 Bug"
    log_error "详细分析请查看:"
    log_error "  - HPFS_FLOW_CONTROL_BUG_ANALYSIS.md"
    log_error "  - HPFS_BUG_FIX_GUIDE.md"
    log_error ""
    log_error "诊断文件:"
    log_error "  - $LOG_FILE"
    log_error "  - $REPORT_FILE"
    echo ""
    exit 1
else
    echo ""
    log_info "=========================================="
    log_info "  ✓ Bug 未触发"
    log_info "=========================================="
    log_info ""
    log_info "可能的原因:"
    log_info "  1. 测试时间不够（某些情况下需要多次尝试）"
    log_info "  2. 数据量太小，任务完成太快"
    log_info "  3. Bug 已被修复（检查 HPFS 代码版本）"
    log_info "  4. 随机性因素（建议重试）"
    echo ""
    log_info "建议: 重新运行此脚本或手动创建更多备份任务"
    echo ""
    exit 0
fi
