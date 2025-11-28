package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"

	"polardbx-ui-backend/pkg/api/domain/platform/diagnostics/service"
	"polardbx-ui-backend/pkg/api/util"
)

// Start 触发集群诊断任务
// @Summary 启动诊断
// @Description 创建 polardbx-clinic Pod 收集集群诊断信息
// @Tags diagnostics
// @Accept json
// @Produce json
// @Param namespace path string true "命名空间"
// @Param cluster path string true "集群名称"
// @Success 202 {object} service.DiagnosticJob "诊断任务已启动"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/diagnostics/{namespace}/{cluster}/start [post]
func Start(c *gin.Context) {
	namespace := c.Param("namespace")
	cluster := c.Param("cluster")

	if namespace == "" || cluster == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_parameters",
			"message": "命名空间和集群名称不能为空",
		})
		return
	}

	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "k8s_client_error",
			"message": "无法获取 Kubernetes 客户端",
		})
		return
	}

	svc := service.NewDiagnosticsService(cli)
	job, err := svc.StartDiagnosis(c.Request.Context(), namespace, cluster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "start_diagnosis_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, job)
}

// GetStatus 返回诊断任务的进度/状态
// @Summary 获取诊断状态
// @Description 获取指定诊断任务的当前状态和进度
// @Tags diagnostics
// @Accept json
// @Produce json
// @Param namespace path string true "命名空间"
// @Param id path string true "诊断任务ID"
// @Success 200 {object} service.DiagnosticJob "诊断任务状态"
// @Failure 404 {object} map[string]string "任务不存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/diagnostics/{namespace}/{id}/status [get]
func GetStatus(c *gin.Context) {
	namespace := c.Param("namespace")
	id := c.Param("id")

	if namespace == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_parameters",
			"message": "命名空间和任务ID不能为空",
		})
		return
	}

	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "k8s_client_error",
			"message": "无法获取 Kubernetes 客户端",
		})
		return
	}

	svc := service.NewDiagnosticsService(cli)
	job, err := svc.GetDiagnosisStatus(c.Request.Context(), namespace, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "job_not_found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListReports 列出诊断报告
// @Summary 列出诊断报告
// @Description 列出指定命名空间下的所有诊断报告
// @Tags diagnostics
// @Accept json
// @Produce json
// @Param namespace query string false "命名空间，不指定则列出所有命名空间"
// @Success 200 {array} service.DiagnosticJob "诊断报告列表"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/diagnostics/reports [get]
func ListReports(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "")

	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "k8s_client_error",
			"message": "无法获取 Kubernetes 客户端",
		})
		return
	}

	svc := service.NewDiagnosticsService(cli)
	reports, err := svc.ListDiagnosisReports(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "list_reports_failed",
			"message": err.Error(),
		})
		return
	}

	if reports == nil {
		reports = []service.DiagnosticJob{}
	}

	c.JSON(http.StatusOK, gin.H{
		"namespace": namespace,
		"reports":   reports,
		"total":     len(reports),
	})
}

// Download 返回报告下载信息
// @Summary 下载诊断报告
// @Description 获取诊断报告的下载链接或路径
// @Tags diagnostics
// @Accept json
// @Produce json
// @Param namespace path string true "命名空间"
// @Param id path string true "诊断任务ID"
// @Success 200 {object} map[string]string "下载信息"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 404 {object} map[string]string "报告不存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/diagnostics/{namespace}/{id}/download [get]
func Download(c *gin.Context) {
	namespace := c.Param("namespace")
	id := c.Param("id")

	if namespace == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_parameters",
			"message": "命名空间和任务ID不能为空",
		})
		return
	}

	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "k8s_client_error",
			"message": "无法获取 Kubernetes 客户端",
		})
		return
	}

	svc := service.NewDiagnosticsService(cli)
	outputPath, err := svc.GetDownloadInfo(c.Request.Context(), namespace, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "download_not_available",
			"message": err.Error(),
		})
		return
	}

	// 返回下载信息
	// 注意: 在实际生产环境中，可能需要:
	// 1. 从 Pod 中复制文件到可访问的存储
	// 2. 生成预签名 URL
	// 3. 通过 API 代理流式传输文件
	c.JSON(http.StatusOK, gin.H{
		"id":        id,
		"namespace": namespace,
		"path":      outputPath,
		"url":       "/api/v1/diagnostics/" + namespace + "/" + id + "/file",
		"message":   "诊断报告准备就绪，可通过 kubectl cp 命令或 url 下载",
		"command":   "kubectl cp " + namespace + "/polardbx-clinic-" + id + ":" + outputPath + " ./" + id + ".tar.gz",
	})
}

// GetFile 获取诊断报告文件（流式下载）
// @Summary 获取诊断报告文件
// @Description 从 Pod 中获取诊断报告文件内容
// @Tags diagnostics
// @Produce application/gzip
// @Param namespace path string true "命名空间"
// @Param id path string true "诊断任务ID"
// @Success 200 {file} binary "诊断报告文件"
// @Failure 404 {object} map[string]string "文件不存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/diagnostics/{namespace}/{id}/file [get]
func GetFile(c *gin.Context) {
	namespace := c.Param("namespace")
	id := c.Param("id")

	// TODO: 实现从 Pod 中读取文件并流式传输
	// 这需要使用 kubernetes client-go 的 exec/cp 功能
	// 或者使用 kubectl cp 的底层实现

	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "文件下载功能需要额外配置，请使用 kubectl cp 命令下载",
		"command": "kubectl cp " + namespace + "/polardbx-clinic-" + id + ":/tmp/polardbx-clinic/" + id + ".tar.gz ./" + id + ".tar.gz",
	})
}

// DeleteJob 删除诊断任务（清理 Pod）
// @Summary 删除诊断任务
// @Description 删除诊断 Pod 和相关资源
// @Tags diagnostics
// @Accept json
// @Produce json
// @Param namespace path string true "命名空间"
// @Param id path string true "诊断任务ID"
// @Success 200 {object} map[string]string "删除成功"
// @Failure 404 {object} map[string]string "任务不存在"
// @Failure 500 {object} map[string]string "服务器错误"
// @Router /api/v1/diagnostics/{namespace}/{id} [delete]
func DeleteJob(c *gin.Context) {
	namespace := c.Param("namespace")
	id := c.Param("id")

	if namespace == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_parameters",
			"message": "命名空间和任务ID不能为空",
		})
		return
	}

	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "k8s_client_error",
			"message": "无法获取 Kubernetes 客户端",
		})
		return
	}

	// 删除诊断 Pod
	podName := service.ClinicPodPrefix + id
	pod := &corev1.Pod{}
	pod.SetName(podName)
	pod.SetNamespace(namespace)

	if err := cli.Delete(c.Request.Context(), pod); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "delete_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "诊断任务已删除",
		"id":      id,
	})
}
