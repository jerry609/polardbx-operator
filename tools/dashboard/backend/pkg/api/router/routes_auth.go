package router

import (
	"polardbx-ui-backend/pkg/api"
	domain_auth "polardbx-ui-backend/pkg/api/domain/platform/auth/handler"
	domain_settings "polardbx-ui-backend/pkg/api/domain/platform/settings/handler"

	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers routes that don't require authentication
func RegisterPublicRoutes(v1 *gin.RouterGroup) {
	// Image registry configuration (public for UI initialization)
	v1.GET("/image-registry/config", domain_settings.GetImageRegistryConfig)
	v1.PUT("/image-registry/config", domain_settings.UpdateImageRegistryConfig)
	v1.GET("/image-registry/presets", domain_settings.GetAvailableRegistries)
	v1.POST("/image-registry/test", domain_settings.TestImageRegistry)

	// Connect endpoint - establishes client for subsequent requests
	v1.POST("/connect", api.KubeconfigAuthMiddleware(), api.Connect)
}

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(v1 *gin.RouterGroup) {
	// JWT login endpoints (optional)
	v1.POST("/auth/login", domain_auth.Login)
	v1.GET("/auth/me", domain_auth.Me)
}
