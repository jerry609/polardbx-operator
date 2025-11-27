// Package settings 提供设置管理的 API 端点
// 实际业务逻辑已迁移到 domain/platform/settings
package settings

import (
	"polardbx-ui-backend/pkg/api/domain/platform/settings/handler"

	"github.com/gin-gonic/gin"
)

// 导出类型别名，保持兼容性
type UpdateImageRegistryRequest = handler.UpdateImageRegistryRequest
type TestImageRegistryRequest = handler.TestImageRegistryRequest

// 委托给 domain handler
var (
	GetImageRegistryConfig    = handler.GetImageRegistryConfig
	UpdateImageRegistryConfig = handler.UpdateImageRegistryConfig
	GetAvailableRegistries    = handler.GetAvailableRegistries
	TestImageRegistry         = handler.TestImageRegistry
)

// RegisterImageRegistryRoutes 注册镜像仓库相关路由
func RegisterImageRegistryRoutes(r *gin.RouterGroup) {
	r.GET("/image-registry/config", GetImageRegistryConfig)
	r.PUT("/image-registry/config", UpdateImageRegistryConfig)
	r.GET("/image-registry/presets", GetAvailableRegistries)
	r.POST("/image-registry/test", TestImageRegistry)
}
