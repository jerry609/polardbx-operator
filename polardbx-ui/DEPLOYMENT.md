# 部署指南

本文档提供了 PolarDB-X UI 项目在不同环境中的部署方案。

## 目录

- [环境要求](#环境要求)
- [本地部署](#本地部署)
- [Docker 部署](#docker-部署)
- [Kubernetes 部署](#kubernetes-部署)
- [生产环境部署](#生产环境部署)
- [监控和日志](#监控和日志)
- [故障排除](#故障排除)

## 环境要求

### 最低要求

- **CPU**: 2 核心
- **内存**: 4GB RAM
- **存储**: 10GB 可用空间
- **网络**: 稳定的网络连接

### 推荐配置

- **CPU**: 4 核心或更多
- **内存**: 8GB RAM 或更多
- **存储**: 50GB SSD
- **网络**: 千兆网络

### 软件依赖

- **Docker**: 20.10+ (用于容器化部署)
- **Kubernetes**: 1.20+ (用于 K8s 部署)
- **kubectl**: 与 Kubernetes 版本匹配
- **Helm**: 3.0+ (可选，用于 Helm 部署)

## 本地部署

### 开发环境部署

适用于开发和测试环境。

#### 1. 准备环境

```bash
# 克隆项目
git clone <repository-url>
cd polardbx-operator/polardbx-ui

# 安装前端依赖
npm install

# 安装后端依赖
cd ../backend
go mod tidy
```

#### 2. 启动服务

```bash
# 启动后端服务
cd backend
go run main.go

# 在新终端启动前端服务
cd polardbx-ui
ng serve
```

#### 3. 访问应用

- 前端: http://localhost:4200
- 后端 API: http://localhost:8080

### 生产模式本地部署

#### 1. 构建前端

```bash
cd polardbx-ui
npm run build
```

#### 2. 构建后端

```bash
cd backend
go build -o polardbx-ui-backend main.go
```

#### 3. 配置 Nginx

创建 `nginx.conf` 文件：

```nginx
server {
    listen 80;
    server_name localhost;
    
    # 前端静态文件
    location / {
        root /path/to/polardbx-ui/dist/polardbx-ui;
        try_files $uri $uri/ /index.html;
    }
    
    # 后端 API 代理
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # 健康检查
    location /ping {
        proxy_pass http://localhost:8080;
    }
}
```

#### 4. 启动服务

```bash
# 启动后端
./polardbx-ui-backend

# 启动 Nginx
nginx -c /path/to/nginx.conf
```

## Docker 部署

### 单容器部署

#### 1. 创建 Dockerfile

**前端 Dockerfile** (`polardbx-ui/Dockerfile`):

```dockerfile
# 多阶段构建
FROM node:18-alpine AS build

# 设置工作目录
WORKDIR /app

# 复制 package 文件
COPY package*.json ./

# 安装依赖
RUN npm ci --only=production

# 复制源代码
COPY . .

# 构建应用
RUN npm run build

# 生产阶段
FROM nginx:alpine

# 复制构建结果
COPY --from=build /app/dist/polardbx-ui /usr/share/nginx/html

# 复制 Nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 暴露端口
EXPOSE 80

# 启动命令
CMD ["nginx", "-g", "daemon off;"]
```

**后端 Dockerfile** (`backend/Dockerfile`):

```dockerfile
# 构建阶段
FROM golang:1.23-alpine AS builder

# 安装必要工具
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 生产阶段
FROM alpine:latest

# 安装 ca-certificates
RUN apk --no-cache add ca-certificates

# 创建工作目录
WORKDIR /root/

# 复制二进制文件
COPY --from=builder /app/main .

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./main"]
```

#### 2. 构建镜像

```bash
# 构建前端镜像
cd polardbx-ui
docker build -t polardbx-ui-frontend .

# 构建后端镜像
cd ../backend
docker build -t polardbx-ui-backend .
```

#### 3. 运行容器

```bash
# 启动后端容器
docker run -d \
  --name polardbx-backend \
  -p 8080:8080 \
  polardbx-ui-backend

# 启动前端容器
docker run -d \
  --name polardbx-frontend \
  -p 80:80 \
  polardbx-ui-frontend
```

### 镜像命名规范（推荐）

- 前端镜像：`<repo>/polardbx-ui-frontend:<tag>`（示例：`registry.example.com/polardbx-ui-frontend:v1.0.0`）
- 后端镜像：`<repo>/polardbx-ui-backend:<tag>`（示例：`registry.example.com/polardbx-ui-backend:v1.0.0`）
- 标签建议：`<semver>-<gitsha>`，如 `v1.0.0-a1b2c3d`

### 容器环境变量清单（示例）

> 注：以下为部署时常用的建议变量名称，需确保在启动脚本或应用配置中被读取与生效。

- `PORT`：后端服务监听端口（默认 `8080`）
- `CORS_ENABLED`：是否启用 CORS（`true|false`）
- `CORS_ALLOW_ORIGINS`：允许的 Origin 列表，逗号分隔（示例：`http://localhost:4200`）
- `LOG_LEVEL`：日志级别（`debug|info|warn|error`，默认 `info`）
- `UI_BASE_HREF`：前端部署子路径（默认 `/`）

### Docker 健康检查（可选）

在后端镜像中添加健康检查命令：

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD wget -qO- http://localhost:8080/ping || exit 1
```

### Docker Compose 部署

#### 1. 创建 docker-compose.yml

```yaml
version: '3.8'

services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: polardbx-backend
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=release
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/ping"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  frontend:
    build:
      context: ./polardbx-ui
      dockerfile: Dockerfile
    container_name: polardbx-frontend
    ports:
      - "80:80"
    depends_on:
      - backend
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:80"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

networks:
  default:
    name: polardbx-network
```

#### 2. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## Kubernetes 部署

### 基础 Kubernetes 部署

#### 1. 创建命名空间

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: polardbx-ui
  labels:
    name: polardbx-ui
```

#### 2. 后端部署

```yaml
# backend-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: polardbx-backend
  namespace: polardbx-ui
  labels:
    app: polardbx-backend
spec:
  replicas: 2
  selector:
    matchLabels:
      app: polardbx-backend
  template:
    metadata:
      labels:
        app: polardbx-backend
    spec:
      containers:
      - name: backend
        image: polardbx-ui-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: GIN_MODE
          value: "release"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /ping
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ping
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: polardbx-backend-service
  namespace: polardbx-ui
spec:
  selector:
    app: polardbx-backend
  ports:
  - protocol: TCP
    port: 8080
    targetPort: 8080
  type: ClusterIP
```

#### 3. 前端部署

```yaml
# frontend-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: polardbx-frontend
  namespace: polardbx-ui
  labels:
    app: polardbx-frontend
spec:
  replicas: 2
  selector:
    matchLabels:
      app: polardbx-frontend
  template:
    metadata:
      labels:
        app: polardbx-frontend
    spec:
      containers:
      - name: frontend
        image: polardbx-ui-frontend:latest
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: polardbx-frontend-service
  namespace: polardbx-ui
spec:
  selector:
    app: polardbx-frontend
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
  type: LoadBalancer
```

#### 4. Ingress 配置

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: polardbx-ui-ingress
  namespace: polardbx-ui
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
spec:
  rules:
  - host: polardbx-ui.local
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: polardbx-backend-service
            port:
              number: 8080
      - path: /
        pathType: Prefix
        backend:
          service:
            name: polardbx-frontend-service
            port:
              number: 80
```

#### 5. 部署应用

```bash
# 应用所有配置
kubectl apply -f namespace.yaml
kubectl apply -f backend-deployment.yaml
kubectl apply -f frontend-deployment.yaml
kubectl apply -f ingress.yaml

# 检查部署状态
kubectl get pods -n polardbx-ui
kubectl get services -n polardbx-ui
kubectl get ingress -n polardbx-ui
```

### Helm 部署

#### 1. 创建 Helm Chart

```bash
# 创建 Helm chart
helm create polardbx-ui-chart
cd polardbx-ui-chart
```

#### 2. 配置 values.yaml

```yaml
# values.yaml
replicaCount:
  backend: 2
  frontend: 2

image:
  backend:
    repository: polardbx-ui-backend
    tag: latest
    pullPolicy: IfNotPresent
  frontend:
    repository: polardbx-ui-frontend
    tag: latest
    pullPolicy: IfNotPresent

service:
  backend:
    type: ClusterIP
    port: 8080
  frontend:
    type: LoadBalancer
    port: 80

ingress:
  enabled: true
  className: nginx
  annotations: {}
  hosts:
    - host: polardbx-ui.local
      paths:
        - path: /
          pathType: Prefix
  tls: []

resources:
  backend:
    limits:
      cpu: 500m
      memory: 512Mi
    requests:
      cpu: 250m
      memory: 256Mi
  frontend:
    limits:
      cpu: 200m
      memory: 256Mi
    requests:
      cpu: 100m
      memory: 128Mi

autoscaling:
  enabled: false
  minReplicas: 1
  maxReplicas: 100
  targetCPUUtilizationPercentage: 80

nodeSelector: {}

tolerations: []

affinity: {}
```

#### 3. 部署 Helm Chart

```bash
# 安装 chart
helm install polardbx-ui ./polardbx-ui-chart -n polardbx-ui --create-namespace

# 升级 chart
helm upgrade polardbx-ui ./polardbx-ui-chart -n polardbx-ui

# 卸载 chart
helm uninstall polardbx-ui -n polardbx-ui
```

## 生产环境部署

### 高可用部署

#### 1. 多副本配置

```yaml
# 后端高可用配置
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
```

#### 2. 资源限制

```yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

#### 3. 持久化存储

```yaml
# 如果需要持久化数据
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: polardbx-data
  namespace: polardbx-ui
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: fast-ssd
```

### 安全配置

#### 1. RBAC 配置

```yaml
# rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: polardbx-ui
  namespace: polardbx-ui
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: polardbx-ui-role
rules:
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: polardbx-ui-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: polardbx-ui-role
subjects:
- kind: ServiceAccount
  name: polardbx-ui
  namespace: polardbx-ui
```

#### 2. 网络策略

```yaml
# network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: polardbx-ui-netpol
  namespace: polardbx-ui
spec:
  podSelector:
    matchLabels:
      app: polardbx-backend
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: polardbx-frontend
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to: []
    ports:
    - protocol: TCP
      port: 443
    - protocol: TCP
      port: 6443
```

### SSL/TLS 配置

#### 1. 证书管理

```yaml
# certificate.yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: polardbx-ui-tls
  namespace: polardbx-ui
spec:
  secretName: polardbx-ui-tls-secret
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
  - polardbx-ui.yourdomain.com
```

#### 2. Ingress TLS 配置

```yaml
spec:
  tls:
  - hosts:
    - polardbx-ui.yourdomain.com
    secretName: polardbx-ui-tls-secret
```

## 监控和日志

### Prometheus 监控

#### 1. ServiceMonitor 配置

```yaml
# servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: polardbx-ui-monitor
  namespace: polardbx-ui
spec:
  selector:
    matchLabels:
      app: polardbx-backend
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

#### 2. 添加监控指标

在后端代码中添加 Prometheus 指标：

```go
// 在 main.go 中添加
import "github.com/prometheus/client_golang/prometheus/promhttp"

// 添加路由
r.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

### 日志收集

#### 1. Fluentd 配置

```yaml
# fluentd-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
  namespace: polardbx-ui
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/containers/polardbx-*.log
      pos_file /var/log/fluentd-containers.log.pos
      tag kubernetes.*
      format json
      read_from_head true
    </source>
    
    <match kubernetes.**>
      @type elasticsearch
      host elasticsearch.logging.svc.cluster.local
      port 9200
      index_name polardbx-ui
    </match>
```

### 健康检查

#### 1. 应用级健康检查

```yaml
livenessProbe:
  httpGet:
    path: /ping
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ping
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

## 故障排除

### 常见问题

#### 1. 容器启动失败

```bash
# 查看 Pod 状态
kubectl get pods -n polardbx-ui

# 查看 Pod 详情
kubectl describe pod <pod-name> -n polardbx-ui

# 查看容器日志
kubectl logs <pod-name> -n polardbx-ui

# 进入容器调试
kubectl exec -it <pod-name> -n polardbx-ui -- /bin/sh
```

#### 2. 服务连接问题

```bash
# 测试服务连通性
kubectl run test-pod --image=busybox --rm -it --restart=Never -- /bin/sh

# 在测试 Pod 中
wget -qO- http://polardbx-backend-service.polardbx-ui.svc.cluster.local:8080/ping
```

#### 3. Ingress 问题

```bash
# 检查 Ingress 状态
kubectl get ingress -n polardbx-ui
kubectl describe ingress polardbx-ui-ingress -n polardbx-ui

# 检查 Ingress Controller
kubectl get pods -n ingress-nginx
kubectl logs -n ingress-nginx <ingress-controller-pod>
```

### 性能调优

#### 1. 资源优化

```yaml
# 根据实际使用情况调整资源
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

#### 2. 水平扩展

```yaml
# HPA 配置
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: polardbx-backend-hpa
  namespace: polardbx-ui
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: polardbx-backend
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### 备份和恢复

#### 1. 配置备份

```bash
# 备份 Kubernetes 配置
kubectl get all -n polardbx-ui -o yaml > polardbx-ui-backup.yaml

# 备份 ConfigMaps 和 Secrets
kubectl get configmaps,secrets -n polardbx-ui -o yaml > polardbx-ui-configs-backup.yaml
```

#### 2. 恢复配置

```bash
# 恢复配置
kubectl apply -f polardbx-ui-backup.yaml
kubectl apply -f polardbx-ui-configs-backup.yaml
```

## 升级策略

### 滚动升级

```bash
# 更新镜像
kubectl set image deployment/polardbx-backend backend=polardbx-ui-backend:v2.0.0 -n polardbx-ui

# 查看升级状态
kubectl rollout status deployment/polardbx-backend -n polardbx-ui

# 回滚到上一版本
kubectl rollout undo deployment/polardbx-backend -n polardbx-ui
```

### 蓝绿部署

```bash
# 创建新版本部署
kubectl apply -f backend-deployment-v2.yaml

# 切换服务指向
kubectl patch service polardbx-backend-service -p '{"spec":{"selector":{"version":"v2"}}}'

# 删除旧版本
kubectl delete deployment polardbx-backend-v1
```

---

更多详细信息请参考：
- [README.md](./README.md)
- [开发环境配置指南](./DEVELOPMENT.md)
- [API 文档](./API_DOCUMENTATION.md)