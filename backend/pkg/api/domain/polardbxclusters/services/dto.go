package services

// DTOs for cluster operations (kept minimal to preserve API contract)

type LogConfigRequest struct {
	EnableAuditLog   *bool  `json:"enableAuditLog,omitempty"`
	LogLevel         string `json:"logLevel,omitempty"`
	AuditLogFilter   string `json:"auditLogFilter,omitempty"`
	SlowLogThreshold *int   `json:"slowLogThreshold,omitempty"`
}

type ClusterScalingRequest struct {
	CNReplicas  *int32 `json:"cnReplicas,omitempty"`
	DNReplicas  *int32 `json:"dnReplicas,omitempty"`
	GMSReplicas *int32 `json:"gmsReplicas,omitempty"`
	CDCReplicas *int32 `json:"cdcReplicas,omitempty"`
}

type ClusterUpgradeRequest struct {
	TargetVersion  string `json:"targetVersion" binding:"required"`
	Strategy       string `json:"strategy,omitempty"`
	MaxUnavailable *int32 `json:"maxUnavailable,omitempty"`
}

// NodeResources 节点资源配置
type NodeResources struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// ClusterNodeConfig 集群节点配置
type ClusterNodeConfig struct {
	Replicas  int           `json:"replicas"`
	Resources NodeResources `json:"resources,omitempty"`
}

// ClusterTopologyConfig 集群拓扑配置
type ClusterTopologyConfig struct {
	CN  ClusterNodeConfig  `json:"cn"`
	DN  ClusterNodeConfig  `json:"dn"`
	GMS ClusterNodeConfig  `json:"gms"`
	CDC *ClusterNodeConfig `json:"cdc,omitempty"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	StorageClassName string   `json:"storageClassName,omitempty"`
	Size             string   `json:"size,omitempty"`
	AccessMode       string   `json:"accessMode,omitempty"`
	AccessModes      []string `json:"accessModes,omitempty"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	ServiceType       string `json:"serviceType,omitempty"`
	LoadBalancerClass string `json:"loadBalancerClass,omitempty"`
	HostNetwork       bool   `json:"hostNetwork,omitempty"` // 使用宿主网络模式
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableTLS  bool   `json:"enableTLS,omitempty"`
	SecretName string `json:"secretName,omitempty"`
}

// AdvancedConfig 高级配置
type AdvancedConfig struct {
	EnableMonitoring    bool              `json:"enableMonitoring,omitempty"`
	EnableBackup        bool              `json:"enableBackup,omitempty"`
	EnableLogCollection bool              `json:"enableLogCollection,omitempty"`
	CustomLabels        map[string]string `json:"customLabels,omitempty"`
	CustomAnnotations   map[string]string `json:"customAnnotations,omitempty"`
	NodeSelector        map[string]string `json:"nodeSelector,omitempty"` // 节点选择器
	ShareGMS            bool              `json:"shareGMS,omitempty"`     // GMS共享极简模式
}

// ImageConfig 镜像配置
type ImageConfig struct {
	Repository string `json:"repository,omitempty"` // 镜像仓库
	Tag        string `json:"tag,omitempty"`        // 镜像标签
	PullPolicy string `json:"pullPolicy,omitempty"` // 拉取策略: Always, IfNotPresent, Never
}

// ClusterCreationConfig 集群创建配置（用户友好格式）
type ClusterCreationConfig struct {
	Name        string                `json:"name" binding:"required"`
	Namespace   string                `json:"namespace,omitempty"`
	Description string                `json:"description,omitempty"`
	Version     string                `json:"version,omitempty"`
	Image       *ImageConfig          `json:"image,omitempty"` // 镜像配置
	Topology    ClusterTopologyConfig `json:"topology" binding:"required"`
	Storage     StorageConfig         `json:"storage"`
	Network     *NetworkConfig        `json:"network,omitempty"`
	Security    *SecurityConfig       `json:"security,omitempty"`
	Advanced    *AdvancedConfig       `json:"advanced,omitempty"`
}
