package provider

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	podrepo "polardbx-ui-backend/pkg/api/domain/platform/pod/repository"
	podsvc "polardbx-ui-backend/pkg/api/domain/platform/pod/service"
	"polardbx-ui-backend/pkg/api/domain/polardbxclusters/services"
	"polardbx-ui-backend/pkg/api/util"
)

const providerKey = "serviceProvider"

// Provider defines how handlers acquire services/repositories.
// Kept minimal to stay non-intrusive for existing handlers.
type Provider interface {
	ClusterService(*gin.Context) *services.ClusterService
	PodService(*gin.Context) (*podsvc.PodService, bool)
}

type defaultProvider struct{}

// NewDefaultProvider constructs the default provider using in-process factories.
func NewDefaultProvider() Provider {
	return &defaultProvider{}
}

// Inject attaches provider to gin context.
func Inject(p Provider) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(providerKey, p)
		c.Next()
	}
}

// FromContext fetches provider from gin context.
func FromContext(c *gin.Context) (Provider, bool) {
	v, ok := c.Get(providerKey)
	if !ok {
		return nil, false
	}
	p, ok := v.(Provider)
	return p, ok
}

// Must returns provider or panics (programming error).
func Must(c *gin.Context) Provider {
	p, ok := FromContext(c)
	if !ok || p == nil {
		if !allowProviderFallback() {
			panic("service provider not injected; ensure provider.Inject middleware is registered or enable PROVIDER_FALLBACK_ENABLED for non-production paths")
		}
		p = NewDefaultProvider()
		c.Set(providerKey, p)
	}
	return p
}

func (p *defaultProvider) ClusterService(_ *gin.Context) *services.ClusterService {
	return services.NewClusterService()
}

func (p *defaultProvider) PodService(c *gin.Context) (*podsvc.PodService, bool) {
	cli, cs, _, ok := util.GetK8sClients(c)
	if !ok {
		return nil, false
	}
	repo := podrepo.NewK8sPodRepositorySimple(cli, cs)
	return podsvc.NewPodService(repo), true
}

func allowProviderFallback() bool {
	if gin.Mode() == gin.TestMode {
		return true
	}
	enabled := strings.TrimSpace(strings.ToLower(os.Getenv("PROVIDER_FALLBACK_ENABLED")))
	return enabled == "1" || enabled == "true" || enabled == "yes" || enabled == "on"
}
