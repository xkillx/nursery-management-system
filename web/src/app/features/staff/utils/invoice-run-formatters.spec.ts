import { formatMinutes, formatBillingMonthLabel, defaultCompletedBillingMonth, blockerNextAction, blockerLabel } from './invoice-run-formatters';

describe('invoice-run-formatters', () => {
  describe('formatMinutes', () => {
    it('formats zero', () => {
      expect(formatMinutes(0)).toBe('0h 0m');
    });

    it('formats minutes only', () => {
      expect(formatMinutes(45)).toBe('45m');
    });

    it('formats hours only', () => {
      expect(formatMinutes(120)).toBe('2h');
    });

    it('formats hours and minutes', () => {
      expect(formatMinutes(135)).toBe('2h 15m');
    });
  });

  describe('formatBillingMonthLabel', () => {
    it('formats May 2026', () => {
      expect(formatBillingMonthLabel('2026-05')).toBe('May 2026');
    });

    it('formats January', () => {
      expect(formatBillingMonthLabel('2026-01')).toBe('January 2026');
    });

    it('formats December', () => {
      expect(formatBillingMonthLabel('2025-12')).toBe('December 2025');
    });
  });

  describe('defaultCompletedBillingMonth', () => {
    it('returns May for June 2026', () => {
      const june = new Date('2026-06-15T12:00:00Z');
      expect(defaultCompletedBillingMonth(june)).toBe('2026-05');
    });

    it('returns December previous year for January 2026', () => {
      const january = new Date('2026-01-15T12:00:00Z');
      expect(defaultCompletedBillingMonth(january)).toBe('2025-12');
    });

    it('returns March for April', () => {
      const april = new Date('2026-04-01T00:00:00Z');
      expect(defaultCompletedBillingMonth(april)).toBe('2026-03');
    });
  });

  describe('blockerLabel', () => {
    it('returns human-readable label for known codes', () => {
      expect(blockerLabel('incomplete_attendance')).toBe('Incomplete attendance');
      expect(blockerLabel('missing_funding_profile')).toBe('Missing funding profile');
      expect(blockerLabel('existing_issued_invoice')).toBe('Already issued');
    });
  });

  describe('blockerNextAction', () => {
    it('routes incomplete attendance to corrections', () => {
      const action = blockerNextAction('incomplete_attendance');
      expect(action.label).toBe('Correct attendance');
      expect(action.route).toEqual(['/staff/manager/attendance-corrections']);
    });

    it('routes missing funding profile with child id to child detail', () => {
      const action = blockerNextAction('missing_funding_profile', 'child-1', '2026-05');
      expect(action.label).toBe('Review funding');
      expect(action.route).toEqual(['/staff/manager/children', 'child-1']);
      expect(action.queryParams).toEqual({ billing_month: '2026-05' });
    });

    it('routes missing funding profile without child id to funding overview', () => {
      const action = blockerNextAction('missing_funding_profile');
      expect(action.label).toBe('Funding overview');
      expect(action.route).toEqual(['/staff/manager/funding']);
    });

    it('routes missing core hourly rate to child detail', () => {
      const action = blockerNextAction('missing_core_hourly_rate', 'child-2');
      expect(action.label).toBe('Review child');
      expect(action.route).toEqual(['/staff/manager/children', 'child-2']);
    });

    it('routes missing guardian link to child detail', () => {
      const action = blockerNextAction('missing_guardian_link', 'child-3');
      expect(action.route).toEqual(['/staff/manager/children', 'child-3']);
    });

    it('existing issued invoice has no route', () => {
      const action = blockerNextAction('existing_issued_invoice');
      expect(action.label).toBe('Already issued');
      expect(action.route).toBeUndefined();
    });
  });
});
