package service

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"polardbx-ui-backend/pkg/api/domain/platform/pod/repository"
)

// PodService 定义 Pod 业务逻辑层
type PodService struct {
	repo repository.PodRepository
}

// NewPodService 创建新的 PodService
func NewPodService(repo repository.PodRepository) *PodService {
	return &PodService{repo: repo}
}

// List 列出指定命名空间的所有 Pod
func (s *PodService) List(ctx context.Context, namespace string) ([]corev1.Pod, error) {
	return s.repo.List(ctx, namespace)
}

// ListForCluster 列出 PolarDBX 集群的所有 Pod
func (s *PodService) ListForCluster(ctx context.Context, namespace, clusterName string) ([]corev1.Pod, error) {
	return s.repo.ListForCluster(ctx, namespace, clusterName)
}

// Get 获取指定的 Pod
func (s *PodService) Get(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	return s.repo.Get(ctx, namespace, name)
}

// Delete 删除指定的 Pod
func (s *PodService) Delete(ctx context.Context, namespace, name string) error {
	return s.repo.Delete(ctx, namespace, name)
}

// GetLogs 获取 Pod 日志
func (s *PodService) GetLogs(ctx context.Context, namespace, podName, container string, tailLines int64) (string, error) {
	return s.repo.GetLogs(ctx, namespace, podName, container, tailLines)
}
