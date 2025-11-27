package repository

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

// SystemRepository 定义 system 信息查询的存储层接口
type SystemRepository interface {
	// ListNamespaces 列出所有命名空间
	ListNamespaces(ctx context.Context) ([]corev1.Namespace, error)
}
