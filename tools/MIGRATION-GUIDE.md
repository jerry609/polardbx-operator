# 迁移指南：Backend 和 UI 迁移到 tools 目录

## 📋 概述

本指南详细说明如何将 `backend/` 和 `polardbx-ui/` 迁移到 `tools/` 目录下。

## 🎯 目标结构

```
polardbx-operator/
├── api/                    # 保持不变
├── pkg/                    # 保持不变
├── tools/
│   ├── ui-backend/         # ✅ 新位置（原 backend/）
│   ├── ui-frontend/        # ✅ 新位置（原 polardbx-ui/）
│   ├── ui-all-in-one.Dockerfile
│   └── ...
```

## 🚀 快速迁移（使用脚本）

```bash
# 1. 运行迁移脚本
./tools/migrate-ui-to-tools.sh

# 2. 验证迁移结果
cd tools/ui-backend
go mod tidy
go build -o /tmp/test-build ./main.go

cd ../ui-frontend
npm install
npm run build

# 3. 测试 Docker 构建
cd ../..
docker build -f tools/ui-all-in-one.Dockerfile -t test-migration .

# 4. 如果一切正常，删除原目录
rm -rf backend polardbx-ui
```

## 📝 手动迁移步骤

### 步骤 1: 创建目录并移动文件

```bash
# 创建目标目录
mkdir -p tools/ui-backend
mkdir -p tools/ui-frontend

# 移动文件（保留原目录作为备份）
cp -r backend/* tools/ui-backend/
cp -r polardbx-ui/* tools/ui-frontend/
```

### 步骤 2: 更新 go.mod

编辑 `tools/ui-backend/go.mod`：

```go
// 原：replace github.com/alibaba/polardbx-operator => ../
// 新：从 tools/ui-backend/ 到根目录需要上两级
replace github.com/alibaba/polardbx-operator => ../../
```

### 步骤 3: 更新 Dockerfile

编辑 `tools/ui-all-in-one.Dockerfile`：

**变更前**：
```dockerfile
COPY polardbx-ui/package*.json ./
COPY polardbx-ui/ ./
COPY backend/go.mod backend/go.sum ./backend/
COPY backend/ ./backend/
```

**变更后**：
```dockerfile
COPY tools/ui-frontend/package*.json ./
COPY tools/ui-frontend/ ./
COPY tools/ui-backend/go.mod tools/ui-backend/go.sum ./backend/
COPY tools/ui-backend/ ./backend/
```

同时更新工作目录：
```dockerfile
WORKDIR /workspace/ui-frontend  # 原：/workspace/polardbx-ui
```

### 步骤 4: 更新脚本

更新以下脚本中的路径：
- `tools/deploy-all-in-one.sh`
- `tools/test-all-in-one.sh`

查找并替换：
- `backend/` → `tools/ui-backend/`
- `polardbx-ui/` → `tools/ui-frontend/`

### 步骤 5: 更新文档

更新以下文档中的路径引用：
- `tools/USER-GUIDE.md`
- `tools/QUICK-START.md`
- `tools/README-ALL-IN-ONE.md`
- `tools/ui-frontend/README.md`（如果存在）

## ✅ 验证清单

迁移完成后，验证以下内容：

- [ ] `tools/ui-backend/go.mod` 中的 replace 路径为 `../../`
- [ ] `tools/ui-backend/` 可以正常构建：`cd tools/ui-backend && go build`
- [ ] `tools/ui-frontend/` 可以正常构建：`cd tools/ui-frontend && npm run build`
- [ ] Dockerfile 可以成功构建镜像
- [ ] 所有脚本可以正常运行
- [ ] 文档中的路径引用已更新

## 🔄 回滚

如果迁移出现问题：

```bash
# 恢复原目录
mv tools/ui-backend backend
mv tools/ui-frontend polardbx-ui

# 恢复 go.mod
cd backend
# 修改 go.mod: replace github.com/alibaba/polardbx-operator => ../

# 恢复 Dockerfile 和脚本（使用备份文件）
cp tools/ui-all-in-one.Dockerfile.bak tools/ui-all-in-one.Dockerfile
```

## 📚 开发工作流变更

### 原工作流

```bash
# 后端开发
cd backend
go run main.go

# 前端开发
cd polardbx-ui
npm start
```

### 新工作流

```bash
# 后端开发
cd tools/ui-backend
go run main.go

# 前端开发
cd tools/ui-frontend
npm start
```

## 🎯 优势

1. **工具集中化**：所有工具和辅助组件在 `tools/` 目录下
2. **清晰的命名**：`ui-backend` 和 `ui-frontend` 明确表示用途
3. **保持依赖**：`api/` 和 `pkg/` 保持在根目录，依赖关系不变
4. **易于维护**：UI 相关代码集中在一个目录下

## ⚠️ 注意事项

1. **构建上下文**：Dockerfile 必须在项目根目录执行
2. **Go Module**：确保 `go.mod` 中的 replace 路径正确
3. **CI/CD**：如果使用 CI/CD，需要更新相关配置
4. **IDE 配置**：可能需要更新 IDE 的项目路径配置

