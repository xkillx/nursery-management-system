import { ParentInvoiceListItem, ChildInvoiceGroup } from '../models/parent-invoices.models';

export function formatGbp(minorUnits: number): string {
  const pounds = minorUnits / 100;
  return `£${pounds.toLocaleString('en-GB', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

export function formatBillingMonthLabel(month: string): string {
  const [year, mon] = month.split('-');
  const date = new Date(Number(year), Number(mon) - 1, 1);
  return date.toLocaleDateString('en-GB', { month: 'long', year: 'numeric' });
}

export function formatMinutes(minutes: number): string {
  if (minutes === 0) return '0h 0m';
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h === 0) return `${m}m`;
  if (m === 0) return `${h}h`;
  return `${h}h ${m}m`;
}

export function formatInstant(iso: string | null): string {
  if (!iso) return '';
  const d = new Date(iso);
  return new Intl.DateTimeFormat('en-GB', {
    timeZone: 'Europe/London',
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(d);
}

export function lineKindLabel(kind: string): string {
  return kind
    .split('_')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ');
}

export function balanceDueMinor(invoice: { totalDueMinor: number; amountPaidMinor: number }): number {
  return Math.max(0, invoice.totalDueMinor - invoice.amountPaidMinor);
}

export function isPayableInvoice(invoice: { status: string; totalDueMinor: number; amountPaidMinor: number }): boolean {
  if (invoice.status !== 'issued' && invoice.status !== 'payment_failed' && invoice.status !== 'overdue') {
    return false;
  }
  return invoice.totalDueMinor - invoice.amountPaidMinor > 0;
}

export function attentionPriority(invoice: ParentInvoiceListItem): number {
  if (invoice.status === 'overdue' || invoice.dueStatus === 'overdue') return 0;
  if (invoice.status === 'payment_failed') return 1;
  if (isPayableInvoice(invoice)) return 2;
  return 3;
}

export function isAttentionInvoice(invoice: ParentInvoiceListItem): boolean {
  if (invoice.status === 'overdue' || invoice.dueStatus === 'overdue') return true;
  if (invoice.status === 'payment_failed') return true;
  if (isPayableInvoice(invoice) && invoice.status === 'issued') return true;
  return false;
}

function compareAsc(a: string | null, b: string | null): number {
  if (!a && !b) return 0;
  if (!a) return 1;
  if (!b) return -1;
  return a < b ? -1 : a > b ? 1 : 0;
}

function compareDesc(a: string | null, b: string | null): number {
  if (!a && !b) return 0;
  if (!a) return 1;
  if (!b) return -1;
  return a < b ? 1 : a > b ? -1 : 0;
}


export function sortAttentionInvoices(a: ParentInvoiceListItem, b: ParentInvoiceListItem): number {
  const pa = attentionPriority(a);
  const pb = attentionPriority(b);
  if (pa !== pb) return pa - pb;

  const dueDateComp = compareAsc(a.dueAt, b.dueAt);
  if (dueDateComp !== 0) return dueDateComp;

  const monthComp = compareDesc(a.billingMonth, b.billingMonth);
  if (monthComp !== 0) return monthComp;

  const nameComp = a.childName.localeCompare(b.childName);
  if (nameComp !== 0) return nameComp;

  return a.invoiceId.localeCompare(b.invoiceId);
}

export function sortHistoryInvoices(a: ParentInvoiceListItem, b: ParentInvoiceListItem): number {
  const monthComp = compareDesc(a.billingMonth, b.billingMonth);
  if (monthComp !== 0) return monthComp;

  const dueComp = compareDesc(a.dueAt, b.dueAt);
  if (dueComp !== 0) return dueComp;

  return a.invoiceId.localeCompare(b.invoiceId);
}

export function groupHistoryByChild(items: ParentInvoiceListItem[]): ChildInvoiceGroup[] {
  const map = new Map<string, ParentInvoiceListItem[]>();
  for (const inv of items) {
    const key = inv.childId;
    let group = map.get(key);
    if (!group) {
      group = [];
      map.set(key, group);
    }
    group.push(inv);
  }

  const groups: ChildInvoiceGroup[] = [];
  for (const [childId, invoices] of map) {
    const childName = invoices[0].childName;
    groups.push({
      childId,
      childName,
      invoices: invoices.sort(sortHistoryInvoices),
    });
  }

  groups.sort((a, b) => a.childName.localeCompare(b.childName));
  return groups;
}
