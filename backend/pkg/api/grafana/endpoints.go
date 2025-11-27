// Package grafana 提供 Grafana 相关的 HTTP 处理器
// 这是一个薄包装层，委托到 domain/platform/grafana/handler
package grafana

import (
	"github.com/gin-gonic/gin"

	"polardbx-ui-backend/pkg/api/domain/platform/grafana/handler"
	"polardbx-ui-backend/pkg/api/domain/platform/grafana/repository"
)

// Config 类型别名，保持向后兼容
type Config = repository.GrafanaConfig

// GetConfig GET /grafana/config
var GetConfig = handler.GetConfig

// PutConfig PUT /grafana/config
var PutConfig = handler.PutConfig

// SyncDashboards POST /grafana/dashboards/sync
var SyncDashboards = handler.SyncDashboards

// ListDashboards GET /grafana/dashboards
var ListDashboards = handler.ListDashboards

// ListDashboardVersions GET /grafana/dashboards/:name/versions
var ListDashboardVersions = handler.ListDashboardVersions

// RollbackDashboard POST /grafana/dashboards/:name/rollback
var RollbackDashboard = handler.RollbackDashboard

// 保留原有函数签名用于兼容性
func getConfig(c *gin.Context)             { handler.GetConfig(c) }
func putConfig(c *gin.Context)             { handler.PutConfig(c) }
func syncDashboards(c *gin.Context)        { handler.SyncDashboards(c) }
func listDashboards(c *gin.Context)        { handler.ListDashboards(c) }
func listDashboardVersions(c *gin.Context) { handler.ListDashboardVersions(c) }
func rollbackDashboard(c *gin.Context)     { handler.RollbackDashboard(c) }
