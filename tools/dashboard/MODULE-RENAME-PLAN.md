# Module 名称重命名计划

## 当前状态

- **当前 module 名称**: `polardbx-ui-backend`
- **目录结构**: `tools/dashboard/backend/`
- **需要修改的文件数**: 339 个 Go 文件

## 建议的新名称

- `polardbx-dashboard-backend` (推荐)
- `dashboard-backend` (简洁)

## 重命名步骤

如果决定重命名，需要执行以下步骤：

1. **修改 go.mod**
   ```bash
   cd tools/dashboard/backend
   go mod edit -module polardbx-dashboard-backend
   ```

2. **批量替换所有 import 路径**
   ```bash
   find tools/dashboard/backend -name "*.go" -type f -exec sed -i 's|polardbx-ui-backend|polardbx-dashboard-backend|g' {} \;
   ```

3. **更新 go.mod 的 replace 路径**（如果需要）

4. **运行 go mod tidy** 验证

5. **重新构建和测试**

## 注意事项

- 这是一个破坏性更改
- 需要更新所有测试文件
- 需要更新 Dockerfile 中的引用
- 需要更新文档

## 建议

如果只是目录结构改变，module 名称可以保持不变，因为：
- module 名称是 Go 的内部标识符，不影响功能
- 修改需要大量工作且容易出错
- 保持向后兼容性

但如果要统一命名规范，可以执行重命名。

