package repository

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

// AlertsRepository 定义告警存储层接口
type AlertsRepository interface {
	// GetConfigMap 获取指定的 ConfigMap
	GetConfigMap(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error)

	// CreateConfigMap 创建 ConfigMap
	CreateConfigMap(ctx context.Context, cm *corev1.ConfigMap) error

	// UpdateConfigMap 更新 ConfigMap
	UpdateConfigMap(ctx context.Context, cm *corev1.ConfigMap) error

	// ListEvents 列出事件
	ListEvents(ctx context.Context, namespace string) ([]corev1.Event, error)
}
