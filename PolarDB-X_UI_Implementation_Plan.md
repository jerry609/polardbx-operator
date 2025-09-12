n# PolarDB-X 可视化运维平台实现文档

## 📋 项目概述

基于PolarDB-X Operator现有能力，实现一个完整的可视化运维平台，支持集群全生命周期管理、日志采集、性能监控等核心功能。

## 🎯 实现目标

### 核心目标
- 通过一个kubeconfig文件即可管理开源版PolarDB-X
- 支持集群创建、变配、升级、重搭、备份恢复等基本运维操作  
- 提供日志采集配置、性能监控、诊断工具等高级功能
- 降低普通开发者的运维门槛，减少黑屏操作

### 功能边界
- 专注于PolarDB-X集群运维管理
- 不涉及数据库业务逻辑操作
- 基于Kubernetes CRD进行管理
- 支持多集群、多命名空间场景

## 🏗️ 系统架构

```
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│   前端 Angular UI   │    │   后端 Go API       │    │  Kubernetes Cluster │
│                     │    │                     │    │                     │
│ • 集群管理界面      │    │ • CRD API 封装      │    │ • PolarDB-X CRDs   │
│ • 日志配置界面      │◄──►│ • 监控数据聚合      │◄──►│ • Operator 控制器   │
│ • 监控仪表板        │    │ • 事件流处理        │    │ • 监控组件          │
│ • 诊断工具界面      │    │ • WebSocket 推送    │    │ • 日志采集组件      │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
```

## 📊 功能模块设计

### 1. 集群生命周期管理模块

#### 1.1 集群创建向导
**文件**: `cluster-creation-wizard.component.ts`

**功能描述**: 
- 多步骤表单引导用户创建PolarDB-X集群
- 支持拓扑配置、资源配置、存储配置
- 实时校验和配置预览

**关键特性**:
```typescript
interface ClusterCreationConfig {
  // 基本信息
  name: string;
  namespace: string;
  description?: string;
  
  // 拓扑配置
  topology: {
    cn: { replicas: number; resources: ResourceRequirements };
    dn: { replicas: number; resources: ResourceRequirements };
    gms: { replicas: number; resources: ResourceRequirements };
    cdc?: { replicas: number; resources: ResourceRequirements };
  };
  
  // 存储配置
  storage: {
    storageClassName: string;
    size: string;
  };
  
  // 网络配置
  serviceType: 'ClusterIP' | 'LoadBalancer' | 'NodePort';
  
  // 安全配置
  rootPassword?: string;
  enableTLS?: boolean;
}
```

**实现步骤**:
1. 创建多步骤表单组件
2. 实现配置校验逻辑
3. 集成PolarDBXCluster CRD API
4. 添加创建进度监控

#### 1.2 集群管理列表
**文件**: `cluster-management.component.ts`

**功能描述**:
- 展示所有PolarDB-X集群状态
- 支持集群基本操作和状态监控
- 提供快速操作入口

**数据模型**:
```typescript
interface ClusterSummary {
  name: string;
  namespace: string;
  phase: string;
  version: string;
  nodes: {
    cn: { ready: number; total: number };
    dn: { ready: number; total: number };
    gms: { ready: number; total: number };
  };
  creationTime: string;
  lastUpdateTime: string;
  healthStatus: 'Healthy' | 'Warning' | 'Critical';
}
```

#### 1.3 集群详情页
**文件**: `cluster-detail.component.ts`

**功能描述**:
- 详细展示集群配置和状态
- 提供节点级别的详细信息
- 集成监控指标和日志查看

**页面结构**:
```
├── 概览 Tab
│   ├── 基本信息卡片
│   ├── 拓扑图展示
│   └── 关键指标概览
├── 节点 Tab  
│   ├── CN节点列表
│   ├── DN节点列表
│   └── GMS节点列表
├── 监控 Tab
│   ├── 性能指标图表
│   ├── 资源使用情况
│   └── 告警信息
├── 配置 Tab
│   ├── 集群配置展示
│   ├── 参数配置管理
│   └── 配置历史
└── 日志 Tab
    ├── 实时日志流
    ├── 日志搜索过滤
    └── 日志下载导出
```

#### 1.4 集群扩缩容
**文件**: `cluster-scaling-dialog.component.ts`

**功能描述**:
- 在线调整集群节点数量
- 提供扩容建议和资源预估
- 支持滚动更新策略配置

### 2. 日志采集管理模块

#### 2.1 日志采集配置
**文件**: `log-collection-config.component.ts`

**功能描述**:
- 配置CN/DN节点的日志采集开关
- 设置日志级别和采集规则
- 管理FileBeat和LogStash组件

**配置项**:
```typescript
interface LogCollectionConfig {
  cluster: string;
  namespace: string;
  
  // CN节点日志配置
  cnLogs: {
    enableAuditLog: boolean;
    enableSlowLog: boolean;
    enableErrorLog: boolean;
    logLevel: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';
  };
  
  // DN节点日志配置  
  dnLogs: {
    enableAuditLog: boolean;
    enableSlowLog: boolean;
    enableErrorLog: boolean;
    slowQueryThreshold: number; // 慢查询阈值(秒)
  };
  
  // 采集组件配置
  collectors: {
    fileBeatName: string;
    logStashName: string;
    namespace: string;
  };
}
```

#### 2.2 日志输出配置
**文件**: `log-output-config.component.ts`

**功能描述**:
- 配置日志输出目标(ElasticSearch/Kafka等)
- 编辑LogStash Pipeline配置
- 管理输出插件和格式化规则

**支持的输出类型**:
- ElasticSearch
- Kafka  
- File
- Stdout
- Custom (自定义输出插件)

#### 2.3 日志查看器
**文件**: `log-viewer.component.ts`

**功能描述**:
- 实时查看日志流
- 按组件/类型/级别过滤日志
- 支持关键词搜索和正则匹配

**查看模式**:
- 实时模式: 显示最新日志流
- 历史模式: 查询指定时间范围日志
- 搜索模式: 全文检索日志内容

### 3. 性能监控模块

#### 3.1 监控仪表板
**文件**: `performance-dashboard.component.ts`

**功能描述**:
- 展示集群整体性能指标
- 提供多维度监控视图
- 支持自定义时间范围和指标

**监控指标**:
```typescript
interface MonitoringMetrics {
  // 系统资源指标
  system: {
    cpu: { usage: number; cores: number };
    memory: { usage: number; total: number };
    disk: { usage: number; total: number; iops: number };
    network: { inbound: number; outbound: number };
  };
  
  // 数据库业务指标
  database: {
    qps: number;            // 每秒查询数
    tps: number;            // 每秒事务数
    connections: number;    // 活跃连接数
    slowQueries: number;    // 慢查询数量
    errors: number;         // 错误数量
  };
  
  // 集群健康指标
  cluster: {
    nodeStatus: Record<string, 'Ready' | 'NotReady' | 'Unknown'>;
    serviceStatus: Record<string, 'Running' | 'Pending' | 'Failed'>;
    replicationLag: number; // 主从复制延迟
  };
}
```

#### 3.2 告警管理
**文件**: `alert-management.component.ts`

**功能描述**:
- 配置监控告警规则
- 管理告警通知渠道
- 查看告警历史和处理状态

### 4. 系统诊断模块

#### 4.1 集群诊断工具
**文件**: `cluster-diagnostics.component.ts`

**功能描述**:
- 集成PolarDB-X内置诊断功能
- 提供一键健康检查
- 生成诊断报告和优化建议

#### 4.2 慢查询分析
**文件**: `slow-query-analyzer.component.ts`

**功能描述**:
- 分析慢查询日志
- 提供SQL优化建议
- 展示查询执行计划

## 🔧 技术实现细节

### 前端技术栈
- **框架**: Angular 17 + TypeScript
- **UI组件**: Angular Material 17
- **图表**: ECharts + ngx-echarts
- **编辑器**: Monaco Editor (配置文件编辑)
- **实时通信**: WebSocket / Server-Sent Events
- **状态管理**: RxJS + Services
- **样式**: SCSS + Angular Flex Layout

### 后端技术栈
- **语言**: Go 1.21+
- **框架**: Gin Web Framework
- **Kubernetes**: client-go + controller-runtime
- **监控**: Prometheus client + Grafana API
- **数据库**: 可选(SQLite/PostgreSQL 存储配置)
- **日志**: structured logging (logrus/zap)

### 关键API设计

#### 集群管理API
```go
// 集群创建
POST /api/v1/clusters
Content-Type: application/json
{
  "name": "my-cluster",
  "namespace": "default", 
  "spec": {
    "topology": {...},
    "storage": {...},
    "network": {...}
  }
}

// 集群列表
GET /api/v1/clusters?namespace=default

// 集群详情
GET /api/v1/clusters/{namespace}/{name}

// 集群扩缩容
PATCH /api/v1/clusters/{namespace}/{name}/scale
{
  "cn": {"replicas": 3},
  "dn": {"replicas": 5}
}

// 集群删除
DELETE /api/v1/clusters/{namespace}/{name}
```

#### 日志采集API
```go
// 日志采集配置
POST /api/v1/clusters/{namespace}/{name}/logs/config
{
  "cnLogs": {"enableAuditLog": true},
  "dnLogs": {"enableSlowLog": true}
}

// 日志输出配置
POST /api/v1/logcollectors/{namespace}/{name}/outputs
{
  "type": "elasticsearch",
  "config": {...}
}

// 实时日志流
GET /api/v1/clusters/{namespace}/{name}/logs/stream?component=cn&follow=true
```

#### 监控API
```go
// 监控指标查询
GET /api/v1/clusters/{namespace}/{name}/metrics?range=1h&step=30s

// 告警规则配置
POST /api/v1/clusters/{namespace}/{name}/alerts
{
  "rules": [...]
}
```

### 数据流设计

#### 集群状态同步
```
Kubernetes CRD Events → Controller → WebSocket → Frontend
                     ↓
                 Database Cache
```

#### 监控数据流
```
Prometheus Metrics → Backend Aggregation → REST API → Frontend Charts
                  ↓
              WebSocket Real-time Updates
```

#### 日志数据流  
```
PolarDB-X Logs → FileBeat → LogStash → Output (ES/Kafka/File)
                                    ↓
                              Backend Log API ← Frontend Log Viewer
```

## 📅 实施计划

### 第一阶段 (2-3周): 集群管理核心功能

**周1-2: 集群创建和管理**
- [ ] 实现集群创建向导组件
- [ ] 实现集群管理列表页面  
- [ ] 完善后端集群管理API
- [ ] 集成PolarDBXCluster CRD操作

**周3: 集群详情和操作**
- [ ] 实现集群详情页面
- [ ] 实现集群扩缩容功能
- [ ] 添加集群删除确认机制
- [ ] 完善状态监控和事件展示

### 第二阶段 (1-2周): 日志采集实现

**周1: 日志配置管理**  
- [ ] 实现日志采集配置界面
- [ ] 实现日志输出配置功能
- [ ] 完善LogCollector CRD操作
- [ ] 集成ConfigMap配置管理

**周2: 日志查看功能**
- [ ] 实现实时日志查看器
- [ ] 添加日志过滤和搜索
- [ ] 集成LogStash输出API
- [ ] 添加日志下载导出

### 第三阶段 (2-3周): 监控诊断功能

**周1-2: 性能监控**
- [ ] 实现监控仪表板界面
- [ ] 集成Prometheus指标查询
- [ ] 实现告警规则配置
- [ ] 添加自定义监控面板

**周3: 诊断工具**
- [ ] 实现集群健康检查
- [ ] 添加慢查询分析功能  
- [ ] 集成诊断报告生成
- [ ] 完善优化建议系统

### 第四阶段 (1周): 优化和文档

- [ ] 性能优化和bug修复
- [ ] 完善单元测试覆盖
- [ ] 编写用户使用文档
- [ ] 准备产品演示和部署

## 🧪 测试策略

### 单元测试
- 前端组件测试 (Jasmine + Karma)
- 后端API测试 (Go testing + testify)
- CRD操作测试 (envtest)

### 集成测试  
- End-to-End测试 (Playwright)
- API集成测试 (Postman/Newman)
- Kubernetes集群测试 (kind/minikube)

### 性能测试
- 前端性能测试 (Lighthouse)
- 后端压力测试 (vegeta/ab)
- 监控数据准确性验证

## 📚 文档体系

### 开发文档
- API接口文档 (Swagger/OpenAPI)
- 组件设计文档
- 数据库模型文档
- 部署配置文档

### 用户文档
- 快速开始指南
- 功能使用手册
- 常见问题解答
- 故障排查指南

## 🚀 部署和运维

### 部署方式
- Docker容器化部署
- Kubernetes YAML部署
- Helm Chart部署
- 开发环境本地部署

### 监控运维
- 应用性能监控(APM)
- 日志集中收集
- 健康检查和自愈
- 版本升级策略

---

## 下一步行动

1. **确认技术选型和架构设计**
2. **搭建开发环境和CI/CD流水线**  
3. **开始第一阶段:集群管理功能实现**

请确认此实现文档是否符合预期，我将根据此文档开始具体的代码实现。