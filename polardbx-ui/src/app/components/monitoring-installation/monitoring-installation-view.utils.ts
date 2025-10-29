import type { EnvironmentSnapshot } from '../../services/monitoring-installation-api.service';
import type { components } from '../../models/generated/monitoring-installation';

export type InstallIntent = components['schemas']['CreatePlanRequest']['intent'];

export interface WizardStepTitleKeys {
  detect: string;
  plan: string;
  progress: string;
}

export const getWizardStepTitleKeys = (intent: InstallIntent | null | undefined): WizardStepTitleKeys => {
  if (intent === 'repair') {
    return {
      detect: 'steps.detect.title.repair',
      plan: 'steps.plan.title.repair',
      progress: 'steps.progress.title.repair'
    };
  }
  return {
    detect: 'steps.detect.title',
    plan: 'steps.plan.title',
    progress: 'steps.progress.title'
  };
};

interface RepairEntryContext {
  planIntent: InstallIntent | null | undefined;
  detectionSnapshot: EnvironmentSnapshot | null | undefined;
  detectionLoading: boolean;
  planLoading: boolean;
  installPolling: boolean;
  detectionError?: string | null;
  hasComponents: boolean;
}

const meetsRepairEntryPrerequisites = (context: RepairEntryContext): boolean => {
  if (context.planIntent !== 'repair') {
    return false;
  }
  if (context.detectionLoading || context.planLoading || context.installPolling) {
    return false;
  }
  if (!context.detectionSnapshot) {
    return false;
  }
  if (context.detectionError) {
    return false;
  }
  if (!context.hasComponents) {
    return false;
  }
  return true;
};

export interface RepairBannerContext extends RepairEntryContext {
  dismissed: boolean;
}

export const shouldShowRepairPlanBanner = (context: RepairBannerContext): boolean => {
  if (context.dismissed) {
    return false;
  }
  return meetsRepairEntryPrerequisites(context);
};

export type RepairQuickJumpContext = RepairEntryContext;

export const shouldShowRepairQuickJump = (context: RepairQuickJumpContext): boolean => {
  return meetsRepairEntryPrerequisites(context);
};
