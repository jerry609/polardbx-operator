# CRD ↔ Backend API 映射（别名路由）

- 基础前缀：`/api/v1/crd/*`（仅新增别名，不改变现有接口与入参/出参）

| CRD Kind | 别名包 | 别名路由 | 现有主路由（示例） |
| --- | --- | --- | --- |
| PolarDBXCluster | `crd/polardbxclusters` | `/api/v1/crd/polardbxclusters` | `/api/v1/clusters` (domain: `/api/v1/polardbxclusters`) |
| SystemTask | `crd/systemtasks` | `/api/v1/crd/systemtasks` | `/api/v1/system-tasks` |
| XStore | `crd/xstores` | `/api/v1/crd/xstores` | `/api/v1/xstores` |
| PolarDBXBackup | `crd/polardbxbackups` | `/api/v1/crd/polardbxbackups` | `/api/v1/backups/*` |
| PolarDBXBackupSchedule | `crd/polardbxbackupschedules` | `/api/v1/crd/polardbxbackupschedules` | `/api/v1/backup-schedules/*` |
| PolarDBXBackupBinlog | `crd/polardbxbackupbinlogs` | `/api/v1/crd/polardbxbackupbinlogs` | `/api/v1/backup-binlogs/*` |
| PolarDBXParameter | `crd/polardbxparameters` | `/api/v1/crd/polardbxparameters` | `/api/v1/parameters/*` |
| PolarDBXParameterTemplate | `crd/polardbxparametertemplates` | `/api/v1/crd/polardbxparametertemplates` | `/api/v1/parameter-templates/*` |
| PolarDBXMonitor | `crd/polardbxmonitors` | `/api/v1/crd/polardbxmonitors` | `/api/v1/monitors/*` |
| PolarDBXLogCollector | `crd/polardbxlogcollectors` | `/api/v1/crd/polardbxlogcollectors` | `/api/v1/log-collectors/*` |

说明：别名路由由 `api/router` 统一装配，底层转发到现有 handlers；UI 可继续使用旧路由。

---

# 领域入口（Domain Entrances）

- 仅新增域级入口（薄 handlers 转发），不替换旧路由：
  - 逻辑集群域：`/api/v1/polardbxclusters/*`（聚合 clusters/backups/backup-schedules/backup-binlogs/parameters/parameter-templates/prechange/restore/cluster-knobs）
  - 存储域：`/api/v1/xstores/*`（聚合 xstores/xstore-backups/xstore-followers/rebuild）
  - 平台任务域：`/api/v1/systemtasks/*`（聚合 system-tasks）
  - 平台横切：`/api/v1/platform/*`（聚合 monitoring/grafana/logs/system/pod 等）

装配位置：`pkg/api/router` 暴露 `RegisterCRDAliasRoutes` 与 `RegisterDomainRoutes`，主程序初始化时注册并通过 `LogGroupedRoutes` 输出分组日志，方便发现性。

---

# 包归属清单（指导迁移，不影响现有路由）

- 逻辑集群域 `domain/polardbxclusters`
  - 归属：`cluster/`（CRUD/运维已接线到 services）、`backup/`, `backupbinlog/`, `parameters/`, `clusterknobs/`, `prechange/`, `restore/`
- 存储域 `domain/xstores`
  - 归属：`xstore/`（含 followers/rebuild 与 xstore-backups）
- 平台任务域 `domain/systemtasks`
  - 归属：`systemtask/`
- 平台横切 `domain/platform`
  - 归属：`monitor/`, `monitoring/`, `grafana/`, `logs/`, `logservice/`, `logstrategy/`, `system/`, `pod/`, `alerts/`, `settings/`, `auth/`

说明：第一阶段仅建立入口与骨架；第二阶段按上述归属逐步迁移 handlers → 领域包，并将编排下沉至 `services/`，K8s 访问归一到 `k8srepo/`，常量放入 `meta/`。
