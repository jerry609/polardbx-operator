// Package handler provides HTTP handlers for cluster parameter knobs

// Follows Clean Architecture design pattern
package handler

import (
	apierr "polardbx-ui-backend/pkg/api/errors"
	"polardbx-ui-backend/pkg/api/util"
	"polardbx-ui-backend/pkg/k8s"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"github.com/gin-gonic/gin"
)

// GetList gets all cluster parameter knobs list
func GetList(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	list, err := k8s.GetClusterKnobsList(cli)
	if err != nil {
		util.HandleK8sError(c, "failed to get cluster knobs list", err)
		return
	}
	apierr.OK(c, list)
}

// Create creates cluster parameter knobs
func Create(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	var payload polardbxv1.PolarDBXClusterKnobs
	if err := c.ShouldBindJSON(&payload); err != nil {
		apierr.AbortValidation(c, "invalid cluster knobs data: "+err.Error())
		return
	}
	created, err := k8s.CreateClusterKnobs(cli, &payload)
	if err != nil {
		util.HandleK8sError(c, "failed to create cluster knobs", err)
		return
	}
	apierr.Created(c, created)
}

// Get gets specified cluster parameter knobs
func Get(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	namespace := c.Param("namespace")
	name := c.Param("name")
	knobs, err := k8s.GetClusterKnobs(cli, namespace, name)
	if err != nil {
		util.HandleK8sError(c, "failed to get cluster knobs", err)
		return
	}
	apierr.OK(c, knobs)
}

// Update updates cluster parameter knobs
func Update(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	namespace := c.Param("namespace")
	name := c.Param("name")
	var payload polardbxv1.PolarDBXClusterKnobs
	if err := c.ShouldBindJSON(&payload); err != nil {
		apierr.AbortValidation(c, "invalid cluster knobs data: "+err.Error())
		return
	}
	payload.Namespace = namespace
	payload.Name = name
	updated, err := k8s.UpdateClusterKnobs(cli, &payload)
	if err != nil {
		util.HandleK8sError(c, "failed to update cluster knobs", err)
		return
	}
	apierr.OK(c, updated)
}

// Delete deletes cluster parameter knobs
func Delete(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	namespace := c.Param("namespace")
	name := c.Param("name")
	if err := k8s.DeleteClusterKnobs(cli, namespace, name); err != nil {
		util.HandleK8sError(c, "failed to delete cluster knobs", err)
		return
	}
	apierr.OK(c, gin.H{"message": "cluster knobs deleted successfully"})
}
