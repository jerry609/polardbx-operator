package router

import (
	domain_settings "polardbx-ui-backend/pkg/api/domain/platform/settings/handler"
	domain_pxc "polardbx-ui-backend/pkg/api/domain/polardbxclusters"

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
	v1.GET("/backup-schedules", domain_pxc.ListSchedules)
	v1.POST("/backup-schedules", domain_pxc.CreateSchedule)
	v1.GET("/backup-schedules/:namespace/:name", domain_pxc.GetSchedule)
	v1.PUT("/backup-schedules/:namespace/:name", domain_pxc.UpdateSchedule)
	v1.DELETE("/backup-schedules/:namespace/:name", domain_pxc.DeleteSchedule)

	// Backup Binlogs
	v1.GET("/backup-binlogs", domain_pxc.ListBackupBinlogs)
	v1.POST("/backup-binlogs", domain_pxc.CreateBackupBinlog)
	v1.GET("/backup-binlogs/:namespace/:name", domain_pxc.GetBackupBinlog)
	v1.PUT("/backup-binlogs/:namespace/:name", domain_pxc.UpdateBackupBinlog)
	v1.DELETE("/backup-binlogs/:namespace/:name", domain_pxc.DeleteBackupBinlog)

	// HPFS sinks
	v1.GET("/hpfs/sinks", domain_pxc.ListHpfsSinks)
	v1.POST("/hpfs/sinks/validate", domain_pxc.ValidateHpfsSink)

	// Settings for backup dashboard
	v1.GET("/settings/backup-dashboard", domain_settings.Get)
	v1.PUT("/settings/backup-dashboard", domain_settings.Update)
	v1.GET("/settings", domain_settings.Get)
	v1.PUT("/settings", domain_settings.Update)
}
