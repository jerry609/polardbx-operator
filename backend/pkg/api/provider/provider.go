package provider

import (
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
		// Fallback to default provider for legacy code paths (e.g., tests without middleware).
		p = NewDefaultProvider()
		// cache it to avoid repeated creation
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
