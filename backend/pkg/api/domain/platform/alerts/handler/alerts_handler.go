package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"polardbx-ui-backend/pkg/api/domain/platform/alerts/repository"
	"polardbx-ui-backend/pkg/api/domain/platform/alerts/service"
	"polardbx-ui-backend/pkg/api/util"
)

// AlertsHandler 处理告警相关的 HTTP 请求
type AlertsHandler struct {
	service *service.AlertsService
}

// NewAlertsHandler 创建新的 AlertsHandler
func NewAlertsHandler(svc *service.AlertsService) *AlertsHandler {
	return &AlertsHandler{service: svc}
}

// NewAlertsHandlerFromContext 从 gin.Context 创建完整的 handler 链
func NewAlertsHandlerFromContext(c *gin.Context) (*AlertsHandler, bool) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return nil, false
	}
	repo := repository.NewK8sAlertsRepository(cli)
	svc := service.NewAlertsService(repo)
	return NewAlertsHandler(svc), true
}

// ListProfiles 列出所有配置文件
func ListProfiles(c *gin.Context) {
	h, ok := NewAlertsHandlerFromContext(c)
	if !ok {
		return
	}
	items, _ := h.service.ListProfiles(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// CreateProfile 创建配置文件
func CreateProfile(c *gin.Context) {
	h, ok := NewAlertsHandlerFromContext(c)
	if !ok {
		return
	}
	var body struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if err := h.service.CreateProfile(c.Request.Context(), body.Name, body.Content); err != nil {
		if err == service.ErrProfileExists {
			c.JSON(http.StatusConflict, gin.H{"error": "profile exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create profiles cm", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"name": body.Name})
}

// GetProfile 获取配置文件
func GetProfile(c *gin.Context) {
	h, ok := NewAlertsHandlerFromContext(c)
	if !ok {
		return
	}
	name := c.Param("name")
	profile, err := h.service.GetProfile(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"name": profile.Name, "content": profile.Content})
}

// UpdateProfile 更新配置文件
func UpdateProfile(c *gin.Context) {
	h, ok := NewAlertsHandlerFromContext(c)
	if !ok {
		return
	}
	name := c.Param("name")
	var body struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if err := h.service.UpdateProfile(c.Request.Context(), name, body.Content); err != nil {
		if err == service.ErrProfileNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "profiles not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update profiles cm", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"name": name})
}

// DeleteProfile 删除配置文件
func DeleteProfile(c *gin.Context) {
	h, ok := NewAlertsHandlerFromContext(c)
	if !ok {
		return
	}
	name := c.Param("name")
	if err := h.service.DeleteProfile(c.Request.Context(), name); err != nil {
		if err == service.ErrProfileNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update profiles cm", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": name})
}

// DryRunProfile 验证 Alertmanager YAML
func DryRunProfile(c *gin.Context) {
	h, ok := NewAlertsHandlerFromContext(c)
	if !ok {
		return
	}
	var body struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	result, err := h.service.DryRunProfile(body.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tempfile"})
		return
	}
	if !result.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"valid": false, "details": result.Details})
		return
	}
	c.JSON(http.StatusOK, gin.H{"valid": true})
}

// GetRoutes 获取路由配置
func GetRoutes(c *gin.Context) {
	h, ok := NewAlertsHandlerFromContext(c)
	if !ok {
		return
	}
	content, _ := h.service.GetRoutes(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"content": content})
}

// PutRoutes 更新路由配置
func PutRoutes(c *gin.Context) {
	h, ok := NewAlertsHandlerFromContext(c)
	if !ok {
		return
	}
	var body struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	_ = h.service.PutRoutes(c.Request.Context(), body.Content)
	c.JSON(http.StatusOK, gin.H{"message": "routes updated"})
}

// List 聚合告警列表
func List(c *gin.Context) {
	h, ok := NewAlertsHandlerFromContext(c)
	if !ok {
		return
	}
	namespace := c.DefaultQuery("namespace", "")
	cluster := c.DefaultQuery("cluster", "")
	alertmanagerURL := c.Query("alertmanager")
	items, _ := h.service.ListAlerts(c.Request.Context(), namespace, cluster, alertmanagerURL)
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// ListSilences Alertmanager 静默列表代理
func ListSilences(c *gin.Context) {
	_, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	base := strings.TrimRight(c.DefaultQuery("alertmanager", ""), "/")
	if base == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alertmanager required"})
		return
	}
	resp, err := http.Get(base + "/api/v2/silences")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	var out any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	c.JSON(resp.StatusCode, out)
}

// CreateSilence 创建静默
func CreateSilence(c *gin.Context) {
	_, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	base := strings.TrimRight(c.DefaultQuery("alertmanager", ""), "/")
	if base == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alertmanager required"})
		return
	}
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(base+"/api/v2/silences", "application/json", bytes.NewReader(b))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	var out any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	c.JSON(resp.StatusCode, out)
}

// DeleteSilence 删除静默
func DeleteSilence(c *gin.Context) {
	_, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	base := strings.TrimRight(c.DefaultQuery("alertmanager", ""), "/")
	if base == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alertmanager required"})
		return
	}
	id := c.Param("id")
	req, _ := http.NewRequest(http.MethodDelete, base+"/api/v2/silence/"+url.PathEscape(id), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	var out any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	c.JSON(resp.StatusCode, out)
}

// TestAlert 发送测试告警
func TestAlert(c *gin.Context) {
	_, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	base := strings.TrimRight(c.DefaultQuery("alertmanager", ""), "/")
	if base == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alertmanager required"})
		return
	}
	labels := c.DefaultQuery("labels", "severity=warning,service=test")
	var pairs []string
	for _, p := range strings.Split(labels, ",") {
		if strings.Contains(p, "=") {
			pairs = append(pairs, p)
		}
	}
	alert := []map[string]any{{"labels": map[string]string{}}}
	for _, kv := range pairs {
		parts := strings.SplitN(kv, "=", 2)
		alert[0]["labels"].(map[string]string)[parts[0]] = parts[1]
	}
	b, _ := json.Marshal(alert)
	resp, err := http.Post(base+"/api/v1/alerts", "application/json", bytes.NewReader(b))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	var out any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	c.JSON(resp.StatusCode, out)
}
