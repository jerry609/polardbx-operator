package polardbxmonitors

import (
	api_monitoring "polardbx-ui-backend/pkg/api/monitoring"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(crd *gin.RouterGroup) {
	r := crd.Group("/polardbxmonitors")
	r.GET("", api_monitoring.List)
	r.POST("", api_monitoring.Create)
	item := r.Group("/:namespace/:name")
	item.GET("", api_monitoring.Get)
	item.PUT("", api_monitoring.Update)
	item.DELETE("", api_monitoring.Delete)
}
