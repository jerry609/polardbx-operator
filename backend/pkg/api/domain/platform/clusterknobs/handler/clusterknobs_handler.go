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
// @Summary List cluster knobs
// @Description Lists all cluster parameter knobs configurations
// @Tags cluster-knobs
// @Accept json
// @Produce json
// @Success 200 {array} polardbxv1.PolarDBXClusterKnobs "List of cluster knobs"
// @Failure 502 {object} map[string]any "Kubernetes API error"
// @Router /api/v1/polardbxclusters/cluster-knobs [get]
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
// @Summary Create cluster knobs
// @Description Creates a new cluster parameter knobs configuration
// @Tags cluster-knobs
// @Accept json
// @Produce json
// @Param body body polardbxv1.PolarDBXClusterKnobs true "Cluster knobs specification"
// @Success 201 {object} polardbxv1.PolarDBXClusterKnobs "Created cluster knobs"
// @Failure 400 {object} map[string]any "Invalid cluster knobs specification"
// @Failure 409 {object} map[string]any "Cluster knobs already exists"
// @Failure 502 {object} map[string]any "Kubernetes API error"
// @Router /api/v1/polardbxclusters/cluster-knobs [post]
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
// @Summary Get cluster knobs
// @Description Retrieves details of a specific cluster parameter knobs configuration
// @Tags cluster-knobs
// @Accept json
// @Produce json
// @Param namespace path string true "Kubernetes namespace"
// @Param name path string true "Name of the cluster knobs"
// @Success 200 {object} polardbxv1.PolarDBXClusterKnobs "Cluster knobs details"
// @Failure 404 {object} map[string]any "Cluster knobs not found"
// @Failure 502 {object} map[string]any "Kubernetes API error"
// @Router /api/v1/polardbxclusters/cluster-knobs/{namespace}/{name} [get]
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
// @Summary Update cluster knobs
// @Description Updates an existing cluster parameter knobs configuration
// @Tags cluster-knobs
// @Accept json
// @Produce json
// @Param namespace path string true "Kubernetes namespace"
// @Param name path string true "Name of the cluster knobs"
// @Param body body polardbxv1.PolarDBXClusterKnobs true "Updated cluster knobs specification"
// @Success 200 {object} polardbxv1.PolarDBXClusterKnobs "Updated cluster knobs"
// @Failure 400 {object} map[string]any "Invalid cluster knobs specification"
// @Failure 404 {object} map[string]any "Cluster knobs not found"
// @Failure 502 {object} map[string]any "Kubernetes API error"
// @Router /api/v1/polardbxclusters/cluster-knobs/{namespace}/{name} [put]
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
// @Summary Delete cluster knobs
// @Description Deletes a cluster parameter knobs configuration
// @Tags cluster-knobs
// @Accept json
// @Produce json
// @Param namespace path string true "Kubernetes namespace"
// @Param name path string true "Name of the cluster knobs"
// @Success 200 {object} map[string]any "Deletion confirmation"
// @Failure 404 {object} map[string]any "Cluster knobs not found"
// @Failure 502 {object} map[string]any "Kubernetes API error"
// @Router /api/v1/polardbxclusters/cluster-knobs/{namespace}/{name} [delete]
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
