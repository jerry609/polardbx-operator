package monitoring

import (
	"net/http"
	"time"
	"context"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"

	"polardbx-ui-backend/pkg/api/util"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Bootstrap installs or registers monitoring stack. For now persists a bootstrap plan ConfigMap.
func Bootstrap(c *gin.Context) {
	type req struct {
		Mode string `json:"mode"` // managed|assisted|byo
		Dry  bool   `json:"dryRun"`
		NS   string `json:"namespace"`
		Name string `json:"releaseName"`
	}
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	var r req
	_ = c.ShouldBindJSON(&r)
	if r.NS == "" {
		r.NS = "polardbx-operator-system"
	}
	cm := corev1.ConfigMap{}
	key := client.ObjectKey{Namespace: r.NS, Name: "polardbx-monitoring-plan"}
	if err := cli.Get(c.Request.Context(), key, &cm); err != nil {
		cm = corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: r.NS, Name: "polardbx-monitoring-plan"}, Data: map[string]string{}}
		_ = cli.Create(c.Request.Context(), &cm)
	}
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	cm.Data["mode"] = r.Mode
	cm.Data["releaseName"] = r.Name
	cm.Data["dryRun"] = map[bool]string{true: "true", false: "false"}[r.Dry]
	_ = cli.Update(c.Request.Context(), &cm)
	c.JSON(http.StatusAccepted, gin.H{"message": "monitoring bootstrap accepted", "namespace": r.NS, "mode": r.Mode, "releaseName": r.Name, "dryRun": r.Dry})
}

// Status summarizes discovered components readiness.
func Status(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "polardbx-operator-system")

	checkDeploy := func(name string) (ready, desired int32, ok bool) {
		dep := appsv1.Deployment{}
		if err := cli.Get(c.Request.Context(), client.ObjectKey{Namespace: ns, Name: name}, &dep); err == nil {
			return dep.Status.ReadyReplicas, dep.Status.Replicas, true
		}
		return 0, 0, false
	}
	checkStateful := func(name string) (ready, desired int32, ok bool) {
		sts := appsv1.StatefulSet{}
		if err := cli.Get(c.Request.Context(), client.ObjectKey{Namespace: ns, Name: name}, &sts); err == nil {
			return sts.Status.ReadyReplicas, *sts.Spec.Replicas, true
		}
		return 0, 0, false
	}
	checkService := func(name string) bool {
		svc := corev1.Service{}
		return cli.Get(c.Request.Context(), client.ObjectKey{Namespace: ns, Name: name}, &svc) == nil
	}

	prom := gin.H{"ready": false}
	if r, d, ok := checkStateful("prometheus-k8s"); ok {
		prom = gin.H{"ready": r == d, "readyReplicas": r, "replicas": d}
	}
	if !prom["ready"].(bool) {
		if r, d, ok := checkStateful("kube-prometheus-stack-prometheus"); ok {
			prom = gin.H{"ready": r == d, "readyReplicas": r, "replicas": d}
		}
	}
	prom["service"] = checkService("prometheus-k8s") || checkService("kube-prometheus-stack-prometheus")

	graf := gin.H{"ready": false}
	if r, d, ok := checkDeploy("grafana"); ok {
		graf = gin.H{"ready": r == d, "readyReplicas": r, "replicas": d}
	}
	if !graf["ready"].(bool) {
		if r, d, ok := checkDeploy("kube-prometheus-stack-grafana"); ok {
			graf = gin.H{"ready": r == d, "readyReplicas": r, "replicas": d}
		}
	}
	graf["service"] = checkService("grafana") || checkService("kube-prometheus-stack-grafana")

	am := gin.H{"configured": checkService("alertmanager-main") || checkService("kube-prometheus-stack-alertmanager")}

	c.JSON(http.StatusOK, gin.H{
		"namespace": ns,
		"components": gin.H{
			"prometheus":   prom,
			"grafana":      graf,
			"alertmanager": am,
		},
	})
}

// Uninstall removes lightweight markers (stub). In managed mode, a SystemTask will handle teardown later.
func Uninstall(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "polardbx-operator-system")
	_ = cli.Delete(c.Request.Context(), &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "polardbx-monitoring-plan"}})
	c.JSON(http.StatusOK, gin.H{"message": "monitoring uninstall request accepted"})
}

// runIOPSBench launches a short-lived pod to estimate write IOPS on node ephemeral storage.
func runIOPSBench(ctx context.Context, cs kubernetes.Interface, ns string) map[string]interface{} {
	name := fmt.Sprintf("iops-bench-%d", rand.Intn(1_000_000))
	script := strings.Join([]string{
		"set -e",
		"cd /data",
		"COUNT=${COUNT:-50000}", // 50k ops @4k ≈ 200MB
		"BS=${BS:-4096}",
		"rm -f testfile || true",
		"OUT=$(dd if=/dev/zero of=testfile bs=$BS count=$COUNT conv=fdatasync 2>&1 | tail -1)",
		"SEC=$(echo \"$OUT\" | awk -F', ' '{print $(NF-1)}' | awk '{print $1}')",
		"if [ -z \"$SEC\" ]; then SEC=0; fi",
		"if [ \"$SEC\" = \"0\" ]; then IOPS=0; else IOPS=$(awk -v c=$COUNT -v s=$SEC 'BEGIN{printf(\"%d\", c/s)}'); fi",
		"echo {\\\"write\\\":{\\\"bs\\\":$BS,\\\"ops\\\":$COUNT,\\\"seconds\\\":$SEC,\\\"iops\\\":$IOPS}}",
	}, "; ")
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Volumes: []corev1.Volume{{Name: "data", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}},
			Containers: []corev1.Container{{
				Name:         "bench",
				Image:        "busybox:1.36",
				Command:      []string{"sh", "-c", script},
				VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: "/data"}},
			}},
		},
	}
	_, err := cs.CoreV1().Pods(ns).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return map[string]interface{}{"estimated": false, "ok": false, "message": "无权限或创建 Pod 失败: " + err.Error()}
	}
	defer func() { _ = cs.CoreV1().Pods(ns).Delete(context.Background(), name, metav1.DeleteOptions{}) }()
	// wait for completion
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		p, e := cs.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})
		if e == nil {
			phase := p.Status.Phase
			if phase == corev1.PodSucceeded || phase == corev1.PodFailed {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
	// fetch logs
	logReq := cs.CoreV1().Pods(ns).GetLogs(name, &corev1.PodLogOptions{Container: "bench"})
	rc, e := logReq.Stream(ctx)
	if e != nil {
		return map[string]interface{}{"estimated": false, "ok": false, "message": "无法读取基准日志: " + e.Error()}
	}
	defer rc.Close()
	b, _ := io.ReadAll(rc)
	line := strings.TrimSpace(string(b))
	// expected: {"write":{"bs":4096,"ops":50000,"seconds":1.23,"iops":40650}}
	res := map[string]interface{}{"estimated": true, "ok": false, "message": "未能解析输出"}
	if strings.HasPrefix(line, "{") {
		// very small parser
		ok := false
		var iops int64 = 0
		var seconds float64 = 0
		var ops int64 = 0
		var bs int64 = 0
		// parse by splitting (avoid bringing full json dep)
		get := func(key string) string {
			idx := strings.Index(line, key)
			if idx < 0 { return "" }
			s := line[idx+len(key):]
			s = strings.TrimLeft(s, ":")
			s = strings.TrimLeft(s, " ")
			i := strings.IndexAny(s, ",}")
			if i < 0 { return s }
			return s[:i]
		}
		if v := get("\"iops\""); v != "" { if n, err := strconv.ParseInt(strings.Trim(v, " \""), 10, 64); err == nil { iops = n; ok = true } }
		if v := get("\"seconds\""); v != "" { if f, err := strconv.ParseFloat(strings.Trim(v, " \""), 64); err == nil { seconds = f } }
		if v := get("\"ops\""); v != "" { if n, err := strconv.ParseInt(strings.Trim(v, " \""), 10, 64); err == nil { ops = n } }
		if v := get("\"bs\""); v != "" { if n, err := strconv.ParseInt(strings.Trim(v, " \""), 10, 64); err == nil { bs = n } }
		res = map[string]interface{}{
			"estimated": ok,
			"ok":       ok && iops > 0,
			"iops":     iops,
			"seconds":  seconds,
			"ops":      ops,
			"bs":       bs,
		}
	}
	return res
}

// Preflight performs lightweight checks before installing/using monitoring stack.
// It provides placeholders for IOPS and clock skew checks and a real AZ distribution summary.
func Preflight(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}

	// List nodes to infer AZ distribution
	nodes := corev1.NodeList{}
	_ = cli.List(c.Request.Context(), &nodes)

	zonesSet := map[string]struct{}{}
	nodesWithZone := 0
	for _, n := range nodes.Items {
		lbl := n.Labels
		zone := lbl["topology.kubernetes.io/zone"]
		if zone == "" {
			zone = lbl["failure-domain.beta.kubernetes.io/zone"]
		}
		if zone != "" {
			zonesSet[zone] = struct{}{}
			nodesWithZone++
		}
	}
	zones := make([]string, 0, len(zonesSet))
	for z := range zonesSet {
		zones = append(zones, z)
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// Run a short IOPS benchmark via a dedicated pod (best-effort)
	iops := gin.H{"estimated": false, "ok": false, "message": "IOPS 测试未执行"}
	if cs, ok2 := util.ClientsetFromContext(c); ok2 {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
		defer cancel()
		ns := util.DefaultNamespace(c, "polardbx-operator-system")
		res := runIOPSBench(ctx, cs, ns)
		iops = gin.H(res)
	}

	c.JSON(http.StatusOK, gin.H{
		"timestamp": now,
		"az": gin.H{
			"zones":         zones,
			"count":         len(zones),
			"nodeCount":     len(nodes.Items),
			"nodesWithZone": nodesWithZone,
			"hasMultiple":   len(zones) >= 2,
		},
		"clock": gin.H{
			"controllerTime": now,
			"skewAssessed":   false,
			"ok":             true,
			"message":        "未校验集群节点时钟漂移（占位）",
		},
		"iops": iops,
	})
}
