# Context 管理最佳实践

## 概述

本文档描述了在 polardbx-ui-backend 中正确使用 `context.Context` 的最佳实践。

## 核心原则

### 1. 永远不要使用 `context.TODO()` 或 `context.Background()` 在请求处理中

**错误示例：**
```go
// ❌ 错误：丢失了请求的取消信号和超时
func ListClusters(c *gin.Context) {
    clusters, err := k8s.ListPolarDBXClusters(context.TODO(), client, namespace)
    // ...
}
```

**正确示例：**
```go
// ✅ 正确：使用请求的 context
func ListClusters(c *gin.Context) {
    ctx, cancel := util.ListCtx(c) // 从请求 context 派生，带超时
    defer cancel()
    
    clusters, err := k8s.ListPolarDBXClustersWithContext(ctx, client, namespace)
    // ...
}
```

### 2. 从 HTTP 请求 Context 派生新的 Context

所有 HTTP handler 应该从 `c.Request.Context()` 派生新的 context：

```go
func Handler(c *gin.Context) {
    // 使用工具函数创建带超时的 context
    ctx, cancel := util.CrudCtx(c) // 默认 60 秒超时
    defer cancel()
    
    // 或者自定义超时
    ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
    defer cancel()
    
    // 执行操作
    result, err := service.DoSomething(ctx, ...)
}
```

### 3. 使用工具函数管理超时

项目提供了便捷的工具函数：

```go
// 列表操作：15 秒超时
ctx, cancel := util.ListCtx(c)
defer cancel()

// CRUD 操作：60 秒超时
ctx, cancel := util.CrudCtx(c)
defer cancel()

// 自定义超时
ctx, cancel := util.BindValidateAndCtx(c, &request, 30*time.Second, validators...)
defer cancel()
```

### 4. 总是 defer cancel()

创建带 cancel 的 context 后，必须立即 defer cancel()：

```go
ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
defer cancel() // 必须调用，释放资源
```

### 5. 在 K8s 客户端调用中使用 Context

所有 K8s 客户端操作都应该使用带超时的 context：

```go
// ✅ 正确
func CreateCluster(ctx context.Context, client client.Client, namespace string, cluster *v1.PolarDBXCluster) error {
    return client.Create(ctx, cluster)
}

// ❌ 错误
func CreateCluster(client client.Client, namespace string, cluster *v1.PolarDBXCluster) error {
    return client.Create(context.TODO(), cluster) // 丢失超时和取消
}
```

## 迁移指南

### 步骤 1: 识别需要迁移的函数

查找所有使用 `context.TODO()` 或 `context.Background()` 的地方：

```bash
grep -r "context.TODO()" backend/pkg/k8s/
grep -r "context.Background()" backend/pkg/api/
```

### 步骤 2: 添加 WithContext 版本

为每个函数创建带 context 参数的版本：

```go
// 旧版本（保持向后兼容）
func ListPolarDBXClusters(c client.Client, namespace string) ([]v1.PolarDBXCluster, error) {
    return ListPolarDBXClustersWithContext(context.TODO(), c, namespace)
}

// 新版本（带 context）
func ListPolarDBXClustersWithContext(ctx context.Context, c client.Client, namespace string) ([]v1.PolarDBXCluster, error) {
    list := &v1.PolarDBXClusterList{}
    if err := c.List(ctx, list, client.InNamespace(namespace)); err != nil {
        return nil, err
    }
    return list.Items, nil
}
```

### 步骤 3: 更新 Handler 调用

在 handler 中使用新的 WithContext 版本：

```go
func ListClustersHandler(c *gin.Context) {
    ctx, cancel := util.ListCtx(c)
    defer cancel()
    
    cli, _, _, ok := util.GetK8sClients(c)
    if !ok {
        return
    }
    
    namespace := util.GetNamespace(c, "default")
    clusters, err := k8s.ListPolarDBXClustersWithContext(ctx, cli, namespace)
    if err != nil {
        apierr.AbortK8sError(c, "list clusters", err)
        return
    }
    
    apierr.OK(c, clusters)
}
```

## Context 超时配置

### 推荐超时时间

- **列表操作**: 15 秒 (`util.DefaultListTimeout`)
- **CRUD 操作**: 60 秒 (`util.DefaultCRUDTimeout`)
- **长时间操作**: 根据实际情况设置（如备份、恢复等）

### 配置示例

```go
const (
    DefaultListTimeout = 15 * time.Second
    DefaultCRUDTimeout = 60 * time.Second
    LongOperationTimeout = 5 * time.Minute // 用于备份、恢复等
)
```

## 错误处理

当 context 超时或取消时，应该正确处理：

```go
ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
defer cancel()

result, err := operation(ctx)
if err != nil {
    // 检查是否是 context 错误
    if ctxErr := apierr.FromContextError(err); ctxErr != nil {
        apierr.Abort(c, ctxErr)
        return
    }
    // 处理其他错误
    apierr.AbortWithError(c, err)
    return
}
```

## 测试中的 Context 使用

在测试中可以使用 `context.Background()`，但建议使用带超时的 context：

```go
func TestOperation(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    result, err := operation(ctx)
    // ...
}
```

## 常见错误

### ❌ 错误 1: 忘记传递 context

```go
func Handler(c *gin.Context) {
    // 丢失了请求的 context
    result, err := service.DoSomething(context.TODO(), ...)
}
```

### ❌ 错误 2: 忘记 cancel

```go
func Handler(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
    // 忘记 defer cancel() - 会导致资源泄漏
    result, err := operation(ctx)
}
```

### ❌ 错误 3: 使用错误的 context

```go
func Handler(c *gin.Context) {
    // 使用 background context 而不是请求 context
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    // 这样会丢失请求的取消信号
}
```

## 总结

1. ✅ 总是从 `c.Request.Context()` 派生新的 context
2. ✅ 使用工具函数 `util.ListCtx()` 和 `util.CrudCtx()` 创建带超时的 context
3. ✅ 总是 `defer cancel()` 释放资源
4. ✅ 在 K8s 客户端调用中使用 context
5. ✅ 正确处理 context 超时和取消错误
6. ❌ 不要在请求处理中使用 `context.TODO()` 或 `context.Background()`

