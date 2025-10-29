package polardbxbackupbinlogs

import (
	"polardbx-ui-backend/pkg/api/crd/common"
	domain_pxc "polardbx-ui-backend/pkg/api/domain/polardbxclusters"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(crd *gin.RouterGroup) {
	common.RegisterNamespaceScopedCRUD(crd, "polardbxbackupbinlogs", common.CRUDHandlers{
		List:   domain_pxc.ListBackupBinlogs,
		Create: domain_pxc.CreateBackupBinlog,
		Get:    domain_pxc.GetBackupBinlog,
		Update: domain_pxc.UpdateBackupBinlog,
		Delete: domain_pxc.DeleteBackupBinlog,
	})
}
