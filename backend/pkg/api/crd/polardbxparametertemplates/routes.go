package polardbxparametertemplates

import (
	domain_parameters "polardbx-ui-backend/pkg/api/domain/platform/parameters/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(crd *gin.RouterGroup) {
	r := crd.Group("/polardbxparametertemplates")
	r.GET("", domain_parameters.ListTemplates)
	r.POST("", domain_parameters.CreateTemplate)
	item := r.Group("/:namespace/:name")
	item.GET("", domain_parameters.GetTemplate)
	item.PUT("", domain_parameters.UpdateTemplate)
	item.DELETE("", domain_parameters.DeleteTemplate)
}
