package xstores

import (
	"polardbx-ui-backend/pkg/api/crd/common"
	domain_xs "polardbx-ui-backend/pkg/api/domain/xstores"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes adds /crd/xstores CRUD aliases.
func RegisterRoutes(crd *gin.RouterGroup) {
	r := common.RegisterNamespaceScopedCRUD(crd, "xstores", common.CRUDHandlers{
		List:   domain_xs.List,
		Create: domain_xs.Create,
		Get:    domain_xs.Get,
		Update: domain_xs.Update,
		Delete: domain_xs.Delete,
	})
	
	// Additional endpoint for namespace-scoped item
	item := r.Group("/:namespace/:name")
	item.GET("/pods", domain_xs.ListPods)
}
