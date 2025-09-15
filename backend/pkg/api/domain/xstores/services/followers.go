package services

import (
	"net/http"

	"polardbx-ui-backend/pkg/api/domain/xstores/k8srepo"
	"polardbx-ui-backend/pkg/api/util"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	polardbxv1xstore "github.com/alibaba/polardbx-operator/api/v1/xstore"
	"github.com/gin-gonic/gin"
)

// FollowersService 封装 XStoreFollower 相关编排（使用 k8srepo）。
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
	c.JSON(http.StatusOK, items)
}

func (s *FollowersService) Create(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.DefaultQuery("namespace", util.DefaultNamespace(c, "default"))
	// 兼容两种请求体：
	// 1) 直接提交完整 CR（含 spec）
	// 2) 简化体：{"name":"...","xStoreName":"...","role":"(可选)"}
	type createPayload struct {
		polardbxv1.XStoreFollower `json:",inline"`
		Name                      string `json:"name,omitempty"`
		XStoreName                string `json:"xStoreName,omitempty"`
		Role                      string `json:"role,omitempty"`
	}
	var payload createPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid xstore follower", "details": err.Error()})
		return
	}
	obj := payload.XStoreFollower
	if obj.Name == "" {
		obj.Name = payload.Name
	}
	if obj.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if obj.Namespace == "" {
		obj.Namespace = ns
	}
	if obj.Spec.XStoreName == "" {
		obj.Spec.XStoreName = payload.XStoreName
	}
	if obj.Spec.XStoreName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "xStoreName is required"})
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
	c.JSON(http.StatusCreated, created)
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
	c.JSON(http.StatusOK, item)
}

func (s *FollowersService) Update(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	var body polardbxv1.XStoreFollower
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid xstore follower", "details": err.Error()})
		return
	}
	body.Namespace = ns
	updated, err := s.repo.UpdateFollower(c.Request.Context(), cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to update xstore follower", err)
		return
	}
	c.JSON(http.StatusOK, updated)
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
	c.JSON(http.StatusOK, gin.H{"message": "xstore follower deleted"})
}
