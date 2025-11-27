package platform

import (
	api_monitoring_v2 "polardbx-ui-backend/pkg/api/monitoringv2"

	"github.com/gin-gonic/gin"
)

// All monitoring handlers now use v2 (unified workflow)
func MonitoringBootstrap(c *gin.Context)       { api_monitoring_v2.Bootstrap(c) }
func MonitoringBootstrapStatus(c *gin.Context) { api_monitoring_v2.BootstrapStatus(c) }
func MonitoringBootstrapLogs(c *gin.Context)   { api_monitoring_v2.GetBootstrapLogs(c) }
func MonitoringStatus(c *gin.Context)          { api_monitoring_v2.Status(c) }
func MonitoringPreflight(c *gin.Context)       { api_monitoring_v2.DetectEnvironment(c) }
func MonitoringUninstall(c *gin.Context)       { api_monitoring_v2.Uninstall(c) }
func MonitoringDetect(c *gin.Context)          { api_monitoring_v2.DetectEnvironment(c) }
func MonitoringCreatePlan(c *gin.Context)      { api_monitoring_v2.CreatePlan(c) }
func MonitoringInstall(c *gin.Context)         { api_monitoring_v2.StartInstallation(c) }
func MonitoringInstallStatus(c *gin.Context)   { api_monitoring_v2.GetInstallStatus(c) }
func MonitoringInstallRetry(c *gin.Context)    { api_monitoring_v2.TriggerRetry(c) }
func MonitoringDiagnose(c *gin.Context)        { api_monitoring_v2.DiagnoseFailure(c) }
func MonitoringAutoFix(c *gin.Context)         { api_monitoring_v2.ApplyAutoFix(c) }
