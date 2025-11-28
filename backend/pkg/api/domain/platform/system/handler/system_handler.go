package handler

import (
	"net/http"

	"polardbx-ui-backend/pkg/api/domain/platform/system/repository"
	"polardbx-ui-backend/pkg/api/domain/platform/system/service"
	"polardbx-ui-backend/pkg/api/util"

	"github.com/gin-gonic/gin"
)

// SystemHandler 处理 system 相关的 HTTP 请求
type SystemHandler struct {
	service *service.SystemService
}

// NewSystemHandler 创建新的 SystemHandler
func NewSystemHandler(svc *service.SystemService) *SystemHandler {
	return &SystemHandler{service: svc}
}

// NewSystemHandlerFromClient 从 K8s client 创建完整的 handler 链
func NewSystemHandlerFromClient(c *gin.Context) (*SystemHandler, bool) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return nil, false
	}
	repo := repository.NewK8sSystemRepository(cli)
	svc := service.NewSystemService(repo)
	return NewSystemHandler(svc), true
}

// ContextInfo 返回当前 k8s 用户、上下文和默认命名空间
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
	c.JSON(http.StatusOK, gin.H{"user": user, "context": ctxName, "defaultNamespace": defNS})
}

// ListNamespaces 返回所有可见的命名空间
func ListNamespaces(c *gin.Context) {
	h, ok := NewSystemHandlerFromClient(c)
	if !ok {
		// 返回空列表以避免前端错误
		c.JSON(http.StatusOK, gin.H{"items": []any{}, "count": 0, "warning": "k8s client not initialized"})
		return
	}
	h.listNamespaces(c)
}

// listNamespaces 实例方法处理命名空间列表
func (h *SystemHandler) listNamespaces(c *gin.Context) {
	items, err := h.service.ListNamespaces(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"items": []any{}, "count": 0, "warning": "failed to list namespaces", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}

// ListStorageClasses 返回所有可用的存储类
func ListStorageClasses(c *gin.Context) {
	h, ok := NewSystemHandlerFromClient(c)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"items": []any{}, "count": 0, "warning": "k8s client not initialized"})
		return
	}
	h.listStorageClasses(c)
}

// listStorageClasses 实例方法处理存储类列表
func (h *SystemHandler) listStorageClasses(c *gin.Context) {
	items, err := h.service.ListStorageClasses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"items": []any{}, "count": 0, "warning": "failed to list storage classes", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}

// ListPolarDBXVersions 返回支持的 PolarDB-X 版本列表
func ListPolarDBXVersions(c *gin.Context) {
	h, ok := NewSystemHandlerFromClient(c)
	if !ok {
		// 即使没有 k8s client，版本列表也可以返回
		svc := service.NewSystemService(nil)
		versions := svc.GetPolarDBXVersions()
		c.JSON(http.StatusOK, gin.H{"items": versions, "count": len(versions)})
		return
	}
	versions := h.service.GetPolarDBXVersions()
	c.JSON(http.StatusOK, gin.H{"items": versions, "count": len(versions)})
}
