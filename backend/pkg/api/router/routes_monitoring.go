package router

import (
	domain_monitoring "polardbx-ui-backend/pkg/api/domain/monitoring"
	domain_grafana "polardbx-ui-backend/pkg/api/domain/platform/grafana/handler"
	domain_prometheusrule "polardbx-ui-backend/pkg/api/domain/platform/prometheusrule/handler"

	"github.com/gin-gonic/gin"
)

// RegisterMonitoringRoutes registers monitoring-related routes
func RegisterMonitoringRoutes(v1 *gin.RouterGroup) {
	// Monitor CRD routes
	v1.GET("/monitors", domain_monitoring.ListMonitors)
	v1.POST("/monitors", domain_monitoring.CreateMonitor)
	v1.GET("/monitors/:namespace/:name", domain_monitoring.GetMonitor)
	v1.PUT("/monitors/:namespace/:name", domain_monitoring.UpdateMonitor)
	v1.DELETE("/monitors/:namespace/:name", domain_monitoring.DeleteMonitor)

	// Monitoring workflow endpoints
	v1.POST("/monitoring/bootstrap", domain_monitoring.Bootstrap)
	v1.GET("/monitoring/bootstrap/status", domain_monitoring.BootstrapStatus)
	v1.GET("/monitoring/bootstrap/logs", domain_monitoring.GetBootstrapLogs)
	v1.GET("/monitoring/status", domain_monitoring.Status)
	v1.GET("/monitoring/preflight", domain_monitoring.DetectEnvironment)
	v1.DELETE("/monitoring/uninstall", domain_monitoring.Uninstall)

	// Monitoring v2 workflow endpoints
	v1.GET("/monitoring/detect", domain_monitoring.DetectEnvironment)
	v1.POST("/monitoring/plan", domain_monitoring.CreatePlan)
	v1.POST("/monitoring/install", domain_monitoring.StartInstallation)
	v1.GET("/monitoring/install/:sessionId/status", domain_monitoring.GetInstallStatus)
	v1.GET("/monitoring/install/:sessionId/logs", domain_monitoring.GetBootstrapLogs)
	v1.POST("/monitoring/install/:sessionId/retry", domain_monitoring.TriggerRetry)
	v1.POST("/monitoring/diagnose", domain_monitoring.DiagnoseFailure)
	v1.POST("/monitoring/auto-fix", domain_monitoring.ApplyAutoFix)

	// Grafana routes
	v1.GET("/monitoring/grafana/config", domain_grafana.GetConfig)
	v1.PUT("/monitoring/grafana/config", domain_grafana.PutConfig)
	v1.POST("/monitoring/grafana/dashboards/sync", domain_grafana.SyncDashboards)
	v1.GET("/monitoring/grafana/dashboards", domain_grafana.ListDashboards)
	v1.GET("/monitoring/grafana/dashboards/:name/versions", domain_grafana.ListDashboardVersions)
	v1.POST("/monitoring/grafana/dashboards/:name/rollback", domain_grafana.RollbackDashboard)
	v1.GET("/monitoring/grafana/templates", domain_grafana.ListTemplates)
	v1.GET("/monitoring/grafana/templates/:name", domain_grafana.GetTemplate)

	// Prometheus Rules
	v1.GET("/prometheus-rules", domain_prometheusrule.List)
	v1.GET("/prometheus-rules/:namespace/:name/yaml", domain_prometheusrule.GetYAML)
	v1.POST("/prometheus-rules/validate", domain_prometheusrule.ValidateRule)
	v1.GET("/prometheus-rules/templates", domain_prometheusrule.ListTemplates)
	v1.GET("/prometheus-rules/templates/:name", domain_prometheusrule.GetTemplate)
	v1.POST("/prometheus-rules/apply", domain_prometheusrule.ApplyTemplate)
}
