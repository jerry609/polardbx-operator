package services

import (
	"net/http"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	polardbxcommon "github.com/alibaba/polardbx-operator/api/v1/common"
	polardbx "github.com/alibaba/polardbx-operator/api/v1/polardbx"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"polardbx-ui-backend/pkg/api/domain/polardbxclusters/k8srepo"
	"polardbx-ui-backend/pkg/api/util"
)

// --- Cluster CRUD (行为保持不变) ---

func (s *ClusterService) List(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.DefaultQuery("namespace", "")
	ctx, cancel := util.ListCtx(c)
	defer cancel()
	clusters, err := k8srepo.NewClusterRepository().List(ctx, cli, ns)
	if err != nil {
		util.HandleK8sError(c, "failed to list clusters", err)
		return
	}
	c.JSON(http.StatusOK, clusters)
}

func (s *ClusterService) Create(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	var obj polardbxv1.PolarDBXCluster
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse cluster data", "details": err.Error()})
		return
	}
	ns := obj.GetNamespace()
	if ns == "" {
		ns = "default"
	}
	ctx, cancel := util.CrudCtx(c)
	defer cancel()
	created, err := k8srepo.NewClusterRepository().Create(ctx, cli, ns, &obj)
	if err != nil {
		util.HandleK8sError(c, "failed to create cluster", err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

// CreateFromConfig 从用户友好的配置格式创建集群
func (s *ClusterService) CreateFromConfig(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}

	// 从URL参数获取namespace
	ns := c.Param("namespace")
	if ns == "" {
		ns = "default"
	}

	var config ClusterCreationConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse cluster config", "details": err.Error()})
		return
	}

	// 转换为 PolarDBXCluster 对象
	cluster := convertConfigToCluster(&config, ns)

	ctx, cancel := util.CrudCtx(c)
	defer cancel()
	created, err := k8srepo.NewClusterRepository().Create(ctx, cli, ns, cluster)
	if err != nil {
		util.HandleK8sError(c, "failed to create cluster", err)
		return
	}
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
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	ctx, cancel := util.ListCtx(c)
	defer cancel()
	cluster, err := k8srepo.NewClusterRepository().Get(ctx, cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "cluster not found", err)
		return
	}
	c.JSON(http.StatusOK, cluster)
}

func (s *ClusterService) Update(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	var body polardbxv1.PolarDBXCluster
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse cluster data", "details": err.Error()})
		return
	}
	ctx, cancel := util.CrudCtx(c)
	defer cancel()
	existing, err := k8srepo.NewClusterRepository().Get(ctx, cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "cluster not found", err)
		return
	}
	body.SetResourceVersion(existing.GetResourceVersion())
	updated, err := k8srepo.NewClusterRepository().Update(ctx, cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to update cluster", err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (s *ClusterService) Delete(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	ctx, cancel := util.CrudCtx(c)
	defer cancel()
	if err := k8srepo.NewClusterRepository().Delete(ctx, cli, ns, name); err != nil {
		util.HandleK8sError(c, "failed to delete cluster", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cluster deletion initiated successfully"})
}
