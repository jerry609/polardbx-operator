package main

import (
	"log"
	"net/http"
	"polardbx-ui-backend/pkg/api"
	api_alerts "polardbx-ui-backend/pkg/api/alerts"
	api_auth "polardbx-ui-backend/pkg/api/auth"
	api_backup "polardbx-ui-backend/pkg/api/backup"
	api_backupbinlog "polardbx-ui-backend/pkg/api/backupbinlog"
	api_cluster "polardbx-ui-backend/pkg/api/cluster"
	api_clusterknobs "polardbx-ui-backend/pkg/api/clusterknobs"
	api_diagnostics "polardbx-ui-backend/pkg/api/diagnostics"
	api_grafana "polardbx-ui-backend/pkg/api/grafana"
	api_logcollector "polardbx-ui-backend/pkg/api/logcollector"
	api_monitor "polardbx-ui-backend/pkg/api/monitor"
	api_monitoring "polardbx-ui-backend/pkg/api/monitoring"
	api_parameters "polardbx-ui-backend/pkg/api/parameters"
	api_pod "polardbx-ui-backend/pkg/api/pod"
	api_prechange "polardbx-ui-backend/pkg/api/prechange"
	api_restore "polardbx-ui-backend/pkg/api/restore"
	api_settings "polardbx-ui-backend/pkg/api/settings"
	api_system "polardbx-ui-backend/pkg/api/system"
	api_systemtask "polardbx-ui-backend/pkg/api/systemtask"
	api_xstore "polardbx-ui-backend/pkg/api/xstore"

	api_logservice "polardbx-ui-backend/pkg/api/logservice"
	api_logstrategy "polardbx-ui-backend/pkg/api/logstrategy"

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

	// The connect endpoint is special, it establishes the client for subsequent requests
	v1.POST("/connect", api.Connect)

	// JWT login endpoints (optional). When JWT_SECRET is set, protect routes with JWT.
	v1.POST("/auth/login", api_auth.Login)
	v1.GET("/auth/me", api_auth.Me)

	// All other endpoints require a valid kubeconfig
	v1.Use(api.KubeconfigAuthMiddleware())
	// Optional JWT middleware
	v1.Use(api_auth.JWTAuthMiddleware())
	{
		v1.GET("/clusters", api_cluster.List)
		v1.POST("/clusters", api_cluster.Create)
		v1.POST("/clusters/:namespace/create", api_cluster.CreateFromConfig)
		v1.GET("/clusters/:namespace/:name", api_cluster.Get)

		// 集群日志配置API
		v1.PATCH("/clusters/:namespace/:name/log-config/:nodeType", api_cluster.UpdateLogConfig)

		// 集群扩缩容API
		v1.PATCH("/clusters/:namespace/:name/scale", api_cluster.Scale)

		// 集群升级API
		v1.PATCH("/clusters/:namespace/:name/upgrade", api_cluster.Upgrade)
		v1.DELETE("/clusters/:namespace/:name", api_cluster.Delete)
		v1.PUT("/clusters/:namespace/:name", api_cluster.Update)
		v1.GET("/clusters/:namespace/:name/alerts-summary", api_cluster.GetAlertsSummary)
		v1.GET("/pods/:namespace/:name/exec", api_pod.ExecWS)
		v1.GET("/pods/:namespace/:name", api_pod.Get)
		v1.DELETE("/pods/:namespace/:name", api_pod.Delete)
		v1.GET("/clusters/:namespace/:name/pods", api_pod.ListForCluster)
		// List pods by namespace only (no cluster filter)
		v1.GET("/pods", api_pod.List)
		v1.GET("/clusters/:namespace/:name/backups", api_backup.List)
		v1.POST("/clusters/:namespace/:name/backups", api_backup.Create)
		// Dry-run validation (does not persist)
		v1.POST("/backups/validate", api_backup.Validate)
		// SSE for backup events (progress/phase skeleton)
		v1.GET("/backups/:namespace/:name/stream", api_backup.StreamEvents)
		v1.GET("/backups/:namespace/:name/metrics", api_backup.GetMetrics)
		v1.DELETE("/backups/:namespace/:name", api_backup.Delete)
		// Overview & Aggregations
		v1.GET("/backups/overview", api_backup.GetBackupOverview)
		v1.GET("/backups/binlog/metrics", api_backup.GetBinlogMetrics)
		v1.GET("/backup-schedules/next-run", api_backup.GetNextRuns)

		// Settings for dashboard thresholds
		v1.GET("/settings/backup-dashboard", api_settings.Get)
		v1.PUT("/settings/backup-dashboard", api_settings.Update)

		// Pre-change safety checklist
		v1.GET("/clusters/:namespace/:name/prechange-check", api_prechange.GetPrechangeChecklist)
		v1.POST("/clusters/:namespace/:name/precheck-systemtask", api_prechange.CreatePrecheckSystemTask)
		v1.GET("/alerts", api_alerts.List)

		// Log endpoint
		v1.GET("/logs/:namespace/:pod_name", api_pod.GetLogs)

		// Parameter routes
		v1.GET("/parameters", api_parameters.List)
		v1.POST("/parameters", api_parameters.Create)
		v1.GET("/parameters/:name", api_parameters.Get)
		v1.PUT("/parameters/:name", api_parameters.Update)
		v1.DELETE("/parameters/:name", api_parameters.Delete)

		// XStore routes
		v1.GET("/xstores", api_xstore.List)
		v1.POST("/xstores", api_xstore.Create)
		v1.GET("/xstores/:namespace/:name", api_xstore.Get)
		// List pods under specific XStore (namespace + xstoreName)
		v1.GET("/xstores/:namespace/:name/pods", api_xstore.ListPods)
		v1.PUT("/xstores/:namespace/:name", api_xstore.Update)
		v1.DELETE("/xstores/:namespace/:name", api_xstore.Delete)

		// Monitor routes
		v1.GET("/monitors", api_monitor.List)
		v1.POST("/monitors", api_monitor.Create)
		v1.GET("/monitors/:namespace/:name", api_monitor.Get)
		v1.PUT("/monitors/:namespace/:name", api_monitor.Update)
		v1.DELETE("/monitors/:namespace/:name", api_monitor.Delete)

		// BackupSchedule routes (modularized)
		v1.GET("/backup-schedules", api_backup.ListSchedules)
		v1.POST("/backup-schedules", api_backup.CreateSchedule)
		v1.GET("/backup-schedules/:namespace/:name", api_backup.GetSchedule)
		v1.PUT("/backup-schedules/:namespace/:name", api_backup.UpdateSchedule)
		v1.DELETE("/backup-schedules/:namespace/:name", api_backup.DeleteSchedule)
		// Force delete backup endpoint removed from root; use specific modules if needed
		v1.GET("/backup-schedules/next-run", api_backup.GetNextRuns)

		// Diagnostics routes (modularized)
		v1.POST("/diagnostics/:namespace/:cluster/start", api_diagnostics.Start)
		v1.GET("/diagnostics/:namespace/:id/status", api_diagnostics.GetStatus)
		v1.GET("/diagnostics/reports", api_diagnostics.ListReports)
		v1.GET("/diagnostics/:namespace/:id/download", api_diagnostics.Download)

		// ParameterTemplate routes
		v1.GET("/parameter-templates", api_parameters.ListTemplates)
		v1.POST("/parameter-templates", api_parameters.CreateTemplate)
		v1.GET("/parameter-templates/:namespace/:name", api_parameters.GetTemplate)
		v1.PUT("/parameter-templates/:namespace/:name", api_parameters.UpdateTemplate)
		v1.DELETE("/parameter-templates/:namespace/:name", api_parameters.DeleteTemplate)

		// SystemTask routes (modularized)
		v1.GET("/system-tasks", api_systemtask.List)
		v1.POST("/system-tasks", api_systemtask.Create)
		v1.GET("/system-tasks/:namespace/:name", api_systemtask.Get)
		v1.PUT("/system-tasks/:namespace/:name", api_systemtask.Update)
		v1.DELETE("/system-tasks/:namespace/:name", api_systemtask.Delete)

		// LogCollector routes
		v1.GET("/log-collectors", api_logcollector.List)
		v1.POST("/log-collectors", api_logcollector.Create)
		v1.GET("/log-collectors/:namespace/:name", api_logcollector.Get)
		v1.PUT("/log-collectors/:namespace/:name", api_logcollector.Update)
		v1.DELETE("/log-collectors/:namespace/:name", api_logcollector.Delete)

		// LogCollector pipeline/certs/status/log streaming/test routes
		v1.GET("/log-collectors/:namespace/pipeline", api_logcollector.GetLogstashPipeline)
		v1.PUT("/log-collectors/:namespace/pipeline", api_logcollector.UpdateLogstashPipeline)
		v1.GET("/log-collectors/:namespace/elastic-certs", api_logcollector.GetElasticsearchCert)
		v1.PUT("/log-collectors/:namespace/elastic-certs", api_logcollector.UpdateElasticsearchCert)
		v1.GET("/log-collectors/:namespace/:name/status", api_logcollector.GetLogCollectorStatus)
		v1.GET("/log-collectors/:namespace/logstash/logs", api_logcollector.StreamLogstashLogs)
		v1.POST("/log-collectors/:namespace/test", api_logcollector.TestLogCollector)

		// Log Service module
		v1.GET("/log-service/status", api_logservice.Status)

		// Log Strategy module (MVP: list/get/create/update/delete/precheck/apply)
		v1.GET("/log-strategies", api_logstrategy.List)
		v1.POST("/log-strategies", api_logstrategy.Create)
		v1.POST("/log-strategies/precheck", api_logstrategy.Precheck)
		v1.GET("/log-strategies/:name", api_logstrategy.Get)
		v1.PUT("/log-strategies/:name", api_logstrategy.Update)
		v1.DELETE("/log-strategies/:name", api_logstrategy.Delete)
		v1.POST("/log-strategies/:name/apply", api_logstrategy.Apply)

		// Monitoring module (modularized)
		v1.POST("/monitoring/bootstrap", api_monitoring.Bootstrap)
		v1.GET("/monitoring/status", api_monitoring.Status)
		v1.GET("/monitoring/preflight", api_monitoring.Preflight)
		v1.DELETE("/monitoring/uninstall", api_monitoring.Uninstall)

		// Grafana module
		v1.GET("/monitoring/grafana/config", api_grafana.GetConfig)
		v1.PUT("/monitoring/grafana/config", api_grafana.PutConfig)
		v1.POST("/monitoring/grafana/dashboards/sync", api_grafana.SyncDashboards)
		v1.GET("/monitoring/grafana/dashboards", api_grafana.ListDashboards)
		v1.GET("/monitoring/grafana/dashboards/:name/versions", api_grafana.ListDashboardVersions)
		v1.POST("/monitoring/grafana/dashboards/:name/rollback", api_grafana.RollbackDashboard)

		// Alerts module
		v1.GET("/alerts/profiles", api_alerts.ListProfiles)
		v1.POST("/alerts/profiles", api_alerts.CreateProfile)
		v1.GET("/alerts/profiles/:name", api_alerts.GetProfile)
		v1.PUT("/alerts/profiles/:name", api_alerts.UpdateProfile)
		v1.DELETE("/alerts/profiles/:name", api_alerts.DeleteProfile)
		v1.POST("/alerts/profiles/dry-run", api_alerts.DryRunProfile)
		v1.GET("/alerts/routes", api_alerts.GetRoutes)
		v1.PUT("/alerts/routes", api_alerts.PutRoutes)
		v1.GET("/alerts/silences", api_alerts.ListSilences)
		v1.POST("/alerts/silences", api_alerts.CreateSilence)
		v1.DELETE("/alerts/silences/:id", api_alerts.DeleteSilence)
		v1.POST("/alerts/test", api_alerts.TestAlert)

		// BackupBinlog routes
		v1.GET("/backup-binlogs", api_backupbinlog.List)
		v1.POST("/backup-binlogs", api_backupbinlog.Create)
		v1.GET("/backup-binlogs/:namespace/:name", api_backupbinlog.Get)
		v1.PUT("/backup-binlogs/:namespace/:name", api_backupbinlog.Update)
		v1.DELETE("/backup-binlogs/:namespace/:name", api_backupbinlog.Delete)

		// HPFS sinks
		v1.GET("/hpfs/sinks", api_backup.ListHpfsSinks)
		v1.POST("/hpfs/sinks/validate", api_backup.ValidateHpfsSink)

		// Backup advice
		v1.GET("/clusters/:namespace/:name/backup-advice", api_backup.GetBackupAdvice)

		// 🚨 CRITICAL RECOVERY API ROUTES - Based on document analysis
		// These are the most critical missing APIs identified in the completeness report
		// PolarDB-X Operator has complete recovery capabilities but Management Platform had ZERO recovery support

		// Cluster restore operations
		v1.POST("/clusters/:namespace/:name/restore", api_restore.RestoreCluster)
		v1.POST("/clusters/:namespace/:name/pitr", api_restore.InitiatePITR)
		v1.GET("/clusters/:namespace/:name/restore-status", api_restore.GetRestoreStatus)

		// Restore job management
		v1.GET("/restore-jobs", api_restore.ListJobs)
		v1.GET("/restore-jobs/:namespace/:name", api_restore.GetJob)
		v1.DELETE("/restore-jobs/:namespace/:name", api_restore.CancelJob)

		// 🚨 HIGH PRIORITY MISSING FUNCTIONALITY: XStoreFollower API Routes
		// XStoreFollower for DN replica fault recovery (备库重搭) - identified as critical missing feature
		v1.GET("/xstore-followers", api_xstore.ListFollowers)
		v1.POST("/xstore-followers", api_xstore.CreateFollower)
		v1.GET("/xstore-followers/:namespace/:name", api_xstore.GetFollower)
		v1.PUT("/xstore-followers/:namespace/:name", api_xstore.UpdateFollower)
		v1.DELETE("/xstore-followers/:namespace/:name", api_xstore.DeleteFollower)

		// Rebuild wrappers aligned with docs terminology
		v1.POST("/xstore-rebuild/:namespace/:name/logger", api_xstore.RebuildLogger)
		v1.POST("/xstore-rebuild/:namespace/:name/learner", api_xstore.RebuildLearner)
		v1.POST("/xstore-rebuild/:namespace/:name/auto", func(c *gin.Context) { c.Set("autoMode", true); api_xstore.AutoRebuild(c) })
		v1.GET("/xstore-rebuild/:namespace/:name/status", api_xstore.RebuildStatus)

		// XStoreBackup API Routes - Complete the unified backup module
		v1.GET("/xstore-backups", api_xstore.ListBackups)
		v1.POST("/xstore-backups", api_xstore.CreateBackup)
		v1.GET("/xstore-backups/:namespace/:name", api_xstore.GetBackup)
		v1.PUT("/xstore-backups/:namespace/:name", api_xstore.UpdateBackup)
		v1.DELETE("/xstore-backups/:namespace/:name", api_xstore.DeleteBackup)
		v1.POST("/xstore-backups/:namespace/:name/force-delete", api_xstore.ForceDeleteBackup)
		v1.GET("/xstore-backups/:namespace/:name/remote-info", api_xstore.GetBackupRemoteInfo)

		// PolarDBXClusterKnobs API Routes - Performance tuning management (modularized)
		v1.GET("/cluster-knobs", api_clusterknobs.GetList)
		v1.POST("/cluster-knobs", api_clusterknobs.Create)
		v1.GET("/cluster-knobs/:namespace/:name", api_clusterknobs.Get)
		v1.PUT("/cluster-knobs/:namespace/:name", api_clusterknobs.Update)
		v1.DELETE("/cluster-knobs/:namespace/:name", api_clusterknobs.Delete)

		// System module
		v1.GET("/system/context", api_system.ContextInfo)
		v1.GET("/system/namespaces", api_system.ListNamespaces)
	}

	for _, route := range r.Routes() {
		log.Printf("Registered route: %s %s", route.Method, route.Path)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
