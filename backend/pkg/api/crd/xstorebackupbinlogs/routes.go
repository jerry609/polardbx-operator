package xstorebackupbinlogs

import (
	"polardbx-ui-backend/pkg/api/crd/common"
	domain_xstores "polardbx-ui-backend/pkg/api/domain/xstores"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes under /api/v1/crd/xstorebackupbinlogs
func RegisterRoutes(crd *gin.RouterGroup) {
	common.RegisterNamespaceScopedCRUD(crd, "xstorebackupbinlogs", common.CRUDHandlers{
		List:   domain_xstores.ListBackupBinlogs,
		Create: domain_xstores.CreateBackupBinlog,
		Get:    domain_xstores.GetBackupBinlog,
		Update: domain_xstores.UpdateBackupBinlog,
		Delete: domain_xstores.DeleteBackupBinlog,
	})
}
