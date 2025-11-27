package service

import (
	"context"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"polardbx-ui-backend/pkg/api/domain/platform/settings/repository"
	"polardbx-ui-backend/pkg/config"
)

const (
	SettingsNamespace = "polardbx-operator-system"
	SettingsConfigMap = "polardbx-ui-backend-config"
)

// BackupDashboardSettings 备份仪表盘设置
type BackupDashboardSettings struct {
	RPOThresholdSeconds                int     `json:"rpoThresholdSeconds"`
	ThroughputLowerBoundMBps           float64 `json:"throughputLowerBoundMBps"`
	DiagnosisRetentionDays             int     `json:"diagnosisRetentionDays"`
	AutoRebuildLagThresholdSeconds     int     `json:"autoRebuildLagThresholdSeconds"`
	AutoRebuildPreferredNodeLabelKey   string  `json:"autoRebuildPreferredNodeLabelKey"`
	AutoRebuildPreferredNodeLabelValue string  `json:"autoRebuildPreferredNodeLabelValue"`
}

// SettingsService 定义设置业务逻辑层
type SettingsService struct {
	repo repository.SettingsRepository
}

// NewSettingsService 创建新的 SettingsService
func NewSettingsService(repo repository.SettingsRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

// Get 获取所有设置
func (s *SettingsService) Get(ctx context.Context) (map[string]string, error) {
	cm, _ := s.repo.GetConfigMap(ctx, SettingsNamespace, SettingsConfigMap)
	if cm == nil {
		return map[string]string{}, nil
	}
	return cm.Data, nil
}

// Update 更新设置
func (s *SettingsService) Update(ctx context.Context, body map[string]any) error {
	cm, err := s.repo.GetConfigMap(ctx, SettingsNamespace, SettingsConfigMap)
	if err != nil {
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Namespace: SettingsNamespace, Name: SettingsConfigMap},
			Data:       map[string]string{},
		}
		if err2 := s.repo.CreateConfigMap(ctx, cm); err2 != nil {
			return err2
		}
	}
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	for k, v := range body {
		cm.Data[k] = fmt.Sprintf("%v", v)
	}
	return s.repo.UpdateConfigMap(ctx, cm)
}

// GetDashboardSettings 获取备份仪表盘设置
func (s *SettingsService) GetDashboardSettings(ctx context.Context) BackupDashboardSettings {
	settings := defaultSettings()
	cm, err := s.repo.GetConfigMap(ctx, SettingsNamespace, SettingsConfigMap)
	if err != nil || cm == nil || cm.Data == nil {
		return settings
	}
	settings.RPOThresholdSeconds = parseInt(cm.Data["rpoThresholdSeconds"], settings.RPOThresholdSeconds)
	settings.ThroughputLowerBoundMBps = parseFloat(cm.Data["throughputLowerBoundMBps"], settings.ThroughputLowerBoundMBps)
	settings.DiagnosisRetentionDays = parseInt(cm.Data["diagnosisRetentionDays"], settings.DiagnosisRetentionDays)
	settings.AutoRebuildLagThresholdSeconds = parseInt(cm.Data["autoRebuildLagThresholdSeconds"], settings.AutoRebuildLagThresholdSeconds)
	if v, ok := cm.Data["autoRebuildPreferredNodeLabelKey"]; ok {
		settings.AutoRebuildPreferredNodeLabelKey = v
	}
	if v, ok := cm.Data["autoRebuildPreferredNodeLabelValue"]; ok {
		settings.AutoRebuildPreferredNodeLabelValue = v
	}
	return settings
}

// GetImageRegistryConfig 获取镜像仓库配置
func (s *SettingsService) GetImageRegistryConfig() map[string]any {
	cfg := config.GetGlobalConfig()
	return cfg.ToMap()
}

// UpdateImageRegistryConfig 更新镜像仓库配置
func (s *SettingsService) UpdateImageRegistryConfig(registry, customRegistry, defaultRegistry string) (map[string]any, error) {
	cfg := config.GetGlobalConfig()

	reg := registry
	if reg == "" {
		reg = defaultRegistry
	}
	if reg == "custom" && customRegistry != "" {
		reg = customRegistry
	}
	if reg == "" {
		return nil, ErrEmptyRegistry
	}

	cfg.SetDefaultRegistry(reg)
	return cfg.ToMap(), nil
}

// GetAvailableRegistries 获取可用的镜像仓库预设
func (s *SettingsService) GetAvailableRegistries() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "DaoCloud Mirror (推荐)",
			"registry":    "docker.m.daocloud.io",
			"description": "DaoCloud 公共镜像加速服务，完整代理 Docker Hub",
			"region":      "China",
			"status":      "verified",
		},
		{
			"name":        "Docker Hub (官方)",
			"registry":    "docker.io",
			"description": "官方 Docker Hub 镜像仓库 (docker.io)",
			"region":      "Global",
			"status":      "slow",
		},
		{
			"name":        "自定义镜像仓库",
			"registry":    "custom",
			"description": "使用企业私有镜像仓库（如 Harbor），需提前同步 alpine/helm:3.12.3 镜像",
			"region":      "Custom",
			"status":      "custom",
		},
	}
}

func defaultSettings() BackupDashboardSettings {
	return BackupDashboardSettings{
		RPOThresholdSeconds:                3600,
		ThroughputLowerBoundMBps:           1.0,
		DiagnosisRetentionDays:             7,
		AutoRebuildLagThresholdSeconds:     0,
		AutoRebuildPreferredNodeLabelKey:   "",
		AutoRebuildPreferredNodeLabelValue: "",
	}
}

func parseInt(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

func parseFloat(s string, def float64) float64 {
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	return def
}

// 错误定义
type SettingsError string

func (e SettingsError) Error() string { return string(e) }

const (
	ErrEmptyRegistry SettingsError = "registry cannot be empty"
)
