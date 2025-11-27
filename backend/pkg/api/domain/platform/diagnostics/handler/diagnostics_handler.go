package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Start 触发集群诊断任务（占位实现）
func Start(c *gin.Context) {
	namespace := c.Param("namespace")
	cluster := c.Param("cluster")
	c.JSON(http.StatusAccepted, gin.H{
		"id":        time.Now().UnixNano(),
		"namespace": namespace,
		"cluster":   cluster,
		"status":    "running",
		"startedAt": time.Now().Format(time.RFC3339),
		"message":   "pending_implementation",
	})
}

// GetStatus 返回诊断任务的进度/状态（占位）
func GetStatus(c *gin.Context) {
	namespace := c.Param("namespace")
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"id":        id,
		"namespace": namespace,
		"status":    "pending_implementation",
		"progress":  0,
	})
}

// ListReports 列出命名空间的诊断报告（占位）
func ListReports(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "")
	c.JSON(http.StatusOK, gin.H{
		"namespace": namespace,
		"reports":   []any{},
	})
}

// Download 返回报告下载链接（占位）
func Download(c *gin.Context) {
	namespace := c.Param("namespace")
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"id":        id,
		"namespace": namespace,
		"download":  "pending_implementation",
	})
}
