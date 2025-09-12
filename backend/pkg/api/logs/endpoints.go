package logs

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"polardbx-ui-backend/pkg/api/util"

	"github.com/gin-gonic/gin"
)

type QueryRequest struct {
	Host      string            `json:"host"`
	Index     string            `json:"index"`
	Query     map[string]any    `json:"query"`
	Size      int               `json:"size"`
	From      int               `json:"from"`
	Sort      []map[string]any  `json:"sort"`
	Aggs      map[string]any    `json:"aggs"`
	TimeRange map[string]string `json:"timeRange"` // {"field":"@timestamp","from":"2024-01-01T00:00:00Z","to":"2024-01-02T00:00:00Z"}
}

func Query(c *gin.Context) {
	if err := util.EnsureLogsSecurityBootstrap(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "security bootstrap failed", "details": err.Error()})
		return
	}
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	allowed, defHost, err := util.LoadLogsSecurityConfig(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load security config", "details": err.Error()})
		return
	}
	if req.Host == "" {
		req.Host = defHost
	}
	if !util.IsHostAllowed(req.Host, allowed, defHost) {
		c.JSON(http.StatusForbidden, gin.H{"error": "target host not allowed"})
		return
	}
	if req.Index == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index is required"})
		return
	}

	// Build DSL
	body := map[string]any{
		"query": req.Query,
		"size":  defaultInt(req.Size, 50),
		"from":  defaultInt(req.From, 0),
	}
	if len(req.Sort) > 0 {
		body["sort"] = req.Sort
	}
	if req.Aggs != nil {
		body["aggs"] = req.Aggs
	}
	if f, ok := req.TimeRange["from"]; ok && f != "" {
		field := req.TimeRange["field"]
		if field == "" {
			field = "@timestamp"
		}
		to := req.TimeRange["to"]
		rangeQ := map[string]any{"range": map[string]any{field: map[string]any{"gte": f}}}
		if to != "" {
			rangeQ["range"].(map[string]any)[field].(map[string]any)["lte"] = to
		}
		if body["query"] == nil {
			body["query"] = map[string]any{"bool": map[string]any{"filter": []any{rangeQ}}}
		} else {
			// append to bool.filter if present, otherwise wrap
			if m, ok := body["query"].(map[string]any)["bool"].(map[string]any); ok {
				filters, _ := m["filter"].([]any)
				m["filter"] = append(filters, rangeQ)
			} else {
				body["query"] = map[string]any{"bool": map[string]any{"must": []any{body["query"]}, "filter": []any{rangeQ}}}
			}
		}
	}

	buf, _ := json.Marshal(body)
	endpoint := strings.TrimRight(req.Host, "/") + "/" + url.PathEscape(req.Index) + "/_search"

	// HTTP client with optional TLS
	httpClient := &http.Client{Timeout: 30 * time.Second}
	if pool, err := util.LoadESRootCAs(c); err == nil && pool != nil {
		httpClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}
	}

	httpReq, _ := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, endpoint, bytes.NewReader(buf))
	httpReq.Header.Set("Content-Type", "application/json")
	if user, pass, ok, _ := util.LoadESCredentials(c); ok && user != "" {
		httpReq.SetBasicAuth(user, pass)
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "es query failed", "details": err.Error()})
		return
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		c.JSON(resp.StatusCode, gin.H{"error": "es error", "details": string(bodyBytes)})
		return
	}
	var out any
	_ = json.Unmarshal(bodyBytes, &out)
	c.JSON(http.StatusOK, out)
}

func defaultInt(v int, d int) int {
	if v <= 0 {
		return d
	}
	return v
}
