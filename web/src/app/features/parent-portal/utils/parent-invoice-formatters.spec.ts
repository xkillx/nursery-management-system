import {
  balanceDueMinor,
  isPayableInvoice,
  attentionPriority,
  isAttentionInvoice,
  sortAttentionInvoices,
  sortHistoryInvoices,
  groupHistoryByChild,
  formatGbp,
  formatBillingMonthLabel,
  formatMinutes,
} from './parent-invoice-formatters';
import { ParentInvoiceListItem } from '../models/parent-invoices.models';

function makeItem(overrides: Partial<Omit<ParentInvoiceListItem, 'invoiceId' | 'status'>> & { invoiceId: string; status?: ParentInvoiceListItem['status'] }): ParentInvoiceListItem {
  return {
    invoiceKind: 'monthly',
    invoiceNumber: null,
    invoiceNumberDisplay: `INV-${overrides.invoiceId}`,
    childId: 'child-1',
    childName: 'Ada Lovelace',
    billingMonth: '2026-05',
    period: null,
    status: 'issued',
    dueStatus: 'due',
    currencyCode: 'gbp',
    subtotalMinor: 45000,
    fundedDeductionMinor: 0,
    totalDueMinor: 45000,
    amountPaidMinor: 0,
    dueAt: '2026-06-01T00:00:00Z',
    issuedAt: '2026-05-28T00:00:00Z',
    paidAt: null,
    paymentFailedAt: null,
    paymentStatusUpdatedAt: null,
    ...overrides,
  };
}

describe('parent-invoice-formatters', () => {
  describe('formatGbp', () => {
    it('formats zero', () => expect(formatGbp(0)).toBe('£0.00'));
    it('formats positive', () => expect(formatGbp(45000)).toBe('£450.00'));
    it('formats with pence', () => expect(formatGbp(1250)).toBe('£12.50'));
  });

  describe('formatBillingMonthLabel', () => {
    it('formats month', () => expect(formatBillingMonthLabel('2026-05')).toBe('May 2026'));
    it('formats january', () => expect(formatBillingMonthLabel('2026-01')).toBe('January 2026'));
  });

  describe('formatMinutes', () => {
    it('zero', () => expect(formatMinutes(0)).toBe('0h 0m'));
    it('minutes only', () => expect(formatMinutes(45)).toBe('45m'));
    it('hours only', () => expect(formatMinutes(120)).toBe('2h'));
    it('mixed', () => expect(formatMinutes(135)).toBe('2h 15m'));
  });

  describe('balanceDueMinor', () => {
    it('unpaid', () => expect(balanceDueMinor({ totalDueMinor: 45000, amountPaidMinor: 0 })).toBe(45000));
    it('partially paid', () => expect(balanceDueMinor({ totalDueMinor: 45000, amountPaidMinor: 20000 })).toBe(25000));
    it('fully paid clamps to zero', () => expect(balanceDueMinor({ totalDueMinor: 45000, amountPaidMinor: 45000 })).toBe(0));
  });

  describe('isPayableInvoice', () => {
    it('issued with positive balance', () => {
      expect(isPayableInvoice(makeItem({ invoiceId: '1', status: 'issued', totalDueMinor: 45000, amountPaidMinor: 0 }))).toBeTrue();
    });

    it('issued with zero balance', () => {
      expect(isPayableInvoice(makeItem({ invoiceId: '2', status: 'issued', totalDueMinor: 0, amountPaidMinor: 0 }))).toBeFalse();
    });

    it('overdue with positive balance', () => {
      expect(isPayableInvoice(makeItem({ invoiceId: '3', status: 'overdue', totalDueMinor: 45000, amountPaidMinor: 0 }))).toBeTrue();
    });

    it('payment_failed with positive balance', () => {
      expect(isPayableInvoice(makeItem({ invoiceId: '4', status: 'payment_failed', totalDueMinor: 45000, amountPaidMinor: 0 }))).toBeTrue();
    });

    it('paid is not payable', () => {
      expect(isPayableInvoice(makeItem({ invoiceId: '5', status: 'paid', totalDueMinor: 45000, amountPaidMinor: 45000 }))).toBeFalse();
    });
  });

  describe('attentionPriority', () => {
    it('overdue status gets priority 0', () => {
      expect(attentionPriority(makeItem({ invoiceId: '1', status: 'overdue' }))).toBe(0);
    });

    it('overdue dueStatus gets priority 0', () => {
      expect(attentionPriority(makeItem({ invoiceId: '2', status: 'issued', dueStatus: 'overdue' }))).toBe(0);
    });

    it('payment_failed gets priority 1', () => {
      expect(attentionPriority(makeItem({ invoiceId: '3', status: 'payment_failed' }))).toBe(1);
    });

    it('payable issued gets priority 2', () => {
      expect(attentionPriority(makeItem({ invoiceId: '4', status: 'issued', totalDueMinor: 45000, amountPaidMinor: 0 }))).toBe(2);
    });

    it('paid gets priority 3', () => {
      expect(attentionPriority(makeItem({ invoiceId: '5', status: 'paid' }))).toBe(3);
    });
  });

  describe('isAttentionInvoice', () => {
    it('overdue status', () => {
      expect(isAttentionInvoice(makeItem({ invoiceId: '1', status: 'overdue' }))).toBeTrue();
    });

    it('overdue due status', () => {
      expect(isAttentionInvoice(makeItem({ invoiceId: '2', status: 'issued', dueStatus: 'overdue' }))).toBeTrue();
    });

    it('payment_failed', () => {
      expect(isAttentionInvoice(makeItem({ invoiceId: '3', status: 'payment_failed' }))).toBeTrue();
    });

    it('payable issued', () => {
      expect(isAttentionInvoice(makeItem({ invoiceId: '4', status: 'issued', totalDueMinor: 45000, amountPaidMinor: 0 }))).toBeTrue();
    });

    it('zero-balance issued is not attention', () => {
      expect(isAttentionInvoice(makeItem({ invoiceId: '5', status: 'issued', totalDueMinor: 0, amountPaidMinor: 0 }))).toBeFalse();
    });

    it('paid is not attention', () => {
      expect(isAttentionInvoice(makeItem({ invoiceId: '6', status: 'paid' }))).toBeFalse();
    });
  });

  describe('sortAttentionInvoices', () => {
    it('orders overdue before payment_failed before payable issued', () => {
      const overdue = makeItem({ invoiceId: 'c', status: 'overdue' });
      const failed = makeItem({ invoiceId: 'b', status: 'payment_failed' });
      const issued = makeItem({ invoiceId: 'a', status: 'issued', totalDueMinor: 45000, amountPaidMinor: 0 });

      const sorted = [issued, failed, overdue].sort(sortAttentionInvoices);
      expect(sorted[0].invoiceId).toBe('c');
      expect(sorted[1].invoiceId).toBe('b');
      expect(sorted[2].invoiceId).toBe('a');
    });

    it('breaks ties by due date ascending', () => {
      const later = makeItem({ invoiceId: '2', status: 'overdue', dueAt: '2026-06-10T00:00:00Z' });
      const earlier = makeItem({ invoiceId: '1', status: 'overdue', dueAt: '2026-06-01T00:00:00Z' });

      const sorted = [later, earlier].sort(sortAttentionInvoices);
      expect(sorted[0].invoiceId).toBe('1');
    });
  });

  describe('sortHistoryInvoices', () => {
    it('sorts newest billing month first', () => {
      const may = makeItem({ invoiceId: '1', billingMonth: '2026-05' });
      const june = makeItem({ invoiceId: '2', billingMonth: '2026-06' });

      const sorted = [may, june].sort(sortHistoryInvoices);
      expect(sorted[0].billingMonth).toBe('2026-06');
    });
  });

  describe('groupHistoryByChild', () => {
    it('groups by child and sorts groups by name', () => {
      const items = [
        makeItem({ invoiceId: '1', childId: 'c2', childName: 'Zara', billingMonth: '2026-05' }),
        makeItem({ invoiceId: '2', childId: 'c1', childName: 'Ada', billingMonth: '2026-05' }),
        makeItem({ invoiceId: '3', childId: 'c1', childName: 'Ada', billingMonth: '2026-04' }),
      ];

      const groups = groupHistoryByChild(items);
      expect(groups.length).toBe(2);
      expect(groups[0].childName).toBe('Ada');
      expect(groups[0].invoices.length).toBe(2);
      expect(groups[0].invoices[0].billingMonth).toBe('2026-05');
      expect(groups[1].childName).toBe('Zara');
    });

    it('returns empty for no items', () => {
      expect(groupHistoryByChild([])).toEqual([]);
    });
  });
});
