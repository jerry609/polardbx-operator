# Backend 和 UI 迁移到 tools 目录的设计方案

## 📋 目标

将 `backend/` 和 `polardbx-ui/` 迁移到 `tools/` 目录下，重新组织目录结构，同时保持所有依赖关系和构建流程。

## 🎯 设计原则

1. **保持依赖关系**：backend 依赖 `api/` 和根目录 `pkg/`，这些保持不变
2. **清晰的命名**：使用 `ui-backend` 和 `ui-frontend` 明确表示这是 UI 相关组件
3. **最小化影响**：尽量保持现有代码结构，只移动目录
4. **工具集中化**：所有工具和辅助组件集中在 `tools/` 目录

## 📁 目标目录结构

```
polardbx-operator/
├── api/                          # 保持不变（backend 依赖）
├── pkg/                          # 保持不变（backend 依赖）
├── backend/                      # ❌ 删除（迁移到 tools/ui-backend/）
├── polardbx-ui/                  # ❌ 删除（迁移到 tools/ui-frontend/）
└── tools/
    ├── ui-backend/               # ✅ 新位置（原 backend/）
    │   ├── main.go
    │   ├── go.mod
    │   ├── go.sum
    │   ├── pkg/
    │   │   ├── api/
    │   │   ├── config/
    │   │   ├── k8s/
    │   │   ├── cache/
    │   │   └── logger/
    │   └── ...
    ├── ui-frontend/              # ✅ 新位置（原 polardbx-ui/）
    │   ├── src/
    │   ├── angular.json
    │   ├── package.json
    │   └── ...
    ├── ui-all-in-one.Dockerfile  # ✅ 已存在，需要更新路径
    ├── deploy-all-in-one.sh      # ✅ 已存在，需要更新路径
    ├── test-all-in-one.sh        # ✅ 已存在，需要更新路径
    └── ...
```

## 🔄 迁移步骤

### 步骤 1: 创建新目录结构

```bash
# 创建新目录
mkdir -p tools/ui-backend
mkdir -p tools/ui-frontend

# 移动 backend
mv backend/* tools/ui-backend/
rmdir backend

# 移动 polardbx-ui
mv polardbx-ui/* tools/ui-frontend/
rmdir polardbx-ui
```

### 步骤 2: 更新 backend/go.mod

**原路径**：
```go
replace github.com/alibaba/polardbx-operator => ../
```

**新路径**（从 `tools/ui-backend/` 到根目录）：
```go
replace github.com/alibaba/polardbx-operator => ../../
```

### 步骤 3: 更新 Dockerfile

**原路径** (`tools/ui-all-in-one.Dockerfile`)：
```dockerfile
COPY polardbx-ui/package*.json ./
COPY polardbx-ui/ ./
COPY backend/go.mod backend/go.sum ./backend/
COPY backend/ ./backend/
```

**新路径**：
```dockerfile
COPY tools/ui-frontend/package*.json ./
COPY tools/ui-frontend/ ./
COPY tools/ui-backend/go.mod tools/ui-backend/go.sum ./backend/
COPY tools/ui-backend/ ./backend/
```

### 步骤 4: 更新脚本和配置文件

需要更新的文件：
- `tools/ui-all-in-one.Dockerfile`
- `tools/deploy-all-in-one.sh`
- `tools/test-all-in-one.sh`
- `tools/helm-example/templates/*.yaml`（如果有路径引用）
- `polardbx-ui/README.md`（如果存在）

### 步骤 5: 更新文档中的路径引用

需要更新的文档：
- `tools/USER-GUIDE.md`
- `tools/QUICK-START.md`
- `tools/README-ALL-IN-ONE.md`
- `tools/IMPLEMENTATION-SUMMARY.md`
- 其他引用 `backend/` 或 `polardbx-ui/` 的文档

## 📝 详细变更清单

### 文件路径变更

| 原路径 | 新路径 |
|--------|--------|
| `backend/` | `tools/ui-backend/` |
| `polardbx-ui/` | `tools/ui-frontend/` |
| `backend/main.go` | `tools/ui-backend/main.go` |
| `backend/go.mod` | `tools/ui-backend/go.mod` |
| `polardbx-ui/src/` | `tools/ui-frontend/src/` |
| `polardbx-ui/package.json` | `tools/ui-frontend/package.json` |

### 代码变更

#### 1. `tools/ui-backend/go.mod`

```go
// 原：replace github.com/alibaba/polardbx-operator => ../
// 新：从 tools/ui-backend/ 到根目录需要上两级
replace github.com/alibaba/polardbx-operator => ../../
```

#### 2. `tools/ui-all-in-one.Dockerfile`

```dockerfile
# 前端构建阶段
FROM node:18-alpine AS frontend-build
WORKDIR /workspace/ui-frontend
COPY tools/ui-frontend/package*.json ./
RUN npm ci
COPY tools/ui-frontend/ ./
RUN npm run build

# 后端构建阶段
FROM golang:1.21-alpine AS backend-build
WORKDIR /workspace
COPY go.mod go.sum ./
COPY tools/ui-backend/go.mod tools/ui-backend/go.sum ./backend/
RUN go mod download
COPY tools/ui-backend/ ./backend/
COPY api/ ./api/
COPY pkg/ ./pkg/
WORKDIR /workspace/backend
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/polardbx-ui-backend ./main.go

# 最终阶段
FROM alpine:latest
WORKDIR /app
COPY --from=backend-build /bin/polardbx-ui-backend /app/polardbx-ui-backend
COPY --from=frontend-build /workspace/ui-frontend/dist/polardbx-ui/browser/ /app/ui/
ENV UI_STATIC_DIR=/app/ui
EXPOSE 8080
ENTRYPOINT ["/app/polardbx-ui-backend"]
```

#### 3. 脚本更新示例

**`tools/deploy-all-in-one.sh`**：
```bash
# 原：BACKEND_DIR="backend"
# 新：
BACKEND_DIR="tools/ui-backend"
FRONTEND_DIR="tools/ui-frontend"
```

**`tools/test-all-in-one.sh`**：
```bash
# 原：docker build -f tools/ui-all-in-one.Dockerfile -t ...
# 路径保持不变（在项目根目录执行）
```

### 文档更新

#### `tools/USER-GUIDE.md`

```markdown
# 原：
cd polardbx-ui
npm install

# 新：
cd tools/ui-frontend
npm install
```

#### `polardbx-ui/README.md`（移动到 `tools/ui-frontend/README.md`）

```markdown
# 原：
cd polardbx-operator/polardbx-ui

# 新：
cd polardbx-operator/tools/ui-frontend
```

## ⚠️ 注意事项

### 1. Go Module 路径

- `backend/go.mod` 中的 `replace` 指令需要从 `../` 改为 `../../`
- 确保构建时工作目录正确

### 2. Docker 构建上下文

- Dockerfile 需要在**项目根目录**执行
- 构建上下文是项目根目录，所以路径是 `tools/ui-backend/` 和 `tools/ui-frontend/`

### 3. 开发工作流

**原工作流**：
```bash
cd backend
go run main.go

cd ../polardbx-ui
npm start
```

**新工作流**：
```bash
cd tools/ui-backend
go run main.go

cd ../ui-frontend
npm start
```

### 4. CI/CD 配置

如果项目有 CI/CD 配置，需要更新：
- GitHub Actions workflows
- GitLab CI/CD
- Jenkins pipelines
- 其他自动化脚本

### 5. IDE 配置

- VS Code workspace 配置
- GoLand/IntelliJ 项目配置
- 其他 IDE 的项目路径配置

## ✅ 验证清单

迁移完成后，验证以下内容：

- [ ] `tools/ui-backend/go.mod` 中的 replace 路径正确
- [ ] Dockerfile 可以成功构建镜像
- [ ] 后端服务可以正常启动
- [ ] 前端应用可以正常启动
- [ ] 所有脚本可以正常运行
- [ ] 文档中的路径引用已更新
- [ ] CI/CD 配置已更新（如果有）
- [ ] 测试通过

## 🔄 回滚方案

如果迁移出现问题，可以快速回滚：

```bash
# 恢复原目录结构
mv tools/ui-backend backend
mv tools/ui-frontend polardbx-ui

# 恢复 go.mod
cd backend
# 修改 go.mod 中的 replace 路径回 ../

# 恢复 Dockerfile 和脚本中的路径引用
```

## 📚 迁移后目录结构示例

```
polardbx-operator/
├── api/                          # API 定义（不变）
├── pkg/                          # Operator 包（不变）
├── cmd/                          # Operator 命令（不变）
├── charts/                       # Helm Charts（不变）
├── tools/
│   ├── ui-backend/               # ✅ UI 后端（新位置）
│   │   ├── main.go
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── pkg/
│   │   │   ├── api/
│   │   │   ├── config/
│   │   │   ├── k8s/
│   │   │   └── ...
│   │   └── ...
│   ├── ui-frontend/              # ✅ UI 前端（新位置）
│   │   ├── src/
│   │   ├── angular.json
│   │   ├── package.json
│   │   └── ...
│   ├── ui-all-in-one.Dockerfile  # ✅ 已更新路径
│   ├── deploy-all-in-one.sh      # ✅ 已更新路径
│   ├── test-all-in-one.sh        # ✅ 已更新路径
│   ├── helm-example/             # Helm Chart（不变）
│   └── ...
└── ...
```

## 🎯 优势

1. **工具集中化**：所有工具和辅助组件在 `tools/` 目录下
2. **清晰的命名**：`ui-backend` 和 `ui-frontend` 明确表示用途
3. **保持依赖**：`api/` 和 `pkg/` 保持在根目录，backend 依赖关系不变
4. **易于维护**：UI 相关代码集中在一个目录下

## 📝 下一步

1. 创建迁移脚本自动化执行
2. 更新所有文档
3. 测试验证
4. 提交变更

