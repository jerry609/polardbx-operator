package repository

import (
	"context"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RestoreRepository 定义恢复操作的存储层接口
type RestoreRepository interface {
	// GetCluster 获取集群
	GetCluster(ctx context.Context, namespace, name string) (*polardbxv1.PolarDBXCluster, error)

	// CreateCluster 创建集群
	CreateCluster(ctx context.Context, cluster *polardbxv1.PolarDBXCluster) error

	// ListClusters 列出集群
	ListClusters(ctx context.Context, namespace string) ([]polardbxv1.PolarDBXCluster, error)

	// GetBackup 获取备份
	GetBackup(ctx context.Context, namespace, name string) (*polardbxv1.PolarDBXBackup, error)
}

// K8sRestoreRepository 使用 Kubernetes client 实现 RestoreRepository
type K8sRestoreRepository struct {
	client client.Client
}

// NewK8sRestoreRepository 创建新的 K8s 实现
func NewK8sRestoreRepository(cli client.Client) *K8sRestoreRepository {
	return &K8sRestoreRepository{client: cli}
}

// GetCluster 获取集群
func (r *K8sRestoreRepository) GetCluster(ctx context.Context, namespace, name string) (*polardbxv1.PolarDBXCluster, error) {
	var cluster polardbxv1.PolarDBXCluster
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &cluster); err != nil {
		return nil, err
	}
	return &cluster, nil
}

// CreateCluster 创建集群
func (r *K8sRestoreRepository) CreateCluster(ctx context.Context, cluster *polardbxv1.PolarDBXCluster) error {
	return r.client.Create(ctx, cluster)
}

// ListClusters 列出集群
func (r *K8sRestoreRepository) ListClusters(ctx context.Context, namespace string) ([]polardbxv1.PolarDBXCluster, error) {
	var clusterList polardbxv1.PolarDBXClusterList
	opts := []client.ListOption{}
	if namespace != "" {
		opts = append(opts, client.InNamespace(namespace))
	}
	if err := r.client.List(ctx, &clusterList, opts...); err != nil {
		return nil, err
	}
	return clusterList.Items, nil
}

// GetBackup 获取备份
func (r *K8sRestoreRepository) GetBackup(ctx context.Context, namespace, name string) (*polardbxv1.PolarDBXBackup, error) {
	var backup polardbxv1.PolarDBXBackup
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &backup); err != nil {
		return nil, err
	}
	return &backup, nil
}
