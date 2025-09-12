package xstore

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"polardbx-ui-backend/pkg/api/util"
	"polardbx-ui-backend/pkg/k8s"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	polardbxv1xstore "github.com/alibaba/polardbx-operator/api/v1/xstore"
)

// ----- XStore CRUD & Pods -----
func List(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	items, err := k8s.ListXStores(cli, ns)
	if err != nil {
		util.HandleK8sError(c, "failed to list xstores", err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func Create(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	var body polardbxv1.XStore
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid xstore", "details": err.Error()})
		return
	}
	created, err := k8s.CreateXStore(cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to create xstore", err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func Get(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	item, err := k8s.GetXStore(cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "failed to get xstore", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func Update(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	var body polardbxv1.XStore
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid xstore", "details": err.Error()})
		return
	}
	body.Namespace = ns
	updated, err := k8s.UpdateXStore(cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to update xstore", err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func Delete(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	if err := k8s.DeleteXStore(cli, ns, name); err != nil {
		util.HandleK8sError(c, "failed to delete xstore", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "xstore deleted"})
}

func ListPods(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	var podList corev1.PodList
	if err := cli.List(c.Request.Context(), &podList, client.InNamespace(ns), client.MatchingLabels{"xstore/name": name}); err != nil {
		util.HandleK8sError(c, "failed to list pods for xstore", err)
		return
	}
	c.JSON(http.StatusOK, podList.Items)
}

// ----- XStoreBackup APIs -----
func ListBackups(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	view := c.DefaultQuery("view", "detail")
	items, err := k8s.ListXStoreBackups(cli, ns)
	if err != nil {
		util.HandleK8sError(c, "failed to list xstore backups", err)
		return
	}
	if view == "summary" {
		c.JSON(http.StatusOK, items)
		return
	}
	c.JSON(http.StatusOK, items)
}

func CreateBackup(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	var body polardbxv1.XStoreBackup
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid xstore backup", "details": err.Error()})
		return
	}
	created, err := k8s.CreateXStoreBackup(cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to create xstore backup", err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func GetBackup(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	item, err := k8s.GetXStoreBackup(cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "failed to get xstore backup", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func UpdateBackup(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	var body polardbxv1.XStoreBackup
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid xstore backup", "details": err.Error()})
		return
	}
	body.Namespace = ns
	updated, err := k8s.UpdateXStoreBackup(cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to update xstore backup", err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteBackup(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	if err := k8s.DeleteXStoreBackup(cli, ns, name); err != nil {
		util.HandleK8sError(c, "failed to delete xstore backup", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "xstore backup deleted"})
}

func ForceDeleteBackup(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	var xs polardbxv1.XStoreBackup
	if err := cli.Get(c.Request.Context(), client.ObjectKey{Namespace: ns, Name: name}, &xs); err != nil {
		util.HandleK8sError(c, "failed to get xstore backup", err)
		return
	}
	xs.SetFinalizers([]string{})
	if err := cli.Update(c.Request.Context(), &xs); err != nil {
		util.HandleK8sError(c, "failed to remove finalizers", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "xstore backup finalizers removed"})
}

func GetBackupRemoteInfo(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	bk, err := k8s.GetXStoreBackup(cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "failed to get xstore backup", err)
		return
	}
	resp := gin.H{
		"namespace": ns,
		"name":      name,
		"phase":     bk.Status.Phase,
	}
	c.JSON(http.StatusOK, resp)
}

// ----- Rebuild -----
func RebuildLogger(c *gin.Context) {
	c.Set("rebuildRole", polardbxv1xstore.FollowerRoleLogger)
	createFollower(c)
}
func RebuildLearner(c *gin.Context) {
	c.Set("rebuildRole", polardbxv1xstore.FollowerRoleLearner)
	createFollower(c)
}
func AutoRebuild(c *gin.Context) { createFollower(c) }

// RebuildStatus：保持可用，查询进行中的 follower 任务
func RebuildStatus(c *gin.Context) {
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
	c.JSON(http.StatusOK, gin.H{"namespace": ns, "xstore": xname, "active": items})
}

// ----- XStoreFollower CRUD (generic management) -----
func ListFollowers(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.DefaultQuery("namespace", util.DefaultNamespace(c, "default"))
	var list polardbxv1.XStoreFollowerList
	if err := cli.List(c.Request.Context(), &list, client.InNamespace(ns)); err != nil {
		util.HandleK8sError(c, "failed to list xstore followers", err)
		return
	}
	c.JSON(http.StatusOK, list.Items)
}

func CreateFollower(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.DefaultQuery("namespace", util.DefaultNamespace(c, "default"))
	// Accept both CR-style and simplified payload {"name": "...", "xStoreName": "..."}
	var raw map[string]any
	if err := c.ShouldBindJSON(&raw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid xstore follower", "details": err.Error()})
		return
	}

	var body polardbxv1.XStoreFollower
	// metadata fields
	if md, ok := raw["metadata"].(map[string]any); ok {
		if v, ok := md["name"].(string); ok {
			body.Name = v
		}
		if v, ok := md["namespace"].(string); ok {
			body.Namespace = v
		}
	}
	if body.Name == "" {
		if v, ok := raw["name"].(string); ok {
			body.Name = v
		}
	}

	// spec fields
	if spec, ok := raw["spec"].(map[string]any); ok {
		if v, ok := spec["xStoreName"].(string); ok {
			body.Spec.XStoreName = v
		}
		if v, ok := spec["fromPodName"].(string); ok {
			body.Spec.FromPodName = v
		}
		if v, ok := spec["targetPodName"].(string); ok {
			body.Spec.TargetPodName = v
		}
		if v, ok := spec["nodeName"].(string); ok {
			body.Spec.NodeName = v
		}
		if v, ok := spec["local"].(bool); ok {
			body.Spec.Local = v
		}
		if v, ok := spec["role"].(string); ok {
			body.Spec.Role = polardbxv1xstore.FollowerRole(v)
		}
	} else {
		if v, ok := raw["xStoreName"].(string); ok {
			body.Spec.XStoreName = v
		}
	}

	if body.Namespace == "" {
		body.Namespace = ns
	}
	if body.Name == "" || body.Spec.XStoreName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and xStoreName are required"})
		return
	}

	if err := cli.Create(c.Request.Context(), &body); err != nil {
		util.HandleK8sError(c, "failed to create xstore follower", err)
		return
	}
	c.JSON(http.StatusCreated, &body)
}

func GetFollower(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	var f polardbxv1.XStoreFollower
	if err := cli.Get(c.Request.Context(), client.ObjectKey{Namespace: ns, Name: name}, &f); err != nil {
		util.HandleK8sError(c, "failed to get xstore follower", err)
		return
	}
	c.JSON(http.StatusOK, &f)
}

func UpdateFollower(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	var f polardbxv1.XStoreFollower
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid xstore follower", "details": err.Error()})
		return
	}
	f.Namespace = ns
	if err := cli.Update(c.Request.Context(), &f); err != nil {
		util.HandleK8sError(c, "failed to update xstore follower", err)
		return
	}
	c.JSON(http.StatusOK, &f)
}

func DeleteFollower(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	var f polardbxv1.XStoreFollower
	f.Namespace = ns
	f.Name = name
	if err := cli.Delete(c.Request.Context(), &f); err != nil {
		util.HandleK8sError(c, "failed to delete xstore follower", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "xstore follower deleted"})
}

// createFollower implements the core rebuild follower creation logic
func createFollower(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	namespace := c.Param("namespace")
	if namespace == "" {
		namespace = util.DefaultNamespace(c, "default")
	}

	var request struct {
		Name          string            `json:"name"`
		XStoreName    string            `json:"xStoreName"`
		FromXStore    string            `json:"fromXStore,omitempty"`
		TargetPodName string            `json:"targetPodName,omitempty"`
		NodeSelector  map[string]string `json:"nodeSelector,omitempty"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse xstore follower request", "details": err.Error()})
		return
	}

	if xstoreFromPath := c.Param("name"); xstoreFromPath != "" {
		if request.XStoreName == "" {
			request.XStoreName = xstoreFromPath
		} else if request.XStoreName != xstoreFromPath {
			c.JSON(http.StatusBadRequest, gin.H{"error": "xStoreName mismatch with path", "details": fmt.Sprintf("body.xStoreName=%s, path.name=%s", request.XStoreName, xstoreFromPath)})
			return
		}
	}

	if request.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if request.XStoreName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "xStoreName is required"})
		return
	}
	if errs := validation.IsDNS1123Label(request.Name); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid name", "details": strings.Join(errs, "; ")})
		return
	}
	if errs := validation.IsDNS1123Label(request.XStoreName); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid xStoreName", "details": strings.Join(errs, "; ")})
		return
	}

	// Ensure target XStore exists
	if _, err := k8s.GetXStore(cli, namespace, request.XStoreName); err != nil {
		if k8serrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "xstore not found", "details": fmt.Sprintf("xstore %s/%s not found", namespace, request.XStoreName)})
			return
		}
		util.HandleK8sError(c, "failed to get xstore", err)
		return
	}

	// Concurrency: reject if any non-end-phase follower exists for this XStore
	var followerList polardbxv1.XStoreFollowerList
	if err := cli.List(c.Request.Context(), &followerList, client.InNamespace(namespace)); err == nil {
		for _, f := range followerList.Items {
			if f.Spec.XStoreName == request.XStoreName {
				if !polardbxv1xstore.IsEndPhase(f.Status.Phase) {
					c.JSON(http.StatusConflict, gin.H{"error": "rebuild in progress", "details": fmt.Sprintf("active follower '%s' for xstore '%s' phase '%s'", f.Name, request.XStoreName, string(f.Status.Phase))})
					return
				}
			}
		}
	}

	// Resolve target pod
	targetPodName := strings.TrimSpace(request.TargetPodName)
	var targetPodList corev1.PodList
	if err := cli.List(c.Request.Context(), &targetPodList, client.InNamespace(namespace), client.MatchingLabels{"xstore/name": request.XStoreName}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to list target pods", "details": fmt.Sprintf("list pods for target XStore '%s': %v", request.XStoreName, err)})
		return
	}
	if len(targetPodList.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target pods not found", "details": fmt.Sprintf("no pods for xstore '%s'", request.XStoreName)})
		return
	}
	if targetPodName != "" {
		var found *corev1.Pod
		for i := range targetPodList.Items {
			if targetPodList.Items[i].Name == targetPodName {
				found = &targetPodList.Items[i]
				break
			}
		}
		if found == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid targetPodName", "details": fmt.Sprintf("pod '%s' not found", targetPodName)})
			return
		}
		if found.Status.Phase != corev1.PodRunning {
			c.JSON(http.StatusBadRequest, gin.H{"error": "target pod not running", "details": fmt.Sprintf("pod '%s' phase=%s", targetPodName, found.Status.Phase)})
			return
		}
		if role := found.Labels["xstore/role"]; strings.EqualFold(role, "leader") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target role", "details": "target must not be leader"})
			return
		}
	} else {
		for _, pod := range targetPodList.Items {
			if pod.Status.Phase == corev1.PodRunning && !strings.EqualFold(pod.Labels["xstore/role"], "leader") {
				targetPodName = pod.Name
				break
			}
		}
		if targetPodName == "" {
			for _, pod := range targetPodList.Items {
				if pod.Status.Phase == corev1.PodRunning {
					targetPodName = pod.Name
					break
				}
			}
		}
		if targetPodName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no running target pods"})
			return
		}
	}

	// NodeName from first NodeSelector value
	targetNodeName := ""
	if request.NodeSelector != nil {
		for _, v := range request.NodeSelector {
			targetNodeName = v
			break
		}
	}

	// Determine role
	role := polardbxv1xstore.FollowerRoleFollower
	if override, exists := c.Get("rebuildRole"); exists {
		if v, ok := override.(polardbxv1xstore.FollowerRole); ok {
			role = v
		} else if s, ok := override.(string); ok {
			switch strings.ToLower(s) {
			case "logger":
				role = polardbxv1xstore.FollowerRoleLogger
			case "learner":
				role = polardbxv1xstore.FollowerRoleLearner
			}
		}
	}

	follower := &polardbxv1.XStoreFollower{
		ObjectMeta: metav1.ObjectMeta{Name: request.Name, Namespace: namespace, Labels: map[string]string{"xstore/target": request.XStoreName, "xstore/rebuild-type": string(role)}},
		Spec:       polardbxv1.XStoreFollowerSpec{XStoreName: request.XStoreName, FromPodName: "", TargetPodName: targetPodName, NodeName: targetNodeName, Role: role, Local: false},
	}

	if err := cli.Create(c.Request.Context(), follower); err != nil {
		util.HandleK8sError(c, "failed to create xstore follower", err)
		return
	}
	c.JSON(http.StatusCreated, follower)
}
