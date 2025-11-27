// Package grafana 提供 Grafana 模板相关的 HTTP 处理器
// 这是一个薄包装层，委托到 domain/platform/grafana/handler
package grafana

import (
"polardbx-ui-backend/pkg/api/domain/platform/grafana/handler"
)

// ListTemplates GET /grafana/templates
var ListTemplates = handler.ListTemplates

// GetTemplate GET /grafana/templates/:name
var GetTemplate = handler.GetTemplate
