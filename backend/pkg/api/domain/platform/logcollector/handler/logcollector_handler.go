package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/client"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"

	"polardbx-ui-backend/pkg/api/domain/platform/logcollector/repository"
	"polardbx-ui-backend/pkg/api/util"
)

// LogCollectorHandler 处理日志收集器相关的 HTTP 请求
type LogCollectorHandler struct {
	repo repository.LogCollectorRepository
}

// NewLogCollectorHandler 创建新的 LogCollectorHandler
func NewLogCollectorHandler(repo repository.LogCollectorRepository) *LogCollectorHandler {
	return &LogCollectorHandler{repo: repo}
}

// NewLogCollectorHandlerFromContext 从 gin.Context 创建 handler
func NewLogCollectorHandlerFromContext(c *gin.Context) (*LogCollectorHandler, bool) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return nil, false
	}
	repo := repository.NewK8sLogCollectorRepository(cli)
	return NewLogCollectorHandler(repo), true
}

// k8sClientFromContext 从 context 获取 k8s client (兼容旧代码)
func k8sClientFromContext(c *gin.Context) (client.Client, bool) {
	v, ok := c.Get("k8sClient")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "kubernetes client not initialized"})
		return nil, false
	}
	cli, ok := v.(client.Client)
	if !ok || cli == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid kubernetes client in context"})
		return nil, false
	}
	return cli, true
}

// List GET /logcollectors
// 列出日志收集器
func List(c *gin.Context) {
	h, ok := NewLogCollectorHandlerFromContext(c)
	if !ok {
		return
	}
	h.list(c)
}

func (h *LogCollectorHandler) list(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")
	collectors, err := h.repo.List(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list log collectors", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, collectors)
}

// Create POST /logcollectors
// 创建日志收集器
func Create(c *gin.Context) {
	h, ok := NewLogCollectorHandlerFromContext(c)
	if !ok {
		return
	}
	h.create(c)
}

func (h *LogCollectorHandler) create(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")
	var collector polardbxv1.PolarDBXLogCollector
	if err := c.ShouldBindJSON(&collector); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse log collector data", "details": err.Error()})
		return
	}
	collector.Namespace = namespace
	createdCollector, err := h.repo.Create(c.Request.Context(), &collector)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create log collector", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, createdCollector)
}

// Get GET /logcollectors/:namespace/:name
// 获取日志收集器
func Get(c *gin.Context) {
	h, ok := NewLogCollectorHandlerFromContext(c)
	if !ok {
		return
	}
	h.get(c)
}

func (h *LogCollectorHandler) get(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	collector, err := h.repo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "failed to get log collector", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, collector)
}

// Update PUT /logcollectors/:namespace/:name
// 更新日志收集器
func Update(c *gin.Context) {
	h, ok := NewLogCollectorHandlerFromContext(c)
	if !ok {
		return
	}
	h.update(c)
}

func (h *LogCollectorHandler) update(c *gin.Context) {
	namespace := c.Param("namespace")
	var collector polardbxv1.PolarDBXLogCollector
	if err := c.ShouldBindJSON(&collector); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse log collector data", "details": err.Error()})
		return
	}
	collector.Namespace = namespace
	updatedCollector, err := h.repo.Update(c.Request.Context(), &collector)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update log collector", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedCollector)
}

// Delete DELETE /logcollectors/:namespace/:name
// 删除日志收集器
func Delete(c *gin.Context) {
	h, ok := NewLogCollectorHandlerFromContext(c)
	if !ok {
		return
	}
	h.delete(c)
}

func (h *LogCollectorHandler) delete(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	if err := h.repo.Delete(c.Request.Context(), namespace, name); err != nil {
		util.HandleK8sError(c, "failed to delete log collector", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "log collector deleted successfully"})
}
