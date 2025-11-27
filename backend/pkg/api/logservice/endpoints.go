// Package logservice 提供日志服务状态相关的 HTTP 处理器
// 这是一个薄包装层，委托到 domain/platform/logservice/handler
package logservice

import (
	"github.com/gin-gonic/gin"

	"polardbx-ui-backend/pkg/api/domain/platform/logservice/handler"
)

// Status GET /logservice/status
// 聚合日志收集组件的就绪状态和配置
var Status = handler.Status

// 保留原有函数签名用于兼容性
func status(c *gin.Context) { handler.Status(c) }
