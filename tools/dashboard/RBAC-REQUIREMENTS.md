# RBAC 权限要求

## 概述

Dashboard 后端需要足够的 Kubernetes RBAC 权限才能执行 CRUD 操作，包括：
- 创建/删除/更新集群
- 节点扩缩容
- 集群变配（升级、参数修改等）
- 备份和恢复操作

## 权限要求

### 必需的权限

Dashboard 需要以下 Kubernetes 资源的完整 CRUD 权限：

#### 1. PolarDB-X CRD 资源
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: polardbx-dashboard-backend
rules:
  # PolarDB-X Cluster CRD
  - apiGroups: ["polardbx.aliyun.com"]
    resources: ["polardbxclusters"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  
  # XStore CRD
  - apiGroups: ["polardbx.aliyun.com"]
    resources: ["xstores"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  
  # Backup CRDs
  - apiGroups: ["polardbx.aliyun.com"]
    resources: ["polardbxbackups", "polardbxbackupschedules", "polardbxbackupbinlogs"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  
  # XStore Backup CRDs
  - apiGroups: ["polardbx.aliyun.com"]
    resources: ["xstorebackups", "xstorebackupbinlogs"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  
  # System Tasks
  - apiGroups: ["polardbx.aliyun.com"]
    resources: ["systemtasks"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  
  # Parameters
  - apiGroups: ["polardbx.aliyun.com"]
    resources: ["polardbxparameters", "polardbxparametertemplates"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  
  # Monitoring
  - apiGroups: ["polardbx.aliyun.com"]
    resources: ["polardbxmonitors"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  
  # Log Collector
  - apiGroups: ["polardbx.aliyun.com"]
    resources: ["polardbxlogcollectors"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

#### 2. 核心 Kubernetes 资源
```yaml
rules:
  # Pods (用于节点管理、日志查看)
  - apiGroups: [""]
    resources: ["pods", "pods/exec", "pods/log"]
    verbs: ["get", "list", "watch", "create", "delete"]
  
  # Services
  - apiGroups: [""]
    resources: ["services"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  
  # ConfigMaps & Secrets
  - apiGroups: [""]
    resources: ["configmaps", "secrets"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  
  # PersistentVolumeClaims
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "create", "delete"]
  
  # Namespaces
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list", "watch"]
  
  # Nodes (用于节点信息)
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
  
  # Storage Classes
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
```

#### 3. 扩展资源（可选，用于监控）
```yaml
rules:
  # PrometheusRule (如果使用 Prometheus Operator)
  - apiGroups: ["monitoring.coreos.com"]
    resources: ["prometheusrules"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  
  # ServiceMonitor
  - apiGroups: ["monitoring.coreos.com"]
    resources: ["servicemonitors"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

## 部署配置

### 方式 1: 使用 ServiceAccount + ClusterRoleBinding（推荐）

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: polardbx-dashboard-backend
  namespace: polardbx-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: polardbx-dashboard-backend
rules:
  # ... (如上所述)
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: polardbx-dashboard-backend
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: polardbx-dashboard-backend
subjects:
- kind: ServiceAccount
  name: polardbx-dashboard-backend
  namespace: polardbx-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: polardbx-dashboard-backend
spec:
  template:
    spec:
      serviceAccountName: polardbx-dashboard-backend
      containers:
      - name: backend
        env:
        - name: KUBECONFIG
          value: ""  # 空值表示使用 in-cluster config
```

### 方式 2: 使用 kubeconfig（需要管理员权限）

如果使用 kubeconfig，需要确保 kubeconfig 中的用户/ServiceAccount 有足够的权限。

## 权限验证

### 检查当前权限

```bash
# 检查 ServiceAccount 的权限
kubectl auth can-i --list --as=system:serviceaccount:polardbx-system:polardbx-dashboard-backend

# 测试特定操作
kubectl auth can-i create polardbxclusters --as=system:serviceaccount:polardbx-system:polardbx-dashboard-backend
kubectl auth can-i update polardbxclusters --as=system:serviceaccount:polardbx-system:polardbx-dashboard-backend
kubectl auth can-i delete polardbxclusters --as=system:serviceaccount:polardbx-system:polardbx-dashboard-backend
```

## 重要说明

⚠️ **权限不足会导致操作失败**

即使镜像能正常通信，如果没有足够的 RBAC 权限，以下操作会失败：
- ❌ 创建集群 → `403 Forbidden`
- ❌ 删除集群 → `403 Forbidden`
- ❌ 扩缩容 → `403 Forbidden`
- ❌ 集群变配 → `403 Forbidden`
- ❌ 创建备份 → `403 Forbidden`

✅ **只读操作可能成功**
- ✅ 查看集群列表
- ✅ 查看集群详情
- ✅ 查看节点信息

## 最佳实践

1. **使用最小权限原则**：只授予必要的权限
2. **使用 ServiceAccount**：不要使用集群管理员权限
3. **定期审查权限**：确保权限与实际需求匹配
4. **监控权限错误**：在日志中查找 403 错误

