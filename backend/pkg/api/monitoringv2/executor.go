package monitoringv2

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"polardbx-ui-backend/pkg/api/monitoringv2/service"
	spec "polardbx-ui-backend/pkg/api/monitoringv2/spec"
)

type installationExecutor struct {
	store    *sessionStore
	running  sync.Map
	detector func() service.DetectionService
}

var executorLogger = ctrllog.Log.WithName("monitoringv2").WithName("executor")

func newInstallationExecutor(store *sessionStore) *installationExecutor {
	return &installationExecutor{
		store:    store,
		detector: getDetectionService,
	}
}

func (e *installationExecutor) start(ctx context.Context, sessionID string, plan spec.InstallationPlan, persist sessionPersistence) {
	if sessionID == "" {
		executorLogger.Info("executor start aborted: empty session id")
		return
	}
	if !e.store.shouldRun(sessionID) {
		executorLogger.WithValues("sessionId", sessionID).Info("executor start skipped: session not runnable")
		return
	}
	if _, loaded := e.running.LoadOrStore(sessionID, struct{}{}); loaded {
		executorLogger.WithValues("sessionId", sessionID).Info("executor already running")
		return
	}

	execCtx := ctx
	if execCtx == nil {
		execCtx = context.Background()
	}
	executorLogger.WithValues("sessionId", sessionID, "steps", len(plan.Steps)).Info("executor start queued")

	go e.run(execCtx, sessionID, plan, persist)
}

func (e *installationExecutor) run(ctx context.Context, sessionID string, plan spec.InstallationPlan, persist sessionPersistence) {
	defer e.running.Delete(sessionID)
	executorLogger.WithValues("sessionId", sessionID, "steps", len(plan.Steps)).Info("executor loop started")

	detector := e.detector()
	namespace := plan.SessionTemplate.Namespace
	steps := append([]spec.PlanStep{}, plan.Steps...)
	if len(steps) == 0 {
		executorLogger.WithValues("sessionId", sessionID).Info("no steps to execute; marking session complete")
		_, _ = e.store.boost(ctx, sessionID, 1, persist)
		e.store.snapshot(ctx, sessionID, persist)
		executorLogger.WithValues("sessionId", sessionID).Info("executor loop finished")
		return
	}
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Order < steps[j].Order
	})

	for _, step := range steps {
		stepLog := executorLogger.WithValues(
			"sessionId", sessionID,
			"stepOrder", step.Order,
			"component", step.Component,
			"action", step.Action,
		)
		if e.store.isStepCompleted(sessionID, step.Order) {
			stepLog.Info("step already completed; skipping")
			continue
		}
		if !e.store.shouldRun(sessionID) {
			stepLog.Info("session no longer runnable; stopping executor")
			return
		}
		select {
		case <-ctx.Done():
			stepLog.Info("executor context cancelled")
			return
		default:
		}

		if _, err := e.store.setCurrentStep(ctx, sessionID, step.Order, persist); err != nil {
			stepLog.Error(err, "failed to set current step")
			return
		}
		stepLog.Info("verifying step state")

		healthy, detail, err := e.verifyStep(ctx, detector, namespace, step)
		if err != nil {
			installErr := e.newInstallError(step, spec.Unknown, fmt.Sprintf("step verification failed: %v", err))
			e.store.addError(ctx, sessionID, step.Order, installErr, persist)
			stepLog.Error(err, "step verification errored")
			return
		}
		if !healthy {
			installErr := e.newInstallError(step, spec.Dependency, detail)
			e.store.addError(ctx, sessionID, step.Order, installErr, persist)
			stepLog.Info("step verification reported unhealthy", "detail", detail)
			return
		}

		if _, err := e.store.completeStep(ctx, sessionID, step.Order, persist); err != nil {
			stepLog.Error(err, "failed to mark step complete")
			return
		}
		stepLog.Info("step completed successfully")

		select {
		case <-ctx.Done():
			stepLog.Info("executor context cancelled after completion")
			return
		case <-time.After(200 * time.Millisecond):
		}
	}

	e.store.snapshot(ctx, sessionID, persist)
	executorLogger.WithValues("sessionId", sessionID).Info("executor loop finished")
}

func (e *installationExecutor) verifyStep(ctx context.Context, detector service.DetectionService, namespace string, step spec.PlanStep) (bool, string, error) {
	if detector == nil {
		executorLogger.WithValues("namespace", namespace, "component", step.Component).Error(fmt.Errorf("detection service not configured"), "unable to verify step")
		return false, "detection service unavailable", fmt.Errorf("detection service not configured")
	}

	snapshot, err := detector.Detect(ctx, namespace)
	if err != nil {
		executorLogger.WithValues("namespace", namespace, "component", step.Component).Error(err, "failed to run detection during verification")
		return false, "failed to detect monitoring environment", err
	}

	var component *spec.DetectedComponent
	for i := range snapshot.Components {
		if snapshot.Components[i].Name == step.Component {
			component = &snapshot.Components[i]
			break
		}
	}

	exists := component != nil && component.Exists != nil && *component.Exists
	healthy := component != nil && component.Healthy != nil && *component.Healthy

	switch step.Action {
	case spec.PlanStepActionVerify:
		if healthy {
			return true, "", nil
		}
		return false, fmt.Sprintf("component %s not healthy", step.Component), nil
	case spec.PlanStepActionInstall, spec.PlanStepActionUpgrade, spec.PlanStepActionRepair:
		if healthy {
			return true, "", nil
		}
		if !exists {
			return false, fmt.Sprintf("component %s is still missing", step.Component), nil
		}
		return false, fmt.Sprintf("component %s is not healthy", step.Component), nil
	case spec.PlanStepActionUninstall:
		if !exists {
			return true, "", nil
		}
		return false, fmt.Sprintf("component %s still present", step.Component), nil
	default:
		if healthy {
			return true, "", nil
		}
		return false, fmt.Sprintf("component %s verification failed", step.Component), nil
	}
}

func (e *installationExecutor) newInstallError(step spec.PlanStep, category spec.FailureCategory, message string) spec.InstallError {
	component := step.Component
	now := time.Now().UTC()
	return spec.InstallError{
		Category:   category,
		Component:  &component,
		Message:    message,
		OccurredAt: &now,
		StepOrder:  &step.Order,
	}
}
