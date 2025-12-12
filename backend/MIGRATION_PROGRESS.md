# 代码改进进度报告

## 已完成的工作

### 1. 日志统一 ✅

#### Config 包
- ✅ `pkg/config/app.go` - 统一为 logger，带 fallback
- ✅ `pkg/config/server.go` - 统一为 logger，带 fallback
- ✅ `pkg/config/kubernetes.go` - 统一为 logger，带 fallback
- ✅ `pkg/config/config.go` - 统一为 logger，带 fallback

**策略：** 由于 config 包在初始化阶段执行，logger 可能还未初始化，因此使用条件检查：如果 logger 可用则使用 logger，否则 fallback 到标准 log。

#### API 层
- ✅ `pkg/api/util/http.go` - HandleK8sError 使用结构化日志
- ✅ `pkg/api/router/router.go` - 统一为 logger
- ✅ `pkg/api/router/setup.go` - 统一为 logger
- ✅ `pkg/api/domain/polardbxclusters/services/cluster_ops.go` - 统一为 logger

### 2. Context 管理改进 ✅ (100% 完成)

#### K8s 客户端函数
- ✅ `pkg/k8s/cluster.go` - 6 个函数
- ✅ `pkg/k8s/backup.go` - 14 个函数
- ✅ `pkg/k8s/xstore.go` - 20 个函数
- ✅ `pkg/k8s/misc.go` - 31 个函数
- ✅ 所有 deprecated 函数改为使用 `context.Background()` 替代 `context.TODO()`
- ✅ 添加警告日志，提醒开发者迁移到 WithContext 版本
- ✅ 更新函数注释，说明使用 context.Background() 的限制

**改进说明：**
- `context.TODO()` 表示"不知道应该使用什么 context"，不适合用于生产代码
- `context.Background()` 表示"没有父 context 的根 context"，更适合作为默认值
- 添加警告日志帮助识别仍在使用 deprecated 函数的代码
- **总计改进：71 个 deprecated 函数**

## 待完成的工作

### 1. 日志统一（90% 完成）

**已完成：**
- ✅ `pkg/api/handlers.go` (18 处) - 已统一
- ✅ `pkg/api/domain/platform/pod/handler/pod_handler.go` (7 处) - 已统一
- ✅ `pkg/api/domain/xstores/services/rebuild.go` (5 处) - 已统一
- ✅ `pkg/api/domain/polardbxclusters/services/runner/runner.go` (3 处) - 已统一
- ✅ `pkg/api/errors/handler.go` (logError 函数) - 已统一

**剩余（可选）：**
- `pkg/api/middleware/error_logger.go` (6 处) - 专门的错误日志中间件，使用特殊格式（带 emoji），可保留或后续统一
- `pkg/api/errors/handler.go` (1 处) - panic recovery，保留使用 log（因为 logger 可能还没初始化）

**优先级：** 低（剩余的都是特殊情况）
**建议：** 已完成核心文件的统一，剩余文件可根据需要决定是否统一

### 2. 代码审查建议

1. **逐步移除 deprecated 函数**
   - 当前这些函数仍在使用，建议：
     - 先确保所有调用方都迁移到 WithContext 版本
     - 然后移除 deprecated 函数

2. **添加 lint 规则**
   - 禁止使用 `context.TODO()`（除了测试代码）
   - 禁止使用标准 `log` 包（除了初始化阶段）

3. **监控警告日志**
   - 在生产环境中监控 deprecated 函数的使用
   - 定期检查并迁移

## 统计

### 日志统一进度
- **已完成：** 
  - ✅ Config 包（4 个文件）
  - ✅ API 工具层和 Router 层
  - ✅ Services 层（cluster_ops.go, rebuild.go, runner.go）
  - ✅ Handlers 层（handlers.go, pod_handler.go）
  - ✅ Errors 层（handler.go 的 logError 函数）
- **剩余：** ~7 处（middleware/error_logger.go - 专门的错误日志中间件，可保留或后续统一）
- **进度：** ~90% ✅

### Context 管理进度
- **已完成：** 
  - ✅ `cluster.go` (6 个函数)
  - ✅ `backup.go` (14 个函数)
  - ✅ `xstore.go` (20 个函数)
  - ✅ `misc.go` (31 个函数)
- **总计：** 71 个函数已改进
- **进度：** 100% ✅

## 下一步行动

1. **立即执行：**
   - [x] 改进 `pkg/k8s/backup.go` 的 context 使用 ✅
   - [x] 改进 `pkg/k8s/xstore.go` 的 context 使用 ✅
   - [x] 改进 `pkg/k8s/misc.go` 的 context 使用 ✅

2. **近期执行：**
   - [ ] 统一 `pkg/api/handlers.go` 的日志
   - [ ] 统一 services 层的日志

3. **长期优化：**
   - [ ] 添加 lint 规则
   - [ ] 移除所有 deprecated 函数
   - [ ] 完善文档

