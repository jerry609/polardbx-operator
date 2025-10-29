import type { EnvironmentSnapshot } from '../../services/monitoring-installation-api.service';
import { getWizardStepTitleKeys, shouldShowRepairPlanBanner, shouldShowRepairQuickJump } from './monitoring-installation-view.utils';

describe('monitoring-installation-view.utils', () => {
  const snapshot: EnvironmentSnapshot = {
    namespace: 'polardbx-monitor',
    detectedAt: new Date().toISOString(),
    components: [],
  };

  it('returns true when repair intent is active and prerequisites are met', () => {
    const result = shouldShowRepairPlanBanner({
      planIntent: 'repair',
      detectionSnapshot: snapshot,
      detectionLoading: false,
      planLoading: false,
      installPolling: false,
      detectionError: null,
      hasComponents: true,
      dismissed: false,
    });

    expect(result).toBeTrue();
  });

  it('returns false when banner has been dismissed', () => {
    const result = shouldShowRepairPlanBanner({
      planIntent: 'repair',
      detectionSnapshot: snapshot,
      detectionLoading: false,
      planLoading: false,
      installPolling: false,
      detectionError: null,
      hasComponents: true,
      dismissed: true,
    });

    expect(result).toBeFalse();
  });

  it('returns false when intent is not repair', () => {
    const result = shouldShowRepairPlanBanner({
      planIntent: 'install',
      detectionSnapshot: snapshot,
      detectionLoading: false,
      planLoading: false,
      installPolling: false,
      detectionError: null,
      hasComponents: true,
      dismissed: false,
    });

    expect(result).toBeFalse();
  });

  it('uses repair-specific step titles in repair mode', () => {
    const keys = getWizardStepTitleKeys('repair');

    expect(keys.detect).toBe('steps.detect.title.repair');
    expect(keys.plan).toBe('steps.plan.title.repair');
    expect(keys.progress).toBe('steps.progress.title.repair');
  });

  it('falls back to default step titles outside repair mode', () => {
    const keys = getWizardStepTitleKeys('install');

    expect(keys.detect).toBe('steps.detect.title');
    expect(keys.plan).toBe('steps.plan.title');
    expect(keys.progress).toBe('steps.progress.title');
  });

  it('shows quick jump CTA when prerequisites are met', () => {
    const result = shouldShowRepairQuickJump({
      planIntent: 'repair',
      detectionSnapshot: snapshot,
      detectionLoading: false,
      planLoading: false,
      installPolling: false,
      detectionError: null,
      hasComponents: true,
    });

    expect(result).toBeTrue();
  });

  it('hides quick jump CTA when detection is still loading', () => {
    const result = shouldShowRepairQuickJump({
      planIntent: 'repair',
      detectionSnapshot: snapshot,
      detectionLoading: true,
      planLoading: false,
      installPolling: false,
      detectionError: null,
      hasComponents: true,
    });

    expect(result).toBeFalse();
  });
});
