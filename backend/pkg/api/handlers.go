package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	apierr "polardbx-ui-backend/pkg/api/errors"
	apiutil "polardbx-ui-backend/pkg/api/util"
	"polardbx-ui-backend/pkg/k8s"
	"polardbx-ui-backend/pkg/logger"
)

var newAllClientsFromKubeconfig = k8s.NewAllClientsFromKubeconfig

func normalizeKubeconfig(raw []byte) ([]byte, *clientcmdapi.Config, error) {
	cfg, err := clientcmd.Load(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}
	if cfg == nil {
		return nil, nil, fmt.Errorf("kubeconfig is empty")
	}

	for name, cluster := range cfg.Clusters {
		if cluster == nil {
			continue
		}
		if len(cluster.CertificateAuthorityData) == 0 && cluster.CertificateAuthority != "" {
			logger.Info("normalizeKubeconfig: inlining certificate authority",
				"cluster", name,
				"path", cluster.CertificateAuthority)
			data, err := readCredentialFile(cluster.CertificateAuthority)
			if err != nil {
				return nil, nil, fmt.Errorf("cluster %q certificate-authority %q: %w", name, cluster.CertificateAuthority, err)
			}
			cluster.CertificateAuthorityData = data
			cluster.CertificateAuthority = ""
		}
	}

	for name, authInfo := range cfg.AuthInfos {
		if authInfo == nil {
			continue
		}
		if len(authInfo.ClientCertificateData) == 0 && authInfo.ClientCertificate != "" {
			logger.Info("normalizeKubeconfig: inlining client certificate",
				"user", name,
				"path", authInfo.ClientCertificate)
			data, err := readCredentialFile(authInfo.ClientCertificate)
			if err != nil {
				return nil, nil, fmt.Errorf("user %q client-certificate %q: %w", name, authInfo.ClientCertificate, err)
			}
			authInfo.ClientCertificateData = data
			authInfo.ClientCertificate = ""
		}
		if len(authInfo.ClientKeyData) == 0 && authInfo.ClientKey != "" {
			logger.Info("normalizeKubeconfig: inlining client key",
				"user", name,
				"path", authInfo.ClientKey)
			data, err := readCredentialFile(authInfo.ClientKey)
			if err != nil {
				return nil, nil, fmt.Errorf("user %q client-key %q: %w", name, authInfo.ClientKey, err)
			}
			authInfo.ClientKeyData = data
			authInfo.ClientKey = ""
		}
	}

	normalized, err := clientcmd.Write(*cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize kubeconfig: %w", err)
	}

	return normalized, cfg, nil
}

func readCredentialFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("credential path is empty")
	}

	expanded := os.ExpandEnv(path)
	if strings.HasPrefix(expanded, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			switch {
			case expanded == "~":
				expanded = home
			case strings.HasPrefix(expanded, "~/"):
				expanded = filepath.Join(home, expanded[2:])
			case strings.HasPrefix(expanded, "~"+string(os.PathSeparator)):
				expanded = filepath.Join(home, expanded[2:])
			}
		}
	}

	expanded = filepath.Clean(expanded)

	data, err := os.ReadFile(expanded)
	if err != nil {
		return nil, fmt.Errorf("read credential file %q: %w", expanded, err)
	}
	return data, nil
}

// KubeconfigAuthMiddleware validates the provided kubeconfig from the request header.
func KubeconfigAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestPath := c.FullPath()
		if requestPath == "" {
			requestPath = c.Request.URL.Path
		}
		logger.Info("KubeconfigAuthMiddleware: handling request",
			"method", c.Request.Method,
			"path", requestPath,
			"clientIP", c.ClientIP())

		kubeconfigB64 := c.GetHeader("X-Kubeconfig-B64")
		// Allow passing via query parameter when custom headers cannot be set (e.g., WebSocket scenarios)
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
			logger.Warn("KubeconfigAuthMiddleware: missing kubeconfig",
				"path", requestPath,
				"clientIP", c.ClientIP())
			apierr.AbortUnauthorized(c, "kubeconfig not provided")
			return
		}

		// Decode the base64 kubeconfig
		kubeconfig, err := base64.StdEncoding.DecodeString(kubeconfigB64)
		if err != nil {
			logger.Error("KubeconfigAuthMiddleware: base64 decode failed",
				"path", requestPath,
				"clientIP", c.ClientIP(),
				"error", err)
			apierr.Abort(c, apierr.Validation("invalid kubeconfig base64"))
			return
		}

		normalized, cfg, err := normalizeKubeconfig(kubeconfig)
		if err != nil {
			logger.Error("KubeconfigAuthMiddleware: normalize failed",
				"path", requestPath,
				"clientIP", c.ClientIP(),
				"error", err)
			apiErr := apierr.Validation("failed to normalize kubeconfig")
			switch {
			case errors.Is(err, fs.ErrPermission):
				apiErr = apierr.Forbidden("permission denied for kubeconfig")
			case errors.Is(err, fs.ErrNotExist):
				apiErr = apierr.Validation("kubeconfig file not found")
			}
			apierr.Abort(c, apiErr)
			return
		}

		// Create Kubernetes clients with normalized kubeconfig
		ctrlClient, clientset, dynClient, err := newAllClientsFromKubeconfig(normalized)
		if err != nil {
			logger.Error("KubeconfigAuthMiddleware: client creation failed",
				"path", requestPath,
				"clientIP", c.ClientIP(),
				"error", err)
			apierr.AbortUnauthorized(c, "failed to create kubernetes clients")
			return
		}

		c.Set("k8sClient", ctrlClient)
		c.Set("clientset", clientset)
		c.Set("dynamic-client", dynClient)

		// Extract identity for audit (best-effort)
		if cfg != nil {
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
			logger.Info("KubeconfigAuthMiddleware: authenticated",
				"context", ctxName,
				"user", user,
				"namespace", c.GetString("k8sDefaultNamespace"),
				"path", requestPath)
		}
		c.Set("normalizedKubeconfig", normalized)
		logger.Info("KubeconfigAuthMiddleware: kubeconfig normalized and clients stored",
			"path", requestPath)
		c.Next()
	}
}

// Connect handler validates the provided kubeconfig from the request header.
func Connect(c *gin.Context) {
	cs, ok := apiutil.ClientsetFromContext(c)
	if !ok {
		logger.Warn("Connect: clientset missing",
			"clientIP", c.ClientIP())
		apierr.AbortUnauthorized(c, "kubeconfig not provided or invalid")
		return
	}

	defNs := ""
	if v, ok := c.Get("k8sDefaultNamespace"); ok {
		if s, ok2 := v.(string); ok2 {
			defNs = s
		}
	}
	user := ""
	if v, ok := c.Get("k8sUser"); ok {
		if s, ok2 := v.(string); ok2 {
			user = s
		}
	}
	ctxName := ""
	if v, ok := c.Get("k8sContext"); ok {
		if s, ok2 := v.(string); ok2 {
			ctxName = s
		}
	}

	logger.Info("Connect: verifying access",
		"context", ctxName,
		"user", user,
		"namespace", defNs,
		"clientIP", c.ClientIP())

	resp := gin.H{
		"message":          "connection successful",
		"user":             user,
		"context":          ctxName,
		"defaultNamespace": defNs,
	}

	// 1) Communicate with apiserver
	logger.Info("Connect: querying apiserver version",
		"context", ctxName,
		"user", user)
	sv, err := cs.Discovery().ServerVersion()
	if err != nil {
		logger.Error("Connect: server version query failed",
			"context", ctxName,
			"user", user,
			"error", err)
		switch {
		case k8serrors.IsUnauthorized(err):
			apierr.AbortUnauthorized(c, "authentication failed")
		case k8serrors.IsForbidden(err):
			apierr.AbortForbidden(c, "permission denied")
		default:
			var netErr net.Error
			if errors.As(err, &netErr) {
				apiErr := apierr.Timeout("apiserver unreachable")
				if !netErr.Timeout() {
					apiErr = apierr.ServiceUnavailable("apiserver unreachable", 0)
				}
				apierr.Abort(c, apiErr)
			} else {
				apierr.Abort(c, apierr.ServiceUnavailable("apiserver unreachable", 0))
			}
		}
		return
	}
	logger.Info("Connect: apiserver version query succeeded",
		"context", ctxName,
		"user", user)

	// 2) Lightweight RBAC validation: list namespaces (limit 1)
	if _, err := cs.CoreV1().Namespaces().List(c.Request.Context(), metav1.ListOptions{Limit: 1}); err != nil {
		logger.Error("Connect: namespace list failed",
			"context", ctxName,
			"user", user,
			"error", err)
		apierr.AbortWithError(c, err)
		return
	}

	if sv != nil {
		resp["apiserverVersion"] = sv.GitVersion
		resp["platform"] = sv.Platform
		logger.Info("Connect: apiserver info",
			"apiserverVersion", sv.GitVersion,
			"platform", sv.Platform,
			"context", ctxName,
			"user", user)
	}

	apierr.OK(c, resp)
	logger.Info("Connect: connection successful",
		"context", ctxName,
		"user", user)
}

// Note: error handling and client getters are centralized in api/util.
