package prechange

import (
	"net/http"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func computePrechangeChecklist(c *gin.Context, k8sClient client.Client, namespace, name string, windowHours int, now time.Time) (hasRecent bool, lastBackupTime *time.Time, rpoLagSeconds int, storageConnectivity string, rpoOk bool, err error) {
	var backupList polardbxv1.PolarDBXBackupList
	if err = k8sClient.List(c.Request.Context(), &backupList, client.InNamespace(namespace)); err != nil {
		return
	}
	windowStart := now.Add(-time.Duration(windowHours) * time.Hour)
	var maxLrt time.Time
	for _, b := range backupList.Items {
		if b.Spec.Cluster.Name != name {
			continue
		}
		if b.Status.StartTime != nil {
			st := b.Status.StartTime.Time
			if lastBackupTime == nil || st.After(*lastBackupTime) {
				lb := st
				lastBackupTime = &lb
			}
			if st.After(windowStart) && b.Status.Phase == polardbxv1.BackupFinished {
				hasRecent = true
			}
		}
		if b.Status.LatestRecoverableTimestamp != nil {
			lrt := b.Status.LatestRecoverableTimestamp.Time
			if lrt.After(maxLrt) {
				maxLrt = lrt
			}
		}
	}
	rpoLagSeconds = -1
	if !maxLrt.IsZero() {
		rpoLagSeconds = int(now.Sub(maxLrt).Seconds())
	}
	storageConnectivity = "unknown"
	var hpfsCM corev1.ConfigMap
	if e := k8sClient.Get(c.Request.Context(), client.ObjectKey{Namespace: "polardbx-operator-system", Name: "polardbx-hpfs-config"}, &hpfsCM); e == nil {
		storageConnectivity = "configured"
	}
	rpoThreshold := 3600
	var settingsCM corev1.ConfigMap
	if e := k8sClient.Get(c.Request.Context(), client.ObjectKey{Namespace: "polardbx-operator-system", Name: "polardbx-ui-backend-config"}, &settingsCM); e == nil {
		if settingsCM.Data != nil {
			if v := settingsCM.Data["rpoThresholdSeconds"]; v != "" {
				if n, err := strconv.Atoi(v); err == nil {
					rpoThreshold = n
				}
			}
		}
	}
	rpoOk = rpoLagSeconds >= 0 && rpoLagSeconds <= rpoThreshold
	return
}

func GetPrechangeChecklist(c *gin.Context) {
	k8sClient, ok := func() (client.Client, bool) {
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
	}()
	if !ok {
		return
	}
	namespace := c.Param("namespace")
	name := c.Param("name")
	windowHours := 24
	if v := c.DefaultQuery("windowHours", "24"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			windowHours = n
		}
	}
	nowStr := c.DefaultQuery("now", "")
	var now time.Time
	if nowStr != "" {
		if t, err := time.Parse(time.RFC3339, nowStr); err == nil {
			now = t
		}
	}
	if now.IsZero() {
		now = time.Now()
	}

	hasRecent, lastBackupTime, rpoLagSeconds, storageConnectivity, rpoOk, err := computePrechangeChecklist(c, k8sClient, namespace, name, windowHours, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list backups", "details": err.Error()})
		return
	}

	plan := []map[string]any{
		{"id": "checkStorage", "state": ternary(storageConnectivity == "configured", "ok", "warn"), "message": "HPFS 配置存在"},
		{"id": "checkRecentBackup", "state": ternary(hasRecent, "ok", "warn"), "message": "最近窗口内存在成功的全量备份"},
		{"id": "checkRPO", "state": ternary(rpoOk, "ok", "warn"), "message": "RPO 未超过阈值"},
	}
	resp := gin.H{"namespace": namespace, "name": name, "generatedAt": now.Format(time.RFC3339), "checks": gin.H{
		"hasRecentBackup":     hasRecent,
		"lastBackupTime":      ternary(lastBackupTime != nil, lastBackupTime.UTC().Format(time.RFC3339), ""),
		"rpoLagSeconds":       rpoLagSeconds,
		"storageConnectivity": storageConnectivity,
	}, "plan": plan}
	c.JSON(http.StatusOK, resp)
}

func ternary[T any](cond bool, a T, b T) T {
	if cond {
		return a
	}
	return b
}

// Checklist 汇总预变更检查的关键结果
type Checklist struct {
	HasRecentBackup     bool   `json:"hasRecentBackup"`
	RpoLagSeconds       int    `json:"rpoLagSeconds"`
	StorageConnectivity string `json:"storageConnectivity"`
	RpoOk               bool   `json:"rpoOk"`
}

// IsEnforced 读取设置开关，是否开启预变更校验强制门禁
func IsEnforced(c *gin.Context, cli client.Client) bool {
	cm := corev1.ConfigMap{}
	_ = cli.Get(c.Request.Context(), client.ObjectKey{Namespace: "polardbx-operator-system", Name: "polardbx-ui-backend-config"}, &cm)
	if cm.Data == nil {
		return false
	}
	return cm.Data["prechange.enforce"] == "true"
}

// ComputeChecklist 对外导出版本，便于其他包复用
func ComputeChecklist(c *gin.Context, cli client.Client, namespace, name string, windowHours int, now time.Time) (Checklist, error) {
	hasRecent, _, rpoLagSeconds, storageConnectivity, rpoOk, err := computePrechangeChecklist(c, cli, namespace, name, windowHours, now)
	if err != nil {
		return Checklist{}, err
	}
	return Checklist{HasRecentBackup: hasRecent, RpoLagSeconds: rpoLagSeconds, StorageConnectivity: storageConnectivity, RpoOk: rpoOk}, nil
}

// CreatePrecheckSystemTask: 复用预变更检查结果作为系统任务的预检查响应占位
func CreatePrecheckSystemTask(c *gin.Context) {
	// 直接复用同一路由逻辑，保持与根实现一致的行为占位
	GetPrechangeChecklist(c)
}
