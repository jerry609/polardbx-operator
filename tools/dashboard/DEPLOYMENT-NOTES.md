# Dashboard 部署说明

## 本地开发部署（Docker）

### 使用 host 网络模式

对于本地开发环境（如 minikube），需要：

1. **挂载 kubeconfig 文件**
2. **挂载证书目录**（如果 kubeconfig 引用外部证书文件）

```bash
docker run -d --name polardbx-dashboard-test \
  --network host \
  -v ${HOME}/.kube/config:/etc/kube/kubeconfig:ro \
  -v ${HOME}/.minikube:${HOME}/.minikube:ro \
  -e UI_STATIC_DIR=/app/ui \
  -e KUBECONFIG=/etc/kube/kubeconfig \
  polardbx-dashboard:test
```

### 为什么需要挂载 .minikube 目录？

minikube 的 kubeconfig 通常引用外部证书文件：
- `certificate-authority: /home/user/.minikube/ca.crt`
- `client-certificate: /home/user/.minikube/profiles/minikube/client.crt`
- `client-key: /home/user/.minikube/profiles/minikube/client.key`

这些路径在容器内不存在，因此需要挂载整个 `.minikube` 目录。

## Kubernetes 部署

### 推荐方式：使用 ServiceAccount

在 Kubernetes 集群内运行时，推荐使用 ServiceAccount 和 in-cluster config：

```yaml
spec:
  serviceAccountName: polardbx-dashboard-backend
  containers:
    - name: backend
      env:
        - name: KUBECONFIG
          value: ""  # 空值表示使用 in-cluster config
```

### 使用 kubeconfig Secret

如果必须使用 kubeconfig：

1. **将证书内容内联到 kubeconfig**（推荐）
   - 使用 `certificate-authority-data` 而不是 `certificate-authority`
   - 使用 `client-certificate-data` 而不是 `client-certificate`
   - 使用 `client-key-data` 而不是 `client-key`

2. **或者挂载证书文件**
   ```yaml
   volumeMounts:
     - name: kubeconfig
       mountPath: /etc/kube
     - name: minikube-certs
       mountPath: /home/user/.minikube
   volumes:
     - name: kubeconfig
       secret:
         secretName: polardbx-ui-backend-secret
     - name: minikube-certs
       secret:
         secretName: minikube-certs  # 需要单独创建
   ```

## 环境变量

- `KUBECONFIG`: kubeconfig 文件路径（空值表示使用 in-cluster config）
- `UI_STATIC_DIR`: 前端静态文件目录（默认: `/app/ui`）
- `LISTEN_ADDRESS`: 监听地址（默认: `:8080`）

## 网络配置

### 本地开发
- 使用 `--network host` 模式访问主机的 Kubernetes API Server

### Kubernetes 部署
- 使用 ClusterIP Service 在集群内访问
- Pod 可以直接访问 API Server（使用 in-cluster config）

