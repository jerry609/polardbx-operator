# PolarDB-X dashborad现状分析报告

## 报告概述

**技术栈**: Angular 19前端应用 + Go Gin后端API服务  

---

## 架构设计分析

### 1. 业务模块组织

#### 备份管理模块 - 统一数据保护体系
**功能范围**: 涵盖完整的数据备份生命周期管理
- **手动备份**: 支持即时全量备份，用户可选择备份范围和存储位置
- **定时备份**: 基于Cron表达式的自动化备份调度，支持增量和全量策略
- **二进制日志备份**: MySQL binlog实时备份，支持时间点恢复(PITR)能力
- **存储级备份**: XStore存储引擎原生备份，支持多种存储后端(OSS/S3/SFTP)

**实现说明与现状**
- 前端支持 s3/oss/sftp 三种存储提供商，表单默认项为 s3；占位示例以 s3 为主。
- 后端为 CR 透传创建：`CreateBackup` 直接提交 `PolarDBXBackup` 对象，不区分 provider 分支逻辑；可用性取决于 Operator 实现与集群内 Secret/权限配置是否就绪。
- 手动备份列表的状态使用 `status.phase` 映射显示：当 `status` 尚未写回或阶段值超出映射集（pending/running/completed/succeeded/failed/deleting）时显示“未知”。常见原因：
  - 备份对象刚创建，Operator 尚未更新 `status.phase`；
  - 实际阶段大小写或命名与前端映射不一致；
  - 列表筛选依赖 `polardbx/name=<cluster>` 标签，若标签延迟或不一致，可能读到旧对象或状态字段为空。
- 409 冲突：后端将 K8s `AlreadyExists` 映射为 409，建议采用唯一命名或 `metadata.generateName`（若 CRD 支持）。

**未完成/待对齐**
- 提供 OSS/SFTP 的 Secret 模板与 sink 示例，完成与 Operator 的联合验证。
- 对 `status.phase` 完整枚举与大小写规范进行对齐，补齐前端映射与颜色标签。
- 备份进度可观测性：进度推送（WebSocket/SSE）、失败原因细化、结构化日志字段定义。

**下一步工作（建议）**
- 表单联动：根据 s3/oss/sftp 动态展示必填字段与示例，增加 sink/Secret 的基础校验。
- 命名策略：默认追加时间戳后缀，出现 409 时提供友好提示与“自动命名”选项。
- 列表体验：无状态→显示“创建中”；对同名对象短期缓存 lastPhase，未写回时显示“同步中…”。

####  恢复管理模块 - 智能数据恢复
**技术特色**: 向导式操作流程，降低恢复操作复杂度
- **恢复向导**: 多步骤引导式恢复流程，支持数据校验和回滚机制
- **恢复任务管理**: 实时监控恢复进度，支持任务暂停/继续/取消操作
- **时间点恢复(PITR)**: 基于binlog的精确时间点恢复，秒级精度选择

**设计思路**
- 前端：采用分步式向导（恢复来源选择→参数校验→执行确认→进度监控），状态由服务统一管理；PITR 提供时间选择器与时区支持。
- 后端：提供恢复请求API（包括PITR），暴露恢复状态查询端点；对接 Operator 的 Phase/Conditions，输出结构化状态与错误信息。
- 可靠性：前端定时轮询或WebSocket订阅恢复进度；错误重试与失败回滚提示；所有操作留存可审计日志。

####  存储管理模块 - 存储架构治理
**管理对象**: XStore分布式存储节点和副本拓扑
- **存储节点管理**: XStore实例生命周期管理，支持扩缩容和配置热更新
- **存储副本管理**: XStoreFollower故障恢复，自动化备库重搭机制

####  运维管理模块 - 全栈运维体系
**覆盖范围**: 从监控告警到性能调优的完整运维工具链
- **监控管理**: Prometheus集成，自定义指标采集和告警规则配置
- **参数模板管理**: 数据库参数标准化管理，支持版本控制和批量应用
- **系统任务管理**: 后台任务调度和执行状态监控，支持依赖关系定义
- **日志收集管理**: 集中化日志收集配置，支持多种日志源和目标
- **集群调优管理**: PolarDBXClusterKnobs性能参数优化，智能推荐机制

**设计思路**
- 监控：以 CRD 申明监控目标与抓取频率，后端统一生成/维护 Prometheus 配置；前端提供可视化仪表并支持阈值/告警规则配置。
- 系统任务：以 SystemTask CRD 驱动任务编排；后端提供任务CRUD与进度查询；前端显示阶段/进度/日志并支持暂停/重试。
- 日志收集：以 LogCollector CRD 描述采集器组件（FileBeat/LogStash），后端映射状态；前端呈现组件健康度与流量，支持一键启停。
- 调优：以参数模板驱动参数集；前端提供差异比对与影响提示；后端校验参数并执行滚动下发。

**实现细节与现状（按子模块）**
- 监控管理（PolarDBXMonitor）
  - 前端：监控列表/创建表单/详情查看，支持采集间隔与目标配置；后续将提供规则编辑与阈值校验。
  - 后端：提供 CRUD API，透传 CRD；建议在 Operator 侧生成 scrape 配置，或由后端生成 ConfigMap/CR 支撑。
  - 待对齐：默认抓取端点/指标集、样例 Dashboard、规则模板与告警路由约定。
- 参数模板管理（ParameterTemplate）
  - 前端：模板列表/创建/编辑，按 CN/DN/GMS 分类展示；新增模板差异比对与影响提示。
  - 后端：CRUD API 与参数校验透传；建议补充“预检”接口验证风险与重启要求。
  - 待对齐：参数枚举/范围的权威来源、只读参数保护、批量下发与回滚策略。
- 系统任务管理（SystemTask）
  - 前端：任务列表/详情，显示阶段（如 Pending/Running/Succeeded/Failed）与进度/日志；支持暂停/重试操作。
  - 后端：CRUD/状态查询 API；与 Operator 的 Phase/Conditions 对齐。
  - 待对齐：任务类型全集（除资源平衡外是否扩展）、状态机定义、失败重试与并发策略、日志流式查看。
- 日志收集管理（LogCollector）
  - 前端：组件健康度、吞吐/错误率概览，一键启停；后续支持 pipeline 级可视化（输入→过滤→输出）。
  - 后端：CRUD/状态映射 API；建议补充 conditions→UI 标签映射表与常见错误分类。
  - 待对齐：采集路径与过滤规则模板、存储与保留策略、告警联动。
- 集群调优管理（ClusterKnobs）
  - 前端：调优参数分组展示与编辑，提供影响等级提示与安全护栏（危险参数二次确认）。
  - 后端：CRUD 与参数校验透传；“干跑（dry-run）”与“按批滚动下发”暂不支持。
  - 待对齐：A/B 测试方案、效果评估指标、自动化建议来源与采样窗口。

**未完成/风险项**
- 监控：
  - 样例 Dashboard 与常用指标集（CPU/内存/连接/延迟/错误率）未内置；
  - 告警规则模板与通知渠道（邮件/短信/钉钉）联动待落地；
  - 多命名空间/多集群聚合视图与面板权限控制待设计。
- 参数模板：
  - 预检接口与只读参数保护未完成；批量应用/回滚与“变更窗口”未实现；
  - 变更审计维度（谁/何时/变更内容）与回滚点标记需完善。
- 系统任务：
  - 任务类型扩展（如巡检、索引重建、健康修复）与统一状态机未对齐；
  - 任务日志流式拉取与长期归档未实现。
- 日志收集：
  - 常见错误分类与自愈建议未内置；吞吐与延迟的趋势图未完成；
  - Pipeline 可视化与规则模板中心未实现。
- 调优：
  - A/B 测试与效果评估缺少闭环；危险参数的多级审批/风控未落地。


### 2. 路由架构设计
```yaml
设计原则:
  - 模块化: 按业务功能清晰分组
  - 层次化: 主路由 + 子路由结构
  - 便捷性: 提供顶层别名路由
  - 一致性: 统一的命名规范

技术特点:
  - 懒加载: 所有子模块支持按需加载
  - 类型安全: 完整的TypeScript路由定义
  - 守卫机制: 统一的认证守卫保护
```

---

##  前后端对应关系详细分析

###  核心管理功能 - 平台基础

#### `/connect` - 认证连接管理
**前端实现**: 连接配置向导，支持Kubeconfig文件上传和Base64编码
**后端API**: 
- `POST /api/v1/connect` - Kubeconfig连接验证，建立安全会话
**技术特色**: Header传输认证信息，统一的中间件验证机制

#### `/clusters` - 集群生命周期管理  
**前端实现**: 集群列表视图，支持筛选、排序和批量操作
**后端API集群**:
- `GET /api/v1/clusters` - 跨命名空间集群列表查询，支持标签选择器
- `POST /api/v1/clusters` - 集群创建，支持模板化配置和参数验证
- `GET /api/v1/clusters/:namespace/:name` - 集群详情获取，包含状态聚合信息
- `PUT /api/v1/clusters/:namespace/:name` - 集群配置热更新，支持滚动升级
- `DELETE /api/v1/clusters/:namespace/:name` - 安全删除，支持数据保护检查
- `GET /api/v1/clusters/:namespace/:name/pods` - Pod拓扑查看，实时状态监控

#### `/clusters/:ns/:name` - 集成管理中心
**架构设计**: 将高频访问功能集成在集群详情页，提升操作效率
**日志查询功能**:
- `GET /api/v1/logs/:namespace/:pod_name` - 实时日志流，支持关键词高亮和时间范围筛选
**参数管理功能**:
- `GET /api/v1/parameters` - 参数列表查询，支持分类和模糊搜索
- `POST /api/v1/parameters` - 参数创建，支持类型验证和取值范围检查
- `GET /api/v1/parameters/:name` - 参数详情获取，包含历史变更记录
- `PUT /api/v1/parameters/:name` - 参数值更新，支持热更新和回滚机制
- `DELETE /api/v1/parameters/:name` - 参数安全删除，依赖关系检查

---

### 备份管理模块 - 数据保护体系

#### `/backup/manual-backups` - 即时备份操作
**前端组件**: 备份创建表单，支持备份范围选择和存储配置
**后端API能力**:
- `GET /api/v1/clusters/:namespace/:name/backups` - 备份历史查询，支持状态筛选和时间排序
- `POST /api/v1/clusters/:namespace/:name/backups` - 备份任务创建，支持增量/全量策略选择
- `DELETE /api/v1/backups/:namespace/:name` - 备份文件清理，支持批量删除和存储回收

#### `/backup/backup-schedules` - 自动化备份调度
**前端特性**: Cron表达式编辑器，可视化调度策略配置
**后端API功能**:
- `GET /api/v1/backup-schedules` - 调度任务列表，支持状态监控和执行历史
- `POST /api/v1/backup-schedules` - 调度策略创建，支持复杂时间表达式和冲突检测
- `GET /api/v1/backup-schedules/:namespace/:name` - 调度详情查看，包含下次执行时间预测
- `PUT /api/v1/backup-schedules/:namespace/:name` - 调度规则更新，支持暂停/恢复和热更新
- `DELETE /api/v1/backup-schedules/:namespace/:name` - 调度任务删除，安全清理相关资源

#### `/backup/backup-binlogs` - 二进制日志管理
**技术实现**: MySQL binlog实时采集，支持压缩和加密传输
**后端API架构**:
- `GET /api/v1/backup-binlogs` - Binlog配置列表，支持集群维度筛选
- `POST /api/v1/backup-binlogs` - Binlog备份配置，支持多种存储后端选择
- `GET /api/v1/backup-binlogs/:namespace/:name` - 配置详情查看，包含采集统计信息
- `PUT /api/v1/backup-binlogs/:namespace/:name` - 配置参数调整，支持采集频率和保留策略
- `DELETE /api/v1/backup-binlogs/:namespace/:name` - 配置安全删除，停止采集任务

#### `/backup/xstore-backups` - 存储级备份
**技术优势**: XStore引擎原生备份，支持快照和增量备份
**后端API设计**:
- `GET /api/v1/xstore-backups` - 存储备份列表，支持多维度筛选和排序
- `POST /api/v1/xstore-backups` - 备份任务创建，支持存储卷选择和压缩配置
- `GET /api/v1/xstore-backups/:namespace/:name` - 备份详情查看，包含完整性校验结果
- `PUT /api/v1/xstore-backups/:namespace/:name` - 备份属性修改，支持标签和描述更新
- `DELETE /api/v1/xstore-backups/:namespace/:name` - 备份数据清理，支持存储空间回收

---

###  恢复管理模块 - 智能数据恢复

#### `/recovery/restore-wizard` - 向导式恢复流程
**用户体验**: 多步骤向导界面，智能恢复建议和风险评估
**后端API支持**:
- `POST /api/v1/clusters/:namespace/:name/restore` - 集群恢复操作，支持完整性验证
- `GET /api/v1/clusters/:namespace/:name/restore-status` - 恢复状态实时查询，包含进度百分比

#### `/recovery/restore-jobs` - 恢复任务监控
**实时特性**: WebSocket连接，实时进度推送和状态变更通知
**后端API功能**:
- `GET /api/v1/restore-jobs` - 恢复任务列表，支持多集群聚合查看
- `GET /api/v1/restore-jobs/:namespace/:name` - 任务详情查看，包含执行日志和错误诊断
- `DELETE /api/v1/restore-jobs/:namespace/:name` - 任务取消操作，支持优雅停止和资源清理

#### `/recovery/pitr` - 时间点恢复
**精度控制**: 秒级时间选择器，binlog位点精确定位
**后端API实现**:
- `POST /api/v1/clusters/:namespace/:name/pitr` - PITR恢复启动，支持时间范围验证和冲突检测

---

###  存储管理模块 - 存储架构治理

#### `/storage/xstores` - 存储节点管理
**管理范围**: XStore分布式存储节点的完整生命周期
**后端API架构**:
- `GET /api/v1/xstores` - 存储节点列表，支持健康状态和性能指标查看
- `POST /api/v1/xstores` - 新节点创建，支持自动配置和拓扑优化
- `GET /api/v1/xstores/:namespace/:name` - 节点详情查看，包含资源使用和连接状态
- `PUT /api/v1/xstores/:namespace/:name` - 节点配置更新，支持在线扩缩容和参数调优
- `DELETE /api/v1/xstores/:namespace/:name` - 节点安全下线，支持数据迁移和副本重分布

#### `/storage/xstore-followers` - 存储副本管理
**核心功能**: 故障恢复和备库重搭自动化
**后端API设计**:
- `GET /api/v1/xstore-followers` - 副本状态列表，支持同步延迟和健康检查
- `POST /api/v1/xstore-followers` - 副本创建配置，支持同步策略和故障切换设置
- `GET /api/v1/xstore-followers/:namespace/:name` - 副本详情查看，包含同步统计和性能指标
- `PUT /api/v1/xstore-followers/:namespace/:name` - 副本参数调整，支持同步模式和优先级配置
- `DELETE /api/v1/xstore-followers/:namespace/:name` - 副本安全删除，支持数据一致性检查

---

### 运维管理模块 - 全栈运维体系

#### `/operations/monitors` - 监控管理
**集成能力**: Prometheus生态深度集成，自定义指标和告警规则
**后端API功能**:
- `GET /api/v1/monitors` - 监控配置列表，支持启用状态和采集频率查看
- `POST /api/v1/monitors` - 监控规则创建，支持指标选择和阈值配置
- `GET /api/v1/monitors/:namespace/:name` - 监控详情查看，包含历史数据和趋势分析
- `PUT /api/v1/monitors/:namespace/:name` - 规则参数更新，支持动态阈值和告警策略
- `DELETE /api/v1/monitors/:namespace/:name` - 监控规则删除，清理相关告警和数据

#### `/operations/parameter-templates` - 参数模板管理
**标准化特性**: 数据库参数最佳实践模板化，支持版本控制
**后端API架构**:
- `GET /api/v1/parameter-templates` - 模板列表查询，支持分类和版本筛选
- `POST /api/v1/parameter-templates` - 模板创建配置，支持参数验证和兼容性检查
- `GET /api/v1/parameter-templates/:namespace/:name` - 模板详情查看，包含参数说明和应用历史
- `PUT /api/v1/parameter-templates/:namespace/:name` - 模板内容更新，支持版本管理和变更审计
- `DELETE /api/v1/parameter-templates/:namespace/:name` - 模板安全删除，检查依赖关系

#### `/operations/system-tasks` - 系统任务管理
**调度能力**: 后台任务编排，支持依赖关系和并发控制
**后端API设计**:
- `GET /api/v1/system-tasks` - 任务列表查询，支持状态筛选和执行历史
- `POST /api/v1/system-tasks` - 任务创建配置，支持调度策略和资源限制
- `GET /api/v1/system-tasks/:namespace/:name` - 任务详情查看，包含执行日志和性能统计
- `PUT /api/v1/system-tasks/:namespace/:name` - 任务参数调整，支持暂停/恢复和优先级设置
- `DELETE /api/v1/system-tasks/:namespace/:name` - 任务删除清理，停止执行并回收资源

#### `/operations/log-collectors` - 日志收集管理
**集中化架构**: 多源日志统一采集，支持实时处理和存储路由
**后端API功能**:
- `GET /api/v1/log-collectors` - 收集器列表，支持状态监控和流量统计
- `POST /api/v1/log-collectors` - 收集器创建，支持多种日志源和目标配置
- `GET /api/v1/log-collectors/:namespace/:name` - 收集器详情，包含吞吐量和错误率统计
- `PUT /api/v1/log-collectors/:namespace/:name` - 配置参数更新，支持过滤规则和路由策略
- `DELETE /api/v1/log-collectors/:namespace/:name` - 收集器删除，安全停止采集任务

#### `/operations/cluster-knobs` - 集群调优管理
**智能优化**: PolarDBXClusterKnobs性能参数自动调优和建议推荐
**后端API实现**:
- `GET /api/v1/cluster-knobs` - 调优配置列表，支持性能指标关联查看
- `POST /api/v1/cluster-knobs` - 调优策略创建，支持参数组合验证和影响评估
- `GET /api/v1/cluster-knobs/:namespace/:name` - 调优详情查看，包含参数建议和性能对比
- `PUT /api/v1/cluster-knobs/:namespace/:name` - 参数配置更新，支持A/B测试和回滚机制
- `DELETE /api/v1/cluster-knobs/:namespace/:name` - 调优配置删除，恢复默认参数设置

---

## 总结

### 近期变更与待处理事项
- 终端(WebShell)重构完成：
  - 前端接入标准 xterm.js 流程（onData→stdin JSON，onmessage→stdout/stderr JSON，FitAddon 同步尺寸）。
  - 后端 Exec 通道改为 WebSocket→io.Pipe→remotecommand.Stream，协议统一为 JSON（stdin/resize/stdout/stderr）。
  - 交互延迟大幅改善，输入回显稳定，断线重连与错误提示完善。
- 列表页请求频繁问题定位：
  - `cluster-list` 页面每 5 秒轮询 `/api/v1/clusters`，路由切换时需确保 `unsubscribe()`，避免后台持续请求。
- 备份创建 409 冲突：
  - 已确认后端将 `AlreadyExists` 映射为 409；需前端生成唯一备份名或使用 `generateName`。

短期待处理：
1) 备份创建幂等与命名策略
   - 方案A：前端自动加时间戳/随机后缀；
   - 方案B：使用 `metadata.generateName`（如 CRD 支持）。
2) 列表轮询优化
   - 在 `ngOnDestroy()` 中 `this.refreshInterval?.unsubscribe()`；
   - 仅页面可见时轮询（visibilitychange）；
   - 允许用户在 UI 中配置轮询频率（如 5s/30s/关闭）。
3) 终端 UX 细节
   - 暗/亮主题联动、字体大小快捷调节、复制/粘贴提示优化；
   - 终端内容下载保留 ANSI 样式（SerializeAddon 配置）。

### 中期优化计划
- WebShell：
  - 支持容器切换热重连、连接保持心跳、历史会话恢复。
  - Resize 节流与合并，进一步减少无效请求。
- 备份/恢复：
  - 备份执行进度推送（WebSocket/SSE），前端进度条显示；
  - 统一的操作结果页面（成功/失败/可重试）。
- 列表与详情页面：
  - 统一轮询与可见性策略，消除隐藏页后台轮询；
  - 请求合并与缓存（ETag/If-None-Match）。

### 长期规划（考虑）
- 监控与可观测性：整合 Prometheus 指标与告警至界面，操作链路可视化。
- 自动化能力：参数模板 A/B 调优、恢复演练自动化、任务编排可视化。
- 多集群管理：统一视图、跨集群切换。


### 3. 测试与覆盖率口径
- 后端覆盖率 64.4% 为 `pkg/api` 包在当前测试集下的综合覆盖率（单元+集成测试，`go test -coverprofile` 汇总）。
- 集成测试：29 项端点级用例，覆盖 CRUD、错误映射、异常分支；`main` 不计入。
- 其他包：`pkg/k8s` 当前为 27.1%（单元为主），后续补齐 CRD 字段/分支覆盖。
- 前端与 E2E 浏览器测试独立统计，不计入该覆盖率。