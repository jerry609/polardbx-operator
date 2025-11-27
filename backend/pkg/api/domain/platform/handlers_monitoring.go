package platform

import (
	api_monitoring "polardbx-ui-backend/pkg/api/monitoring"
	api_monitoring_v2 "polardbx-ui-backend/pkg/api/monitoringv2"

	"github.com/gin-gonic/gin"
)

// Legacy v1 handlers (kept for complex Job-based bootstrap)
func MonitoringBootstrap(c *gin.Context)       { api_monitoring.Bootstrap(c) }
func MonitoringBootstrapStatus(c *gin.Context) { api_monitoring.BootstrapStatus(c) }
func MonitoringStatus(c *gin.Context)          { api_monitoring.Status(c) }

// v2 handlers (unified workflow)
func MonitoringBootstrapLogs(c *gin.Context) { api_monitoring_v2.GetBootstrapLogs(c) }
func MonitoringPreflight(c *gin.Context)     { api_monitoring_v2.DetectEnvironment(c) }
func MonitoringUninstall(c *gin.Context)     { api_monitoring_v2.Uninstall(c) }
func MonitoringDetect(c *gin.Context)        { api_monitoring_v2.DetectEnvironment(c) }
func MonitoringCreatePlan(c *gin.Context)    { api_monitoring_v2.CreatePlan(c) }
func MonitoringInstall(c *gin.Context)       { api_monitoring_v2.StartInstallation(c) }
func MonitoringInstallStatus(c *gin.Context) { api_monitoring_v2.GetInstallStatus(c) }
func MonitoringInstallRetry(c *gin.Context)  { api_monitoring_v2.TriggerRetry(c) }
func MonitoringDiagnose(c *gin.Context)      { api_monitoring_v2.DiagnoseFailure(c) }
func MonitoringAutoFix(c *gin.Context)       { api_monitoring_v2.ApplyAutoFix(c) }
