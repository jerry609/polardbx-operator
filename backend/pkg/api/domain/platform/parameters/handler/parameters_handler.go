// Package handler 提供参数和参数模板的 HTTP 处理器package handler

// 遵循 Clean Architecture 设计模式
package handler

import (
	apierr "polardbx-ui-backend/pkg/api/errors"
	"polardbx-ui-backend/pkg/api/util"
	"polardbx-ui-backend/pkg/k8s"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"github.com/gin-gonic/gin"
)

// ======================== Parameters ========================

// List 获取参数列表
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

// Get 获取指定参数
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

// Create 创建参数
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

// Update 更新参数
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

// Delete 删除参数
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

// ListTemplates 获取参数模板列表
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

// GetTemplate 获取指定参数模板
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

// CreateTemplate 创建参数模板
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

// UpdateTemplate 更新参数模板
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

// DeleteTemplate 删除参数模板
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
