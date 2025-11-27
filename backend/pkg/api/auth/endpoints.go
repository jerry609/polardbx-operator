// Package auth 提供 JWT 认证的 API 端点
// 实际业务逻辑已迁移到 domain/platform/auth
package auth

import (
	"polardbx-ui-backend/pkg/api/domain/platform/auth/handler"

	"github.com/gin-gonic/gin"
)

// 导出类型别名，保持兼容性
type Claims = handler.Claims

// 委托给 domain handler
var (
	Login             = handler.Login
	Me                = handler.Me
	JWTAuthMiddleware = handler.JWTAuthMiddleware
)

// RegisterRoutes 注册认证相关路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/auth/login", Login)
	r.GET("/auth/me", Me)
}
