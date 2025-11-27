package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// SettingsNamespace Grafana 配置所在命名空间
	SettingsNamespace = "polardbx-operator-system"
	// ConfigMapGrafanaConfig Grafana 配置 ConfigMap 名称
	ConfigMapGrafanaConfig = "polardbx-grafana-config"
	// ConfigMapDashboards Grafana 仪表盘 ConfigMap 名称
	ConfigMapDashboards = "polardbx-grafana-dashboards"
	// SecretGrafana Grafana 认证 Secret 名称
	SecretGrafana = "polardbx-grafana-secret"
)

// GrafanaConfig Grafana 配置
type GrafanaConfig struct {
	URL           string         `json:"url"`
	Auth          string         `json:"auth"` // sso|basic|token
	ApiKey        string         `json:"apiKey"`
	BasicUser     string         `json:"basicUser"`
	BasicPassword string         `json:"basicPassword"`
	DataSource    map[string]any `json:"dataSource"`
}

// GrafanaRepository 定义 Grafana 配置的数据访问接口
type GrafanaRepository interface {
	// GetConfig 获取 Grafana 配置
	GetConfig(ctx context.Context) (*GrafanaConfig, error)
	// SaveConfig 保存 Grafana 配置
	SaveConfig(ctx context.Context, config *GrafanaConfig) error
	// GetDashboards 获取仪表盘数据
	GetDashboards(ctx context.Context) (map[string]string, error)
	// SaveDashboards 保存仪表盘数据
	SaveDashboards(ctx context.Context, dashboards map[string]string, overwrite bool) (int, error)
}

// K8sGrafanaRepository K8s 实现
type K8sGrafanaRepository struct {
	client client.Client
}

// NewK8sGrafanaRepository 创建 K8s Grafana Repository
func NewK8sGrafanaRepository(cli client.Client) *K8sGrafanaRepository {
	return &K8sGrafanaRepository{client: cli}
}

// GetConfig 获取 Grafana 配置
func (r *K8sGrafanaRepository) GetConfig(ctx context.Context) (*GrafanaConfig, error) {
	cm := corev1.ConfigMap{}
	_ = r.client.Get(ctx, client.ObjectKey{Namespace: SettingsNamespace, Name: ConfigMapGrafanaConfig}, &cm)

	sec := corev1.Secret{}
	_ = r.client.Get(ctx, client.ObjectKey{Namespace: SettingsNamespace, Name: SecretGrafana}, &sec)

	config := &GrafanaConfig{
		URL:  cm.Data["url"],
		Auth: cm.Data["auth"],
	}

	if v, ok := sec.Data["apiKey"]; ok {
		config.ApiKey = string(v)
	}
	if v, ok := sec.Data["basicUser"]; ok {
		config.BasicUser = string(v)
	}
	if v, ok := sec.Data["basicPassword"]; ok {
		config.BasicPassword = string(v)
	}
	if v, ok := cm.Data["datasource.json"]; ok {
		config.DataSource = map[string]any{"raw": v}
	}

	return config, nil
}

// SaveConfig 保存 Grafana 配置
func (r *K8sGrafanaRepository) SaveConfig(ctx context.Context, config *GrafanaConfig) error {
	// 更新或创建 ConfigMap
	cm := corev1.ConfigMap{}
	key := client.ObjectKey{Namespace: SettingsNamespace, Name: ConfigMapGrafanaConfig}
	err := r.client.Get(ctx, key, &cm)
	if err != nil {
		cm = corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Namespace: SettingsNamespace, Name: ConfigMapGrafanaConfig},
			Data:       map[string]string{},
		}
		if err2 := r.client.Create(ctx, &cm); err2 != nil {
			return fmt.Errorf("failed to create grafana config: %w", err2)
		}
	}

	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	cm.Data["url"] = config.URL
	cm.Data["auth"] = config.Auth
	if raw, ok := config.DataSource["raw"]; ok {
		cm.Data["datasource.json"] = fmt.Sprintf("%v", raw)
	}

	if err := r.client.Update(ctx, &cm); err != nil {
		return fmt.Errorf("failed to update grafana config: %w", err)
	}

	// 更新或创建 Secret
	sec := corev1.Secret{}
	skey := client.ObjectKey{Namespace: SettingsNamespace, Name: SecretGrafana}
	if err := r.client.Get(ctx, skey, &sec); err != nil {
		sec = corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Namespace: SettingsNamespace, Name: SecretGrafana},
			Type:       corev1.SecretTypeOpaque,
			Data:       map[string][]byte{},
		}
		if err2 := r.client.Create(ctx, &sec); err2 != nil {
			return fmt.Errorf("failed to create grafana secret: %w", err2)
		}
	}

	if sec.Data == nil {
		sec.Data = map[string][]byte{}
	}
	if config.ApiKey != "" {
		sec.Data["apiKey"] = []byte(config.ApiKey)
	}
	if config.BasicUser != "" {
		sec.Data["basicUser"] = []byte(config.BasicUser)
	}
	if config.BasicPassword != "" {
		sec.Data["basicPassword"] = []byte(config.BasicPassword)
	}

	if err := r.client.Update(ctx, &sec); err != nil {
		return fmt.Errorf("failed to update grafana secret: %w", err)
	}

	return nil
}

// GetDashboards 获取仪表盘数据
func (r *K8sGrafanaRepository) GetDashboards(ctx context.Context) (map[string]string, error) {
	cm := corev1.ConfigMap{}
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: SettingsNamespace, Name: ConfigMapDashboards}, &cm); err != nil {
		return map[string]string{}, nil
	}
	if cm.Data == nil {
		return map[string]string{}, nil
	}
	return cm.Data, nil
}

// SaveDashboards 保存仪表盘数据
func (r *K8sGrafanaRepository) SaveDashboards(ctx context.Context, dashboards map[string]string, overwrite bool) (int, error) {
	if len(dashboards) == 0 {
		return 0, nil
	}

	cm := corev1.ConfigMap{}
	key := client.ObjectKey{Namespace: SettingsNamespace, Name: ConfigMapDashboards}
	err := r.client.Get(ctx, key, &cm)
	if err != nil {
		cm = corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Namespace: SettingsNamespace, Name: ConfigMapDashboards},
			Data:       map[string]string{},
		}
		if err2 := r.client.Create(ctx, &cm); err2 != nil {
			return 0, fmt.Errorf("failed to create dashboards: %w", err2)
		}
	}

	if cm.Data == nil {
		cm.Data = map[string]string{}
	}

	for k, v := range dashboards {
		if old, exists := cm.Data[k]; exists {
			if !overwrite && old != "" {
				continue
			}
			if old != v {
				next := nextVersion(cm.Data, k)
				cm.Data[fmt.Sprintf("%s@v%04d", k, next)] = old
			}
		}
		cm.Data[k] = v
	}

	if err := r.client.Update(ctx, &cm); err != nil {
		return 0, fmt.Errorf("failed to update dashboards: %w", err)
	}

	return len(dashboards), nil
}

// nextVersion 计算下一个版本号
func nextVersion(data map[string]string, name string) int {
	maxv := 0
	prefix := name + "@v"
	for k := range data {
		if strings.HasPrefix(k, prefix) {
			if n, err := strconv.Atoi(strings.TrimPrefix(k, prefix)); err == nil && n > maxv {
				maxv = n
			}
		}
	}
	return maxv + 1
}

// SplitVersionKey 分割版本键
func SplitVersionKey(key string) (string, int, bool) {
	idx := strings.LastIndex(key, "@v")
	if idx <= 0 {
		return "", 0, false
	}
	name := key[:idx]
	vn := key[idx+2:]
	n, err := strconv.Atoi(vn)
	if err != nil {
		return "", 0, false
	}
	return name, n, true
}
