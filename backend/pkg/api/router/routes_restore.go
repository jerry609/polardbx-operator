package router

import (
	domain_restore "polardbx-ui-backend/pkg/api/domain/platform/restore/handler"

	"github.com/gin-gonic/gin"
)

// RegisterRestoreRoutes registers restore-related routes
func RegisterRestoreRoutes(v1 *gin.RouterGroup) {
	// Cluster restore operations
	v1.POST("/clusters/:namespace/:name/restore", domain_restore.RestoreCluster)
	v1.POST("/clusters/:namespace/:name/pitr", domain_restore.InitiatePITR)
	v1.GET("/clusters/:namespace/:name/restore-status", domain_restore.GetRestoreStatus)

	// Restore jobs
	v1.GET("/restore-jobs", domain_restore.ListJobs)
	v1.GET("/restore-jobs/:namespace/:name", domain_restore.GetJob)
	v1.DELETE("/restore-jobs/:namespace/:name", domain_restore.CancelJob)
}
