package polardbxparameters

import (
	"polardbx-ui-backend/pkg/api/crd/common"
	api_parameters "polardbx-ui-backend/pkg/api/parameters"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(crd *gin.RouterGroup) {
	common.RegisterClusterScopedCRUD(crd, "polardbxparameters", common.CRUDHandlers{
		List:   api_parameters.List,
		Create: api_parameters.Create,
		Get:    api_parameters.Get,
		Update: api_parameters.Update,
		Delete: api_parameters.Delete,
	})
}
