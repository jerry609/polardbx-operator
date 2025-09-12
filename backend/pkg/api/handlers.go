package api

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"polardbx-ui-backend/pkg/k8s"
)

// KubeconfigAuthMiddleware validates the provided kubeconfig from the request header.
func KubeconfigAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		kubeconfigB64 := c.GetHeader("X-Kubeconfig-B64")
		// WebSocket 等场景无法自定义 Header 时，允许通过查询参数传递
		if kubeconfigB64 == "" {
			if v := c.Query("kubeconfig"); v != "" {
				kubeconfigB64 = v
			}
		}
		if kubeconfigB64 == "" {
			if v := c.Query("k"); v != "" {
				kubeconfigB64 = v
			}
		}
		if kubeconfigB64 == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "kubeconfig not provided"})
			c.Abort()
			return
		}

		// Decode the base64 kubeconfig
		kubeconfig, err := base64.StdEncoding.DecodeString(kubeconfigB64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid kubeconfig base64", "details": err.Error()})
			c.Abort()
			return
		}

		// Create Kubernetes clients
		ctrlClient, clientset, err := k8s.NewClientsFromKubeconfig(kubeconfig)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to create kubernetes clients", "details": err.Error()})
			c.Abort()
			return
		}

		c.Set("k8sClient", ctrlClient)
		c.Set("clientset", clientset)

		// Extract identity for audit (best-effort)
		if cfg, err := clientcmd.Load(kubeconfig); err == nil && cfg != nil {
			ctxName := cfg.CurrentContext
			user := ctxName
			if ctx, ok := cfg.Contexts[ctxName]; ok && ctx != nil {
				if ctx.AuthInfo != "" {
					user = ctx.AuthInfo
				}
				if ctx.Namespace != "" {
					c.Set("k8sDefaultNamespace", ctx.Namespace)
				}
			}
			c.Set("k8sUser", user)
			c.Set("k8sContext", ctxName)
		}
		c.Next()
	}
}

// Connect handler validates the provided kubeconfig from the request header.
func Connect(c *gin.Context) {
	// The actual connection test is handled by the KubeconfigAuthMiddleware.
	// If we reach here, it means the client is valid.
	c.JSON(http.StatusOK, gin.H{"message": "kubeconfig is valid and connection successful"})
}

// handleK8sError checks the error from the kubernetes client and returns the appropriate HTTP status code.
func handleK8sError(c *gin.Context, contextMsg string, err error) {
	if k8serrors.IsInvalid(err) || k8serrors.IsBadRequest(err) {
		c.JSON(http.StatusBadRequest, gin.H{"error": contextMsg, "details": err.Error()})
		return
	}
	if k8serrors.IsAlreadyExists(err) {
		c.JSON(http.StatusConflict, gin.H{"error": contextMsg, "details": err.Error()})
		return
	}
	if k8serrors.IsNotFound(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": contextMsg, "details": err.Error()})
		return
	}
	if k8serrors.IsForbidden(err) {
		c.JSON(http.StatusForbidden, gin.H{"error": contextMsg, "details": err.Error()})
		return
	}
	if k8serrors.IsUnauthorized(err) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": contextMsg, "details": err.Error()})
		return
	}
	if k8serrors.IsTimeout(err) {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": contextMsg, "details": err.Error()})
		return
	}
	if k8serrors.IsTooManyRequests(err) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": contextMsg, "details": err.Error()})
		return
	}
	if k8serrors.IsServiceUnavailable(err) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": contextMsg, "details": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": contextMsg, "details": err.Error()})
}

// clientFromContext retrieves the kubernetes client from the gin context.
func clientFromContext(c *gin.Context) (client.Client, bool) {
	k8sClientVal, ok := c.Get("k8sClient")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "kubeconfig not provided or invalid"})
		c.Abort()
		return nil, false
	}
	k8sClient := k8sClientVal.(client.Client)
	return k8sClient, true
}
