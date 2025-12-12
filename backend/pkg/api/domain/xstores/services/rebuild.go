package services

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	polardbxv1xstore "github.com/alibaba/polardbx-operator/api/v1/xstore"

	"polardbx-ui-backend/pkg/api/domain/xstores/k8srepo"
	apierr "polardbx-ui-backend/pkg/api/errors"
	"polardbx-ui-backend/pkg/api/util"
	"polardbx-ui-backend/pkg/logger"
)

// RebuildService: Migrate Status first, other entries to be orchestrated later
type RebuildService struct {
	repo k8srepo.XStoreRepository
}

func NewRebuildService() *RebuildService { return &RebuildService{repo: k8srepo.NewXStoreRepository()} }

// Compatible with test requirements: Create XStoreFollower CR based on path parameters and simplified body
// Request body example: {"name":"rebuild-logger-x1","xStoreName":"xstore1"}
func (s *RebuildService) Logger(c *gin.Context) {
	createFollowerWithRole(c, polardbxv1xstore.FollowerRole("logger"))
}
func (s *RebuildService) Learner(c *gin.Context) {
	createFollowerWithRole(c, polardbxv1xstore.FollowerRole("learner"))
}
func (s *RebuildService) Auto(c *gin.Context) {
	createFollowerWithRole(c, polardbxv1xstore.FollowerRole("follower"))
}

func (s *RebuildService) Status(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	xname := c.Param("name")
	var list polardbxv1.XStoreFollowerList
	if err := cli.List(c.Request.Context(), &list, client.InNamespace(ns)); err != nil {
		util.HandleK8sError(c, "failed to list followers", err)
		return
	}
	items := make([]map[string]any, 0)
	for _, f := range list.Items {
		if f.Spec.XStoreName == xname && !polardbxv1xstore.IsEndPhase(f.Status.Phase) {
			items = append(items, map[string]any{"name": f.Name, "phase": string(f.Status.Phase), "message": f.Status.Message, "targetPod": f.Status.TargetPodName})
		}
	}
	apierr.OK(c, gin.H{"namespace": ns, "xstore": xname, "active": items})
}

// Wait polls and waits for specified follower to enter terminal state (success/failed/deleting), supports timeout and interval.
func (s *RebuildService) Wait(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Query("follower")
	if name == "" {
		apierr.AbortValidation(c, "follower is required")
		return
	}
	timeoutSec, _ := strconv.Atoi(c.DefaultQuery("timeoutSec", "300"))
	intervalSec, _ := strconv.Atoi(c.DefaultQuery("intervalSec", "3"))
	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)
	for {
		var f polardbxv1.XStoreFollower
		if err := cli.Get(c.Request.Context(), client.ObjectKey{Namespace: ns, Name: name}, &f); err != nil {
			util.HandleK8sError(c, "failed to get follower", err)
			return
		}
		if polardbxv1xstore.IsEndPhase(f.Status.Phase) {
			apierr.OK(c, gin.H{"name": f.Name, "phase": string(f.Status.Phase), "message": f.Status.Message})
			return
		}
		if time.Now().After(deadline) {
			apierr.Abort(c, apierr.New(apierr.ErrTimeout, "timeout waiting for follower "+f.Name+" (phase: "+string(f.Status.Phase)+")"))
			return
		}
		logger.Info("rebuild wait",
			"namespace", ns,
			"follower", name,
			"phase", f.Status.Phase)
		time.Sleep(time.Duration(intervalSec) * time.Second)
	}
}

// Progress: Simply returns current follower's status and message for external polling.
func (s *RebuildService) Progress(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Query("follower")
	if name == "" {
		apierr.AbortValidation(c, "follower is required")
		return
	}
	var f polardbxv1.XStoreFollower
	if err := cli.Get(c.Request.Context(), client.ObjectKey{Namespace: ns, Name: name}, &f); err != nil {
		util.HandleK8sError(c, "failed to get follower", err)
		return
	}
	apierr.OK(c, gin.H{"name": f.Name, "phase": string(f.Status.Phase), "message": f.Status.Message, "targetPod": f.Status.TargetPodName})
}

// Cancel: Cancel (delete) XStoreFollower task
func (s *RebuildService) Cancel(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	xstoreName := c.Param("name")

	// Find all Follower tasks related to this XStore
	followers, err := s.repo.ListFollowers(c.Request.Context(), cli, ns)
	if err != nil {
		util.HandleK8sError(c, "failed to list followers", err)
		return
	}

	var targetFollower *polardbxv1.XStoreFollower
	for _, f := range followers {
		if f.Spec.XStoreName == xstoreName {
			// Find non-terminal state task
			if f.Status.Phase != polardbxv1xstore.FollowerPhaseSuccess &&
				f.Status.Phase != polardbxv1xstore.FollowerPhaseFailed &&
				f.Status.Phase != polardbxv1xstore.FollowerPhaseDeleting {
				targetFollower = &f
				break
			}
		}
	}

	if targetFollower == nil {
		apierr.AbortNotFound(c, "rebuild task", xstoreName)
		return
	}

	// Delete XStoreFollower task
	if err := s.repo.DeleteFollower(c.Request.Context(), cli, ns, targetFollower.Name); err != nil {
		util.HandleK8sError(c, "failed to cancel rebuild task", err)
		return
	}

	apierr.OK(c, gin.H{
		"message": "rebuild task cancelled",
		"task":    targetFollower.Name,
		"xstore":  xstoreName,
	})
}

// Internal helper method: Create XStoreFollower based on role
func createFollowerWithRole(c *gin.Context, role polardbxv1xstore.FollowerRole) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	if ns == "" {
		ns = c.DefaultQuery("namespace", util.DefaultNamespace(c, "default"))
	}
	xstoreName := c.Param("name")
	type payload struct {
		Name       string `json:"name"`
		XStoreName string `json:"xStoreName"`
	}
	var body payload
	_ = c.ShouldBindJSON(&body)
	name := body.Name
	if name == "" {
		logger.Warn("rebuild create missing name",
			"namespace", ns,
			"xstore", xstoreName)
		util.NotFound(c)
		return
	}
	if xstoreName == "" {
		xstoreName = body.XStoreName
	}
	if xstoreName == "" {
		logger.Warn("rebuild create missing xStoreName",
			"namespace", ns,
			"name", name)
		apierr.AbortValidation(c, "xStoreName is required")
		return
	}
	obj := &polardbxv1.XStoreFollower{}
	obj.Namespace = ns
	obj.Name = name
	obj.Spec.XStoreName = xstoreName
	obj.Spec.Role = role
	obj.Spec.Local = false
	// Auto-select target Pod (prefer follower role and Running pods)
	if obj.Spec.TargetPodName == "" || obj.Spec.FromPodName == "" {
		var pods corev1.PodList
		if err := cli.List(c.Request.Context(), &pods, client.InNamespace(ns)); err == nil {
			var candidate string
			for _, p := range pods.Items {
				if p.Labels["xstore/name"] != xstoreName {
					continue
				}
				if p.Status.Phase != corev1.PodRunning {
					continue
				}
				if p.Labels["xstore/role"] == "follower" {
					candidate = p.Name
					break
				}
				if candidate == "" {
					candidate = p.Name
				}
			}
			if candidate != "" {
				if obj.Spec.TargetPodName == "" {
					obj.Spec.TargetPodName = candidate
				}
				if obj.Spec.FromPodName == "" {
					obj.Spec.FromPodName = candidate
				}
			}
		}
	}
	if obj.Labels == nil {
		obj.Labels = map[string]string{}
	}
	obj.Labels["xstore/rebuild-type"] = string(role)
	if err := cli.Create(c.Request.Context(), obj); err != nil {
		logger.Error("rebuild create failed",
			"namespace", ns,
			"name", name,
			"xstore", xstoreName,
			"role", role,
			"error", err)
		util.HandleK8sError(c, "failed to create xstore follower", err)
		return
	}
	logger.Info("rebuild create ok",
		"namespace", ns,
		"name", name,
		"xstore", xstoreName,
		"role", role,
		"target", obj.Spec.TargetPodName)
	apierr.Created(c, obj)
}
