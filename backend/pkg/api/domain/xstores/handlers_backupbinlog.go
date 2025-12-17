package xstores

import (
	"polardbx-ui-backend/pkg/api/domain/xstores/services"

	"github.com/gin-gonic/gin"
)

// ListBackupBinlogs lists binlog backups for XStores.
// @Summary List XStore binlog backups
// @Description List binlog backups for XStore instances in the specified or default namespace.
// @Tags xstores, backup-binlogs
// @Produce json
// @Param namespace query string false "Kubernetes namespace; defaults to 'default' when omitted"
// @Success 200 {array} polardbxv1.XStoreBackupBinlog "List of XStore binlog backups"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func ListBackupBinlogs(c *gin.Context) { services.NewBackupBinlogService().List(c) }

// CreateBackupBinlog creates a binlog backup for an XStore.
// @Summary Create XStore binlog backup
// @Description Create a new binlog backup for an XStore.
// @Tags xstores, backup-binlogs
// @Accept json
// @Produce json
// @Param namespace query string false "Kubernetes namespace; defaults to 'default' when omitted"
// @Param body body polardbxv1.XStoreBackupBinlog true "XStore binlog backup specification"
// @Success 201 {object} polardbxv1.XStoreBackupBinlog "Created XStore binlog backup"
// @Failure 400 {object} map[string]any "Invalid request body"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func CreateBackupBinlog(c *gin.Context) { services.NewBackupBinlogService().Create(c) }

// GetBackupBinlog gets a single XStore binlog backup by namespace and name.
// @Summary Get XStore binlog backup
// @Description Get a binlog backup by namespace and name.
// @Tags xstores, backup-binlogs
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the binlog backup"
// @Param name path string true "Name of the binlog backup"
// @Success 200 {object} polardbxv1.XStoreBackupBinlog "XStore binlog backup"
// @Failure 404 {object} map[string]any "Binlog backup not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func GetBackupBinlog(c *gin.Context) { services.NewBackupBinlogService().Get(c) }

// UpdateBackupBinlog updates an existing XStore binlog backup.
// @Summary Update XStore binlog backup
// @Description Update an existing XStore binlog backup resource.
// @Tags xstores, backup-binlogs
// @Accept json
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the binlog backup"
// @Param name path string true "Name of the binlog backup"
// @Param body body polardbxv1.XStoreBackupBinlog true "Updated XStore binlog backup specification"
// @Success 200 {object} polardbxv1.XStoreBackupBinlog "Updated XStore binlog backup"
// @Failure 400 {object} map[string]any "Invalid request body"
// @Failure 404 {object} map[string]any "Binlog backup not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func UpdateBackupBinlog(c *gin.Context) { services.NewBackupBinlogService().Update(c) }

// DeleteBackupBinlog deletes an XStore binlog backup.
// @Summary Delete XStore binlog backup
// @Description Delete an XStore binlog backup by namespace and name.
// @Tags xstores, backup-binlogs
// @Param namespace path string true "Kubernetes namespace of the binlog backup"
// @Param name path string true "Name of the binlog backup"
// @Success 200 {object} map[string]any "Deletion confirmation"
// @Failure 404 {object} map[string]any "Binlog backup not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func DeleteBackupBinlog(c *gin.Context) { services.NewBackupBinlogService().Delete(c) }
