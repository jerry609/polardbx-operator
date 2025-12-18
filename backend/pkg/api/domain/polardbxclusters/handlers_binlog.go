package polardbxclusters

import (
	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"github.com/gin-gonic/gin"

	"polardbx-ui-backend/pkg/api/domain/polardbxclusters/services"
	apierr "polardbx-ui-backend/pkg/api/errors"
	"polardbx-ui-backend/pkg/api/util"
)

// --- Handlers forwarding to BackupBinlog service (pure business methods) ---

func ListBackupBinlogs(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.DefaultQuery("namespace", "")
	items, err := services.NewBackupBinlogService().ListBinlogs(c.Request.Context(), cli, ns)
	if err != nil {
		apierr.AbortWithError(c, err)
		return
	}
	apierr.OK(c, items)
}

func CreateBackupBinlog(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.DefaultQuery("namespace", "default")
	var obj polardbxv1.PolarDBXBackupBinlog
	if err := c.ShouldBindJSON(&obj); err != nil {
		apierr.AbortValidation(c, "invalid backup binlog: "+err.Error())
		return
	}
	created, err := services.NewBackupBinlogService().CreateBinlog(c.Request.Context(), cli, ns, &obj)
	if err != nil {
		apierr.AbortWithError(c, err)
		return
	}
	apierr.Created(c, created)
}

func GetBackupBinlog(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	obj, err := services.NewBackupBinlogService().GetBinlog(c.Request.Context(), cli, ns, name)
	if err != nil {
		apierr.AbortWithError(c, err)
		return
	}
	apierr.OK(c, obj)
}

func UpdateBackupBinlog(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	var obj polardbxv1.PolarDBXBackupBinlog
	if err := c.ShouldBindJSON(&obj); err != nil {
		apierr.AbortValidation(c, "invalid backup binlog: "+err.Error())
		return
	}
	updated, err := services.NewBackupBinlogService().UpdateBinlog(c.Request.Context(), cli, ns, &obj)
	if err != nil {
		apierr.AbortWithError(c, err)
		return
	}
	apierr.OK(c, updated)
}

func DeleteBackupBinlog(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	if err := services.NewBackupBinlogService().DeleteBinlog(c.Request.Context(), cli, ns, name); err != nil {
		apierr.AbortWithError(c, err)
		return
	}
	apierr.OK(c, gin.H{"message": "backup binlog deleted"})
}
