// Package alerts 提供告警管理的 API 端点
// 实际业务逻辑已迁移到 domain/platform/alerts
package alerts

import (
	"polardbx-ui-backend/pkg/api/domain/platform/alerts/handler"

	"github.com/gin-gonic/gin"
)

// 委托给 domain handler
var (
	ListProfiles  = handler.ListProfiles
	CreateProfile = handler.CreateProfile
	GetProfile    = handler.GetProfile
	UpdateProfile = handler.UpdateProfile
	DeleteProfile = handler.DeleteProfile
	DryRunProfile = handler.DryRunProfile
	GetRoutes     = handler.GetRoutes
	PutRoutes     = handler.PutRoutes
	ListSilences  = handler.ListSilences
	CreateSilence = handler.CreateSilence
	DeleteSilence = handler.DeleteSilence
	TestAlert     = handler.TestAlert
)

// RegisterRoutes 注册告警相关路由
func RegisterRoutes(r *gin.RouterGroup) {
	// Profiles CRUD
	r.GET("/alerts/profiles", ListProfiles)
	r.POST("/alerts/profiles", CreateProfile)
	r.GET("/alerts/profiles/:name", GetProfile)
	r.PUT("/alerts/profiles/:name", UpdateProfile)
	r.DELETE("/alerts/profiles/:name", DeleteProfile)
	r.POST("/alerts/profiles/dryrun", DryRunProfile)

	// Routes
	r.GET("/alerts/routes", GetRoutes)
	r.PUT("/alerts/routes", PutRoutes)

	// Alertmanager proxy
	r.GET("/alerts/silences", ListSilences)
	r.POST("/alerts/silences", CreateSilence)
	r.DELETE("/alerts/silences/:id", DeleteSilence)
	r.POST("/alerts/test", TestAlert)
}
