// Package monitoring 提供监控模块的 API 端点定义
// 业务逻辑实现位于 domain/monitoring 包中
package monitoring

import (
	domain_monitoring "polardbx-ui-backend/pkg/api/domain/monitoring"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有监控相关的 API 路由
// 这是推荐的路由注册方式，可以在 main.go 中调用
func RegisterRoutes(rg *gin.RouterGroup) {
	// Bootstrap endpoints (传统安装)
	rg.POST("/monitoring/bootstrap", domain_monitoring.Bootstrap)
	rg.GET("/monitoring/bootstrap/status", domain_monitoring.BootstrapStatus)
	rg.GET("/monitoring/bootstrap/logs", domain_monitoring.GetBootstrapLogs)

	// Status and detection
	rg.GET("/monitoring/status", domain_monitoring.Status)
	rg.GET("/monitoring/preflight", domain_monitoring.DetectEnvironment)
	rg.GET("/monitoring/detect", domain_monitoring.DetectEnvironment)

	// Uninstall
	rg.DELETE("/monitoring/uninstall", domain_monitoring.Uninstall)

	// Installation workflow
	rg.POST("/monitoring/plan", domain_monitoring.CreatePlan)
	rg.POST("/monitoring/install", domain_monitoring.StartInstallation)
	rg.GET("/monitoring/install/:sessionId/status", domain_monitoring.GetInstallStatus)
	rg.GET("/monitoring/install/:sessionId/logs", domain_monitoring.GetBootstrapLogs)
	rg.POST("/monitoring/install/:sessionId/retry", domain_monitoring.TriggerRetry)

	// Diagnostics and auto-fix
	rg.POST("/monitoring/diagnose", domain_monitoring.DiagnoseFailure)
	rg.POST("/monitoring/auto-fix", domain_monitoring.ApplyAutoFix)
}

// 以下为兼容性导出，允许 main.go 继续使用原有的调用方式
// 推荐逐步迁移到使用 RegisterRoutes

var (
	Bootstrap         = domain_monitoring.Bootstrap
	BootstrapStatus   = domain_monitoring.BootstrapStatus
	GetBootstrapLogs  = domain_monitoring.GetBootstrapLogs
	Status            = domain_monitoring.Status
	DetectEnvironment = domain_monitoring.DetectEnvironment
	Uninstall         = domain_monitoring.Uninstall
	CreatePlan        = domain_monitoring.CreatePlan
	StartInstallation = domain_monitoring.StartInstallation
	GetInstallStatus  = domain_monitoring.GetInstallStatus
	TriggerRetry      = domain_monitoring.TriggerRetry
	DiagnoseFailure   = domain_monitoring.DiagnoseFailure
	ApplyAutoFix      = domain_monitoring.ApplyAutoFix
)
