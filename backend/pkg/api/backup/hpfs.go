package backup

import (
	"net/http"
	"strings"

	"polardbx-ui-backend/pkg/api/util"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListHpfsSinks returns sinks configured in HPFS ConfigMap
func ListHpfsSinks(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	// Allow override of system namespace via query for tests and flexibility
	systemNS := c.DefaultQuery("systemNamespace", "polardbx-operator-system")
	cm, err := getHpfsConfigMap(c, cli, systemNS)
	if err != nil {
		util.HandleK8sError(c, "failed to get HPFS config ConfigMap", err)
		return
	}
	if _, exists := cm.Data["config.yaml"]; !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "config.yaml not found in ConfigMap"})
		return
	}
	cfg, err := decodeHpfsConfig(cm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse config.yaml", "details": err.Error()})
		return
	}
	type SinkDTO struct {
		Name             string `json:"name"`
		Type             string `json:"type"`
		Endpoint         string `json:"endpoint,omitempty"`
		Bucket           string `json:"bucket,omitempty"`
		BucketLookupType string `json:"bucketLookupType,omitempty"`
		Host             string `json:"host,omitempty"`
		Port             int    `json:"port,omitempty"`
		RootPath         string `json:"rootPath,omitempty"`
	}
	sinks := make([]SinkDTO, 0, len(cfg.Sinks))
	for _, s := range cfg.Sinks {
		dto := SinkDTO{
			Name:             s.Name,
			Type:             s.Type,
			Endpoint:         s.Endpoint,
			Bucket:           s.Bucket,
			BucketLookupType: s.MinioSink.BucketLookupType,
			Host:             s.SftpSink.Host,
			Port:             s.SftpSink.Port,
			RootPath:         s.SftpSink.RootPath,
		}
		sinks = append(sinks, dto)
	}
	c.JSON(http.StatusOK, gin.H{"namespace": systemNS, "configMap": "polardbx-hpfs-config", "sinks": sinks})
}

// ValidateHpfsSink validates that a given sink (name+type) exists in HPFS config
func ValidateHpfsSink(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	var req struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	systemNS := c.DefaultQuery("systemNamespace", "polardbx-operator-system")
	cm, err := getHpfsConfigMap(c, cli, systemNS)
	if err != nil {
		util.HandleK8sError(c, "failed to get HPFS config ConfigMap", err)
		return
	}
	if _, exists := cm.Data["config.yaml"]; !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "config.yaml not found in ConfigMap"})
		return
	}
	cfg, err := decodeHpfsConfig(cm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse config.yaml", "details": err.Error()})
		return
	}
	reqType := strings.ToLower(strings.TrimSpace(req.Type))
	reqName := strings.TrimSpace(req.Name)
	status := "not_found"
	message := "sink not found"
	for _, s := range cfg.Sinks {
		if s.Name == reqName && strings.EqualFold(s.Type, reqType) {
			status = "ok"
			message = "sink exists"
			break
		}
	}
	c.JSON(http.StatusOK, gin.H{"name": reqName, "type": reqType, "status": status, "message": message})
}

// GetBackupAdvice returns suggested backup role based on cluster topology (whether followers exist for all XStores)
func GetBackupAdvice(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}
	namespace := c.Param("namespace")
	clusterName := c.Param("name")
	var xstoreList polardbxv1.XStoreList
	if err := cli.List(c.Request.Context(), &xstoreList, client.InNamespace(namespace), client.MatchingLabels{
		"polardbx/name": clusterName,
	}); err != nil {
		util.HandleK8sError(c, "failed to list xstores for cluster", err)
		return
	}
	if len(xstoreList.Items) == 0 {
		c.JSON(http.StatusOK, gin.H{"hasFollower": false, "role": "leader", "reason": "no xstores found for cluster"})
		return
	}
	namesWithoutFollower := make([]string, 0)
	for _, xs := range xstoreList.Items {
		if xs.Status.TotalPods <= 1 {
			namesWithoutFollower = append(namesWithoutFollower, xs.Name)
		}
	}
	if len(namesWithoutFollower) == 0 {
		c.JSON(http.StatusOK, gin.H{"hasFollower": true, "role": "follower"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"hasFollower": false, "role": "leader", "reason": namesWithoutFollower})
}
