package services

import (
	"polardbx-ui-backend/pkg/api/domain/xstores/k8srepo"
	apierr "polardbx-ui-backend/pkg/api/errors"
	"polardbx-ui-backend/pkg/api/util"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	polardbxv1xstore "github.com/alibaba/polardbx-operator/api/v1/xstore"
	"github.com/gin-gonic/gin"
)

// FollowersService encapsulates XStoreFollower related orchestration (using k8srepo).
type FollowersService struct {
	repo k8srepo.XStoreRepository
}

func NewFollowersService() *FollowersService {
	return &FollowersService{repo: k8srepo.NewXStoreRepository()}
}

func (s *FollowersService) List(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.DefaultQuery("namespace", util.DefaultNamespace(c, "default"))
	items, err := s.repo.ListFollowers(c.Request.Context(), cli, ns)
	if err != nil {
		util.HandleK8sError(c, "failed to list xstore followers", err)
		return
	}
	apierr.OK(c, items)
}

func (s *FollowersService) Create(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.DefaultQuery("namespace", util.DefaultNamespace(c, "default"))
	// Compatible with two request body formats:
	// 1) Submit complete CR directly (with spec)
	// 2) Simplified body: {"name":"...","xStoreName":"...","role":"(optional)"}
	type createPayload struct {
		polardbxv1.XStoreFollower `json:",inline"`
		Name                      string `json:"name,omitempty"`
		XStoreName                string `json:"xStoreName,omitempty"`
		Role                      string `json:"role,omitempty"`
	}
	var payload createPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		apierr.AbortValidation(c, "invalid xstore follower: "+err.Error())
		return
	}
	obj := payload.XStoreFollower
	if obj.Name == "" {
		obj.Name = payload.Name
	}
	if obj.Name == "" {
		apierr.AbortValidation(c, "name is required")
		return
	}
	if obj.Namespace == "" {
		obj.Namespace = ns
	}
	if obj.Spec.XStoreName == "" {
		obj.Spec.XStoreName = payload.XStoreName
	}
	if obj.Spec.XStoreName == "" {
		apierr.AbortValidation(c, "xStoreName is required")
		return
	}
	if string(obj.Spec.Role) == "" && payload.Role != "" {
		obj.Spec.Role = polardbxv1xstore.FollowerRole(payload.Role)
	}
	created, err := s.repo.CreateFollower(c.Request.Context(), cli, obj.Namespace, &obj)
	if err != nil {
		util.HandleK8sError(c, "failed to create xstore follower", err)
		return
	}
	apierr.Created(c, created)
}

func (s *FollowersService) Get(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	item, err := s.repo.GetFollower(c.Request.Context(), cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "failed to get xstore follower", err)
		return
	}
	apierr.OK(c, item)
}

func (s *FollowersService) Update(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	var body polardbxv1.XStoreFollower
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.AbortValidation(c, "invalid xstore follower: "+err.Error())
		return
	}
	body.Namespace = ns
	updated, err := s.repo.UpdateFollower(c.Request.Context(), cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to update xstore follower", err)
		return
	}
	apierr.OK(c, updated)
}

func (s *FollowersService) Delete(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	if err := s.repo.DeleteFollower(c.Request.Context(), cli, ns, name); err != nil {
		util.HandleK8sError(c, "failed to delete xstore follower", err)
		return
	}
	apierr.OK(c, gin.H{"message": "xstore follower deleted"})
}

// Retry: Retry failed XStoreFollower task (delete old task and create new task)
func (s *FollowersService) Retry(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")

	// Get original task
	original, err := s.repo.GetFollower(c.Request.Context(), cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "failed to get original follower", err)
		return
	}

	// Check if retry is allowed (only failed tasks can be retried)
	if original.Status.Phase != polardbxv1xstore.FollowerPhaseFailed {
		apierr.AbortValidation(c, "only failed tasks can be retried, current_phase: "+string(original.Status.Phase))
		return
	}

	// Delete original task
	if err := s.repo.DeleteFollower(c.Request.Context(), cli, ns, name); err != nil {
		util.HandleK8sError(c, "failed to delete original follower", err)
		return
	}

	// Create new task (keep original configuration)
	newFollower := &polardbxv1.XStoreFollower{
		ObjectMeta: original.ObjectMeta,
		Spec:       original.Spec,
	}
	// Reset status and resource version
	newFollower.Status = polardbxv1.XStoreFollowerStatus{}
	newFollower.ResourceVersion = ""
	newFollower.Generation = 0

	created, err := s.repo.CreateFollower(c.Request.Context(), cli, ns, newFollower)
	if err != nil {
		util.HandleK8sError(c, "failed to create retry follower", err)
		return
	}

	apierr.OK(c, gin.H{
		"message":       "task retried successfully",
		"original_task": name,
		"new_task":      created.Name,
	})
}

// Cancel: Cancel specified XStoreFollower task
func (s *FollowersService) Cancel(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")

	// Get task information
	follower, err := s.repo.GetFollower(c.Request.Context(), cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "failed to get follower", err)
		return
	}

	// Check if cancellation is allowed (terminal state tasks cannot be cancelled)
	if follower.Status.Phase == polardbxv1xstore.FollowerPhaseSuccess ||
		follower.Status.Phase == polardbxv1xstore.FollowerPhaseFailed ||
		follower.Status.Phase == polardbxv1xstore.FollowerPhaseDeleting {
		apierr.AbortValidation(c, "cannot cancel completed or already deleting task, current_phase: "+string(follower.Status.Phase))
		return
	}

	// Delete task
	if err := s.repo.DeleteFollower(c.Request.Context(), cli, ns, name); err != nil {
		util.HandleK8sError(c, "failed to cancel follower", err)
		return
	}

	apierr.OK(c, gin.H{
		"message": "task cancelled successfully",
		"task":    name,
		"xstore":  follower.Spec.XStoreName,
	})
}
