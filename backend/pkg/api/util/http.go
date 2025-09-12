package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sClientFromContext returns controller-runtime client from gin context.
func K8sClientFromContext(c *gin.Context) (client.Client, bool) {
	v, ok := c.Get("k8sClient")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "kubernetes client not initialized"})
		return nil, false
	}
	cli, ok := v.(client.Client)
	if !ok || cli == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid kubernetes client in context"})
		return nil, false
	}
	return cli, true
}

// ClientsetFromContext returns client-go clientset from gin context.
func ClientsetFromContext(c *gin.Context) (kubernetes.Interface, bool) {
	v, ok := c.Get("clientset")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "kubeconfig not provided or invalid"})
		return nil, false
	}
	cs, ok := v.(kubernetes.Interface)
	if !ok || cs == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid clientset in context"})
		return nil, false
	}
	return cs, true
}

// DefaultNamespace returns query namespace or the middleware-injected default.
func DefaultNamespace(c *gin.Context, fallback string) string {
	if ns := c.Query("namespace"); ns != "" {
		return ns
	}
	if v, ok := c.Get("k8sDefaultNamespace"); ok {
		if s, ok2 := v.(string); ok2 && s != "" {
			return s
		}
	}
	return fallback
}

// HandleK8sError maps common k8s errors to HTTP codes.
func HandleK8sError(c *gin.Context, context string, err error) {
	switch {
	case k8serrors.IsInvalid(err), k8serrors.IsBadRequest(err):
		c.JSON(http.StatusBadRequest, gin.H{"error": context, "details": err.Error()})
	case k8serrors.IsAlreadyExists(err):
		c.JSON(http.StatusConflict, gin.H{"error": context, "details": err.Error()})
	case k8serrors.IsNotFound(err):
		c.JSON(http.StatusNotFound, gin.H{"error": context, "details": err.Error()})
	case k8serrors.IsForbidden(err):
		c.JSON(http.StatusForbidden, gin.H{"error": context, "details": err.Error()})
	case k8serrors.IsUnauthorized(err):
		c.JSON(http.StatusUnauthorized, gin.H{"error": context, "details": err.Error()})
	case k8serrors.IsTimeout(err):
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": context, "details": err.Error()})
	case k8serrors.IsTooManyRequests(err):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": context, "details": err.Error()})
	case k8serrors.IsServiceUnavailable(err):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": context, "details": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": context, "details": err.Error()})
	}
}
