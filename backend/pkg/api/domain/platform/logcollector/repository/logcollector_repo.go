package repository

import (
	"context"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LogCollectorRepository 定义日志收集器的数据访问接口
type LogCollectorRepository interface {
	// List 列出指定命名空间的日志收集器
	List(ctx context.Context, namespace string) (*polardbxv1.PolarDBXLogCollectorList, error)
	// Get 获取指定的日志收集器
	Get(ctx context.Context, namespace, name string) (*polardbxv1.PolarDBXLogCollector, error)
	// Create 创建日志收集器
	Create(ctx context.Context, collector *polardbxv1.PolarDBXLogCollector) (*polardbxv1.PolarDBXLogCollector, error)
	// Update 更新日志收集器
	Update(ctx context.Context, collector *polardbxv1.PolarDBXLogCollector) (*polardbxv1.PolarDBXLogCollector, error)
	// Delete 删除日志收集器
	Delete(ctx context.Context, namespace, name string) error
}

// K8sLogCollectorRepository 使用 Kubernetes client 实现 LogCollectorRepository
type K8sLogCollectorRepository struct {
	client client.Client
}

// NewK8sLogCollectorRepository 创建新的 K8sLogCollectorRepository
func NewK8sLogCollectorRepository(cli client.Client) *K8sLogCollectorRepository {
	return &K8sLogCollectorRepository{client: cli}
}

// List 列出指定命名空间的日志收集器
func (r *K8sLogCollectorRepository) List(ctx context.Context, namespace string) (*polardbxv1.PolarDBXLogCollectorList, error) {
	list := &polardbxv1.PolarDBXLogCollectorList{}
	opts := []client.ListOption{}
	if namespace != "" {
		opts = append(opts, client.InNamespace(namespace))
	}
	if err := r.client.List(ctx, list, opts...); err != nil {
		return nil, err
	}
	return list, nil
}

// Get 获取指定的日志收集器
func (r *K8sLogCollectorRepository) Get(ctx context.Context, namespace, name string) (*polardbxv1.PolarDBXLogCollector, error) {
	collector := &polardbxv1.PolarDBXLogCollector{}
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, collector); err != nil {
		return nil, err
	}
	return collector, nil
}

// Create 创建日志收集器
func (r *K8sLogCollectorRepository) Create(ctx context.Context, collector *polardbxv1.PolarDBXLogCollector) (*polardbxv1.PolarDBXLogCollector, error) {
	if err := r.client.Create(ctx, collector); err != nil {
		return nil, err
	}
	return collector, nil
}

// Update 更新日志收集器
func (r *K8sLogCollectorRepository) Update(ctx context.Context, collector *polardbxv1.PolarDBXLogCollector) (*polardbxv1.PolarDBXLogCollector, error) {
	if err := r.client.Update(ctx, collector); err != nil {
		return nil, err
	}
	return collector, nil
}

// Delete 删除日志收集器
func (r *K8sLogCollectorRepository) Delete(ctx context.Context, namespace, name string) error {
	collector := &polardbxv1.PolarDBXLogCollector{}
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, collector); err != nil {
		return err
	}
	return r.client.Delete(ctx, collector)
}
