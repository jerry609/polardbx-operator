package systemtasks

import (
	domain_st "polardbx-ui-backend/pkg/api/domain/systemtasks"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes adds /crd/systemtasks CRUD aliases.
func RegisterRoutes(crd *gin.RouterGroup) {
	r := crd.Group("/systemtasks")
	r.GET("", domain_st.List)
	r.POST("", domain_st.Create)
	item := r.Group("/:namespace/:name")
	item.GET("", domain_st.Get)
	item.PUT("", domain_st.Update)
	item.DELETE("", domain_st.Delete)
}
