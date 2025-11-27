// Package logcollector 提供日志收集器相关的 HTTP 处理器
// 这是一个薄包装层，委托到 domain/platform/logcollector/handler
package logcollector

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"polardbx-ui-backend/pkg/api/domain/platform/logcollector/handler"
)

// List GET /logcollectors
// 列出日志收集器
var List = handler.List

// Create POST /logcollectors
// 创建日志收集器
var Create = handler.Create

// Get GET /logcollectors/:namespace/:name
// 获取日志收集器
var Get = handler.Get

// Update PUT /logcollectors/:namespace/:name
// 更新日志收集器
var Update = handler.Update

// Delete DELETE /logcollectors/:namespace/:name
// 删除日志收集器
var Delete = handler.Delete

// k8sClientFromContext 从 context 获取 k8s client
// 保留此函数供 pipeline.go 使用
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

// 保留原有函数签名用于兼容性
func list(c *gin.Context)    { handler.List(c) }
func create(c *gin.Context)  { handler.Create(c) }
func get(c *gin.Context)     { handler.Get(c) }
func update(c *gin.Context)  { handler.Update(c) }
func delete_(c *gin.Context) { handler.Delete(c) }
