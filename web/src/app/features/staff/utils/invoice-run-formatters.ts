import { formatGbp } from '../pages/manager-dashboard/manager-dashboard.models';
import { InvoiceRunBlockerCode } from '../models/invoice-run.models';

export { formatGbp };

export function formatMinutes(minutes: number): string {
  if (minutes === 0) return '0h 0m';
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h === 0) return `${m}m`;
  if (m === 0) return `${h}h`;
  return `${h}h ${m}m`;
}

export function formatBillingMonthLabel(month: string): string {
  const [year, mon] = month.split('-');
  const date = new Date(Number(year), Number(mon) - 1, 1);
  return date.toLocaleDateString('en-GB', { month: 'long', year: 'numeric' });
}

export function defaultCompletedBillingMonth(now: Date = new Date()): string {
  const formatter = new Intl.DateTimeFormat('en-GB', {
    timeZone: 'Europe/London',
    year: 'numeric',
    month: '2-digit',
  });
  const parts = formatter.formatToParts(now);
  const year = parts.find(p => p.type === 'year')!.value;
  const month = parts.find(p => p.type === 'month')!.value;

  const y = parseInt(year, 10);
  const m = parseInt(month, 10);

  const prevM = m === 1 ? 12 : m - 1;
  const prevY = m === 1 ? y - 1 : y;

  return `${prevY}-${String(prevM).padStart(2, '0')}`;
}

export interface BlockerNextAction {
  label: string;
  route?: string[];
  queryParams?: Record<string, string>;
}

const BLOCKER_LABELS: Record<InvoiceRunBlockerCode, string> = {
  incomplete_attendance: 'Incomplete attendance',
  missing_funding_profile: 'Missing funding profile',
  missing_core_hourly_rate: 'Missing core hourly rate',
  missing_guardian_link: 'Missing guardian link',
  existing_issued_invoice: 'Already issued',
};

export function blockerLabel(code: InvoiceRunBlockerCode): string {
  return BLOCKER_LABELS[code];
}

export function blockerNextAction(
  blockerCode: InvoiceRunBlockerCode,
  childId?: string,
  billingMonth?: string,
): BlockerNextAction {
  switch (blockerCode) {
    case 'incomplete_attendance':
      return {
        label: 'Correct attendance',
        route: ['/staff/manager/attendance-corrections'],
      };
    case 'missing_funding_profile':
      if (childId) {
        return {
          label: 'Review funding',
          route: ['/staff/manager/children', childId],
          queryParams: billingMonth ? { billing_month: billingMonth } : undefined,
        };
      }
      return {
        label: 'Funding overview',
        route: ['/staff/manager/funding'],
      };
    case 'missing_core_hourly_rate':
    case 'missing_guardian_link':
      return {
        label: 'Review child',
        route: childId ? ['/staff/manager/children', childId] : ['/staff/manager/children'],
      };
    case 'existing_issued_invoice':
      return { label: 'Already issued' };
  }
}
