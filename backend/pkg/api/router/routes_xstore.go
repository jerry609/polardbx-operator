package router

import (
	domain_xs "polardbx-ui-backend/pkg/api/domain/xstores"

	"github.com/gin-gonic/gin"
)

// RegisterXStoreRoutes registers XStore-related routes
func RegisterXStoreRoutes(v1 *gin.RouterGroup) {
	// XStore Backups
	v1.GET("/xstore-backups", domain_xs.ListBackups)
	v1.POST("/xstore-backups", domain_xs.CreateBackup)
	v1.GET("/xstore-backups/:namespace/:name", domain_xs.GetBackup)
	v1.PUT("/xstore-backups/:namespace/:name", domain_xs.UpdateBackup)
	v1.DELETE("/xstore-backups/:namespace/:name", domain_xs.DeleteBackup)
	v1.POST("/xstore-backups/:namespace/:name/force-delete", domain_xs.ForceDeleteBackup)

	// XStore Followers
	v1.GET("/xstore-followers", domain_xs.ListFollowers)
	v1.POST("/xstore-followers", domain_xs.CreateFollower)
	v1.GET("/xstore-followers/:namespace/:name", domain_xs.GetFollower)
	v1.PUT("/xstore-followers/:namespace/:name", domain_xs.UpdateFollower)
	v1.DELETE("/xstore-followers/:namespace/:name", domain_xs.DeleteFollower)
}
