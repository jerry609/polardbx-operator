package polardbxclusters

import (
	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"polardbx-ui-backend/pkg/api/domain/platform/pod/handler"
	domain_prechange "polardbx-ui-backend/pkg/api/domain/platform/prechange/handler"
	domain_restore "polardbx-ui-backend/pkg/api/domain/platform/restore/handler"
	"polardbx-ui-backend/pkg/api/domain/polardbxclusters/services"
	apierr "polardbx-ui-backend/pkg/api/errors"
	"polardbx-ui-backend/pkg/api/util"
)

// --- Thin handlers forwarding to services/others ---

func UpdateLogConfig(c *gin.Context) {
	_ = services.NewClusterService().UpdateLogConfig(c.Request.Context(), c)
}
func Scale(c *gin.Context)            { _ = services.NewClusterService().Scale(c.Request.Context(), c) }
func Upgrade(c *gin.Context)          { _ = services.NewClusterService().Upgrade(c.Request.Context(), c) }
func GetAlertsSummary(c *gin.Context) { services.GetAlertsSummary(c) }

func ListPods(c *gin.Context) { handler.ListForCluster(c) }

func GetPrechangeChecklist(c *gin.Context) { domain_prechange.GetPrechangeChecklist(c) }
func Precheck(c *gin.Context)              { domain_prechange.Precheck(c) }

func RestoreCluster(c *gin.Context)   { domain_restore.RestoreCluster(c) }
func InitiatePITR(c *gin.Context)     { domain_restore.InitiatePITR(c) }
func GetRestoreStatus(c *gin.Context) { domain_restore.GetRestoreStatus(c) }

// UpgradeCandidate upgrade candidate version
type UpgradeCandidate struct {
	Version     string `json:"version"`
	Recommended bool   `json:"recommended,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// UpgradePlanResponse upgrade plan response
type UpgradePlanResponse struct {
	CurrentVersion string             `json:"currentVersion"`
	Candidates     []UpgradeCandidate `json:"candidates"`
	Matrix         map[string]any     `json:"matrix,omitempty"`
}

// GetUpgradePlan gets cluster upgrade plan (candidate version list)
func GetUpgradePlan(c *gin.Context) {
	cli, ok := util.K8sClientFromContext(c)
	if !ok {
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// Get current cluster information
	var cluster polardbxv1.PolarDBXCluster
	if err := cli.Get(c.Request.Context(), client.ObjectKey{Namespace: namespace, Name: name}, &cluster); err != nil {
		apierr.AbortNotFound(c, "cluster", name)
		return
	}

	currentVersion := cluster.Spec.Topology.Version
	if currentVersion == "" {
		currentVersion = "unknown"
	}

	// Build candidate version list
	// In production environment, available versions can be obtained from ConfigMap or external service
	candidates := buildUpgradeCandidates(currentVersion)

	resp := UpgradePlanResponse{
		CurrentVersion: currentVersion,
		Candidates:     candidates,
		Matrix: map[string]any{
			"minVersion": "5.4.13",
			"maxVersion": "5.4.19",
		},
	}

	apierr.OK(c, resp)
}

// buildUpgradeCandidates builds upgrade candidate list
func buildUpgradeCandidates(currentVersion string) []UpgradeCandidate {
	// Predefined version list (can be read from ConfigMap in practice)
	allVersions := []string{"5.4.19", "5.4.18", "5.4.17", "5.4.16", "5.4.15", "5.4.14", "5.4.13"}

	var candidates []UpgradeCandidate
	for i, v := range allVersions {
		// Skip current version and lower versions
		if v == currentVersion {
			break
		}
		candidate := UpgradeCandidate{
			Version: v,
		}
		if i == 0 {
			candidate.Recommended = true
			candidate.Notes = "Latest stable version"
		}
		candidates = append(candidates, candidate)
	}

	// If no higher versions, return default candidates
	if len(candidates) == 0 {
		candidates = []UpgradeCandidate{
			{Version: "5.4.19", Recommended: true, Notes: "Latest stable version"},
			{Version: "5.4.18"},
		}
	}

	return candidates
}
