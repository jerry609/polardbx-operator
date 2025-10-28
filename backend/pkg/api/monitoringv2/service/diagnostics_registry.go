package service

import (
	"context"
	"sort"
)

// DiagnosticProbeFactory 用于延迟构建探针实例，便于测试与依赖注入。
type DiagnosticProbeFactory func() DiagnosticProbe

// registry 默认包含设计阶段定义的探针，后续阶段可替换为具体实现。
var defaultDiagnosticProbes = map[string]struct {
	Description string
	Factory     DiagnosticProbeFactory
}{
	"prometheus-health": {
		Description: "检测 Prometheus StatefulSet、Service 以及核心指标可用性",
		Factory:     func() DiagnosticProbe { return newPrometheusHealthProbe() },
	},
	"grafana-connectivity": {
		Description: "验证 Grafana Pod 与数据源连通性",
		Factory:     func() DiagnosticProbe { return newPlaceholderProbe("grafana-connectivity") },
	},
	"operator-events": {
		Description: "分析监控 Operator 的事件与日志以发现调谐失败",
		Factory:     func() DiagnosticProbe { return newPlaceholderProbe("operator-events") },
	},
	"k8s-events": {
		Description: "聚合命名空间层面的 Warning/Error 事件",
		Factory:     func() DiagnosticProbe { return newPlaceholderProbe("k8s-events") },
	},
}

// ListDefaultDiagnosticProbes 返回排序后的探针 ID 列表，便于 UI 或日志展示。
func ListDefaultDiagnosticProbes() []string {
	ids := make([]string, 0, len(defaultDiagnosticProbes))
	for id := range defaultDiagnosticProbes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// ResolveDiagnosticProbe 根据 ID 构造探针实例；找不到时返回 nil。
func ResolveDiagnosticProbe(id string) DiagnosticProbe {
	if meta, ok := defaultDiagnosticProbes[id]; ok && meta.Factory != nil {
		return meta.Factory()
	}
	return nil
}

// newPlaceholderProbe 提供阶段占位实现，在 Phase 3 中将替换为真实逻辑。
func newPlaceholderProbe(id string) DiagnosticProbe {
	return placeholderProbe{id: id}
}

// placeholderProbe 满足 DiagnosticProbe 接口但不执行任何诊断。
type placeholderProbe struct {
	id string
}

func (p placeholderProbe) ID() string { return p.id }

func (p placeholderProbe) Description() string { return "placeholder" }

func (p placeholderProbe) Run(_ context.Context, _ ProbeInput) ([]DiagnosticFinding, error) {
	return nil, nil
}
