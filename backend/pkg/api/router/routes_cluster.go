package router

import (
	domain_alerts "polardbx-ui-backend/pkg/api/domain/platform/alerts/handler"
	domain_auth "polardbx-ui-backend/pkg/api/domain/platform/auth/handler"
	domain_clusterknobs "polardbx-ui-backend/pkg/api/domain/platform/clusterknobs/handler"
	domain_pod "polardbx-ui-backend/pkg/api/domain/platform/pod/handler"
	domain_pxc "polardbx-ui-backend/pkg/api/domain/polardbxclusters"

	"github.com/gin-gonic/gin"
)

// RegisterClusterRoutes registers cluster-related routes
func RegisterClusterRoutes(v1 *gin.RouterGroup) {
	// Optional JWT middleware for all protected routes
	v1.Use(domain_auth.JWTAuthMiddleware())

	// Cluster CRUD
	v1.GET("/clusters", domain_pxc.List)
	v1.POST("/clusters", domain_pxc.Create)
	v1.POST("/clusters/:namespace/create", domain_pxc.CreateFromConfig)
	v1.GET("/clusters/:namespace/:name", domain_pxc.Get)
	v1.PUT("/clusters/:namespace/:name", domain_pxc.Update)
	v1.DELETE("/clusters/:namespace/:name", domain_pxc.Delete)

	// Cluster operations
	v1.PATCH("/clusters/:namespace/:name/log-config/:nodeType", domain_pxc.UpdateLogConfig)
	v1.PATCH("/clusters/:namespace/:name/scale", domain_pxc.Scale)
	v1.PATCH("/clusters/:namespace/:name/upgrade", domain_pxc.Upgrade)
	v1.GET("/clusters/:namespace/:name/upgrade-plan", domain_pxc.GetUpgradePlan)
	v1.GET("/clusters/:namespace/:name/alerts-summary", domain_pxc.GetAlertsSummary)
	v1.GET("/clusters/:namespace/:name/prechange-check", domain_pxc.GetPrechangeChecklist)
	v1.POST("/clusters/:namespace/:name/precheck", domain_pxc.Precheck)

	// Pods
	v1.GET("/pods", domain_pod.List)
	v1.GET("/pods/:namespace/:name", domain_pod.Get)
	v1.DELETE("/pods/:namespace/:name", domain_pod.Delete)
	v1.GET("/pods/:namespace/:name/exec", domain_pod.ExecWS)
	v1.GET("/clusters/:namespace/:name/pods", domain_pod.ListForCluster)
	v1.GET("/logs/:namespace/:pod_name", domain_pod.GetLogs)

	// Alerts
	v1.GET("/alerts", domain_alerts.List)

	// Parameters
	v1.GET("/parameters", domain_pxc.ListParameters)
	v1.POST("/parameters", domain_pxc.CreateParameter)
	v1.GET("/parameters/:name", domain_pxc.GetParameter)
	v1.PUT("/parameters/:name", domain_pxc.UpdateParameter)
	v1.DELETE("/parameters/:name", domain_pxc.DeleteParameter)

	// Parameter Templates
	v1.GET("/parameter-templates", domain_pxc.ListTemplates)
	v1.POST("/parameter-templates", domain_pxc.CreateTemplate)
	v1.GET("/parameter-templates/:namespace/:name", domain_pxc.GetTemplate)
	v1.PUT("/parameter-templates/:namespace/:name", domain_pxc.UpdateTemplate)
	v1.DELETE("/parameter-templates/:namespace/:name", domain_pxc.DeleteTemplate)

	// Cluster Knobs
	v1.GET("/cluster-knobs", domain_clusterknobs.GetList)
	v1.POST("/cluster-knobs", domain_clusterknobs.Create)
	v1.GET("/cluster-knobs/:namespace/:name", domain_clusterknobs.Get)
	v1.PUT("/cluster-knobs/:namespace/:name", domain_clusterknobs.Update)
	v1.DELETE("/cluster-knobs/:namespace/:name", domain_clusterknobs.Delete)
}
