# 网络连接问题排查指南

## 问题描述

当容器无法连接到 Kubernetes API Server 时，会出现 504 超时错误。这通常发生在以下情况：

1. **容器在 Docker 中运行**（不在 K8s 集群内），无法访问 kubeconfig 中的 API Server 地址
2. **网络不可达**：容器网络无法访问 API Server 的 IP 地址
3. **防火墙/安全组限制**：网络策略阻止了连接

## 诊断步骤

### 1. 检查容器内的网络连接

```bash
# 进入容器
docker exec -it <container-name> sh

# 检查 API Server 地址（从 kubeconfig 中提取）
# 例如：https://192.168.49.2:8443

# 测试连接
wget -O- --no-check-certificate https://<api-server-ip>:<port>/version
# 或
curl -k https://<api-server-ip>:<port>/version
```

### 2. 检查 kubeconfig 中的 API Server 地址

```bash
# 查看 kubeconfig 中的 server 地址
kubectl config view | grep server

# 检查该地址是否可以从容器访问
docker exec <container-name> ping <api-server-ip>
```

### 3. 检查容器日志

```bash
docker logs <container-name> 2>&1 | grep -i "timeout\|unreachable\|connection"
```

## 解决方案

### 方案 1: 修复 Docker 容器网络（适用于本地开发）

#### Minikube

如果使用 minikube，API Server 地址通常是 `https://192.168.49.2:8443`（或其他 minikube IP）。

**选项 A: 使用 host 网络模式**

```bash
docker run --network host polardbx-ui-all-in-one:latest
```

**选项 B: 使用 Docker 的 host.docker.internal（macOS/Windows）**

修改 kubeconfig，将 API Server 地址改为：
- macOS/Windows: `https://host.docker.internal:8443`
- Linux: 需要手动设置，或使用 `--add-host` 参数

```bash
docker run --add-host=host.docker.internal:host-gateway polardbx-ui-all-in-one:latest
```

**选项 C: 使用 minikube 的 IP 地址（需要确保容器能访问）**

```bash
# 获取 minikube IP
minikube ip

# 确保容器能访问该 IP（可能需要配置 Docker 网络）
docker run --network bridge polardbx-ui-all-in-one:latest
```

#### Kind

如果使用 kind，API Server 地址通常是 `https://127.0.0.1:<random-port>`。

**解决方案：使用 host 网络模式**

```bash
docker run --network host polardbx-ui-all-in-one:latest
```

### 方案 2: 部署到 Kubernetes 集群内（推荐生产环境）

当容器部署在 K8s 集群内时，应该使用 **in-cluster config**，这样会自动使用集群内的服务发现。

#### 使用 Helm Chart 部署

```bash
# 部署到集群内，使用 in-cluster config
helm install polardbx-ui ./tools/helm-example \
  --set kubeconfig.useInCluster=true
```

#### 使用原生 YAML 部署

在 Deployment 中设置环境变量：

```yaml
env:
  - name: KUBECONFIG
    value: ""  # 空值表示使用 in-cluster config
```

### 方案 3: 修改 kubeconfig 中的 API Server 地址

如果容器在 Docker 中运行，但需要访问远程 K8s 集群：

1. **获取容器可访问的 API Server 地址**
   - 如果是本地集群，使用 `host.docker.internal` 或主机 IP
   - 如果是远程集群，确保网络可达

2. **修改 kubeconfig**

```bash
# 备份原 kubeconfig
cp ~/.kube/config ~/.kube/config.backup

# 修改 server 地址（使用适合你环境的地址）
kubectl config set-cluster <cluster-name> --server=https://<accessible-ip>:<port>
```

3. **重新部署容器，使用新的 kubeconfig**

```bash
# 创建新的 Secret
kubectl create secret generic polardbx-ui-kubeconfig \
  --from-file=kubeconfig=<path-to-modified-kubeconfig>

# 或使用 Helm
helm upgrade polardbx-ui ./tools/helm-example \
  --set kubeconfig.value="$(cat <path-to-modified-kubeconfig> | base64 -w 0)"
```

## 验证修复

### 1. 检查连接端点

```bash
# 测试 /api/v1/connect 端点
curl -X POST http://localhost:8080/api/v1/connect \
  -H "Content-Type: application/json" \
  -H "X-Kubeconfig-B64: $(cat ~/.kube/config | base64 -w 0)" \
  -v
```

### 2. 检查容器日志

```bash
docker logs <container-name> 2>&1 | grep -i "connect\|apiserver"
```

应该看到：
```
Connect: apiserver version query succeeded
Connect: connection successful
```

而不是：
```
Connect: apiserver connection timeout
```

### 3. 测试前端连接

1. 打开浏览器访问 `http://localhost:8080`
2. 在连接页面输入 kubeconfig
3. 点击连接，应该成功连接而不是超时

## 常见问题

### Q: 为什么容器内无法访问 `192.168.49.2`？

A: Docker 容器的网络是隔离的，默认情况下无法直接访问主机的私有网络。需要使用 `--network host` 或配置 Docker 网络。

### Q: 生产环境应该使用什么方案？

A: **推荐使用方案 2**：将容器部署到 K8s 集群内，使用 in-cluster config。这样可以：
- 自动使用集群内的服务发现
- 无需配置网络路由
- 更安全（使用 ServiceAccount 认证）

### Q: 如何判断应该使用哪种方案？

A:
- **本地开发/测试**：使用方案 1（Docker 网络配置）
- **生产环境**：使用方案 2（部署到集群内）
- **混合环境**：使用方案 3（修改 kubeconfig 地址）

## 相关文件

- `tools/deploy-all-in-one.sh` - 部署脚本
- `tools/helm-example/` - Helm Chart 配置
- `tools/USER-GUIDE.md` - 用户指南
- `backend/pkg/api/handlers.go` - Connect 端点实现（已添加 10 秒超时）

