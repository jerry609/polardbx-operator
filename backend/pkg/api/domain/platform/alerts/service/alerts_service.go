package service

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"polardbx-ui-backend/pkg/api/domain/platform/alerts/repository"
)

const (
	SettingsNS      = "polardbx-operator-system"
	CMAlertProfiles = "polardbx-alert-profiles"
	CMAlertRoutes   = "polardbx-alert-routes"
)

// ProfileInfo 配置文件信息
type ProfileInfo struct {
	Name string `json:"name"`
}

// ProfileDetail 配置文件详情
type ProfileDetail struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// AlertItem 告警条目
type AlertItem struct {
	Source    string            `json:"source"`
	Severity  string            `json:"severity"`
	Labels    map[string]string `json:"labels"`
	Message   string            `json:"message"`
	Timestamp string            `json:"timestamp"`
	Time      string            `json:"time,omitempty"`
}

// DryRunResult 干运行结果
type DryRunResult struct {
	Valid   bool   `json:"valid"`
	Details string `json:"details,omitempty"`
}

// AlertsService 定义告警业务逻辑层
type AlertsService struct {
	repo repository.AlertsRepository
}

// NewAlertsService 创建新的 AlertsService
func NewAlertsService(repo repository.AlertsRepository) *AlertsService {
	return &AlertsService{repo: repo}
}

// ListProfiles 列出所有配置文件
func (s *AlertsService) ListProfiles(ctx context.Context) ([]ProfileInfo, error) {
	cm, err := s.repo.GetConfigMap(ctx, SettingsNS, CMAlertProfiles)
	if err != nil {
		return []ProfileInfo{}, nil
	}
	items := make([]ProfileInfo, 0, len(cm.Data))
	for k := range cm.Data {
		items = append(items, ProfileInfo{Name: k})
	}
	return items, nil
}

// CreateProfile 创建配置文件
func (s *AlertsService) CreateProfile(ctx context.Context, name, content string) error {
	cm, err := s.repo.GetConfigMap(ctx, SettingsNS, CMAlertProfiles)
	if err != nil {
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Namespace: SettingsNS, Name: CMAlertProfiles},
			Data:       map[string]string{},
		}
		if err2 := s.repo.CreateConfigMap(ctx, cm); err2 != nil {
			return err2
		}
	}
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	if _, exists := cm.Data[name]; exists {
		return ErrProfileExists
	}
	cm.Data[name] = content
	return s.repo.UpdateConfigMap(ctx, cm)
}

// GetProfile 获取配置文件
func (s *AlertsService) GetProfile(ctx context.Context, name string) (*ProfileDetail, error) {
	cm, err := s.repo.GetConfigMap(ctx, SettingsNS, CMAlertProfiles)
	if err != nil {
		return nil, ErrProfileNotFound
	}
	content, ok := cm.Data[name]
	if !ok {
		return nil, ErrProfileNotFound
	}
	return &ProfileDetail{Name: name, Content: content}, nil
}

// UpdateProfile 更新配置文件
func (s *AlertsService) UpdateProfile(ctx context.Context, name, content string) error {
	cm, err := s.repo.GetConfigMap(ctx, SettingsNS, CMAlertProfiles)
	if err != nil {
		return ErrProfileNotFound
	}
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	cm.Data[name] = content
	return s.repo.UpdateConfigMap(ctx, cm)
}

// DeleteProfile 删除配置文件
func (s *AlertsService) DeleteProfile(ctx context.Context, name string) error {
	cm, err := s.repo.GetConfigMap(ctx, SettingsNS, CMAlertProfiles)
	if err != nil {
		return ErrProfileNotFound
	}
	if cm.Data == nil || cm.Data[name] == "" {
		return ErrProfileNotFound
	}
	delete(cm.Data, name)
	return s.repo.UpdateConfigMap(ctx, cm)
}

// DryRunProfile 验证 Alertmanager YAML
func (s *AlertsService) DryRunProfile(content string) (*DryRunResult, error) {
	f, err := os.CreateTemp("", "am-profile-*.yaml")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	_, _ = f.Write([]byte(content))
	_ = f.Close()
	cmd := exec.Command("promtool", "check", "rules", f.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &DryRunResult{Valid: false, Details: string(out)}, nil
	}
	return &DryRunResult{Valid: true}, nil
}

// GetRoutes 获取路由配置
func (s *AlertsService) GetRoutes(ctx context.Context) (string, error) {
	cm, _ := s.repo.GetConfigMap(ctx, SettingsNS, CMAlertRoutes)
	if cm == nil {
		return "", nil
	}
	return cm.Data["config.yaml"], nil
}

// PutRoutes 更新路由配置
func (s *AlertsService) PutRoutes(ctx context.Context, content string) error {
	cm, err := s.repo.GetConfigMap(ctx, SettingsNS, CMAlertRoutes)
	if err != nil {
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Namespace: SettingsNS, Name: CMAlertRoutes},
			Data:       map[string]string{"config.yaml": content},
		}
		return s.repo.CreateConfigMap(ctx, cm)
	}
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	cm.Data["config.yaml"] = content
	return s.repo.UpdateConfigMap(ctx, cm)
}

// ListAlerts 聚合 Alertmanager 和 K8s 事件的告警列表
func (s *AlertsService) ListAlerts(ctx context.Context, namespace, cluster, alertmanagerURL string) ([]AlertItem, error) {
	items := make([]AlertItem, 0)

	// 从 Alertmanager 获取
	if alertmanagerURL != "" {
		items = append(items, s.fetchAlertsFromAlertmanager(alertmanagerURL, namespace, cluster)...)
	}

	// 从 K8s 事件获取
	events, err := s.repo.ListEvents(ctx, namespace)
	if err == nil {
		for _, ev := range events {
			if cluster != "" && !strings.Contains(ev.InvolvedObject.Name, cluster) {
				continue
			}
			sev := "info"
			if ev.Type == corev1.EventTypeWarning {
				sev = "warning"
			}
			labels := map[string]string{
				"namespace":      ev.Namespace,
				"reason":         ev.Reason,
				"involvedObject": ev.InvolvedObject.Name,
			}
			if cluster != "" {
				labels["cluster"] = cluster
			}
			ts := ev.LastTimestamp.Time
			if ts.IsZero() && !ev.EventTime.IsZero() {
				ts = ev.EventTime.Time
			}
			if ts.IsZero() {
				ts = ev.ObjectMeta.CreationTimestamp.Time
			}
			tsStr := ts.Format(time.RFC3339)
			items = append(items, AlertItem{
				Source:    "k8s-event",
				Severity:  sev,
				Message:   ev.Message,
				Labels:    labels,
				Time:      tsStr,
				Timestamp: tsStr,
			})
		}
	}

	return items, nil
}

func (s *AlertsService) fetchAlertsFromAlertmanager(url, namespace, cluster string) []AlertItem {
	type amAlert struct {
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		StartsAt    string            `json:"startsAt"`
	}
	items := make([]AlertItem, 0)
	resp, err := http.Get(url + "/api/v2/alerts")
	if err != nil || resp.StatusCode != 200 {
		return items
	}
	defer resp.Body.Close()
	var alerts []amAlert
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		return items
	}
	for _, a := range alerts {
		if (namespace == "" || a.Labels["namespace"] == namespace) && (cluster == "" || a.Labels["cluster"] == cluster) {
			msg := a.Annotations["summary"]
			if msg == "" {
				msg = a.Annotations["description"]
			}
			if msg == "" {
				msg = a.Labels["alertname"]
			}
			items = append(items, AlertItem{
				Source:    "alertmanager",
				Severity:  a.Labels["severity"],
				Labels:    a.Labels,
				Message:   msg,
				Timestamp: a.StartsAt,
			})
		}
	}
	return items
}

// 错误定义
type AlertsError string

func (e AlertsError) Error() string { return string(e) }

const (
	ErrProfileExists   AlertsError = "profile already exists"
	ErrProfileNotFound AlertsError = "profile not found"
)
