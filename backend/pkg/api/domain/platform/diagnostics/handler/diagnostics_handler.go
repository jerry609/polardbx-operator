package handler

import (
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"

	"polardbx-ui-backend/pkg/api/domain/platform/diagnostics/service"
	apierr "polardbx-ui-backend/pkg/api/errors"
	"polardbx-ui-backend/pkg/api/util"
)

// Start triggers cluster diagnostic task
// @Summary Start diagnosis
// @Description Create polardbx-clinic Pod to collect cluster diagnostic information
// @Tags diagnostics
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param cluster path string true "cluster name"
// @Success 202 {object} service.DiagnosticJob "diagnostic task started"
// @Failure 400 {object} map[string]string "request parameter error"
// @Failure 500 {object} map[string]string "server error"
// @Router /api/v1/diagnostics/{namespace}/{cluster}/start [post]
func Start(c *gin.Context) {
	namespace := c.Param("namespace")
	cluster := c.Param("cluster")

	if namespace == "" || cluster == "" {
		apierr.AbortValidation(c, "namespace and cluster name cannot be empty")
		return
	}

	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		apierr.AbortInternal(c, "unable to get Kubernetes client")
		return
	}

	svc := service.NewDiagnosticsService(cli)
	job, err := svc.StartDiagnosis(c.Request.Context(), namespace, cluster)
	if err != nil {
		apierr.AbortInternal(c, err.Error())
		return
	}

	apierr.Accepted(c, job)
}

// GetStatus returns diagnostic task progress/status
// @Summary Get diagnostic status
// @Description Get current status and progress of specified diagnostic task
// @Tags diagnostics
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param id path string true "diagnostic task ID"
// @Success 200 {object} service.DiagnosticJob "diagnostic task status"
// @Failure 404 {object} map[string]string "task not found"
// @Failure 500 {object} map[string]string "server error"
// @Router /api/v1/diagnostics/{namespace}/{id}/status [get]
func GetStatus(c *gin.Context) {
	namespace := c.Param("namespace")
	id := c.Param("id")

	if namespace == "" || id == "" {
		apierr.AbortValidation(c, "namespace and task ID cannot be empty")
		return
	}

	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		apierr.AbortInternal(c, "unable to get Kubernetes client")
		return
	}

	svc := service.NewDiagnosticsService(cli)
	job, err := svc.GetDiagnosisStatus(c.Request.Context(), namespace, id)
	if err != nil {
		apierr.AbortNotFound(c, "diagnostic job", id)
		return
	}

	apierr.OK(c, job)
}

// ListReports lists diagnostic reports
// @Summary List diagnostic reports
// @Description List all diagnostic reports under specified namespace
// @Tags diagnostics
// @Accept json
// @Produce json
// @Param namespace query string false "namespace, list all namespaces if not specified"
// @Success 200 {array} service.DiagnosticJob "diagnostic reports list"
// @Failure 500 {object} map[string]string "server error"
// @Router /api/v1/diagnostics/reports [get]
func ListReports(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "")

	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		apierr.AbortInternal(c, "unable to get Kubernetes client")
		return
	}

	svc := service.NewDiagnosticsService(cli)
	reports, err := svc.ListDiagnosisReports(c.Request.Context(), namespace)
	if err != nil {
		apierr.AbortInternal(c, err.Error())
		return
	}

	if reports == nil {
		reports = []service.DiagnosticJob{}
	}

	apierr.OK(c, gin.H{
		"namespace": namespace,
		"reports":   reports,
		"total":     len(reports),
	})
}

// Download returns report download information
// @Summary Download diagnostic report
// @Description Get download link or path for diagnostic report
// @Tags diagnostics
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param id path string true "diagnostic task ID"
// @Success 200 {object} map[string]string "download information"
// @Failure 400 {object} map[string]string "request parameter error"
// @Failure 404 {object} map[string]string "report not found"
// @Failure 500 {object} map[string]string "server error"
// @Router /api/v1/diagnostics/{namespace}/{id}/download [get]
func Download(c *gin.Context) {
	namespace := c.Param("namespace")
	id := c.Param("id")

	if namespace == "" || id == "" {
		apierr.AbortValidation(c, "namespace and task ID cannot be empty")
		return
	}

	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		apierr.AbortInternal(c, "unable to get Kubernetes client")
		return
	}

	svc := service.NewDiagnosticsService(cli)
	outputPath, err := svc.GetDownloadInfo(c.Request.Context(), namespace, id)
	if err != nil {
		apierr.AbortNotFound(c, "diagnostic report", id)
		return
	}

	// Return download information
	// Note: In production environment, may need to:
	// 1. Copy files from Pod to accessible storage
	// 2. Generate pre-signed URLs
	// 3. Stream files through API proxy
	apierr.OK(c, gin.H{
		"id":        id,
		"namespace": namespace,
		"path":      outputPath,
		"url":       "/api/v1/diagnostics/" + namespace + "/" + id + "/file",
		"message":   "diagnostic report ready, can be downloaded via kubectl cp command or url",
		"command":   "kubectl cp " + namespace + "/polardbx-clinic-" + id + ":" + outputPath + " ./" + id + ".tar.gz",
	})
}

// GetFile gets diagnostic report file (streaming download)
// @Summary Get diagnostic report file
// @Description Get diagnostic report file content from Pod
// @Tags diagnostics
// @Produce application/gzip
// @Param namespace path string true "namespace"
// @Param id path string true "diagnostic task ID"
// @Success 200 {file} binary "diagnostic report file"
// @Failure 404 {object} map[string]string "file not found"
// @Failure 500 {object} map[string]string "server error"
// @Router /api/v1/diagnostics/{namespace}/{id}/file [get]
func GetFile(c *gin.Context) {
	namespace := c.Param("namespace")
	id := c.Param("id")

	// TODO: Implement reading files from Pod and streaming
	// This requires using kubernetes client-go's exec/cp functionality
	// or kubectl cp's underlying implementation

	apierr.Abort(c, apierr.Internal("file download requires additional configuration, please use kubectl cp command to download: kubectl cp "+namespace+"/polardbx-clinic-"+id+":/tmp/polardbx-clinic/"+id+".tar.gz ./"+id+".tar.gz"))
}

// DeleteJob deletes diagnostic task (cleanup Pod)
// @Summary Delete diagnostic task
// @Description Delete diagnostic Pod and related resources
// @Tags diagnostics
// @Accept json
// @Produce json
// @Param namespace path string true "namespace"
// @Param id path string true "diagnostic task ID"
// @Success 200 {object} map[string]string "deletion successful"
// @Failure 404 {object} map[string]string "task not found"
// @Failure 500 {object} map[string]string "server error"
// @Router /api/v1/diagnostics/{namespace}/{id} [delete]
func DeleteJob(c *gin.Context) {
	namespace := c.Param("namespace")
	id := c.Param("id")

	if namespace == "" || id == "" {
		apierr.AbortValidation(c, "namespace and task ID cannot be empty")
		return
	}

	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		apierr.AbortInternal(c, "unable to get Kubernetes client")
		return
	}

	// Delete diagnostic Pod
	podName := service.ClinicPodPrefix + id
	pod := &corev1.Pod{}
	pod.SetName(podName)
	pod.SetNamespace(namespace)

	if err := cli.Delete(c.Request.Context(), pod); err != nil {
		apierr.AbortK8sError(c, "delete diagnostic pod", err)
		return
	}

	apierr.OK(c, gin.H{
		"message": "diagnostic task deleted",
		"id":      id,
	})
}
