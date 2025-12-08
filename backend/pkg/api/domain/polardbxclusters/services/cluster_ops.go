package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	apierr "polardbx-ui-backend/pkg/api/errors"
	"polardbx-ui-backend/pkg/api/util"

	"github.com/gin-gonic/gin"
)

// patchClusterJSON is a test-hookable wrapper around util.K8sPatchClusterJSON.
var patchClusterJSON = util.K8sPatchClusterJSON

// ClusterService 聚合与集群相关的编排。
type ClusterService struct{}

func NewClusterService() *ClusterService { return &ClusterService{} }

// UpdateLogConfig：最小实现，构造 JSON Patch 并调用 K8s。
func (s *ClusterService) UpdateLogConfig(ctx context.Context, c *gin.Context) error {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return nil
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	nodeType := c.Param("nodeType")
	start := time.Now()
	log.Printf("ops UpdateLogConfig begin: %s/%s nodeType=%s", ns, name, nodeType)
	switch nodeType {
	case "cn", "dn", "gms", "cdc":
	default:
		apierr.AbortValidation(c, "invalid nodeType: "+nodeType)
		return nil
	}
	var req LogConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.AbortValidation(c, "invalid log config data: "+err.Error())
		return nil
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
	b, _ := json.Marshal(patchData)
	if _, err := patchClusterJSON(ctx, cli, ns, name, b); err != nil {
		log.Printf("ops UpdateLogConfig failed: %s duration=%s", err, time.Since(start))
		util.HandleK8sError(c, "failed to update cluster log config", err)
		return nil
	}
	log.Printf("ops UpdateLogConfig ok duration=%s", time.Since(start))
	apierr.OK(c, gin.H{"message": nodeType + " log config updated successfully"})
	return nil
}

// Scale：最小实现，构造 replicas JSON Patch。
func (s *ClusterService) Scale(ctx context.Context, c *gin.Context) error {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return nil
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	start := time.Now()
	log.Printf("ops Scale begin: %s/%s", ns, name)
	var req ClusterScalingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.AbortValidation(c, "invalid scaling request: "+err.Error())
		return nil
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
		apierr.AbortValidation(c, "no replica changes specified")
		return nil
	}
	b, _ := json.Marshal(patch)
	if _, err := patchClusterJSON(ctx, cli, ns, name, b); err != nil {
		log.Printf("ops Scale failed: %s duration=%s", err, time.Since(start))
		util.HandleK8sError(c, "failed to scale cluster", err)
		return nil
	}
	log.Printf("ops Scale ok duration=%s", time.Since(start))
	apierr.OK(c, gin.H{"message": "Cluster scaling initiated successfully"})
	return nil
}

// Upgrade：最小实现，设置目标版本与可选策略。
func (s *ClusterService) Upgrade(ctx context.Context, c *gin.Context) error {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return nil
	}
	ns := c.Param("namespace")
	name := c.Param("name")
	start := time.Now()
	log.Printf("ops Upgrade begin: %s/%s", ns, name)
	var req ClusterUpgradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.AbortValidation(c, "invalid upgrade request: "+err.Error())
		return nil
	}
	patch := map[string]any{"spec": map[string]any{"topology": map[string]any{"version": req.TargetVersion}}}
	if req.Strategy != "" {
		patch["spec"].(map[string]any)["upgradeStrategy"] = req.Strategy
	}
	b, _ := json.Marshal(patch)
	if _, err := patchClusterJSON(ctx, cli, ns, name, b); err != nil {
		log.Printf("ops Upgrade failed: %s duration=%s", err, time.Since(start))
		util.HandleK8sError(c, "failed to upgrade cluster", err)
		return nil
	}
	log.Printf("ops Upgrade ok duration=%s", time.Since(start))
	apierr.OK(c, gin.H{"message": "Cluster upgrade initiated successfully", "upgrade": gin.H{"targetVersion": req.TargetVersion, "strategy": req.Strategy, "status": "升级已启动"}})
	return nil
}
