package backup

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"polardbx-ui-backend/pkg/api/util"

	"bytes"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	hpfsconfig "github.com/alibaba/polardbx-operator/pkg/hpfs/config"
	"github.com/gin-gonic/gin"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetBackupOverview moved from root api to backup package
func GetBackupOverview(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	namespace := c.DefaultQuery("namespace", "")
	now := time.Now()
	windowStart := now.Add(-24 * time.Hour)

	var backupList polardbxv1.PolarDBXBackupList
	listOpts := []client.ListOption{}
	if namespace != "" {
		listOpts = append(listOpts, client.InNamespace(namespace))
	}
	if err := cli.List(c.Request.Context(), &backupList, listOpts...); err != nil {
		util.HandleK8sError(c, "failed to list backups", err)
		return
	}

	total24h, success24h, failed24h, runningNow := 0, 0, 0, 0
	prefixes := make([]string, 0, len(backupList.Items))
	for _, b := range backupList.Items {
		switch b.Status.Phase {
		case polardbxv1.FullBackuping, polardbxv1.BackupCollecting, polardbxv1.BackupCalculating, polardbxv1.BinlogBackuping, polardbxv1.MetadataBackuping:
			runningNow++
		}
		if b.Status.StartTime != nil && b.Status.StartTime.Time.After(windowStart) {
			total24h++
			switch b.Status.Phase {
			case polardbxv1.BackupFinished:
				success24h++
			case polardbxv1.BackupFailed:
				failed24h++
			}
			if b.Status.BackupRootPath != "" {
				prefixes = append(prefixes, b.Status.BackupRootPath)
			}
		}
	}
	successRate := 0
	if total24h > 0 {
		successRate = int(float64(success24h)*100.0/float64(total24h) + 0.5)
	}

	storageConnectivity := "pending_implementation"
	storageConnectivityStatus := "unknown"
	if c.DefaultQuery("evaluateConnectivity", "false") == "true" {
		systemNS := c.DefaultQuery("systemNamespace", "polardbx-operator-system")
		if cm, err := getHpfsConfigMap(c, cli, systemNS); err == nil && cm != nil {
			data := cm.Data["config.yaml"]
			if data == "" {
				storageConnectivity = "connected"
				storageConnectivityStatus = "ok"
			} else {
				if cfg, err2 := decodeHpfsConfig(cm); err2 == nil && len(cfg.Sinks) > 0 {
					mode := strings.ToLower(c.DefaultQuery("connectivityMode", "present"))
					if mode == "online" {
						if checkS3BucketOnline(c, cfg) {
							storageConnectivity = "connected"
							storageConnectivityStatus = "ok"
						} else {
							storageConnectivity = "unreachable"
							storageConnectivityStatus = "warn"
						}
					} else {
						storageConnectivity = "configured"
						storageConnectivityStatus = "ok"
					}
				} else {
					storageConnectivity = "unknown"
					storageConnectivityStatus = "warn"
				}
			}
		} else {
			storageConnectivity = "unknown"
			storageConnectivityStatus = "warn"
		}
	}

	var totalStorageBytes *int64
	if c.DefaultQuery("evaluateStorage", "false") == "true" {
		if v, ok := estimateTotalStorageBytes(c, cli, prefixes); ok {
			totalStorageBytes = &v
		}
	}

	kpi := gin.H{"successRate24h": successRate, "running": runningNow, "failed24h": failed24h, "totalBackups24h": total24h, "totalStorage": "pending_implementation", "storageConnectivity": storageConnectivity}
	if totalStorageBytes != nil {
		kpi["totalStorageBytes"] = *totalStorageBytes
	}
	if c.DefaultQuery("evaluateConnectivity", "false") == "true" {
		kpi["storageConnectivityStatus"] = storageConnectivityStatus
	}
	c.JSON(http.StatusOK, gin.H{"namespace": namespace, "timeWindowHours": 24, "generatedAt": now.Format(time.RFC3339), "kpi": kpi})
}

// GetClusterBackupState aggregates per-cluster backup state: latest full backup, next schedule time, and RPO
func GetClusterBackupState(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	namespace := c.DefaultQuery("namespace", "")
	now := time.Now()

	var clusterList polardbxv1.PolarDBXClusterList
	clusterOpts := []client.ListOption{}
	if namespace != "" {
		clusterOpts = append(clusterOpts, client.InNamespace(namespace))
	}
	if err := cli.List(c.Request.Context(), &clusterList, clusterOpts...); err != nil {
		util.HandleK8sError(c, "failed to list clusters", err)
		return
	}

	var backupList polardbxv1.PolarDBXBackupList
	backupOpts := []client.ListOption{}
	if namespace != "" {
		backupOpts = append(backupOpts, client.InNamespace(namespace))
	}
	if err := cli.List(c.Request.Context(), &backupList, backupOpts...); err != nil {
		util.HandleK8sError(c, "failed to list backups", err)
		return
	}

	var scheduleList polardbxv1.PolarDBXBackupScheduleList
	scheduleOpts := []client.ListOption{}
	if namespace != "" {
		scheduleOpts = append(scheduleOpts, client.InNamespace(namespace))
	}
	if err := cli.List(c.Request.Context(), &scheduleList, scheduleOpts...); err != nil {
		util.HandleK8sError(c, "failed to list backup schedules", err)
		return
	}

	latestByCluster := map[string]polardbxv1.PolarDBXBackup{}
	maxLrtByCluster := map[string]time.Time{}
	for _, b := range backupList.Items {
		clusterName := b.Spec.Cluster.Name
		if clusterName == "" {
			continue
		}
		if prev, ok := latestByCluster[clusterName]; ok {
			if isBackupNewer(b, prev) {
				latestByCluster[clusterName] = b
			}
		} else {
			latestByCluster[clusterName] = b
		}
		if b.Status.LatestRecoverableTimestamp != nil {
			lrt := b.Status.LatestRecoverableTimestamp.Time
			if prev, ok := maxLrtByCluster[clusterName]; !ok || lrt.After(prev) {
				maxLrtByCluster[clusterName] = lrt
			}
		}
	}

	nextByCluster := map[string]*metav1.Time{}
	for _, s := range scheduleList.Items {
		name := s.Spec.BackupSpec.Cluster.Name
		if name == "" {
			continue
		}
		if _, exists := nextByCluster[name]; exists {
			continue
		}
		if s.Status.NextBackupTime != nil {
			nextByCluster[name] = s.Status.NextBackupTime
		}
	}

	out := make([]map[string]any, 0, len(clusterList.Items))
	for _, cl := range clusterList.Items {
		entry := map[string]any{
			"clusterName": cl.Name,
			"namespace":   cl.Namespace,
		}
		if lb, ok := latestByCluster[cl.Name]; ok {
			lbInfo := map[string]any{"name": lb.Name, "phase": string(lb.Status.Phase)}
			if lb.Status.StartTime != nil {
				lbInfo["startTime"] = lb.Status.StartTime.Time.Format(time.RFC3339)
			}
			if lb.Status.EndTime != nil {
				lbInfo["endTime"] = lb.Status.EndTime.Time.Format(time.RFC3339)
			}
			if lb.Status.LatestRecoverableTimestamp != nil {
				lbInfo["latestRecoverableTimestamp"] = lb.Status.LatestRecoverableTimestamp.Time.Format(time.RFC3339)
			}
			entry["latestBackup"] = lbInfo
		}
		if nt, ok := nextByCluster[cl.Name]; ok && nt != nil {
			entry["nextScheduledTime"] = nt.Time.Format(time.RFC3339)
		} else {
			entry["nextScheduledTime"] = nil
		}
		if lrt, ok := maxLrtByCluster[cl.Name]; ok && !lrt.IsZero() {
			entry["rpoSeconds"] = int(now.Sub(lrt).Seconds())
		} else {
			entry["rpoSeconds"] = nil
		}
		out = append(out, entry)
	}

	c.JSON(http.StatusOK, gin.H{"namespace": namespace, "total": len(out), "clusters": out})
}

func isBackupNewer(a, b polardbxv1.PolarDBXBackup) bool {
	endA := time.Time{}
	if a.Status.EndTime != nil {
		endA = a.Status.EndTime.Time
	}
	endB := time.Time{}
	if b.Status.EndTime != nil {
		endB = b.Status.EndTime.Time
	}
	if !endA.Equal(endB) {
		return endA.After(endB)
	}
	startA := time.Time{}
	if a.Status.StartTime != nil {
		startA = a.Status.StartTime.Time
	}
	startB := time.Time{}
	if b.Status.StartTime != nil {
		startB = b.Status.StartTime.Time
	}
	if !startA.Equal(startB) {
		return startA.After(startB)
	}
	return a.CreationTimestamp.After(b.CreationTimestamp.Time)
}

func getHpfsConfigMap(c *gin.Context, cli client.Client, namespace string) (*corev1.ConfigMap, error) {
	var cm corev1.ConfigMap
	if err := cli.Get(c.Request.Context(), client.ObjectKey{Namespace: namespace, Name: "polardbx-hpfs-config"}, &cm); err != nil {
		return nil, err
	}
	return &cm, nil
}

func decodeHpfsConfig(cm *corev1.ConfigMap) (hpfsconfig.Config, error) {
	var cfg hpfsconfig.Config
	data, ok := cm.Data["config.yaml"]
	if !ok {
		return cfg, nil
	}
	dec := yamlutil.NewYAMLOrJSONDecoder(bytes.NewBufferString(data), 4096)
	if err := dec.Decode(&cfg); err != nil {
		return hpfsconfig.Config{}, err
	}
	return cfg, nil
}

func createS3ClientFromSink(s hpfsconfig.Sink) (*minio.Client, string, bool) {
	endpoint := s.Endpoint
	secure := false
	low := strings.ToLower(endpoint)
	if strings.HasPrefix(low, "https://") {
		secure = true
		endpoint = endpoint[len("https://"):]
	} else if strings.HasPrefix(low, "http://") {
		secure = false
		endpoint = endpoint[len("http://"):]
	}
	if endpoint == "" || s.AccessKey == "" || s.AccessSecret == "" || s.Bucket == "" {
		return nil, "", false
	}
	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.AccessKey, s.AccessSecret, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, "", false
	}
	return cli, s.Bucket, true
}

func checkS3BucketOnline(c *gin.Context, cfg hpfsconfig.Config) bool {
	for _, s := range cfg.Sinks {
		if strings.EqualFold(s.Type, hpfsconfig.SinkTypeMinio) || strings.EqualFold(s.Type, hpfsconfig.SinkTypeOss) || strings.EqualFold(s.Type, "s3") {
			cli, bucket, ok := createS3ClientFromSink(s)
			if !ok {
				continue
			}
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()
			exists, err := cli.BucketExists(ctx, bucket)
			if err == nil && exists {
				return true
			}
		}
	}
	return false
}

func estimateTotalStorageBytes(c *gin.Context, cli client.Client, prefixes []string) (int64, bool) {
	systemNS := c.DefaultQuery("systemNamespace", "polardbx-operator-system")
	cm, err := getHpfsConfigMap(c, cli, systemNS)
	if err != nil {
		return 0, false
	}
	cfg, err := decodeHpfsConfig(cm)
	if err != nil {
		return 0, false
	}
	var sink *hpfsconfig.Sink
	for i := range cfg.Sinks {
		s := cfg.Sinks[i]
		if strings.EqualFold(s.Type, hpfsconfig.SinkTypeMinio) || strings.EqualFold(s.Type, hpfsconfig.SinkTypeOss) || strings.EqualFold(s.Type, "s3") {
			sink = &s
			break
		}
	}
	if sink == nil {
		return 0, false
	}
	cliS3, bucket, ok := createS3ClientFromSink(*sink)
	if !ok {
		return 0, false
	}
	if override := strings.TrimSpace(c.DefaultQuery("storagePrefixes", "")); override != "" {
		prefixes = strings.Split(override, ",")
	}
	norm := make([]string, 0, len(prefixes))
	for _, p := range prefixes {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		lp := strings.ToLower(p)
		if idx := strings.Index(lp, "://"); idx >= 0 {
			p = p[idx+3:]
		}
		if strings.HasPrefix(p, bucket+"/") {
			p = p[len(bucket)+1:]
		}
		norm = append(norm, p)
	}
	if len(norm) == 0 {
		return 0, false
	}
	maxObjects := 1000
	if v := strings.TrimSpace(c.DefaultQuery("maxObjects", "")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxObjects = n
		}
	}
	var total int64
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	processed := 0
	for _, pr := range norm {
		ch := cliS3.ListObjects(ctx, bucket, minio.ListObjectsOptions{Prefix: pr, Recursive: true})
		for obj := range ch {
			if obj.Err != nil {
				continue
			}
			total += obj.Size
			processed++
			if processed >= maxObjects {
				break
			}
		}
		if processed >= maxObjects {
			break
		}
	}
	return total, true
}

// GetBinlogMetrics moved from root api to backup package
func GetBinlogMetrics(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	namespace := c.DefaultQuery("namespace", "")
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

	var binlogList polardbxv1.PolarDBXBackupBinlogList
	binlogOpts := []client.ListOption{}
	if namespace != "" {
		binlogOpts = append(binlogOpts, client.InNamespace(namespace))
	}
	_ = cli.List(c.Request.Context(), &binlogList, binlogOpts...)

	var backupList polardbxv1.PolarDBXBackupList
	backupOpts := []client.ListOption{}
	if namespace != "" {
		backupOpts = append(backupOpts, client.InNamespace(namespace))
	}
	_ = cli.List(c.Request.Context(), &backupList, backupOpts...)

	maxLrtByCluster := map[string]time.Time{}
	for _, b := range backupList.Items {
		clusterName := b.Spec.Cluster.Name
		if clusterName == "" || b.Status.LatestRecoverableTimestamp == nil {
			continue
		}
		lrt := b.Status.LatestRecoverableTimestamp.Time
		if prev, ok := maxLrtByCluster[clusterName]; !ok || lrt.After(prev) {
			maxLrtByCluster[clusterName] = lrt
		}
	}

	items := make([]map[string]any, 0, len(binlogList.Items))
	for _, b := range binlogList.Items {
		clusterName := b.Spec.PxcName
		entry := map[string]any{
			"name":                b.Name,
			"namespace":           b.Namespace,
			"cluster":             clusterName,
			"phase":               string(b.Status.Phase),
			"lastCheckExpireTime": b.Status.CheckExpireFileLastTime,
			"recentDeletedFiles":  b.Status.LastDeletedFiles,
			"recentFiles":         b.Status.LastDeletedFiles,
			"throughputMBps":      "pending_implementation",
		}
		if lrt, ok := maxLrtByCluster[clusterName]; ok && !lrt.IsZero() {
			entry["latestBackupTime"] = lrt.UTC().Format(time.RFC3339)
			entry["lagSeconds"] = int(now.Sub(lrt).Seconds())
		}
		items = append(items, entry)
	}
	c.JSON(http.StatusOK, gin.H{"namespace": namespace, "total": len(items), "binlogs": items})
}
