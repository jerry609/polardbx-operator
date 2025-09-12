package backupbinlog

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"polardbx-ui-backend/pkg/api/util"
	"polardbx-ui-backend/pkg/k8s"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
)

// List backup binlogs
func List(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.DefaultQuery("namespace", "default")
	items, err := k8s.ListPolarDBXBackupBinlogs(cli, ns)
	if err != nil {
		util.HandleK8sError(c, "failed to list backup binlogs", err)
		return
	}
	c.JSON(http.StatusOK, items)
}

// Create backup binlog
func Create(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.DefaultQuery("namespace", "default")
	var body polardbxv1.PolarDBXBackupBinlog
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}
	if body.Namespace == "" {
		body.Namespace = ns
	}
	created, err := k8s.CreatePolarDBXBackupBinlog(cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to create backup binlog", err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

// Get backup binlog
func Get(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	item, err := k8s.GetPolarDBXBackupBinlog(cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "failed to get backup binlog", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

// Update backup binlog
func Update(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	var body polardbxv1.PolarDBXBackupBinlog
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}
	body.Namespace = ns
	body.Name = name
	updated, err := k8s.UpdatePolarDBXBackupBinlog(cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to update backup binlog", err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

// Delete backup binlog
func Delete(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	// Use direct object to delete to avoid extra GET
	obj := &polardbxv1.PolarDBXBackupBinlog{}
	obj.Namespace = ns
	obj.Name = name
	if err := cli.Delete(c.Request.Context(), obj, &client.DeleteOptions{}); err != nil {
		util.HandleK8sError(c, "failed to delete backup binlog", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "backup binlog deleted successfully"})
}
