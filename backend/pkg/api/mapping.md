# Backend API 架构设计文档# CRD ↔ Backend API Mapping (Alias Routes)



## 目标架构（Clean Architecture）- Base prefix: `/api/v1/crd/*` (only adds aliases, does not change existing endpoints or request/response)



采用分层架构，每层职责明确，依赖单向流动：| CRD Kind | Alias Package | Alias Route | Existing Primary Route (example) | Status |

| --- | --- | --- | --- | --- |

```| PolarDBXCluster | `crd/polardbxclusters` | `/api/v1/crd/polardbxclusters` | `/api/v1/clusters` (domain: `/api/v1/polardbxclusters`) | ✅ Done |

┌─────────────────────────────────────────────────────────────┐| SystemTask | `crd/systemtasks` | `/api/v1/crd/systemtasks` | `/api/v1/system-tasks` | ✅ Done |

│                     HTTP Layer (Gin)                        │| XStore | `crd/xstores` | `/api/v1/crd/xstores` | `/api/v1/xstores` | ✅ Done |

│  main.go / routes.go → 路由注册                              │| PolarDBXBackup | `crd/polardbxbackups` | `/api/v1/crd/polardbxbackups` | `/api/v1/backups/*` | ✅ Done |

└─────────────────────────────────────────────────────────────┘| PolarDBXBackupSchedule | `crd/polardbxbackupschedules` | `/api/v1/crd/polardbxbackupschedules` | `/api/v1/backup-schedules/*` | ✅ Done |

                              ↓| PolarDBXBackupBinlog | `crd/polardbxbackupbinlogs` | `/api/v1/crd/polardbxbackupbinlogs` | `/api/v1/backup-binlogs/*` | ✅ Done |

┌─────────────────────────────────────────────────────────────┐| PolarDBXParameter | `crd/polardbxparameters` | `/api/v1/crd/polardbxparameters` | `/api/v1/parameters/*` | ✅ Done |

│                    Handler Layer                            │| PolarDBXParameterTemplate | `crd/polardbxparametertemplates` | `/api/v1/crd/polardbxparametertemplates` | `/api/v1/parameter-templates/*` | ✅ Done |

│  domain/*/handler/ → HTTP 请求/响应处理，DTO 转换            │| PolarDBXMonitor | `crd/polardbxmonitors` | `/api/v1/crd/polardbxmonitors` | `/api/v1/monitors/*` | ✅ Done |

│  - 解析请求参数                                              │| PolarDBXLogCollector | `crd/polardbxlogcollectors` | `/api/v1/crd/polardbxlogcollectors` | `/api/v1/log-collectors/*` | ✅ Done |

│  - 调用 Service                                              │| XStoreBackupBinlog | `crd/xstorebackupbinlogs` | `/api/v1/crd/xstorebackupbinlogs` | (new) | ✅ Done |

│  - 格式化响应                                                │

└─────────────────────────────────────────────────────────────┘Note: Alias routes are assembled by `api/router`, forwarding to existing handlers; the UI can continue using legacy routes.

                              ↓

┌─────────────────────────────────────────────────────────────┐---

│                    Service Layer                            │

│  domain/*/service/ → 业务逻辑，不依赖 HTTP                   │# Domain Entrances

│  - 接收纯业务参数 (context.Context, 业务对象)                │

│  - 返回业务对象或 error                                      │- Add domain-level entrances only (thin handlers forwarding), do not replace legacy routes:

│  - 依赖 Repository 接口                                      │  - Logical cluster domain: `/api/v1/polardbxclusters/*` ✅ Done

└─────────────────────────────────────────────────────────────┘  - Storage domain: `/api/v1/xstores/*` ✅ Done

                              ↓  - System tasks domain: `/api/v1/systemtasks/*` ✅ Done

┌─────────────────────────────────────────────────────────────┐  - Platform cross-cutting: `/api/v1/platform/*` ✅ Done

│                   Repository Layer                          │

│  domain/*/repository/ → 数据访问抽象                         │Where assembled: `pkg/api/router` exposes `RegisterCRDAliasRoutes` and `RegisterDomainRoutes`; they are registered during app init, and `LogGroupedRoutes` prints grouped logs for discoverability.

│  - 定义接口 (interface)                                      │

│  - K8s 实现 (k8s_*.go)                                       │---

│  - 便于 mock 测试                                            │

└─────────────────────────────────────────────────────────────┘# Package Ownership List (Migration Progress)

```

## ✅ Completed (Migrated to Domain)

---

| Domain | Package | Content |

## 目录结构规范| --- | --- | --- |

| `domain/polardbxclusters` | handlers.go | Cluster CRUD/ops, backup/schedule/binlog, parameters/templates, knobs, prechange, restore |

```| `domain/xstores` | handlers.go | XStore CRUD, followers, backups, rebuild |

pkg/api/| `domain/systemtasks` | handlers.go | SystemTask CRUD |

├── domain/                          # 业务域| `domain/monitoring` | handlers.go, crd_handlers.go | 监控安装工作流 + PolarDBXMonitor CRD CRUD |

│   ├── polardbxclusters/            # 集群域| `domain/platform` | handlers_*.go | 代理转发到旧包（alerts, grafana, logs, pod, settings, system, prometheusrule） |

│   │   ├── handler/                 # HTTP 处理层

│   │   │   ├── cluster_handler.go   # Handler 实现## 🔄 Phase 2: Pending Migration (Still in pkg/api/)

│   │   │   ├── dto.go               # 请求/响应 DTO

│   │   │   └── routes.go            # 路由注册These packages have thin forwarding wrappers in `domain/platform/handlers_*.go` but the actual logic is still in the old location:

│   │   ├── service/                 # 业务逻辑层

│   │   │   ├── cluster_service.go   # Service 实现| Old Package | Target Domain | Notes |

│   │   │   ├── interface.go         # Repository 接口定义| --- | --- | --- |

│   │   │   └── *_test.go            # 单元测试（mock repo）| `alerts/` | `domain/platform` | AlertManager integration |

│   │   ├── repository/              # 数据访问层| `auth/` | `domain/platform` | JWT auth (optional) |

│   │   │   ├── interface.go         # 接口定义（可与 service 共用）| `clusterknobs/` | `domain/polardbxclusters` | Already forwarded via domain handlers |

│   │   │   ├── k8s_cluster.go       # K8s 实现| `diagnostics/` | `domain/platform` | Cluster diagnostics |

│   │   │   └── mock_cluster.go      # Mock 实现（测试用）| `grafana/` | `domain/platform` | Dashboard sync/config |

│   │   └── entity/                  # 领域实体（可选）| `logcollector/` | `domain/platform` | LogCollector CRD management |

│   │       └── cluster.go| `logs/` | `domain/platform` | Log query/bootstrap |

│   │| `logservice/` | `domain/platform` | ES/Loki service status |

│   ├── xstores/                     # 存储域（同上结构）| `logstrategy/` | `domain/platform` | Log strategy CRUD |

│   ├── systemtasks/                 # 系统任务域| `parameters/` | `domain/polardbxclusters` | Already forwarded |

│   ├── monitoring/                  # 监控域| `pod/` | `domain/platform` | Pod CRUD/exec/logs |

│   └── platform/                    # 平台横切域| `prechange/` | `domain/polardbxclusters` | Already forwarded |

│       ├── handler/| `prometheusrule/` | `domain/platform` | PrometheusRule templates |

│       ├── service/| `restore/` | `domain/polardbxclusters` | Restore/PITR |

│       └── repository/| `settings/` | `domain/platform` | Dashboard thresholds, image registry |

│| `system/` | `domain/platform` | Namespaces, context info |

├── middleware/                      # 中间件| `systemtask/` | `domain/systemtasks` | Duplicate of domain/systemtasks |

│   ├── auth.go                      # 认证

│   ├── kubeconfig.go               # K8s 客户端注入---

│   └── cors.go                      # CORS

│# Recommended Next Steps

├── router/                          # 路由聚合

│   └── router.go                    # 注册所有路由## Priority 1: Remove Duplicates

│- [ ] Delete `pkg/api/systemtask/` (已迁移到 `domain/systemtasks`)

└── util/                            # 通用工具- [ ] Verify `clusterknobs/`, `parameters/`, `prechange/` are only used via domain forwarding

    ├── response.go                  # 响应格式化

    ├── errors.go                    # 错误处理## Priority 2: Consolidate Platform Handlers

    └── context.go                   # Context 工具Move actual logic from old packages to `domain/platform/`:

```- [ ] `system/` → `domain/platform/handlers_system.go` (move logic, not just forward)

- [ ] `pod/` → `domain/platform/handlers_pod.go`

---- [ ] `alerts/` → `domain/platform/handlers_alerts.go`

- [ ] `grafana/` → `domain/platform/handlers_grafana.go`

## 代码规范- [ ] `logs/`, `logservice/`, `logstrategy/` → `domain/platform/handlers_logs.go`

- [ ] `settings/` → `domain/platform/handlers_settings.go`

### 1. Handler 层（HTTP 适配）

## Priority 3: Final Cleanup

```go- [ ] Create `domain/platform/services/` for business logic

// handler/cluster_handler.go- [ ] Create `domain/platform/k8srepo/` for K8s access

package handler- [ ] Remove empty/forwarding-only old packages

- [ ] Update imports in main.go to use domain directly

type ClusterHandler struct {

    service *service.ClusterService---

    logger  *zap.Logger

}# Current Architecture Summary



func NewClusterHandler(svc *service.ClusterService) *ClusterHandler {```

    return &ClusterHandler{service: svc}pkg/api/

}├── monitoring/endpoints.go     # API 入口，只做路由定义

├── crd/                        # CRD 别名路由（11个子包）

// List 只处理 HTTP，不含业务逻辑├── router/router.go            # 注册 CRD 和 Domain 路由

func (h *ClusterHandler) List(c *gin.Context) {├── domain/

    // 1. 解析请求│   ├── monitoring/             # ✅ 完全迁移（handlers + service）

    namespace := c.DefaultQuery("namespace", "")│   ├── platform/               # 🔄 代理层（handlers_*.go 转发到旧包）

    │   ├── polardbxclusters/       # ✅ 完全迁移

    // 2. 调用 Service（传递 context.Context，不是 gin.Context）│   ├── systemtasks/            # ✅ 完全迁移

    clusters, err := h.service.List(c.Request.Context(), namespace)│   └── xstores/                # ✅ 完全迁移

    if err != nil {└── [旧包]/                      # 待迁移到 domain

        h.handleError(c, err)```

        return
    }
    
    // 3. 格式化响应
    c.JSON(http.StatusOK, clusters)
}
```

### 2. Service 层（业务逻辑）

```go
// service/cluster_service.go
package service

type ClusterService struct {
    repo   ClusterRepository  // 依赖接口
    logger *zap.Logger
}

func NewClusterService(repo ClusterRepository) *ClusterService {
    return &ClusterService{repo: repo}
}

// List 纯业务逻辑，不知道 HTTP
func (s *ClusterService) List(ctx context.Context, namespace string) ([]*dto.ClusterResponse, error) {
    clusters, err := s.repo.List(ctx, namespace)
    if err != nil {
        return nil, fmt.Errorf("list clusters: %w", err)
    }
    return toClusterResponses(clusters), nil
}
```

### 3. Repository 层（数据访问）

```go
// service/interface.go (或 repository/interface.go)
package service

type ClusterRepository interface {
    List(ctx context.Context, namespace string) ([]*polardbxv1.PolarDBXCluster, error)
    Get(ctx context.Context, namespace, name string) (*polardbxv1.PolarDBXCluster, error)
    Create(ctx context.Context, cluster *polardbxv1.PolarDBXCluster) (*polardbxv1.PolarDBXCluster, error)
    Update(ctx context.Context, cluster *polardbxv1.PolarDBXCluster) (*polardbxv1.PolarDBXCluster, error)
    Delete(ctx context.Context, namespace, name string) error
}

// repository/k8s_cluster.go
package repository

type K8sClusterRepository struct {
    // 注入 K8s client 工厂或直接使用 context 获取
}

func (r *K8sClusterRepository) List(ctx context.Context, namespace string) ([]*polardbxv1.PolarDBXCluster, error) {
    client := util.K8sClientFromCtx(ctx)
    // ... K8s 操作
}
```

---

## CRD 路由映射

| CRD Kind | 路由前缀 | Domain |
| --- | --- | --- |
| PolarDBXCluster | `/api/v1/clusters`, `/api/v1/crd/polardbxclusters` | `polardbxclusters` |
| XStore | `/api/v1/xstores`, `/api/v1/crd/xstores` | `xstores` |
| SystemTask | `/api/v1/system-tasks`, `/api/v1/crd/systemtasks` | `systemtasks` |
| PolarDBXMonitor | `/api/v1/monitors`, `/api/v1/crd/polardbxmonitors` | `monitoring` |
| PolarDBXBackup | `/api/v1/backups` | `polardbxclusters` |
| PolarDBXBackupSchedule | `/api/v1/backup-schedules` | `polardbxclusters` |
| PolarDBXBackupBinlog | `/api/v1/backup-binlogs` | `polardbxclusters` |
| PolarDBXParameter | `/api/v1/parameters` | `polardbxclusters` |
| PolarDBXParameterTemplate | `/api/v1/parameter-templates` | `polardbxclusters` |
| PolarDBXLogCollector | `/api/v1/log-collectors` | `platform` |

---

## 迁移进度

### ✅ Phase 1: 路由别名（已完成）
- [x] CRD 别名路由 `/api/v1/crd/*`
- [x] Domain 入口路由 `/api/v1/polardbxclusters/*` 等

### 🔄 Phase 2: 架构重构（进行中）

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
