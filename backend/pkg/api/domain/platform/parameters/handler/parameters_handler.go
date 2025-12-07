// Package handler provides HTTP handlers for parameters and parameter templates

// Follows Clean Architecture design pattern
package handler

import (
	apierr "polardbx-ui-backend/pkg/api/errors"
	"polardbx-ui-backend/pkg/api/util"
	"polardbx-ui-backend/pkg/k8s"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"github.com/gin-gonic/gin"
)

// ======================== Parameters ========================

// List gets parameter list
func List(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	items, err := k8s.ListPolarDBXParametersWithContext(c.Request.Context(), cli, ns)
	if err != nil {
		util.HandleK8sError(c, "failed to list parameters", err)
		return
	}
	apierr.OK(c, items)
}

// Get gets specified parameter
func Get(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	name := c.Param("name")
	item, err := k8s.GetPolarDBXParameterWithContext(c.Request.Context(), cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "failed to get parameter", err)
		return
	}
	apierr.OK(c, item)
}

// Create creates parameter
func Create(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	var body polardbxv1.PolarDBXParameter
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.AbortValidation(c, "invalid parameter: "+err.Error())
		return
	}
	created, err := k8s.CreatePolarDBXParameterWithContext(c.Request.Context(), cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to create parameter", err)
		return
	}
	apierr.Created(c, created)
}

// Update updates parameter
func Update(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	var body polardbxv1.PolarDBXParameter
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.AbortValidation(c, "invalid parameter: "+err.Error())
		return
	}
	updated, err := k8s.UpdatePolarDBXParameterWithContext(c.Request.Context(), cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to update parameter", err)
		return
	}
	apierr.OK(c, updated)
}

// Delete deletes parameter
func Delete(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	name := c.Param("name")
	if err := k8s.DeletePolarDBXParameterWithContext(c.Request.Context(), cli, ns, name); err != nil {
		util.HandleK8sError(c, "failed to delete parameter", err)
		return
	}
	apierr.OK(c, gin.H{"message": "parameter deleted"})
}

// ======================== Parameter Templates ========================

// ListTemplates gets parameter template list
func ListTemplates(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	items, err := k8s.ListPolarDBXParameterTemplatesWithContext(c.Request.Context(), cli, ns)
	if err != nil {
		util.HandleK8sError(c, "failed to list parameter templates", err)
		return
	}
	apierr.OK(c, items)
}

// GetTemplate gets specified parameter template
func GetTemplate(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	name := c.Param("name")
	item, err := k8s.GetPolarDBXParameterTemplateWithContext(c.Request.Context(), cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "failed to get parameter template", err)
		return
	}
	apierr.OK(c, item)
}

// CreateTemplate creates parameter template
func CreateTemplate(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	var body polardbxv1.PolarDBXParameterTemplate
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.AbortValidation(c, "invalid parameter template: "+err.Error())
		return
	}
	created, err := k8s.CreatePolarDBXParameterTemplateWithContext(c.Request.Context(), cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to create parameter template", err)
		return
	}
	apierr.Created(c, created)
}

// UpdateTemplate updates parameter template
func UpdateTemplate(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	var body polardbxv1.PolarDBXParameterTemplate
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.AbortValidation(c, "invalid parameter template: "+err.Error())
		return
	}
	updated, err := k8s.UpdatePolarDBXParameterTemplateWithContext(c.Request.Context(), cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to update parameter template", err)
		return
	}
	apierr.OK(c, updated)
}

// DeleteTemplate deletes parameter template
func DeleteTemplate(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := util.DefaultNamespace(c, "default")
	name := c.Param("name")
	if err := k8s.DeletePolarDBXParameterTemplateWithContext(c.Request.Context(), cli, ns, name); err != nil {
		util.HandleK8sError(c, "failed to delete parameter template", err)
		return
	}
	apierr.OK(c, gin.H{"message": "parameter template deleted"})
}
