package polardbxclusters

import (
	"polardbx-ui-backend/pkg/api/domain/polardbxclusters/services"

	"github.com/gin-gonic/gin"
)

// ListBackups lists backups for a given cluster by delegating to BackupService.
// @Summary List backups for a cluster
// @Description List all backups that belong to the specified PolarDB-X cluster.
// @Tags polardbxclusters, backups
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the cluster"
// @Param name path string true "Name of the PolarDB-X cluster"
// @Success 200 {array} polardbxv1.PolarDBXBackup "List of backups"
// @Failure 500 {object} map[string]any "Internal server error"
func ListBackups(c *gin.Context) { services.NewBackupService().List(c) }

// CreateBackup creates a new backup for a given cluster.
// @Summary Create backup
// @Description Create a PolarDB-X backup resource bound to the specified cluster.
// @Tags polardbxclusters, backups
// @Accept json
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the cluster"
// @Param name path string true "Name of the PolarDB-X cluster"
// @Param body body polardbxv1.PolarDBXBackup true "Backup specification"
// @Success 201 {object} polardbxv1.PolarDBXBackup "Created backup"
// @Failure 400 {object} map[string]any "Invalid request body or parameters"
// @Failure 500 {object} map[string]any "Internal server error"
func CreateBackup(c *gin.Context) { services.NewBackupService().Create(c) }

// GetBackupAdvice returns backup-related recommendations for a cluster.
// @Summary Get backup advice
// @Description Get recommendations and advice for backup strategies of a given cluster.
// @Tags polardbxclusters, backups
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the cluster"
// @Param name path string true "Name of the PolarDB-X cluster"
// @Success 200 {object} map[string]any "Advice payload"
// @Failure 500 {object} map[string]any "Internal server error"
func GetBackupAdvice(c *gin.Context) { services.NewBackupService().GetBackupAdvice(c) }

// ValidateBackup validates a backup specification using a dry-run request.
// @Summary Validate backup
// @Description Validate a backup specification by performing a dry-run create against Kubernetes.
// @Tags polardbxclusters, backups
// @Accept json
// @Produce json
// @Param namespace query string false "Target namespace for validation; falls back to backup namespace or 'default'"
// @Param body body polardbxv1.PolarDBXBackup true "Backup specification to validate"
// @Success 200 {object} map[string]bool "Validation result (valid: true)"
// @Failure 400 {object} map[string]any "Invalid request body or parameters"
// @Failure 422 {object} map[string]any "Backup validation failed"
// @Failure 500 {object} map[string]any "Internal server error"
func ValidateBackup(c *gin.Context) { services.NewBackupService().Validate(c) }

// StreamBackupEvents streams backup phase changes and events using server-sent events.
// @Summary Stream backup events
// @Description Stream backup phase changes and events as Server-Sent Events (SSE).
// @Tags polardbxclusters, backups
// @Produce text/event-stream
// @Param namespace path string true "Kubernetes namespace of the backup"
// @Param name path string true "Name of the PolarDB-X backup"
// @Success 200 {string} string "SSE stream of backup events"
// @Failure 404 {object} map[string]any "Backup not found"
// @Failure 500 {object} map[string]any "Internal server error or streaming unsupported"
func StreamBackupEvents(c *gin.Context) { services.NewBackupService().StreamEvents(c) }

// GetBackupMetrics returns coarse-grained backup progress metrics for a backup.
// @Summary Get backup metrics
// @Description Get coarse-grained backup progress and aggregated child XStore backup status.
// @Tags polardbxclusters, backups
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the backup"
// @Param name path string true "Name of the PolarDB-X backup"
// @Success 200 {object} map[string]any "Progress metrics and child backup stats"
// @Failure 404 {object} map[string]any "Backup not found"
// @Failure 500 {object} map[string]any "Internal server error"
func GetBackupMetrics(c *gin.Context) { services.NewBackupService().GetMetrics(c) }

// GetBinlogMetrics returns binlog backup related metrics for a cluster.
// @Summary Get binlog backup metrics
// @Description Get metrics related to binlog backup for the specified cluster.
// @Tags polardbxclusters, backups
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the cluster"
// @Param name path string true "Name of the PolarDB-X cluster"
// @Success 200 {object} map[string]any "Binlog backup metrics"
// @Failure 404 {object} map[string]any "Cluster or metrics not found"
// @Failure 500 {object} map[string]any "Internal server error"
func GetBinlogMetrics(c *gin.Context) { services.NewBackupService().GetBinlogMetrics(c) }

// DeleteBackup deletes a PolarDB-X backup resource.
// @Summary Delete backup
// @Description Delete the specified PolarDB-X backup resource.
// @Tags polardbxclusters, backups
// @Param namespace path string true "Kubernetes namespace of the backup"
// @Param name path string true "Name of the PolarDB-X backup"
// @Success 204 "Backup deleted successfully"
// @Failure 404 {object} map[string]any "Backup not found"
// @Failure 500 {object} map[string]any "Internal server error"
func DeleteBackup(c *gin.Context) { services.NewBackupService().Delete(c) }

// ForceDeleteBackup forcefully deletes a backup by removing finalizers and then deleting the resource.
// @Summary Force delete backup
// @Description Remove finalizers from the backup and then delete it, bypassing normal protection.
// @Tags polardbxclusters, backups
// @Param namespace path string true "Kubernetes namespace of the backup"
// @Param name path string true "Name of the PolarDB-X backup"
// @Success 202 {object} map[string]any "Deletion requested"
// @Failure 404 {object} map[string]any "Backup not found"
// @Failure 500 {object} map[string]any "Internal server error"
func ForceDeleteBackup(c *gin.Context) { services.NewBackupService().ForceDelete(c) }

// ListSchedules lists backup schedules for a cluster or namespace.
// @Summary List backup schedules
// @Description List backup schedules, optionally filtered by namespace and name.
// @Tags polardbxclusters, backups, schedules
// @Produce json
// @Param namespace query string false "Filter by namespace"
// @Param name query string false "Filter by schedule name"
// @Success 200 {array} polardbxv1.PolarDBXBackupSchedule "List of backup schedules"
// @Failure 500 {object} map[string]any "Internal server error"
func ListSchedules(c *gin.Context) { services.NewBackupScheduleService().List(c) }

// CreateSchedule creates a new backup schedule.
// @Summary Create backup schedule
// @Description Create a new PolarDB-X backup schedule resource.
// @Tags polardbxclusters, backups, schedules
// @Accept json
// @Produce json
// @Param body body polardbxv1.PolarDBXBackupSchedule true "Backup schedule specification"
// @Success 201 {object} polardbxv1.PolarDBXBackupSchedule "Created backup schedule"
// @Failure 400 {object} map[string]any "Invalid request body or parameters"
// @Failure 500 {object} map[string]any "Internal server error"
func CreateSchedule(c *gin.Context) { services.NewBackupScheduleService().Create(c) }

// GetSchedule gets a specific backup schedule.
// @Summary Get backup schedule
// @Description Get a backup schedule by namespace and name.
// @Tags polardbxclusters, backups, schedules
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the schedule"
// @Param name path string true "Name of the backup schedule"
// @Success 200 {object} polardbxv1.PolarDBXBackupSchedule "Backup schedule"
// @Failure 404 {object} map[string]any "Schedule not found"
// @Failure 500 {object} map[string]any "Internal server error"
func GetSchedule(c *gin.Context) { services.NewBackupScheduleService().Get(c) }

// UpdateSchedule updates an existing backup schedule.
// @Summary Update backup schedule
// @Description Update an existing PolarDB-X backup schedule resource.
// @Tags polardbxclusters, backups, schedules
// @Accept json
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the schedule"
// @Param name path string true "Name of the backup schedule"
// @Param body body polardbxv1.PolarDBXBackupSchedule true "Updated schedule specification"
// @Success 200 {object} polardbxv1.PolarDBXBackupSchedule "Updated backup schedule"
// @Failure 400 {object} map[string]any "Invalid request body or parameters"
// @Failure 404 {object} map[string]any "Schedule not found"
// @Failure 500 {object} map[string]any "Internal server error"
func UpdateSchedule(c *gin.Context) { services.NewBackupScheduleService().Update(c) }

// DeleteSchedule deletes a backup schedule.
// @Summary Delete backup schedule
// @Description Delete a backup schedule by namespace and name.
// @Tags polardbxclusters, backups, schedules
// @Param namespace path string true "Kubernetes namespace of the schedule"
// @Param name path string true "Name of the backup schedule"
// @Success 204 "Backup schedule deleted successfully"
// @Failure 404 {object} map[string]any "Schedule not found"
// @Failure 500 {object} map[string]any "Internal server error"
func DeleteSchedule(c *gin.Context) { services.NewBackupScheduleService().Delete(c) }

// GetScheduleNextRuns returns the next execution times for backup schedules.
// @Summary Get backup schedule next runs
// @Description Get upcoming execution times for backup schedules, optionally filtered by namespace.
// @Tags polardbxclusters, backups, schedules
// @Produce json
// @Param namespace query string false "Filter by namespace"
// @Success 200 {object} map[string]any "Next-run aggregation"
// @Failure 500 {object} map[string]any "Internal server error"
func GetScheduleNextRuns(c *gin.Context) { services.NewBackupScheduleService().GetNextRuns(c) }

// GetBackupOverview returns an aggregated overview of backup status for a cluster.
// @Summary Get backup overview
// @Description Get high-level backup overview information for the specified cluster.
// @Tags polardbxclusters, backups
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the cluster"
// @Param name path string true "Name of the PolarDB-X cluster"
// @Success 200 {object} map[string]any "Backup overview"
// @Failure 404 {object} map[string]any "Cluster not found"
// @Failure 500 {object} map[string]any "Internal server error"
func GetBackupOverview(c *gin.Context) { services.NewBackupService().GetOverview(c) }

// GetClusterBackupState returns a summarized backup state for the given cluster.
// @Summary Get cluster backup state
// @Description Get summarized backup state for the specified cluster, for use in UI.
// @Tags polardbxclusters, backups
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the cluster"
// @Param name path string true "Name of the PolarDB-X cluster"
// @Success 200 {object} map[string]any "Cluster backup state"
// @Failure 404 {object} map[string]any "Cluster not found"
// @Failure 500 {object} map[string]any "Internal server error"
func GetClusterBackupState(c *gin.Context) { services.NewBackupService().GetClusterState(c) }

// ListHpfsSinks lists configured HPFS sinks from the HPFS configuration.
// @Summary List HPFS sinks
// @Description List sinks configured in the HPFS configuration ConfigMap.
// @Tags polardbxclusters, hpfs
// @Produce json
// @Param systemNamespace query string false "Namespace of the HPFS config ConfigMap; defaults to polardbx-operator-system"
// @Success 200 {object} map[string]any "HPFS sink definitions"
// @Failure 500 {object} map[string]any "Internal server error or invalid config"
func ListHpfsSinks(c *gin.Context) { services.NewBackupService().ListHpfsSinks(c) }

// ValidateHpfsSink validates that a given HPFS sink exists in the HPFS configuration.
// @Summary Validate HPFS sink
// @Description Validate that an HPFS sink with given name and type exists in HPFS config.
// @Tags polardbxclusters, hpfs
// @Accept json
// @Produce json
// @Param systemNamespace query string false "Namespace of the HPFS config ConfigMap; defaults to polardbx-operator-system"
// @Param body body struct{name string `json:"name"`; type string `json:"type"`} true "Sink name and type"
// @Success 200 {object} map[string]any "Validation status and message"
// @Failure 400 {object} map[string]any "Invalid request body"
// @Failure 500 {object} map[string]any "Internal server error"
func ValidateHpfsSink(c *gin.Context) { services.NewBackupService().ValidateHpfsSink(c) }
