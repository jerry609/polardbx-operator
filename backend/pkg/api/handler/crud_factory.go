package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"polardbx-ui-backend/pkg/api/util"
)

// ResourceOperations defines the contract for K8s resource CRUD operations.
// T is the resource type (e.g., *polardbxv1.PolarDBXMonitor).
// ListResult is typically []T or a custom list type.
type ResourceOperations[T any, ListResult any] interface {
	List(ctx context.Context, cli client.Client, namespace string) (ListResult, error)
	Get(ctx context.Context, cli client.Client, namespace, name string) (T, error)
	Create(ctx context.Context, cli client.Client, namespace string, obj T) (T, error)
	Update(ctx context.Context, cli client.Client, namespace string, obj T) (T, error)
	Delete(ctx context.Context, cli client.Client, namespace, name string) error
}

// CRUDHandlerFactory creates standard CRUD handlers for a Kubernetes resource type.
// This implements the Factory Pattern to eliminate repetitive handler code.
type CRUDHandlerFactory[T any, ListResult any] struct {
	ops          ResourceOperations[T, ListResult]
	resourceName string
	newInstance  func() T // Factory function to create new instance
}

// NewCRUDHandlerFactory creates a new CRUD handler factory for a resource type.
func NewCRUDHandlerFactory[T any, ListResult any](
	ops ResourceOperations[T, ListResult],
	resourceName string,
	newInstance func() T,
) *CRUDHandlerFactory[T, ListResult] {
	return &CRUDHandlerFactory[T, ListResult]{
		ops:          ops,
		resourceName: resourceName,
		newInstance:  newInstance,
	}
}

// List returns a handler function for listing resources.
func (f *CRUDHandlerFactory[T, ListResult]) List() gin.HandlerFunc {
	return func(c *gin.Context) {
		cli, ok := util.K8sClientFromContext(c)
		if !ok {
			return
		}
		namespace := util.DefaultNamespace(c, "default")
		items, err := f.ops.List(c.Request.Context(), cli, namespace)
		if err != nil {
			util.HandleK8sError(c, "failed to list "+f.resourceName, err)
			return
		}
		c.JSON(http.StatusOK, items)
	}
}

// Get returns a handler function for getting a single resource.
func (f *CRUDHandlerFactory[T, ListResult]) Get() gin.HandlerFunc {
	return func(c *gin.Context) {
		cli, ok := util.K8sClientFromContext(c)
		if !ok {
			return
		}
		namespace := c.Param("namespace")
		name := c.Param("name")
		item, err := f.ops.Get(c.Request.Context(), cli, namespace, name)
		if err != nil {
			util.HandleK8sError(c, "failed to get "+f.resourceName, err)
			return
		}
		c.JSON(http.StatusOK, item)
	}
}

// Create returns a handler function for creating a resource.
func (f *CRUDHandlerFactory[T, ListResult]) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		cli, ok := util.K8sClientFromContext(c)
		if !ok {
			return
		}
		namespace := util.DefaultNamespace(c, "default")
		obj := f.newInstance()
		if err := c.ShouldBindJSON(&obj); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid " + f.resourceName,
				"details": err.Error(),
			})
			return
		}
		created, err := f.ops.Create(c.Request.Context(), cli, namespace, obj)
		if err != nil {
			util.HandleK8sError(c, "failed to create "+f.resourceName, err)
			return
		}
		c.JSON(http.StatusCreated, created)
	}
}

// Update returns a handler function for updating a resource.
func (f *CRUDHandlerFactory[T, ListResult]) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		cli, ok := util.K8sClientFromContext(c)
		if !ok {
			return
		}
		namespace := c.Param("namespace")
		obj := f.newInstance()
		if err := c.ShouldBindJSON(&obj); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid " + f.resourceName,
				"details": err.Error(),
			})
			return
		}
		updated, err := f.ops.Update(c.Request.Context(), cli, namespace, obj)
		if err != nil {
			util.HandleK8sError(c, "failed to update "+f.resourceName, err)
			return
		}
		c.JSON(http.StatusOK, updated)
	}
}

// Delete returns a handler function for deleting a resource.
func (f *CRUDHandlerFactory[T, ListResult]) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		cli, ok := util.K8sClientFromContext(c)
		if !ok {
			return
		}
		namespace := c.Param("namespace")
		name := c.Param("name")
		if err := f.ops.Delete(c.Request.Context(), cli, namespace, name); err != nil {
			util.HandleK8sError(c, "failed to delete "+f.resourceName, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": f.resourceName + " deleted"})
	}
}

// ClusterScopedCRUDHandlerFactory creates CRUD handlers for cluster-scoped (non-namespaced) resources.
type ClusterScopedCRUDHandlerFactory[T any, ListResult any] struct {
	ops          ResourceOperations[T, ListResult]
	resourceName string
	newInstance  func() T
}

// NewClusterScopedCRUDHandlerFactory creates a factory for cluster-scoped resources.
func NewClusterScopedCRUDHandlerFactory[T any, ListResult any](
	ops ResourceOperations[T, ListResult],
	resourceName string,
	newInstance func() T,
) *ClusterScopedCRUDHandlerFactory[T, ListResult] {
	return &ClusterScopedCRUDHandlerFactory[T, ListResult]{
		ops:          ops,
		resourceName: resourceName,
		newInstance:  newInstance,
	}
}

// Get returns a handler for cluster-scoped get (/:name instead of /:namespace/:name).
func (f *ClusterScopedCRUDHandlerFactory[T, ListResult]) Get() gin.HandlerFunc {
	return func(c *gin.Context) {
		cli, ok := util.K8sClientFromContext(c)
		if !ok {
			return
		}
		namespace := util.DefaultNamespace(c, "default")
		name := c.Param("name")
		item, err := f.ops.Get(c.Request.Context(), cli, namespace, name)
		if err != nil {
			util.HandleK8sError(c, "failed to get "+f.resourceName, err)
			return
		}
		c.JSON(http.StatusOK, item)
	}
}

// Update returns a handler for cluster-scoped update.
func (f *ClusterScopedCRUDHandlerFactory[T, ListResult]) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		cli, ok := util.K8sClientFromContext(c)
		if !ok {
			return
		}
		namespace := util.DefaultNamespace(c, "default")
		obj := f.newInstance()
		if err := c.ShouldBindJSON(&obj); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid " + f.resourceName,
				"details": err.Error(),
			})
			return
		}
		updated, err := f.ops.Update(c.Request.Context(), cli, namespace, obj)
		if err != nil {
			util.HandleK8sError(c, "failed to update "+f.resourceName, err)
			return
		}
		c.JSON(http.StatusOK, updated)
	}
}

// Delete returns a handler for cluster-scoped delete.
func (f *ClusterScopedCRUDHandlerFactory[T, ListResult]) Delete() gin.HandlerFunc {
	return func(c *gin.Context) {
		cli, ok := util.K8sClientFromContext(c)
		if !ok {
			return
		}
		namespace := util.DefaultNamespace(c, "default")
		name := c.Param("name")
		if err := f.ops.Delete(c.Request.Context(), cli, namespace, name); err != nil {
			util.HandleK8sError(c, "failed to delete "+f.resourceName, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": f.resourceName + " deleted"})
	}
}

// List and Create are the same as namespace-scoped, so we can reuse them.
func (f *ClusterScopedCRUDHandlerFactory[T, ListResult]) List() gin.HandlerFunc {
	// Delegate to namespace-scoped list handler
	nsFactory := NewCRUDHandlerFactory(f.ops, f.resourceName, f.newInstance)
	return nsFactory.List()
}

func (f *ClusterScopedCRUDHandlerFactory[T, ListResult]) Create() gin.HandlerFunc {
	// Delegate to namespace-scoped create handler
	nsFactory := NewCRUDHandlerFactory(f.ops, f.resourceName, f.newInstance)
	return nsFactory.Create()
}
