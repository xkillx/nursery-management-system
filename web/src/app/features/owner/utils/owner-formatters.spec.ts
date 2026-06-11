import { formatGbp, formatSetupStatus, formatGrantOutcome, isExceptionSite } from './owner-formatters';

describe('owner-formatters', () => {
  describe('formatGbp', () => {
    it('formats minor units to GBP string', () => {
      expect(formatGbp(150000)).toBe('£1,500.00');
      expect(formatGbp(0)).toBe('£0.00');
      expect(formatGbp(99)).toBe('£0.99');
    });
  });

  describe('formatSetupStatus', () => {
    it('maps known statuses to clear labels', () => {
      expect(formatSetupStatus('complete')).toBe('Complete');
      expect(formatSetupStatus('incomplete_attendance')).toBe('Incomplete attendance');
    });

    it('formats unknown statuses safely', () => {
      expect(formatSetupStatus('some_new_status')).toBe('Some New Status');
    });
  });

  describe('formatGrantOutcome', () => {
    it('maps Go-style outcomes', () => {
      expect(formatGrantOutcome('manager_membership_granted')).toBe('Manager access granted.');
      expect(formatGrantOutcome('manager_membership_reactivated')).toBe('Manager access reactivated.');
      expect(formatGrantOutcome('manager_membership_already_active')).toBe('Manager already has active access.');
      expect(formatGrantOutcome('manager_invite_pending')).toContain('invite sent');
    });

    it('maps OpenAPI short outcomes', () => {
      expect(formatGrantOutcome('granted')).toBe('Manager access granted.');
      expect(formatGrantOutcome('reactivated')).toBe('Manager access reactivated.');
      expect(formatGrantOutcome('already_active')).toBe('Manager already has active access.');
      expect(formatGrantOutcome('invite_pending')).toContain('invite sent');
    });
  });

  describe('isExceptionSite', () => {
    const healthySite = {
      activeManagerCount: 1,
      invoicePaymentHealth: { overdueOutstandingMinor: 0, paymentFailedCount: 0, outstandingMinor: 0 },
      attendance: { incompleteAttendanceCount: 0 },
      fundingReadiness: { flaggedChildCount: 0 },
    };

    it('returns false for healthy site', () => {
      expect(isExceptionSite(healthySite)).toBeFalse();
    });

    it('flags no active manager', () => {
      expect(isExceptionSite({ ...healthySite, activeManagerCount: 0 })).toBeTrue();
    });

    it('flags overdue outstanding', () => {
      expect(isExceptionSite({
        ...healthySite,
        invoicePaymentHealth: { ...healthySite.invoicePaymentHealth, overdueOutstandingMinor: 500 },
      })).toBeTrue();
    });

    it('flags payment failures', () => {
      expect(isExceptionSite({
        ...healthySite,
        invoicePaymentHealth: { ...healthySite.invoicePaymentHealth, paymentFailedCount: 1 },
      })).toBeTrue();
    });

    it('flags outstanding balance', () => {
      expect(isExceptionSite({
        ...healthySite,
        invoicePaymentHealth: { ...healthySite.invoicePaymentHealth, outstandingMinor: 100 },
      })).toBeTrue();
    });

    it('flags incomplete attendance', () => {
      expect(isExceptionSite({
        ...healthySite,
        attendance: { incompleteAttendanceCount: 3 },
      })).toBeTrue();
    });

    it('flags funding readiness issues', () => {
      expect(isExceptionSite({
        ...healthySite,
        fundingReadiness: { flaggedChildCount: 2 },
      })).toBeTrue();
    });
  });
});
