package alerts

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"polardbx-ui-backend/pkg/api/util"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// List aggregates alerts from Alertmanager and K8s Events
func List(c *gin.Context) {
	k8sClient, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	namespace := c.DefaultQuery("namespace", "")
	cluster := c.DefaultQuery("cluster", "")
	alertmanagerURL := c.Query("alertmanager")
	items := make([]map[string]any, 0)

	// from alertmanager
	if alertmanagerURL != "" {
		type amAlert struct {
			Labels      map[string]string `json:"labels"`
			Annotations map[string]string `json:"annotations"`
			StartsAt    string            `json:"startsAt"`
		}
		var alerts []amAlert
		if resp, err := http.Get(alertmanagerURL + "/api/v2/alerts"); err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			if err := json.NewDecoder(resp.Body).Decode(&alerts); err == nil {
				for _, a := range alerts {
					if (namespace == "" || a.Labels["namespace"] == namespace) && (cluster == "" || a.Labels["cluster"] == cluster) {
						msg := a.Annotations["summary"]
						if msg == "" {
							msg = a.Annotations["description"]
						}
						if msg == "" {
							msg = a.Labels["alertname"]
						}
						items = append(items, map[string]any{
							"source":    "alertmanager",
							"severity":  a.Labels["severity"],
							"labels":    a.Labels,
							"message":   msg,
							"timestamp": a.StartsAt,
						})
					}
				}
			}
		}
	}

	// from k8s events
	var evList corev1.EventList
	opts := []client.ListOption{}
	if namespace != "" {
		opts = append(opts, client.InNamespace(namespace))
	}
	if err := k8sClient.List(c.Request.Context(), &evList, opts...); err == nil {
		for _, ev := range evList.Items {
			if cluster != "" && !strings.Contains(ev.InvolvedObject.Name, cluster) {
				continue
			}
			sev := "info"
			if ev.Type == corev1.EventTypeWarning {
				sev = "warning"
			}
			labels := map[string]string{
				"namespace":      ev.Namespace,
				"cluster":        cluster,
				"reason":         ev.Reason,
				"involvedObject": ev.InvolvedObject.Name,
			}
			items = append(items, map[string]any{
				"source":    "k8s-event",
				"severity":  sev,
				"message":   ev.Message,
				"labels":    labels,
				"time":      ev.LastTimestamp.Time.Format(time.RFC3339),
				"timestamp": ev.LastTimestamp.Time.Format(time.RFC3339),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}
