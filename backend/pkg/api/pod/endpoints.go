// Package pod 提供 Pod 操作的 API 端点
// 实际业务逻辑已迁移到 domain/platform/pod
package pod

import (
	"polardbx-ui-backend/pkg/api/domain/platform/pod/handler"

	"github.com/gin-gonic/gin"
)

// GetLogs 获取 Pod 日志 - 委托给 domain handler
var GetLogs = handler.GetLogs

// ListForCluster 列出集群的所有 Pod - 委托给 domain handler
var ListForCluster = handler.ListForCluster

// List 列出命名空间的所有 Pod - 委托给 domain handler
var List = handler.List

// Get 获取指定的 Pod - 委托给 domain handler
var Get = handler.Get

// Delete 删除指定的 Pod - 委托给 domain handler
var Delete = handler.Delete

// ExecWS WebSocket 代理 K8s Exec - 委托给 domain handler
var ExecWS = handler.ExecWS

// RegisterRoutes 注册 Pod 相关路由
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/pods", List)
	r.GET("/pods/:namespace/:name", Get)
	r.DELETE("/pods/:namespace/:name", Delete)
	r.GET("/pods/:namespace/:pod_name/logs", GetLogs)
}
