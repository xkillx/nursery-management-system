import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { ManagerInvoicesApiService } from './manager-invoices-api.service';

describe('ManagerInvoicesApiService', () => {
  let service: ManagerInvoicesApiService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [ManagerInvoicesApiService],
    });
    service = TestBed.inject(ManagerInvoicesApiService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  describe('listInvoices', () => {
    it('omits status query param when filter is all', () => {
      const response = {
        items: [
          {
            invoice_id: 'inv-1',
            invoice_kind: 'monthly',
            invoice_number: null,
            invoice_number_display: 'Draft invoice',
            child_id: 'c1',child_first_name: 'Ben',
child_middle_name: null,
child_last_name: null,
            billing_month: '2026-05',
            status: 'draft',
            due_status: 'not_due',
            currency_code: 'gbp',
            subtotal_minor: 33000,
            funded_deduction_minor: 9000,
            total_due_minor: 24000,
            amount_paid_minor: 0,
            due_at: null,
            issued_at: null,
            paid_at: null,
            payment_failed_at: null,
            payment_status_updated_at: null,
            generated_run_id: null,
            generated_run_status: null,
            generated_run_started_at: null,
            generated_run_completed_at: null,
            generated_run_exception_count: null,
            created_at: '2026-06-09T10:00:00Z',
            updated_at: '2026-06-09T10:00:00Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 50,
      };

      service.listInvoices({ billingMonth: '2026-05', status: 'all', limit: 50, offset: 0 }).subscribe((result) => {
        expect(result.items.length).toBe(1);
        expect(result.items[0].invoiceId).toBe('inv-1');
        expect(result.items[0].invoiceNumberDisplay).toBe('Draft invoice');
        expect(result.items[0].dueStatus).toBe('not_due');
        expect(result.total).toBe(1);
        expect(result.page).toBe(1);
        expect(result.page_size).toBe(50);
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices');
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('billing_month')).toBe('2026-05');
      expect(req.request.params.get('status')).toBeNull();
      expect(req.request.params.get('limit')).toBe('50');
      expect(req.request.params.get('offset')).toBe('0');
      req.flush(response);
    });

    it('includes status query param for concrete filter', () => {
      service.listInvoices({ billingMonth: '2026-05', status: 'issued', limit: 50, offset: 0 }).subscribe();

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices');
      expect(req.request.params.get('status')).toBe('issued');
      req.flush({ items: [], total: 0, page: 1, page_size: 50 });
    });

    it('sends limit and offset for pagination', () => {
      service.listInvoices({ billingMonth: '2026-05', status: 'all', limit: 50, offset: 50 }).subscribe();

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices');
      expect(req.request.params.get('limit')).toBe('50');
      expect(req.request.params.get('offset')).toBe('50');
      req.flush({ items: [], total: 0, page: 2, page_size: 50 });
    });

    it('maps camelCase fields from snake_case response', () => {
      const response = {
        items: [
          {
            invoice_id: 'inv-2',
            invoice_kind: 'monthly',
            invoice_number: 'INV-202605-0001',
            invoice_number_display: 'INV-202605-0001',
            child_id: 'c2',child_first_name: 'Alice',
child_middle_name: null,
child_last_name: null,
            billing_month: '2026-05',
            period: { start_date: '2026-05-01', end_date: '2026-05-31' },
            status: 'issued',
            due_status: 'due',
            currency_code: 'gbp',
            subtotal_minor: 33000,
            funded_deduction_minor: 9000,
            total_due_minor: 24000,
            amount_paid_minor: 0,
            due_at: '2026-06-10T00:00:00Z',
            issued_at: '2026-06-09T12:00:00Z',
            paid_at: null,
            payment_failed_at: null,
            payment_status_updated_at: null,
            generated_run_id: 'run-1',
            generated_run_status: 'completed',
            generated_run_started_at: '2026-06-09T11:00:00Z',
            generated_run_completed_at: '2026-06-09T11:05:00Z',
            generated_run_exception_count: 0,
            created_at: '2026-06-09T10:00:00Z',
            updated_at: '2026-06-09T12:00:00Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 50,
      };

      service.listInvoices({ billingMonth: '2026-05', status: 'all', limit: 50, offset: 0 }).subscribe((result) => {
        const item = result.items[0];
        expect(item.childName).toBe('Alice');
        expect(item.invoiceNumber).toBe('INV-202605-0001');
        expect(item.period?.startDate).toBe('2026-05-01');
        expect(item.period?.endDate).toBe('2026-05-31');
        expect(item.generatedRunId).toBe('run-1');
        expect(item.amountPaidMinor).toBe(0);
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices');
      req.flush(response);
    });

    it('uses null-safe defaults for optional fields', () => {
      const minimalItem = {
        invoice_id: 'inv-3',
        invoice_kind: 'monthly',
        invoice_number: null,
        child_id: 'c3',child_first_name: 'Chloe',
child_middle_name: null,
child_last_name: null,
        billing_month: '2026-05',
        status: 'draft',
        subtotal_minor: 10000,
        funded_deduction_minor: 0,
        total_due_minor: 10000,
        issued_at: null,
        created_at: '2026-06-09T10:00:00Z',
        updated_at: '2026-06-09T10:00:00Z',
      };

      service.listInvoices({ billingMonth: '2026-05', status: 'all', limit: 50, offset: 0 }).subscribe((result) => {
        const item = result.items[0];
        expect(item.dueStatus).toBe('not_due');
        expect(item.amountPaidMinor).toBe(0);
        expect(item.invoiceNumberDisplay).toBe('');
        expect(item.period).toBeNull();
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices');
      req.flush({ items: [minimalItem], total: 1, page: 1, page_size: 50 });
    });

    it('maps null generated run fields to null', () => {
      const itemWithNullRun = {
        invoice_id: 'inv-4',
        invoice_kind: 'monthly',
        invoice_number: 'INV-001',
        child_id: 'c4',child_first_name: 'Dan',
child_middle_name: null,
child_last_name: null,
        billing_month: '2026-05',
        status: 'draft',
        subtotal_minor: 20000,
        funded_deduction_minor: 5000,
        total_due_minor: 15000,
        generated_run_id: null,
        generated_run_status: null,
        generated_run_started_at: null,
        generated_run_completed_at: null,
        generated_run_exception_count: null,
        issued_at: null,
        created_at: '2026-06-09T10:00:00Z',
        updated_at: '2026-06-09T10:00:00Z',
      };

      service.listInvoices({ billingMonth: '2026-05', status: 'all', limit: 50, offset: 0 }).subscribe((result) => {
        const item = result.items[0];
        expect(item.generatedRunId).toBeNull();
        expect(item.generatedRunStatus).toBeNull();
        expect(item.generatedRunStartedAt).toBeNull();
        expect(item.generatedRunCompletedAt).toBeNull();
        expect(item.generatedRunExceptionCount).toBeNull();
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices');
      req.flush({ items: [itemWithNullRun], total: 1, page: 1, page_size: 50 });
    });
  });

  describe('bulkIssueInvoices', () => {
    it('sends correct request body and maps response', () => {
      const response = {
        run_id: 'run-1',
        billing_month: '2026-05',
        status: 'completed',
        summary: { eligible_count: 3, success_count: 2, blocked_count: 1, total_due_minor: 50000 },
        issued: [
          {
            invoice_id: 'inv-1',
            child_id: 'c1',
            child_first_name: 'Ben',
            child_middle_name: null,
            child_last_name: null,
            invoice_number: 'INV-001',
            issued_at: '2026-06-09T12:00:00Z',
            due_at: '2026-06-10T00:00:00Z',
            total_due_minor: 24000,
          },
          {
            invoice_id: 'inv-2',
            child_id: 'c2',
            child_first_name: 'Alice',
            child_middle_name: null,
            child_last_name: null,
            invoice_number: 'INV-002',
            issued_at: '2026-06-09T12:00:00Z',
            due_at: '2026-06-10T00:00:00Z',
            total_due_minor: 26000,
          },
        ],
        blocked: [
          {
            invoice_id: 'inv-3',
            child_id: 'c3',
            child_first_name: 'Chloe',
            child_middle_name: null,
            child_last_name: null,
            blockers: [{ code: 'incomplete_attendance', message: 'Attendance incomplete' }],
          },
        ],
      };

      service.bulkIssueInvoices({ billingMonth: '2026-05', invoiceIds: ['inv-1', 'inv-2', 'inv-3'] }).subscribe((result) => {
        expect(result.runId).toBe('run-1');
        expect(result.billingMonth).toBe('2026-05');
        expect(result.status).toBe('completed');
        expect(result.summary.successCount).toBe(2);
        expect(result.summary.blockedCount).toBe(1);
        expect(result.issued.length).toBe(2);
        expect(result.issued[0].invoiceId).toBe('inv-1');
        expect(result.issued[0].childName).toBe('Ben');
        expect(result.blocked.length).toBe(1);
        expect(result.blocked[0].blockers[0].code).toBe('incomplete_attendance');
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices/drafts/bulk-issue');
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({
        billing_month: '2026-05',
        invoice_ids: ['inv-1', 'inv-2', 'inv-3'],
        confirm: true,
      });
      req.flush(response);
    });

    it('handles empty invoice IDs array', () => {
      const response = {
        run_id: 'run-2',
        billing_month: '2026-05',
        status: 'completed',
        summary: { eligible_count: 0, success_count: 0, blocked_count: 0, total_due_minor: 0 },
        issued: [],
        blocked: [],
      };

      service.bulkIssueInvoices({ billingMonth: '2026-05', invoiceIds: [] }).subscribe((result) => {
        expect(result.issued).toEqual([]);
        expect(result.blocked).toEqual([]);
        expect(result.summary.successCount).toBe(0);
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices/drafts/bulk-issue');
      expect(req.request.body.invoice_ids).toEqual([]);
      req.flush(response);
    });

    it('propagates HTTP errors', () => {
      service.bulkIssueInvoices({ billingMonth: '2026-05', invoiceIds: ['inv-1'] }).subscribe({
        next: () => fail('should have errored'),
        error: (err) => {
          expect(err.status).toBe(400);
        },
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices/drafts/bulk-issue');
      req.flush({ code: 'invalid_request', message: 'Bad request' }, { status: 400, statusText: 'Bad Request' });
    });
  });

  describe('getInvoice', () => {
    it('maps detail with due status, locked timestamp, calculation, and line sorting', () => {
      const detailResponse = {
        invoice_id: 'inv-1',
        invoice_kind: 'monthly',
        invoice_number: 'INV-202605-0001',
        invoice_number_display: 'INV-202605-0001',
        child_id: 'c1',child_first_name: 'Ben',
child_middle_name: null,
child_last_name: null,
        billing_month: '2026-05',
        status: 'issued',
        due_status: 'due',
        currency_code: 'gbp',
        subtotal_minor: 33000,
        funded_deduction_minor: 9000,
        total_due_minor: 24000,
        amount_paid_minor: 0,
        issued_at: '2026-06-09T12:00:00Z',
        locked_at: '2026-06-09T12:00:00Z',
        due_at: '2026-06-10T00:00:00Z',
        paid_at: null,
        payment_failed_at: null,
        payment_status_updated_at: null,
        adjusts_invoice_id: null,
        adjustment_reason_code: null,
        adjustment_reason_note: null,
        generated_run_id: 'run-1',
        generated_run_status: 'completed',
        generated_run_exception_count: 1,
        generated_run_exceptions: [
          { child_id: 'c2',child_first_name: 'Alice',
child_middle_name: null,
child_last_name: null, blocker_codes: ['incomplete_attendance'] },
        ],
        calculation: {
          core_hourly_rate_minor: 25,
          raw_attended_minutes: 1315,
          rounded_attended_minutes: 1320,
          funded_allowance_minutes: 360,
          funded_deduction_minutes: 360,
          core_billable_minutes: 960,
          included_session_count: 4,
          core_subtotal_minor: 24000,
          extras_total_minor: 9000,
        },
        lines: [
          { line_id: 'l2', line_kind: 'funded_deduction', description: 'Funded deduction', sort_order: 2, quantity_minutes: 360, unit_amount_minor: 25, line_amount_minor: -9000 },
          { line_id: 'l1', line_kind: 'core_childcare', description: 'Core childcare', sort_order: 1, quantity_minutes: 1320, unit_amount_minor: 25, line_amount_minor: 33000 },
        ],
        created_at: '2026-06-09T10:00:00Z',
        updated_at: '2026-06-09T12:00:00Z',
      };

      service.getInvoice('inv-1').subscribe((detail) => {
        expect(detail.invoiceId).toBe('inv-1');
        expect(detail.dueStatus).toBe('due');
        expect(detail.lockedAt).toBe('2026-06-09T12:00:00Z');
        expect(detail.calculation?.siteCoreHourlyRateMinor).toBe(25);
        expect(detail.calculation?.roundedAttendedMinutes).toBe(1320);
        expect(detail.calculation?.coreBillableMinutes).toBe(960);
        expect(detail.generatedRunExceptions.length).toBe(1);
        expect(detail.generatedRunExceptions[0].childName).toBe('Alice');
        expect(detail.generatedRunExceptions[0].blockerCodes).toEqual(['incomplete_attendance']);
        // Lines sorted by sort_order
        expect(detail.lines[0].lineId).toBe('l1');
        expect(detail.lines[1].lineId).toBe('l2');
        expect(detail.lines[0].lineKind).toBe('core_childcare');
      });

      const req = httpMock.expectOne('/api/v1/invoices/inv-1');
      expect(req.request.method).toBe('GET');
      req.flush(detailResponse);
    });

    it('handles null calculation and empty lines gracefully', () => {
      const detailResponse = {
        invoice_id: 'inv-2',
        invoice_kind: 'monthly',
        invoice_number: null,
        child_id: 'c1',child_first_name: 'Ben',
child_middle_name: null,
child_last_name: null,
        billing_month: '2026-05',
        status: 'draft',
        subtotal_minor: 10000,
        funded_deduction_minor: 0,
        total_due_minor: 10000,
        issued_at: null,
        created_at: '2026-06-09T10:00:00Z',
        updated_at: '2026-06-09T10:00:00Z',
      };

      service.getInvoice('inv-2').subscribe((detail) => {
        expect(detail.calculation).toBeNull();
        expect(detail.lines).toEqual([]);
        expect(detail.generatedRunExceptions).toEqual([]);
      });

      const req = httpMock.expectOne('/api/v1/invoices/inv-2');
      req.flush(detailResponse);
    });
  });

  describe('getPaymentStatus', () => {
    it('calls GET /invoices/:id/payment-status and maps camelCase fields', () => {
      const response = {
        invoice_id: 'inv-1',
        status: 'issued',
        due_status: 'due',
        currency_code: 'gbp',
        total_due_minor: 24000,
        amount_paid_minor: 0,
        paid_at: null,
        payment_failed_at: null,
        payment_status_updated_at: null,
        checkout_retry_available: true,
        checkout_retry_reason_code: 'no_payment_collected',
        latest_payment_attempt: {
          payment_attempt_id: 'pa-1',
          status: 'checkout_created',
          amount_minor: 24000,
          currency_code: 'gbp',
          stripe_checkout_session_id: 'cs_123',
          stripe_payment_intent_id: null,
          stripe_expires_at: '2026-06-10T12:00:00Z',
          failure_reason: null,
          provider_error_code: null,
          provider_error_message: null,
          created_at: '2026-06-09T14:00:00Z',
          updated_at: '2026-06-09T14:00:00Z',
        },
        latest_payment_event: {
          payment_event_id: 'pe-1',
          payment_attempt_id: 'pa-1',
          stripe_event_id: 'evt_123',
          stripe_event_type: 'checkout.session.completed',
          stripe_checkout_session_id: 'cs_123',
          stripe_payment_intent_id: null,
          outcome: 'checkout_session_created',
          reason_code: 'checkout_initiated',
          previous_invoice_status: 'issued',
          new_invoice_status: 'issued',
          attempt_previous_status: 'checkout_creation_started',
          attempt_new_status: 'checkout_created',
          amount_minor: 24000,
          currency_code: 'gbp',
          webhook_processing_status: 'processed',
          webhook_processing_reason: null,
          webhook_received_at: '2026-06-09T14:00:00Z',
          webhook_processed_at: '2026-06-09T14:00:01Z',
          created_at: '2026-06-09T14:00:00Z',
        },
      };

      service.getPaymentStatus('inv-1').subscribe((status) => {
        expect(status.invoiceId).toBe('inv-1');
        expect(status.checkoutRetryAvailable).toBe(true);
        expect(status.checkoutRetryReasonCode).toBe('no_payment_collected');
        expect(status.latestPaymentAttempt).not.toBeNull();
        expect(status.latestPaymentAttempt!.paymentAttemptId).toBe('pa-1');
        expect(status.latestPaymentAttempt!.status).toBe('checkout_created');
        expect(status.latestPaymentAttempt!.amountMinor).toBe(24000);
        expect(status.latestPaymentAttempt!.stripeCheckoutSessionId).toBe('cs_123');
        expect(status.latestPaymentEvent).not.toBeNull();
        expect(status.latestPaymentEvent!.outcome).toBe('checkout_session_created');
        expect(status.latestPaymentEvent!.webhookProcessingStatus).toBe('processed');
        expect(status.latestPaymentEvent!.previousInvoiceStatus).toBe('issued');
        expect(status.latestPaymentEvent!.newInvoiceStatus).toBe('issued');
      });

      const req = httpMock.expectOne('/api/v1/invoices/inv-1/payment-status');
      expect(req.request.method).toBe('GET');
      req.flush(response);
    });

    it('handles null latest attempt and null latest event', () => {
      const response = {
        invoice_id: 'inv-2',
        status: 'draft',
        due_status: 'not_due',
        currency_code: 'gbp',
        total_due_minor: 10000,
        amount_paid_minor: 0,
        paid_at: null,
        payment_failed_at: null,
        payment_status_updated_at: null,
        checkout_retry_available: false,
        checkout_retry_reason_code: 'not_issued',
        latest_payment_attempt: null,
        latest_payment_event: null,
      };

      service.getPaymentStatus('inv-2').subscribe((status) => {
        expect(status.latestPaymentAttempt).toBeNull();
        expect(status.latestPaymentEvent).toBeNull();
        expect(status.checkoutRetryAvailable).toBe(false);
      });

      const req = httpMock.expectOne('/api/v1/invoices/inv-2/payment-status');
      req.flush(response);
    });
  });

  describe('listPaymentEvents', () => {
    it('calls GET /invoices/:id/payment-events with limit and offset params', () => {
      const response = {
        items: [
          {
            payment_event_id: 'pe-1',
            payment_attempt_id: 'pa-1',
            stripe_event_id: 'evt_1',
            stripe_event_type: 'checkout.session.completed',
            stripe_checkout_session_id: 'cs_1',
            stripe_payment_intent_id: null,
            outcome: 'payment_succeeded',
            reason_code: 'checkout_paid',
            previous_invoice_status: 'issued',
            new_invoice_status: 'paid',
            attempt_previous_status: 'checkout_created',
            attempt_new_status: 'paid',
            amount_minor: 24000,
            currency_code: 'gbp',
            webhook_processing_status: 'processed',
            webhook_processing_reason: null,
            webhook_received_at: '2026-06-09T15:00:00Z',
            webhook_processed_at: '2026-06-09T15:00:01Z',
            created_at: '2026-06-09T15:00:00Z',
          },
        ],
        limit: 50,
        offset: 0,
      };

      service.listPaymentEvents('inv-1', { limit: 50, offset: 0 }).subscribe((result) => {
        expect(result.items.length).toBe(1);
        expect(result.items[0].paymentEventId).toBe('pe-1');
        expect(result.items[0].outcome).toBe('payment_succeeded');
        expect(result.items[0].reasonCode).toBe('checkout_paid');
        expect(result.items[0].amountMinor).toBe(24000);
        expect(result.items[0].webhookProcessingStatus).toBe('processed');
        expect(result.items[0].stripeEventId).toBe('evt_1');
        expect(result.items[0].previousInvoiceStatus).toBe('issued');
        expect(result.items[0].newInvoiceStatus).toBe('paid');
        expect(result.items[0].attemptPreviousStatus).toBe('checkout_created');
        expect(result.items[0].attemptNewStatus).toBe('paid');
        expect(result.limit).toBe(50);
        expect(result.offset).toBe(0);
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices/inv-1/payment-events');
      expect(req.request.method).toBe('GET');
      expect(req.request.params.get('limit')).toBe('50');
      expect(req.request.params.get('offset')).toBe('0');
      req.flush(response);
    });

    it('handles empty event history', () => {
      service.listPaymentEvents('inv-1', { limit: 50, offset: 0 }).subscribe((result) => {
        expect(result.items).toEqual([]);
        expect(result.limit).toBe(50);
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices/inv-1/payment-events');
      req.flush({ items: [], limit: 50, offset: 0 });
    });

    it('maps nullable amount and currency fields', () => {
      const response = {
        items: [
          {
            payment_event_id: 'pe-2',
            payment_attempt_id: 'pa-2',
            stripe_event_id: null,
            stripe_event_type: null,
            stripe_checkout_session_id: null,
            stripe_payment_intent_id: null,
            outcome: 'checkout_expired',
            reason_code: 'session_expired',
            previous_invoice_status: null,
            new_invoice_status: null,
            attempt_previous_status: 'checkout_created',
            attempt_new_status: 'expired',
            amount_minor: null,
            currency_code: null,
            webhook_processing_status: 'processed',
            webhook_processing_reason: null,
            webhook_received_at: null,
            webhook_processed_at: null,
            created_at: '2026-06-09T16:00:00Z',
          },
        ],
        limit: 50,
        offset: 0,
      };

      service.listPaymentEvents('inv-1', { limit: 50, offset: 0 }).subscribe((result) => {
        const event = result.items[0];
        expect(event.amountMinor).toBeNull();
        expect(event.currencyCode).toBeNull();
        expect(event.stripeEventId).toBeNull();
        expect(event.previousInvoiceStatus).toBeNull();
      });

      const req = httpMock.expectOne((r) => r.url === '/api/v1/invoices/inv-1/payment-events');
      req.flush(response);
    });
  });
});
