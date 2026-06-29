import { formatGbp } from '../pages/manager-dashboard/manager-dashboard.models';

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

const BLOCKER_LABELS: Record<string, string> = {
  incomplete_attendance: 'Incomplete attendance',
  missing_funding_profile: 'Missing funding profile',
  missing_billing_rate: 'Missing site billing rate',
  missing_child_name: 'Missing child name',
  missing_child_date_of_birth: 'Missing date of birth',
  missing_child_start_date: 'Missing start date',
  missing_parent_carer_contact: 'Missing parent carer contact',
  invoice_already_issued: 'Already issued',
  child_not_found: 'Child not found',
  child_not_in_billing_month: 'Child not in billing month',
  invoice_not_found: 'Invoice not found',
  invoice_not_in_billing_month: 'Invoice not in billing month',
  invoice_not_draft: 'Invoice not draft',
  invoice_not_monthly: 'Invoice not monthly',
};

function humanizeCode(code: string): string {
  return code
    .split('_')
    .map(w => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ');
}

export function blockerLabel(code: string): string {
  return BLOCKER_LABELS[code] ?? humanizeCode(code);
}

export function blockerNextAction(
  blockerCode: string,
  childId?: string,
  billingMonth?: string,
): BlockerNextAction {
  switch (blockerCode) {
    case 'incomplete_attendance':
      return {
        label: 'Correct attendance',
        route: ['/manager/attendance-corrections'],
      };
    case 'missing_funding_profile':
      if (childId) {
        return {
          label: 'Review funding',
          route: ['/manager/children', childId],
          queryParams: billingMonth ? { billing_month: billingMonth } : undefined,
        };
      }
      return { label: 'Funding overview' };
    case 'missing_billing_rate':
    case 'missing_child_name':
    case 'missing_child_date_of_birth':
    case 'missing_child_start_date':
    case 'missing_parent_carer_contact':
      return {
        label: 'Review child',
        route: childId ? ['/manager/children', childId] : ['/manager/children'],
      };
    case 'invoice_already_issued':
    case 'invoice_not_draft':
    case 'invoice_not_monthly':
    case 'invoice_not_found':
    case 'invoice_not_in_billing_month':
    case 'child_not_found':
    case 'child_not_in_billing_month':
      return { label: blockerLabel(blockerCode) };
    default:
      return { label: humanizeCode(blockerCode) };
  }
}
