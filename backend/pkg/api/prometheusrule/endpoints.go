// Package prometheusrule 提供 PrometheusRule 告警规则管理的 API 端点
// 这是一个薄包装层，委托到 domain/platform/prometheusrule/handler
package prometheusrule

import (
"github.com/gin-gonic/gin"
"k8s.io/apimachinery/pkg/runtime/schema"

"polardbx-ui-backend/pkg/api/domain/platform/prometheusrule/handler"
"polardbx-ui-backend/pkg/api/util"
)

// ======================== 保留原始类型定义用于向后兼容 ========================

// PrometheusRule represents a simplified view of PrometheusRule resource
type PrometheusRule struct {
APIVersion string                 `json:"apiVersion"`
Kind       string                 `json:"kind"`
Metadata   PrometheusRuleMetadata `json:"metadata"`
Spec       PrometheusRuleSpec     `json:"spec"`
}

// PrometheusRuleMetadata 元数据
type PrometheusRuleMetadata struct {
Name              string            `json:"name"`
Namespace         string            `json:"namespace"`
CreationTimestamp string            `json:"creationTimestamp"`
Labels            map[string]string `json:"labels,omitempty"`
}

// PrometheusRuleSpec 规格
type PrometheusRuleSpec struct {
Groups []PrometheusRuleGroup `json:"groups"`
}

// PrometheusRuleGroup 规则组
type PrometheusRuleGroup struct {
Name     string                `json:"name"`
Interval string                `json:"interval,omitempty"`
Rules    []PrometheusRuleEntry `json:"rules"`
}

// PrometheusRuleEntry 规则条目
type PrometheusRuleEntry struct {
Alert       string            `json:"alert,omitempty"`
Expr        string            `json:"expr"`
For         string            `json:"for,omitempty"`
Labels      map[string]string `json:"labels,omitempty"`
Annotations map[string]string `json:"annotations,omitempty"`
Record      string            `json:"record,omitempty"`
}

// ======================== 模板类型（用于测试兼容） ========================

// AlertRuleTemplateSummary 告警规则模板摘要
type AlertRuleTemplateSummary = handler.AlertRuleTemplateSummary

// AlertRuleTemplateDetail 告警规则模板详情
type AlertRuleTemplateDetail = handler.AlertRuleTemplateDetail

// alertTemplateDirEnv 模板目录环境变量名（用于测试）
const alertTemplateDirEnv = "ALERT_TEMPLATE_DIR"

// prometheusRuleGVR 用于测试兼容
var prometheusRuleGVR = schema.GroupVersionResource{
Group:    "monitoring.coreos.com",
Version:  "v1",
Resource: "prometheusrules",
}

// isValidPromQLBasic 基础 PromQL 验证（用于测试兼容）
func isValidPromQLBasic(expr string) bool {
return handler.IsValidPromQLBasic(expr)
}

// ======================== 工厂函数 ========================

// getHandler 从 context 获取 handler
func getHandler(c *gin.Context) (*handler.PrometheusRuleHandler, bool) {
dynClient, ok := util.DynamicClientFromContext(c)
if !ok {
return nil, false
}
return handler.NewPrometheusRuleHandler(dynClient), true
}

// ======================== PrometheusRule CRUD (委托到 handler) ========================

// List GET /prometheusrules
func List(c *gin.Context) {
h, ok := getHandler(c)
if !ok {
return
}
h.List(c)
}

// GetYAML GET /prometheusrules/:namespace/:name/yaml
func GetYAML(c *gin.Context) {
h, ok := getHandler(c)
if !ok {
return
}
h.GetYAML(c)
}

// ValidateRule POST /prometheusrules/validate
func ValidateRule(c *gin.Context) {
h, ok := getHandler(c)
if !ok {
return
}
h.ValidateRule(c)
}

// ======================== Template Management (委托到 handler) ========================

// ListTemplates GET /prometheusrules/templates
func ListTemplates(c *gin.Context) {
h, ok := getHandler(c)
if !ok {
return
}
h.ListTemplates(c)
}

// GetTemplate GET /prometheusrules/templates/:name
func GetTemplate(c *gin.Context) {
h, ok := getHandler(c)
if !ok {
return
}
h.GetTemplate(c)
}

// ApplyTemplate POST /prometheusrules/templates/:name/apply
func ApplyTemplate(c *gin.Context) {
h, ok := getHandler(c)
if !ok {
return
}
h.ApplyTemplate(c)
}

// ======================== RegisterRoutes ========================

// RegisterRoutes 注册 PrometheusRule 相关路由
func RegisterRoutes(r *gin.RouterGroup) {
// CRUD
r.GET("/prometheusrules", List)
r.GET("/prometheusrules/:namespace/:name/yaml", GetYAML)

// Validation
r.POST("/prometheusrules/validate", ValidateRule)

// Templates
r.GET("/prometheusrules/templates", ListTemplates)
r.GET("/prometheusrules/templates/:name", GetTemplate)
r.POST("/prometheusrules/templates/:name/apply", ApplyTemplate)
}

// loadChartTemplates 加载 chart 模板（用于测试兼容）
func loadChartTemplates() ([]AlertRuleTemplateSummary, error) {
return handler.LoadChartTemplates()
}
