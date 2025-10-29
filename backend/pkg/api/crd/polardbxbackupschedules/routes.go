package polardbxbackupschedules

import (
	"polardbx-ui-backend/pkg/api/crd/common"
	domain_pxc "polardbx-ui-backend/pkg/api/domain/polardbxclusters"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(crd *gin.RouterGroup) {
	common.RegisterNamespaceScopedCRUD(crd, "polardbxbackupschedules", common.CRUDHandlers{
		List:   domain_pxc.ListSchedules,
		Create: domain_pxc.CreateSchedule,
		Get:    domain_pxc.GetSchedule,
		Update: domain_pxc.UpdateSchedule,
		Delete: domain_pxc.DeleteSchedule,
	})
}
