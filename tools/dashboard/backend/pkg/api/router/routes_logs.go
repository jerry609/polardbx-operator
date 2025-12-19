package router

import (
	domain_logcollector "polardbx-dashboard-backend/pkg/api/domain/platform/logcollector/handler"
	domain_logs "polardbx-dashboard-backend/pkg/api/domain/platform/logs/handler"
	domain_logservice "polardbx-dashboard-backend/pkg/api/domain/platform/logservice/handler"
	domain_logstrategy "polardbx-dashboard-backend/pkg/api/domain/platform/logstrategy/handler"
	"polardbx-dashboard-backend/pkg/api/routerutil"

	"github.com/gin-gonic/gin"
)

// RegisterLogsRoutes registers log-related routes
func RegisterLogsRoutes(v1 *gin.RouterGroup) {
	// Log Collectors
	base := "/log-collectors"
	routerutil.RegisterCRUDWithItemPattern(v1, base, base+"/:namespace/:name", routerutil.CRUDHandlers{
		List:   domain_logcollector.List,
		Create: domain_logcollector.Create,
		Get:    domain_logcollector.Get,
		Update: domain_logcollector.Update,
		Delete: domain_logcollector.Delete,
	})
	v1.GET("/log-collectors/:namespace/pipeline", domain_logcollector.GetLogstashPipeline)
	v1.PUT("/log-collectors/:namespace/pipeline", domain_logcollector.UpdateLogstashPipeline)
	v1.GET("/log-collectors/:namespace/elastic-certs", domain_logcollector.GetElasticsearchCert)
	v1.PUT("/log-collectors/:namespace/elastic-certs", domain_logcollector.UpdateElasticsearchCert)
	v1.GET("/log-collectors/:namespace/:name/status", domain_logcollector.GetLogCollectorStatus)
	v1.GET("/log-collectors/:namespace/logstash/logs", domain_logcollector.StreamLogstashLogs)
	v1.POST("/log-collectors/:namespace/test", domain_logcollector.TestLogCollector)

	// Log Service
	v1.GET("/log-service/status", domain_logservice.Status)

	// Log Strategies
	v1.GET("/log-strategies", domain_logstrategy.List)
	v1.POST("/log-strategies", domain_logstrategy.Create)
	v1.POST("/log-strategies/precheck", domain_logstrategy.Precheck)
	v1.GET("/log-strategies/apply-records", domain_logstrategy.ListApplyRecords)
	v1.GET("/log-strategies/:name", domain_logstrategy.Get)
	v1.PUT("/log-strategies/:name", domain_logstrategy.Update)
	v1.DELETE("/log-strategies/:name", domain_logstrategy.Delete)
	v1.POST("/log-strategies/:name/apply", domain_logstrategy.Apply)
	v1.POST("/log-strategies/test-connection", domain_logstrategy.TestConnection)

	// Logs Bootstrap (installation wizard)
	v1.POST("/logs/bootstrap", domain_logs.Bootstrap)
	v1.GET("/logs/bootstrap/status", domain_logs.BootstrapStatus)
	v1.GET("/logs/bootstrap/logs", domain_logs.BootstrapLogs)
	v1.POST("/logs/query", domain_logs.Query)
	v1.GET("/logs/presets", domain_logs.Presets)
	v1.GET("/logs/presets/:pattern", domain_logs.PresetByPattern)
}
