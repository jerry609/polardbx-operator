package backup

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"polardbx-ui-backend/pkg/api/util"
	"polardbx-ui-backend/pkg/k8s"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// List backups for a cluster
func List(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	clusterName := c.Param("name")
	namespace := c.Param("namespace")
	backups, err := k8s.ListPolarDBXBackupsWithContext(c.Request.Context(), cli, namespace, clusterName)
	if err != nil {
		util.HandleK8sError(c, "failed to list backups", err)
		return
	}
	c.JSON(http.StatusOK, backups)
}

// Create backup for a cluster
func Create(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	var backup polardbxv1.PolarDBXBackup
	if err := c.ShouldBindJSON(&backup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse backup data", "details": err.Error()})
		return
	}
	clusterName := c.Param("name")
	namespace := c.Param("namespace")
	backup.Spec.Cluster.Name = clusterName
	created, err := k8s.CreatePolarDBXBackupWithContext(c.Request.Context(), cli, namespace, &backup)
	if err != nil {
		util.HandleK8sError(c, "failed to create backup", err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

// Validate backup via dry-run
func Validate(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	var backup polardbxv1.PolarDBXBackup
	if err := c.ShouldBindJSON(&backup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse backup data", "details": err.Error()})
		return
	}
	ns := c.Query("namespace")
	if ns == "" {
		ns = backup.Namespace
	}
	if ns == "" {
		ns = "default"
	}
	if _, err := k8s.CreatePolarDBXBackupDryRunWithContext(c.Request.Context(), cli, ns, &backup); err != nil {
		util.HandleK8sError(c, "backup validation failed", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"valid": true})
}

// Stream backup events (SSE, polling-based)
func StreamEvents(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	namespace := c.Param("namespace")
	name := c.Param("name")
	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	write := func(event string, payload gin.H) {
		payload["timestamp"] = time.Now().UTC().Format(time.RFC3339)
		b, _ := json.Marshal(gin.H{"type": event, "payload": payload})
		_, _ = w.Write([]byte("event: " + event + "\n"))
		_, _ = w.Write([]byte("data: " + string(b) + "\n\n"))
		flusher.Flush()
	}
	ctx := c.Request.Context()
	var lastPhase string
	var backup polardbxv1.PolarDBXBackup
	if err := cli.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &backup); err == nil {
		lastPhase = strings.ToLower(string(backup.Status.Phase))
		write("phaseChanged", gin.H{"phase": lastPhase})
	} else {
		write("error", gin.H{"message": "backup not found", "details": err.Error()})
		return
	}
	keep := time.NewTicker(10 * time.Second)
	defer keep.Stop()
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-keep.C:
			_, _ = w.Write([]byte(": ping\n\n"))
			flusher.Flush()
		case <-tick.C:
			var cur polardbxv1.PolarDBXBackup
			if err := cli.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &cur); err != nil {
				write("error", gin.H{"message": "failed to get backup", "details": err.Error()})
				return
			}
			phase := strings.ToLower(string(cur.Status.Phase))
			if phase != lastPhase {
				lastPhase = phase
				write("phaseChanged", gin.H{"phase": phase})
			}
			switch phase {
			case "succeeded", "completed", "finished", "failed", "deleting":
				return
			}
		}
	}
}

// Get backup metrics (coarse progress)
func GetMetrics(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	var backup polardbxv1.PolarDBXBackup
	if err := cli.Get(c.Request.Context(), client.ObjectKey{Namespace: ns, Name: name}, &backup); err != nil {
		util.HandleK8sError(c, "failed to get backup", err)
		return
	}
	var xsList polardbxv1.XStoreBackupList
	_ = cli.List(c.Request.Context(), &xsList, client.InNamespace(ns))
	total, finished, failed := 0, 0, 0
	for i := range xsList.Items {
		b := xsList.Items[i]
		if b.Labels["polardbx/top-backup"] != name {
			continue
		}
		total++
		p := strings.ToLower(string(b.Status.Phase))
		if p == "finished" || p == "completed" {
			finished++
		} else if p == "failed" {
			failed++
		}
	}
	phase := strings.ToLower(string(backup.Status.Phase))
	mapping := map[string]int{"": 5, "new": 5, "fullbackuping": 25, "collecting": 50, "calculating": 65, "binlogbackuping": 85, "metadatabackuping": 95, "finished": 100, "succeeded": 100, "completed": 100, "failed": 0, "deleting": 0}
	progress := mapping[phase]
	if progress > 0 && progress < 100 && total > 0 {
		fromChildren := int(float64(finished) / float64(total) * 100.0)
		progress = (progress + fromChildren) / 2
		if failed > 0 && phase != "finished" && progress > 90 {
			progress = 90
		}
	}
	c.JSON(http.StatusOK, gin.H{"phase": phase, "progress": progress, "estimated": true, "children": gin.H{"total": total, "finished": finished, "failed": failed}})
}

// Delete backup
func Delete(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	if err := k8s.DeletePolarDBXBackupWithContext(c.Request.Context(), cli, ns, name); err != nil {
		util.HandleK8sError(c, "failed to delete backup", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "backup deleted successfully"})
}
