package common

import (
	"github.com/gin-gonic/gin"
)

// CRUDHandlers defines the standard CRUD operation handlers for a resource.
type CRUDHandlers struct {
	List   gin.HandlerFunc
	Create gin.HandlerFunc
	Get    gin.HandlerFunc
	Update gin.HandlerFunc
	Delete gin.HandlerFunc
}

// RegisterNamespaceScopedCRUD registers standard CRUD routes for a namespace-scoped resource.
// The routes follow the pattern:
//   GET    /<resourcePath>                    -> List
//   POST   /<resourcePath>                    -> Create
//   GET    /<resourcePath>/:namespace/:name   -> Get
//   PUT    /<resourcePath>/:namespace/:name   -> Update
//   DELETE /<resourcePath>/:namespace/:name   -> Delete
func RegisterNamespaceScopedCRUD(parent *gin.RouterGroup, resourcePath string, handlers CRUDHandlers) *gin.RouterGroup {
	r := parent.Group("/" + resourcePath)
	r.GET("", handlers.List)
	r.POST("", handlers.Create)
	item := r.Group("/:namespace/:name")
	item.GET("", handlers.Get)
	item.PUT("", handlers.Update)
	item.DELETE("", handlers.Delete)
	return r
}

// RegisterClusterScopedCRUD registers standard CRUD routes for a cluster-scoped (non-namespaced) resource.
// The routes follow the pattern:
//   GET    /<resourcePath>       -> List
//   POST   /<resourcePath>       -> Create
//   GET    /<resourcePath>/:name -> Get
//   PUT    /<resourcePath>/:name -> Update
//   DELETE /<resourcePath>/:name -> Delete
func RegisterClusterScopedCRUD(parent *gin.RouterGroup, resourcePath string, handlers CRUDHandlers) *gin.RouterGroup {
	r := parent.Group("/" + resourcePath)
	r.GET("", handlers.List)
	r.POST("", handlers.Create)
	r.GET("/:name", handlers.Get)
	r.PUT("/:name", handlers.Update)
	r.DELETE("/:name", handlers.Delete)
	return r
}
