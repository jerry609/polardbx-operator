package handler

import (
	"polardbx-ui-backend/pkg/api/domain/platform/system/repository"
	"polardbx-ui-backend/pkg/api/domain/platform/system/service"
	apierr "polardbx-ui-backend/pkg/api/errors"
	"polardbx-ui-backend/pkg/api/util"

	"github.com/gin-gonic/gin"
)

// SystemHandler handles HTTP requests related to system
type SystemHandler struct {
	service *service.SystemService
}

// NewSystemHandler creates a new SystemHandler
func NewSystemHandler(svc *service.SystemService) *SystemHandler {
	return &SystemHandler{service: svc}
}

// NewSystemHandlerFromClient creates complete handler chain from K8s client
func NewSystemHandlerFromClient(c *gin.Context) (*SystemHandler, bool) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return nil, false
	}
	repo := repository.NewK8sSystemRepository(cli)
	svc := service.NewSystemService(repo)
	return NewSystemHandler(svc), true
}

// ContextInfo returns current k8s user, context and default namespace
func ContextInfo(c *gin.Context) {
	_, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	user, _ := c.Get("k8sUser")
	ctxName, _ := c.Get("k8sContext")
	defNS := c.DefaultQuery("namespace", "")
	if defNS == "" {
		if v, exists := c.Get("k8sDefaultNamespace"); exists {
			if s, ok := v.(string); ok {
				defNS = s
			}
		}
	}
	apierr.OK(c, gin.H{"user": user, "context": ctxName, "defaultNamespace": defNS})
}

// ListNamespaces returns all visible namespaces
func ListNamespaces(c *gin.Context) {
	h, ok := NewSystemHandlerFromClient(c)
	if !ok {
		// Return empty list to avoid frontend errors
		apierr.OK(c, gin.H{"items": []any{}, "count": 0, "warning": "k8s client not initialized"})
		return
	}
	h.listNamespaces(c)
}

// listNamespaces instance method handles namespace list
func (h *SystemHandler) listNamespaces(c *gin.Context) {
	items, err := h.service.ListNamespaces(c.Request.Context())
	if err != nil {
		apierr.OK(c, gin.H{"items": []any{}, "count": 0, "warning": "failed to list namespaces", "details": err.Error()})
		return
	}
	apierr.OK(c, gin.H{"items": items, "count": len(items)})
}

// ListStorageClasses returns all available storage classes
func ListStorageClasses(c *gin.Context) {
	h, ok := NewSystemHandlerFromClient(c)
	if !ok {
		apierr.OK(c, gin.H{"items": []any{}, "count": 0, "warning": "k8s client not initialized"})
		return
	}
	h.listStorageClasses(c)
}

// listStorageClasses instance method handles storage class list
func (h *SystemHandler) listStorageClasses(c *gin.Context) {
	items, err := h.service.ListStorageClasses(c.Request.Context())
	if err != nil {
		apierr.OK(c, gin.H{"items": []any{}, "count": 0, "warning": "failed to list storage classes", "details": err.Error()})
		return
	}
	apierr.OK(c, gin.H{"items": items, "count": len(items)})
}

// ListPolarDBXVersions returns supported PolarDB-X version list
func ListPolarDBXVersions(c *gin.Context) {
	h, ok := NewSystemHandlerFromClient(c)
	if !ok {
		// Version list can be returned even without k8s client
		svc := service.NewSystemService(nil)
		versions := svc.GetPolarDBXVersions()
		apierr.OK(c, gin.H{"items": versions, "count": len(versions)})
		return
	}
	versions := h.service.GetPolarDBXVersions()
	apierr.OK(c, gin.H{"items": versions, "count": len(versions)})
}
