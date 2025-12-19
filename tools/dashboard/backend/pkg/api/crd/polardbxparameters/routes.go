package polardbxparameters

import (
	domain_parameters "polardbx-ui-backend/pkg/api/domain/platform/parameters/handler"
	"polardbx-ui-backend/pkg/api/routerutil"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(crd *gin.RouterGroup) {
	routerutil.RegisterCRUD(crd, "/polardbxparameters", routerutil.CRUDHandlers{
		List:   domain_parameters.List,
		Create: domain_parameters.Create,
		Get:    domain_parameters.Get,
		Update: domain_parameters.Update,
		Delete: domain_parameters.Delete,
	})
}
