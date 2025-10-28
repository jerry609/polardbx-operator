package service

import (
	"context"

	spec "polardbx-ui-backend/pkg/api/monitoringv2/spec"
)

// SuggestedFixType标识推荐修复的形态，用于区分自动执行、脚本化或手动操作步骤。
type SuggestedFixType string

const (
	// SuggestedFixTypeOperatorHook 表示可以通过 Operator/Helm 钩子自动触发的修复动作。
	SuggestedFixTypeOperatorHook SuggestedFixType = "operator-hook"
	// SuggestedFixTypeManualProcedure 表示需要用户按步骤手动执行的修复动作。
	SuggestedFixTypeManualProcedure SuggestedFixType = "manual-procedure"
	// SuggestedFixTypeKnowledgeBase 表示跳转到知识库或外部文档的辅助信息。
	SuggestedFixTypeKnowledgeBase SuggestedFixType = "knowledge-base"
)

// SuggestedFix描述诊断建议的结构，后续会映射到 OpenAPI 中的 AutoFix/Fix 相关模型。
type SuggestedFix struct {
	// ID 在同一次诊断结果中需要唯一，便于 UI 触发自动修复或埋点。
	ID string
	// Title 面向用户的短标题。
	Title string
	// Description 提供更详细的步骤或执行背景。
	Description string
	// Type 用于区分修复手段，例如自动化 Hook、手动操作或知识库链接。
	Type SuggestedFixType
	// Automated 指示该修复是否可以由后端自动执行。
	Automated bool
	// Steps 是手动方案的结构化步骤描述。
	Steps []string
	// Verification 给出修复完成后建议的验证方法。
	Verification string
	// RelatedComponents 关联到的监控组件，便于 UI 高亮。
	RelatedComponents []spec.ComponentName
	// Metadata 保留额外信息，例如需要调用的 Operator Hook 名称、RBAC 前提等。
	Metadata map[string]string
}

// DiagnosticFinding 表示探针给出的单条诊断结果。
type DiagnosticFinding struct {
	// ProbeID 产出该结论的探针标识。
	ProbeID string
	// Summary 面向用户的诊断摘要。
	Summary string
	// Category 对应 OpenAPI 中的 FailureCategory，方便与服务端枚举保持一致。
	Category spec.FailureCategory
	// Severity 对应 DiagnosisSeverity，标识严重程度。
	Severity spec.DiagnosisSeverity
	// PossibleCauses 可能的根因候选。
	PossibleCauses []string
	// SuggestedFixes 建议的修复方案列表。
	SuggestedFixes []SuggestedFix
	// Evidence 记录截图、日志片段等佐证信息。
	Evidence []string
	// Details 附加的键值说明，用于 UI 透出更多上下文。
	Details map[string]string
}

// ProbeInput 为诊断探针提供上下文信息，包含集群命名空间、会话信息以及已有的安装状态。
type ProbeInput struct {
	Namespace string
	SessionID string
	// Checkpoint 为最近一次持久化的安装检查点，可为空。
	Checkpoint *spec.Checkpoint
	// Status 为安装执行状态快照，包含组件进度和错误列表。
	Status *spec.InstallStatusResponse
	// Plan 为当前安装计划，用于比对期望动作。
	Plan *spec.InstallationPlan
	// Context 额外的诊断上下文，例如用户手动收集的日志路径。
	Context map[string]string
}

// DiagnosticProbe 定义诊断探针需要实现的最小接口。
type DiagnosticProbe interface {
	// ID 返回全局唯一的探针标识，例如 "prometheus-health"。
	ID() string
	// Description 提供简短的人类可读说明。
	Description() string
	// Run 执行诊断逻辑，返回一组诊断结果；若无问题可返回空切片。
	Run(ctx context.Context, input ProbeInput) ([]DiagnosticFinding, error)
}
