# PolarDB-X Management Platform - All-in-One Deployment

本文档说明如何构建和部署整合了前后端的 all-in-one 镜像，以及如何在 Kubernetes 中使用 Helm Chart 进行配置管理。

## 架构说明

### All-in-One 镜像

- **前端**：Angular 应用打包后的静态文件（`polardbx-ui/dist/`）
- **后端**：Go HTTP 服务（`backend/main.go`）
- **统一服务**：后端同时提供：
  - `/api/v1/*` - REST API
  - `/` - 静态前端页面（Angular history mode）

### 配置管理（K8s 原生）

- **ConfigMap**：非敏感配置（日志级别、端口、路径等）
- **Secret**：敏感数据（kubeconfig、访问密钥等）
- **环境变量**：通过 `envFrom` 自动注入到 Pod

## 快速开始

### 1. 构建 All-in-One 镜像

```bash
# 在项目根目录执行
docker build -f tools/ui-all-in-one.Dockerfile -t polardbx-ui-all-in-one:v1.0.0 .
```

### 2. 本地测试

```bash
# 运行测试脚本（自动构建、启动、测试）
./tools/test-all-in-one.sh

# 或手动运行
docker run -d \
  --name polardbx-ui-test \
  -p 8080:8080 \
  -e UI_STATIC_DIR=/app/ui \
  polardbx-ui-all-in-one:v1.0.0

# 访问 http://localhost:8080
```

### 3. Kubernetes 部署（Helm）

```bash
# 安装 Helm Chart
helm install polardbx-ui tools/helm-example \
  --set backend.image.repository=polardbx-ui-all-in-one \
  --set backend.image.tag=v1.0.0 \
  --set backend.secrets.kubeconfig="<base64-encoded-kubeconfig>" \
  --namespace polardbx-system \
  --create-namespace

# 查看状态
kubectl get pods -n polardbx-system
kubectl get svc -n polardbx-system

# 访问（通过 Service 或 Ingress）
kubectl port-forward -n polardbx-system svc/polardbx-management-platform-backend 8080:8080
# 然后访问 http://localhost:8080
```

## 配置说明

### 环境变量（ConfigMap）

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `LOG_LEVEL` | 日志级别 | `info` |
| `LISTEN_ADDRESS` | 监听地址 | `:8080` |
| `UI_STATIC_DIR` | 前端静态文件目录 | `/app/ui` |
| `KUBE_MODE` | K8s 连接模式 | `incluster` |
| `KUBECONFIG_PATH` | kubeconfig 文件路径 | `/etc/kube/kubeconfig` |

### 敏感配置（Secret）

| 变量名 | 说明 |
|--------|------|
| `kubeconfig` | 外部集群的 kubeconfig（base64 编码） |
| `HPFS_ACCESS_KEY` | HPFS 访问密钥 |
| `HPFS_SECRET_KEY` | HPFS 密钥 |

### 修改配置

```bash
# 编辑 values.yaml
vim tools/helm-example/values.yaml

# 升级部署
helm upgrade polardbx-ui tools/helm-example \
  --namespace polardbx-system \
  -f tools/helm-example/values.yaml
```

## 开发模式 vs 生产模式

### 开发模式（本地）

- 前端：`ng serve` (端口 4200)
- 后端：`go run backend/main.go` (端口 8080)
- 代理：`proxy.conf.js` 将 `/api` 代理到后端

### 生产模式（All-in-One）

- 前端：静态文件打包到镜像
- 后端：同一个服务提供 API 和静态文件
- 访问：统一通过 `http://<host>/` 和 `http://<host>/api/v1/*`

## 故障排查

### 镜像构建失败

```bash
# 检查 Dockerfile 路径
docker build -f tools/ui-all-in-one.Dockerfile -t test .

# 查看构建日志
docker build -f tools/ui-all-in-one.Dockerfile -t test . 2>&1 | tee build.log
```

### 容器无法启动

```bash
# 查看容器日志
docker logs polardbx-ui-test

# 检查环境变量
docker exec polardbx-ui-test env | grep UI_STATIC_DIR
```

### K8s 部署问题

```bash
# 检查 Pod 状态
kubectl describe pod -n polardbx-system -l app.kubernetes.io/component=backend

# 查看 ConfigMap/Secret
kubectl get configmap -n polardbx-system
kubectl get secret -n polardbx-system

# 检查环境变量注入
kubectl exec -n polardbx-system <pod-name> -- env | grep UI_STATIC_DIR
```

## 下一步

1. **集成到现有 Helm Chart**：将 `tools/helm-example` 的内容合并到 `charts/polardbx-operator`
2. **CI/CD 集成**：在构建流程中自动构建 all-in-one 镜像
3. **多环境配置**：为 dev/test/prod 准备不同的 values 文件

