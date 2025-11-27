// Package settings 提供设置管理的 API 端点
// 实际业务逻辑已迁移到 domain/platform/settings
package settings

import (
	"polardbx-ui-backend/pkg/api/domain/platform/settings/handler"
	"polardbx-ui-backend/pkg/api/domain/platform/settings/service"

	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// 导出类型别名，保持兼容性
type BackupDashboardSettings = service.BackupDashboardSettings

// 委托给 domain handler
var (
	Get    = handler.Get
	Update = handler.Update
)

// ReadDashboardSettings 读取备份仪表盘设置（供其他包使用）
// 保持原有接口，内部委托给新的 handler
func ReadDashboardSettings(c *gin.Context, cli client.Client) (BackupDashboardSettings, error) {
	return handler.ReadDashboardSettings(c.Request.Context(), cli), nil
}

// RegisterRoutes 注册设置相关路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/settings", Get)
	r.PUT("/settings", Update)
}
