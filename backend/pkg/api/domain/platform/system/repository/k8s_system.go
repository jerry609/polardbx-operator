package repository

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// K8sSystemRepository 使用 Kubernetes client 实现 SystemRepository
type K8sSystemRepository struct {
	client client.Client
}

// NewK8sSystemRepository 创建新的 K8s 实现
func NewK8sSystemRepository(cli client.Client) *K8sSystemRepository {
	return &K8sSystemRepository{client: cli}
}

// ListNamespaces 列出所有命名空间
func (r *K8sSystemRepository) ListNamespaces(ctx context.Context) ([]corev1.Namespace, error) {
	var nsList corev1.NamespaceList
	if err := r.client.List(ctx, &nsList, &client.ListOptions{}); err != nil {
		return nil, err
	}
	return nsList.Items, nil
}

// ListStorageClasses 列出所有存储类
func (r *K8sSystemRepository) ListStorageClasses(ctx context.Context) ([]storagev1.StorageClass, error) {
	var scList storagev1.StorageClassList
	if err := r.client.List(ctx, &scList, &client.ListOptions{}); err != nil {
		return nil, err
	}
	return scList.Items, nil
}
