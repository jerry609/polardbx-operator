# 故障排除指南

本文档提供了 PolarDB-X UI 项目常见问题的解决方案和调试方法。

## 目录

- [前端问题](#前端问题)
- [后端问题](#后端问题)
- [连接问题](#连接问题)
- [部署问题](#部署问题)
- [性能问题](#性能问题)
- [安全问题](#安全问题)
- [调试工具](#调试工具)
- [日志分析](#日志分析)

## 前端问题

### Angular 编译错误

#### 问题：`NG5002: Unexpected closing tag`

**症状**：
```
Error: src/app/app.component.html:X:Y - error NG5002: Unexpected closing tag "div". It may happen when the tag has already been closed by another tag.
```

**解决方案**：
1. 检查 HTML 模板中的标签配对
2. 确保所有开始标签都有对应的结束标签
3. 使用 IDE 的 HTML 验证功能

```bash
# 清理并重新构建
rm -rf node_modules .angular
npm install
ng build
```

#### 问题：依赖版本冲突

**症状**：
```
npm ERR! peer dep missing: @angular/core@^18.0.0
```

**解决方案**：
```bash
# 检查依赖兼容性
npm ls

# 更新依赖
npm update

# 强制解决冲突（谨慎使用）
npm install --force

# 或者使用 legacy peer deps
npm install --legacy-peer-deps
```

#### 问题：内存不足

**症状**：
```
JavaScript heap out of memory
```

**解决方案**：
```bash
# 增加 Node.js 内存限制
export NODE_OPTIONS="--max-old-space-size=8192"
ng build

# 或者在 package.json 中配置
{
  "scripts": {
    "build": "node --max-old-space-size=8192 ./node_modules/@angular/cli/bin/ng build"
  }
}
```

### 运行时错误

#### 问题：路由不工作

**症状**：页面刷新后显示 404 错误

**解决方案**：
1. 配置服务器支持 HTML5 路由
2. 在 Nginx 中添加：

```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

#### 问题：API 调用失败

**症状**：
```
CORS error: Access to XMLHttpRequest at 'http://localhost:8080/api/v1/clusters' from origin 'http://localhost:4200' has been blocked
```

**解决方案**：
1. 在开发环境中配置代理：

```json
// proxy.conf.json
{
  "/api/*": {
    "target": "http://localhost:8080",
    "secure": true,
    "changeOrigin": true
  }
}
```

2. 启动开发服务器时使用代理：

```bash
ng serve --proxy-config proxy.conf.json
```

## 后端问题

### Go 编译错误

#### 问题：模块依赖问题

**症状**：
```
go: module example.com/polardbx-ui requires go >= 1.19
```

**解决方案**：
```bash
# 检查 Go 版本
go version

# 更新 Go 版本或调整 go.mod
go mod edit -go=1.19
go mod tidy
```

#### 问题：包导入错误

**症状**：
```
package k8s.io/client-go/kubernetes: cannot find package
```

**解决方案**：
```bash
# 清理模块缓存
go clean -modcache

# 重新下载依赖
go mod download
go mod tidy

# 如果仍有问题，检查网络和代理设置
go env GOPROXY
```

### 运行时错误

#### 问题：Kubernetes 客户端初始化失败

**症状**：
```
failed to create Kubernetes client: invalid configuration
```

**解决方案**：
1. 检查 kubeconfig 文件格式
2. 验证集群连接：

```bash
# 测试 kubectl 连接
kubectl cluster-info

# 检查 kubeconfig 文件
kubectl config view

# 验证当前上下文
kubectl config current-context
```

#### 问题：端口被占用

**症状**：
```
listen tcp :8080: bind: address already in use
```

**解决方案**：
```bash
# 查找占用端口的进程
lsof -i :8080

# 杀死进程
kill -9 <PID>

# 或者使用不同端口
PORT=8081 go run main.go
```

#### 问题：内存泄漏

**症状**：应用运行一段时间后内存使用持续增长

**解决方案**：
1. 使用 pprof 进行内存分析：

```go
// 在 main.go 中添加
import _ "net/http/pprof"

// 启动 pprof 服务器
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

2. 分析内存使用：

```bash
# 获取内存 profile
go tool pprof http://localhost:6060/debug/pprof/heap

# 在 pprof 交互模式中
(pprof) top
(pprof) list <function_name>
```

## 连接问题

### Kubeconfig 问题

#### 问题：认证失败

**症状**：
```
HTTP 401: Unauthorized
```

**解决方案**：
1. 检查 kubeconfig 文件完整性：

```bash
# 验证 kubeconfig 格式
kubectl config view --validate

# 测试集群连接
kubectl get nodes
```

2. 检查证书有效期：

```bash
# 查看证书信息
openssl x509 -in ~/.kube/config -text -noout
```

3. 重新生成 kubeconfig：

```bash
# 对于 kubeadm 集群
sudo cp /etc/kubernetes/admin.conf ~/.kube/config
sudo chown $(id -u):$(id -g) ~/.kube/config
```

#### 问题：Base64 编码错误

**症状**：
```
invalid character '\n' in base64 input
```

**解决方案**：
```bash
# 正确的 base64 编码（去除换行符）
cat ~/.kube/config | base64 -w 0

# 或者在 macOS 上
cat ~/.kube/config | base64
```

### 网络连接问题

#### 问题：无法连接到 Kubernetes API

**症状**：
```
connection refused: dial tcp 192.168.1.100:6443: connect: connection refused
```

**解决方案**：
1. 检查网络连通性：

```bash
# 测试 API 服务器连接
telnet <api-server-ip> 6443

# 检查防火墙设置
sudo iptables -L

# 检查 DNS 解析
nslookup <api-server-hostname>
```

2. 检查集群状态：

```bash
# 检查 API 服务器状态
sudo systemctl status kube-apiserver

# 查看 API 服务器日志
sudo journalctl -u kube-apiserver -f
```

## 部署问题

### Docker 问题

#### 问题：镜像构建失败

**症状**：
```
ERROR: failed to solve: process "/bin/sh -c npm install" did not complete successfully
```

**解决方案**：
1. 检查 Dockerfile 语法
2. 使用多阶段构建优化：

```dockerfile
# 使用特定版本的基础镜像
FROM node:18-alpine AS build

# 设置工作目录
WORKDIR /app

# 先复制 package 文件
COPY package*.json ./

# 安装依赖
RUN npm ci --only=production

# 再复制源代码
COPY . .

# 构建应用
RUN npm run build
```

#### 问题：容器启动失败

**症状**：
```
container exited with code 1
```

**解决方案**：
```bash
# 查看容器日志
docker logs <container-id>

# 进入容器调试
docker run -it --entrypoint /bin/sh <image-name>

# 检查容器资源使用
docker stats <container-id>
```

### Kubernetes 部署问题

#### 问题：Pod 无法启动

**症状**：
```
Pod status: ImagePullBackOff
```

**解决方案**：
1. 检查镜像是否存在：

```bash
# 检查镜像
docker images | grep polardbx-ui

# 推送镜像到仓库
docker push <registry>/polardbx-ui:latest
```

2. 检查镜像拉取策略：

```yaml
spec:
  containers:
  - name: backend
    image: polardbx-ui-backend:latest
    imagePullPolicy: IfNotPresent  # 或 Always
```

#### 问题：服务无法访问

**症状**：
```
connection timed out
```

**解决方案**：
1. 检查服务配置：

```bash
# 检查服务
kubectl get svc -n polardbx-ui
kubectl describe svc <service-name> -n polardbx-ui

# 检查端点
kubectl get endpoints -n polardbx-ui
```

2. 测试服务连通性：

```bash
# 在集群内测试
kubectl run test-pod --image=busybox --rm -it --restart=Never -- /bin/sh
wget -qO- http://<service-name>.<namespace>.svc.cluster.local:<port>/ping
```

## 性能问题

### 前端性能优化

#### 问题：页面加载缓慢

**解决方案**：
1. 启用生产模式构建：

```bash
ng build --configuration production
```

2. 启用 gzip 压缩：

```nginx
gzip on;
gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
```

3. 使用 CDN 加速静态资源

#### 问题：内存使用过高

**解决方案**：
1. 使用 OnPush 变更检测策略
2. 实现虚拟滚动
3. 使用 TrackBy 函数优化 *ngFor

### 后端性能优化

#### 问题：API 响应缓慢

**解决方案**：
1. 添加性能监控：

```go
// 添加中间件记录请求时间
func LoggerMiddleware() gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\
",
            param.ClientIP,
            param.TimeStamp.Format(time.RFC1123),
            param.Method,
            param.Path,
            param.Request.Proto,
            param.StatusCode,
            param.Latency,
            param.Request.UserAgent(),
            param.ErrorMessage,
        )
    })
}
```

2. 实现缓存机制：

```go
// 使用 Redis 缓存
func CacheMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        key := c.Request.URL.Path
        if cached, exists := cache.Get(key); exists {
            c.JSON(200, cached)
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## 安全问题

### HTTPS 配置

#### 问题：SSL 证书错误

**症状**：
```
SSL certificate problem: self signed certificate
```

**解决方案**：
1. 使用有效的 SSL 证书
2. 配置 cert-manager 自动管理证书：

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
```

### 认证和授权

#### 问题：RBAC 权限不足

**症状**：
```
forbidden: User "system:serviceaccount:default:default" cannot get resource "pods" in API group "" in the namespace "default"
```

**解决方案**：
1. 创建适当的 RBAC 规则：

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: polardbx-ui-role
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

## 调试工具

### 前端调试

#### Angular DevTools

```bash
# 安装 Angular DevTools 浏览器扩展
# Chrome: https://chrome.google.com/webstore/detail/angular-devtools/
# Firefox: https://addons.mozilla.org/en-US/firefox/addon/angular-devtools/
```

#### 浏览器开发者工具

1. **Network 面板**：检查 API 请求和响应
2. **Console 面板**：查看 JavaScript 错误和日志
3. **Performance 面板**：分析页面性能
4. **Application 面板**：检查本地存储和缓存

### 后端调试

#### pprof 性能分析

```go
// 在 main.go 中添加
import _ "net/http/pprof"

func main() {
    // 启动 pprof 服务器
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // 其他代码...
}
```

使用 pprof：

```bash
# CPU 分析
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine 分析
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

#### Delve 调试器

```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动调试
dlv debug main.go

# 在代码中设置断点
(dlv) break main.main
(dlv) continue
```

### Kubernetes 调试

#### kubectl 调试命令

```bash
# 查看 Pod 详情
kubectl describe pod <pod-name> -n <namespace>

# 查看 Pod 日志
kubectl logs <pod-name> -n <namespace> -f

# 进入 Pod 调试
kubectl exec -it <pod-name> -n <namespace> -- /bin/sh

# 端口转发
kubectl port-forward <pod-name> 8080:8080 -n <namespace>

# 查看事件
kubectl get events -n <namespace> --sort-by='.lastTimestamp'
```

#### 临时调试 Pod

```bash
# 创建调试 Pod
kubectl run debug-pod --image=busybox --rm -it --restart=Never -- /bin/sh

# 或使用 nicolaka/netshoot 进行网络调试
kubectl run netshoot --image=nicolaka/netshoot --rm -it --restart=Never -- /bin/bash
```

## 日志分析

### 日志级别配置

#### 前端日志

```typescript
// 在 environment.ts 中配置
export const environment = {
  production: false,
  logLevel: 'debug' // debug, info, warn, error
};

// 创建日志服务
@Injectable({
  providedIn: 'root'
})
export class LoggerService {
  private logLevel = environment.logLevel;
  
  debug(message: string, ...args: any[]) {
    if (this.shouldLog('debug')) {
      console.debug(`[DEBUG] ${message}`, ...args);
    }
  }
  
  info(message: string, ...args: any[]) {
    if (this.shouldLog('info')) {
      console.info(`[INFO] ${message}`, ...args);
    }
  }
  
  private shouldLog(level: string): boolean {
    const levels = ['debug', 'info', 'warn', 'error'];
    return levels.indexOf(level) >= levels.indexOf(this.logLevel);
  }
}
```

#### 后端日志

```go
// 使用 logrus 进行结构化日志
import "github.com/sirupsen/logrus"

func init() {
    // 设置日志格式
    logrus.SetFormatter(&logrus.JSONFormatter{})
    
    // 设置日志级别
    logrus.SetLevel(logrus.InfoLevel)
    
    // 设置输出
    logrus.SetOutput(os.Stdout)
}

func LoggerMiddleware() gin.HandlerFunc {
    return gin.LoggerWithWriter(logrus.StandardLogger().Writer())
}
```

### 日志聚合

#### ELK Stack 配置

```yaml
# filebeat.yml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/containers/polardbx-*.log
  json.keys_under_root: true
  json.add_error_key: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "polardbx-ui-%{+yyyy.MM.dd}"

setup.kibana:
  host: "kibana:5601"
```

### 监控告警

#### Prometheus 告警规则

```yaml
# alerts.yml
groups:
- name: polardbx-ui
  rules:
  - alert: PolarDBXUIDown
    expr: up{job="polardbx-ui"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "PolarDB-X UI is down"
      description: "PolarDB-X UI has been down for more than 1 minute."
      
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High error rate detected"
      description: "Error rate is {{ $value }} errors per second."
```

## 常用命令速查

### 开发环境

```bash
# 前端开发
npm install                    # 安装依赖
ng serve                       # 启动开发服务器
ng build                       # 构建应用
ng test                        # 运行测试
npm run lint                   # 代码检查

# 后端开发
go mod tidy                    # 整理依赖
go run main.go                 # 运行应用
go build                       # 构建应用
go test ./...                  # 运行测试
golangci-lint run              # 代码检查
```

### Docker 操作

```bash
# 镜像操作
docker build -t app:latest .   # 构建镜像
docker images                  # 查看镜像
docker rmi <image-id>          # 删除镜像

# 容器操作
docker run -d -p 8080:8080 app # 运行容器
docker ps                      # 查看运行中的容器
docker logs <container-id>     # 查看日志
docker exec -it <container-id> /bin/sh  # 进入容器
docker stop <container-id>     # 停止容器
docker rm <container-id>       # 删除容器
```

### Kubernetes 操作

```bash
# 基本操作
kubectl get pods               # 查看 Pod
kubectl get svc                # 查看服务
kubectl get ingress            # 查看 Ingress
kubectl describe <resource> <name>  # 查看详情
kubectl logs <pod-name> -f     # 查看日志
kubectl exec -it <pod-name> -- /bin/sh  # 进入 Pod

# 部署操作
kubectl apply -f <file.yaml>   # 应用配置
kubectl delete -f <file.yaml>  # 删除资源
kubectl rollout restart deployment/<name>  # 重启部署
kubectl rollout status deployment/<name>   # 查看部署状态
```

---

如果遇到本文档未涵盖的问题，请：
1. 查看项目的 [GitHub Issues](https://github.com/your-repo/issues)
2. 创建新的 Issue 并提供详细的错误信息
3. 参考相关文档：[README.md](./README.md)、[开发指南](./DEVELOPMENT.md)、[部署指南](./DEPLOYMENT.md)