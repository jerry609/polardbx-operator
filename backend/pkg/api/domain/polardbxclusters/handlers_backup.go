package polardbxclusters

import (
	"polardbx-ui-backend/pkg/api/domain/polardbxclusters/services"

	"github.com/gin-gonic/gin"
)

// ListBackups lists backups for a given cluster by delegating to BackupService.
func ListBackups(c *gin.Context) { services.NewBackupService().List(c) }

// CreateBackup creates a new backup for a given cluster.
func CreateBackup(c *gin.Context) { services.NewBackupService().Create(c) }

// GetBackupAdvice returns backup-related recommendations for a cluster.
func GetBackupAdvice(c *gin.Context) { services.NewBackupService().GetBackupAdvice(c) }

// ValidateBackup validates a backup specification using a dry-run request.
func ValidateBackup(c *gin.Context) { services.NewBackupService().Validate(c) }

// StreamBackupEvents streams backup phase changes and events using server-sent events.
func StreamBackupEvents(c *gin.Context) { services.NewBackupService().StreamEvents(c) }

// GetBackupMetrics returns coarse-grained backup progress metrics for a backup.
func GetBackupMetrics(c *gin.Context) { services.NewBackupService().GetMetrics(c) }

// GetBinlogMetrics returns binlog backup related metrics for a cluster.
func GetBinlogMetrics(c *gin.Context) { services.NewBackupService().GetBinlogMetrics(c) }

// DeleteBackup deletes a PolarDB-X backup resource.
func DeleteBackup(c *gin.Context) { services.NewBackupService().Delete(c) }

// ForceDeleteBackup forcefully deletes a backup by removing finalizers and then deleting the resource.
func ForceDeleteBackup(c *gin.Context) { services.NewBackupService().ForceDelete(c) }

// ListSchedules lists backup schedules for a cluster.
func ListSchedules(c *gin.Context) { services.NewBackupScheduleService().List(c) }

// CreateSchedule creates a new backup schedule.
func CreateSchedule(c *gin.Context) { services.NewBackupScheduleService().Create(c) }

// GetSchedule gets a specific backup schedule.
func GetSchedule(c *gin.Context) { services.NewBackupScheduleService().Get(c) }

// UpdateSchedule updates an existing backup schedule.
func UpdateSchedule(c *gin.Context) { services.NewBackupScheduleService().Update(c) }

// DeleteSchedule deletes a backup schedule.
func DeleteSchedule(c *gin.Context) { services.NewBackupScheduleService().Delete(c) }

// GetScheduleNextRuns returns the next execution times for backup schedules.
func GetScheduleNextRuns(c *gin.Context) { services.NewBackupScheduleService().GetNextRuns(c) }

// GetBackupOverview returns an aggregated overview of backup status for a cluster.
func GetBackupOverview(c *gin.Context) { services.NewBackupService().GetOverview(c) }

// GetClusterBackupState returns a summarized backup state for the given cluster.
func GetClusterBackupState(c *gin.Context) { services.NewBackupService().GetClusterState(c) }

// ListHpfsSinks lists configured HPFS sinks from the HPFS configuration.
func ListHpfsSinks(c *gin.Context) { services.NewBackupService().ListHpfsSinks(c) }

// ValidateHpfsSink validates that a given HPFS sink exists in the HPFS configuration.
func ValidateHpfsSink(c *gin.Context) { services.NewBackupService().ValidateHpfsSink(c) }
