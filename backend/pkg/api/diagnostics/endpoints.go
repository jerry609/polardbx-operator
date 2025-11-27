// Package diagnostics 提供集群诊断的 API 端点
// 实际业务逻辑已迁移到 domain/platform/diagnostics
package diagnostics

import (
	"polardbx-ui-backend/pkg/api/domain/platform/diagnostics/handler"

	"github.com/gin-gonic/gin"
)

// 委托给 domain handler
var (
	Start       = handler.Start
	GetStatus   = handler.GetStatus
	ListReports = handler.ListReports
	Download    = handler.Download
)

// RegisterRoutes 注册诊断相关路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/diagnostics/:namespace/:cluster/start", Start)
	r.GET("/diagnostics/:namespace/:id/status", GetStatus)
	r.GET("/diagnostics/reports", ListReports)
	r.GET("/diagnostics/:namespace/:id/download", Download)
}
