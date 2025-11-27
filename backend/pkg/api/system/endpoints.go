// Package system 提供系统信息查询的 API 端点
// 实际业务逻辑已迁移到 domain/platform/system
package system

import (
	"polardbx-ui-backend/pkg/api/domain/platform/system/handler"

	"github.com/gin-gonic/gin"
)

// ContextInfo returns current k8s user, context and default namespace if provided.
// 委托给 domain handler
var ContextInfo = handler.ContextInfo

// ListNamespaces returns all namespaces visible to the provided kubeconfig.
// 委托给 domain handler
var ListNamespaces = handler.ListNamespaces

// RegisterRoutes 注册 system 相关路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/context-info", ContextInfo)
	r.GET("/namespaces", ListNamespaces)
}
