package xstores

import (
	"polardbx-ui-backend/pkg/api/domain/xstores/services"

	"github.com/gin-gonic/gin"
)

// ListFollowers lists XStore followers.
// @Summary List XStore followers
// @Description List follower XStores in the specified or default namespace.
// @Tags xstores, followers
// @Produce json
// @Param namespace query string false "Kubernetes namespace; defaults to 'default' when omitted"
// @Success 200 {array} polardbxv1.XStore "List of follower XStores"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func ListFollowers(c *gin.Context) { services.NewFollowersService().List(c) }

// CreateFollower creates a follower XStore.
// @Summary Create XStore follower
// @Description Create a follower XStore that replicates from a primary XStore.
// @Tags xstores, followers
// @Accept json
// @Produce json
// @Param namespace query string false "Kubernetes namespace; defaults to 'default' when omitted"
// @Param body body polardbxv1.XStore true "Follower XStore specification"
// @Success 201 {object} polardbxv1.XStore "Created follower XStore"
// @Failure 400 {object} map[string]any "Invalid request body"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func CreateFollower(c *gin.Context) { services.NewFollowersService().Create(c) }

// GetFollower gets a follower XStore by namespace and name.
// @Summary Get XStore follower
// @Description Get a follower XStore by namespace and name.
// @Tags xstores, followers
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the follower"
// @Param name path string true "Name of the follower XStore"
// @Success 200 {object} polardbxv1.XStore "Follower XStore"
// @Failure 404 {object} map[string]any "Follower not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func GetFollower(c *gin.Context) { services.NewFollowersService().Get(c) }

// UpdateFollower updates a follower XStore.
// @Summary Update XStore follower
// @Description Update an existing follower XStore.
// @Tags xstores, followers
// @Accept json
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the follower"
// @Param name path string true "Name of the follower XStore"
// @Param body body polardbxv1.XStore true "Updated follower specification"
// @Success 200 {object} polardbxv1.XStore "Updated follower XStore"
// @Failure 400 {object} map[string]any "Invalid request body"
// @Failure 404 {object} map[string]any "Follower not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func UpdateFollower(c *gin.Context) { services.NewFollowersService().Update(c) }

// DeleteFollower deletes a follower XStore.
// @Summary Delete XStore follower
// @Description Delete a follower XStore by namespace and name.
// @Tags xstores, followers
// @Param namespace path string true "Kubernetes namespace of the follower"
// @Param name path string true "Name of the follower XStore"
// @Success 200 {object} map[string]any "Deletion confirmation"
// @Failure 404 {object} map[string]any "Follower not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func DeleteFollower(c *gin.Context) { services.NewFollowersService().Delete(c) }

// RebuildLogger triggers a rebuild of the logger role for an XStore.
// @Summary Rebuild logger
// @Description Trigger rebuild of the logger role for the specified XStore.
// @Tags xstores, rebuild
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the XStore"
// @Param name path string true "Name of the XStore"
// @Success 202 {object} map[string]any "Rebuild requested"
// @Failure 404 {object} map[string]any "XStore not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func RebuildLogger(c *gin.Context) { services.NewRebuildService().Logger(c) }

// RebuildLearner triggers a rebuild of the learner role for an XStore.
// @Summary Rebuild learner
// @Description Trigger rebuild of the learner role for the specified XStore.
// @Tags xstores, rebuild
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the XStore"
// @Param name path string true "Name of the XStore"
// @Success 202 {object} map[string]any "Rebuild requested"
// @Failure 404 {object} map[string]any "XStore not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func RebuildLearner(c *gin.Context) { services.NewRebuildService().Learner(c) }

// AutoRebuild triggers an automatic rebuild based on internal heuristics.
// @Summary Auto rebuild XStore
// @Description Trigger automatic rebuild of XStore roles using internal heuristics.
// @Tags xstores, rebuild
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the XStore"
// @Param name path string true "Name of the XStore"
// @Success 202 {object} map[string]any "Rebuild requested"
// @Failure 404 {object} map[string]any "XStore not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func AutoRebuild(c *gin.Context) { services.NewRebuildService().Auto(c) }

// RebuildStatus returns the status of a rebuild operation.
// @Summary Get rebuild status
// @Description Get status of an ongoing rebuild operation for the specified XStore.
// @Tags xstores, rebuild
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the XStore"
// @Param name path string true "Name of the XStore"
// @Success 200 {object} map[string]any "Rebuild status"
// @Failure 404 {object} map[string]any "Rebuild or XStore not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func RebuildStatus(c *gin.Context) { services.NewRebuildService().Status(c) }

// RebuildWait waits until a rebuild operation finishes.
// @Summary Wait rebuild
// @Description Block until a rebuild operation finishes or times out.
// @Tags xstores, rebuild
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the XStore"
// @Param name path string true "Name of the XStore"
// @Success 200 {object} map[string]any "Final rebuild status"
// @Failure 404 {object} map[string]any "Rebuild or XStore not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func RebuildWait(c *gin.Context) { services.NewRebuildService().Wait(c) }

// RebuildProgress returns detailed progress information for a rebuild operation.
// @Summary Get rebuild progress
// @Description Get detailed progress information for an ongoing rebuild.
// @Tags xstores, rebuild
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the XStore"
// @Param name path string true "Name of the XStore"
// @Success 200 {object} map[string]any "Rebuild progress details"
// @Failure 404 {object} map[string]any "Rebuild or XStore not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func RebuildProgress(c *gin.Context) { services.NewRebuildService().Progress(c) }

// RebuildCancel cancels a running rebuild operation.
// @Summary Cancel rebuild
// @Description Cancel an ongoing rebuild operation for the specified XStore.
// @Tags xstores, rebuild
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the XStore"
// @Param name path string true "Name of the XStore"
// @Success 200 {object} map[string]any "Cancellation requested"
// @Failure 404 {object} map[string]any "Rebuild or XStore not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func RebuildCancel(c *gin.Context) { services.NewRebuildService().Cancel(c) }

// RetryFollower retries a failed follower provisioning or sync.
// @Summary Retry follower
// @Description Retry a failed follower XStore operation.
// @Tags xstores, followers
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the follower"
// @Param name path string true "Name of the follower XStore"
// @Success 202 {object} map[string]any "Retry requested"
// @Failure 404 {object} map[string]any "Follower not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func RetryFollower(c *gin.Context) { services.NewFollowersService().Retry(c) }

// CancelFollower cancels a follower provisioning or sync operation.
// @Summary Cancel follower
// @Description Cancel a follower provisioning or sync operation.
// @Tags xstores, followers
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the follower"
// @Param name path string true "Name of the follower XStore"
// @Success 200 {object} map[string]any "Cancellation requested"
// @Failure 404 {object} map[string]any "Follower not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func CancelFollower(c *gin.Context) { services.NewFollowersService().Cancel(c) }
