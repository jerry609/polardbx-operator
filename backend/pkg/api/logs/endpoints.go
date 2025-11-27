package logs

import (
	"polardbx-ui-backend/pkg/api/domain/platform/logs/handler"

	"github.com/gin-gonic/gin"
)

// Query delegates to domain handler
var Query = handler.Query

// Re-export types for backward compatibility
type QueryRequest = handler.QueryRequest
type FacetSpec = handler.FacetSpec
type HistogramSpec = handler.HistogramSpec
type FacetBucket = handler.FacetBucket
type HistogramBucket = handler.HistogramBucket
type NormalizedResponse = handler.NormalizedResponse

// QueryHandler is an alias for Query (backward compatibility)
func QueryHandler(c *gin.Context) {
	Query(c)
}
