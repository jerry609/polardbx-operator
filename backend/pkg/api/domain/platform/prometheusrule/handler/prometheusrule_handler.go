// Package handler 提供 PrometheusRule 告警规则管理的 HTTP 处理器
// 遵循 Clean Architecture 设计模式
package handler

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

// ======================== Types ========================

// PrometheusRule 表示告警规则资源
type PrometheusRule struct {
	Name           string            `json:"name"`
	Namespace      string            `json:"namespace"`
	Labels         map[string]string `json:"labels,omitempty"`
	GroupsCount    int               `json:"groupsCount"`
	RulesCount     int               `json:"rulesCount"`
	AlertsCount    int               `json:"alertsCount"`
	RecordingCount int               `json:"recordingCount"`
	Groups         []interface{}     `json:"groups,omitempty"`
}

// AlertRuleTemplateSummary 表示告警规则模板的摘要信息
type AlertRuleTemplateSummary struct {
	Name            string                  `json:"name"`
	DisplayName     string                  `json:"displayName"`
	Title           string                  `json:"title"` // 前端使用
	Description     string                  `json:"description"`
	Category        string                  `json:"category"`
	Categories      []string                `json:"categories,omitempty"`
	PrimarySeverity string                  `json:"primarySeverity,omitempty"`
	Groups          []AlertRuleGroupSummary `json:"groups,omitempty"`
	Labels          map[string]string       `json:"labels,omitempty"`
	Annotations     map[string]string       `json:"annotations,omitempty"`
	Source          string                  `json:"source"`    // chart 或 embedded
	File            string                  `json:"file"`      // 来源文件名
	Size            int64                   `json:"size"`      // 文件大小
	UpdatedAt       string                  `json:"updatedAt"` // 更新时间
}

// AlertRuleGroupSummary 表示告警规则组的摘要信息
// 字段名与前端 AlertRuleGroupSummary 接口匹配
type AlertRuleGroupSummary struct {
	Name       string   `json:"name"`
	Rules      int      `json:"rules"`                // 前端使用 rules 而非 rulesCount
	Interval   string   `json:"interval,omitempty"`   // 规则评估间隔
	Severities []string `json:"severities,omitempty"` // 组内规则的严重级别列表
}

// AlertRuleTemplateDetail 表示告警规则模板的详细信息
type AlertRuleTemplateDetail struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Content     string `json:"content"`
}

// ruleValidationOutcome 表示规则验证结果
type ruleValidationOutcome struct {
	Success  bool
	Message  string
	Details  []map[string]string
	Errors   []string
	Warnings []string
}

// ======================== Embedded Templates ========================

//go:embed templates/*.yaml
var embeddedTemplates embed.FS

// ======================== Constants ========================

const (
	// 模板目录在 chart 中的相对路径
	// 注意：PrometheusRule 模板直接在 templates 目录下，不是 alertrules 子目录
	alertRulesDir = "charts/polardbx-monitor/templates"
)

var (
	// prometheusRuleGVR 定义 PrometheusRule 资源的 GroupVersionResource
	prometheusRuleGVR = schema.GroupVersionResource{
		Group:    "monitoring.coreos.com",
		Version:  "v1",
		Resource: "prometheusrules",
	}

	// 模板目录缓存
	templateDirCache    string
	templateDirCacheMu  sync.RWMutex
	templateDirResolved bool
)

// ======================== Handler ========================

// PrometheusRuleHandler 处理告警规则相关的请求
type PrometheusRuleHandler struct {
	dynamicClient dynamic.Interface
}

// NewPrometheusRuleHandler 创建新的 PrometheusRuleHandler 实例
func NewPrometheusRuleHandler(dynamicClient dynamic.Interface) *PrometheusRuleHandler {
	return &PrometheusRuleHandler{
		dynamicClient: dynamicClient,
	}
}

// ======================== PrometheusRule CRUD ========================

// List 获取 PrometheusRule 列表
func (h *PrometheusRuleHandler) List(c *gin.Context) {
	namespace := c.Query("namespace")

	ctx := c.Request.Context()

	var list *unstructured.UnstructuredList
	var err error

	if namespace != "" {
		list, err = h.dynamicClient.Resource(prometheusRuleGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	} else {
		list, err = h.dynamicClient.Resource(prometheusRuleGVR).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to list PrometheusRules: %v", err),
		})
		return
	}

	rules := make([]PrometheusRule, 0, len(list.Items))
	for _, item := range list.Items {
		rule := parsePrometheusRule(item)
		rules = append(rules, rule)
	}

	c.JSON(http.StatusOK, rules)
}

// GetYAML 获取 PrometheusRule 的 YAML 内容
func (h *PrometheusRuleHandler) GetYAML(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "namespace and name are required",
		})
		return
	}

	ctx := c.Request.Context()

	obj, err := h.dynamicClient.Resource(prometheusRuleGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get PrometheusRule: %v", err),
		})
		return
	}

	// 移除 managedFields 以获得更干净的输出
	unstructured.RemoveNestedField(obj.Object, "metadata", "managedFields")
	unstructured.RemoveNestedField(obj.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(obj.Object, "metadata", "uid")
	unstructured.RemoveNestedField(obj.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(obj.Object, "metadata", "generation")

	yamlBytes, err := yaml.Marshal(obj.Object)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to marshal to YAML: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"yaml": string(yamlBytes),
	})
}

// ValidateRule 验证 PrometheusRule YAML 内容
func (h *PrometheusRuleHandler) ValidateRule(c *gin.Context) {
	var body struct {
		YAML string `json:"yaml" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: yaml field is required",
		})
		return
	}

	_, outcome := runRuleValidation(body.YAML)

	c.JSON(http.StatusOK, gin.H{
		"success":  outcome.Success,
		"message":  outcome.Message,
		"details":  outcome.Details,
		"errors":   outcome.Errors,
		"warnings": outcome.Warnings,
	})
}

// ======================== Template Management ========================

// ListTemplates 列出所有告警规则模板
func (h *PrometheusRuleHandler) ListTemplates(c *gin.Context) {
	templates, err := loadAllTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to load templates: %v", err),
		})
		return
	}

	// 获取模板目录路径用于前端显示
	directory := ""
	if dir, err := resolveAlertTemplateDir(); err == nil {
		directory = dir
	}

	// 返回前端期望的格式: { items: [...], directory: "..." }
	c.JSON(http.StatusOK, gin.H{
		"items":     templates,
		"directory": directory,
	})
}

// GetTemplate 获取指定模板的详细信息
func (h *PrometheusRuleHandler) GetTemplate(c *gin.Context) {
	templateName := c.Param("name")
	if templateName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "template name is required",
		})
		return
	}

	template, err := loadTemplateDetail(templateName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Template not found: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, template)
}

// ApplyTemplate 应用模板到集群
func (h *PrometheusRuleHandler) ApplyTemplate(c *gin.Context) {
	var body struct {
		TemplateName string `json:"templateName" binding:"required"`
		Namespace    string `json:"namespace" binding:"required"`
		Name         string `json:"name"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// 获取模板详情
	template, err := loadTemplateDetail(body.TemplateName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Template not found: %v", err),
		})
		return
	}

	// 解析 YAML 内容
	var obj unstructured.Unstructured
	if err := yaml.Unmarshal([]byte(template.Content), &obj.Object); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to parse template: %v", err),
		})
		return
	}

	// 设置 namespace
	obj.SetNamespace(body.Namespace)

	// 如果提供了自定义名称，则使用它
	if body.Name != "" {
		obj.SetName(body.Name)
	}

	ctx := c.Request.Context()

	// 尝试创建或更新
	existing, err := h.dynamicClient.Resource(prometheusRuleGVR).Namespace(body.Namespace).Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err == nil {
		// 资源已存在，进行更新
		obj.SetResourceVersion(existing.GetResourceVersion())
		_, err = h.dynamicClient.Resource(prometheusRuleGVR).Namespace(body.Namespace).Update(ctx, &obj, metav1.UpdateOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to update PrometheusRule: %v", err),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "PrometheusRule updated successfully",
			"name":    obj.GetName(),
		})
	} else {
		// 创建新资源
		_, err = h.dynamicClient.Resource(prometheusRuleGVR).Namespace(body.Namespace).Create(ctx, &obj, metav1.CreateOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to create PrometheusRule: %v", err),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "PrometheusRule created successfully",
			"name":    obj.GetName(),
		})
	}
}

// ======================== Helper Functions ========================

// parsePrometheusRule 从 unstructured 对象解析 PrometheusRule
func parsePrometheusRule(item unstructured.Unstructured) PrometheusRule {
	rule := PrometheusRule{
		Name:      item.GetName(),
		Namespace: item.GetNamespace(),
		Labels:    item.GetLabels(),
	}

	// 解析 groups
	groups, found, _ := unstructured.NestedSlice(item.Object, "spec", "groups")
	if found {
		rule.GroupsCount = len(groups)
		rule.Groups = groups

		for _, g := range groups {
			groupMap, ok := g.(map[string]interface{})
			if !ok {
				continue
			}

			rules, ok := groupMap["rules"].([]interface{})
			if !ok {
				continue
			}

			rule.RulesCount += len(rules)

			for _, r := range rules {
				ruleMap, ok := r.(map[string]interface{})
				if !ok {
					continue
				}

				if _, hasAlert := ruleMap["alert"]; hasAlert {
					rule.AlertsCount++
				}
				if _, hasRecord := ruleMap["record"]; hasRecord {
					rule.RecordingCount++
				}
			}
		}
	}

	return rule
}

// runRuleValidation 验证 PrometheusRule YAML 内容
func runRuleValidation(yamlContent string) (*unstructured.Unstructured, ruleValidationOutcome) {
	trimmed := strings.TrimSpace(yamlContent)
	if trimmed == "" {
		return nil, ruleValidationOutcome{
			Success: false,
			Message: "YAML content is empty",
			Details: []map[string]string{{
				"level":   "error",
				"message": "YAML content is required",
			}},
			Errors: []string{"YAML content is required"},
		}
	}

	var obj unstructured.Unstructured
	if err := yaml.Unmarshal([]byte(trimmed), &obj.Object); err != nil {
		errMsg := fmt.Sprintf("Invalid YAML format: %v", err)
		return nil, ruleValidationOutcome{
			Success: false,
			Message: errMsg,
			Details: []map[string]string{{
				"level":   "error",
				"message": errMsg,
			}},
			Errors: []string{errMsg},
		}
	}

	errors := []string{}
	warnings := []string{}
	details := []map[string]string{}

	// 验证 apiVersion 和 kind
	if obj.GetAPIVersion() != "monitoring.coreos.com/v1" {
		errors = append(errors, "apiVersion should be 'monitoring.coreos.com/v1'")
	}
	if obj.GetKind() != "PrometheusRule" {
		errors = append(errors, "kind should be 'PrometheusRule'")
	}
	if obj.GetName() == "" {
		errors = append(errors, "metadata.name is required")
	}

	// 验证 groups
	groups, found, _ := unstructured.NestedSlice(obj.Object, "spec", "groups")
	if !found {
		errors = append(errors, "spec.groups is required")
	} else if len(groups) == 0 {
		warnings = append(warnings, "No rule groups defined")
	}

	for i, groupInterface := range groups {
		groupMap, ok := groupInterface.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Sprintf("Group %d: invalid structure", i))
			continue
		}

		name, hasName := groupMap["name"].(string)
		if !hasName || strings.TrimSpace(name) == "" {
			errors = append(errors, fmt.Sprintf("Group %d: name is required", i))
			continue
		}

		rulesInterface, ok := groupMap["rules"].([]interface{})
		if !ok {
			errors = append(errors, fmt.Sprintf("Group '%s': rules field must be an array", name))
			continue
		}
		if len(rulesInterface) == 0 {
			warnings = append(warnings, fmt.Sprintf("Group '%s': No rules defined", name))
		}

		for j, ruleInterface := range rulesInterface {
			ruleMap, ok := ruleInterface.(map[string]interface{})
			if !ok {
				errors = append(errors, fmt.Sprintf("Group '%s', Rule %d: invalid structure", name, j))
				continue
			}

			expr, hasExpr := ruleMap["expr"].(string)
			if !hasExpr || strings.TrimSpace(expr) == "" {
				errors = append(errors, fmt.Sprintf("Group '%s', Rule %d: expr is required", name, j))
			}

			_, hasAlert := ruleMap["alert"].(string)
			_, hasRecord := ruleMap["record"].(string)
			if !hasAlert && !hasRecord {
				errors = append(errors, fmt.Sprintf("Group '%s', Rule %d: either 'alert' or 'record' must be specified", name, j))
			}
			if hasAlert && hasRecord {
				errors = append(errors, fmt.Sprintf("Group '%s', Rule %d: cannot have both 'alert' and 'record'", name, j))
			}

			if hasExpr && strings.TrimSpace(expr) != "" && !IsValidPromQLBasic(expr) {
				warnings = append(warnings, fmt.Sprintf("Group '%s', Rule %d: potentially invalid PromQL expression", name, j))
			}
		}
	}

	// 构建 details
	for _, err := range errors {
		details = append(details, map[string]string{
			"level":   "error",
			"message": err,
		})
	}
	for _, warn := range warnings {
		details = append(details, map[string]string{
			"level":   "warning",
			"message": warn,
		})
	}

	success := len(errors) == 0
	message := "Validation passed"
	if !success {
		message = "Validation failed"
	} else if len(warnings) > 0 {
		message = "Validation passed with warnings"
	}

	return &obj, ruleValidationOutcome{
		Success:  success,
		Message:  message,
		Details:  details,
		Errors:   errors,
		Warnings: warnings,
	}
}

// IsValidPromQLBasic 执行基本的 PromQL 语法检查（导出用于测试兼容）
func IsValidPromQLBasic(expr string) bool {
	// 基本检查：括号平衡
	parenCount := 0
	braceCount := 0
	bracketCount := 0

	for _, ch := range expr {
		switch ch {
		case '(':
			parenCount++
		case ')':
			parenCount--
		case '{':
			braceCount++
		case '}':
			braceCount--
		case '[':
			bracketCount++
		case ']':
			bracketCount--
		}

		if parenCount < 0 || braceCount < 0 || bracketCount < 0 {
			return false
		}
	}

	return parenCount == 0 && braceCount == 0 && bracketCount == 0
}

// ======================== Template Loading ========================

// loadAllTemplates 加载所有模板的摘要信息
func loadAllTemplates() ([]AlertRuleTemplateSummary, error) {
	// 首先尝试从文件系统加载 chart 模板
	chartTemplates, err := LoadChartTemplates()
	if err == nil && len(chartTemplates) > 0 {
		return chartTemplates, nil
	}

	// 回退到嵌入的模板
	return loadEmbeddedTemplates()
}

// LoadChartTemplates 从 chart 目录加载模板（导出用于测试兼容）
func LoadChartTemplates() ([]AlertRuleTemplateSummary, error) {
	templateDir, err := resolveAlertTemplateDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(templateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read template directory: %w", err)
	}

	templates := []AlertRuleTemplateSummary{}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		filePath := filepath.Join(templateDir, entry.Name())

		// 获取文件信息
		fileInfo, err := os.Stat(filePath)
		var fileSize int64
		var updatedAt string
		if err == nil {
			fileSize = fileInfo.Size()
			updatedAt = fileInfo.ModTime().Format("2006-01-02T15:04:05Z")
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		// 处理 Helm 模板内容
		cleanContent := sanitizeHelmPlaceholders(string(content))

		// 解析可能包含多个文档的 YAML
		docs := splitYAMLDocuments(cleanContent)

		for _, doc := range docs {
			template, err := parseTemplateFromYAMLWithInfo(doc, entry.Name(), "chart", fileSize, updatedAt)
			if err != nil {
				continue
			}
			templates = append(templates, template)
		}
	}

	// 按名称排序
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})

	return templates, nil
}

// loadEmbeddedTemplates 从嵌入资源加载模板
func loadEmbeddedTemplates() ([]AlertRuleTemplateSummary, error) {
	templates := []AlertRuleTemplateSummary{}

	err := fs.WalkDir(embeddedTemplates, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".yaml") {
			return nil
		}

		content, err := embeddedTemplates.ReadFile(path)
		if err != nil {
			return nil
		}

		// 获取嵌入文件信息
		fileInfo, _ := d.Info()
		var fileSize int64
		var updatedAt string
		if fileInfo != nil {
			fileSize = fileInfo.Size()
			updatedAt = fileInfo.ModTime().Format("2006-01-02T15:04:05Z")
		}

		template, err := parseTemplateFromYAMLWithInfo(string(content), d.Name(), "embedded", fileSize, updatedAt)
		if err != nil {
			return nil
		}

		templates = append(templates, template)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return templates, nil
}

// loadTemplateDetail 加载模板详细信息
func loadTemplateDetail(templateName string) (*AlertRuleTemplateDetail, error) {
	// 首先尝试从 chart 加载
	templateDir, err := resolveAlertTemplateDir()
	if err == nil {
		entries, err := os.ReadDir(templateDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
					continue
				}

				filePath := filepath.Join(templateDir, entry.Name())
				content, err := os.ReadFile(filePath)
				if err != nil {
					continue
				}

				cleanContent := sanitizeHelmPlaceholders(string(content))
				docs := splitYAMLDocuments(cleanContent)

				for _, doc := range docs {
					var obj map[string]interface{}
					if err := yaml.Unmarshal([]byte(doc), &obj); err != nil {
						continue
					}

					metadata, _ := obj["metadata"].(map[string]interface{})
					name, _ := metadata["name"].(string)

					if name == templateName {
						return &AlertRuleTemplateDetail{
							Name:        name,
							DisplayName: formatDisplayName(name),
							Description: extractDescription(obj),
							Category:    extractCategory(entry.Name()),
							Content:     doc,
						}, nil
					}
				}
			}
		}
	}

	// 回退到嵌入模板
	return loadEmbeddedTemplateDetail(templateName)
}

// loadEmbeddedTemplateDetail 从嵌入资源加载模板详情
func loadEmbeddedTemplateDetail(templateName string) (*AlertRuleTemplateDetail, error) {
	var result *AlertRuleTemplateDetail

	err := fs.WalkDir(embeddedTemplates, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".yaml") {
			return nil
		}

		content, err := embeddedTemplates.ReadFile(path)
		if err != nil {
			return nil
		}

		var obj map[string]interface{}
		if err := yaml.Unmarshal(content, &obj); err != nil {
			return nil
		}

		metadata, _ := obj["metadata"].(map[string]interface{})
		name, _ := metadata["name"].(string)

		if name == templateName {
			result = &AlertRuleTemplateDetail{
				Name:        name,
				DisplayName: formatDisplayName(name),
				Description: extractDescription(obj),
				Category:    extractCategory(d.Name()),
				Content:     string(content),
			}
			return fs.SkipAll
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}

	return result, nil
}

// resolveAlertTemplateDir 解析告警模板目录路径
func resolveAlertTemplateDir() (string, error) {
	templateDirCacheMu.RLock()
	if templateDirResolved {
		dir := templateDirCache
		templateDirCacheMu.RUnlock()
		if dir == "" {
			return "", fmt.Errorf("template directory not found")
		}
		return dir, nil
	}
	templateDirCacheMu.RUnlock()

	templateDirCacheMu.Lock()
	defer templateDirCacheMu.Unlock()

	// 双重检查
	if templateDirResolved {
		if templateDirCache == "" {
			return "", fmt.Errorf("template directory not found")
		}
		return templateDirCache, nil
	}

	// 尝试不同的路径
	possiblePaths := []string{
		alertRulesDir,
		filepath.Join("..", alertRulesDir),
		filepath.Join("..", "..", alertRulesDir),
		"/app/charts/polardbx-monitor/templates",
	}

	// 添加基于可执行文件位置的路径
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		possiblePaths = append(possiblePaths,
			filepath.Join(exeDir, alertRulesDir),
			filepath.Join(exeDir, "..", alertRulesDir),
		)
	}

	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			templateDirCache = path
			templateDirResolved = true
			return path, nil
		}
	}

	templateDirResolved = true
	templateDirCache = ""
	return "", fmt.Errorf("template directory not found in any of the expected locations")
}

// splitYAMLDocuments 分割多文档 YAML
func splitYAMLDocuments(content string) []string {
	docs := []string{}
	separator := regexp.MustCompile(`(?m)^---\s*$`)
	parts := separator.Split(content, -1)

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			docs = append(docs, trimmed)
		}
	}

	return docs
}

// sanitizeHelmPlaceholders 清理 Helm 模板占位符
func sanitizeHelmPlaceholders(content string) string {
	// 处理条件语句块 - 移除整个 if/else/end 结构
	ifPattern := regexp.MustCompile(`\{\{-?\s*if[^}]*\}\}`)
	elsePattern := regexp.MustCompile(`\{\{-?\s*else[^}]*\}\}`)
	endPattern := regexp.MustCompile(`\{\{-?\s*end\s*-?\}\}`)
	rangePattern := regexp.MustCompile(`\{\{-?\s*range[^}]*\}\}`)
	withPattern := regexp.MustCompile(`\{\{-?\s*with[^}]*\}\}`)

	result := content
	result = ifPattern.ReplaceAllString(result, "")
	result = elsePattern.ReplaceAllString(result, "")
	result = endPattern.ReplaceAllString(result, "")
	result = rangePattern.ReplaceAllString(result, "")
	result = withPattern.ReplaceAllString(result, "")

	// 处理 include 语句
	includePattern := regexp.MustCompile(`\{\{-?\s*include\s+"[^"]*"\s*\.\s*\|\s*nindent\s+\d+\s*-?\}\}`)
	result = includePattern.ReplaceAllString(result, "")

	// 处理简单值替换 {{ .Values.xxx }}
	valuePattern := regexp.MustCompile(`\{\{[^}]+\}\}`)
	result = valuePattern.ReplaceAllStringFunc(result, func(match string) string {
		// 保留一些常见的默认值
		if strings.Contains(match, ".Release.Namespace") {
			return "default"
		}
		if strings.Contains(match, ".Release.Name") {
			return "polardbx-monitor"
		}
		return ""
	})

	return result
}

// parseTemplateFromYAML 从 YAML 内容解析模板摘要
func parseTemplateFromYAML(content string, filename string) (AlertRuleTemplateSummary, error) {
	return parseTemplateFromYAMLWithInfo(content, filename, "chart", 0, "")
}

// parseTemplateFromYAMLWithInfo 从 YAML 内容解析模板摘要（带文件信息）
func parseTemplateFromYAMLWithInfo(content string, filename string, source string, size int64, updatedAt string) (AlertRuleTemplateSummary, error) {
	var obj map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &obj); err != nil {
		return AlertRuleTemplateSummary{}, err
	}

	// 验证是 PrometheusRule
	kind, _ := obj["kind"].(string)
	if kind != "PrometheusRule" {
		return AlertRuleTemplateSummary{}, fmt.Errorf("not a PrometheusRule")
	}

	metadata, _ := obj["metadata"].(map[string]interface{})
	name, _ := metadata["name"].(string)
	if name == "" {
		return AlertRuleTemplateSummary{}, fmt.Errorf("missing name")
	}

	displayName := formatDisplayName(name)
	category := extractCategory(filename)

	// 提取 labels 和 annotations
	labels := make(map[string]string)
	if labelsRaw, ok := metadata["labels"].(map[string]interface{}); ok {
		for k, v := range labelsRaw {
			if s, ok := v.(string); ok {
				labels[k] = s
			}
		}
	}
	annotations := make(map[string]string)
	if annotationsRaw, ok := metadata["annotations"].(map[string]interface{}); ok {
		for k, v := range annotationsRaw {
			if s, ok := v.(string); ok {
				annotations[k] = s
			}
		}
	}

	template := AlertRuleTemplateSummary{
		Name:            name,
		DisplayName:     displayName,
		Title:           displayName, // 前端使用 title 字段
		Description:     extractDescription(obj),
		Category:        category,
		Categories:      []string{category},
		PrimarySeverity: extractPrimarySeverity(obj),
		Groups:          []AlertRuleGroupSummary{},
		Labels:          labels,
		Annotations:     annotations,
		Source:          source,
		File:            filename,
		Size:            size,
		UpdatedAt:       updatedAt,
	}

	// 解析 groups
	spec, _ := obj["spec"].(map[string]interface{})
	groups, _ := spec["groups"].([]interface{})

	for _, g := range groups {
		groupMap, ok := g.(map[string]interface{})
		if !ok {
			continue
		}

		groupName, _ := groupMap["name"].(string)
		interval, _ := groupMap["interval"].(string)
		rules, _ := groupMap["rules"].([]interface{})

		// 收集该组内的所有严重级别
		severitySet := make(map[string]bool)
		for _, r := range rules {
			ruleMap, ok := r.(map[string]interface{})
			if !ok {
				continue
			}
			ruleLabels, _ := ruleMap["labels"].(map[string]interface{})
			if sev, ok := ruleLabels["severity"].(string); ok && sev != "" {
				severitySet[sev] = true
			}
		}
		severities := make([]string, 0, len(severitySet))
		for sev := range severitySet {
			severities = append(severities, sev)
		}

		template.Groups = append(template.Groups, AlertRuleGroupSummary{
			Name:       groupName,
			Rules:      len(rules),
			Interval:   interval,
			Severities: severities,
		})
	}

	return template, nil
}

// extractPrimarySeverity 从对象中提取主要严重级别
func extractPrimarySeverity(obj map[string]interface{}) string {
	spec, _ := obj["spec"].(map[string]interface{})
	groups, _ := spec["groups"].([]interface{})

	severityCounts := make(map[string]int)
	for _, g := range groups {
		groupMap, ok := g.(map[string]interface{})
		if !ok {
			continue
		}
		rules, _ := groupMap["rules"].([]interface{})
		for _, r := range rules {
			ruleMap, ok := r.(map[string]interface{})
			if !ok {
				continue
			}
			labels, _ := ruleMap["labels"].(map[string]interface{})
			if sev, ok := labels["severity"].(string); ok {
				severityCounts[sev]++
			}
		}
	}

	// 返回最常见的严重级别
	maxCount := 0
	primarySev := ""
	for sev, count := range severityCounts {
		if count > maxCount {
			maxCount = count
			primarySev = sev
		}
	}
	return primarySev
}

// formatDisplayName 格式化显示名称
func formatDisplayName(name string) string {
	// 移除常见前缀
	name = strings.TrimPrefix(name, "polardbx-")
	name = strings.TrimPrefix(name, "pxc-")

	// 转换为标题格式
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}

	return strings.Join(parts, " ")
}

// extractDescription 从对象中提取描述
func extractDescription(obj map[string]interface{}) string {
	metadata, _ := obj["metadata"].(map[string]interface{})
	annotations, _ := metadata["annotations"].(map[string]interface{})

	if desc, ok := annotations["description"].(string); ok {
		return desc
	}

	// 尝试从名称生成描述
	name, _ := metadata["name"].(string)
	return fmt.Sprintf("Alert rules for %s", formatDisplayName(name))
}

// extractCategory 从文件名提取分类
func extractCategory(filename string) string {
	// 根据文件名推断分类
	lower := strings.ToLower(filename)

	categories := map[string]string{
		"cn":      "CN (Compute Node)",
		"dn":      "DN (Data Node)",
		"gms":     "GMS (Global Meta Service)",
		"cdc":     "CDC (Change Data Capture)",
		"storage": "Storage",
		"cluster": "Cluster",
	}

	for key, category := range categories {
		if strings.Contains(lower, key) {
			return category
		}
	}

	return "General"
}

// ======================== Context-based Functions ========================

// ListWithContext 使用 context 获取 PrometheusRule 列表（用于内部调用）
func (h *PrometheusRuleHandler) ListWithContext(ctx context.Context, namespace string) ([]PrometheusRule, error) {
	var list *unstructured.UnstructuredList
	var err error

	if namespace != "" {
		list, err = h.dynamicClient.Resource(prometheusRuleGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	} else {
		list, err = h.dynamicClient.Resource(prometheusRuleGVR).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, err
	}

	rules := make([]PrometheusRule, 0, len(list.Items))
	for _, item := range list.Items {
		rule := parsePrometheusRule(item)
		rules = append(rules, rule)
	}

	return rules, nil
}

// GetYAMLWithContext 使用 context 获取 YAML（用于内部调用）
func (h *PrometheusRuleHandler) GetYAMLWithContext(ctx context.Context, namespace, name string) (string, error) {
	obj, err := h.dynamicClient.Resource(prometheusRuleGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	unstructured.RemoveNestedField(obj.Object, "metadata", "managedFields")
	unstructured.RemoveNestedField(obj.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(obj.Object, "metadata", "uid")
	unstructured.RemoveNestedField(obj.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(obj.Object, "metadata", "generation")

	yamlBytes, err := yaml.Marshal(obj.Object)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

// ValidateWithContext 使用 context 验证规则（用于内部调用）
func (h *PrometheusRuleHandler) ValidateWithContext(yamlContent string) ruleValidationOutcome {
	_, outcome := runRuleValidation(yamlContent)
	return outcome
}

// ApplyTemplateWithContext 使用 context 应用模板（用于内部调用）
func (h *PrometheusRuleHandler) ApplyTemplateWithContext(ctx context.Context, templateName, namespace, name string) error {
	template, err := loadTemplateDetail(templateName)
	if err != nil {
		return err
	}

	var obj unstructured.Unstructured
	if err := yaml.Unmarshal([]byte(template.Content), &obj.Object); err != nil {
		return err
	}

	obj.SetNamespace(namespace)
	if name != "" {
		obj.SetName(name)
	}

	existing, err := h.dynamicClient.Resource(prometheusRuleGVR).Namespace(namespace).Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err == nil {
		obj.SetResourceVersion(existing.GetResourceVersion())
		_, err = h.dynamicClient.Resource(prometheusRuleGVR).Namespace(namespace).Update(ctx, &obj, metav1.UpdateOptions{})
	} else {
		_, err = h.dynamicClient.Resource(prometheusRuleGVR).Namespace(namespace).Create(ctx, &obj, metav1.CreateOptions{})
	}

	return err
}
