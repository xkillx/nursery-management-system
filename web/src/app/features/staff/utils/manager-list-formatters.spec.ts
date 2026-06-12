import {
  formatHourlyRateGbp,
  formatSiteRate,
  minorToPounds,
  missingRequirementLabel,
  poundsToMinor,
  statusFilterLabel,
} from './manager-list-formatters';

describe('manager-list-formatters', () => {
  describe('formatHourlyRateGbp', () => {
    it('formats 750 minor units as £7.50/hr', () => {
      expect(formatHourlyRateGbp(750)).toBe('£7.50/hr');
    });

    it('formats 0 as £0.00/hr', () => {
      expect(formatHourlyRateGbp(0)).toBe('£0.00/hr');
    });

    it('formats 100 as £1.00/hr', () => {
      expect(formatHourlyRateGbp(100)).toBe('£1.00/hr');
    });

    it('formats 1255 as £12.55/hr', () => {
      expect(formatHourlyRateGbp(1255)).toBe('£12.55/hr');
    });
  });

  describe('minorToPounds', () => {
    it('converts 750 to 7.5', () => {
      expect(minorToPounds(750)).toBe(7.5);
    });

    it('converts 0 to 0', () => {
      expect(minorToPounds(0)).toBe(0);
    });
  });

  describe('poundsToMinor', () => {
    it('converts 7.5 to 750', () => {
      expect(poundsToMinor(7.5)).toBe(750);
    });

    it('converts string "7.50" to 750', () => {
      expect(poundsToMinor('7.50')).toBe(750);
    });

    it('rounds 7.555 to 756', () => {
      expect(poundsToMinor(7.555)).toBe(756);
    });

    it('converts 0 to 0', () => {
      expect(poundsToMinor(0)).toBe(0);
    });
  });

  describe('statusFilterLabel', () => {
    it('returns Active for active', () => {
      expect(statusFilterLabel('active')).toBe('Active');
    });

    it('returns Inactive for inactive', () => {
      expect(statusFilterLabel('inactive')).toBe('Inactive');
    });

    it('returns All for all', () => {
      expect(statusFilterLabel('all')).toBe('All');
    });
  });

  describe('formatSiteRate', () => {
    it('formats 750 minor units as £7.50/hr', () => {
      expect(formatSiteRate(750)).toBe('£7.50/hr');
    });

    it('formats 0 as £0.00/hr', () => {
      expect(formatSiteRate(0)).toBe('£0.00/hr');
    });

    it('returns Not set for null', () => {
      expect(formatSiteRate(null)).toBe('Not set');
    });

    it('returns Not set for undefined', () => {
      expect(formatSiteRate(undefined)).toBe('Not set');
    });
  });

  describe('missingRequirementLabel', () => {
    it('maps guardian_link to Linked guardian', () => {
      expect(missingRequirementLabel('guardian_link')).toBe('Linked guardian');
    });

    it('maps full_name to Full name', () => {
      expect(missingRequirementLabel('full_name')).toBe('Full name');
    });

    it('maps date_of_birth to Date of birth', () => {
      expect(missingRequirementLabel('date_of_birth')).toBe('Date of birth');
    });

    it('maps start_date to Start date', () => {
      expect(missingRequirementLabel('start_date')).toBe('Start date');
    });

    it('returns raw code for billing_rate (no longer a recognized label)', () => {
      expect(missingRequirementLabel('billing_rate')).toBe('billing_rate');
    });

    it('returns raw code for unknown codes', () => {
      expect(missingRequirementLabel('unknown_code')).toBe('unknown_code');
    });
  });
});
