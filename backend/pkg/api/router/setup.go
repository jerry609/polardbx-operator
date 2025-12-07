package router

import (
	"log"
	"strconv"
	"strings"

	"polardbx-ui-backend/pkg/api"
	apierr "polardbx-ui-backend/pkg/api/errors"
	"polardbx-ui-backend/pkg/api/middleware"
	"polardbx-ui-backend/pkg/config"

	"github.com/gin-gonic/gin"
)

// SetupRouter creates and configures the main Gin router
func SetupRouter() *gin.Engine {
	cfg := config.GetServerConfig()

	// Set Gin mode from config
	gin.SetMode(cfg.Mode)

	r := gin.New()

	// Setup middleware
	setupMiddleware(r, cfg)

	// Setup routes
	setupRoutes(r)

	return r
}

// setupMiddleware configures all middleware
func setupMiddleware(r *gin.Engine, cfg *config.ServerConfig) {
	// Request ID middleware (first, for tracing)
	r.Use(apierr.RequestIDMiddleware())

	// Recovery middleware - use our enhanced version that returns APIError
	r.Use(apierr.RecoveryHandler())

	// Global error handler - catches any unhandled errors and formats them
	r.Use(apierr.Handler())

	// Request logging middleware
	r.Use(middleware.RequestLogger(middleware.RequestLogConfig{
		LogRequestBody:  cfg.LogRequestBody,
		LogResponseBody: cfg.LogResponseBody,
		MaxBodyLogSize:  cfg.LogMaxBodySize,
		SkipPaths:       cfg.LogSkipPaths,
		SensitiveFields: cfg.LogSensitiveFields,
	}))

	// CORS middleware
	r.Use(CORSMiddleware(cfg))
}

// setupRoutes registers all route groups
func setupRoutes(r *gin.Engine) {
	// Health check endpoints (no auth required)
	RegisterHealthRoutes(r)

	// Ping endpoint
	r.GET("/ping", func(c *gin.Context) {
		apierr.OK(c, gin.H{"message": "pong"})
	})

	// API v1 group
	v1 := r.Group("/api/v1")

	// Public routes (no kubeconfig required)
	RegisterPublicRoutes(v1)

	// Auth routes
	RegisterAuthRoutes(v1)

	// Protected routes (require kubeconfig)
	v1.Use(api.KubeconfigAuthMiddleware())
	{
		RegisterClusterRoutes(v1)
		RegisterBackupRoutes(v1)
		RegisterXStoreRoutes(v1)
		RegisterMonitoringRoutes(v1)
		RegisterLogsRoutes(v1)
		RegisterSystemRoutes(v1)
		RegisterDiagnosticsRoutes(v1)
		RegisterRestoreRoutes(v1)
	}

	// Register CRD-aligned alias routes
	RegisterCRDAliasRoutes(v1)
	RegisterDomainRoutes(v1)
}

// CORSMiddleware creates CORS middleware with config
func CORSMiddleware(cfg *config.ServerConfig) gin.HandlerFunc {
	// Pre-compute headers
	allowOrigin := "*"
	if len(cfg.CORSAllowOrigins) > 0 {
		allowOrigin = cfg.CORSAllowOrigins[0]
	}
	allowMethods := strings.Join(cfg.CORSAllowMethods, ", ")
	allowHeaders := strings.Join(cfg.CORSAllowHeaders, ", ")
	exposeHeaders := strings.Join(cfg.CORSExposeHeaders, ", ")
	maxAge := strconv.Itoa(cfg.CORSMaxAge)
	allowCredentials := "false"
	if cfg.CORSAllowCredentials {
		allowCredentials = "true"
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		originAllowed := allowOrigin == "*"
		if !originAllowed && origin != "" {
			for _, allowed := range cfg.CORSAllowOrigins {
				if allowed == origin || allowed == "*" {
					originAllowed = true
					break
				}
			}
		}

		if originAllowed {
			if allowOrigin == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			}
			c.Header("Access-Control-Allow-Credentials", allowCredentials)
			c.Header("Access-Control-Allow-Headers", allowHeaders)
			c.Header("Access-Control-Allow-Methods", allowMethods)
			c.Header("Access-Control-Expose-Headers", exposeHeaders)
			c.Header("Access-Control-Max-Age", maxAge)
		}

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// LogRoutes logs all registered routes
func LogRoutes(r *gin.Engine) {
	log.Println("=== Registered Routes ===")
	for _, route := range r.Routes() {
		log.Printf("  %s %s", route.Method, route.Path)
	}
	log.Printf("=== Total: %d routes ===", len(r.Routes()))
}
