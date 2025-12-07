package services

import (
	"polardbx-ui-backend/pkg/api/domain/xstores/k8srepo"
	apierr "polardbx-ui-backend/pkg/api/errors"
	"polardbx-ui-backend/pkg/api/util"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"github.com/gin-gonic/gin"
)

// XStoreService encapsulates XStore basic CRUD and Pod listing
type XStoreService struct {
	repo k8srepo.XStoreRepository
}

func NewXStoreService() *XStoreService { return &XStoreService{repo: k8srepo.NewXStoreRepository()} }

func (s *XStoreService) List(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	items, err := s.repo.List(c.Request.Context(), cli, ns)
	if err != nil {
		util.HandleK8sError(c, "failed to list xstores", err)
		return
	}
	apierr.OK(c, items)
}

func (s *XStoreService) Create(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	var body polardbxv1.XStore
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.AbortValidation(c, "invalid xstore: "+err.Error())
		return
	}
	created, err := s.repo.Create(c.Request.Context(), cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to create xstore", err)
		return
	}
	apierr.Created(c, created)
}

func (s *XStoreService) Get(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	item, err := s.repo.Get(c.Request.Context(), cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "failed to get xstore", err)
		return
	}
	apierr.OK(c, item)
}

func (s *XStoreService) Update(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	var body polardbxv1.XStore
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.AbortValidation(c, "invalid xstore: "+err.Error())
		return
	}
	body.Namespace = ns
	updated, err := s.repo.Update(c.Request.Context(), cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to update xstore", err)
		return
	}
	apierr.OK(c, updated)
}

func (s *XStoreService) Delete(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	if err := s.repo.Delete(c.Request.Context(), cli, ns, name); err != nil {
		util.HandleK8sError(c, "failed to delete xstore", err)
		return
	}
	apierr.OK(c, gin.H{"message": "xstore deleted"})
}

func (s *XStoreService) ListPods(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	items, err := s.repo.ListPods(c.Request.Context(), cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "failed to list pods for xstore", err)
		return
	}
	apierr.OK(c, items)
}
