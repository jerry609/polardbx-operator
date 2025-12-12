# Swagger/OpenAPI 文档设置指南

## 概述

项目已集成 Swagger UI，提供交互式 API 文档。Swagger 支持 OpenAPI 3.0 规范，可以自动生成和展示 API 文档。

## 启用方式

### 方式 1: 通过环境变量（推荐）

```bash
export ENABLE_SWAGGER=true
make run
```

### 方式 2: Debug 模式自动启用

在 debug 模式下，Swagger 会自动启用：

```bash
export GIN_MODE=debug
make run
```

## 访问地址

启动服务后，可以通过以下地址访问：

- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **OpenAPI JSON**: `http://localhost:8080/swagger/doc.json`
- **Swagger 根路径**: `http://localhost:8080/swagger` (自动重定向到 UI)

## 功能特性

### 1. 交互式 API 文档

- 浏览所有 API 端点
- 查看请求/响应模型
- 直接在浏览器中测试 API
- 查看认证要求

### 2. OpenAPI 规范

- 符合 OpenAPI 3.0 标准
- 支持 JSON 格式导出
- 可与其他工具集成（Postman、Insomnia 等）

### 3. 安全配置

- 默认仅在 debug 模式或明确启用时可用
- 生产环境可通过 `ENABLE_SWAGGER=true` 显式启用
- 支持 Kubeconfig 认证说明

## 当前文档覆盖

目前文档包含以下端点：

- ✅ Health Check 端点 (`/health`, `/ready`, `/version`)
- ⚠️ 其他 API 端点（可通过代码注释扩展）

## 扩展文档

### 使用代码注释生成文档

可以使用 `swag` 工具从代码注释自动生成文档：

1. 安装 swag：
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

2. 在 handler 函数上添加注释：
```go
// @Summary      Get cluster
// @Description  Get PolarDB-X cluster by name
// @Tags         Clusters
// @Accept       json
// @Produce      json
// @Param        namespace  path  string  true  "Namespace"
// @Param        name       path  string  true  "Cluster name"
// @Success      200  {object}  polardbxv1.PolarDBXCluster
// @Failure      404  {object}  errors.APIError
// @Router       /api/v1/clusters/{namespace}/{name} [get]
func GetCluster(c *gin.Context) {
    // ...
}
```

3. 生成文档：
```bash
make generate-swagger
```

### 手动更新 OpenAPI 规范

如果需要手动更新 OpenAPI 规范，可以编辑 `pkg/api/router/swagger.go` 中的 `getSwaggerSpec()` 函数。

## 配置说明

### 环境变量

- `ENABLE_SWAGGER`: 是否启用 Swagger（默认: `false`）
- `GIN_MODE`: Gin 运行模式，`debug` 模式下自动启用（默认: `debug`）

### 代码配置

Swagger 路由注册在 `pkg/api/router/setup.go` 中：

```go
// setupRoutes registers all route groups
func setupRoutes(r *gin.Engine) {
    // Swagger/OpenAPI documentation (only in debug mode or when enabled)
    RegisterSwaggerRoutes(r)
    // ...
}
```

## 最佳实践

1. **开发环境**: 使用 debug 模式，自动启用 Swagger
2. **生产环境**: 通过 `ENABLE_SWAGGER=true` 显式启用（如需要）
3. **文档维护**: 使用代码注释自动生成，保持文档与代码同步
4. **安全考虑**: 生产环境建议限制 Swagger 访问（通过反向代理或 IP 白名单）

## 故障排查

### Swagger UI 无法访问

1. 检查服务是否运行在 debug 模式或 `ENABLE_SWAGGER=true`
2. 检查端口是否正确（默认 8080）
3. 查看日志确认路由已注册

### 文档不完整

1. 使用 `make generate-swagger` 重新生成文档
2. 检查代码注释格式是否正确
3. 查看 `swagger/doc.json` 确认规范是否正确

## 相关文件

- `pkg/api/router/swagger.go` - Swagger 路由和规范定义
- `pkg/api/router/setup.go` - 路由注册
- `pkg/config/server.go` - 配置管理
- `Makefile` - 包含 `generate-swagger` 命令

## 参考资料

- [Swagger UI](https://swagger.io/tools/swagger-ui/)
- [OpenAPI Specification](https://swagger.io/specification/)
- [swaggo/swag](https://github.com/swaggo/swag)

