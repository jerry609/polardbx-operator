package router

import (
	domain_xs "polardbx-ui-backend/pkg/api/domain/xstores"
	"polardbx-ui-backend/pkg/api/routerutil"

	"github.com/gin-gonic/gin"
)

// RegisterXStoreRoutes registers XStore-related routes
func RegisterXStoreRoutes(v1 *gin.RouterGroup) {
	// XStore Backups
	baseBackups := "/xstore-backups"
	routerutil.RegisterCRUDWithItemPattern(v1, baseBackups, baseBackups+"/:namespace/:name", routerutil.CRUDHandlers{
		List:   domain_xs.ListBackups,
		Create: domain_xs.CreateBackup,
		Get:    domain_xs.GetBackup,
		Update: domain_xs.UpdateBackup,
		Delete: domain_xs.DeleteBackup,
	})
	v1.POST("/xstore-backups/:namespace/:name/force-delete", domain_xs.ForceDeleteBackup)

	// XStore Followers
	baseFollowers := "/xstore-followers"
	routerutil.RegisterCRUDWithItemPattern(v1, baseFollowers, baseFollowers+"/:namespace/:name", routerutil.CRUDHandlers{
		List:   domain_xs.ListFollowers,
		Create: domain_xs.CreateFollower,
		Get:    domain_xs.GetFollower,
		Update: domain_xs.UpdateFollower,
		Delete: domain_xs.DeleteFollower,
	})
}
