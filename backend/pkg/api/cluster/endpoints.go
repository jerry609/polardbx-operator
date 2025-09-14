package cluster

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"github.com/alibaba/polardbx-operator/api/v1/common"
	polardbx "github.com/alibaba/polardbx-operator/api/v1/polardbx"

	"polardbx-ui-backend/pkg/api/prechange"
	"polardbx-ui-backend/pkg/api/util"
	"polardbx-ui-backend/pkg/k8s"
)

// ----- Cluster basic CRUD -----

func List(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ctx, cancel := util.ListCtx(c)
	defer cancel()
	clusters, err := k8s.ListPolarDBXClustersWithContext(ctx, cli, "")
	if err != nil {
		util.HandleK8sError(c, "failed to list clusters", err)
		return
	}
	c.JSON(http.StatusOK, clusters)
}

func Create(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ctx, cancel := util.CrudCtx(c)
	defer cancel()
	var cluster polardbxv1.PolarDBXCluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse cluster data", "details": err.Error()})
		return
	}
	ns := cluster.GetNamespace()
	if ns == "" {
		ns = "default"
	}
	created, err := k8s.CreatePolarDBXClusterWithContext(ctx, cli, ns, &cluster)
	if err != nil {
		util.HandleK8sError(c, "failed to create cluster", err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

// ClusterCreationRequest mirrors the root request structure
type ClusterCreationRequest struct {
	Name        string `json:"name" binding:"required"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
	Version     string `json:"version"`

	Topology struct {
		CN struct {
			Replicas  int `json:"replicas" binding:"required,min=1"`
			Resources struct {
				CPU    string `json:"cpu" binding:"required"`
				Memory string `json:"memory" binding:"required"`
			} `json:"resources" binding:"required"`
		} `json:"cn" binding:"required"`
		DN struct {
			Replicas  int `json:"replicas" binding:"required,min=1"`
			Resources struct {
				CPU    string `json:"cpu" binding:"required"`
				Memory string `json:"memory" binding:"required"`
			} `json:"resources" binding:"required"`
		} `json:"dn" binding:"required"`
		GMS struct {
			Replicas  int `json:"replicas" binding:"required,min=1"`
			Resources struct {
				CPU    string `json:"cpu" binding:"required"`
				Memory string `json:"memory" binding:"required"`
			} `json:"resources" binding:"required"`
		} `json:"gms" binding:"required"`
		CDC *struct {
			Replicas  int `json:"replicas" binding:"min=1"`
			Resources struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
			} `json:"resources"`
		} `json:"cdc,omitempty"`
	} `json:"topology" binding:"required"`

	Storage struct {
		StorageClassName string   `json:"storageClassName" binding:"required"`
		Size             string   `json:"size" binding:"required"`
		AccessModes      []string `json:"accessModes"`
	} `json:"storage" binding:"required"`

	Network struct {
		ServiceType       string `json:"serviceType" binding:"required"`
		LoadBalancerClass string `json:"loadBalancerClass"`
	} `json:"network" binding:"required"`

	Security struct {
		EnableTLS  bool   `json:"enableTLS"`
		SecretName string `json:"secretName"`
	} `json:"security"`

	Advanced struct {
		EnableMonitoring    bool              `json:"enableMonitoring"`
		EnableBackup        bool              `json:"enableBackup"`
		EnableLogCollection bool              `json:"enableLogCollection"`
		CustomLabels        map[string]string `json:"customLabels"`
		CustomAnnotations   map[string]string `json:"customAnnotations"`
	} `json:"advanced"`
}

func CreateFromConfig(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	var req ClusterCreationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse cluster configuration", "details": err.Error()})
		return
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if req.Version == "" {
		req.Version = "8.0.18"
	}
	if req.Storage.AccessModes == nil {
		req.Storage.AccessModes = []string{"ReadWriteOnce"}
	}
	cluster, err := convertToCluster(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to convert cluster configuration", "details": err.Error()})
		return
	}
	created, err := k8s.CreatePolarDBXClusterWithContext(c.Request.Context(), cli, req.Namespace, cluster)
	if err != nil {
		util.HandleK8sError(c, "failed to create cluster", err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func convertToCluster(req *ClusterCreationRequest) (*polardbxv1.PolarDBXCluster, error) {
	cluster := &polardbxv1.PolarDBXCluster{ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: req.Namespace, Labels: map[string]string{}, Annotations: map[string]string{}}, Spec: polardbxv1.PolarDBXClusterSpec{ServiceName: req.Name, Topology: polardbx.Topology{Version: req.Version}}}
	if req.Description != "" {
		cluster.Annotations["polardbx.aliyun.com/description"] = req.Description
	}
	for k, v := range req.Advanced.CustomLabels {
		cluster.Labels[k] = v
	}
	for k, v := range req.Advanced.CustomAnnotations {
		cluster.Annotations[k] = v
	}
	cnCPU, err := resource.ParseQuantity(req.Topology.CN.Resources.CPU)
	if err != nil {
		return nil, fmt.Errorf("invalid CN CPU resource: %v", err)
	}
	cnMem, err := resource.ParseQuantity(req.Topology.CN.Resources.Memory)
	if err != nil {
		return nil, fmt.Errorf("invalid CN memory resource: %v", err)
	}
	cnRep := int32(req.Topology.CN.Replicas)
	cluster.Spec.Topology.Nodes.CN = polardbx.TopologyNodeCN{Replicas: &cnRep, Template: polardbx.CNTemplate{Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: cnCPU, corev1.ResourceMemory: cnMem}, Limits: corev1.ResourceList{corev1.ResourceCPU: cnCPU, corev1.ResourceMemory: cnMem}}}}
	dnCPU, err := resource.ParseQuantity(req.Topology.DN.Resources.CPU)
	if err != nil {
		return nil, fmt.Errorf("invalid DN CPU resource: %v", err)
	}
	dnMem, err := resource.ParseQuantity(req.Topology.DN.Resources.Memory)
	if err != nil {
		return nil, fmt.Errorf("invalid DN memory resource: %v", err)
	}
	if _, err := resource.ParseQuantity(req.Storage.Size); err != nil {
		return nil, fmt.Errorf("invalid storage size: %v", err)
	}
	cluster.Spec.Topology.Nodes.DN = polardbx.TopologyNodeDN{Replicas: int32(req.Topology.DN.Replicas), Template: polardbx.XStoreTemplate{Resources: common.ExtendedResourceRequirements{ResourceRequirements: corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: dnCPU, corev1.ResourceMemory: dnMem}, Limits: corev1.ResourceList{corev1.ResourceCPU: dnCPU, corev1.ResourceMemory: dnMem}}}}}
	gmsCPU, err := resource.ParseQuantity(req.Topology.GMS.Resources.CPU)
	if err != nil {
		return nil, fmt.Errorf("invalid GMS CPU resource: %v", err)
	}
	gmsMem, err := resource.ParseQuantity(req.Topology.GMS.Resources.Memory)
	if err != nil {
		return nil, fmt.Errorf("invalid GMS memory resource: %v", err)
	}
	cluster.Spec.Topology.Nodes.GMS = polardbx.TopologyNodeGMS{Template: &polardbx.XStoreTemplate{Resources: common.ExtendedResourceRequirements{ResourceRequirements: corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: gmsCPU, corev1.ResourceMemory: gmsMem}, Limits: corev1.ResourceList{corev1.ResourceCPU: gmsCPU, corev1.ResourceMemory: gmsMem}}}}}
	if req.Topology.CDC != nil {
		cdcCPU, err := resource.ParseQuantity(req.Topology.CDC.Resources.CPU)
		if err != nil {
			return nil, fmt.Errorf("invalid CDC CPU resource: %v", err)
		}
		cdcMem, err := resource.ParseQuantity(req.Topology.CDC.Resources.Memory)
		if err != nil {
			return nil, fmt.Errorf("invalid CDC memory resource: %v", err)
		}
		cluster.Spec.Topology.Nodes.CDC = &polardbx.TopologyNodeCDC{Replicas: intstr.FromInt(req.Topology.CDC.Replicas), Template: polardbx.CDCTemplate{Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: cdcCPU, corev1.ResourceMemory: cdcMem}, Limits: corev1.ResourceList{corev1.ResourceCPU: cdcCPU, corev1.ResourceMemory: cdcMem}}}}
	}
	switch req.Network.ServiceType {
	case "NodePort":
		cluster.Spec.ServiceType = corev1.ServiceTypeNodePort
	case "LoadBalancer":
		cluster.Spec.ServiceType = corev1.ServiceTypeLoadBalancer
	default:
		cluster.Spec.ServiceType = corev1.ServiceTypeClusterIP
	}
	if req.Advanced.EnableMonitoring {
		if cluster.Annotations == nil {
			cluster.Annotations = map[string]string{}
		}
		cluster.Annotations["polardbx.aliyun.com/enable-monitoring"] = "true"
	}
	if req.Advanced.EnableBackup {
		if cluster.Annotations == nil {
			cluster.Annotations = map[string]string{}
		}
		cluster.Annotations["polardbx.aliyun.com/enable-backup"] = "true"
	}
	if req.Advanced.EnableLogCollection {
		if cluster.Annotations == nil {
			cluster.Annotations = map[string]string{}
		}
		cluster.Annotations["polardbx.aliyun.com/enable-log-collection"] = "true"
	}
	return cluster, nil
}

func Get(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	cluster, err := k8s.GetPolarDBXClusterWithContext(c.Request.Context(), cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "cluster not found", err)
		return
	}
	c.JSON(http.StatusOK, cluster)
}

func Delete(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	if isPrecheckEnforced(c, cli) {
		// Require precheck token to mitigate TOCTOU
		token := c.GetHeader("X-Precheck-Token")
		if !prechange.ValidatePrecheckToken(token, "scale", ns, name, 10*time.Minute) {
			c.JSON(http.StatusPreconditionFailed, gin.H{"error": "invalid or missing precheck token"})
			return
		}
	}
	if err := k8s.DeletePolarDBXClusterWithContext(c.Request.Context(), cli, ns, name); err != nil {
		util.HandleK8sError(c, "failed to delete cluster", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cluster deletion initiated successfully"})
}

func Update(c *gin.Context) {
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
	existing, err := k8s.GetPolarDBXClusterWithContext(c.Request.Context(), cli, ns, name)
	if err != nil {
		util.HandleK8sError(c, "cluster not found", err)
		return
	}
	body.SetResourceVersion(existing.GetResourceVersion())
	updated, err := k8s.UpdatePolarDBXClusterWithContext(c.Request.Context(), cli, ns, &body)
	if err != nil {
		util.HandleK8sError(c, "failed to update cluster", err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

// UpdateClusterLogConfig
type LogConfigRequest struct {
	EnableAuditLog   *bool  `json:"enableAuditLog,omitempty"`
	LogLevel         string `json:"logLevel,omitempty"`
	AuditLogFilter   string `json:"auditLogFilter,omitempty"`
	SlowLogThreshold *int   `json:"slowLogThreshold,omitempty"`
}

func UpdateLogConfig(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	clusterName := c.Param("name")
	nodeType := c.Param("nodeType")
	var req LogConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log config data", "details": err.Error()})
		return
	}
	patchData := map[string]any{"spec": map[string]any{"config": map[string]any{nodeType: map[string]any{}}}}
	nodeConfig := patchData["spec"].(map[string]any)["config"].(map[string]any)[nodeType].(map[string]any)
	if req.EnableAuditLog != nil {
		nodeConfig["enableAuditLog"] = *req.EnableAuditLog
	}
	if req.LogLevel != "" {
		nodeConfig["logLevel"] = req.LogLevel
	}
	if req.AuditLogFilter != "" {
		nodeConfig["auditLogFilter"] = req.AuditLogFilter
	}
	if req.SlowLogThreshold != nil {
		nodeConfig["slowLogThreshold"] = *req.SlowLogThreshold
	}
	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create patch data"})
		return
	}
	cluster, err := k8s.PatchPolarDBXClusterWithContext(c.Request.Context(), cli, ns, clusterName, patchBytes)
	if err != nil {
		util.HandleK8sError(c, "failed to update cluster log config", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": strings.ToUpper(nodeType) + " log config updated successfully", "cluster": cluster})
}

// ScaleCluster
type ClusterScalingRequest struct {
	CNReplicas  *int32 `json:"cnReplicas,omitempty"`
	DNReplicas  *int32 `json:"dnReplicas,omitempty"`
	GMSReplicas *int32 `json:"gmsReplicas,omitempty"`
	CDCReplicas *int32 `json:"cdcReplicas,omitempty"`
}

func Scale(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	var req ClusterScalingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scaling request", "details": err.Error()})
		return
	}
	if isPrecheckEnforced(c, cli) {
		// Require precheck token to mitigate TOCTOU
		token := c.GetHeader("X-Precheck-Token")
		sig := c.GetHeader("X-Precheck-Token-Signature")
		if !prechange.ValidatePrecheckToken(token, "scale", ns, name, 10*time.Minute) {
			c.JSON(http.StatusPreconditionFailed, gin.H{"error": "invalid or missing precheck token"})
			return
		}
		// Signature check if configured
		secret := prechange.GetPrecheckSecretForValidation(c)
		if secret != "" && !prechange.VerifyPrecheckTokenForValidation(secret, token, sig) {
			c.JSON(http.StatusPreconditionFailed, gin.H{"error": "invalid precheck token signature"})
			return
		}
	}
	patch := map[string]any{"spec": map[string]any{"topology": map[string]any{"nodes": map[string]any{}}}}
	nodes := patch["spec"].(map[string]any)["topology"].(map[string]any)["nodes"].(map[string]any)
	if req.CNReplicas != nil {
		nodes["cn"] = map[string]any{"replicas": *req.CNReplicas}
	}
	if req.DNReplicas != nil {
		nodes["dn"] = map[string]any{"replicas": *req.DNReplicas}
	}
	if req.CDCReplicas != nil {
		nodes["cdc"] = map[string]any{"replicas": *req.CDCReplicas}
	}
	if len(nodes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no replica changes specified"})
		return
	}
	b, _ := json.Marshal(patch)
	cluster, err := k8s.PatchPolarDBXClusterWithContext(c.Request.Context(), cli, ns, name, b)
	if err != nil {
		util.HandleK8sError(c, "failed to scale cluster", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cluster scaling initiated successfully", "cluster": cluster})
}

// UpgradeCluster
type ClusterUpgradeRequest struct {
	TargetVersion  string `json:"targetVersion" binding:"required"`
	Strategy       string `json:"strategy,omitempty"`
	MaxUnavailable *int32 `json:"maxUnavailable,omitempty"`
}

func Upgrade(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	var req ClusterUpgradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid upgrade request", "details": err.Error()})
		return
	}
	if isPrecheckEnforced(c, cli) {
		// Require precheck token and optional signature
		token := c.GetHeader("X-Precheck-Token")
		sig := c.GetHeader("X-Precheck-Token-Signature")
		if !prechange.ValidatePrecheckToken(token, "upgrade", ns, name, 10*time.Minute) {
			c.JSON(http.StatusPreconditionFailed, gin.H{"error": "invalid or missing precheck token"})
			return
		}
		secret := prechange.GetPrecheckSecretForValidation(c)
		if secret != "" && !prechange.VerifyPrecheckTokenForValidation(secret, token, sig) {
			c.JSON(http.StatusPreconditionFailed, gin.H{"error": "invalid precheck token signature"})
			return
		}
	}
	patch := map[string]any{"spec": map[string]any{"topology": map[string]any{"version": req.TargetVersion}}}
	if req.Strategy != "" {
		patch["spec"].(map[string]any)["upgradeStrategy"] = req.Strategy
	}
	b, _ := json.Marshal(patch)
	cluster, err := k8s.PatchPolarDBXClusterWithContext(c.Request.Context(), cli, ns, name, b)
	if err != nil {
		util.HandleK8sError(c, "failed to upgrade cluster", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cluster upgrade initiated successfully", "cluster": cluster, "upgrade": gin.H{"targetVersion": req.TargetVersion, "strategy": req.Strategy, "status": "升级已启动"}})
}

// Alerts summary
func GetAlertsSummary(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	amURL := c.Query("alertmanager")
	summary := map[string]any{"namespace": ns, "name": name, "critical": 0, "warning": 0, "info": 0, "total": 0, "source": "none"}
	if amURL != "" {
		type amAlert struct {
			Labels map[string]string `json:"labels"`
			Status map[string]any    `json:"status"`
		}
		var alerts []amAlert
		if resp, err := http.Get(amURL + "/api/v2/alerts"); err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			if err := json.NewDecoder(resp.Body).Decode(&alerts); err == nil {
				for _, a := range alerts {
					if a.Labels["namespace"] != ns || a.Labels["cluster"] != name {
						continue
					}
					switch strings.ToLower(a.Labels["severity"]) {
					case "critical":
						summary["critical"] = summary["critical"].(int) + 1
					case "warning":
						summary["warning"] = summary["warning"].(int) + 1
					default:
						summary["info"] = summary["info"].(int) + 1
					}
					summary["total"] = summary["total"].(int) + 1
				}
				summary["source"] = "alertmanager"
				c.JSON(http.StatusOK, summary)
				return
			}
		}
	}
	var evList corev1.EventList
	if err := cli.List(c.Request.Context(), &evList, client.InNamespace(ns)); err == nil {
		warn := 0
		for _, ev := range evList.Items {
			if strings.Contains(ev.InvolvedObject.Name, name) && strings.EqualFold(ev.Type, "Warning") {
				warn++
			}
		}
		summary["warning"], summary["total"], summary["source"] = warn, warn, "events"
	}
	c.JSON(http.StatusOK, summary)
}

// isPrecheckEnforced reads backend config to decide if precheck enforcement is enabled
func isPrecheckEnforced(c *gin.Context, cli client.Client) bool {
	cm := corev1.ConfigMap{}
	_ = cli.Get(c.Request.Context(), client.ObjectKey{Namespace: "polardbx-operator-system", Name: "polardbx-ui-backend-config"}, &cm)
	if cm.Data == nil {
		return false
	}
	return cm.Data["prechange.enforce"] == "true"
}
