import { containsInstallError, isSameInstallError, type InstallError } from './monitoring-installation-utils';

const buildError = (overrides: Partial<InstallError> = {}): InstallError => ({
  component: 'prometheus',
  category: 'ImagePull',
  message: 'failed to pull image',
  stepOrder: 2,
  ...overrides,
});

describe('monitoring-installation-utils', () => {
  describe('isSameInstallError', () => {
    it('returns true when component and step match', () => {
      const errorA = buildError();
      const errorB = buildError();
      expect(isSameInstallError(errorA, errorB)).toBeTrue();
    });

    it('returns true when component and message match', () => {
      const errorA = buildError({ stepOrder: 1 });
      const errorB = buildError({ stepOrder: 5 });
      expect(isSameInstallError(errorA, errorB)).toBeTrue();
    });

    it('returns true when category and message match', () => {
      const errorA = buildError({ component: 'prometheus', stepOrder: undefined });
      const errorB = buildError({ component: 'grafana', stepOrder: undefined });
      expect(isSameInstallError(errorA, errorB)).toBeTrue();
    });

    it('returns false when relevant fields differ', () => {
      const errorA = buildError({ message: 'permission denied' });
      const errorB = buildError({ stepOrder: 3 });
      expect(isSameInstallError(errorA, errorB)).toBeFalse();
    });

    it('returns false when either error is null', () => {
      const errorA = buildError();
      expect(isSameInstallError(errorA, null)).toBeFalse();
    });

    it('ignores surrounding whitespace when comparing fields', () => {
      const errorA = buildError({
        component: '  prometheus  ' as unknown as InstallError['component'],
        category: ' ImagePull ' as unknown as InstallError['category'],
        message: ' failed to pull image ',
        stepOrder: 3,
      });
      const errorB = buildError({
        component: 'prometheus',
        category: 'ImagePull',
        message: 'failed to pull image',
        stepOrder: 3,
      });
      expect(isSameInstallError(errorA, errorB)).toBeTrue();
    });
  });

  describe('containsInstallError', () => {
    it('returns true when target exists in list', () => {
      const errors = [buildError(), buildError({ component: 'grafana' })];
      expect(containsInstallError(errors, buildError({ stepOrder: 5 }))).toBeTrue();
    });

    it('returns false when list is empty', () => {
      expect(containsInstallError([], buildError())).toBeFalse();
    });

    it('returns false when target is undefined', () => {
      const errors = [buildError()];
      expect(containsInstallError(errors, undefined)).toBeFalse();
    });

    it('returns false when target is null', () => {
      const errors = [buildError()];
      expect(containsInstallError(errors, null)).toBeFalse();
    });
  });
});
