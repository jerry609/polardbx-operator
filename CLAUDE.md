# CLAUDE.md - PolarDB-X Operator 开发指南

**项目**: PolarDB-X Operator  
**版本**: v1.7.0  
**更新日期**: 2025-02-02  
**架构**: Kubernetes Operator + Management Platform UI  

## 📋 项目概览

PolarDB-X Operator 是一个用于在 Kubernetes 集群中部署和管理 PolarDB-X 分布式数据库的 Operator。项目采用双重架构设计：

1. **Kubernetes Operator**: 基于 controller-runtime 的核心 Operator 组件
2. **Management Platform UI**: 基于 Angular 19 的 Web 管理界面 + Go Backend API

## 🏗️ 项目架构

### 目录结构
```
polardbx-operator/
├── api/                    # CRD 定义和类型
│   └── v1/                # v1 版本的 API 定义
├── cmd/                   # 各种可执行程序入口
│   ├── polardbx-operator/ # 主 Operator 程序
│   ├── polardbx-binlog/   # Binlog 处理工具
│   └── ...               # 其他工具程序
├── pkg/                   # 核心业务逻辑
│   ├── operator/         # Operator 控制器逻辑
│   ├── k8s/              # Kubernetes 客户端封装
│   └── ...              # 其他核心包
├── backend/              # Management Platform Backend
│   ├── main.go          # Backend API 服务入口
│   ├── pkg/api/         # API 处理器
│   └── pkg/k8s/         # K8s 客户端封装
├── polardbx-ui/         # Management Platform Frontend
│   ├── src/app/         # Angular 应用源码
│   └── package.json     # 前端依赖配置
├── charts/              # Helm 部署图表
├── test/                # 测试用例
└── Makefile            # 构建和开发工具
```

### 核心组件

#### 1. Kubernetes Operator
- **语言**: Go 1.21+
- **框架**: controller-runtime
- **功能**: 管理 PolarDB-X 集群生命周期
- **入口**: `cmd/polardbx-operator/main.go`

#### 2. Management Platform Backend  
- **语言**: Go 1.21+
- **框架**: Gin
- **功能**: 提供 REST API 管理界面
- **入口**: `backend/main.go`

#### 3. Management Platform Frontend
- **语言**: TypeScript
- **框架**: Angular 19 + Angular Material
- **功能**: Web UI 管理界面
- **入口**: `polardbx-ui/src/main.ts`

## 🚀 快速开发指南

### Environment Setup

**必需工具**:
```bash
# Go 开发环境
go version  # 需要 1.21+

# Node.js 前端环境  
node --version  # 需要 18+
npm --version

# Kubernetes 工具
kubectl version
helm version

# 构建工具
make --version
docker --version
```

### 构建命令

**Backend API 构建**:
```bash
# 进入 backend 目录
cd backend

# 本地开发运行
go run main.go

# 构建二进制
go build -o polardbx-ui-backend main.go

# 运行测试
go test ./...
```

**Frontend UI 构建**:
```bash
# 进入 UI 目录
cd polardbx-ui

# 安装依赖
npm install

# 本地开发服务器
npm run start

# 生产构建
npm run build

# 运行测试
npm run test

# 代码规范检查
npm run lint
```

**Operator 构建**:
```bash
# 项目根目录构建
make all

# 构建特定组件
make TARGET=cmd/polardbx-operator

# 构建 Docker 镜像
make build

# 运行单元测试
make unit-test
```

### 开发工作流

#### Backend API 开发
1. **添加新 API 端点**:
   - 在 `backend/pkg/api/handlers.go` 添加处理器函数
   - 在 `backend/main.go` 注册路由
   - 在 `backend/pkg/k8s/client.go` 添加 K8s 操作函数
   - 编写对应的测试用例

2. **API 设计规范**:
   ```go
   // RESTful API 设计
   GET    /api/v1/resources                    // 列表
   POST   /api/v1/resources                    // 创建
   GET    /api/v1/resources/:namespace/:name   // 详情
   PUT    /api/v1/resources/:namespace/:name   // 更新
   DELETE /api/v1/resources/:namespace/:name   // 删除
   ```

#### Frontend UI 开发
1. **添加新页面组件**:
   - 在 `src/app/components/` 创建组件
   - 在 `src/app/models/` 定义 TypeScript 模型
   - 在 `src/app/services/api.service.ts` 添加 API 调用
   - 更新路由配置

2. **UI 设计规范**:
   - 使用 Angular Material 组件
   - 遵循 Material Design 规范  
   - 支持响应式设计
   - 统一的错误处理和加载状态

#### CRD 开发
1. **定义新 CRD**:
   - 在 `api/v1/` 添加类型定义
   - 运行 `make controller-gen` 生成代码
   - 更新 RBAC 权限

2. **控制器开发**:
   - 在 `pkg/operator/` 添加控制器逻辑
   - 实现 Reconcile 方法
   - 添加对应的测试用例

## 🔧 核心 CRD 资源

项目定义了 14 个 CRD 资源，Management Platform 现已完整支持所有资源：

### ✅ 已完整支持 (14/14) - 🎉 100% 完成！
1. **PolarDBXCluster** - 集群管理
2. **PolarDBXBackup** - 备份管理  
3. **PolarDBXParameter** - 参数配置
4. **XStore** - 存储节点管理
5. **PolarDBXMonitor** - 监控配置
6. **PolarDBXBackupSchedule** - 定时备份
7. **PolarDBXParameterTemplate** - 参数模板
8. **SystemTask** - 系统任务
9. **PolarDBXLogCollector** - 日志收集
10. **PolarDBXBackupBinlog** - Binlog 备份
11. **XStoreFollower** - DN副本故障恢复
12. **XStoreBackup** - 存储级备份
13. **XStoreBackupBinlog** - 存储级日志备份
14. **PolarDBXClusterKnobs** - 集群性能调优 🆕

### 🆕 最新实现的关键功能 (v4.0 重大更新)

#### PolarDBXClusterKnobs - 集群性能调优 🎯
- **完整的性能调优管理** - 支持所有14个CRD资源的最后一块拼图
- **分类化参数配置** - 连接管理、内存优化、查询优化、日志配置四大分类
- **30+ 预定义调优参数** - 涵盖关键性能调优场景
- **智能参数验证** - 根据参数类型进行值验证和格式化
- **影响等级可视化** - 清晰标识参数影响范围（低/中/高）
- **重启要求标识** - 明确标识是否需要重启集群
- **自定义参数支持** - 支持添加自定义调优参数
- **Material Design UI** - 现代化的用户体验

#### Recovery Management - 恢复管理
- **RestoreCluster API** - 集群备份恢复
- **InitiatePITR API** - 时间点恢复(PITR)
- **GetRestoreStatus API** - 恢复状态查询
- **Recovery Wizard UI** - 分步式恢复向导界面

#### XStoreFollower Management - DN副本故障恢复  
- **XStoreFollower CRUD API** - DN 备库重搭
- **XStoreFollower Management UI** - 故障恢复管理界面

#### XStoreBackup Management - 统一备份模块
- **XStoreBackup CRUD API** - 存储级备份
- **XStoreBackup Management UI** - 备份策略管理界面

#### 完整路由组织化架构 🗂️
- **模块化导航结构** - 备份管理、恢复管理、存储管理、运维管理四大模块
- **Material Expansion Panel** - 分层级导航体验
- **懒加载路由** - 优化应用性能和资源加载
- **统一的用户体验** - 跨模块的一致性设计

## 🧪 测试指南

### 单元测试
```bash
# Backend 测试
cd backend && go test ./...

# 覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Frontend 测试
cd polardbx-ui && npm run test
```

### 集成测试  
```bash
# E2E 测试
make e2e-test

# 特定测试并行度
make e2e-test E2E_PARALLELISM=5
```

### API 测试
推荐使用 Postman 或 curl 测试 Backend API：

```bash
# 健康检查
curl http://localhost:8080/health

# 获取集群列表  
curl -H "Kubeconfig: <base64-encoded-kubeconfig>" \
     http://localhost:8080/api/v1/clusters

# 创建集群性能调优配置
curl -X POST -H "Content-Type: application/json" \
     -H "Kubeconfig: <base64-encoded-kubeconfig>" \
     -d '{
       "name": "performance-tuning",
       "namespace": "default",
       "clusterName": "my-cluster",
       "knobs": {
         "max_connections": 1000,
         "innodb_buffer_pool_size": "8G",
         "query_cache_size": "256M"
       }
     }' \
     http://localhost:8080/api/v1/cluster-knobs
```

## 🔍 故障排查

### 常见问题

#### 1. Backend API 启动失败
```bash
# 检查端口占用
lsof -i :8080

# 检查 Kubeconfig 配置
kubectl cluster-info

# 查看详细日志
./polardbx-ui-backend --v=2
```

#### 2. Frontend 构建失败
```bash
# 清理 node_modules
rm -rf node_modules package-lock.json
npm install

# 检查 TypeScript 编译
npx tsc --noEmit

# 检查 ESLint 规范
npm run lint
```

#### 3. Operator 控制器问题
```bash
# 查看 Operator 日志
kubectl logs -n polardbx-operator-system deployment/polardbx-operator-controller-manager

# 检查 CRD 状态
kubectl get crd | grep polardbx

# 查看集群状态
kubectl get polardbx -A
```

### 调试技巧

#### Backend 调试
```go
// 使用 Go 调试工具
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug main.go
```

#### Frontend 调试
- 使用 Chrome DevTools
- Angular DevTools 扩展
- 检查 Network 面板 API 调用
- Console 错误日志分析

## 📊 性能优化

### Backend 优化
- **连接池**: 合理配置 K8s 客户端连接池
- **缓存策略**: 实现适当的 API 响应缓存
- **并发控制**: 使用 Go routine 池控制并发度

### Frontend 优化  
- **懒加载**: 实现路由级别的代码分割
- **虚拟滚动**: 大列表使用 CDK Virtual Scrolling
- **状态管理**: 合理使用 RxJS 管理状态
- **Bundle 分析**: 定期分析打包体积

## 🚀 部署指南

### 本地开发部署
```bash
# 1. 启动 Backend API
cd backend && go run main.go

# 2. 启动 Frontend (新终端)
cd polardbx-ui && npm run start

# 3. 访问 UI
open http://localhost:4200
```

### 生产环境部署
```bash
# 使用 Helm 部署 Operator
helm install polardbx-operator charts/polardbx-operator

# 构建和部署 Management Platform
make build IMAGE_SOURCE=your-registry.com
kubectl apply -f manifests/
```

## 📋 代码规范

### Go 代码规范
- 遵循 `gofmt` 标准格式
- 使用 `golint` 检查代码质量
- 函数注释遵循 Go Doc 规范
- 错误处理要完整和一致

### TypeScript 代码规范
- 使用 ESLint + Prettier
- 严格的 TypeScript 类型检查
- 遵循 Angular 官方风格指南
- 组件和服务的命名约定

### 提交信息规范
```
feat: add recovery wizard component
fix: resolve cluster status display issue  
docs: update API documentation
test: add unit tests for restore handlers
```

## 🔗 相关资源

- **官方文档**: https://doc.polardbx.com/quickstart/topics/quickstart-k8s.html
- **项目文档**：/Users/jerry/polardbx-operator/PolarDB-X_Management_Platform_Completeness_Analysis_Report.md
- **API 参考**: `/api/v1/` 目录下的类型定义
- **监控配置**: Prometheus + Grafana 集成
- **日志收集**: EFK Stack 支持

## 🤝 贡献指南

1. Fork 项目仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)  
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

确保所有测试通过，并遵循代码规范，并更新项目文档：/Users/jerry/polardbx-operator/PolarDB-X_Management_Platform_Completeness_Analysis_Report.md。

---

**最后更新**: 2025-02-02  
**文档版本**: v4.0  
**状态**: ✅ 已完成 - 系统功能完整性达到 100%，生产就绪

## 🎯 完整性里程碑

### 📊 功能完成度统计
- **CRD 支持率**: 14/14 (100%) ✅
- **UI 模块化**: 4 大业务模块完整实现 ✅
- **API 完整性**: 所有 CRUD 操作完整支持 ✅
- **路由组织**: 现代化分层导航架构 ✅

### 🏆 核心成就
1. **完整的CRD生态系统** - 14个CRD资源全部支持，无缺失
2. **企业级用户体验** - Material Design + 模块化架构
3. **性能调优能力** - 专业的集群性能优化管理
4. **生产就绪状态** - 完整的错误处理、测试覆盖、文档

### 🚀 技术栈完整性
- **Frontend**: Angular 19 + Material Design + TypeScript
- **Backend**: Go 1.21+ + Gin + Kubernetes Controller Runtime  
- **Infrastructure**: Kubernetes + Helm + Prometheus + Grafana
- **Testing**: Unit Tests + E2E Tests + API Tests

PolarDB-X Management Platform 现已成为功能完整、生产就绪的企业级数据库管理平台！