package xstores

import (
	"polardbx-ui-backend/pkg/api/domain/xstores/services"

	"github.com/gin-gonic/gin"
)

// ListBackups lists XStore backups in the given namespace.
// @Summary List XStore backups
// @Description List backups for XStore instances in the specified or default namespace.
// @Tags xstores, backups
// @Produce json
// @Param namespace query string false "Kubernetes namespace; defaults to 'default' when omitted"
// @Success 200 {array} k8srepo.XStoreBackupAlias "List of XStore backups"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func ListBackups(c *gin.Context) { services.NewBackupsService().List(c) }

// CreateBackup creates an XStore backup.
// @Summary Create XStore backup
// @Description Create a new backup for an XStore.
// @Tags xstores, backups
// @Accept json
// @Produce json
// @Param namespace query string false "Kubernetes namespace; defaults to 'default' when omitted"
// @Param body body k8srepo.XStoreBackupAlias true "XStore backup specification"
// @Success 201 {object} k8srepo.XStoreBackupAlias "Created XStore backup"
// @Failure 400 {object} map[string]any "Invalid request body"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func CreateBackup(c *gin.Context) { services.NewBackupsService().Create(c) }

// GetBackup gets a single XStore backup by namespace and name.
// @Summary Get XStore backup
// @Description Get a backup for a given XStore by namespace and name.
// @Tags xstores, backups
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the backup"
// @Param name path string true "Name of the backup"
// @Success 200 {object} k8srepo.XStoreBackupAlias "XStore backup"
// @Failure 404 {object} map[string]any "Backup not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func GetBackup(c *gin.Context) { services.NewBackupsService().Get(c) }

// UpdateBackup updates an existing XStore backup.
// @Summary Update XStore backup
// @Description Update an existing XStore backup resource.
// @Tags xstores, backups
// @Accept json
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the backup"
// @Param name path string true "Name of the backup"
// @Param body body k8srepo.XStoreBackupAlias true "Updated XStore backup specification"
// @Success 200 {object} k8srepo.XStoreBackupAlias "Updated XStore backup"
// @Failure 400 {object} map[string]any "Invalid request body"
// @Failure 404 {object} map[string]any "Backup not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func UpdateBackup(c *gin.Context) { services.NewBackupsService().Update(c) }

// DeleteBackup deletes an XStore backup.
// @Summary Delete XStore backup
// @Description Delete an XStore backup by namespace and name.
// @Tags xstores, backups
// @Param namespace path string true "Kubernetes namespace of the backup"
// @Param name path string true "Name of the backup"
// @Success 200 {object} map[string]any "Deletion confirmation"
// @Failure 404 {object} map[string]any "Backup not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func DeleteBackup(c *gin.Context) { services.NewBackupsService().Delete(c) }

// ForceDeleteBackup forcefully deletes an XStore backup by removing protection and then deleting it.
// @Summary Force delete XStore backup
// @Description Force delete an XStore backup resource by bypassing safeguards.
// @Tags xstores, backups
// @Param namespace path string true "Kubernetes namespace of the backup"
// @Param name path string true "Name of the backup"
// @Success 202 {object} map[string]any "Deletion requested"
// @Failure 404 {object} map[string]any "Backup not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func ForceDeleteBackup(c *gin.Context) { services.NewBackupsService().ForceDelete(c) }

// GetBackupRemoteInfo gets remote storage information for an XStore backup.
// @Summary Get XStore backup remote info
// @Description Get remote storage information for an XStore backup (e.g., remote path or location).
// @Tags xstores, backups
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the backup"
// @Param name path string true "Name of the backup"
// @Success 200 {object} map[string]any "Remote backup information"
// @Failure 404 {object} map[string]any "Backup not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func GetBackupRemoteInfo(c *gin.Context) { services.NewBackupsService().RemoteInfo(c) }
