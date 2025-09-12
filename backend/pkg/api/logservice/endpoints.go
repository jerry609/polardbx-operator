package logservice

import (
	"net/http"

	"polardbx-ui-backend/pkg/api/util"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Status aggregates log-collector components readiness and config presence
// Query: namespace (default: polardbx-logcollector)
func Status(c *gin.Context) {
	clientset, ok := util.ClientsetFromContext(c)
	if !ok {
		return
	}
	ns := c.DefaultQuery("namespace", "polardbx-logcollector")

	// Probe components
	var ds *appsv1.DaemonSet
	dsObj, dsErr := clientset.AppsV1().DaemonSets(ns).Get(c.Request.Context(), "filebeat", metav1.GetOptions{})
	if dsErr == nil {
		ds = dsObj
	}
	var dep *appsv1.Deployment
	depObj, depErr := clientset.AppsV1().Deployments(ns).Get(c.Request.Context(), "logstash", metav1.GetOptions{})
	if depErr == nil {
		dep = depObj
	}
	// Pipeline config
	var cm *corev1.ConfigMap
	cmObj, _ := clientset.CoreV1().ConfigMaps(ns).Get(c.Request.Context(), "logstash-pipeline", metav1.GetOptions{})
	if cmObj != nil {
		cm = cmObj
	}

	existsFB := ds != nil
	existsLS := dep != nil
	readyFB := int32(0)
	desiredFB := int32(0)
	if ds != nil {
		readyFB = ds.Status.NumberReady
		desiredFB = ds.Status.DesiredNumberScheduled
	}
	readyLS := int32(0)
	desiredLS := int32(0)
	if dep != nil {
		readyLS = dep.Status.ReadyReplicas
		desiredLS = dep.Status.Replicas
	}

	state := "not_installed"
	if existsFB || existsLS {
		state = "degraded"
		if existsFB && existsLS && desiredFB > 0 && desiredLS > 0 && readyFB == desiredFB && readyLS == desiredLS {
			state = "running"
		}
	}

	resp := gin.H{
		"namespace": ns,
		"components": gin.H{
			"filebeat": gin.H{"exists": existsFB, "ready": readyFB, "desired": desiredFB, "error": errString(dsErr)},
			"logstash": gin.H{"exists": existsLS, "ready": readyLS, "desired": desiredLS, "error": errString(depErr)},
		},
		"pipelineConfigMapExists": cm != nil,
		"state":                   state,
	}
	if state == "not_installed" {
		resp["installHint"] = "helm install polardbx-logcollector charts/polardbx-logcollector -n polardbx-logcollector --create-namespace"
	}
	c.JSON(http.StatusOK, resp)
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
