import type { ManagerPaymentStatus, PaymentAttempt } from '../models/manager-invoices.models';

const OPEN_ATTEMPT_STATUSES = new Set(['checkout_creation_started', 'checkout_created']);

export function isOpenPaymentAttempt(status: string | null | undefined): boolean {
  return !!status && OPEN_ATTEMPT_STATUSES.has(status);
}

export function canShowParentRetry(status: ManagerPaymentStatus): boolean {
  if (!status.checkoutRetryAvailable) return false;
  if (isOpenPaymentAttempt(status.latestPaymentAttempt?.status)) return false;
  return true;
}

export type PaymentDisplayState =
  | 'paid'
  | 'payment_failed'
  | 'unpaid'
  | 'unpaid_overdue'
  | 'awaiting_provider_update'
  | 'not_issued'
  | 'no_payment_due';

export function getPaymentDisplayState(
  invoiceStatus: string,
  dueStatus: string,
  amountPaidMinor: number,
  latestAttemptStatus: string | null | undefined,
): PaymentDisplayState {
  if (invoiceStatus === 'draft') return 'not_issued';
  if (invoiceStatus === 'paid') return 'paid';
  if (invoiceStatus === 'payment_failed') return 'payment_failed';
  if (isOpenPaymentAttempt(latestAttemptStatus)) return 'awaiting_provider_update';
  if ((invoiceStatus === 'issued' || invoiceStatus === 'overdue') && amountPaidMinor === 0) {
    return dueStatus === 'overdue' ? 'unpaid_overdue' : 'unpaid';
  }
  return 'no_payment_due';
}

export function paymentDisplayLabel(state: PaymentDisplayState): string {
  switch (state) {
    case 'paid': return 'Paid';
    case 'payment_failed': return 'Payment failed';
    case 'unpaid': return 'Unpaid';
    case 'unpaid_overdue': return 'Unpaid';
    case 'awaiting_provider_update': return 'Awaiting provider update';
    case 'not_issued': return 'Not issued';
    case 'no_payment_due': return 'No payment due';
  }
}

const ATTEMPT_STATUS_LABELS: Record<string, string> = {
  checkout_creation_started: 'Checkout creation started',
  checkout_created: 'Checkout session created',
  checkout_creation_failed: 'Checkout creation failed',
  paid: 'Paid',
  payment_failed: 'Payment failed',
  cancelled: 'Cancelled',
  expired: 'Expired',
};

export function attemptStatusLabel(status: string): string {
  return ATTEMPT_STATUS_LABELS[status] ?? humanizeCode(status);
}

const EVENT_OUTCOME_LABELS: Record<string, string> = {
  checkout_session_created: 'Checkout session created',
  checkout_completed: 'Checkout completed',
  checkout_expired: 'Checkout expired',
  payment_succeeded: 'Payment succeeded',
  payment_failed: 'Payment failed',
  payment_cancelled: 'Payment cancelled',
  webhook_processing_failed: 'Webhook processing failed',
  checkout_creation_failed: 'Checkout creation failed',
};

export function eventOutcomeLabel(outcome: string): string {
  return EVENT_OUTCOME_LABELS[outcome] ?? humanizeCode(outcome);
}

const WEBHOOK_STATUS_LABELS: Record<string, string> = {
  pending: 'Pending',
  processed: 'Processed',
  failed: 'Failed',
  skipped: 'Skipped',
};

export function webhookStatusLabel(status: string): string {
  return WEBHOOK_STATUS_LABELS[status] ?? humanizeCode(status);
}

const RETRY_REASON_LABELS: Record<string, string> = {
  no_payment_collected: 'No payment collected yet',
  already_paid: 'Already paid',
  zero_total_invoice: 'Zero-total invoice',
  not_issued: 'Invoice not issued',
  partial_payment: 'Partial payment received',
  non_monthly_invoice: 'Non-monthly invoice',
  unsupported_currency: 'Unsupported currency',
  open_checkout_attempt: 'Awaiting provider update',
};

export function retryReasonLabel(code: string): string {
  return RETRY_REASON_LABELS[code] ?? humanizeCode(code);
}

export function paymentAttemptSummary(attempt: PaymentAttempt | null): string {
  if (!attempt) return '';
  const parts = [attemptStatusLabel(attempt.status)];
  if (attempt.amountMinor > 0) parts.push(`£${(attempt.amountMinor / 100).toFixed(2)}`);
  return parts.join(' — ');
}

function humanizeCode(code: string): string {
  const words = code.replace(/[-_]+/g, ' ').trim().split(' ');
  if (words.length === 0) return '';
  return [
    words[0].charAt(0).toUpperCase() + words[0].slice(1),
    ...words.slice(1),
  ].join(' ');
}
