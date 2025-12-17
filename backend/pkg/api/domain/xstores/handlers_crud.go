package xstores

import (
	"polardbx-ui-backend/pkg/api/domain/xstores/services"

	"github.com/gin-gonic/gin"
)

// List lists XStores in the given namespace (or default namespace when not specified).
// @Summary List XStores
// @Description List XStore instances in the specified namespace.
// @Tags xstores
// @Produce json
// @Param namespace query string false "Kubernetes namespace; defaults to 'default' when omitted"
// @Success 200 {array} polardbxv1.XStore "List of XStores"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func List(c *gin.Context) { services.NewXStoreService().List(c) }

// Create creates a new XStore.
// @Summary Create XStore
// @Description Create a new XStore resource in the specified or default namespace.
// @Tags xstores
// @Accept json
// @Produce json
// @Param namespace query string false "Kubernetes namespace; defaults to 'default' when omitted"
// @Param body body polardbxv1.XStore true "XStore specification"
// @Success 201 {object} polardbxv1.XStore "Created XStore"
// @Failure 400 {object} map[string]any "Invalid request body"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func Create(c *gin.Context) { services.NewXStoreService().Create(c) }

// Get returns a single XStore by namespace and name.
// @Summary Get XStore
// @Description Get an XStore resource by namespace and name.
// @Tags xstores
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the XStore"
// @Param name path string true "Name of the XStore"
// @Success 200 {object} polardbxv1.XStore "XStore"
// @Failure 404 {object} map[string]any "XStore not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func Get(c *gin.Context) { services.NewXStoreService().Get(c) }

// Update updates an existing XStore.
// @Summary Update XStore
// @Description Update an existing XStore resource.
// @Tags xstores
// @Accept json
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the XStore"
// @Param name path string true "Name of the XStore"
// @Param body body polardbxv1.XStore true "Updated XStore specification"
// @Success 200 {object} polardbxv1.XStore "Updated XStore"
// @Failure 400 {object} map[string]any "Invalid request body"
// @Failure 404 {object} map[string]any "XStore not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func Update(c *gin.Context) { services.NewXStoreService().Update(c) }

// Delete deletes an XStore.
// @Summary Delete XStore
// @Description Delete an XStore resource by namespace and name.
// @Tags xstores
// @Param namespace path string true "Kubernetes namespace of the XStore"
// @Param name path string true "Name of the XStore"
// @Success 200 {object} map[string]any "Deletion confirmation"
// @Failure 404 {object} map[string]any "XStore not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func Delete(c *gin.Context) { services.NewXStoreService().Delete(c) }

// ListPods lists Pods that belong to the given XStore.
// @Summary List XStore pods
// @Description List Pods under the specified XStore.
// @Tags xstores
// @Produce json
// @Param namespace path string true "Kubernetes namespace of the XStore"
// @Param name path string true "Name of the XStore"
// @Success 200 {array} corev1.Pod "List of Pods"
// @Failure 404 {object} map[string]any "XStore or Pods not found"
// @Failure 502 {object} map[string]any "Upstream Kubernetes error"
func ListPods(c *gin.Context) { services.NewXStoreService().ListPods(c) }
