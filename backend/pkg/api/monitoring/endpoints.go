package monitoring

import (
	"net/http"
	"time"

	"polardbx-ui-backend/pkg/api/util"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		"iops": gin.H{
			"estimated": false,
			"ok":        false,
			"message":   "IOPS 测试占位，未执行真实磁盘基准",
		},
	})
}
