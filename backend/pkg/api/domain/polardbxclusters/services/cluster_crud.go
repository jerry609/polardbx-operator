package services

import (
	"fmt"
	"regexp"
	"strings"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	polardbxcommon "github.com/alibaba/polardbx-operator/api/v1/common"
	polardbx "github.com/alibaba/polardbx-operator/api/v1/polardbx"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"polardbx-ui-backend/pkg/api/domain/polardbxclusters/k8srepo"
	apierr "polardbx-ui-backend/pkg/api/errors"
	"polardbx-ui-backend/pkg/api/middleware"
	"polardbx-ui-backend/pkg/api/util"
)

// ValidationError validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateClusterCreationConfig validates cluster creation configuration
func ValidateClusterCreationConfig(config *ClusterCreationConfig) []ValidationError {
	var errors []ValidationError

	// 1. Name validation
	if config.Name == "" {
		errors = append(errors, ValidationError{Field: "name", Message: "cluster name cannot be empty"})
	} else if !isValidK8sName(config.Name) {
		errors = append(errors, ValidationError{Field: "name", Message: "cluster name can only contain lowercase letters, numbers, and hyphens, and must start with a letter"})
	} else if len(config.Name) > 63 {
		errors = append(errors, ValidationError{Field: "name", Message: "cluster name cannot exceed 63 characters"})
	}

	// 2. Namespace validation
	if config.Namespace != "" && !isValidK8sName(config.Namespace) {
		errors = append(errors, ValidationError{Field: "namespace", Message: "namespace format is incorrect"})
	}

	// 3. Version validation
	if config.Version != "" && !isValidVersion(config.Version) {
		errors = append(errors, ValidationError{Field: "version", Message: "version format is incorrect, should be in x.y.z format"})
	}

	// 4. Topology validation
	if err := validateNodeConfig("topology.cn", config.Topology.CN, 100); err != nil {
		errors = append(errors, *err)
	}
	if err := validateNodeConfig("topology.dn", config.Topology.DN, 100); err != nil {
		errors = append(errors, *err)
	}
	if err := validateNodeConfig("topology.gms", config.Topology.GMS, 3); err != nil {
		errors = append(errors, *err)
	}
	if config.Topology.CDC != nil {
		if err := validateNodeConfig("topology.cdc", *config.Topology.CDC, 10); err != nil {
			errors = append(errors, *err)
		}
	}

	// 5. Storage validation
	if config.Storage.Size != "" && !isValidStorageSize(config.Storage.Size) {
		errors = append(errors, ValidationError{Field: "storage.size", Message: "storage size format is incorrect, should be in format like 10Gi, 100Gi, 1Ti"})
	}

	// 6. Network validation
	if config.Network != nil {
		if config.Network.ServiceType != "" &&
			config.Network.ServiceType != "ClusterIP" &&
			config.Network.ServiceType != "NodePort" &&
			config.Network.ServiceType != "LoadBalancer" {
			errors = append(errors, ValidationError{Field: "network.serviceType", Message: "service type must be ClusterIP, NodePort, or LoadBalancer"})
		}
	}

	// 7. Security configuration validation
	if config.Security != nil {
		if config.Security.EnableTLS && config.Security.SecretName == "" {
			errors = append(errors, ValidationError{Field: "security.secretName", Message: "secret name must be specified when TLS is enabled"})
		}
	}

	// 8. Image configuration validation
	if config.Image != nil {
		if config.Image.PullPolicy != "" &&
			config.Image.PullPolicy != "Always" &&
			config.Image.PullPolicy != "IfNotPresent" &&
			config.Image.PullPolicy != "Never" {
			errors = append(errors, ValidationError{Field: "image.pullPolicy", Message: "pull policy must be Always, IfNotPresent, or Never"})
		}
	}

	return errors
}

// isValidK8sName checks if it's a valid Kubernetes name
func isValidK8sName(name string) bool {
	pattern := regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$|^[a-z]$`)
	return pattern.MatchString(name)
}

// isValidVersion checks version format
func isValidVersion(version string) bool {
	pattern := regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`)
	return pattern.MatchString(version)
}

// isValidStorageSize checks storage size format
func isValidStorageSize(size string) bool {
	pattern := regexp.MustCompile(`^\d+(\.\d+)?(Ki|Mi|Gi|Ti|Pi|Ei)?$`)
	return pattern.MatchString(size)
}

// isValidResourceQuantity checks resource quantity format (CPU/Memory)
func isValidResourceQuantity(quantity string) bool {
	if quantity == "" {
		return true
	}
	cpuPattern := regexp.MustCompile(`^\d+(\.\d+)?(m)?$`)
	memPattern := regexp.MustCompile(`^\d+(\.\d+)?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)?$`)
	return cpuPattern.MatchString(quantity) || memPattern.MatchString(quantity)
}

// validateNodeConfig validates node configuration
func validateNodeConfig(fieldPrefix string, node ClusterNodeConfig, maxReplicas int) *ValidationError {
	if node.Replicas < 1 {
		return &ValidationError{
			Field:   fieldPrefix + ".replicas",
			Message: fmt.Sprintf("%s replicas must be at least 1", fieldPrefix),
		}
	}

	if node.Replicas > maxReplicas {
		return &ValidationError{
			Field:   fieldPrefix + ".replicas",
			Message: fmt.Sprintf("%s replicas cannot exceed %d", fieldPrefix, maxReplicas),
		}
	}

	if node.Resources.CPU != "" && !isValidResourceQuantity(node.Resources.CPU) {
		return &ValidationError{
			Field:   fieldPrefix + ".resources.cpu",
			Message: "CPU resource format is incorrect",
		}
	}
	if node.Resources.Memory != "" && !isValidResourceQuantity(node.Resources.Memory) {
		return &ValidationError{
			Field:   fieldPrefix + ".resources.memory",
			Message: "memory resource format is incorrect",
		}
	}

	return nil
}

// --- Cluster CRUD (behavior remains unchanged) ---

func (s *ClusterService) List(c *gin.Context) {
	logger := middleware.NewBusinessLogger(c, "ClusterService")
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		logger.Error(nil, "failed to get k8s client from context")
		return
	}
	ns := util.GetNamespace(c, "")
	logger.Info("listing clusters in namespace=%s", ns)

	ctx, cancel := util.ListCtx(c)
	defer cancel()
	clusters, err := k8srepo.NewClusterRepository().List(ctx, cli, ns)
	if err != nil {
		logger.Error(err, "failed to list clusters")
		middleware.LogK8sError(c, "List", "PolarDBXCluster", ns, "*", err)
		util.HandleK8sError(c, "failed to list clusters", err)
		return
	}
	logger.Info("listed %d clusters", len(clusters))
	apierr.OK(c, clusters)
}

func (s *ClusterService) Create(c *gin.Context) {
	logger := middleware.NewBusinessLogger(c, "ClusterService")
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		logger.Error(nil, "failed to get k8s client from context")
		return
	}
	var obj polardbxv1.PolarDBXCluster
	ctx, cancel, ok := util.BindValidateAndCtx(c, &obj, util.DefaultCRUDTimeout)
	if !ok {
		return
	}
	ns := util.GetNamespace(c, obj.GetNamespace())
	if ns == "" {
		ns = "default"
	}
	logger.Info("creating cluster name=%s namespace=%s", obj.GetName(), ns)

	defer cancel()
	created, err := k8srepo.NewClusterRepository().Create(ctx, cli, ns, &obj)
	if err != nil {
		logger.Error(err, "failed to create cluster name=%s namespace=%s", obj.GetName(), ns)
		middleware.LogK8sError(c, "Create", "PolarDBXCluster", ns, obj.GetName(), err)
		middleware.LogAudit(c, "CREATE", "PolarDBXCluster", ns, obj.GetName(), false)
		util.HandleK8sError(c, "failed to create cluster", err)
		return
	}
	logger.Info("cluster created successfully name=%s namespace=%s", obj.GetName(), ns)
	middleware.LogAudit(c, "CREATE", "PolarDBXCluster", ns, obj.GetName(), true)
	apierr.Created(c, created)
}

// CreateFromConfig creates cluster from user-friendly configuration format
func (s *ClusterService) CreateFromConfig(c *gin.Context) {
	logger := middleware.NewBusinessLogger(c, "ClusterService")
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		logger.Error(nil, "failed to get k8s client from context")
		return
	}

	// Get namespace from URL parameter
	ns := util.GetNamespace(c, "default")

	var config ClusterCreationConfig
	ctx, cancel, ok := util.BindValidateAndCtx(c, &config, util.DefaultCRUDTimeout)
	if !ok {
		return
	}
	defer cancel()

	logger.Info("creating cluster from config name=%s namespace=%s", config.Name, ns)

	// Parameter validation
	if validationErrors := ValidateClusterCreationConfig(&config); len(validationErrors) > 0 {
		for _, ve := range validationErrors {
			middleware.LogValidationError(c, "ClusterService", ve.Field, ve.Message)
		}
		logger.Warn("cluster config validation failed with %d errors", len(validationErrors))
		apierr.Abort(c, apierr.Validation("configuration validation failed").WithDetails(map[string]string{
			"validationErrors": fmt.Sprintf("%+v", validationErrors),
		}))
		return
	}

	// Convert to PolarDBXCluster object
	cluster := convertConfigToCluster(&config, ns)
	logger.Debug("converted config to PolarDBXCluster: CN=%d DN=%d",
		config.Topology.CN.Replicas, config.Topology.DN.Replicas)

	created, err := k8srepo.NewClusterRepository().Create(ctx, cli, ns, cluster)
	if err != nil {
		// Extract more meaningful error information
		errMsg := err.Error()
		if strings.Contains(errMsg, "already exists") {
			logger.Warn("cluster already exists name=%s namespace=%s", config.Name, ns)
			apierr.Abort(c, apierr.AlreadyExists("cluster", config.Name))
			return
		}
		logger.Error(err, "failed to create cluster from config name=%s namespace=%s", config.Name, ns)
		middleware.LogK8sError(c, "Create", "PolarDBXCluster", ns, config.Name, err)
		middleware.LogAudit(c, "CREATE", "PolarDBXCluster", ns, config.Name, false)
		util.HandleK8sError(c, "failed to create cluster", err)
		return
	}
	logger.Info("cluster created successfully from config name=%s namespace=%s", config.Name, ns)
	middleware.LogAudit(c, "CREATE", "PolarDBXCluster", ns, config.Name, true)
	apierr.Created(c, created)
}

// convertConfigToCluster converts user-friendly configuration to PolarDBXCluster object
func convertConfigToCluster(config *ClusterCreationConfig, namespace string) *polardbxv1.PolarDBXCluster {
	// Build CN replicas pointer
	cnReplicas := int32(config.Topology.CN.Replicas)

	// Handle HostNetwork configuration
	var cnHostNetwork, dnHostNetwork bool
	if config.Network != nil {
		cnHostNetwork = config.Network.HostNetwork
		dnHostNetwork = config.Network.HostNetwork
	}

	cluster := &polardbxv1.PolarDBXCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: namespace,
		},
		Spec: polardbxv1.PolarDBXClusterSpec{
			Topology: polardbx.Topology{
				Version: config.Version,
				Nodes: polardbx.TopologyNodes{
					CN: polardbx.TopologyNodeCN{
						Replicas: &cnReplicas,
						Template: polardbx.CNTemplate{
							Resources:   buildResourceRequirements(config.Topology.CN.Resources),
							HostNetwork: cnHostNetwork,
						},
					},
					DN: polardbx.TopologyNodeDN{
						Replicas: int32(config.Topology.DN.Replicas),
						Template: polardbx.XStoreTemplate{
							Resources:   buildExtendedResourceRequirements(config.Topology.DN.Resources),
							HostNetwork: boolPtr(dnHostNetwork),
						},
					},
				},
			},
		},
	}

	// Set image configuration
	if config.Image != nil {
		if config.Image.Repository != "" || config.Image.Tag != "" {
			image := config.Image.Repository
			if config.Image.Tag != "" {
				if image != "" {
					image = image + ":" + config.Image.Tag
				} else {
					image = config.Image.Tag
				}
			}
			// Set CN image
			cluster.Spec.Topology.Nodes.CN.Template.Image = image
			// Set DN image
			cluster.Spec.Topology.Nodes.DN.Template.Image = image
		}
		// Set image pull policy
		if config.Image.PullPolicy != "" {
			pullPolicy := corev1.PullPolicy(config.Image.PullPolicy)
			cluster.Spec.Topology.Nodes.CN.Template.ImagePullPolicy = pullPolicy
			cluster.Spec.Topology.Nodes.DN.Template.ImagePullPolicy = pullPolicy
		}
	}

	// Set ShareGMS mode
	if config.Advanced != nil && config.Advanced.ShareGMS {
		cluster.Spec.ShareGMS = true
	}

	// Set description
	if config.Description != "" {
		if cluster.Annotations == nil {
			cluster.Annotations = make(map[string]string)
		}
		cluster.Annotations["description"] = config.Description
	}

	// Set CDC configuration (if any)
	if config.Topology.CDC != nil && config.Topology.CDC.Replicas > 0 {
		cluster.Spec.Topology.Nodes.CDC = &polardbx.TopologyNodeCDC{
			Replicas: intstr.FromInt(config.Topology.CDC.Replicas),
			Template: polardbx.CDCTemplate{
				Resources:   buildResourceRequirements(config.Topology.CDC.Resources),
				HostNetwork: cnHostNetwork, // CDC also uses the same HostNetwork configuration
			},
		}
	}

	// Set network service type
	if config.Network != nil && config.Network.ServiceType != "" {
		serviceType := corev1.ServiceType(config.Network.ServiceType)
		cluster.Spec.ServiceType = serviceType
	}

	// Set TLS configuration
	if config.Security != nil && config.Security.EnableTLS {
		if cluster.Spec.Security == nil {
			cluster.Spec.Security = &polardbx.Security{}
		}
		cluster.Spec.Security.TLS = &polardbx.TLS{
			SecretName: config.Security.SecretName,
		}
	}

	// Set custom labels, annotations, and node selectors
	if config.Advanced != nil {
		if len(config.Advanced.CustomLabels) > 0 {
			if cluster.Labels == nil {
				cluster.Labels = make(map[string]string)
			}
			for k, v := range config.Advanced.CustomLabels {
				cluster.Labels[k] = v
			}
		}
		if len(config.Advanced.CustomAnnotations) > 0 {
			if cluster.Annotations == nil {
				cluster.Annotations = make(map[string]string)
			}
			for k, v := range config.Advanced.CustomAnnotations {
				cluster.Annotations[k] = v
			}
		}
		// Set node selector (via TopologyRules)
		if len(config.Advanced.NodeSelector) > 0 {
			nodeSelector := buildNodeSelector(config.Advanced.NodeSelector)
			cluster.Spec.Topology.Rules = polardbx.TopologyRules{
				Selectors: []polardbx.NodeSelectorItem{
					{
						Name:         "default",
						NodeSelector: nodeSelector,
					},
				},
			}
		}
	}

	return cluster
}

// boolPtr returns pointer to bool value
func boolPtr(b bool) *bool {
	return &b
}

// buildNodeSelector converts map to corev1.NodeSelector
func buildNodeSelector(labels map[string]string) corev1.NodeSelector {
	var matchExpressions []corev1.NodeSelectorRequirement
	for key, value := range labels {
		matchExpressions = append(matchExpressions, corev1.NodeSelectorRequirement{
			Key:      key,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{value},
		})
	}
	return corev1.NodeSelector{
		NodeSelectorTerms: []corev1.NodeSelectorTerm{
			{
				MatchExpressions: matchExpressions,
			},
		},
	}
}

// buildExtendedResourceRequirements builds extended resource requirements (for CN/DN)
func buildExtendedResourceRequirements(res NodeResources) polardbxcommon.ExtendedResourceRequirements {
	return polardbxcommon.ExtendedResourceRequirements{
		ResourceRequirements: buildResourceRequirements(res),
	}
}

// buildResourceRequirements builds resource requirements
func buildResourceRequirements(res NodeResources) corev1.ResourceRequirements {
	requirements := corev1.ResourceRequirements{
		Limits:   corev1.ResourceList{},
		Requests: corev1.ResourceList{},
	}

	if res.CPU != "" {
		cpuQty := resource.MustParse(res.CPU)
		requirements.Limits[corev1.ResourceCPU] = cpuQty
		requirements.Requests[corev1.ResourceCPU] = cpuQty
	}

	if res.Memory != "" {
		memQty := resource.MustParse(res.Memory)
		requirements.Limits[corev1.ResourceMemory] = memQty
		requirements.Requests[corev1.ResourceMemory] = memQty
	}

	return requirements
}

func (s *ClusterService) Get(c *gin.Context) {
	logger := middleware.NewBusinessLogger(c, "ClusterService")
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		logger.Error(nil, "failed to get k8s client from context")
		return
	}
	ns := util.GetNamespace(c, "default")
	name := c.Param("name")
	logger.Debug("getting cluster name=%s namespace=%s", name, ns)

	ctx, cancel := util.ListCtx(c)
	defer cancel()
	cluster, err := k8srepo.NewClusterRepository().Get(ctx, cli, ns, name)
	if err != nil {
		logger.Error(err, "cluster not found name=%s namespace=%s", name, ns)
		middleware.LogK8sError(c, "Get", "PolarDBXCluster", ns, name, err)
		util.HandleK8sError(c, "cluster not found", err)
		return
	}
	logger.Debug("cluster retrieved successfully name=%s namespace=%s phase=%s",
		name, ns, cluster.Status.Phase)
	apierr.OK(c, cluster)
}

func (s *ClusterService) Update(c *gin.Context) {
	logger := middleware.NewBusinessLogger(c, "ClusterService")
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		logger.Error(nil, "failed to get k8s client from context")
		return
	}
	ns := util.GetNamespace(c, "default")
	name := c.Param("name")
	var body polardbxv1.PolarDBXCluster
	ctx, cancel, ok := util.BindValidateAndCtx(c, &body, util.DefaultCRUDTimeout)
	if !ok {
		return
	}
	logger.Info("updating cluster name=%s namespace=%s", name, ns)

	defer cancel()
	existing, err := k8srepo.NewClusterRepository().Get(ctx, cli, ns, name)
	if err != nil {
		logger.Error(err, "cluster not found for update name=%s namespace=%s", name, ns)
		middleware.LogK8sError(c, "Get", "PolarDBXCluster", ns, name, err)
		util.HandleK8sError(c, "cluster not found", err)
		return
	}
	body.SetResourceVersion(existing.GetResourceVersion())
	updated, err := k8srepo.NewClusterRepository().Update(ctx, cli, ns, &body)
	if err != nil {
		logger.Error(err, "failed to update cluster name=%s namespace=%s", name, ns)
		middleware.LogK8sError(c, "Update", "PolarDBXCluster", ns, name, err)
		middleware.LogAudit(c, "UPDATE", "PolarDBXCluster", ns, name, false)
		util.HandleK8sError(c, "failed to update cluster", err)
		return
	}
	logger.Info("cluster updated successfully name=%s namespace=%s", name, ns)
	middleware.LogAudit(c, "UPDATE", "PolarDBXCluster", ns, name, true)
	apierr.OK(c, updated)
}

func (s *ClusterService) Delete(c *gin.Context) {
	logger := middleware.NewBusinessLogger(c, "ClusterService")
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		logger.Error(nil, "failed to get k8s client from context")
		return
	}
	ns := util.GetNamespace(c, "default")
	name := c.Param("name")
	logger.Info("deleting cluster name=%s namespace=%s", name, ns)

	ctx, cancel := util.CrudCtx(c)
	defer cancel()
	if err := k8srepo.NewClusterRepository().Delete(ctx, cli, ns, name); err != nil {
		logger.Error(err, "failed to delete cluster name=%s namespace=%s", name, ns)
		middleware.LogK8sError(c, "Delete", "PolarDBXCluster", ns, name, err)
		middleware.LogAudit(c, "DELETE", "PolarDBXCluster", ns, name, false)
		util.HandleK8sError(c, "failed to delete cluster", err)
		return
	}
	logger.Info("cluster deletion initiated successfully name=%s namespace=%s", name, ns)
	middleware.LogAudit(c, "DELETE", "PolarDBXCluster", ns, name, true)
	apierr.OK(c, gin.H{"message": "cluster deletion initiated successfully"})
}
