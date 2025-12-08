package router

import (
	domain_settings "polardbx-ui-backend/pkg/api/domain/platform/settings/handler"
	domain_pxc "polardbx-ui-backend/pkg/api/domain/polardbxclusters"
	"polardbx-ui-backend/pkg/api/routerutil"

	"github.com/gin-gonic/gin"
)

// RegisterBackupRoutes registers backup-related routes
func RegisterBackupRoutes(v1 *gin.RouterGroup) {
	// Backups under cluster
	v1.GET("/clusters/:namespace/:name/backups", domain_pxc.ListBackups)
	v1.POST("/clusters/:namespace/:name/backups", domain_pxc.CreateBackup)
	v1.GET("/clusters/:namespace/:name/backup-advice", domain_pxc.GetBackupAdvice)

	// Root-level backup operations
	v1.POST("/backups/validate", domain_pxc.ValidateBackup)
	v1.GET("/backups/:namespace/:name/stream", domain_pxc.StreamBackupEvents)
	v1.GET("/backups/:namespace/:name/metrics", domain_pxc.GetBackupMetrics)
	v1.DELETE("/backups/:namespace/:name", domain_pxc.DeleteBackup)
	v1.POST("/backups/:namespace/:name/force-delete", domain_pxc.ForceDeleteBackup)
	v1.GET("/backups/overview", domain_pxc.GetBackupOverview)
	v1.GET("/backups/cluster-state", domain_pxc.GetClusterBackupState)
	v1.GET("/backups/binlog/metrics", domain_pxc.GetBinlogMetrics)

	// Backup Schedules
	baseSchedules := "/backup-schedules"
	routerutil.RegisterCRUDWithItemPattern(v1, baseSchedules, baseSchedules+"/:namespace/:name", routerutil.CRUDHandlers{
		List:   domain_pxc.ListSchedules,
		Create: domain_pxc.CreateSchedule,
		Get:    domain_pxc.GetSchedule,
		Update: domain_pxc.UpdateSchedule,
		Delete: domain_pxc.DeleteSchedule,
	})

	// Backup Binlogs
	baseBinlogs := "/backup-binlogs"
	routerutil.RegisterCRUDWithItemPattern(v1, baseBinlogs, baseBinlogs+"/:namespace/:name", routerutil.CRUDHandlers{
		List:   domain_pxc.ListBackupBinlogs,
		Create: domain_pxc.CreateBackupBinlog,
		Get:    domain_pxc.GetBackupBinlog,
		Update: domain_pxc.UpdateBackupBinlog,
		Delete: domain_pxc.DeleteBackupBinlog,
	})

	// HPFS sinks
	v1.GET("/hpfs/sinks", domain_pxc.ListHpfsSinks)
	v1.POST("/hpfs/sinks/validate", domain_pxc.ValidateHpfsSink)

	// Settings for backup dashboard
	v1.GET("/settings/backup-dashboard", domain_settings.Get)
	v1.PUT("/settings/backup-dashboard", domain_settings.Update)
	v1.GET("/settings", domain_settings.Get)
	v1.PUT("/settings", domain_settings.Update)
}
