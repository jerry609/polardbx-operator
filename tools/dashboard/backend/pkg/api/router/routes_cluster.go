package router

import (
	domain_alerts "polardbx-dashboard-backend/pkg/api/domain/platform/alerts/handler"
	domain_auth "polardbx-dashboard-backend/pkg/api/domain/platform/auth/handler"
	domain_clusterknobs "polardbx-dashboard-backend/pkg/api/domain/platform/clusterknobs/handler"
	domain_pod "polardbx-dashboard-backend/pkg/api/domain/platform/pod/handler"
	domain_pxc "polardbx-dashboard-backend/pkg/api/domain/polardbxclusters"
	"polardbx-dashboard-backend/pkg/api/routerutil"

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
	routerutil.RegisterCRUD(v1, "/parameters", routerutil.CRUDHandlers{
		List:   domain_pxc.ListParameters,
		Create: domain_pxc.CreateParameter,
		Get:    domain_pxc.GetParameter,
		Update: domain_pxc.UpdateParameter,
		Delete: domain_pxc.DeleteParameter,
	})

	// Parameter Templates
	baseTpl := "/parameter-templates"
	routerutil.RegisterCRUDWithItemPattern(v1, baseTpl, baseTpl+"/:namespace/:name", routerutil.CRUDHandlers{
		List:   domain_pxc.ListTemplates,
		Create: domain_pxc.CreateTemplate,
		Get:    domain_pxc.GetTemplate,
		Update: domain_pxc.UpdateTemplate,
		Delete: domain_pxc.DeleteTemplate,
	})

	// Cluster Knobs
	baseKnobs := "/cluster-knobs"
	routerutil.RegisterCRUDWithItemPattern(v1, baseKnobs, baseKnobs+"/:namespace/:name", routerutil.CRUDHandlers{
		List:   domain_clusterknobs.GetList,
		Create: domain_clusterknobs.Create,
		Get:    domain_clusterknobs.Get,
		Update: domain_clusterknobs.Update,
		Delete: domain_clusterknobs.Delete,
	})
}
