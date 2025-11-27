// Package alerts 提供告警管理的 API 端点
// 实际业务逻辑已迁移到 domain/platform/alerts
package alerts

import (
	"polardbx-ui-backend/pkg/api/domain/platform/alerts/handler"

	"github.com/gin-gonic/gin"
)

// List 聚合告警列表 - 委托给 domain handler
var List gin.HandlerFunc = handler.List
