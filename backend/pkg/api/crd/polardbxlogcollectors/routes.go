package polardbxlogcollectors

import (
	domain_logcollector "polardbx-ui-backend/pkg/api/domain/platform/logcollector/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(crd *gin.RouterGroup) {
	r := crd.Group("/polardbxlogcollectors")
	r.GET("", domain_logcollector.List)
	r.POST("", domain_logcollector.Create)
	item := r.Group("/:namespace/:name")
	item.GET("", domain_logcollector.Get)
	item.PUT("", domain_logcollector.Update)
	item.DELETE("", domain_logcollector.Delete)
	// namespace-scoped extra endpoints
	ns := r.Group("/:namespace")
	ns.GET("/pipeline", domain_logcollector.GetLogstashPipeline)
	ns.PUT("/pipeline", domain_logcollector.UpdateLogstashPipeline)
	ns.GET("/elastic-certs", domain_logcollector.GetElasticsearchCert)
	ns.PUT("/elastic-certs", domain_logcollector.UpdateElasticsearchCert)
	ns.GET("/:name/status", domain_logcollector.GetLogCollectorStatus)
	ns.GET("/logstash/logs", domain_logcollector.StreamLogstashLogs)
	ns.POST("/test", domain_logcollector.TestLogCollector)
}
