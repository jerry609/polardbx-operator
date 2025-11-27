package platform

import (
	api_monitoring "polardbx-ui-backend/pkg/api/monitoring"

	"github.com/gin-gonic/gin"
)

// All monitoring handlers now use v2 (unified workflow)
func MonitoringBootstrap(c *gin.Context)       { api_monitoring.Bootstrap(c) }
func MonitoringBootstrapStatus(c *gin.Context) { api_monitoring.BootstrapStatus(c) }
func MonitoringBootstrapLogs(c *gin.Context)   { api_monitoring.GetBootstrapLogs(c) }
func MonitoringStatus(c *gin.Context)          { api_monitoring.Status(c) }
func MonitoringPreflight(c *gin.Context)       { api_monitoring.DetectEnvironment(c) }
func MonitoringUninstall(c *gin.Context)       { api_monitoring.Uninstall(c) }
func MonitoringDetect(c *gin.Context)          { api_monitoring.DetectEnvironment(c) }
func MonitoringCreatePlan(c *gin.Context)      { api_monitoring.CreatePlan(c) }
func MonitoringInstall(c *gin.Context)         { api_monitoring.StartInstallation(c) }
func MonitoringInstallStatus(c *gin.Context)   { api_monitoring.GetInstallStatus(c) }
func MonitoringInstallRetry(c *gin.Context)    { api_monitoring.TriggerRetry(c) }
func MonitoringDiagnose(c *gin.Context)        { api_monitoring.DiagnoseFailure(c) }
func MonitoringAutoFix(c *gin.Context)         { api_monitoring.ApplyAutoFix(c) }
