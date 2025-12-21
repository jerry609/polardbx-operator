package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"polardbx-dashboard-backend/pkg/cache"
)

type K8sClients struct {
	Client    client.Client
	Clientset kubernetes.Interface
	Dynamic   dynamic.Interface
}

type NewAllClientsFunc func(kubeconfig []byte) (client.Client, kubernetes.Interface, dynamic.Interface, error)

const (
	kubeconfigClientCacheKeyPrefix = "auth:kubeconfig:clients:"
	defaultKubeconfigClientTTL     = 5 * time.Minute
	kubeconfigClientTTLEnv         = "KUBECONFIG_CLIENT_CACHE_TTL"
)

func kubeconfigClientCacheTTL() time.Duration {
	raw := strings.TrimSpace(os.Getenv(kubeconfigClientTTLEnv))
	if raw == "" {
		return defaultKubeconfigClientTTL
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return defaultKubeconfigClientTTL
	}
	return d
}

func kubeconfigClientCacheKey(normalizedKubeconfig []byte) string {
	sum := sha256.Sum256(normalizedKubeconfig)
	return kubeconfigClientCacheKeyPrefix + hex.EncodeToString(sum[:])
}

// GetOrCreateK8sClientsFromKubeconfig reuses a cached client bundle for the same kubeconfig (by hash),
// to avoid rebuilding REST configs and clients on every request.
func GetOrCreateK8sClientsFromKubeconfig(normalizedKubeconfig []byte, newClients NewAllClientsFunc) (*K8sClients, bool, error) {
	if len(normalizedKubeconfig) == 0 {
		return nil, false, fmt.Errorf("normalized kubeconfig is empty")
	}
	if newClients == nil {
		return nil, false, fmt.Errorf("newClients func is nil")
	}

	key := kubeconfigClientCacheKey(normalizedKubeconfig)
	c := cache.GetGlobalCache()

	if v, ok := c.Get(key); ok {
		if clients, ok := v.(*K8sClients); ok && clients != nil {
			return clients, true, nil
		}
		c.Delete(key)
	}

	ttl := kubeconfigClientCacheTTL()
	v, err := c.GetOrSetWithExpiration(key, ttl, func() (interface{}, error) {
		cli, cs, dyn, err := newClients(normalizedKubeconfig)
		if err != nil {
			return nil, err
		}
		return &K8sClients{
			Client:    cli,
			Clientset: cs,
			Dynamic:   dyn,
		}, nil
	})
	if err != nil {
		return nil, false, err
	}
	clients, ok := v.(*K8sClients)
	if !ok || clients == nil {
		return nil, false, fmt.Errorf("cached value has unexpected type: %T", v)
	}
	return clients, false, nil
}

