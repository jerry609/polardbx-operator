package polardbxmonitors

import (
	"polardbx-ui-backend/pkg/api/crd/common"
	api_monitor "polardbx-ui-backend/pkg/api/monitor"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(crd *gin.RouterGroup) {
	common.RegisterNamespaceScopedCRUD(crd, "polardbxmonitors", common.CRUDHandlers{
		List:   api_monitor.List,
		Create: api_monitor.Create,
		Get:    api_monitor.Get,
		Update: api_monitor.Update,
		Delete: api_monitor.Delete,
	})
}
