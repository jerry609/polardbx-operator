package repository

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

// SettingsRepository 定义设置存储层接口
type SettingsRepository interface {
	// GetConfigMap 获取指定的 ConfigMap
	GetConfigMap(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error)

	// CreateConfigMap 创建 ConfigMap
	CreateConfigMap(ctx context.Context, cm *corev1.ConfigMap) error

	// UpdateConfigMap 更新 ConfigMap
	UpdateConfigMap(ctx context.Context, cm *corev1.ConfigMap) error
}
