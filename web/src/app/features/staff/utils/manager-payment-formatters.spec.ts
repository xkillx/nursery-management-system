import {
  isOpenPaymentAttempt,
  canShowParentRetry,
  getPaymentDisplayState,
  paymentDisplayLabel,
  attemptStatusLabel,
  eventOutcomeLabel,
  webhookStatusLabel,
  retryReasonLabel,
  paymentAttemptSummary,
} from './manager-payment-formatters';
import type { ManagerPaymentStatus, PaymentAttempt } from '../models/manager-invoices.models';

describe('manager-payment-formatters', () => {
  describe('isOpenPaymentAttempt', () => {
    it('returns true for checkout_creation_started', () => {
      expect(isOpenPaymentAttempt('checkout_creation_started')).toBe(true);
    });

    it('returns true for checkout_created', () => {
      expect(isOpenPaymentAttempt('checkout_created')).toBe(true);
    });

    it('returns false for paid', () => {
      expect(isOpenPaymentAttempt('paid')).toBe(false);
    });

    it('returns false for payment_failed', () => {
      expect(isOpenPaymentAttempt('payment_failed')).toBe(false);
    });

    it('returns false for null', () => {
      expect(isOpenPaymentAttempt(null)).toBe(false);
    });

    it('returns false for undefined', () => {
      expect(isOpenPaymentAttempt(undefined)).toBe(false);
    });
  });

  describe('canShowParentRetry', () => {
    function makeStatus(overrides: Partial<ManagerPaymentStatus> = {}): ManagerPaymentStatus {
      return {
        invoiceId: 'inv-1',
        status: 'issued',
        dueStatus: 'due',
        currencyCode: 'gbp',
        totalDueMinor: 24000,
        amountPaidMinor: 0,
        paidAt: null,
        paymentFailedAt: null,
        paymentStatusUpdatedAt: null,
        checkoutRetryAvailable: true,
        checkoutRetryReasonCode: 'no_payment_collected',
        latestPaymentAttempt: null,
        latestPaymentEvent: null,
        ...overrides,
      };
    }

    it('returns true when retry available and no open attempt', () => {
      expect(canShowParentRetry(makeStatus())).toBe(true);
    });

    it('returns false when retry not available', () => {
      expect(canShowParentRetry(makeStatus({ checkoutRetryAvailable: false }))).toBe(false);
    });

    it('returns false when latest attempt is checkout_created', () => {
      const attempt: PaymentAttempt = {
        paymentAttemptId: 'pa-1',
        status: 'checkout_created',
        amountMinor: 24000,
        currencyCode: 'gbp',
        stripeCheckoutSessionId: 'cs_1',
        stripePaymentIntentId: null,
        stripeExpiresAt: null,
        failureReason: null,
        providerErrorCode: null,
        providerErrorMessage: null,
        createdAt: '2026-06-09T14:00:00Z',
        updatedAt: '2026-06-09T14:00:00Z',
      };
      expect(canShowParentRetry(makeStatus({ latestPaymentAttempt: attempt }))).toBe(false);
    });

    it('returns false when latest attempt is checkout_creation_started', () => {
      const attempt: PaymentAttempt = {
        paymentAttemptId: 'pa-1',
        status: 'checkout_creation_started',
        amountMinor: 24000,
        currencyCode: 'gbp',
        stripeCheckoutSessionId: null,
        stripePaymentIntentId: null,
        stripeExpiresAt: null,
        failureReason: null,
        providerErrorCode: null,
        providerErrorMessage: null,
        createdAt: '2026-06-09T14:00:00Z',
        updatedAt: '2026-06-09T14:00:00Z',
      };
      expect(canShowParentRetry(makeStatus({ latestPaymentAttempt: attempt }))).toBe(false);
    });

    it('returns true when retry available and latest attempt is paid', () => {
      const attempt: PaymentAttempt = {
        paymentAttemptId: 'pa-1',
        status: 'paid',
        amountMinor: 24000,
        currencyCode: 'gbp',
        stripeCheckoutSessionId: 'cs_1',
        stripePaymentIntentId: 'pi_1',
        stripeExpiresAt: null,
        failureReason: null,
        providerErrorCode: null,
        providerErrorMessage: null,
        createdAt: '2026-06-09T14:00:00Z',
        updatedAt: '2026-06-09T15:00:00Z',
      };
      expect(canShowParentRetry(makeStatus({ latestPaymentAttempt: attempt }))).toBe(true);
    });
  });

  describe('getPaymentDisplayState', () => {
    it('returns not_issued for draft', () => {
      expect(getPaymentDisplayState('draft', 'not_due', 0, null)).toBe('not_issued');
    });

    it('returns paid for paid status', () => {
      expect(getPaymentDisplayState('paid', 'paid', 24000, null)).toBe('paid');
    });

    it('returns payment_failed for payment_failed status', () => {
      expect(getPaymentDisplayState('payment_failed', 'due', 0, null)).toBe('payment_failed');
    });

    it('returns awaiting_provider_update for open attempt', () => {
      expect(getPaymentDisplayState('issued', 'due', 0, 'checkout_created')).toBe('awaiting_provider_update');
    });

    it('returns unpaid for issued with zero paid', () => {
      expect(getPaymentDisplayState('issued', 'due', 0, null)).toBe('unpaid');
    });

    it('returns unpaid_overdue for overdue with zero paid', () => {
      expect(getPaymentDisplayState('overdue', 'overdue', 0, null)).toBe('unpaid_overdue');
    });

    it('returns no_payment_due for issued with partial payment', () => {
      expect(getPaymentDisplayState('issued', 'due', 10000, null)).toBe('no_payment_due');
    });

    it('open attempt takes priority over overdue', () => {
      expect(getPaymentDisplayState('overdue', 'overdue', 0, 'checkout_creation_started')).toBe('awaiting_provider_update');
    });
  });

  describe('paymentDisplayLabel', () => {
    it('returns human-readable labels for all states', () => {
      expect(paymentDisplayLabel('paid')).toBe('Paid');
      expect(paymentDisplayLabel('payment_failed')).toBe('Payment failed');
      expect(paymentDisplayLabel('unpaid')).toBe('Unpaid');
      expect(paymentDisplayLabel('unpaid_overdue')).toBe('Unpaid');
      expect(paymentDisplayLabel('awaiting_provider_update')).toBe('Awaiting provider update');
      expect(paymentDisplayLabel('not_issued')).toBe('Not issued');
      expect(paymentDisplayLabel('no_payment_due')).toBe('No payment due');
    });
  });

  describe('attemptStatusLabel', () => {
    it('returns known labels', () => {
      expect(attemptStatusLabel('checkout_created')).toBe('Checkout session created');
      expect(attemptStatusLabel('paid')).toBe('Paid');
      expect(attemptStatusLabel('payment_failed')).toBe('Payment failed');
      expect(attemptStatusLabel('expired')).toBe('Expired');
    });

    it('falls back to humanized code for unknown', () => {
      expect(attemptStatusLabel('some_new_status')).toBe('Some new status');
    });
  });

  describe('eventOutcomeLabel', () => {
    it('returns known labels', () => {
      expect(eventOutcomeLabel('payment_succeeded')).toBe('Payment succeeded');
      expect(eventOutcomeLabel('checkout_expired')).toBe('Checkout expired');
      expect(eventOutcomeLabel('webhook_processing_failed')).toBe('Webhook processing failed');
    });

    it('falls back to humanized code', () => {
      expect(eventOutcomeLabel('new_outcome_type')).toBe('New outcome type');
    });
  });

  describe('webhookStatusLabel', () => {
    it('returns known labels', () => {
      expect(webhookStatusLabel('processed')).toBe('Processed');
      expect(webhookStatusLabel('failed')).toBe('Failed');
      expect(webhookStatusLabel('pending')).toBe('Pending');
    });
  });

  describe('retryReasonLabel', () => {
    it('returns known labels', () => {
      expect(retryReasonLabel('no_payment_collected')).toBe('No payment collected yet');
      expect(retryReasonLabel('already_paid')).toBe('Already paid');
      expect(retryReasonLabel('open_checkout_attempt')).toBe('Awaiting provider update');
    });

    it('falls back to humanized code', () => {
      expect(retryReasonLabel('unknown_reason')).toBe('Unknown reason');
    });
  });

  describe('paymentAttemptSummary', () => {
    it('returns empty string for null attempt', () => {
      expect(paymentAttemptSummary(null)).toBe('');
    });

    it('returns status label for zero amount', () => {
      const attempt: PaymentAttempt = {
        paymentAttemptId: 'pa-1',
        status: 'checkout_created',
        amountMinor: 0,
        currencyCode: 'gbp',
        stripeCheckoutSessionId: null,
        stripePaymentIntentId: null,
        stripeExpiresAt: null,
        failureReason: null,
        providerErrorCode: null,
        providerErrorMessage: null,
        createdAt: '2026-06-09T14:00:00Z',
        updatedAt: '2026-06-09T14:00:00Z',
      };
      expect(paymentAttemptSummary(attempt)).toBe('Checkout session created');
    });

    it('returns status with amount', () => {
      const attempt: PaymentAttempt = {
        paymentAttemptId: 'pa-1',
        status: 'paid',
        amountMinor: 24000,
        currencyCode: 'gbp',
        stripeCheckoutSessionId: 'cs_1',
        stripePaymentIntentId: 'pi_1',
        stripeExpiresAt: null,
        failureReason: null,
        providerErrorCode: null,
        providerErrorMessage: null,
        createdAt: '2026-06-09T14:00:00Z',
        updatedAt: '2026-06-09T15:00:00Z',
      };
      expect(paymentAttemptSummary(attempt)).toBe('Paid — £240.00');
    });
  });
});
