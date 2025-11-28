package router

import (
	domain_system "polardbx-ui-backend/pkg/api/domain/platform/system/handler"
	domain_st "polardbx-ui-backend/pkg/api/domain/systemtasks"

	"github.com/gin-gonic/gin"
)

// RegisterSystemRoutes registers system-related routes
func RegisterSystemRoutes(v1 *gin.RouterGroup) {
	// System context
	v1.GET("/system/context", domain_system.ContextInfo)
	v1.GET("/system/namespaces", domain_system.ListNamespaces)

	// Direct namespace route for frontend compatibility
	v1.GET("/namespaces", domain_system.ListNamespaces)

	// System Tasks
	v1.GET("/system-tasks", domain_st.List)
	v1.POST("/system-tasks", domain_st.Create)
	v1.GET("/system-tasks/:namespace/:name", domain_st.Get)
	v1.PUT("/system-tasks/:namespace/:name", domain_st.Update)
	v1.DELETE("/system-tasks/:namespace/:name", domain_st.Delete)
}
