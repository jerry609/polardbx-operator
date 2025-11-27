// Package logcollector 提供日志收集器相关的 API 端点
// 这是一个薄包装层，实际业务逻辑在 domain/platform/logcollector/handler 中实现
package logcollector

import (
"github.com/gin-gonic/gin"

"polardbx-ui-backend/pkg/api/domain/platform/logcollector/handler"
)

// GetLogstashPipeline GET /logcollectors/:namespace/pipeline
// 获取 logstash pipeline ConfigMap 内容
func GetLogstashPipeline(c *gin.Context) {
handler.GetLogstashPipeline(c)
}

// UpdateLogstashPipeline PUT /logcollectors/:namespace/pipeline
// 更新或创建 logstash pipeline ConfigMap
func UpdateLogstashPipeline(c *gin.Context) {
handler.UpdateLogstashPipeline(c)
}

// GetElasticsearchCert GET /logcollectors/:namespace/es-cert
// 获取 Elasticsearch 证书信息
func GetElasticsearchCert(c *gin.Context) {
handler.GetElasticsearchCert(c)
}

// UpdateElasticsearchCert PUT /logcollectors/:namespace/es-cert
// 更新 Elasticsearch 证书
func UpdateElasticsearchCert(c *gin.Context) {
handler.UpdateElasticsearchCert(c)
}

// GetLogCollectorStatus GET /logcollectors/:namespace/:name/status
// 聚合 Filebeat/Logstash 就绪状态和输出模式
func GetLogCollectorStatus(c *gin.Context) {
handler.GetLogCollectorStatus(c)
}

// StreamLogstashLogs GET /logcollectors/:namespace/logs
// 流式输出 logstash pod 日志
func StreamLogstashLogs(c *gin.Context) {
handler.StreamLogstashLogs(c)
}

// TestLogCollector GET /logcollectors/:namespace/test
// 执行简单检查以查找常见配置错误
func TestLogCollector(c *gin.Context) {
handler.TestLogCollector(c)
}
