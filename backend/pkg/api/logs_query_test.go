package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	api_logs "polardbx-ui-backend/pkg/api/logs"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func setupLogsRouterWithClientset(cs *k8sfake.Clientset) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/api/v1")
	{
		v1.Use(func(c *gin.Context) { c.Set("clientset", cs) })
		v1.POST("/logs/query", api_logs.Query)
	}
	return r
}

func TestLogsQuery_ForbiddenWhenHostNotAllowed(t *testing.T) {
	cs := k8sfake.NewSimpleClientset(
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "logs-config", Namespace: "polardbx-logcollector"}, Data: map[string]string{
			"allowedHosts": "http://allowed.local",
			"defaultHost":  "",
		}},
	)
	r := setupLogsRouterWithClientset(cs)

	body := `{"host":"http://not-allowed.local","index":"logs-*","query":{"match_all":{}}}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/logs/query", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestLogsQuery_TimeRangeDSL_BuildsAndPosts(t *testing.T) {
	var captured struct {
		Path string
		Body map[string]any
	}
	es := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.Path = r.URL.Path
		var b map[string]any
		_ = json.NewDecoder(r.Body).Decode(&b)
		captured.Body = b
		_, _ = w.Write([]byte(`{"hits":{"total":{"value":1},"hits":[{"_source":{"msg":"ok"}}]}}`))
	}))
	defer es.Close()

	cs := k8sfake.NewSimpleClientset(
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "logs-config", Namespace: "polardbx-logcollector"}, Data: map[string]string{
			"allowedHosts": es.URL,
			"defaultHost":  "",
		}},
	)
	r := setupLogsRouterWithClientset(cs)

	body := `{"host":"` + es.URL + `","index":"logs-test","query":{"match_all":{}},"timeRange":{"field":"@timestamp","from":"2024-05-01T00:00:00Z","to":"2024-05-01T01:00:00Z"},"size":5}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/logs/query", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}

	if captured.Path != "/logs-test/_search" {
		t.Fatalf("unexpected path to ES: %s", captured.Path)
	}
	// Verify range filter is present
	q := captured.Body["query"].(map[string]any)
	b, ok := q["bool"].(map[string]any)
	if !ok {
		t.Fatalf("expected bool query wrapper")
	}
	filters, ok := b["filter"].([]any)
	if !ok || len(filters) == 0 {
		t.Fatalf("expected at least one filter")
	}
	found := false
	for _, f := range filters {
		m, ok := f.(map[string]any)["range"].(map[string]any)
		if ok {
			if ts, ok2 := m["@timestamp"].(map[string]any); ok2 {
				if ts["gte"] == "2024-05-01T00:00:00Z" && ts["lte"] == "2024-05-01T01:00:00Z" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatalf("expected range filter on @timestamp with gte/lte")
	}
}
