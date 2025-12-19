# 用户使用指南 - 从零开始部署

## 📋 场景说明

假设你：
- ✅ 已下载本仓库
- ✅ 已启动 Kubernetes 集群（minikube/kind/云集群等）
- ✅ 有 kubeconfig 文件（`~/.kube/config` 或指定路径）
- ✅ 已安装 Docker、kubectl

## 🚀 快速开始（3 步）

### 步骤 1：构建镜像

```bash
# 在项目根目录执行
docker build -f tools/ui-all-in-one.Dockerfile -t polardbx-ui-all-in-one:latest .
```

**如果遇到镜像拉取问题**（网络问题）：
```bash
# 方案 1：使用镜像加速器（推荐）
# 配置 Docker 镜像加速器后重试

# 方案 2：手动拉取基础镜像
docker pull node:18-alpine
docker pull golang:1.21-alpine
docker pull alpine:latest

# 然后重新构建
docker build -f tools/ui-all-in-one.Dockerfile -t polardbx-ui-all-in-one:latest .
```

### 步骤 2：运行自动化部署脚本

```bash
# 使用默认 kubeconfig
./tools/deploy-all-in-one.sh

# 或指定 kubeconfig 路径
./tools/deploy-all-in-one.sh --kubeconfig /path/to/kubeconfig
```

脚本会自动完成：
- ✅ 检查环境（docker、kubectl、kubeconfig）
- ✅ 创建命名空间 `polardbx-system`
- ✅ 创建 ConfigMap（配置）
- ✅ 创建 Secret（kubeconfig）
- ✅ 部署 Deployment 和 Service
- ✅ 等待 Pod 就绪
- ✅ 显示访问方式

### 步骤 3：访问应用

```bash
# 端口转发（在另一个终端执行）
kubectl port-forward -n polardbx-system svc/polardbx-ui-backend 8080:8080

# 然后打开浏览器访问
# http://localhost:8080
```

## 📝 详细步骤（手动部署）

如果自动化脚本不适用，可以手动执行：

### 1. 准备 kubeconfig

```bash
# 检查 kubeconfig 是否有效
kubectl cluster-info

# 如果使用自定义路径
export KUBECONFIG=/path/to/kubeconfig
kubectl cluster-info
```

### 2. 创建命名空间

```bash
kubectl create namespace polardbx-system
```

### 3. 创建 ConfigMap

```bash
kubectl create configmap polardbx-ui-backend-config \
  --from-literal=LOG_LEVEL=info \
  --from-literal=LISTEN_ADDRESS=:8080 \
  --from-literal=UI_STATIC_DIR=/app/ui \
  --from-literal=KUBE_MODE=kubeconfig \
  --from-literal=KUBECONFIG_PATH=/etc/kube/kubeconfig \
  -n polardbx-system
```

### 4. 创建 Secret（包含 kubeconfig）

```bash
kubectl create secret generic polardbx-ui-backend-secret \
  --from-literal=kubeconfig="$(cat ~/.kube/config)" \
  -n polardbx-system
```

### 5. 部署应用

```bash
# 使用提供的 manifest 文件
kubectl apply -f tools/k8s-manifests/ -n polardbx-system

# 或使用 Helm
helm install polardbx-ui tools/helm-example \
  --set backend.image.repository=polardbx-ui-all-in-one \
  --set backend.image.tag=latest \
  --set backend.secrets.kubeconfig="$(cat ~/.kube/config | base64 -w 0)" \
  --namespace polardbx-system
```

### 6. 检查状态

```bash
# 查看 Pod
kubectl get pods -n polardbx-system

# 查看日志
kubectl logs -n polardbx-system -l app=polardbx-ui-backend

# 查看 Service
kubectl get svc -n polardbx-system
```

### 7. 访问应用

```bash
# 方式 1：端口转发
kubectl port-forward -n polardbx-system svc/polardbx-ui-backend 8080:8080

# 方式 2：NodePort（修改 service.yaml 中 type: NodePort）
kubectl get svc -n polardbx-system polardbx-ui-backend

# 方式 3：LoadBalancer（云环境）
kubectl get svc -n polardbx-system polardbx-ui-backend
```

## 🔍 验证部署

### 健康检查

```bash
# 通过端口转发测试
kubectl port-forward -n polardbx-system svc/polardbx-ui-backend 8080:8080 &

# 测试 API
curl http://localhost:8080/api/v1/health

# 测试前端
curl http://localhost:8080/ | head -20
```

### 功能验证

1. **打开浏览器**：访问 `http://localhost:8080`
2. **连接集群**：在 UI 中使用 kubeconfig 连接
3. **查看节点**：进入「节点」页面，应该能看到集群中的 Pod
4. **测试终端**：点击某个 Pod 的「终端」按钮，应该能打开 WebShell

## 🐛 常见问题

### 问题 1：镜像构建失败（网络问题）

**症状**：`failed to resolve source metadata`

**解决方案**：
```bash
# 配置 Docker 镜像加速器
# 编辑 /etc/docker/daemon.json
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://registry.docker-cn.com"
  ]
}

# 重启 Docker
sudo systemctl restart docker

# 重新构建
docker build -f tools/ui-all-in-one.Dockerfile -t polardbx-ui-all-in-one:latest .
```

### 问题 2：Pod 无法启动

**检查步骤**：
```bash
# 查看 Pod 状态
kubectl describe pod -n polardbx-system -l app=polardbx-ui-backend

# 查看日志
kubectl logs -n polardbx-system -l app=polardbx-ui-backend

# 检查镜像是否存在
docker images | grep polardbx-ui-all-in-one
```

**可能原因**：
- 镜像未构建或未推送到集群可访问的仓库
- 如果是远程集群，需要将镜像推送到镜像仓库

### 问题 3：无法连接 K8s API（504 超时）

**症状**：前端连接时出现 504 Gateway Timeout，日志显示 "apiserver connection timeout"

**快速检查**：
```bash
# 检查 kubeconfig 是否正确
kubectl exec -n polardbx-system <pod-name> -- cat /etc/kube/kubeconfig

# 检查环境变量
kubectl exec -n polardbx-system <pod-name> -- env | grep KUBE

# 测试 kubeconfig
kubectl --kubeconfig=/etc/kube/kubeconfig cluster-info
```

**详细排查**：请参考 [NETWORK-TROUBLESHOOTING.md](./NETWORK-TROUBLESHOOTING.md)

**常见原因**：
- 容器在 Docker 中运行，无法访问 kubeconfig 中的 API Server 地址（如 minikube 的 `192.168.49.2:8443`）
- 网络不可达或防火墙阻止
- kubeconfig 中的 server 地址不正确

**解决方案**：
1. **本地开发**：使用 `--network host` 运行容器，或修改 kubeconfig 使用 `host.docker.internal`
2. **生产环境**：部署到 K8s 集群内，使用 in-cluster config（推荐）

### 问题 4：前端页面无法加载

**检查步骤**：
```bash
# 检查静态文件
kubectl exec -n polardbx-system <pod-name> -- ls -la /app/ui/

# 检查环境变量
kubectl exec -n polardbx-system <pod-name> -- env | grep UI_STATIC_DIR

# 查看后端日志
kubectl logs -n polardbx-system -l app=polardbx-ui-backend | grep "Serving static"
```

## 📦 远程集群部署

如果 Kubernetes 集群在远程（不在本地），需要：

### 1. 将镜像推送到镜像仓库

```bash
# 标记镜像
docker tag polardbx-ui-all-in-one:latest your-registry/polardbx-ui-all-in-one:latest

# 推送镜像
docker push your-registry/polardbx-ui-all-in-one:latest
```

### 2. 修改部署配置

```bash
# 修改 deployment.yaml 中的 image
# 或使用 Helm values
helm install polardbx-ui tools/helm-example \
  --set backend.image.repository=your-registry/polardbx-ui-all-in-one \
  --set backend.image.tag=latest \
  --namespace polardbx-system
```

## 🧹 清理

```bash
# 删除部署
kubectl delete -f tools/k8s-manifests/ -n polardbx-system

# 删除命名空间（会删除所有资源）
kubectl delete namespace polardbx-system

# 删除本地镜像
docker rmi polardbx-ui-all-in-one:latest
```

## 📚 更多资源

- [README-ALL-IN-ONE.md](./README-ALL-IN-ONE.md) - 详细技术文档
- [QUICK-START.md](./QUICK-START.md) - 快速开始指南
- [IMPLEMENTATION-SUMMARY.md](./IMPLEMENTATION-SUMMARY.md) - 实施总结
- [NETWORK-TROUBLESHOOTING.md](./NETWORK-TROUBLESHOOTING.md) - 网络连接问题排查指南

## 💡 提示

1. **开发环境**：可以使用 `minikube` 或 `kind` 快速搭建本地集群
2. **生产环境**：建议使用 Helm Chart 部署，便于配置管理和升级
3. **监控**：部署后可以配置 Prometheus/Grafana 监控
4. **安全**：生产环境建议使用 RBAC、网络策略等安全措施

