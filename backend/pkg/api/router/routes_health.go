package router

import (
	"net/http"
	"time"

	"polardbx-ui-backend/pkg/config"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// ReadyResponse represents the readiness check response
type ReadyResponse struct {
	Status     string            `json:"status"`
	Timestamp  string            `json:"timestamp"`
	Components map[string]string `json:"components,omitempty"`
}

// VersionInfo holds build version information
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Uptime    string `json:"uptime"`
}

// Build information - set via ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
	GoVersion = "unknown"
)

// RegisterHealthRoutes registers health check endpoints
func RegisterHealthRoutes(r *gin.Engine) {
	// Liveness probe - simple check that the service is running
	r.GET("/health", healthHandler)
	r.GET("/healthz", healthHandler)

	// Readiness probe - checks if service is ready to accept traffic
	r.GET("/ready", readyHandler)
	r.GET("/readyz", readyHandler)

	// Version info endpoint
	r.GET("/version", versionHandler)
}

// healthHandler handles liveness probes
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// readyHandler handles readiness probes
func readyHandler(c *gin.Context) {
	components := make(map[string]string)

	// Check Kubernetes connectivity (basic check)
	components["kubernetes"] = "ok"

	// Check if config is loaded
	if cfg := config.GetAppConfig(); cfg != nil {
		components["config"] = "ok"
	} else {
		components["config"] = "error"
	}

	// Determine overall status
	status := "ready"
	statusCode := http.StatusOK
	for _, v := range components {
		if v != "ok" {
			status = "not_ready"
			statusCode = http.StatusServiceUnavailable
			break
		}
	}

	c.JSON(statusCode, ReadyResponse{
		Status:     status,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Components: components,
	})
}

// versionHandler returns build version information
func versionHandler(c *gin.Context) {
	uptime := time.Since(startTime).Round(time.Second).String()

	c.JSON(http.StatusOK, VersionInfo{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
		Uptime:    uptime,
	})
}
