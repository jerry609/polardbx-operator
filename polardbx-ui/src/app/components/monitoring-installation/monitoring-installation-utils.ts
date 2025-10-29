import type { components } from '../../models/generated/monitoring-installation';

export type InstallError = components['schemas']['InstallError'];

interface NormalizedError {
  component: string;
  category: string;
  message: string;
  step: string;
}

const normalizeError = (error: InstallError | null | undefined): NormalizedError | null => {
  if (!error) {
    return null;
  }
  const component = error.component ? String(error.component).trim() : '';
  const category = (error.category ?? '').trim();
  const message = (error.message ?? '').trim();
  const step = error.stepOrder != null ? String(error.stepOrder) : '';
  return { component, category, message, step };
};

export const isSameInstallError = (
  target: InstallError | null | undefined,
  candidate: InstallError | null | undefined
): boolean => {
  const normalizedTarget = normalizeError(target);
  const normalizedCandidate = normalizeError(candidate);
  if (!normalizedTarget || !normalizedCandidate) {
    return false;
  }

  const componentMatch = normalizedTarget.component !== '' && normalizedTarget.component === normalizedCandidate.component;
  const stepMatch = normalizedTarget.step !== '' && normalizedTarget.step === normalizedCandidate.step;
  const messageMatch = normalizedTarget.message !== '' && normalizedTarget.message === normalizedCandidate.message;
  const categoryMatch = normalizedTarget.category !== '' && normalizedTarget.category === normalizedCandidate.category;

  if (componentMatch && stepMatch) {
    return true;
  }

  if (componentMatch && messageMatch) {
    return true;
  }

  if (categoryMatch && messageMatch) {
    return true;
  }

  return false;
};

export const containsInstallError = (
  errors: ReadonlyArray<InstallError> | null | undefined,
  target: InstallError | null | undefined
): boolean => {
  if (!errors || errors.length === 0 || !target) {
    return false;
  }
  return errors.some(err => isSameInstallError(target, err));
};
