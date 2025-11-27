# CRD ↔ Backend API Mapping (Alias Routes)

- Base prefix: `/api/v1/crd/*` (only adds aliases, does not change existing endpoints or request/response)

| CRD Kind | Alias Package | Alias Route | Existing Primary Route (example) | Status |
| --- | --- | --- | --- | --- |
| PolarDBXCluster | `crd/polardbxclusters` | `/api/v1/crd/polardbxclusters` | `/api/v1/clusters` (domain: `/api/v1/polardbxclusters`) | ✅ Done |
| SystemTask | `crd/systemtasks` | `/api/v1/crd/systemtasks` | `/api/v1/system-tasks` | ✅ Done |
| XStore | `crd/xstores` | `/api/v1/crd/xstores` | `/api/v1/xstores` | ✅ Done |
| PolarDBXBackup | `crd/polardbxbackups` | `/api/v1/crd/polardbxbackups` | `/api/v1/backups/*` | ✅ Done |
| PolarDBXBackupSchedule | `crd/polardbxbackupschedules` | `/api/v1/crd/polardbxbackupschedules` | `/api/v1/backup-schedules/*` | ✅ Done |
| PolarDBXBackupBinlog | `crd/polardbxbackupbinlogs` | `/api/v1/crd/polardbxbackupbinlogs` | `/api/v1/backup-binlogs/*` | ✅ Done |
| PolarDBXParameter | `crd/polardbxparameters` | `/api/v1/crd/polardbxparameters` | `/api/v1/parameters/*` | ✅ Done |
| PolarDBXParameterTemplate | `crd/polardbxparametertemplates` | `/api/v1/crd/polardbxparametertemplates` | `/api/v1/parameter-templates/*` | ✅ Done |
| PolarDBXMonitor | `crd/polardbxmonitors` | `/api/v1/crd/polardbxmonitors` | `/api/v1/monitors/*` | ✅ Done |
| PolarDBXLogCollector | `crd/polardbxlogcollectors` | `/api/v1/crd/polardbxlogcollectors` | `/api/v1/log-collectors/*` | ✅ Done |
| XStoreBackupBinlog | `crd/xstorebackupbinlogs` | `/api/v1/crd/xstorebackupbinlogs` | (new) | ✅ Done |

Note: Alias routes are assembled by `api/router`, forwarding to existing handlers; the UI can continue using legacy routes.

---

# Domain Entrances

- Add domain-level entrances only (thin handlers forwarding), do not replace legacy routes:
  - Logical cluster domain: `/api/v1/polardbxclusters/*` ✅ Done
  - Storage domain: `/api/v1/xstores/*` ✅ Done
  - System tasks domain: `/api/v1/systemtasks/*` ✅ Done
  - Platform cross-cutting: `/api/v1/platform/*` ✅ Done

Where assembled: `pkg/api/router` exposes `RegisterCRDAliasRoutes` and `RegisterDomainRoutes`; they are registered during app init, and `LogGroupedRoutes` prints grouped logs for discoverability.

---

# Package Ownership List (Migration Progress)

## ✅ Completed (Migrated to Domain)

| Domain | Package | Content |
| --- | --- | --- |
| `domain/polardbxclusters` | handlers.go | Cluster CRUD/ops, backup/schedule/binlog, parameters/templates, knobs, prechange, restore |
| `domain/xstores` | handlers.go | XStore CRUD, followers, backups, rebuild |
| `domain/systemtasks` | handlers.go | SystemTask CRUD |
| `domain/monitoring` | handlers.go, crd_handlers.go | 监控安装工作流 + PolarDBXMonitor CRD CRUD |
| `domain/platform` | handlers_*.go | 代理转发到旧包（alerts, grafana, logs, pod, settings, system, prometheusrule） |

## 🔄 Phase 2: Pending Migration (Still in pkg/api/)

These packages have thin forwarding wrappers in `domain/platform/handlers_*.go` but the actual logic is still in the old location:

| Old Package | Target Domain | Notes |
| --- | --- | --- |
| `alerts/` | `domain/platform` | AlertManager integration |
| `auth/` | `domain/platform` | JWT auth (optional) |
| `clusterknobs/` | `domain/polardbxclusters` | Already forwarded via domain handlers |
| `diagnostics/` | `domain/platform` | Cluster diagnostics |
| `grafana/` | `domain/platform` | Dashboard sync/config |
| `logcollector/` | `domain/platform` | LogCollector CRD management |
| `logs/` | `domain/platform` | Log query/bootstrap |
| `logservice/` | `domain/platform` | ES/Loki service status |
| `logstrategy/` | `domain/platform` | Log strategy CRUD |
| `parameters/` | `domain/polardbxclusters` | Already forwarded |
| `pod/` | `domain/platform` | Pod CRUD/exec/logs |
| `prechange/` | `domain/polardbxclusters` | Already forwarded |
| `prometheusrule/` | `domain/platform` | PrometheusRule templates |
| `restore/` | `domain/polardbxclusters` | Restore/PITR |
| `settings/` | `domain/platform` | Dashboard thresholds, image registry |
| `system/` | `domain/platform` | Namespaces, context info |
| `systemtask/` | `domain/systemtasks` | Duplicate of domain/systemtasks |

---

# Recommended Next Steps

## Priority 1: Remove Duplicates
- [ ] Delete `pkg/api/systemtask/` (已迁移到 `domain/systemtasks`)
- [ ] Verify `clusterknobs/`, `parameters/`, `prechange/` are only used via domain forwarding

## Priority 2: Consolidate Platform Handlers
Move actual logic from old packages to `domain/platform/`:
- [ ] `system/` → `domain/platform/handlers_system.go` (move logic, not just forward)
- [ ] `pod/` → `domain/platform/handlers_pod.go`
- [ ] `alerts/` → `domain/platform/handlers_alerts.go`
- [ ] `grafana/` → `domain/platform/handlers_grafana.go`
- [ ] `logs/`, `logservice/`, `logstrategy/` → `domain/platform/handlers_logs.go`
- [ ] `settings/` → `domain/platform/handlers_settings.go`

## Priority 3: Final Cleanup
- [ ] Create `domain/platform/services/` for business logic
- [ ] Create `domain/platform/k8srepo/` for K8s access
- [ ] Remove empty/forwarding-only old packages
- [ ] Update imports in main.go to use domain directly

---

# Current Architecture Summary

```
pkg/api/
├── monitoring/endpoints.go     # API 入口，只做路由定义
├── crd/                        # CRD 别名路由（11个子包）
├── router/router.go            # 注册 CRD 和 Domain 路由
├── domain/
│   ├── monitoring/             # ✅ 完全迁移（handlers + service）
│   ├── platform/               # 🔄 代理层（handlers_*.go 转发到旧包）
│   ├── polardbxclusters/       # ✅ 完全迁移
│   ├── systemtasks/            # ✅ 完全迁移
│   └── xstores/                # ✅ 完全迁移
└── [旧包]/                      # 待迁移到 domain
```
