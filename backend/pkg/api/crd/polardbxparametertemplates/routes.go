package polardbxparametertemplates

import (
	"polardbx-ui-backend/pkg/api/crd/common"
	api_parameters "polardbx-ui-backend/pkg/api/parameters"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(crd *gin.RouterGroup) {
	common.RegisterNamespaceScopedCRUD(crd, "polardbxparametertemplates", common.CRUDHandlers{
		List:   api_parameters.ListTemplates,
		Create: api_parameters.CreateTemplate,
		Get:    api_parameters.GetTemplate,
		Update: api_parameters.UpdateTemplate,
		Delete: api_parameters.DeleteTemplate,
	})
}
