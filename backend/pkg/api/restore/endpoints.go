// Package restore 提供恢复相关的 HTTP 处理器
// 这是一个薄包装层，委托到 domain/platform/restore/handler
package restore

import (
	"github.com/gin-gonic/gin"

	"polardbx-ui-backend/pkg/api/domain/platform/restore/handler"
)

// RestoreCluster POST /clusters/:namespace/:name/restore
// 从备份恢复集群
var RestoreCluster = handler.RestoreCluster

// InitiatePITR POST /clusters/:namespace/:name/pitr
// 发起 PITR (Point-in-Time Recovery)
var InitiatePITR = handler.InitiatePITR

// GetRestoreStatus GET /clusters/:namespace/:name/restore-status
// 获取集群恢复状态
var GetRestoreStatus = handler.GetRestoreStatus

// ListJobs GET /restore-jobs
// 列出恢复任务
var ListJobs = handler.ListJobs

// GetJob GET /restore-jobs/:namespace/:name
// 获取恢复任务详情
var GetJob = handler.GetJob

// CancelJob DELETE /restore-jobs/:namespace/:name
// 取消恢复任务
var CancelJob = handler.CancelJob

// 如果需要保留原有函数签名用于兼容性，可以使用包装函数
func restoreCluster(c *gin.Context)   { handler.RestoreCluster(c) }
func initiatePITR(c *gin.Context)     { handler.InitiatePITR(c) }
func getRestoreStatus(c *gin.Context) { handler.GetRestoreStatus(c) }
func listJobs(c *gin.Context)         { handler.ListJobs(c) }
func getJob(c *gin.Context)           { handler.GetJob(c) }
func cancelJob(c *gin.Context)        { handler.CancelJob(c) }
