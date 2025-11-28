package polardbxparameters

import (
	domain_parameters "polardbx-ui-backend/pkg/api/domain/platform/parameters/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(crd *gin.RouterGroup) {
	r := crd.Group("/polardbxparameters")
	r.GET("", domain_parameters.List)
	r.POST("", domain_parameters.Create)
	r.GET("/:name", domain_parameters.Get)
	r.PUT("/:name", domain_parameters.Update)
	r.DELETE("/:name", domain_parameters.Delete)
}
