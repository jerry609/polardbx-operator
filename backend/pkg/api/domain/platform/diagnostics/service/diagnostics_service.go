package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// 诊断 Pod 名称前缀
	ClinicPodPrefix = "polardbx-clinic-"
	// 诊断 Pod 标签
	ClinicLabelKey = "polardbx/clinic"
	// 诊断状态
	DiagStatusRunning   = "running"
	DiagStatusSucceeded = "succeeded"
	DiagStatusFailed    = "failed"
	DiagStatusPending   = "pending"
)

// DiagnosticJob 诊断任务信息
type DiagnosticJob struct {
	ID          string    `json:"id"`
	Namespace   string    `json:"namespace"`
	Cluster     string    `json:"cluster"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt,omitempty"`
	OutputPath  string    `json:"outputPath,omitempty"`
	Message     string    `json:"message,omitempty"`
}

// DiagnosticsService 诊断服务
type DiagnosticsService struct {
	cli client.Client
}

// NewDiagnosticsService 创建诊断服务
func NewDiagnosticsService(cli client.Client) *DiagnosticsService {
	return &DiagnosticsService{cli: cli}
}

// StartDiagnosis 启动诊断任务
// 创建 polardbx-clinic Pod 来收集集群诊断信息
func (s *DiagnosticsService) StartDiagnosis(ctx context.Context, namespace, clusterName string) (*DiagnosticJob, error) {
	// 生成唯一的诊断任务 ID
	jobID := fmt.Sprintf("%s-%d", clusterName, time.Now().Unix())
	podName := ClinicPodPrefix + jobID

	// 构建诊断 Pod
	pod := s.buildClinicPod(namespace, podName, clusterName, jobID)

	// 创建 Pod
	if err := s.cli.Create(ctx, pod); err != nil {
		return nil, fmt.Errorf("创建诊断 Pod 失败: %w", err)
	}

	return &DiagnosticJob{
		ID:        jobID,
		Namespace: namespace,
		Cluster:   clusterName,
		Status:    DiagStatusRunning,
		Progress:  0,
		StartedAt: time.Now(),
		Message:   "诊断任务已启动",
	}, nil
}

// GetDiagnosisStatus 获取诊断任务状态
func (s *DiagnosticsService) GetDiagnosisStatus(ctx context.Context, namespace, jobID string) (*DiagnosticJob, error) {
	podName := ClinicPodPrefix + jobID

	var pod corev1.Pod
	key := client.ObjectKey{Namespace: namespace, Name: podName}
	if err := s.cli.Get(ctx, key, &pod); err != nil {
		return nil, fmt.Errorf("获取诊断 Pod 状态失败: %w", err)
	}

	job := &DiagnosticJob{
		ID:        jobID,
		Namespace: namespace,
		Cluster:   pod.Labels["polardbx/cluster"],
	}

	// 解析 Pod 状态
	switch pod.Status.Phase {
	case corev1.PodPending:
		job.Status = DiagStatusPending
		job.Progress = 0
		job.Message = "诊断 Pod 正在启动"
	case corev1.PodRunning:
		job.Status = DiagStatusRunning
		job.Progress = 50
		job.Message = "正在收集诊断信息"
	case corev1.PodSucceeded:
		job.Status = DiagStatusSucceeded
		job.Progress = 100
		job.Message = "诊断完成"
		job.OutputPath = fmt.Sprintf("/tmp/polardbx-clinic/%s.tar.gz", jobID)
	case corev1.PodFailed:
		job.Status = DiagStatusFailed
		job.Progress = 0
		job.Message = "诊断任务失败"
		// 尝试获取失败原因
		if len(pod.Status.ContainerStatuses) > 0 {
			cs := pod.Status.ContainerStatuses[0]
			if cs.State.Terminated != nil && cs.State.Terminated.Reason != "" {
				job.Message = fmt.Sprintf("诊断失败: %s", cs.State.Terminated.Reason)
			}
		}
	default:
		job.Status = DiagStatusPending
		job.Progress = 0
	}

	// 设置时间
	if !pod.CreationTimestamp.IsZero() {
		job.StartedAt = pod.CreationTimestamp.Time
	}
	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Terminated != nil {
				job.CompletedAt = cs.State.Terminated.FinishedAt.Time
				break
			}
		}
	}

	return job, nil
}

// ListDiagnosisReports 列出诊断报告
func (s *DiagnosticsService) ListDiagnosisReports(ctx context.Context, namespace string) ([]DiagnosticJob, error) {
	var podList corev1.PodList

	// 构建标签选择器
	labelSelector := labels.SelectorFromSet(map[string]string{
		ClinicLabelKey: "true",
	})

	listOpts := &client.ListOptions{
		LabelSelector: labelSelector,
	}
	if namespace != "" {
		listOpts.Namespace = namespace
	}

	if err := s.cli.List(ctx, &podList, listOpts); err != nil {
		return nil, fmt.Errorf("列出诊断 Pod 失败: %w", err)
	}

	var reports []DiagnosticJob
	for _, pod := range podList.Items {
		jobID := strings.TrimPrefix(pod.Name, ClinicPodPrefix)
		job := DiagnosticJob{
			ID:        jobID,
			Namespace: pod.Namespace,
			Cluster:   pod.Labels["polardbx/cluster"],
			StartedAt: pod.CreationTimestamp.Time,
		}

		// 解析状态
		switch pod.Status.Phase {
		case corev1.PodSucceeded:
			job.Status = DiagStatusSucceeded
			job.Progress = 100
		case corev1.PodFailed:
			job.Status = DiagStatusFailed
		case corev1.PodRunning:
			job.Status = DiagStatusRunning
			job.Progress = 50
		default:
			job.Status = DiagStatusPending
		}

		reports = append(reports, job)
	}

	return reports, nil
}

// GetDownloadInfo 获取诊断报告下载信息
func (s *DiagnosticsService) GetDownloadInfo(ctx context.Context, namespace, jobID string) (string, error) {
	// 检查 Pod 状态
	job, err := s.GetDiagnosisStatus(ctx, namespace, jobID)
	if err != nil {
		return "", err
	}

	if job.Status != DiagStatusSucceeded {
		return "", fmt.Errorf("诊断任务尚未完成，当前状态: %s", job.Status)
	}

	// 返回诊断报告的路径
	// 实际使用时可能需要通过 kubectl cp 或者其他方式获取文件
	return job.OutputPath, nil
}

// buildClinicPod 构建诊断 Pod 配置
func (s *DiagnosticsService) buildClinicPod(namespace, podName, clusterName, jobID string) *corev1.Pod {
	// 诊断脚本 - 收集各种诊断信息
	diagScript := `#!/bin/bash
set -e

CLUSTER_NAME="${CLUSTER_NAME:-unknown}"
OUTPUT_DIR="/tmp/polardbx-clinic"
REPORT_FILE="${OUTPUT_DIR}/${JOB_ID}.tar.gz"

mkdir -p ${OUTPUT_DIR}/data

echo "=== PolarDB-X Clinic Diagnostic Tool ==="
echo "Cluster: ${CLUSTER_NAME}"
echo "Namespace: ${NAMESPACE}"
echo "Start Time: $(date)"

# 收集集群信息
echo "Collecting cluster information..."
kubectl get pxc ${CLUSTER_NAME} -n ${NAMESPACE} -o yaml > ${OUTPUT_DIR}/data/pxc.yaml 2>/dev/null || echo "Failed to get PXC"

# 收集 XStore 信息
echo "Collecting XStore information..."
kubectl get xstore -n ${NAMESPACE} -l polardbx/name=${CLUSTER_NAME} -o yaml > ${OUTPUT_DIR}/data/xstores.yaml 2>/dev/null || echo "No XStore found"

# 收集 Pod 信息
echo "Collecting Pod information..."
kubectl get pods -n ${NAMESPACE} -l polardbx/name=${CLUSTER_NAME} -o wide > ${OUTPUT_DIR}/data/pods.txt 2>/dev/null || echo "No Pods found"

# 收集事件
echo "Collecting events..."
kubectl get events -n ${NAMESPACE} --field-selector involvedObject.name=${CLUSTER_NAME} > ${OUTPUT_DIR}/data/events.txt 2>/dev/null || echo "No events"

# 收集 CN 日志
echo "Collecting CN logs..."
for pod in $(kubectl get pods -n ${NAMESPACE} -l polardbx/name=${CLUSTER_NAME},polardbx/role=cn -o jsonpath='{.items[*].metadata.name}' 2>/dev/null); do
    kubectl logs ${pod} -n ${NAMESPACE} --tail=1000 > ${OUTPUT_DIR}/data/cn-${pod}.log 2>/dev/null || true
done

# 收集 DN 日志
echo "Collecting DN logs..."
for pod in $(kubectl get pods -n ${NAMESPACE} -l polardbx/name=${CLUSTER_NAME},polardbx/role=dn -o jsonpath='{.items[*].metadata.name}' 2>/dev/null); do
    kubectl logs ${pod} -n ${NAMESPACE} --tail=1000 > ${OUTPUT_DIR}/data/dn-${pod}.log 2>/dev/null || true
done

# 收集 GMS 日志
echo "Collecting GMS logs..."
for pod in $(kubectl get pods -n ${NAMESPACE} -l polardbx/name=${CLUSTER_NAME},polardbx/role=gms -o jsonpath='{.items[*].metadata.name}' 2>/dev/null); do
    kubectl logs ${pod} -n ${NAMESPACE} --tail=1000 > ${OUTPUT_DIR}/data/gms-${pod}.log 2>/dev/null || true
done

# 收集 CDC 日志
echo "Collecting CDC logs..."
for pod in $(kubectl get pods -n ${NAMESPACE} -l polardbx/name=${CLUSTER_NAME},polardbx/role=cdc -o jsonpath='{.items[*].metadata.name}' 2>/dev/null); do
    kubectl logs ${pod} -n ${NAMESPACE} --tail=1000 > ${OUTPUT_DIR}/data/cdc-${pod}.log 2>/dev/null || true
done

# 收集节点信息
echo "Collecting node information..."
kubectl get nodes -o wide > ${OUTPUT_DIR}/data/nodes.txt 2>/dev/null || echo "Failed to get nodes"

# 收集 ConfigMap
echo "Collecting ConfigMaps..."
kubectl get configmap -n ${NAMESPACE} -l polardbx/name=${CLUSTER_NAME} -o yaml > ${OUTPUT_DIR}/data/configmaps.yaml 2>/dev/null || echo "No ConfigMaps"

# 收集 Secret (仅元数据)
echo "Collecting Secrets metadata..."
kubectl get secrets -n ${NAMESPACE} -l polardbx/name=${CLUSTER_NAME} -o jsonpath='{.items[*].metadata.name}' > ${OUTPUT_DIR}/data/secrets.txt 2>/dev/null || echo "No Secrets"

# 收集 PVC 信息
echo "Collecting PVC information..."
kubectl get pvc -n ${NAMESPACE} -l polardbx/name=${CLUSTER_NAME} -o yaml > ${OUTPUT_DIR}/data/pvcs.yaml 2>/dev/null || echo "No PVCs"

# 生成报告摘要
echo "Generating summary..."
cat > ${OUTPUT_DIR}/data/summary.txt << EOF
=== PolarDB-X Diagnostic Report ===
Generated: $(date)
Cluster: ${CLUSTER_NAME}
Namespace: ${NAMESPACE}

Files collected:
$(ls -la ${OUTPUT_DIR}/data/)
EOF

# 打包
echo "Creating archive..."
cd ${OUTPUT_DIR}
tar -czf ${REPORT_FILE} data/

echo "=== Diagnostic completed ==="
echo "Report: ${REPORT_FILE}"
echo "End Time: $(date)"
`

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels: map[string]string{
				ClinicLabelKey:     "true",
				"polardbx/cluster": clusterName,
				"polardbx/job-id":  jobID,
			},
			Annotations: map[string]string{
				"polardbx/clinic-version": "1.0",
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy:      corev1.RestartPolicyNever,
			ServiceAccountName: "polardbx-operator", // 需要有足够权限的 ServiceAccount
			Containers: []corev1.Container{
				{
					Name:  "clinic",
					Image: "bitnami/kubectl:latest",
					Command: []string{
						"/bin/bash",
						"-c",
						diagScript,
					},
					Env: []corev1.EnvVar{
						{Name: "CLUSTER_NAME", Value: clusterName},
						{Name: "NAMESPACE", Value: namespace},
						{Name: "JOB_ID", Value: jobID},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "output",
							MountPath: "/tmp/polardbx-clinic",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "output",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}
}
