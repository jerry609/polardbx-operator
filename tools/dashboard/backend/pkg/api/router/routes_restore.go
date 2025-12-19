package router

import (
	domain_restore "polardbx-dashboard-backend/pkg/api/domain/platform/restore/handler"
	"polardbx-dashboard-backend/pkg/api/routerutil"

	"github.com/gin-gonic/gin"
)

// RegisterRestoreRoutes registers restore-related routes
func RegisterRestoreRoutes(v1 *gin.RouterGroup) {
	// Cluster restore operations
	v1.POST("/clusters/:namespace/:name/restore", domain_restore.RestoreCluster)
	v1.POST("/clusters/:namespace/:name/pitr", domain_restore.InitiatePITR)
	v1.GET("/clusters/:namespace/:name/restore-status", domain_restore.GetRestoreStatus)

	// Restore jobs
	routerutil.RegisterLGD(v1, "/restore-jobs", "/restore-jobs/:namespace/:name", routerutil.LGDHandlers{
		List:   domain_restore.ListJobs,
		Get:    domain_restore.GetJob,
		Delete: domain_restore.CancelJob,
	})
}
