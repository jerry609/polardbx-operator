package services

import (
	"fmt"
	"net/http"
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
	"polardbx-ui-backend/pkg/api/middleware"
	"polardbx-ui-backend/pkg/api/util"
)

// ValidationError 校验错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateClusterCreationConfig 校验集群创建配置
func ValidateClusterCreationConfig(config *ClusterCreationConfig) []ValidationError {
	var errors []ValidationError

	// 1. 名称校验
	if config.Name == "" {
		errors = append(errors, ValidationError{Field: "name", Message: "集群名称不能为空"})
	} else if !isValidK8sName(config.Name) {
		errors = append(errors, ValidationError{Field: "name", Message: "集群名称只能包含小写字母、数字和连字符，且必须以字母开头"})
	} else if len(config.Name) > 63 {
		errors = append(errors, ValidationError{Field: "name", Message: "集群名称不能超过63个字符"})
	}

	// 2. 命名空间校验
	if config.Namespace != "" && !isValidK8sName(config.Namespace) {
		errors = append(errors, ValidationError{Field: "namespace", Message: "命名空间格式不正确"})
	}

	// 3. 版本校验
	if config.Version != "" && !isValidVersion(config.Version) {
		errors = append(errors, ValidationError{Field: "version", Message: "版本格式不正确，应为 x.y.z 格式"})
	}

	// 4. 拓扑校验
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

	// 5. 存储校验
	if config.Storage.Size != "" && !isValidStorageSize(config.Storage.Size) {
		errors = append(errors, ValidationError{Field: "storage.size", Message: "存储大小格式不正确，应为如 10Gi, 100Gi, 1Ti 格式"})
	}

	// 6. 网络校验
	if config.Network != nil {
		if config.Network.ServiceType != "" &&
			config.Network.ServiceType != "ClusterIP" &&
			config.Network.ServiceType != "NodePort" &&
			config.Network.ServiceType != "LoadBalancer" {
			errors = append(errors, ValidationError{Field: "network.serviceType", Message: "服务类型必须是 ClusterIP, NodePort 或 LoadBalancer"})
		}
	}

	// 7. 安全配置校验
	if config.Security != nil {
		if config.Security.EnableTLS && config.Security.SecretName == "" {
			errors = append(errors, ValidationError{Field: "security.secretName", Message: "启用 TLS 时必须指定 Secret 名称"})
		}
	}

	// 8. 镜像配置校验
	if config.Image != nil {
		if config.Image.PullPolicy != "" &&
			config.Image.PullPolicy != "Always" &&
			config.Image.PullPolicy != "IfNotPresent" &&
			config.Image.PullPolicy != "Never" {
			errors = append(errors, ValidationError{Field: "image.pullPolicy", Message: "拉取策略必须是 Always, IfNotPresent 或 Never"})
		}
	}

	return errors
}

// isValidK8sName 检查是否是有效的 Kubernetes 名称
func isValidK8sName(name string) bool {
	pattern := regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$|^[a-z]$`)
	return pattern.MatchString(name)
}

// isValidVersion 检查版本格式
func isValidVersion(version string) bool {
	pattern := regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`)
	return pattern.MatchString(version)
}

// isValidStorageSize 检查存储大小格式
func isValidStorageSize(size string) bool {
	pattern := regexp.MustCompile(`^\d+(\.\d+)?(Ki|Mi|Gi|Ti|Pi|Ei)?$`)
	return pattern.MatchString(size)
}

// isValidResourceQuantity 检查资源数量格式 (CPU/Memory)
func isValidResourceQuantity(quantity string) bool {
	if quantity == "" {
		return true
	}
	cpuPattern := regexp.MustCompile(`^\d+(\.\d+)?(m)?$`)
	memPattern := regexp.MustCompile(`^\d+(\.\d+)?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)?$`)
	return cpuPattern.MatchString(quantity) || memPattern.MatchString(quantity)
}

// validateNodeConfig 校验节点配置
func validateNodeConfig(fieldPrefix string, node ClusterNodeConfig, maxReplicas int) *ValidationError {
	if node.Replicas < 1 {
		return &ValidationError{
			Field:   fieldPrefix + ".replicas",
			Message: fmt.Sprintf("%s 副本数必须至少为 1", fieldPrefix),
		}
	}

	if node.Replicas > maxReplicas {
		return &ValidationError{
			Field:   fieldPrefix + ".replicas",
			Message: fmt.Sprintf("%s 副本数不能超过 %d", fieldPrefix, maxReplicas),
		}
	}

	if node.Resources.CPU != "" && !isValidResourceQuantity(node.Resources.CPU) {
		return &ValidationError{
			Field:   fieldPrefix + ".resources.cpu",
			Message: "CPU 资源格式不正确",
		}
	}
	if node.Resources.Memory != "" && !isValidResourceQuantity(node.Resources.Memory) {
		return &ValidationError{
			Field:   fieldPrefix + ".resources.memory",
			Message: "内存资源格式不正确",
		}
	}

	return nil
}

// --- Cluster CRUD (行为保持不变) ---

func (s *ClusterService) List(c *gin.Context) {
	logger := middleware.NewBusinessLogger(c, "ClusterService")
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		logger.Error(nil, "failed to get k8s client from context")
		return
	}
	ns := c.DefaultQuery("namespace", "")
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
	c.JSON(http.StatusOK, clusters)
}

func (s *ClusterService) Create(c *gin.Context) {
	logger := middleware.NewBusinessLogger(c, "ClusterService")
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		logger.Error(nil, "failed to get k8s client from context")
		return
	}
	var obj polardbxv1.PolarDBXCluster
	if err := c.ShouldBindJSON(&obj); err != nil {
		logger.Error(err, "failed to parse cluster data")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse cluster data", "details": err.Error()})
		return
	}
	ns := obj.GetNamespace()
	if ns == "" {
		ns = "default"
	}
	logger.Info("creating cluster name=%s namespace=%s", obj.GetName(), ns)

	ctx, cancel := util.CrudCtx(c)
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
	c.JSON(http.StatusCreated, created)
}

// CreateFromConfig 从用户友好的配置格式创建集群
func (s *ClusterService) CreateFromConfig(c *gin.Context) {
	logger := middleware.NewBusinessLogger(c, "ClusterService")
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		logger.Error(nil, "failed to get k8s client from context")
		return
	}

	// 从URL参数获取namespace
	ns := c.Param("namespace")
	if ns == "" {
		ns = "default"
	}

	var config ClusterCreationConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		logger.Error(err, "failed to parse cluster creation config")
		c.JSON(http.StatusBadRequest, gin.H{"error": "配置解析失败", "details": err.Error()})
		return
	}

	logger.Info("creating cluster from config name=%s namespace=%s", config.Name, ns)

	// 参数校验
	if validationErrors := ValidateClusterCreationConfig(&config); len(validationErrors) > 0 {
		for _, ve := range validationErrors {
			middleware.LogValidationError(c, "ClusterService", ve.Field, ve.Message)
		}
		logger.Warn("cluster config validation failed with %d errors", len(validationErrors))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":            "配置校验失败",
			"validationErrors": validationErrors,
		})
		return
	}

	// 转换为 PolarDBXCluster 对象
	cluster := convertConfigToCluster(&config, ns)
	logger.Debug("converted config to PolarDBXCluster: CN=%d DN=%d",
		config.Topology.CN.Replicas, config.Topology.DN.Replicas)

	ctx, cancel := util.CrudCtx(c)
	defer cancel()
	created, err := k8srepo.NewClusterRepository().Create(ctx, cli, ns, cluster)
	if err != nil {
		// 提取更有意义的错误信息
		errMsg := err.Error()
		if strings.Contains(errMsg, "already exists") {
			logger.Warn("cluster already exists name=%s namespace=%s", config.Name, ns)
			c.JSON(http.StatusConflict, gin.H{"error": "集群已存在", "details": fmt.Sprintf("名为 %s 的集群在命名空间 %s 中已存在", config.Name, ns)})
			return
		}
		logger.Error(err, "failed to create cluster from config name=%s namespace=%s", config.Name, ns)
		middleware.LogK8sError(c, "Create", "PolarDBXCluster", ns, config.Name, err)
		middleware.LogAudit(c, "CREATE", "PolarDBXCluster", ns, config.Name, false)
		util.HandleK8sError(c, "创建集群失败", err)
		return
	}
	logger.Info("cluster created successfully from config name=%s namespace=%s", config.Name, ns)
	middleware.LogAudit(c, "CREATE", "PolarDBXCluster", ns, config.Name, true)
	c.JSON(http.StatusCreated, created)
}

// convertConfigToCluster 将用户友好配置转换为 PolarDBXCluster 对象
func convertConfigToCluster(config *ClusterCreationConfig, namespace string) *polardbxv1.PolarDBXCluster {
	// 构建CN副本数指针
	cnReplicas := int32(config.Topology.CN.Replicas)

	// 处理 HostNetwork 配置
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

	// 设置镜像配置
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
			// 设置CN镜像
			cluster.Spec.Topology.Nodes.CN.Template.Image = image
			// 设置DN镜像
			cluster.Spec.Topology.Nodes.DN.Template.Image = image
		}
		// 设置镜像拉取策略
		if config.Image.PullPolicy != "" {
			pullPolicy := corev1.PullPolicy(config.Image.PullPolicy)
			cluster.Spec.Topology.Nodes.CN.Template.ImagePullPolicy = pullPolicy
			cluster.Spec.Topology.Nodes.DN.Template.ImagePullPolicy = pullPolicy
		}
	}

	// 设置 ShareGMS 模式
	if config.Advanced != nil && config.Advanced.ShareGMS {
		cluster.Spec.ShareGMS = true
	}

	// 设置描述
	if config.Description != "" {
		if cluster.Annotations == nil {
			cluster.Annotations = make(map[string]string)
		}
		cluster.Annotations["description"] = config.Description
	}

	// 设置CDC配置（如果有）
	if config.Topology.CDC != nil && config.Topology.CDC.Replicas > 0 {
		cluster.Spec.Topology.Nodes.CDC = &polardbx.TopologyNodeCDC{
			Replicas: intstr.FromInt(config.Topology.CDC.Replicas),
			Template: polardbx.CDCTemplate{
				Resources:   buildResourceRequirements(config.Topology.CDC.Resources),
				HostNetwork: cnHostNetwork, // CDC 也使用相同的 HostNetwork 配置
			},
		}
	}

	// 设置网络服务类型
	if config.Network != nil && config.Network.ServiceType != "" {
		serviceType := corev1.ServiceType(config.Network.ServiceType)
		cluster.Spec.ServiceType = serviceType
	}

	// 设置TLS配置
	if config.Security != nil && config.Security.EnableTLS {
		if cluster.Spec.Security == nil {
			cluster.Spec.Security = &polardbx.Security{}
		}
		cluster.Spec.Security.TLS = &polardbx.TLS{
			SecretName: config.Security.SecretName,
		}
	}

	// 设置自定义标签、注解和节点选择器
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
		// 设置节点选择器（通过 TopologyRules）
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

// boolPtr 返回 bool 值的指针
func boolPtr(b bool) *bool {
	return &b
}

// buildNodeSelector 将 map 转换为 corev1.NodeSelector
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

// buildExtendedResourceRequirements 构建扩展资源需求（用于CN/DN）
func buildExtendedResourceRequirements(res NodeResources) polardbxcommon.ExtendedResourceRequirements {
	return polardbxcommon.ExtendedResourceRequirements{
		ResourceRequirements: buildResourceRequirements(res),
	}
}

// buildResourceRequirements 构建资源需求
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
	ns := c.Param("namespace")
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
	c.JSON(http.StatusOK, cluster)
}

func (s *ClusterService) Update(c *gin.Context) {
	logger := middleware.NewBusinessLogger(c, "ClusterService")
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		logger.Error(nil, "failed to get k8s client from context")
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	var body polardbxv1.PolarDBXCluster
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Error(err, "failed to parse cluster data for update")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse cluster data", "details": err.Error()})
		return
	}
	logger.Info("updating cluster name=%s namespace=%s", name, ns)

	ctx, cancel := util.CrudCtx(c)
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
	c.JSON(http.StatusOK, updated)
}

func (s *ClusterService) Delete(c *gin.Context) {
	logger := middleware.NewBusinessLogger(c, "ClusterService")
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		logger.Error(nil, "failed to get k8s client from context")
		return
	}
	ns := c.Param("namespace")
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
	c.JSON(http.StatusOK, gin.H{"message": "cluster deletion initiated successfully"})
}
