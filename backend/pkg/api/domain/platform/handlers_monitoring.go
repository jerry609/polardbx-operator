package platform

import (
	api_monitoring "polardbx-ui-backend/pkg/api/monitoring"

	"github.com/gin-gonic/gin"
)

func MonitoringBootstrap(c *gin.Context) { api_monitoring.Bootstrap(c) }
func MonitoringStatus(c *gin.Context)    { api_monitoring.Status(c) }
func MonitoringPreflight(c *gin.Context) { api_monitoring.Preflight(c) }
func MonitoringUninstall(c *gin.Context) { api_monitoring.Uninstall(c) }
