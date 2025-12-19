# All-in-One 部署实施总结

## ✅ 已完成的工作

### 1. 后端路由改造 (`backend/pkg/api/router/router.go`)

- ✅ 添加静态文件服务支持
- ✅ 当 `UI_STATIC_DIR` 环境变量设置时，自动服务前端静态文件
- ✅ 支持 Angular history mode（未匹配路径返回 `index.html`）
- ✅ API 路由保持 `/api/v1/*` 不变

**关键代码：**
```go
if uiDir := strings.TrimSpace(os.Getenv("UI_STATIC_DIR")); uiDir != "" {
    r.StaticFS("/", http.Dir(uiDir))
    r.NoRoute(func(c *gin.Context) {
        // 非 API 路径返回 index.html
    })
}
```

### 2. All-in-One Dockerfile (`tools/ui-all-in-one.Dockerfile`)

- ✅ 多阶段构建：前端构建 + 后端构建 + 最终镜像
- ✅ 前端：Node.js 构建 Angular 应用
- ✅ 后端：Go 构建静态二进制
- ✅ 最终镜像：Alpine + 二进制 + 静态文件

**构建命令：**
```bash
docker build -f tools/ui-all-in-one.Dockerfile -t polardbx-ui-all-in-one:v1.0.0 .
```

### 3. Helm Chart 示例 (`tools/helm-example/`)

- ✅ `values.yaml` - 配置模板
- ✅ `templates/configmap.yaml` - 非敏感配置
- ✅ `templates/secret.yaml` - 敏感数据
- ✅ `templates/deployment.yaml` - 部署配置（envFrom 注入）
- ✅ `templates/service.yaml` - 服务暴露
- ✅ `templates/_helpers.tpl` - Helm 辅助函数

**部署命令：**
```bash
helm install polardbx-ui tools/helm-example \
  --set backend.image.repository=polardbx-ui-all-in-one \
  --set backend.image.tag=v1.0.0 \
  --namespace polardbx-system
```

### 4. 测试脚本 (`tools/test-all-in-one.sh`)

- ✅ 自动构建镜像
- ✅ 启动容器
- ✅ 测试健康检查端点
- ✅ 测试前端页面
- ✅ 测试 API 端点

**使用方法：**
```bash
./tools/test-all-in-one.sh
```

## 📋 配置说明

### 环境变量（ConfigMap）

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `UI_STATIC_DIR` | 前端静态文件目录 | `/app/ui` |
| `LOG_LEVEL` | 日志级别 | `info` |
| `LISTEN_ADDRESS` | 监听地址 | `:8080` |
| `KUBE_MODE` | K8s 连接模式 | `incluster` |

### 敏感配置（Secret）

- `kubeconfig` - 外部集群配置（可选）
- `HPFS_ACCESS_KEY` - HPFS 访问密钥
- `HPFS_SECRET_KEY` - HPFS 密钥

## 🚀 使用流程

### 开发环境

1. **前端开发**：`cd polardbx-ui && npm start` (端口 4200)
2. **后端开发**：`cd backend && go run main.go` (端口 8080)
3. **代理配置**：`proxy.conf.js` 自动代理 `/api` 到后端

### 生产环境（All-in-One）

1. **构建镜像**：
   ```bash
   docker build -f tools/ui-all-in-one.Dockerfile -t polardbx-ui-all-in-one:v1.0.0 .
   ```

2. **本地测试**：
   ```bash
   docker run -d -p 8080:8080 -e UI_STATIC_DIR=/app/ui polardbx-ui-all-in-one:v1.0.0
   # 访问 http://localhost:8080
   ```

3. **K8s 部署**：
   ```bash
   # 修改 values.yaml 中的镜像地址
   helm install polardbx-ui tools/helm-example -n polardbx-system
   ```

## 🔍 验证清单

- [ ] 镜像构建成功
- [ ] 容器启动正常
- [ ] `/api/v1/health` 返回 200
- [ ] `/` 返回前端页面
- [ ] `/api/v1/*` API 正常工作
- [ ] ConfigMap/Secret 正确注入
- [ ] K8s 部署成功

## 📝 下一步建议

1. **集成到现有 Chart**：将 `tools/helm-example` 合并到 `charts/polardbx-operator`
2. **CI/CD 集成**：在构建流程中自动构建 all-in-one 镜像
3. **多环境配置**：准备 dev/test/prod 的 values 文件
4. **文档完善**：更新主 README，添加部署章节

## 🐛 已知问题

- Docker 镜像拉取可能因网络问题失败，建议使用镜像加速或离线构建
- 如果使用 `scratch` 基础镜像（如原 backend Dockerfile），需要确保静态文件服务不依赖系统库

## 📚 相关文件

- `tools/ui-all-in-one.Dockerfile` - All-in-one 镜像构建
- `tools/helm-example/` - Helm Chart 示例
- `tools/test-all-in-one.sh` - 测试脚本
- `tools/README-ALL-IN-ONE.md` - 详细文档
- `backend/pkg/api/router/router.go` - 路由配置

