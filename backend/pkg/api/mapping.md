# Backend API 架构设计文档# Backend API 架构设计文档# Backend API 架构设计文档# CRD ↔ Backend API Mapping (Alias Routes)



## 目标架构（Clean Architecture）



采用分层架构，每层职责明确，依赖单向流动：## 目标架构（Clean Architecture）



```

┌─────────────────────────────────────────────────────────────┐

│                     HTTP Layer (Gin)                        │采用分层架构，每层职责明确，依赖单向流动：## 目标架构（Clean Architecture）- Base prefix: `/api/v1/crd/*` (only adds aliases, does not change existing endpoints or request/response)

│  main.go / routes.go → 路由注册                              │

└─────────────────────────────────────────────────────────────┘

                              ↓

┌─────────────────────────────────────────────────────────────┐```

│                    Handler Layer                            │

│  domain/*/handler/ → HTTP 请求/响应处理，DTO 转换            │┌─────────────────────────────────────────────────────────────┐

│  - 解析请求参数                                              │

│  - 调用 Service                                              ││                     HTTP Layer (Gin)                        │采用分层架构，每层职责明确，依赖单向流动：| CRD Kind | Alias Package | Alias Route | Existing Primary Route (example) | Status |

│  - 格式化响应                                                │

└─────────────────────────────────────────────────────────────┘│  main.go / routes.go → 路由注册                              │

                              ↓

┌─────────────────────────────────────────────────────────────┐└─────────────────────────────────────────────────────────────┘| --- | --- | --- | --- | --- |

│                    Service Layer                            │

│  domain/*/service/ → 业务逻辑，不依赖 HTTP                   │                              ↓

│  - 接收纯业务参数 (context.Context, 业务对象)                │

│  - 返回业务对象或 error                                      │┌─────────────────────────────────────────────────────────────┐```| PolarDBXCluster | `crd/polardbxclusters` | `/api/v1/crd/polardbxclusters` | `/api/v1/clusters` (domain: `/api/v1/polardbxclusters`) | ✅ Done |

│  - 依赖 Repository 接口                                      │

└─────────────────────────────────────────────────────────────┘│                    Handler Layer                            │

                              ↓

┌─────────────────────────────────────────────────────────────┐│  domain/*/handler/ → HTTP 请求/响应处理，DTO 转换            │┌─────────────────────────────────────────────────────────────┐| SystemTask | `crd/systemtasks` | `/api/v1/crd/systemtasks` | `/api/v1/system-tasks` | ✅ Done |

│                   Repository Layer                          │

│  domain/*/repository/ → 数据访问抽象                         ││  - 解析请求参数                                              │

│  - 定义接口 (interface.go)                                   │

│  - K8s 实现 (k8s_*.go)                                       ││  - 调用 Service                                              ││                     HTTP Layer (Gin)                        │| XStore | `crd/xstores` | `/api/v1/crd/xstores` | `/api/v1/xstores` | ✅ Done |

│  - 便于 mock 测试                                            │

└─────────────────────────────────────────────────────────────┘│  - 格式化响应                                                │

```

└─────────────────────────────────────────────────────────────┘│  main.go / routes.go → 路由注册                              │| PolarDBXBackup | `crd/polardbxbackups` | `/api/v1/crd/polardbxbackups` | `/api/v1/backups/*` | ✅ Done |

---

                              ↓

## 目录结构

┌─────────────────────────────────────────────────────────────┐└─────────────────────────────────────────────────────────────┘| PolarDBXBackupSchedule | `crd/polardbxbackupschedules` | `/api/v1/crd/polardbxbackupschedules` | `/api/v1/backup-schedules/*` | ✅ Done |

```

pkg/api/│                    Service Layer                            │

├── domain/                          # 业务域

│   ├── polardbxclusters/            # 集群域 ✅ 完成│  domain/*/service/ → 业务逻辑，不依赖 HTTP                   │                              ↓| PolarDBXBackupBinlog | `crd/polardbxbackupbinlogs` | `/api/v1/crd/polardbxbackupbinlogs` | `/api/v1/backup-binlogs/*` | ✅ Done |

│   │   ├── handler/

│   │   └── (直接 K8s 访问)│  - 接收纯业务参数 (context.Context, 业务对象)                │

│   ├── xstores/                     # 存储域 ✅ 完成

│   │   └── handler/│  - 返回业务对象或 error                                      │┌─────────────────────────────────────────────────────────────┐| PolarDBXParameter | `crd/polardbxparameters` | `/api/v1/crd/polardbxparameters` | `/api/v1/parameters/*` | ✅ Done |

│   ├── systemtasks/                 # 系统任务域 ✅ 完成

│   │   ├── handler/│  - 依赖 Repository 接口                                      │

│   │   ├── service/

│   │   └── repository/└─────────────────────────────────────────────────────────────┘│                    Handler Layer                            │| PolarDBXParameterTemplate | `crd/polardbxparametertemplates` | `/api/v1/crd/polardbxparametertemplates` | `/api/v1/parameter-templates/*` | ✅ Done |

│   ├── monitoring/                  # 监控域 ✅ 完成

│   │   ├── handler/                              ↓

│   │   ├── service/

│   │   └── spec/┌─────────────────────────────────────────────────────────────┐│  domain/*/handler/ → HTTP 请求/响应处理，DTO 转换            │| PolarDBXMonitor | `crd/polardbxmonitors` | `/api/v1/crd/polardbxmonitors` | `/api/v1/monitors/*` | ✅ Done |

│   └── platform/                    # 平台横切域

│       ├── system/                  # ✅ 完成 (Handler/Service/Repository)│                   Repository Layer                          │

│       ├── pod/                     # ✅ 完成

│       ├── alerts/                  # ✅ 完成│  domain/*/repository/ → 数据访问抽象                         ││  - 解析请求参数                                              │| PolarDBXLogCollector | `crd/polardbxlogcollectors` | `/api/v1/crd/polardbxlogcollectors` | `/api/v1/log-collectors/*` | ✅ Done |

│       ├── settings/                # ✅ 完成

│       ├── diagnostics/             # ✅ 完成 (Handler only)│  - 定义接口 (interface.go)                                   │

│       ├── auth/                    # ✅ 完成 (Handler only)

│       ├── restore/                 # ✅ 完成 (Handler/Repository)│  - K8s 实现 (k8s_*.go)                                       ││  - 调用 Service                                              │| XStoreBackupBinlog | `crd/xstorebackupbinlogs` | `/api/v1/crd/xstorebackupbinlogs` | (new) | ✅ Done |

│       ├── logservice/              # ✅ 完成 (Handler only)

│       └── logcollector/            # ✅ 完成 (Handler/Repository)│  - 便于 mock 测试                                            │

├── crd/                             # CRD 别名路由 (11个子包)

├── router/                          # 路由聚合└─────────────────────────────────────────────────────────────┘│  - 格式化响应                                                │

├── middleware/                      # 中间件

└── util/                            # 通用工具```

```

└─────────────────────────────────────────────────────────────┘Note: Alias routes are assembled by `api/router`, forwarding to existing handlers; the UI can continue using legacy routes.

---

---

## 迁移进度

                              ↓

### ✅ 已完成（Clean Architecture）

## 目录结构规范

| 模块 | Handler | Service | Repository | 旧包状态 |

| --- | --- | --- | --- | --- |┌─────────────────────────────────────────────────────────────┐---

| `domain/systemtasks` | ✅ | ✅ | ✅ | 可删除 `systemtask/` |

| `domain/monitoring` | ✅ | ✅ | ✅ | 已合并 |```

| `domain/platform/system` | ✅ | ✅ | ✅ | 薄包装 |

| `domain/platform/pod` | ✅ | ✅ | ✅ | 薄包装 |pkg/api/│                    Service Layer                            │

| `domain/platform/alerts` | ✅ | ✅ | ✅ | 薄包装 |

| `domain/platform/settings` | ✅ | ✅ | ✅ | 薄包装 |├── domain/                          # 业务域

| `domain/platform/diagnostics` | ✅ | - | - | 薄包装 |

| `domain/platform/auth` | ✅ | - | - | 薄包装 |│   ├── polardbxclusters/            # 集群域│  domain/*/service/ → 业务逻辑，不依赖 HTTP                   │# Domain Entrances

| `domain/platform/restore` | ✅ | - | ✅ | 薄包装 |

| `domain/platform/logservice` | ✅ | - | - | 薄包装 |│   │   ├── handler/                 # HTTP 处理层

| `domain/platform/logcollector` | ✅ | - | ✅ | 薄包装 |

│   │   ├── service/                 # 业务逻辑层│  - 接收纯业务参数 (context.Context, 业务对象)                │

### 🔄 待迁移（仍在 pkg/api/ 旧位置）

│   │   └── repository/              # 数据访问层

| 旧包 | 目标位置 | 行数 | 优先级 |

| --- | --- | --- | --- |│   ├── xstores/                     # 存储域│  - 返回业务对象或 error                                      │- Add domain-level entrances only (thin handlers forwarding), do not replace legacy routes:

| `grafana/` | `domain/platform/grafana` | 673 | P3 |

| `logs/` | `domain/platform/logs` | 1045 | P3 |│   ├── systemtasks/                 # 系统任务域 ✅ Clean Architecture

| `logstrategy/` | `domain/platform/logstrategy` | 1006 | P3 |

| `prometheusrule/` | `domain/platform/prometheusrule` | 2032 | P3 |│   │   ├── handler/│  - 依赖 Repository 接口                                      │  - Logical cluster domain: `/api/v1/polardbxclusters/*` ✅ Done



---│   │   ├── service/



## 薄包装层删除策略│   │   └── repository/└─────────────────────────────────────────────────────────────┘  - Storage domain: `/api/v1/xstores/*` ✅ Done



### 什么是薄包装层？│   ├── monitoring/                  # 监控域 ✅ Clean Architecture



迁移后旧包仅保留变量别名，将调用委托给 domain handler：│   │   ├── handler/                              ↓  - System tasks domain: `/api/v1/systemtasks/*` ✅ Done



```go│   │   ├── service/

// pkg/api/system/endpoints.go（薄包装）

package system│   │   └── spec/┌─────────────────────────────────────────────────────────────┐  - Platform cross-cutting: `/api/v1/platform/*` ✅ Done



import "polardbx-ui-backend/pkg/api/domain/platform/system/handler"│   └── platform/                    # 平台横切域



var ContextInfo = handler.ContextInfo│       ├── system/                  # ✅ Clean Architecture│                   Repository Layer                          │

var ListNamespaces = handler.ListNamespaces

```│       │   ├── handler/



### 可以删除吗？│       │   ├── service/│  domain/*/repository/ → 数据访问抽象                         │Where assembled: `pkg/api/router` exposes `RegisterCRDAliasRoutes` and `RegisterDomainRoutes`; they are registered during app init, and `LogGroupedRoutes` prints grouped logs for discoverability.



**可以！** 薄包装层只是为了保持向后兼容，待以下条件满足后可安全删除：│       │   └── repository/



1. **更新 main.go 导入**：改为直接导入 `domain/platform/*/handler`│       ├── pod/                     # ✅ Clean Architecture│  - 定义接口 (interface)                                      │

2. **更新路由注册**：使用 domain handler 函数

3. **检查交叉引用**：确保没有其他包依赖旧包│       ├── alerts/                  # ✅ Clean Architecture



### 删除步骤│       ├── settings/                # ✅ Clean Architecture│  - K8s 实现 (k8s_*.go)                                       │---



```bash│       ├── diagnostics/             # 🔄 待迁移

# 1. 查找旧包引用

grep -r "polardbx-ui-backend/pkg/api/system" --include="*.go" | grep -v "_test.go"│       ├── auth/                    # 🔄 待迁移│  - 便于 mock 测试                                            │



# 2. 更新 main.go 导入│       └── ...

# 3. 删除旧包

rm -rf pkg/api/system/├── crd/                             # CRD 别名路由 (11个子包)└─────────────────────────────────────────────────────────────┘# Package Ownership List (Migration Progress)



# 4. 验证编译├── router/                          # 路由聚合

go build -o /dev/null .

```├── middleware/                      # 中间件```



### 推荐时机└── util/                            # 通用工具



- **所有 platform 包迁移完成后**```## ✅ Completed (Migrated to Domain)

- **批量删除**：一次性删除所有薄包装，减少多次修改 main.go



---

------

## CRD 路由映射



| CRD Kind | 路由前缀 | Domain |

| --- | --- | --- |## 迁移进度| Domain | Package | Content |

| PolarDBXCluster | `/api/v1/clusters`, `/api/v1/crd/polardbxclusters` | `polardbxclusters` |

| XStore | `/api/v1/xstores`, `/api/v1/crd/xstores` | `xstores` |

| SystemTask | `/api/v1/system-tasks`, `/api/v1/crd/systemtasks` | `systemtasks` |

| PolarDBXMonitor | `/api/v1/monitors`, `/api/v1/crd/polardbxmonitors` | `monitoring` |### ✅ 已完成（Clean Architecture 三层结构）## 目录结构规范| --- | --- | --- |

| PolarDBXBackup | `/api/v1/backups` | `polardbxclusters` |

| PolarDBXBackupSchedule | `/api/v1/backup-schedules` | `polardbxclusters` |

| PolarDBXBackupBinlog | `/api/v1/backup-binlogs` | `polardbxclusters` |

| PolarDBXParameter | `/api/v1/parameters` | `polardbxclusters` || 模块 | Handler | Service | Repository | 旧包状态 || `domain/polardbxclusters` | handlers.go | Cluster CRUD/ops, backup/schedule/binlog, parameters/templates, knobs, prechange, restore |

| PolarDBXParameterTemplate | `/api/v1/parameter-templates` | `polardbxclusters` |

| PolarDBXLogCollector | `/api/v1/log-collectors` | `platform` || --- | --- | --- | --- | --- |



---| `domain/systemtasks` | ✅ | ✅ | ✅ | 可删除 `systemtask/` |```| `domain/xstores` | handlers.go | XStore CRUD, followers, backups, rebuild |



## 代码规范| `domain/monitoring` | ✅ | ✅ | ✅ | 已合并 |



### 1. Handler 层（HTTP 适配）| `domain/platform/system` | ✅ | ✅ | ✅ | 薄包装 |pkg/api/| `domain/systemtasks` | handlers.go | SystemTask CRUD |



```go| `domain/platform/pod` | ✅ | ✅ | ✅ | 薄包装 |

// handler/system_handler.go

package handler| `domain/platform/alerts` | ✅ | ✅ | ✅ | 薄包装 |├── domain/                          # 业务域| `domain/monitoring` | handlers.go, crd_handlers.go | 监控安装工作流 + PolarDBXMonitor CRD CRUD |



type SystemHandler struct {| `domain/platform/settings` | ✅ | ✅ | ✅ | 薄包装 |

    service *service.SystemService

}│   ├── polardbxclusters/            # 集群域| `domain/platform` | handlers_*.go | 代理转发到旧包（alerts, grafana, logs, pod, settings, system, prometheusrule） |



func NewSystemHandler(svc *service.SystemService) *SystemHandler {### 🔄 待迁移（仍在 pkg/api/ 旧位置）

    return &SystemHandler{service: svc}

}│   │   ├── handler/                 # HTTP 处理层



// ListNamespaces 只处理 HTTP，不含业务逻辑| 旧包 | 目标位置 | 行数 | 优先级 |

func ListNamespaces(c *gin.Context) {

    h, ok := NewSystemHandlerFromClient(c)| --- | --- | --- | --- |│   │   │   ├── cluster_handler.go   # Handler 实现## 🔄 Phase 2: Pending Migration (Still in pkg/api/)

    if !ok {

        return| `diagnostics/` | `domain/platform/diagnostics` | 54 | P1 |

    }

    items, err := h.service.ListNamespaces(c.Request.Context())| `auth/` | `domain/platform/auth` | 132 | P1 |│   │   │   ├── dto.go               # 请求/响应 DTO

    if err != nil {

        c.JSON(http.StatusOK, gin.H{"items": []any{}, "warning": err.Error()})| `restore/` | `domain/platform/restore` | 416 | P2 |

        return

    }| `logservice/` | `domain/platform/logservice` | 241 | P2 |│   │   │   └── routes.go            # 路由注册These packages have thin forwarding wrappers in `domain/platform/handlers_*.go` but the actual logic is still in the old location:

    c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})

}| `logcollector/` | `domain/platform/logcollector` | 462 | P2 |

```

| `grafana/` | `domain/platform/grafana` | 673 | P3 |│   │   ├── service/                 # 业务逻辑层

### 2. Service 层（业务逻辑）

| `logs/` | `domain/platform/logs` | 1045 | P3 |

```go

// service/system_service.go| `logstrategy/` | `domain/platform/logstrategy` | 1006 | P3 |│   │   │   ├── cluster_service.go   # Service 实现| Old Package | Target Domain | Notes |

package service

| `prometheusrule/` | `domain/platform/prometheusrule` | 2032 | P3 |

type SystemService struct {

    repo repository.SystemRepository│   │   │   ├── interface.go         # Repository 接口定义| --- | --- | --- |

}

---

func NewSystemService(repo repository.SystemRepository) *SystemService {

    return &SystemService{repo: repo}│   │   │   └── *_test.go            # 单元测试（mock repo）| `alerts/` | `domain/platform` | AlertManager integration |

}

## 薄包装层删除策略

// ListNamespaces 纯业务逻辑，不知道 HTTP

func (s *SystemService) ListNamespaces(ctx context.Context) ([]NamespaceInfo, error) {│   │   ├── repository/              # 数据访问层| `auth/` | `domain/platform` | JWT auth (optional) |

    namespaces, err := s.repo.ListNamespaces(ctx)

    if err != nil {### 什么是薄包装层？

        return nil, err

    }│   │   │   ├── interface.go         # 接口定义（可与 service 共用）| `clusterknobs/` | `domain/polardbxclusters` | Already forwarded via domain handlers |

    // 转换为业务对象

    items := make([]NamespaceInfo, 0, len(namespaces))迁移后旧包仅保留变量别名，将调用委托给 domain handler：

    for _, ns := range namespaces {

        items = append(items, NamespaceInfo{│   │   │   ├── k8s_cluster.go       # K8s 实现| `diagnostics/` | `domain/platform` | Cluster diagnostics |

            Name:      ns.Name,

            Status:    string(ns.Status.Phase),```go

            CreatedAt: ns.CreationTimestamp.Time,

        })// pkg/api/system/endpoints.go（薄包装）│   │   │   └── mock_cluster.go      # Mock 实现（测试用）| `grafana/` | `domain/platform` | Dashboard sync/config |

    }

    return items, nilpackage system

}

```│   │   └── entity/                  # 领域实体（可选）| `logcollector/` | `domain/platform` | LogCollector CRD management |



### 3. Repository 层（数据访问）import "polardbx-ui-backend/pkg/api/domain/platform/system/handler"



```go│   │       └── cluster.go| `logs/` | `domain/platform` | Log query/bootstrap |

// repository/interface.go

package repositoryvar ContextInfo = handler.ContextInfo



type SystemRepository interface {var ListNamespaces = handler.ListNamespaces│   │| `logservice/` | `domain/platform` | ES/Loki service status |

    ListNamespaces(ctx context.Context) ([]corev1.Namespace, error)

}```



// repository/k8s_system.go│   ├── xstores/                     # 存储域（同上结构）| `logstrategy/` | `domain/platform` | Log strategy CRUD |

type K8sSystemRepository struct {

    client client.Client### 可以删除吗？

}

│   ├── systemtasks/                 # 系统任务域| `parameters/` | `domain/polardbxclusters` | Already forwarded |

func NewK8sSystemRepository(cli client.Client) *K8sSystemRepository {

    return &K8sSystemRepository{client: cli}**可以！** 薄包装层只是为了保持向后兼容，待以下条件满足后可安全删除：

}

│   ├── monitoring/                  # 监控域| `pod/` | `domain/platform` | Pod CRUD/exec/logs |

func (r *K8sSystemRepository) ListNamespaces(ctx context.Context) ([]corev1.Namespace, error) {

    var nsList corev1.NamespaceList1. **更新 main.go 导入**：改为直接导入 `domain/platform/*/handler`

    if err := r.client.List(ctx, &nsList, &client.ListOptions{}); err != nil {

        return nil, err2. **更新路由注册**：使用 domain handler 函数│   └── platform/                    # 平台横切域| `prechange/` | `domain/polardbxclusters` | Already forwarded |

    }

    return nsList.Items, nil3. **检查交叉引用**：确保没有其他包依赖旧包

}

```│       ├── handler/| `prometheusrule/` | `domain/platform` | PrometheusRule templates |



---### 删除步骤



## 测试策略│       ├── service/| `restore/` | `domain/polardbxclusters` | Restore/PITR |



```go```bash

// service/system_service_test.go

func TestSystemService_ListNamespaces(t *testing.T) {# 1. 查找旧包引用│       └── repository/| `settings/` | `domain/platform` | Dashboard thresholds, image registry |

    // 使用 mock repository

    mockRepo := &MockSystemRepository{}grep -r "polardbx-ui-backend/pkg/api/system" --include="*.go" | grep -v "_test.go"

    mockRepo.On("ListNamespaces", mock.Anything).Return([]corev1.Namespace{

        {ObjectMeta: metav1.ObjectMeta{Name: "default"}},│| `system/` | `domain/platform` | Namespaces, context info |

    }, nil)

    # 2. 更新 main.go 导入

    svc := NewSystemService(mockRepo)

    result, err := svc.ListNamespaces(context.Background())# 3. 删除旧包├── middleware/                      # 中间件| `systemtask/` | `domain/systemtasks` | Duplicate of domain/systemtasks |

    

    assert.NoError(t, err)rm -rf pkg/api/system/

    assert.Len(t, result, 1)

    assert.Equal(t, "default", result[0].Name)│   ├── auth.go                      # 认证

    mockRepo.AssertExpectations(t)

}# 4. 验证编译

```

go build -o /dev/null .│   ├── kubeconfig.go               # K8s 客户端注入---

---

```

## 下一步计划

│   └── cors.go                      # CORS

### Phase 2C: 完成大包迁移

1. [ ] `grafana/` → `domain/platform/grafana/`### 推荐时机

2. [ ] `logs/` → `domain/platform/logs/`

3. [ ] `logstrategy/` → `domain/platform/logstrategy/`│# Recommended Next Steps

4. [ ] `prometheusrule/` → `domain/platform/prometheusrule/`

- **Phase 2 完成后**：所有 platform 包迁移完成

### Phase 3: 清理薄包装

1. [ ] 更新 main.go 直接使用 domain handlers- **批量删除**：一次性删除所有薄包装，减少多次修改 main.go├── router/                          # 路由聚合

2. [ ] 删除所有薄包装旧包

3. [ ] 更新文档


---│   └── router.go                    # 注册所有路由## Priority 1: Remove Duplicates



## CRD 路由映射│- [ ] Delete `pkg/api/systemtask/` (已迁移到 `domain/systemtasks`)



| CRD Kind | 路由前缀 | Domain |└── util/                            # 通用工具- [ ] Verify `clusterknobs/`, `parameters/`, `prechange/` are only used via domain forwarding

| --- | --- | --- |

| PolarDBXCluster | `/api/v1/clusters`, `/api/v1/crd/polardbxclusters` | `polardbxclusters` |    ├── response.go                  # 响应格式化

| XStore | `/api/v1/xstores`, `/api/v1/crd/xstores` | `xstores` |

| SystemTask | `/api/v1/system-tasks`, `/api/v1/crd/systemtasks` | `systemtasks` |    ├── errors.go                    # 错误处理## Priority 2: Consolidate Platform Handlers

| PolarDBXMonitor | `/api/v1/monitors`, `/api/v1/crd/polardbxmonitors` | `monitoring` |

| PolarDBXBackup | `/api/v1/backups` | `polardbxclusters` |    └── context.go                   # Context 工具Move actual logic from old packages to `domain/platform/`:

| PolarDBXBackupSchedule | `/api/v1/backup-schedules` | `polardbxclusters` |

| PolarDBXBackupBinlog | `/api/v1/backup-binlogs` | `polardbxclusters` |```- [ ] `system/` → `domain/platform/handlers_system.go` (move logic, not just forward)

| PolarDBXParameter | `/api/v1/parameters` | `polardbxclusters` |

| PolarDBXParameterTemplate | `/api/v1/parameter-templates` | `polardbxclusters` |- [ ] `pod/` → `domain/platform/handlers_pod.go`

| PolarDBXLogCollector | `/api/v1/log-collectors` | `platform` |

---- [ ] `alerts/` → `domain/platform/handlers_alerts.go`

---

- [ ] `grafana/` → `domain/platform/handlers_grafana.go`

## 代码规范

## 代码规范- [ ] `logs/`, `logservice/`, `logstrategy/` → `domain/platform/handlers_logs.go`

### 1. Handler 层（HTTP 适配）

- [ ] `settings/` → `domain/platform/handlers_settings.go`

```go

// handler/system_handler.go### 1. Handler 层（HTTP 适配）

package handler

## Priority 3: Final Cleanup

type SystemHandler struct {

    service *service.SystemService```go- [ ] Create `domain/platform/services/` for business logic

}

// handler/cluster_handler.go- [ ] Create `domain/platform/k8srepo/` for K8s access

func NewSystemHandler(svc *service.SystemService) *SystemHandler {

    return &SystemHandler{service: svc}package handler- [ ] Remove empty/forwarding-only old packages

}

- [ ] Update imports in main.go to use domain directly

// ListNamespaces 只处理 HTTP，不含业务逻辑

func ListNamespaces(c *gin.Context) {type ClusterHandler struct {

    h, ok := NewSystemHandlerFromClient(c)

    if !ok {    service *service.ClusterService---

        return

    }    logger  *zap.Logger

    items, err := h.service.ListNamespaces(c.Request.Context())

    if err != nil {}# Current Architecture Summary

        c.JSON(http.StatusOK, gin.H{"items": []any{}, "warning": err.Error()})

        return

    }

    c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})func NewClusterHandler(svc *service.ClusterService) *ClusterHandler {```

}

```    return &ClusterHandler{service: svc}pkg/api/



### 2. Service 层（业务逻辑）}├── monitoring/endpoints.go     # API 入口，只做路由定义



```go├── crd/                        # CRD 别名路由（11个子包）

// service/system_service.go

package service// List 只处理 HTTP，不含业务逻辑├── router/router.go            # 注册 CRD 和 Domain 路由



type SystemService struct {func (h *ClusterHandler) List(c *gin.Context) {├── domain/

    repo repository.SystemRepository

}    // 1. 解析请求│   ├── monitoring/             # ✅ 完全迁移（handlers + service）



func NewSystemService(repo repository.SystemRepository) *SystemService {    namespace := c.DefaultQuery("namespace", "")│   ├── platform/               # 🔄 代理层（handlers_*.go 转发到旧包）

    return &SystemService{repo: repo}

}    │   ├── polardbxclusters/       # ✅ 完全迁移



// ListNamespaces 纯业务逻辑，不知道 HTTP    // 2. 调用 Service（传递 context.Context，不是 gin.Context）│   ├── systemtasks/            # ✅ 完全迁移

func (s *SystemService) ListNamespaces(ctx context.Context) ([]NamespaceInfo, error) {

    namespaces, err := s.repo.ListNamespaces(ctx)    clusters, err := h.service.List(c.Request.Context(), namespace)│   └── xstores/                # ✅ 完全迁移

    if err != nil {

        return nil, err    if err != nil {└── [旧包]/                      # 待迁移到 domain

    }

    // 转换为业务对象        h.handleError(c, err)```

    items := make([]NamespaceInfo, 0, len(namespaces))

    for _, ns := range namespaces {        return

        items = append(items, NamespaceInfo{    }

            Name:      ns.Name,    

            Status:    string(ns.Status.Phase),    // 3. 格式化响应

            CreatedAt: ns.CreationTimestamp.Time,    c.JSON(http.StatusOK, clusters)

        })}

    }```

    return items, nil

}### 2. Service 层（业务逻辑）

```

```go

### 3. Repository 层（数据访问）// service/cluster_service.go

package service

```go

// repository/interface.gotype ClusterService struct {

package repository    repo   ClusterRepository  // 依赖接口

    logger *zap.Logger

type SystemRepository interface {}

    ListNamespaces(ctx context.Context) ([]corev1.Namespace, error)

}func NewClusterService(repo ClusterRepository) *ClusterService {

    return &ClusterService{repo: repo}

// repository/k8s_system.go}

type K8sSystemRepository struct {

    client client.Client// List 纯业务逻辑，不知道 HTTP

}func (s *ClusterService) List(ctx context.Context, namespace string) ([]*dto.ClusterResponse, error) {

    clusters, err := s.repo.List(ctx, namespace)

func NewK8sSystemRepository(cli client.Client) *K8sSystemRepository {    if err != nil {

    return &K8sSystemRepository{client: cli}        return nil, fmt.Errorf("list clusters: %w", err)

}    }

    return toClusterResponses(clusters), nil

func (r *K8sSystemRepository) ListNamespaces(ctx context.Context) ([]corev1.Namespace, error) {}

    var nsList corev1.NamespaceList```

    if err := r.client.List(ctx, &nsList, &client.ListOptions{}); err != nil {

        return nil, err### 3. Repository 层（数据访问）

    }

    return nsList.Items, nil```go

}// service/interface.go (或 repository/interface.go)

```package service



---type ClusterRepository interface {

    List(ctx context.Context, namespace string) ([]*polardbxv1.PolarDBXCluster, error)

## 测试策略    Get(ctx context.Context, namespace, name string) (*polardbxv1.PolarDBXCluster, error)

    Create(ctx context.Context, cluster *polardbxv1.PolarDBXCluster) (*polardbxv1.PolarDBXCluster, error)

```go    Update(ctx context.Context, cluster *polardbxv1.PolarDBXCluster) (*polardbxv1.PolarDBXCluster, error)

// service/system_service_test.go    Delete(ctx context.Context, namespace, name string) error

func TestSystemService_ListNamespaces(t *testing.T) {}

    // 使用 mock repository

    mockRepo := &MockSystemRepository{}// repository/k8s_cluster.go

    mockRepo.On("ListNamespaces", mock.Anything).Return([]corev1.Namespace{package repository

        {ObjectMeta: metav1.ObjectMeta{Name: "default"}},

    }, nil)type K8sClusterRepository struct {

        // 注入 K8s client 工厂或直接使用 context 获取

    svc := NewSystemService(mockRepo)}

    result, err := svc.ListNamespaces(context.Background())

    func (r *K8sClusterRepository) List(ctx context.Context, namespace string) ([]*polardbxv1.PolarDBXCluster, error) {

    assert.NoError(t, err)    client := util.K8sClientFromCtx(ctx)

    assert.Len(t, result, 1)    // ... K8s 操作

    assert.Equal(t, "default", result[0].Name)}

    mockRepo.AssertExpectations(t)```

}

```---



---## CRD 路由映射



## 下一步计划| CRD Kind | 路由前缀 | Domain |

| --- | --- | --- |

### Phase 2A: 完成小包迁移| PolarDBXCluster | `/api/v1/clusters`, `/api/v1/crd/polardbxclusters` | `polardbxclusters` |

1. [x] `system/` → `domain/platform/system/` ✅| XStore | `/api/v1/xstores`, `/api/v1/crd/xstores` | `xstores` |

2. [x] `pod/` → `domain/platform/pod/` ✅| SystemTask | `/api/v1/system-tasks`, `/api/v1/crd/systemtasks` | `systemtasks` |

3. [x] `alerts/` → `domain/platform/alerts/` ✅| PolarDBXMonitor | `/api/v1/monitors`, `/api/v1/crd/polardbxmonitors` | `monitoring` |

4. [x] `settings/` → `domain/platform/settings/` ✅| PolarDBXBackup | `/api/v1/backups` | `polardbxclusters` |

5. [ ] `diagnostics/` → `domain/platform/diagnostics/`| PolarDBXBackupSchedule | `/api/v1/backup-schedules` | `polardbxclusters` |

6. [ ] `auth/` → `domain/platform/auth/`| PolarDBXBackupBinlog | `/api/v1/backup-binlogs` | `polardbxclusters` |

7. [ ] `restore/` → `domain/platform/restore/`| PolarDBXParameter | `/api/v1/parameters` | `polardbxclusters` |

8. [ ] `logservice/` → `domain/platform/logservice/`| PolarDBXParameterTemplate | `/api/v1/parameter-templates` | `polardbxclusters` |

9. [ ] `logcollector/` → `domain/platform/logcollector/`| PolarDBXLogCollector | `/api/v1/log-collectors` | `platform` |



### Phase 2B: 大包迁移---

10. [ ] `grafana/` → `domain/platform/grafana/`

11. [ ] `logs/` → `domain/platform/logs/`## 迁移进度

12. [ ] `logstrategy/` → `domain/platform/logstrategy/`

13. [ ] `prometheusrule/` → `domain/platform/prometheusrule/`### ✅ Phase 1: 路由别名（已完成）

- [x] CRD 别名路由 `/api/v1/crd/*`

### Phase 3: 清理薄包装- [x] Domain 入口路由 `/api/v1/polardbxclusters/*` 等

14. [ ] 更新 main.go 直接使用 domain handlers

15. [ ] 删除所有薄包装旧包### 🔄 Phase 2: 架构重构（进行中）

16. [ ] 更新文档

| Domain | Handler | Service | Repository | 状态 |
| --- | --- | --- | --- | --- |
| `monitoring` | ✅ | ✅ | ✅ | 完成 |
| `systemtasks` | 🔄 | 待重构 | 待重构 | 需要解耦 |
| `polardbxclusters` | 🔄 | 部分 | 部分 | 需要解耦 |
| `xstores` | 🔄 | 部分 | 部分 | 需要解耦 |
| `platform` | 🔄 | 待迁移 | 待迁移 | 代理层 |

### 📋 Phase 2 任务清单

1. **删除重复包**
   - [ ] 删除 `pkg/api/systemtask/`（已迁移）

2. **重构 systemtasks 为示范**
   - [ ] 创建 `service/interface.go` 定义 Repository 接口
   - [ ] 重构 Service 方法签名，不接收 `*gin.Context`
   - [ ] Handler 只做 HTTP 处理

3. **迁移 platform 旧包**
   - [ ] `system/` → `platform/service/system.go`
   - [ ] `pod/` → `platform/service/pod.go`
   - [ ] `alerts/` → `platform/service/alerts.go`
   - [ ] `grafana/` → `platform/service/grafana.go`
   - [ ] `logs/`, `logservice/`, `logstrategy/` → `platform/service/logs.go`
   - [ ] `settings/` → `platform/service/settings.go`
   - [ ] `diagnostics/` → `platform/service/diagnostics.go`
   - [ ] `prometheusrule/` → `platform/service/prometheusrule.go`

4. **清理旧包**
   - [ ] 删除已迁移的旧包
   - [ ] 更新 main.go 导入

---

## 测试策略

```go
// service/cluster_service_test.go
func TestClusterService_List(t *testing.T) {
    // 使用 mock repository
    mockRepo := &MockClusterRepository{}
    mockRepo.On("List", mock.Anything, "default").Return([]*polardbxv1.PolarDBXCluster{...}, nil)
    
    svc := NewClusterService(mockRepo)
    result, err := svc.List(context.Background(), "default")
    
    assert.NoError(t, err)
    assert.Len(t, result, 1)
    mockRepo.AssertExpectations(t)
}
```

---

## 依赖注入

```go
// main.go 或 wire.go
func initializeDependencies(k8sClient client.Client) {
    // Repository
    clusterRepo := repository.NewK8sClusterRepository()
    
    // Service
    clusterService := service.NewClusterService(clusterRepo)
    
    // Handler
    clusterHandler := handler.NewClusterHandler(clusterService)
    
    // 路由注册
    clusterHandler.RegisterRoutes(v1)
}
```
