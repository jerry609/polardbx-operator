package systemtasks

import (
	"polardbx-ui-backend/pkg/api/crd/common"
	api_systemtask "polardbx-ui-backend/pkg/api/systemtask"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes adds /crd/systemtasks CRUD aliases.
func RegisterRoutes(crd *gin.RouterGroup) {
	common.RegisterNamespaceScopedCRUD(crd, "systemtasks", common.CRUDHandlers{
		List:   api_systemtask.List,
		Create: api_systemtask.Create,
		Get:    api_systemtask.Get,
		Update: api_systemtask.Update,
		Delete: api_systemtask.Delete,
	})
}
