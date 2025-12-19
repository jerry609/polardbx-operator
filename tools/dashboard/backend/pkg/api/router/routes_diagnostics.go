package router

import (
	domain_diagnostics "polardbx-dashboard-backend/pkg/api/domain/platform/diagnostics/handler"

	"github.com/gin-gonic/gin"
)

// RegisterDiagnosticsRoutes registers diagnostics-related routes
func RegisterDiagnosticsRoutes(v1 *gin.RouterGroup) {
	v1.POST("/diagnostics/:namespace/:cluster/start", domain_diagnostics.Start)
	v1.GET("/diagnostics/:namespace/:id/status", domain_diagnostics.GetStatus)
	v1.GET("/diagnostics/reports", domain_diagnostics.ListReports)
	v1.GET("/diagnostics/:namespace/:id/download", domain_diagnostics.Download)
	v1.GET("/diagnostics/:namespace/:id/file", domain_diagnostics.GetFile)
	v1.DELETE("/diagnostics/:namespace/:id", domain_diagnostics.DeleteJob)
}
