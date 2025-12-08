package polardbxclusters

import (
	"polardbx-ui-backend/pkg/api/domain/polardbxclusters/services"
	"polardbx-ui-backend/pkg/api/provider"

	"github.com/gin-gonic/gin"
)

// --- Thin handlers forwarding to services ---

func svc(c *gin.Context) *services.ClusterService { return provider.Must(c).ClusterService(c) }

func List(c *gin.Context)             { svc(c).List(c) }
func Create(c *gin.Context)           { svc(c).Create(c) }
func CreateFromConfig(c *gin.Context) { svc(c).CreateFromConfig(c) }
func Get(c *gin.Context)              { svc(c).Get(c) }
func Update(c *gin.Context)           { svc(c).Update(c) }
func Delete(c *gin.Context)           { svc(c).Delete(c) }
