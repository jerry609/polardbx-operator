package router

import (
	domain_system "polardbx-ui-backend/pkg/api/domain/platform/system/handler"
	domain_st "polardbx-ui-backend/pkg/api/domain/systemtasks"
	"polardbx-ui-backend/pkg/api/routerutil"

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
	base := "/system-tasks"
	routerutil.RegisterCRUDWithItemPattern(v1, base, base+"/:namespace/:name", routerutil.CRUDHandlers{
		List:   domain_st.List,
		Create: domain_st.Create,
		Get:    domain_st.Get,
		Update: domain_st.Update,
		Delete: domain_st.Delete,
	})
}
