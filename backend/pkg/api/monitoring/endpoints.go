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
	// ========================================
	// PolarDBXMonitor CRD CRUD (/monitors)
	// ========================================
	rg.GET("/monitors", domain_monitoring.ListMonitors)
	rg.POST("/monitors", domain_monitoring.CreateMonitor)
	rg.GET("/monitors/:namespace/:name", domain_monitoring.GetMonitor)
	rg.PUT("/monitors/:namespace/:name", domain_monitoring.UpdateMonitor)
	rg.DELETE("/monitors/:namespace/:name", domain_monitoring.DeleteMonitor)

	// ========================================
	// 监控栈安装工作流 (/monitoring/*)
	// ========================================

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

// ========================================
// 兼容性导出 - 安装工作流
// ========================================

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

// ========================================
// 兼容性导出 - CRD CRUD (原 api_monitor)
// ========================================

var (
	List   = domain_monitoring.ListMonitors
	Create = domain_monitoring.CreateMonitor
	Get    = domain_monitoring.GetMonitor
	Update = domain_monitoring.UpdateMonitor
	Delete = domain_monitoring.DeleteMonitor
)
