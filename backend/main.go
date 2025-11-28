package main

import (
	"log"
	"net/http"
	"os"
	"polardbx-ui-backend/pkg/api"

	// domain handlers - platform
	domain_alerts "polardbx-ui-backend/pkg/api/domain/platform/alerts/handler"
	domain_auth "polardbx-ui-backend/pkg/api/domain/platform/auth/handler"
	domain_clusterknobs "polardbx-ui-backend/pkg/api/domain/platform/clusterknobs/handler"
	domain_diagnostics "polardbx-ui-backend/pkg/api/domain/platform/diagnostics/handler"
	domain_grafana "polardbx-ui-backend/pkg/api/domain/platform/grafana/handler"
	domain_logcollector "polardbx-ui-backend/pkg/api/domain/platform/logcollector/handler"
	domain_logs "polardbx-ui-backend/pkg/api/domain/platform/logs/handler"
	domain_logservice "polardbx-ui-backend/pkg/api/domain/platform/logservice/handler"
	domain_logstrategy "polardbx-ui-backend/pkg/api/domain/platform/logstrategy/handler"
	domain_pod "polardbx-ui-backend/pkg/api/domain/platform/pod/handler"
	domain_prometheusrule "polardbx-ui-backend/pkg/api/domain/platform/prometheusrule/handler"
	domain_restore "polardbx-ui-backend/pkg/api/domain/platform/restore/handler"
	domain_settings "polardbx-ui-backend/pkg/api/domain/platform/settings/handler"
	domain_system "polardbx-ui-backend/pkg/api/domain/platform/system/handler"

	// domain handlers - monitoring (special location)
	domain_monitoring "polardbx-ui-backend/pkg/api/domain/monitoring"

	// domain handlers - business
	domain_pxc "polardbx-ui-backend/pkg/api/domain/polardbxclusters"
	domain_st "polardbx-ui-backend/pkg/api/domain/systemtasks"
	domain_xs "polardbx-ui-backend/pkg/api/domain/xstores"

	// router
	api_router "polardbx-ui-backend/pkg/api/router"

	"github.com/gin-gonic/gin"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Kubeconfig-B64")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// monitoringV2Enabled removed: v2 endpoints are now always enabled

func main() {
	// 设置controller-runtime日志器
	logger := zap.New(zap.UseDevMode(true))
	ctrllog.SetLogger(logger)

	r := gin.Default()

	// CORS middleware
	r.Use(CORSMiddleware())

	// A dummy ping endpoint
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// API v1 group
	v1 := r.Group("/api/v1")

	// Public endpoints that don't require kubeconfig authentication
	v1.GET("/image-registry/config", domain_settings.GetImageRegistryConfig)
	v1.PUT("/image-registry/config", domain_settings.UpdateImageRegistryConfig)
	v1.GET("/image-registry/presets", domain_settings.GetAvailableRegistries)
	v1.POST("/image-registry/test", domain_settings.TestImageRegistry)

	// The connect endpoint is special, it establishes the client for subsequent requests
	v1.POST("/connect", api.KubeconfigAuthMiddleware(), api.Connect)

	// JWT login endpoints (optional). When JWT_SECRET is set, protect routes with JWT.
	v1.POST("/auth/login", domain_auth.Login)
	v1.GET("/auth/me", domain_auth.Me)

	// All other endpoints require a valid kubeconfig
	v1.Use(api.KubeconfigAuthMiddleware())
	// Optional JWT middleware
	v1.Use(domain_auth.JWTAuthMiddleware())
	{
		// Cluster routes → domain/polardbxclusters
		v1.GET("/clusters", domain_pxc.List)
		v1.POST("/clusters", domain_pxc.Create)
		v1.POST("/clusters/:namespace/create", domain_pxc.CreateFromConfig)
		v1.GET("/clusters/:namespace/:name", domain_pxc.Get)
		v1.PATCH("/clusters/:namespace/:name/log-config/:nodeType", domain_pxc.UpdateLogConfig)
		v1.PATCH("/clusters/:namespace/:name/scale", domain_pxc.Scale)
		v1.PATCH("/clusters/:namespace/:name/upgrade", domain_pxc.Upgrade)
		v1.GET("/clusters/:namespace/:name/upgrade-plan", domain_pxc.GetUpgradePlan)
		v1.DELETE("/clusters/:namespace/:name", domain_pxc.Delete)
		v1.PUT("/clusters/:namespace/:name", domain_pxc.Update)
		v1.GET("/clusters/:namespace/:name/alerts-summary", domain_pxc.GetAlertsSummary)

		// Pods (保持原实现)
		v1.GET("/pods/:namespace/:name/exec", domain_pod.ExecWS)
		v1.GET("/pods/:namespace/:name", domain_pod.Get)
		v1.DELETE("/pods/:namespace/:name", domain_pod.Delete)
		v1.GET("/clusters/:namespace/:name/pods", domain_pod.ListForCluster)
		v1.GET("/pods", domain_pod.List)

		// Backups under cluster
		v1.GET("/clusters/:namespace/:name/backups", domain_pxc.ListBackups)
		v1.POST("/clusters/:namespace/:name/backups", domain_pxc.CreateBackup)
		// Root-level backup ops（保留原行为）
		v1.POST("/backups/validate", domain_pxc.ValidateBackup)
		v1.GET("/backups/:namespace/:name/stream", domain_pxc.StreamBackupEvents)
		v1.GET("/backups/:namespace/:name/metrics", domain_pxc.GetBackupMetrics)
		v1.DELETE("/backups/:namespace/:name", domain_pxc.DeleteBackup)
		v1.POST("/backups/:namespace/:name/force-delete", domain_pxc.ForceDeleteBackup)
		v1.GET("/backups/overview", domain_pxc.GetBackupOverview)
		v1.GET("/backups/binlog/metrics", domain_pxc.GetBinlogMetrics)

		// Settings for dashboard thresholds
		v1.GET("/settings/backup-dashboard", domain_settings.Get)
		v1.PUT("/settings/backup-dashboard", domain_settings.Update)
		// Generic settings alias (for frontend compatibility)
		v1.GET("/settings", domain_settings.Get)
		v1.PUT("/settings", domain_settings.Update)

		// XStore Backup alias routes (frontend expects /xstore-backups)
		v1.GET("/xstore-backups", domain_xs.ListBackups)
		v1.POST("/xstore-backups", domain_xs.CreateBackup)
		v1.GET("/xstore-backups/:namespace/:name", domain_xs.GetBackup)
		v1.PUT("/xstore-backups/:namespace/:name", domain_xs.UpdateBackup)
		v1.DELETE("/xstore-backups/:namespace/:name", domain_xs.DeleteBackup)
		v1.POST("/xstore-backups/:namespace/:name/force-delete", domain_xs.ForceDeleteBackup)

		// XStore Follower alias routes (frontend expects /xstore-followers)
		v1.GET("/xstore-followers", domain_xs.ListFollowers)
		v1.POST("/xstore-followers", domain_xs.CreateFollower)
		v1.GET("/xstore-followers/:namespace/:name", domain_xs.GetFollower)
		v1.PUT("/xstore-followers/:namespace/:name", domain_xs.UpdateFollower)
		v1.DELETE("/xstore-followers/:namespace/:name", domain_xs.DeleteFollower)

		// System module（前端依赖：/api/v1/system/*）
		v1.GET("/system/context", domain_system.ContextInfo)
		v1.GET("/system/namespaces", domain_system.ListNamespaces)

		// Direct routes for frontend compatibility
		v1.GET("/namespaces", domain_system.ListNamespaces)                              // Maps to /platform/system/namespaces
		v1.GET("/prometheus-rules", domain_prometheusrule.List)                          // PrometheusRule resources
		v1.GET("/prometheus-rules/:namespace/:name/yaml", domain_prometheusrule.GetYAML) // Get YAML
		v1.POST("/prometheus-rules/validate", domain_prometheusrule.ValidateRule)        // Validate YAML
		v1.GET("/prometheus-rules/templates", domain_prometheusrule.ListTemplates)
		v1.GET("/prometheus-rules/templates/:name", domain_prometheusrule.GetTemplate)
		v1.POST("/prometheus-rules/apply", domain_prometheusrule.ApplyTemplate)

		// Pre-change safety checklist
		v1.GET("/clusters/:namespace/:name/prechange-check", domain_pxc.GetPrechangeChecklist)
		v1.POST("/clusters/:namespace/:name/precheck", domain_pxc.Precheck)
		v1.GET("/alerts", domain_alerts.List)

		// Logs endpoint (pod logs)
		v1.GET("/logs/:namespace/:pod_name", domain_pod.GetLogs)

		// Parameter routes → domain
		v1.GET("/parameters", domain_pxc.ListParameters)
		v1.POST("/parameters", domain_pxc.CreateParameter)
		v1.GET("/parameters/:name", domain_pxc.GetParameter)
		v1.PUT("/parameters/:name", domain_pxc.UpdateParameter)
		v1.DELETE("/parameters/:name", domain_pxc.DeleteParameter)

		// Monitor CRD routes (统一到 domain_monitoring)
		v1.GET("/monitors", domain_monitoring.ListMonitors)
		v1.POST("/monitors", domain_monitoring.CreateMonitor)
		v1.GET("/monitors/:namespace/:name", domain_monitoring.GetMonitor)
		v1.PUT("/monitors/:namespace/:name", domain_monitoring.UpdateMonitor)
		v1.DELETE("/monitors/:namespace/:name", domain_monitoring.DeleteMonitor)

		// BackupSchedule routes → domain
		v1.GET("/backup-schedules", domain_pxc.ListSchedules)
		v1.POST("/backup-schedules", domain_pxc.CreateSchedule)
		v1.GET("/backup-schedules/:namespace/:name", domain_pxc.GetSchedule)
		v1.PUT("/backup-schedules/:namespace/:name", domain_pxc.UpdateSchedule)
		v1.DELETE("/backup-schedules/:namespace/:name", domain_pxc.DeleteSchedule)

		// Diagnostics routes (→ domain handlers)
		v1.POST("/diagnostics/:namespace/:cluster/start", domain_diagnostics.Start)
		v1.GET("/diagnostics/:namespace/:id/status", domain_diagnostics.GetStatus)
		v1.GET("/diagnostics/reports", domain_diagnostics.ListReports)
		v1.GET("/diagnostics/:namespace/:id/download", domain_diagnostics.Download)

		// ParameterTemplate routes → domain
		v1.GET("/parameter-templates", domain_pxc.ListTemplates)
		v1.POST("/parameter-templates", domain_pxc.CreateTemplate)
		v1.GET("/parameter-templates/:namespace/:name", domain_pxc.GetTemplate)
		v1.PUT("/parameter-templates/:namespace/:name", domain_pxc.UpdateTemplate)
		v1.DELETE("/parameter-templates/:namespace/:name", domain_pxc.DeleteTemplate)

		// ClusterKnobs routes → domain handlers (add direct aliases for frontend)
		v1.GET("/cluster-knobs", domain_clusterknobs.GetList)
		v1.POST("/cluster-knobs", domain_clusterknobs.Create)
		v1.GET("/cluster-knobs/:namespace/:name", domain_clusterknobs.Get)
		v1.PUT("/cluster-knobs/:namespace/:name", domain_clusterknobs.Update)
		v1.DELETE("/cluster-knobs/:namespace/:name", domain_clusterknobs.Delete)

		// SystemTask routes → domain/systemtasks
		v1.GET("/system-tasks", domain_st.List)
		v1.POST("/system-tasks", domain_st.Create)
		v1.GET("/system-tasks/:namespace/:name", domain_st.Get)
		v1.PUT("/system-tasks/:namespace/:name", domain_st.Update)
		v1.DELETE("/system-tasks/:namespace/:name", domain_st.Delete)

		// LogCollector routes（→ domain handlers）
		v1.GET("/log-collectors", domain_logcollector.List)
		v1.POST("/log-collectors", domain_logcollector.Create)
		v1.GET("/log-collectors/:namespace/:name", domain_logcollector.Get)
		v1.PUT("/log-collectors/:namespace/:name", domain_logcollector.Update)
		v1.DELETE("/log-collectors/:namespace/:name", domain_logcollector.Delete)
		v1.GET("/log-collectors/:namespace/pipeline", domain_logcollector.GetLogstashPipeline)
		v1.PUT("/log-collectors/:namespace/pipeline", domain_logcollector.UpdateLogstashPipeline)
		v1.GET("/log-collectors/:namespace/elastic-certs", domain_logcollector.GetElasticsearchCert)
		v1.PUT("/log-collectors/:namespace/elastic-certs", domain_logcollector.UpdateElasticsearchCert)
		v1.GET("/log-collectors/:namespace/:name/status", domain_logcollector.GetLogCollectorStatus)
		v1.GET("/log-collectors/:namespace/logstash/logs", domain_logcollector.StreamLogstashLogs)
		v1.POST("/log-collectors/:namespace/test", domain_logcollector.TestLogCollector)

		// Log Service / Strategy / Logs（→ domain handlers）
		v1.GET("/log-service/status", domain_logservice.Status)
		v1.GET("/log-strategies", domain_logstrategy.List)
		v1.POST("/log-strategies", domain_logstrategy.Create)
		v1.POST("/log-strategies/precheck", domain_logstrategy.Precheck)
		v1.GET("/log-strategies/apply-records", domain_logstrategy.ListApplyRecords) // New endpoint
		v1.GET("/log-strategies/:name", domain_logstrategy.Get)
		v1.PUT("/log-strategies/:name", domain_logstrategy.Update)
		v1.DELETE("/log-strategies/:name", domain_logstrategy.Delete)
		v1.POST("/log-strategies/:name/apply", domain_logstrategy.Apply)
		v1.POST("/log-strategies/test-connection", domain_logstrategy.TestConnection)
		// Logs Bootstrap (安装向导后端)
		v1.POST("/logs/bootstrap", domain_logs.Bootstrap)
		v1.GET("/logs/bootstrap/status", domain_logs.BootstrapStatus)
		v1.GET("/logs/bootstrap/logs", domain_logs.BootstrapLogs)
		v1.POST("/logs/query", domain_logs.Query)
		v1.GET("/logs/presets", domain_logs.Presets)
		v1.GET("/logs/presets/:pattern", domain_logs.PresetByPattern)

		// Monitoring v2 workflow endpoints (unified - all v1 logic migrated to v2)
		v1.POST("/monitoring/bootstrap", domain_monitoring.Bootstrap)
		v1.GET("/monitoring/bootstrap/status", domain_monitoring.BootstrapStatus)
		v1.GET("/monitoring/bootstrap/logs", domain_monitoring.GetBootstrapLogs)
		v1.GET("/monitoring/status", domain_monitoring.Status)
		v1.GET("/monitoring/preflight", domain_monitoring.DetectEnvironment) // preflight → detect
		v1.DELETE("/monitoring/uninstall", domain_monitoring.Uninstall)

		// v2 workflow endpoints
		v1.GET("/monitoring/detect", domain_monitoring.DetectEnvironment)
		v1.POST("/monitoring/plan", domain_monitoring.CreatePlan)
		v1.POST("/monitoring/install", domain_monitoring.StartInstallation)
		v1.GET("/monitoring/install/:sessionId/status", domain_monitoring.GetInstallStatus)
		v1.GET("/monitoring/install/:sessionId/logs", domain_monitoring.GetBootstrapLogs)
		v1.POST("/monitoring/install/:sessionId/retry", domain_monitoring.TriggerRetry)
		v1.POST("/monitoring/diagnose", domain_monitoring.DiagnoseFailure)
		v1.POST("/monitoring/auto-fix", domain_monitoring.ApplyAutoFix)

		v1.GET("/monitoring/grafana/config", domain_grafana.GetConfig)
		v1.PUT("/monitoring/grafana/config", domain_grafana.PutConfig)
		v1.POST("/monitoring/grafana/dashboards/sync", domain_grafana.SyncDashboards)
		v1.GET("/monitoring/grafana/dashboards", domain_grafana.ListDashboards)
		v1.GET("/monitoring/grafana/dashboards/:name/versions", domain_grafana.ListDashboardVersions)
		v1.POST("/monitoring/grafana/dashboards/:name/rollback", domain_grafana.RollbackDashboard)
		v1.GET("/monitoring/grafana/templates", domain_grafana.ListTemplates)
		v1.GET("/monitoring/grafana/templates/:name", domain_grafana.GetTemplate)

		// BackupBinlog routes（改由 domain handlers 接管，路径保持不变）
		v1.GET("/backup-binlogs", domain_pxc.ListBackupBinlogs)
		v1.POST("/backup-binlogs", domain_pxc.CreateBackupBinlog)
		v1.GET("/backup-binlogs/:namespace/:name", domain_pxc.GetBackupBinlog)
		v1.PUT("/backup-binlogs/:namespace/:name", domain_pxc.UpdateBackupBinlog)
		v1.DELETE("/backup-binlogs/:namespace/:name", domain_pxc.DeleteBackupBinlog)

		// HPFS sinks
		v1.GET("/hpfs/sinks", domain_pxc.ListHpfsSinks)
		v1.POST("/hpfs/sinks/validate", domain_pxc.ValidateHpfsSink)

		// Backup advice
		v1.GET("/clusters/:namespace/:name/backup-advice", domain_pxc.GetBackupAdvice)

		// Restore API routes（→ domain handlers）
		v1.POST("/clusters/:namespace/:name/restore", domain_restore.RestoreCluster)
		v1.POST("/clusters/:namespace/:name/pitr", domain_restore.InitiatePITR)
		v1.GET("/clusters/:namespace/:name/restore-status", domain_restore.GetRestoreStatus)
		v1.GET("/restore-jobs", domain_restore.ListJobs)
		v1.GET("/restore-jobs/:namespace/:name", domain_restore.GetJob)
		v1.DELETE("/restore-jobs/:namespace/:name", domain_restore.CancelJob)
	}

	// Register CRD-aligned alias routes (no behavior change)
	api_router.RegisterCRDAliasRoutes(v1)
	// Register domain entrance aliases (safe, non-conflicting)
	api_router.RegisterDomainRoutes(v1)

	// Grouped route logging for discoverability
	api_router.LogGroupedRoutes(r)

	if os.Getenv("GIN_LOG_ROUTES") == "1" {
		for _, route := range r.Routes() {
			log.Printf("Registered route: %s %s", route.Method, route.Path)
		}
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
