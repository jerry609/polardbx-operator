#!/bin/bash

# PolarDB-X 备份和 PITR 自动化测试脚本
# 用法: ./test-backup-pitr.sh [oss|sftp|s3|all]

set -e

# 配置
NAMESPACE="default"
CLUSTER_NAME="test-pitr"
BACKEND_URL="http://localhost:8080/api/v1"
STORAGE_TYPE="${1:-oss}"  # 默认 OSS

# 颜色输出
RED='\033[0:31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖工具..."
    
    for cmd in kubectl curl jq; do
        if ! command -v $cmd &> /dev/null; then
            log_error "$cmd 未安装"
            exit 1
        fi
    done
    
    log_success "所有依赖已安装"
}

# 检查集群状态
check_cluster() {
    log_info "检查集群 $CLUSTER_NAME 状态..."
    
    if ! kubectl get polardbxcluster $CLUSTER_NAME -n $NAMESPACE &>/dev/null; then
        log_error "集群 $CLUSTER_NAME 不存在"
        exit 1
    fi
    
    PHASE=$(kubectl get polardbxcluster $CLUSTER_NAME -n $NAMESPACE -o jsonpath='{.status.phase}')
    if [ "$PHASE" != "Running" ]; then
        log_error "集群状态不是 Running，当前: $PHASE"
        exit 1
    fi
    
    log_success "集群运行正常"
}

# 创建测试数据
create_test_data() {
    log_info "创建测试数据..."
    
    CN_POD=$(kubectl get pods -n $NAMESPACE -l polardbx/name=$CLUSTER_NAME,polardbx/role=cn -o jsonpath='{.items[0].metadata.name}')
    
    kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root << 'EOF'
CREATE DATABASE IF NOT EXISTS backup_test mode='auto';
USE backup_test;

CREATE TABLE IF NOT EXISTS events (
    id INT AUTO_INCREMENT PRIMARY KEY,
    event_type VARCHAR(50),
    description VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

DELETE FROM events;

INSERT INTO events (event_type, description) VALUES
    ('INIT', 'Initial data - before any backup'),
    ('SETUP', 'Setup phase completed'),
    ('CONFIG', 'Configuration applied');
EOF
    
    log_success "测试数据创建完成"
    
    # 显示数据
    kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root backup_test -e "SELECT * FROM events;"
}

# 获取存储配置
get_storage_config() {
    local storage_type=$1
    local backup_name=$2
    
    case $storage_type in
        oss)
            echo '{
                "storageName": "oss",
                "sink": "oss://polardbx-test-bucket/backups/'$CLUSTER_NAME'/'$backup_name'"
            }'
            ;;
        sftp)
            echo '{
                "storageName": "sftp",
                "sink": "sftp://backup-user@sftp-server/polardbx/'$CLUSTER_NAME'/'$backup_name'"
            }'
            ;;
        s3)
            echo '{
                "storageName": "s3",
                "sink": "s3://polardbx-backups/'$CLUSTER_NAME'/'$backup_name'"
            }'
            ;;
        *)
            log_error "不支持的存储类型: $storage_type"
            exit 1
            ;;
    esac
}

# 创建备份
create_backup() {
    local storage_type=$1
    local backup_name="${CLUSTER_NAME}-${storage_type}-backup-$(date +%Y%m%d-%H%M%S)"
    
    log_info "创建 $storage_type 备份: $backup_name"
    
    STORAGE_CONFIG=$(get_storage_config $storage_type $backup_name)
    
    RESPONSE=$(curl -s -X POST $BACKEND_URL/clusters/$NAMESPACE/$CLUSTER_NAME/backups \
        -H "Content-Type: application/json" \
        -d '{
            "metadata": {
                "name": "'$backup_name'"
            },
            "spec": {
                "retentionTime": "168h",
                "cleanPolicy": "Retain",
                "storageProvider": '$STORAGE_CONFIG',
                "preferredBackupRole": "follower"
            }
        }')
    
    if echo "$RESPONSE" | jq -e '.metadata.name' &>/dev/null; then
        log_success "备份创建成功: $backup_name"
        echo $backup_name
    else
        log_error "备份创建失败"
        echo "$RESPONSE" | jq .
        exit 1
    fi
}

# 等待备份完成
wait_for_backup() {
    local backup_name=$1
    local max_wait=600  # 10 分钟超时
    local elapsed=0
    
    log_info "等待备份完成: $backup_name"
    
    while [ $elapsed -lt $max_wait ]; do
        PHASE=$(kubectl get polardbxbackup $backup_name -n $NAMESPACE -o jsonpath='{.status.phase}' 2>/dev/null || echo "")
        
        case $PHASE in
            "Finished"|"Completed")
                log_success "备份完成"
                return 0
                ;;
            "Failed")
                log_error "备份失败"
                kubectl describe polardbxbackup $backup_name -n $NAMESPACE
                return 1
                ;;
            *)
                echo -n "."
                sleep 10
                elapsed=$((elapsed + 10))
                ;;
        esac
    done
    
    log_error "备份超时"
    return 1
}

# 创建 Binlog 备份
create_binlog_backup() {
    local storage_type=$1
    local binlog_name="${CLUSTER_NAME}-${storage_type}-binlog"
    
    log_info "创建 $storage_type Binlog 备份: $binlog_name"
    
    STORAGE_CONFIG=$(get_storage_config $storage_type "binlog/")
    
    CLUSTER_UID=$(kubectl get polardbxcluster $CLUSTER_NAME -n $NAMESPACE -o jsonpath='{.metadata.uid}')
    
    RESPONSE=$(curl -s -X POST $BACKEND_URL/backup-binlogs \
        -H "Content-Type: application/json" \
        -d '{
            "metadata": {
                "name": "'$binlog_name'"
            },
            "spec": {
                "pxcName": "'$CLUSTER_NAME'",
                "pxcUid": "'$CLUSTER_UID'",
                "storageProvider": '$STORAGE_CONFIG'
            }
        }')
    
    if echo "$RESPONSE" | jq -e '.metadata.name' &>/dev/null; then
        log_success "Binlog 备份创建成功: $binlog_name"
        echo $binlog_name
    else
        log_error "Binlog 备份创建失败"
        echo "$RESPONSE" | jq .
        exit 1
    fi
}

# 添加 PITR 目标点数据
add_pitr_target_data() {
    log_info "添加 PITR 目标点数据..."
    
    CN_POD=$(kubectl get pods -n $NAMESPACE -l polardbx/name=$CLUSTER_NAME,polardbx/role=cn -o jsonpath='{.items[0].metadata.name}')
    
    kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root backup_test << 'EOF'
INSERT INTO events (event_type, description) VALUES
    ('BACKUP1', 'Data added after first backup'),
    ('PROCESS', 'Processing started'),
    ('CHECKPOINT', 'This is the PITR target point');
SELECT NOW() as pitr_target_time;
EOF
    
    # 等待几秒确保 binlog 写入
    sleep 5
    
    # 记录 PITR 时间点
    PITR_TIME=$(kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root -N -e "SELECT NOW()")
    log_success "PITR 目标时间点: $PITR_TIME"
    
    # 保存到文件
    echo $PITR_TIME > /tmp/pitr_time.txt
}

# 添加错误数据
add_error_data() {
    log_info "添加需要回滚的错误数据..."
    
    CN_POD=$(kubectl get pods -n $NAMESPACE -l polardbx/name=$CLUSTER_NAME,polardbx/role=cn -o jsonpath='{.items[0].metadata.name}')
    
    sleep 3
    
    kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root backup_test << 'EOF'
INSERT INTO events (event_type, description) VALUES
    ('ERROR', 'Wrong data - should be rolled back'),
    ('MISTAKE', 'Another mistake to revert'),
    ('OOPS', 'This should not exist after PITR');
EOF
    
    log_warning "错误数据已添加（稍后将通过 PITR 回滚）"
    
    # 显示当前所有数据
    kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root backup_test -e "SELECT * FROM events ORDER BY id;"
}

# 执行 PITR
execute_pitr() {
    local storage_type=$1
    local backup_name=$2
    local target_name="${CLUSTER_NAME}-restored-${storage_type}"
    
    log_info "执行 $storage_type PITR 恢复..."
    
    # 读取 PITR 时间
    if [ ! -f /tmp/pitr_time.txt ]; then
        log_error "PITR 时间点未记录"
        exit 1
    fi
    
    PITR_TIME=$(cat /tmp/pitr_time.txt)
    # 转换为 ISO 8601 格式
    PITR_TIME_ISO=$(date -d "$PITR_TIME" -u +"%Y-%m-%dT%H:%M:%SZ")
    
    log_info "PITR 时间点: $PITR_TIME_ISO"
    log_info "目标集群: $target_name"
    
    RESPONSE=$(curl -s -X POST $BACKEND_URL/clusters/$NAMESPACE/$CLUSTER_NAME/pitr \
        -H "Content-Type: application/json" \
        -d '{
            "time": "'$PITR_TIME_ISO'",
            "targetCluster": "'$target_name'",
            "backupSet": "'$backup_name'",
            "timezone": "UTC"
        }')
    
    if echo "$RESPONSE" | jq -e '.targetCluster' &>/dev/null; then
        log_success "PITR 恢复已启动"
        echo "$RESPONSE" | jq .
        echo $target_name
    else
        log_error "PITR 启动失败"
        echo "$RESPONSE" | jq .
        exit 1
    fi
}

# 等待 PITR 完成
wait_for_pitr() {
    local target_name=$1
    local max_wait=900  # 15 分钟超时
    local elapsed=0
    
    log_info "等待 PITR 恢复完成: $target_name"
    
    while [ $elapsed -lt $max_wait ]; do
        if kubectl get polardbxcluster $target_name -n $NAMESPACE &>/dev/null; then
            PHASE=$(kubectl get polardbxcluster $target_name -n $NAMESPACE -o jsonpath='{.status.phase}')
            
            case $PHASE in
                "Running")
                    log_success "PITR 恢复完成，集群运行中"
                    return 0
                    ;;
                "Failed")
                    log_error "PITR 恢复失败"
                    kubectl describe polardbxcluster $target_name -n $NAMESPACE
                    return 1
                    ;;
                *)
                    echo -n "."
                    sleep 15
                    elapsed=$((elapsed + 15))
                    ;;
            esac
        else
            echo -n "."
            sleep 15
            elapsed=$((elapsed + 15))
        fi
    done
    
    log_error "PITR 恢复超时"
    return 1
}

# 验证恢复数据
verify_restored_data() {
    local target_name=$1
    
    log_info "验证恢复的数据..."
    
    # 等待 CN Pod 就绪
    log_info "等待 CN Pod 就绪..."
    kubectl wait --for=condition=Ready pod -l polardbx/name=$target_name,polardbx/role=cn -n $NAMESPACE --timeout=300s || true
    
    CN_POD=$(kubectl get pods -n $NAMESPACE -l polardbx/name=$target_name,polardbx/role=cn -o jsonpath='{.items[0].metadata.name}')
    
    if [ -z "$CN_POD" ]; then
        log_error "未找到 CN Pod"
        return 1
    fi
    
    log_info "连接到 Pod: $CN_POD"
    
    # 查询恢复的数据
    echo ""
    log_info "恢复的数据内容:"
    kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root backup_test -e "
        SELECT * FROM events ORDER BY id;
        SELECT '---' as separator;
        SELECT COUNT(*) as total_records FROM events;
        SELECT '---' as separator;
        SELECT MAX(event_type) as last_event_type FROM events;
    " || log_error "查询失败"
    
    # 验证数据
    RECORD_COUNT=$(kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root -N backup_test -e "SELECT COUNT(*) FROM events;" 2>/dev/null || echo "0")
    
    log_info "记录总数: $RECORD_COUNT"
    
    if [ "$RECORD_COUNT" -eq 6 ]; then
        log_success "数据验证通过！包含 6 条记录（PITR 目标点之前的数据）"
        
        # 检查不应该存在的错误数据
        ERROR_COUNT=$(kubectl exec -n $NAMESPACE $CN_POD -c engine -- mysql -h 127.0.0.1 -P 3306 -u polardbx_root -N backup_test -e "SELECT COUNT(*) FROM events WHERE event_type IN ('ERROR', 'MISTAKE', 'OOPS');" 2>/dev/null || echo "0")
        
        if [ "$ERROR_COUNT" -eq 0 ]; then
            log_success "错误数据已成功回滚！"
            return 0
        else
            log_error "发现 $ERROR_COUNT 条错误数据，PITR 失败"
            return 1
        fi
    else
        log_error "数据验证失败！预期 6 条记录，实际 $RECORD_COUNT 条"
        return 1
    fi
}

# 清理测试资源
cleanup() {
    local storage_type=$1
    
    log_info "清理 $storage_type 测试资源..."
    
    # 删除恢复的集群
    TARGET_NAME="${CLUSTER_NAME}-restored-${storage_type}"
    if kubectl get polardbxcluster $TARGET_NAME -n $NAMESPACE &>/dev/null; then
        log_info "删除恢复的集群: $TARGET_NAME"
        kubectl delete polardbxcluster $TARGET_NAME -n $NAMESPACE --wait=false
    fi
    
    # 删除备份
    BACKUPS=$(kubectl get polardbxbackup -n $NAMESPACE -o name | grep "$storage_type")
    if [ ! -z "$BACKUPS" ]; then
        log_info "删除备份..."
        echo "$BACKUPS" | xargs kubectl delete -n $NAMESPACE --wait=false
    fi
    
    # 删除 Binlog 备份
    BINLOG_BACKUPS=$(kubectl get polardbxbackupbinlog -n $NAMESPACE -o name | grep "$storage_type")
    if [ ! -z "$BINLOG_BACKUPS" ]; then
        log_info "删除 Binlog 备份..."
        echo "$BINLOG_BACKUPS" | xargs kubectl delete -n $NAMESPACE --wait=false
    fi
    
    log_success "清理完成"
}

# 运行单个存储类型的完整测试
run_storage_test() {
    local storage_type=$1
    
    echo ""
    echo "=========================================="
    log_info "开始测试存储类型: $storage_type"
    echo "=========================================="
    
    # 1. 创建备份
    BACKUP_NAME=$(create_backup $storage_type)
    
    # 2. 等待备份完成
    if ! wait_for_backup $BACKUP_NAME; then
        log_error "$storage_type 备份失败"
        return 1
    fi
    
    # 3. 创建 Binlog 备份
    BINLOG_NAME=$(create_binlog_backup $storage_type)
    log_info "Binlog 备份已启动: $BINLOG_NAME"
    
    # 4. 添加 PITR 目标点数据
    add_pitr_target_data
    
    # 5. 添加错误数据
    add_error_data
    
    # 6. 执行 PITR
    TARGET_NAME=$(execute_pitr $storage_type $BACKUP_NAME)
    
    # 7. 等待 PITR 完成
    if ! wait_for_pitr $TARGET_NAME; then
        log_error "$storage_type PITR 失败"
        return 1
    fi
    
    # 8. 验证数据
    if verify_restored_data $TARGET_NAME; then
        log_success "$storage_type 测试全部通过！✅"
        return 0
    else
        log_error "$storage_type 数据验证失败"
        return 1
    fi
}

# 主函数
main() {
    log_info "PolarDB-X 备份和 PITR 自动化测试"
    log_info "存储类型: $STORAGE_TYPE"
    
    # 检查依赖
    check_dependencies
    
    # 检查集群
    check_cluster
    
    # 创建测试数据
    create_test_data
    
    # 根据参数运行测试
    case $STORAGE_TYPE in
        oss|sftp|s3)
            run_storage_test $STORAGE_TYPE
            RESULT=$?
            ;;
        all)
            TOTAL=0
            PASSED=0
            
            for st in oss sftp s3; do
                TOTAL=$((TOTAL + 1))
                if run_storage_test $st; then
                    PASSED=$((PASSED + 1))
                fi
                
                # 清理以便下一个测试
                if [ $st != "s3" ]; then
                    log_info "等待 30 秒后继续下一个测试..."
                    sleep 30
                fi
            done
            
            echo ""
            echo "=========================================="
            log_info "测试总结"
            echo "=========================================="
            log_info "总测试数: $TOTAL"
            log_success "通过: $PASSED"
            log_error "失败: $((TOTAL - PASSED))"
            
            if [ $PASSED -eq $TOTAL ]; then
                log_success "所有测试通过！🎉"
                RESULT=0
            else
                log_error "部分测试失败"
                RESULT=1
            fi
            ;;
        *)
            log_error "不支持的存储类型: $STORAGE_TYPE"
            log_info "用法: $0 [oss|sftp|s3|all]"
            exit 1
            ;;
    esac
    
    # 询问是否清理
    echo ""
    read -p "是否清理测试资源？(y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cleanup $STORAGE_TYPE
    fi
    
    exit $RESULT
}

# 运行主函数
main
