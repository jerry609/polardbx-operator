package handler


import (
	"net/http"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"github.com/gin-gonic/gin"

	"polardbx-ui-backend/pkg/api/domain/systemtasks/service"
	"polardbx-ui-backend/pkg/api/util"
)

// SystemTaskHandler 处理 SystemTask 相关的 HTTP 请求
// 只负责 HTTP 请求/响应处理，业务逻辑委托给 Service
type SystemTaskHandler struct {
	service *service.SystemTaskService
}

// NewSystemTaskHandler 创建 Handler 实例
func NewSystemTaskHandler(svc *service.SystemTaskService) *SystemTaskHandler {
	return &SystemTaskHandler{service: svc}
}

// List 列出 SystemTask
func (h *SystemTaskHandler) List(c *gin.Context) {
	// 1. 获取 K8s 客户端
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}

	// 2. 解析请求参数
	namespace := util.DefaultNamespace(c, "default")

	// 3. 调用 Service（传递 context.Context，不是 gin.Context）
	tasks, err := h.service.List(c.Request.Context(), cli, namespace)
	if err != nil {
		util.HandleK8sError(c, "failed to list system tasks", err)
		return
	}

	// 4. 返回响应
	c.JSON(http.StatusOK, tasks)
}

// Get 获取单个 SystemTask
func (h *SystemTaskHandler) Get(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	task, err := h.service.Get(c.Request.Context(), cli, namespace, name)
	if err != nil {
		util.HandleK8sError(c, "failed to get system task", err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// Create 创建 SystemTask
func (h *SystemTaskHandler) Create(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}

	namespace := util.DefaultNamespace(c, "default")

	var body polardbxv1.SystemTask
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid system task", "details": err.Error()})
		return
	}

	created, err := h.service.Create(c.Request.Context(), cli, namespace, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to create system task", err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

// Update 更新 SystemTask
func (h *SystemTaskHandler) Update(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}

	namespace := c.Param("namespace")

	var body polardbxv1.SystemTask
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid system task", "details": err.Error()})
		return
	}

	updated, err := h.service.Update(c.Request.Context(), cli, namespace, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to update system task", err)
		return
	}

	c.JSON(http.StatusOK, updated)
}

// Delete 删除 SystemTask
func (h *SystemTaskHandler) Delete(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	if err := h.service.Delete(c.Request.Context(), cli, namespace, name); err != nil {
		util.HandleK8sError(c, "failed to delete system task", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "system task deleted"})
}

// RegisterRoutes 注册路由
func (h *SystemTaskHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/system-tasks", h.List)
	rg.POST("/system-tasks", h.Create)
	rg.GET("/system-tasks/:namespace/:name", h.Get)
	rg.PUT("/system-tasks/:namespace/:name", h.Update)
	rg.DELETE("/system-tasks/:namespace/:name", h.Delete)
}
