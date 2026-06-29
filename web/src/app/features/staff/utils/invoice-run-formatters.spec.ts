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
    it('returns human-readable label for preflight codes', () => {
      expect(blockerLabel('incomplete_attendance')).toBe('Incomplete attendance');
      expect(blockerLabel('missing_funding_profile')).toBe('Missing funding profile');
      expect(blockerLabel('missing_billing_rate')).toBe('Missing site billing rate');
      expect(blockerLabel('missing_child_name')).toBe('Missing child name');
      expect(blockerLabel('missing_child_date_of_birth')).toBe('Missing date of birth');
      expect(blockerLabel('missing_child_start_date')).toBe('Missing start date');
      expect(blockerLabel('missing_parent_carer_contact')).toBe('Missing parent carer contact');
      expect(blockerLabel('invoice_already_issued')).toBe('Already issued');
    });

    it('returns human-readable label for issue blockers', () => {
      expect(blockerLabel('invoice_not_found')).toBe('Invoice not found');
      expect(blockerLabel('invoice_not_draft')).toBe('Invoice not draft');
    });

    it('returns humanized label for unknown codes', () => {
      expect(blockerLabel('future_blocker_code')).toBe('Future Blocker Code');
    });
  });

  describe('blockerNextAction', () => {
    it('routes incomplete attendance to corrections', () => {
      const action = blockerNextAction('incomplete_attendance');
      expect(action.label).toBe('Correct attendance');
      expect(action.route).toEqual(['/manager/attendance-corrections']);
    });

    it('routes missing funding profile with child id to child detail', () => {
      const action = blockerNextAction('missing_funding_profile', 'child-1', '2026-05');
      expect(action.label).toBe('Review funding');
      expect(action.route).toEqual(['/manager/children', 'child-1']);
      expect(action.queryParams).toEqual({ billing_month: '2026-05' });
    });

    it('shows funding overview label without route when no child id', () => {
      const action = blockerNextAction('missing_funding_profile');
      expect(action.label).toBe('Funding overview');
      expect(action.route).toBeUndefined();
    });

    it('routes missing site billing rate to child detail', () => {
      const action = blockerNextAction('missing_billing_rate', 'child-2');
      expect(action.label).toBe('Review child');
      expect(action.route).toEqual(['/manager/children', 'child-2']);
    });

    it('routes missing parent carer contact to child detail', () => {
      const action = blockerNextAction('missing_parent_carer_contact', 'child-3');
      expect(action.route).toEqual(['/manager/children', 'child-3']);
    });

    it('routes missing child name to child detail', () => {
      const action = blockerNextAction('missing_child_name', 'child-4');
      expect(action.route).toEqual(['/manager/children', 'child-4']);
    });

    it('invoice already issued has no route', () => {
      const action = blockerNextAction('invoice_already_issued');
      expect(action.label).toBe('Already issued');
      expect(action.route).toBeUndefined();
    });

    it('invoice not draft has no route', () => {
      const action = blockerNextAction('invoice_not_draft');
      expect(action.label).toBe('Invoice not draft');
      expect(action.route).toBeUndefined();
    });

    it('unknown code returns humanized label with no route', () => {
      const action = blockerNextAction('some_new_blocker');
      expect(action.label).toBe('Some New Blocker');
      expect(action.route).toBeUndefined();
    });
  });
});
