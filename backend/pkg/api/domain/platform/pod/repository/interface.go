package repository

import (
	"context"
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/remotecommand"
)

// PodRepository 定义 Pod 操作的存储层接口
type PodRepository interface {
	// List 列出指定命名空间的所有 Pod
	List(ctx context.Context, namespace string) ([]corev1.Pod, error)

	// ListForCluster 列出 PolarDBX 集群的所有 Pod
	ListForCluster(ctx context.Context, namespace, clusterName string) ([]corev1.Pod, error)

	// Get 获取指定的 Pod
	Get(ctx context.Context, namespace, name string) (*corev1.Pod, error)

	// Delete 删除指定的 Pod
	Delete(ctx context.Context, namespace, name string) error

	// GetLogs 获取 Pod 日志
	GetLogs(ctx context.Context, namespace, podName, container string, tailLines int64) (string, error)
}

// ExecConfig 执行命令的配置
type ExecConfig struct {
	Namespace string
	PodName   string
	Container string
	Command   []string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	TTY       bool
	SizeQueue remotecommand.TerminalSizeQueue
}

// ExecRepository 定义 Pod Exec 操作的存储层接口
type ExecRepository interface {
	// Exec 在 Pod 中执行命令
	Exec(ctx context.Context, cfg ExecConfig) error
}
