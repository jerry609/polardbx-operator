# 开发环境配置指南

本文档提供了 PolarDB-X UI 项目的详细开发环境配置指南。

## 环境要求

### 必需软件

- **Node.js**: 18.x 或更高版本
- **npm**: 9.x 或更高版本
- **Go**: 1.19 或更高版本
- **Angular CLI**: 19.x
- **Git**: 最新版本

### 可选软件

- **Docker**: 用于容器化部署
- **kubectl**: 用于 Kubernetes 集群管理
- **VS Code**: 推荐的开发编辑器

## 快速设置

### 1. 环境检查

```bash
# 检查 Node.js 版本
node --version
# 应该显示 v18.x.x 或更高

# 检查 npm 版本
npm --version
# 应该显示 9.x.x 或更高

# 检查 Go 版本
go version
# 应该显示 go1.19 或更高

# 检查 Angular CLI
ng version
# 如果未安装，运行: npm install -g @angular/cli
```

### 2. 项目设置

```bash
# 进入项目目录
cd polardbx-operator/polardbx-ui

# 安装前端依赖
npm install

# 进入后端目录
cd ../backend

# 下载 Go 依赖
go mod tidy
```

### 3. 开发服务器启动

#### 启动后端服务

```bash
# 在 backend 目录下
go run main.go

# 或者使用 air 进行热重载（需要先安装 air）
# go install github.com/cosmtrek/air@latest
# air
```

#### 启动前端服务

```bash
# 在 polardbx-ui 目录下
ng serve

# 或者指定端口
ng serve --port 4200

# 开启热重载和自动打开浏览器
ng serve --open --hmr
```

## 开发工具配置

### VS Code 推荐扩展

创建 `.vscode/extensions.json` 文件：

```json
{
  "recommendations": [
    "angular.ng-template",
    "ms-vscode.vscode-typescript-next",
    "golang.go",
    "bradlc.vscode-tailwindcss",
    "esbenp.prettier-vscode",
    "ms-vscode.vscode-json",
    "redhat.vscode-yaml"
  ]
}
```

### VS Code 设置

创建 `.vscode/settings.json` 文件：

```json
{
  "typescript.preferences.importModuleSpecifier": "relative",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.testFlags": ["-v"],
  "files.exclude": {
    "**/node_modules": true,
    "**/dist": true,
    "**/.angular": true
  }
}
```

### 调试配置

创建 `.vscode/launch.json` 文件：

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Backend",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/backend/main.go",
      "cwd": "${workspaceFolder}/backend",
      "env": {},
      "args": []
    },
    {
      "name": "Debug Frontend",
      "type": "node",
      "request": "launch",
      "program": "${workspaceFolder}/polardbx-ui/node_modules/@angular/cli/bin/ng",
      "args": ["serve"],
      "cwd": "${workspaceFolder}/polardbx-ui",
      "console": "integratedTerminal"
    }
  ]
}
```

## 代码质量工具

### 前端代码质量

#### ESLint 配置

```bash
# 安装 ESLint
npm install --save-dev @angular-eslint/builder @angular-eslint/eslint-plugin @angular-eslint/eslint-plugin-template @angular-eslint/schematics @angular-eslint/template-parser @typescript-eslint/eslint-plugin @typescript-eslint/parser eslint

# 初始化 ESLint 配置
ng add @angular-eslint/schematics
```

#### Prettier 配置

创建 `.prettierrc` 文件：

```json
{
  "semi": true,
  "trailingComma": "es5",
  "singleQuote": true,
  "printWidth": 80,
  "tabWidth": 2,
  "useTabs": false
}
```

### 后端代码质量

#### golangci-lint 配置

创建 `.golangci.yml` 文件：

```yaml
run:
  timeout: 5m
  issues-exit-code: 1
  tests: true

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
    - typecheck

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

## 测试配置

### 前端测试

```bash
# 运行单元测试
ng test

# 运行测试并生成覆盖率报告
ng test --code-coverage

# 运行端到端测试
ng e2e
```

### 后端测试

```bash
# 运行所有测试
go test ./...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 运行基准测试
go test -bench=. ./...
```

## 构建和部署

### 本地构建

```bash
# 构建前端生产版本
ng build --configuration production

# 构建后端二进制文件
cd backend
go build -o polardbx-ui-backend main.go
```

### Docker 构建

#### 前端 Dockerfile

```dockerfile
# 多阶段构建
FROM node:18-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist/polardbx-ui /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

#### 后端 Dockerfile

```dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

## 常见问题解决

### Node.js 版本问题

```bash
# 使用 nvm 管理 Node.js 版本
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
nvm install 18
nvm use 18
```

### 依赖安装问题

```bash
# 清理 npm 缓存
npm cache clean --force

# 删除 node_modules 重新安装
rm -rf node_modules package-lock.json
npm install
```

### Go 模块问题

```bash
# 清理 Go 模块缓存
go clean -modcache

# 重新下载依赖
go mod download
go mod tidy
```

### 端口冲突

```bash
# 查找占用端口的进程
lsof -i :4200
lsof -i :8080

# 杀死进程
kill -9 <PID>
```

## 开发工作流

### Git 工作流

1. **创建功能分支**
   ```bash
   git checkout -b feature/new-feature
   ```

2. **提交代码**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

3. **推送分支**
   ```bash
   git push origin feature/new-feature
   ```

4. **创建 Pull Request**

### 提交信息规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

- `feat:` 新功能
- `fix:` 修复 bug
- `docs:` 文档更新
- `style:` 代码格式化
- `refactor:` 代码重构
- `test:` 测试相关
- `chore:` 构建过程或辅助工具的变动

## 性能优化

### 前端优化

- 使用 OnPush 变更检测策略
- 实现虚拟滚动
- 使用 TrackBy 函数
- 懒加载模块
- 优化包大小

### 后端优化

- 使用连接池
- 实现缓存机制
- 优化数据库查询
- 使用 pprof 进行性能分析

## 监控和日志

### 开发环境监控

```bash
# 监控前端构建
ng build --watch

# 监控后端变化（使用 air）
air
```

### 日志配置

- 前端：使用浏览器开发者工具
- 后端：配置结构化日志输出

---

如有问题，请参考 [README.md](./README.md) 或创建 Issue。