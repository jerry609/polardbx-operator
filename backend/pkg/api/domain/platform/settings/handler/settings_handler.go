package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"polardbx-ui-backend/pkg/api/domain/platform/settings/repository"
	"polardbx-ui-backend/pkg/api/domain/platform/settings/service"
	apierr "polardbx-ui-backend/pkg/api/errors"
	"polardbx-ui-backend/pkg/api/util"
)

// SettingsHandler 处理设置相关的 HTTP 请求
type SettingsHandler struct {
	service *service.SettingsService
}

// NewSettingsHandler 创建新的 SettingsHandler
func NewSettingsHandler(svc *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{service: svc}
}

// NewSettingsHandlerFromContext 从 gin.Context 创建完整的 handler 链
func NewSettingsHandlerFromContext(c *gin.Context) (*SettingsHandler, bool) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return nil, false
	}
	repo := repository.NewK8sSettingsRepository(cli)
	svc := service.NewSettingsService(repo)
	return NewSettingsHandler(svc), true
}

// NewSettingsServiceFromClient 创建 SettingsService（供其他包使用）
func NewSettingsServiceFromClient(cli client.Client) *service.SettingsService {
	repo := repository.NewK8sSettingsRepository(cli)
	return service.NewSettingsService(repo)
}

// Get 获取所有设置
func Get(c *gin.Context) {
	h, ok := NewSettingsHandlerFromContext(c)
	if !ok {
		return
	}
	data, _ := h.service.Get(c.Request.Context())
	apierr.OK(c, data)
}

// Update 更新设置
func Update(c *gin.Context) {
	h, ok := NewSettingsHandlerFromContext(c)
	if !ok {
		return
	}
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.AbortValidation(c, "invalid payload: "+err.Error())
		return
	}
	if err := h.service.Update(c.Request.Context(), body); err != nil {
		apierr.AbortInternal(c, "failed to update settings: "+err.Error())
		return
	}
	apierr.OK(c, body)
}

// ReadDashboardSettings 读取备份仪表盘设置（供其他包使用）
func ReadDashboardSettings(ctx context.Context, cli client.Client) service.BackupDashboardSettings {
	svc := NewSettingsServiceFromClient(cli)
	return svc.GetDashboardSettings(ctx)
}

// GetImageRegistryConfig 获取镜像仓库配置
func GetImageRegistryConfig(c *gin.Context) {
	h, ok := NewSettingsHandlerFromContext(c)
	if !ok {
		return
	}
	apierr.OK(c, gin.H{
		"success": true,
		"data":    h.service.GetImageRegistryConfig(),
	})
}

// UpdateImageRegistryRequest 更新镜像仓库请求
type UpdateImageRegistryRequest struct {
	Registry        string   `json:"registry"`
	CustomRegistry  string   `json:"customRegistry"`
	DefaultRegistry string   `json:"defaultRegistry"`
	Mirrors         []string `json:"mirrors"`
}

// UpdateImageRegistryConfig 更新镜像仓库配置
func UpdateImageRegistryConfig(c *gin.Context) {
	h, ok := NewSettingsHandlerFromContext(c)
	if !ok {
		return
	}
	var req UpdateImageRegistryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.AbortValidation(c, "Invalid request: "+err.Error())
		return
	}

	data, err := h.service.UpdateImageRegistryConfig(req.Registry, req.CustomRegistry, req.DefaultRegistry)
	if err != nil {
		apierr.AbortValidation(c, err.Error())
		return
	}

	apierr.OK(c, gin.H{
		"success": true,
		"message": "镜像源配置已更新",
		"data":    data,
	})
}

// GetAvailableRegistries 获取可用的镜像仓库预设
func GetAvailableRegistries(c *gin.Context) {
	h, ok := NewSettingsHandlerFromContext(c)
	if !ok {
		return
	}
	apierr.OK(c, gin.H{
		"success": true,
		"data":    h.service.GetAvailableRegistries(),
	})
}

// TestImageRegistryRequest 测试镜像仓库请求
type TestImageRegistryRequest struct {
	Registry string `json:"registry" binding:"required"`
}

// TestImageRegistry 测试镜像仓库连通性
func TestImageRegistry(c *gin.Context) {
	var req TestImageRegistryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.AbortValidation(c, "Invalid request: "+err.Error())
		return
	}

	apierr.OK(c, gin.H{
		"success":   true,
		"message":   "Registry test not yet implemented",
		"registry":  req.Registry,
		"reachable": true,
	})
}
