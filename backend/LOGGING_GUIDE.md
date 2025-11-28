# 后端日志埋点设计文档

## 概述

本文档描述了 PolarDB-X UI 后端的日志埋点方案，用于解决问题排查时日志不足的问题。

## 日志架构

### 1. 日志模块位置

```
backend/pkg/api/middleware/
├── logger.go          # 请求日志中间件 + 业务日志工具
└── error_logger.go    # 错误日志增强 + 审计日志
```

### 2. 日志类型

| 类型 | 用途 | 示例 |
|------|------|------|
| 请求日志 | 记录 HTTP 请求/响应 | `[abc123] ▶ GET /api/v1/clusters` |
| 业务日志 | 记录业务逻辑执行 | `[abc123] [ClusterService] ℹ️ listing clusters` |
| K8s 操作日志 | 记录 Kubernetes API 调用 | `[abc123] 🔄 K8s List PolarDBXCluster/default/*` |
| 错误日志 | 记录错误详情和堆栈 | `[abc123] ❌ ERROR [NOT_FOUND] ...` |
| 审计日志 | 记录关键操作 | `[abc123] ✓ AUDIT [SUCCESS] CREATE PolarDBXCluster` |
| 安全日志 | 记录安全相关事件 | `[abc123] 🔒 SECURITY [AUTH_FAILURE] ...` |

### 3. 请求 ID 追踪

每个请求都会生成唯一的 8 位请求 ID（如 `abc12345`），贯穿整个请求生命周期：

```
[abc12345] ▶ POST /api/v1/clusters/default/create from 192.168.1.1
[abc12345] [ClusterService] ℹ️ creating cluster from config name=my-cluster namespace=default
[abc12345] [ClusterService] 🔍 converted config to PolarDBXCluster: CN=2 DN=2
[abc12345] 🔄 K8s Create PolarDBXCluster/default/my-cluster starting
[abc12345] ✓ K8s Create PolarDBXCluster/default/my-cluster completed in 125ms
[abc12345] ✓ AUDIT [SUCCESS] CREATE PolarDBXCluster/default/my-cluster | user=admin | ip=192.168.1.1
[abc12345] ✓ INFO POST /api/v1/clusters/default/create | 201 | 156ms | user=admin context=minikube
```

## 使用方法

### 1. 在业务服务中添加日志

```go
import "polardbx-ui-backend/pkg/api/middleware"

func (s *MyService) DoSomething(c *gin.Context) {
    // 创建业务日志器
    logger := middleware.NewBusinessLogger(c, "MyService")
    
    // 记录信息
    logger.Info("starting operation param=%s", param)
    
    // 记录警告
    logger.Warn("deprecated feature used")
    
    // 记录错误
    if err != nil {
        logger.Error(err, "operation failed for resource=%s", name)
        middleware.LogK8sError(c, "Create", "MyResource", namespace, name, err)
        return
    }
    
    // 记录调试信息
    logger.Debug("processed %d items", count)
}
```

### 2. 记录 K8s 操作

```go
// 方式1: 使用 LogK8sError 记录错误
if err := cli.Create(ctx, obj); err != nil {
    middleware.LogK8sError(c, "Create", "PolarDBXCluster", namespace, name, err)
    return
}

// 方式2: 使用 K8sOperationLogger 记录操作时间
done := middleware.K8sOperationLogger(c, "Create", "PolarDBXCluster", namespace, name)
err := cli.Create(ctx, obj)
done(err) // 自动记录成功/失败和耗时
```

### 3. 记录审计日志

```go
// 操作成功
middleware.LogAudit(c, "CREATE", "PolarDBXCluster", namespace, name, true)

// 操作失败
middleware.LogAudit(c, "DELETE", "PolarDBXBackup", namespace, name, false)
```

### 4. 记录安全事件

```go
middleware.LogSecurityEvent(c, "AUTH_FAILURE", "invalid kubeconfig provided")
middleware.LogSecurityEvent(c, "PERMISSION_DENIED", "user lacks cluster-admin role")
```

## 日志级别说明

| 级别 | Emoji | 使用场景 |
|------|-------|----------|
| DEBUG | 🔍 | 详细调试信息，生产环境可关闭 |
| INFO | ℹ️ | 正常业务流程 |
| WARN | ⚠️ | 警告但不影响功能 |
| ERROR | ❌ | 错误需要关注 |

## 敏感信息处理

请求日志中间件自动脱敏以下字段：
- `password`
- `token`
- `secret`
- `kubeconfig`
- `authorization`

示例：
```json
// 原始请求
{"password": "my-secret-pass", "name": "cluster1"}

// 日志输出
{"password": "***MASKED***", "name": "cluster1"}
```

## 日志输出示例

### 成功请求
```
[a1b2c3d4] ▶ POST /api/v1/clusters/default/create from 192.168.1.100 | User-Agent: Mozilla/5.0...
[a1b2c3d4] 📤 Request Body: {"name":"my-cluster","topology":{"cn":{"replicas":2}...}}
[a1b2c3d4] [ClusterService] ℹ️ creating cluster from config name=my-cluster namespace=default
[a1b2c3d4] [ClusterService] 🔍 converted config to PolarDBXCluster: CN=2 DN=2
[a1b2c3d4] ✓ AUDIT [SUCCESS] CREATE PolarDBXCluster/default/my-cluster | user=admin | ip=192.168.1.100
[a1b2c3d4] ✓ INFO POST /api/v1/clusters/default/create | 201 | 234ms | user=admin context=minikube
```

### 失败请求
```
[e5f6g7h8] ▶ DELETE /api/v1/clusters/default/nonexistent from 192.168.1.100
[e5f6g7h8] [ClusterService] ℹ️ deleting cluster name=nonexistent namespace=default
[e5f6g7h8] [ClusterService] ❌ failed to delete cluster name=nonexistent namespace=default: ...
[e5f6g7h8] ❌ K8s ERROR | op=Delete | resource=PolarDBXCluster/default/nonexistent | status=404 | type=NOT_FOUND
[e5f6g7h8] ✗ AUDIT [FAILURE] DELETE PolarDBXCluster/default/nonexistent | user=admin | ip=192.168.1.100
[e5f6g7h8] ⚠ WARN DELETE /api/v1/clusters/default/nonexistent | 404 | 45ms | user=admin context=minikube
[e5f6g7h8] 📥 Response Body: {"error":"cluster not found","details":"..."}
```

### Panic 恢复
```
[i9j0k1l2] 💥 PANIC RECOVERED | path=/api/v1/test | error=nil pointer dereference
    /app/pkg/api/domain/test/handler.go:45 Handler
    /app/pkg/api/middleware/logger.go:120 RequestLogger.func1
```

## 配置

在 `main.go` 中配置日志中间件：

```go
r.Use(middleware.RequestLogger(middleware.RequestLogConfig{
    LogRequestBody:  true,           // 记录请求体
    LogResponseBody: true,           // 记录响应体（错误时）
    MaxBodyLogSize:  4096,           // 最大记录大小
    SkipPaths:       []string{"/ping", "/health"}, // 跳过的路径
    SensitiveFields: []string{"password", "token"}, // 需要脱敏的字段
}))
```

## 已埋点的服务

| 服务 | 文件 | 状态 |
|------|------|------|
| ClusterService | `cluster_crud.go` | ✅ 完成 |
| BackupService | `backup_core.go` | ✅ 完成 |
| 其他服务 | - | 待补充 |

## 待办事项

- [ ] 为其他 domain handler 添加日志
- [ ] 添加日志级别控制（通过环境变量）
- [ ] 考虑集成结构化日志库（如 zap）
- [ ] 添加日志采集配置（如 fluentd/loki）

## 问题排查指南

1. **找到请求 ID**：从前端错误响应或日志中获取 `request_id`
2. **搜索日志**：`grep "abc12345" backend.log`
3. **分析请求流程**：按时间顺序查看该请求的所有日志
4. **定位错误**：查找 `❌` 或 `ERROR` 标记的日志行
5. **检查堆栈**：如果有 Stack 信息，定位到具体代码行
