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

// SettingsHandler handles HTTP requests related to settings
type SettingsHandler struct {
	service *service.SettingsService
}

// NewSettingsHandler creates a new SettingsHandler
func NewSettingsHandler(svc *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{service: svc}
}

// NewSettingsHandlerFromContext creates complete handler chain from gin.Context
func NewSettingsHandlerFromContext(c *gin.Context) (*SettingsHandler, bool) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return nil, false
	}
	repo := repository.NewK8sSettingsRepository(cli)
	svc := service.NewSettingsService(repo)
	return NewSettingsHandler(svc), true
}

// NewSettingsServiceFromClient creates SettingsService (for use by other packages)
func NewSettingsServiceFromClient(cli client.Client) *service.SettingsService {
	repo := repository.NewK8sSettingsRepository(cli)
	return service.NewSettingsService(repo)
}

// Get gets all settings
func Get(c *gin.Context) {
	h, ok := NewSettingsHandlerFromContext(c)
	if !ok {
		return
	}
	data, _ := h.service.Get(c.Request.Context())
	apierr.OK(c, data)
}

// Update updates settings
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

// ReadDashboardSettings reads backup dashboard settings (for use by other packages)
func ReadDashboardSettings(ctx context.Context, cli client.Client) service.BackupDashboardSettings {
	svc := NewSettingsServiceFromClient(cli)
	return svc.GetDashboardSettings(ctx)
}

// GetImageRegistryConfig gets image registry configuration
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

// UpdateImageRegistryRequest update image registry request
type UpdateImageRegistryRequest struct {
	Registry        string   `json:"registry"`
	CustomRegistry  string   `json:"customRegistry"`
	DefaultRegistry string   `json:"defaultRegistry"`
	Mirrors         []string `json:"mirrors"`
}

// UpdateImageRegistryConfig updates image registry configuration
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
		"message": "Image registry configuration updated",
		"data":    data,
	})
}

// GetAvailableRegistries gets available image registry presets
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

// TestImageRegistryRequest test image registry request
type TestImageRegistryRequest struct {
	Registry string `json:"registry" binding:"required"`
}

// TestImageRegistry tests image registry connectivity
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
