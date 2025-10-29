package polardbxlogcollectors

import (
	"polardbx-ui-backend/pkg/api/crd/common"
	api_logcollector "polardbx-ui-backend/pkg/api/logcollector"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(crd *gin.RouterGroup) {
	r := common.RegisterNamespaceScopedCRUD(crd, "polardbxlogcollectors", common.CRUDHandlers{
		List:   api_logcollector.List,
		Create: api_logcollector.Create,
		Get:    api_logcollector.Get,
		Update: api_logcollector.Update,
		Delete: api_logcollector.Delete,
	})
	
	// namespace-scoped extra endpoints
	ns := r.Group("/:namespace")
	ns.GET("/pipeline", api_logcollector.GetLogstashPipeline)
	ns.PUT("/pipeline", api_logcollector.UpdateLogstashPipeline)
	ns.GET("/elastic-certs", api_logcollector.GetElasticsearchCert)
	ns.PUT("/elastic-certs", api_logcollector.UpdateElasticsearchCert)
	ns.GET("/:name/status", api_logcollector.GetLogCollectorStatus)
	ns.GET("/logstash/logs", api_logcollector.StreamLogstashLogs)
	ns.POST("/test", api_logcollector.TestLogCollector)
}
