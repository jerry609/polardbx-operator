package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	polardbx "github.com/alibaba/polardbx-operator/api/v1/polardbx"

	"polardbx-ui-backend/pkg/api/domain/platform/restore/repository"
	"polardbx-ui-backend/pkg/api/util"
)

// RestoreHandler 处理恢复相关的 HTTP 请求
type RestoreHandler struct {
	repo repository.RestoreRepository
}

// NewRestoreHandler 创建新的 RestoreHandler
func NewRestoreHandler(repo repository.RestoreRepository) *RestoreHandler {
	return &RestoreHandler{repo: repo}
}

// NewRestoreHandlerFromContext 从 gin.Context 创建 handler
func NewRestoreHandlerFromContext(c *gin.Context) (*RestoreHandler, bool) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return nil, false
	}
	repo := repository.NewK8sRestoreRepository(cli)
	return NewRestoreHandler(repo), true
}

// RestoreCluster POST /clusters/:namespace/:name/restore
func RestoreCluster(c *gin.Context) {
	h, ok := NewRestoreHandlerFromContext(c)
	if !ok {
		return
	}
	h.restoreCluster(c)
}

func (h *RestoreHandler) restoreCluster(c *gin.Context) {
	namespace := c.Param("namespace")
	clusterName := c.Param("name")

	var req struct {
		BackupSet       string `json:"backupSet"`
		BackupName      string `json:"backupName"`
		TargetCluster   string `json:"targetCluster,omitempty"`
		TargetName      string `json:"targetName,omitempty"`
		StorageProvider *struct {
			Type   string            `json:"type"`
			Config map[string]string `json:"config"`
		} `json:"storageProvider,omitempty"`
		Time     string `json:"time,omitempty"`
		TimeZone string `json:"timezone,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid restore request", "details": err.Error()})
		return
	}
	if req.BackupSet == "" && req.BackupName != "" {
		req.BackupSet = req.BackupName
	}
	if req.TargetCluster == "" && req.TargetName != "" {
		req.TargetCluster = req.TargetName
	}
	if strings.TrimSpace(req.BackupSet) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid restore request", "details": "backupSet (or backupName) is required"})
		return
	}

	// 验证备份存在且已完成
	if req.BackupSet != "" {
		backup, err := h.repo.GetBackup(c.Request.Context(), namespace, req.BackupSet)
		if err != nil {
			util.HandleK8sError(c, "backup not found", err)
			return
		}
		if backup.Status.Phase != polardbxv1.BackupFinished {
			c.JSON(http.StatusBadRequest, gin.H{"error": "backup not ready", "details": fmt.Sprintf("phase=%s", backup.Status.Phase)})
			return
		}
	}

	target := req.TargetCluster
	if target == "" {
		target = clusterName + "-restored"
	}

	// 确保目标集群不存在
	if _, err := h.repo.GetCluster(c.Request.Context(), namespace, target); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "target cluster exists", "details": fmt.Sprintf("%s/%s", namespace, target)})
		return
	}

	// 加载源集群
	source, err := h.repo.GetCluster(c.Request.Context(), namespace, clusterName)
	if err != nil {
		util.HandleK8sError(c, "source cluster not found", err)
		return
	}

	// 构建恢复集群
	restored := &polardbxv1.PolarDBXCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      target,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":     "polardbx",
				"app.kubernetes.io/instance": target,
				"polardbx/restore-source":    clusterName,
				"polardbx/restore-backup":    req.BackupSet,
			},
			Annotations: map[string]string{
				"polardbx/restore-from":   fmt.Sprintf("%s/%s", namespace, clusterName),
				"polardbx/restore-backup": req.BackupSet,
				"polardbx/restore-time":   time.Now().Format(time.RFC3339),
			},
		},
		Spec: source.Spec,
	}
	restored.Spec.ServiceName = target
	if restored.Spec.Restore == nil {
		restored.Spec.Restore = &polardbx.RestoreSpec{}
	}
	restored.Spec.Restore.BackupSet = req.BackupSet
	if req.Time != "" {
		restored.Spec.Restore.Time = req.Time
	}
	if req.TimeZone != "" {
		restored.Spec.Restore.TimeZone = req.TimeZone
	}
	restored.Spec.Restore.From = polardbx.PolarDBXRestoreFrom{PolarBDXName: clusterName}

	if err := h.repo.CreateCluster(c.Request.Context(), restored); err != nil {
		util.HandleK8sError(c, "failed to create restored cluster", err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "cluster restore initiated successfully", "sourceCluster": clusterName, "targetCluster": target, "namespace": namespace, "backupSet": req.BackupSet, "restoreTime": req.Time, "status": "creating"})
}

// InitiatePITR POST /clusters/:namespace/:name/pitr
func InitiatePITR(c *gin.Context) {
	h, ok := NewRestoreHandlerFromContext(c)
	if !ok {
		return
	}
	h.initiatePITR(c)
}

func (h *RestoreHandler) initiatePITR(c *gin.Context) {
	namespace := c.Param("namespace")
	sourceName := c.Param("name")

	var req struct {
		Time            string `json:"time"`
		TargetTime      string `json:"targetTime"`
		TimeZone        string `json:"timezone"`
		BackupSet       string `json:"backupSet"`
		BackupName      string `json:"backupName"`
		TargetCluster   string `json:"targetCluster"`
		TargetName      string `json:"targetName"`
		StorageProvider *struct {
			Type   string            `json:"type"`
			Config map[string]string `json:"config"`
		} `json:"storageProvider"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid PITR request", "details": err.Error()})
		return
	}

	// 规范化字段
	if strings.TrimSpace(req.Time) == "" {
		req.Time = req.TargetTime
	}
	if req.BackupSet == "" && req.BackupName != "" {
		req.BackupSet = req.BackupName
	}
	if req.TargetCluster == "" && req.TargetName != "" {
		req.TargetCluster = req.TargetName
	}

	if strings.TrimSpace(req.Time) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid PITR request", "details": "time (or targetTime) is required"})
		return
	}

	// 验证备份
	if strings.TrimSpace(req.BackupSet) != "" {
		backup, err := h.repo.GetBackup(c.Request.Context(), namespace, req.BackupSet)
		if err != nil {
			util.HandleK8sError(c, "backup not found", err)
			return
		}
		if backup.Status.Phase != polardbxv1.BackupFinished {
			c.JSON(http.StatusBadRequest, gin.H{"error": "backup not ready", "details": fmt.Sprintf("phase=%s", backup.Status.Phase)})
			return
		}
	}

	target := req.TargetCluster
	if target == "" {
		target = sourceName + "-pitr"
	}

	// 确保目标集群不存在
	if _, err := h.repo.GetCluster(c.Request.Context(), namespace, target); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "target cluster exists", "details": fmt.Sprintf("%s/%s", namespace, target)})
		return
	}

	// 加载源集群
	source, err := h.repo.GetCluster(c.Request.Context(), namespace, sourceName)
	if err != nil {
		util.HandleK8sError(c, "source cluster not found", err)
		return
	}

	// 构建 PITR 恢复集群
	restored := &polardbxv1.PolarDBXCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      target,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":     "polardbx",
				"app.kubernetes.io/instance": target,
				"polardbx/restore-source":    sourceName,
			},
			Annotations: map[string]string{
				"polardbx/restore-from": fmt.Sprintf("%s/%s", namespace, sourceName),
				"polardbx/restore-time": req.Time,
			},
		},
		Spec: source.Spec,
	}
	restored.Spec.ServiceName = target
	if restored.Spec.Restore == nil {
		restored.Spec.Restore = &polardbx.RestoreSpec{}
	}
	restored.Spec.Restore.Time = req.Time
	if req.TimeZone != "" {
		restored.Spec.Restore.TimeZone = req.TimeZone
	}
	if req.BackupSet != "" {
		restored.Spec.Restore.BackupSet = req.BackupSet
		if restored.Labels == nil {
			restored.Labels = map[string]string{}
		}
		restored.Labels["polardbx/restore-backup"] = req.BackupSet
	}
	restored.Spec.Restore.From = polardbx.PolarDBXRestoreFrom{PolarBDXName: sourceName}

	if err := h.repo.CreateCluster(c.Request.Context(), restored); err != nil {
		util.HandleK8sError(c, "failed to create PITR restored cluster", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "PITR initiated successfully",
		"sourceCluster": sourceName,
		"targetCluster": target,
		"namespace":     namespace,
		"pitrTime":      req.Time,
		"restoreSpec": gin.H{
			"time":      req.Time,
			"timezone":  req.TimeZone,
			"backupSet": req.BackupSet,
		},
	})
}

// GetRestoreStatus GET /clusters/:namespace/:name/restore-status
func GetRestoreStatus(c *gin.Context) {
	h, ok := NewRestoreHandlerFromContext(c)
	if !ok {
		return
	}
	h.getRestoreStatus(c)
}

func (h *RestoreHandler) getRestoreStatus(c *gin.Context) {
	ns := c.Param("namespace")
	name := c.Param("name")

	cluster, err := h.repo.GetCluster(c.Request.Context(), ns, name)
	if err != nil {
		util.HandleK8sError(c, "failed to get cluster", err)
		return
	}

	isRestore := false
	var restoreSpec *polardbx.RestoreSpec
	var sourceCluster, restoreBackup, restoreTime string

	if cluster.Spec.Restore != nil {
		isRestore = true
		restoreSpec = cluster.Spec.Restore
	}
	if cluster.Labels != nil {
		if s, ok := cluster.Labels["polardbx/restore-source"]; ok {
			sourceCluster = s
			isRestore = true
		}
		if s, ok := cluster.Labels["polardbx/restore-backup"]; ok {
			restoreBackup = s
		}
		if s, ok := cluster.Labels["polardbx/restore-time"]; ok {
			restoreTime = s
		}
	}
	if cluster.Annotations != nil {
		if s, ok := cluster.Annotations["polardbx/restore-backup"]; ok && restoreBackup == "" {
			restoreBackup = s
		}
	}
	if restoreSpec != nil {
		if restoreSpec.BackupSet != "" {
			restoreBackup = restoreSpec.BackupSet
		}
		if restoreSpec.Time != "" {
			restoreTime = restoreSpec.Time
		}
		if restoreSpec.From.PolarBDXName != "" {
			sourceCluster = restoreSpec.From.PolarBDXName
		}
	}

	phase := "Unknown"
	stage := "Unknown"
	isRestoring := false
	if cluster.Status.Phase != "" {
		switch cluster.Status.Phase {
		case polardbx.PhasePending:
			phase = "Pending"
			isRestoring = true
		case polardbx.PhaseCreating:
			phase = "Creating"
			stage = "Initializing"
			isRestoring = true
		case polardbx.PhaseRunning:
			if isRestore {
				phase = "Completed"
				stage = "Running"
			} else {
				phase = "Running"
				stage = "Normal"
			}
		case polardbx.PhaseFailed:
			phase = "Failed"
			stage = "Error"
		default:
			phase = string(cluster.Status.Phase)
		}
		if strings.EqualFold(phase, "restoring") {
			stage = "Creating"
			isRestoring = true
		}
	}

	conditions := make([]map[string]any, 0)
	for _, cond := range cluster.Status.Conditions {
		conditions = append(conditions, map[string]any{
			"type":               cond.Type,
			"status":             cond.Status,
			"lastTransitionTime": cond.LastTransitionTime.Format(time.RFC3339),
			"reason":             cond.Reason,
			"message":            cond.Message,
		})
	}

	resp := gin.H{
		"clusterName":        name,
		"namespace":          ns,
		"phase":              phase,
		"stage":              stage,
		"isRestoring":        isRestoring,
		"observedGeneration": cluster.Status.ObservedGeneration,
		"conditions":         conditions,
	}

	if isRestore {
		specInfo := gin.H{}
		if restoreBackup != "" {
			specInfo["backupSet"] = restoreBackup
		}
		if restoreTime != "" {
			specInfo["time"] = restoreTime
		}
		if sourceCluster != "" {
			specInfo["from"] = gin.H{"polardbxName": sourceCluster}
		}
		if restoreSpec != nil && restoreSpec.TimeZone != "" {
			specInfo["timezone"] = restoreSpec.TimeZone
		}
		resp["restoreSpec"] = specInfo
		resp["isRestoreCluster"] = true
		resp["sourceCluster"] = sourceCluster
	} else {
		resp["isRestoreCluster"] = false
	}
	c.JSON(http.StatusOK, resp)
}

// ListJobs GET /restore-jobs
func ListJobs(c *gin.Context) {
	h, ok := NewRestoreHandlerFromContext(c)
	if !ok {
		return
	}
	h.listJobs(c)
}

func (h *RestoreHandler) listJobs(c *gin.Context) {
	ns := c.DefaultQuery("namespace", "")

	clusters, err := h.repo.ListClusters(c.Request.Context(), ns)
	if err != nil {
		util.HandleK8sError(c, "failed to list clusters", err)
		return
	}

	jobs := make([]gin.H, 0)
	for _, cl := range clusters {
		isRestore := cl.Spec.Restore != nil || (cl.Labels != nil && (cl.Labels["polardbx/restore-source"] != ""))
		if !isRestore {
			continue
		}
		job := gin.H{"name": cl.Name, "namespace": cl.Namespace, "phase": string(cl.Status.Phase)}
		jobs = append(jobs, job)
	}
	c.JSON(http.StatusOK, gin.H{"items": jobs, "count": len(jobs)})
}

// GetJob GET /restore-jobs/:namespace/:name
func GetJob(c *gin.Context) {
	GetRestoreStatus(c)
}

// CancelJob DELETE /restore-jobs/:namespace/:name
func CancelJob(c *gin.Context) {
	h, ok := NewRestoreHandlerFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")

	// 检查 job 是否存在 (使用 PolarDBXBackup 作为代理检查)
	_, err := h.repo.GetBackup(c.Request.Context(), ns, name)
	if err != nil {
		// Job 不存在
		c.JSON(http.StatusNotFound, gin.H{"error": "restore job not found", "namespace": ns, "name": name})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cancel requested", "namespace": ns, "name": name, "status": "pending_implementation"})
}
